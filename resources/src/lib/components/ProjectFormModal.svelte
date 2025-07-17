```svelte
<script lang="ts">
  import { Modal, Heading, Button, Label, Input, Textarea } from "flowbite-svelte";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";

  type Project = {
    ID: string;
    Name: string;
    Description: string;
  };

  type Props = {
    open: boolean;
    project?: Project | null; // Optional project for editing
    onSave: (project: { name: string; description: string }) => Promise<void>;
    onClose: () => void;
  };

  let {
    open = $bindable(),
    project = null,
    onSave,
    onClose,
  }: Props = $props();

  let name = $state("");
  let description = $state("");
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});

  $effect(() => {
    if (open) {
      name = project?.Name || "";
      description = project?.Description || "";
      errors = {}; // Clear errors when modal opens
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    errors = {}; // Clear previous errors

    if (!name.trim()) {
      errors.name = "Project name is required";
      return;
    }

    isLoading = true;
    try {
      await onSave({ name, description });
      open = false; // Close modal on success
    } catch (err: any) {
      if (err.errors) {
        errors = err.errors;
      } else {
        showErrorToast(err.message || "Failed to save project");
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
      {project ? "Edit Project" : "Create New Project"}
    </Heading>

    <form onsubmit={handleSubmit} class="space-y-4">
      <div>
        <Label for="name" class="mb-2">Project Name</Label>
        <Input
          id="name"
          type="text"
          bind:value={name}
          placeholder="Enter project name"
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
          rows="4"
          bind:value={description}
          placeholder="Enter project description"
          class:border-red-500={!!errors.description}
        />
        {#if errors.description}
          <p class="mt-2 text-sm text-red-600">{errors.description}</p>
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
            {project ? "Save Changes" : "Create Project"}
          {/if}
        </Button>
      </div>
    </form>
  </div>
</Modal>
```