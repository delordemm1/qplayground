<script lang="ts">
  import { Modal, Button } from "flowbite-svelte";
  import { ChevronLeftOutline, ChevronRightOutline, DownloadOutline, XmarkOutline } from "flowbite-svelte-icons";

  type Props = {
    open: boolean;
    imageUrls: string[];
    startIndex?: number;
    onClose: () => void;
  };

  let { open = $bindable(), imageUrls, startIndex = 0, onClose }: Props = $props();

  let currentIndex = $state(startIndex);

  // Reset current index when modal opens or startIndex changes
  $effect(() => {
    if (open) {
      currentIndex = startIndex;
    }
  });

  // Navigation functions
  function goToPrevious() {
    if (currentIndex > 0) {
      currentIndex--;
    }
  }

  function goToNext() {
    if (currentIndex < imageUrls.length - 1) {
      currentIndex++;
    }
  }

  // Download function
  function downloadCurrentImage() {
    if (imageUrls[currentIndex]) {
      const link = document.createElement('a');
      link.href = imageUrls[currentIndex];
      link.download = getImageFileName(imageUrls[currentIndex]);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    }
  }

  // Helper to extract filename from URL
  function getImageFileName(url: string): string {
    const parts = url.split('/');
    return parts[parts.length - 1] || 'image.png';
  }

  // Keyboard navigation
  function handleKeydown(event: KeyboardEvent) {
    if (!open) return;
    
    switch (event.key) {
      case 'ArrowLeft':
        event.preventDefault();
        goToPrevious();
        break;
      case 'ArrowRight':
        event.preventDefault();
        goToNext();
        break;
      case 'Escape':
        event.preventDefault();
        onClose();
        break;
    }
  }

  // Current image info
  const currentImageUrl = $derived(imageUrls[currentIndex]);
  const hasMultipleImages = $derived(imageUrls.length > 1);
  const canGoBack = $derived(currentIndex > 0);
  const canGoForward = $derived(currentIndex < imageUrls.length - 1);
</script>

<svelte:window onkeydown={handleKeydown} />

<Modal bind:open outsideclose={true} class="w-full max-w-6xl" size="xl">
  <div class="relative">
    <!-- Header -->
    <div class="flex items-center justify-between p-4 border-b">
      <h3 class="text-lg font-semibold text-gray-900">
        Image Viewer {hasMultipleImages ? `(${currentIndex + 1} of ${imageUrls.length})` : ''}
      </h3>
      <div class="flex items-center space-x-2">
        <Button size="sm" color="primary" onclick={downloadCurrentImage}>
          <DownloadOutline class="w-4 h-4 mr-2" />
          Download
        </Button>
        <Button size="sm" color="alternative" onclick={onClose}>
          <XmarkOutline class="w-4 h-4" />
        </Button>
      </div>
    </div>

    <!-- Image Display -->
    <div class="relative bg-gray-100 flex items-center justify-center min-h-96 max-h-[70vh]">
      {#if currentImageUrl}
        <img
          src={currentImageUrl}
          alt="Automation output {currentIndex + 1}"
          class="max-w-full max-h-full object-contain"
          loading="lazy"
        />
      {:else}
        <div class="text-center text-gray-500">
          <p>No image to display</p>
        </div>
      {/if}

      <!-- Navigation Arrows (only show if multiple images) -->
      {#if hasMultipleImages}
        <!-- Previous Button -->
        <button
          onclick={goToPrevious}
          disabled={!canGoBack}
          class="absolute left-4 top-1/2 transform -translate-y-1/2 bg-black bg-opacity-50 hover:bg-opacity-70 text-white p-2 rounded-full disabled:opacity-30 disabled:cursor-not-allowed transition-all"
          title="Previous image (←)"
        >
          <ChevronLeftOutline class="w-6 h-6" />
        </button>

        <!-- Next Button -->
        <button
          onclick={goToNext}
          disabled={!canGoForward}
          class="absolute right-4 top-1/2 transform -translate-y-1/2 bg-black bg-opacity-50 hover:bg-opacity-70 text-white p-2 rounded-full disabled:opacity-30 disabled:cursor-not-allowed transition-all"
          title="Next image (→)"
        >
          <ChevronRightOutline class="w-6 h-6" />
        </button>
      {/if}
    </div>

    <!-- Footer with thumbnails (only show if multiple images) -->
    {#if hasMultipleImages}
      <div class="p-4 border-t bg-gray-50">
        <div class="flex space-x-2 overflow-x-auto">
          {#each imageUrls as imageUrl, index (index)}
            <button
              onclick={() => currentIndex = index}
              class="flex-shrink-0 w-16 h-16 border-2 rounded-md overflow-hidden transition-all {currentIndex === index ? 'border-primary-500' : 'border-gray-300 hover:border-gray-400'}"
              title="Go to image {index + 1}"
            >
              <img
                src={imageUrl}
                alt="Thumbnail {index + 1}"
                class="w-full h-full object-cover"
                loading="lazy"
              />
            </button>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Image Info -->
    <div class="p-4 bg-gray-50 text-sm text-gray-600">
      <p class="truncate">
        <span class="font-medium">File:</span> {getImageFileName(currentImageUrl)}
      </p>
      <p class="mt-1">
        <span class="font-medium">URL:</span> 
        <a href={currentImageUrl} target="_blank" rel="noopener noreferrer" class="text-primary-600 hover:underline truncate">
          {currentImageUrl}
        </a>
      </p>
    </div>
  </div>
</Modal>

<style>
  /* Custom scrollbar for thumbnail strip */
  .overflow-x-auto::-webkit-scrollbar {
    height: 6px;
  }
  
  .overflow-x-auto::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 3px;
  }
  
  .overflow-x-auto::-webkit-scrollbar-thumb {
    background: #c1c1c1;
    border-radius: 3px;
  }
  
  .overflow-x-auto::-webkit-scrollbar-thumb:hover {
    background: #a8a8a8;
  }
</style>