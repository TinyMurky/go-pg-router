package backend

import "net"

// Connector is the interface PGHandler depends on.
// Connect performs the full backend handshake and returns
// a connection ready to relay queries.
type Connector interface {
	// Connect creates net.Conn with PostgreSQL and pass authentication challenge,
	// the returned net.Conn is in the stage of ReadyForQuery
	Connect(params map[string]string) (net.Conn, error)
}

// PGConnector is the real implementation of Connector
// PGConnector create connection between PostgreSQL and go-pg-router,
// and pass the authentication challenge, then return the connection
type PGConnector struct {
	address  string // e.g. "localhost:5432"
	password string
}

var _ Connector = (*PGConnector)(nil)

func (c *PGConnector) Connect(params map[string]string) (net.Conn, error) {
	return nil, nil
}
