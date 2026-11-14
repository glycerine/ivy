package aig

import (
	"reflect"
	"testing"
)

func TestSupportCollectorRespectsCurrentFrontier(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	x := p.AND(a, b)
	root := p.AND(x, c)
	Aig_ManIncrementTravId(p)
	for _, e := range []Edge{a, x, root} {
		Aig_ObjSetTravIdCurrent(p, e)
	}
	got := Aig_SuppMinCollectSupport(p, []Edge{x, root})
	if !reflect.DeepEqual(got, []Edge{b, c}) {
		t.Fatalf("support = %v, want unmarked frontier %v", got, []Edge{b, c})
	}
	for _, e := range got {
		if !Aig_ObjIsTravIdCurrent(p, e) {
			t.Error("collected support missing source traversal mark")
		}
	}
}

func TestSupportReconstructionLeavesSourceCopies(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	x := p.AND(a, b)
	root := p.AND(x, c)
	unrelated := p.NewInput()
	Aig_ObjSetCopy(p, unrelated, c.Not())
	result := Aig_SuppMinReconstruct(p, []Edge{a}, []Edge{x, root}, []Edge{a, b, c})
	for e, want := range map[Edge]Edge{a: p.Const1(), b: b, c: c, x: b, root: result, unrelated: c.Not()} {
		if got := Aig_ObjCopy(p, e); got != want {
			t.Errorf("copy of %d = %d, want %d", e, got, want)
		}
	}
	if result != p.AND(b, c) {
		t.Fatal("incorrect reconstructed cone")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSupportCofactorsAtSixteenVariableBoundary(t *testing.T) {
	p := NewGraph()
	support := make([]Edge, aigSupportLimit)
	for i := range support {
		support[i] = p.NewInput()
	}
	root := p.NewXor(support[0], support[15].Not())
	truth := supportTruth(p, []Edge{root}, support)[root.NodeID()]
	if len(truth) != 2048 {
		t.Fatalf("sixteen-variable table uses %d words, want 2048", len(truth))
	}
	for mask := 0; mask < 1<<16; mask++ {
		got := truth[mask>>5]>>uint(mask&31)&1 != 0
		want := (mask&1 != 0) != (mask&(1<<15) == 0)
		if got != want {
			t.Fatalf("packed XOR differs at assignment %d", mask)
		}
	}
	got := Aig_SuppMinPerform(p, []Edge{support[15]}, []Edge{root}, support)
	want := make([]Edge, 0, 15)
	for _, e := range support[1:15] {
		want = append(want, e.Not())
	}
	want = append(want, support[15])
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cofactors = %v, want %v", got, want)
	}
	result := Aig_SuppMinReconstruct(p, got, []Edge{root}, support)
	if result != support[0] {
		t.Fatalf("cofactored XOR = %d, want %d", result, support[0])
	}
}

func TestPackedSupportTruthAcrossWordBoundary(t *testing.T) {
	for n := 2; n <= 6; n++ {
		p := NewGraph()
		support := make([]Edge, n)
		for i := range support {
			support[i] = p.NewInput()
		}
		x := p.AND(support[0].Not(), support[n-1])
		y := p.NewXor(x, support[1].Not())
		out := p.AddOutput(y.Not())
		tables := supportTruth(p, []Edge{x, y, out}, support)
		for mask := 0; mask < 1<<n; mask++ {
			a, b, last := mask&1 != 0, mask&2 != 0, mask&(1<<(n-1)) != 0
			xv := !a && last
			yv := xv != !b
			for e, want := range map[Edge]bool{x: xv, y: yv, out: !yv} {
				if got := tables[e.NodeID()][mask>>5]>>uint(mask&31)&1 != 0; got != want {
					t.Fatalf("n=%d assignment=%d object=%d: got %v want %v", n, mask, e, got, want)
				}
			}
		}
	}
}
