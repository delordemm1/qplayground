```svelte
<script lang="ts">
  import { showErrorToast, showSuccessToast } from "$lib/utils/toast";
  import ProjectFormModal from "$lib/components/ProjectFormModal.svelte";
  import { formatDate } from "$lib/utils/date";

  type Project = {
    ID: string;
    Name: string;
    Description: string;
    CreatedAt: string;
  };

  let { projects = [], user }: { projects: Project[]; user: any } = $props();

  let showCreateModal = $state(false);

  async function handleSaveProject(data: { name: string; description: string }) {
    try {
      const response = await fetch("/projects", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Project created successfully");
        projects = [...projects, result.project]; // Add new project to the list
      } else {
        throw result;
      }
    } catch (err: any) {
      console.error("Failed to create project:", err);
      if (err.errors) {
        throw err; // Re-throw to be caught by modal
      } else {
        showErrorToast(err.message || "Failed to create project");
        throw new Error(err.message || "Failed to create project");
      }
    }
  }
</script>

<svelte:head>
  <title>Projects - QPlayground</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
  <!-- Header -->
  <div class="md:flex md:items-center md:justify-between">
    <div class="flex-1 min-w-0">
      <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
        Projects
      </h2>
    </div>
    <div class="mt-4 flex md:mt-0 md:ml-4">
      <button
        onclick={() => (showCreateModal = true)}
        class="ml-3 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
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
        New Project
      </button>
    </div>
  </div>

  <!-- Projects Grid -->
  <div class="mt-8">
    {#if projects?.length === 0}
      <div class="text-center">
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
            d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
          />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">No projects</h3>
        <p class="mt-1 text-sm text-gray-500">Get started by creating a new project.</p>
        <div class="mt-6">
          <button
            onclick={() => (showCreateModal = true)}
            class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
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
            New Project
          </button>
        </div>
      </div>
    {:else}
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {#each projects as project (project.ID)}
          <div
            class="bg-white overflow-hidden shadow rounded-lg hover:shadow-md transition-shadow"
          >
            <div class="p-6">
              <div class="flex items-center">
                <div class="flex-shrink-0">
                  <svg
                    class="h-8 w-8 text-primary-600"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                    />
                  </svg>
                </div>
                <div class="ml-4 flex-1">
                  <h3 class="text-lg font-medium text-gray-900">
                    <a
                      href="/projects/{project.ID}"
                      class="hover:text-primary-600"
                    >
                      {project.Name}
                    </a>
                  </h3>
                  {#if project.Description}
                    <p class="text-sm text-gray-500 mt-1">
                      {project.Description}
                    </p>
                  {/if}
                </div>
              </div>
              <div class="mt-4">
                <div class="flex items-center justify-between text-sm text-gray-500">
                  <span>0 automations</span>
                  <span>Created {formatDate(project.CreatedAt)}</span>
                </div>
              </div>
            </div>
            <div class="bg-gray-50 px-6 py-3">
              <div class="flex justify-between">
                <a
                  href="/projects/{project.ID}"
                  class="text-sm font-medium text-primary-700 hover:text-primary-900"
                >
                  View Project <span aria-hidden="true">&rarr;</span>
                </a>
                <!-- Settings button can be added here if needed -->
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Create Project Modal -->
<ProjectFormModal
  bind:open={showCreateModal}
  onSave={handleSaveProject}
  onClose={() => (showCreateModal = false)}
/>
```