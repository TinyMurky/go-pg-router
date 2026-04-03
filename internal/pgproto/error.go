package pgproto

import (
	"errors"
	"fmt"
	"net"
)

var (
	// ErrConnectionClose means that tcp connection is closed,
	// it can be identified as net.ErrClosed too.
	ErrConnectionClose = fmt.Errorf("connection closed: %w", net.ErrClosed)

	// ErrInvalidMsgFormat means that msg read from client or PostgreSQL
	// does not follow PostgreSQL message format
	ErrInvalidMsgFormat = errors.New("invalid message format")
)
