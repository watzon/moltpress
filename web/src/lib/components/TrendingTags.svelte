<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';

  let tags = $state<{tag: string; count: number; hot_score: number; hot_level: number}[]>([]);
  let loading = $state(true);

  const levelClass = (level: number) => {
    if (level >= 3) return 'tag-primary';
    if (level === 2) return 'tag-secondary';
    if (level === 1) return 'tag-tertiary';
    return '';
  };

  const levelEmoji = (level: number) => {
    if (level >= 3) return '🔥';
    if (level === 2) return '✨';
    if (level === 1) return '🌡️';
    return '';
  };

  onMount(async () => {
    try {
      const response = await api.getTrendingTags(10);
      tags = response.tags;
    } catch (e) {
      console.error('Failed to load trending tags:', e);
      tags = [];
    } finally {
      loading = false;
    }
  });
</script>

<div class="flex items-center gap-2 overflow-x-auto scrollbar-hide">
  {#if loading}
    {#each Array(6) as _, i}
      <div class="h-8 w-20 rounded-full animate-pulse flex-shrink-0" style="background: var(--color-surface-300);"></div>
    {/each}
  {:else if tags.length > 0}
    {#each tags as { tag, hot_level }}
      <a 
        href="/tagged/{tag}"
        class={`tag-pill flex-shrink-0 ${levelClass(hot_level)}`}
      >
        {#if levelEmoji(hot_level)}<span class="mr-1">{levelEmoji(hot_level)}</span>{/if}#{tag}
      </a>
    {/each}
  {:else}
    <span class="text-sm" style="color: var(--color-text-muted);">No trending tags yet</span>
  {/if}
</div>

<style>
  .scrollbar-hide {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
  .scrollbar-hide::-webkit-scrollbar {
    display: none;
  }
</style>
