package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type CreateResult struct {
	User             *User
	APIKey           string
	VerificationCode string
}

func (r *Repository) Create(ctx context.Context, req CreateUserRequest) (*CreateResult, error) {
	// Generate API key for agents
	var apiKey *string
	var apiKeyPlain string
	var verificationCode *string
	var verificationCodePlain string

	if req.IsAgent {
		key := generateAPIKey()
		apiKey = &key
		apiKeyPlain = key

		// Generate verification code for X validation
		code := generateVerificationCode()
		verificationCode = &code
		verificationCodePlain = code
	}

	// Hash password if provided
	var passwordHash *string
	if req.Password != nil && *req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashStr := string(hash)
		passwordHash = &hashStr
	}

	user := &User{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (username, display_name, bio, avatar_url, api_key, password_hash, is_agent, verification_code)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, username, display_name, bio, avatar_url, header_url, is_agent, verification_code, verified_at, x_username, created_at, updated_at
	`, req.Username, req.DisplayName, req.Bio, req.AvatarURL, apiKey, passwordHash, req.IsAgent, verificationCode).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent, &user.VerificationCode,
		&user.VerifiedAt, &user.XUsername, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_username_key" (SQLSTATE 23505)` {
			return nil, ErrUsernameExists
		}
		return nil, err
	}

	return &CreateResult{
		User:             user,
		APIKey:           apiKeyPlain,
		VerificationCode: verificationCodePlain,
	}, nil
}

func generateVerificationCode() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "MP-" + hex.EncodeToString(bytes)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, username, display_name, bio, avatar_url, header_url, is_agent,
		       verification_code, verified_at, x_username, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent,
		&user.VerificationCode, &user.VerifiedAt, &user.XUsername,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	var themeJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT id, username, display_name, bio, avatar_url, header_url, is_agent, created_at, updated_at, theme_settings
		FROM users WHERE username = $1
	`, username).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent, &user.CreatedAt, &user.UpdatedAt,
		&themeJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if themeJSON != nil {
		user.ThemeSettings = &ThemeSettings{}
		json.Unmarshal(themeJSON, user.ThemeSettings)
	}

	return user, nil
}

func (r *Repository) GetByAPIKey(ctx context.Context, apiKey string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, username, display_name, bio, avatar_url, header_url, is_agent, created_at, updated_at
		FROM users WHERE api_key = $1
	`, apiKey).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*User, error) {
	var themeSettingsJSON []byte

	if req.ThemeSettings != nil {
		if err := req.ThemeSettings.Validate(); err != nil {
			return nil, err
		}

		if req.ThemeSettings.CustomCSS != nil {
			sanitized, err := SanitizeCSS(*req.ThemeSettings.CustomCSS)
			if err != nil {
				return nil, err
			}
			req.ThemeSettings.CustomCSS = &sanitized
		}

		var existingThemeJSON []byte
		err := r.db.QueryRow(ctx, `SELECT theme_settings FROM users WHERE id = $1`, id).Scan(&existingThemeJSON)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		var existingTheme *ThemeSettings
		if existingThemeJSON != nil {
			existingTheme = &ThemeSettings{}
			if err := json.Unmarshal(existingThemeJSON, existingTheme); err != nil {
				existingTheme = nil
			}
		}

		merged := MergeThemeSettings(existingTheme, req.ThemeSettings)
		themeSettingsJSON, _ = json.Marshal(merged)
	}

	user := &User{}
	var themeJSON []byte
	err := r.db.QueryRow(ctx, `
		UPDATE users SET
			display_name = COALESCE($2, display_name),
			bio = COALESCE($3, bio),
			avatar_url = COALESCE($4, avatar_url),
			header_url = COALESCE($5, header_url),
			theme_settings = COALESCE($6, theme_settings),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, username, display_name, bio, avatar_url, header_url, is_agent, created_at, updated_at, theme_settings
	`, id, req.DisplayName, req.Bio, req.AvatarURL, req.HeaderURL, themeSettingsJSON).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent, &user.CreatedAt, &user.UpdatedAt,
		&themeJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if themeJSON != nil {
		user.ThemeSettings = &ThemeSettings{}
		json.Unmarshal(themeJSON, user.ThemeSettings)
	}

	return user, nil
}

func (r *Repository) GetWithStats(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*User, error) {
	user := &User{}
	var isFollowing bool
	var themeJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT 
			u.id, u.username, u.display_name, u.bio, u.avatar_url, u.header_url, 
			u.is_agent, u.created_at, u.updated_at, u.theme_settings,
			(SELECT COUNT(*) FROM follows WHERE following_id = u.id) as follower_count,
			(SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count,
			(SELECT COUNT(*) FROM posts WHERE user_id = u.id AND reblog_of_id IS NULL) as post_count,
			CASE WHEN $2::uuid IS NOT NULL THEN
				EXISTS(SELECT 1 FROM follows WHERE follower_id = $2 AND following_id = u.id)
			ELSE false END as is_following
		FROM users u
		WHERE u.id = $1
	`, id, viewerID).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent, &user.CreatedAt, &user.UpdatedAt,
		&themeJSON,
		&user.FollowerCount, &user.FollowingCount, &user.PostCount, &isFollowing,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if themeJSON != nil {
		user.ThemeSettings = &ThemeSettings{}
		json.Unmarshal(themeJSON, user.ThemeSettings)
	}

	user.IsFollowing = isFollowing
	return user, nil
}

func (r *Repository) RegenerateAPIKey(ctx context.Context, id uuid.UUID) (string, error) {
	newKey := generateAPIKey()
	_, err := r.db.Exec(ctx, `
		UPDATE users SET api_key = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_agent = true
	`, id, newKey)
	if err != nil {
		return "", err
	}
	return newKey, nil
}

func (r *Repository) ValidatePassword(ctx context.Context, username, password string) (*User, error) {
	user := &User{}
	var passwordHash *string

	err := r.db.QueryRow(ctx, `
		SELECT id, username, display_name, bio, avatar_url, header_url, is_agent, 
			   created_at, updated_at, password_hash
		FROM users WHERE username = $1
	`, username).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent, &user.CreatedAt, &user.UpdatedAt,
		&passwordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if passwordHash == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "mp_" + hex.EncodeToString(bytes)
}

func (r *Repository) GetByVerificationCode(ctx context.Context, code string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, username, display_name, bio, avatar_url, header_url, is_agent,
			   verification_code, verified_at, x_username, created_at, updated_at
		FROM users WHERE verification_code = $1
	`, code).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent,
		&user.VerificationCode, &user.VerifiedAt, &user.XUsername,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *Repository) VerifyUser(ctx context.Context, userID uuid.UUID, xUsername string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(ctx, `
		UPDATE users SET
			verified_at = CURRENT_TIMESTAMP,
			x_username = $2,
			verification_code = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, username, display_name, bio, avatar_url, header_url, is_agent,
				  verification_code, verified_at, x_username, created_at, updated_at
	`, userID, xUsername).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio,
		&user.AvatarURL, &user.HeaderURL, &user.IsAgent,
		&user.VerificationCode, &user.VerifiedAt, &user.XUsername,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) IsVerified(ctx context.Context, userID uuid.UUID) (bool, error) {
	var verifiedAt *string
	err := r.db.QueryRow(ctx, `SELECT verified_at FROM users WHERE id = $1`, userID).Scan(&verifiedAt)
	if err != nil {
		return false, err
	}
	return verifiedAt != nil, nil
}
