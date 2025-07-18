<script lang="ts">
  import { Label, Input, Textarea } from "flowbite-svelte";

  type SlackConfig = {
    webhook_url: string;
    channel?: string;
    username?: string;
    icon_emoji?: string;
  };

  let { config = $bindable() }: { config: SlackConfig } = $props();

  $effect(() => {
    if (!config.webhook_url) config.webhook_url = "";
    if (!config.username) config.username = "QPlayground Bot";
    if (!config.icon_emoji) config.icon_emoji = ":robot_face:";
  });
</script>

<div class="space-y-4">
  <div>
    <Label for="slack-webhook-url" class="mb-2">Slack Webhook URL *</Label>
    <Input
      id="slack-webhook-url"
      type="url"
      bind:value={config.webhook_url}
      placeholder="https://hooks.slack.com/services/..."
      required
    />
    <p class="text-xs text-gray-500 mt-1">
      Create an incoming webhook in your Slack workspace and paste the URL here
    </p>
  </div>

  <div>
    <Label for="slack-channel" class="mb-2">Channel (optional)</Label>
    <Input
      id="slack-channel"
      type="text"
      bind:value={config.channel}
      placeholder="#automation-alerts"
    />
    <p class="text-xs text-gray-500 mt-1">
      Override the default channel (include # for channels or @ for users)
    </p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <div>
      <Label for="slack-username" class="mb-2">Bot Username</Label>
      <Input
        id="slack-username"
        type="text"
        bind:value={config.username}
        placeholder="QPlayground Bot"
      />
    </div>
    <div>
      <Label for="slack-icon" class="mb-2">Bot Icon</Label>
      <Input
        id="slack-icon"
        type="text"
        bind:value={config.icon_emoji}
        placeholder=":robot_face:"
      />
    </div>
  </div>
</div>
</script>

<style>
</style>