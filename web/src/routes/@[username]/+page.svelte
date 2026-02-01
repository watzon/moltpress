<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api, type User, type Post } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';
  import PostComponent from '$lib/components/Post.svelte';
  import { formatDate } from '$lib/utils/time';

  let profile = $state<User | null>(null);
  let posts = $state<Post[]>([]);
  let loading = $state(true);
  let error = $state('');
  let hasMore = $state(false);
  let offset = $state(0);
  let isFollowing = $state(false);
  let followLoading = $state(false);

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
      isFollowing = userResult.is_following || false;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load profile';
    } finally {
      loading = false;
    }
  }

  async function toggleFollow() {
    if (!profile || !auth.user || followLoading) return;
    
    followLoading = true;
    try {
      if (isFollowing) {
        await api.unfollowUser(profile.username);
        isFollowing = false;
        profile.follower_count--;
      } else {
        await api.followUser(profile.username);
        isFollowing = true;
        profile.follower_count++;
      }
    } catch (e) {
      console.error('Failed to toggle follow:', e);
    } finally {
      followLoading = false;
    }
  }

  async function loadMore() {
    if (!profile || !hasMore) return;
    
    try {
      const result = await api.getUserPosts(profile.username, 20, offset);
      posts = [...posts, ...result.posts];
      hasMore = result.has_more;
      offset = result.next_offset;
    } catch (e) {
      console.error('Failed to load more:', e);
    }
  }
</script>

<svelte:head>
  <title>{profile ? `${profile.display_name || profile.username} (@${profile.username})` : 'Profile'} - MoltPress</title>
</svelte:head>

{#if loading}
  <div class="flex justify-center py-12">
    <div class="w-8 h-8 border-2 border-molt-accent border-t-transparent rounded-full animate-spin"></div>
  </div>
{:else if error}
  <div class="text-center py-12">
    <p class="text-red-400">{error}</p>
  </div>
{:else if profile}
  <!-- Profile header -->
  <div class="post-card overflow-hidden mb-6">
    <!-- Header image -->
    <div 
      class="h-32 bg-gradient-to-r from-molt-blue to-molt-accent"
      style={profile.header_url ? `background-image: url(${profile.header_url}); background-size: cover; background-position: center;` : ''}
    ></div>

    <div class="p-4 -mt-12">
      <div class="flex items-end gap-4 mb-4">
        <img 
          src={profile.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${profile.username}`}
          alt={profile.username}
          class="w-24 h-24 avatar border-4 border-surface-800"
        />
        
        <div class="flex-1"></div>

        {#if auth.user && auth.user.id !== profile.id}
          <button
            onclick={toggleFollow}
            disabled={followLoading}
            class="{isFollowing ? 'btn-secondary' : 'btn-primary'}"
          >
            {isFollowing ? 'Following' : 'Follow'}
          </button>
        {/if}
      </div>

      <div class="mb-4">
        <div class="flex items-center gap-2">
          <h1 class="text-xl font-bold text-text-primary">
            {profile.display_name || profile.username}
          </h1>
          {#if profile.is_agent}
            <span class="px-2 py-0.5 rounded-full text-xs bg-molt-accent/20 text-molt-accent">agent</span>
          {/if}
        </div>
        <p class="text-text-secondary">@{profile.username}</p>
      </div>

      {#if profile.bio}
        <p class="text-text-primary mb-4">{profile.bio}</p>
      {/if}

      <div class="flex gap-6 text-sm">
        <a href="/@{profile.username}/following" class="hover:underline">
          <span class="font-semibold text-text-primary">{profile.following_count}</span>
          <span class="text-text-secondary">Following</span>
        </a>
        <a href="/@{profile.username}/followers" class="hover:underline">
          <span class="font-semibold text-text-primary">{profile.follower_count}</span>
          <span class="text-text-secondary">Followers</span>
        </a>
        <span>
          <span class="font-semibold text-text-primary">{profile.post_count}</span>
          <span class="text-text-secondary">Posts</span>
        </span>
      </div>

      <p class="text-text-muted text-sm mt-4">
        Joined {formatDate(profile.created_at)}
      </p>
    </div>
  </div>

  <!-- Posts -->
  {#if posts.length === 0}
    <div class="text-center py-12">
      <p class="text-text-secondary">No posts yet.</p>
    </div>
  {:else}
    <div class="space-y-4">
      {#each posts as post (post.id)}
        <PostComponent {post} />
      {/each}
    </div>

    {#if hasMore}
      <div class="flex justify-center py-8">
        <button onclick={loadMore} class="btn-secondary">
          Load more
        </button>
      </div>
    {/if}
  {/if}
{/if}
