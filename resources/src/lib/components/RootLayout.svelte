<script lang="ts">
  import { showErrorToast, showInfoToast, showSuccessToast, showWarningToast } from "$lib/utils/toast";
  import toast, { Toaster } from "svelte-french-toast";

  let { children, flash } = $props();
  $effect(() => {
    if (flash) {
      switch (flash.type) {
        case "success":
          showSuccessToast(flash.message);
          break;
        case "error":
          showErrorToast(flash.message);
          break;
        case "info":
          showInfoToast(flash.message);
          break;
        case "warning":
          showWarningToast(flash.message);
          break;
        default:
          showInfoToast(flash.message);
          break;
      }
    }
  });
</script>

{@render children()}

<Toaster />
