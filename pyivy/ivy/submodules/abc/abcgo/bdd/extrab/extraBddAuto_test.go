package extrab

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func autoTruth(dd *DdManager, table uint64, n int) *DdNode {
	var build func(int, uint64) *DdNode
	build = func(level int, table uint64) *DdNode {
		if level == n {
			if table&1 != 0 {
				return dd.One
			}
			return dd.LogicZero
		}
		width := uint(1) << uint(n-level-1)
		mask := (uint64(1) << width) - 1
		return cudd.Cudd_bddIte(dd, dd.Vars[level], build(level+1, table>>width), build(level+1, table&mask))
	}
	return build(0, table)
}
func autoTruthBits(t *testing.T, dd *DdManager, f *DdNode, n int) uint64 {
	t.Helper()
	if f == nil {
		t.Fatal("nil space")
	}
	var table uint64
	for a := 0; a < 1<<uint(n); a++ {
		values := make([]bool, len(dd.Vars))
		for i := 0; i < n; i++ {
			values[i] = a&(1<<uint(n-1-i)) != 0
		}
		value, err := dd.Evaluate(f, values)
		if err != nil {
			t.Fatal(err)
		}
		if value {
			table |= 1 << uint(a)
		}
	}
	return table
}
func autoEquationBits(dd *DdManager, f *DdNode) byte {
	var result byte
	var visit func(*DdNode, int)
	visit = func(f *DdNode, mask int) {
		if f == dd.Zero {
			return
		}
		if f == dd.One {
			result |= 1 << uint(mask)
			return
		}
		visit(f.E, mask)
		visit(f.T, mask|1<<uint(f.Index/2))
	}
	visit(f, 0)
	return result
}
func TestAutoAllTwoFunctionTranslationSpaces(t *testing.T) {
	dd := cudd.NewBDDManager(6)
	functions := make([]*DdNode, 256)
	for i := range functions {
		functions[i] = autoTruth(dd, uint64(i), 3)
	}
	for a, f := range functions {
		fast := Extra_bddSpaceFromFunctionFast(dd, f)
		if fast != Extra_bddSpaceFromFunction(dd, f, f) {
			t.Fatal("fast space", a)
		}
		for b, g := range functions {
			space := autoTruthBits(t, dd, Extra_bddSpaceFromFunction(dd, f, g), 3)
			var want uint64
			for shift := 0; shift < 8; shift++ {
				equal := true
				for x := 0; x < 8; x++ {
					if a>>uint(x)&1 != b>>uint(x^shift)&1 {
						equal = false
						break
					}
				}
				if equal {
					want |= 1 << uint(shift)
				}
			}
			if space != want {
				t.Fatalf("shift a%x b%x got%x want%x", a, b, space, want)
			}
		}
	}
}
func TestAutoLinearEquationsReductionAndGates(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	f := cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
	periods := Extra_bddSpaceFromFunction(dd, f, f)
	equations := Extra_bddSpaceEquationsNeg(dd, periods)
	if Extra_bddSpaceFromMatrixPos(dd, equations) != periods {
		t.Fatal("equation reconstruction")
	}
	canonical := Extra_bddSpaceCanonVars(dd, periods)
	reduced := Extra_bddSpaceReduce(dd, f, canonical)
	if reduced != dd.Vars[1] {
		t.Fatal("canonical reduction")
	}
	gates := Extra_bddSpaceExorGates(dd, reduced, equations)
	if len(gates) != 4 || gates[1] == nil || Extra_zddLitCountComb(dd, gates[1]) != 2 || gates[1].Ref == 0 {
		t.Fatal("EXOR gates", gates)
	}
	for _, g := range gates {
		if g != nil {
			cudd.Cudd_RecursiveDeref(dd, g)
		}
	}
	if Extra_bddSpaceExorGates(dd, dd.One, equations) != nil {
		t.Fatal("invalid gate basis accepted")
	}
	if Extra_bddSpaceFromFunctionFast(cudd.NewBDDManager(2), nil) != nil {
		t.Fatal("invalid fast input")
	}
}
func TestAutoReorderingAndInvalidDomains(t *testing.T) {
	dd := cudd.NewBDDManager(6)
	f := cudd.Cudd_bddXor(dd, dd.Vars[1], dd.Vars[4])
	cudd.Cudd_Ref(f)
	cudd.Cudd_ShuffleHeap(dd, []int{4, 2, 0, 5, 1, 3})
	if Extra_bddSpaceFromFunctionFast(dd, f) != Extra_bddSpaceFromFunction(dd, f, f) {
		t.Fatal("reordered fastspace")
	}
	periods := Extra_bddSpaceFromFunction(dd, f, f)
	if Extra_bddSpaceFromMatrixPos(dd, Extra_bddSpaceEquationsNeg(dd, periods)) != periods {
		t.Fatal("reordered equation roundtrip")
	}
	other := cudd.NewBDDManager(6)
	if Extra_bddSpaceFromFunction(dd, f, other.Vars[0]) != nil || Extra_bddSpaceEquationsPos(dd, cudd.Cudd_addIthVar(dd, 0)) != nil {
		t.Fatal("invalid BDD accepted")
	}
	odd := cudd.Cudd_zddChange(dd, dd.One, 1)
	if Extra_bddSpaceFromMatrixPos(dd, odd) != nil {
		t.Fatal("odd-index matrix accepted")
	}
	small := cudd.NewBDDManager(2)
	if Extra_bddSpaceFromFunctionFast(small, cudd.Cudd_bddAnd(small, small.Vars[0], small.Vars[1])) != nil {
		t.Fatal("insufficient scratch variables accepted")
	}
}

