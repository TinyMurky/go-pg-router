package pgserver

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func BuildSSLRequestMessage(t testing.TB) []byte {
	t.Helper()

	// total message lenght should be 4 bytes too
	var testTotalLength uint32 = 4 + 4

	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, testTotalLength)
	_ = binary.Write(buf, binary.BigEndian, SSLRequest())

	return buf.Bytes()
}

func BuildStarupMessage(t testing.TB, kv map[string]string) ([]byte, uint32) {
	t.Helper()

	binaryTestKVs := BuildBinaryKV(t, kv)

	// protocol high 16 bites are Major Version
	// low 16 bites are Minor Version
	// so this is version 3.0
	var testProtocol uint32 = 0x00030000
	// total message lenght should be 4 bytes too
	testTotalLength := 4 + 4 + uint32(len(binaryTestKVs))

	msg := BuildCustomStartupMsg(t, testTotalLength, testProtocol, binaryTestKVs)
	return msg, testProtocol
}

func BuildBinaryKV(t testing.TB, kv map[string]string) []byte {
	t.Helper()

	nullChar := []byte("\000")
	buf := new(bytes.Buffer)
	for k, v := range kv {
		binary.Write(buf, binary.BigEndian, []byte(k))
		binary.Write(buf, binary.BigEndian, nullChar)
		binary.Write(buf, binary.BigEndian, []byte(v))
		binary.Write(buf, binary.BigEndian, nullChar)
	}

	// the last one means the end of sentence
	binary.Write(buf, binary.BigEndian, nullChar)

	return buf.Bytes()
}

func BuildCustomStartupMsg(t testing.TB, totalLength uint32, protocol uint32, binaryKV []byte) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, totalLength)
	binary.Write(buf, binary.BigEndian, protocol)
	binary.Write(buf, binary.BigEndian, binaryKV)
	return buf.Bytes()
}
