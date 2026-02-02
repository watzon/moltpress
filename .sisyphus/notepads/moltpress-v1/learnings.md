# MoltPress V1 Learnings

## 2026-02-01 Initial Analysis

### Codebase Patterns
- Go backend uses repository pattern (no ORM, raw SQL via pgx)
- Handlers call repositories directly (no service layer)
- Svelte 5 with runes ($state, $derived, $props) - NOT stores
- API client is singleton at `$lib/api/client.ts`
- Auth: Bearer tokens for agents, session cookies for humans

### Key Files
- `internal/api/handlers.go` - All HTTP handlers
- `internal/api/router.go` - Route registration
- `internal/api/middleware.go` - Auth middleware (withAuth)
- `web/src/lib/stores/auth.svelte.ts` - Auth state with runes

### Verification Flow (Current - STUBBED)
1. Agent registers → gets `verification_code` (MP-xxxx format)
2. Agent gets `verification_url` (pre-filled tweet intent)
3. Agent calls `/api/v1/verify` with `x_username`
4. Backend just accepts it without checking (line 149 handlers.go)

### What Moltbook Does
- 1.5M+ agents registered
- X/Twitter verification proves human ownership
- Karma system for reputation
- Submolts (like subreddits)
- OpenClaw integration for agent creation

## 2026-02-01 Twitter Syndication Client

- Added twitter client to parse tweet IDs and fetch via syndication API with a 10s timeout

- **Verification Badge**: Implemented using inline SVG from `Post.svelte` with `var(--color-molt-accent)` for consistency. Added conditionally based on `profile.is_verified`.
- **X Linking**: Added conditional link to X profile if user is verified and has `x_username`.
