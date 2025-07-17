<script lang="ts">
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";
  import ProjectFormModal from "$lib/components/ProjectFormModal.svelte";
  import AutomationFormModal from "$lib/components/AutomationFormModal.svelte";
  import ConfirmDeleteModal from "$lib/components/ConfirmDeleteModal.svelte";
  import { formatDate } from "$lib/utils/date";
  import { router,page } from "@inertiajs/svelte";

  type Project = {
    ID: string;
    Name: string;
    Description: string;
    CreatedAt: string;
    UpdatedAt: string;
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
    user: any;
    params: Record<string, string>;
  };

  let { project, automations, params }: Props = $props();

  let showEditProjectModal = $state(false);
  let showCreateAutomationModal = $state(false);
  let showDeleteProjectConfirm = $state(false);
  let isDeletingProject = $state(false);

  const {id:projectId} = params;

  async function handleSaveProject(data: { name: string; description: string }) {
    try {
      const response = await fetch(`/projects/${projectId}`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Project updated successfully");
        // Update the local project prop to reflect changes
        project.Name = result.project.Name;
        project.Description = result.project.Description;
        project.UpdatedAt = result.project.UpdatedAt;
      } else {
        throw result;
      }
    } catch (err: any) {
      console.error("Failed to update project:", err);
      if (err.errors) {
        throw err; // Re-throw to be caught by modal
      } else {
        showErrorToast(err.message || "Failed to update project");
        throw new Error(err.message || "Failed to update project");
      }
    }
  }

  async function handleDeleteProject() {
    isDeletingProject = true;
    try {
      const response = await fetch(`/projects/${projectId}`, {
        method: "DELETE",
      });

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Project deleted successfully");
        router.visit("/projects"); // Redirect to projects list
      } else {
        showErrorToast(result.error || "Failed to delete project");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isDeletingProject = false;
      showDeleteProjectConfirm = false;
    }
  }

  async function handleSaveAutomation(data: {
    name: string;
    description: string;
    config_json: string;
  }) {
    try {
      const response = await fetch(`/projects/${projectId}/automations`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Automation created successfully");
        // Add new automation to the list
        automations = [...automations, result.automation];
      } else {
        throw result;
      }
    } catch (err: any) {
      console.error("Failed to create automation:", err);
      if (err.errors) {
        throw err; // Re-throw to be caught by modal
      } else {
        showErrorToast(err.message || "Failed to create automation");
        throw new Error(err.message || "Failed to create automation");
      }
    }
  }
</script>

<svelte:head>
  <title>{project.Name} - QPlayground</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
  <!-- Project Header -->
  <div class="md:flex md:items-center md:justify-between mb-6">
    <div class="flex-1 min-w-0">
      <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
        {project.Name}
      </h2>
      {#if project.Description}
        <p class="mt-2 text-sm text-gray-600">{project.Description}</p>
      {/if}
      <p class="mt-1 text-sm text-gray-500">
        Created: {formatDate(project.CreatedAt)} | Last Updated: {formatDate(project.UpdatedAt)}
      </p>
    </div>
    <div class="mt-4 flex md:mt-0 md:ml-4">
      <button
        onclick={() => (showEditProjectModal = true)}
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <svg
          class="-ml-1 mr-2 h-5 w-5 text-gray-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
          />
        </svg>
        Edit Project
      </button>
      <button
        onclick={() => (showDeleteProjectConfirm = true)}
        class="ml-3 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
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
            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
          />
        </svg>
        Delete Project
      </button>
    </div>
  </div>

  <!-- Automations Section -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg leading-6 font-medium text-gray-900">Automations</h3>
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
            <div>
              <a
                href="/projects/{projectId}/automations/{automation.ID}"
                class="text-sm font-medium text-gray-600 hover:text-gray-900"
              >
                View Details <span aria-hidden="true">&rarr;</span>
              </a>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>

<!-- Modals -->
<ProjectFormModal
  bind:open={showEditProjectModal}
  project={project}
  onSave={handleSaveProject}
  onClose={() => (showEditProjectModal = false)}
/>

<AutomationFormModal
  bind:open={showCreateAutomationModal}
  onSave={handleSaveAutomation}
  onClose={() => (showCreateAutomationModal = false)}
/>

<ConfirmDeleteModal
  bind:open={showDeleteProjectConfirm}
  title="Delete Project"
  message="Are you sure you want to delete this project? All associated automations will also be deleted. This action cannot be undone."
  onConfirm={handleDeleteProject}
  onCancel={() => (showDeleteProjectConfirm = false)}
  loading={isDeletingProject}
/>
