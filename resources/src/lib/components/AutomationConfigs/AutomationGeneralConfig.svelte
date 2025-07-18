<script lang="ts">
  import {
    Label,
    Input,
    Select,
    Checkbox,
    Button,
    Textarea,
  } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import SlackNotificationConfig from "../NotificationConfigs/SlackNotificationConfig.svelte";

  type Variable = {
    key: string;
    type: "static" | "dynamic" | "environment";
    value: string;
    description?: string;
  };

  type MultiRunConfig = {
    enabled: boolean;
    mode: "sequential" | "parallel";
    count: number;
    delay: number; // delay between runs in ms
  };

  type NotificationChannelConfig = {
    id: string;
    type: "slack" | "email" | "webhook";
    onComplete: boolean;
    onError: boolean;
    config: any;
  };

  type AutomationConfig = {
    variables: Variable[];
    multirun: MultiRunConfig;
    timeout: number; // global timeout in seconds
    retries: number; // number of retries on failure
    screenshots: {
      enabled: boolean;
      onError: boolean;
      onSuccess: boolean;
      path: string;
    };
    notifications: NotificationChannelConfig[];
  };

  let { config = $bindable() }: { config: AutomationConfig } = $props();
  // Initialize config with defaults if empty
  function applyDefaults(targetConfig: AutomationConfig) {
    if (!targetConfig.variables) targetConfig.variables = [];
    if (!targetConfig.multirun) {
      targetConfig.multirun = {
        enabled: false,
        mode: "sequential",
        count: 1,
        delay: 1000,
      };
    }
    if (!targetConfig.timeout) targetConfig.timeout = 300; // 5 minutes default
    if (!targetConfig.retries) targetConfig.retries = 0;
    if (!targetConfig.screenshots) {
      targetConfig.screenshots = {
        enabled: true,
        onError: true,
        onSuccess: false,
        path: "screenshots/{{timestamp}}-{{loopIndex}}.png",
      };
    }
    if (!config?.notifications) {
      config.notifications = [];
    }
  }
  config = config ?? {};
  applyDefaults(config);
  $effect(() => {
    applyDefaults(config);
  });

  function addVariable() {
    config.variables = [
      ...config.variables,
      {
        key: "",
        type: "static",
        value: "",
        description: "",
      },
    ];
  }

  function removeVariable(index: number) {
    config.variables = config.variables.filter((_, i) => i !== index);
  }

  function addNotificationChannel() {
    const newChannel: NotificationChannelConfig = {
      id: `channel-${Date.now()}`,
      type: "slack",
      onComplete: false,
      onError: true,
      config: {},
    };
    config.notifications = [
      ...(config?.notifications?.length ? config.notifications : []),
      newChannel,
    ];
  }

  function removeNotificationChannel(index: number) {
    config.notifications = config.notifications.filter((_, i) => i !== index);
  }

  // Predefined dynamic variable options for gofakeit
  const dynamicVariableOptions = [
    { value: "{{faker.name}}", label: "Random Name" },
    { value: "{{faker.email}}", label: "Random Email" },
    { value: "{{faker.phone}}", label: "Random Phone" },
    { value: "{{faker.address}}", label: "Random Address" },
    { value: "{{faker.company}}", label: "Random Company" },
    { value: "{{faker.username}}", label: "Random Username" },
    { value: "{{faker.password}}", label: "Random Password" },
    { value: "{{faker.uuid}}", label: "Random UUID" },
    { value: "{{faker.number}}", label: "Random Number" },
    { value: "{{faker.date}}", label: "Random Date" },
  ];

  // Predefined environment variable options
  const environmentVariableOptions = [
    { value: "{{loopIndex}}", label: "Loop Index (Multi-run)" },
    { value: "{{timestamp}}", label: "Current Timestamp" },
    { value: "{{runId}}", label: "Automation Run ID" },
    { value: "{{userId}}", label: "User ID" },
    { value: "{{projectId}}", label: "Project ID" },
    { value: "{{automationId}}", label: "Automation ID" },
  ];
</script>

