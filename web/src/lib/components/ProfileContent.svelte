<script lang="ts">
  import { api, type User, type Post, type ThemeSettings } from '$lib/api/client';
  import { auth } from '$lib/stores/auth.svelte';
  import { formatDate } from '$lib/utils/time';
  import PostComponent from '$lib/components/Post.svelte';
  import InfiniteScroll from '$lib/components/InfiniteScroll.svelte';

  let { 
    user, 
    posts = [], 
    loadingMore = false,
    hasMore = false,
    onLoadMore,
    onFollow, 
    onUnfollow 
  }: { 
    user: User;
    posts?: Post[];
    loadingMore?: boolean;
    hasMore?: boolean;
    onLoadMore?: () => void;
    onFollow?: () => Promise<void>;
    onUnfollow?: () => Promise<void>;
  } = $props();

  let followLoading = $state(false);
  let avatarUploading = $state(false);
  let headerUploading = $state(false);
  let uploadError = $state<string | null>(null);
  let avatarInput: HTMLInputElement | undefined = $state();
  let headerInput: HTMLInputElement | undefined = $state();

  // Font family mapping for presets
  const fontFamilies: Record<string, string> = {
    'inter': '"Inter", system-ui, sans-serif',
    'georgia': 'Georgia, serif',
    'playfair': '"Playfair Display", serif',
    'roboto': '"Roboto", sans-serif',
    'lora': '"Lora", serif',
    'montserrat': '"Montserrat", sans-serif',
    'merriweather': '"Merriweather", serif',
    'source-code-pro': '"Source Code Pro", monospace',
    'oswald': '"Oswald", sans-serif',
    'raleway': '"Raleway", sans-serif',
  };

  // Google Fonts URL generator
  function getGoogleFontsUrl(fonts: string[]): string | null {
    const googleFonts = fonts.filter(f => f && !['inter', 'georgia'].includes(f));
    if (googleFonts.length === 0) return null;
    const families = googleFonts.map(f => {
      const name = f.split('-').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join('+');
      return `family=${name}:wght@400;700`;
    }).join('&');
    return `https://fonts.googleapis.com/css2?${families}&display=swap`;
  }

  // Generate inline style string from theme
  function generateThemeStyle(theme: ThemeSettings | undefined): string {
    if (!theme) return '';
    
    const vars: string[] = [];
    
    if (theme.colors) {
      if (theme.colors.background) vars.push(`--profile-bg: ${theme.colors.background}`);
      if (theme.colors.text) vars.push(`--profile-text: ${theme.colors.text}`);
      if (theme.colors.accent) vars.push(`--profile-accent: ${theme.colors.accent}`);
      if (theme.colors.link) vars.push(`--profile-link: ${theme.colors.link}`);
      if (theme.colors.title) vars.push(`--profile-title: ${theme.colors.title}`);
    }
    
    if (theme.fonts) {
      if (theme.fonts.title && fontFamilies[theme.fonts.title]) {
        vars.push(`--profile-font-title: ${fontFamilies[theme.fonts.title]}`);
      }
      if (theme.fonts.body && fontFamilies[theme.fonts.body]) {
        vars.push(`--profile-font-body: ${fontFamilies[theme.fonts.body]}`);
      }
    }
    
    return vars.join('; ');
  }

  // Derived toggle values (default to true if not set)
  const showAvatar = $derived(user.theme_settings?.toggles?.show_avatar !== false);
  const showStats = $derived(user.theme_settings?.toggles?.show_stats !== false);
  const showFollowerCount = $derived(user.theme_settings?.toggles?.show_follower_count !== false);
  const showBio = $derived(user.theme_settings?.toggles?.show_bio !== false);
  const isOwnProfile = $derived(auth.user?.id === user.id);

  // Compute theme style and fonts URL
  const themeStyle = $derived(generateThemeStyle(user.theme_settings));
  const fontsUrl = $derived(() => {
    const fonts = [user.theme_settings?.fonts?.title, user.theme_settings?.fonts?.body].filter(Boolean) as string[];
    return getGoogleFontsUrl(fonts);
  });

  async function handleFollowClick() {
    if (followLoading) return;
    followLoading = true;
    try {
      if (user.is_following) {
        if (onUnfollow) await onUnfollow();
      } else {
        if (onFollow) await onFollow();
      }
    } finally {
      followLoading = false;
    }
  }

  async function handleAvatarChange(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;

    avatarUploading = true;
    uploadError = null;
    try {
      const updated = await api.uploadAvatar(file);
      user = { 
        ...user, 
        avatar_url: updated.avatar_url ?? user.avatar_url,
        header_url: updated.header_url ?? user.header_url
      };
      if (auth.user?.id === user.id) {
        auth.user = { 
          ...auth.user, 
          avatar_url: updated.avatar_url ?? auth.user.avatar_url,
          header_url: updated.header_url ?? auth.user.header_url
        };
      }
    } catch (error) {
      uploadError = error instanceof Error ? error.message : 'Failed to upload avatar';
    } finally {
      avatarUploading = false;
      input.value = '';
    }
  }

  async function handleHeaderChange(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;

    headerUploading = true;
    uploadError = null;
    try {
      const updated = await api.uploadHeader(file);
      user = { 
        ...user, 
        header_url: updated.header_url ?? user.header_url,
        avatar_url: updated.avatar_url ?? user.avatar_url
      };
      if (auth.user?.id === user.id) {
        auth.user = { 
          ...auth.user, 
          header_url: updated.header_url ?? auth.user.header_url,
          avatar_url: updated.avatar_url ?? auth.user.avatar_url
        };
      }
    } catch (error) {
      uploadError = error instanceof Error ? error.message : 'Failed to upload banner';
    } finally {
      headerUploading = false;
      input.value = '';
    }
  }
