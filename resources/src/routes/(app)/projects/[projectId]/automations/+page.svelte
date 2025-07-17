<script lang="ts">
  import { page } from "@inertiajs/svelte";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";
  import AutomationFormModal from "$lib/components/AutomationFormModal.svelte";
  import ConfirmDeleteModal from "$lib/components/ConfirmDeleteModal.svelte";
  import { formatDate } from "$lib/utils/date";

  type Project = {
    ID: string;
    Name: string;
  };

  type Automation = {
    ID: string;
    Name: string;
    Description: string;
    CreatedAt: string;
  };

  type Props = {
    project: Project;
    automations: Automation[];
    user: any; // Assuming user type is defined elsewhere
  };

  let { project, automations }: Props = $props();

  let showCreateAutomationModal = $state(false);
  let showEditAutomationModal = $state(false);
  let showDeleteAutomationConfirm = $state(false);
  let isDeletingAutomation = $state(false);
  let selectedAutomation = $state<Automation | null>(null);

  const projectId = $derived($page.props.params.projectId);

  function openEditModal(automation: Automation) {
    selectedAutomation = automation;
    showEditAutomationModal = true;
  }

  function openDeleteConfirm(automation: Automation) {
    selectedAutomation = automation;
    showDeleteAutomationConfirm = true;
  }

  async function handleSaveAutomation(data: {
    name: string;
    description: string;
    config_json: string;
  }) {
    try {
      const method = selectedAutomation ? "PUT" : "POST";
      const url = selectedAutomation
        ? `/projects/${projectId}/automations/${selectedAutomation.ID}`
        : `/projects/${projectId}/automations`;

      const response = await fetch(url, {
        method: method,
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (response.ok) {
        showSuccessToast(
          selectedAutomation ? "Automation updated successfully" : "Automation created successfully"
        );
        if (selectedAutomation) {
          // Update existing automation in the list
          automations = automations.map((a) =>
            a.ID === result.automation.ID ? result.automation : a
          );
        } else {
          // Add new automation to the list
          automations = [...automations, result.automation];
        }
      } else {
        throw result;
      }
    } catch (err: any) {
      console.error("Failed to save automation:", err);
      if (err.errors) {
        throw err; // Re-throw to be caught by modal
      } else {
        showErrorToast(err.message || "Failed to save automation");
        throw new Error(err.message || "Failed to save automation");
      }
    }
  }

  async function handleDeleteAutomation() {
    if (!selectedAutomation) return;

    isDeletingAutomation = true;
    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${selectedAutomation.ID}`,
        {
          method: "DELETE",
        }
      );

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Automation deleted successfully");
        automations = automations.filter((a) => a.ID !== selectedAutomation?.ID);
        selectedAutomation = null;
      } else {
        showErrorToast(result.error || "Failed to delete automation");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isDeletingAutomation = false;
      showDeleteAutomationConfirm = false;
    }
  }
</script>

<svelte:head>
  <title>Automations for {project.Name} - QPlayground</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
  <!-- Header -->
  <div class="md:flex md:items-center md:justify-between mb-6">
    <div class="flex-1 min-w-0">
      <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
        Automations for {project.Name}
      </h2>
      <p class="mt-2 text-sm text-gray-600">
        Manage your automated workflows for this project.
      </p>
    </div>
    <div class="mt-4 flex md:mt-0 md:ml-4">
      <button
        onclick={() => (showCreateAutomationModal = true)}
        class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <svg
          class="-ml-1 mr-2 h-5 w-5"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 6v6m0 0v6m0-6h6m-6 0H6"
          />
        </svg>
        New Automation
      </button>
    </div>
  </div>

  <!-- Automations List -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6">
    {#if automations?.length === 0}
      <div class="text-center py-8">
        <svg
          class="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 10V3L4 14h7v7l9-11h-7z"
          />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">No automations yet</h3>
        <p class="mt-1 text-sm text-gray-500">
          Get started by creating your first automation for this project.
        </p>
      </div>
    {:else}
      <ul role="list" class="divide-y divide-gray-200">
        {#each automations as automation (automation.ID)}
          <li class="py-4 flex justify-between items-center">
            <div>
              <a
                href="/projects/{projectId}/automations/{automation.ID}"
                class="text-lg font-medium text-primary-600 hover:text-primary-800"
              >
                {automation.Name}
              </a>
              {#if automation.Description}
                <p class="text-sm text-gray-500">
                  {automation.Description}
                </p>
              {/if}
              <p class="text-xs text-gray-400 mt-1">
                Created: {formatDate(automation.CreatedAt)}
              </p>
            </div>
            <div class="flex space-x-3">
              <button
                onclick={() => openEditModal(automation)}
                class="text-sm font-medium text-gray-600 hover:text-gray-900"
              >
                Edit
              </button>
              <button
                onclick={() => openDeleteConfirm(automation)}
                class="text-sm font-medium text-red-600 hover:text-red-900"
              >
                Delete
              </button>
              <a
                href="/projects/{projectId}/automations/{automation.ID}"
                class="text-sm font-medium text-primary-600 hover:text-primary-800"
              >
                View <span aria-hidden="true">&rarr;</span>
              </a>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>

<!-- Modals -->
<AutomationFormModal
  bind:open={showCreateAutomationModal}
  onSave={handleSaveAutomation}
  onClose={() => (showCreateAutomationModal = false)}
/>

<AutomationFormModal
  bind:open={showEditAutomationModal}
  automation={selectedAutomation}
  onSave={handleSaveAutomation}
  onClose={() => (showEditAutomationModal = false)}
/>

<ConfirmDeleteModal
  bind:open={showDeleteAutomationConfirm}
  title="Delete Automation"
  message="Are you sure you want to delete '{selectedAutomation?.Name}'? All associated steps, actions, and runs will also be deleted. This action cannot be undone."
  onConfirm={handleDeleteAutomation}
  onCancel={() => (showDeleteAutomationConfirm = false)}
  loading={isDeletingAutomation}
/>
