package pgproto

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
)

// PGHandler is for handling tcp connection formatted in PostgreSQL protocal.
//
// PGHandler implements the listener.Handler interface
type PGHandler struct {
	startupMsgHandler *StartupMessage
}

// NewPGHandler creates a handler
// that handles connections from clients wanting to
// connect to PostgreSQL.
//
// This handler acts as proxy to forward connections from
// client to PostgreSQL.
func NewPGHandler() *PGHandler {
	return &PGHandler{
		startupMsgHandler: &StartupMessage{},
	}
}

// Handle can hande tcp connetions from a client which were formatted in PostgreSQL protocal
func (h *PGHandler) Handle(conn net.Conn) {
	if err := h.handleStartupMessage(conn); err != nil {
		if errors.Is(err, ErrConnectionClose) {
			slog.Error("handleStartupMessage: unexpected connection closed", "error", err)
			return
		}

		if errors.Is(err, ErrInvalidMsgFormat) {
			slog.Error("handleStartupMessage: invalid msg format", "error", err)
			return
		}
		// TODO: use zaplogger instead
		slog.Error("handleStartupMessage: ", "error", err)
		return
	}
}

// handleStartupMessage will handle the first message client send to go-pg-route
// when client try to establish connection.
//
// First, client will send Startup Message, handler will parse key-value pairs from message
// (ex: user name, which database to connect to ...)
// and protocol version that client use.
//
// Second, handler will send back AuthenticationOK to client, this makes the client believe
// that they pass the Authentication (which go-pg-router won't check)
//
// Third, handler will send back ReadyForQuery to client, this will tell client to
// send the rest of SQL queries to us
func (h *PGHandler) handleStartupMessage(rw io.ReadWriter) error {
	if err := h.startupMsgHandler.ReadStartupMessage(rw); err != nil {
		if errors.Is(err, ErrInvalidMsgFormat) {
			return fmt.Errorf("startupMsgHandler.ReadStartupMessage: %w", err)
		}

		if errors.Is(err, io.EOF) {
			return fmt.Errorf("startupMsgHandler.ReadStartupMessage: %w", ErrConnectionClose)
		}

		return fmt.Errorf("startupMsgHandler.ReadStartupMessage: %w", err)
	}

	if err := h.startupMsgHandler.WriteAuthOK(rw); err != nil {
		if errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("startupMsgHandler.WriteAuthOK: %w", ErrConnectionClose)
		}
		return fmt.Errorf("startupMsgHandler.WriteAuthOK: %w", err)
	}

	if err := h.startupMsgHandler.WriteReadyForQuery(rw); err != nil {
		if errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("startupMsgHandler.WriteReadyForQuery: %w", ErrConnectionClose)
		}
		return fmt.Errorf("startupMsgHandler.WriteReadyForQuery: %w", err)
	}

	return nil
}
