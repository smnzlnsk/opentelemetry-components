# mqttexporter

## Configuration

The `opentelemetry-config.yaml` for the `mqttexporter` accepts the following configuration:

```yaml
exporters:
  mqtt:
    interval: 1s
    client_id: <string> 
    topic: <string>     # default: telemetry/metrics
    broker:
      host: <string>    # default: localhost
      port: <int>       # default: 1883
```

Those settings are overwritten by the respective environment variables:

- `MQTTEXPORTER_MONITORING_CLIENT_ID` overwrites `exporters::mqtt::client_id`
- `MQTTEXPORTER_MONITORING_BROKER_HOST` overwrites `exporters::mqtt::broker::host`
- `MQTTEXPORTER_MONITORING_BROKER_PORT` overwrites `exporters::mqtt::broker::port`