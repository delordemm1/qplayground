<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightTypeConfig = {
    selector: string;
    text: string;
    delay?: number;
  };

  let { config = $bindable() }: { config: PlaywrightTypeConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightTypeConfig) {
    if (!targetConfig.selector) targetConfig.selector = "";
    if (!targetConfig.text) targetConfig.text = "";
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="type-selector" class="mb-2">Selector *</Label>
    <Input
      id="type-selector"
      type="text"
      bind:value={config.selector}
      placeholder="input[type='text'], textarea"
      required
    />
  </div>

  <div>
    <Label for="type-text" class="mb-2">Text *</Label>
    <Input
      id="type-text"
      type="text"
      bind:value={config.text}
      placeholder="Text to type"
      required
    />
  </div>

  <div>
    <Label for="type-delay" class="mb-2">Delay between keystrokes (ms)</Label>
    <Input
      id="type-delay"
      type="number"
      bind:value={config.delay}
      placeholder="100"
      min={0}
    />
  </div>
</div>