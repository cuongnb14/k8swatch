# K8swatch
This module is designed to check the state of all pods every minute and send a notification via a specified channel (e.g., Discord, Slack) for every detected restart of a pod.

# Setup

- Update webhook URLs in deployment.yaml:
  - `DISCORD_WEBHOOK_URL` for Discord notifications (optional)
  - `SLACK_WEBHOOK_URL` for Slack notifications (optional)
  - At least one webhook URL must be provided
- Apply
```
kubectl apply -f deployment.yaml
```

## Webhook Setup

### Discord
1. Create a webhook in your Discord server settings
2. Copy the webhook URL and update `DISCORD_WEBHOOK_URL` in deployment.yaml

### Slack
1. Create a Slack app at https://api.slack.com/apps
2. Enable Incoming Webhooks and create a webhook for your channel
3. Copy the webhook URL and update `SLACK_WEBHOOK_URL` in deployment.yaml
