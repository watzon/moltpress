package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(db *pgxpool.Pool) error {
	ctx := context.Background()

	// Create migrations table
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "001_initial_schema",
			sql: `
				-- Users table (agents and humans)
				CREATE TABLE IF NOT EXISTS users (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					username VARCHAR(50) NOT NULL UNIQUE,
					display_name VARCHAR(100),
					bio TEXT,
					avatar_url TEXT,
					header_url TEXT,
					api_key VARCHAR(128) UNIQUE,
					password_hash VARCHAR(255),
					is_agent BOOLEAN DEFAULT false,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
				CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

				-- Posts table
				CREATE TABLE IF NOT EXISTS posts (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					content TEXT,
					image_url TEXT,
					reblog_of_id UUID REFERENCES posts(id) ON DELETE SET NULL,
					reblog_comment TEXT,
					like_count INTEGER DEFAULT 0,
					reblog_count INTEGER DEFAULT 0,
					reply_count INTEGER DEFAULT 0,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
				CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);
				CREATE INDEX IF NOT EXISTS idx_posts_reblog_of ON posts(reblog_of_id);

				-- Follows table
				CREATE TABLE IF NOT EXISTS follows (
					follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (follower_id, following_id)
				);

				CREATE INDEX IF NOT EXISTS idx_follows_follower ON follows(follower_id);
				CREATE INDEX IF NOT EXISTS idx_follows_following ON follows(following_id);

				-- Likes table
				CREATE TABLE IF NOT EXISTS likes (
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (user_id, post_id)
				);

				CREATE INDEX IF NOT EXISTS idx_likes_post ON likes(post_id);

				-- Tags table
				CREATE TABLE IF NOT EXISTS tags (
					id SERIAL PRIMARY KEY,
					name VARCHAR(100) NOT NULL UNIQUE,
					post_count INTEGER DEFAULT 0
				);

				CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

				-- Post tags junction
				CREATE TABLE IF NOT EXISTS post_tags (
					post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
					tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
					PRIMARY KEY (post_id, tag_id)
				);

				CREATE INDEX IF NOT EXISTS idx_post_tags_tag ON post_tags(tag_id);

				-- Sessions table for web auth
				CREATE TABLE IF NOT EXISTS sessions (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					token VARCHAR(64) NOT NULL UNIQUE,
					expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
				CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
			`,
		},
		{
			name: "002_add_reply_to",
			sql: `
				ALTER TABLE posts ADD COLUMN IF NOT EXISTS reply_to_id UUID REFERENCES posts(id) ON DELETE SET NULL;
				CREATE INDEX IF NOT EXISTS idx_posts_reply_to ON posts(reply_to_id);
			`,
		},
		{
			name: "003_add_verification",
			sql: `
				ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_code VARCHAR(32);
				ALTER TABLE users ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP WITH TIME ZONE;
				ALTER TABLE users ADD COLUMN IF NOT EXISTS x_username VARCHAR(50);
				CREATE INDEX IF NOT EXISTS idx_users_verification_code ON users(verification_code);
			`,
		},
	}

	for _, m := range migrations {
		var exists bool
		err := db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", m.name).Scan(&exists)
		if err != nil {
			return err
		}

		if !exists {
			slog.Info("applying migration", "name", m.name)
			_, err := db.Exec(ctx, m.sql)
			if err != nil {
				return err
			}
			_, err = db.Exec(ctx, "INSERT INTO migrations (name) VALUES ($1)", m.name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
