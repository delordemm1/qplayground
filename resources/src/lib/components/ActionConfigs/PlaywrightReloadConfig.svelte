<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightReloadConfig = {
    timeout?: number;
  };

  let { config = $bindable() }: { config: PlaywrightReloadConfig } = $props();
</script>

<div class="space-y-4">
  <div>
    <Label for="reload-timeout" class="mb-2">Timeout (ms)</Label>
    <Input
      id="reload-timeout"
      type="number"
      bind:value={config.timeout}
      placeholder="30000"
      min={0}
    />
    <p class="text-xs text-gray-500 mt-1">
      Maximum time to wait for the page to reload (default: 30 seconds)
    </p>
  </div>
</div>