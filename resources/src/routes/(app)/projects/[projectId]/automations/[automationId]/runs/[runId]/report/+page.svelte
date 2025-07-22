<script lang="ts">
  import { page } from "@inertiajs/svelte";
  import { formatDate } from "$lib/utils/date";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";
  import ImageViewerModal from "$lib/components/ImageViewerModal.svelte";
  import { ChevronDownOutline, ChevronRightOutline, DownloadOutline, TableColumnOutline } from "flowbite-svelte-icons";

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
  
  // Image viewer modal state
  let showImageViewerModal = $state(false);
  let currentImageIndex = $state(0);
  let imageFilesForModal = $state<string[]>([]);
  
  // Expanded state for steps and actions
  let expandedSteps = $state<Set<string>>(new Set());
  let expandedActions = $state<Set<string>>(new Set());

  // Parse logs and organize by steps and actions
  const parsedLogs = $derived.by(() => {
    try {
      const logs = JSON.parse(run.LogsJSON);
      return Array.isArray(logs) ? logs : [];
    } catch (e) {
      console.error("Failed to parse logs JSON:", e);
      return [];
    }
  });

  const parsedOutputFiles = $derived.by(() => {
    try {
      const files = JSON.parse(run.OutputFilesJSON);
      return Array.isArray(files) ? files : [];
    } catch (e) {
      console.error("Failed to parse output files JSON:", e);
      return [];
    }
  });

  // Organize data by steps and actions
  const reportData = $derived.by(() => {
    const stepMap = new Map();
    
    // Process logs to build step and action structure
    parsedLogs.forEach(log => {
      const stepId = log.step_id;
      const actionId = log.action_id;
      
      if (!stepId) return;
      
      // Initialize step if not exists
      if (!stepMap.has(stepId)) {
        stepMap.set(stepId, {
          id: stepId,
          name: log.step_name || 'Unknown Step',
          actions: new Map(),
          totalDuration: 0,
          status: 'success',
          startTime: log.timestamp,
          endTime: log.timestamp,
          logs: []
        });
      }
      
      const step = stepMap.get(stepId);
      step.endTime = log.timestamp;
      step.logs.push(log);
      
      if (log.status === 'failed') {
        step.status = 'failed';
      }
      
      // Process action if actionId exists
      if (actionId) {
        if (!step.actions.has(actionId)) {
          step.actions.set(actionId, {
            id: actionId,
            type: log.action_type || 'Unknown Action',
            duration: log.duration_ms || 0,
            status: log.status || 'success',
            error: log.error || null,
            logs: [],
            outputFiles: []
          });
        }
        
        const action = step.actions.get(actionId);
        action.logs.push(log);
        
        if (log.output_file) {
          action.outputFiles.push(log.output_file);
        }
        
        if (log.status === 'failed') {
          action.status = 'failed';
          action.error = log.error;
        }
      }
    });
    
    // Calculate step durations
    stepMap.forEach(step => {
      if (step.startTime && step.endTime) {
        step.totalDuration = new Date(step.endTime).getTime() - new Date(step.startTime).getTime();
      }
    });
    
    return Array.from(stepMap.values()).sort((a, b) => 
      new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
    );
  });

  // Filter output files to get only images for the modal
  const imageFiles = $derived.by(() => {
    return parsedOutputFiles.filter(fileUrl => getFileType(fileUrl) === "image");
  });

  // Update imageFilesForModal when imageFiles changes
  $effect(() => {
    imageFilesForModal = imageFiles;
  });

  // Function to open image viewer
  function openImageViewer(imageUrl: string) {
    const index = imageFilesForModal.indexOf(imageUrl);
    if (index !== -1) {
      currentImageIndex = index;
      showImageViewerModal = true;
    }
  }

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

  // Toggle functions for expand/collapse
  function toggleStep(stepId: string) {
    if (expandedSteps.has(stepId)) {
      expandedSteps.delete(stepId);
    } else {
      expandedSteps.add(stepId);
    }
    expandedSteps = new Set(expandedSteps);
  }

  function toggleAction(actionId: string) {
    if (expandedActions.has(actionId)) {
      expandedActions.delete(actionId);
    } else {
      expandedActions.add(actionId);
    }
    expandedActions = new Set(expandedActions);
  }

  // Export functions
  function exportToCSV() {
    try {
      const csvRows = [];
      
      // CSV Headers
      csvRows.push([
        'Step ID',
        'Step Name', 
        'Step Duration (ms)',
        'Step Status',
        'Action ID',
        'Action Type',
        'Action Duration (ms)',
        'Action Status',
        'Error Message',
        'Output Files',
        'Timestamp'
      ]);

      // Process each step and action
      reportData.forEach(step => {
        if (step.actions.size === 0) {
          // Step with no actions
          csvRows.push([
            step.id,
            step.name,
            step.totalDuration,
            step.status,
            '',
            '',
            '',
            '',
            '',
            '',
            step.startTime
          ]);
        } else {
          // Step with actions
          Array.from(step.actions.values()).forEach(action => {
            csvRows.push([
              step.id,
              step.name,
              step.totalDuration,
              step.status,
              action.id,
              action.type,
              action.duration,
              action.status,
              action.error || '',
              action.outputFiles.join('; '),
              step.startTime
            ]);
          });
        }
      });

      // Convert to CSV string
      const csvContent = csvRows.map(row => 
        row.map(field => `"${String(field).replace(/"/g, '""')}"`).join(',')
      ).join('\n');

      // Download CSV
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `automation_report_${runId}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      showSuccessToast('Report exported to CSV successfully');
    } catch (error) {
      console.error('Failed to export CSV:', error);
      showErrorToast('Failed to export CSV report');
    }
  }

  function exportToJSON() {
    try {
      const reportJson = {
        run: {
          id: run.ID,
          status: run.Status,
          startTime: run.StartTime,
          endTime: run.EndTime,
          errorMessage: run.ErrorMessage
        },
        automation: {
          id: automation.ID,
          name: automation.Name
        },
        project: {
          id: project.ID,
          name: project.Name
        },
        steps: reportData.map(step => ({
          id: step.id,
          name: step.name,
          totalDuration: step.totalDuration,
          status: step.status,
          startTime: step.startTime,
          endTime: step.endTime,
          actions: Array.from(step.actions.values()).map(action => ({
            id: action.id,
            type: action.type,
            duration: action.duration,
            status: action.status,
            error: action.error,
            outputFiles: action.outputFiles,
            logs: action.logs
          })),
          logs: step.logs
        })),
        summary: {
          totalSteps: reportData.length,
          totalActions: reportData.reduce((sum, step) => sum + step.actions.size, 0),
          totalDuration: run.StartTime && run.EndTime ? 
            new Date(run.EndTime).getTime() - new Date(run.StartTime).getTime() : 0,
          outputFiles: parsedOutputFiles
        }
      };

      const jsonContent = JSON.stringify(reportJson, null, 2);
      const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `automation_report_${runId}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      showSuccessToast('Report exported to JSON successfully');
    } catch (error) {
      console.error('Failed to export JSON:', error);
      showErrorToast('Failed to export JSON report');
    }
  }
