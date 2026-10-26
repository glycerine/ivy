package abc

// Go-only tests for src/base/abci/abcDarUnfold2.c.

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

func darUnfold2SequentialFixture(t testing.TB) (*AigManager, AigLit, AigLit, AigLit) {
	t.Helper()
	graph := NewAigManager()
	graph.Name, graph.Spec = "unfold", "fixture.aig"
	input, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	state, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	next, err := graph.Xor(input, state)
	if err != nil {
		t.Fatal(err)
	}
	ordinary, err := graph.And(input, state)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(ordinary); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(input); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(next); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	if err = graph.Check(); err != nil {
		t.Fatal(err)
	}
	return graph, input, state, next
}

func TestDarUnfold2ClassifiedConstraints(t *testing.T) {
	source, input, state, next := darUnfold2SequentialFixture(t)
	before := source.Clone()
	result, err := DarUnfold2(source, DarUnfold2Options{TypeI: []AigLit{state.Not()}, TypeII: []AigLit{next}})
	if err != nil {
		t.Fatal(err)
	}
	graph := result.Graph
	if result.TypeICount != 1 || result.TypeIICount != 1 || graph.Name != source.Name || graph.Spec != source.Spec {
		t.Fatal(result.TypeICount, result.TypeIICount, graph.Name, graph.Spec)
	}
	if graph.PrimaryInputCount() != 1 || graph.RegisterCount() != 2 || graph.PrimaryOutputCount() != 4 || graph.ConstraintCount() != 3 {
		t.Fatal(graph.PrimaryInputCount(), graph.RegisterCount(), graph.PrimaryOutputCount(), graph.ConstraintCount())
	}
	for mask := 0; mask < 8; mask++ {
		primary := []bool{mask&1 != 0}
		oldState := mask&2 != 0
		newState := mask&4 != 0
		outputs, nextState, stepErr := graph.Step(primary, []bool{oldState, newState})
		if stepErr != nil {
			t.Fatal(stepErr)
		}
		nextValue := primary[0] != oldState
		wantOutputs := []bool{primary[0] && oldState, primary[0], !oldState, newState && nextValue}
		wantState := []bool{nextValue, nextValue}
		if !reflect.DeepEqual(outputs, wantOutputs) || !reflect.DeepEqual(nextState, wantState) {
			t.Fatal(mask, outputs, wantOutputs, nextState, wantState)
		}
	}
	if !reflect.DeepEqual(source, before) || input.ID() == 0 {
		t.Fatal("source mutated")
	}
}

func TestDarFold2SequentialStickyViolation(t *testing.T) {
	source, _, _, _ := darUnfold2SequentialFixture(t)
	before := source.Clone()
	folded, err := DarFold2(source, DarFold2Options{})
	if err != nil {
		t.Fatal(err)
	}
	if folded.Name != source.Name || folded.Spec != source.Spec || folded.PrimaryInputCount() != 1 || folded.PrimaryOutputCount() != 1 || folded.ConstraintCount() != 0 || folded.RegisterCount() != 2 {
		t.Fatal(folded.Name, folded.Spec, folded.PrimaryInputCount(), folded.PrimaryOutputCount(), folded.ConstraintCount(), folded.RegisterCount())
	}
	for mask := 0; mask < 8; mask++ {
		input, oldState, sticky := mask&1 != 0, mask&2 != 0, mask&4 != 0
		outputs, nextState, stepErr := folded.Step([]bool{input}, []bool{oldState, sticky})
		if stepErr != nil {
			t.Fatal(stepErr)
		}
		violation := input
		stickyNext := violation || sticky
		wantOutput := input && oldState && !stickyNext
		wantOldNext := oldState
		if !stickyNext {
			wantOldNext = input != oldState
		}
		if !reflect.DeepEqual(outputs, []bool{wantOutput}) || !reflect.DeepEqual(nextState, []bool{wantOldNext, stickyNext}) {
			t.Fatal(mask, outputs, nextState, wantOutput, wantOldNext, stickyNext)
		}
	}
	if !reflect.DeepEqual(source, before) {
		t.Fatal("source mutated")
	}
}

