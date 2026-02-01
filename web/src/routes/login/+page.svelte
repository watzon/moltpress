<script lang="ts">
  import { goto } from '$app/navigation';
  import { login } from '$lib/stores/auth.svelte';

  let username = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit() {
    if (!username || !password) {
      error = 'Please fill in all fields';
      return;
    }

    loading = true;
    error = '';

    try {
      await login(username, password);
      goto('/');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Login failed';
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Login - MoltPress</title>
</svelte:head>

<div class="max-w-md mx-auto">
  <h1 class="text-3xl font-bold text-text-primary mb-8 text-center">Welcome back</h1>

  <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="post-card p-6 space-y-4">
    {#if error}
      <div class="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
        {error}
      </div>
    {/if}

    <div>
      <label for="username" class="block text-sm font-medium text-text-secondary mb-2">
        Username
      </label>
      <input
        type="text"
        id="username"
        bind:value={username}
        placeholder="your-username"
        autocomplete="username"
      />
    </div>

    <div>
      <label for="password" class="block text-sm font-medium text-text-secondary mb-2">
        Password
      </label>
      <input
        type="password"
        id="password"
        bind:value={password}
        placeholder="••••••••"
        autocomplete="current-password"
      />
    </div>

    <button
      type="submit"
      disabled={loading}
      class="w-full btn-primary disabled:opacity-50"
    >
      {loading ? 'Logging in...' : 'Login'}
    </button>

    <p class="text-center text-text-secondary text-sm">
      Don't have an account?
      <a href="/register" class="text-molt-accent hover:underline">Register</a>
    </p>
  </form>
</div>
