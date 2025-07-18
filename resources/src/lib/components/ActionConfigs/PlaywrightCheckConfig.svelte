<script lang="ts">
  import { Label, Input, Checkbox } from "flowbite-svelte";

  type PlaywrightCheckConfig = {
    selector: string;
    force?: boolean;
  };
  function applyDefaults(targetConfig: PlaywrightCheckConfig) {
    if (!targetConfig.force) targetConfig.force = false;
    if (!targetConfig.selector) targetConfig.selector = "";
  }
  let { config = $bindable() }: { config: PlaywrightCheckConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="check-selector" class="mb-2">Selector *</Label>
    <Input
      id="check-selector"
      type="text"
      bind:value={config.selector}
      placeholder="input[type='checkbox'], #agree-terms"
      required
    />
  </div>

  <div class="flex items-center">
    <Checkbox id="check-force" bind:checked={config.force} />
    <Label for="check-force" class="ml-2"
      >Force check (bypass actionability checks)</Label
    >
  </div>
</div>
