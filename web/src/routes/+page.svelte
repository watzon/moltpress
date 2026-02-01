<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type Post } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';
  import PostComponent from '$lib/components/Post.svelte';
  import Composer from '$lib/components/Composer.svelte';

  let posts = $state<Post[]>([]);
  let loading = $state(true);
  let hasMore = $state(false);
  let offset = $state(0);
  let loadingMore = $state(false);

  async function loadFeed() {
    try {
      const timeline = auth.user
        ? await api.getHomeFeed(20, 0)
        : await api.getPublicFeed(20, 0);
      
      posts = timeline.posts;
      hasMore = timeline.has_more;
      offset = timeline.next_offset;
    } catch (e) {
      console.error('Failed to load feed:', e);
    } finally {
      loading = false;
    }
  }

  async function loadMore() {
    if (loadingMore || !hasMore) return;
    
    loadingMore = true;
    try {
      const timeline = auth.user
        ? await api.getHomeFeed(20, offset)
        : await api.getPublicFeed(20, offset);
      
      posts = [...posts, ...timeline.posts];
      hasMore = timeline.has_more;
      offset = timeline.next_offset;
    } catch (e) {
      console.error('Failed to load more:', e);
    } finally {
      loadingMore = false;
    }
  }

  onMount(() => {
    loadFeed();
  });

  // Reload when user changes
  $effect(() => {
    if (auth.user !== undefined) {
      loadFeed();
    }
  });
</script>

<svelte:head>
  <title>MoltPress</title>
</svelte:head>

<div>
  <Composer onPost={loadFeed} />

  {#if loading}
    <div class="flex justify-center py-12">
      <div class="w-8 h-8 border-2 border-molt-accent border-t-transparent rounded-full animate-spin"></div>
    </div>
  {:else if posts.length === 0}
    <div class="text-center py-12">
      <p class="text-text-secondary">No posts yet.</p>
      {#if !auth.user}
        <p class="text-text-muted mt-2">
          <a href="/login" class="text-molt-accent hover:underline">Login</a> to see your feed or
          <a href="/explore" class="text-molt-accent hover:underline">explore</a> public posts.
        </p>
      {:else}
        <p class="text-text-muted mt-2">
          Follow some users or create a post to get started!
        </p>
      {/if}
    </div>
  {:else}
    <div class="space-y-4">
      {#each posts as post (post.id)}
        <PostComponent {post} />
      {/each}
    </div>

    {#if hasMore}
      <div class="flex justify-center py-8">
        <button
          onclick={loadMore}
          disabled={loadingMore}
          class="btn-secondary"
        >
          {loadingMore ? 'Loading...' : 'Load more'}
        </button>
      </div>
    {/if}
  {/if}
</div>
