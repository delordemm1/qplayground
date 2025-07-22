<script lang="ts">
  import {
    Modal,
    Heading,
    Button,
    Label,
    Input,
    Textarea,
    Select,
  } from "flowbite-svelte";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";

  // Import shared action configuration utilities
  import { actionTypes, actionConfigComponents, validateActionConfig } from "$lib/utils/actionConfigMap";

  type Action = {
    ID: string;
    ActionType: string;
    ActionConfigJSON: string;
    ActionOrder: number;
  };

  type Props = {
    open: boolean;
    action?: Action | null;
    maxOrder?: number; // Maximum current order for default calculation
    onSave: (action: {
      action_type: string;
      action_config_json: string;
      action_order: number;
    }) => Promise<void>;
    onClose: () => void;
  };

  let { open = $bindable(), action = null, maxOrder = 0, onSave, onClose }: Props = $props();

  let actionType = $state("");
  let actionOrder = $state(0);
  let isLoading = $state(false);
  let isDuplicating = $state(false);
  let errors = $state<Record<string, string>>({});

  // This will hold the structured configuration data
  let currentActionConfig = $state<Record<string, any>>({});

  // Derived state for the current config component
  const CurrentConfigComponent = $derived(actionConfigComponents[actionType]);

  // Effect to initialize form fields when modal opens or action prop changes
  $effect(() => {
    if (open) {
      actionType = action?.ActionType || "";
      actionOrder = action?.ActionOrder || (maxOrder + 1);
      errors = {};

      // Parse existing JSON config into structured object
      try {
        currentActionConfig = action?.ActionConfigJSON
          ? JSON.parse(action.ActionConfigJSON)
          : {};
      } catch (e) {
        console.error("Failed to parse existing action config JSON:", e);
        currentActionConfig = {};
        errors.action_config_json = "Invalid existing JSON format";
      }
    }
  });

  // Reset config when action type changes
  $effect(() => {
    if (actionType && !action) {
      // Only reset for new actions, not when editing existing ones
      currentActionConfig = {};
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    errors = {};

    if (!actionType.trim()) {
      errors.action_type = "Action type is required";
      return;
    }
    if (actionOrder < 0) {
      errors.action_order = "Action order cannot be negative";
      return;
    }

    // Basic validation for currentActionConfig
    if (
      typeof currentActionConfig !== "object" ||
      currentActionConfig === null
    ) {
      errors.action_config_json = "Invalid configuration data";
      return;
    }

    // Validate required fields based on action type
    const validationErrors = validateActionConfig(actionType, currentActionConfig);
    if (validationErrors.length > 0) {
      errors.action_config_json = validationErrors.join(", ");
      return;
    }

    let actionConfigJsonString: string;
    try {
      actionConfigJsonString = JSON.stringify(currentActionConfig);
    } catch (err) {
      errors.action_config_json = "Failed to serialize configuration to JSON";
      return;
    }

    isLoading = true;
    try {
      await onSave({
        action_type: actionType,
        action_config_json: actionConfigJsonString,
        action_order: actionOrder,
      });
      open = false;
    } catch (err: any) {
      console.error("Failed to save action:", err);
      if (err.errors) {
        errors = err.errors;
      } else {
        showErrorToast(err.message || "Failed to save action");
        throw new Error(err.message || "Failed to save action");
      }
    } finally {
      isLoading = false;
    }
  }

  function handleClose() {
    onClose();
    open = false;
  }

  function handleDuplicateAction() {
    if (!action) return;
    
    isDuplicating = true;
    // Reset to creation mode with copied data
    actionType = action.ActionType;
    actionOrder = (maxOrder || 0) + 1;
    
    // Parse and copy the action config
    try {
      currentActionConfig = action.ActionConfigJSON 
        ? JSON.parse(action.ActionConfigJSON) 
        : {};
    } catch (e) {
      console.error("Failed to parse action config for duplication:", e);
      currentActionConfig = {};
    }
    
    // Clear the action prop to signal creation mode
    action = null;
    isDuplicating = false;
  }
</script>

<Modal bind:open outsideclose={false} class="" size="lg">
  <div class="p-6">
    <Heading tag="h3" class="text-xl font-semibold mb-4">
      {action ? "Edit Action" : "Create New Action"}
    </Heading>
    
    {#if action && !isDuplicating}
      <div class="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
        <div class="flex items-center justify-between">
          <span class="text-sm text-blue-800">
            Want to create a similar action? You can duplicate this action and modify it.
          </span>
          <button
            type="button"
            onclick={handleDuplicateAction}
            class="text-sm font-medium text-blue-600 hover:text-blue-800"
          >
            Duplicate Action
          </button>
        </div>
      </div>
    {/if}

    <form onsubmit={handleSubmit} class="space-y-4">
      <div>
        <Label for="actionType" class="mb-2">Action Type</Label>
        <Select
          id="actionType"
          bind:value={actionType}
          required
          class={errors.action_type ? "border-red-500" : ""}
          items={[
            { value: "", name: "Select an action type" },
            ...actionTypes.map((type) => ({ value: type, name: type })),
          ]}
        />
        {#if errors.action_type}
          <p class="mt-2 text-sm text-red-600">{errors.action_type}</p>
        {/if}
      </div>

      <div>
        <Label for="actionOrder" class="mb-2">Order</Label>
        <Input
          id="actionOrder"
          type="number"
          bind:value={actionOrder}
          placeholder="0"
          required
          class={errors.action_order ? "border-red-500" : ""}
        />
        {#if errors.action_order}
          <p class="mt-2 text-sm text-red-600">{errors.action_order}</p>
        {/if}
      </div>

      <!-- Dynamic Configuration Fields -->
      {#if CurrentConfigComponent}
        <div class="border p-4 rounded-md bg-gray-50">
          <h4 class="text-md font-semibold mb-3">Action Configuration</h4>
          <!-- <svelte:component
            this={CurrentConfigComponent}
            bind:config={currentActionConfig}
            {actionType}
          /> -->
          <CurrentConfigComponent bind:config={currentActionConfig} {actionType} />
        </div>
      {:else}
        <p class="text-sm text-gray-500">
          Select an action type to configure its parameters.
        </p>
      {/if}

      {#if errors.action_config_json}
        <p class="mt-2 text-sm text-red-600">{errors.action_config_json}</p>
      {/if}

      <div class="flex justify-end space-x-3 pt-4">
        <Button color="alternative" onclick={handleClose} disabled={isLoading}>
          Cancel
        </Button>
        <Button type="submit" color="primary" disabled={isLoading}>
          {#if isLoading}
            <svg
              class="animate-spin -ml-1 mr-3 h-5 w-5 text-white"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
          {#if actionType}
            {#if CurrentConfigComponent}
              <div class="border p-4 rounded-md bg-gray-50">
                <h4 class="text-md font-semibold mb-3">Action Configuration</h4>
                <CurrentConfigComponent bind:config={currentActionConfig} {actionType} />
              </div>
            {:else}
              <div class="border p-4 rounded-md bg-gray-100">
                <p class="text-sm text-gray-500 italic">
                  No configuration available for action type: {actionType}
                </p>
              </div>
            {/if}
          {:else}
            <div class="border p-4 rounded-md bg-gray-100">
              <p class="text-sm text-gray-500">
                Select an action type to configure its parameters.
              </p>
            </div>
          {/if}
        </Button>
      </div>
    </form>
  </div>
</Modal>
