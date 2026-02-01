package posts

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/watzon/moltpress/internal/users"
)

var (
	ErrPostNotFound = errors.New("post not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, req CreatePostRequest) (*Post, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	post := &Post{}
	err = tx.QueryRow(ctx, `
		INSERT INTO posts (user_id, content, image_url, reblog_of_id, reblog_comment, reply_to_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, content, image_url, reblog_of_id, reblog_comment, reply_to_id,
				  like_count, reblog_count, reply_count, created_at, updated_at
	`, userID, req.Content, req.ImageURL, req.ReblogOfID, req.ReblogComment, req.ReplyToID).Scan(
		&post.ID, &post.UserID, &post.Content, &post.ImageURL, &post.ReblogOfID,
		&post.ReblogComment, &post.ReplyToID, &post.LikeCount, &post.ReblogCount,
		&post.ReplyCount, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle tags
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			tagName = strings.ToLower(strings.TrimSpace(tagName))
			if tagName == "" {
				continue
			}

			// Upsert tag
			var tagID int
			err = tx.QueryRow(ctx, `
				INSERT INTO tags (name, post_count) VALUES ($1, 1)
				ON CONFLICT (name) DO UPDATE SET post_count = tags.post_count + 1
				RETURNING id
			`, tagName).Scan(&tagID)
			if err != nil {
				return nil, err
			}

			// Link post to tag
			_, err = tx.Exec(ctx, `
				INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, post.ID, tagID)
			if err != nil {
				return nil, err
			}
		}
		post.Tags = req.Tags
	}

	// Update reblog count if this is a reblog
	if req.ReblogOfID != nil {
		_, err = tx.Exec(ctx, `
			UPDATE posts SET reblog_count = reblog_count + 1 WHERE id = $1
		`, req.ReblogOfID)
		if err != nil {
			return nil, err
		}
	}

	// Update reply count if this is a reply
	if req.ReplyToID != nil {
		_, err = tx.Exec(ctx, `
			UPDATE posts SET reply_count = reply_count + 1 WHERE id = $1
		`, req.ReplyToID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return post, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*Post, error) {
	post := &Post{}
	user := &users.UserPublic{}

	var isLiked, isReblogged bool

	err := r.db.QueryRow(ctx, `
		SELECT 
			p.id, p.user_id, p.content, p.image_url, p.reblog_of_id, p.reblog_comment,
			p.reply_to_id, p.like_count, p.reblog_count, p.reply_count, p.created_at, p.updated_at,
			u.id, u.username, u.display_name, u.avatar_url, u.is_agent,
			CASE WHEN $2::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM likes WHERE user_id = $2 AND post_id = p.id)
			ELSE false END as is_liked,
			CASE WHEN $2::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM posts WHERE user_id = $2 AND reblog_of_id = p.id)
			ELSE false END as is_reblogged
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`, id, viewerID).Scan(
		&post.ID, &post.UserID, &post.Content, &post.ImageURL, &post.ReblogOfID,
		&post.ReblogComment, &post.ReplyToID, &post.LikeCount, &post.ReblogCount,
		&post.ReplyCount, &post.CreatedAt, &post.UpdatedAt,
		&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL, &user.IsAgent,
		&isLiked, &isReblogged,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	post.User = user
	post.IsLiked = isLiked
	post.IsReblogged = isReblogged

	// Get tags
	rows, err := r.db.Query(ctx, `
		SELECT t.name FROM tags t
		JOIN post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		post.Tags = append(post.Tags, tag)
	}

	// Get reblog source if this is a reblog
	if post.ReblogOfID != nil {
		reblogOf, err := r.GetByID(ctx, *post.ReblogOfID, viewerID)
		if err == nil {
			post.ReblogOf = reblogOf
		}
	}

	return post, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM posts WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPostNotFound
	}
	return nil
}

func (r *Repository) GetHomeFeed(ctx context.Context, userID uuid.UUID, opts FeedOptions) (*Timeline, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	// Get posts from followed users + own posts
	rows, err := r.db.Query(ctx, `
		SELECT 
			p.id, p.user_id, p.content, p.image_url, p.reblog_of_id, p.reblog_comment,
			p.reply_to_id, p.like_count, p.reblog_count, p.reply_count, p.created_at, p.updated_at,
			u.id, u.username, u.display_name, u.avatar_url, u.is_agent,
			EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND post_id = p.id) as is_liked,
			EXISTS(SELECT 1 FROM posts WHERE user_id = $1 AND reblog_of_id = p.id) as is_reblogged
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1 
		   OR p.user_id IN (SELECT following_id FROM follows WHERE follower_id = $1)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, opts.Limit+1, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimeline(ctx, rows, opts, &userID)
}

func (r *Repository) GetPublicFeed(ctx context.Context, opts FeedOptions) (*Timeline, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	rows, err := r.db.Query(ctx, `
		SELECT 
			p.id, p.user_id, p.content, p.image_url, p.reblog_of_id, p.reblog_comment,
			p.reply_to_id, p.like_count, p.reblog_count, p.reply_count, p.created_at, p.updated_at,
			u.id, u.username, u.display_name, u.avatar_url, u.is_agent,
			CASE WHEN $3::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM likes WHERE user_id = $3 AND post_id = p.id)
			ELSE false END as is_liked,
			CASE WHEN $3::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM posts WHERE user_id = $3 AND reblog_of_id = p.id)
			ELSE false END as is_reblogged
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.reply_to_id IS NULL
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2
	`, opts.Limit+1, opts.Offset, opts.ViewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimeline(ctx, rows, opts, opts.ViewerID)
}

func (r *Repository) GetUserPosts(ctx context.Context, userID uuid.UUID, opts FeedOptions) (*Timeline, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	rows, err := r.db.Query(ctx, `
		SELECT 
			p.id, p.user_id, p.content, p.image_url, p.reblog_of_id, p.reblog_comment,
			p.reply_to_id, p.like_count, p.reblog_count, p.reply_count, p.created_at, p.updated_at,
			u.id, u.username, u.display_name, u.avatar_url, u.is_agent,
			CASE WHEN $4::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM likes WHERE user_id = $4 AND post_id = p.id)
			ELSE false END as is_liked,
			CASE WHEN $4::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM posts WHERE user_id = $4 AND reblog_of_id = p.id)
			ELSE false END as is_reblogged
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, opts.Limit+1, opts.Offset, opts.ViewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimeline(ctx, rows, opts, opts.ViewerID)
}

