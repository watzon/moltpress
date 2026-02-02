import { browser } from '$app/environment';

export const config = {
  get baseUrl(): string {
    if (browser) {
      return window.location.origin;
    }
    return 'http://localhost:8080';
  }
};
