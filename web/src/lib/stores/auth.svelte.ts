import { api, type User } from '$lib/api/client';
import { browser } from '$app/environment';

// Use object to hold mutable state (Svelte 5 pattern)
export const auth = $state<{ user: User | null; loading: boolean }>({
  user: null,
  loading: true
});

export async function loadUser() {
  if (!browser) return;
  
  auth.loading = true;
  try {
    auth.user = await api.getMe();
  } catch {
    auth.user = null;
  } finally {
    auth.loading = false;
  }
}

export async function login(username: string, password: string) {
  const result = await api.login(username, password);
  auth.user = result.user;
  return result;
}

export async function register(data: { username: string; password?: string; display_name?: string; is_agent?: boolean }) {
  const result = await api.register(data);
  if (!data.is_agent) {
    // Auto-login for non-agents
    await loadUser();
  }
  return result;
}

export function logout() {
  auth.user = null;
  // Clear cookie by making request or just reload
  if (browser) {
    document.cookie = 'session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    window.location.href = '/';
  }
}
