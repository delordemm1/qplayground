<script lang="ts">
  import { page } from "@inertiajs/svelte";
  import { formatDate } from "$lib/utils/date";
  import { calculatePercentile } from "$lib/utils/date";
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";
  import ImageViewerModal from "$lib/components/ImageViewerModal.svelte";
  import RunPerformanceChart from "$lib/components/RunPerformanceChart.svelte";
  import UserExplorerModal from "$lib/components/UserExplorerModal.svelte";
  import {
    ChevronDownOutline,
    ChevronRightOutline,
    DownloadOutline,
    TableColumnOutline,
    UserOutline,
    CaretDownOutline
  } from "flowbite-svelte-icons";
  import { Dropdown, DropdownItem } from "flowbite-svelte";

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
  let expandedFailureActions = $state<Set<string>>(new Set());

  // State for step image viewer
  let showStepImageViewerModal = $state(false);
  let stepImageFiles = $state<string[]>([]);

  // User Explorer Modal state
  let showUserExplorerModal = $state(false);

  // Live step summaries from SSE
  let liveStepSummaries = $state<Map<string, any>>(new Map());

  // Initialize SSE connection for real-time updates
  $effect(() => {
    if (typeof window !== "undefined") {
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
        console.error("Failed to parse SSE message:", error);
      }
    };

    eventSource.onerror = (error) => {
      console.error("SSE connection error:", error);
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
      case "status":
        liveStatus = data.status;
        if (
          data.status === "cancelled" ||
          data.status === "completed" ||
          data.status === "failed"
        ) {
          // Close SSE connection for final states
          if (eventSource) {
            eventSource.close();
            eventSource = null;
          }
        }
        break;

      case "step":
        currentStep = data.stepName;
        liveProgress = data.progress || 0;
        break;

      case "step_summary":
        // Update live step summaries for dashboard
        liveStepSummaries.set(data.stepId, {
          stepId: data.stepId,
          stepName: data.stepName,
          completedCount: data.completedCount,
          inProgressCount: data.inProgressCount,
          failedCount: data.failedCount,
          totalUsersForStep: data.totalUsersForStep,
          averageDurationMs: data.averageDurationMs,
          filesCount: data.filesCount,
        });
        liveStepSummaries = new Map(liveStepSummaries);
        break;

      case "log":
        liveLogs = [
          ...liveLogs,
          {
            timestamp: data.timestamp,
            stepName: data.stepName,
            stepId: data.stepId,
            actionId: data.actionId,
            actionName: data.actionName,
            actionType: data.actionType,
            message: data.message,
            duration: data.duration,
            loopIndex: data.loopIndex,
            status: "success",
          },
        ];
        break;

      case "error":
        liveLogs = [
          ...liveLogs,
          {
            timestamp: data.timestamp,
            stepName: data.stepName,
            stepId: data.stepId,
            actionId: data.actionId,
            actionName: data.actionName,
            actionType: data.actionType,
            error: data.error,
            loopIndex: data.loopIndex,
            status: "failed",
          },
        ];
        break;

      case "output":
        if (data.outputFile && !liveOutputFiles.includes(data.outputFile)) {
          liveOutputFiles = [...liveOutputFiles, data.outputFile];
        }
        break;

      case "complete":
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
  const enhancedReportData = $derived.by(() => {
    const stepMap = new Map();

    // Process logs to build step and action structure
    parsedLogs.forEach((log) => {
      const stepId = log.step_id;
      const actionId = log.action_id;
      const actionName = log.action_name;
      const loopIndex = log.loop_index || 0;

      if (!stepId) return;

      // Initialize step if not exists
      if (!stepMap.has(stepId)) {
        stepMap.set(stepId, {
          id: stepId,
          name: log.step_name || "Unknown Step",
          aggregatedActions: new Map(),
          rawActions: new Map(), // Keep original actions for drill-down
          totalDuration: 0,
          status: "success",
          startTime: log.timestamp,
          endTime: log.timestamp,
          logs: [],
          stepImageFiles: [],
          concurrentUsers: new Set(),
          loopIndexes: new Set(),
          totalOutputFiles: 0,
          totalFailures: 0,
          totalExecutions: 0,
        });
      }

      const step = stepMap.get(stepId);
      step.endTime = log.timestamp;
      step.logs.push(log);
      step.concurrentUsers.add(loopIndex);
      step.loopIndexes.add(loopIndex);
      step.totalExecutions++;

      if (log.status === "failed") {
        step.status = "failed";
        step.totalFailures++;
      }

      // Process action if actionId exists
      if (actionId) {
        // Store raw action data for drill-down
        const rawActionKey = `${actionId}-${loopIndex}`;
        if (!step.rawActions.has(rawActionKey)) {
          step.rawActions.set(rawActionKey, {
            id: actionId,
            name: actionName,
            loopIndex: loopIndex,
            type: log.action_type || "Unknown Action",
            parentActionId: log.parent_action_id,
            duration: log.duration_ms || 0,
            status: log.status || "success",
            error: log.error || null,
            logs: [],
            outputFiles: [],
          });
        }

        const rawAction = step.rawActions.get(rawActionKey);
        rawAction.logs.push(log);

        // Aggregate actions by action ID (unique action definition)
        const aggregateKey = actionId;

        if (!step.aggregatedActions.has(aggregateKey)) {
          step.aggregatedActions.set(aggregateKey, {
            id: actionId,
            name: actionName,
            type: log.action_type || "Unknown Action",
            executions: 0,
            successCount: 0,
            failureCount: 0,
            durations: [],
            stats: { avg: 0, min: 0, max: 0, p50: 0, p95: 0, count: 0 },
            failedExecutions: [],
          });
        }

        const aggregatedAction = step.aggregatedActions.get(aggregateKey);
        aggregatedAction.executions++;
        aggregatedAction.durations.push(log.duration_ms || 0);

        if (log.status === "failed") {
          aggregatedAction.failureCount++;
          aggregatedAction.failedExecutions.push({
            loopIndex: loopIndex,
            errorMessage: log.error || "Unknown error",
            outputFiles: log.output_file ? [log.output_file] : [],
          });
        } else {
          aggregatedAction.successCount++;
        }

        const action = step.rawActions.get(rawActionKey);
        action.logs.push(log);

        if (log.output_file) {
          action.outputFiles.push(log.output_file);
          step.totalOutputFiles++; // Increment step's total output files
          // Add image files to step-level collection
          if (getFileType(log.output_file) === "image") {
            step.stepImageFiles.push(log.output_file);
          }
        }
      }
    });

    // Calculate step durations, convert Sets to arrays, and compute aggregated action stats
    stepMap.forEach((step) => {
      if (step.startTime && step.endTime) {
        step.totalDuration =
          new Date(step.endTime).getTime() - new Date(step.startTime).getTime();
      }
      step.concurrentUsers = Array.from(step.concurrentUsers).sort(
        (a, b) => a - b
      );
      step.loopIndexes = Array.from(step.loopIndexes).sort((a, b) => a - b);

      // Calculate statistics for each aggregated action
      step.aggregatedActions.forEach((action) => {
        action.stats = calculateDurationStats(action.durations);
      });
    });

    return Array.from(stepMap.values()).sort(
      (a, b) =>
        new Date(a.startTime).getTime() - new Date(b.startTime).getTime()
    );
  });

  // Enhanced performance metrics with KPIs
  const enhancedPerformanceMetrics = $derived.by(() => {
    const stepMetrics = new Map();
    const runMetrics = new Map(); // Group by loop index
    let totalUsers = new Set();
    let totalActionExecutions = 0;
    let totalActionFailures = 0;
    let allDurations: number[] = [];

    parsedLogs.forEach((log) => {
      const stepName = log.step_name || "Unknown Step";
      const loopIndex = log.loop_index || 0;
      const duration = log.duration_ms || 0;
      const status = log.status || "success";

      totalUsers.add(loopIndex);
      if (duration > 0) allDurations.push(duration);

      // Count action executions for accurate success rate
      if (log.action_id) {
        totalActionExecutions++;
        if (status === "failed") totalActionFailures++;
      }

      // Step-level metrics
      if (!stepMetrics.has(stepName)) {
        stepMetrics.set(stepName, {
          name: stepName,
          durations: [],
          failures: 0,
          totalRuns: 0,
        });
      }

      const stepMetric = stepMetrics.get(stepName);
      stepMetric.durations.push(duration);
      stepMetric.totalRuns++;
      if (status === "failed") {
        stepMetric.failures++;
      }

      // Run-level metrics (by loop index)
      if (!runMetrics.has(loopIndex)) {
        runMetrics.set(loopIndex, {
          loopIndex,
          steps: new Map(),
          totalDuration: 0,
          status: "success",
        });
      }

      const runMetric = runMetrics.get(loopIndex);
      if (!runMetric.steps.has(stepName)) {
        runMetric.steps.set(stepName, { duration: 0, status: "success" });
      }

      const runStepMetric = runMetric.steps.get(stepName);
      runStepMetric.duration += duration;
      if (status === "failed") {
        runStepMetric.status = "failed";
        runMetric.status = "failed";
      }

      runMetric.totalDuration += duration;
    });

    // Calculate averages and prepare chart data
    const stepAverages = Array.from(stepMetrics.values()).map((metric) => ({
      name: metric.name,
      averageDuration:
        metric.durations.reduce((sum, d) => sum + d, 0) /
        metric.durations.length,
      failureRate: (metric.failures / metric.totalRuns) * 100,
      totalRuns: metric.totalRuns,
    }));

    const runData = Array.from(runMetrics.values()).sort(
      (a, b) => a.loopIndex - b.loopIndex
    );

    // Calculate KPIs
    const totalUserCount = totalUsers.size;
    // More accurate success rate based on action executions
    const successRate =
      totalActionExecutions > 0
        ? ((totalActionExecutions - totalActionFailures) /
            totalActionExecutions) *
          100
        : 100;
    const overallFailureRate = 100 - successRate;
    const avgResponseTime =
      allDurations.length > 0
        ? allDurations.reduce((sum, d) => sum + d, 0) / allDurations.length
        : 0;
    const p95ResponseTime = calculatePercentile(allDurations, 95);

    // Generate automated insights
    const insights: string[] = [];

    // High failure rate detection
    stepAverages.forEach((step) => {
      if (step.failureRate > 10) {
        insights.push(
          `⚠️ High Failure Rate: Step "${step.name}" failed for ${step.failureRate.toFixed(1)}% of users.`
        );
      }
    });

    // Performance bottleneck detection
    stepAverages.forEach((step) => {
      if (step.averageDuration > avgResponseTime * 2) {
        insights.push(
          `🐌 Performance Bottleneck: Step "${step.name}" is ${((step.averageDuration / avgResponseTime) * 100).toFixed(0)}% slower than average.`
        );
      }
    });

    // P95 vs Average detection
    const overallP95 = calculatePercentile(allDurations, 95);
    if (overallP95 > avgResponseTime * 3) {
      insights.push(
        `📊 High Variance: P95 response time (${formatDuration(overallP95)}) is significantly higher than average (${formatDuration(avgResponseTime)}).`
      );
    }

    if (insights.length === 0) {
      insights.push(
        "✅ No significant issues detected. Performance looks healthy!"
      );
    }

    return {
      // KPIs
      totalUsers: totalUserCount,
      successRate: successRate,
      avgResponseTime: avgResponseTime,
      p95ResponseTime: p95ResponseTime,
      totalErrors: totalActionFailures,
      insights: insights,
      // Original metrics
      stepAverages,
      runData,
      totalRuns: runMetrics.size,
      overallFailureRate: overallFailureRate,
    };
  });

  // Import the enhanced calculation function
  import { calculateDurationStats } from "$lib/utils/date";
  import { Button } from "flowbite-svelte";
  // Filter output files to get only images for the modal
  const imageFiles = $derived.by(() => {
    return parsedOutputFiles.filter(
      (fileUrl) => getFileType(fileUrl) === "image"
    );
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

  function toggleFailureAction(actionKey: string) {
    if (expandedFailureActions.has(actionKey)) {
      expandedFailureActions.delete(actionKey);
    } else {
      expandedFailureActions.add(actionKey);
    }
    expandedFailureActions = new Set(expandedFailureActions);
  }

  // Export functions
  function exportUserStepDetailCSV() {
    try {
      // 1. Prepare a structured summary from the raw logs
      const userRunSummaries = new Map();

      // The 'parsedLogs' variable should contain all log entries from your run
      parsedLogs.forEach(log => {
        const userId = log.loop_index ?? 0;
        const stepName = log.step_name;

        if (!stepName) return; // Skip logs without a step context

        // Initialize the main entry for the user if it doesn't exist
        if (!userRunSummaries.has(userId)) {
          userRunSummaries.set(userId, {
            overallStatus: 'success',
            overallDuration: 0,
            steps: new Map()
          });
        }
        const userRun = userRunSummaries.get(userId);

        // Initialize the summary for this specific step if it doesn't exist
        if (!userRun.steps.has(stepName)) {
          userRun.steps.set(stepName, {
            status: 'success',
            duration: 0,
            actionCount: 0,
            errorCount: 0,
          });
        }
        const stepSummary = userRun.steps.get(stepName);

        // Aggregate the data
        const duration = log.duration_ms || 0;
        userRun.overallDuration += duration;
        stepSummary.duration += duration;
        stepSummary.actionCount++;

        if (log.status === 'failed') {
          userRun.overallStatus = 'failed';
          stepSummary.status = 'failed';
          stepSummary.errorCount++;
        }
      });

      // 2. Define the new, more useful CSV headers
      const csvRows = [];
      csvRows.push([
        'User ID',
        'Overall Run Status',
        'Overall Run Duration (ms)',
        'Step Name',
        'Step Status',
        'Step Duration (ms)',
        'Actions in Step',
        'Errors in Step'
      ]);

      // 3. Build the CSV rows from the structured summary
      userRunSummaries.forEach((runData, userId) => {
        if (runData.steps.size === 0) {
          // Handle cases where a user might have failed before completing any steps
          csvRows.push([
            userId,
            runData.overallStatus,
            runData.overallDuration,
            'No Steps Recorded',
            runData.overallStatus,
            0,
            0,
            runData.overallStatus === 'failed' ? 1 : 0
          ]);
        } else {
          runData.steps.forEach((stepData, stepName) => {
            csvRows.push([
              userId,
              runData.overallStatus,
              runData.overallDuration,
              stepName,
              stepData.status,
              stepData.duration,
              stepData.actionCount,
              stepData.errorCount
            ]);
          });
        }
      });

      // 4. Convert to CSV string and trigger download
      const csvContent = csvRows
        .map((row) =>
          row.map((field) => `"${String(field).replace(/"/g, '""')}"`).join(",")
        )
        .join("\n");

      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.download = `automation_per_user_detail_${runId}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      showSuccessToast('User step detail report exported to CSV successfully');
    } catch (error) {
      console.error('Failed to export user step detail CSV:', error);
      showErrorToast('Failed to export user step detail CSV');
    }
  }

  function exportAggregatedSummaryCSV() {
    try {
      // 1. Prepare data by aggregating stats for each unique step name
      const stepAggregates = new Map();

      parsedLogs.forEach(log => {
        const stepName = log.step_name;
        if (!stepName) return;

        if (!stepAggregates.has(stepName)) {
          stepAggregates.set(stepName, {
            durations: [],
            successCount: 0,
            failureCount: 0,
          });
        }
        const stepData = stepAggregates.get(stepName);
        
        stepData.durations.push(log.duration_ms || 0);
        if (log.status === 'failed') {
          stepData.failureCount++;
        } else {
          stepData.successCount++;
        }
      });

      // 2. Define headers for the aggregated summary
      const csvRows = [];
      csvRows.push([
        'Step Name',
        'Total Executions',
        'Success Rate (%)',
        'Failure Count',
        'Average Duration (ms)',
        'Min Duration (ms)',
        'Max Duration (ms)',
        '95th Percentile Duration (ms)'
      ]);

      // 3. Calculate final metrics and build the CSV rows
      stepAggregates.forEach((data, stepName) => {
        const totalExecutions = data.successCount + data.failureCount;
        const successRate = totalExecutions > 0 ? (data.successCount / totalExecutions) * 100 : 0;
        
        const sortedDurations = [...data.durations].sort((a, b) => a - b);
        const sum = sortedDurations.reduce((a, b) => a + b, 0);
        const avg = totalExecutions > 0 ? sum / totalExecutions : 0;
        const min = sortedDurations.length > 0 ? sortedDurations[0] : 0;
        const max = sortedDurations.length > 0 ? sortedDurations[sortedDurations.length - 1] : 0;
        
        // Calculate P95
        const p95 = sortedDurations.length > 0 ? calculatePercentile(sortedDurations, 95) : 0;
        
        csvRows.push([
          stepName,
          totalExecutions,
          successRate.toFixed(2),
          data.failureCount,
          avg.toFixed(2),
          min,
          max,
          p95.toFixed(2)
        ]);
      });

      // 4. Generate and download the CSV
      const csvContent = csvRows.map(row => row.map(field => `"${String(field).replace(/"/g, '""')}"`).join(",")).join("\n");
      const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" });
      const link = document.createElement("a");
      link.href = URL.createObjectURL(blob);
      link.download = `automation_aggregated_summary_${runId}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      showSuccessToast("Aggregated summary exported to CSV successfully");
    } catch (error) {
      console.error("Failed to export aggregated CSV:", error);
      showErrorToast("Failed to export aggregated CSV");
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
          errorMessage: run.ErrorMessage,
        },
        automation: {
          id: automation.ID,
          name: automation.Name,
        },
        project: {
          id: project.ID,
          name: project.Name,
        },
        steps: enhancedReportData.map((step) => ({
          id: step.id,
          name: step.name,
          totalDuration: step.totalDuration,
          status: step.status,
          startTime: step.startTime,
          endTime: step.endTime,
          concurrentUsers: step.concurrentUsers,
          aggregatedActions: Array.from(step.aggregatedActions.values()),
          rawActions: Array.from(step.rawActions.values()).map((action) => ({
            id: action.id,
            parentActionId: action.parentActionId,
            type: action.type,
            duration: action.duration,
            status: action.status,
            error: action.error,
            loopIndex: action.loopIndex,
            outputFiles: action.outputFiles,
            logs: action.logs,
          })),
          logs: step.logs,
        })),
        performanceMetrics: enhancedPerformanceMetrics,
        summary: {
          totalSteps: enhancedReportData.length,
          totalActions: enhancedReportData.reduce(
            (sum, step) => sum + step.rawActions.size,
            0
          ),
          totalDuration:
            run.StartTime && run.EndTime
              ? new Date(run.EndTime).getTime() -
                new Date(run.StartTime).getTime()
              : 0,
          totalConcurrentUsers: Math.max(
            ...enhancedReportData.map((step) => step.concurrentUsers.length),
            0
          ),
          outputFiles: parsedOutputFiles,
        },
      };

      const jsonContent = JSON.stringify(reportJson, null, 2);
      const blob = new Blob([jsonContent], {
        type: "application/json;charset=utf-8;",
      });
      const link = document.createElement("a");
      link.href = URL.createObjectURL(blob);
      link.download = `automation_report_${runId}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      showSuccessToast("Report exported to JSON successfully");
    } catch (error) {
      console.error("Failed to export JSON:", error);
      showErrorToast("Failed to export JSON report");
    }
  }

  function exportToHTML() {
    try {
      // Get the current page content
      const reportContent =
        document.querySelector(".report-container")?.innerHTML || "";

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
      const blob = new Blob([htmlContent], {
        type: "text/html;charset=utf-8;",
      });
      const link = document.createElement("a");
      link.href = URL.createObjectURL(blob);
      link.download = `automation_report_${runId}.html`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      showSuccessToast("Report exported to HTML successfully");
    } catch (error) {
      console.error("Failed to export HTML:", error);
      showErrorToast("Failed to export HTML report");
    }
  }

  // Auto-scroll logs to bottom when new entries are added
  $effect(() => {
    if (liveLogs.length > 0) {
      const logsContainer = document.getElementById("logs-container");
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
    <div class="mt-4 flex md:mt-0 md:ml-4 space-x-3">
      <div class="relative">
        <button
          id="csv-export-dropdown"
          class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
        >
          <TableColumnOutline class="-ml-1 mr-2 h-5 w-5" />
          Export CSV
          <CaretDownOutline class="ml-2 h-4 w-4" />
        </button>
        <Dropdown triggeredBy="#csv-export-dropdown" class="w-56">
          <DropdownItem onclick={exportUserStepDetailCSV}>
            <div class="flex flex-col">
              <span class="font-medium">User Step Detail</span>
              <span class="text-xs text-gray-500">Per-user step execution details</span>
            </div>
          </DropdownItem>
          <DropdownItem onclick={exportAggregatedSummaryCSV}>
            <div class="flex flex-col">
              <span class="font-medium">Aggregated Summary</span>
              <span class="text-xs text-gray-500">Step performance statistics</span>
            </div>
          </DropdownItem>
        </Dropdown>
      </div>
      <button
        onclick={exportToJSON}
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <DownloadOutline class="-ml-1 mr-2 h-5 w-5" />
        Export JSON
      </button>
      <button
        onclick={exportToHTML}
        class="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
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
            d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
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
        <span
          class="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-500 bg-gray-100"
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
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          Queued
        </span>
      {/if}
    </div>
  </div>

  <div class="report-container">
    <!-- High-Level Summary & Triage Section -->
    <div
      class="bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-lg p-6 mb-6"
    >
      <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">
        Executive Summary
      </h3>

      <!-- Key Performance Indicators -->
      <dl
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-x-4 gap-y-6 mb-6"
      >
        <div class="text-center">
          <dt class="text-sm font-medium text-gray-500">Total Users</dt>
          <dd class="mt-1 text-3xl font-bold text-blue-600">
            {enhancedPerformanceMetrics.totalUsers}
          </dd>
        </div>
        <div class="text-center">
          <dt class="text-sm font-medium text-gray-500">Success Rate</dt>
          <dd
            class="mt-1 text-3xl font-bold {enhancedPerformanceMetrics.successRate >=
            95
              ? 'text-green-600'
              : enhancedPerformanceMetrics.successRate >= 90
                ? 'text-yellow-600'
                : 'text-red-600'}"
          >
            {enhancedPerformanceMetrics.successRate.toFixed(1)}%
          </dd>
        </div>
        <div class="text-center">
          <dt class="text-sm font-medium text-gray-500">Avg Response Time</dt>
          <dd class="mt-1 text-3xl font-bold text-gray-900">
            {formatDuration(enhancedPerformanceMetrics.avgResponseTime)}
          </dd>
        </div>
        <div class="text-center">
          <dt class="text-sm font-medium text-gray-500">P95 Response Time</dt>
          <dd class="mt-1 text-3xl font-bold text-gray-900">
            {formatDuration(enhancedPerformanceMetrics.p95ResponseTime)}
          </dd>
        </div>
        <div class="text-center">
          <dt class="text-sm font-medium text-gray-500">Total Errors</dt>
          <dd
            class="mt-1 text-3xl font-bold {enhancedPerformanceMetrics.totalErrors ===
            0
              ? 'text-green-600'
              : 'text-red-600'}"
          >
            {enhancedPerformanceMetrics.totalErrors}
          </dd>
        </div>
      </dl>

      <!-- Automated Insights -->
      <div class="bg-white border border-gray-200 rounded-lg p-4">
        <h4 class="text-md font-semibold text-gray-800 mb-3">
          🔍 Automated Insights
        </h4>
        <div class="space-y-2">
          {#each enhancedPerformanceMetrics.insights as insight}
            <p class="text-sm text-gray-700">{insight}</p>
          {/each}
        </div>
      </div>
    </div>

    <!-- Original Run Summary (Simplified) -->
    <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
      <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">
        Run Details
      </h3>
      <dl
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-x-4 gap-y-6"
      >
        <div>
          <dt class="text-sm font-medium text-gray-500">Total Duration</dt>
          <dd class="mt-1 text-2xl font-semibold text-gray-900">
            {#if run.StartTime && run.EndTime}
              {formatDuration(
                new Date(run.EndTime).getTime() -
                  new Date(run.StartTime).getTime()
              )}
            {:else}
              N/A
            {/if}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Total Steps</dt>
          <dd class="mt-1 text-2xl font-semibold text-gray-900">
            {enhancedReportData.length}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Total Actions</dt>
          <dd class="mt-1 text-2xl font-semibold text-gray-900">
            {enhancedReportData.reduce(
              (sum, step) => sum + step.rawActions.size,
              0
            )}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Output Files</dt>
          <dd class="mt-1 text-2xl font-semibold text-gray-900">
            {parsedOutputFiles.length}
          </dd>
        </div>
      </dl>
    </div>

    <!-- Performance Visualization -->
    {#if enhancedPerformanceMetrics.totalRuns > 1}
      <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">
          Performance Analysis
        </h3>

        <!-- Charts -->
        <RunPerformanceChart
          reportData={enhancedReportData}
          performanceMetrics={enhancedPerformanceMetrics}
        />
      </div>
    {/if}
    <!-- Run Details -->
    <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
      <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">Details</h3>

      {#if liveStatus === "running" && currentStep}
        <div class="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-md">
          <div class="flex items-center justify-between">
            <div>
              <h4 class="text-sm font-medium text-blue-800">
                Currently Executing
              </h4>
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
                <span class="text-sm font-medium text-blue-800"
                  >{liveProgress}%</span
                >
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
              class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadgeClass(
                liveStatus
              )}"
            >
              {liveStatus}
              {#if liveStatus === "running" && liveProgress > 0}
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
    <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg leading-6 font-medium text-gray-900">
          Live Step Dashboard
        </h3>
        <Button onclick={() => (showUserExplorerModal = true)}>
          <UserOutline class="w-4 h-4 mr-2" />
          Explore Users
        </Button>
      </div>

      {#if enhancedReportData.length === 0}
        <p class="text-sm text-gray-500">
          No step data available for this run.
        </p>
      {:else}
        <div class="space-y-3">
          {#each enhancedReportData as step, stepIndex (step.id)}
            {@const liveSummary = liveStepSummaries.get(step.id)}
            {@const completedCount =
              liveSummary?.completedCount || step.concurrentUsers.length}
            {@const inProgressCount = liveSummary?.inProgressCount || 0}
            {@const failedCount =
              liveSummary?.failedCount || step.totalFailures}
            {@const totalUsersForStep =
              liveSummary?.totalUsersForStep || step.concurrentUsers.length}
            {@const avgDurationMs =
              liveSummary?.averageDurationMs ||
              step.totalDuration / Math.max(step.concurrentUsers.length, 1)}
            {@const filesCount =
              liveSummary?.filesCount || step.totalOutputFiles}
            {@const completedPercent =
              totalUsersForStep > 0
                ? (completedCount / totalUsersForStep) * 100
                : 0}
            {@const inProgressPercent =
              totalUsersForStep > 0
                ? (inProgressCount / totalUsersForStep) * 100
                : 0}
            {@const failedPercent =
              totalUsersForStep > 0
                ? (failedCount / totalUsersForStep) * 100
                : 0}
            {@const pendingPercent =
              100 - completedPercent - inProgressPercent - failedPercent}

            <div
              class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
            >
              <!-- Mission Control Row -->
              <div class="grid grid-cols-12 gap-4 items-center">
                <!-- Step Name & Status -->
                <div class="col-span-3">
                  <h4 class="font-medium text-gray-900">{step.name}</h4>
                  <p class="text-sm text-gray-500">
                    {#if inProgressCount > 0}
                      <span class="text-blue-600">In Progress</span>
                    {:else if failedCount > 0}
                      <span class="text-red-600">Failed</span>
                    {:else if completedCount === totalUsersForStep}
                      <span class="text-green-600">Completed</span>
                    {:else}
                      <span class="text-gray-600">Pending</span>
                    {/if}
                  </p>
                </div>
              </div>

              <!-- User Progress Bar -->
              <div class="col-span-4">
                <div
                  class="w-full bg-gray-200 rounded-full h-6 relative overflow-hidden"
                >
                  <!-- Completed (Green) -->
                  <div
                    class="absolute left-0 top-0 h-full bg-green-500 transition-all duration-300"
                    style="width: {completedPercent}%"
                  ></div>

                  <!-- In Progress (Blue) -->
                  <div
                    class="absolute top-0 h-full bg-blue-500 transition-all duration-300"
                    style="left: {completedPercent}%; width: {inProgressPercent}%"
                  ></div>

                  <!-- Failed (Red) -->
                  <div
                    class="absolute top-0 h-full bg-red-500 transition-all duration-300"
                    style="left: {completedPercent +
                      inProgressPercent}%; width: {failedPercent}%"
                  ></div>

                  <!-- Progress Text Overlay -->
                  <div
                    class="absolute inset-0 flex items-center justify-center"
                  >
                    <span class="text-xs font-medium text-white drop-shadow">
                      {completedPercent.toFixed(0)}% ✅ | {inProgressPercent.toFixed(
                        0
                      )}% 🏃 | {failedPercent.toFixed(0)}% ❌
                    </span>
                  </div>
                </div>
              </div>

              <!-- Key Metrics -->
              <div class="col-span-3 text-center">
                <div class="text-sm font-medium text-gray-900">
                  Users: {completedCount}/{totalUsersForStep}
                </div>
                <div class="text-xs text-gray-500">
                  Avg Time: {formatDuration(avgDurationMs)}
                </div>
              </div>

              <!-- Files Button -->
              <div class="col-span-2 text-right">
                <Button
                  size="sm"
                  color={filesCount > 0 ? "primary" : "alternative"}
                  disabled={filesCount === 0}
                  onclick={() => (showUserExplorerModal = true)}
                >
                  View {filesCount} files
                </Button>
              </div>
            </div>

            <!-- Expandable Details -->
            <div class="mt-4">
              <button
                onclick={() => toggleStep(step.id)}
                class="flex items-center text-sm text-gray-600 hover:text-gray-900"
              >
                {#if expandedSteps.has(step.id)}
                  <ChevronDownOutline class="h-4 w-4 mr-1" />
                  Hide Details
                {:else}
                  <ChevronRightOutline class="h-4 w-4 mr-1" />
                  Show Details ({step.aggregatedActions.size} actions)
                {/if}
              </button>

              {#if expandedSteps.has(step.id)}
                <div class="mt-4 space-y-3">
                  <!-- Aggregated Actions View -->
                  {#each Array.from(step.aggregatedActions.entries()) as [actionKey, action] (actionKey)}
                    <div
                      class="bg-gray-50 border border-gray-200 rounded-md p-3"
                    >
                      <div class="flex items-center justify-between">
                        <div>
                          <h6 class="text-sm font-medium text-gray-900">
                            ▶ {action.name || action.type}
                            {#if action.name}
                              <span class="text-xs text-gray-500"
                                >({action.type})</span
                              >
                            {/if}
                          </h6>
                          <div
                            class="flex items-center space-x-4 text-xs text-gray-600 mt-1"
                          >
                            <span
                              >Executions: {(
                                (action.successCount / action.executions) *
                                100
                              ).toFixed(0)}% ({action.successCount}/{action.executions})
                              Success</span
                            >
                            <span
                              >Duration: Avg: {formatDuration(action.stats.avg)}
                              | P95: {formatDuration(action.stats.p95)}</span
                            >
                          </div>
                        </div>
                        {#if action.failureCount > 0}
                          <button
                            onclick={() => toggleFailureAction(actionKey)}
                            class="text-sm font-medium text-red-600 hover:text-red-800 underline"
                          >
                            {action.failureCount} Failure{action.failureCount >
                            1
                              ? "s"
                              : ""}
                          </button>
                        {/if}
                      </div>

                      <!-- Failure Details (Drill-down on demand) -->
                      {#if action.failureCount > 0 && expandedFailureActions.has(actionKey)}
                        <div class="border-t border-gray-300 pt-3 mt-3">
                          <h6 class="text-xs font-semibold text-red-700 mb-2">
                            Failure Details:
                          </h6>
                          <div class="space-y-2">
                            {#each action.failedExecutions as failure}
                              <div
                                class="bg-red-50 border border-red-200 rounded p-3"
                              >
                                <div class="flex items-start justify-between">
                                  <div>
                                    <p class="text-sm font-medium text-red-800">
                                      User {failure.loopIndex}
                                    </p>
                                    <p class="text-xs text-red-700 mt-1">
                                      ❌ {failure.errorMessage}
                                    </p>
                                  </div>
                                  <div class="flex space-x-2">
                                    {#if failure.outputFiles.length > 0}
                                      {#each failure.outputFiles as fileUrl}
                                        {#if getFileType(fileUrl) === "image"}
                                          <button
                                            onclick={() =>
                                              openImageViewer(fileUrl)}
                                            class="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded hover:bg-blue-200"
                                          >
                                            View Screenshot
                                          </button>
                                        {:else}
                                          <a
                                            href={fileUrl}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            class="text-xs bg-gray-100 text-gray-800 px-2 py-1 rounded hover:bg-gray-200"
                                          >
                                            View File
                                          </a>
                                        {/if}
                                      {/each}
                                    {/if}
                                  </div>
                                </div>
                              </div>
                            {/each}
                          </div>
                        </div>
                      {/if}
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
            <!-- </div> -->
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
  onClose={() => (showImageViewerModal = false)}
/>

<!-- Step Image Viewer Modal -->
<ImageViewerModal
  bind:open={showStepImageViewerModal}
  imageUrls={stepImageFiles}
  startIndex={0}
  onClose={() => (showStepImageViewerModal = false)}
/>

<!-- User Explorer Modal -->
<UserExplorerModal
  bind:open={showUserExplorerModal}
  reportData={enhancedReportData}
  onClose={() => (showUserExplorerModal = false)}
/>

<style>
  .bg-gray-25 {
    background-color: #fafafa;
  }
</style>