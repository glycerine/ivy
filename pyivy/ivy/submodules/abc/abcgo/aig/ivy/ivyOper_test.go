package ivy

import "testing"

func TestBooleanOperationsAndFanoutLifecycle(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	Ivy_ManStartFanout(p)
	roots := []*Ivy_Obj_t{Ivy_And(p, a, b), Ivy_Or(p, a, b), Ivy_Exor(p, a, b), Ivy_Mux(p, a, b, c), Ivy_Maj(p, a, b, c), Ivy_Mux(p, a, a, p.pConst1)}
	for _, root := range roots {
		Ivy_ObjCreatePo(p, root)
	}
	if Ivy_ManCheck(p) == 0 {
		t.Fatal("constructed graph failed check")
	}
	for bits := 0; bits < 8; bits++ {
		av, bv, cv := bits&1 != 0, bits&2 != 0, bits&4 != 0
		want := []bool{av && bv, av || bv, av != bv, av && bv || !av && cv, av && bv || av && cv || bv && cv, true}
		got := ivyBalanceEvaluate(p, bits)
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("bits=%d output=%d got=%v want=%v", bits, i, got[i], want[i])
			}
		}
	}
	for len(p.vPos) > 0 {
		Ivy_ObjDelete_rec(p, p.vPos[len(p.vPos)-1], 1)
	}
	Ivy_ManCleanup(p)
	if Ivy_ManCheck(p) == 0 || Ivy_ManNodeNum(p) != 0 {
		t.Fatal("deletion left inconsistent graph")
	}
	for _, n := range p.vPis {
		if n.nRefs != 0 || n.pFanout != nil {
			t.Fatal("stale fanout")
		}
	}
	Ivy_ManStopFanout(p)
}
func TestNativeXorPhase(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	x := Ivy_CanonExor(p, a, b)
	if Ivy_Regular(x).fPhase != 0 {
		t.Fatal("wrong all-zero XOR phase")
	}
	// Complemented AND has phase 1; XOR with phase-0 input must have phase 1.
	high := Ivy_Not(Ivy_And(p, a, b))
	native := Ivy_ObjCreate(p, Ivy_ObjCreateGhost(p, high, a, IVY_EXOR, IVY_INIT_NONE))
	if native.fPhase != 1 {
		t.Fatal("XOR creation used AND phase rule")
	}
}
