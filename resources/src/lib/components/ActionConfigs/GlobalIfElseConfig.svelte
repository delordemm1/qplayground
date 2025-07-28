<script lang="ts">
  import { Label, Input, Select, Button } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import { nestedActionTypes, conditionTypes } from "$lib/utils/actionConfigMap";
  import NestedActionConfigurator from "../NestedActionConfigurator.svelte";
  import ConditionConfigurator from "../ConditionConfigurator.svelte";

  type NestedAction = {
    action_type: string;
    action_config: Record<string, any>;
  };

  type ElseIfCondition = {
    condition_type: string;
    condition_config: Record<string, any>;
    actions: NestedAction[];
  };

  type GlobalIfElseConfig = {
    condition_type: string;
    condition_config: Record<string, any>;
    if_actions: NestedAction[];
    else_if_conditions: ElseIfCondition[];
    else_actions: NestedAction[];
    final_actions: NestedAction[];
  };

  let { config = $bindable() }: { config: GlobalIfElseConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: GlobalIfElseConfig) {
    if (!targetConfig.condition_type) targetConfig.condition_type = "";
    if (!targetConfig.condition_config) targetConfig.condition_config = {};
    if (!targetConfig.if_actions) targetConfig.if_actions = [];
    if (!targetConfig.else_if_conditions) targetConfig.else_if_conditions = [];
    if (!targetConfig.else_actions) targetConfig.else_actions = [];
    if (!targetConfig.final_actions) targetConfig.final_actions = [];
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  // Helper functions for managing nested actions
  function addIfAction() {
    config.if_actions = [
      ...config.if_actions,
      { action_type: "", action_config: {} },
    ];
  }

  function removeIfAction(index: number) {
    config.if_actions = config.if_actions.filter((_, i) => i !== index);
  }

  function addElseIfCondition() {
    config.else_if_conditions = [
      ...config.else_if_conditions,
      { condition_type: "", condition_config: {}, actions: [] },
    ];
  }

  function removeElseIfCondition(index: number) {
    config.else_if_conditions = config.else_if_conditions.filter(
      (_, i) => i !== index
    );
  }

  function addElseIfAction(conditionIndex: number) {
    config.else_if_conditions[conditionIndex].actions = [
      ...config.else_if_conditions[conditionIndex].actions,
      { action_type: "", action_config: {} },
    ];
  }

  function removeElseIfAction(conditionIndex: number, actionIndex: number) {
    config.else_if_conditions[conditionIndex].actions =
      config.else_if_conditions[conditionIndex].actions.filter(
        (_, i) => i !== actionIndex
      );
  }

  function addElseAction() {
    config.else_actions = [
      ...config.else_actions,
      { action_type: "", action_config: {} },
    ];
  }

  function removeElseAction(index: number) {
    config.else_actions = config.else_actions.filter((_, i) => i !== index);
  }

  function addFinalAction() {
    config.final_actions = [
      ...config.final_actions,
      { action_type: "", action_config: {} },
    ];
  }

  function removeFinalAction(index: number) {
    config.final_actions = config.final_actions.filter((_, i) => i !== index);
  }
</script>

<div class="space-y-6">
  <!-- Main Condition -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-3">Main Condition (IF)</h4>

    <div class="mb-4">
      <Label for="if-condition-type" class="mb-2">Condition Type *</Label>
      <Select
        id="if-condition-type"
        bind:value={config.condition_type}
        items={[
          { value: "", name: "Select condition type" },
          ...conditionTypes.map((type) => ({
            value: type,
            name: type,
          })),
        ]}
      />
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

    <!-- IF Actions -->
    <div>
      <div class="flex items-center justify-between mb-3">
        <Label class="text-sm font-medium"
          >Actions to execute if condition is TRUE</Label
        >
        <Button size="sm" onclick={addIfAction}>
          <PlusOutline class="w-4 h-4 mr-2" />
          Add Action
        </Button>
      </div>

      {#if config.if_actions?.length === 0}
        <p class="text-sm text-gray-500 italic">No actions defined</p>
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
                <Label for="if-action-type-{index}" class="mb-2"
                  >Action Type *</Label
                >
                <select
                  id="if-action-type-{index}"
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
  </div>

  <!-- ELSE IF Conditions -->
  <div class="border p-4 rounded-md bg-gray-50">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-md font-semibold">ELSE IF Conditions</h4>
      <Button size="sm" onclick={addElseIfCondition}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Else If
      </Button>
    </div>

    {#if config.else_if_conditions?.length === 0}
      <p class="text-sm text-gray-500 italic">No else-if conditions defined</p>
    {:else}
      <div class="space-y-4">
        {#each config.else_if_conditions as elseIfCondition, conditionIndex (conditionIndex)}
          <div class="border p-4 rounded-md bg-white">
            <div class="flex items-center justify-between mb-3">
              <h5 class="text-sm font-semibold">
                Else If #{conditionIndex + 1}
              </h5>
              <Button
                size="sm"
                color="red"
                onclick={() => removeElseIfCondition(conditionIndex)}
              >
                <TrashBinOutline class="w-4 h-4" />
              </Button>
            </div>

            <div class="mb-4">
              <Label for="elseif-condition-type-{conditionIndex}" class="mb-2"
                >Condition Type *</Label
              >
              <select
                id="elseif-condition-type-{conditionIndex}"
                bind:value={elseIfCondition.condition_type}
                class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
              >
                <option value="">Select condition type</option>
                {#each conditionTypes as type}
                  <option value={type}>{type}</option>
                {/each}
              </select>
            </div>

            {#if elseIfCondition.condition_type}
              <div class="mb-4">
                <Label class="mb-2">Condition Configuration</Label>
                <ConditionConfigurator
                  conditionType={elseIfCondition.condition_type}
                  bind:config={elseIfCondition.condition_config}
                />
              </div>
            {/if}

            <!-- Else If Actions -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <Label class="text-sm font-medium"
                  >Actions to execute if this condition is TRUE</Label
                >
                <Button
                  size="sm"
                  onclick={() => addElseIfAction(conditionIndex)}
                >
                  <PlusOutline class="w-4 h-4 mr-2" />
                  Add Action
                </Button>
              </div>

              {#if elseIfCondition.actions?.length === 0}
                <p class="text-sm text-gray-500 italic">No actions defined</p>
              {:else}
                <div class="space-y-3">
                  {#each elseIfCondition.actions as action, actionIndex (actionIndex)}
                    <div class="border p-3 rounded-md bg-gray-100">
                      <div class="flex items-center justify-between mb-3">
                        <h6 class="text-xs font-semibold">
                          Action #{actionIndex + 1}
                        </h6>
                        <Button
                          size="sm"
                          color="red"
                          onclick={() =>
                            removeElseIfAction(conditionIndex, actionIndex)}
                        >
                          <TrashBinOutline class="w-4 h-4" />
                        </Button>
                      </div>

                      <div class="mb-3">
                        <Label
                          for="elseif-action-type-{conditionIndex}-{actionIndex}"
                          class="mb-1 text-xs">Action Type *</Label
                        >
                        <select
                          id="elseif-action-type-{conditionIndex}-{actionIndex}"
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
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- ELSE Actions -->
  <div class="border p-4 rounded-md bg-gray-50">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-md font-semibold">ELSE Actions</h4>
      <Button size="sm" onclick={addElseAction}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Action
      </Button>
    </div>

    {#if config.else_actions?.length === 0}
      <p class="text-sm text-gray-500 italic">No else actions defined</p>
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
              <Label for="else-action-type-{index}" class="mb-2"
                >Action Type *</Label
              >
              <select
                id="else-action-type-{index}"
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

  <!-- FINAL Actions -->
  <div class="border p-4 rounded-md bg-green-50 border-green-200">
    <div class="flex items-center justify-between mb-3">
      <h4 class="text-md font-semibold text-green-800">FINAL Actions</h4>
      <Button size="sm" onclick={addFinalAction}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Action
      </Button>
    </div>
    <p class="text-sm text-green-700 mb-4">
      These actions will always execute after the IF/ELSE IF/ELSE logic
      completes, regardless of which path was taken.
    </p>

    {#if config.final_actions?.length === 0}
      <p class="text-sm text-gray-500 italic">No final actions defined</p>
    {:else}
      <div class="space-y-4">
        {#each config.final_actions as action, index (index)}
          <div class="border p-4 rounded-md bg-white">
            <div class="flex items-center justify-between mb-3">
              <h5 class="text-sm font-semibold">FINAL Action #{index + 1}</h5>
              <Button
                size="sm"
                color="red"
                onclick={() => removeFinalAction(index)}
              >
                <TrashBinOutline class="w-4 h-4" />
              </Button>
            </div>

            <div class="mb-4">
              <Label for="final-action-type-{index}" class="mb-2"
                >Action Type *</Label
              >
              <select
                id="final-action-type-{index}"
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
      <p><strong>Scenario:</strong> Check if submit button is visible and enabled</p>
      <p><strong>Main Condition:</strong> playwright:wait_for_selector with selector "#submit-btn" and state "visible"</p>
      <p><strong>IF Actions:</strong> Click submit button, take screenshot</p>
      <p><strong>ELSE Actions:</strong> Log error message, take error screenshot</p>
      <p><strong>FINAL Actions:</strong> Always log completion status</p>
    </div>
  </div>
</div>