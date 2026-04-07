package backend

import (
	"bytes"
	"encoding/binary"
)

func buildCustomReadMessage(msgType byte, customLength uint32, payload []byte) []byte {
	buf := new(bytes.Buffer)

	buf.WriteByte(msgType)
	binary.Write(buf, binary.BigEndian, customLength)
	buf.Write(payload)

	return buf.Bytes()
}
