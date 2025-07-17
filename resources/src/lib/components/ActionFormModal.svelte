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

  // Import configuration components
  import PlaywrightGotoConfig from "./ActionConfigs/PlaywrightGotoConfig.svelte";
  import PlaywrightClickConfig from "./ActionConfigs/PlaywrightClickConfig.svelte";
  import PlaywrightFillConfig from "./ActionConfigs/PlaywrightFillConfig.svelte";
  import PlaywrightTypeConfig from "./ActionConfigs/PlaywrightTypeConfig.svelte";
  import PlaywrightScreenshotConfig from "./ActionConfigs/PlaywrightScreenshotConfig.svelte";
  import PlaywrightWaitConfig from "./ActionConfigs/PlaywrightWaitConfig.svelte";
  import PlaywrightEvaluateConfig from "./ActionConfigs/PlaywrightEvaluateConfig.svelte";
  import R2UploadConfig from "./ActionConfigs/R2UploadConfig.svelte";
  import R2DeleteConfig from "./ActionConfigs/R2DeleteConfig.svelte";

  type Action = {
    ID: string;
    ActionType: string;
    ActionConfigJSON: string;
    ActionOrder: number;
  };

  type Props = {
    open: boolean;
    action?: Action | null;
    onSave: (action: {
      action_type: string;
      action_config_json: string;
      action_order: number;
    }) => Promise<void>;
    onClose: () => void;
  };

  let { open = $bindable(), action = null, onSave, onClose }: Props = $props();

  let actionType = $state("");
  let actionOrder = $state(0);
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});

  // This will hold the structured configuration data
  let currentActionConfig = $state<Record<string, any>>({});

  // List of supported action types
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
    "playwright:wait_for_load_state",
    "playwright:screenshot",
    "playwright:evaluate",
    "playwright:hover",
    "playwright:scroll",
    "playwright:get_text",
    "playwright:get_attribute",
    "playwright:set_viewport",
    "playwright:reload",
    "playwright:go_back",
    "playwright:go_forward",
    "r2:upload",
    "r2:delete",
  ];

  // Map action types to their respective config components
  const actionConfigComponents: Record<string, any> = {
    "playwright:goto": PlaywrightGotoConfig,
    "playwright:click": PlaywrightClickConfig,
    "playwright:fill": PlaywrightFillConfig,
    "playwright:type": PlaywrightTypeConfig,
    "playwright:screenshot": PlaywrightScreenshotConfig,
    "playwright:wait_for_selector": PlaywrightWaitConfig,
    "playwright:wait_for_timeout": PlaywrightWaitConfig,
    "playwright:wait_for_load_state": PlaywrightWaitConfig,
    "playwright:evaluate": PlaywrightEvaluateConfig,
    "r2:upload": R2UploadConfig,
    "r2:delete": R2DeleteConfig,
  };

  // Derived state for the current config component
  const CurrentConfigComponent = $derived(actionConfigComponents[actionType]);

  // Effect to initialize form fields when modal opens or action prop changes
  $effect(() => {
    if (open) {
      actionType = action?.ActionType || "";
      actionOrder = action?.ActionOrder || 0;
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
    const validationErrors = validateActionConfig(
      actionType,
      currentActionConfig
    );
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

  function validateActionConfig(
    actionType: string,
    config: Record<string, any>
  ): string[] {
    const errors: string[] = [];

    switch (actionType) {
      case "playwright:goto":
        if (!config.url) errors.push("URL is required");
        break;
      case "playwright:click":
      case "playwright:fill":
      case "playwright:type":
      case "playwright:wait_for_selector":
        if (!config.selector) errors.push("Selector is required");
        if (actionType === "playwright:fill" && !config.value)
          errors.push("Value is required");
        if (actionType === "playwright:type" && !config.text)
          errors.push("Text is required");
        break;
      case "playwright:wait_for_timeout":
        if (!config.timeout_ms || config.timeout_ms <= 0)
          errors.push("Timeout (ms) is required and must be positive");
        break;
      case "playwright:screenshot":
        if (config.upload_to_r2 && !config.r2_key)
          errors.push("R2 key is required when uploading to R2");
        break;
      case "playwright:evaluate":
        if (!config.expression)
          errors.push("JavaScript expression is required");
        break;
      case "r2:upload":
        if (!config.key) errors.push("Object key is required");
        if (!config.content) errors.push("Content is required");
        break;
      case "r2:delete":
        if (!config.key) errors.push("Object key is required");
        break;
    }

    return errors;
  }

  function handleClose() {
    onClose();
    open = false;
  }
</script>

<Modal bind:open outsideclose={false} class="w-full max-w-2xl">
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
          <svelte:component
            this={CurrentConfigComponent}
            bind:config={currentActionConfig}
            {actionType}
          />
        </div>
      {:else if actionType}
        <!-- Fallback for action types without a specific component -->
        <div>
          <Label for="actionConfigJson" class="mb-2">Configuration (JSON)</Label
          >
          <Textarea
            id="actionConfigJson"
            rows={8}
            bind:value={currentActionConfig[actionType]}
            class={"font-mono text-sm " +
              (errors.action_config_json ? "border-red-500" : "")}
          />
          <p class="text-xs text-gray-500 mt-1">
            No specific configuration form available for this action type.
            Please enter JSON configuration manually.
          </p>
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