<div class="space-y-6">
  <!-- Variables Section -->
  <div class="border p-4 rounded-md bg-gray-50">
    <div class="flex items-center justify-between mb-4">
      <h4 class="text-md font-semibold">Variables</h4>
      <Button size="sm" onclick={addVariable}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Variable
      </Button>
    </div>

    {#if config.variables.length === 0}
      <p class="text-sm text-gray-500">
        No variables defined. Click "Add Variable" to create one.
      </p>
    {:else}
      <div class="space-y-3">
        {#each config.variables as variable, index (index)}
          <div class="border p-3 rounded-md bg-white">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-3">
              <div>
                <Label for="var-key-{index}" class="mb-1 text-xs">Key</Label>
                <Input
                  id="var-key-{index}"
                  type="text"
                  bind:value={variable.key}
                  placeholder="variableName"
                  size="sm"
                />
              </div>
              <div>
                <Label for="var-type-{index}" class="mb-1 text-xs">Type</Label>
                <Select
                  id="var-type-{index}"
                  bind:value={variable.type}
                  size="sm"
                  items={[
                    { value: "static", name: "Static" },
                    { value: "dynamic", name: "Dynamic (Faker)" },
                    { value: "environment", name: "Environment" },
                  ]}
                />
              </div>
              <div>
                <Label for="var-value-{index}" class="mb-1 text-xs">Value</Label
                >
                {#if variable.type === "dynamic"}
                  <Select
                    id="var-value-{index}"
                    bind:value={variable.value}
                    size="sm"
                    items={[
                      { value: "", name: "Select a dynamic value..." },
                      ...dynamicVariableOptions.map((opt) => ({
                        value: opt.value,
                        name: opt.label,
                      })),
                    ]}
                  />
                {:else if variable.type === "environment"}
                  <Select
                    id="var-value-{index}"
                    bind:value={variable.value}
                    size="sm"
                    items={[
                      { value: "", name: "Select an environment value..." },
                      ...environmentVariableOptions.map((opt) => ({
                        value: opt.value,
                        name: opt.label,
                      })),
                    ]}
                  />
                {:else}
                  <Input
                    id="var-value-{index}"
                    type="text"
                    bind:value={variable.value}
                    placeholder="Enter static value"
                    size="sm"
                  />
                {/if}
              </div>
              <div class="flex items-end">
                <Button
                  size="sm"
                  color="red"
                  onclick={() => removeVariable(index)}
                  class="w-full"
                >
                  <TrashBinOutline class="w-4 h-4" />
                </Button>
              </div>
            </div>
            <div class="mt-2">
              <Label for="var-desc-{index}" class="mb-1 text-xs"
                >Description (optional)</Label
              >
              <Input
                id="var-desc-{index}"
                type="text"
                bind:value={variable.description}
                placeholder="Describe what this variable is used for"
                size="sm"
              />
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Multi-Run Configuration -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-4">Multi-Run Configuration</h4>

    <div class="space-y-4">
      <div class="flex items-center">
        <Checkbox
          id="multirun-enabled"
          bind:checked={config.multirun.enabled}
        />
        <Label for="multirun-enabled" class="ml-2">Enable Multi-Run</Label>
      </div>

      {#if config.multirun.enabled}
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label for="multirun-mode" class="mb-2">Execution Mode</Label>
            <Select
              id="multirun-mode"
              bind:value={config.multirun.mode}
              items={[
                { value: "sequential", name: "Sequential" },
                { value: "parallel", name: "Parallel" },
              ]}
            />
          </div>
          <div>
            <Label for="multirun-count" class="mb-2">Run Count</Label>
            <Input
              id="multirun-count"
              type="number"
              bind:value={config.multirun.count}
              min={1}
              max={100}
            />
          </div>
          <div>
            <Label for="multirun-delay" class="mb-2"
              >Delay Between Runs (ms)</Label
            >
            <Input
              id="multirun-delay"
              type="number"
              bind:value={config.multirun.delay}
              min={0}
            />
          </div>
        </div>
      {/if}
    </div>
  </div>

  <!-- General Settings -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-4">General Settings</h4>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <Label for="timeout" class="mb-2">Global Timeout (seconds)</Label>
        <Input
          id="timeout"
          type="number"
          bind:value={config.timeout}
          min={1}
          placeholder="300"
        />
        <p class="text-xs text-gray-500 mt-1">
          Maximum time for the entire automation to complete
        </p>
      </div>
      <div>
        <Label for="retries" class="mb-2">Retry Count</Label>
        <Input
          id="retries"
          type="number"
          bind:value={config.retries}
          min={0}
          max={10}
          placeholder="0"
        />
        <p class="text-xs text-gray-500 mt-1">
          Number of times to retry on failure
        </p>
      </div>
    </div>
  </div>

  <!-- Screenshot Configuration -->
  <div class="border p-4 rounded-md bg-gray-50">
    <h4 class="text-md font-semibold mb-4">Screenshot Configuration</h4>

    <div class="space-y-4">
      <div class="flex items-center">
        <Checkbox
          id="screenshots-enabled"
          bind:checked={config.screenshots.enabled}
        />
        <Label for="screenshots-enabled" class="ml-2">Enable Screenshots</Label>
      </div>

      {#if config.screenshots.enabled}
        <div class="space-y-3">
          <div class="flex items-center space-x-4">
            <div class="flex items-center">
              <Checkbox
                id="screenshots-error"
                bind:checked={config.screenshots.onError}
              />
              <Label for="screenshots-error" class="ml-2">On Error</Label>
            </div>
            <div class="flex items-center">
              <Checkbox
                id="screenshots-success"
                bind:checked={config.screenshots.onSuccess}
              />
              <Label for="screenshots-success" class="ml-2">On Success</Label>
            </div>
          </div>
          <div>
            <Label for="screenshots-path" class="mb-2"
              >Screenshot Path Template</Label
            >
            <Input
              id="screenshots-path"
              type="text"
              bind:value={config.screenshots.path}
              placeholder={`screenshots/{{timestamp}}-{{loopIndex}}.png`}
            />
            <p class="text-xs text-gray-500 mt-1">
              {`Use variables like {{timestamp}}, {{loopIndex}}, {{runId}} in the path`}
            </p>
          </div>
        </div>
      {/if}
    </div>
  </div>

  <!-- Notification Configuration -->
  <div class="border p-4 rounded-md bg-gray-50">
    <div class="flex items-center justify-between mb-4">
      <h4 class="text-md font-semibold">Notification Channels</h4>
      <Button size="sm" onclick={addNotificationChannel}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Channel
      </Button>
    </div>

    {#if config?.notifications?.length === 0}
      <p class="text-sm text-gray-500">
        No notification channels configured. Click "Add Channel" to create one.
      </p>
    {:else}
      <div class="space-y-4">
        {#each config?.notifications as channel, index (channel.id)}
          <div class="border p-4 rounded-md bg-white">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
              <div>
                <Label for="channel-type-{index}" class="mb-2"
                  >Channel Type</Label
                >
                <Select
                  id="channel-type-{index}"
                  bind:value={channel.type}
                  items={[
                    { value: "slack", name: "Slack Webhook" },
                    { value: "email", name: "Email (Coming Soon)" },
                    { value: "webhook", name: "Generic Webhook" },
                  ]}
                />
              </div>
              <div class="flex items-center space-x-4">
                <div class="flex items-center">
                  <Checkbox
                    id="channel-complete-{index}"
                    bind:checked={channel.onComplete}
                  />
                  <Label for="channel-complete-{index}" class="ml-2"
                    >On Complete</Label
                  >
                </div>
                <div class="flex items-center">
                  <Checkbox
                    id="channel-error-{index}"
                    bind:checked={channel.onError}
                  />
                  <Label for="channel-error-{index}" class="ml-2"
                    >On Error</Label
                  >
                </div>
              </div>
              <div class="flex items-end">
                <Button
                  size="sm"
                  color="red"
                  onclick={() => removeNotificationChannel(index)}
                  class="w-full"
                >
                  <TrashBinOutline class="w-4 h-4" />
                </Button>
              </div>
            </div>

            <!-- Channel-specific configuration -->
            {#if channel.type === "slack"}
              <SlackNotificationConfig bind:config={channel.config} />
            {:else if channel.type === "webhook"}
              <div>
                <Label for="webhook-url-{index}" class="mb-2">Webhook URL</Label
                >
                <Input
                  id="webhook-url-{index}"
                  type="url"
                  bind:value={channel.config.webhook_url}
                  placeholder="https://your-webhook-url.com/notify"
                />
              </div>
            {:else if channel.type === "email"}
              <div class="text-sm text-gray-500 italic">
                Email notifications are coming soon. Please use Slack or generic
                webhook for now.
              </div>
            {/if}
          </div>
        {/each}
      </div>
      <!-- </div> -->
    {/if}
  </div>
</div>
