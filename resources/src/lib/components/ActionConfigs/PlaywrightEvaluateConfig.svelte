<script lang="ts">
  import { Label, Textarea } from "flowbite-svelte";

  type PlaywrightEvaluateConfig = {
    expression: string;
  };

  let { config = $bindable() }: { config: PlaywrightEvaluateConfig } = $props();

  $effect(() => {
    if (!config.expression) config.expression = "";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="evaluate-expression" class="mb-2">JavaScript Expression *</Label>
    <Textarea
      id="evaluate-expression"
      rows="6"
      bind:value={config.expression}
      placeholder="console.log('Hello from browser'); return document.title;"
      class="font-mono text-sm"
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      JavaScript code to execute in the browser context
    </p>
  </div>
</div>