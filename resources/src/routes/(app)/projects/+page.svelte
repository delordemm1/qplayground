<script lang="ts">
  import { showErrorToast, showSuccessToast } from "$lib/utils/toast";

  let { projects = [], user } = $props();
  
  let showCreateModal = $state(false);
  let isLoading = $state(false);
  let errors = $state<Record<string, string>>({});
  
  let newProject = $state({
    name: "",
    description: ""
  });

  async function createProject(e) {
    e.preventDefault();
    if (!newProject.name.trim()) {
      errors.name = "Project name is required";
      return;
    }

    isLoading = true;
    errors = {};

    try {
      const response = await fetch("/projects", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(newProject),
      });

      const data = await response.json();

      if (response.ok) {
        showSuccessToast("Project created successfully");
        showCreateModal = false;
        newProject = { name: "", description: "" };
        // Refresh the page to show the new project
        window.location.reload();
      } else {
        if (data.errors) {
          errors = data.errors;
        } else {
          showErrorToast(data.error || "Failed to create project");
        }
      }
    } catch (error) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isLoading = false;
    }
  }

  function closeModal() {
    showCreateModal = false;
    newProject = { name: "", description: "" };
    errors = {};
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
        onclick={() => showCreateModal = true}
        class="ml-3 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <svg class="-ml-1 mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        New Project
      </button>
    </div>
  </div>

  <!-- Projects Grid -->
  <div class="mt-8">
    {#if projects.length === 0}
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
            onclick={() => showCreateModal = true}
            class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
          >
            <svg class="-ml-1 mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            New Project
          </button>
        </div>
      </div>
    {:else}
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {#each projects as project}
          <div class="bg-white overflow-hidden shadow rounded-lg hover:shadow-md transition-shadow">
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
                    <a href="/projects/{project.ID}/automations" class="hover:text-primary-600">
                      {project.Name}
                    </a>
                  </h3>
                  {#if project.Description}
                    <p class="text-sm text-gray-500 mt-1">{project.Description}</p>
                  {/if}
                </div>
              </div>
              <div class="mt-4">
                <div class="flex items-center justify-between text-sm text-gray-500">
                  <span>0 automations</span>
                  <span>Created {new Date(project.CreatedAt).toLocaleDateString()}</span>
                </div>
              </div>
            </div>
            <div class="bg-gray-50 px-6 py-3">
              <div class="flex justify-between">
                <a
                  href="/projects/{project.ID}/automations"
                  class="text-sm font-medium text-primary-700 hover:text-primary-900"
                >
                  View automations
                </a>
                <button class="text-sm font-medium text-gray-500 hover:text-gray-700">
                  Settings
                </button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Create Project Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
    <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
      <div class="mt-3">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-medium text-gray-900">Create New Project</h3>
          <button
            onclick={closeModal}
            class="text-gray-400 hover:text-gray-600"
          >
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <form onsubmit={createProject} class="mt-6 space-y-4">
          <div>
            <label for="name" class="block text-sm font-medium text-gray-700">
              Project Name
            </label>
            <input
              id="name"
              type="text"
              required
              bind:value={newProject.name}
              class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
              class:border-red-300={errors.name}
              placeholder="Enter project name"
            />
            {#if errors.name}
              <p class="mt-2 text-sm text-red-600">{errors.name}</p>
            {/if}
          </div>

          <div>
            <label for="description" class="block text-sm font-medium text-gray-700">
              Description (optional)
            </label>
            <textarea
              id="description"
              rows="3"
              bind:value={newProject.description}
              class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-primary-500 focus:border-primary-500 sm:text-sm"
              placeholder="Enter project description"
            ></textarea>
          </div>

          <div class="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              onclick={closeModal}
              class="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isLoading}
              class="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50"
            >
              {#if isLoading}
                Creating...
              {:else}
                Create Project
              {/if}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
{/if}