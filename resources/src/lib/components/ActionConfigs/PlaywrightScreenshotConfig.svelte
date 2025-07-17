<script lang="ts">
  import { Label, Input, Select, Checkbox } from "flowbite-svelte";

  type PlaywrightScreenshotConfig = {
    full_page?: boolean;
    format?: "png" | "jpeg";
    quality?: number;
    upload_to_r2?: boolean;
    r2_key?: string;
  };

  let { config = $bindable() }: { config: PlaywrightScreenshotConfig } = $props();

  $effect(() => {
    if (config.full_page === undefined) config.full_page = true;
    if (!config.format) config.format = "png";
  });
</script>

<div class="space-y-4">
  <div class="flex items-center">
    <Checkbox id="screenshot-full-page" bind:checked={config.full_page} />
    <Label for="screenshot-full-page" class="ml-2">Full page screenshot</Label>
  </div>

  <div>
    <Label for="screenshot-format" class="mb-2">Format</Label>
    <Select id="screenshot-format" bind:value={config.format} items={[{ value: "png", name: "PNG" }, { value: "jpeg", name: "JPEG" }]} />
  </div>

  {#if config.format === "jpeg"}
    <div>
      <Label for="screenshot-quality" class="mb-2">Quality (1-100)</Label>
      <Input
        id="screenshot-quality"
        type="number"
        bind:value={config.quality}
        placeholder="80"
        min={1}
        max={100}
      />
    </div>
  {/if}

  <div class="flex items-center">
    <Checkbox id="screenshot-upload-r2" bind:checked={config.upload_to_r2} />
    <Label for="screenshot-upload-r2" class="ml-2">Upload to R2 storage</Label>
  </div>

  {#if config.upload_to_r2}
    <div>
      <Label for="screenshot-r2-key" class="mb-2">R2 Key *</Label>
      <Input
        id="screenshot-r2-key"
        type="text"
        bind:value={config.r2_key}
        placeholder="screenshots/test-screenshot.png"
        required={config.upload_to_r2}
      />
    </div>
  {/if}
</div>