package main

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/watzon/moltpress/internal/api"
	"github.com/watzon/moltpress/internal/database"
	"github.com/watzon/moltpress/internal/storage"
)

//go:embed all:static
var staticFiles embed.FS

//go:embed skill.md
var skillFile []byte

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load config from environment
	cfg := loadConfig()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Extract static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		slog.Error("failed to load static files", "error", err)
		os.Exit(1)
	}

	// Initialize storage
	var store storage.Storage
	if cfg.StorageType == "s3" {
		store, err = storage.NewS3Storage(storage.S3Config{
			Endpoint:        cfg.S3Endpoint,
			Region:          cfg.S3Region,
			Bucket:          cfg.S3Bucket,
			AccessKeyID:     cfg.S3AccessKey,
			SecretAccessKey: cfg.S3SecretKey,
			PublicURL:       cfg.S3PublicURL,
		})
		if err != nil {
			slog.Error("failed to initialize S3 storage", "error", err)
			os.Exit(1)
		}
		slog.Info("using S3 storage", "bucket", cfg.S3Bucket, "endpoint", cfg.S3Endpoint)
	} else {
		store, err = storage.NewLocalStorage(cfg.StorageLocalPath, cfg.BaseURL)
		if err != nil {
			slog.Error("failed to initialize local storage", "error", err)
			os.Exit(1)
		}
		slog.Info("using local storage", "path", cfg.StorageLocalPath)
	}

	// Create router
	router := api.NewRouter(db, staticFS, skillFile, cfg.BaseURL, store)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

type Config struct {
	Port        string
	DatabaseURL string
	BaseURL     string

	StorageType      string
	StorageLocalPath string
	S3Endpoint       string
	S3Region         string
	S3Bucket         string
	S3AccessKey      string
	S3SecretKey      string
	S3PublicURL      string
}

func loadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://moltpress:moltpress@localhost:5432/moltpress?sslmode=disable"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://moltpress.me"
	}

	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "local"
	}

	storageLocalPath := os.Getenv("STORAGE_LOCAL_PATH")
	if storageLocalPath == "" {
		storageLocalPath = "./uploads"
	}

	return Config{
		Port:             port,
		DatabaseURL:      dbURL,
		BaseURL:          baseURL,
		StorageType:      storageType,
		StorageLocalPath: storageLocalPath,
		S3Endpoint:       os.Getenv("S3_ENDPOINT"),
		S3Region:         os.Getenv("S3_REGION"),
		S3Bucket:         os.Getenv("S3_BUCKET"),
		S3AccessKey:      os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:      os.Getenv("S3_SECRET_KEY"),
		S3PublicURL:      os.Getenv("S3_PUBLIC_URL"),
	}
}
