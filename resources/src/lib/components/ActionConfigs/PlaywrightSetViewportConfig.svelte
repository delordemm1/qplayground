<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightSetViewportConfig = {
    width: number;
    height: number;
  };

  let { config = $bindable() }: { config: PlaywrightSetViewportConfig } = $props();

  $effect(() => {
    if (!config.width) config.width = 1920;
    if (!config.height) config.height = 1080;
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="viewport-width" class="mb-2">Width *</Label>
    <Input
      id="viewport-width"
      type="number"
      bind:value={config.width}
      placeholder="1920"
      min={1}
      required
    />
  </div>

  <div>
    <Label for="viewport-height" class="mb-2">Height *</Label>
    <Input
      id="viewport-height"
      type="number"
      bind:value={config.height}
      placeholder="1080"
      min={1}
      required
    />
  </div>

  <p class="text-xs text-gray-500">
    Common sizes: 1920x1080 (Desktop), 1366x768 (Laptop), 375x667 (Mobile)
  </p>
</div>