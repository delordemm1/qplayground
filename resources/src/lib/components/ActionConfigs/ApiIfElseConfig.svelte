<script lang="ts">
  import { Label, Input, Select, Button } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import { nestedActionTypes } from "$lib/utils/actionConfigMap";
  import NestedActionConfigurator from "../NestedActionConfigurator.svelte";

  type NestedAction = {
    id?: string;
    action_type: string;
    action_config: Record<string, any>;
  };

  type ElseIfCondition = {
    variable_path: string;
    condition_type: string;
    expected_value: any;
    actions: NestedAction[];
  };

  type ApiIfElseConfig = {
    variable_path: string;
    condition_type: string;
    expected_value: any;
    if_actions: NestedAction[];
    else_if_conditions: ElseIfCondition[];
    else_actions: NestedAction[];
    final_actions: NestedAction[];
  };

  let { config = $bindable() }: { config: ApiIfElseConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: ApiIfElseConfig) {
    if (!targetConfig.variable_path) targetConfig.variable_path = "";
    if (!targetConfig.condition_type) targetConfig.condition_type = "equals";
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

  const conditionTypes = [
    { value: "equals", name: "Equals" },
    { value: "not_equals", name: "Not Equals" },
    { value: "contains", name: "Contains" },
    { value: "not_contains", name: "Not Contains" },
    { value: "greater_than", name: "Greater Than" },
    { value: "less_than", name: "Less Than" },
    { value: "greater_than_or_equal", name: "Greater Than or Equal" },
    { value: "less_than_or_equal", name: "Less Than or Equal" },
    { value: "is_null", name: "Is Null" },
    { value: "is_not_null", name: "Is Not Null" },
    { value: "is_true", name: "Is True" },
    { value: "is_false", name: "Is False" },
  ];

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
      { variable_path: "", condition_type: "equals", expected_value: "", actions: [] },
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

  // Check if expected value input should be disabled for certain condition types
  const shouldDisableExpectedValue = $derived(
    config.condition_type === "is_null" || 
    config.condition_type === "is_not_null" ||
    config.condition_type === "is_true" ||
    config.condition_type === "is_false"
  );
</script>

<div class="space-y-6">
  <!-- Main Condition -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-3">Main Condition (IF)</h4>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
      <div>
        <Label for="if-variable-path" class="mb-2">Runtime Variable Path *</Label>
        <Input
          id="if-variable-path"
          type="text"
          bind:value={config.variable_path}
          placeholder="runtime.lastResponse.successCode"
          required
        />
        <p class="text-xs text-gray-500 mt-1">
          Path to runtime variable (e.g., runtime.response.data.options[0].id)
        </p>
      </div>
      <div>
        <Label for="if-condition-type" class="mb-2">Condition *</Label>
        <Select
          id="if-condition-type"
          bind:value={config.condition_type}
          items={conditionTypes}
        />
      </div>
      <div>
        <Label for="if-expected-value" class="mb-2">Expected Value</Label>
        <Input
          id="if-expected-value"
          type="text"
          bind:value={config.expected_value}
          placeholder="201"
          disabled={shouldDisableExpectedValue}
        />
        <p class="text-xs text-gray-500 mt-1">
          {shouldDisableExpectedValue ? "Not required for this condition type" : "Value to compare against (supports variables)"}
        </p>
      </div>
    </div>

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
                <Select
                  id="if-action-type-{index}"
                  bind:value={action.action_type}
                  items={[
                    { value: "", name: "Select action type" },
                    ...nestedActionTypes.map((type) => ({
                      value: type,
                      name: type,
                    })),
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

            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
              <div>
                <Label for="elseif-variable-path-{conditionIndex}" class="mb-2"
                  >Runtime Variable Path *</Label
                >
                <Input
                  id="elseif-variable-path-{conditionIndex}"
                  type="text"
                  bind:value={elseIfCondition.variable_path}
                  placeholder="runtime.response.data.status"
                  required
                />
              </div>
              <div>
                <Label for="elseif-condition-type-{conditionIndex}" class="mb-2"
                  >Condition *</Label
                >
                <Select
                  id="elseif-condition-type-{conditionIndex}"
                  bind:value={elseIfCondition.condition_type}
                  items={conditionTypes}
                />
              </div>
              <div>
                <Label for="elseif-expected-value-{conditionIndex}" class="mb-2"
                  >Expected Value</Label
                >
                <Input
                  id="elseif-expected-value-{conditionIndex}"
                  type="text"
                  bind:value={elseIfCondition.expected_value}
                  placeholder="success"
                  disabled={elseIfCondition.condition_type === "is_null" || 
                           elseIfCondition.condition_type === "is_not_null" ||
                           elseIfCondition.condition_type === "is_true" ||
                           elseIfCondition.condition_type === "is_false"}
                />
              </div>
            </div>

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
                        <Select
                          id="elseif-action-type-{conditionIndex}-{actionIndex}"
                          bind:value={action.action_type}
                          items={[
                            { value: "", name: "Select action type" },
                            ...nestedActionTypes.map((type) => ({
                              value: type,
                              name: type,
                            })),
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
              <Select
                id="else-action-type-{index}"
                bind:value={action.action_type}
                items={[
                  { value: "", name: "Select action type" },
                  ...nestedActionTypes.map((type) => ({
                    value: type,
                    name: type,
                  })),
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
              <Select
                id="final-action-type-{index}"
                bind:value={action.action_type}
                items={[
                  { value: "", name: "Select action type" },
                  ...nestedActionTypes.map((type) => ({
                    value: type,
                    name: type,
                  })),
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