func TestDarFold2CombinationalPolarityAndClone(t *testing.T) {
	for _, complement := range []bool{false, true} {
		graph := NewAigManager()
		a, _ := graph.CreateCI()
		b, _ := graph.CreateCI()
		ordinary, _ := graph.Xor(a, b)
		graph.CreateCO(ordinary)
		graph.CreateCO(b)
		graph.SetConstraintCount(1)
		folded, err := DarFold2(graph, DarFold2Options{Complement: complement})
		if err != nil {
			t.Fatal(err)
		}
		if folded.RegisterCount() != 0 || folded.PrimaryOutputCount() != 1 || folded.ConstraintCount() != 0 {
			t.Fatal(folded.RegisterCount(), folded.PrimaryOutputCount(), folded.ConstraintCount())
		}
		for mask := 0; mask < 4; mask++ {
			av, bv := mask&1 != 0, mask&2 != 0
			outputs, _, stepErr := folded.Step([]bool{av, bv}, nil)
			if stepErr != nil {
				t.Fatal(stepErr)
			}
			violation := bv
			if complement {
				violation = !violation
			}
			want := (av != bv) && !violation
			if len(outputs) != 1 || outputs[0] != want {
				t.Fatal(complement, mask, outputs, want)
			}
		}
	}
	plain := NewAigManager()
	a, _ := plain.CreateCI()
	plain.CreateCO(a)
	clone, err := DarFold2(plain, DarFold2Options{})
	if err != nil || clone == plain || !reflect.DeepEqual(clone, plain) {
		t.Fatal(err, clone == plain, reflect.DeepEqual(clone, plain))
	}
}

