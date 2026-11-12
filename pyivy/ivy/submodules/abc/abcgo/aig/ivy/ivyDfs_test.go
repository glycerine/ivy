package ivy

import "testing"

func TestDfsConeAndRequiredLevels(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	ab := Ivy_And(p, a, b)
	root := Ivy_And(p, ab, c)
	Ivy_ObjCreatePo(p, root)
	ids := Ivy_ManDfs(p)
	if len(ids) != 2 || ids[0] != ab.Id || ids[1] != root.Id {
		t.Fatalf("DFS order %v", ids)
	}
	var cone []*Ivy_Obj_t
	Ivy_ManCollectCone(root, []*Ivy_Obj_t{a, b, c, a}, &cone)
	if len(cone) != 2 || cone[0] != ab || cone[1] != root {
		t.Fatal("wrong cone")
	}
	for _, n := range p.vObjs {
		if n != nil && n.fMarkA != 0 {
			t.Fatal("cone leaked marks")
		}
	}
	root.Level = 99
	ab.Level = 88
	if Ivy_ManSetLevels(p, 0) != 2 || root.Level != 2 || ab.Level != 1 {
		t.Fatal("levels not recomputed")
	}
	required := Ivy_ManRequiredLevels(p)
	if required[a.Id] != 0 || required[ab.Id] != 1 || required[root.Id] != 2 || required[c.Id] != 1 {
		t.Fatalf("required levels %v", required)
	}
}
func TestAcyclicDetectsChoiceDependency(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	root := Ivy_And(p, a, b)
	Ivy_ObjCreatePo(p, root)
	// A CI choice depending on that same CI is a combinational choice cycle.
	a.pEquiv = root
	root.pEquiv = a
	if Ivy_ManIsAcyclic(p) != 0 {
		t.Fatal("choice-induced dependency cycle accepted")
	}
}

func TestSequentialDfsCollectsLatchInputs(t *testing.T) {
	p := Ivy_ManStart()
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	next := Ivy_And(p, a, b)
	latch := Ivy_Latch(p, next, IVY_INIT_0)
	output := Ivy_And(p, latch, a)
	Ivy_ObjCreatePo(p, output)
	var latches []int32
	ids := Ivy_ManDfsSeq(p, &latches)
	if len(latches) != 1 || latches[0] != Ivy_Regular(latch).Id || len(ids) != 2 {
		t.Fatalf("sequential DFS nodes=%v latches=%v", ids, latches)
	}
	if Ivy_ManCheck(p) == 0 {
		t.Fatal("sequential graph check failed")
	}
	if Ivy_ManSetLevels(p, 0) != 1 {
		t.Fatal("latches should cut combinational levels")
	}
}
