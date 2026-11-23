package goivy

import "testing"

func TestFragmentCycleChoosesFirstParallelArc(t *testing.T) {
	start, left, alias, right, leaf := NewUFNode(), NewUFNode(), NewUFNode(), NewUFNode(), NewUFNode()
	UFUnify(left, alias)
	first := arc{from: left, to: right, fmla: True, loc: Location{Filename: "cycle.ivy", Line: 7}, hasIdx: true, argIdx: 2}
	later := arc{from: alias, to: right, fmla: False, loc: Location{Filename: "cycle.ivy", Line: 8}, hasIdx: true, argIdx: 1}
	back := arc{from: right, to: alias, fmla: True, loc: Location{Filename: "cycle.ivy", Line: 9}}
	c := checker{arcs: []arc{{from: start, to: alias}, {from: left, to: leaf}, first, first, later}}
	if got := c.findCycle(); len(got) != 0 {
		t.Fatalf("acyclic graph returned %v", got)
	}
	c.arcs = append(c.arcs, back)
	if got := c.findCycle(); len(got) != 2 || got[0] != first || got[1] != back {
		t.Fatalf("unexpected cycle with parallel arcs: %v", got)
	}
	self := arc{from: alias, to: left, fmla: True}
	c.arcs = []arc{self, first, back}
	if got := c.findCycle(); len(got) != 1 || got[0] != self {
		t.Fatalf("expected the first self-loop: %v", got)
	}
}

func BenchmarkFragmentCycleTraversal(b *testing.B) {
	nodes := make([]*UFNode, 129)
	for i := range nodes {
		nodes[i] = NewUFNode()
	}
	var formula Expr = &LogicNot{Body: True}
	for i := 0; i < 8; i++ {
		formula = &LogicOr{Terms: []Expr{True, formula}}
	}
	for _, tc := range []struct {
		name    string
		formula Expr
	}{{"Leaf", True}, {"Nested", formula}} {
		var c checker
		for i := 0; i+1 < len(nodes); i++ {
			a := arc{from: nodes[i], to: nodes[i+1], fmla: tc.formula}
			c.arcs = append(c.arcs, a, a)
			if i+2 < len(nodes) {
				a.to = nodes[i+2]
				c.arcs = append(c.arcs, a)
			}
		}
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if cycle := c.findCycle(); len(cycle) != 0 {
					b.Fatal(cycle)
				}
			}
		})
	}
}