func TestDarUnfold2ValidationLimitsAndCancellation(t *testing.T) {
	source, input, _, _ := darUnfold2SequentialFixture(t)
	cases := []struct {
		unfold DarUnfold2Options
		want   error
	}{
		{DarUnfold2Options{MaxObjects: -1}, ErrDimension},
		{DarUnfold2Options{MaxObjects: 1}, ErrLimit},
		{DarUnfold2Options{Stop: func() bool { return true }}, ErrCancelled},
		{DarUnfold2Options{TypeI: []AigLit{AigConst1}}, ErrInvalidLiteral},
		{DarUnfold2Options{TypeI: []AigLit{input}, TypeII: []AigLit{input.Not()}}, ErrInvariant},
		{DarUnfold2Options{TypeI: []AigLit{InvalidAigLit}}, ErrInvalidLiteral},
	}
	for _, tc := range cases {
		result, err := DarUnfold2(source, tc.unfold)
		if result.Graph != nil || !errors.Is(err, tc.want) {
			t.Fatal(tc.unfold.MaxObjects, result.Graph, err, tc.want)
		}
	}
	if result, err := DarUnfold2(nil, DarUnfold2Options{}); result.Graph != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result.Graph, err)
	}
	for _, tc := range []struct {
		options DarFold2Options
		want    error
	}{
		{DarFold2Options{MaxObjects: -1}, ErrDimension},
		{DarFold2Options{MaxObjects: 1}, ErrLimit},
		{DarFold2Options{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := DarFold2(source, tc.options); result != nil || !errors.Is(err, tc.want) {
			t.Fatal(result, err, tc.want)
		}
	}
	constraintsOnly := NewAigManager()
	a, _ := constraintsOnly.CreateCI()
	constraintsOnly.CreateCO(a)
	constraintsOnly.SetConstraintCount(1)
	if result, err := DarFold2(constraintsOnly, DarFold2Options{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	plain := NewAigManager()
	plainInput, _ := plain.CreateCI()
	plain.CreateCO(plainInput)
	calls := 0
	result, err := DarFold2(plain, DarFold2Options{Stop: func() bool {
		calls++
		return calls > 4
	}})
	if result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal("late cancellation while copying outputs", calls, result, err)
	}
}

func TestDarFold2IndependentModelCampaign(t *testing.T) {
	rng := rand.New(rand.NewSource(0xf01d2))
	for execution := 0; execution < 5000; execution++ {
		graph := NewAigManager()
		a, _ := graph.CreateCI()
		b, _ := graph.CreateCI()
		state, _ := graph.CreateCI()
		ordinaryKind := rng.Intn(4)
		constraintKind := rng.Intn(4)
		nextKind := rng.Intn(4)
		ordinary := darFold2ModelLiteral(graph, a, b, ordinaryKind)
		constraint := darFold2ModelLiteral(graph, b, state, constraintKind)
		next := darFold2ModelLiteral(graph, a, state, nextKind)
		graph.CreateCO(ordinary)
		graph.CreateCO(constraint)
		graph.CreateCO(next)
		graph.SetRegisterCount(1)
		graph.SetConstraintCount(1)
		complement := rng.Intn(2) != 0
		folded, err := DarFold2(graph, DarFold2Options{Complement: complement})
		if err != nil {
			t.Fatal(execution, err)
		}
		for mask := 0; mask < 16; mask++ {
			av, bv := mask&1 != 0, mask&2 != 0
			oldState, sticky := mask&4 != 0, mask&8 != 0
			outputs, nextState, stepErr := folded.Step([]bool{av, bv}, []bool{oldState, sticky})
			if stepErr != nil {
				t.Fatal(execution, mask, stepErr)
			}
			ordinaryValue := darFold2ModelValue(av, bv, ordinaryKind)
			constraintValue := darFold2ModelValue(bv, oldState, constraintKind)
			violation := constraintValue != complement
			stickyNext := sticky || violation
			oldNext := oldState
			if !stickyNext {
				oldNext = darFold2ModelValue(av, oldState, nextKind)
			}
			if len(outputs) != 1 || outputs[0] != (ordinaryValue && !stickyNext) ||
				len(nextState) != 2 || nextState[0] != oldNext || nextState[1] != stickyNext {
				t.Fatal(execution, mask, outputs, nextState, ordinaryValue, oldNext, stickyNext)
			}
		}
	}
}

func darFold2ModelLiteral(graph *AigManager, left, right AigLit, kind int) AigLit {
	switch kind {
	case 0:
		result, _ := graph.And(left, right)
		return result
	case 1:
		result, _ := graph.Xor(left, right)
		return result
	case 2:
		result, _ := graph.And(left, right)
		return result.Not()
	default:
		result, _ := graph.Xor(left, right)
		return result.Not()
	}
}

func darFold2ModelValue(left, right bool, kind int) bool {
	switch kind {
	case 0:
		return left && right
	case 1:
		return left != right
	case 2:
		return !(left && right)
	default:
		return left == right
	}
}

func TestDarUnfold2IndependentModelCampaign(t *testing.T) {
	rng := rand.New(rand.NewSource(0xabcda2))
	for execution := 0; execution < 5000; execution++ {
		graph := NewAigManager()
		a, _ := graph.CreateCI()
		b, _ := graph.CreateCI()
		ordinaryFlip := rng.Intn(2) != 0
		ordinary := a
		if ordinaryFlip {
			ordinary = ordinary.Not()
		}
		typeIFlip := rng.Intn(2) != 0
		candidateI := b.NotIf(typeIFlip)
		candidateII, _ := graph.Xor(a, b)
		typeIIFlip := rng.Intn(2) != 0
		candidateII = candidateII.NotIf(typeIIFlip)
		graph.CreateCO(ordinary)
		result, err := DarUnfold2(graph, DarUnfold2Options{TypeI: []AigLit{candidateI}, TypeII: []AigLit{candidateII}})
		if err != nil {
			t.Fatal(execution, err)
		}
		for mask := 0; mask < 8; mask++ {
			av, bv, latch := mask&1 != 0, mask&2 != 0, mask&4 != 0
			outputs, state, stepErr := result.Graph.Step([]bool{av, bv}, []bool{latch})
			if stepErr != nil {
				t.Fatal(execution, mask, stepErr)
			}
			ordinaryValue := av != ordinaryFlip
			typeIValue := bv != typeIFlip
			xorValue := av != bv
			typeIIValue := xorValue != typeIIFlip
			want := []bool{ordinaryValue, typeIValue, latch && typeIIValue}
			if !reflect.DeepEqual(outputs, want) || !reflect.DeepEqual(state, []bool{typeIIValue}) {
				t.Fatal(execution, mask, outputs, want, state, typeIIValue)
			}
		}
	}
}

func FuzzDarUnfold2Semantics(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x55, 0xaa, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		graph := NewAigManager()
		a, _ := graph.CreateCI()
		b, _ := graph.CreateCI()
		candidate, _ := graph.Xor(a, b)
		candidate = candidate.NotIf(seed&1 != 0)
		graph.CreateCO(a.NotIf(seed&2 != 0))
		result, err := DarUnfold2(graph, DarUnfold2Options{TypeII: []AigLit{candidate}})
		if err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 8; mask++ {
			av, bv, state := mask&1 != 0, mask&2 != 0, mask&4 != 0
			outputs, next, stepErr := result.Graph.Step([]bool{av, bv}, []bool{state})
			if stepErr != nil {
				t.Fatal(stepErr)
			}
			cv := (av != bv) != (seed&1 != 0)
			if len(outputs) != 2 || outputs[0] != (av != (seed&2 != 0)) || outputs[1] != (state && cv) || len(next) != 1 || next[0] != cv {
				t.Fatal(seed, mask, outputs, next, cv)
			}
		}
	})
}
