package gia

import (
	"encoding/binary"
	"testing"
)

func TestOriginalAigerIntegers(t *testing.T) {
	for _, value := range []int32{0, 1, -1, 0x12345678, -2147483648} {
		var buffer [4]byte
		Gia_AigerWriteInt(buffer[:], value)
		if Gia_AigerReadInt(buffer[:]) != value {
			t.Fatal(value)
		}
	}
	for _, value := range []uint32{0, 1, 127, 128, 16384, 0xffffffff} {
		data := binary.AppendUvarint(nil, uint64(value))
		data = append(data, 42)
		if got := Gia_AigerReadUnsigned(&data); got != value || len(data) != 1 || data[0] != 42 {
			t.Fatal(value, got, data)
		}
	}
}

func TestAigerUnsignedRejectsMalformed32BitValues(t *testing.T) {
	for _, input := range [][]byte{{}, {0x80}, {0xff, 0xff, 0xff, 0xff, 0x1f}, {0x80, 0x80, 0x80, 0x80, 0x80, 0}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("malformed value accepted", input)
				}
			}()
			data := input
			Gia_AigerReadUnsigned(&data)
		}()
	}
}