</script>

<svelte:head>
  <title>Detailed Report - Run {run.ID.substring(0, 8)}... - QPlayground</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
  <!-- Header -->
  <div class="md:flex md:items-center md:justify-between mb-6">
    <div class="flex-1 min-w-0">
      <h2 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
        Detailed Report - Run {run.ID.substring(0, 8)}...
      </h2>
      <p class="mt-2 text-sm text-gray-600">
        Automation: <a
          href="/projects/{projectId}/automations/{automationId}"
          class="text-primary-600 hover:underline">{automation.Name}</a>
      </p>
      <p class="mt-1 text-sm text-gray-500">
        Project: <a
          href="/projects/{project.ID}"
          class="text-primary-600 hover:underline">{project.Name}</a>
      </p>
    </div>
    <div class="mt-4 flex md:mt-0 md:ml-4 space-x-3">
      <button
        onclick={exportToCSV}
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <TableColumnOutline class="-ml-1 mr-2 h-5 w-5" />
        Export CSV
      </button>
      <button
        onclick={exportToJSON}
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <DownloadOutline class="-ml-1 mr-2 h-5 w-5" />
        Export JSON
      </button>
      <a
        href="/projects/{projectId}/automations/{automationId}/runs/{runId}"
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
        Back to Run
      </a>
    </div>
  </div>

  <!-- Summary Stats -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Summary</h3>
    <dl class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-x-4 gap-y-6">
      <div>
        <dt class="text-sm font-medium text-gray-500">Total Steps</dt>
        <dd class="mt-1 text-2xl font-semibold text-gray-900">{reportData.length}</dd>
      </div>
      <div>
        <dt class="text-sm font-medium text-gray-500">Total Actions</dt>
        <dd class="mt-1 text-2xl font-semibold text-gray-900">
          {reportData.reduce((sum, step) => sum + step.actions.size, 0)}
        </dd>
      </div>
      <div>
        <dt class="text-sm font-medium text-gray-500">Total Duration</dt>
        <dd class="mt-1 text-2xl font-semibold text-gray-900">
          {#if run.StartTime && run.EndTime}
            {formatDuration(new Date(run.EndTime).getTime() - new Date(run.StartTime).getTime())}
          {:else}
            N/A
          {/if}
        </dd>
      </div>
      <div>
        <dt class="text-sm font-medium text-gray-500">Output Files</dt>
        <dd class="mt-1 text-2xl font-semibold text-gray-900">{parsedOutputFiles.length}</dd>
      </div>
    </dl>
  </div>

  <!-- Detailed Step Report -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Step-by-Step Report</h3>
    
    {#if reportData.length === 0}
      <p class="text-sm text-gray-500">No step data available for this run.</p>
    {:else}
      <div class="space-y-4">
        {#each reportData as step, stepIndex (step.id)}
          <div class="border border-gray-200 rounded-lg">
            <!-- Step Header -->
            <button
              onclick={() => toggleStep(step.id)}
              class="w-full px-6 py-4 text-left hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-inset"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center space-x-4">
                  <div class="flex-shrink-0">
                    {#if expandedSteps.has(step.id)}
                      <ChevronDownOutline class="h-5 w-5 text-gray-400" />
                    {:else}
                      <ChevronRightOutline class="h-5 w-5 text-gray-400" />
                    {/if}
                  </div>
                  <div>
                    <h4 class="text-lg font-medium text-gray-900">
                      Step {stepIndex + 1}: {step.name}
                    </h4>
                    <p class="text-sm text-gray-500">
                      {step.actions.size} actions • {formatDuration(step.totalDuration)}
                    </p>
                  </div>
                </div>
                <div class="flex items-center space-x-3">
                  <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadgeClass(step.status)}">
                    {step.status.toUpperCase()}
                  </span>
                  <span class="text-sm text-gray-500">
                    {new Date(step.startTime).toLocaleTimeString()}
                  </span>
                </div>
              </div>
            </button>

            <!-- Step Details (Expandable) -->
            {#if expandedSteps.has(step.id)}
              <div class="border-t border-gray-200 px-6 py-4">
                {#if step.actions.size === 0}
                  <p class="text-sm text-gray-500 italic">No actions recorded for this step.</p>
                {:else}
                  <div class="space-y-3">
                    {#each Array.from(step.actions.values()) as action, actionIndex (action.id)}
                      <div class="border border-gray-100 rounded-md">
                        <!-- Action Header -->
                        <button
                          onclick={() => toggleAction(action.id)}
                          class="w-full px-4 py-3 text-left hover:bg-gray-25 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:ring-inset"
                        >
                          <div class="flex items-center justify-between">
                            <div class="flex items-center space-x-3">
                              <div class="flex-shrink-0">
                                {#if expandedActions.has(action.id)}
                                  <ChevronDownOutline class="h-4 w-4 text-gray-400" />
                                {:else}
                                  <ChevronRightOutline class="h-4 w-4 text-gray-400" />
                                {/if}
                              </div>
                              <div>
                                <h5 class="text-sm font-medium text-gray-900">
                                  Action {actionIndex + 1}: {action.type}
                                </h5>
                                {#if action.error}
                                  <p class="text-xs text-red-600">{action.error}</p>
                                {/if}
                              </div>
                            </div>
                            <div class="flex items-center space-x-3">
                              <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {getStatusBadgeClass(action.status)}">
                                {action.status.toUpperCase()}
                              </span>
                              <span class="text-xs text-gray-500">
                                {formatDuration(action.duration)}
                              </span>
                            </div>
                          </div>
                        </button>

                        <!-- Action Details (Expandable) -->
                        {#if expandedActions.has(action.id)}
                          <div class="border-t border-gray-100 px-4 py-3 bg-gray-25">
                            <!-- Action Logs -->
                            {#if action.logs.length > 0}
                              <div class="mb-3">
                                <h6 class="text-xs font-medium text-gray-700 mb-2">Logs:</h6>
                                <div class="space-y-1">
                                  {#each action.logs as log}
                                    <div class="text-xs text-gray-600 bg-gray-100 p-2 rounded">
                                      <div class="flex justify-between items-start">
                                        <span>{log.message || 'Action executed'}</span>
                                        <span class="text-gray-400">{new Date(log.timestamp).toLocaleTimeString()}</span>
                                      </div>
                                      {#if log.error}
                                        <div class="mt-1 text-red-600 font-medium">Error: {log.error}</div>
                                      {/if}
                                    </div>
                                  {/each}
                                </div>
                              </div>
                            {/if}

                            <!-- Action Output Files -->
                            {#if action.outputFiles.length > 0}
                              <div>
                                <h6 class="text-xs font-medium text-gray-700 mb-2">Output Files:</h6>
                                <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
                                  {#each action.outputFiles as fileUrl}
                                    <div class="border border-gray-200 rounded p-2 hover:shadow-sm transition-shadow">
                                      <div class="flex items-center space-x-2">
                                        <div class="flex-shrink-0">
                                          {#if getFileType(fileUrl) === "image"}
                                            <svg class="h-4 w-4 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                            </svg>
                                          {:else}
                                            <svg class="h-4 w-4 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                                            </svg>
                                          {/if}
                                        </div>
                                        <div class="flex-1 min-w-0">
                                          <p class="text-xs font-medium text-gray-900 truncate">
                                            {fileUrl.split("/").pop() || "Unknown File"}
                                          </p>
                                        </div>
                                      </div>
                                      <div class="mt-2">
                                        {#if getFileType(fileUrl) === "image"}
                                          <button
                                            onclick={() => openImageViewer(fileUrl)}
                                            class="text-xs text-primary-600 hover:text-primary-800 font-medium"
                                          >
                                            View Image
                                          </button>
                                        {:else}
                                          <a
                                            href={fileUrl}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            class="text-xs text-primary-600 hover:text-primary-800 font-medium"
                                          >
                                            Open File
                                          </a>
                                        {/if}
                                      </div>
                                    </div>
                                  {/each}
                                </div>
                              </div>
                            {/if}
                          </div>
                        {/if}
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Image Viewer Modal -->
<ImageViewerModal
  bind:open={showImageViewerModal}
  imageUrls={imageFilesForModal}
  startIndex={currentImageIndex}
  onClose={() => showImageViewerModal = false}
/>

<style>
  .bg-gray-25 {
    background-color: #fafafa;
  }
</style>