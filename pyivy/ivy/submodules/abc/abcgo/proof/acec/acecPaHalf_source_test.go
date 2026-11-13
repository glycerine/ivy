package acec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// acecPa.c:122-150 propagates all three sign bits even when the third
// input is constant zero, and records a complemented half-adder separately.
func TestSourceHalfAdderPhaseAndConstant(t *testing.T) {
	for sign := int32(0); sign < 16; sign++ {
		g := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		sum := gia.Gia_ManAppendAnd(g, a, b^1)
		carry := gia.Gia_ManAppendAnd(g, a, b)
		adds := vec.Vec_IntAllocArrayCopy([]int32{a >> 1, b >> 1, 0, sum >> 1, carry >> 1, sign}, 6)
		mapping := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(g)))
		vec.Vec_IntWriteEntry(mapping, int(carry>>1), 0)
		roots := vec.Vec_IntAllocArrayCopy([]int32{carry >> 1}, 1)
		var constants *Vec_Bit_t
		phases := Pas_ManPhase(g, adds, mapping, roots, &constants)
		if vec.Vec_BitEntry(constants, 0) != 1 {
			t.Fatalf("sign=%d missing complemented half-adder", sign)
		}
		expectedSum := (sign >> 3) & 1
		for i, id := range []int32{a >> 1, b >> 1, 0} {
			phase := ((sign >> i) & 1) ^ 1
			expectedSum ^= phase
			if id != 0 && int32(vec.Vec_BitEntry(phases, int(id))) != phase {
				t.Fatalf("sign=%d input=%d", sign, i)
			}
		}
		if got := int32(vec.Vec_BitEntry(phases, int(sum>>1))); got != expectedSum {
			t.Fatalf("sign=%d sum=%d want=%d", sign, got, expectedSum)
		}
	}
}
