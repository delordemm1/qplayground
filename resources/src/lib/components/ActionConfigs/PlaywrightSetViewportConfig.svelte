<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightSetViewportConfig = {
    width: number;
    height: number;
  };

  let { config = $bindable() }: { config: PlaywrightSetViewportConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightSetViewportConfig) {
    if (!targetConfig.width) targetConfig.width = 1920;
    if (!targetConfig.height) targetConfig.height = 1080;
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
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