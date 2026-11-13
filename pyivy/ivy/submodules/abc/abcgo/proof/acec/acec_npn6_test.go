package acec

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSixInputNPNOrbit(t *testing.T) {
	truth := uint64(0x9f4a7c1582b396d0)
	members := vec.Vec_WrdAlloc(1024)
	canonical := Extra_TruthCanonNPN3(truth, 6, members)
	seen := make(map[uint64]bool, vec.Vec_WrdSize(members))
	minimum := ^uint64(0)
	for _, v := range vec.Vec_WrdArray(members) {
		if seen[v] {
			t.Fatalf("duplicate orbit member %x", v)
		}
		seen[v] = true
		if v < minimum {
			minimum = v
		}
	}
	if minimum != canonical || !seen[truth] || !seen[^truth] {
		t.Fatalf("bad orbit canonical=%x min=%x members=%d", canonical, minimum, len(seen))
	}
	for v := int32(0); v < 6; v++ {
		if !seen[util.Abc_Tt6Flip(truth, v)] {
			t.Fatalf("missing input phase %d", v)
		}
	}
	for v := int32(0); v < 5; v++ {
		if !seen[util.Abc_Tt6SwapAdjacent(truth, v)] {
			t.Fatalf("missing swap %d", v)
		}
	}
	if len(seen) != 92160 {
		t.Fatalf("unexpected generic orbit size %d", len(seen))
	}
}
