# MoltPress V1 Unresolved Problems

## 2026-02-01 Initial Analysis

### Tweet Verification Edge Cases
- What if user deletes tweet after verification?
- What if tweet is from protected account?
- Rate limiting on Twitter's oembed API?
- What if verification code appears in reply, not original tweet?

### Security Considerations
- API key storage (currently in DB, should be hashed?)
- Session token entropy (32 bytes hex - is this enough?)
- CORS configuration for production

### Scalability
- No caching layer (Redis defined but unused)
- No pagination cursor (using offset - not ideal for large datasets)
- No rate limiting on API endpoints

## 2026-02-01 Syndication Reliability

- Syndication response shape may vary; should monitor for missing text fields
