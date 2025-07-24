<script lang="ts">
  import { onMount } from 'svelte';
  import { Chart, registerables } from 'chart.js';
  import { formatDuration } from '$lib/utils/date';

  type PerformanceMetrics = {
    stepAverages: Array<{
      name: string;
      averageDuration: number;
      failureRate: number;
      totalRuns: number;
    }>;
    runData: Array<{
      loopIndex: number;
      steps: Map<string, { duration: number; status: string }>;
      totalDuration: number;
      status: string;
    }>;
    totalRuns: number;
    overallFailureRate: number;
  };

  let { metrics }: { metrics: PerformanceMetrics } = $props();

  let stepDurationChartCanvas: HTMLCanvasElement;
  let runLatencyChartCanvas: HTMLCanvasElement;
  let stepDurationChart: Chart | null = null;
  let runLatencyChart: Chart | null = null;

  onMount(() => {
    Chart.register(...registerables);
    createCharts();
    
    return () => {
      if (stepDurationChart) {
        stepDurationChart.destroy();
      }
      if (runLatencyChart) {
        runLatencyChart.destroy();
      }
    };
  });

  $effect(() => {
    if (stepDurationChart && runLatencyChart) {
      updateCharts();
    }
  });

  function createCharts() {
    createStepDurationChart();
    createRunLatencyChart();
  }

  function createStepDurationChart() {
    const ctx = stepDurationChartCanvas.getContext('2d');
    if (!ctx) return;

    const stepNames = metrics.stepAverages.map(step => step.name);
    const averageDurations = metrics.stepAverages.map(step => step.averageDuration);
    const failureRates = metrics.stepAverages.map(step => step.failureRate);

    stepDurationChart = new Chart(ctx, {
      type: 'bar',
      data: {
        labels: stepNames,
        datasets: [
          {
            label: 'Average Duration (ms)',
            data: averageDurations,
            backgroundColor: 'rgba(59, 130, 246, 0.6)',
            borderColor: 'rgba(59, 130, 246, 1)',
            borderWidth: 1,
            yAxisID: 'y'
          },
          {
            label: 'Failure Rate (%)',
            data: failureRates,
            type: 'line',
            backgroundColor: 'rgba(239, 68, 68, 0.6)',
            borderColor: 'rgba(239, 68, 68, 1)',
            borderWidth: 2,
            yAxisID: 'y1',
            tension: 0.4
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Step Performance Overview'
          },
          legend: {
            display: true,
            position: 'top'
          }
        },
        scales: {
          x: {
            title: {
              display: true,
              text: 'Steps'
            }
          },
          y: {
            type: 'linear',
            display: true,
            position: 'left',
            title: {
              display: true,
              text: 'Duration (ms)'
            }
          },
          y1: {
            type: 'linear',
            display: true,
            position: 'right',
            title: {
              display: true,
              text: 'Failure Rate (%)'
            },
            grid: {
              drawOnChartArea: false,
            },
            min: 0,
            max: 100
          }
        }
      }
    });
  }

  function createRunLatencyChart() {
    const ctx = runLatencyChartCanvas.getContext('2d');
    if (!ctx) return;

    // Prepare data for run latency chart
    const runLabels = metrics.runData.map(run => `Run ${run.loopIndex + 1}`);
    const runDurations = metrics.runData.map(run => run.totalDuration);
    
    // Create datasets for each step to show individual step performance across runs
    const stepNames = [...new Set(metrics.stepAverages.map(step => step.name))];
    const colors = [
      'rgba(59, 130, 246, 0.8)',   // blue
      'rgba(16, 185, 129, 0.8)',   // green
      'rgba(245, 158, 11, 0.8)',   // yellow
      'rgba(239, 68, 68, 0.8)',    // red
      'rgba(139, 92, 246, 0.8)',   // purple
      'rgba(236, 72, 153, 0.8)',   // pink
      'rgba(14, 165, 233, 0.8)',   // sky
      'rgba(34, 197, 94, 0.8)',    // emerald
    ];

    const datasets = [
      {
        label: 'Total Run Duration',
        data: runDurations,
        backgroundColor: 'rgba(17, 24, 39, 0.8)',
        borderColor: 'rgba(17, 24, 39, 1)',
        borderWidth: 2,
        tension: 0.4
      }
    ];

    // Add individual step datasets
    stepNames.forEach((stepName, index) => {
      const stepDurations = metrics.runData.map(run => {
        const stepData = run.steps.get(stepName);
        return stepData ? stepData.duration : 0;
      });

      datasets.push({
        label: stepName,
        data: stepDurations,
        backgroundColor: colors[index % colors.length],
        borderColor: colors[index % colors.length].replace('0.8', '1'),
        borderWidth: 1,
        tension: 0.4
      });
    });

    runLatencyChart = new Chart(ctx, {
      type: 'line',
      data: {
        labels: runLabels,
        datasets: datasets
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: 'Run Latency Analysis (Multi-Run Performance)'
          },
          legend: {
            display: true,
            position: 'top'
          }
        },
        scales: {
          x: {
            title: {
              display: true,
              text: 'Run Number (Loop Index)'
            }
          },
          y: {
            title: {
              display: true,
              text: 'Duration (ms)'
            },
            beginAtZero: true
          }
        },
        interaction: {
          intersect: false,
          mode: 'index'
        }
      }
    });
  }

  function updateCharts() {
    if (stepDurationChart) {
      stepDurationChart.destroy();
      createStepDurationChart();
    }
    if (runLatencyChart) {
      runLatencyChart.destroy();
      createRunLatencyChart();
    }
  }
</script>

<div class="space-y-6">
  <!-- Step Duration Chart -->
  <div class="bg-gray-50 p-4 rounded-lg">
    <div style="height: 400px; position: relative;">
      <canvas bind:this={stepDurationChartCanvas}></canvas>
    </div>
  </div>

  <!-- Run Latency Chart -->
  <div class="bg-gray-50 p-4 rounded-lg">
    <div style="height: 400px; position: relative;">
      <canvas bind:this={runLatencyChartCanvas}></canvas>
    </div>
  </div>

  <!-- Performance Insights -->
  <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
    <h4 class="text-md font-semibold text-yellow-800 mb-3">Performance Insights</h4>
    <div class="space-y-2 text-sm text-yellow-700">
      {#if metrics.overallFailureRate > 10}
        <p>⚠️ High failure rate detected ({metrics.overallFailureRate.toFixed(1)}%). Consider reviewing failed steps.</p>
      {:else if metrics.overallFailureRate > 0}
        <p>✅ Low failure rate ({metrics.overallFailureRate.toFixed(1)}%) - Good stability.</p>
      {:else}
        <p>🎉 Perfect run! No failures detected across all {metrics.totalRuns} runs.</p>
      {/if}
      
      {#if metrics.stepAverages.length > 0}
        {@const slowestStep = metrics.stepAverages.reduce((prev, current) => 
          prev.averageDuration > current.averageDuration ? prev : current
        )}
        <p>🐌 Slowest step: "{slowestStep.name}" (avg: {formatDuration(slowestStep.averageDuration)})</p>
        
        {@const fastestStep = metrics.stepAverages.reduce((prev, current) => 
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