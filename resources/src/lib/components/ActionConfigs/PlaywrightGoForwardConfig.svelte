<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightGoForwardConfig = {
    timeout?: number;
  };

  let { config = $bindable() }: { config: PlaywrightGoForwardConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightGoForwardConfig) {
    // No required defaults for go forward config, all fields are optional
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="go-forward-timeout" class="mb-2">Timeout (ms)</Label>
    <Input
      id="go-forward-timeout"
      type="number"
      bind:value={config.timeout}
      placeholder="30000"
      min={0}
    />
    <p class="text-xs text-gray-500 mt-1">
      Maximum time to wait for navigation to complete (default: 30 seconds)
    </p>
  </div>
</div>