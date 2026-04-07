package backend

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

// maxBackendMsgSize will be 1 GiB
const maxBackendMsgSize = 1 << 30

// readBackendMessage reads one backend-format message.
// Returns the type byte and the payload (excluding type and length).
//
// return:
//  1. msgType: the message type that was defined in PostgreSQL message flow
//     it will be first byte to be read.
//     ex : 'R' for Authentication
//  2. payload: payload will be the message after msgType and length of payload.
//     That means payload does not contain length of payload at begining
//  3. err: error
func readBackendMessage(r io.Reader) (msgType byte, payload []byte, err error) {

	msgTypeBuf := make([]byte, 1)

	if _, err := io.ReadFull(r, msgTypeBuf); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return 0, nil, fmt.Errorf("readBackendMessage: read msgTypeBuf bytes: %w: %w", ErrConnectionClosed, err)
		}
		return 0, nil, fmt.Errorf("readBackendMessage: read msgTypeBuf bytes: %w: %w", ErrInvalidMsgFormat, err)
	}

	var totalLength uint32
	if err := binary.Read(r, binary.BigEndian, &totalLength); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return 0, nil, fmt.Errorf("readBackendMessage: read total length bytes: %w: %w", ErrConnectionClosed, err)
		}
		return 0, nil, fmt.Errorf("readBackendMessage: read total length bytes: %w: %w", ErrInvalidMsgFormat, err)
	}

	// 4 for length of totalLength
	if totalLength < 4 {
		return 0, nil, fmt.Errorf("readBackendMessage: %w: total length %d is too short (minimum 4)", ErrInvalidMsgFormat, totalLength)
	}
	if totalLength > maxBackendMsgSize {
		// When totalLength > maxBackendMsgSize , most likely because grabage input.
		return 0, nil, fmt.Errorf("readBackendMessage: %w: total length %d is too long (maximum %d)", ErrInvalidMsgFormat, totalLength, maxBackendMsgSize)
	}

	// rewrite totalLength back into payload
	payloadBuf := new(bytes.Buffer)

	lenOfRestIfPayload := totalLength - 4
	if _, err := io.CopyN(payloadBuf, r, int64(lenOfRestIfPayload)); err != nil {
		if errors.Is(err, io.EOF) {

			return 0, nil, fmt.Errorf("readBackendMessage: %w: %w", ErrConnectionClosed, err)
		}
		return 0, nil, fmt.Errorf("readBackendMessage: %w", err)
	}
	return msgTypeBuf[0], payloadBuf.Bytes(), nil
}

// doAuthHandshake sends the startup message to backend and handles
// the auth challenge until the backend sends ReadyForQuery.
func doAuthHandshake(rw io.ReadWriter, params map[string]string, password string) error {
	return nil
}

// parseErrorResponse extracts the 'M' (message) field from an ErrorResponse payload
func parseErrorResponse(payload []byte) string {
	return ""
}

func buildStartupMessage(kv map[string]string) []byte {
	// filter password
	kvWithoutSensitive := make(map[string]string)
	for k, v := range kv {
		if strings.ToLower(k) == "password" {
			continue
		}
		kvWithoutSensitive[k] = v
	}
	binaryKVs := buildBinaryKV(kvWithoutSensitive)

	// protocol high 16 bites are Major Version
	// low 16 bites are Minor Version
	// so this is version 3.0
	var protocol uint32 = 0x00030000
	// total message lenght should be 4 bytes too
	totalLength := 4 + 4 + uint32(len(binaryKVs))

	msg := buildCustomStartupMsg(totalLength, protocol, binaryKVs)
	return msg
}

func buildBinaryKV(kv map[string]string) []byte {
	// it will be same as []byte("\000")[0]
	nullChar := byte(0)
	buf := new(bytes.Buffer)
	for k, v := range kv {
		_, _ = buf.WriteString(k)
		_ = buf.WriteByte(nullChar)
		_, _ = buf.WriteString(v)
		_ = buf.WriteByte(nullChar)
	}
	// the last one means the end of sentence

	_ = buf.WriteByte(nullChar)

	return buf.Bytes()
}

func buildCustomStartupMsg(totalLength uint32, protocol uint32, binaryKV []byte) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, totalLength)
	_ = binary.Write(buf, binary.BigEndian, protocol)
	_, _ = buf.Write(binaryKV)
	return buf.Bytes()
}
