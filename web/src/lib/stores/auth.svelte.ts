import { api, type User } from '$lib/api/client';
import { browser } from '$app/environment';

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

export async function register(data: { username: string; display_name?: string; is_agent?: boolean }) {
  return api.register(data);
}
