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
    LogsJSON: string;
    OutputFilesJSON: string;
    ErrorMessage: string;
    CreatedAt: string;
  };

  type Props = {
    project: Project;
    automation: Automation;
    run: Run;
    user: any;
  };

  let { project, automation, run }: Props = $props();

  const projectId = $derived($page.props.params.projectId);
  const automationId = $derived($page.props.params.automationId);
  const runId = $derived($page.props.params.id);

  let parsedLogs = $derived.by(() => {
    try {
      return JSON.parse(run.LogsJSON);
    } catch (e) {
      console.error("Failed to parse logs JSON:", e);
      return [];
    }
  });

  let parsedOutputFiles = $derived.by(() => {
    try {
      return JSON.parse(run.OutputFilesJSON);
    } catch (e) {
      console.error("Failed to parse output files JSON:", e);
      return [];
    }
  });
</script>

<svelte:head>
  <title>Run {run.ID.substring(0, 8)}... - QPlayground</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
  <!-- Header -->
  <div class="md:flex md:items-center md:justify-between mb-6">
    <div class="flex-1 min-w-0">
      <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
        Automation Run: {run.ID.substring(0, 8)}...
      </h2>
      <p class="mt-2 text-sm text-gray-600">
        Automation: <a
          href="/projects/{projectId}/automations/{automationId}"
          class="text-primary-600 hover:underline"
          >{automation.Name}</a
        >
      </p>
      <p class="mt-1 text-sm text-gray-500">
        Project: <a href="/projects/{project.ID}" class="text-primary-600 hover:underline"
          >{project.Name}</a
        >
      </p>
    </div>
    <div class="mt-4 flex md:mt-0 md:ml-4">
      <a
        href="/projects/{projectId}/automations/{automationId}/runs"
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
        Back to All Runs
      </a>
    </div>
  </div>

  <!-- Run Details -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Details</h3>
    <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-8">
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Status</dt>
        <dd class="mt-1 text-sm text-gray-900">{run.Status}</dd>
      </div>
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Started At</dt>
        <dd class="mt-1 text-sm text-gray-900">{formatDate(run.StartTime)}</dd>
      </div>
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Ended At</dt>
        <dd class="mt-1 text-sm text-gray-900">
          {run.EndTime ? formatDate(run.EndTime) : "Still running..."}
        </dd>
      </div>
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Duration</dt>
        <dd class="mt-1 text-sm text-gray-900">
          {#if run.StartTime && run.EndTime}
            {((new Date(run.EndTime).getTime() - new Date(run.StartTime).getTime()) / 1000).toFixed(2)} seconds
          {:else}
            N/A
          {/if}
        </dd>
      </div>
      {#if run.ErrorMessage}
        <div class="sm:col-span-2">
          <dt class="text-sm font-medium text-red-500">Error Message</dt>
          <dd class="mt-1 text-sm text-red-900">{run.ErrorMessage}</dd>
        </div>
      {/if}
    </dl>
  </div>

  <!-- Logs -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Logs</h3>
    {#if parsedLogs.length === 0}
      <p class="text-sm text-gray-500">No logs available for this run.</p>
    {:else}
      <div class="bg-gray-100 p-4 rounded-md text-sm font-mono overflow-auto max-h-96">
        {#each parsedLogs as logEntry}
          <div class="mb-2 pb-2 border-b border-gray-200 last:border-b-0 last:mb-0 last:pb-0">
            <p class="text-gray-700">
              <span class="text-gray-500">{logEntry.timestamp}</span> -
              <span
                class:text-green-600={logEntry.status === 'success'}
                class:text-red-600={logEntry.status === 'failed'}
              >
                {logEntry.status.toUpperCase()}</span
              >: {logEntry.type} (Step: {logEntry.step_name})
            </p>
            {#if logEntry.error}
              <p class="text-red-500 ml-4">Error: {logEntry.error}</p>
            {/if}
            {#if logEntry.output_file}
              <p class="text-blue-500 ml-4">Output: <a href={logEntry.output_file} target="_blank" rel="noopener noreferrer" class="underline">View File</a></p>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Output Files -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Output Files</h3>
    {#if parsedOutputFiles.length === 0}
      <p class="text-sm text-gray-500">No output files generated for this run.</p>
    {:else}
      <ul role="list" class="divide-y divide-gray-200">
        {#each parsedOutputFiles as fileUrl}
          <li class="py-3">
            <a href={fileUrl} target="_blank" rel="noopener noreferrer" class="text-primary-600 hover:underline text-sm">
              {fileUrl}
            </a>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>
