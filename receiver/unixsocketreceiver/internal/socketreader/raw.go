package socketreader

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"go.uber.org/zap"
)

// RawSocketReader implements SocketReader for raw socket connections
type RawSocketReader struct {
	conn       net.Conn
	socketPath string
	logger     *zap.Logger
	buffer     []byte
}

// NewRawSocketReader creates a new raw socket reader
func NewRawSocketReader(conn net.Conn, socketPath string, logger *zap.Logger) *RawSocketReader {
	reader := &RawSocketReader{
		conn:       conn,
		socketPath: socketPath,
		logger:     logger,
		buffer:     make([]byte, 65536), // 64KB buffer
	}

	// Set the socket to non-blocking mode
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetReadBuffer(65536); err != nil {
			logger.Error("Failed to set read buffer size",
				zap.String("socket", socketPath),
				zap.Error(err))
		}
	}

	return reader
}

func (r *RawSocketReader) Read(ctx context.Context) ([]byte, error) {
	r.logger.Debug("Attempting to read from socket", zap.String("socket", r.socketPath))

	// Set a very short deadline for the read operation
	if err := r.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}
	defer r.conn.SetReadDeadline(time.Time{})

	// Read data from socket
	n, err := r.conn.Read(r.buffer)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("connection closed by peer: %w", err)
		}
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil // Return empty data for timeout
		}
		return nil, fmt.Errorf("error reading from socket: %w", err)
	}

	if n == 0 {
		return nil, nil
	}

	// Return a copy of the read data
	data := make([]byte, n)
	copy(data, r.buffer[:n])
	return data, nil
}

func (r *RawSocketReader) Close() error {
	return r.conn.Close()
}
