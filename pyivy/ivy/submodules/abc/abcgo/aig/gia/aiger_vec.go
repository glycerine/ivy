package gia

import (
	"encoding/binary"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_AigerWriteUnsigned(out *Vec_Str_t, value uint32) {
	var buffer [binary.MaxVarintLen32]byte
	n := binary.PutUvarint(buffer[:], uint64(value))
	for _, b := range buffer[:n] {
		vec.Vec_StrPush(out, b)
	}
}
