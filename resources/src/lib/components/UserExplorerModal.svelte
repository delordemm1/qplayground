<script lang="ts">
  import { Modal, Button, Input } from "flowbite-svelte";
  import { SearchOutline, UserOutline, CheckCircleOutline, XCircleOutline, ClockOutline } from "flowbite-svelte-icons";
  import ImageViewerModal from "./ImageViewerModal.svelte";
  import { formatDuration } from "$lib/utils/date";

  type UserJourneyStep = {
    stepId: string;
    stepName: string;
    status: "success" | "failed" | "skipped" | "in_progress";
    duration: number;
    error?: string;
    outputFiles: string[];
    timestamp: string;
  };

  type UserData = {
    loopIndex: number;
    status: "completed" | "failed" | "in_progress";
    completedSteps: number;
    totalSteps: number;
    journey: UserJourneyStep[];
    totalDuration: number;
    totalFiles: number;
  };

  type Props = {
    open: boolean;
    reportData: any[];
    onClose: () => void;
  };

  let { open = $bindable(), reportData, onClose }: Props = $props();

  let searchTerm = $state("");
  let selectedUser = $state<UserData | null>(null);
  let showImageViewer = $state(false);
  let currentImageIndex = $state(0);
  let currentStepImages = $state<string[]>([]);

  // Process report data to extract user data
  const userData = $derived.by(() => {
    const userMap = new Map<number, UserData>();

    reportData.forEach(step => {
      Array.from(step.rawActions.values()).forEach(action => {
        const loopIndex = action.loopIndex;
        
        if (!userMap.has(loopIndex)) {
          userMap.set(loopIndex, {
            loopIndex,
            status: "in_progress",
            completedSteps: 0,
            totalSteps: reportData.length,
            journey: [],
            totalDuration: 0,
            totalFiles: 0,
          });
        }

        const user = userMap.get(loopIndex)!;
        
        // Find or create journey step
        let journeyStep = user.journey.find(j => j.stepId === step.id);
        if (!journeyStep) {
          journeyStep = {
            stepId: step.id,
            stepName: step.name,
            status: "in_progress",
            duration: 0,
            outputFiles: [],
            timestamp: action.logs[0]?.timestamp || "",
          };
          user.journey.push(journeyStep);
        }

        // Update journey step
        journeyStep.duration += action.duration;
        journeyStep.outputFiles.push(...action.outputFiles);
        user.totalFiles += action.outputFiles.length;
        user.totalDuration += action.duration;

        if (action.status === "failed") {
          journeyStep.status = "failed";
          journeyStep.error = action.error;
          user.status = "failed";
        } else if (journeyStep.status !== "failed") {
          journeyStep.status = "success";
          if (user.status !== "failed") {
            user.completedSteps = user.journey.filter(j => j.status === "success").length;
            if (user.completedSteps === user.totalSteps) {
              user.status = "completed";
            }
          }
        }
      });
    });

    // Sort journey steps by step order
    userMap.forEach(user => {
      user.journey.sort((a, b) => {
        const stepA = reportData.find(s => s.id === a.stepId);
        const stepB = reportData.find(s => s.id === b.stepId);
        return (stepA?.stepOrder || 0) - (stepB?.stepOrder || 0);
      });
    });

    return Array.from(userMap.values()).sort((a, b) => a.loopIndex - b.loopIndex);
  });

  // Filter users based on search term
  const filteredUsers = $derived.by(() => {
    if (!searchTerm) return userData;
    
    const term = searchTerm.toLowerCase();
    return userData.filter(user => 
      `user ${user.loopIndex}`.includes(term) ||
      user.status.includes(term) ||
      user.journey.some(step => step.stepName.toLowerCase().includes(term))
    );
  });

  function selectUser(user: UserData) {
    selectedUser = user;
  }

  function openImageViewer(stepImages: string[], startIndex: number = 0) {
    currentStepImages = stepImages;
    currentImageIndex = startIndex;
    showImageViewer = true;
  }

  function getStatusIcon(status: string) {
    switch (status) {
      case "success":
        return CheckCircleOutline;
      case "failed":
        return XCircleOutline;
      case "skipped":
        return "⏭️";
      case "in_progress":
        return ClockOutline;
      default:
        return ClockOutline;
    }
  }

  function getStatusColor(status: string) {
    switch (status) {
      case "completed":
        return "text-green-600";
      case "failed":
        return "text-red-600";
      case "in_progress":
        return "text-blue-600";
      default:
        return "text-gray-600";
    }
  }

  function getStepStatusColor(status: string) {
    switch (status) {
      case "success":
        return "text-green-600 bg-green-50";
      case "failed":
        return "text-red-600 bg-red-50";
      case "skipped":
        return "text-yellow-600 bg-yellow-50";
      case "in_progress":
        return "text-blue-600 bg-blue-50";
      default:
        return "text-gray-600 bg-gray-50";
    }
  }

  // Filter images from output files
  function getImageFiles(files: string[]): string[] {
    return files.filter(file => {
      const ext = file.split('.').pop()?.toLowerCase();
      return ['png', 'jpg', 'jpeg', 'gif', 'webp'].includes(ext || '');
    });
  }
