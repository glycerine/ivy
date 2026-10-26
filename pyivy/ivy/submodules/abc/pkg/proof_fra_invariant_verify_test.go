package abc

// Standalone exhaustive, SAT-model, limit, concurrency and fuzz tests for fraIndVer.c.

import (
	"errors"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func fraInvariantValues(graph *AigManager, inputs []bool) []bool {
	values := make([]bool, graph.ObjectCount())
	values[0] = true
	for index, id := range graph.cis {
		values[id] = inputs[index]
	}
	read := func(literal AigLit) bool { return values[literal.ID()] != literal.Complemented() }
	for id, object := range graph.objects {
		switch object.Type {
		case AigAND:
			values[id] = read(object.Fanin0) && read(object.Fanin1)
		case AigBuffer, AigCO:
			values[id] = read(object.Fanin0)
		}
	}
	return values
}

func fraInvariantClauseValue(clause []AigLit, values []bool) bool {
	for _, literal := range clause {
		if values[literal.ID()] != literal.Complemented() {
			return true
		}
	}
	return false
}

type fraInvariantExpected struct {
	base, induction []int
	premise         bool
}

func fraInvariantExhaustive(t testing.TB, graph *AigManager, clauses [][]AigLit) fraInvariantExpected {
	t.Helper()
	primaryInputs, registers := graph.PrimaryInputCount(), graph.RegisterCount()
	if primaryInputs+registers > 12 {
		t.Fatal("exhaustive fixture too wide")
	}
	baseFailed := make([]bool, len(clauses))
	zeroState := make([]bool, registers)
	for inputMask := 0; inputMask < 1<<primaryInputs; inputMask++ {
		inputs := make([]bool, primaryInputs+registers)
		for index := 0; index < primaryInputs; index++ {
			inputs[index] = inputMask&(1<<index) != 0
		}
		copy(inputs[primaryInputs:], zeroState)
		values := fraInvariantValues(graph, inputs)
		for index, clause := range clauses {
			baseFailed[index] = baseFailed[index] || !fraInvariantClauseValue(clause, values)
		}
	}
	inductionFailed := make([]bool, len(clauses))
	premise := false
	for stateMask := 0; stateMask < 1<<registers; stateMask++ {
		state := make([]bool, registers)
		for index := range state {
			state[index] = stateMask&(1<<index) != 0
		}
		for firstMask := 0; firstMask < 1<<primaryInputs; firstMask++ {
			primary := make([]bool, primaryInputs)
			for index := range primary {
				primary[index] = firstMask&(1<<index) != 0
			}
			firstInputs := append(slices.Clone(primary), state...)
			firstValues := fraInvariantValues(graph, firstInputs)
			all := true
			for _, clause := range clauses {
				all = all && fraInvariantClauseValue(clause, firstValues)
			}
			if !all {
				continue
			}
			premise = true
			_, next, err := graph.Step(primary, state)
			if err != nil {
				t.Fatal(err)
			}
			for secondMask := 0; secondMask < 1<<primaryInputs; secondMask++ {
				secondInputs := make([]bool, primaryInputs+registers)
				for index := 0; index < primaryInputs; index++ {
					secondInputs[index] = secondMask&(1<<index) != 0
				}
				copy(secondInputs[primaryInputs:], next)
				secondValues := fraInvariantValues(graph, secondInputs)
				for index, clause := range clauses {
					inductionFailed[index] = inductionFailed[index] || !fraInvariantClauseValue(clause, secondValues)
				}
			}
		}
	}
	result := fraInvariantExpected{premise: premise}
	for index := range clauses {
		if baseFailed[index] {
			result.base = append(result.base, index)
		}
		if premise && inductionFailed[index] {
			result.induction = append(result.induction, index)
		}
	}
	return result
}

func fraInvariantFailureIndices(failures []FRAInvariantFailure) []int {
	result := make([]int, len(failures))
	for index, failure := range failures {
		result[index] = failure.Clause
	}
	return result
}

func fraInvariantFixture(t testing.TB, stable bool) (*AigManager, AigLit) {
	t.Helper()
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	state, _ := graph.CreateCI()
	graph.CreateCO(state)
	next := input
	if stable {
		next = state
	}
	graph.CreateCO(next)
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return graph, state
}

func TestFRAInvariantBaseInductionAndModels(t *testing.T) {
	unstable, state := fraInvariantFixture(t, false)
	for _, clauses := range [][][]AigLit{
		{{state.Not()}},
		{{state}},
		{{state.Not()}, {state, state.Not()}},
	} {
		want := fraInvariantExhaustive(t, unstable, clauses)
		result, err := VerifyFRAInvariant(unstable, 1, clauses, FRAInvariantOptions{})
		if err != nil {
			t.Fatal(clauses, err)
		}
		if result.InductionPremiseConsistent != want.premise || !slices.Equal(fraInvariantFailureIndices(result.BaseFailures), want.base) || !slices.Equal(fraInvariantFailureIndices(result.InductiveFailures), want.induction) || result.Verified != (len(want.base) == 0 && len(want.induction) == 0) {
			t.Fatalf("clauses=%v got=%+v want=%+v", clauses, result, want)
		}
		for _, failure := range append(slices.Clone(result.BaseFailures), result.InductiveFailures...) {
			if failure.Status != SatSatisfiable || len(failure.Model) == 0 {
				t.Fatalf("failure lacks replayed SAT model: %+v", failure)
			}
		}
	}

	stableGraph, stableState := fraInvariantFixture(t, true)
	verified, err := VerifyFRAInvariant(stableGraph, 1, [][]AigLit{{stableState.Not()}}, FRAInvariantOptions{})
	if err != nil || !verified.Verified || len(verified.BaseFailures) != 0 || len(verified.InductiveFailures) != 0 {
		t.Fatalf("verified=%+v err=%v", verified, err)
	}
	inconsistent, err := VerifyFRAInvariant(stableGraph, 1, [][]AigLit{{}}, FRAInvariantOptions{})
	if err != nil || inconsistent.Verified || inconsistent.InductionPremiseConsistent || len(inconsistent.BaseFailures) != 1 {
		t.Fatalf("inconsistent=%+v err=%v", inconsistent, err)
	}

	bounded := NewAigManager()
	for index := 0; index < 4; index++ {
		bounded.CreateCI()
	}
	boundedState, _ := bounded.CreateCI()
	bounded.CreateCO(boundedState)
	bounded.CreateCO(boundedState)
	bounded.SetRegisterCount(1)
	unknown, err := VerifyFRAInvariant(bounded, 1, [][]AigLit{{boundedState}}, FRAInvariantOptions{SAT: SatLimits{Decisions: 1}})
	if err != nil || len(unknown.BaseFailures) != 1 || unknown.BaseFailures[0].Status != SatUnknown || unknown.BaseFailures[0].Reason != SatDecisionLimit {
		t.Fatalf("unknown=%+v err=%v", unknown, err)
	}

	owned, err := VerifyFRAInvariant(unstable, 1, [][]AigLit{{state.Not()}}, FRAInvariantOptions{})
	if err != nil || len(owned.InductiveFailures) != 1 {
		t.Fatal(err, owned)
	}
	owned.InductiveFailures[0].Model[0] = !owned.InductiveFailures[0].Model[0]
	again, againErr := VerifyFRAInvariant(unstable, 1, [][]AigLit{{state.Not()}}, FRAInvariantOptions{})
	if againErr != nil || reflect.DeepEqual(owned, again) {
		t.Fatal("failure models do not own their state")
	}
}

func TestFRAInvariantLimitsDepthAndConcurrency(t *testing.T) {
	graph, state := fraInvariantFixture(t, false)
	clauses := [][]AigLit{{state.Not()}}
	if result, err := VerifyFRAInvariant(nil, 1, clauses, FRAInvariantOptions{}); result.VariablesPerFrame != 0 || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil result=%+v err=%v", result, err)
	}
	if result, err := VerifyFRAInvariant(graph, 2, clauses, FRAInvariantOptions{}); result.VariablesPerFrame != 0 || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("frames result=%+v err=%v", result, err)
	}
	for _, options := range []FRAInvariantOptions{{MaxObjects: -1}, {MaxVariables: -1}, {MaxClauses: -1}, {MaxLiterals: -1}, {MaxWork: -1}} {
		if result, err := VerifyFRAInvariant(graph, 1, clauses, options); result.VariablesPerFrame != 0 || !errors.Is(err, ErrDimension) {
			t.Fatalf("options=%+v result=%+v err=%v", options, result, err)
		}
	}
	for _, options := range []FRAInvariantOptions{{MaxObjects: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {MaxWork: 1}, {Stop: func() bool { return true }}, {SAT: SatLimits{Stop: func() bool { return true }}}} {
		if result, err := VerifyFRAInvariant(graph, 1, clauses, options); result.VariablesPerFrame != 0 || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatalf("options=%+v result=%+v err=%v", options, result, err)
		}
	}
	if result, err := VerifyFRAInvariant(graph, 1, [][]AigLit{{InvalidAigLit}}, FRAInvariantOptions{}); result.VariablesPerFrame != 0 || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatalf("literal result=%+v err=%v", result, err)
	}

	deep := NewAigManager()
	root, _ := deep.CreateCI()
	for index := 0; index < 10000; index++ {
		input, _ := deep.CreateCI()
		root, _ = deep.And(root.NotIf(index&1 != 0), input.NotIf(index&2 != 0))
	}
	stateInput, _ := deep.CreateCI()
	deep.CreateCO(root)
	deep.CreateCO(stateInput)
	if err := deep.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	deepResult, err := VerifyFRAInvariant(deep, 1, [][]AigLit{{root, root.Not()}}, FRAInvariantOptions{MaxObjects: 50000, MaxVariables: 100000, MaxClauses: 100000, MaxLiterals: 500000, MaxWork: 1000000})
	if err != nil || !deepResult.Verified {
		t.Fatalf("deep result=%+v err=%v", deepResult, err)
	}

	baseline, err := VerifyFRAInvariant(graph, 1, clauses, FRAInvariantOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, callErr := VerifyFRAInvariant(graph, 1, clauses, FRAInvariantOptions{})
			if callErr != nil || !reflect.DeepEqual(got, baseline) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzFRAInvariantVerification(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{0xff, 0, 17})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		at := func(index int) byte { return data[index%len(data)] }
		graph := NewAigManager()
		a, _ := graph.CreateCI()
		b, _ := graph.CreateCI()
		state, _ := graph.CreateCI()
		pool := []AigLit{AigConst1, AigConst0, a, b, state}
		for gate := 0; gate < 1+int(at(0)%4); gate++ {
			left := pool[int(at(2*gate+1))%len(pool)].NotIf(at(gate+1)&1 != 0)
			right := pool[int(at(2*gate+2))%len(pool)].NotIf(at(gate+2)&2 != 0)
			root, err := graph.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			pool = append(pool, root)
		}
		graph.CreateCO(pool[int(at(3))%len(pool)].NotIf(at(4)&4 != 0))
		graph.CreateCO(pool[int(at(5))%len(pool)].NotIf(at(6)&8 != 0))
		if err := graph.SetRegisterCount(1); err != nil {
			t.Fatal(err)
		}
		clauseCount := 1 + int(at(7)%4)
		clauses := make([][]AigLit, clauseCount)
		for clause := range clauses {
			width := 1 + int(at(8+clause)%3)
			for literal := 0; literal < width; literal++ {
				value := pool[int(at(9+3*clause+literal))%len(pool)].NotIf(at(10+clause+literal)&16 != 0)
				clauses[clause] = append(clauses[clause], value)
			}
		}
		want := fraInvariantExhaustive(t, graph, clauses)
		got, err := VerifyFRAInvariant(graph, 1, clauses, FRAInvariantOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.InductionPremiseConsistent != want.premise || !slices.Equal(fraInvariantFailureIndices(got.BaseFailures), want.base) || !slices.Equal(fraInvariantFailureIndices(got.InductiveFailures), want.induction) || got.Verified != (want.premise && len(want.base) == 0 && len(want.induction) == 0) {
			t.Fatalf("clauses=%v got=%+v want=%+v", clauses, got, want)
		}
	})
}
