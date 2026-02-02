# Profile Redesign - Tumblr-Style Modal + Enhanced Design

**Created:** 2026-02-01
**Status:** In Progress

## Overview

Implement Tumblr-inspired profile experience:
- Profile modals that appear over the current page when clicking usernames
- URL changes to profile URL, but current page stays in background
- Full refresh shows actual profile page (graceful degradation)
- Enhanced profile design with banner, centered avatar, better UX
- No right sidebar on profile pages

## Architecture

### Component Structure

```
ProfileContent.svelte     # Shared profile UI (banner, avatar, bio, posts)
ProfileModal.svelte       # Modal wrapper with overlay, close button
ProfileContext.svelte.ts  # State for modal management
```

### Routing Strategy

1. Click on `@username` link → intercept, open modal, `pushState` to `/@username`
2. In modal → scroll profile content, posts load with infinite scroll
3. Close modal → `popstate` back to previous URL
4. Direct navigation / refresh → render full profile page

### Key Technical Details

- Use SvelteKit's `$app/navigation` for `pushState`/`replaceState`
- Modal uses `position: fixed` with backdrop blur
- Profile page detects if opened via modal (check navigation state)
- Layout conditionally hides right sidebar for profile routes

## Tasks

### Phase 1: Core Infrastructure
- [x] Create ProfileContent component (shared UI between modal/page)
- [x] Create ProfileModal component (overlay, close button, back button)
- [x] Create profile modal state management
- [x] Update layout to support modal rendering

### Phase 2: Profile Design Enhancement
- [x] Design enhanced profile header (taller banner, overlapping avatar)
- [x] Add profile links section (website, social links)
- [x] Add posts section with tabs/filtering
- [x] Implement follow/unfollow button improvements

### Phase 3: Modal Behavior
- [x] Implement link interception for @username links
- [x] Add shallow routing (pushState on modal open)
- [x] Handle back button (close modal on popstate)
- [x] Handle Escape key to close modal

### Phase 4: Layout Adaptation  
- [x] Hide right sidebar on profile pages
- [x] Full-width profile content
- [x] Responsive adjustments

### Phase 5: Testing & Polish
- [ ] Test modal/page transitions
- [ ] Test browser back/forward
- [ ] Test direct URL access
- [ ] Verify mobile responsiveness
- [ ] Fix any LSP/build errors

## Dependencies

- Existing: `@[username]/+page.svelte`, User model, API client
- New components must use Svelte 5 runes
- Follow existing styling patterns (app.css CSS vars)

## Notes

- Backend already supports `header_url` for banner images
- Profile page currently has basic design - needs complete overhaul
- Must maintain existing functionality (follow, posts, stats)
