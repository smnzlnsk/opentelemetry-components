
CURL=curl

REPO=github.com/smnzlnsk/opentelemetry-components
GOPROXY=https://proxy.golang.org

BACKEND_EXPORTER_MODULE=exporters/backend
MQTT_RECEIVER_MODULE=receivers/mqtt
MQTT_EXPORTER_MODULE=exporters/mqtt

VERSION=v0.0.0

.PHONY: check-proxy
check-proxy:
	$(CURL) $(GOPROXY)/$(REPO)/$(BACKEND_EXPORTER_MODULE)/@v/$(VERSION).info
	$(CURL) $(GOPROXY)/$(REPO)/$(MQTT_EXPORTER_MODULE)/@v/$(VERSION).info
	$(CURL) $(GOPROXY)/$(REPO)/$(MQTT_RECEIVER_MODULE)/@v/$(VERSION).info

