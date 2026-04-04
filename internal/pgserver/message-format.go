package pgserver

// This file store "Message Formats" that was mentioned in:
// https://www.postgresql.org/docs/18/protocol-message-formats.html
// The meaning of these "Message Formats" can be found in Chapter 54.2:
// https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-START-UP

// StartUPAuthenticationOk :
// The authentication exchange is successfully completed.
func StartUPAuthenticationOk() []byte {
	return []byte{
		byte('R'),
		0x00, 0x00, 0x00, 0x08,
		0x00, 0x00, 0x00, 0x00,
	}
}

// StartUPReadyForQuery :
// Identifies the message type. ReadyForQuery is sent whenever the backend is ready for a new query cycle.
func StartUPReadyForQuery() []byte {
	return []byte{
		byte('Z'),
		0x00, 0x00, 0x00, 0x05,
		byte('I'),
	}
}
