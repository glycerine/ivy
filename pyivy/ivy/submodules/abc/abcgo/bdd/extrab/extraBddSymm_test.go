package extrab

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"testing"
)

func TestSymmetryAllThreeVariableFunctions(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	for table := 0; table < 256; table++ {
		f := autoTruth(dd, uint64(table), 3)
		fast := Extra_SymmPairsCompute(dd, f)
		slow := Extra_SymmPairsComputeNaive(dd, f)
		if fast == nil || slow == nil || fast.NSymms != slow.NSymms || fast.NVars != slow.NVars {
			t.Fatalf("table %x fast=%+v slow=%+v", table, fast, slow)
		}
		for i := 0; i < fast.NVars; i++ {
			for j := i + 1; j < fast.NVars; j++ {
				if fast.Symms[i][j] != slow.Symms[i][j] {
					t.Fatal("pair", table, i, j)
				}
			}
		}
		for a := 0; a < 3; a++ {
			for b := a + 1; b < 3; b++ {
				if Extra_bddCheckVarsSymmetricNaive(dd, f, a, b) && !Extra_bddCheckVarsSymmetric(dd, f, a, b) {
					t.Fatal("prefilter false negative", table, a, b)
				}
			}
		}
	}
}
func TestSymmetryTuplesAndCofactorRelations(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	vars := cudd.Cudd_bddAnd(dd, dd.Vars[0], cudd.Cudd_bddAnd(dd, dd.Vars[1], dd.Vars[2]))
	for k, want := range []byte{1, 22, 104, 128} {
		if got := autoEquationBits(dd, Extra_zddTuplesFromBdd(dd, k, vars)); got != want {
			t.Fatal("tuple", k, got, want)
		}
	}
	for a := 0; a < 256; a++ {
		f := autoTruth(dd, uint64(a), 3)
		for b := 0; b < 256; b++ {
			g := autoTruth(dd, uint64(b), 3)
			var want byte
			if f == g {
				want |= 1
			}
			for i := 0; i < 3; i++ {
				if cudd.Cudd_Cofactor(dd, f, cudd.Cudd_Not(dd.Vars[i])) == cudd.Cudd_Cofactor(dd, g, dd.Vars[i]) {
					want |= 1 << uint(1<<i)
				}
			}
			if got := autoEquationBits(dd, Extra_zddGetSymmetricVars(dd, f, g, vars)); got != want {
				t.Fatalf("get vars a%x b%x got%x want%x", a, b, got, want)
			}
		}
	}
}

func TestOriginalCSymmetryFixture(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	vars := cudd.Cudd_bddAnd(dd, dd.Vars[0], cudd.Cudd_bddAnd(dd, dd.Vars[1], dd.Vars[2]))
	functions := make([]*DdNode, 256)
	for i := range functions {
		functions[i] = autoTruth(dd, uint64(i), 3)
	}
	hash := uint64(14695981039346656037)
	put := func(b byte) { hash ^= uint64(b); hash *= 1099511628211 }
	for _, f := range functions {
		p := Extra_SymmPairsCompute(dd, f)
		put(byte(p.NVars))
		put(byte(p.NSymms))
		for i := 0; i < p.NVars; i++ {
			put(byte(p.Vars[i]))
			for j := i + 1; j < p.NVars; j++ {
				if p.Symms[i][j] {
					put(1)
				} else {
					put(0)
				}
			}
		}
		for i := 0; i < 3; i++ {
			for j := i + 1; j < 3; j++ {
				if Extra_bddCheckVarsSymmetric(dd, f, i, j) {
					put(1)
				} else {
					put(0)
				}
			}
		}
		for _, g := range functions {
			put(autoEquationBits(dd, Extra_zddGetSymmetricVars(dd, f, g, vars)))
		}
	}
	if hash != 0x72c5fd61011d557d {
		t.Fatalf("native symmetry hash %016x", hash)
	}
}
func TestSymmetryReorderedAndInvalidDomains(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	f := cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[3])
	cudd.Cudd_Ref(f)
	if cudd.Cudd_ShuffleHeap(dd, []int{3, 1, 0, 2}) != 1 {
		t.Fatal("reorder")
	}
	fast, slow := Extra_SymmPairsCompute(dd, f), Extra_SymmPairsComputeNaive(dd, f)
	if fast == nil || fast.NSymms != slow.NSymms {
		t.Fatal("reordered symmetry")
	}
	if Extra_zddSymmPairsCompute(dd, f, dd.Vars[0]) != nil {
		t.Fatal("missing support")
	}
	if Extra_zddTuplesFromBdd(dd, -1, dd.One) != nil {
		t.Fatal("negative tuple")
	}
	if Extra_zddGetSingletons(dd, cudd.Cudd_Not(dd.Vars[0])) != nil {
		t.Fatal("negative cube")
	}
	if Extra_SymmPairsCreateFromZdd(dd, cudd.Cudd_zddChange(dd, dd.One, 1), dd.Vars[0]) != nil {
		t.Fatal("odd literal")
	}
	if Extra_SymmPairsCreateFromZdd(dd, cudd.Cudd_zddChange(dd, dd.One, 0), dd.Vars[0]) != nil {
		t.Fatal("singleton pair")
	}
	if Extra_SymmPairsAllocate(-1) != nil {
		t.Fatal("negative size")
	}
	if p := Extra_SymmPairsCompute(dd, dd.One); p == nil || p.NVars != 0 || p.NSymms != 0 {
		t.Fatal("constant info")
	}
}
