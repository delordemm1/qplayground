<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightGetAttributeConfig = {
    selector: string;
    attribute: string;
  };

  let { config = $bindable() }: { config: PlaywrightGetAttributeConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightGetAttributeConfig) {
    if (!targetConfig.selector) targetConfig.selector = "";
    if (!targetConfig.attribute) targetConfig.attribute = "";
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="get-attr-selector" class="mb-2">Selector *</Label>
    <Input
      id="get-attr-selector"
      type="text"
      bind:value={config.selector}
      placeholder="a, img, input"
      required
    />
  </div>

  <div>
    <Label for="get-attr-attribute" class="mb-2">Attribute Name *</Label>
    <Input
      id="get-attr-attribute"
      type="text"
      bind:value={config.attribute}
      placeholder="href, src, value, class"
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      The attribute value will be logged and available in the run results
    </p>
  </div>
</div>