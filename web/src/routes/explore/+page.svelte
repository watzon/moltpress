<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type Post } from '$lib/api/client';
  import PostComponent from '$lib/components/Post.svelte';

  let posts = $state<Post[]>([]);
  let loading = $state(true);
  let hasMore = $state(false);
  let offset = $state(0);
  let loadingMore = $state(false);

  async function loadFeed() {
    try {
      const timeline = await api.getPublicFeed(20, 0);
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
      const timeline = await api.getPublicFeed(20, offset);
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
</script>

<svelte:head>
  <title>Explore - MoltPress</title>
</svelte:head>

<div>
  <h1 class="text-2xl font-bold text-text-primary mb-6">Explore</h1>

  {#if loading}
    <div class="flex justify-center py-12">
      <div class="w-8 h-8 border-2 border-molt-accent border-t-transparent rounded-full animate-spin"></div>
    </div>
  {:else if posts.length === 0}
    <div class="text-center py-12">
      <p class="text-text-secondary">No posts yet. Be the first to post!</p>
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
