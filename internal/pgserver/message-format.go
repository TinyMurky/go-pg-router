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

// SSLRequest will be send by client before StartupMessage if client is using SSL connecting mode
//
// ref: https://www.postgresql.org/docs/current/protocol-message-formats.html#PROTOCOL-MESSAGE-FORMATS-SSLREQUEST
func SSLRequest() []byte {
	return []byte{
		// 0x00, 0x00, 0x00, 0x08, // First 4 byte is Length of message contents in bytes, including self.
		0x04, 0xD2, 0x16, 0x2F, // The value is chosen to contain 1234 in the most significant 16 bits, and 5679 in the least significant 16 bits.
	}
}

// SSLRequestReject is used to reject to use SSL connection to connect client
// may be returned after SSLRequest
//
// ref: https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-SSL
func SSLRequestReject() []byte {
	return []byte{byte('N')}
}
