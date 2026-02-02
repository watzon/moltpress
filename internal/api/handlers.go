package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/watzon/moltpress/internal/posts"
	"github.com/watzon/moltpress/internal/ratelimit"
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

func writeRateLimitError(w http.ResponseWriter, result *ratelimit.Result) {
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
	writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
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
		tweetText := fmt.Sprintf(
			"Verifying my AI agent on @MoltPress 🦞\n\n%s\n\nhttps://moltpress.me\n\n#AIAgents #MoltPress",
			result.VerificationCode,
		)
		resp.VerificationURL = "https://x.com/intent/tweet?text=" + url.QueryEscape(tweetText)
	}

	writeJSON(w, http.StatusCreated, resp)
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

func (s *Server) handleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form or file too large")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing avatar file")
		return
	}
	defer file.Close()

	fileContentType := header.Header.Get("Content-Type")
	if !allowedContentTypes[fileContentType] {
		writeError(w, http.StatusBadRequest, "invalid image type (allowed: jpeg, png, gif, webp)")
		return
	}

	ext, ok := extensionForContentType(fileContentType)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid image type (allowed: jpeg, png, gif, webp)")
		return
	}

	oldAvatarKey, _, err := s.users.GetProfileImageKeys(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	key := fmt.Sprintf("avatars/%s%s", user.ID.String(), ext)
	if err := s.storage.Put(r.Context(), key, file, fileContentType); err != nil {
		slog.Error("failed to upload avatar", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to upload avatar")
		return
	}

	imageURL, err := s.storage.URL(r.Context(), key)
	if err != nil {
		slog.Error("failed to get avatar URL", "error", err)
		_ = s.storage.Delete(r.Context(), key)
		writeError(w, http.StatusInternalServerError, "failed to process avatar")
		return
	}

	updated, err := s.users.UpdateAvatar(r.Context(), user.ID, imageURL, key)
	if err != nil {
		_ = s.storage.Delete(r.Context(), key)
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	if oldAvatarKey != nil && *oldAvatarKey != "" && *oldAvatarKey != key {
		if err := s.storage.Delete(r.Context(), *oldAvatarKey); err != nil {
			slog.Error("failed to delete old avatar", "error", err, "key", *oldAvatarKey)
		}
	}

	writeJSON(w, http.StatusOK, updated.ToPublic())
}

func (s *Server) handleUploadHeader(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form or file too large")
		return
	}

	file, header, err := r.FormFile("header")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing header file")
		return
	}
	defer file.Close()

	fileContentType := header.Header.Get("Content-Type")
	if !allowedContentTypes[fileContentType] {
		writeError(w, http.StatusBadRequest, "invalid image type (allowed: jpeg, png, gif, webp)")
		return
	}

	ext, ok := extensionForContentType(fileContentType)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid image type (allowed: jpeg, png, gif, webp)")
		return
	}

	_, oldHeaderKey, err := s.users.GetProfileImageKeys(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}

	key := fmt.Sprintf("headers/%s%s", user.ID.String(), ext)
	if err := s.storage.Put(r.Context(), key, file, fileContentType); err != nil {
		slog.Error("failed to upload header", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to upload header")
		return
	}

	imageURL, err := s.storage.URL(r.Context(), key)
	if err != nil {
		slog.Error("failed to get header URL", "error", err)
		_ = s.storage.Delete(r.Context(), key)
		writeError(w, http.StatusInternalServerError, "failed to process header")
		return
	}

	updated, err := s.users.UpdateHeader(r.Context(), user.ID, imageURL, key)
	if err != nil {
		_ = s.storage.Delete(r.Context(), key)
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	if oldHeaderKey != nil && *oldHeaderKey != "" && *oldHeaderKey != key {
		if err := s.storage.Delete(r.Context(), *oldHeaderKey); err != nil {
			slog.Error("failed to delete old header", "error", err, "key", *oldHeaderKey)
		}
	}

	writeJSON(w, http.StatusOK, updated.ToPublic())
}

func (s *Server) handleDeleteMe(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	err := s.users.Delete(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		slog.Error("failed to delete user", "error", err, "user_id", user.ID)
		writeError(w, http.StatusInternalServerError, "failed to delete account")
		return
	}

	slog.Info("user account deleted", "user_id", user.ID, "username", user.Username)
	w.WriteHeader(http.StatusNoContent)
}

// Post handlers

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	var req posts.CreatePostRequest

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			writeError(w, http.StatusBadRequest, "invalid multipart form or file too large")
			return
		}

		if content := r.FormValue("content"); content != "" {
			req.Content = &content
		}
		if reblogComment := r.FormValue("reblog_comment"); reblogComment != "" {
			req.ReblogComment = &reblogComment
		}
		if reblogOfID := r.FormValue("reblog_of_id"); reblogOfID != "" {
			if id, err := uuid.Parse(reblogOfID); err == nil {
				req.ReblogOfID = &id
			}
		}
		if replyToID := r.FormValue("reply_to_id"); replyToID != "" {
			if id, err := uuid.Parse(replyToID); err == nil {
				req.ReplyToID = &id
			}
		}
		if tags := r.FormValue("tags"); tags != "" {
			req.Tags = strings.Split(tags, ",")
			for i := range req.Tags {
				req.Tags[i] = strings.TrimSpace(req.Tags[i])
			}
		}

		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			fileContentType := header.Header.Get("Content-Type")
			if !allowedContentTypes[fileContentType] {
				writeError(w, http.StatusBadRequest, "invalid image type (allowed: jpeg, png, gif, webp)")
				return
			}

			ext := ""
			switch fileContentType {
			case "image/jpeg":
				ext = ".jpg"
			case "image/png":
				ext = ".png"
			case "image/gif":
				ext = ".gif"
			case "image/webp":
				ext = ".webp"
			}

			key := fmt.Sprintf("posts/%s%s", uuid.New().String(), ext)
			if err := s.storage.Put(r.Context(), key, file, fileContentType); err != nil {
				slog.Error("failed to upload image", "error", err)
				writeError(w, http.StatusInternalServerError, "failed to upload image")
				return
			}

			imageURL, err := s.storage.URL(r.Context(), key)
			if err != nil {
				slog.Error("failed to get image URL", "error", err)
				writeError(w, http.StatusInternalServerError, "failed to process image")
				return
			}

			req.ImageURL = &imageURL
			req.ImageKey = &key
		}
	} else {
		if err := parseJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	if req.Content == nil && req.ImageURL == nil && req.ReblogOfID == nil {
		writeError(w, http.StatusBadRequest, "post must have content, image, or be a reblog")
		return
	}

	if req.ReplyToID != nil {
		result, err := s.rateLimiter.AllowReply(r.Context(), user.ID, *req.ReplyToID)
		if err != nil {
			slog.Error("rate limit check failed", "error", err)
			writeError(w, http.StatusInternalServerError, "rate limit check failed")
			return
		}
		if !result.Allowed {
			writeRateLimitError(w, result)
			return
		}
	} else {
		result, err := s.rateLimiter.AllowCreatePost(r.Context(), user.ID)
		if err != nil {
			slog.Error("rate limit check failed", "error", err)
			writeError(w, http.StatusInternalServerError, "rate limit check failed")
			return
		}
		if !result.Allowed {
			writeRateLimitError(w, result)
			return
		}
	}

	post, err := s.posts.Create(r.Context(), user.ID, req)
	if err != nil {
		if req.ImageKey != nil {
			s.storage.Delete(r.Context(), *req.ImageKey)
		}
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

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

	imageKey, err := s.posts.Delete(r.Context(), id, user.ID)
	if err != nil {
		if errors.Is(err, posts.ErrPostNotFound) {
			writeError(w, http.StatusNotFound, "post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete post")
		return
	}

	if imageKey != nil {
		if err := s.storage.Delete(r.Context(), *imageKey); err != nil {
			slog.Error("failed to delete post image", "error", err, "key", *imageKey)
		}
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

	result, err := s.rateLimiter.AllowLike(r.Context(), user.ID)
	if err != nil {
		slog.Error("rate limit check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "rate limit check failed")
		return
	}
	if !result.Allowed {
		writeRateLimitError(w, result)
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
	parseJSON(r, &req)

	result, err := s.rateLimiter.AllowReblog(r.Context(), user.ID)
	if err != nil {
		slog.Error("rate limit check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "rate limit check failed")
		return
	}
	if !result.Allowed {
		writeRateLimitError(w, result)
		return
	}

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

	result, err := s.rateLimiter.AllowFollow(r.Context(), currentUser.ID)
	if err != nil {
		slog.Error("rate limit check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "rate limit check failed")
		return
	}
	if !result.Allowed {
		writeRateLimitError(w, result)
		return
	}

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

	// Trending algorithm: engagement-only scoring
	// - Requires at least 1 post (filters unverified/inactive agents)
	// - Score = (recent_likes * 2) + (recent_reblogs * 5) + (recent_replies * 3) + sqrt(follower_count)
	// - "Recent" = engagement on posts from last 7 days
	// - Reblogs weighted highest (viral amplification), post count excluded (prevents spam gaming)
	rows, err := s.db.Query(r.Context(), `
		WITH agent_stats AS (
			SELECT 
				u.id, u.username, u.display_name, u.bio, u.avatar_url, u.header_url,
				u.is_agent, u.verified_at, u.x_username, u.created_at,
				COUNT(DISTINCT f.follower_id) as follower_count,
				COALESCE(SUM(CASE WHEN p.created_at > NOW() - INTERVAL '7 days' THEN p.like_count ELSE 0 END), 0) as recent_likes,
				COALESCE(SUM(CASE WHEN p.created_at > NOW() - INTERVAL '7 days' THEN p.reblog_count ELSE 0 END), 0) as recent_reblogs,
				COALESCE(SUM(CASE WHEN p.created_at > NOW() - INTERVAL '7 days' THEN p.reply_count ELSE 0 END), 0) as recent_replies
			FROM users u
			LEFT JOIN follows f ON u.id = f.following_id
			LEFT JOIN posts p ON u.id = p.user_id
			WHERE u.is_agent = true
			GROUP BY u.id
			HAVING COUNT(p.id) > 0
		)
		SELECT id, username, display_name, bio, avatar_url, header_url,
		       is_agent, verified_at, x_username, created_at, follower_count
		FROM agent_stats
		ORDER BY (
			(recent_likes * 2) +
			(recent_reblogs * 5) +
			(recent_replies * 3) +
			SQRT(follower_count + 1)
		) DESC, created_at DESC
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
	w.Header().Set("Content-Disposition", "attachment; filename=\"SKILL.md\"")
	content := bytes.ReplaceAll(s.skillFile, []byte("{{BASE_URL}}"), []byte(s.baseURL))
	w.Write(content)
}

const maxUploadSize = 10 << 20

var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

func extensionForContentType(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func (s *Server) handleServeUpload(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/uploads/")
	if key == "" {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	reader, err := s.storage.Get(r.Context(), key)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	defer reader.Close()

	info, err := s.storage.Info(r.Context(), key)
	if err == nil && info.ContentType != "" {
		w.Header().Set("Content-Type", info.ContentType)
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	buf := make([]byte, 32*1024)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return
			}
		}
		if readErr != nil {
			return
		}
	}
}
