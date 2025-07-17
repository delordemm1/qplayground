<script lang="ts">
  import { Label, Input, Select, Option } from "flowbite-svelte";

  type PlaywrightGotoConfig = {
    url: string;
    timeout?: number;
    wait_until?: "load" | "domcontentloaded" | "networkidle";
  };

  let { config = $bindable() }: { config: PlaywrightGotoConfig } = $props();

  // Initialize config if it's empty or missing properties
  $effect(() => {
    if (!config.url) config.url = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="goto-url" class="mb-2">URL *</Label>
    <Input
      id="goto-url"
      type="url"
      bind:value={config.url}
      placeholder="https://example.com"
      required
    />
  </div>

  <div>
    <Label for="goto-timeout" class="mb-2">Timeout (ms)</Label>
    <Input
      id="goto-timeout"
      type="number"
      bind:value={config.timeout}
      placeholder="30000"
      min={0}
    />
  </div>

  <div>
    <Label for="goto-wait-until" class="mb-2">Wait Until</Label>
    <Select id="goto-wait-until" bind:value={config.wait_until}>
      <Option value="">(Default)</Option>
      <Option value="load">Load</Option>
      <Option value="domcontentloaded">DOM Content Loaded</Option>
      <Option value="networkidle">Network Idle</Option>
    </Select>
  </div>
</div>