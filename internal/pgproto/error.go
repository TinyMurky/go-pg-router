package pgproto

import (
	"errors"
	"fmt"
	"io"
)

var (
	// ErrConnectionClose means that tcp connection is closed,
	// it can be identified as EOF too.
	ErrConnectionClose = fmt.Errorf("connection closed: %w", io.EOF)

	// ErrInvaidMsgFormat means that msg read from client or PostgreSQL
	// does not follow PostgreSQL message format
	ErrInvaidMsgFormat = errors.New("invalid message format")
)
