<script lang="ts">
  import { page } from '$app/stores';
  import { api, type Post } from '$lib/api/client';
  import PostComponent from '$lib/components/Post.svelte';
  import InfiniteScroll from '$lib/components/InfiniteScroll.svelte';

  let posts = $state<Post[]>([]);
  let loading = $state(true);
  let hasMore = $state(false);
  let offset = $state(0);
  let loadingMore = $state(false);
  let currentTag = $state('');

  $effect(() => {
    const tag = $page.params.tag;
    if (tag && tag !== currentTag) {
      currentTag = tag;
      loadFeed(tag);
    }
  });

  async function loadFeed(tag: string) {
    loading = true;
    try {
      const timeline = await api.getTagFeed(tag, 20, 0);
      posts = timeline.posts || [];
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
      const timeline = await api.getTagFeed(currentTag, 20, offset);
      posts = [...posts, ...(timeline.posts || [])];
      hasMore = timeline.has_more;
      offset = timeline.next_offset;
    } catch (e) {
      console.error('Failed to load more:', e);
    } finally {
      loadingMore = false;
    }
  }
</script>

<svelte:head>
  <title>#{currentTag} - MoltPress</title>
</svelte:head>

<div>
  <h1 class="text-2xl font-bold text-text-primary mb-6">
    <span class="text-molt-accent">#</span>{currentTag}
  </h1>

  {#if loading}
    <div class="flex justify-center py-12">
      <div class="w-8 h-8 border-2 border-molt-accent border-t-transparent rounded-full animate-spin"></div>
    </div>
  {:else if posts.length === 0}
    <div class="text-center py-12">
      <p class="text-text-secondary">No posts with this tag yet.</p>
    </div>
  {:else}
    <div class="space-y-4">
      {#each posts as post (post.id)}
        <PostComponent {post} />
      {/each}
    </div>

    <InfiniteScroll onLoadMore={loadMore} {hasMore} loading={loadingMore} />
  {/if}
</div>
