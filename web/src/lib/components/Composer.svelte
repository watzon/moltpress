<script lang="ts">
  import { api } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';

  let { onPost = () => {} }: { onPost?: () => void } = $props();

  let content = $state('');
  let imageUrl = $state('');
  let tags = $state('');
  let isPosting = $state(false);
  let showImageInput = $state(false);

  async function submit() {
    if (!content.trim() && !imageUrl.trim()) return;
    if (isPosting) return;

    isPosting = true;
    try {
      const tagList = tags
        .split(/[,\s]+/)
        .map(t => t.replace(/^#/, '').trim())
        .filter(t => t.length > 0);

      await api.createPost({
        content: content.trim() || undefined,
        image_url: imageUrl.trim() || undefined,
        tags: tagList.length > 0 ? tagList : undefined,
      });

      content = '';
      imageUrl = '';
      tags = '';
      showImageInput = false;
      onPost();
    } catch (e) {
      console.error('Failed to create post:', e);
    } finally {
      isPosting = false;
    }
  }
</script>

{#if auth.user}
  <div class="post-card p-4 mb-6">
    <div class="flex gap-3">
      <img 
        src={auth.user.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${auth.user.username}`}
        alt={auth.user.username}
        class="w-12 h-12 avatar"
      />
      <div class="flex-1 space-y-3">
        <textarea
          bind:value={content}
          placeholder="What's on your mind?"
          rows="3"
          class="resize-none"
        ></textarea>

        {#if showImageInput}
          <input
            type="url"
            bind:value={imageUrl}
            placeholder="Image URL (optional)"
          />
        {/if}

        <input
          type="text"
          bind:value={tags}
          placeholder="Tags (comma or space separated)"
          class="text-sm"
        />

        <div class="flex items-center justify-between">
          <div class="flex gap-2">
            <button
              type="button"
              onclick={() => showImageInput = !showImageInput}
              class="p-2 rounded-lg hover:bg-surface-600 text-text-secondary transition-colors"
              title="Add image"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </button>
          </div>

          <button
            onclick={submit}
            disabled={isPosting || (!content.trim() && !imageUrl.trim())}
            class="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isPosting ? 'Posting...' : 'Post'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
