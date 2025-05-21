package socketreader

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
)

// NewSocketReader creates a new socket reader based on protocol detection
func NewSocketReader(socketPath string, logger *zap.Logger) (SocketReader, error) {
	// Try HTTP first
	httpReader := NewHTTPSocketReader(socketPath, logger)
	data, err := httpReader.Read(context.Background())
	if err == nil && len(data) > 0 {
		logger.Info("Detected HTTP protocol on socket", zap.String("socket", socketPath))
		return httpReader, nil
	}

	// Fall back to raw socket
	logger.Info("HTTP protocol not detected, falling back to raw socket reading",
		zap.String("socket", socketPath),
		zap.Error(err))

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to socket: %w", err)
	}

	return NewRawSocketReader(conn, socketPath, logger), nil
}
