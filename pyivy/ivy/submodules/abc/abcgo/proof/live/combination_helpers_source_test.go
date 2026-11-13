package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestCombinationSourceHelpersBuildDescendingOr(t *testing.T) {
	old := aig.Aig_ManStart(20)
	a, b, c := aig.Aig_ObjCreateCi(old), aig.Aig_ObjCreateCi(old), aig.Aig_ObjCreateCi(old)
	for _, e := range []aig.Edge{a, b, c, a, b, c} {
		aig.Aig_ObjCreateCo(old, e)
	}
	aig.Aig_ManSetRegNum(old, 3)
	for _, exhaust := range []bool{false, true} {
		dst := aig.Aig_ManStart(20)
		ac, bc, cc := aig.Aig_ObjCreateCi(dst), aig.Aig_ObjCreateCi(dst), aig.Aig_ObjCreateCi(dst)
		aig.Aig_ObjSetCopy(old, a, ac)
		aig.Aig_ObjSetCopy(old, b, bc)
		aig.Aig_ObjSetCopy(old, c, cc)
		var got []aig.Edge
		var count int32
		if exhaust {
			count = generateCombinatorialStabilExhaust(dst, old, &got, 3, 3)
		} else {
			count = generateCombinatorialStabil(dst, old, []int32{0, 1, 2}, &got, 3, 3)
		}
		expected := aig.Aig_Or(dst, ac, aig.Aig_Or(dst, cc, bc))
		if count != 1 || len(got) != 1 || got[0] != expected {
			t.Fatalf("exhaust=%v got=%v count=%d expected %v", exhaust, got, count, expected)
		}
	}
}
