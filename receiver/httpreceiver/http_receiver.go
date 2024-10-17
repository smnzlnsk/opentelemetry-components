package httpreceiver // import github.com/smnzlnsk/opentelemetry-components/receiver/httpreceiver

import (
	"context"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"io"
	"net/http"
)

var _ receiver.Metrics = (*httpReceiver)(nil)

type httpReceiver struct {
	config    *Config
	logger    *zap.Logger
	consumer  consumer.Metrics
	marshaler *marshaler
	host      component.Host
	cancel    context.CancelFunc
}

func newHTTPReceiver(cfg *Config, logger *zap.Logger, consumer consumer.Metrics) (*httpReceiver, error) {
	r := &httpReceiver{
		config:   cfg,
		logger:   logger,
		consumer: consumer,
	}
	return r, nil
}

func (hr *httpReceiver) Start(ctx context.Context, host component.Host) error {
	ctx = context.Background()
	ctx, hr.cancel = context.WithCancel(ctx)
	marshaler, err := newMarshaler()
	if err != nil {
		return err
	}
	hr.marshaler = marshaler
	hr.host = host

	hr.startHTTPServer()

	return nil
}

func (hr *httpReceiver) Shutdown(ctx context.Context) error {
	if hr.cancel != nil {
		hr.cancel()
	}
	return nil
}

func (hr *httpReceiver) ConsumeMetrics(ctx context.Context, metrics pmetric.Metrics) error {
	if hr.consumer == nil {
		hr.logger.Error("no next consumer available, dropping metric data")
		return nil
	}

	err := hr.consumer.ConsumeMetrics(ctx, metrics)
	if err != nil {
		hr.logger.Error("failed to forward metric data", zap.Error(err))
		return err
	}
	hr.logger.Debug("successfully consumed metric data")
	return nil
}

func (hr *httpReceiver) startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/metrics", hr.metricsHandler)

	err := http.ListenAndServe(
		fmt.Sprintf("%s:%d", hr.config.Endpoint.IP, hr.config.Endpoint.Port),
		mux)
	if err != nil {
		hr.logger.Error("failed to start http server", zap.Error(err))
	}
}

// handler for /v1/metrics
// supported operations: POST
func (hr *httpReceiver) metricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		hr.logger.Error("failed to read request body", zap.Error(err))
		return
	}
	defer r.Body.Close()

	metrics, err := hr.marshaler.metricsUnmarshaler.UnmarshalMetrics(body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		hr.logger.Error("failed to unmarshal metrics", zap.Error(err))
		return
	}
	err = hr.consumer.ConsumeMetrics(r.Context(), metrics)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		hr.logger.Error("failed to consume metrics", zap.Error(err))
		return
	}
}
