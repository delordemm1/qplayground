<script lang="ts">
  import { onMount } from 'svelte';
  import { Chart, registerables } from 'chart.js';
  import { formatDuration, calculateDurationStats } from '$lib/utils/date';

  type AggregatedAction = {
    type: string;
    selector?: string;
    executions: number;
    successCount: number;
    failureCount: number;
    durations: number[];
    stats: {
      avg: number;
      min: number;
      max: number;
      p50: number;
      p95: number;
      count: number;
    };
    failedExecutions: Array<{
      loopIndex: number;
      errorMessage: string;
      outputFiles: string[];
    }>;
  };

  type EnhancedStepData = {
    id: string;
    name: string;
    aggregatedActions: Map<string, AggregatedAction>;
    totalDuration: number;
    status: string;
    startTime: string;
    endTime: string;
    concurrentUsers: number[];
    totalExecutions: number;
    totalFailures: number;
    totalOutputFiles: number;
    stepImageFiles: string[];
  };

  type HeatmapData = {
    stepName: string;
    userIndex: number;
    duration: number;
    status: string;
  };

  let { 
    reportData, 
    performanceMetrics 
  }: { 
    reportData: EnhancedStepData[];
    performanceMetrics: any;
  } = $props();

  let stepPerformanceChartCanvas: HTMLCanvasElement;
  let heatmapChartCanvas: HTMLCanvasElement;
  let stepPerformanceChart: Chart | null = null;
  let heatmapChart: Chart | null = null;

  onMount(() => {
    Chart.register(...registerables);
    createCharts();
    
    return () => {
      if (stepPerformanceChart) {
        stepPerformanceChart.destroy();
      }
      if (heatmapChart) {
        heatmapChart.destroy();
      }
    };
  });

  $effect(() => {
    if (stepPerformanceChart && heatmapChart) {
      updateCharts();
    }
  });

  function createCharts() {
    createStepPerformanceChart();
    createHeatmapChart();
  }

  function createStepPerformanceChart() {
    const ctx = stepPerformanceChartCanvas.getContext('2d');
    if (!ctx) return;

    // Prepare data for horizontal floating bar chart
    const stepNames = reportData.map(step => step.name);
    const stepStats = reportData.map(step => {
      // Calculate overall step statistics from all actions
      const allDurations: number[] = [];
      step.aggregatedActions.forEach(action => {
        allDurations.push(...action.durations);
      });
      return calculateDurationStats(allDurations);
    });

    const failureRates = reportData.map(step => {
      const totalExecutions = step.totalExecutions;
      return totalExecutions > 0 ? (step.totalFailures / totalExecutions) * 100 : 0;
    });

    // Create floating bar data (P50 to P95 range)
    const performanceRanges = stepStats.map(stats => [stats.p50, stats.p95]);
    const averages = stepStats.map(stats => stats.avg);

    stepPerformanceChart = new Chart(ctx, {
      type: 'bar',
      data: {
        labels: stepNames,
        datasets: [
          {
            label: 'Performance Range (P50-P95)',
            data: performanceRanges,
            backgroundColor: 'rgba(59, 130, 246, 0.6)',
            borderColor: 'rgba(59, 130, 246, 1)',
            borderWidth: 1,
            barThickness: 20,
          },
          {
            label: 'Average Duration',
            data: averages,
            type: 'scatter',
            backgroundColor: 'rgba(239, 68, 68, 0.8)',
            borderColor: 'rgba(239, 68, 68, 1)',
            pointRadius: 6,
            pointStyle: 'circle',
          }
        ]
      },
      options: {
        indexAxis: 'y',
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Step Performance Overview (P50-P95 Range with Average)'
          },
          legend: {
            display: true,
            position: 'top'
          },
          tooltip: {
            callbacks: {
              label: function(context) {
                if (context.dataset.type === 'scatter') {
                  return `Average: ${formatDuration(context.parsed.x)}`;
                } else {
                  const [p50, p95] = context.parsed;
                  return `P50-P95: ${formatDuration(p50)} - ${formatDuration(p95)}`;
                }
              },
              afterLabel: function(context) {
                const stepIndex = context.dataIndex;
                const failureRate = failureRates[stepIndex];
                return `Failure Rate: ${failureRate.toFixed(1)}%`;
              }
            }
          }
        },
        scales: {
          x: {
            title: {
              display: true,
              text: 'Duration (ms)'
            },
            beginAtZero: true
          },
          y: {
            title: {
              display: true,
              text: 'Steps'
            }
          }
        }
      }
    });
  }

  function createHeatmapChart() {
    const ctx = heatmapChartCanvas.getContext('2d');
    if (!ctx) return;

    // Prepare heatmap data
    const heatmapData: HeatmapData[] = [];
    const stepNames = reportData.map(step => step.name);
    const maxUsers = Math.max(...reportData.map(step => step.concurrentUsers.length), 1);
    
    // Create a matrix of step performance by user
    reportData.forEach(step => {
      step.concurrentUsers.forEach(userIndex => {
        // Calculate average duration for this user in this step
        let userStepDuration = 0;
        let userStepStatus = 'success';
        let actionCount = 0;

        step.aggregatedActions.forEach(action => {
          // Find durations for this specific user (this is simplified - in reality you'd need to track per-user data)
          if (action.durations.length > 0) {
            userStepDuration += action.durations[userIndex % action.durations.length] || 0;
            actionCount++;
          }
          if (action.failedExecutions.some(failure => failure.loopIndex === userIndex)) {
            userStepStatus = 'failed';
          }
        });

        if (actionCount > 0) {
          userStepDuration = userStepDuration / actionCount;
        }

        heatmapData.push({
          stepName: step.name,
          userIndex: userIndex,
          duration: userStepDuration,
          status: userStepStatus
        });
      });
    });

    // Calculate color scale bounds
    const allDurations = heatmapData.map(d => d.duration);
    const minDuration = Math.min(...allDurations);
    const maxDuration = Math.max(...allDurations);

    // Create scatter plot to simulate heatmap
    const datasets = stepNames.map((stepName, stepIndex) => {
      const stepData = heatmapData.filter(d => d.stepName === stepName);
      
      return {
        label: stepName,
        data: stepData.map(d => ({
          x: d.userIndex,
          y: stepIndex,
          duration: d.duration,
          status: d.status
        })),
        backgroundColor: (context: any) => {
          const point = context.raw;
          if (point.status === 'failed') {
            return 'rgba(239, 68, 68, 0.8)'; // Red for failures
          }
          // Color scale from green (fast) to red (slow)
          const ratio = (point.duration - minDuration) / (maxDuration - minDuration);
          const red = Math.floor(255 * ratio);
          const green = Math.floor(255 * (1 - ratio));
          return `rgba(${red}, ${green}, 0, 0.7)`;
        },
        pointRadius: 8,
        pointHoverRadius: 10,
      };
    });

    heatmapChart = new Chart(ctx, {
      type: 'scatter',
      data: {
        datasets: datasets
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Run Latency Heatmap (User vs Step Performance)'
          },
          legend: {
            display: false // Too many datasets for legend
          },
          tooltip: {
            callbacks: {
              title: function(context) {
                const point = context[0].raw as any;
                return `User ${point.x} - ${stepNames[point.y]}`;
              },
              label: function(context) {
                const point = context.raw as any;
                return [
                  `Duration: ${formatDuration(point.duration)}`,
                  `Status: ${point.status.toUpperCase()}`
                ];
              }
            }
          }
        },
        scales: {
          x: {
            type: 'linear',
            title: {
              display: true,
              text: 'User Index'
            },
            min: 0,
            max: maxUsers
          },
          y: {
            type: 'linear',
            title: {
              display: true,
              text: 'Steps'
            },
            min: -0.5,
            max: stepNames.length - 0.5,
            ticks: {
              stepSize: 1,
              callback: function(value) {
                return stepNames[value as number] || '';
              }
            }
          }
        },
        interaction: {
          intersect: false,
          mode: 'point'
        }
      }
    });
  }

  function updateCharts() {
    if (stepPerformanceChart) {
      stepPerformanceChart.destroy();
      createStepPerformanceChart();
    }
    if (heatmapChart) {
      heatmapChart.destroy();
      createHeatmapChart();
    }
  }
