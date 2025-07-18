<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type PlaywrightScrollConfig = {
    delta_x?: number;
    delta_y?: number;
  };

  let { config = $bindable() }: { config: PlaywrightScrollConfig } = $props();

  $effect(() => {
    if (config.delta_x === undefined) config.delta_x = 0;
    if (config.delta_y === undefined) config.delta_y = 1000;
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="scroll-delta-x" class="mb-2">Horizontal Scroll (Delta X)</Label>
    <Input
      id="scroll-delta-x"
      type="number"
      bind:value={config.delta_x}
      placeholder="0"
    />
    <p class="text-xs text-gray-500 mt-1">
      Positive values scroll right, negative values scroll left
    </p>
  </div>

  <div>
    <Label for="scroll-delta-y" class="mb-2">Vertical Scroll (Delta Y)</Label>
    <Input
      id="scroll-delta-y"
      type="number"
      bind:value={config.delta_y}
      placeholder="1000"
    />
    <p class="text-xs text-gray-500 mt-1">
      Positive values scroll down, negative values scroll up
    </p>
  </div>
</div>