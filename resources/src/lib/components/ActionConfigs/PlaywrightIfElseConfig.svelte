<script lang="ts">
  import { Label, Input, Select, Button, Checkbox } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import { nestedActionTypes } from "$lib/utils/actionConfigMap";
  import NestedActionConfigurator from "../NestedActionConfigurator.svelte";

  type NestedAction = {
    action_type: string;
    action_config: Record<string, any>;
  };

  type PlaywrightIfElseConfig = {
    selector?: string;
    condition_type?: string;
    timeout?: number;
    if_actions: NestedAction[];
    else_actions: NestedAction[];
  };

  let { config = $bindable() }: { config: PlaywrightIfElseConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightIfElseConfig) {
    if (!targetConfig.if_actions) targetConfig.if_actions = [];
    if (!targetConfig.else_actions) targetConfig.else_actions = [];
    if (!targetConfig.selector) targetConfig.selector = "";
    if (!targetConfig.condition_type) targetConfig.condition_type = "is_visible";
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  const conditionTypes = [
    { value: "is_visible", name: "Is Visible" },
    { value: "is_hidden", name: "Is Hidden" },
    { value: "is_enabled", name: "Is Enabled" },
    { value: "is_disabled", name: "Is Disabled" },
    { value: "is_checked", name: "Is Checked" },
    { value: "is_editable", name: "Is Editable" },
    { value: "loop_index_is_even", name: "Loop Index is Even" },
    { value: "loop_index_is_odd", name: "Loop Index is Odd" },
    { value: "loop_index_is_prime", name: "Loop Index is Prime" },
    { value: "random", name: "Random (50% chance)" },
  ];

  // Check if selector is required for the current condition type
  const selectorRequired = $derived(!['loop_index_is_even', 'loop_index_is_odd', 'loop_index_is_prime', 'random'].includes(config.condition_type || ''));

  // Helper functions for managing if/else actions
  function addIfAction() {
    config.if_actions = [...config.if_actions, { action_type: "", action_config: {} }];
  }

  function removeIfAction(index: number) {
    config.if_actions = config.if_actions.filter((_, i) => i !== index);
  }

  function addElseAction() {
    config.else_actions = [...config.else_actions, { action_type: "", action_config: {} }];
  }

  function removeElseAction(index: number) {
    config.else_actions = config.else_actions.filter((_, i) => i !== index);
  }
</script>

<div class="space-y-6">
  <!-- Condition Configuration -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-3">Condition</h4>
    
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <Label for="if-else-condition-type" class="mb-2">Condition Type</Label>
        <Select
          id="if-else-condition-type"
          bind:value={config.condition_type}
          items={conditionTypes}
        />
      </div>
      
      {#if selectorRequired}
        <div>
          <Label for="if-else-selector" class="mb-2">Selector *</Label>
          <Input
            id="if-else-selector"
            type="text"
            bind:value={config.selector}
            placeholder="button, #element"
            required
          />
        </div>
      {:else}
        <div class="flex items-center justify-center bg-gray-100 rounded-md p-4">
          <p class="text-sm text-gray-500 italic">
            No selector needed for this condition type
          </p>
        </div>
      {/if}
    </div>

    <div class="mt-4">
      <Label for="if-else-timeout" class="mb-2">Timeout (ms)</Label>
      <Input
        id="if-else-timeout"
        type="number"
        bind:value={config.timeout}
        placeholder="30000"
        min={0}
      />
    </div>
  </div>

  <!-- IF Actions -->
  <div class="border p-4 rounded-md bg-green-50 border-green-200">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-md font-semibold text-green-800">IF Actions (Condition True)</h4>
      <Button size="sm" color="green" onclick={addIfAction}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Action
      </Button>
    </div>

    {#if config.if_actions?.length === 0}
      <p class="text-sm text-green-700 italic">No actions defined for when condition is true.</p>
    {:else}
      <div class="space-y-4">
        {#each config.if_actions as action, index (index)}
          <div class="border p-4 rounded-md bg-white">
            <div class="flex items-center justify-between mb-3">
              <h5 class="text-sm font-semibold">IF Action #{index + 1}</h5>
              <Button
                size="sm"
                color="red"
                onclick={() => removeIfAction(index)}
              >
                <TrashBinOutline class="w-4 h-4" />
              </Button>
            </div>
            
            <div class="mb-4">
              <Label for="if-action-type-{index}" class="mb-2">Action Type *</Label>
              <Select
                id="if-action-type-{index}"
                bind:value={action.action_type}
                items={[
                  { value: "", name: "Select action type" },
                  ...nestedActionTypes.map(type => ({ value: type, name: type }))
                ]}
              />
            </div>

            {#if action.action_type}
              <NestedActionConfigurator 
                actionType={action.action_type} 
                bind:config={action.action_config} 
              />
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- ELSE Actions -->
  <div class="border p-4 rounded-md bg-red-50 border-red-200">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-md font-semibold text-red-800">ELSE Actions (Condition False)</h4>
      <Button size="sm" color="red" onclick={addElseAction}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Action
      </Button>
    </div>

    {#if config.else_actions?.length === 0}
      <p class="text-sm text-red-700 italic">No actions defined for when condition is false.</p>
    {:else}
      <div class="space-y-4">
        {#each config.else_actions as action, index (index)}
          <div class="border p-4 rounded-md bg-white">
            <div class="flex items-center justify-between mb-3">
              <h5 class="text-sm font-semibold">ELSE Action #{index + 1}</h5>
              <Button
                size="sm"
                color="red"
                onclick={() => removeElseAction(index)}
              >
                <TrashBinOutline class="w-4 h-4" />
              </Button>
            </div>
            
            <div class="mb-4">
              <Label for="else-action-type-{index}" class="mb-2">Action Type *</Label>
              <Select
                id="else-action-type-{index}"
                bind:value={action.action_type}
                items={[
                  { value: "", name: "Select action type" },
                  ...nestedActionTypes.map(type => ({ value: type, name: type }))
                ]}
              />
            </div>

            {#if action.action_type}
              <NestedActionConfigurator 
                actionType={action.action_type} 
                bind:config={action.action_config} 
              />
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>