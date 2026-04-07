package backend

import (
	"net"
)

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

// New creates PGConnector that implement Connector interface
func New(address, password string) *PGConnector {
	return &PGConnector{
		address:  address,
		password: password,
	}
}

// Connect will be performed as below
//
// # The backend connection sequence
//
// go-pg-router                    Real PostgreSQL
// ────────────────────────────────────────────────────
// net.Dial("tcp", address) ──────► (TCP connection established)
//
// Send StartupMessage ────────────►
//
//	totalLength (4 bytes)
//	protocol version 0x00030000 (4 bytes)
//	"user\0alice\0database\0mydb\0\0"
//
//	                       ◄────── Authentication challenge (one of):
//	                                'R' + length + 0  → AuthenticationOK (trust
//
// auth)
//
//	'R' + length + 3  → CleartextPassword
//	'R' + length + 5  → MD5Password + 4-byte salt
//
// If password requested:
// Send PasswordMessage ───────────►
//
//	'p' + length + "password\0"
//
//	                       ◄────── 'R' + 0x00000008 + 0x00000000
//
// (AuthenticationOK)
//
//	◄────── 'S' + ... (ParameterStatus, zero or more)
//	◄────── 'K' + ... (BackendKeyData)
//	◄────── 'Z' + 0x00000005 + 'I'  (ReadyForQuery)
//
// Connection is now ready to relay queries.
func (c *PGConnector) Connect(params map[string]string) (net.Conn, error) {
	return nil, nil
}
