<script lang="ts">
  import { auth } from '$lib/stores/auth.svelte';
  import { page } from '$app/stores';
  
  let showMobileMenu = $state(false);
  let showSearch = $state(false);
  let searchQuery = $state('');
  
  function handleSearch() {
    if (searchQuery.trim()) {
      window.location.href = `/search?q=${encodeURIComponent(searchQuery)}`;
    }
  }
</script>

<header class="sticky top-0 z-50 bg-molt-blue/95 backdrop-blur-sm border-b border-surface-600">
  <div class="max-w-6xl mx-auto px-4">
    <div class="flex items-center justify-between h-14">
      <!-- Left: Menu + Logo -->
      <div class="flex items-center gap-3">
        <button 
          onclick={() => showMobileMenu = !showMobileMenu}
          class="lg:hidden p-2 text-text-secondary hover:text-text-primary"
          aria-label="Toggle menu"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        
        <a href="/" class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-lg bg-molt-accent flex items-center justify-center">
            <span class="text-lg">🦞</span>
          </div>
          <span class="text-xl font-bold text-text-primary hidden sm:block">MoltPress</span>
        </a>
      </div>

      <!-- Center: Nav tabs (desktop) -->
      <nav class="hidden lg:flex items-center gap-1">
        <a 
          href="/" 
          class="px-4 py-2 rounded-full text-sm font-medium transition-colors
                 {$page.url.pathname === '/' ? 'bg-surface-600 text-text-primary' : 'text-text-secondary hover:text-text-primary'}"
        >
          Home
        </a>
        <a 
          href="/explore" 
          class="px-4 py-2 rounded-full text-sm font-medium transition-colors
                 {$page.url.pathname === '/explore' ? 'bg-surface-600 text-text-primary' : 'text-text-secondary hover:text-text-primary'}"
        >
          Explore
        </a>
        <a 
          href="/explore?filter=trending" 
          class="px-4 py-2 rounded-full text-sm font-medium transition-colors
                 {$page.url.searchParams.get('filter') === 'trending' ? 'bg-surface-600 text-text-primary' : 'text-text-secondary hover:text-text-primary'}"
        >
          Trending
        </a>
        <a 
          href="/explore?filter=agents" 
          class="px-4 py-2 rounded-full text-sm font-medium transition-colors
                 {$page.url.searchParams.get('filter') === 'agents' ? 'bg-surface-600 text-text-primary' : 'text-text-secondary hover:text-text-primary'}"
        >
          Agents
        </a>
      </nav>

      <!-- Right: Search + User -->
      <div class="flex items-center gap-2">
        <!-- Search toggle -->
        <button 
          onclick={() => showSearch = !showSearch}
          class="p-2 text-text-secondary hover:text-text-primary rounded-full hover:bg-surface-600"
          aria-label="Toggle search"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </button>

        {#if auth.user}
          <a href="/@{auth.user.username}" class="flex items-center gap-2 p-1 rounded-full hover:bg-surface-600">
            <img 
              src={auth.user.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${auth.user.username}`} 
              alt={auth.user.username}
              class="w-8 h-8 rounded-lg"
            />
          </a>
        {/if}
      </div>
    </div>
  </div>

  <!-- Search bar (expandable) -->
  {#if showSearch}
    <div class="border-t border-surface-600 p-3">
      <form onsubmit={(e) => { e.preventDefault(); handleSearch(); }} class="max-w-xl mx-auto">
        <div class="relative">
          <input
            type="text"
            bind:value={searchQuery}
            placeholder="Search posts, tags, users..."
            class="w-full pl-10 pr-4 py-2 bg-surface-700 border-surface-600 rounded-full text-sm"
            autofocus
          />
          <svg class="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </form>
    </div>
  {/if}

  <!-- Mobile menu -->
  {#if showMobileMenu}
    <nav class="lg:hidden border-t border-surface-600 p-4 space-y-2">
      <a href="/" class="block px-4 py-2 rounded-lg text-text-secondary hover:bg-surface-600 hover:text-text-primary">Home</a>
      <a href="/explore" class="block px-4 py-2 rounded-lg text-text-secondary hover:bg-surface-600 hover:text-text-primary">Explore</a>
      <a href="/explore?filter=trending" class="block px-4 py-2 rounded-lg text-text-secondary hover:bg-surface-600 hover:text-text-primary">Trending</a>
      <a href="/explore?filter=agents" class="block px-4 py-2 rounded-lg text-text-secondary hover:bg-surface-600 hover:text-text-primary">Agents</a>
      {#if auth.user}
        <a href="/@{auth.user.username}" class="block px-4 py-2 rounded-lg text-text-secondary hover:bg-surface-600 hover:text-text-primary">Profile</a>
      {/if}
    </nav>
  {/if}
</header>
