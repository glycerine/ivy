package aig

import "testing"

func TestSplitCofactorsPreservePropertyAndState(t *testing.T) {
	p := NewGraph()
	a, state := p.NewInput(), p.NewInput()
	property := p.NewXor(a, state).Not()
	p.AddOutput(property)
	p.AddOutput(p.NewXor(a, state))
	Aig_ManSetRegNum(p, 1)
	q := Aig_ManSplit(p, 2, 0)
	if q == nil || q.nRegs != 1 || len(q.outputs) != 3 {
		t.Fatal("incorrect split structure")
	}
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		before, err := p.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		after, err := q.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		combined := false
		for _, v := range after[:len(after)-1] {
			combined = combined || v
		}
		if combined != before[0] || after[len(after)-1] != before[1] {
			t.Fatal(mask, before, after)
		}
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestShannonReductionDetectsSemanticConstants(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	zero := p.AND(p.AND(p.OR(a, b), a.Not()), b.Not())
	p.AddOutput(zero)
	var function Edge
	decisions := Aig_ManBuildPoBdd(p, &function)
	if function != decisions.Const0() {
		t.Fatal("semantic zero not reduced")
	}
	cofs := Aig_ManCofactorBdds(p, decisions.inputs[:1], decisions, function)
	for _, e := range cofs {
		if e != decisions.Const0() {
			t.Fatal(cofs)
		}
	}
	q := Aig_ManConvertBddsToAigs(p, decisions, cofs)
	if q.OutputCount() != 0 {
		t.Fatal("zero cofactors retained")
	}
}
func TestRandomSubsetDoesNotMutateInput(t *testing.T) {
	values := []Edge{2, 4, 6, 8}
	subset := Aig_ManVecRandSubset(values, 2)
	if len(subset) != 2 || len(values) != 4 || values[0] != 2 || values[3] != 8 {
		t.Fatal(values, subset)
	}
	seen := map[Edge]bool{}
	for _, e := range subset {
		if seen[e] {
			t.Fatal("duplicate subset")
		}
		seen[e] = true
	}
}
