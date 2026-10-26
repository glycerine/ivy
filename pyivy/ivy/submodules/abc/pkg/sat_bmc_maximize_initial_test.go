package abc

// Stand-alone tests for initial-state maximization from bmcMaxi.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func giaMaxTransitionMachine(t testing.TB) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	graph.Name, graph.Spec = "maxi", "source-spec"
	input, _ := graph.AppendCI()
	register0, _ := graph.AppendCI()
	register1, _ := graph.AppendCI()
	next0, err := graph.AppendAnd(input.Not(), register0)
	if err != nil {
		t.Fatal(err)
	}
	and, err := graph.AppendAnd(input.Not(), register1.Not())
	if err != nil {
		t.Fatal(err)
	}
	next1 := and.Not()
	graph.AppendCO(next0)
	graph.AppendCO(next1)
	if err = graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	return graph
}

func giaMaxInitialValue(class GiaInitialClass, useVariables, control, data bool) bool {
	switch class {
	case GiaInitialKnownZero:
		return useVariables && control && data
	case GiaInitialKnownOne:
		return !useVariables || !control || data
	case GiaInitialRelaxedZero:
		return control && data
	case GiaInitialRelaxedOne:
		return !control || data
	case GiaInitialFixedZero:
		return false
	case GiaInitialFixedOne:
		return true
	default:
		panic("invalid initial class")
	}
}

func TestGiaMaxUnfoldAllInitialClasses(t *testing.T) {
	source := giaMaxTransitionMachine(t)
	before := source.Clone()
	classes := []GiaInitialClass{GiaInitialKnownZero, GiaInitialKnownOne, GiaInitialRelaxedZero, GiaInitialRelaxedOne, GiaInitialFixedZero, GiaInitialFixedOne}
	for _, first := range classes {
		for _, second := range classes {
			for _, useVariables := range []bool{false, true} {
				initial := []GiaInitialClass{first, second}
				unfolding, err := source.UnfoldMaxInitial(initial, GiaMaxUnfoldOptions{Frames: 2, UseVariables: useVariables})
				if err != nil {
					t.Fatal(initial, useVariables, err)
				}
				graph := unfolding.Graph()
				if graph.Name != source.Name || graph.Spec != "" || graph.InputCount() != 6 || graph.OutputCount() != 2 || graph.RegisterCount() != 0 || graph.Check() != nil {
					t.Fatal(initial, useVariables, graph.Name, graph.Spec, graph.InputCount(), graph.OutputCount(), graph.Check())
				}
				controls, data := unfolding.Controls(), unfolding.Data()
				if len(controls) != 2 || len(data) != 2 {
					t.Fatal(len(controls), len(data))
				}
				controls[0] = InvalidGiaLit
				data[0] = InvalidGiaLit
				if unfolding.Controls()[0] == InvalidGiaLit || unfolding.Data()[0] == InvalidGiaLit {
					t.Fatal("accessor aliases")
				}
				for assignment := 0; assignment < 64; assignment++ {
					values := make([]bool, 6)
					for index := range values {
						values[index] = assignment>>index&1 != 0
					}
					state := []bool{
						giaMaxInitialValue(first, useVariables, values[0], values[2]),
						giaMaxInitialValue(second, useVariables, values[1], values[3]),
					}
					for frame := 0; frame < 2; frame++ {
						_, state, err = source.Step([]bool{values[4+frame]}, state)
						if err != nil {
							t.Fatal(err)
						}
					}
					got, evalErr := graph.Evaluate(values)
					if evalErr != nil || !slices.Equal(got, state) {
						t.Fatal(initial, useVariables, assignment, got, state, evalErr)
					}
				}
			}
		}
	}
	if !reflect.DeepEqual(source, before) {
		t.Fatal("source mutated")
	}
	unfolding, _ := source.UnfoldMaxInitial([]GiaInitialClass{0, 0}, GiaMaxUnfoldOptions{Frames: 2})
	if row, err := unfolding.FrameInputs(0); err != nil || len(row) != 1 {
		t.Fatal(row, err)
	}
	for _, frame := range []int{-1, 2} {
		if row, err := unfolding.FrameInputs(frame); row != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(frame, row, err)
		}
	}
}

func giaMaxCoreMachine(t testing.TB) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	register0, _ := graph.AppendCI()
	graph.AppendCI()
	graph.AppendCO(register0)
	graph.AppendCO(GiaConst0)
	if err := graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	return graph
}

