# mqttreceiver

## Configuration

```yaml
receivers:
  mqtt:
    interval: 1s
    client_id: <string> 
    topic: <string>     # default: telemetry/metrics
    broker:
      host: <string>    # default: localhost
      port: <int>       # default: 1883
```

Those settings are overwritten by the respective environment variables:

- `MQTTRECEIVER_MONITORING_CLIENT_ID` overwrites `receivers::mqtt::client_id`
- `MQTTRECEIVER_MONITORING_BROKER_HOST` overwrites `receivers::mqtt::broker::host`
- `MQTTRECEIVER_MONITORING_BROKER_PORT` overwrites `receivers::mqtt::broker::port`