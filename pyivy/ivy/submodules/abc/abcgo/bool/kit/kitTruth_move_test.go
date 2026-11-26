package kit

import (
	"math/bits"
	"math/rand"
	"testing"
)

// referenceMove is the bit-level definition of Kit_TruthStretch/Shrink:
// variable i of the small function is the i-th variable set in phase.
func referenceMove(in []uint32, total int32, phase uint32, shrink bool) []uint32 {
	var order []int
	for v := 0; v < int(total); v++ {
		if phase&(1<<uint(v)) != 0 {
			order = append(order, v)
		}
	}
	for v := 0; v < int(total); v++ {
		if phase&(1<<uint(v)) == 0 {
			order = append(order, v)
		}
	}
	out := make([]uint32, Kit_TruthWordNum(total))
	for x := 0; x < len(out)*32; x++ {
		old := x &^ int(Kit_BitMask(total))
		for n, o := range order {
			if shrink {
				old |= (x >> uint(n) & 1) << uint(o)
			} else {
				old |= (x >> uint(o) & 1) << uint(n)
			}
		}
		if in[old>>5]&(1<<uint(old&31)) != 0 {
			out[x>>5] |= 1 << uint(x&31)
		}
	}
	return out
}

func TestTruthStretchShrinkMatchReference(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for it := 0; it < 20000; it++ {
		total := int32(1 + r.Intn(12))
		phase := r.Uint32() & Kit_BitMask(total)
		nVars := int32(bits.OnesCount32(phase))
		words := Kit_TruthWordNum(total)
		in := make([]uint32, words)
		for i := range in {
			in[i] = r.Uint32()
		}
		for _, shrink := range []bool{false, true} {
			saved := append([]uint32(nil), in...)
			out := make([]uint32, words)
			if shrink {
				Kit_TruthShrink(out, in, nVars, total, phase, 0)
			} else {
				Kit_TruthStretch(out, in, nVars, total, phase, 0)
			}
			want := referenceMove(saved, total, phase, shrink)
			for i := range want {
				if out[i] != want[i] || in[i] != saved[i] {
					t.Fatalf("total %d phase %x shrink %v word %d: got %08x want %08x", total, phase, shrink, i, out[i], want[i])
				}
			}
		}
	}
}
