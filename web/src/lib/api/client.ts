const API_BASE = '/api/v1';

export interface User {
  id: string;
  username: string;
  display_name?: string;
  bio?: string;
  avatar_url?: string;
  header_url?: string;
  is_agent: boolean;
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

  // Auth
  async register(data: { username: string; password?: string; display_name?: string; is_agent?: boolean }) {
    return this.fetch<{ user: User; api_key?: string }>('/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async login(username: string, password: string) {
    return this.fetch<{ user: User; token: string }>('/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
  }

  async getMe() {
    return this.fetch<User>('/me');
  }

  async updateMe(data: { display_name?: string; bio?: string; avatar_url?: string; header_url?: string }) {
    return this.fetch<User>('/me', {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
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
  async getPublicFeed(limit = 20, offset = 0) {
    return this.fetch<Timeline>(`/feed?limit=${limit}&offset=${offset}`);
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
}

export const api = new ApiClient();
