<script lang="ts">
  import { Label, Input } from "flowbite-svelte";

  type R2UploadConfig = {
    key: string;
    content: string;
    content_type?: string;
  };

  let { config = $bindable() }: { config: R2UploadConfig } = $props();

  $effect(() => {
    if (!config.key) config.key = "";
    if (!config.content) config.content = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="r2-key" class="mb-2">Object Key *</Label>
    <Input
      id="r2-key"
      type="text"
      bind:value={config.key}
      placeholder="files/document.txt"
      required
    />
  </div>

  <div>
    <Label for="r2-content" class="mb-2">Content *</Label>
    <Input
      id="r2-content"
      type="text"
      bind:value={config.content}
      placeholder="File content or text"
      required
    />
  </div>

  <div>
    <Label for="r2-content-type" class="mb-2">Content Type</Label>
    <Input
      id="r2-content-type"
      type="text"
      bind:value={config.content_type}
      placeholder="text/plain, image/png, application/json"
    />
    <p class="text-xs text-gray-500 mt-1">
      Leave empty for auto-detection based on file extension
    </p>
  </div>
</div>