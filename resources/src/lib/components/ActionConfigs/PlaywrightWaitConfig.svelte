<script lang="ts">
  import { Label, Input, Select } from "flowbite-svelte";

  type PlaywrightWaitConfig = {
    selector?: string;
    timeout?: number;
    state?: "attached" | "detached" | "visible" | "hidden";
    timeout_ms?: number; // For wait_for_timeout action
  };

  let {
    config = $bindable(),
    actionType,
  }: { config: PlaywrightWaitConfig; actionType: string } = $props();

  $effect(() => {
    if (actionType === "playwright:wait_for_selector" && !config.selector) {
      config.selector = "";
    }
  });
</script>

<div class="space-y-4">
  {#if actionType === "playwright:wait_for_selector"}
    <div>
      <Label for="wait-selector" class="mb-2">Selector *</Label>
      <Input
        id="wait-selector"
        type="text"
        bind:value={config.selector}
        placeholder=".loading, #content"
        required
      />
    </div>

    <div>
      <Label for="wait-state" class="mb-2">State</Label>
      <Select
        id="wait-state"
        bind:value={config.state}
        items={[
          { value: "", name: "(Default - Visible)" },
          { value: "attached", name: "Attached" },
          { value: "detached", name: "Detached" },
          { value: "visible", name: "Visible" },
          { value: "hidden", name: "Hidden" },
        ]}
      />
    </div>

    <div>
      <Label for="wait-timeout" class="mb-2">Timeout (ms)</Label>
      <Input
        id="wait-timeout"
        type="number"
        bind:value={config.timeout}
        placeholder="30000"
        min={0}
      />
    </div>
  {:else if actionType === "playwright:wait_for_timeout"}
    <div>
      <Label for="wait-timeout-ms" class="mb-2">Timeout (ms) *</Label>
      <Input
        id="wait-timeout-ms"
        type="number"
        bind:value={config.timeout_ms}
        placeholder="5000"
        min={0}
        required
      />
    </div>
  {:else if actionType === "playwright:wait_for_load_state"}
    <div>
      <Label for="wait-load-state" class="mb-2">Load State</Label>
      <Select
        id="wait-load-state"
        bind:value={config.state}
        items={[
          { value: "", name: "(Default - Load)" },
          { value: "load", name: "Load" },
          { value: "domcontentloaded", name: "DOM Content Loaded" },
          { value: "networkidle", name: "Network Idle" },
        ]}
      />
    </div>

    <div>
      <Label for="wait-load-timeout" class="mb-2">Timeout (ms)</Label>
      <Input
        id="wait-load-timeout"
        type="number"
        bind:value={config.timeout}
        placeholder="30000"
        min={0}
      />
    </div>
  {/if}
</div>
