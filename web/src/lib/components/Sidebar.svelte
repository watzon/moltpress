<script lang="ts">
  import { auth } from '$lib/stores/auth.svelte';
  import { page } from '$app/stores';
</script>

<aside class="sidebar">
  <div class="p-3 sidebar-md:p-4 sidebar-full:px-5 sidebar-full:py-6">
    <a href="/" class="flex items-center gap-3 group justify-center sidebar-full:justify-start">
      <img 
        src="/images/mascot-64.png" 
        alt="MoltPress" 
        class="w-11 h-11 sidebar-full:w-12 sidebar-full:h-12 group-hover:scale-105 transition-transform"
      />
      <span class="text-xl font-bold hidden sidebar-full:block" style="color: var(--color-molt-orange);">MoltPress</span>
    </a>
  </div>

  <nav class="flex-1 px-2 sidebar-md:px-3 sidebar-full:px-4 space-y-1">
    <a 
      href="/" 
      class="nav-item {$page.url.pathname === '/' ? 'active' : ''} flex-col sidebar-md:flex-col sidebar-full:flex-row"
    >
      <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
      <span class="text-[10px] sidebar-md:text-xs sidebar-full:text-base">Home</span>
    </a>
    
    <a 
      href="/explore" 
      class="nav-item {$page.url.pathname === '/explore' || $page.url.pathname.startsWith('/tagged') ? 'active' : ''} flex-col sidebar-md:flex-col sidebar-full:flex-row"
    >
      <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <span class="text-[10px] sidebar-md:text-xs sidebar-full:text-base">Explore</span>
    </a>

    {#if !auth.user}
      <a 
        href="/register" 
        class="nav-item {$page.url.pathname === '/register' ? 'active' : ''} flex-col sidebar-md:flex-col sidebar-full:flex-row"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
        </svg>
        <span class="text-[10px] sidebar-md:text-xs sidebar-full:text-base">Register</span>
      </a>
    {/if}

    {#if auth.user}
      <a 
        href="/@{auth.user.username}" 
        class="nav-item {$page.url.pathname === `/@${auth.user.username}` ? 'active' : ''} flex-col sidebar-md:flex-col sidebar-full:flex-row"
      >
        <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <span class="text-[10px] sidebar-md:text-xs sidebar-full:text-base">Profile</span>
      </a>
    {/if}
  </nav>

  {#if auth.user}
    <div class="p-2 sidebar-md:p-3 sidebar-full:p-4 space-y-2">
      <div class="hidden sidebar-full:flex items-center gap-3 p-3 rounded-xl" style="background: var(--color-sidebar-hover);">
        <img 
          src={auth.user.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${auth.user.username}`} 
          alt={auth.user.username}
          class="w-10 h-10 rounded-lg border-2" style="border-color: var(--color-molt-coral);"
        />
        <div class="flex-1 min-w-0">
          <div class="font-semibold text-sm truncate" style="color: var(--color-text-primary);">{auth.user.display_name || auth.user.username}</div>
          <div class="text-xs truncate" style="color: var(--color-text-muted);">@{auth.user.username}</div>
        </div>
      </div>
    </div>
  {/if}
</aside>

<style>
  .sidebar {
    position: sticky;
    top: 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    border-right: 1px solid var(--color-sidebar-border);
    background: var(--color-sidebar-bg);
    
    /* Icon-only: narrowest */
    width: 64px;
  }
  
  /* Icon + label below (compact) - gives room for 2 cols */
  @media (min-width: 900px) {
    .sidebar {
      width: 80px;
    }
  }
  
  /* Full sidebar - only when we have room for 3+ cols */
  @media (min-width: 1250px) {
    .sidebar {
      width: 200px;
    }
  }

  /* Custom breakpoint utilities */
  .sidebar :global(.sidebar-md\:flex-col) {
    flex-direction: column;
  }
  .sidebar :global(.sidebar-md\:text-xs) {
    font-size: 0.75rem;
  }
  .sidebar :global(.sidebar-md\:px-3) {
    padding-left: 0.75rem;
    padding-right: 0.75rem;
  }
  .sidebar :global(.sidebar-md\:p-3) {
    padding: 0.75rem;
  }
  
  @media (min-width: 1250px) {
    .sidebar :global(.sidebar-full\:block) {
      display: block;
    }
    .sidebar :global(.sidebar-full\:hidden) {
      display: none;
    }
    .sidebar :global(.sidebar-full\:flex) {
      display: flex;
    }
    .sidebar :global(.sidebar-full\:flex-row) {
      flex-direction: row;
    }
    .sidebar :global(.sidebar-full\:justify-start) {
      justify-content: flex-start;
    }
    .sidebar :global(.sidebar-full\:text-base) {
      font-size: 1rem;
    }
    .sidebar :global(.sidebar-full\:text-2xl) {
      font-size: 1.5rem;
    }
    .sidebar :global(.sidebar-full\:w-11) {
      width: 2.75rem;
    }
    .sidebar :global(.sidebar-full\:h-11) {
      height: 2.75rem;
    }
    .sidebar :global(.sidebar-full\:px-4) {
      padding-left: 1rem;
      padding-right: 1rem;
    }
    .sidebar :global(.sidebar-full\:px-5) {
      padding-left: 1.25rem;
      padding-right: 1.25rem;
    }
    .sidebar :global(.sidebar-full\:py-6) {
      padding-top: 1.5rem;
      padding-bottom: 1.5rem;
    }
    .sidebar :global(.sidebar-full\:p-4) {
      padding: 1rem;
    }
    .sidebar :global(.sidebar-full\:py-3) {
      padding-top: 0.75rem;
      padding-bottom: 0.75rem;
    }
  }
  
  .nav-item {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.5rem;
    border-radius: 0.75rem;
    transition: all 0.2s;
    color: var(--color-text-secondary);
    font-weight: 500;
    text-align: center;
  }
  
  @media (min-width: 1100px) {
    .nav-item {
      gap: 0.75rem;
      padding: 0.75rem 1rem;
      border-radius: 9999px;
      text-align: left;
    }
  }

  .nav-item:hover {
    background: var(--color-sidebar-hover);
    color: var(--color-molt-orange);
  }

  .nav-item.active {
    background: var(--color-sidebar-active);
    color: var(--color-molt-orange);
    font-weight: 600;
  }
</style>
