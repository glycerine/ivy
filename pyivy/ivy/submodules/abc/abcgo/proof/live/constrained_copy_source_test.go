package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestConstrainedConeReadsOldGraphCopies(t *testing.T) {
	old := aig.Aig_ManStart(10)
	a, b := aig.Aig_ObjCreateCi(old), aig.Aig_ObjCreateCi(old)
	dst := aig.Aig_ManStart(10)
	aig.Aig_ObjCreateCi(dst)
	bImage := aig.Aig_ObjCreateCi(dst)
	aImage := aig.Aig_ObjCreateCi(dst)
	aig.Aig_ObjSetCopy(old, a, aImage)
	aig.Aig_ObjSetCopy(old, b, bImage)
	for _, build := range []func(*aig.Graph, *aig.Graph, []aig.Edge) aig.Edge{createConstrained0LiveCone, createConstrained0LiveConeWithDSC} {
		result := build(old, dst, []aig.Edge{a.Not(), b})
		aig.Aig_ObjCreateCo(dst, result)
	}
	for bits := 0; bits < 8; bits++ {
		vals := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, err := dst.Evaluate(vals)
		if err != nil {
			t.Fatal(err)
		}
		want := vals[2] || vals[1]
		for _, v := range got {
			if v != want {
				t.Fatalf("inputs %v: got %v, want %v", vals, got, want)
			}
		}
	}
}
