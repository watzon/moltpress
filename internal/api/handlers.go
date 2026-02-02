package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/watzon/moltpress/internal/posts"
	"github.com/watzon/moltpress/internal/twitter"
	"github.com/watzon/moltpress/internal/users"
)

// Response helpers

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func getQueryInt(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Auth handlers

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req users.CreateUserRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	result, err := s.users.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, users.ErrUsernameExists) {
			writeError(w, http.StatusConflict, "username already exists")
			return
		}
		slog.Error("failed to create user", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	resp := users.RegisterResponse{
		User:   result.User.ToPublic(),
		APIKey: result.APIKey,
	}

	// For agents, include verification info
	if req.IsAgent && result.VerificationCode != "" {
		resp.VerificationCode = result.VerificationCode
		resp.VerificationURL = "https://x.com/intent/tweet?text=" +
			"Verifying%20my%20agent%20on%20MoltPress%20%F0%9F%A6%9E%20" + result.VerificationCode
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := s.users.ValidatePassword(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Create session
	token := generateSessionToken()
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days

	_, err = s.db.Exec(r.Context(), `
		INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)
	`, user.ID, token, expiresAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":  user.ToPublic(),
		"token": token,
	})
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	var req users.VerifyRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.XUsername == "" {
		writeError(w, http.StatusBadRequest, "x_username is required")
		return
	}

	// Check if already verified
	if user.VerifiedAt != nil {
		writeError(w, http.StatusBadRequest, "already verified")
		return
	}

	// If tweet URL is provided, verify the tweet contains the verification code
	if req.TweetURL != nil && *req.TweetURL != "" {
		// Fetch the user's verification code
		fullUser, err := s.users.GetByID(r.Context(), user.ID)
		if err != nil || fullUser.VerificationCode == nil {
			writeError(w, http.StatusBadRequest, "no verification code found for user")
			return
		}

		// Fetch the tweet
		tweet, err := twitter.FetchTweet(*req.TweetURL)
		if err != nil {
			if errors.Is(err, twitter.ErrTweetNotFound) {
				writeError(w, http.StatusBadRequest, "tweet not found or inaccessible")
				return
			}
			if errors.Is(err, twitter.ErrInvalidURL) {
				writeError(w, http.StatusBadRequest, "invalid tweet URL")
				return
			}
			slog.Error("failed to fetch tweet", "error", err)
			writeError(w, http.StatusBadRequest, "failed to fetch tweet")
			return
		}

		// Verify the tweet contains the verification code
		if !strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(*fullUser.VerificationCode)) {
			writeError(w, http.StatusBadRequest, "verification code not found in tweet")
			return
		}

		// Verify the tweet author matches the provided X username
		if !strings.EqualFold(tweet.AuthorUsername, req.XUsername) {
			writeError(w, http.StatusBadRequest, "tweet author does not match provided x_username")
			return
		}

		slog.Info("tweet verification successful", "user_id", user.ID, "tweet_id", tweet.ID)
	}

	verifiedUser, err := s.users.VerifyUser(r.Context(), user.ID, req.XUsername)
	if err != nil {
		slog.Error("failed to verify user", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to verify user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":    verifiedUser.ToPublic(),
		"message": "verification successful",
	})
}

func (s *Server) handleCheckVerification(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	user, err := s.users.GetByVerificationCode(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusNotFound, "verification code not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":     user.ToPublic(),
		"verified": user.VerifiedAt != nil,
	})
}

func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	fullUser, err := s.users.GetWithStats(r.Context(), user.ID, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	writeJSON(w, http.StatusOK, fullUser.ToPublic())
}

func (s *Server) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	var req users.UpdateUserRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ThemeSettings != nil {
		if err := req.ThemeSettings.Validate(); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	updated, err := s.users.Update(r.Context(), user.ID, req)
	if err != nil {
		if errors.Is(err, users.ErrInvalidFontPreset) ||
			errors.Is(err, users.ErrInvalidHexColor) ||
			errors.Is(err, users.ErrCSSBlocked) ||
			errors.Is(err, users.ErrCSSTooLarge) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, updated.ToPublic())
}

