package pgserver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

// readSSLOrStartup reads the first message from the client.
// If it is an SSL request, replies 'N' and reads the next message.
// Returns the startup message bytes.
//
// ref: https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-SSL
// SSLRequest: https://www.postgresql.org/docs/current/protocol-message-formats.html#PROTOCOL-MESSAGE-FORMATS-SSLREQUEST
func readSSLOrStartup(rw io.ReadWriter) ([]byte, error) {
	for {
		var totalLength uint32
		if err := binary.Read(rw, binary.BigEndian, &totalLength); err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("readSSLOrStartup: read total length bytes: %w: %w", ErrConnectionClosed, err)
			}
			return nil, fmt.Errorf("readSSLOrStartup: read total length bytes: %w: %w", ErrInvalidMsgFormat, err)
		}

		// 4 for length of totalLength
		// 4 for length of SSLRequest or ProtocolVersion
		if totalLength < 8 {
			return nil, fmt.Errorf("readSSLOrStartup: %w: total length %d is too short (minimum 8)", ErrInvalidMsgFormat, totalLength)
		}
		if totalLength > maxStartupMsgSize {
			// When totalLength > maxStartupMsgSize, most likely because grabage input.
			// SSLRequest will be only 8 bytes long
			return nil, fmt.Errorf("readSSLOrStartup: %w: total length %d is too long (maximum %d)", ErrInvalidMsgFormat, totalLength, maxStartupMsgSize)
		}

		code := make([]byte, 4)

		if _, err := io.ReadFull(rw, code); err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("readSSLOrStartup: read code bytes: %w: %w", ErrConnectionClosed, err)
			}
			return nil, fmt.Errorf("readSSLOrStartup: code bytes: %w: %w", ErrInvalidMsgFormat, err)
		}

		if isSSLRequest(code) {
			if _, err := rw.Write(SSLRequestReject()); err != nil {
				// io.ErrClosedPipe to detect connection close if w is net.Pipe
				if errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrClosedPipe) {
					return nil, fmt.Errorf("readSSLOrStartup: %w: %w", ErrConnectionClosed, err)
				}
				return nil, fmt.Errorf("readSSLOrStartup: %w", err)
			}

			continue
		}

		buf := new(bytes.Buffer)
		_ = binary.Write(buf, binary.BigEndian, totalLength)
		_ = binary.Write(buf, binary.BigEndian, code)
		lenOfKV := totalLength - 4 - 4
		if _, err := io.CopyN(buf, rw, int64(lenOfKV)); err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {

				return nil, fmt.Errorf("readSSLOrStartup: %w: %w", ErrConnectionClosed, err)
			}
			return nil, fmt.Errorf("readSSLOrStartup: %w", err)
		}

		return buf.Bytes(), nil
	}
}

// isSSLRequest check certain []byte is same as SSLRequest code
func isSSLRequest(b []byte) bool {
	return bytes.Equal(SSLRequest(), b)
}
