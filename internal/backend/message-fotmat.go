package backend

import (
	"bytes"
	"encoding/binary"
)

// This file store "Message Formats" that was mentioned in:
// https://www.postgresql.org/docs/18/protocol-message-formats.html
// The meaning of these "Message Formats" can be found in Chapter 54.2:
// https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-START-UP

// AuthenticationOk :
// The authentication exchange is successfully completed.
func AuthenticationOk() []byte {
	return []byte{
		byte('R'),
		0x00, 0x00, 0x00, 0x08,
		0x00, 0x00, 0x00, 0x00,
	}
}

// AuthenticationCleartextPassword :
// The authentication callenge by plain password.
//
// ref: https://www.postgresql.org/docs/18/protocol-message-formats.html#PROTOCOL-MESSAGE-FORMATS-AUTHENTICATIONCLEARTEXTPASSWORD
func AuthenticationCleartextPassword() []byte {
	return []byte{
		byte('R'),
		0x00, 0x00, 0x00, 0x08,
		0x00, 0x00, 0x00, 0x03,
	}
}

// AuthenticationMD5Password :
// The authentication callenge by MD5 password.
//
// salt should be 4 byte
//
// ref: https://www.postgresql.org/docs/18/protocol-message-formats.html#PROTOCOL-MESSAGE-FORMATS-AUTHENTICATIONMD5PASSWORD
func AuthenticationMD5Password(salt uint32) []byte {
	res := []byte{
		'R',
		0x00, 0x00, 0x00, 0x0C,
		0x00, 0x00, 0x00, 0x05,
	}

	saltByte := uint32toBytes(salt)

	return append(res, saltByte...)
}

// ReadyForQuery :
// Identifies the message type. ReadyForQuery is sent whenever the backend is ready for a new query cycle.
func ReadyForQuery() []byte {
	return []byte{
		byte('Z'),
		0x00, 0x00, 0x00, 0x05,
		byte('I'),
	}
}

// ParameterStatus returns a ParameterStatus (S) message.
//
// ref: https://www.postgresql.org/docs/18/protocol-message-formats.html#PROTOCOL-MESSAGE-FORMATS-PARAMETERSTATUS
func ParameterStatus(name, value string) []byte {
	buf := new(bytes.Buffer)

	buf.WriteByte('S')

	// lengthItself + name + null(\000) + value + null(\000)
	length := int32(4 + len(name) + 1 + len(value) + 1)

	binary.Write(buf, binary.BigEndian, length)

	buf.WriteString(name)
	buf.WriteByte(0)

	buf.WriteString(value)
	buf.WriteByte(0)

	return buf.Bytes()
}

// BackendKeyData returns a BackendKeyData (B) message.
//
// ref: https://www.postgresql.org/docs/18/protocol-message-formats.html#PROTOCOL-MESSAGE-FORMATS-BACKENDKEYDATA
func BackendKeyData(pid uint32, secretKey []byte) []byte {
	buf := new(bytes.Buffer)

	buf.WriteByte('K')

	// lengthItself + name + null(\000) + value + null + null(\000)
	length := int32(4 + 4 + len(secretKey))

	binary.Write(buf, binary.BigEndian, length)
	binary.Write(buf, binary.BigEndian, pid)
	buf.Write(secretKey)
	return buf.Bytes()
}