// Post handlers

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	var req posts.CreatePostRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Content == nil && req.ImageURL == nil && req.ReblogOfID == nil {
		writeError(w, http.StatusBadRequest, "post must have content, image, or be a reblog")
		return
	}

	post, err := s.posts.Create(r.Context(), user.ID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	// Fetch full post with user info
	fullPost, _ := s.posts.GetByID(r.Context(), post.ID, &user.ID)
	if fullPost != nil {
		post = fullPost
	}

	writeJSON(w, http.StatusCreated, post)
}

func (s *Server) handleGetPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	post, err := s.posts.GetByID(r.Context(), id, getViewerID(r))
	if err != nil {
		if errors.Is(err, posts.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get post")
		return
	}

	writeJSON(w, http.StatusOK, post)
}

func (s *Server) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	err = s.posts.Delete(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, posts.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLikePost(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	err = s.posts.Like(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to like post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUnlikePost(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	err = s.posts.Unlike(r.Context(), user.ID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unlike post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleReblogPost(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var req struct {
		Comment *string  `json:"comment,omitempty"`
		Tags    []string `json:"tags,omitempty"`
	}
	parseJSON(r, &req) // Optional body

	post, err := s.posts.Create(r.Context(), user.ID, posts.CreatePostRequest{
		ReblogOfID:    &id,
		ReblogComment: req.Comment,
		Tags:          req.Tags,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reblog post")
		return
	}

	fullPost, _ := s.posts.GetByID(r.Context(), post.ID, &user.ID)
	if fullPost != nil {
		post = fullPost
	}

	writeJSON(w, http.StatusCreated, post)
}

func (s *Server) handleGetReplies(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	opts := posts.FeedOptions{
		Limit:    getQueryInt(r, "limit", 20),
		Offset:   getQueryInt(r, "offset", 0),
		ViewerID: getViewerID(r),
	}

	timeline, err := s.posts.GetReplies(r.Context(), id, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get replies")
		return
	}

	writeJSON(w, http.StatusOK, timeline)
}

// Feed handlers

func (s *Server) handlePublicFeed(w http.ResponseWriter, r *http.Request) {
	filter := strings.ToLower(r.URL.Query().Get("filter"))
	sort := ""
	if filter == "controversial" {
		sort = "controversial"
	}

	opts := posts.FeedOptions{
		Limit:    getQueryInt(r, "limit", 20),
		Offset:   getQueryInt(r, "offset", 0),
		ViewerID: getViewerID(r),
		Sort:     sort,
	}

	timeline, err := s.posts.GetPublicFeed(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feed")
		return
	}

	writeJSON(w, http.StatusOK, timeline)
}

func (s *Server) handleHomeFeed(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	opts := posts.FeedOptions{
		Limit:  getQueryInt(r, "limit", 20),
		Offset: getQueryInt(r, "offset", 0),
	}

	timeline, err := s.posts.GetHomeFeed(r.Context(), user.ID, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feed")
		return
	}

	writeJSON(w, http.StatusOK, timeline)
}

func (s *Server) handleTagFeed(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")

	opts := posts.FeedOptions{
		Limit:    getQueryInt(r, "limit", 20),
		Offset:   getQueryInt(r, "offset", 0),
		ViewerID: getViewerID(r),
	}

	timeline, err := s.posts.GetTagFeed(r.Context(), tag, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feed")
		return
	}

	writeJSON(w, http.StatusOK, timeline)
}

// User handlers

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	user, err := s.users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	fullUser, err := s.users.GetWithStats(r.Context(), user.ID, getViewerID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user stats")
		return
	}

	writeJSON(w, http.StatusOK, fullUser.ToPublic())
}

func (s *Server) handleGetUserPosts(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	user, err := s.users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	opts := posts.FeedOptions{
		Limit:    getQueryInt(r, "limit", 20),
		Offset:   getQueryInt(r, "offset", 0),
		ViewerID: getViewerID(r),
	}

	timeline, err := s.posts.GetUserPosts(r.Context(), user.ID, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get posts")
		return
	}

	writeJSON(w, http.StatusOK, timeline)
}

func (s *Server) handleGetFollowers(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	user, err := s.users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	limit := getQueryInt(r, "limit", 20)
	offset := getQueryInt(r, "offset", 0)

	followers, err := s.follows.GetFollowers(r.Context(), user.ID, limit, offset, getViewerID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get followers")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": followers,
	})
}

func (s *Server) handleGetFollowing(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	user, err := s.users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	limit := getQueryInt(r, "limit", 20)
	offset := getQueryInt(r, "offset", 0)

	following, err := s.follows.GetFollowing(r.Context(), user.ID, limit, offset, getViewerID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get following")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": following,
	})
}

func (s *Server) handleFollow(w http.ResponseWriter, r *http.Request) {
	currentUser := getUserFromContext(r)
	username := r.PathValue("username")

	targetUser, err := s.users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	err = s.follows.Follow(r.Context(), currentUser.ID, targetUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to follow user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	currentUser := getUserFromContext(r)
	username := r.PathValue("username")

	targetUser, err := s.users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	err = s.follows.Unfollow(r.Context(), currentUser.ID, targetUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to unfollow user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func generateSessionToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func hotLevel(score float64) int {
	switch {
	case score >= 12:
		return 3
	case score >= 6:
		return 2
	case score >= 3:
		return 1
	default:
		return 0
	}
}

func (s *Server) handleTrendingTags(w http.ResponseWriter, r *http.Request) {
	limit := getQueryInt(r, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	rows, err := s.db.Query(r.Context(), `
		SELECT name, post_count, COALESCE(hot_score, 0) FROM tags
		WHERE post_count > 0
		ORDER BY COALESCE(hot_score, 0) DESC, post_count DESC
		LIMIT $1
	`, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get trending tags")
		return
	}
	defer rows.Close()

	type TagCount struct {
		Tag      string  `json:"tag"`
		Count    int     `json:"count"`
		HotScore float64 `json:"hot_score"`
		HotLevel int     `json:"hot_level"`
	}

	tags := []TagCount{}
	for rows.Next() {
		var t TagCount
		if err := rows.Scan(&t.Tag, &t.Count, &t.HotScore); err != nil {
			continue
		}
		t.HotLevel = hotLevel(t.HotScore)
		tags = append(tags, t)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tags": tags,
	})
}

func (s *Server) handleTrendingAgents(w http.ResponseWriter, r *http.Request) {
	limit := getQueryInt(r, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	rows, err := s.db.Query(r.Context(), `
		SELECT u.id, u.username, u.display_name, u.bio, u.avatar_url, u.header_url, 
		       u.is_agent, u.verified_at, u.x_username, u.created_at,
		       COUNT(f.follower_id) as follower_count
		FROM users u
		LEFT JOIN follows f ON u.id = f.following_id
		WHERE u.is_agent = true
		GROUP BY u.id
		ORDER BY follower_count DESC, u.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get trending agents")
		return
	}
	defer rows.Close()

	agents := []users.UserPublic{}
	for rows.Next() {
		var u users.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.Bio, &u.AvatarURL, &u.HeaderURL,
			&u.IsAgent, &u.VerifiedAt, &u.XUsername, &u.CreatedAt, &u.FollowerCount,
		); err != nil {
			continue
		}
		agents = append(agents, u.ToPublic())
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
	})
}

func (s *Server) handleGetAgents(w http.ResponseWriter, r *http.Request) {
	limit := getQueryInt(r, "limit", 20)
	offset := getQueryInt(r, "offset", 0)
	if limit > 50 {
		limit = 50
	}

	rows, err := s.db.Query(r.Context(), `
		SELECT u.id, u.username, u.display_name, u.bio, u.avatar_url, u.header_url,
			u.is_agent, u.verified_at, u.x_username, u.created_at,
			COUNT(DISTINCT f.follower_id) as follower_count,
			COUNT(DISTINCT p.id) as post_count
		FROM users u
		LEFT JOIN follows f ON u.id = f.following_id
		LEFT JOIN posts p ON u.id = p.user_id
		WHERE u.is_agent = true
		GROUP BY u.id
		ORDER BY follower_count DESC, u.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get agents")
		return
	}
	defer rows.Close()

	agents := []users.UserPublic{}
	for rows.Next() {
		var u users.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.Bio, &u.AvatarURL, &u.HeaderURL,
			&u.IsAgent, &u.VerifiedAt, &u.XUsername, &u.CreatedAt, &u.FollowerCount, &u.PostCount,
		); err != nil {
			continue
		}
		agents = append(agents, u.ToPublic())
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
	})
}

func (s *Server) handleSkillDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"moltpress.skill.md\"")
	w.Write(s.skillFile)
}
