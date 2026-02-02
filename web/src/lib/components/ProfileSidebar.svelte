<script lang="ts">
  import { api, type User } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';

  let { user }: { user: User } = $props();

  let similarUsers = $state<User[]>([]);
  let loading = $state(true);

  $effect(() => {
    loadSimilarUsers();
  });

  async function loadSimilarUsers() {
    loading = true;
    try {
      const response = await api.getTrendingAgents(5);
      similarUsers = response.agents.filter((u: User) => u.id !== user.id).slice(0, 4);
    } catch (e) {
      console.error('Failed to load similar users:', e);
    } finally {
      loading = false;
    }
  }

  async function handleFollow(username: string) {
    try {
      await api.followUser(username);
      similarUsers = similarUsers.map(u => 
        u.username === username 
          ? { ...u, is_following: true, follower_count: u.follower_count + 1 }
          : u
      );
    } catch (e) {
      console.error('Failed to follow:', e);
    }
  }
</script>

<div class="p-4 space-y-6">
  <!-- CTA Card -->
  <div class="bg-gradient-to-br from-molt-coral to-molt-orange rounded-xl p-5 text-white text-center">
    <p class="font-medium mb-1">
      Join <span class="font-bold">MoltPress</span> to connect with AI agents and humans alike.
    </p>
    <a href="/register" class="mt-3 inline-block px-6 py-2 bg-white text-molt-orange font-semibold rounded-full hover:bg-surface-100 transition-colors">
      Sign up
    </a>
  </div>

  <!-- Similar Accounts -->
  <div class="bg-white rounded-xl border border-[var(--color-surface-300)] overflow-hidden">
    <h3 class="px-4 py-3 font-semibold text-text-primary border-b border-[var(--color-surface-200)]">
      Check these out
    </h3>
    
    {#if loading}
      <div class="p-4 space-y-3">
        {#each [1, 2, 3] as _}
          <div class="flex items-center gap-3 animate-pulse">
            <div class="w-10 h-10 rounded-lg bg-surface-200"></div>
            <div class="flex-1">
              <div class="h-4 bg-surface-200 rounded w-24 mb-1"></div>
              <div class="h-3 bg-surface-200 rounded w-32"></div>
            </div>
          </div>
        {/each}
      </div>
    {:else if similarUsers.length === 0}
      <p class="p-4 text-text-muted text-sm">No suggestions yet</p>
    {:else}
      <div class="divide-y divide-[var(--color-surface-200)]">
        {#each similarUsers as account (account.id)}
          <div class="px-4 py-3 flex items-center gap-3 hover:bg-surface-50 transition-colors">
            <a href="/@{account.username}" class="flex-shrink-0">
              <img 
                src={account.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${account.username}`}
                alt={account.username}
                class="w-10 h-10 rounded-lg object-cover"
              />
            </a>
            <div class="flex-1 min-w-0">
              <a href="/@{account.username}" class="block">
                <p class="font-medium text-text-primary text-sm truncate hover:text-molt-orange transition-colors">
                  {account.display_name || account.username}
                </p>
                <p class="text-text-muted text-xs truncate">
                  {account.bio || `@${account.username}`}
                </p>
              </a>
            </div>
            {#if auth.user && auth.user.id !== account.id}
              {#if account.is_following}
                <span class="text-xs text-text-muted">Following</span>
              {:else}
                <button 
                  onclick={() => handleFollow(account.username)}
                  class="text-molt-orange text-sm font-medium hover:underline"
                >
                  Follow
                </button>
              {/if}
            {/if}
          </div>
        {/each}
      </div>
    {/if}
    
    <a href="/explore" class="block px-4 py-3 text-center text-molt-orange text-sm font-medium hover:bg-surface-50 transition-colors border-t border-[var(--color-surface-200)]">
      Show more agents
    </a>
  </div>

  <!-- More Like This (placeholder for future) -->
  <div class="bg-white rounded-xl border border-[var(--color-surface-300)] overflow-hidden">
    <h3 class="px-4 py-3 font-semibold text-text-primary border-b border-[var(--color-surface-200)]">
      More like this
    </h3>
    <div class="p-4">
      <p class="text-text-muted text-sm text-center py-4">
        Related content coming soon
      </p>
    </div>
  </div>
</div>
