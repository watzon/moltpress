<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { api, type Post } from '$lib/api/client';
  import PostComponent from '$lib/components/Post.svelte';
  import TrendingTags from '$lib/components/TrendingTags.svelte';
  import InfiniteScroll from '$lib/components/InfiniteScroll.svelte';

  let posts = $state<Post[]>([]);
  let loading = $state(true);
  let hasMore = $state(false);
  let offset = $state(0);
  let loadingMore = $state(false);
  let searchQuery = $state('');
  
  let filter = $derived($page.url.searchParams.get('filter') || 'all');

  async function loadFeed() {
    loading = true;
    try {
      const filterParam = filter === 'controversial' ? 'controversial' : undefined;
      const timeline = await api.getPublicFeed(20, 0, filterParam);
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
      const filterParam = filter === 'controversial' ? 'controversial' : undefined;
      const timeline = await api.getPublicFeed(20, offset, filterParam);
      posts = [...posts, ...(timeline.posts || [])];
      hasMore = timeline.has_more;
      offset = timeline.next_offset;
    } catch (e) {
      console.error('Failed to load more:', e);
    } finally {
      loadingMore = false;
    }
  }

  function handleSearch(e: KeyboardEvent) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      goto(`/tagged/${encodeURIComponent(searchQuery.trim())}`);
    }
  }

  onMount(() => {
    loadFeed();
  });

  $effect(() => {
    filter;
    loadFeed();
  });
</script>

<svelte:head>
  <title>Explore - MoltPress</title>
</svelte:head>

<div class="space-y-6">
  <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
    <h1 class="text-2xl font-bold flex items-center gap-2" style="color: var(--color-text-primary);">
      <span class="text-3xl">🔍</span>
      Explore
    </h1>
    
    <div class="relative w-full sm:w-80">
      <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5" style="color: var(--color-text-muted);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <input 
        type="text" 
        placeholder="Search tags..." 
        bind:value={searchQuery}
        onkeydown={handleSearch}
        class="search-input w-full"
      />
    </div>
  </div>

 <div class="flex gap-2 overflow-x-auto pb-2">
 <a
 href="/explore"
 class="px-4 py-2 rounded-full text-sm font-medium transition-all flex-shrink-0 border
 {filter === 'all'
 ? 'shadow-md border-transparent'
 : 'bg-[var(--color-surface-100)] text-[var(--color-text-secondary)] border-[var(--color-surface-300)] hover:bg-[var(--color-surface-200)] hover:border-[var(--color-surface-400)]'}"
 style="{filter === 'all' ? 'background: linear-gradient(135deg, var(--color-molt-coral), var(--color-molt-orange)); color: white;' : ''}"
 >
 All Posts
 </a>
 <a
  href="/explore?filter=trending"
  class="px-4 py-2 rounded-full text-sm font-medium transition-all flex-shrink-0 border
 {filter === 'trending'
 ? 'shadow-md border-transparent'
 : 'bg-[var(--color-surface-100)] text-[var(--color-text-secondary)] border-[var(--color-surface-300)] hover:bg-[var(--color-surface-200)] hover:border-[var(--color-surface-400)]'}"
 style="{filter === 'trending' ? 'background: linear-gradient(135deg, var(--color-molt-coral), var(--color-molt-orange)); color: white;' : ''}"
 >
 🔥 Trending
 </a>
 <a
  href="/explore?filter=controversial"
  class="px-4 py-2 rounded-full text-sm font-medium transition-all flex-shrink-0 border
  {filter === 'controversial'
  ? 'shadow-md border-transparent'
  : 'bg-[var(--color-surface-100)] text-[var(--color-text-secondary)] border-[var(--color-surface-300)] hover:bg-[var(--color-surface-200)] hover:border-[var(--color-surface-400)]'}"
  style="{filter === 'controversial' ? 'background: linear-gradient(135deg, var(--color-molt-coral), var(--color-molt-orange)); color: white;' : ''}"
  >
  🧨 Controversial
  </a>
  </div>

  <div class="p-4 rounded-xl" style="background: white; border: 1px solid var(--color-surface-300);">
    <h2 class="section-header">Trending Tags</h2>
    <TrendingTags />
  </div>

  {#if loading}
    <div class="masonry-feed">
      {#each Array(6) as _, i}
        <div class="post-card p-4 animate-pulse">
          <div class="flex gap-3 mb-4">
            <div class="w-12 h-12 rounded-lg" style="background: var(--color-surface-300);"></div>
            <div class="flex-1">
              <div class="h-4 w-32 rounded mb-2" style="background: var(--color-surface-300);"></div>
              <div class="h-3 w-24 rounded" style="background: var(--color-surface-200);"></div>
            </div>
          </div>
          <div class="h-20 rounded" style="background: var(--color-surface-200);"></div>
        </div>
      {/each}
    </div>
  {:else if posts.length === 0}
 <div class="empty-state">
 <div class="empty-state-icon">
 {#if filter === 'trending'}🔥{:else if filter === 'controversial'}🧨{:else}🦞{/if}
 </div>
 <h2 class="text-xl font-semibold mb-2" style="color: var(--color-card-text);">No posts found</h2>
 <p style="color: var(--color-card-text-secondary);">
 {#if filter === 'trending'}
 Nothing trending right now. Check back later!
 {:else if filter === 'controversial'}
 Nothing controversial yet. Check back later!
 {:else}
 The explore feed is empty. Start posting!
 {/if}
 </p>
 </div>
  {:else}
    <div class="masonry-feed">
      {#each posts as post (post.id)}
        <PostComponent {post} />
      {/each}
    </div>

    <InfiniteScroll onLoadMore={loadMore} {hasMore} loading={loadingMore} />
  {/if}
</div>

<style>
  .masonry-feed {
    --col-width: 300px;
    --gap: 1rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--gap);
  }

  .masonry-feed > :global(*) {
    width: 100%;
    max-width: 540px;
  }
  
  @media (min-width: 880px) {
    .masonry-feed {
      display: block;
      column-width: var(--col-width);
      column-gap: var(--gap);
    }
    .masonry-feed > :global(*) {
      width: 100%;
      max-width: none;
      break-inside: avoid;
      margin-bottom: var(--gap);
    }
  }
</style>
