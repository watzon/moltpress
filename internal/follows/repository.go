package follows

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/watzon/moltpress/internal/users"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	if followerID == followingID {
		return nil // Can't follow yourself
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO follows (follower_id, following_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, followerID, followingID)
	return err
}

func (r *Repository) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM follows WHERE follower_id = $1 AND following_id = $2
	`, followerID, followingID)
	return err
}

func (r *Repository) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)
	`, followerID, followingID).Scan(&exists)
	return exists, err
}

func (r *Repository) GetFollowers(ctx context.Context, userID uuid.UUID, limit, offset int, viewerID *uuid.UUID) ([]users.UserPublic, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(ctx, `
		SELECT 
			u.id, u.username, u.display_name, u.bio, u.avatar_url, u.is_agent, u.created_at,
			(SELECT COUNT(*) FROM follows WHERE following_id = u.id) as follower_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
			CASE WHEN $4::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM follows WHERE follower_id = $4 AND following_id = u.id)
			ELSE false END as is_following
		FROM users u
		JOIN follows f ON u.id = f.follower_id
		WHERE f.following_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []users.UserPublic
	for rows.Next() {
		var u users.UserPublic
		err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.Bio, &u.AvatarURL, &u.IsAgent, &u.CreatedAt,
			&u.FollowerCount, &u.FollowingCount, &u.IsFollowing,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}

func (r *Repository) GetFollowing(ctx context.Context, userID uuid.UUID, limit, offset int, viewerID *uuid.UUID) ([]users.UserPublic, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(ctx, `
		SELECT 
			u.id, u.username, u.display_name, u.bio, u.avatar_url, u.is_agent, u.created_at,
			(SELECT COUNT(*) FROM follows WHERE following_id = u.id) as follower_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
			CASE WHEN $4::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM follows WHERE follower_id = $4 AND following_id = u.id)
			ELSE false END as is_following
		FROM users u
		JOIN follows f ON u.id = f.following_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []users.UserPublic
	for rows.Next() {
		var u users.UserPublic
		err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.Bio, &u.AvatarURL, &u.IsAgent, &u.CreatedAt,
			&u.FollowerCount, &u.FollowingCount, &u.IsFollowing,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}
