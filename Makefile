
CURL=curl

REPO=github.com/smnzlnsk/opentelemetry-components
GOPROXY=https://proxy.golang.org

MQTT_RECEIVER_MODULE=receiver/mqttreceiver
MQTT_EXPORTER_MODULE=exporter/mqttexporter

VERSION=v0.0.1
MQTT_RECEIVER_VERSION=v0.0.1
MQTT_EXPORTER_VERSION=v0.0.2

.PHONY: check-proxy
check-proxy:
	$(CURL) $(GOPROXY)/$(REPO)/$(MQTT_EXPORTER_MODULE)/@v/$(MQTT_EXPORTER_VERSION).info
	$(CURL) $(GOPROXY)/$(REPO)/$(MQTT_RECEIVER_MODULE)/@v/$(MQTT_RECEIVER_VERSION).info
	$(CURL) $(GOPROXY)/$(REPO)/$(MQTT_EXPORTER_MODULE)/@latest
	$(CURL) $(GOPROXY)/$(REPO)/$(MQTT_RECEIVER_MODULE)/@latest

