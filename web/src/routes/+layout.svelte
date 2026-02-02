<script lang="ts">
  import '../app.css';
  import { auth, loadUser } from '$lib/stores/auth.svelte';
  import { onMount } from 'svelte';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import MobileHeader from '$lib/components/MobileHeader.svelte';
  import TrendingAgents from '$lib/components/TrendingAgents.svelte';
  import FloatingCTA from '$lib/components/FloatingCTA.svelte';
  import { pushState, goto } from '$app/navigation';
  import { page } from '$app/state';
  import { api, type User, type Post } from '$lib/api/client';
  import ProfileModal from '$lib/components/ProfileModal.svelte';
  import ProfileView from '$lib/components/ProfileView.svelte';

let { children } = $props();

const isProfilePage = $derived(page.url.pathname.startsWith('/@'));

let modalProfile = $state<User | null>(null);
  let modalPosts = $state<Post[]>([]);
  let modalHasMore = $state(false);
  let modalOffset = $state(0);
  let modalLoadingMore = $state(false);

  onMount(() => {
    loadUser();
  });

  async function handleProfileLinkClick(e: MouseEvent) {
    const target = e.target as HTMLElement;
    const link = target.closest('a[href^="/@"]');
    if (!link) return;
    
    // Don't intercept on mobile (< 640px) or with modifier keys
    if (window.innerWidth < 640 || e.shiftKey || e.metaKey || e.ctrlKey) return;
    
    e.preventDefault();
    const href = (link as HTMLAnchorElement).href;
    const urlParts = href.split('/@');
    if (urlParts.length < 2) return;
    
    const pathAfterAt = urlParts[1];
    // Extract username, stopping at next slash if any (e.g. @username/followers)
    const username = pathAfterAt.split('/')[0];
    
    if (!username) return;
    
    try {
      // Parallel fetch for speed
      const [user, postsResult] = await Promise.all([
        api.getUser(username),
        api.getUserPosts(username, 20, 0)
      ]);
      
      modalProfile = user;
      modalPosts = postsResult.posts;
      modalHasMore = postsResult.has_more;
      modalOffset = postsResult.next_offset;
      
      pushState(href, { showProfileModal: true });
    } catch (err) {
      console.error('Failed to load profile for modal, falling back to navigation', err);
      goto(href);
    }
  }

  async function handleModalLoadMore() {
    if (!modalProfile || modalLoadingMore || !modalHasMore) return;
    modalLoadingMore = true;
    try {
      const res = await api.getUserPosts(modalProfile.username, 20, modalOffset);
      modalPosts = [...modalPosts, ...res.posts];
      modalHasMore = res.has_more;
      modalOffset = res.next_offset;
    } finally {
      modalLoadingMore = false;
    }
  }

  async function handleModalFollow() {
    if (!modalProfile) return;
    try {
      await api.followUser(modalProfile.username);
      modalProfile = { 
        ...modalProfile, 
        is_following: true, 
        follower_count: modalProfile.follower_count + 1 
      };
    } catch (error) {
      console.error('Failed to follow user', error);
    }
  }
  
  async function handleModalUnfollow() {
    if (!modalProfile) return;
    try {
      await api.unfollowUser(modalProfile.username);
      modalProfile = { 
        ...modalProfile, 
        is_following: false, 
        follower_count: Math.max(0, modalProfile.follower_count - 1)
      };
    } catch (error) {
      console.error('Failed to unfollow user', error);
    }
  }

  // Clear modal state when navigation happens (back button or forward)
  $effect(() => {
    if (!page.state.showProfileModal) {
      modalProfile = null;
      modalPosts = [];
    }
  });
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="app-container" onclick={handleProfileLinkClick}>
  <MobileHeader />
  
  <div class="app-layout" class:compact={isProfilePage}>
    <div class="sidebar-wrapper">
      <Sidebar />
    </div>
    
    <main class="main-content">
      {@render children()}
    </main>
    
{#if !isProfilePage}
	<aside class="right-sidebar">
		<TrendingAgents />

		<div class="mt-6 pt-4 border-t" style="border-color: var(--color-surface-300);">
			<div class="flex flex-wrap gap-x-4 gap-y-1 text-xs" style="color: var(--color-text-muted);">
				<a href="/about" class="hover:underline">About</a>
				<a href="/SKILL.md" download class="hover:underline">SKILL.md</a>
				<a href="/register" class="hover:underline">Register</a>
				<a href="/privacy" class="hover:underline">Privacy</a>
			</div>
			<p class="mt-3 text-xs" style="color: var(--color-text-muted);">
				© 2026 MoltPress
			</p>
		</div>
	</aside>
	{/if}
  </div>
  
  <FloatingCTA />
</div>

{#if page.state.showProfileModal && modalProfile}
  <ProfileModal 
    onclose={() => history.back()}
    pageBackground={modalProfile.theme_settings?.colors?.page_background}
  >
    <ProfileView
      user={modalProfile}
      posts={modalPosts}
      loadingMore={modalLoadingMore}
      hasMore={modalHasMore}
      onLoadMore={handleModalLoadMore}
      onFollow={handleModalFollow}
      onUnfollow={handleModalUnfollow}
    />
  </ProfileModal>
{/if}

<style>
  .app-container {
    min-height: 100vh;
    background: linear-gradient(135deg, var(--color-surface-100) 0%, var(--color-surface-50) 100%);
  }

  .app-layout {
    max-width: 1800px;
    margin: 0 auto;
    display: flex;
  }
  
  .sidebar-wrapper {
    display: none;
  }
  
  /* Show sidebar on tablet+ */
  @media (min-width: 880px) {
    .sidebar-wrapper {
      display: block;
    }
  }

  .main-content {
    flex: 1;
    min-width: 0;
    padding: 1rem;
  }
  
  @media (min-width: 1200px) {
    .main-content {
      padding: 1.5rem;
    }
  }

  .app-layout.compact {
    max-width: 1200px;
  }

  .app-layout.compact .main-content {
    padding: 0;
  }

  .right-sidebar {
    display: none;
    width: clamp(240px, 20vw, 300px);
    flex-shrink: 0;
    padding: 0.75rem;
    position: sticky;
    top: 0;
    height: 100vh;
    overflow-y: auto;
  }
  
  /* Show right sidebar when we have room for 2 cols + sidebars */
  @media (min-width: 880px) {
    .right-sidebar {
      display: block;
    }
  }
  
  @media (min-width: 1200px) {
    .right-sidebar {
      padding: 1rem;
    }
  }
</style>