func giaMaxVerifyGeneralization(t testing.TB, source *GiaManager, original, generalized []GiaInitialClass, frames int) {
	t.Helper()
	baseline, err := source.UnfoldMaxInitial(original, GiaMaxUnfoldOptions{Frames: frames})
	if err != nil {
		t.Fatal(err)
	}
	result, err := source.UnfoldMaxInitial(generalized, GiaMaxUnfoldOptions{Frames: frames})
	if err != nil {
		t.Fatal(err)
	}
	if baseline.Graph().InputCount() != result.Graph().InputCount() || baseline.Graph().InputCount() > 14 {
		t.Fatal("generalization dimensions", baseline.Graph().InputCount(), result.Graph().InputCount())
	}
	inputs := make([]bool, baseline.Graph().InputCount())
	for assignment := 0; assignment < 1<<len(inputs); assignment++ {
		for index := range inputs {
			inputs[index] = assignment>>index&1 != 0
		}
		want, firstErr := baseline.Graph().Evaluate(inputs)
		got, secondErr := result.Graph().Evaluate(inputs)
		if firstErr != nil || secondErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, firstErr, secondErr, original, generalized)
		}
	}
}

func TestGiaMaximizeInitialStateCoresAndSoundness(t *testing.T) {
	graph := giaMaxCoreMachine(t)
	for _, test := range []struct {
		initial, want []GiaInitialClass
		core          []int
	}{
		{[]GiaInitialClass{GiaInitialKnownZero, GiaInitialKnownZero}, []GiaInitialClass{GiaInitialKnownZero, GiaInitialRelaxedZero}, []int{0}},
		{[]GiaInitialClass{GiaInitialKnownOne, GiaInitialKnownOne}, []GiaInitialClass{GiaInitialKnownOne, GiaInitialRelaxedOne}, []int{0}},
		{[]GiaInitialClass{GiaInitialFixedZero, GiaInitialKnownZero}, []GiaInitialClass{GiaInitialFixedZero, GiaInitialRelaxedZero}, nil},
		{[]GiaInitialClass{GiaInitialRelaxedOne, GiaInitialFixedOne}, []GiaInitialClass{GiaInitialRelaxedOne, GiaInitialFixedOne}, nil},
	} {
		result, err := graph.MaximizeInitialState(test.initial, GiaMaxInitialOptions{Frames: 2})
		if err != nil || result.Status != SatUnsatisfiable || !result.FixedPoint || !slices.Equal(result.Classes, test.want) || !slices.Equal(result.CoreRegisters, test.core) {
			t.Fatal(test, result, err)
		}
		giaMaxVerifyGeneralization(t, graph, test.initial, result.Classes, 2)
		result.Classes[0] = GiaInitialFixedOne
		again, againErr := graph.MaximizeInitialState(test.initial, GiaMaxInitialOptions{Frames: 2})
		if againErr != nil || again.Classes[0] != test.want[0] {
			t.Fatal("result aliases", again, againErr)
		}
	}
	nilInitial, err := graph.MaximizeInitialState(nil, GiaMaxInitialOptions{Frames: 1})
	if err != nil || !slices.Equal(nilInitial.Classes, []GiaInitialClass{GiaInitialKnownZero, GiaInitialRelaxedZero}) {
		t.Fatal(nilInitial, err)
	}
}

