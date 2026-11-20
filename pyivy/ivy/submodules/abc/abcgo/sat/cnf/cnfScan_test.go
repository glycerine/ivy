package cnf

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestScanSelectedCuts(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	ab := g.AND(a, b)
	abc := g.AND(ab, c)
	top := g.AND(abc, d)
	g.AddOutput(top.Not())
	g.AddOutput(ab)
	g.AddOutput(top)
	g.AddOutput(g.Const1())
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	m.selectedCuts = make([]*Cnf_Cut_t, aig.Aig_ManObjNumMax(g))
	m.selectedCuts[ab.NodeID()] = Cnf_CutFromTruth([]int32{int32(a.NodeID()), int32(b.NodeID())}, []uint32{0x88888888})
	m.selectedCuts[top.NodeID()] = Cnf_CutFromTruth([]int32{int32(ab.NodeID()), int32(c.NodeID()), int32(d.NodeID())}, []uint32{0x80808080})
	m.selectedCuts[ab.NodeID()].Cost = 3
	m.selectedCuts[top.NodeID()].Cost = 4
	for _, tc := range []struct {
		name string
		scan func() *Vec_Ptr_t
		want []any
	}{
		{"DAR preorder", func() *Vec_Ptr_t { return Aig_ManScanMapping(m, 1) }, []any{top, ab}},
		{"CNF preorder", func() *Vec_Ptr_t { return Cnf_ManScanMapping(m, 1, 1) }, []any{top, ab}},
		{"CNF postorder", func() *Vec_Ptr_t { return Cnf_ManScanMapping(m, 1, 0) }, []any{ab, top}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.scan()
			if !slices.Equal(vec.Vec_PtrArray(got), tc.want) {
				t.Errorf("mapped edges %v, want %v", vec.Vec_PtrArray(got), tc.want)
			}
			if m.aArea != 7 {
				t.Errorf("mapped area %d, want 7", m.aArea)
			}
			if g.Node(abc).Refs != 0 || g.Node(top).Refs != 2 || g.Node(ab).Refs != 2 || g.Node(g.Const1()).Refs != 1 {
				t.Errorf("mapped refs for absorbed, root, shared and constant nodes = %d %d %d %d", g.Node(abc).Refs, g.Node(top).Refs, g.Node(ab).Refs, g.Node(g.Const1()).Refs)
			}
			if cost := Cnf_ManScanMapping_rec(m, &top, got, 0); cost != 0 || g.Node(top).Refs != 3 || vec.Vec_PtrSize(got) != 2 {
				t.Errorf("revisiting root cost %d, refs %d, nodes %d", cost, g.Node(top).Refs, vec.Vec_PtrSize(got))
			}
		})
	}
	if got := Cnf_ManScanMapping(m, 0, 1); got != nil || m.aArea != 7 {
		t.Errorf("area-only scan returned %+v with area %d", got, m.aArea)
	}
	aig.Aig_ObjSetMarkB(g, top)
	Cnf_ManScanMapping(m, 0, 0)
	if m.aArea != 8 || g.Node(a).Refs != 2 || g.Node(b).Refs != 2 || g.Node(ab).Refs != 1 || !g.Node(top).MarkB {
		t.Errorf("supergate area %d, refs of a, b and ab = %d %d %d, mark %t", m.aArea, g.Node(a).Refs, g.Node(b).Refs, g.Node(ab).Refs, g.Node(top).MarkB)
	}
	aig.Aig_ObjClearMarkB(g, top)
	m.selectedCuts[ab.NodeID()].Cost = 9
	Aig_ManScanMapping(m, 0)
	if m.aArea != 7 {
		t.Errorf("DAR scan computed SOP cost %d, want 7", m.aArea)
	}
	Cnf_ManScanMapping(m, 0, 0)
	if m.aArea != 13 {
		t.Errorf("transferred cut cost %d, want 13", m.aArea)
	}
}

func TestCNFDerivationRestoresStructuralReferences(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	for _, derive := range []func(){
		func() { Cnf_DeriveMappingArray(g) },
		func() { Cnf_Derive(g, 0) },
		func() {
			if _, err := Cnf_DeriveOtherMapped(g, 0); err != nil {
				t.Fatal(err)
			}
		},
	} {
		derive()
		if err := g.Check(); err != nil || g.Node(inner).Refs != 1 {
			t.Fatalf("derivation left structural references inconsistent: %v", err)
		}
	}
}

func mappedAndChain(gates int) *Cnf_Man_t {
	g := aig.NewGraph()
	m := Cnf_ManStart()
	m.pManAig = g
	m.selectedCuts = make([]*Cnf_Cut_t, 2*gates+3)
	last := g.NewInput()
	for range gates {
		input := g.NewInput()
		gate := g.AND(last, input)
		m.selectedCuts[gate.NodeID()] = Cnf_CutFromTruth([]int32{int32(last.NodeID()), int32(input.NodeID())}, []uint32{0x88888888})
		m.selectedCuts[gate.NodeID()].Cost = 3
		last = gate
	}
	g.AddOutput(last.Not())
	return m
}

func BenchmarkScanBinaryAndChain(b *testing.B) {
	const gates = 1024
	m := mappedAndChain(gates)
	defer Cnf_ManStop(m)
	b.ReportAllocs()
	for b.Loop() {
		mapped := Cnf_ManScanMapping(m, 1, 1)
		if vec.Vec_PtrSize(mapped) != gates {
			b.Fatalf("visited %d gates", vec.Vec_PtrSize(mapped))
		}
	}
}

func BenchmarkWriteMappedBinaryAndChain(b *testing.B) {
	const gates = 1024
	m := mappedAndChain(gates)
	defer Cnf_ManStop(m)
	mapped := Cnf_ManScanMapping(m, 1, 1)
	b.ReportAllocs()
	for b.Loop() {
		d := Cnf_ManWriteCnf(m, mapped, 0)
		if d.nVars != 2*gates+3 || d.nClauses != 3*gates+2 || d.nLiterals != 7*gates+2 {
			b.Fatalf("wrote %d variables, %d clauses and %d literals", d.nVars, d.nClauses, d.nLiterals)
		}
		Cnf_DataFree(d)
	}
}
