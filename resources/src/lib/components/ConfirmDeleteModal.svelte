<script lang="ts">
  import { Modal, Button } from "flowbite-svelte";

  type Props = {
    open: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
    onCancel: () => void;
    confirmText?: string;
    cancelText?: string;
    loading?: boolean;
  };

  let {
    open = $bindable(),
    title,
    message,
    onConfirm,
    onCancel,
    confirmText = "Delete",
    cancelText = "Cancel",
    loading = false,
  }: Props = $props();

  function handleConfirm() {
    onConfirm();
  }

  function handleCancel() {
    onCancel();
  }
</script>

<Modal bind:open outsideclose={false} class="w-full max-w-md">
  <div class="p-6 text-center">
    <svg
      class="mx-auto mb-4 text-gray-400 w-12 h-12 dark:text-gray-200"
      aria-hidden="true"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 20 20"
    >
      <path
        stroke="currentColor"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
        d="M10 11V6m0 8h.01M19 10a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
      />
    </svg>
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
      {title}
    </h3>
    <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">
      {message}
    </p>
    <Button color="red" class="me-2" onclick={handleConfirm} disabled={loading}>
      {#if loading}
        <svg
          class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
        Deleting...
      {:else}
        {confirmText}
      {/if}
    </Button>
    <Button color="alternative" onclick={handleCancel} disabled={loading}>
      {cancelText}
    </Button>
  </div>
</Modal>
