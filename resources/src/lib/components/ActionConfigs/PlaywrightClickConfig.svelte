<script lang="ts">
  import { Label, Input, Select, Checkbox } from "flowbite-svelte";

  type PlaywrightClickConfig = {
    selector: string;
    button?: "left" | "right" | "middle";
    click_count?: number;
    force?: boolean;
  };

  const applyDefaults = (targetConfig: PlaywrightClickConfig) => {
    if (!targetConfig.button) targetConfig.button = "left";
    if (!targetConfig.click_count) targetConfig.click_count = 1;
    if (!targetConfig.force) targetConfig.force = false;
    if (!targetConfig.selector) targetConfig.selector = "";
  };
  let { config = $bindable() }: { config: PlaywrightClickConfig } = $props();
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
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
    <Select
      id="click-button"
      bind:value={config.button}
      items={[
        { value: "", name: "(Default - Left)" },
        { value: "left", name: "Left" },
        { value: "right", name: "Right" },
        { value: "middle", name: "Middle" },
      ]}
    />
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
    <Label for="click-force" class="ml-2"
      >Force click (bypass actionability checks)</Label
    >
  </div>
</div>
