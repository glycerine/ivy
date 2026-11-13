package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// A compiled C probe of Gia_ManForEachAnd plus Gia_ManHashAnd shows that
// native XOR nodes are copied as AND gates by the checked-in C helpers.
func TestSourceCopiesNativeXorAsAnd(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	xor := gia.Gia_ManAppendXorReal(g, a, b)
	gia.Gia_ManAppendCo(g, xor)
	gia.Gia_ManAppendCo(g, xor)
	gia.Gia_ManSetRegNum(g, 1)
	out := Int2_ManFrameInit(g, 1, 0)
	defer gia.Gia_ManStop(out)
	m := Int2_GoCreate(g, &Int2_ManPars_t{})
	defer Int2_ManStop(m)
	m.pSatPref = sat_solver_new()
	Int2_GoCreateFrames(m, 0, vec.Vec_IntAllocArrayCopy([]int32{0}, 1))
	output := Int2_ManCheckFrames(m, 0, gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0)))
	for mask := int32(0); mask < 4; mask++ {
		x, y := mask&1 != 0, mask&2 != 0
		values, err := out.Evaluate([]bool{x, y})
		if err != nil {
			t.Fatal(err)
		}
		if values[0] != (x && y) {
			t.Fatalf("frame inputs %d: got %v", mask, values)
		}
		av := Int2_ManCheckFrames(m, 0, a>>1)
		bv := Int2_ManCheckFrames(m, 0, b>>1)
		for v := int32(0); v < 2; v++ {
			status := sat_solver_solve(m.pSatPref, []int32{av ^ (1 - (mask & 1)), bv ^ (1 - ((mask >> 1) & 1)), output ^ (1 - v)}, 0, 0, 0, 0)
			if (status == l_True) != ((v != 0) == (x && y)) {
				t.Fatalf("mask=%d output=%d status=%d", mask, v, status)
			}
		}
	}
}
