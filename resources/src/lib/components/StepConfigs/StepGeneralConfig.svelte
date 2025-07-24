<script lang="ts">
  import { Label, Select } from "flowbite-svelte";

  type StepConfig = {
    skip_on?: string;
    run_only_on?: string;
  };

  let { config = $bindable() }: { config: StepConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: StepConfig) {
    // No required defaults for step config, all fields are optional
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  const conditionOptions = [
    { value: "", name: "No condition" },
    { value: "loop_index_is_even", name: "Loop Index is Even" },
    { value: "loop_index_is_odd", name: "Loop Index is Odd" },
    { value: "loop_index_is_prime", name: "Loop Index is Prime" },
  ];

  // Clear the other condition when one is selected
  $effect(() => {
    if (config.skip_on && config.run_only_on) {
      // If both are set, clear run_only_on (prioritize skip_on)
      config.run_only_on = "";
    }
  });
</script>

<div class="space-y-4">
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-3">Step Execution Conditions</h4>
    <p class="text-sm text-gray-600 mb-4">
      Configure when this step should be executed or skipped based on the current loop index.
      These conditions are useful for multi-run automations where you want different behavior for different concurrent users.
    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <Label for="skip-on" class="mb-2">Skip Step When</Label>
        <Select
          id="skip-on"
          bind:value={config.skip_on}
          items={conditionOptions}
        />
        <p class="text-xs text-gray-500 mt-1">
          Skip this step entirely when the condition is met
        </p>
      </div>

      <div>
        <Label for="run-only-on" class="mb-2">Run Step Only When</Label>
        <Select
          id="run-only-on"
          bind:value={config.run_only_on}
          items={conditionOptions}
        />
        <p class="text-xs text-gray-500 mt-1">
          Only execute this step when the condition is met
        </p>
      </div>
    </div>

    {#if config.skip_on && config.run_only_on}
      <div class="mt-3 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
        <p class="text-sm text-yellow-800">
          ⚠️ Both skip and run-only conditions are set. Skip condition takes precedence.
        </p>
      </div>
    {/if}

    <div class="mt-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
      <h5 class="text-sm font-medium text-blue-800 mb-2">Examples:</h5>
      <ul class="text-xs text-blue-700 space-y-1">
        <li><strong>Even:</strong> Loop indexes 0, 2, 4, 6, 8... (useful for alternating behavior)</li>
        <li><strong>Odd:</strong> Loop indexes 1, 3, 5, 7, 9... (complementary to even)</li>
        <li><strong>Prime:</strong> Loop indexes 2, 3, 5, 7, 11... (useful for special test cases)</li>
      </ul>
    </div>
  </div>
</div>