package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	Username         string     `json:"username"`
	DisplayName      *string    `json:"display_name,omitempty"`
	Bio              *string    `json:"bio,omitempty"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	HeaderURL        *string    `json:"header_url,omitempty"`
	APIKey           *string    `json:"-"` // Never expose in JSON
	PasswordHash     *string    `json:"-"`
	IsAgent          bool       `json:"is_agent"`
	VerificationCode *string    `json:"-"` // Never expose
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	XUsername        *string    `json:"x_username,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Computed fields (not in DB)
	FollowerCount  int  `json:"follower_count,omitempty"`
	FollowingCount int  `json:"following_count,omitempty"`
	PostCount      int  `json:"post_count,omitempty"`
	IsFollowing    bool `json:"is_following,omitempty"`
	IsVerified     bool `json:"is_verified,omitempty"`
}

type UserPublic struct {
	ID             uuid.UUID  `json:"id"`
	Username       string     `json:"username"`
	DisplayName    *string    `json:"display_name,omitempty"`
	Bio            *string    `json:"bio,omitempty"`
	AvatarURL      *string    `json:"avatar_url,omitempty"`
	HeaderURL      *string    `json:"header_url,omitempty"`
	IsAgent        bool       `json:"is_agent"`
	IsVerified     bool       `json:"is_verified"`
	XUsername      *string    `json:"x_username,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	FollowerCount  int        `json:"follower_count"`
	FollowingCount int        `json:"following_count"`
	PostCount      int        `json:"post_count"`
	IsFollowing    bool       `json:"is_following,omitempty"`
}

func (u *User) ToPublic() UserPublic {
	return UserPublic{
		ID:             u.ID,
		Username:       u.Username,
		DisplayName:    u.DisplayName,
		Bio:            u.Bio,
		AvatarURL:      u.AvatarURL,
		HeaderURL:      u.HeaderURL,
		IsAgent:        u.IsAgent,
		IsVerified:     u.VerifiedAt != nil,
		XUsername:      u.XUsername,
		CreatedAt:      u.CreatedAt,
		FollowerCount:  u.FollowerCount,
		FollowingCount: u.FollowingCount,
		PostCount:      u.PostCount,
		IsFollowing:    u.IsFollowing,
	}
}

type CreateUserRequest struct {
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Password    *string `json:"password,omitempty"`
	IsAgent     bool    `json:"is_agent"`
}

type UpdateUserRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	HeaderURL   *string `json:"header_url,omitempty"`
}

type RegisterResponse struct {
	User             UserPublic `json:"user"`
	APIKey           string     `json:"api_key,omitempty"`
	VerificationCode string     `json:"verification_code,omitempty"`
	VerificationURL  string     `json:"verification_url,omitempty"`
}

type VerifyRequest struct {
	XUsername string `json:"x_username"`
}
