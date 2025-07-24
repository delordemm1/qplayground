<script lang="ts">
  import { page } from "@inertiajs/svelte";
  import { formatDate } from "$lib/utils/date";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";
  import ImageViewerModal from "$lib/components/ImageViewerModal.svelte";
  import RunPerformanceChart from "$lib/components/RunPerformanceChart.svelte";
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
  
  let isCancelling = $state(false);
  let liveStatus = $state(run.Status);
  let liveProgress = $state(0);
  let currentStep = $state("");
  let liveLogs = $state<any[]>([]);
  let liveOutputFiles = $state<string[]>([]);
  let eventSource: EventSource | null = null;
  
  // Image viewer modal state
  let showImageViewerModal = $state(false);
  let currentImageIndex = $state(0);
  let imageFilesForModal = $state<string[]>([]);
  
  // Expanded state for steps and actions
  let expandedSteps = $state<Set<string>>(new Set());
  let expandedActions = $state<Set<string>>(new Set());
  
  // State for step image viewer
  let showStepImageViewerModal = $state(false);
  let stepImageFiles = $state<string[]>([]);
  
  // Initialize SSE connection for real-time updates
  $effect(() => {
    if (typeof window !== 'undefined') {
      connectToSSE();
    }
    
    return () => {
      if (eventSource) {
        eventSource.close();
      }
    };
  });
  
  function connectToSSE() {
    const sseUrl = `/projects/${projectId}/automations/${automationId}/runs/${runId}/events`;
    eventSource = new EventSource(sseUrl);
    
    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        handleSSEMessage(data);
      } catch (error) {
        console.error('Failed to parse SSE message:', error);
      }
    };
    
    eventSource.onerror = (error) => {
      console.error('SSE connection error:', error);
      // Attempt to reconnect after 5 seconds
      setTimeout(() => {
        if (eventSource?.readyState === EventSource.CLOSED) {
          connectToSSE();
        }
      }, 5000);
    };
  }
  
  function handleSSEMessage(data: any) {
    switch (data.type) {
      case 'status':
        liveStatus = data.status;
        if (data.status === 'cancelled' || data.status === 'completed' || data.status === 'failed') {
          // Close SSE connection for final states
          if (eventSource) {
            eventSource.close();
            eventSource = null;
          }
        }
        break;
        
      case 'step':
        currentStep = data.stepName;
        liveProgress = data.progress || 0;
        break;
        
      case 'log':
        liveLogs = [...liveLogs, {
          timestamp: data.timestamp,
          stepName: data.stepName,
          actionType: data.actionType,
          message: data.message,
          duration: data.duration,
          status: 'success'
        }];
        break;
        
      case 'error':
        liveLogs = [...liveLogs, {
          timestamp: data.timestamp,
          stepName: data.stepName,
          actionType: data.actionType,
          error: data.error,
          status: 'failed'
        }];
        break;
        
      case 'output':
        if (data.outputFile && !liveOutputFiles.includes(data.outputFile)) {
          liveOutputFiles = [...liveOutputFiles, data.outputFile];
        }
        break;
        
      case 'complete':
        liveStatus = data.status;
        if (data.data?.outputFiles) {
          liveOutputFiles = data.data.outputFiles;
        }
        // Close SSE connection
        if (eventSource) {
          eventSource.close();
          eventSource = null;
        }
        break;
    }
  }
  
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
      // Combine initial logs with live logs
      const initialLogs = JSON.parse(run.LogsJSON);
      const logs = Array.isArray(initialLogs) ? initialLogs : [];
      return [...logs, ...liveLogs];
    } catch (e) {
      console.error("Failed to parse logs JSON:", e);
      return liveLogs;
    }
  });

  let parsedOutputFiles = $derived.by(() => {
    try {
      // Combine initial files with live files
      const initialFiles = JSON.parse(run.OutputFilesJSON);
      const files = Array.isArray(initialFiles) ? initialFiles : [];
      return [...files, ...liveOutputFiles];
    } catch (e) {
      console.error("Failed to parse output files JSON:", e);
      return liveOutputFiles;
    }
  });

  // Organize data by steps and actions for detailed view
  const reportData = $derived.by(() => {
    const stepMap = new Map();
    
    // Process logs to build step and action structure
    parsedLogs.forEach(log => {
      const stepId = log.step_id;
      const actionId = log.action_id;
      const loopIndex = log.loop_index || 0;
      
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
          logs: [],
          stepImageFiles: [],
          concurrentUsers: new Set(),
          loopIndexes: new Set(),
          totalOutputFiles: 0, // Initialize new property
          totalFailures: 0,    // Initialize new property
        });
      }
      
      const step = stepMap.get(stepId);
      step.endTime = log.timestamp;
      step.logs.push(log);
      step.concurrentUsers.add(loopIndex);
      step.loopIndexes.add(loopIndex);
      
      if (log.status === 'failed') {
        step.status = 'failed';
      }
      
      // Process action if actionId exists
      if (actionId) {
        const actionKey = `${actionId}-${loopIndex}`;
        if (!step.actions.has(actionKey)) {
          step.actions.set(actionKey, {
            id: actionId,
            loopIndex: loopIndex,
            type: log.action_type || 'Unknown Action',
            parentActionId: log.parent_action_id,
            duration: log.duration_ms || 0,
            status: log.status || 'success',
            error: log.error || null,
            logs: [],
            outputFiles: []
          });
        }
        
        const action = step.actions.get(actionKey);
        action.logs.push(log);
        
        if (log.output_file) {
          action.outputFiles.push(log.output_file);
          step.totalOutputFiles++; // Increment step's total output files
          // Add image files to step-level collection
          if (getFileType(log.output_file) === "image") {
            step.stepImageFiles.push(log.output_file);
          }
        }
        
        if (log.status === 'failed') {
          action.status = 'failed';
          action.error = log.error;
          step.totalFailures++; // Increment step's total failures
        }
      }
    });
    
    // Calculate step durations and convert Sets to arrays
    stepMap.forEach(step => {
      if (step.startTime && step.endTime) {
        step.totalDuration = new Date(step.endTime).getTime() - new Date(step.startTime).getTime();
      }
      step.concurrentUsers = Array.from(step.concurrentUsers).sort((a, b) => a - b);
      step.loopIndexes = Array.from(step.loopIndexes).sort((a, b) => a - b);
    });
    
    return Array.from(stepMap.values()).sort((a, b) => 
      new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
    );
  });
  
  // Performance metrics for visualization
  const performanceMetrics = $derived.by(() => {
    const stepMetrics = new Map();
    const runMetrics = new Map(); // Group by loop index
    
    parsedLogs.forEach(log => {
      const stepName = log.step_name || 'Unknown Step';
      const loopIndex = log.loop_index || 0;
      const duration = log.duration_ms || 0;
      const status = log.status || 'success';
      
      // Step-level metrics
      if (!stepMetrics.has(stepName)) {
        stepMetrics.set(stepName, {
          name: stepName,
          durations: [],
          failures: 0,
          totalRuns: 0
        });
      }
      
      const stepMetric = stepMetrics.get(stepName);
      stepMetric.durations.push(duration);
      stepMetric.totalRuns++;
      if (status === 'failed') {
        stepMetric.failures++;
      }
      
      // Run-level metrics (by loop index)
      if (!runMetrics.has(loopIndex)) {
        runMetrics.set(loopIndex, {
          loopIndex,
          steps: new Map(),
          totalDuration: 0,
          status: 'success'
        });
      }
      
      const runMetric = runMetrics.get(loopIndex);
      if (!runMetric.steps.has(stepName)) {
        runMetric.steps.set(stepName, { duration: 0, status: 'success' });
      }
      
      const runStepMetric = runMetric.steps.get(stepName);
      runStepMetric.duration += duration;
      if (status === 'failed') {
        runStepMetric.status = 'failed';
        runMetric.status = 'failed';
      }
      
      runMetric.totalDuration += duration;
    });
    
    // Calculate averages and prepare chart data
    const stepAverages = Array.from(stepMetrics.values()).map(metric => ({
      name: metric.name,
      averageDuration: metric.durations.reduce((sum, d) => sum + d, 0) / metric.durations.length,
      failureRate: (metric.failures / metric.totalRuns) * 100,
      totalRuns: metric.totalRuns
    }));
    
    const runData = Array.from(runMetrics.values()).sort((a, b) => a.loopIndex - b.loopIndex);
    
    return {
      stepAverages,
      runData,
      totalRuns: runMetrics.size,
      overallFailureRate: runData.filter(r => r.status === 'failed').length / runData.length * 100
    };
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

  // Function to open step image viewer
  function openStepImageViewer(images: string[]) {
    stepImageFiles = images;
    currentImageIndex = 0;
    showStepImageViewerModal = true;
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
        'Parent Action ID',
        'Action Type',
        'Action Duration (ms)',
        'Action Status',
        'Error Message',
        'Output Files',
        'Loop Index',
        'Local Loop Index',
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
              action.parentActionId || '',
              action.type,
              action.duration,
              action.status,
              action.error || '',
              action.outputFiles.join('; '),
              action.loopIndex,
              '', // Local loop index would need to be tracked separately
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
          concurrentUsers: step.concurrentUsers,
          actions: Array.from(step.actions.values()).map(action => ({
            id: action.id,
            parentActionId: action.parentActionId,
            type: action.type,
            duration: action.duration,
            status: action.status,
            error: action.error,
            loopIndex: action.loopIndex,
            outputFiles: action.outputFiles,
            logs: action.logs
          })),
          logs: step.logs
        })),
        performanceMetrics: performanceMetrics,
        summary: {
          totalSteps: reportData.length,
          totalActions: reportData.reduce((sum, step) => sum + step.actions.size, 0),
          totalDuration: run.StartTime && run.EndTime ? 
            new Date(run.EndTime).getTime() - new Date(run.StartTime).getTime() : 0,
          totalConcurrentUsers: Math.max(...reportData.map(step => step.concurrentUsers.length), 0),
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
  
  function exportToHTML() {
    try {
      // Get the current page content
      const reportContent = document.querySelector('.report-container')?.innerHTML || '';
      
      // Create a complete HTML document
      const htmlContent = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Automation Report - ${automation.Name}</title>
    <style>
        /* Tailwind CSS Reset and Base Styles */
        *, ::before, ::after { box-sizing: border-box; border-width: 0; border-style: solid; border-color: #e5e7eb; }
        html { line-height: 1.5; -webkit-text-size-adjust: 100%; -moz-tab-tab-size: 4; tab-size: 4; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif; }
        body { margin: 0; line-height: inherit; }
        
        /* Utility Classes */
        .px-4 { padding-left: 1rem; padding-right: 1rem; }
        .py-6 { padding-top: 1.5rem; padding-bottom: 1.5rem; }
        .mb-6 { margin-bottom: 1.5rem; }
        .mb-4 { margin-bottom: 1rem; }
        .text-2xl { font-size: 1.5rem; line-height: 2rem; }
        .text-lg { font-size: 1.125rem; line-height: 1.75rem; }
        .text-sm { font-size: 0.875rem; line-height: 1.25rem; }
        .text-xs { font-size: 0.75rem; line-height: 1rem; }
        .font-bold { font-weight: 700; }
        .font-semibold { font-weight: 600; }
        .font-medium { font-weight: 500; }
        .text-gray-900 { color: rgb(17 24 39); }
        .text-gray-700 { color: rgb(55 65 81); }
        .text-gray-600 { color: rgb(75 85 99); }
        .text-gray-500 { color: rgb(107 114 128); }
        .bg-white { background-color: rgb(255 255 255); }
        .bg-gray-50 { background-color: rgb(249 250 251); }
        .bg-green-100 { background-color: rgb(220 252 231); }
        .bg-red-100 { background-color: rgb(254 226 226); }
        .bg-blue-100 { background-color: rgb(219 234 254); }
        .bg-yellow-100 { background-color: rgb(254 249 195); }
        .text-green-800 { color: rgb(22 101 52); }
        .text-red-800 { color: rgb(153 27 27); }
        .text-blue-800 { color: rgb(30 64 175); }
        .text-yellow-800 { color: rgb(146 64 14); }
        .shadow { box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1); }
        .rounded-lg { border-radius: 0.5rem; }
        .rounded-md { border-radius: 0.375rem; }
        .rounded-full { border-radius: 9999px; }
        .border { border-width: 1px; }
        .border-gray-200 { border-color: rgb(229 231 235); }
        .p-6 { padding: 1.5rem; }
        .p-4 { padding: 1rem; }
        .px-6 { padding-left: 1.5rem; padding-right: 1.5rem; }
        .py-4 { padding-top: 1rem; padding-bottom: 1rem; }
        .px-2\\.5 { padding-left: 0.625rem; padding-right: 0.625rem; }
        .py-0\\.5 { padding-top: 0.125rem; padding-bottom: 0.125rem; }
        .space-y-4 > :not([hidden]) ~ :not([hidden]) { margin-top: 1rem; }
        .space-y-6 > :not([hidden]) ~ :not([hidden]) { margin-top: 1.5rem; }
        .grid { display: grid; }
        .grid-cols-1 { grid-template-columns: repeat(1, minmax(0, 1fr)); }
        .grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
        .grid-cols-4 { grid-template-columns: repeat(4, minmax(0, 1fr)); }
        .gap-4 { gap: 1rem; }
        .gap-6 { gap: 1.5rem; }
        .flex { display: flex; }
        .items-center { align-items: center; }
        .justify-between { justify-content: space-between; }
        .space-x-3 > :not([hidden]) ~ :not([hidden]) { margin-left: 0.75rem; }
        .space-x-4 > :not([hidden]) ~ :not([hidden]) { margin-left: 1rem; }
        .inline-flex { display: inline-flex; }
        .mt-1 { margin-top: 0.25rem; }
        .mt-2 { margin-top: 0.5rem; }
        .overflow-hidden { overflow: hidden; }
        .cursor-pointer { cursor: pointer; }
        .hover\\:bg-gray-50:hover { background-color: rgb(249 250 251); }
        .transition-colors { transition-property: color, background-color, border-color, text-decoration-color, fill, stroke; transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1); transition-duration: 150ms; }
        
        @media (min-width: 640px) {
            .sm\\:grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
        }
        @media (min-width: 1024px) {
            .lg\\:grid-cols-4 { grid-template-columns: repeat(4, minmax(0, 1fr)); }
        }
        
        /* Custom styles for report */
        .report-container { max-width: 1200px; margin: 0 auto; }
        .step-files-button { 
            background: none; 
            border: none; 
            color: rgb(59 130 246); 
            text-decoration: underline; 
            cursor: pointer; 
            font-size: 0.875rem;
        }
        .step-files-button:hover { color: rgb(37 99 235); }
    </style>
</head>
<body>
    <div class="report-container px-4 py-6">
        <div class="mb-6">
            <h1 class="text-2xl font-bold text-gray-900 mb-2">
                Automation Report - ${automation.Name}
            </h1>
            <p class="text-sm text-gray-600">
                Run ID: ${run.ID} | Project: ${project.Name}
            </p>
            <p class="text-sm text-gray-500">
                Generated: ${new Date().toLocaleString()}
            </p>
        </div>
        
        ${reportContent}
    </div>
</body>
</html>`;

      // Create and download the HTML file
      const blob = new Blob([htmlContent], { type: 'text/html;charset=utf-8;' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `automation_report_${runId}.html`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      showSuccessToast('Report exported to HTML successfully');
    } catch (error) {
      console.error('Failed to export HTML:', error);
      showErrorToast('Failed to export HTML report');
    }
  }
  
  // Auto-scroll logs to bottom when new entries are added
  $effect(() => {
    if (liveLogs.length > 0) {
      const logsContainer = document.getElementById('logs-container');
      if (logsContainer) {
        setTimeout(() => {
          logsContainer.scrollTop = logsContainer.scrollHeight;
        }, 100);
      }
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
      <button
        onclick={exportToCSV}
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <TableColumnOutline class="-ml-1 mr-2 h-5 w-5" />
        Export CSV
      </button>
      <button
        onclick={exportToJSON}
        class="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <DownloadOutline class="-ml-1 mr-2 h-5 w-5" />
        Export JSON
      </button>
      <button
        onclick={exportToHTML}
        class="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <svg class="-ml-1 mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        Export HTML
      </button>
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
        {#if isCancelling}
          Cancelling...
        {:else}
          All Runs
        {/if}
      </a>
      {#if liveStatus === "running" || liveStatus === "pending"}
        <button
          onclick={handleCancelRun}
          disabled={isCancelling}
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
      {:else if liveStatus === "queued"}
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

  <div class="report-container">
    <!-- Run Summary -->
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
          <dt class="text-sm font-medium text-gray-500">Concurrent Users</dt>
          <dd class="mt-1 text-2xl font-semibold text-gray-900">
            {Math.max(...reportData.map(step => step.concurrentUsers.length), 0)}
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
      </dl>
    </div>

    <!-- Performance Visualization -->
    {#if performanceMetrics.totalRuns > 1}
      <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Performance Analysis</h3>
        
        <!-- Performance Summary -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <div class="bg-blue-50 p-4 rounded-lg">
            <dt class="text-sm font-medium text-blue-600">Total Runs</dt>
            <dd class="mt-1 text-2xl font-semibold text-blue-900">{performanceMetrics.totalRuns}</dd>
          </div>
          <div class="bg-green-50 p-4 rounded-lg">
            <dt class="text-sm font-medium text-green-600">Success Rate</dt>
            <dd class="mt-1 text-2xl font-semibold text-green-900">
              {(100 - performanceMetrics.overallFailureRate).toFixed(1)}%
            </dd>
          </div>
          <div class="bg-red-50 p-4 rounded-lg">
            <dt class="text-sm font-medium text-red-600">Failure Rate</dt>
            <dd class="mt-1 text-2xl font-semibold text-red-900">
              {performanceMetrics.overallFailureRate.toFixed(1)}%
            </dd>
          </div>
          <div class="bg-purple-50 p-4 rounded-lg">
            <dt class="text-sm font-medium text-purple-600">Avg Run Duration</dt>
            <dd class="mt-1 text-2xl font-semibold text-purple-900">
              {formatDuration(performanceMetrics.runData.reduce((sum, run) => sum + run.totalDuration, 0) / performanceMetrics.runData.length)}
            </dd>
          </div>
        </div>
        
        <!-- Charts -->
        <RunPerformanceChart metrics={performanceMetrics} />
      </div>
    {/if}
  <!-- Run Details -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Details</h3>
    
    {#if liveStatus === 'running' && currentStep}
      <div class="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-md">
        <div class="flex items-center justify-between">
          <div>
            <h4 class="text-sm font-medium text-blue-800">Currently Executing</h4>
            <p class="text-sm text-blue-600">{currentStep}</p>
          </div>
          {#if liveProgress > 0}
            <div class="flex items-center">
              <div class="w-32 bg-blue-200 rounded-full h-2 mr-3">
                <div 
                  class="bg-blue-600 h-2 rounded-full transition-all duration-300" 
                  style="width: {liveProgress}%"
                ></div>
              </div>
              <span class="text-sm font-medium text-blue-800">{liveProgress}%</span>
            </div>
          {/if}
        </div>
      </div>
    {/if}
    
    <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-8">
      <div class="sm:col-span-1">
        <dt class="text-sm font-medium text-gray-500">Status</dt>
        <dd class="mt-1">
          <span
            class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadgeClass(liveStatus)}"
          >
            {liveStatus}
            {#if liveStatus === 'running' && liveProgress > 0}
              <span class="ml-2 text-xs">({liveProgress}%)</span>
            {/if}
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
            : liveStatus === "running"
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

    <!-- Detailed Step Report -->
    <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
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
                      <div class="flex items-center space-x-3 text-sm text-gray-500">
                        <span>{step.concurrentUsers.length} users</span>
                        {#if step.totalOutputFiles > 0}
                          <span>{step.totalOutputFiles} files</span>
                        {/if}
                        {#if step.totalFailures > 0}
                          <span class="text-red-600">{step.totalFailures} failures</span>
                        {/if}
                        <span>{formatDuration(step.totalDuration)}</span>
                      </div>
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
                  <!-- Concurrent Users Info -->
                  {#if step.concurrentUsers.length > 1}
                    <div class="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
                      <h5 class="text-sm font-medium text-blue-800 mb-2">Concurrent Execution</h5>
                      <p class="text-sm text-blue-700">
                        This step was executed by {step.concurrentUsers.length} concurrent users: 
                        {step.concurrentUsers.map(idx => `User ${idx}`).join(', ')}
                      </p>
                    </div>
                  {/if}

                  {#if step.actions.size === 0}
                    <p class="text-sm text-gray-500 italic">No actions recorded for this step.</p>
                  {:else}
                    <div class="space-y-3">
                      {#each Array.from(step.actions.values()) as action, actionIndex (action.id + '-' + action.loopIndex)}
                        <div class="border border-gray-100 rounded-md">
                          <!-- Action Header -->
                          <button
                            onclick={(event) => { event.stopPropagation(); toggleAction(action.id + '-' + action.loopIndex); }}
                            class="w-full px-4 py-3 text-left hover:bg-gray-25 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:ring-inset"
                          >
                            <div class="flex items-center justify-between">
                              <div class="flex items-center space-x-3">
                                <div class="flex-shrink-0">
                                  {#if expandedActions.has(action.id + '-' + action.loopIndex)}
                                    <ChevronDownOutline class="h-4 w-4 text-gray-400" />
                                  {:else}
                                    <ChevronRightOutline class="h-4 w-4 text-gray-400" />
                                  {/if}
                                </div>
                                <div>
                                  <h5 class="text-sm font-medium text-gray-900">
                                    {action.type}
                                    {#if action.parentActionId}
                                      <span class="text-xs text-gray-500">(nested)</span>
                                    {/if}
                                  </h5>
                                  <div class="flex items-center space-x-2 text-xs text-gray-500">
                                    <span>User {action.loopIndex}</span>
                                    {#if action.parentActionId}
                                      <span>Parent: {action.parentActionId.substring(0, 8)}...</span>
                                    {/if}
                                  </div>
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
                          {#if expandedActions.has(action.id + '-' + action.loopIndex)}
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
                                              onclick={(event) => { event.stopPropagation(); openImageViewer(fileUrl); }}
                                              class="text-xs text-primary-600 hover:text-primary-800 font-medium"
                                            >
                                              View Image
                                            </button>
                                          {:else}
                                            <a
                                              href={fileUrl}
                                              target="_blank"
                                              rel="noopener noreferrer"
                                              onclick={(event) => event.stopPropagation()}
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
</div>

<!-- Image Viewer Modal -->
<ImageViewerModal
  bind:open={showImageViewerModal}
  imageUrls={imageFilesForModal}
  startIndex={currentImageIndex}
  onClose={() => showImageViewerModal = false}
/>

<!-- Step Image Viewer Modal -->
<ImageViewerModal
  bind:open={showStepImageViewerModal}
  imageUrls={stepImageFiles}
  startIndex={0}
  onClose={() => showStepImageViewerModal = false}
/>

<style>
  .bg-gray-25 {
    background-color: #fafafa;
  }
</style>
