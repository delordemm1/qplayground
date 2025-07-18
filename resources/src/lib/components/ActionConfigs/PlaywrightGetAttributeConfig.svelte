<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightGetAttributeConfig = {
    selector: string;
    attribute: string;
  };

  let { config = $bindable() }: { config: PlaywrightGetAttributeConfig } = $props();

  $effect(() => {
    if (!config.selector) config.selector = "";
    if (!config.attribute) config.attribute = "";
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