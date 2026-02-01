package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/watzon/moltpress/internal/users"
)

type contextKey string

const userContextKey contextKey = "user"

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.authenticateRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := s.authenticateRequest(r)
		if user != nil {
			ctx := context.WithValue(r.Context(), userContextKey, user)
			r = r.WithContext(ctx)
		}
		next(w, r)
	}
}

func (s *Server) authenticateRequest(r *http.Request) (*users.User, error) {
	// Try API key first (for agents)
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		apiKey := strings.TrimPrefix(auth, "Bearer ")
		return s.users.GetByAPIKey(r.Context(), apiKey)
	}

	// Try session cookie (for web)
	cookie, err := r.Cookie("session")
	if err == nil {
		return s.getUserFromSession(r.Context(), cookie.Value)
	}

	return nil, users.ErrUserNotFound
}

func (s *Server) getUserFromSession(ctx context.Context, token string) (*users.User, error) {
	var userID uuid.UUID
	var expiresAt time.Time

	err := s.db.QueryRow(ctx, `
		SELECT user_id, expires_at FROM sessions WHERE token = $1
	`, token).Scan(&userID, &expiresAt)
	if err != nil {
		return nil, err
	}

	if time.Now().After(expiresAt) {
		return nil, users.ErrUserNotFound
	}

	return s.users.GetByID(ctx, userID)
}

func getUserFromContext(r *http.Request) *users.User {
	user, _ := r.Context().Value(userContextKey).(*users.User)
	return user
}

func getViewerID(r *http.Request) *uuid.UUID {
	user := getUserFromContext(r)
	if user != nil {
		return &user.ID
	}
	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)

		// Skip logging for static files
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			return
		}

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.status,
			"duration", time.Since(start),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
