<script lang="ts">
  import { Label, Input, Textarea, Button, Select } from "flowbite-svelte";
  import { PlusOutline, TrashBinOutline } from "flowbite-svelte-icons";

  type AfterHook = {
    path: string;
    save_as: string;
    scope: "local" | "global";
  };

  type AuthConfig = {
    type: "bearer" | "basic" | "api_key" | "custom";
    token: string;
    header?: string;
  };

  type ApiPutConfig = {
    url: string;
    headers: Record<string, string>;
    body: string;
    timeout?: number;
    auth?: AuthConfig;
    after_hooks: AfterHook[];
  };

  let { config = $bindable() }: { config: ApiPutConfig } = $props();

  // Ensure config is always an object
  config = config ?? {};

  function applyDefaults(targetConfig: ApiPutConfig) {
    if (!targetConfig.url) targetConfig.url = "";
    if (!targetConfig.headers) targetConfig.headers = {};
    if (!targetConfig.body) targetConfig.body = "";
    if (!targetConfig.after_hooks) targetConfig.after_hooks = [];
    if (!targetConfig.timeout) targetConfig.timeout = 30000;
  }

  // Apply defaults immediately for initial render
  applyDefaults(config);

  $effect(() => {
    applyDefaults(config);
  });

  // Helper to manage headers as key-value pairs
  let headerEntries = $state<Array<{key: string, value: string}>>([]);

  $effect(() => {
    // Convert headers object to array for editing
    headerEntries = Object.entries(config.headers || {}).map(([key, value]) => ({ key, value }));
    if (headerEntries.length === 0) {
      headerEntries = [{ key: "", value: "" }];
    }
  });

  $effect(() => {
    // Convert array back to headers object
    const newHeaders: Record<string, string> = {};
    headerEntries.forEach(entry => {
      if (entry.key.trim() && entry.value.trim()) {
        newHeaders[entry.key.trim()] = entry.value.trim();
      }
    });
    config.headers = newHeaders;
  });

  function addHeader() {
    headerEntries = [...headerEntries, { key: "", value: "" }];
  }

  function removeHeader(index: number) {
    headerEntries = headerEntries.filter((_, i) => i !== index);
    if (headerEntries.length === 0) {
      headerEntries = [{ key: "", value: "" }];
    }
  }

  function addAfterHook() {
    config.after_hooks = [...config.after_hooks, { path: "", save_as: "", scope: "local" }];
  }

  function removeAfterHook(index: number) {
    config.after_hooks = config.after_hooks.filter((_, i) => i !== index);
  }

  function toggleAuth() {
    if (config.auth) {
      config.auth = undefined;
    } else {
      config.auth = { type: "bearer", token: "" };
    }
  }

  const authTypes = [
    { value: "bearer", name: "Bearer Token" },
    { value: "basic", name: "Basic Auth" },
    { value: "api_key", name: "API Key" },
    { value: "custom", name: "Custom" },
  ];

  const scopeTypes = [
    { value: "local", name: "Local (current run only)" },
    { value: "global", name: "Global (all runs)" },
  ];
</script>

