<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightPressConfig = {
    selector: string;
    key: string;
    delay?: number;
  };

  let { config = $bindable() }: { config: PlaywrightPressConfig } = $props();

  $effect(() => {
    if (!config.selector) config.selector = "";
    if (!config.key) config.key = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="press-selector" class="mb-2">Selector *</Label>
    <Input
      id="press-selector"
      type="text"
      bind:value={config.selector}
      placeholder="input, button, #element"
      required
    />
  </div>

  <div>
    <Label for="press-key" class="mb-2">Key *</Label>
    <Input
      id="press-key"
      type="text"
      bind:value={config.key}
      placeholder="Enter, Tab, Escape, ArrowDown"
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      Key name (e.g., Enter, Tab, Escape, ArrowDown, a, A, 1, etc.)
    </p>
  </div>

  <div>
    <Label for="press-delay" class="mb-2">Delay (ms)</Label>
    <Input
      id="press-delay"
      type="number"
      bind:value={config.delay}
      placeholder="100"
      min={0}
    />
    <p class="text-xs text-gray-500 mt-1">
      Time to wait between key down and key up
    </p>
  </div>
</div>