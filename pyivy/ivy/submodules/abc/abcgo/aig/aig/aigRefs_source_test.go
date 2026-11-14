package aig

import "testing"

func TestOutputReferenceCountsExcludeManagerOwnership(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	out := g.AddOutput(gate)
	if got := Aig_ObjRefs(g, out); got != 0 {
		t.Errorf("CO references = %d, want C's zero actual fanouts", got)
	}
	if Aig_ObjRefs(g, gate) != 1 || Aig_ObjRefs(g, a) != 1 || Aig_ObjRefs(g, b) != 1 {
		t.Error("real pin references were not counted")
	}
	for _, reset := range []func(*Graph){(*Graph).recountReferences, Aig_ManResetRefs} {
		reset(g)
		if Aig_ObjRefs(g, out) != 0 || Aig_ObjRefs(g, gate) != 1 {
			t.Error("reference reset restored synthetic output ownership")
		}
		if err := g.Check(); err != nil {
			t.Fatal(err)
		}
	}
	Aig_ManCleanup(g)
	if !g.valid(out) || !g.valid(gate) {
		t.Error("cleanup lost a zero-fanout CO or its cone")
	}
}

func TestOutputReferenceHelpersMatchSourceCounters(t *testing.T) {
	g := NewGraph()
	out := g.AddOutput(g.NewInput())
	Aig_ObjRef(g, out)
	if Aig_ObjRefs(g, out) != 1 {
		t.Error("CO reference increment included an ownership anchor")
	}
	Aig_ObjDeref(g, out)
	if Aig_ObjRefs(g, out) != 0 {
		t.Error("CO dereference did not restore zero")
	}
	Aig_ObjRef(g, out)
	Aig_ObjClearRef(g, out)
	if err := g.Check(); err != nil {
		t.Errorf("clearing a CO's real references should preserve ownership: %v", err)
	}
}

func TestOutputDereferenceCannotConsumeOwnership(t *testing.T) {
	g := NewGraph()
	out := g.AddOutput(g.NewInput())
	defer func() {
		if recover() == nil {
			t.Error("dereference consumed output ownership instead of rejecting zero references")
		}
	}()
	Aig_ObjDeref(g, out)
}

func TestAttachedOutputStillRequiresOutputDeletion(t *testing.T) {
	g := NewGraph()
	out := g.AddOutput(g.NewInput())
	func() {
		defer func() {
			if recover() == nil {
				t.Error("low-level deletion accepted an attached output")
			}
		}()
		Aig_ObjDelete(g, out)
	}()
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	Aig_ObjDeletePo(g, out)
	if g.valid(out) || g.OutputCount() != 0 {
		t.Fatal("output-aware deletion failed")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestFanoutObjectAccessorUsesMaintainedHead(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	first, second := g.AND(a, b), g.AND(a, c)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	Aig_ObjRemoveFanoutPin(g, a, first, 0)
	Aig_ObjAddFanoutPin(g, a, first, 0)
	if got := Aig_ObjFanout0(g, a); got != second {
		t.Errorf("fanout object = %d, want maintained head %d", got, second)
	}
}

func TestFanoutObjectAccessorRequiresSourceIndex(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	g.AddOutput(a)
	defer func() {
		if recover() == nil {
			t.Error("fanout accessor accepted missing source fanout allocation")
		}
	}()
	Aig_ObjFanout0(g, a)
}

func TestFanoutObjectAccessorPreservesSourceZeroHandleConversion(t *testing.T) {
	g := NewGraph()
	out := g.AddOutput(g.NewInput())
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	// C's raw accessor shifts the handle and indexes vObjs unconditionally.
	// With a zero head that yields object zero; iteration is bounded by refs.
	if got := Aig_ObjFanout0(g, out); got != g.Const1() {
		t.Fatalf("zero fanout handle converted to %d, want source object zero", got)
	}
}

func TestReferenceResetPreservesLegacyRawOutputEdges(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	g.outputs = []Edge{root.Not()}
	g.recountReferences()
	Aig_ManResetRefs(g)
	if Aig_ObjRefs(g, root) != 1 {
		t.Fatal("legacy raw output lost its implicit CO use")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	Aig_ObjDeletePo(g, root.Not())
	if Aig_ObjRefs(g, root) != 0 || g.OutputCount() != 0 {
		t.Fatal("legacy output deletion retained its implicit use")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}
