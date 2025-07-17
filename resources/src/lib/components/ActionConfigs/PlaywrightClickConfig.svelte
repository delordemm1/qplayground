<script lang="ts">
  import { Label, Input, Select, Option, Checkbox } from "flowbite-svelte";

  type PlaywrightClickConfig = {
    selector: string;
    button?: "left" | "right" | "middle";
    click_count?: number;
    force?: boolean;
  };

  let { config = $bindable() }: { config: PlaywrightClickConfig } = $props();

  $effect(() => {
    if (!config.selector) config.selector = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="click-selector" class="mb-2">Selector *</Label>
    <Input
      id="click-selector"
      type="text"
      bind:value={config.selector}
      placeholder="button, #submit, .btn-primary"
      required
    />
  </div>

  <div>
    <Label for="click-button" class="mb-2">Mouse Button</Label>
    <Select id="click-button" bind:value={config.button}>
      <Option value="">(Default - Left)</Option>
      <Option value="left">Left</Option>
      <Option value="right">Right</Option>
      <Option value="middle">Middle</Option>
    </Select>
  </div>

  <div>
    <Label for="click-count" class="mb-2">Click Count</Label>
    <Input
      id="click-count"
      type="number"
      bind:value={config.click_count}
      placeholder="1"
      min={1}
      max={10}
    />
  </div>

  <div class="flex items-center">
    <Checkbox id="click-force" bind:checked={config.force} />
    <Label for="click-force" class="ml-2">Force click (bypass actionability checks)</Label>
  </div>
</div>