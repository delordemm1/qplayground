<script lang="ts">
  import { page } from "@inertiajs/svelte";
  import { formatDate } from "$lib/utils/date";

  type Project = {
    ID: string;
    Name: string;
  };

  type Automation = {
    ID: string;
    Name: string;
  };

  type Run = {
    ID: string;
    Status: string;
    StartTime: string;
    EndTime: string;
    ErrorMessage: string;
    CreatedAt: string;
  };

  type Props = {
    project: Project;
    automation: Automation;
    runs: Run[];
    user: any;
  };

  let { project, automation, runs }: Props = $props();

  const projectId = $derived($page.props.params.projectId);
  const automationId = $derived($page.props.params.automationId);
</script>

<svelte:head>
  <title>Runs for {automation.Name} - QPlayground</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
  <!-- Header -->
  <div class="md:flex md:items-center md:justify-between mb-6">
    <div class="flex-1 min-w-0">
      <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
        Runs for {automation.Name}
      </h2>
      <p class="mt-2 text-sm text-gray-600">
        Project: <a href="/projects/{project.ID}" class="text-primary-600 hover:underline"
          >{project.Name}</a
        >
      </p>
    </div>
    <div class="mt-4 flex md:mt-0 md:ml-4">
      <a
        href="/projects/{projectId}/automations/{automationId}"
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
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
            d="M10 19l-7-7m0 0l7-7m-7 7h18"
          />
        </svg>
        Back to Automation
      </a>
    </div>
  </div>

  <!-- Runs List -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6">
    {#if runs.length === 0}
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
            d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">No runs yet</h3>
        <p class="mt-1 text-sm text-gray-500">
          Trigger an automation run from the automation details page to see results here.
        </p>
      </div>
    {:else}
      <ul role="list" class="divide-y divide-gray-200">
        {#each runs as run (run.ID)}
          <li class="py-4 flex justify-between items-center">
            <div>
              <a
                href="/projects/{projectId}/automations/{automationId}/runs/{run.ID}"
                class="text-lg font-medium text-primary-600 hover:text-primary-800"
              >
                Run ID: {run.ID.substring(0, 8)}...
              </a>
              <p class="text-sm text-gray-500">Status: {run.Status}</p>
              <p class="text-xs text-gray-400 mt-1">
                Started: {formatDate(run.StartTime)} | Ended: {run.EndTime ? formatDate(run.EndTime) : 'N/A'}
              </p>
              {#if run.ErrorMessage}
                <p class="text-sm text-red-600 mt-1">Error: {run.ErrorMessage}</p>
              {/if}
            </div>
            <div>
              <a
                href="/projects/{projectId}/automations/{automationId}/runs/{run.ID}"
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
