<script lang="ts">
  import { conditionConfigComponents } from "$lib/utils/actionConfigMap";

  type Props = {
    conditionType: string;
    config: Record<string, any>;
  };

  let { conditionType, config = $bindable() }: Props = $props();

  // Ensure config is always an object
  config = config ?? {};

  // Derived state for the current config component
  const CurrentConfigComponent = $derived(conditionConfigComponents[conditionType]);

  // Reset config when condition type changes
  $effect(() => {
    if (conditionType && !Object.keys(config).length) {
      // Only reset for new conditions with empty config
      config = {};
    }
  });
</script>

{#if CurrentConfigComponent}
  <div class="border p-3 rounded-md bg-gray-50">
    <h5 class="text-sm font-semibold mb-3">Condition Configuration</h5>
    {#await CurrentConfigComponent() then Component}
      <Component.default bind:config {conditionType} />
    {/await}
  </div>
{:else if conditionType && !conditionType.startsWith("loop_index_") && conditionType !== "random"}
  <div class="border p-3 rounded-md bg-gray-100">
    <p class="text-sm text-gray-500 italic">
      No configuration available for condition type: {conditionType}
    </p>
  </div>
{:else if conditionType === "random"}
  <div class="border p-3 rounded-md bg-gray-50">
    <label for="probability" class="block text-sm font-medium text-gray-700 mb-2">
      Probability (0.0 - 1.0)
    </label>
    <input
      id="probability"
      type="number"
      bind:value={config.probability}
      min="0"
      max="1"
      step="0.1"
      placeholder="0.5"
      class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm"
    />
    <p class="text-xs text-gray-500 mt-1">
      Probability for random condition (0.5 = 50% chance)
    </p>
  </div>
{:else if conditionType.startsWith("loop_index_")}
  <div class="border p-3 rounded-md bg-gray-100">
    <p class="text-sm text-gray-500 italic">
      No additional configuration needed for loop index conditions.
    </p>
  </div>
{:else}
  <div class="border p-3 rounded-md bg-gray-100">
    <p class="text-sm text-gray-500 italic">
      Select a condition type to configure its parameters.
    </p>
  </div>
{/if}