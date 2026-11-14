package res

import (
	"testing"
)

type fixedRandSource struct{ value uint32 }

func (s fixedRandSource) next() uint32 { return s.value }

func TestSourceRandomBytesOverwritesAndOrdersBits(t *testing.T) {
	// The source processes bit zero first into the most significant byte.
	p := []uint32{0x12345678, 0x87654321, 99}
	Abc_InfoRandomBytes(p, 2, fixedRandSource{value: 5})
	if p[0] != 0xff00ff00 || p[1] != 0xff00ff00 || p[2] != 99 {
		t.Fatalf("got %08x", p)
	}
}
