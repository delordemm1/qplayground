<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type R2DeleteConfig = {
    key: string;
  };

  let { config = $bindable() }: { config: R2DeleteConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: R2DeleteConfig) {
    if (!targetConfig.key) targetConfig.key = "";
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="r2-delete-key" class="mb-2">Object Key *</Label>
    <Input
      id="r2-delete-key"
      type="text"
      bind:value={config.key}
      placeholder="files/document.txt"
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      The key of the object to delete from R2 storage
    </p>
  </div>
</div>