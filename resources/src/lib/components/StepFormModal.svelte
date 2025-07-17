```svelte
<script lang="ts">
  import { Modal, Heading, Button, Label, Input } from "flowbite-svelte";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";

  type Step = {
    ID: string;
    Name: string;
    StepOrder: number;
  };

  type Props = {
    open: boolean;
    step?: Step | null; // Optional step for editing
    onSave: (step: { name: string; step_order: number }) => Promise<void>;
    onClose: () => void;
  };

  let { open = $bindable(), step = null, onSave, onClose }: Props = $props();

  let name = $state("");
  let stepOrder = $state(0);
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});

  $effect(() => {
    if (open) {
      name = step?.Name || "";
      stepOrder = step?.StepOrder || 0;
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

    isLoading = true;
    try {
      await onSave({ name, step_order: stepOrder });
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
          class:border-red-500={!!errors.name}
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
          class:border-red-500={!!errors.step_order}
        />
        {#if errors.step_order}
          <p class="mt-2 text-sm text-red-600">{errors.step_order}</p>
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
          {#else}
            {step ? "Save Changes" : "Create Step"}
          {/if}
        </Button>
      </div>
    </form>
  </div>
</Modal>
```