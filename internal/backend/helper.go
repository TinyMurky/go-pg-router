package backend

import "encoding/binary"

func uint32toBytes(val uint32) []byte {
	r := make([]byte, 4)
	binary.BigEndian.PutUint32(r, val)
	return r
}

// b should be 4 bytes
func bytesToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}
