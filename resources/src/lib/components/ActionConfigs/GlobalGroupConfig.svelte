<script lang="ts">
  import { Label, Button } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import { nestedActionTypes } from "$lib/utils/actionConfigMap";
  import NestedActionConfigurator from "../NestedActionConfigurator.svelte";

  type NestedAction = {
    action_type: string;
    action_config: Record<string, any>;
  };

  type GlobalGroupConfig = {
    actions: NestedAction[];
  };

  let { config = $bindable() }: { config: GlobalGroupConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: GlobalGroupConfig) {
    if (!targetConfig.actions) targetConfig.actions = [];
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  // Helper functions for managing actions
  function addAction() {
    config.actions = [...config.actions, { action_type: "", action_config: {} }];
  }

  function removeAction(index: number) {
    config.actions = config.actions.filter((_, i) => i !== index);
  }
</script>

<div class="space-y-6">
  <!-- Group Actions -->
  <div class="border p-4 rounded-md bg-gray-50">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-md font-semibold">Group Actions</h4>
      <Button size="sm" onclick={addAction}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Action
      </Button>
    </div>
    <p class="text-sm text-gray-600 mb-4">
      Group multiple actions together. Only the last output file from the group will be saved.
    </p>

    {#if config.actions?.length === 0}
      <p class="text-sm text-gray-500 italic">No actions defined. Add actions to group together.</p>
    {:else}
      <div class="space-y-4">
        {#each config.actions as action, index (index)}
          <div class="border p-4 rounded-md bg-white">
            <div class="flex items-center justify-between mb-3">
              <h5 class="text-sm font-semibold">Action #{index + 1}</h5>
              <Button
                size="sm"
                color="red"
                onclick={() => removeAction(index)}
              >
                <TrashBinOutline class="w-4 h-4" />
              </Button>
            </div>
            
            <div class="mb-4">
              <Label for="action-type-{index}" class="mb-2">Action Type *</Label>
              <select
                id="action-type-{index}"
                bind:value={action.action_type}
                class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              >
                <option value="">Select action type</option>
                {#each nestedActionTypes.filter(type => type !== "global:group") as type}
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
      <p><strong>Scenario:</strong> Select country code and fill phone number</p>
      <p><strong>Actions:</strong></p>
      <ul class="list-disc list-inside ml-4">
        <li>Click country dropdown</li>
        <li>Select country option</li>
        <li>Fill phone number input</li>
        <li>Take screenshot (only this file will be saved)</li>
      </ul>
      <p><strong>Result:</strong> All actions execute as a unit, only the final screenshot is saved</p>
    </div>
  </div>
</div>

<style>
  code {
    background-color: #f3f4f6;
    padding: 0.125rem 0.25rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
  }
</style>