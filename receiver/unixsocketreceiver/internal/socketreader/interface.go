package socketreader

import (
	"context"
)

// SocketReader defines the interface for reading from a socket
type SocketReader interface {
	Read(ctx context.Context) ([]byte, error)
	Close() error
}
