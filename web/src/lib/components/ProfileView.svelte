<script lang="ts">
  import type { User, Post } from '$lib/api/client';
  import ProfileContent from './ProfileContent.svelte';
  import ProfileSidebar from './ProfileSidebar.svelte';

  let {
    user,
    posts,
    loadingMore = false,
    hasMore = false,
    onLoadMore,
    onFollow,
    onUnfollow,
  }: {
    user: User;
    posts: Post[];
    loadingMore?: boolean;
    hasMore?: boolean;
    onLoadMore?: () => void;
    onFollow?: () => Promise<void>;
    onUnfollow?: () => Promise<void>;
  } = $props();
</script>

<div class="profile-view">
  <div class="flex h-full">
    <div class="flex-1 min-w-0">
      <ProfileContent 
        {user}
        {posts}
        {loadingMore}
        {hasMore}
        {onLoadMore}
        {onFollow}
        {onUnfollow}
      />
    </div>
    <div class="hidden md:block w-[280px] flex-shrink-0 border-l border-[var(--color-surface-300)] bg-[var(--color-surface-100)]">
      <ProfileSidebar {user} />
    </div>
  </div>
</div>

<style>
  .profile-view {
    max-width: 950px;
    height: 100%;
  }
</style>
