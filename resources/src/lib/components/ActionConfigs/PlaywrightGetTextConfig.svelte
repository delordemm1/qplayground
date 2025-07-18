<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightGetTextConfig = {
    selector: string;
  };

  let { config = $bindable() }: { config: PlaywrightGetTextConfig } = $props();

  $effect(() => {
    if (!config.selector) config.selector = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="get-text-selector" class="mb-2">Selector *</Label>
    <Input
      id="get-text-selector"
      type="text"
      bind:value={config.selector}
      placeholder="h1, .title, #content"
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      The text content will be logged and available in the run results
    </p>
  </div>
</div>