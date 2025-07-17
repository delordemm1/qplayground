```svelte
<script lang="ts">
  import { Modal, Heading, Button, Label, Input, Textarea } from "flowbite-svelte";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";

  type Automation = {
    ID: string;
    Name: string;
    Description: string;
    ConfigJSON: string;
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
  let configJson = $state("{}");
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});

  $effect(() => {
    if (open) {
      name = automation?.Name || "";
      description = automation?.Description || "";
      configJson = automation?.ConfigJSON || "{}";
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

    // Basic JSON validation
    try {
      JSON.parse(configJson);
    } catch (err) {
      errors.config_json = "Invalid JSON format";
      return;
    }

    isLoading = true;
    try {
      await onSave({ name, description, config_json: configJson });
      open = false; // Close modal on success
    } catch (err: any) {
      if (err.errors) {
        errors = err.errors;
      } else {
        showErrorToast(err.message || "Failed to save automation");
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
          class:border-red-500={!!errors.name}
        />
        {#if errors.name}
          <p class="mt-2 text-sm text-red-600">{errors.name}</p>
        {/if}
      </div>

      <div>
        <Label for="description" class="mb-2">Description (optional)</Label>
        <Textarea
          id="description"
          rows="3"
          bind:value={description}
          placeholder="Enter automation description"
          class:border-red-500={!!errors.description}
        />
        {#if errors.description}
          <p class="mt-2 text-sm text-red-600">{errors.description}</p>
        {/if}
      </div>

      <div>
        <Label for="configJson" class="mb-2">Configuration (JSON)</Label>
        <Textarea
          id="configJson"
          rows="6"
          bind:value={configJson}
          placeholder="{}"
          class="font-mono text-sm"
          class:border-red-500={!!errors.config_json}
        />
        {#if errors.config_json}
          <p class="mt-2 text-sm text-red-600">{errors.config_json}</p>
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
```