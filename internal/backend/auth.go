package backend

import "io"

// readBackendMessage reads one backend-format message.
// Returns the type byte and the payload (excluding type and length).
func readBackendMessage(r io.Reader) (msgType byte, payload []byte, err error) {
	return 0, nil, nil
}

// doAuthHandshake sends the startup message to backend and handles
// the auth challenge until the backend sends ReadyForQuery.
func doAuthHandshake(rw io.ReadWriter, params map[string]string, password string) error {
	return nil
}
