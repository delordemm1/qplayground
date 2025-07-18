<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightGoBackConfig = {
    timeout?: number;
  };

  let { config = $bindable() }: { config: PlaywrightGoBackConfig } = $props();
</script>

<div class="space-y-4">
  <div>
    <Label for="go-back-timeout" class="mb-2">Timeout (ms)</Label>
    <Input
      id="go-back-timeout"
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