<script lang="ts">
  import { page } from "@inertiajs/svelte";
  import { formatDate } from "$lib/utils/date";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";

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
  const runId = $derived($page.props.params.runId);
  
  let isCancelling = $state(false);
  
  async function handleCancelRun() {
    if (isCancelling) return;
    
    isCancelling = true;
    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${automationId}/runs/${runId}/cancel`,
        {
          method: "POST",
        }
      );

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Automation run cancelled successfully");
        // Refresh the page to show updated status
        window.location.reload();
      } else {
        showErrorToast(result.error || "Failed to cancel automation run");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isCancelling = false;
    }
  }
  let parsedLogs = $derived.by(() => {
    try {
      const logs = JSON.parse(run.LogsJSON);
      return Array.isArray(logs) ? logs : [];
    } catch (e) {
      console.error("Failed to parse logs JSON:", e);
      return [];
    }
  });

  let parsedOutputFiles = $derived.by(() => {
    try {
      const files = JSON.parse(run.OutputFilesJSON);
      return Array.isArray(files) ? files : [];
    } catch (e) {
      console.error("Failed to parse output files JSON:", e);
      return [];
    }
  });

  // Helper function to get status badge styling
  function getStatusBadgeClass(status: string): string {
    switch (status?.toLowerCase()) {
      case "success":
        return "bg-green-100 text-green-800";
      case "failed":
      case "error":
        return "bg-red-100 text-red-800";
      case "running":
        return "bg-blue-100 text-blue-800";
      case "pending":
        return "bg-yellow-100 text-yellow-800";
      case "queued":
        return "bg-purple-100 text-purple-800";
      case "cancelled":
        return "bg-gray-100 text-gray-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  }

  // Helper function to format duration
  function formatDuration(durationMs: number): string {
    if (durationMs < 1000) {
      return `${durationMs}ms`;
    } else if (durationMs < 60000) {
      return `${(durationMs / 1000).toFixed(2)}s`;
    } else {
      const minutes = Math.floor(durationMs / 60000);
      const seconds = ((durationMs % 60000) / 1000).toFixed(2);
      return `${minutes}m ${seconds}s`;
    }
  }

  // Helper function to get file type from URL
  function getFileType(url: string): string {
    const extension = url.split(".").pop()?.toLowerCase();
    switch (extension) {
      case "png":
      case "jpg":
      case "jpeg":
      case "gif":
      case "webp":
        return "image";
      case "pdf":
        return "pdf";
      case "json":
        return "json";
      case "txt":
        return "text";
      default:
        return "file";
    }
  }
</script>

<svelte:head>
  <title>Run {run.ID.substring(0, 8)}... - QPlayground</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
  <!-- Header -->
  <div class="md:flex md:items-center md:justify-between mb-6">
    <div class="flex-1 min-w-0">
      <h2
        class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate"
      >
        Automation Run: {run.ID.substring(0, 8)}...
      </h2>
      <p class="mt-2 text-sm text-gray-600">
        Automation: <a
          href="/projects/{projectId}/automations/{automationId}"
          class="text-primary-600 hover:underline">{automation.Name}</a
        >
      </p>
      <p class="mt-1 text-sm text-gray-500">
        Project: <a
          href="/projects/{project.ID}"
          class="text-primary-600 hover:underline">{project.Name}</a
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
      {#if run.Status === "running" || run.Status === "pending"}
        <button
          onclick={handleCancelRun}
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
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
          Cancel Run
        </button>
        {:else if run.Status === "queued"}
        <span class="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-500 bg-gray-100">
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
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          Queued
        </span>
      {/if}
    </div>
  </div>

  <!-- Run Details -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Details</h3>
    <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-8">
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Status</dt>
        <dd class="mt-1">
          <span
            class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadgeClass(
              run.Status
            )}"
          >
            {run.Status}
          </span>
        </dd>
      </div>
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Started At</dt>
        <dd class="mt-1 text-sm text-gray-900">
          {run.StartTime ? formatDate(run.StartTime) : "Not started"}
        </dd>
      </div>
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Ended At</dt>
        <dd class="mt-1 text-sm text-gray-900">
          {run.EndTime
            ? formatDate(run.EndTime)
            : run.Status === "running"
              ? "Still running..."
              : "N/A"}
        </dd>
      </div>
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Duration</dt>
        <dd class="mt-1 text-sm text-gray-900">
          {#if run.StartTime && run.EndTime}
            {(
              (new Date(run.EndTime).getTime() -
                new Date(run.StartTime).getTime()) /
              1000
            ).toFixed(2)} seconds
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
      <div class="space-y-3 max-h-96 overflow-auto">
        {#each parsedLogs as logEntry, index (index)}
          <div class="border border-gray-200 rounded-lg p-4 bg-gray-50">
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center space-x-3">
                <span
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadgeClass(
                    logEntry.status
                  )}"
                >
                  {logEntry.status?.toUpperCase() || "UNKNOWN"}
                </span>
                <span class="text-sm font-medium text-gray-900">
                  {logEntry.action_type || "Unknown Action"}
                </span>
                {#if logEntry.loop_index !== undefined}
                  <span
                    class="text-xs text-gray-500 bg-gray-200 px-2 py-1 rounded"
                  >
                    Run #{logEntry.loop_index}
                  </span>
                {/if}
              </div>
              <div class="flex items-center space-x-2 text-xs text-gray-500">
                {#if logEntry.duration_ms}
                  <span>{formatDuration(logEntry.duration_ms)}</span>
                {/if}
                {#if logEntry.timestamp}
                  <span
                    >{new Date(logEntry.timestamp).toLocaleTimeString()}</span
                  >
                {/if}
              </div>
            </div>

            {#if logEntry.step_name}
              <p class="text-sm text-gray-600 mb-2">
                <span class="font-medium">Step:</span>
                {logEntry.step_name}
              </p>
            {/if}

            {#if logEntry.error}
              <div class="bg-red-50 border border-red-200 rounded-md p-3 mb-2">
                <p class="text-sm text-red-800">
                  <span class="font-medium">Error:</span>
                  {logEntry.error}
                </p>
              </div>
            {/if}

            {#if logEntry.output_file}
              <div class="bg-blue-50 border border-blue-200 rounded-md p-3">
                <p class="text-sm text-blue-800">
                  <span class="font-medium">Output File:</span>
                  <a
                    href={logEntry.output_file}
                    target="_blank"
                    rel="noopener noreferrer"
                    class="ml-2 underline hover:text-blue-900"
                  >
                    View File →
                  </a>
                </p>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Output Files -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">
      Output Files
    </h3>
    {#if parsedOutputFiles.length === 0}
      <p class="text-sm text-gray-500">
        No output files generated for this run.
      </p>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {#each parsedOutputFiles as fileUrl, index (index)}
          <div
            class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
          >
            <div class="flex items-center space-x-3">
              <div class="flex-shrink-0">
                {#if getFileType(fileUrl) === "image"}
                  <svg
                    class="h-8 w-8 text-green-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                    />
                  </svg>
                {:else if getFileType(fileUrl) === "pdf"}
                  <svg
                    class="h-8 w-8 text-red-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
                    />
                  </svg>
                {:else}
                  <svg
                    class="h-8 w-8 text-blue-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                {/if}
              </div>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-gray-900 truncate">
                  {fileUrl.split("/").pop() || "Unknown File"}
                </p>
                <p class="text-xs text-gray-500 capitalize">
                  {getFileType(fileUrl)} file
                </p>
              </div>
            </div>
            <div class="mt-3">
              <a
                href={fileUrl}
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-primary-700 bg-primary-100 hover:bg-primary-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors"
              >
                <svg
                  class="mr-2 h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                  />
                </svg>
                Open File
              </a>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
