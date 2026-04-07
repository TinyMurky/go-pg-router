package backend

import (
	"errors"
	"fmt"
	"net"
)

var (
	// ErrConnectionClosed means that tcp connection is closed,
	// it can be identified as net.ErrClosed too.
	ErrConnectionClosed = fmt.Errorf("connection closed: %w", net.ErrClosed)

	// ErrInvalidMsgFormat means that msg read from client or PostgreSQL
	// does not follow PostgreSQL message format
	ErrInvalidMsgFormat = errors.New("invalid message format")
)
