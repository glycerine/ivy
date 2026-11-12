package gia

import (
	"encoding/binary"
	"os"
)

func Gia_AigerReadInt(position []byte) int32 { return int32(binary.BigEndian.Uint32(position)) }
func Gia_AigerWriteInt(position []byte, value int32) {
	binary.BigEndian.PutUint32(position, uint32(value))
}
func Gia_AigerReadUnsigned(position *[]byte) uint32 {
	value, n := binary.Uvarint(*position)
	if n <= 0 || n > binary.MaxVarintLen32 || value > uint64(^uint32(0)) {
		panic("gia: malformed or overflowing AIGER unsigned integer")
	}
	*position = (*position)[n:]
	return uint32(value)
}
func Gia_AigerWriteUnsignedFile(file *os.File, value uint32) {
	var buffer [binary.MaxVarintLen32]byte
	n := binary.PutUvarint(buffer[:], uint64(value))
	if written, err := file.Write(buffer[:n]); err != nil || written != n {
		panic("gia: cannot write AIGER unsigned integer")
	}
}

func Gia_AigerWriteUnsignedBuffer(buffer []byte, position int32, value uint32) int32 {
	if position < 0 || int(position) > len(buffer) {
		panic("gia: invalid AIGER buffer position")
	}
	var encoded [binary.MaxVarintLen32]byte
	n := binary.PutUvarint(encoded[:], uint64(value))
	if len(buffer)-int(position) < n {
		panic("gia: short AIGER output buffer")
	}
	copy(buffer[int(position):], encoded[:n])
	return position + int32(n)
}
