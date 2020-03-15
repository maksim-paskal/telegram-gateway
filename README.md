# Background
Prometheus and Sentry has no adapters for sending telegram messages for Ops peoples

# Install in your kubernetes cluster
```
helm repo add paskal-dev https://maksim-paskal.github.io/helm/

# standart installation
helm install --namespace telegram-gateway paskal-dev/telegram-gateway

# installation without tiller
helm template --namespace telegram-gateway paskal-dev/telegram-gateway | kubectl apply --dry-run -f -
```

# Prometheus configuration
add this block to your Prometheus installation (values.yaml)
```yaml
alertmanagerFiles:
  alertmanager.yml:
    route:
      receiver: "prod-notify"
      group_by: ['alertname']
      group_wait:      15s
      group_interval:  15s
      repeat_interval: 15m

    inhibit_rules:
    - source_match:
        severity: 'critical'
      target_match_re:
        severity: '^(warning|info|)$'
      equal: ['alertname']

    receivers:
    - name: "prod-notify"
      webhook_configs:
      - url: 'http://telegram-gateway.telegram-gateway.svc.cluster.local:9090/prom'
        send_resolved: true
```