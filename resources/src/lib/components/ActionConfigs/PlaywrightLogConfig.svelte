<script lang="ts">
  import { Label, Textarea, Select } from "flowbite-svelte";

  type PlaywrightLogConfig = {
    message: string;
    level?: string;
  };

  let { config = $bindable() }: { config: PlaywrightLogConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: PlaywrightLogConfig) {
    if (!targetConfig.message) targetConfig.message = "";
    if (!targetConfig.level) targetConfig.level = "info";
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  const logLevels = [
    { value: "info", name: "Info" },
    { value: "debug", name: "Debug" },
    { value: "warn", name: "Warning" },
    { value: "error", name: "Error" },
  ];
</script>

<div class="space-y-4">
  <div>
    <Label for="log-message" class="mb-2">Log Message *</Label>
    <Textarea
      id="log-message"
      rows={3}
      bind:value={config.message}
      placeholder="Enter the message to log..."
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      This message will be logged during automation execution
    </p>
  </div>

  <div>
    <Label for="log-level" class="mb-2">Log Level</Label>
    <Select
      id="log-level"
      bind:value={config.level}
      items={logLevels}
    />
    <p class="text-xs text-gray-500 mt-1">
      Choose the severity level for this log message
    </p>
  </div>
</div>