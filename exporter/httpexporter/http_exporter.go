package httpexporter // import github.com/smnzlnsk/opentelemetry-components/exporter/httpexporter

import (
	"bytes"
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type httpExporter struct {
	config *Config
	logger *zap.Logger
	*marshaler
	host   component.Host
	cancel context.CancelFunc
}

func newHTTPExporter(cfg *Config, logger *zap.Logger) (*httpExporter, error) {
	exporter := &httpExporter{
		logger: logger,
		config: cfg,
	}
	return exporter, nil
}

func (he *httpExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data, err := he.metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		he.logger.Error("failed to marshal metrics", zap.Error(err))
		return err
	}

	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://%s:%d/v1/metrics", he.config.Endpoint.IP, he.config.Endpoint.Port),
		bytes.NewBuffer(data),
	)
	if err != nil {
		he.logger.Error("failed to create request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "application/x-protobuf")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			he.logger.Error("request canceled or timed out", zap.Error(err))
			return ctx.Err()
		default:
			he.logger.Error("failed to send POST request", zap.Error(err))
			return err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (he *httpExporter) start(ctx context.Context, host component.Host) error {
	ctx, cancel := context.WithCancel(ctx)
	he.cancel = cancel

	marshaler, err := newMarshaler()
	if err != nil {
		return err
	}
	he.marshaler = marshaler
	he.host = host
	return nil
}

func (he *httpExporter) shutdown(_ context.Context) error {
	if he.cancel != nil {
		he.cancel()
	}
	return nil
}
