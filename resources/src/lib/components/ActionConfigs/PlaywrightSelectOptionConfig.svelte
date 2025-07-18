<script lang="ts">
  import { Label, Input, Select, Radio } from "flowbite-svelte";

  type PlaywrightSelectOptionConfig = {
    selector: string;
    selection_type?: "value" | "label" | "index";
    value?: string;
    values?: string[];
    label?: string;
    index?: number;
  };

  let { config = $bindable() }: { config: PlaywrightSelectOptionConfig } = $props();
  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightSelectOptionConfig) {
    if (!targetConfig.selector) targetConfig.selector = "";
    if (!targetConfig.selection_type) targetConfig.selection_type = "value";
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  // Helper to manage multiple values as a comma-separated string
  let valuesString = $state("");
  
  $effect(() => {
    if (config.values && Array.isArray(config.values)) {
      valuesString = config.values.join(", ");
    }
  });

  $effect(() => {
    if (config.selection_type === "value" && valuesString) {
      config.values = valuesString.split(",").map(v => v.trim()).filter(v => v);
    }
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="select-selector" class="mb-2">Selector *</Label>
    <Input
      id="select-selector"
      type="text"
      bind:value={config.selector}
      placeholder="select, #country-dropdown"
      required
    />
  </div>

  <div>
    <Label class="mb-2">Selection Method</Label>
    <div class="space-y-2">
      <div class="flex items-center">
        <Radio
          name="selection-type"
          value="value"
          bind:group={config.selection_type}
        />
        <Label class="ml-2">By Value</Label>
      </div>
      <div class="flex items-center">
        <Radio
          name="selection-type"
          value="label"
          bind:group={config.selection_type}
        />
        <Label class="ml-2">By Label (visible text)</Label>
      </div>
      <div class="flex items-center">
        <Radio
          name="selection-type"
          value="index"
          bind:group={config.selection_type}
        />
        <Label class="ml-2">By Index (position)</Label>
      </div>
    </div>
  </div>

  {#if config.selection_type === "value"}
    <div>
      <Label for="select-value" class="mb-2">Value *</Label>
      <Input
        id="select-value"
        type="text"
        bind:value={config.value}
        placeholder="option-value"
        required
      />
      <p class="text-xs text-gray-500 mt-1">
        The value attribute of the option to select
      </p>
    </div>
    
    <div>
      <Label for="select-values" class="mb-2">Multiple Values (optional)</Label>
      <Input
        id="select-values"
        type="text"
        bind:value={valuesString}
        placeholder="value1, value2, value3"
      />
      <p class="text-xs text-gray-500 mt-1">
        Comma-separated values for multi-select dropdowns
      </p>
    </div>
  {:else if config.selection_type === "label"}
    <div>
      <Label for="select-label" class="mb-2">Label *</Label>
      <Input
        id="select-label"
        type="text"
        bind:value={config.label}
        placeholder="Option Text"
        required
      />
      <p class="text-xs text-gray-500 mt-1">
        The visible text of the option to select
      </p>
    </div>
  {:else if config.selection_type === "index"}
    <div>
      <Label for="select-index" class="mb-2">Index *</Label>
      <Input
        id="select-index"
        type="number"
        bind:value={config.index}
        placeholder="0"
        min={0}
        required
      />
      <p class="text-xs text-gray-500 mt-1">
        Zero-based index of the option to select (0 = first option)
      </p>
    </div>
  {/if}
</div>