<div class="space-y-4">
  <div>
    <Label for="api-url" class="mb-2">URL *</Label>
    <Input
      id="api-url"
      type="text"
      bind:value={config.url}
      placeholder={"https://api.example.com/users/{{runtime.user_id}}"}
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      {"Supports variables: {{`{runtime.baseUrl}`}}, {{`{faker.uuid}`}}, etc."}
    </p>
  </div>

  <div>
    <Label for="api-timeout" class="mb-2">Timeout (ms)</Label>
    <Input
      id="api-timeout"
      type="number"
      bind:value={config.timeout}
      placeholder="30000"
      min={1000}
    />
  </div>

  <!-- Headers Section -->
  <div class="border p-4 rounded-md bg-gray-50">
    <div class="flex items-center justify-between mb-3">
      <Label class="text-sm font-medium">Headers</Label>
      <Button size="sm" onclick={addHeader}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Header
      </Button>
    </div>

    <div class="space-y-2">
      {#each headerEntries as header, index (index)}
        <div class="grid grid-cols-5 gap-2 items-center">
          <div class="col-span-2">
            <Input
              type="text"
              bind:value={header.key}
              placeholder="Header name"
              size="sm"
            />
          </div>
          <div class="col-span-2">
            <Input
              type="text"
              bind:value={header.value}
              placeholder="Header value (supports variables)"
              size="sm"
            />
          </div>
          <div>
            <Button
              size="sm"
              color="red"
              onclick={() => removeHeader(index)}
              disabled={headerEntries.length === 1}
            >
              <TrashBinOutline class="w-4 h-4" />
            </Button>
          </div>
        </div>
      {/each}
    </div>
  </div>

  <!-- Request Body Section -->
  <div>
    <Label for="api-body" class="mb-2">Request Body</Label>
    <Textarea
      id="api-body"
      rows={6}
      bind:value={config.body}
      placeholder={'{"name": "{{faker.name}}", "email": "{{runtime.user_email}}"}'}
      class="font-mono text-sm"
    />
    <p class="text-xs text-gray-500 mt-1">
      JSON request body. Supports all variable types including runtime variables.
    </p>
  </div>

  <!-- Authentication Section -->
  <div class="border p-4 rounded-md bg-gray-50">
    <div class="flex items-center justify-between mb-3">
      <Label class="text-sm font-medium">Authentication</Label>
      <Button size="sm" onclick={toggleAuth}>
        {config.auth ? "Remove Auth" : "Add Auth"}
      </Button>
    </div>

    {#if config.auth}
      <div class="space-y-3">
        <div>
          <Label for="auth-type" class="mb-2">Auth Type</Label>
          <Select
            id="auth-type"
            bind:value={config.auth.type}
            items={authTypes}
          />
        </div>

        <div>
          <Label for="auth-token" class="mb-2">Token/Credentials</Label>
          <Input
            id="auth-token"
            type="text"
            bind:value={config.auth.token}
            placeholder={config.auth.type === "bearer" ? "{{runtime.access_token}}" : 
                        config.auth.type === "basic" ? "base64encodedcredentials" :
                        config.auth.type === "api_key" ? "{{runtime.api_key}}" : "Custom auth value"}
          />
          <p class="text-xs text-gray-500 mt-1">
            {"Supports runtime variables like {{runtime.access_token}}"}
          </p>
        </div>

        {#if config.auth.type === "api_key"}
          <div>
            <Label for="auth-header" class="mb-2">Header Name</Label>
            <Input
              id="auth-header"
              type="text"
              bind:value={config.auth.header}
              placeholder="X-API-Key"
            />
          </div>
        {/if}
      </div>
    {:else}
      <p class="text-sm text-gray-500 italic">
        No authentication configured. Runtime variables 'access_token' or 'api_key' will be auto-detected.
      </p>
    {/if}
  </div>

  <!-- After Hooks Section -->
  <div class="border p-4 rounded-md bg-green-50 border-green-200">
    <div class="flex items-center justify-between mb-3">
      <Label class="text-sm font-medium text-green-800">After Hooks (Data Extraction)</Label>
      <Button size="sm" onclick={addAfterHook}>
        <PlusOutline class="w-4 h-4 mr-2" />
        Add Hook
      </Button>
    </div>

    {#if config.after_hooks.length === 0}
      <p class="text-sm text-gray-500 italic">
        No after hooks defined. Add hooks to extract data from API responses.
      </p>
    {:else}
      <div class="space-y-3">
        {#each config.after_hooks as hook, index (index)}
          <div class="border p-3 rounded-md bg-white">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-3">
              <div>
                <Label for="hook-path-{index}" class="mb-1 text-xs">JSON Path</Label>
                <Input
                  id="hook-path-{index}"
                  type="text"
                  bind:value={hook.path}
                  placeholder="data.user.id"
                  size="sm"
                />
              </div>
              <div>
                <Label for="hook-save-as-{index}" class="mb-1 text-xs">Save As</Label>
                <Input
                  id="hook-save-as-{index}"
                  type="text"
                  bind:value={hook.save_as}
                  placeholder="user_id"
                  size="sm"
                />
              </div>
              <div>
                <Label for="hook-scope-{index}" class="mb-1 text-xs">Scope</Label>
                <Select
                  id="hook-scope-{index}"
                  bind:value={hook.scope}
                  size="sm"
                  items={scopeTypes}
                />
              </div>
              <div class="flex items-end">
                <Button
                  size="sm"
                  color="red"
                  onclick={() => removeAfterHook(index)}
                  class="w-full"
                >
                  <TrashBinOutline class="w-4 h-4" />
                </Button>
              </div>
            </div>
            <p class="text-xs text-gray-500 mt-2">
              Extract <code>{hook.path || "JSON.path"}</code> and save as runtime variable <code>{`{{runtime.${hook.save_as || "var_name"}}`}</code>
            </p>
          </div>
        {/each}
      </div>
    {/if}
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