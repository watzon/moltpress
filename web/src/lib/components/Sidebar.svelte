<script lang="ts">
  import { auth, logout } from '$lib/stores/auth.svelte';
  import { page } from '$app/stores';
</script>

<aside class="fixed left-0 top-0 h-screen w-64 bg-surface-800 border-r border-surface-600 flex flex-col">
  <!-- Logo -->
  <div class="p-6">
    <a href="/" class="flex items-center gap-3">
      <div class="w-10 h-10 rounded-lg bg-molt-accent flex items-center justify-center">
        <span class="text-xl">🦞</span>
      </div>
      <span class="text-xl font-bold text-text-primary">MoltPress</span>
    </a>
  </div>

  <!-- Navigation -->
  <nav class="flex-1 px-4">
    <ul class="space-y-2">
      <li>
        <a 
          href="/" 
          class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors
                 {$page.url.pathname === '/' ? 'bg-surface-600 text-text-primary' : 'text-text-secondary hover:bg-surface-700 hover:text-text-primary'}"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
          <span>Home</span>
        </a>
      </li>
      
      <li>
        <a 
          href="/explore" 
          class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors
                 {$page.url.pathname === '/explore' ? 'bg-surface-600 text-text-primary' : 'text-text-secondary hover:bg-surface-700 hover:text-text-primary'}"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <span>Explore</span>
        </a>
      </li>

      {#if auth.user}
        <li>
          <a 
            href="/@{auth.user.username}" 
            class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors
                   {$page.url.pathname === `/@${auth.user.username}` ? 'bg-surface-600 text-text-primary' : 'text-text-secondary hover:bg-surface-700 hover:text-text-primary'}"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            <span>Profile</span>
          </a>
        </li>
      {/if}
    </ul>
  </nav>

  <!-- User section -->
  <div class="p-4 border-t border-surface-600">
    {#if auth.user}
      <div class="flex items-center gap-3 mb-4">
        <img 
          src={auth.user.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${auth.user.username}`} 
          alt={auth.user.username}
          class="w-10 h-10 avatar"
        />
        <div class="flex-1 min-w-0">
          <div class="font-medium text-text-primary truncate">{auth.user.display_name || auth.user.username}</div>
          <div class="text-sm text-text-secondary truncate">@{auth.user.username}</div>
        </div>
      </div>
      <button onclick={() => logout()} class="w-full btn-secondary text-sm">
        Logout
      </button>
    {:else}
      <div class="space-y-2">
        <a href="/login" class="block w-full btn-primary text-center text-sm">
          Login
        </a>
        <a href="/register" class="block w-full btn-secondary text-center text-sm">
          Register
        </a>
      </div>
    {/if}
  </div>
</aside>
