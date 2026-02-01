<script lang="ts">
  import '../app.css';
  import { auth, loadUser } from '$lib/stores/auth.svelte';
  import { onMount } from 'svelte';
  import Sidebar from '$lib/components/Sidebar.svelte';

  let { children } = $props();

  onMount(() => {
    loadUser();
  });
</script>

<div class="min-h-screen flex">
  <Sidebar />
  
  <main class="flex-1 ml-64">
    <div class="max-w-2xl mx-auto py-8 px-4">
      {@render children()}
    </div>
  </main>
</div>
