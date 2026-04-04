package listener

import "net"

// Handler will do "Handle" to the connection
type Handler interface {
	Handle(conn net.Conn)
}
