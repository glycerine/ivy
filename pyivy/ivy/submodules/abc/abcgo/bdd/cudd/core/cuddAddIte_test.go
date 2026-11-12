package core

import (
	"math"
	"testing"
)

func TestADDITEPermutationAndInverse(t *testing.T) {
	dd := NewBDDManager(3)
	x, y, z := Cudd_addIthVar(dd, 0), Cudd_addIthVar(dd, 1), Cudd_addIthVar(dd, 2)
	a, b, c := Cudd_addConst(dd, 2.5), Cudd_addConst(dd, -7), Cudd_addConst(dd, 11)
	f := Cudd_addIte(dd, x, Cudd_addIte(dd, y, a, b), Cudd_addIte(dd, z, c, dd.Zero))
	if f == nil || !Cudd_IsAdd(f) || Cudd_IsAdd(dd.Vars[0]) {
		t.Fatal("ADD family detection")
	}
	perm := []int{2, 0, 1}
	p := Cudd_addPermute(dd, f, perm)
	if p == nil || Cudd_addPermute(dd, p, []int{1, 2, 0}) != f {
		t.Fatal("inverse permutation not canonical")
	}
	for bits := 0; bits < 8; bits++ {
		original := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		mapped := make([]bool, 3)
		for i, v := range original {
			mapped[perm[i]] = v
		}
		want := 0.0
		if original[0] {
			if original[1] {
				want = 2.5
			} else {
				want = -7
			}
		} else if original[2] {
			want = 11
		}
		got, err := dd.EvaluateADD(f, original)
		if err != nil || got != want {
			t.Fatal(bits, got, want, err)
		}
		got, err = dd.EvaluateADD(p, mapped)
		if err != nil || got != want {
			t.Fatal("permuted", bits, got, want, err)
		}
	}
	if Cudd_addIte(dd, x, a, a) != a || Cudd_addConst(dd, 2.5) != a {
		t.Fatal("ADD reduction/constant canonicity")
	}
	for _, n := range dd.addUnique {
		if n.family != 2 || n.T == n.E || Cudd_IsComplement(n.T) || Cudd_IsComplement(n.E) || dd.level(n.T) <= dd.level(n) || dd.level(n.E) <= dd.level(n) {
			t.Fatal("ADD ordering invariant")
		}
	}
}
func TestADDBoundariesAndManagerIsolation(t *testing.T) {
	dd := NewBDDManager(2)
	x := Cudd_addIthVar(dd, 0)
	a := Cudd_addConst(dd, 3)
	f := Cudd_addIte(dd, x, a, dd.Zero)
	before := len(dd.Vars)
	if Cudd_addPermute(dd, f, []int{-1}) != nil || len(dd.Vars) != before {
		t.Fatal("invalid permutation mutated manager")
	}
	if Cudd_addIte(dd, dd.Vars[0], a, dd.Zero) != nil || Cudd_bddAnd(dd, f, dd.One) != nil {
		t.Fatal("mixed family accepted")
	}
	if Cudd_addIte(dd, x, NewBDDManager(0).Zero, a) != nil {
		t.Fatal("foreign manager accepted")
	}
	if n := Cudd_addConst(dd, math.NaN()); n == nil || !math.IsNaN(n.Value) {
		t.Fatal("IEEE NaN constant lost")
	}
	if _, err := dd.EvaluateADD(f, nil); err == nil {
		t.Fatal("missing assignment accepted")
	}
	if Cudd_addIte(dd, dd.Zero, a, dd.One) != dd.One {
		t.Fatal("ADD zero not distinguished from BDD complement")
	}
	Cudd_Quit(dd)
	if len(dd.addUnique) != 0 || len(dd.addCache) != 0 || Cudd_addConst(dd, 2) != nil {
		t.Fatal("ADD lifecycle cleanup")
	}
}

func TestReviewADDGenericSupportAndCounting(t *testing.T) {
	dd := NewBDDManager(2)
	x := Cudd_addIthVar(dd, 0)
	f := Cudd_addIte(dd, x, Cudd_addConst(dd, -2.5), dd.Zero)
	if support := dd.Support(f); len(support) != 1 || support[0] != 0 || Cudd_SupportSize(dd, f) != 1 {
		t.Fatal("ADD omitted from generic support", support)
	}
	if Cudd_Support(dd, f) != dd.Vars[0] {
		t.Fatal("ADD support cube")
	}
	if Cudd_CountMinterm(dd, f, 2) != 2 || Cudd_CountMinterm(dd, dd.Zero, 2) != 0 {
		t.Fatal("ADD omitted from generic minterm count")
	}
	var digits int
	n := Cudd_ApaCountMinterm(dd, f, 2, &digits)
	if apaInteger(digits, n).Int64() != 2 {
		t.Fatal("APA ADD count")
	}
	if Cudd_ReadNodeCount(dd) != len(dd.unique)+len(dd.zddUnique)+len(dd.addUnique)+len(dd.constants) {
		t.Fatal("ADD missing from statistics")
	}
}
