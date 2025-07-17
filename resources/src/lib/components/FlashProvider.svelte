<script lang="ts">
  import toast, { Toaster } from "svelte-french-toast";
  import { onMount } from "svelte";
  import CustomToast from "./CustomToast.svelte"; // Import our new component
  import type { FlashMessage } from "$lib/types/index"; // Assuming types are in src/lib/types.ts

  type Props = {
    children: any; // Using Svelte 5's Snippet type
    flash: FlashMessage | null;
  };

  let { children, flash }: Props = $props();

  // We use a separate variable to track the last shown message ID
  // to prevent re-triggering on every render.
  let lastFlashId = "";

  $effect(() => {
    // Only trigger if there is a flash message AND it's a new one.
    if (flash && flash.message !== lastFlashId) {
      lastFlashId = flash.message; // Update the last shown message

      toast.custom(CustomToast, {
        // Pass the flash data as props to the CustomToast component
        props: {
          type: flash.type,
          message: flash.message,
        },
        // Sensible defaults
        duration: 5000, // 5 seconds
        position: "top-right",
      });
    }
  });
</script>

{@render children()}

<Toaster />
