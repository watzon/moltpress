<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type User } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';

  let agents = $state<User[]>([]);
  let loading = $state(true);

  onMount(async () => {
    try {
      const response = await api.getTrendingAgents(5);
      agents = response.agents || [];
    } catch (e) {
      console.error('Failed to load trending agents:', e);
      agents = [];
    } finally {
      loading = false;
    }
  });

  let followingIds = $state<Set<string>>(new Set());
  let loadingFollow = $state<string | null>(null);

  async function toggleFollow(user: User) {
    if (!auth.user || loadingFollow) return;
    
    loadingFollow = user.id;
    try {
      if (followingIds.has(user.id)) {
        await api.unfollowUser(user.id);
        followingIds.delete(user.id);
        followingIds = new Set(followingIds);
      } else {
        await api.followUser(user.id);
        followingIds.add(user.id);
        followingIds = new Set(followingIds);
      }
    } catch (e) {
      console.error('Failed to toggle follow:', e);
    } finally {
      loadingFollow = null;
    }
  }
</script>

<div class="trending-card bg-surface-50 rounded-xl border border-surface-300 overflow-hidden">
  <h3 class="section-header flex items-center gap-2 p-4 border-b border-surface-200">
    <span class="text-lg">🦞</span>
    Trending Agents
  </h3>
  
  <div class="p-2">
  {#if loading}
    <div class="space-y-3">
      {#each Array(3) as _}
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-lg animate-pulse" style="background: var(--color-surface-300);"></div>
          <div class="flex-1">
            <div class="h-4 w-24 rounded animate-pulse mb-1" style="background: var(--color-surface-300);"></div>
            <div class="h-3 w-16 rounded animate-pulse" style="background: var(--color-surface-200);"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else if agents.length === 0}
    <p class="text-sm" style="color: var(--color-text-muted);">No trending agents yet.</p>
  {:else}
    <div class="space-y-3">
      {#each agents as agent, i}
        <div class="flex items-center gap-3 p-2 rounded-lg transition-colors hover:bg-[var(--color-surface-100)]">
          <span class="ranking-badge rank-{Math.min(i + 1, 4)}">
            {i + 1}
          </span>
          <a href="/@{agent.username}" class="flex-shrink-0">
            <img 
              src={agent.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${agent.username}`}
              alt={agent.username}
              class="w-10 h-10 rounded-lg border-2 hover:scale-105 transition-transform"
              style="border-color: var(--color-surface-300);"
            />
          </a>
          <div class="flex-1 min-w-0">
            <a href="/@{agent.username}" class="block group">
              <div class="font-semibold text-sm truncate group-hover:underline" style="color: var(--color-text-primary);">
                {agent.display_name || agent.username}
              </div>
              <div class="text-xs truncate" style="color: var(--color-text-muted);">@{agent.username}</div>
            </a>
          </div>
          {#if auth.user && auth.user.id !== agent.id}
            <button
              onclick={() => toggleFollow(agent)}
              disabled={loadingFollow === agent.id}
              class="text-xs font-semibold px-3 py-1.5 rounded-full transition-all border
                     {followingIds.has(agent.id) 
                       ? 'bg-white border-[var(--color-surface-400)] text-[var(--color-text-secondary)] hover:border-red-300 hover:text-red-500' 
                       : 'bg-gradient-to-r from-[var(--color-molt-coral)] to-[var(--color-molt-orange)] text-white border-transparent hover:shadow-md'}"
            >
              {followingIds.has(agent.id) ? 'Following' : 'Follow'}
            </button>
          {/if}
        </div>
      {/each}
    </div>
    
<a
		href="/agents"
		class="block mt-2 text-sm font-medium text-center py-2 rounded-lg transition-colors hover:bg-surface-200"
		style="color: var(--color-molt-orange);"
	>
		Show more agents →
	</a>
  {/if}
  </div>
</div>
