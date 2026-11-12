package ivy

import "testing"

func TestIvyMultiBuildsEquivalentTreeAndUsesSharedNode(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a, b, c, d := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	shared := Ivy_And(p, a, b)
	result := Ivy_Multi(p, []*Ivy_Obj_t{a, b, c, d}, 4, IVY_AND)
	Ivy_ObjCreatePo(p, result)
	for assignment := 0; assignment < 16; assignment++ {
		want := assignment&1 != 0 && assignment&2 != 0 && assignment&4 != 0 && assignment&8 != 0
		if got := ivyBalanceEvaluate(p, assignment)[0]; got != want {
			t.Fatalf("assignment %04b: got %t, want %t (shared node %d)", assignment, got, want, shared.Id)
		}
	}
	if Ivy_ManCheck(p) == 0 {
		t.Fatal("multi-input construction produced an invalid graph")
	}
	if result == nil || Ivy_ObjLevel(result) == 0 {
		t.Fatal("multi-input operation did not build a gate")
	}
}

func TestIvyMultiPlusFindsExistingCover(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	shared := Ivy_And(p, a, b)
	leaves := Vec_PtrAlloc(3)
	cone := Vec_PtrAlloc(0)
	solutions := Vec_PtrAlloc(3)
	defer Vec_PtrFree(leaves)
	defer Vec_PtrFree(cone)
	defer Vec_PtrFree(solutions)
	for _, leaf := range []*Ivy_Obj_t{a, b, c} {
		Vec_PtrPush(leaves, leaf)
	}
	if Ivy_MultiPlus(p, leaves, cone, IVY_AND, 4, solutions) == 0 {
		t.Fatal("expected the existing two-leaf node and final leaf to cover the support")
	}
	if Vec_PtrSize(solutions) != 2 {
		t.Fatalf("cover contains %d entries, want 2", Vec_PtrSize(solutions))
	}
	seenShared, seenC := false, false
	for i := int32(0); i < Vec_PtrSize(solutions); i++ {
		solution := ivyVecObj(solutions, i)
		seenShared = seenShared || solution == shared
		seenC = seenC || solution == c
	}
	if !seenShared || !seenC {
		t.Fatalf("cover did not select shared node and final leaf: shared=%t c=%t", seenShared, seenC)
	}
}

func TestIvyMultiRejectsInvalidCount(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	defer func() {
		if recover() == nil {
			t.Fatal("invalid multi-input count did not panic")
		}
	}()
	Ivy_Multi(p, nil, 1, IVY_AND)
}