func TestOriginalCAutoFixture(t *testing.T) {
	dd := cudd.NewBDDManager(6)
	functions := make([]*DdNode, 256)
	for i := range functions {
		functions[i] = autoTruth(dd, uint64(i), 3)
	}
	var hash uint64 = 14695981039346656037
	byteHash := func(v byte) { hash ^= uint64(v); hash *= 1099511628211 }
	bcollect := func(f *DdNode) { byteHash(byte(autoTruthBits(t, dd, f, 3))) }
	zcollect := func(f *DdNode) { byteHash(autoEquationBits(dd, f)) }
	for _, f := range functions {
		bcollect(Extra_bddSpaceFromFunctionPos(dd, f))
		bcollect(Extra_bddSpaceFromFunctionNeg(dd, f))
		bcollect(Extra_bddSpaceCanonVars(dd, f))
		bcollect(Extra_bddSpaceFromFunction(dd, f, f))
		bcollect(Extra_bddSpaceFromFunctionFast(dd, f))
		p, n := Extra_bddSpaceEquationsPos(dd, f), Extra_bddSpaceEquationsNeg(dd, f)
		zcollect(p)
		zcollect(n)
		zcollect(Extra_bddSpaceEquations(dd, f))
		bcollect(Extra_bddSpaceFromMatrixPos(dd, n))
		bcollect(Extra_bddSpaceFromMatrixNeg(dd, p))
		for _, g := range functions {
			bcollect(Extra_bddSpaceFromFunction(dd, f, g))
		}
	}
	for family := 0; family < 256; family++ {
		m := dd.Zero
		for member := 0; member < 8; member++ {
			if family&(1<<uint(member)) == 0 {
				continue
			}
			cube := dd.One
			for v := 0; v < 3; v++ {
				if member&(1<<uint(v)) != 0 {
					cube = cudd.Cudd_zddChange(dd, cube, 2*v)
				}
			}
			m = cudd.Cudd_zddUnion(dd, m, cube)
		}
		bcollect(Extra_bddSpaceFromMatrixPos(dd, m))
		bcollect(Extra_bddSpaceFromMatrixNeg(dd, m))
	}
	if hash != 0x4b8b81243ecf473a {
		t.Fatalf("original C auto fixture: %016x", hash)
	}
}
