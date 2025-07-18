<script lang="ts">
  import { showSuccessToast, showErrorToast } from "$lib/utils/toast";
  import AutomationFormModal from "$lib/components/AutomationFormModal.svelte";
  import StepFormModal from "$lib/components/StepFormModal.svelte";
  import ActionFormModal from "$lib/components/ActionFormModal.svelte";
  import ConfirmDeleteModal from "$lib/components/ConfirmDeleteModal.svelte";
  import { formatDate } from "$lib/utils/date";
  import { router } from "@inertiajs/svelte";

  type Project = {
    ID: string;
    Name: string;
  };

  type Automation = {
    ID: string;
    Name: string;
    Description: string;
    ConfigJSON: string;
    CreatedAt: string;
    UpdatedAt: string;
  };

  type Step = {
    ID: string;
    Name: string;
    StepOrder: number;
    CreatedAt: string;
    Actions: Action[]; // Nested actions
  };

  type Action = {
    ID: string;
    ActionType: string;
    ActionConfigJSON: string;
    ActionOrder: number;
    CreatedAt: string;
  };

  type Props = {
    project: Project;
    automation: Automation;
    steps: { step: Step; actions: Action[]; maxActionOrder: number }[]; // Backend sends steps with nested actions and max order
    maxStepOrder: number; // Maximum step order for this automation
    user: any;
    params: Record<string, string>;
  };

  let { project, automation, steps, maxStepOrder, params }: Props = $props();
  const { projectId, automationId } = params;

  let showEditAutomationModal = $state(false);
  let showDeleteAutomationConfirm = $state(false);
  let isDeletingAutomation = $state(false);

  let showCreateStepModal = $state(false);
  let showEditStepModal = $state(false);
  let showDeleteStepConfirm = $state(false);
  let selectedStep = $state<Step | null>(null);
  let isDeletingStep = $state(false);

  let showCreateActionModal = $state(false);
  let showEditActionModal = $state(false);
  let showDeleteActionConfirm = $state(false);
  let selectedAction = $state<Action | null>(null);
  let isDeletingAction = $state(false);
  let currentStepForAction = $state<Step | null>(null); // To know which step an action belongs to
  let currentMaxActionOrder = $state(0); // To track max action order for the current step

  // --- Automation Handlers ---
  function openEditAutomationModal() {
    showEditAutomationModal = true;
  }

  function openDeleteAutomationConfirm() {
    showDeleteAutomationConfirm = true;
  }

  async function handleSaveAutomation(data: {
    name: string;
    description: string;
    config_json: string;
  }) {
    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${automationId}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(data),
        }
      );

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Automation updated successfully");
        // Update local automation prop
        automation.Name = result.automation.Name;
        automation.Description = result.automation.Description;
        automation.ConfigJSON = result.automation.ConfigJSON;
        automation.UpdatedAt = result.automation.UpdatedAt;
      } else {
        throw result;
      }
    } catch (err: any) {
      console.error("Failed to save automation:", err);
      if (err.errors) {
        throw err;
      } else {
        showErrorToast(err.message || "Failed to save automation");
        throw new Error(err.message || "Failed to save automation");
      }
    }
  }

  async function handleDeleteAutomation() {
    isDeletingAutomation = true;
    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${automationId}`,
        {
          method: "DELETE",
        }
      );

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Automation deleted successfully");
        router.visit(`/projects/${projectId}/automations`); // Redirect to automations list
      } else {
        showErrorToast(result.error || "Failed to delete automation");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isDeletingAutomation = false;
      showDeleteAutomationConfirm = false;
    }
  }

  async function handleTriggerRun() {
    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${automationId}/runs`,
        {
          method: "POST",
        }
      );

      const result = await response.json();
      console.log(result);

      if (response.ok) {
        showSuccessToast("Automation run triggered successfully!");
        // Optionally redirect to runs page or update runs list
        router.visit(`/projects/${projectId}/automations/${automationId}/runs/${result.run.ID}`);
      } else {
        showErrorToast(result.error || "Failed to trigger automation run");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    }
  }

  // --- Step Handlers ---
  function openCreateStepModal() {
    selectedStep = null; // Clear for creation
    showCreateStepModal = true;
  }

  function openEditStepModal(step: Step) {
    selectedStep = step;
    showEditStepModal = true;
  }

  function openDeleteStepConfirm(step: Step) {
    selectedStep = step;
    showDeleteStepConfirm = true;
  }

  async function handleSaveStep(data: { name: string; step_order: number }) {
    try {
      const method = selectedStep ? "PUT" : "POST";
      const url = selectedStep
        ? `/projects/${projectId}/automations/${automationId}/steps/${selectedStep.ID}`
        : `/projects/${projectId}/automations/${automationId}/steps`;

      const response = await fetch(url, {
        method: method,
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (response.ok) {
        showSuccessToast(
          selectedStep
            ? "Step updated successfully"
            : "Step created successfully"
        );
        // Refresh page to get updated steps and actions
        window.location.reload();
      } else {
        throw result;
      }
    } catch (err: any) {
      console.error("Failed to save step:", err);
      if (err.errors) {
        throw err;
      } else {
        showErrorToast(err.message || "Failed to save step");
        throw new Error(err.message || "Failed to save step");
      }
    }
  }

  async function handleDeleteStep() {
    if (!selectedStep) return;

    isDeletingStep = true;
    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${automationId}/steps/${selectedStep.ID}`,
        {
          method: "DELETE",
        }
      );

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Step deleted successfully");
        // Filter out the deleted step and its actions
        steps = steps.filter((s) => s.step.ID !== selectedStep?.ID);
        selectedStep = null;
      } else {
        showErrorToast(result.error || "Failed to delete step");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isDeletingStep = false;
      showDeleteStepConfirm = false;
    }
  }

  // --- Action Handlers ---
  function openCreateActionModal(step: Step) {
    currentStepForAction = step;
    // Find the max action order for this step
    const stepData = steps.find(s => s.step.ID === step.ID);
    currentMaxActionOrder = stepData?.maxActionOrder || 0;
    selectedAction = null; // Clear for creation
    showCreateActionModal = true;
  }

  function openEditActionModal(step: Step, action: Action) {
    currentStepForAction = step;
    // Find the max action order for this step
    const stepData = steps.find(s => s.step.ID === step.ID);
    currentMaxActionOrder = stepData?.maxActionOrder || 0;
    selectedAction = action;
    showEditActionModal = true;
  }

  function openDeleteActionConfirm(step: Step, action: Action) {
    currentStepForAction = step;
    selectedAction = action;
    showDeleteActionConfirm = true;
  }

  async function handleSaveAction(data: {
    action_type: string;
    action_config_json: string;
    action_order: number;
  }) {
    if (!currentStepForAction) return;

    try {
      const method = selectedAction ? "PUT" : "POST";
      const url = selectedAction
        ? `/projects/${projectId}/automations/${automationId}/steps/${currentStepForAction.ID}/actions/${selectedAction.ID}`
        : `/projects/${projectId}/automations/${automationId}/steps/${currentStepForAction.ID}/actions`;

      const response = await fetch(url, {
        method: method,
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (response.ok) {
        showSuccessToast(
          selectedAction
            ? "Action updated successfully"
            : "Action created successfully"
        );
        // Refresh page to get updated steps and actions
        window.location.reload();
      } else {
        throw result;
      }
    } catch (err: any) {
      console.error("Failed to save action:", err);
      if (err.errors) {
        throw err;
      } else {
        showErrorToast(err.message || "Failed to save action");
        throw new Error(err.message || "Failed to save action");
      }
    }
  }

  async function handleDeleteAction() {
    if (!currentStepForAction || !selectedAction) return;

    isDeletingAction = true;
    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${automationId}/steps/${currentStepForAction.ID}/actions/${selectedAction.ID}`,
        {
          method: "DELETE",
        }
      );

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Action deleted successfully");
        // Filter out the deleted action from the correct step
        steps = steps.map((s) => {
          if (s.step.ID === currentStepForAction?.ID) {
            s.actions = s.actions.filter((a) => a.ID !== selectedAction?.ID);
          }
          return s;
        });
        selectedAction = null;
        currentStepForAction = null;
      } else {
        showErrorToast(result.error || "Failed to delete action");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    } finally {
      isDeletingAction = false;
      showDeleteActionConfirm = false;
    }
  }

  // --- Move Handlers ---
  async function handleMoveStep(step: Step, direction: 'up' | 'down') {
    const newOrder = direction === 'up' ? step.StepOrder - 1 : step.StepOrder + 1;
    
    // Validate bounds
    if (newOrder < 1 || newOrder > maxStepOrder) {
      return; // Invalid move
    }

    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${automationId}/steps/${step.ID}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            name: step.Name,
            step_order: newOrder,
          }),
        }
      );

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Step order updated");
        // Refresh page to get updated order
        window.location.reload();
      } else {
        showErrorToast(result.error || "Failed to update step order");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    }
  }

  async function handleMoveAction(step: Step, action: Action, direction: 'up' | 'down') {
    const stepData = steps.find(s => s.step.ID === step.ID);
    const maxActionOrderForStep = stepData?.maxActionOrder || 0;
    const newOrder = direction === 'up' ? action.ActionOrder - 1 : action.ActionOrder + 1;
    
    // Validate bounds
    if (newOrder < 1 || newOrder > maxActionOrderForStep) {
      return; // Invalid move
    }

    try {
      const response = await fetch(
        `/projects/${projectId}/automations/${automationId}/steps/${step.ID}/actions/${action.ID}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            action_type: action.ActionType,
            action_config_json: action.ActionConfigJSON,
            action_order: newOrder,
          }),
        }
      );

      const result = await response.json();

      if (response.ok) {
        showSuccessToast("Action order updated");
        // Refresh page to get updated order
        window.location.reload();
      } else {
        showErrorToast(result.error || "Failed to update action order");
      }
    } catch (err: any) {
      showErrorToast("Network error. Please try again.");
    }
  }
</script>

<svelte:head>
  <title>{automation.Name} - QPlayground</title>
</svelte:head>

<div class="px-4 py-6 sm:px-0">
  <!-- Automation Header -->
  <div class="md:flex md:items-center md:justify-between mb-6">
    <div class="flex-1 min-w-0">
      <h2
        class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate"
      >
        {automation.Name}
      </h2>
      <p class="mt-2 text-sm text-gray-600">
        Project: <a
          href="/projects/{project.ID}"
          class="text-primary-600 hover:underline">{project.Name}</a
        >
      </p>
      {#if automation.Description}
        <p class="mt-2 text-sm text-gray-600">{automation.Description}</p>
      {/if}
      <p class="mt-1 text-sm text-gray-500">
        Created: {formatDate(automation.CreatedAt)} | Last Updated: {formatDate(
          automation.UpdatedAt
        )}
      </p>
    </div>
    <div class="mt-4 flex md:mt-0 md:ml-4">
      <button
        onclick={handleTriggerRun}
        class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
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
            d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
          />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        Run Automation
      </button>
      <button
        onclick={openEditAutomationModal}
        class="ml-3 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        <svg
          class="-ml-1 mr-2 h-5 w-5 text-gray-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
          />
        </svg>
        Edit Automation
      </button>
      <button
        onclick={openDeleteAutomationConfirm}
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
            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
          />
        </svg>
        Delete Automation
      </button>
    </div>
  </div>

  <!-- Automation Config -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
    <h3 class="text-lg leading-6 font-medium text-gray-900 mb-4">
      Configuration
    </h3>
    <pre
      class="bg-gray-100 p-4 rounded-md text-sm overflow-auto">{JSON.stringify(
        JSON.parse(automation.ConfigJSON),
        null,
        2
      )}</pre>
  </div>

  <!-- Steps Section -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6 mb-6">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg leading-6 font-medium text-gray-900">Steps</h3>
      <button
        onclick={openCreateStepModal}
        class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
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
            d="M12 6v6m0 0v6m0-6h6m-6 0H6"
          />
        </svg>
        New Step
      </button>
    </div>

    {#if steps?.length === 0}
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
            d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
          />
        </svg>
        <h3 class="mt-2 text-sm font-medium text-gray-900">No steps defined</h3>
        <p class="mt-1 text-sm text-gray-500">
          Add steps to define the sequence of actions for this automation.
        </p>
      </div>
    {:else}
      <ul role="list" class="divide-y divide-gray-200">
        {#each steps as { step, actions, maxActionOrder } (step.ID)}
          <li class="py-4">
            <div class="flex justify-between items-center mb-2">
              <h4 class="text-lg font-medium text-gray-900">
                {step.StepOrder}. {step.Name}
              </h4>
              <div class="flex space-x-3">
                <button
                  onclick={() => handleMoveStep(step, 'up')}
                  disabled={step.StepOrder <= 1}
                  class="text-sm font-medium text-gray-600 hover:text-gray-900 disabled:text-gray-400 disabled:cursor-not-allowed"
                  title="Move step up"
                >
                  ↑
                </button>
                <button
                  onclick={() => handleMoveStep(step, 'down')}
                  disabled={step.StepOrder >= maxStepOrder}
                  class="text-sm font-medium text-gray-600 hover:text-gray-900 disabled:text-gray-400 disabled:cursor-not-allowed"
                  title="Move step down"
                >
                  ↓
                </button>
                <button
                  onclick={() => openCreateActionModal(step)}
                  class="text-sm font-medium text-primary-600 hover:text-primary-800"
                >
                  Add Action
                </button>
                <button
                  onclick={() => openEditStepModal(step)}
                  class="text-sm font-medium text-gray-600 hover:text-gray-900"
                >
                  Edit
                </button>
                <button
                  onclick={() => openDeleteStepConfirm(step)}
                  class="text-sm font-medium text-red-600 hover:text-red-900"
                >
                  Delete
                </button>
              </div>
            </div>
            {#if actions?.length === 0}
              <p class="text-sm text-gray-500 ml-6">No actions in this step.</p>
            {:else}
              <ul
                role="list"
                class="divide-y divide-gray-100 border-t border-gray-100 mt-2"
              >
                {#each actions as action (action.ID)}
                  <li class="py-3 flex justify-between items-center ml-6">
                    <div>
                      <p class="text-sm font-medium text-gray-700">
                        {action.ActionOrder}. {action.ActionType}
                      </p>
                      <pre
                        class="bg-gray-50 p-2 rounded-md text-xs overflow-auto max-w-md">{JSON.stringify(
                          JSON.parse(action.ActionConfigJSON),
                          null,
                          2
                        )}</pre>
                    </div>
                    <div class="flex space-x-3">
                      <button
                        onclick={() => handleMoveAction(step, action, 'up')}
                        disabled={action.ActionOrder <= 1}
                        class="text-sm font-medium text-gray-600 hover:text-gray-900 disabled:text-gray-400 disabled:cursor-not-allowed"
                        title="Move action up"
                      >
                        ↑
                      </button>
                      <button
                        onclick={() => handleMoveAction(step, action, 'down')}
                        disabled={action.ActionOrder >= maxActionOrder}
                        class="text-sm font-medium text-gray-600 hover:text-gray-900 disabled:text-gray-400 disabled:cursor-not-allowed"
                        title="Move action down"
                      >
                        ↓
                      </button>
                      <button
                        onclick={() => openEditActionModal(step, action)}
                        class="text-sm font-medium text-gray-600 hover:text-gray-900"
                      >
                        Edit
                      </button>
                      <button
                        onclick={() => openDeleteActionConfirm(step, action)}
                        class="text-sm font-medium text-red-600 hover:text-red-900"
                      >
                        Delete
                      </button>
                    </div>
                  </li>
                {/each}
              </ul>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  <!-- Runs Section -->
  <div class="bg-white shadow overflow-hidden sm:rounded-lg p-6">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg leading-6 font-medium text-gray-900">Recent Runs</h3>
      <a
        href="/projects/{projectId}/automations/{automationId}/runs"
        class="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
      >
        View All Runs
      </a>
    </div>
    <!-- Placeholder for recent runs, actual list will be on /runs page -->
    <div class="text-center py-8">
      <p class="text-sm text-gray-500">
        No recent runs to display. Trigger an automation run to see results.
      </p>
    </div>
  </div>
</div>

<!-- Modals -->
<AutomationFormModal
  bind:open={showEditAutomationModal}
  {automation}
  onSave={handleSaveAutomation}
  onClose={() => (showEditAutomationModal = false)}
/>

<ConfirmDeleteModal
  bind:open={showDeleteAutomationConfirm}
  title="Delete Automation"
  message="Are you sure you want to delete '{automation.Name}'? All associated steps, actions, and runs will also be deleted. This action cannot be undone."
  onConfirm={handleDeleteAutomation}
  onCancel={() => (showDeleteAutomationConfirm = false)}
  loading={isDeletingAutomation}
/>

<StepFormModal
  bind:open={showCreateStepModal}
  maxOrder={maxStepOrder}
  onSave={handleSaveStep}
  onClose={() => (showCreateStepModal = false)}
/>

<StepFormModal
  bind:open={showEditStepModal}
  step={selectedStep}
  maxOrder={maxStepOrder}
  onSave={handleSaveStep}
  onClose={() => (showEditStepModal = false)}
/>

<ConfirmDeleteModal
  bind:open={showDeleteStepConfirm}
  title="Delete Step"
  message="Are you sure you want to delete '{selectedStep?.Name}'? All associated actions will also be deleted. This action cannot be undone."
  onConfirm={handleDeleteStep}
  onCancel={() => (showDeleteStepConfirm = false)}
  loading={isDeletingStep}
/>

<ActionFormModal
  bind:open={showCreateActionModal}
  maxOrder={currentMaxActionOrder}
  onSave={handleSaveAction}
  onClose={() => (showCreateActionModal = false)}
/>

<ActionFormModal
  bind:open={showEditActionModal}
  action={selectedAction}
  maxOrder={currentMaxActionOrder}
  onSave={handleSaveAction}
  onClose={() => (showEditActionModal = false)}
/>

<ConfirmDeleteModal
  bind:open={showDeleteActionConfirm}
  title="Delete Action"
  message="Are you sure you want to delete this action? This action cannot be undone."
  onConfirm={handleDeleteAction}
  onCancel={() => (showDeleteActionConfirm = false)}
  loading={isDeletingAction}
/>
