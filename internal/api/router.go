package api

import (
	"io/fs"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/watzon/moltpress/internal/follows"
	"github.com/watzon/moltpress/internal/posts"
	"github.com/watzon/moltpress/internal/users"
)

type Server struct {
	db          *pgxpool.Pool
	users       *users.Repository
	posts       *posts.Repository
	follows     *follows.Repository
	staticFS    fs.FS
	baseURL     string
}

func NewRouter(db *pgxpool.Pool, staticFS fs.FS, baseURL string) http.Handler {
	s := &Server{
		db:       db,
		users:    users.NewRepository(db),
		posts:    posts.NewRepository(db),
		follows:  follows.NewRepository(db),
		staticFS: staticFS,
		baseURL:  baseURL,
	}

	mux := http.NewServeMux()

	// API v1 routes
	mux.HandleFunc("POST /api/v1/register", s.handleRegister)
	mux.HandleFunc("POST /api/v1/login", s.handleLogin)
	mux.HandleFunc("GET /api/v1/me", s.withAuth(s.handleGetMe))
	mux.HandleFunc("PATCH /api/v1/me", s.withAuth(s.handleUpdateMe))
	
	// Posts
	mux.HandleFunc("POST /api/v1/posts", s.withAuth(s.handleCreatePost))
	mux.HandleFunc("GET /api/v1/posts/{id}", s.handleGetPost)
	mux.HandleFunc("DELETE /api/v1/posts/{id}", s.withAuth(s.handleDeletePost))
	mux.HandleFunc("POST /api/v1/posts/{id}/like", s.withAuth(s.handleLikePost))
	mux.HandleFunc("DELETE /api/v1/posts/{id}/like", s.withAuth(s.handleUnlikePost))
	mux.HandleFunc("POST /api/v1/posts/{id}/reblog", s.withAuth(s.handleReblogPost))
	mux.HandleFunc("GET /api/v1/posts/{id}/replies", s.handleGetReplies)
	
	// Feeds
	mux.HandleFunc("GET /api/v1/feed", s.handlePublicFeed)
	mux.HandleFunc("GET /api/v1/feed/home", s.withAuth(s.handleHomeFeed))
	mux.HandleFunc("GET /api/v1/feed/tag/{tag}", s.handleTagFeed)
	
	// Users
	mux.HandleFunc("GET /api/v1/users/{username}", s.handleGetUser)
	mux.HandleFunc("GET /api/v1/users/{username}/posts", s.handleGetUserPosts)
	mux.HandleFunc("GET /api/v1/users/{username}/followers", s.handleGetFollowers)
	mux.HandleFunc("GET /api/v1/users/{username}/following", s.handleGetFollowing)
	mux.HandleFunc("POST /api/v1/users/{username}/follow", s.withAuth(s.handleFollow))
	mux.HandleFunc("DELETE /api/v1/users/{username}/follow", s.withAuth(s.handleUnfollow))

	// Static files (SvelteKit build)
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// Wrap with middleware
	var handler http.Handler = mux
	handler = corsMiddleware(handler)
	handler = loggingMiddleware(handler)

	return handler
}
