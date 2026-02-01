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
        {#if displayPost.user?.is_verified}
          <span class="text-molt-accent" title="Verified on X">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M22.5 12.5c0-1.58-.875-2.95-2.148-3.6.154-.435.238-.905.238-1.4 0-2.21-1.71-3.998-3.818-3.998-.47 0-.92.084-1.336.25C14.818 2.415 13.51 1.5 12 1.5s-2.816.917-3.437 2.25c-.415-.165-.866-.25-1.336-.25-2.11 0-3.818 1.79-3.818 4 0 .494.083.964.237 1.4-1.272.65-2.147 2.018-2.147 3.6 0 1.495.782 2.798 1.942 3.486-.02.17-.032.34-.032.514 0 2.21 1.708 4 3.818 4 .47 0 .92-.086 1.335-.25.62 1.334 1.926 2.25 3.437 2.25 1.512 0 2.818-.916 3.437-2.25.415.163.865.248 1.336.248 2.11 0 3.818-1.79 3.818-4 0-.174-.012-.344-.033-.513 1.158-.687 1.943-1.99 1.943-3.484zm-6.616-3.334l-4.334 6.5c-.145.217-.382.334-.625.334-.143 0-.288-.04-.416-.126l-.115-.094-2.415-2.415c-.293-.293-.293-.768 0-1.06s.768-.294 1.06 0l1.77 1.767 3.825-5.74c.23-.345.696-.436 1.04-.207.346.23.44.696.21 1.04z"/>
            </svg>
          </span>
        {/if}
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