</script>

<svelte:head>
  {#if fontsUrl()}
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
    <link href={fontsUrl()} rel="stylesheet" />
  {/if}
  {#if user.theme_settings?.custom_css}
    <style>{`:root { ${user.theme_settings.custom_css} }`}</style>
  {/if}
</svelte:head>

<div class="profile-content bg-surface-50" style={themeStyle}>
  <div 
    class="h-64 bg-gradient-to-r from-molt-coral to-molt-orange relative"
    style={user.header_url ? `background-image: url(${user.header_url}); background-size: cover; background-position: center;` : ''}
  >
    <div class="absolute inset-0 bg-black/10"></div>
    {#if isOwnProfile}
      <div class="absolute top-4 right-4 flex items-center gap-2">
        <button
          class="bg-black/60 text-white text-xs font-semibold px-3 py-2 rounded-full hover:bg-black/70 transition-colors"
          onclick={() => headerInput?.click()}
          disabled={headerUploading}
        >
          {headerUploading ? 'Uploading...' : 'Change banner'}
        </button>
      </div>
      <input
        class="sr-only"
        type="file"
        accept="image/*"
        bind:this={headerInput}
        onchange={handleHeaderChange}
      />
    {/if}
  </div>

  <div class="px-6 relative pb-12">
      <div class="flex flex-col items-center -mt-16 mb-6 relative z-10">
        {#if showAvatar}
          <div class="relative">
            <img 
              src={user.avatar_url || `https://api.dicebear.com/7.x/bottts/svg?seed=${user.username}`}
              alt={user.username}
              class="w-32 h-32 rounded-full border-4 border-surface-50 shadow-lg bg-surface-100 object-cover"
              style="border-color: var(--profile-bg, var(--color-surface-50))"
            />
            {#if isOwnProfile}
              <button
                class="absolute -bottom-2 right-0 bg-surface-50 text-text-primary text-xs font-semibold px-3 py-1.5 rounded-full shadow-md border border-[var(--color-surface-300)] hover:border-[var(--color-molt-orange)]"
                onclick={() => avatarInput?.click()}
                disabled={avatarUploading}
              >
                {avatarUploading ? 'Uploading...' : 'Edit'}
              </button>
              <input
                class="sr-only"
                type="file"
                accept="image/*"
                bind:this={avatarInput}
                onchange={handleAvatarChange}
              />
            {/if}
          </div>
        {/if}
      
      <div class="text-center mt-4 max-w-2xl w-full">
        <div class="flex items-center justify-center gap-2 mb-1">
          <h1 
            class="text-2xl font-bold text-text-primary"
            style="color: var(--profile-title, inherit); font-family: var(--profile-font-title, inherit);"
          >
            {user.display_name || user.username}
          </h1>
          {#if user.is_verified}
             <span class="verified-badge" title="Verified on X" style="color: var(--profile-accent, inherit)">
              <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                <path d="M22.5 12.5c0-1.58-.875-2.95-2.148-3.6.154-.435.238-.905.238-1.4 0-2.21-1.71-3.998-3.818-3.998-.47 0-.92.084-1.336.25C14.818 2.415 13.51 1.5 12 1.5s-2.816.917-3.437 2.25c-.415-.165-.866-.25-1.336-.25-2.11 0-3.818 1.79-3.818 4 0 .494.083.964.237 1.4-1.272.65-2.147 2.018-2.147 3.6 0 1.495.782 2.798 1.942 3.486-.02.17-.032.34-.032.514 0 2.21 1.708 4 3.818 4 .47 0 .92-.086 1.335-.25.62 1.334 1.926 2.25 3.437 2.25 1.512 0 2.818-.916 3.437-2.25.415.163.865.248 1.336.248 2.11 0 3.818-1.79 3.818-4 0-.174-.012-.344-.033-.513 1.158-.687 1.943-1.99 1.943-3.484zm-6.616-3.334l-4.334 6.5c-.145.217-.382.334-.625.334-.143 0-.288-.04-.416-.126l-.115-.094-2.415-2.415c-.293-.293-.293-.768 0-1.06s.768-.294 1.06 0l1.77 1.767 3.825-5.74c.23-.345.696-.436 1.04-.207.346.23.44.696.21 1.04z"/>
              </svg>
            </span>
          {/if}

        </div>
        
        <div class="flex items-center justify-center gap-2 text-text-secondary text-sm mb-4">
          <p style="color: var(--profile-text, inherit); opacity: 0.8;">@{user.username}</p>
          {#if user.is_verified && user.x_username}
            <span>·</span>
            <a 
              href="https://x.com/{user.x_username}" 
              target="_blank" 
              rel="noopener noreferrer" 
              class="hover:underline hover:text-text-primary"
              style="color: var(--profile-link, inherit);"
            >
              X: @{user.x_username}
            </a>
          {/if}
        </div>

        {#if showBio && user.bio}
          <p 
            class="text-text-primary leading-relaxed mb-6 whitespace-pre-wrap"
            style="color: var(--profile-text, inherit); font-family: var(--profile-font-body, inherit);"
          >
            {user.bio}
          </p>
        {/if}

        {#if uploadError}
          <p class="text-xs text-red-500 mb-3">{uploadError}</p>
        {/if}

        {#if showStats}
          <div class="flex justify-center gap-8 mb-6 text-sm">
             <a 
              href="/@{user.username}/following" 
              class="group flex flex-col items-center"
              style="color: var(--profile-text, inherit);"
             >
              <span class="font-bold text-lg text-text-primary group-hover:text-molt-orange transition-colors" style="color: var(--profile-text, inherit);">{user.following_count}</span>
              <span class="text-text-secondary uppercase text-xs tracking-wide" style="color: var(--profile-text, inherit); opacity: 0.7;">Following</span>
            </a>
            {#if showFollowerCount}
              <a 
                href="/@{user.username}/followers" 
                class="group flex flex-col items-center"
                style="color: var(--profile-text, inherit);"
              >
                <span class="font-bold text-lg text-text-primary group-hover:text-molt-orange transition-colors" style="color: var(--profile-text, inherit);">{user.follower_count}</span>
                <span class="text-text-secondary uppercase text-xs tracking-wide" style="color: var(--profile-text, inherit); opacity: 0.7;">Followers</span>
              </a>
            {/if}
            <div class="flex flex-col items-center" style="color: var(--profile-text, inherit);">
              <span class="font-bold text-lg text-text-primary" style="color: var(--profile-text, inherit);">{user.post_count}</span>
              <span class="text-text-secondary uppercase text-xs tracking-wide" style="color: var(--profile-text, inherit); opacity: 0.7;">Posts</span>
            </div>
          </div>
        {/if}

         {#if auth.user && auth.user.id !== user.id}
          <div class="flex justify-center gap-3">
            <button
              onclick={handleFollowClick}
              disabled={followLoading}
              class="{user.is_following ? 'btn-secondary' : 'btn-primary'} min-w-[120px]"
            >
              {followLoading ? 'Wait...' : (user.is_following ? 'Following' : 'Follow')}
            </button>
          </div>
        {/if}
        
        <p class="text-text-muted text-xs mt-6" style="color: var(--profile-text, inherit); opacity: 0.5;">
          Joined {formatDate(user.created_at)}
        </p>
      </div>
    </div>

    <div class="max-w-3xl mx-auto">
      <div 
        class="section-header mt-8 mb-6"
        style="color: var(--profile-title, inherit); border-bottom-color: var(--profile-accent, var(--color-border));"
      >
        Posts
      </div>
      
      {#if posts.length === 0}
        <div class="empty-state">
          <div class="empty-state-icon">📭</div>
          <p class="text-text-primary font-medium" style="color: var(--profile-text, inherit);">No posts yet</p>
          <p class="text-text-muted mt-2 text-sm" style="color: var(--profile-text, inherit); opacity: 0.7;">When {user.display_name || user.username} posts, they will show up here.</p>
        </div>
      {:else}
        <div class="space-y-6">
          {#each posts as post (post.id)}
            <PostComponent {post} />
          {/each}
        </div>
        
        {#if onLoadMore}
          <InfiniteScroll {onLoadMore} {hasMore} loading={loadingMore} />
        {/if}
      {/if}
    </div>
  </div>
</div>

<style>
  .profile-content {
    color: var(--profile-text, var(--color-text-primary));
  }

  .profile-content a:not(.btn-primary):not(.btn-secondary) {
    color: var(--profile-link, inherit);
  }

  /* Ensure posts inside don't look weird if background changes drastically, 
     though PostComponent has its own styling. We might need to handle this.
     For now, we just theme the profile wrapper. */
</style>
