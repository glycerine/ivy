package core_test

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec"
	"github.com/glycerine/abcgo/proof/cec/core"
)

// Source-named CEC node-two queries are conjunction checks, not pairwise
// equivalence checks. Compare both backends and an independent truth oracle.
func TestSourceWrapperAndCoreConjunction(t *testing.T) {
	g := gia.Gia_ManStart(32)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	ab := gia.Gia_ManAppendAnd(g, a, b)
	mix := gia.Gia_ManAppendAnd(g, a, b^1)
	literals := []int32{0, 1, a, a ^ 1, b, b ^ 1, ab, ab ^ 1, mix, mix ^ 1}
	for _, lit := range literals {
		gia.Gia_ManAppendCo(g, lit)
	}
	source := cec.Cec_ManSatCreate(g, nil)
	defer cec.Cec_ManSatStop(source)
	manager, err := core.NewManager(g, core.Options{MaxSATVariables: 2, RecycleQueries: 1})
	if err != nil {
		t.Fatal(err)
	}
	object := func(lit int32) *gia.Gia_Obj_t {
		p := gia.Gia_ManObj(g, lit>>1)
		if lit&1 != 0 {
			return gia.Gia_Not(p)
		}
		return p
	}
	for i, x := range literals {
		for j, y := range literals {
			if x == 1 && y == 1 {
				continue
			} // C asserts on this case.
			want := core.Equivalent
			for mask := 0; mask < 4; mask++ {
				out, e := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
				if e != nil {
					t.Fatal(e)
				}
				if out[i] && out[j] {
					want = core.Different
					break
				}
			}
			got, e := manager.CheckConjunction(x, y)
			if e != nil || got.Status != want {
				t.Fatalf("core %d,%d = %+v err=%v want=%d", x, y, got, e, want)
			}
			sourceStatus := cec.Cec_ManSatCheckNodeTwo(source, object(x), object(y))
			if source.Error != nil || sourceStatus != int32(want) {
				t.Fatalf("source wrapper %d,%d = %d err=%v want=%d", x, y, sourceStatus, source.Error, want)
			}
		}
	}
}