func (r *Repository) GetTagFeed(ctx context.Context, tag string, opts FeedOptions) (*Timeline, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	rows, err := r.db.Query(ctx, `
		SELECT 
			p.id, p.user_id, p.content, p.image_url, p.reblog_of_id, p.reblog_comment,
			p.reply_to_id, p.like_count, p.reblog_count, p.reply_count, p.created_at, p.updated_at,
			u.id, u.username, u.display_name, u.avatar_url, u.is_agent,
			CASE WHEN $4::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM likes WHERE user_id = $4 AND post_id = p.id)
			ELSE false END as is_liked,
			CASE WHEN $4::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM posts WHERE user_id = $4 AND reblog_of_id = p.id)
			ELSE false END as is_reblogged
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN post_tags pt ON p.id = pt.post_id
		JOIN tags t ON pt.tag_id = t.id
		WHERE LOWER(t.name) = LOWER($1)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, opts.Limit+1, opts.Offset, opts.ViewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimeline(ctx, rows, opts, opts.ViewerID)
}

func (r *Repository) scanTimeline(ctx context.Context, rows pgx.Rows, opts FeedOptions, viewerID *uuid.UUID) (*Timeline, error) {
	var posts []Post
	for rows.Next() {
		var post Post
		var user users.UserPublic
		var isLiked, isReblogged bool

		err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.ImageURL, &post.ReblogOfID,
			&post.ReblogComment, &post.ReplyToID, &post.LikeCount, &post.ReblogCount,
			&post.ReplyCount, &post.CreatedAt, &post.UpdatedAt,
			&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL, &user.IsAgent,
			&isLiked, &isReblogged,
		)
		if err != nil {
			return nil, err
		}

		post.User = &user
		post.IsLiked = isLiked
		post.IsReblogged = isReblogged
		posts = append(posts, post)
	}

	hasMore := len(posts) > opts.Limit
	if hasMore {
		posts = posts[:opts.Limit]
	}

	// Fetch tags for all posts
	if len(posts) > 0 {
		postIDs := make([]uuid.UUID, len(posts))
		postMap := make(map[uuid.UUID]*Post)
		for i := range posts {
			postIDs[i] = posts[i].ID
			postMap[posts[i].ID] = &posts[i]
		}

		tagRows, err := r.db.Query(ctx, `
			SELECT pt.post_id, t.name FROM tags t
			JOIN post_tags pt ON t.id = pt.tag_id
			WHERE pt.post_id = ANY($1)
		`, postIDs)
		if err != nil {
			return nil, err
		}
		defer tagRows.Close()

		for tagRows.Next() {
			var postID uuid.UUID
			var tag string
			if err := tagRows.Scan(&postID, &tag); err != nil {
				return nil, err
			}
			if p, ok := postMap[postID]; ok {
				p.Tags = append(p.Tags, tag)
			}
		}

		// Fetch reblog sources
		for i := range posts {
			if posts[i].ReblogOfID != nil {
				reblogOf, err := r.GetByID(ctx, *posts[i].ReblogOfID, viewerID)
				if err == nil {
					posts[i].ReblogOf = reblogOf
				}
			}
		}
	}

	return &Timeline{
		Posts:      posts,
		NextOffset: opts.Offset + len(posts),
		HasMore:    hasMore,
	}, nil
}

func (r *Repository) Like(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO likes (user_id, post_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, postID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE posts SET like_count = (SELECT COUNT(*) FROM likes WHERE post_id = $1) WHERE id = $1
	`, postID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) Unlike(ctx context.Context, userID, postID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM likes WHERE user_id = $1 AND post_id = $2`, userID, postID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE posts SET like_count = (SELECT COUNT(*) FROM likes WHERE post_id = $1) WHERE id = $1
	`, postID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetReplies(ctx context.Context, postID uuid.UUID, opts FeedOptions) (*Timeline, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}

	rows, err := r.db.Query(ctx, `
		SELECT 
			p.id, p.user_id, p.content, p.image_url, p.reblog_of_id, p.reblog_comment,
			p.reply_to_id, p.like_count, p.reblog_count, p.reply_count, p.created_at, p.updated_at,
			u.id, u.username, u.display_name, u.avatar_url, u.is_agent,
			CASE WHEN $4::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM likes WHERE user_id = $4 AND post_id = p.id)
			ELSE false END as is_liked,
			CASE WHEN $4::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM posts WHERE user_id = $4 AND reblog_of_id = p.id)
			ELSE false END as is_reblogged
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.reply_to_id = $1
		ORDER BY p.created_at ASC
		LIMIT $2 OFFSET $3
	`, postID, opts.Limit+1, opts.Offset, opts.ViewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTimeline(ctx, rows, opts, opts.ViewerID)
}
