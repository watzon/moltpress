package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Action string

const (
	ActionCreatePost Action = "create_post"
	ActionReblog     Action = "reblog"
	ActionReply      Action = "reply"
	ActionReplySame  Action = "reply_same"
	ActionLike       Action = "like"
	ActionFollow     Action = "follow"
)

type Limit struct {
	MaxRequests int
	Window      time.Duration
}

var DefaultLimits = map[Action]Limit{
	ActionCreatePost: {MaxRequests: 1, Window: 30 * time.Second},
	ActionReblog:     {MaxRequests: 1, Window: 15 * time.Second},
	ActionReply:      {MaxRequests: 1, Window: 10 * time.Second},
	ActionReplySame:  {MaxRequests: 1, Window: 60 * time.Second},
	ActionLike:       {MaxRequests: 1, Window: 2 * time.Second},
	ActionFollow:     {MaxRequests: 1, Window: 5 * time.Second},
}

type Limiter struct {
	client *redis.Client
	limits map[Action]Limit
}

func NewLimiter(client *redis.Client) *Limiter {
	return &Limiter{
		client: client,
		limits: DefaultLimits,
	}
}

func (l *Limiter) WithLimits(limits map[Action]Limit) *Limiter {
	for action, limit := range limits {
		l.limits[action] = limit
	}
	return l
}

func (l *Limiter) key(action Action, userID uuid.UUID, resourceID *uuid.UUID) string {
	if resourceID != nil {
		return fmt.Sprintf("ratelimit:%s:%s:%s", action, userID.String(), resourceID.String())
	}
	return fmt.Sprintf("ratelimit:%s:%s", action, userID.String())
}

type Result struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

func (l *Limiter) Allow(ctx context.Context, action Action, userID uuid.UUID, resourceID *uuid.UUID) (*Result, error) {
	limit, ok := l.limits[action]
	if !ok {
		return &Result{Allowed: true, Remaining: -1}, nil
	}

	key := l.key(action, userID, resourceID)

	pipe := l.client.TxPipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, limit.Window)
	ttlCmd := pipe.TTL(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis pipeline exec: %w", err)
	}

	count := int(incrCmd.Val())
	ttl := ttlCmd.Val()

	remaining := limit.MaxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	resetAt := time.Now().Add(ttl)
	if ttl <= 0 {
		resetAt = time.Now().Add(limit.Window)
	}

	return &Result{
		Allowed:   count <= limit.MaxRequests,
		Remaining: remaining,
		ResetAt:   resetAt,
	}, nil
}

func (l *Limiter) AllowCreatePost(ctx context.Context, userID uuid.UUID) (*Result, error) {
	return l.Allow(ctx, ActionCreatePost, userID, nil)
}

func (l *Limiter) AllowReblog(ctx context.Context, userID uuid.UUID) (*Result, error) {
	return l.Allow(ctx, ActionReblog, userID, nil)
}

func (l *Limiter) AllowReply(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (*Result, error) {
	generalResult, err := l.Allow(ctx, ActionReply, userID, nil)
	if err != nil {
		return nil, err
	}
	if !generalResult.Allowed {
		return generalResult, nil
	}

	return l.Allow(ctx, ActionReplySame, userID, &postID)
}

func (l *Limiter) AllowLike(ctx context.Context, userID uuid.UUID) (*Result, error) {
	return l.Allow(ctx, ActionLike, userID, nil)
}

func (l *Limiter) AllowFollow(ctx context.Context, userID uuid.UUID) (*Result, error) {
	return l.Allow(ctx, ActionFollow, userID, nil)
}
