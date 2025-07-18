<script lang="ts">
  import { Label, Input, Checkbox } from "flowbite-svelte";

  type PlaywrightUncheckConfig = {
    selector: string;
    force?: boolean;
  };

  let { config = $bindable() }: { config: PlaywrightUncheckConfig } = $props();

  $effect(() => {
    if (!config.selector) config.selector = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="uncheck-selector" class="mb-2">Selector *</Label>
    <Input
      id="uncheck-selector"
      type="text"
      bind:value={config.selector}
      placeholder="input[type='checkbox'], #newsletter-opt-in"
      required
    />
  </div>

  <div class="flex items-center">
    <Checkbox id="uncheck-force" bind:checked={config.force} />
    <Label for="uncheck-force" class="ml-2">Force uncheck (bypass actionability checks)</Label>
  </div>
</div>