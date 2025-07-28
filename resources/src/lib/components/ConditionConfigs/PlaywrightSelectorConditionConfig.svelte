<script lang="ts">
  import { Label, Input, Select } from "flowbite-svelte";

  type PlaywrightSelectorConditionConfig = {
    selector: string;
    state?: "attached" | "detached" | "visible" | "hidden";
    timeout?: number;
  };

  let { config = $bindable() }: { config: PlaywrightSelectorConditionConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightSelectorConditionConfig) {
    if (!targetConfig.selector) targetConfig.selector = "";
    if (!targetConfig.state) targetConfig.state = "visible";
    if (!targetConfig.timeout) targetConfig.timeout = 5000;
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="condition-selector" class="mb-2">Selector *</Label>
    <Input
      id="condition-selector"
      type="text"
      bind:value={config.selector}
      placeholder="#submit-button, .loading-spinner"
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      CSS selector to check for the condition
    </p>
  </div>

  <div>
    <Label for="condition-state" class="mb-2">State</Label>
    <Select
      id="condition-state"
      bind:value={config.state}
      items={[
        { value: "visible", name: "Visible" },
        { value: "hidden", name: "Hidden" },
        { value: "attached", name: "Attached" },
        { value: "detached", name: "Detached" },
      ]}
    />
    <p class="text-xs text-gray-500 mt-1">
      What state to check for the selector
    </p>
  </div>

  <div>
    <Label for="condition-timeout" class="mb-2">Timeout (ms)</Label>
    <Input
      id="condition-timeout"
      type="number"
      bind:value={config.timeout}
      placeholder="5000"
      min={100}
    />
    <p class="text-xs text-gray-500 mt-1">
      Maximum time to wait for the condition (default: 5 seconds)
    </p>
  </div>
</div>