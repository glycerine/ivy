package aig

import "testing"

func TestIncrementalForwardTimingPreservesUnrelatedAndSeededState(t *testing.T) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	child := g.AND(root, c)
	other := g.AND(b, d)
	g.AddOutput(child)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	g.node(a).Level = 4
	g.node(other).Level = 17
	g.node(child).MarkA = true
	Aig_ManUpdateLevel(g, root)
	if g.node(root).Level != 5 || g.node(root).MarkA {
		t.Error("changed node did not receive its new level and clear its work mark")
	}
	if g.node(other).Level != 17 || g.node(child).Level != 2 || !g.node(child).MarkA {
		t.Error("incremental update changed unrelated or premarked timing state")
	}
}

func TestIncrementalForwardTimingPropagatesIncreasesAndDecreases(t *testing.T) {
	for _, increase := range []bool{false, true} {
		g := NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		if !increase {
			g.node(a).Level = 4
		}
		root := g.AND(a, b)
		child := g.AND(root, c)
		out := g.AddOutput(child)
		oldCo := g.node(out).Level
		Aig_ManFanoutStart(g)
		if increase {
			g.node(a).Level = 4
		} else {
			g.node(a).Level = 0
		}
		Aig_ManUpdateLevel(g, root)
		want := g.node(a).Level + 1
		if g.node(root).Level != want || g.node(child).Level != want+1 {
			t.Errorf("increase=%v: levels = %d,%d, want %d,%d", increase, g.node(root).Level, g.node(child).Level, want, want+1)
		}
		if g.node(out).Level != oldCo {
			t.Error("source incremental propagation should leave CO metadata alone")
		}
		if g.node(root).MarkA || g.node(child).MarkA {
			t.Error("incremental propagation leaked queue marks")
		}
		if err := g.Check(); err != nil {
			t.Error(err)
		}
		Aig_ManFanoutStop(g)
	}
}

func TestIncrementalReverseTimingProcessesNewSameLevelNodes(t *testing.T) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	first := g.AND(a, b)
	other := g.AND(c, d)
	g.AddOutput(first)
	g.AddOutput(other)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	Aig_ManStartReverseLevels(g, 0)
	defer Aig_ManStopReverseLevels(g)
	Aig_ObjSetReverseLevel(g, other, 13)
	middle := g.AND(first, c)
	top := g.AND(middle, d)
	g.AddOutput(top)
	oldTrav := g.travID
	Aig_ManUpdateReverseLevel(g, top)
	if g.travID != oldTrav {
		t.Error("incremental timing changed the shared traversal generation")
	}
	if Aig_ObjReverseLevel(g, top) != 1 || Aig_ObjReverseLevel(g, middle) != 2 || Aig_ObjReverseLevel(g, first) != 3 {
		t.Error("reverse propagation missed a node appended to the current old-level bucket")
	}
	if Aig_ObjReverseLevel(g, other) != 13 {
		t.Error("reverse update overwrote an unrelated stored level")
	}
	for _, e := range []Edge{first, middle, top} {
		if g.node(e).MarkA {
			t.Error("reverse propagation leaked queue marks")
		}
	}
	Aig_ObjSetReverseLevel(g, other, 1)
	if err := Aig_ManVerifyReverseLevel(g); err != nil {
		t.Fatal(err)
	}
}

func TestIncrementalTimingRequiresSourceNodeAndFanoutSetup(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, kind := range []string{"input", "output", "constant", "missing-fanouts"} {
			func() {
				g := NewGraph()
				a, b := g.NewInput(), g.NewInput()
				root := g.AND(a, b)
				out := g.AddOutput(root)
				if kind != "missing-fanouts" {
					Aig_ManFanoutStart(g)
					defer Aig_ManFanoutStop(g)
				}
				if reverse {
					g.reverseLevels = make([]int32, len(g.nodes))
				}
				e := root
				switch kind {
				case "input":
					e = a
				case "output":
					e = out
				case "constant":
					e = g.Const1()
				}
				defer func() {
					if recover() == nil {
						t.Errorf("reverse=%v kind=%s: source precondition ignored", reverse, kind)
					}
				}()
				if reverse {
					Aig_ManUpdateReverseLevel(g, e)
				} else {
					Aig_ManUpdateLevel(g, e)
				}
			}()
		}
	}
}

func TestCheckAndLevelVerificationIgnoreNonLogicLevelMetadata(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	out := g.AddOutput(root)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	g.node(out).Level = 23
	if err := g.Check(); err != nil {
		t.Errorf("CO metadata is not a gate level: %v", err)
	}
	if err := Aig_ManVerifyLevel(g); err != nil {
		t.Errorf("source level verification visits only logic gates: %v", err)
	}
}

func TestVerifyLevelRequiresSourceFanoutSetup(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	if Aig_ManVerifyLevel(g) == nil {
		t.Fatal("level verification ignored C's fanout setup precondition")
	}
}