</script>

<Modal bind:open outsideclose={true} class="w-full max-w-7xl" size="xl">
  <div class="flex h-[80vh]">
    <!-- Pane 1: User Selector (Left Side) -->
    <div class="w-1/3 border-r border-gray-200 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <h3 class="text-lg font-semibold text-gray-900 mb-3">User Explorer</h3>
        <div class="relative">
          <SearchOutline class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <Input
            type="text"
            bind:value={searchTerm}
            placeholder="Search users..."
            class="pl-10"
          />
        </div>
      </div>
      
      <div class="flex-1 overflow-y-auto p-4">
        <div class="space-y-2">
          {#each filteredUsers as user (user.loopIndex)}
            <button
              onclick={() => selectUser(user)}
              class="w-full text-left p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors {selectedUser?.loopIndex === user.loopIndex ? 'bg-blue-50 border-blue-300' : ''}"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center space-x-3">
                  <UserOutline class="h-5 w-5 text-gray-400" />
                  <div>
                    <p class="font-medium text-gray-900">User {user.loopIndex}</p>
                    <p class="text-sm {getStatusColor(user.status)}">
                      {user.status.charAt(0).toUpperCase() + user.status.slice(1)}
                    </p>
                  </div>
                </div>
                <div class="text-right">
                  <p class="text-sm text-gray-600">{user.completedSteps}/{user.totalSteps} steps</p>
                  <p class="text-xs text-gray-500">{formatDuration(user.totalDuration)}</p>
                </div>
              </div>
            </button>
          {/each}
        </div>
      </div>
    </div>

    <!-- Pane 2: Selected User's Journey (Right Side) -->
    <div class="w-2/3 flex flex-col">
      {#if selectedUser}
        <div class="p-4 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">
                User {selectedUser.loopIndex} - {selectedUser.status.toUpperCase()}
              </h3>
              <p class="text-sm text-gray-600">
                {selectedUser.completedSteps}/{selectedUser.totalSteps} steps completed • 
                {formatDuration(selectedUser.totalDuration)} total • 
                {selectedUser.totalFiles} files
              </p>
            </div>
            <div class="flex items-center space-x-2">
              <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {selectedUser.status === 'completed' ? 'bg-green-100 text-green-800' : selectedUser.status === 'failed' ? 'bg-red-100 text-red-800' : 'bg-blue-100 text-blue-800'}">
                {selectedUser.status}
              </span>
            </div>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto p-4">
          <!-- Vertical Timeline -->
          <div class="space-y-4">
            {#each selectedUser.journey as step, index (step.stepId)}
              <div class="flex">
                <!-- Timeline Line -->
                <div class="flex flex-col items-center mr-4">
                  <div class="flex items-center justify-center w-8 h-8 rounded-full {getStepStatusColor(step.status)}">
                    {#if step.status === "success"}
                      <CheckCircleOutline class="h-5 w-5" />
                    {:else if step.status === "failed"}
                      <XCircleOutline class="h-5 w-5" />
                    {:else if step.status === "skipped"}
                      <span class="text-sm">⏭️</span>
                    {:else}
                      <ClockOutline class="h-5 w-5" />
                    {/if}
                  </div>
                  {#if index < selectedUser.journey.length - 1}
                    <div class="w-0.5 h-8 bg-gray-300 mt-2"></div>
                  {/if}
                </div>

                <!-- Step Content -->
                <div class="flex-1 pb-8">
                  <div class="bg-white border border-gray-200 rounded-lg p-4">
                    <div class="flex items-center justify-between mb-2">
                      <h4 class="font-medium text-gray-900">{step.stepName}</h4>
                      <span class="text-sm text-gray-500">{formatDuration(step.duration)}</span>
                    </div>

                    {#if step.error}
                      <div class="mb-3 p-3 bg-red-50 border border-red-200 rounded-md">
                        <p class="text-sm font-medium text-red-800">Error:</p>
                        <p class="text-sm text-red-700">{step.error}</p>
                      </div>
                    {/if}

                    {#if step.outputFiles.length > 0}
                      {@const imageFiles = getImageFiles(step.outputFiles)}
                      <div class="mt-3">
                        <p class="text-sm font-medium text-gray-700 mb-2">
                          Files ({step.outputFiles.length})
                          {#if imageFiles.length > 0}
                            • {imageFiles.length} images
                          {/if}
                        </p>
                        
                        {#if imageFiles.length > 0}
                          <div class="grid grid-cols-4 gap-2 mb-3">
                            {#each imageFiles.slice(0, 8) as imageUrl, imgIndex (imageUrl)}
                              <button
                                onclick={() => openImageViewer(imageFiles, imgIndex)}
                                class="aspect-square bg-gray-100 rounded-md overflow-hidden hover:ring-2 hover:ring-blue-500 transition-all"
                              >
                                <img
                                  src={imageUrl}
                                  alt="Step output {imgIndex + 1}"
                                  class="w-full h-full object-cover"
                                  loading="lazy"
                                />
                              </button>
                            {/each}
                            {#if imageFiles.length > 8}
                              <button
                                onclick={() => openImageViewer(imageFiles, 0)}
                                class="aspect-square bg-gray-200 rounded-md flex items-center justify-center text-sm font-medium text-gray-600 hover:bg-gray-300 transition-colors"
                              >
                                +{imageFiles.length - 8}
                              </button>
                            {/if}
                          </div>
                        {/if}

                        <!-- Non-image files -->
                        {#if step.outputFiles.length > imageFiles.length}
                          <div class="space-y-1">
                            {#each step.outputFiles.filter(f => !getImageFiles([f]).length) as fileUrl}
                              <a
                                href={fileUrl}
                                target="_blank"
                                rel="noopener noreferrer"
                                class="block text-sm text-blue-600 hover:text-blue-800 hover:underline"
                              >
                                📄 {fileUrl.split('/').pop() || 'Unknown file'}
                              </a>
                            {/each}
                          </div>
                        {/if}
                      </div>
                    {/if}

                    <div class="mt-2 text-xs text-gray-500">
                      {new Date(step.timestamp).toLocaleString()}
                    </div>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {:else}
        <div class="flex-1 flex items-center justify-center">
          <div class="text-center text-gray-500">
            <UserOutline class="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p class="text-lg font-medium">Select a user to view their journey</p>
            <p class="text-sm">Choose a user from the left panel to see their step-by-step execution details</p>
          </div>
        </div>
      {/if}
    </div>
  </div>

  <div class="flex justify-end p-4 border-t border-gray-200">
    <Button color="alternative" onclick={onClose}>
      Close
    </Button>
  </div>
</Modal>

<!-- Image Viewer Modal -->
<ImageViewerModal
  bind:open={showImageViewer}
  imageUrls={currentStepImages}
  startIndex={currentImageIndex}
  onClose={() => showImageViewer = false}
/>