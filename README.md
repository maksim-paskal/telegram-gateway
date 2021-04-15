# Motivation
Prometheus and Sentry has no adapters for sending telegram messages for Ops peoples

# Install in your kubernetes cluster
```
helm repo add paskal-dev https://maksim-paskal.github.io/helm/
helm install --namespace telegram-gateway paskal-dev/telegram-gateway
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

# use curl to test sending messages
```
for curl messages. example:

curl -sS -X GET localhost:9090/message --get \
--data-urlencode "test=value" \
--data-urlencode "test.empty=" \
--data-urlencode "url=https://test.com" \
--data-urlencode "msg=hello world" \
--data-urlencode "url.title=Open report".
```

# how to create new telegram group with bot
* create new telegram group 
* add youp telegram bot to new group
* create config with bot token
* start application with `-enableChatServer`
* in telegram group will be shown new `chatID` in new group
