package aig

import (
	"slices"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestSplitDoesNotDependOnTerminalScratchIDs(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(a, b))
	// CI IDs are scratch state, not the variable numbering used by C's BDD
	// builder. NewInput leaves both at zero until an explicit indexing pass.
	q := Aig_ManSplit(p, 2, 0)
	if q == nil || len(q.outputs) != 1 {
		t.Fatal("AND property should have one nonzero minterm")
	}
	for mask := 0; mask < 4; mask++ {
		input := []bool{mask&1 != 0, mask&2 != 0}
		want, _ := p.Evaluate(input)
		got, err := q.Evaluate(input)
		if err != nil || !slices.Equal(got, want) {
			t.Fatalf("mask %d: got %v (%v), want %v", mask, got, err, want)
		}
	}
}

func TestSplitRandomSubsetUsesSourceStream(t *testing.T) {
	values := []Edge{2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24}
	rng := NewAigRandomSource()
	rng.Uint32() // Aig_ManRandom(1) returns and consumes the first word.
	want := slices.Clone(values)
	for len(want) > 4 {
		i := int(rng.Uint32() % uint32(len(want)))
		want = append(want[:i], want[i+1:]...)
	}
	Aig_ManRandom(1)
	defer Aig_ManRandom(1)
	if got := Aig_ManVecRandSubset(values, 4); !slices.Equal(got, want) {
		t.Fatalf("subset = %v, want source stream %v", got, want)
	}
}

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
func TestSplitBddReductionDetectsSemanticConstants(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	zero := p.AND(p.AND(p.OR(a, b), a.Not()), b.Not())
	p.AddOutput(zero)
	var function *cudd.DdNode
	decisions := Aig_ManBuildPoBdd(p, &function)
	defer cudd.Cudd_Quit(decisions)
	defer cudd.Cudd_RecursiveDeref(decisions, function)
	if function != cudd.Cudd_ReadLogicZero(decisions) {
		t.Fatal("semantic zero not reduced")
	}
	cofs := Aig_ManCofactorBdds(p, decisions.Vars[:1], decisions, function)
	for _, e := range cofs {
		defer cudd.Cudd_RecursiveDeref(decisions, e)
		if e != cudd.Cudd_ReadLogicZero(decisions) {
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
