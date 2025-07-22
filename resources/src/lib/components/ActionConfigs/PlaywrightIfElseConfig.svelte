<script lang="ts">
  import { Label, Input, Select, Button } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import { nestedActionTypes } from "$lib/utils/actionConfigMap";
  import NestedActionConfigurator from "../NestedActionConfigurator.svelte";

  type NestedAction = {
    action_type: string;
    action_config: Record<string, any>;
  };

  type ElseIfCondition = {
    selector: string;
    condition_type: string;
    actions: NestedAction[];
  };

  type PlaywrightIfElseConfig = {
    selector: string;
    condition_type: string;
    if_actions: NestedAction[];
    else_if_conditions: ElseIfCondition[];
    else_actions: NestedAction[];
  };

  let { config = $bindable() }: { config: PlaywrightIfElseConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightIfElseConfig) {
    if (!targetConfig.selector) targetConfig.selector = "";
    if (!targetConfig.condition_type) targetConfig.condition_type = "is_enabled";
    if (!targetConfig.if_actions) targetConfig.if_actions = [];
    if (!targetConfig.else_if_conditions) targetConfig.else_if_conditions = [];
    if (!targetConfig.else_actions) targetConfig.else_actions = [];
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  const conditionTypes = [
    { value: "is_enabled", name: "Is Enabled" },
    { value: "is_disabled", name: "Is Disabled" },
    { value: "is_visible", name: "Is Visible" },
    { value: "is_hidden", name: "Is Hidden" },
    { value: "is_checked", name: "Is Checked" },
    { value: "is_editable", name: "Is Editable" },
  ];

  // Helper functions for managing nested actions
  function addIfAction() {
    config.if_actions = [...config.if_actions, { action_type: "", action_config: {} }];
  }

  function removeIfAction(index: number) {
    config.if_actions = config.if_actions.filter((_, i) => i !== index);
  }

  function addElseIfCondition() {
    config.else_if_conditions = [
      ...config.else_if_conditions,
      { selector: "", condition_type: "is_enabled", actions: [] }
    ];
  }

  function removeElseIfCondition(index: number) {
    config.else_if_conditions = config.else_if_conditions.filter((_, i) => i !== index);
  }

  function addElseIfAction(conditionIndex: number) {
    config.else_if_conditions[conditionIndex].actions = [
      ...config.else_if_conditions[conditionIndex].actions,
      { action_type: "", action_config: {} }
    ];
  }

  function removeElseIfAction(conditionIndex: number, actionIndex: number) {
    config.else_if_conditions[conditionIndex].actions = 
      config.else_if_conditions[conditionIndex].actions.filter((_, i) => i !== actionIndex);
  }

  function addElseAction() {
    config.else_actions = [...config.else_actions, { action_type: "", action_config: {} }];
  }

  function removeElseAction(index: number) {
    config.else_actions = config.else_actions.filter((_, i) => i !== index);
  }
</script>

<div class="space-y-6">
  <!-- Main Condition -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-3">Main Condition (IF)</h4>
    
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
      <div>
        <Label for="if-selector" class="mb-2">Selector *</Label>
        <Input
          id="if-selector"
          type="text"
          bind:value={config.selector}
          placeholder="input#chatbox-reply-input:not([disabled])"
          required
        />
      </div>
      <div>
        <Label for="if-condition-type" class="mb-2">Condition *</Label>
        <Select
          id="if-condition-type"
          bind:value={config.condition_type}
          items={conditionTypes}
        />
      </div>
    </div>

    <!-- IF Actions -->
    <div>
      <div class="flex items-center justify-between mb-3">
        <Label class="text-sm font-medium">Actions to execute if condition is TRUE</Label>
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
              <h5 class="text-sm font-semibold">Else If #{conditionIndex + 1}</h5>
              <Button
                size="sm"
                color="red"
                onclick={() => removeElseIfCondition(conditionIndex)}
              >
                <TrashBinOutline class="w-4 h-4" />
              </Button>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
              <div>
                <Label for="elseif-selector-{conditionIndex}" class="mb-2">Selector *</Label>
                <Input
                  id="elseif-selector-{conditionIndex}"
                  type="text"
                  bind:value={elseIfCondition.selector}
                  placeholder="div#chatbox-hints button:first-child"
                  required
                />
              </div>
              <div>
                <Label for="elseif-condition-type-{conditionIndex}" class="mb-2">Condition *</Label>
                <Select
                  id="elseif-condition-type-{conditionIndex}"
                  bind:value={elseIfCondition.condition_type}
                  items={conditionTypes}
                />
              </div>
            </div>

            <!-- Else If Actions -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <Label class="text-sm font-medium">Actions to execute if this condition is TRUE</Label>
                <Button size="sm" onclick={() => addElseIfAction(conditionIndex)}>
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
                        <h6 class="text-xs font-semibold">Action #{actionIndex + 1}</h6>
                        <Button
                          size="sm"
                          color="red"
                          onclick={() => removeElseIfAction(conditionIndex, actionIndex)}
                        >
                          <TrashBinOutline class="w-4 h-4" />
                        </Button>
                      </div>
                      
                      <div class="mb-3">
                        <Label for="elseif-action-type-{conditionIndex}-{actionIndex}" class="mb-1 text-xs">Action Type *</Label>
                        <Select
                          id="elseif-action-type-{conditionIndex}-{actionIndex}"
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