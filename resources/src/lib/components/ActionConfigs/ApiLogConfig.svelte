<script lang="ts">
  import { Label, Textarea, Select } from "flowbite-svelte";

  type ApiLogConfig = {
    message: string;
    level?: string;
  };

  let { config = $bindable() }: { config: ApiLogConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: ApiLogConfig) {
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
      rows={4}
      bind:value={config.message}
      placeholder="Enter the message to log..."
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      {"Supports runtime variables like {{runtime.user_id}}, {{runtime.api_response.status}}, etc."}
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

  <!-- Usage Examples -->
  <div class="border p-4 rounded-md bg-blue-50 border-blue-200">
    <h4 class="text-md font-semibold mb-3 text-blue-800">Usage Examples</h4>
    <div class="text-sm text-blue-700 space-y-2">
      <p><strong>Basic logging:</strong></p>
      <p><code>Processing user registration for loop {{`{{loopIndex}}`}}</code></p>
      
      <p><strong>Runtime variable logging:</strong></p>
      <p><code>{"User ID: {{runtime.user_id}}, Status: {{runtime.api_response.status}}"}</code></p>
      
      <p><strong>Deep nested variables:</strong></p>
      <p><code>{"First item name: {{runtime.api_response.data[0].name}}"}</code></p>
      
      <p><strong>Mixed variables:</strong></p>
      <p><code>{"Loop {{loopIndex}}: Processing {{faker.name}} with ID {{runtime.user_id}}"}</code></p>
    </div>
  </div>
</div>

<style>
  code {
    background-color: #f3f4f6;
    padding: 0.125rem 0.25rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
  }
</style>