func TestGiaMaxInitialErrorsLimitsCancellationAndDepth(t *testing.T) {
	graph := giaMaxCoreMachine(t)
	initial := []GiaInitialClass{0, 0}
	for _, options := range []GiaMaxUnfoldOptions{
		{}, {Frames: -1}, {Frames: 1, MaxObjects: -1}, {Frames: 1, MaxMappingEntries: -1},
		{Frames: 1, MaxObjects: 1}, {Frames: 1, MaxMappingEntries: 1},
		{Frames: 1, Stop: func() bool { return true }},
	} {
		if result, err := graph.UnfoldMaxInitial(initial, options); result != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	for _, bad := range [][]GiaInitialClass{{0}, {0, 0, 0}, {0, 9}} {
		if result, err := graph.UnfoldMaxInitial(bad, GiaMaxUnfoldOptions{Frames: 1}); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(bad, result, err)
		}
	}
	if result, err := (*GiaManager)(nil).UnfoldMaxInitial(initial, GiaMaxUnfoldOptions{Frames: 1}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	combinational := NewGiaManager()
	input, _ := combinational.AppendCI()
	combinational.AppendCO(input)
	if result, err := combinational.UnfoldMaxInitial(nil, GiaMaxUnfoldOptions{Frames: 1}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	buffered := NewGiaManager()
	input, _ = buffered.AppendCI()
	register, _ := buffered.AppendCI()
	buffer, _ := buffered.AppendBuffer(input)
	buffered.AppendCO(buffer)
	buffered.AppendCO(register)
	buffered.SetRegisterCount(1)
	if result, err := buffered.UnfoldMaxInitial([]GiaInitialClass{0}, GiaMaxUnfoldOptions{Frames: 1}); result != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	for _, options := range []GiaMaxInitialOptions{
		{}, {Frames: -1}, {Frames: 1, MaxObjects: -1}, {Frames: 1, MaxMappingEntries: -1},
		{Frames: 1, MaxVariables: -1}, {Frames: 1, MaxClauses: -1}, {Frames: 1, MaxLiterals: -1},
		{Frames: 1, MaxIterations: -1}, {Frames: 1, MaxVariables: 1}, {Frames: 1, MaxClauses: 1},
		{Frames: 1, MaxLiterals: 1}, {Frames: 1, MaxIterations: 1}, {Frames: 1, Stop: func() bool { return true }},
	} {
		if result, err := graph.MaximizeInitialState(initial, options); result.Classes != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	unknown, err := graph.MaximizeInitialState(initial, GiaMaxInitialOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || unknown.Status != SatUnknown || !slices.Equal(unknown.Classes, initial) || !slices.Equal(unknown.CoreRegisters, []int{0, 1}) {
		t.Fatal(unknown, err)
	}

	deep := NewGiaManager()
	root, _ := deep.AppendCI()
	for index := 0; index < 10000; index++ {
		input, _ := deep.AppendCI()
		root, _ = deep.AppendAnd(root, input.NotIf(index&1 != 0))
	}
	_, _ = deep.AppendCI()
	deep.AppendCO(root)
	deep.SetRegisterCount(1)
	unfolding, err := deep.UnfoldMaxInitial([]GiaInitialClass{0}, GiaMaxUnfoldOptions{Frames: 1, MaxObjects: 30000, MaxMappingEntries: 30000})
	if err != nil || unfolding.Graph().AndCount() != 10000 || unfolding.Graph().Check() != nil {
		t.Fatal(unfolding, err)
	}
}

func FuzzGiaMaximizeInitialState(f *testing.F) {
	f.Add([]byte{1, 3, 5, 7, 9, 11})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		graph := NewGiaManager()
		input0, _ := graph.AppendCI()
		input1, _ := graph.AppendCI()
		register0, _ := graph.AppendCI()
		register1, _ := graph.AppendCI()
		signals := []GiaLit{input0, input1, register0, register1, GiaConst0, GiaConst1}
		for index := 0; index < len(data) && len(signals) < 24; index++ {
			leftIndex := int(data[index]) % len(signals)
			rightIndex := (leftIndex + 1 + int(data[index]>>4)) % len(signals)
			if signals[leftIndex].ID() == signals[rightIndex].ID() {
				continue
			}
			left := signals[leftIndex].NotIf(data[index]&1 != 0)
			right := signals[rightIndex].NotIf(data[index]&2 != 0)
			if literal, err := graph.AppendAnd(left, right); err == nil {
				signals = append(signals, literal)
			}
		}
		graph.AppendCO(signals[int(data[0])%len(signals)].NotIf(data[0]&4 != 0))
		graph.AppendCO(signals[int(data[len(data)-1])%len(signals)].NotIf(data[len(data)-1]&8 != 0))
		graph.SetRegisterCount(2)
		initial := []GiaInitialClass{GiaInitialClass(data[0] & 1), GiaInitialClass(data[len(data)-1] & 1)}
		frames := 1 + int(data[0]>>1&1)
		result, err := graph.MaximizeInitialState(initial, GiaMaxInitialOptions{Frames: frames, MaxObjects: 10000, MaxMappingEntries: 10000, MaxVariables: 10000, MaxClauses: 40000, MaxLiterals: 120000})
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != SatUnsatisfiable || !result.FixedPoint {
			t.Fatal(result)
		}
		giaMaxVerifyGeneralization(t, graph, initial, result.Classes, frames)
	})
}
