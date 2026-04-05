package pgserver

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

// Startup Massage maximum will be 10_000 bytes
// Ref: PostgreSQL src/include/libpq/pqcomm.h
const maxStartupMsgSize = 10_000

// StartupMessage will parse Start Up Messages from client,
// then interact with client.
//
// PostgreSQL doc: https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-START-UP
type StartupMessage struct {
	ProtocolVersion uint32
	Parameters      map[string]string
}

// ReadStartupMessage should parse the start up message.
//
// Structure of Start up message looks like:
//
//	| 4 bytes: total message length (including these 4 bytes) |
//	| 4 bytes: protocol version (3.0 = 0x00030000)            |
//	| key\0value\0key\0value\0 ... \0                          |
//
// The key-value pairs are null-terminated strings. For example:
//
//	user\0tinymurky\0database\0mydb\0\0
//
// (The final \0 signals the end of the list.)
func (sm *StartupMessage) ReadStartupMessage(r io.Reader) error {
	totalLength, protocolVersionBuf, err := readMsgHeader(r)

	if err != nil {
		return fmt.Errorf("ReadStartupMessage: %w", err)
	}
	// Should we check version?

	sm.ProtocolVersion = binary.BigEndian.Uint32(protocolVersionBuf)

	lenOfKV := totalLength - 4 - 4

	kvBuf := make([]byte, lenOfKV)

	if _, err := io.ReadFull(r, kvBuf); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return fmt.Errorf("ReadStartupMessage read key value pairs bytes: %w: %w", ErrConnectionClosed, err)
		}
		return fmt.Errorf("ReadStartupMessage read key value pairs bytes: %w: %w", ErrInvalidMsgFormat, err)
	}

	// \000 in golang is \0 in c
	splitedKV := strings.Split(string(kvBuf), "\000")

	if len(splitedKV)%2 != 0 {
		return fmt.Errorf("ReadStartupMessage: %w : read key value pairs bytes: key value is not paired one by one", ErrInvalidMsgFormat)
	}

	sm.Parameters = make(map[string]string)
	for i := 0; i+1 < len(splitedKV); i += 2 {
		k := splitedKV[i]

		// KV looks like: user\0tinymurky\0database\0mydb\0\0
		// it will end with extra \0
		// the last part of kv split will be k = "" and v = ""
		if k == "" {
			break
		}
		v := splitedKV[i+1]
		sm.Parameters[k] = v
	}

	return nil
}

// WriteAuthOK will return client with AuthenticationOk,
// AuthenticationOk originally returned then client pass authentication by provided password,
// but go-pg-router will return WriteAuthOK when connect to client no matter what (fake it),
// So that we can guarantee create connection with client (and since go-pg-router is just work as proxy, we don't need authentication)
func (sm *StartupMessage) WriteAuthOK(w io.Writer) error {
	if _, err := w.Write(StartUPAuthenticationOk()); err != nil {
		// io.ErrClosedPipe to detect connection close if w is net.Pipe
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrClosedPipe) {
			return fmt.Errorf("WriteAuthOK: %w: %w", ErrConnectionClosed, err)
		}
		return fmt.Errorf("WriteAuthOK: %w", err)
	}

	return nil
}

// WriteReadyForQuery will return ReadyForQuery Message to client,
// message should be send after WriteAuthOK,
// this message is telling client that they can send the rest of the SQL queries
func (sm *StartupMessage) WriteReadyForQuery(w io.Writer) error {
	if _, err := w.Write(StartUPReadyForQuery()); err != nil {
		// io.ErrClosedPipe to detect connection close if w is net.Pipe
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrClosedPipe) {
			return fmt.Errorf("WriteReadyForQuery: %w: %w", ErrConnectionClosed, err)
		}
		return fmt.Errorf("WriteReadyForQuery: %w", err)
	}

	return nil
}

// readMsgHeader will read first four bytes to get totalLength of TCP Message from client,
// then read next 4 bytes to get whether SLL request code or protocol version
// return will be (totalLength uint32, code []byte, err error)
func readMsgHeader(r io.Reader) (uint32, []byte, error) {
	var totalLength uint32
	if err := binary.Read(r, binary.BigEndian, &totalLength); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return 0, nil, fmt.Errorf("readMsgHeader: read total length bytes: %w: %w", ErrConnectionClosed, err)
		}
		return 0, nil, fmt.Errorf("readMsgHeader: read total length bytes: %w: %w", ErrInvalidMsgFormat, err)
	}

	// Since the incoming net.Conn will not be closed
	// we used totalLength to determine how many bytes we will read
	// 4 for length of totalLength
	// 4 for length of SSLRequest or ProtocolVersion
	if totalLength < 8 {
		return 0, nil, fmt.Errorf("readMsgHeader: %w: total length %d is too short (minimum 8)", ErrInvalidMsgFormat, totalLength)
	}
	if totalLength > maxStartupMsgSize {
		// When totalLength > maxStartupMsgSize, most likely because grabage input.
		// SSLRequest will be only 8 bytes long
		return 0, nil, fmt.Errorf("readMsgHeader: %w: total length %d is too long (maximum %d)", ErrInvalidMsgFormat, totalLength, maxStartupMsgSize)
	}

	code := make([]byte, 4)

	if _, err := io.ReadFull(r, code); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return 0, nil, fmt.Errorf("readMsgHeader: read code bytes: %w: %w", ErrConnectionClosed, err)
		}
		return 0, nil, fmt.Errorf("readMsgHeader: code bytes: %w: %w", ErrInvalidMsgFormat, err)
	}

	return totalLength, code, nil
}
