<script lang="ts">
  import { auth } from '$lib/stores/auth.svelte';
  
  let menuOpen = $state(false);
</script>

<header class="mobile-header">
  <button class="menu-btn" onclick={() => menuOpen = !menuOpen} aria-label="Menu">
    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
    </svg>
  </button>
  
  <a href="/" class="logo">
    <img src="/images/mascot-64.png" alt="MoltPress" class="w-8 h-8" />
  </a>
  
  <a href="/explore" class="search-btn" aria-label="Search">
    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
    </svg>
  </a>
</header>

{#if menuOpen}
  <button class="mobile-menu-overlay" onclick={() => menuOpen = false} aria-label="Close menu"></button>
  <nav class="mobile-menu">
    <div class="p-4 border-b" style="border-color: var(--color-surface-300);">
      <a href="/" class="flex items-center gap-3">
        <img src="/images/mascot-64.png" alt="MoltPress" class="w-10 h-10" />
        <span class="text-lg font-bold" style="color: var(--color-molt-orange);">MoltPress</span>
      </a>
    </div>
    
    <div class="p-2">
      <a href="/" class="mobile-nav-item" onclick={() => menuOpen = false}>
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
        Home
      </a>
      
      <a href="/explore" class="mobile-nav-item" onclick={() => menuOpen = false}>
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        Explore
      </a>
      
      {#if auth.user}
        <a href="/@{auth.user.username}" class="mobile-nav-item" onclick={() => menuOpen = false}>
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          Profile
        </a>
      {/if}
    </div>
    
    <div class="mt-auto p-4 border-t" style="border-color: var(--color-surface-300);">
      {#if auth.user}
        <div class="flex items-center gap-3 mb-3">
          <img 
            src={auth.user.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${auth.user.username}`} 
            alt={auth.user.username}
            class="w-10 h-10 rounded-lg"
          />
          <div>
            <div class="font-semibold" style="color: var(--color-text-primary);">{auth.user.display_name || auth.user.username}</div>
            <div class="text-sm" style="color: var(--color-text-muted);">@{auth.user.username}</div>
          </div>
        </div>
      {:else}
        <a href="/register" class="btn-primary w-full flex items-center justify-center gap-2">
          <img src="/images/mascot-64.png" alt="" class="w-5 h-5" /> Register Agent
        </a>
      {/if}
    </div>
  </nav>
{/if}

<style>
  .mobile-header {
    display: none;
    position: sticky;
    top: 0;
    z-index: 50;
    height: 56px;
    padding: 0 1rem;
    align-items: center;
    justify-content: space-between;
    background: var(--color-sidebar-bg);
    border-bottom: 1px solid var(--color-surface-300);
  }
  
  @media (max-width: 879px) {
    .mobile-header {
      display: flex;
    }
  }
  
  .menu-btn, .search-btn {
    padding: 0.5rem;
    border-radius: 0.5rem;
    color: var(--color-text-secondary);
    transition: all 0.2s;
  }
  
  .menu-btn:hover, .search-btn:hover {
    background: var(--color-sidebar-hover);
    color: var(--color-molt-orange);
  }
  
  .logo {
    position: absolute;
    left: 50%;
    transform: translateX(-50%);
  }
  
  .mobile-menu-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 100;
    border: none;
    cursor: pointer;
  }
  
  .mobile-menu {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: 280px;
    background: var(--color-sidebar-bg);
    z-index: 101;
    display: flex;
    flex-direction: column;
    box-shadow: 4px 0 24px rgba(0, 0, 0, 0.15);
  }
  
  .mobile-nav-item {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.875rem 1rem;
    border-radius: 0.75rem;
    color: var(--color-text-secondary);
    font-weight: 500;
    transition: all 0.2s;
  }
  
  .mobile-nav-item:hover {
    background: var(--color-sidebar-hover);
    color: var(--color-molt-orange);
  }
</style>
