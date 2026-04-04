package listener

import (
	"errors"
	"log/slog"
	"net"
)

// TCPListener will handle connection
type TCPListener struct {
	handler Handler
}

// New will create new *TCPListener
func New(handler Handler) *TCPListener {
	return &TCPListener{
		handler: handler,
	}
}

// Start will use Handler to handle connection
func (tl *TCPListener) Start(l net.Listener) error {
	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			// if the conn is closed intentionally,
			// stop the loop
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			// TODO: use zaplogger instead
			slog.Error("Accepting Connection", "error", err)
			// Be careful, this will keep on reconnect
			continue
		}

		// when we accept (get from net.Dial)
		// we use another go routine to handle it
		go func() {
			tl.handler.Handle(conn)
		}()
	}
}
