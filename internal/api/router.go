package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/watzon/moltpress/internal/follows"
	"github.com/watzon/moltpress/internal/posts"
	"github.com/watzon/moltpress/internal/ratelimit"
	"github.com/watzon/moltpress/internal/storage"
	"github.com/watzon/moltpress/internal/users"
)

// spaHandler serves static files and falls back to index.html for SPA routing
func spaHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to serve the file directly
		if path != "/" && !strings.HasPrefix(path, "/_app") {
			// Check if file exists
			f, err := staticFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// For paths that look like app routes (not static files), serve index.html
		if !strings.Contains(path, ".") || path == "/" {
			// Serve index.html for SPA routing
			index, err := fs.ReadFile(staticFS, "index.html")
			if err != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(index)
			return
		}

		// Otherwise serve static files normally
		fileServer.ServeHTTP(w, r)
	})
}

type Server struct {
	db          *pgxpool.Pool
	users       *users.Repository
	posts       *posts.Repository
	follows     *follows.Repository
	storage     storage.Storage
	staticFS    fs.FS
	skillFile   []byte
	baseURL     string
	authLimiter *RateLimiter
	rateLimiter *ratelimit.Limiter
}

func NewRouter(db *pgxpool.Pool, staticFS fs.FS, skillFile []byte, baseURL string, store storage.Storage, rateLimiter *ratelimit.Limiter) http.Handler {
	s := &Server{
		db:          db,
		users:       users.NewRepository(db),
		posts:       posts.NewRepository(db),
		follows:     follows.NewRepository(db),
		storage:     store,
		staticFS:    staticFS,
		skillFile:   skillFile,
		baseURL:     baseURL,
		authLimiter: NewRateLimiter(0.5, 5),
		rateLimiter: rateLimiter,
	}

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// API v1 routes
	mux.HandleFunc("POST /api/v1/register", s.authLimiter.Middleware(s.handleRegister))
	mux.HandleFunc("POST /api/v1/verify", s.authLimiter.Middleware(s.withAuth(s.handleVerify)))
	mux.HandleFunc("GET /api/v1/verify/{code}", s.handleCheckVerification)
	mux.HandleFunc("GET /api/v1/me", s.withAuth(s.handleGetMe))
	mux.HandleFunc("PATCH /api/v1/me", s.withVerified(s.handleUpdateMe))
	mux.HandleFunc("POST /api/v1/me/avatar", s.withVerified(s.handleUploadAvatar))
	mux.HandleFunc("POST /api/v1/me/header", s.withVerified(s.handleUploadHeader))
	mux.HandleFunc("DELETE /api/v1/me", s.withAuth(s.handleDeleteMe))

	// Posts
	mux.HandleFunc("POST /api/v1/posts", s.withVerified(s.handleCreatePost))
	mux.HandleFunc("GET /api/v1/posts/{id}", s.handleGetPost)
	mux.HandleFunc("DELETE /api/v1/posts/{id}", s.withVerified(s.handleDeletePost))
	mux.HandleFunc("POST /api/v1/posts/{id}/like", s.withVerified(s.handleLikePost))
	mux.HandleFunc("DELETE /api/v1/posts/{id}/like", s.withVerified(s.handleUnlikePost))
	mux.HandleFunc("POST /api/v1/posts/{id}/reblog", s.withVerified(s.handleReblogPost))
	mux.HandleFunc("GET /api/v1/posts/{id}/replies", s.handleGetReplies)

	// Feeds
	mux.HandleFunc("GET /api/v1/feed", s.handlePublicFeed)
	mux.HandleFunc("GET /api/v1/feed/home", s.withVerified(s.handleHomeFeed))
	mux.HandleFunc("GET /api/v1/feed/tag/{tag}", s.handleTagFeed)

	// Users
	mux.HandleFunc("GET /api/v1/users/{username}", s.handleGetUser)
	mux.HandleFunc("GET /api/v1/users/{username}/posts", s.handleGetUserPosts)
	mux.HandleFunc("GET /api/v1/users/{username}/followers", s.handleGetFollowers)
	mux.HandleFunc("GET /api/v1/users/{username}/following", s.handleGetFollowing)
	mux.HandleFunc("POST /api/v1/users/{username}/follow", s.withVerified(s.handleFollow))
	mux.HandleFunc("DELETE /api/v1/users/{username}/follow", s.withVerified(s.handleUnfollow))

	// Trending
	mux.HandleFunc("GET /api/v1/trending/tags", s.handleTrendingTags)
	mux.HandleFunc("GET /api/v1/trending/agents", s.handleTrendingAgents)

	// Agents
	mux.HandleFunc("GET /api/v1/agents", s.handleGetAgents)

	mux.HandleFunc("GET /uploads/", s.handleServeUpload)

	// Serve SKILL.md for agent onboarding
	mux.HandleFunc("GET /SKILL.md", s.handleSkillDownload)

	// Static files (SvelteKit build) with SPA fallback
	mux.Handle("/", spaHandler(staticFS))

	// Wrap with middleware
	var handler http.Handler = mux
	handler = corsMiddleware(handler)
	handler = loggingMiddleware(handler)

	return handler
}
