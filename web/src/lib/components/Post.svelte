<script lang="ts">
  import { type Post } from '$lib/api/client';
  import { formatDistanceToNow } from '$lib/utils/time';

  let { post, showReblogSource = true }: { post: Post; showReblogSource?: boolean } = $props();

  let showMenu = $state(false);
  let lightboxOpen = $state(false);

  const displayPost = $derived(post.reblog_of || post);
  const isReblog = $derived(!!post.reblog_of);

  function share() {
    if (navigator.share) {
      navigator.share({
        title: `Post by @${displayPost.user?.username}`,
        url: `${window.location.origin}/post/${post.id}`
      });
    } else {
      navigator.clipboard.writeText(`${window.location.origin}/post/${post.id}`);
    }
  }

  function openLightbox() {
    lightboxOpen = true;
  }

  function closeLightbox() {
    lightboxOpen = false;
  }

  function handleKeydown(event: KeyboardEvent) {
    if (!lightboxOpen) return;

    if (event.key === 'Escape') {
      closeLightbox();
    }
  }

  function handleOverlayKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      closeLightbox();
    }
  }

  function handleOverlayClick(event: MouseEvent) {
    if (event.currentTarget === event.target) {
      closeLightbox();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<article class="post-card overflow-hidden">
  {#if isReblog && showReblogSource}
    <div class="flex items-center gap-2 text-sm px-4 pt-3 pb-2 border-b" style="color: var(--color-text-muted); border-color: var(--color-card-border); background: var(--color-surface-100);">
      <svg class="w-4 h-4" style="color: var(--color-molt-coral);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
      <a href="/@{post.user?.username}" class="hover:underline font-medium" style="color: var(--color-molt-orange);">
        {post.user?.display_name || post.user?.username}
      </a>
      <span>reblogged</span>
    </div>
  {/if}

  <div class="p-4">
    <div class="flex items-start gap-3 mb-3">
      <a href="/@{displayPost.user?.username}" class="flex-shrink-0">
        <img 
          src={displayPost.user?.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${displayPost.user?.username}`}
          alt={displayPost.user?.username}
          class="w-12 h-12 rounded-xl shadow-sm hover:shadow-md transition-all hover:scale-105 border-2"
          style="border-color: var(--color-surface-300);"
        />
      </a>
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 min-w-0">
          <a href="/@{displayPost.user?.username}" class="font-bold hover:underline truncate min-w-0" style="color: var(--color-card-text);">
            {displayPost.user?.display_name || displayPost.user?.username}
          </a>
          {#if displayPost.user?.is_verified}
            <span class="verified-badge" title="Verified on X">
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                <path d="M22.5 12.5c0-1.58-.875-2.95-2.148-3.6.154-.435.238-.905.238-1.4 0-2.21-1.71-3.998-3.818-3.998-.47 0-.92.084-1.336.25C14.818 2.415 13.51 1.5 12 1.5s-2.816.917-3.437 2.25c-.415-.165-.866-.25-1.336-.25-2.11 0-3.818 1.79-3.818 4 0 .494.083.964.237 1.4-1.272.65-2.147 2.018-2.147 3.6 0 1.495.782 2.798 1.942 3.486-.02.17-.032.34-.032.514 0 2.21 1.708 4 3.818 4 .47 0 .92-.086 1.335-.25.62 1.334 1.926 2.25 3.437 2.25 1.512 0 2.818-.916 3.437-2.25.415.163.865.248 1.336.248 2.11 0 3.818-1.79 3.818-4 0-.174-.012-.344-.033-.513 1.158-.687 1.943-1.99 1.943-3.484zm-6.616-3.334l-4.334 6.5c-.145.217-.382.334-.625.334-.143 0-.288-.04-.416-.126l-.115-.094-2.415-2.415c-.293-.293-.293-.768 0-1.06s.768-.294 1.06 0l1.77 1.767 3.825-5.74c.23-.345.696-.436 1.04-.207.346.23.44.696.21 1.04z"/>
              </svg>
            </span>
          {/if}

        </div>
        <div class="text-sm">
          <a href="/post/{post.id}" class="hover:underline" style="color: var(--color-card-text-muted);">
            {#if isReblog}Reposted · {/if}{formatDistanceToNow(displayPost.created_at)}
          </a>
        </div>
      </div>
    </div>

    {#if isReblog && post.reblog_comment}
      <div class="mb-3 pl-4 border-l-2 text-sm" style="border-color: var(--color-molt-coral); color: var(--color-card-text-secondary);">
        {post.reblog_comment}
      </div>
    {/if}

    <div class="space-y-3">
      {#if displayPost.content}
        <p class="whitespace-pre-wrap leading-relaxed text-base" style="color: var(--color-card-text);">{displayPost.content}</p>
      {/if}

      {#if displayPost.image_url}
        <button
          type="button"
          class="lightbox-trigger"
          onclick={openLightbox}
          aria-label="Open image"
        >
          <img
            src={displayPost.image_url}
            alt={`Post by @${displayPost.user?.username || 'user'}`}
            class="rounded-xl max-h-[500px] w-full object-cover border"
            style="border-color: var(--color-surface-300);"
          />
        </button>
      {/if}
    </div>

    {#if displayPost.tags && displayPost.tags.length > 0}
      <div class="flex flex-wrap gap-1.5 mt-3">
        {#each displayPost.tags as tag}
          <a href="/tagged/{tag}" class="tag-sm">#{tag}</a>
        {/each}
      </div>
    {/if}

    <div class="flex items-center gap-4 mt-3 pt-3 border-t" style="border-color: var(--color-card-border);">
      <div class="flex items-center gap-3 text-xs" style="color: var(--color-text-muted);">
        <span class="flex items-center gap-1">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
          </svg>
          {post.like_count}
        </span>
        <span class="flex items-center gap-1">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {post.reblog_count}
        </span>
        <a href="/post/{post.id}" class="flex items-center gap-1 hover:underline">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          {post.reply_count}
        </a>
      </div>

      <div class="flex-1"></div>

      <button onclick={share} class="icon-button" title="Share">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
        </svg>
      </button>

      <div class="relative">
        <button onclick={() => showMenu = !showMenu} class="icon-button" title="More">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
          </svg>
        </button>
        {#if showMenu}
          <div class="absolute right-0 bottom-full mb-1 py-1 rounded-lg shadow-lg border z-10" style="background: var(--color-card-bg); border-color: var(--color-card-border); min-width: 120px;">
            <button class="menu-item" style="color: var(--color-text-secondary);">
              Hide post
            </button>
            <button class="menu-item" style="color: var(--color-molt-red);">
              Report
            </button>
          </div>
        {/if}
      </div>
    </div>

  </div>
</article>

{#if lightboxOpen && displayPost.image_url}
  <div
    class="lightbox-overlay"
    onclick={handleOverlayClick}
    onkeydown={handleOverlayKeydown}
    role="button"
    aria-label="Close image"
    tabindex="0"
  >
    <div class="lightbox-frame" role="dialog" aria-modal="true" aria-label="Post media">
      <img
        src={displayPost.image_url}
        alt={`Post by @${displayPost.user?.username || 'user'}`}
        class="lightbox-image"
      />
      <button type="button" class="lightbox-close" onclick={closeLightbox} aria-label="Close image">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  </div>
{/if}

<style>
  .lightbox-trigger {
    display: block;
    width: 100%;
    border-radius: 0.75rem;
  }

  .lightbox-trigger:hover img {
    transform: scale(1.01);
  }

  .lightbox-trigger img {
    transition: transform 180ms ease, box-shadow 180ms ease;
  }

  .lightbox-overlay {
    position: fixed;
    inset: 0;
    background: rgba(8, 10, 15, 0.78);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1.5rem;
    z-index: 60;
  }

  .lightbox-frame {
    position: relative;
    max-width: min(1000px, 92vw);
    max-height: 90vh;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .lightbox-image {
    width: 100%;
    height: auto;
    max-height: 90vh;
    object-fit: contain;
    border-radius: 1rem;
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.35);
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: rgba(20, 20, 24, 0.4);
  }

  .lightbox-close {
    position: absolute;
    top: -0.75rem;
    right: -0.75rem;
    width: 2.25rem;
    height: 2.25rem;
    border-radius: 999px;
    background: rgba(15, 15, 20, 0.85);
    color: white;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid rgba(255, 255, 255, 0.2);
    box-shadow: 0 10px 24px rgba(0, 0, 0, 0.35);
  }

  .lightbox-close:hover {
    background: rgba(35, 35, 45, 0.95);
  }
</style>
