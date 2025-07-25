```svelte
<script lang="ts">
  import { Label, Input, Select, Button, Checkbox } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import { nestedActionTypes } from "$lib/utils/actionConfigMap";
  import NestedActionConfigurator from "../NestedActionConfigurator.svelte";

  type NestedAction = {
    id: string; // Added for tracking nested actions
    action_type: string;
    action_config: Record<string, any>;
  };

  type ApiLoopUntilConfig = {
    variable_path: string;
    condition_type: string;
    expected_value?: any;
    max_loops?: number;
    timeout_ms?: number;
    fail_on_force_stop?: boolean;
    loop_actions: NestedAction[];
  };

  let { config = $bindable() }: { config: ApiLoopUntilConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: ApiLoopUntilConfig) {
    if (!targetConfig.variable_path) targetConfig.variable_path = "";
    if (!targetConfig.condition_type) targetConfig.condition_type = "equals";
    if (!targetConfig.loop_actions) targetConfig.loop_actions = [];
    if (targetConfig.fail_on_force_stop === undefined) targetConfig.fail_on_force_stop = false;
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  const conditionTypes = [
    { value: "equals", name: "Equals" },
    { value: "not_equals", name: "Not Equals" },
    { value: "greater_than", name: "Greater Than" },
    { value: "less_than", name: "Less Than" },
    { value: "greater_than_or_equal", name: "Greater Than or Equal" },
    { value: "less_than_or_equal", name: "Less Than or Equal" },
    { value: "contains", name: "Contains (string)" },
    { value: "not_contains", name: "Does Not Contain (string)" },
    { value: "is_null", name: "Is Null" },
    { value: "is_not_null", name: "Is Not Null" },
    { value: "is_true", name: "Is True (boolean)" },
    { value: "is_false", name: "Is False (boolean)" },
  ];

  const requiresExpectedValue = $derived(
    !["is_null", "is_not_null", "is_true", "is_false"].includes(config.condition_type)
  );

  // Helper functions for managing loop actions
  function addLoopAction() {
    config.loop_actions = [...config.loop_actions, { id: crypto.randomUUID(), action_type: "", action_config: {} }];
  }

  function removeLoopAction(index: number) {
    config.loop_actions = config.loop_actions.filter((_, i) => i !== index);
  }
</script>

<div class="space-y-6">
  <!-- Loop Condition -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-3">Loop Condition</h4>
    <p class="text-sm text-gray-600 mb-4">
      The loop will continue until this condition on a runtime variable is met.
    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
      <div>
        <Label for="loop-var-path" class="mb-2">Runtime Variable Path *</Label>
        <Input
          id="loop-var-path"
          type="text"
          bind:value={config.variable_path}
          placeholder="runtime.lastResponse.data.status"
          required
        />
        <p class="text-xs text-gray-500 mt-1">
          e.g., <code>runtime.lastResponse.data.session.isCompleted</code> or <code>runtime.extracted_id</code>
        </p>
      </div>
      <div>
        <Label for="loop-condition-type" class="mb-2">Condition Type *</Label>
        <Select
          id="loop-condition-type"
          bind:value={config.condition_type}
          items={conditionTypes}
          required
        />
      </div>
    </div>

    {#if requiresExpectedValue}
      <div>
        <Label for="loop-expected-value" class="mb-2">Expected Value</Label>
        <Input
          id="loop-expected-value"
          type="text"
          bind:value={config.expected_value}
          placeholder="true, 'completed', 123"
        />
        <p class="text-xs text-gray-500 mt-1">
          Value to compare against (e.g., <code>true</code>, <code>"completed"</code>, <code>123</code>)
        </p>
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
        {#each config.loop_actions as action, index (action.id)}
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
              <Select
                id="loop-action-type-{index}"
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
```