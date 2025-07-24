<script lang="ts">
  import { Label, Select } from "flowbite-svelte";

  type StepConfig = {
    skip_on_condition?: string;
    run_only_on_condition?: string;
  };

  let { config = $bindable() }: { config: StepConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: StepConfig) {
    if (!targetConfig.skip_on_condition) targetConfig.skip_on_condition = "";
    if (!targetConfig.run_only_on_condition) targetConfig.run_only_on_condition = "";
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  const conditionOptions = [
    { value: "", name: "None" },
    { value: "loop_index_is_even", name: "Loop Index is Even" },
    { value: "loop_index_is_odd", name: "Loop Index is Odd" },
    { value: "loop_index_is_prime", name: "Loop Index is Prime" },
    { value: "random", name: "Random (50% chance)" },
  ];
</script>

<div class="space-y-4">
  <div>
    <Label for="skip-condition" class="mb-2">Skip Step If</Label>
    <Select
      id="skip-condition"
      bind:value={config.skip_on_condition}
      items={conditionOptions}
    />
    <p class="text-xs text-gray-500 mt-1">
      If this condition is true, the entire step will be skipped.
    </p>
  </div>

  <div>
    <Label for="run-only-on-condition" class="mb-2">Run Step Only If</Label>
    <Select
      id="run-only-on-condition"
      bind:value={config.run_only_on_condition}
      items={conditionOptions}
    />
    <p class="text-xs text-gray-500 mt-1">
      If this condition is false, the entire step will be skipped.
    </p>
  </div>
</div>