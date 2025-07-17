<script lang="ts">
  import { Modal, Heading, Button, Label, Input, Textarea } from "flowbite-svelte";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";
  import AutomationGeneralConfig from "./AutomationConfigs/AutomationGeneralConfig.svelte";

  type Automation = {
    ID: string;
    Name: string;
    Description: string;
    ConfigJSON: string;
  };

  type Variable = {
    key: string;
    type: "static" | "dynamic" | "environment";
    value: string;
    description?: string;
  };

  type MultiRunConfig = {
    enabled: boolean;
    mode: "sequential" | "parallel";
    count: number;
    delay: number;
  };

  type AutomationConfig = {
    variables: Variable[];
    multirun: MultiRunConfig;
    timeout: number;
    retries: number;
    screenshots: {
      enabled: boolean;
      onError: boolean;
      onSuccess: boolean;
      path: string;
    };
    notifications: {
      onComplete: boolean;
      onError: boolean;
      webhook?: string;
    };
  };

  type Props = {
    open: boolean;
    automation?: Automation | null; // Optional automation for editing
    onSave: (automation: {
      name: string;
      description: string;
      config_json: string;
    }) => Promise<void>;
    onClose: () => void;
  };

  let {
    open = $bindable(),
    automation = null,
    onSave,
    onClose,
  }: Props = $props();

  let name = $state("");
  let description = $state("");
  let automationConfig = $state<AutomationConfig>({
    variables: [],
    multirun: {
      enabled: false,
      mode: "sequential",
      count: 1,
      delay: 1000,
    },
    timeout: 300,
    retries: 0,
    screenshots: {
      enabled: true,
      onError: true,
      onSuccess: false,
      path: "screenshots/{{timestamp}}-{{loopIndex}}.png",
    },
    notifications: {
      onComplete: false,
      onError: true,
      webhook: "",
    },
  });
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});

  $effect(() => {
    if (open) {
      name = automation?.Name || "";
      description = automation?.Description || "";
      
      // Parse existing ConfigJSON or use defaults
      try {
        if (automation?.ConfigJSON) {
          const parsed = JSON.parse(automation.ConfigJSON);
          // Merge with defaults to ensure all properties exist
          automationConfig = {
            variables: parsed.variables || [],
            multirun: {
              enabled: parsed.multirun?.enabled || false,
              mode: parsed.multirun?.mode || "sequential",
              count: parsed.multirun?.count || 1,
              delay: parsed.multirun?.delay || 1000,
            },
            timeout: parsed.timeout || 300,
            retries: parsed.retries || 0,
            screenshots: {
              enabled: parsed.screenshots?.enabled !== undefined ? parsed.screenshots.enabled : true,
              onError: parsed.screenshots?.onError !== undefined ? parsed.screenshots.onError : true,
              onSuccess: parsed.screenshots?.onSuccess || false,
              path: parsed.screenshots?.path || "screenshots/{{timestamp}}-{{loopIndex}}.png",
            },
            notifications: {
              onComplete: parsed.notifications?.onComplete || false,
              onError: parsed.notifications?.onError !== undefined ? parsed.notifications.onError : true,
              webhook: parsed.notifications?.webhook || "",
            },
          };
        }
      } catch (err) {
        console.error("Failed to parse existing automation config JSON:", err);
        // Keep default values if parsing fails
      }
      
      errors = {}; // Clear errors when modal opens
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    errors = {}; // Clear previous errors

    if (!name.trim()) {
      errors.name = "Automation name is required";
      return;
    }

    // Validate automation config
    try {
      // Basic validation for required fields
      if (automationConfig.timeout <= 0) {
        errors.config = "Timeout must be greater than 0";
        return;
      }
      if (automationConfig.retries < 0) {
        errors.config = "Retries cannot be negative";
        return;
      }
      if (automationConfig.multirun.enabled && automationConfig.multirun.count <= 0) {
        errors.config = "Multi-run count must be greater than 0";
        return;
      }
      
      // Validate variables
      for (const variable of automationConfig.variables) {
        if (!variable.key.trim()) {
          errors.config = "All variables must have a key";
          return;
        }
        if (!variable.value.trim()) {
          errors.config = "All variables must have a value";
          return;
        }
      }
      
      // Check for duplicate variable keys
      const variableKeys = automationConfig.variables.map(v => v.key);
      const uniqueKeys = new Set(variableKeys);
      if (variableKeys.length !== uniqueKeys.size) {
        errors.config = "Variable keys must be unique";
        return;
      }
    } catch (err) {
      errors.config = "Invalid configuration";
      return;
    }

    isLoading = true;
    try {
      // Serialize the automation config to JSON
      const configJsonString = JSON.stringify(automationConfig);
      await onSave({ name, description, config_json: configJsonString });
      open = false; // Close modal on success
    } catch (err: any) {
      if (err.errors) {
        errors = err.errors;
      } else {
        // showErrorToast(err.message || "Failed to save automation");
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
      {automation ? "Edit Automation" : "Create New Automation"}
    </Heading>

    <form onsubmit={handleSubmit} class="space-y-4">
      <div>
        <Label for="name" class="mb-2">Automation Name</Label>
        <Input
          id="name"
          type="text"
          bind:value={name}
          placeholder="Enter automation name"
          required
          class={errors.name ? "border-red-500" : ""}
        />
        {#if errors.name}
          <p class="mt-2 text-sm text-red-600">{errors.name}</p>
        {/if}
      </div>

      <div>
        <Label for="description" class="mb-2">Description (optional)</Label>
        <Textarea
          id="description"
          rows={3}
          bind:value={description}
          placeholder="Enter automation description"
          class={errors.description ? "border-red-500" : ""}
        />
        {#if errors.description}
          <p class="mt-2 text-sm text-red-600">{errors.description}</p>
        {/if}
      </div>

      <div>
        <Label class="mb-2">Automation Configuration</Label>
        <div class={errors.config ? "border border-red-500 rounded-md p-1" : ""}>
          <AutomationGeneralConfig bind:config={automationConfig} />
        </div>
        {#if errors.config}
          <p class="mt-2 text-sm text-red-600">{errors.config}</p>
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
            {automation ? "Save Changes" : "Create Automation"}
          {/if}
        </Button>
      </div>
    </form>
  </div>
</Modal>
