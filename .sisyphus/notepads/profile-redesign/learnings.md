
## Profile Modal & Shallow Routing (Svelte 5)

Implemented a Tumblr-style profile modal using SvelteKit's shallow routing and `pushState`.

### Pattern
1. **Event Delegation**: Added `onclick` to the root layout container to intercept all clicks.
2. **Filtering**: 
   - Check `e.target.closest('a[href^="/@"]')`.
   - Ignore mobile (`< 640px`) to fallback to full page navigation.
   - Ignore modifier keys (Cmd/Ctrl/Shift + Click).
3. **Data Fetching**:
   - `e.preventDefault()`.
   - Fetch `api.getUser` and `api.getUserPosts` in parallel.
   - Fallback to `goto(href)` if fetch fails.
4. **State Management**:
   - Use `pushState(href, { showProfileModal: true })`.
   - Store profile data in local `$state` (runes) within `+layout.svelte`.
   - Watch `page.state.showProfileModal` with `$effect` to clear data when modal closes (e.g. back button).
5. **Modal Component**:
   - Conditionally render `ProfileModal` based on `page.state.showProfileModal && modalProfile`.
   - Pass interactions (`onFollow`, `onUnfollow`, `onLoadMore`) as props to `ProfileContent`.

### Key Benefits
- **URL Sync**: URL updates to `/@username` so links can be shared/bookmarked (though reloading goes to full page, which is desired).
- **History**: Browser back button closes the modal naturally.
- **Performance**: Preloads data before showing modal; falls back gracefully.
- **Experience**: Keeps context of the current feed while exploring profiles.

### Code Example
```typescript
async function handleProfileLinkClick(e: MouseEvent) {
  // ... checks ...
  e.preventDefault();
  const [user, posts] = await fetchProfileData(username);
  modalProfile = user;
  pushState(href, { showProfileModal: true });
}

$effect(() => {
  if (!page.state.showProfileModal) modalProfile = null;
});
```
