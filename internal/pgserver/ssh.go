package pgserver

import "io"

// readSSLOrStartup reads the first message from the client.
// If it is an SSL request, replies 'N' and reads the next message.
// Returns the startup message bytes.
func readSSLOrStartup(rw io.ReadWriter) ([]byte, error) {
	return nil, nil
}
