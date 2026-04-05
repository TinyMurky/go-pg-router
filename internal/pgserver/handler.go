package pgserver

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
)

// PGHandler is for handling tcp connection formatted in PostgreSQL protocol.
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

// Handle can handle tcp connections from a client which were formatted in PostgreSQL protocol
func (h *PGHandler) Handle(conn net.Conn) {
	// if any error happend
	// close connection immediately
	defer conn.Close()

	startupMsgBytes, err := readSSLOrStartup(conn)
	if err != nil {
		err = fmt.Errorf("readSSLOrStartup: %w", err)
		slog.Error("PGHandler.Handle:", "error", err)
		return
	}

	// Below code if for handling StartupMessage
	// these codes will handle the first message client send to go-pg-router
	// when client try to establish connection.
	// First, client will send Startup Message, handler will parse key-value pairs from message
	// (ex: user name, which database to connect to ...)
	// and protocol version that client use.
	//
	// Second, handler will send back AuthenticationOK to client, this makes the client believe
	// that they pass the Authentication (which go-pg-router won't check)
	//
	// Third, handler will send back ReadyForQuery to client, this will tell client to
	// send the rest of SQL queries to us

	if err := h.startupMsgHandler.ReadStartupMessage(bytes.NewReader(startupMsgBytes)); err != nil {
		err = fmt.Errorf("startupMsgHandler.ReadStartupMessage: %w", err)

		slog.Error("PGHandler.Handle:", "error", err)
		return
	}

	if err := h.startupMsgHandler.WriteAuthOK(conn); err != nil {
		err = fmt.Errorf("startupMsgHandler.WriteAuthOK: %w", err)
		slog.Error("PGHandler.Handle:", "error", err)
		return
	}

	if err := h.startupMsgHandler.WriteReadyForQuery(conn); err != nil {
		err = fmt.Errorf("startupMsgHandler.WriteReadyForQuery: %w", err)

		slog.Error("PGHandler.Handle:", "error", err)
		return
	}

}
