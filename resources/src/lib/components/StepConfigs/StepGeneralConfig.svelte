<script lang="ts">
  import { Label, Select } from "flowbite-svelte";

  type StepConfig = {
    skip_condition?: string;
    run_only_condition?: string;
    probability?: number;
  };

  let { config = $bindable() }: { config: StepConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: StepConfig) {
    if (!targetConfig.probability) targetConfig.probability = 0.5;
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  const conditionOptions = [
    { value: "", name: "None (Always Execute)" },
    { value: "loop_index_is_even", name: "Skip when Loop Index is Even" },
    { value: "loop_index_is_odd", name: "Skip when Loop Index is Odd" },
    { value: "loop_index_is_prime", name: "Skip when Loop Index is Prime" },
    { value: "random", name: "Skip Randomly" },
  ];

  const runOnlyConditionOptions = [
    { value: "", name: "None (Always Execute)" },
    { value: "loop_index_is_even", name: "Run only when Loop Index is Even" },
    { value: "loop_index_is_odd", name: "Run only when Loop Index is Odd" },
    { value: "loop_index_is_prime", name: "Run only when Loop Index is Prime" },
    { value: "random", name: "Run Randomly" },
  ];

  // Clear the other condition when one is selected
  $effect(() => {
    if (config.skip_condition && config.run_only_condition) {
      // If both are set, clear run_only_condition (skip takes precedence)
      config.run_only_condition = "";
    }
  });

  const showProbability = $derived(
    config.skip_condition === "random" || config.run_only_condition === "random"
  );
</script>

<div class="space-y-4">
  <div>
    <Label for="skip-condition" class="mb-2">Skip Condition</Label>
    <Select
      id="skip-condition"
      bind:value={config.skip_condition}
      items={conditionOptions}
    />
    <p class="text-xs text-gray-500 mt-1">
      Choose when this step should be skipped during multi-run execution
    </p>
  </div>

  <div>
    <Label for="run-only-condition" class="mb-2">Run Only Condition</Label>
    <Select
      id="run-only-condition"
      bind:value={config.run_only_condition}
      items={runOnlyConditionOptions}
      disabled={!!config.skip_condition}
    />
    <p class="text-xs text-gray-500 mt-1">
      Choose when this step should run (alternative to skip condition)
    </p>
  </div>

  {#if showProbability}
    <div>
      <Label for="probability" class="mb-2">Probability (0.0 - 1.0)</Label>
      <input
        id="probability"
        type="number"
        bind:value={config.probability}
        min="0"
        max="1"
        step="0.1"
        class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
      />
      <p class="text-xs text-gray-500 mt-1">
        Probability for random condition (0.5 = 50% chance)
      </p>
    </div>
  {/if}

  {#if config.skip_condition || config.run_only_condition}
    <div class="p-3 bg-yellow-50 border border-yellow-200 rounded-md">
      <p class="text-sm text-yellow-800">
        <strong>Note:</strong> This step will be conditionally executed based on the loop index 
        during multi-run automation. Single runs (loop index 0) will follow the condition logic.
      </p>
    </div>
  {/if}
</div>