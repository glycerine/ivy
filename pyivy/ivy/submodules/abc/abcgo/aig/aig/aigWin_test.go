package aig

import "testing"

func TestReconvergentWindowCut(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	ab := p.AND(a, b)
	ac := p.AND(a, c)
	root := p.AND(ab, ac)
	var front, visited []Edge
	Aig_ManFindCut(p, root, &front, &visited, 3, 100)
	if len(front) != 3 {
		t.Fatal(front)
	}
	seen := map[Edge]bool{}
	for _, e := range front {
		seen[e] = true
	}
	if !seen[a] || !seen[b] || !seen[c] {
		t.Fatal(front)
	}
	Aig_ManCheckMarkA(p)
	table, err := Aig_ManCutTruth(p, root, front)
	if err != nil || table[0]&255 != 128 {
		t.Fatal(table, err)
	}
}
func TestWindowSizeAndFanoutLimits(t *testing.T) {
	p := NewGraph()
	a, b, c, d := p.NewInput(), p.NewInput(), p.NewInput(), p.NewInput()
	ab, cd := p.AND(a, b), p.AND(c, d)
	root := p.AND(ab, cd)
	p.AddOutput(ab)
	p.AddOutput(ab)
	p.AddOutput(cd)
	p.AddOutput(cd)
	var front, visited []Edge
	Aig_ManFindCut(p, root, &front, &visited, 4, 1)
	if len(front) != 2 {
		t.Fatal("high fanout limit ignored", front)
	}
	Aig_ManCheckMarkA(p)
	Aig_ManFindCut(p, root, &front, &visited, 3, 100)
	if len(front) > 3 {
		t.Fatal("size bound ignored", front)
	}
	Aig_ManCheckMarkA(p)
}
