<script lang="ts">
  import { Label, Input, Button, Checkbox } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import { nestedActionTypes, conditionTypes } from "$lib/utils/actionConfigMap";
  import NestedActionConfigurator from "../NestedActionConfigurator.svelte";
  import ConditionConfigurator from "../ConditionConfigurator.svelte";

  type NestedAction = {
    action_type: string;
    action_config: Record<string, any>;
  };

  type GlobalLoopConfig = {
    condition_type: string;
    condition_config: Record<string, any>;
    max_loops?: number;
    timeout_ms?: number;
    fail_on_force_stop?: boolean;
    loop_actions: NestedAction[];
  };

  let { config = $bindable() }: { config: GlobalLoopConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: GlobalLoopConfig) {
    if (!targetConfig.condition_config) targetConfig.condition_config = {};
    if (!targetConfig.loop_actions) targetConfig.loop_actions = [];
    if (targetConfig.fail_on_force_stop === undefined) targetConfig.fail_on_force_stop = false;
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  // Helper functions for managing loop actions
  function addLoopAction() {
    config.loop_actions = [...config.loop_actions, { action_type: "", action_config: {} }];
  }

  function removeLoopAction(index: number) {
    config.loop_actions = config.loop_actions.filter((_, i) => i !== index);
  }
</script>

<div class="space-y-6">
  <!-- Loop Condition (Optional) -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-3">Loop Condition (Optional)</h4>
    <p class="text-sm text-gray-600 mb-4">
      If specified, the loop will continue until this condition is met. If not specified, the loop will only stop when force stop conditions are reached.
    </p>
    
    <div class="mb-4">
      <Label for="loop-condition-type" class="mb-2">Condition Type</Label>
      <select
        id="loop-condition-type"
        bind:value={config.condition_type}
        class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
      >
        <option value="">No condition (force stop only)</option>
        {#each conditionTypes as type}
          <option value={type}>{type}</option>
        {/each}
      </select>
      <p class="text-xs text-gray-500 mt-1">
        What condition to check for loop exit
      </p>
    </div>

    {#if config.condition_type}
      <div class="mb-4">
        <Label class="mb-2">Condition Configuration</Label>
        <ConditionConfigurator
          conditionType={config.condition_type}
          bind:config={config.condition_config}
        />
      </div>
    {/if}
  </div>

  <!-- Force Stop Conditions (Required) -->
  <div class="border p-4 rounded-md bg-yellow-50 border-yellow-200">
    <h4 class="text-md font-semibold mb-3 text-yellow-800">Force Stop Conditions (Required)</h4>
    <p class="text-sm text-yellow-700 mb-4">
      At least one force stop condition must be specified to prevent infinite loops.
    </p>
    
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
      <div>
        <Label for="loop-max-loops" class="mb-2">Max Loops</Label>
        <Input
          id="loop-max-loops"
          type="number"
          bind:value={config.max_loops}
          placeholder="10"
          min={1}
        />
        <p class="text-xs text-gray-500 mt-1">
          Maximum number of loop iterations
        </p>
      </div>
      <div>
        <Label for="loop-timeout" class="mb-2">Timeout (ms)</Label>
        <Input
          id="loop-timeout"
          type="number"
          bind:value={config.timeout_ms}
          placeholder="30000"
          min={1000}
        />
        <p class="text-xs text-gray-500 mt-1">
          Maximum time to wait before stopping the loop
        </p>
      </div>
    </div>

    <div class="flex items-center">
      <Checkbox id="loop-fail-on-force-stop" bind:checked={config.fail_on_force_stop} />
      <Label for="loop-fail-on-force-stop" class="ml-2">
        Fail automation if force stop is triggered
      </Label>
    </div>
    <p class="text-xs text-gray-500 mt-1">
      If unchecked, reaching max loops/timeout will log a warning but continue the automation
    </p>
  </div>

  <!-- Loop Actions -->
  <div class="border p-4 rounded-md bg-gray-50">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-md font-semibold">Actions to Repeat</h4>
      <Button size="sm" onclick={addLoopAction}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Action
      </Button>
    </div>

    {#if config.loop_actions?.length === 0}
      <p class="text-sm text-gray-500 italic">No actions defined. Add actions that will be repeated in the loop.</p>
    {:else}
      <div class="space-y-4">
        {#each config.loop_actions as action, index (index)}
          <div class="border p-4 rounded-md bg-white">
            <div class="flex items-center justify-between mb-3">
              <h5 class="text-sm font-semibold">Loop Action #{index + 1}</h5>
              <Button
                size="sm"
                color="red"
                onclick={() => removeLoopAction(index)}
              >
                <TrashBinOutline class="w-4 h-4" />
              </Button>
            </div>
            
            <div class="mb-4">
              <Label for="loop-action-type-{index}" class="mb-2">Action Type *</Label>
              <select
                id="loop-action-type-{index}"
                bind:value={action.action_type}
                class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              >
                <option value="">Select action type</option>
                {#each nestedActionTypes as type}
                  <option value={type}>{type}</option>
                {/each}
              </select>
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

  <!-- Usage Example -->
  <div class="border p-4 rounded-md bg-blue-50 border-blue-200">
    <h4 class="text-md font-semibold mb-3 text-blue-800">Usage Example</h4>
    <div class="text-sm text-blue-700 space-y-2">
      <p><strong>Scenario:</strong> Wait for loading spinner to disappear</p>
      <p><strong>Condition:</strong> playwright:wait_for_selector with selector ".loading-spinner" and state "hidden"</p>
      <p><strong>Loop Actions:</strong> Wait 1 second, check page status</p>
      <p><strong>Force Stop:</strong> Max 30 loops or 60 seconds timeout</p>
    </div>
  </div>
</div>