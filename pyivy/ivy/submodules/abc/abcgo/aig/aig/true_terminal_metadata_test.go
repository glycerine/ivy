package aig

import "testing"

func TestTrueTerminalMetadataConstructionAndDuplication(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	g.NewInput()
	g.AddOutput(a)
	if g.TruePiCount() != 0 || g.TruePoCount() != 0 {
		t.Fatal("new graph must begin with zero metadata")
	}
	d := Aig_ManDupSimple(g)
	if d.TruePiCount() != 2 || d.TruePoCount() != 1 {
		t.Fatalf("duplicate counts %d/%d", d.TruePiCount(), d.TruePoCount())
	}
	Aig_ManSetRegNum(g, 0)
	if g.TruePiCount() != 2 || g.TruePoCount() != 1 {
		t.Fatal("set register count did not initialize metadata")
	}
	g.NewInput()
	g.AddOutput(a)
	if g.TruePiCount() != 2 || g.TruePoCount() != 1 {
		t.Fatal("append must not implicitly refresh metadata")
	}
}

func TestTrueInputMetadataCleanup(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	g.NewInput()
	g.AddOutput(a)
	Aig_ManSetRegNum(g, 0)
	Aig_ManCiCleanup(g)
	if g.TruePiCount() != 2 {
		t.Fatalf("combinational cleanup should retain C's stale count, got %d", g.TruePiCount())
	}
	h := NewGraph()
	x := h.NewInput()
	h.NewInput()
	q := h.NewInput()
	h.AddOutput(x)
	h.AddOutput(q)
	Aig_ManSetRegNum(h, 1)
	Aig_ManCiCleanup(h)
	if h.TruePiCount() != 1 {
		t.Fatalf("sequential cleanup count = %d, want 1", h.TruePiCount())
	}
}

func TestTrueOutputMetadataCleanup(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	g.AddOutput(a)
	g.AddOutput(g.Const0())
	Aig_ManSetRegNum(g, 0)
	Aig_ManCoCleanup(g)
	if g.TruePoCount() != 2 {
		t.Fatalf("combinational cleanup count = %d, want unchanged 2", g.TruePoCount())
	}
	h := NewGraph()
	x, q := h.NewInput(), h.NewInput()
	h.AddOutput(x)
	h.AddOutput(h.Const0())
	h.AddOutput(q)
	Aig_ManSetRegNum(h, 1)
	Aig_ManCoCleanup(h)
	if h.TruePoCount() != 1 {
		t.Fatalf("sequential cleanup count = %d, want 1", h.TruePoCount())
	}
}

func TestSequentialCleanupRefreshesTrueCounts(t *testing.T) {
	g := NewGraph()
	x, q := g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(q)
	Aig_ManSetRegNum(g, 1)
	g.SetTruePiCount(0)
	g.SetTruePoCount(0)
	Aig_ManSeqCleanup(g)
	if g.TruePiCount() != int32(g.InputCount())-g.RegisterCount() || g.TruePoCount() != int32(g.OutputCount())-g.RegisterCount() {
		t.Fatalf("cleanup left stale counts %d/%d", g.TruePiCount(), g.TruePoCount())
	}
}
