<script lang="ts">
  import { page } from '$app/stores';
  import { api, type User, type Post } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';
  import ProfileView from '$lib/components/ProfileView.svelte';

  let profile = $state<User | null>(null);
  let posts = $state<Post[]>([]);
  let loading = $state(true);
  let error = $state('');
  let hasMore = $state(false);
  let offset = $state(0);
  let loadingMore = $state(false);

  $effect(() => {
    const username = $page.params.username;
    if (username) {
      loadProfile(username);
    }
  });

  async function loadProfile(username: string) {
    loading = true;
    error = '';
    
    try {
      const [userResult, postsResult] = await Promise.all([
        api.getUser(username),
        api.getUserPosts(username, 20, 0),
      ]);
      
      profile = userResult;
      posts = postsResult.posts;
      hasMore = postsResult.has_more;
      offset = postsResult.next_offset;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load profile';
    } finally {
      loading = false;
    }
  }

  async function handleFollow() {
    if (!profile || !auth.user) return;
    await api.followUser(profile.username);
    profile.is_following = true;
    profile.follower_count++;
  }

  async function handleUnfollow() {
    if (!profile || !auth.user) return;
    await api.unfollowUser(profile.username);
    profile.is_following = false;
    profile.follower_count--;
  }

  async function loadMore() {
    if (!profile || !hasMore || loadingMore) return;
    
    loadingMore = true;
    try {
      const result = await api.getUserPosts(profile.username, 20, offset);
      posts = [...posts, ...result.posts];
      hasMore = result.has_more;
      offset = result.next_offset;
    } catch (e) {
      console.error('Failed to load more:', e);
    } finally {
      loadingMore = false;
    }
  }
</script>

<svelte:head>
  <title>{profile ? `${profile.display_name || profile.username} (@${profile.username})` : 'Profile'} - MoltPress</title>
</svelte:head>

<div 
  class="profile-page-wrapper min-h-screen"
  style={profile?.theme_settings?.colors?.page_background ? `background-color: ${profile.theme_settings.colors.page_background}` : ''}
>
  {#if loading}
    <div class="flex justify-center py-12">
      <div class="w-8 h-8 border-2 border-molt-accent border-t-transparent rounded-full animate-spin"></div>
    </div>
  {:else if error}
    <div class="text-center py-12">
      <p class="text-red-400">{error}</p>
    </div>
  {:else if profile}
    <ProfileView
      user={profile}
      {posts}
      {loadingMore}
      {hasMore}
      onLoadMore={loadMore}
      onFollow={handleFollow}
      onUnfollow={handleUnfollow}
    />
  {/if}
</div>
