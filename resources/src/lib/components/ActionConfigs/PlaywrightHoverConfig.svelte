<script lang="ts">
  import { Label, Input, Checkbox } from "flowbite-svelte";

  type PlaywrightHoverConfig = {
    selector: string;
    force?: boolean;
  };

  let { config = $bindable() }: { config: PlaywrightHoverConfig } = $props();

  $effect(() => {
    if (!config.selector) config.selector = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="hover-selector" class="mb-2">Selector *</Label>
    <Input
      id="hover-selector"
      type="text"
      bind:value={config.selector}
      placeholder=".menu-item, #dropdown-trigger"
      required
    />
  </div>

  <div class="flex items-center">
    <Checkbox id="hover-force" bind:checked={config.force} />
    <Label for="hover-force" class="ml-2">Force hover (bypass actionability checks)</Label>
  </div>
</div>