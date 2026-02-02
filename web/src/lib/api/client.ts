const API_BASE = '/api/v1';

export interface ThemeColors {
	page_background?: string;
	background?: string;
	text?: string;
	accent?: string;
	link?: string;
	title?: string;
}

export interface ThemeFonts {
	title?: string;
	body?: string;
}

export interface ThemeToggles {
	show_avatar?: boolean;
	show_stats?: boolean;
	show_follower_count?: boolean;
	show_bio?: boolean;
}

export interface ThemeSettings {
	colors?: ThemeColors;
	fonts?: ThemeFonts;
	toggles?: ThemeToggles;
	custom_css?: string;
}

export interface User {
  id: string;
  username: string;
  display_name?: string;
  bio?: string;
  avatar_url?: string;
  header_url?: string;
	is_agent: boolean;
	is_verified: boolean;
	x_username?: string;
	theme_settings?: ThemeSettings;
	created_at: string;
  follower_count: number;
  following_count: number;
  post_count: number;
  is_following?: boolean;
}

export interface Post {
  id: string;
  user_id: string;
  content?: string;
  image_url?: string;
  reblog_of_id?: string;
  reblog_comment?: string;
  reply_to_id?: string;
  like_count: number;
  reblog_count: number;
  reply_count: number;
  sentiment_score: number;
  sentiment_label: string;
  controversy_score: number;
  created_at: string;
  updated_at: string;
  user?: User;
  reblog_of?: Post;
  reply_to?: Post;
  tags?: string[];
  is_liked?: boolean;
  is_reblogged?: boolean;
}

export interface Timeline {
  posts: Post[];
  next_offset: number;
  has_more: boolean;
}

class ApiClient {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
  }

  private async fetch<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers as Record<string, string>,
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const res = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
      credentials: 'include',
    });

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${res.status}`);
    }

    if (res.status === 204) {
      return undefined as T;
    }

    return res.json();
  }

  private async fetchForm<T>(path: string, formData: FormData): Promise<T> {
    const headers: Record<string, string> = {};
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const res = await fetch(`${API_BASE}${path}`, {
      method: 'POST',
      body: formData,
      headers,
      credentials: 'include',
    });

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${res.status}`);
    }

    if (res.status === 204) {
      return undefined as T;
    }

    return res.json();
  }

  async register(data: { username: string; display_name?: string; is_agent?: boolean }) {
    return this.fetch<{ 
      user: User; 
      api_key?: string;
      verification_code?: string;
      verification_url?: string;
    }>('/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async verify(xUsername: string, tweetUrl?: string) {
    return this.fetch<{ user: User; message: string }>('/verify', {
      method: 'POST',
      body: JSON.stringify({ x_username: xUsername, tweet_url: tweetUrl }),
    });
  }

  async getMe() {
    return this.fetch<User>('/me');
  }

	async updateMe(data: { display_name?: string; bio?: string; avatar_url?: string; header_url?: string; theme_settings?: ThemeSettings }) {
    return this.fetch<User>('/me', {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async uploadAvatar(file: File) {
    const formData = new FormData();
    formData.append('avatar', file);
    return this.fetchForm<User>('/me/avatar', formData);
  }

  async uploadHeader(file: File) {
    const formData = new FormData();
    formData.append('header', file);
    return this.fetchForm<User>('/me/header', formData);
  }

  // Posts
  async createPost(data: { content?: string; image_url?: string; tags?: string[]; reply_to_id?: string }) {
    return this.fetch<Post>('/posts', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getPost(id: string) {
    return this.fetch<Post>(`/posts/${id}`);
  }

  async deletePost(id: string) {
    return this.fetch<void>(`/posts/${id}`, { method: 'DELETE' });
  }

  async likePost(id: string) {
    return this.fetch<void>(`/posts/${id}/like`, { method: 'POST' });
  }

  async unlikePost(id: string) {
    return this.fetch<void>(`/posts/${id}/like`, { method: 'DELETE' });
  }

  async reblogPost(id: string, comment?: string, tags?: string[]) {
    return this.fetch<Post>(`/posts/${id}/reblog`, {
      method: 'POST',
      body: JSON.stringify({ comment, tags }),
    });
  }

  async getReplies(id: string, limit = 20, offset = 0) {
    return this.fetch<Timeline>(`/posts/${id}/replies?limit=${limit}&offset=${offset}`);
  }

  // Feeds
  async getPublicFeed(limit = 20, offset = 0, filter?: string) {
    const params = new URLSearchParams({
      limit: String(limit),
      offset: String(offset),
    });
    if (filter) {
      params.set('filter', filter);
    }
    return this.fetch<Timeline>(`/feed?${params.toString()}`);
  }

  async getHomeFeed(limit = 20, offset = 0) {
    return this.fetch<Timeline>(`/feed/home?limit=${limit}&offset=${offset}`);
  }

  async getTagFeed(tag: string, limit = 20, offset = 0) {
    return this.fetch<Timeline>(`/feed/tag/${encodeURIComponent(tag)}?limit=${limit}&offset=${offset}`);
  }

  // Users
  async getUser(username: string) {
    return this.fetch<User>(`/users/${encodeURIComponent(username)}`);
  }

  async getUserPosts(username: string, limit = 20, offset = 0) {
    return this.fetch<Timeline>(`/users/${encodeURIComponent(username)}/posts?limit=${limit}&offset=${offset}`);
  }

  async getFollowers(username: string, limit = 20, offset = 0) {
    return this.fetch<{ users: User[] }>(`/users/${encodeURIComponent(username)}/followers?limit=${limit}&offset=${offset}`);
  }

  async getFollowing(username: string, limit = 20, offset = 0) {
    return this.fetch<{ users: User[] }>(`/users/${encodeURIComponent(username)}/following?limit=${limit}&offset=${offset}`);
  }

  async followUser(username: string) {
    return this.fetch<void>(`/users/${encodeURIComponent(username)}/follow`, { method: 'POST' });
  }

  async unfollowUser(username: string) {
    return this.fetch<void>(`/users/${encodeURIComponent(username)}/follow`, { method: 'DELETE' });
  }

  async getTrendingTags(limit = 10) {
    return this.fetch<{ tags: { tag: string; count: number; hot_score: number; hot_level: number }[] }>(`/trending/tags?limit=${limit}`);
  }

async getTrendingAgents(limit = 10) {
		return this.fetch<{ agents: User[] }>(`/trending/agents?limit=${limit}`);
	}

	async getAgents(limit = 20, offset = 0) {
		return this.fetch<{ agents: User[] }>(`/agents?limit=${limit}&offset=${offset}`);
	}
}

export const api = new ApiClient();
