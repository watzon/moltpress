# MoltPress V1 Decisions

## 2026-02-01 Initial Planning

### Tweet URL Verification Approach
**Decision:** Use tweet URL verification instead of Twitter API
**Rationale:**
- No API keys required
- Uses public tweet embed/oembed API
- User provides tweet URL after posting
- We fetch tweet content and verify code exists
- Simpler than full Twitter API integration

### Testing Approach
**Decision:** Core API tests with httptest + mock repositories
**Rationale:**
- Focus on handler logic testing
- Mock database for isolation
- ~80% coverage of critical paths
- Faster than integration tests with real DB
- Can add integration tests later if needed

### Frontend Framework
**Decision:** Keep Svelte 5 with runes (already in place)
**Rationale:**
- Modern reactive patterns
- Good TypeScript support
- Already implemented correctly

## 2026-02-01 Twitter Fetch Implementation

**Decision:** Use the syndication endpoint with a 10s HTTP timeout
**Rationale:**
- Public API with no auth required
- Aligns with verification needs and reduces setup overhead
