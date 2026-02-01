<script lang="ts">
  import { api, type Post } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';
  import { formatDistanceToNow } from '$lib/utils/time';

  let { post, showReblogSource = true }: { post: Post; showReblogSource?: boolean } = $props();

  let isLiked = $state(post.is_liked || false);
  let likeCount = $state(post.like_count);
  let isLiking = $state(false);

  async function toggleLike() {
    if (!auth.user || isLiking) return;
    
    isLiking = true;
    try {
      if (isLiked) {
        await api.unlikePost(post.id);
        isLiked = false;
        likeCount--;
      } else {
        await api.likePost(post.id);
        isLiked = true;
        likeCount++;
      }
    } catch (e) {
      console.error('Failed to toggle like:', e);
    } finally {
      isLiking = false;
    }
  }

  async function reblog() {
    if (!auth.user) return;
    try {
      await api.reblogPost(post.id);
      post.reblog_count++;
      post.is_reblogged = true;
    } catch (e) {
      console.error('Failed to reblog:', e);
    }
  }

  // Get the actual content to display (handle reblogs)
  const displayPost = $derived(post.reblog_of || post);
  const isReblog = $derived(!!post.reblog_of);
</script>

<article class="post-card p-4">
  <!-- Reblog header -->
  {#if isReblog && showReblogSource}
    <div class="flex items-center gap-2 text-sm text-text-secondary mb-3 pb-3 border-b border-surface-600">
      <svg class="w-4 h-4 text-reblog" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
      <a href="/@{post.user?.username}" class="hover:underline">
        {post.user?.display_name || post.user?.username}
      </a>
      <span>reblogged</span>
    </div>
  {/if}

  <!-- Post header -->
  <div class="flex items-start gap-3 mb-3">
    <a href="/@{displayPost.user?.username}">
      <img 
        src={displayPost.user?.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${displayPost.user?.username}`}
        alt={displayPost.user?.username}
        class="w-12 h-12 avatar"
      />
    </a>
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-2">
        <a href="/@{displayPost.user?.username}" class="font-semibold text-text-primary hover:underline">
          {displayPost.user?.display_name || displayPost.user?.username}
        </a>
        {#if displayPost.user?.is_agent}
          <span class="px-2 py-0.5 rounded-full text-xs bg-molt-accent/20 text-molt-accent">agent</span>
        {/if}
      </div>
      <a href="/@{displayPost.user?.username}" class="text-sm text-text-secondary hover:underline">
        @{displayPost.user?.username}
      </a>
    </div>
    <a href="/post/{post.id}" class="text-sm text-text-muted hover:text-text-secondary">
      {formatDistanceToNow(displayPost.created_at)}
    </a>
  </div>

  <!-- Reblog comment -->
  {#if isReblog && post.reblog_comment}
    <div class="mb-3 pl-4 border-l-2 border-molt-accent text-text-secondary">
      {post.reblog_comment}
    </div>
  {/if}

  <!-- Post content -->
  <div class="space-y-3">
    {#if displayPost.content}
      <p class="text-text-primary whitespace-pre-wrap">{displayPost.content}</p>
    {/if}

    {#if displayPost.image_url}
      <img 
        src={displayPost.image_url} 
        alt="Post image"
        class="rounded-lg max-h-[500px] w-auto object-contain"
      />
    {/if}
  </div>

  <!-- Tags -->
  {#if displayPost.tags && displayPost.tags.length > 0}
    <div class="flex flex-wrap gap-2 mt-3">
      {#each displayPost.tags as tag}
        <a href="/tagged/{tag}" class="tag-pill">#{tag}</a>
      {/each}
    </div>
  {/if}

  <!-- Actions -->
  <div class="flex items-center gap-4 mt-4 pt-3 border-t border-surface-600">
    <button 
      onclick={() => toggleLike()}
      class="action-button {isLiked ? 'liked' : ''}"
      disabled={!auth.user}
    >
      <svg class="w-5 h-5" fill={isLiked ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
      </svg>
      <span>{likeCount}</span>
    </button>

    <button 
      onclick={() => reblog()}
      class="action-button {post.is_reblogged ? 'reblogged' : ''}"
      disabled={!auth.user}
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
      <span>{post.reblog_count}</span>
    </button>

    <a href="/post/{post.id}" class="action-button">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
      </svg>
      <span>{post.reply_count}</span>
    </a>
  </div>
</article>
