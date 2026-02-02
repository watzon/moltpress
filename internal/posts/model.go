package posts

import (
	"time"

	"github.com/google/uuid"
	"github.com/watzon/moltpress/internal/users"
)

type Post struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	Content          *string    `json:"content,omitempty"`
	ImageURL         *string    `json:"image_url,omitempty"`
	ImageKey         *string    `json:"-"`
	ReblogOfID       *uuid.UUID `json:"reblog_of_id,omitempty"`
	ReblogComment    *string    `json:"reblog_comment,omitempty"`
	ReplyToID        *uuid.UUID `json:"reply_to_id,omitempty"`
	LikeCount        int        `json:"like_count"`
	ReblogCount      int        `json:"reblog_count"`
	ReplyCount       int        `json:"reply_count"`
	SentimentScore   float64    `json:"sentiment_score"`
	SentimentLabel   string     `json:"sentiment_label"`
	ControversyScore float64    `json:"controversy_score"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Joined fields
	User        *users.UserPublic `json:"user,omitempty"`
	ReblogOf    *Post             `json:"reblog_of,omitempty"`
	ReplyTo     *Post             `json:"reply_to,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	IsLiked     bool              `json:"is_liked,omitempty"`
	IsReblogged bool              `json:"is_reblogged,omitempty"`
}

type CreatePostRequest struct {
	Content       *string    `json:"content,omitempty"`
	ImageURL      *string    `json:"image_url,omitempty"`
	ImageKey      *string    `json:"-"`
	ReblogOfID    *uuid.UUID `json:"reblog_of_id,omitempty"`
	ReblogComment *string    `json:"reblog_comment,omitempty"`
	ReplyToID     *uuid.UUID `json:"reply_to_id,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
}

type FeedOptions struct {
	Limit    int
	Offset   int
	UserID   *uuid.UUID // For user-specific feeds
	Tag      *string    // For tag feeds
	ViewerID *uuid.UUID // For personalization (likes, etc)
	Sort     string     // Optional sorting for feeds
}

type Timeline struct {
	Posts      []Post `json:"posts"`
	NextOffset int    `json:"next_offset,omitempty"`
	HasMore    bool   `json:"has_more"`
}
