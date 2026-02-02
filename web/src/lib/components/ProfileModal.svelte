<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import type { Snippet } from 'svelte';

  let { children, onclose, pageBackground }: { 
    children: Snippet, 
    onclose: () => void,
    pageBackground?: string 
  } = $props();
  
  const backdropStyle = $derived(pageBackground ? `background-color: ${pageBackground}` : '');

  let modalElement: HTMLDivElement | undefined = $state();

  $effect(() => {
    // Lock body scroll
    const originalOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    // Focus the modal for accessibility
    modalElement?.focus();

    const handleKeydown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onclose();
      }
    };

    window.addEventListener('keydown', handleKeydown);

    return () => {
      document.body.style.overflow = originalOverflow;
      window.removeEventListener('keydown', handleKeydown);
    };
  });

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      onclose();
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
class="fixed inset-0 z-50 flex justify-center backdrop-blur-sm {pageBackground ? '' : 'bg-black/60'}"
style={backdropStyle}
onclick={handleBackdropClick}
transition:fade={{ duration: 200 }}
>
<div
bind:this={modalElement}
class="relative w-full max-w-[950px] h-full flex focus:outline-none"
transition:fly={{ y: 20, duration: 250, easing: cubicOut }}
tabindex="-1"
role="dialog"
aria-modal="true"
onclick={(e) => e.stopPropagation()}
>
<div class="absolute right-full mr-4 top-6 hidden md:flex flex-col gap-3 z-50">
<button
onclick={onclose}
class="w-10 h-10 flex items-center justify-center rounded-full bg-[var(--color-surface-50)] text-[var(--color-text-secondary)] shadow-lg hover:text-[var(--color-molt-orange)] transition-colors hover:scale-105 active:scale-95 cursor-pointer border border-[var(--color-surface-300)]"
aria-label="Close"
title="Close (Esc)"
>
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
<path d="M18 6 6 18"/><path d="M6 6l12 12"/>
</svg>
</button>

<button
onclick={onclose}
class="w-10 h-10 flex items-center justify-center rounded-full bg-[var(--color-surface-50)] text-[var(--color-text-secondary)] shadow-lg hover:text-[var(--color-molt-orange)] transition-colors hover:scale-105 active:scale-95 cursor-pointer border border-[var(--color-surface-300)]"
aria-label="Go back"
title="Go back"
>
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
<path d="m12 19-7-7 7-7"/><path d="M19 12H5"/>
</svg>
</button>
</div>

<button
onclick={onclose}
class="md:hidden absolute top-4 left-4 z-50 w-10 h-10 flex items-center justify-center rounded-full bg-[var(--color-surface-50)] text-[var(--color-text-secondary)] shadow-sm border border-[var(--color-surface-300)]"
aria-label="Close"
>
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
<path d="M18 6 6 18"/><path d="M6 6l12 12"/>
</svg>
</button>

<div class="h-full w-full bg-[var(--color-surface-50)] shadow-2xl overflow-hidden flex flex-col border-x border-[var(--color-surface-300)]">
<div class="h-full overflow-y-auto custom-scrollbar relative">
{@render children()}
</div>
</div>
</div>
</div>

<style>
    .custom-scrollbar::-webkit-scrollbar {
        width: 8px;
    }
    .custom-scrollbar::-webkit-scrollbar-track {
        background: transparent;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb {
        background: var(--color-surface-300);
        border-radius: 4px;
    }
    .custom-scrollbar::-webkit-scrollbar-thumb:hover {
        background: var(--color-surface-400);
    }
</style>
