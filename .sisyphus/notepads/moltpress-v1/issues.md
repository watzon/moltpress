# MoltPress V1 Issues

## 2026-02-01 Initial Analysis

### Critical Issues
1. **X/Twitter verification is stubbed** (handlers.go:149)
   - Just accepts any X username without checking
   - Need to implement tweet URL verification

2. **No login UI** 
   - Missing `/login` route entirely
   - Humans can't log in via web interface

3. **No backend tests**
   - Zero test coverage
   - Need testing infrastructure

### Medium Issues
1. **Hardcoded base URL** in register/+page.svelte:4
   - `https://moltpress.nova.dev` hardcoded
   - Should use environment variable

2. **Trending endpoints missing**
   - TrendingTags.svelte uses placeholder data
   - TrendingAgents.svelte hacks data from posts

3. **Profile page missing verification badge**
   - Post component shows badge
   - Profile header doesn't

### Low Issues
1. **No logout endpoint**
   - Session deletion not implemented
   - Cookie not cleared

2. **No password reset flow**
   - No UI or API for password reset

## 2026-02-01 Build Verification

1. **go build ./... fails without embedded static assets**
   - cmd/server expects static files for embed
   - Frontend build artifacts need to be present
