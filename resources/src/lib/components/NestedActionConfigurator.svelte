<script lang="ts">
  import { actionConfigComponents } from "$lib/utils/actionConfigMap";

  type Props = {
    actionType: string;
    config: Record<string, any>;
  };

  let { actionType, config = $bindable() }: Props = $props();

  // Ensure config is always an object
  config = config ?? {};

  // Derived state for the current config component
  const CurrentConfigComponent = $derived(actionConfigComponents[actionType]);

  // Reset config when action type changes
  $effect(() => {
    if (actionType && !Object.keys(config).length) {
      // Only reset for new actions with empty config
      config = {};
    }
  });
</script>

{#if CurrentConfigComponent}
  <div class="border p-3 rounded-md bg-gray-50">
    <h5 class="text-sm font-semibold mb-3">Action Configuration</h5>
    <CurrentConfigComponent bind:config {actionType} />
  </div>
{:else if actionType}
  <div class="border p-3 rounded-md bg-gray-100">
    <p class="text-sm text-gray-500 italic">
      No configuration available for action type: {actionType}
    </p>
  </div>
{:else}
  <div class="border p-3 rounded-md bg-gray-100">
    <p class="text-sm text-gray-500 italic">
      Select an action type to configure its parameters.
    </p>
  </div>
{/if}
</script>