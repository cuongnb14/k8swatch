# K8swatch
This module is designed to check the state of all pods every minute and send a notification via a specified channel (e.g., Discord) for every detected restart of a pod.

# Setup

- update `DISCORD_WEBHOOK_URL` in deployment.yaml
- Apply
```
kubectl apply -f deployment.yaml
```
