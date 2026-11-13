package aig

import "testing"

func TestDupTerminalMetadataMatchesSource(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	root := p.AND(a, b)
	p.AddOutput(root)
	Aig_ManReprStart(p, int32(len(p.nodes)))
	cases := []struct {
		name string
		fn   func() *Graph
	}{
		{"simple", func() *Graph { return Aig_ManDupSimple(p) }},
		{"array", func() *Graph { return Aig_ManDupArray([]*Graph{p}) }},
		{"nodes", func() *Graph { return Aig_ManDupNodes(p, []Edge{root}) }},
		{"repr", func() *Graph { return Aig_ManDupRepr(p, 0) }},
		{"repr basic", func() *Graph { return Aig_ManDupReprBasic(p) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := tc.fn()
			if q.TruePiCount() != 2 || q.TruePoCount() != 1 {
				t.Fatalf("counts %d/%d", q.TruePiCount(), q.TruePoCount())
			}
			if got := Aig_ObjCioId(q, q.inputs[1]); got != 1 {
				t.Fatalf("second CI ID=%d", got)
			}
		})
	}
}

func TestDupWithoutPosLeavesTerminalMetadataUnset(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(a, b))
	Aig_ManSetRegNum(p, 0)
	q := Aig_ManDupWithoutPos(p)
	if q.TruePiCount() != 0 || q.TruePoCount() != 0 {
		t.Fatalf("counts %d/%d, want zero", q.TruePiCount(), q.TruePoCount())
	}
	if got := Aig_ObjCioId(q, q.inputs[1]); got != 0 {
		t.Fatalf("second CI ID=%d, want uninitialized zero", got)
	}
}

func TestDupOneOutputCopiesStaleMetadataWithoutSettingCioIds(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(a, b))
	// Source starts with zero nTruePis; this routine copies it directly.
	q := Aig_ManDupOneOutput(p, 0, 1)
	if q.TruePiCount() != 0 || q.TruePoCount() != 1 {
		t.Fatalf("counts %d/%d", q.TruePiCount(), q.TruePoCount())
	}
	if got := Aig_ObjCioId(q, q.inputs[1]); got != 0 {
		t.Fatalf("second CI ID=%d, want uninitialized zero", got)
	}
}
