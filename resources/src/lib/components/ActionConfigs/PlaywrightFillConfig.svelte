<script lang="ts">
  import { Label, Input, Checkbox } from "flowbite-svelte";

  type PlaywrightFillConfig = {
    selector: string;
    value: string;
    force?: boolean;
  };

  let { config = $bindable() }: { config: PlaywrightFillConfig } = $props();

  $effect(() => {
    if (!config.selector) config.selector = "";
    if (!config.value) config.value = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="fill-selector" class="mb-2">Selector *</Label>
    <Input
      id="fill-selector"
      type="text"
      bind:value={config.selector}
      placeholder="input[name='email'], #username"
      required
    />
  </div>

  <div>
    <Label for="fill-value" class="mb-2">Value *</Label>
    <Input
      id="fill-value"
      type="text"
      bind:value={config.value}
      placeholder="Text to fill"
      required
    />
  </div>

  <div class="flex items-center">
    <Checkbox id="fill-force" bind:checked={config.force} />
    <Label for="fill-force" class="ml-2">Force fill (bypass actionability checks)</Label>
  </div>
</div>