</script>

<div class="space-y-6">
  <!-- Step Performance Chart -->
  <div class="bg-gray-50 p-4 rounded-lg">
    <div style="height: 500px; position: relative;">
      <canvas bind:this={stepPerformanceChartCanvas}></canvas>
    </div>
  </div>

  <!-- Heatmap Chart -->
  <div class="bg-gray-50 p-4 rounded-lg">
    <div style="height: 400px; position: relative;">
      <canvas bind:this={heatmapChartCanvas}></canvas>
    </div>
  </div>

  <!-- Performance Insights -->
  <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
    <h4 class="text-md font-semibold text-yellow-800 mb-3">Performance Insights</h4>
    <div class="space-y-2 text-sm text-yellow-700">
      {#if performanceMetrics.overallFailureRate > 10}
        <p>⚠️ High failure rate detected ({performanceMetrics.overallFailureRate.toFixed(1)}%). Consider reviewing failed steps.</p>
      {:else if performanceMetrics.overallFailureRate > 0}
        <p>✅ Low failure rate ({performanceMetrics.overallFailureRate.toFixed(1)}%) - Good stability.</p>
      {:else}
        <p>🎉 Perfect run! No failures detected across all {performanceMetrics.totalRuns} runs.</p>
      {/if}
      
      {#if performanceMetrics.stepAverages.length > 0}
        {@const slowestStep = performanceMetrics.stepAverages.reduce((prev, current) => 
          prev.averageDuration > current.averageDuration ? prev : current
        )}
        <p>🐌 Slowest step: "{slowestStep.name}" (avg: {formatDuration(slowestStep.averageDuration)})</p>
        
        {@const fastestStep = performanceMetrics.stepAverages.reduce((prev, current) => 
          prev.averageDuration < current.averageDuration ? prev : current
        )}
        <p>⚡ Fastest step: "{fastestStep.name}" (avg: {formatDuration(fastestStep.averageDuration)})</p>
      {/if}
    </div>
  </div>
</div>

<style>
  /* Ensure charts are responsive */
  canvas {
    max-width: 100%;
    height: auto !important;
  }
</style>