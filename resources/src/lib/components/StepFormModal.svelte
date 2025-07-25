<script lang="ts">
  import { Modal, Heading, Button, Label, Input } from "flowbite-svelte";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";
  import StepGeneralConfig from "./StepConfigs/StepGeneralConfig.svelte";

  type Step = {
    ID: string;
    Name: string;
    StepOrder: number;
    ConfigJSON?: string;
  };

  type StepConfig = {
    skip_condition?: string;
    run_only_condition?: string;
    probability?: number;
  };

  type Props = {
    open: boolean;
    step?: Step | null; // Optional step for editing
    maxOrder?: number; // Maximum current order for default calculation
    onSave: (step: { name: string; step_order: number; config_json: string }) => Promise<void>;
    onClose: () => void;
  };

  let { open = $bindable(), step = null, maxOrder = 0, onSave, onClose }: Props = $props();

  let name = $state("");
  let stepOrder = $state(0);
  let stepConfig = $state<StepConfig>({});
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});

  $effect(() => {
    if (open) {
      name = step?.Name || "";
      stepOrder = step?.StepOrder || (maxOrder + 1);
      
      // Parse existing ConfigJSON or use defaults
      try {
        if (step?.ConfigJSON) {
          stepConfig = JSON.parse(step.ConfigJSON);
        } else {
          stepConfig = {};
        }
      } catch (err) {
        console.error("Failed to parse step config JSON:", err);
        stepConfig = {};
      }
      
      errors = {}; // Clear errors when modal opens
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    errors = {}; // Clear previous errors

    if (!name.trim()) {
      errors.name = "Step name is required";
      return;
    }
    if (stepOrder < 0) {
      errors.step_order = "Step order cannot be negative";
      return;
    }

    // Validate step config
    if (stepConfig.skip_condition && stepConfig.run_only_condition) {
      errors.config = "Cannot have both skip condition and run-only condition";
      return;
    }
    
    if ((stepConfig.skip_condition === "random" || stepConfig.run_only_condition === "random") && 
        (stepConfig.probability === undefined || stepConfig.probability < 0 || stepConfig.probability > 1)) {
      errors.config = "Probability must be between 0.0 and 1.0 for random conditions";
      return;
    }
    isLoading = true;
    try {
      // Serialize step config to JSON
      const configJsonString = JSON.stringify(stepConfig);
      await onSave({ name, step_order: stepOrder, config_json: configJsonString });
      open = false; // Close modal on success
    } catch (err: any) {
      if (err.errors) {
        errors = err.errors;
      } else {
        showErrorToast(err.message || "Failed to save step");
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
      {step ? "Edit Step" : "Create New Step"}
    </Heading>

    <form onsubmit={handleSubmit} class="space-y-4">
      <div>
        <Label for="name" class="mb-2">Step Name</Label>
        <Input
          id="name"
          type="text"
          bind:value={name}
          placeholder="Enter step name"
          required
          class={errors.name ? "border-red-500" : ""}
        />
        {#if errors.name}
          <p class="mt-2 text-sm text-red-600">{errors.name}</p>
        {/if}
      </div>

      <div>
        <Label for="stepOrder" class="mb-2">Order</Label>
        <Input
          id="stepOrder"
          type="number"
          bind:value={stepOrder}
          placeholder="0"
          required
          class={errors.step_order ? "border-red-500" : ""} 
        />
        {#if errors.step_order}
          <p class="mt-2 text-sm text-red-600">{errors.step_order}</p>
        {/if}
      </div>

      <div>
        <Label class="mb-2">Step Configuration</Label>
        <div class={errors.config ? "border border-red-500 rounded-md p-3" : "border border-gray-200 rounded-md p-3"}>
          <StepGeneralConfig bind:config={stepConfig} />
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
            {step ? "Save Changes" : "Create Step"}
          {/if}
        </Button>
      </div>
    </form>
  </div>
</Modal>
