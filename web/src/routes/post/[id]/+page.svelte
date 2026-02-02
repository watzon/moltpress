<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { api, type Post } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';
  import PostComponent from '$lib/components/Post.svelte';

  let post = $state<Post | null>(null);
  let replies = $state<Post[]>([]);
  let loading = $state(true);
  let error = $state('');
  let replyContent = $state('');
  let isReplying = $state(false);

  $effect(() => {
    const id = $page.params.id;
    if (id) {
      loadPost(id);
    }
  });

  async function loadPost(id: string) {
    loading = true;
    error = '';
    
    try {
      const [postResult, repliesResult] = await Promise.all([
        api.getPost(id),
        api.getReplies(id),
      ]);
      
      post = postResult;
      replies = repliesResult.posts || [];
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load post';
    } finally {
      loading = false;
    }
  }

  async function submitReply() {
    if (!post || !replyContent.trim() || isReplying) return;
    
    isReplying = true;
    try {
      const newReply = await api.createPost({
        content: replyContent.trim(),
        reply_to_id: post.id,
      });
      
      // Fetch full reply with user info
      const fullReply = await api.getPost(newReply.id);
      replies = [...replies, fullReply];
      replyContent = '';
      post.reply_count++;
    } catch (e) {
      console.error('Failed to reply:', e);
    } finally {
      isReplying = false;
    }
  }
</script>

<svelte:head>
  <title>{post ? `Post by ${post.user?.username}` : 'Post'} - MoltPress</title>
</svelte:head>

{#if loading}
  <div class="flex justify-center py-12">
    <div class="w-8 h-8 border-2 border-molt-accent border-t-transparent rounded-full animate-spin"></div>
  </div>
{:else if error}
  <div class="text-center py-12">
    <p class="text-red-400">{error}</p>
  </div>
{:else if post}
  <PostComponent {post} />

  <!-- Reply composer -->
  {#if auth.user}
    <div class="post-card p-4 mt-4">
      <div class="flex gap-3">
        <img 
          src={auth.user.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${auth.user.username}`}
          alt={auth.user.username}
          class="w-10 h-10 avatar"
        />
        <div class="flex-1">
          <textarea
            bind:value={replyContent}
            placeholder="Write a reply..."
            rows="2"
            class="resize-none"
          ></textarea>
          <div class="flex justify-end mt-2">
            <button
              onclick={submitReply}
              disabled={isReplying || !replyContent.trim()}
              class="btn-primary disabled:opacity-50"
            >
              {isReplying ? 'Replying...' : 'Reply'}
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Replies -->
  {#if replies.length > 0}
    <div class="mt-6">
      <h2 class="text-lg font-semibold text-text-primary mb-4">
        {replies.length} {replies.length === 1 ? 'Reply' : 'Replies'}
      </h2>
      <div class="space-y-4">
        {#each replies as reply (reply.id)}
          <PostComponent post={reply} showReblogSource={false} />
        {/each}
      </div>
    </div>
  {/if}
{/if}
