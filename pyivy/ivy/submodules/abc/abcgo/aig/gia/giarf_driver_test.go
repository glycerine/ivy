package gia

import "testing"

func TestGiarfEquivalenceDriver(t *testing.T) {
	for _, mini := range []int32{0, 1} {
		g := Gia_ManStart(12)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x := Gia_ManAppendAnd(g, a, b)
		y := Gia_ManAppendAnd(g, a, b)
		z := Gia_ManAppendAnd(g, y, c)
		Gia_ManAppendCo(g, z)
		Gia_ComputeEquivalences(g, 1000, mini, 0)
		if Gia_ObjRepr(g, y>>1) != x>>1 || Gia_ObjProved(g, y>>1) == 0 {
			t.Fatalf("backend=%d: repr=%d proved=%d", mini, Gia_ObjRepr(g, y>>1), Gia_ObjProved(g, y>>1))
		}
	}
}

func TestGiarfUnknownDoesNotProveClass(t *testing.T) {
	g := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x, y := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, a, c)
	Gia_ManLevelNum(g)
	Gia_ManSetPhase(g)
	// TAS treats a negative conflict limit as already exhausted. Supply a
	// deliberately unresolved simulation class whose members differ.
	p := Gia_ManEquivStart(g, -1, 0)
	Hcd_ManClassesCreate(p)
	Gia_ObjUnsetRepr(g, x>>1)
	Gia_ObjSetRepr(g, y>>1, x>>1)
	Gia_ObjSetNext(g, x>>1, y>>1)
	var roots *Vec_Ptr_t
	reduced := Gia_GenerateReducedLevel(g, 1, &roots)
	Gia_ComputeEquivalencesLevel(p, reduced, roots, 1, 0)
	if Gia_ObjFailed(g, x>>1) == 0 || Gia_ObjFailed(g, y>>1) == 0 {
		t.Fatal("fixture did not exhaust SAT limits")
	}
	if Gia_ObjProved(g, x>>1) != 0 || Gia_ObjProved(g, y>>1) != 0 {
		t.Fatal("unknown SAT results were recorded as proofs")
	}
}

func TestGiarfDropsMembersOfUnprovedRepresentative(t *testing.T) {
	for _, failedHead := range []bool{false, true} {
		g := Gia_ManStart(1)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		ab := Gia_ManAppendAnd(g, a, b)
		head := Gia_ManAppendAnd(g, ab, c)
		low, tail := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, a, b)
		g.StartRepresentatives()
		Gia_ObjSetRepr(g, low>>1, head>>1)
		Gia_ObjSetRepr(g, tail>>1, head>>1)
		Gia_ObjSetNext(g, head>>1, low>>1)
		Gia_ObjSetNext(g, low>>1, tail>>1)
		// The level-one comparison proves low == tail. It says nothing about
		// their higher-level head (a & b & c). A later unknown check marks the
		// head and temporary lower representative, leaving tail's flag intact.
		Gia_ObjSetProved(g, low>>1)
		Gia_ObjSetProved(g, tail>>1)
		if failedHead {
			Gia_ObjSetFailed(g, head>>1)
			Gia_ObjSetFailed(g, low>>1)
		}
		giaDropUnprovedEquivalences(g)
		if Gia_ObjHasRepr(g, low>>1) != 0 || Gia_ObjHasRepr(g, tail>>1) != 0 || Gia_ManCountChoices(g) != 0 {
			t.Fatalf("failed head %t: kept members without proving their actual representative", failedHead)
		}
	}
}

func TestGiarfProvesHighestLevelCandidates(t *testing.T) {
	for _, mini := range []int32{0, 1} {
		g := Gia_ManStart(1)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		x, y := Gia_ManAppendAnd(g, a, b), Gia_ManAppendAnd(g, a, b)
		Gia_ManAppendCo(g, x)
		Gia_ManAppendCo(g, y)
		Gia_ComputeEquivalences(g, 1000, mini, 0)
		if Gia_ObjRepr(g, y>>1) != x>>1 || Gia_ObjProved(g, y>>1) == 0 {
			t.Fatalf("backend %d: highest-level representative=%d proved=%d", mini, Gia_ObjRepr(g, y>>1), Gia_ObjProved(g, y>>1))
		}
	}
}

func TestGiarfDropsUnprovedBudgetLimitedConstant(t *testing.T) {
	for _, mini := range []int32{0, 1} {
		g := Gia_ManStart(1)
		var holes [4][3]int32
		for pigeon := range holes {
			for hole := range holes[pigeon] {
				holes[pigeon][hole] = Gia_ManAppendCi(g)
			}
		}
		Gia_ManHashAlloc(g)
		root := int32(1)
		// Four pigeons into three holes is UNSAT, but needs more than this
		// deliberately tiny SAT budget to prove. Each prefix is satisfiable.
		for _, row := range holes {
			clause := Gia_ManHashOr(g, Gia_ManHashOr(g, row[0], row[1]), row[2])
			root = Gia_ManHashAnd(g, root, clause)
		}
		for first := 0; first < 4; first++ {
			for second := first + 1; second < 4; second++ {
				for hole := 0; hole < 3; hole++ {
					clause := Gia_ManHashOr(g, holes[first][hole]^1, holes[second][hole]^1)
					root = Gia_ManHashAnd(g, root, clause)
				}
			}
		}
		Gia_ManHashStop(g)
		Gia_ManAppendCo(g, root)
		Gia_ComputeEquivalences(g, 1, mini, 0)
		if Gia_ObjFailed(g, root>>1) == 0 {
			t.Fatalf("backend %d: fixture did not exhaust the positive conflict budget", mini)
		}
		if Gia_ObjProved(g, root>>1) != 0 || Gia_ObjHasRepr(g, root>>1) != 0 {
			t.Fatalf("backend %d: unchecked constant retained as a proved equivalence", mini)
		}
	}
}
