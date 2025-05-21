package socketreader

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPSocketReader implements SocketReader for HTTP-based sockets
type HTTPSocketReader struct {
	client     *http.Client
	socketPath string
	logger     *zap.Logger
}

// NewHTTPSocketReader creates a new HTTP socket reader
func NewHTTPSocketReader(socketPath string, logger *zap.Logger) *HTTPSocketReader {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second,
	}

	return &HTTPSocketReader{
		client:     client,
		socketPath: socketPath,
		logger:     logger,
	}
}

func (r *HTTPSocketReader) Read(ctx context.Context) ([]byte, error) {
	r.logger.Debug("Making HTTP request to socket", zap.String("socket", r.socketPath))

	resp, err := r.client.Get("http://unix/metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func (r *HTTPSocketReader) Close() error {
	return nil // HTTP client doesn't need explicit cleanup
}
