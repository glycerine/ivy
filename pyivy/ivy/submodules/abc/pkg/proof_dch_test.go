package abc

// Stand-alone tests for src/proof/dch.

import (
	"errors"
	"slices"
	"testing"
)

func dchTestCI(t testing.TB, g *AigManager) AigLit {
	t.Helper()
	literal, err := g.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func dchTestAnd(t testing.TB, g *AigManager, first, second AigLit) AigLit {
	t.Helper()
	literal, err := g.And(first, second)
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func dchTestCO(t testing.TB, g *AigManager, driver AigLit) {
	t.Helper()
	if _, err := g.CreateCO(driver); err != nil {
		t.Fatal(err)
	}
}

func dchTestAssociative(t testing.TB, signs uint8) (*AigManager, AigID, AigID) {
	t.Helper()
	g := NewAigManager()
	x, y, z := dchTestCI(t, g), dchTestCI(t, g), dchTestCI(t, g)
	x, y, z = x.NotIf(signs&1 != 0), y.NotIf(signs&2 != 0), z.NotIf(signs&4 != 0)
	left := dchTestAnd(t, g, dchTestAnd(t, g, x, y), z)
	right := dchTestAnd(t, g, x, dchTestAnd(t, g, y, z))
	if left.ID() == right.ID() || left.Complemented() || right.Complemented() {
		t.Fatalf("associative roots collapsed: %d %d", left, right)
	}
	dchTestCO(t, g, left)
	dchTestCO(t, g, right)
	return g, left.ID(), right.ID()
}

func dchTestObjectValues(t testing.TB, g *AigManager, inputs []bool) []bool {
	t.Helper()
	if len(inputs) != len(g.cis) {
		t.Fatal("input dimension")
	}
	values := make([]bool, len(g.objects))
	values[0] = true
	for index, id := range g.cis {
		values[id] = inputs[index]
	}
	read := func(literal AigLit) bool { return values[literal.ID()] != literal.Complemented() }
	for id, object := range g.objects {
		switch object.Type {
		case AigAND:
			values[id] = read(object.Fanin0) && read(object.Fanin1)
		case AigBuffer, AigCO:
			values[id] = read(object.Fanin0)
		}
	}
	return values
}

func dchTestCheckResult(t testing.TB, source *AigManager, result DchResult) {
	t.Helper()
	if result.Swept == nil || len(result.Replacements) != source.ObjectCount() {
		t.Fatalf("incomplete result: %+v", result.Stats)
	}
	for assignment := 0; assignment < 1<<source.InputCount(); assignment++ {
		inputs := make([]bool, source.InputCount())
		for index := range inputs {
			inputs[index] = assignment>>index&1 != 0
		}
		want, err := source.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := result.Swept.Evaluate(inputs)
		if err != nil || !slices.Equal(got, want) {
			t.Fatalf("swept assignment %d = %v, %v; want %v", assignment, got, err, want)
		}
		if result.Graph != nil {
			choice, choiceErr := result.Graph.Evaluate(inputs)
			if choiceErr != nil || !slices.Equal(choice, want) {
				t.Fatalf("choice assignment %d = %v, %v; want %v", assignment, choice, choiceErr, want)
			}
			choiceValues := dchTestObjectValues(t, result.Graph, inputs)
			for _, chain := range result.Graph.ChoiceChains() {
				for _, alternative := range chain[1:] {
					if choiceValues[alternative] != choiceValues[chain[0]] {
						t.Fatalf("choice chain %v differs on assignment %d", chain, assignment)
					}
				}
			}
		}
		values := dchTestObjectValues(t, source, inputs)
		for _, class := range result.Classes {
			representative := values[class.Representative]
			if class.Representative != class.Members[0].ID || class.Members[0].Inverted {
				t.Fatalf("bad class head: %+v", class)
			}
			for _, member := range class.Members {
				if values[member.ID] != representative != member.Inverted {
					t.Fatalf("class %+v fails assignment %d", class, assignment)
				}
			}
		}
		for id, replacement := range result.Replacements {
			if source.objects[id].Type == AigNone || source.objects[id].Type == AigCO {
				continue
			}
			if values[id] != values[replacement.ID()] != replacement.Complemented() {
				t.Fatalf("replacement %d -> %d fails assignment %d", id, replacement, assignment)
			}
		}
	}
}

func TestDchClassesSATChoicesSweepAndOwnership(t *testing.T) {
	g, left, right := dchTestAssociative(t, 0)
	x := AigLit(g.cis[0]) << 1
	y := AigLit(g.cis[1]) << 1
	or, err := g.Or(x, y)
	if err != nil {
		t.Fatal(err)
	}
	zero := dchTestAnd(t, g, dchTestAnd(t, g, or, x.Not()), y.Not())
	dchTestCO(t, g, zero)
	before := g.Clone()
	result, err := ComputeDchChoices(g, DchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(g.objects, before.objects) || !slices.Equal(g.equivs, before.equivs) {
		t.Fatal("source graph mutated")
	}
	if result.Graph == g || result.Swept == g || result.Stats.Proved == 0 || result.Stats.SATCalls == 0 {
		t.Fatalf("missing owned SAT result: %+v", result.Stats)
	}
	foundPair, foundZero := false, false
	for _, class := range result.Classes {
		ids := make([]AigID, len(class.Members))
		for index, member := range class.Members {
			ids[index] = member.ID
		}
		if slices.Contains(ids, left) && slices.Contains(ids, right) {
			foundPair = true
		}
		if class.Representative == 0 && slices.Contains(ids, zero.ID()) {
			foundZero = true
		}
	}
	if !foundPair || !foundZero {
		t.Fatalf("classes did not contain pair/constant: %+v", result.Classes)
	}
	chains := result.Graph.ChoiceChains()
	if len(chains) == 0 || result.Stats.Choices == 0 {
		t.Fatalf("missing choices: %v %+v", chains, result.Stats)
	}
	for _, chain := range chains {
		for _, alternative := range chain[1:] {
			object, objectErr := result.Graph.Object(alternative)
			if objectErr != nil || object.RefCount != 0 {
				t.Fatalf("choice alternative %d is not dangling: %+v %v", alternative, object, objectErr)
			}
		}
	}
	if _, err = result.Graph.DFSChoices(AigTraversalOptions{}); err != nil {
		t.Fatalf("choice-aware traversal: %v", err)
	}
	second, err := ComputeDchChoices(g, DchOptions{})
	if err != nil || len(second.Classes) != len(result.Classes) || !slices.Equal(second.Replacements, result.Replacements) {
		t.Fatalf("nondeterministic dimensions/replacements: %v", err)
	}
	for index := range result.Classes {
		if second.Classes[index].Representative != result.Classes[index].Representative || !slices.Equal(second.Classes[index].Members, result.Classes[index].Members) {
			t.Fatalf("nondeterministic class %d: %+v %+v", index, result.Classes[index], second.Classes[index])
		}
	}
	secondChains := second.Graph.ChoiceChains()
	if len(secondChains) != len(chains) {
		t.Fatalf("nondeterministic chain count: %v %v", chains, secondChains)
	}
	for index := range chains {
		if !slices.Equal(secondChains[index], chains[index]) {
			t.Fatalf("nondeterministic chain %d: %v %v", index, chains[index], secondChains[index])
		}
	}
	dchTestCheckResult(t, g, result)
	result.Classes[0].Members[0].ID = InvalidAigID
	result.Replacements[0] = InvalidAigLit
	if g.objects[0].Type != AigConstant1 || len(g.ChoiceChains()) != 0 {
		t.Fatal("result aliases source")
	}
}

func TestDchTotalAigAndValidation(t *testing.T) {
	first, _, _ := dchTestAssociative(t, 0)
	second, _, _ := dchTestAssociative(t, 7)
	total, err := DeriveDchTotalAig([]*AigManager{first, second}, DchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if total.InputCount() != first.InputCount() || total.OutputCount() != first.OutputCount() || total.AndCount() <= first.AndCount() {
		t.Fatalf("total dimensions: inputs %d outputs %d ands %d", total.InputCount(), total.OutputCount(), total.AndCount())
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, _ := first.Evaluate(inputs)
		got, evalErr := total.Evaluate(inputs)
		if evalErr != nil || !slices.Equal(got, want) {
			t.Fatalf("total assignment %d = %v, %v; want %v", assignment, got, evalErr, want)
		}
	}
	bad := NewAigManager()
	dchTestCI(t, bad)
	if result, err := DeriveDchTotalAig([]*AigManager{first, bad}, DchOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("dimension result %v error %v", result, err)
	}
}

func TestDchChoiceRebuildCollapsesRepresentativeFanins(t *testing.T) {
	g := NewAigManager()
	x, y, z, w := dchTestCI(t, g), dchTestCI(t, g), dchTestCI(t, g), dchTestCI(t, g)
	left := dchTestAnd(t, g, dchTestAnd(t, g, x, y), z)
	right := dchTestAnd(t, g, x, dchTestAnd(t, g, y, z))
	leftUse := dchTestAnd(t, g, left, w)
	rightUse := dchTestAnd(t, g, right, w)
	dchTestCO(t, g, leftUse)
	dchTestCO(t, g, rightUse)
	result, err := ComputeDchChoices(g, DchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Stats.Proved < 2 {
		t.Fatalf("expected both equivalences: %+v", result.Stats)
	}
	if result.Stats.Choices != 1 || len(result.Graph.ChoiceChains()) != 1 {
		t.Fatalf("representative fanins were not structurally collapsed: %v %+v", result.Graph.ChoiceChains(), result.Stats)
	}
	dchTestCheckResult(t, g, result)
}

func TestDchComplementPhaseReplacementDoesNotBecomeChoice(t *testing.T) {
	g := NewAigManager()
	a, b, z, w := dchTestCI(t, g), dchTestCI(t, g), dchTestCI(t, g), dchTestCI(t, g)
	function := dchTestAnd(t, g, a, b)
	zw := dchTestAnd(t, g, z, w)
	znw := dchTestAnd(t, g, z, w.Not())
	zAgain, err := g.Or(zw, znw)
	if err != nil {
		t.Fatal(err)
	}
	tautology, err := g.Or(zAgain, z.Not())
	if err != nil {
		t.Fatal(err)
	}
	complement := dchTestAnd(t, g, function.Not(), tautology)
	dchTestCO(t, g, function)
	dchTestCO(t, g, complement)
	result, err := ComputeDchChoices(g, DchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, class := range result.Classes {
		if class.Representative != function.ID() {
			continue
		}
		for _, member := range class.Members {
			if member.ID == complement.ID() && member.Inverted {
				found = true
			}
		}
	}
	if !found || result.Replacements[complement.ID()] != function.Not() {
		t.Fatalf("complement phase relation missing: classes=%+v replacement=%v", result.Classes, result.Replacements[complement.ID()])
	}
	if result.Stats.Choices != 0 || len(result.Graph.ChoiceChains()) != 0 {
		t.Fatalf("complement entered choice metadata: %v", result.Graph.ChoiceChains())
	}
	dchTestCheckResult(t, g, result)
}

func TestDchLimitsCancellationAndAtomicFailure(t *testing.T) {
	g, _, _ := dchTestAssociative(t, 0)
	for _, options := range []DchOptions{
		{SimulationWords: -1}, {MaxObjects: -1}, {MaxSimulationWords: -1}, {MaxCandidatePairs: -1}, {MaxSATVariables: -1},
	} {
		if result, err := ComputeDchChoices(g, options); result.Graph != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("options %+v result %+v error %v", options, result.Stats, err)
		}
	}
	for _, options := range []DchOptions{
		{MaxObjects: g.ObjectCount() - 1},
		{SimulationWords: 2, MaxSimulationWords: 1},
		{MaxSATVariables: 1},
	} {
		if result, err := ComputeDchChoices(g, options); result.Graph != nil || !errors.Is(err, ErrLimit) {
			t.Fatalf("limit %+v result %+v error %v", options, result.Stats, err)
		}
	}
	x, y, z := AigLit(g.cis[0])<<1, AigLit(g.cis[1])<<1, AigLit(g.cis[2])<<1
	third := dchTestAnd(t, g, dchTestAnd(t, g, x, z), y)
	dchTestCO(t, g, third)
	before := g.Clone()
	if result, err := ComputeDchChoices(g, DchOptions{MaxCandidatePairs: 1}); result.Graph != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("candidate-pair limit result %+v error %v", result.Stats, err)
	}
	calls := 0
	if result, err := ComputeDchChoices(g, DchOptions{Stop: func() bool { calls++; return true }}); result.Graph != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel result %+v error %v", result.Stats, err)
	}
	if !slices.Equal(g.objects, before.objects) || !slices.Equal(g.equivs, before.equivs) {
		t.Fatal("failed operation mutated source")
	}
}

func TestDchIndependentCombinationalModel5000(t *testing.T) {
	for trial := 0; trial < 5000; trial++ {
		g, _, _ := dchTestAssociative(t, uint8(trial))
		result, err := ComputeDchChoices(g, DchOptions{SimulationWords: 1})
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		if result.Stats.Proved == 0 {
			t.Fatalf("trial %d did not prove associativity", trial)
		}
		dchTestCheckResult(t, g, result)
	}
}

func FuzzDchCombinationalEquivalence(f *testing.F) {
	for _, seed := range []uint8{0, 1, 3, 7, 19, 255} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, code uint8) {
		g, _, _ := dchTestAssociative(t, code)
		result, err := ComputeDchChoices(g, DchOptions{SimulationWords: 1})
		if err != nil {
			t.Fatal(err)
		}
		if result.Stats.Proved == 0 {
			t.Fatal("associative equivalence was not proved")
		}
		dchTestCheckResult(t, g, result)
	})
}
