package aig

import "testing"

func TestXorStructuralLevelsMatchSourceCost(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x := g.NewXor(a, b)
	root := g.AND(x, c)
	out := g.AddOutput(root)
	if g.node(x).Level != 2 || g.node(root).Level != 3 || Aig_ObjLevelNew(g, x) != 2 {
		t.Errorf("structural XOR/AND levels = %d/%d, calculated XOR = %d; want 2/3 and 2", g.node(x).Level, g.node(root).Level, Aig_ObjLevelNew(g, x))
	}
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	g.node(a).Level = 4
	Aig_ManUpdateLevel(g, x)
	if g.node(x).Level != 6 || g.node(root).Level != 7 {
		t.Errorf("incremental XOR/AND levels = %d/%d, want 6/7", g.node(x).Level, g.node(root).Level)
	}
	g.recomputeLevels()
	if g.node(x).Level != 6 || g.node(root).Level != 7 || g.node(out).Level != 7 {
		t.Error("full recomputation lost the source XOR delay")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestXorChoiceLevelsRemainSourceUnitCost(t *testing.T) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	x := g.NewXor(a, b)
	root := g.AND(x, c)
	g.AddOutput(root)
	dangling := g.AND(c, d)
	if level := Aig_ManChoiceLevel(g); level != 2 || g.node(x).Level != 1 || g.node(dangling).Level != 0 {
		t.Fatal("choice levels must retain C's unit-cost and reset semantics")
	}
	if err := g.Check(); err != nil {
		t.Errorf("valid source choice annotations rejected: %v", err)
	}
	g.node(x).Level = 0
	if g.Check() == nil {
		t.Error("checker accepted corrupt annotated XOR level")
	}
	g.recomputeLevels()
	if len(g.levelModes) != 0 {
		t.Error("structural recomputation retained unnecessary annotation storage")
	}
	if g.node(x).Level != 2 || g.node(root).Level != 3 || g.node(dangling).Level != 1 {
		t.Error("structural recomputation retained choice annotations")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestIncrementalXorTimingReplacesChoiceCostLocally(t *testing.T) {
	g := NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	x := g.NewXor(a, b)
	root := g.AND(x, c)
	g.AddOutput(root)
	Aig_ManChoiceLevel(g)
	appended := g.AND(x, d)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	Aig_ManUpdateLevel(g, x)
	if g.node(x).Level != 2 || g.node(root).Level != 3 || g.node(appended).Level != 3 {
		t.Fatal("structural update failed to replace choice costs on affected nodes")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsUnitCostXorWithoutChoiceAnnotation(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x := g.NewXor(a, b)
	g.AddOutput(x)
	g.node(x).Level = 1
	if g.Check() == nil {
		t.Fatal("structural checker accepted XOR delay 1 instead of C's delay 2")
	}
	Aig_ManChoiceLevel(g)
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	Aig_ObjSetLevel(g, x, 2)
	if err := g.Check(); err != nil {
		t.Fatalf("explicit structural level write retained a choice annotation: %v", err)
	}
}
