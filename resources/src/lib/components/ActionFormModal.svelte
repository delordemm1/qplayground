```svelte
<script lang="ts">
  import { Modal, Heading, Button, Label, Input, Textarea, Select, Option } from "flowbite-svelte";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";

  type Action = {
    ID: string;
    ActionType: string;
    ActionConfigJSON: string;
    ActionOrder: number;
  };

  type Props = {
    open: boolean;
    action?: Action | null; // Optional action for editing
    onSave: (action: {
      action_type: string;
      action_config_json: string;
      action_order: number;
    }) => Promise<void>;
    onClose: () => void;
  };

  let {
    open = $bindable(),
    action = null,
    onSave,
    onClose,
  }: Props = $props();

  let actionType = $state("");
  let actionConfigJson = $state("{}");
  let actionOrder = $state(0);
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});

  // List of supported action types (extend as needed)
  const actionTypes = [
    "playwright:goto",
    "playwright:click",
    "playwright:fill",
    "playwright:type",
    "playwright:press",
    "playwright:check",
    "playwright:uncheck",
    "playwright:select_option",
    "playwright:wait_for_selector",
    "playwright:wait_for_timeout",
    "playwright:screenshot",
    "playwright:evaluate",
    "playwright:hover",
    "playwright:scroll",
    "playwright:get_text",
    "playwright:get_attribute",
    "playwright:wait_for_load_state",
    "playwright:set_viewport",
    "playwright:reload",
    "playwright:go_back",
    "playwright:go_forward",
    "r2:upload",
    "r2:delete",
  ];

  $effect(() => {
    if (open) {
      actionType = action?.ActionType || "";
      actionConfigJson = action?.ActionConfigJSON || "{}";
      actionOrder = action?.ActionOrder || 0;
      errors = {}; // Clear errors when modal opens
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    errors = {}; // Clear previous errors

    if (!actionType.trim()) {
      errors.action_type = "Action type is required";
      return;
    }
    if (actionOrder < 0) {
      errors.action_order = "Action order cannot be negative";
      return;
    }

    // Basic JSON validation for config
    try {
      JSON.parse(actionConfigJson);
    } catch (err) {
      errors.action_config_json = "Invalid JSON format";
      return;
    }

    isLoading = true;
    try {
      await onSave({
        action_type: actionType,
        action_config_json: actionConfigJson,
        action_order: actionOrder,
      });
      open = false; // Close modal on success
    } catch (err: any) {
      if (err.errors) {
        errors = err.errors;
      } else {
        showErrorToast(err.message || "Failed to save action");
      }
    } finally {
      isLoading = false;
    }
  }

  function handleClose() {
    onClose();
    open = false;
  }
</script>

<Modal bind:open outsideclose={false} class="w-full max-w-md">
  <div class="p-6">
    <Heading tag="h3" class="text-xl font-semibold mb-4">
      {action ? "Edit Action" : "Create New Action"}
    </Heading>

    <form onsubmit={handleSubmit} class="space-y-4">
      <div>
        <Label for="actionType" class="mb-2">Action Type</Label>
        <Select
          id="actionType"
          bind:value={actionType}
          required
          class:border-red-500={!!errors.action_type}
        >
          <Option value="" disabled>Select an action type</Option>
          {#each actionTypes as type}
            <Option value={type}>{type}</Option>
          {/each}
        </Select>
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
          class:border-red-500={!!errors.action_order}
        />
        {#if errors.action_order}
          <p class="mt-2 text-sm text-red-600">{errors.action_order}</p>
        {/if}
      </div>

      <div>
        <Label for="actionConfigJson" class="mb-2">Configuration (JSON)</Label>
        <Textarea
          id="actionConfigJson"
          rows="8"
          bind:value={actionConfigJson}
          placeholder="{}"
          class="font-mono text-sm"
          class:border-red-500={!!errors.action_config_json}
        />
        {#if errors.action_config_json}
          <p class="mt-2 text-sm text-red-600">{errors.action_config_json}</p>
        {/if}
      </div>

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
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            Saving...
          {:else}
            {action ? "Save Changes" : "Create Action"}
          {/if}
        </Button>
      </div>
    </form>
  </div>
</Modal>
```