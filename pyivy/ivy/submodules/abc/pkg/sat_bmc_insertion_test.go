package abc

// Stand-alone tests for insertion-based initial-state discovery from bmcInse.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func giaInsertionMachine(t testing.TB) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	graph.Name, graph.Spec = "insertion", "owned"
	input0, _ := graph.AppendCI()
	input1, _ := graph.AppendCI()
	register0, _ := graph.AppendCI()
	register1, _ := graph.AppendCI()
	next0, err := graph.AppendAnd(input0.Not(), register0)
	if err != nil {
		t.Fatal(err)
	}
	and, err := graph.AppendAnd(input1.Not(), register0.Not())
	if err != nil {
		t.Fatal(err)
	}
	graph.AppendCO(next0)
	graph.AppendCO(and.Not())
	if err = graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	_ = register1 // A deliberately unconstrained previous-state bit.
	return graph
}

func giaInsertionRandom() func() uint64 {
	z, w := uint32(3716960521), uint32(2174103536)
	random32 := func() uint32 {
		z = 36969*(z&65535) + (z >> 16)
		w = 18000*(w&65535) + (w >> 16)
		return (z << 16) + w
	}
	random32()
	return func() uint64 { return uint64(random32())<<32 | uint64(random32()) }
}

func giaInsertionToTernary(value GiaInsertionValue) GiaTernary {
	switch value {
	case GiaInsertionZero:
		return GiaTernaryZero
	case GiaInsertionOne:
		return GiaTernaryOne
	default:
		return GiaTernaryUnknown
	}
}

func giaInsertionReference(t testing.TB, graph *GiaManager, initial []bool, frames, words int) GiaInsertionResult {
	t.Helper()
	if initial == nil {
		initial = make([]bool, graph.RegisterCount())
	}
	target := make([]GiaInsertionValue, graph.RegisterCount())
	for index := range target {
		target[index] = GiaInsertionUnknown
	}
	randomWord := giaInsertionRandom()
	result := GiaInsertionResult{
		Inputs:           []GiaInsertionValue{},
		SelectedPatterns: []int{},
		UnknownCounts:    []int{},
	}
	for frame := 0; frame < frames; frame++ {
		piWords := make([]uint64, graph.PrimaryInputCount()*words)
		for input := 0; input < graph.PrimaryInputCount(); input++ {
			for word := 0; word < words; word++ {
				piWords[input*words+word] = randomWord()
			}
		}
		bestPattern, bestKnown := 0, -1
		bestState := make([]GiaInsertionValue, graph.RegisterCount())
		for sample := 0; sample < 64*words; sample++ {
			values := make([]GiaTernary, graph.InputCount())
			for input := 0; input < graph.PrimaryInputCount(); input++ {
				if piWords[input*words+sample/64]>>uint(sample&63)&1 != 0 {
					values[input] = GiaTernaryZero
				} else {
					values[input] = GiaTernaryOne
				}
			}
			for register := range target {
				values[graph.PrimaryInputCount()+register] = giaInsertionToTernary(target[register])
			}
			outputs, err := graph.EvaluateTernary(values, GiaCofactorOptions{})
			if err != nil {
				t.Fatal(err)
			}
			known := 0
			candidate := make([]GiaInsertionValue, graph.RegisterCount())
			for register, value := range outputs[graph.PrimaryOutputCount():] {
				switch value {
				case GiaTernaryZero:
					candidate[register], known = GiaInsertionZero, known+1
				case GiaTernaryOne:
					candidate[register], known = GiaInsertionOne, known+1
				default:
					candidate[register] = GiaInsertionUnknown
				}
			}
			if known > bestKnown {
				bestPattern, bestKnown = sample, known
				copy(bestState, candidate)
			}
		}
		target = bestState
		result.SelectedPatterns = append(result.SelectedPatterns, bestPattern)
		result.UnknownCounts = append(result.UnknownCounts, graph.RegisterCount()-bestKnown)
		for input := 0; input < graph.PrimaryInputCount(); input++ {
			value := GiaInsertionOne
			if piWords[input*words+bestPattern/64]>>uint(bestPattern&63)&1 != 0 {
				value = GiaInsertionZero
			}
			result.Inputs = append(result.Inputs, value)
		}
	}
	state := slices.Clone(initial)
	primary := make([]bool, graph.PrimaryInputCount())
	for frame := 0; frame < frames; frame++ {
		for input := range primary {
			primary[input] = result.Inputs[frame*len(primary)+input] == GiaInsertionOne
		}
		_, state, _ = graph.Step(primary, state)
	}
	result.Target = slices.Clone(target)
	result.Classes = make([]GiaInitialClass, graph.RegisterCount())
	for register, value := range state {
		result.Classes[register] = GiaInitialKnownZero
		if value {
			result.Classes[register] = GiaInitialKnownOne
		}
		if target[register] != GiaInsertionUnknown {
			result.Classes[register] += 4
		}
	}
	return result
}

func TestGiaInsertionInitialStateIndependentModel(t *testing.T) {
	graph := giaInsertionMachine(t)
	before := graph.Clone()
	initial := []bool{true, false}
	for frames := 0; frames <= 4; frames++ {
		for words := 1; words <= 2; words++ {
			got, err := graph.DiscoverInsertionInitialState(initial, GiaInsertionOptions{Frames: frames, Words: words})
			want := giaInsertionReference(t, graph, initial, frames, words)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatal(frames, words, got, want, err)
			}
			clone := got.Clone()
			if len(clone.Classes) != 0 {
				clone.Classes[0] ^= 1
				if slices.Equal(clone.Classes, got.Classes) {
					t.Fatal("clone aliases classes")
				}
			}
		}
	}
	if !reflect.DeepEqual(graph, before) {
		t.Fatal("source graph mutated")
	}
	zero, err := graph.DiscoverInsertionInitialState(nil, GiaInsertionOptions{Frames: 0, Words: 1})
	if err != nil || !slices.Equal(zero.Classes, []GiaInitialClass{GiaInitialKnownZero, GiaInitialKnownZero}) || !slices.Equal(zero.Target, []GiaInsertionValue{GiaInsertionUnknown, GiaInsertionUnknown}) {
		t.Fatal(zero, err)
	}
}

func TestGiaInsertionInitialStateFixedResults(t *testing.T) {
	graph := NewGiaManager()
	graph.AppendCI()
	graph.AppendCI()
	graph.AppendCO(GiaConst0)
	graph.AppendCO(GiaConst1)
	graph.SetRegisterCount(2)
	result, err := graph.DiscoverInsertionInitialState([]bool{true, false}, GiaInsertionOptions{Frames: 1, Words: 1})
	if err != nil || !slices.Equal(result.Classes, []GiaInitialClass{GiaInitialFixedZero, GiaInitialFixedOne}) || !slices.Equal(result.Target, []GiaInsertionValue{GiaInsertionZero, GiaInsertionOne}) || !slices.Equal(result.UnknownCounts, []int{0}) || !slices.Equal(result.SelectedPatterns, []int{0}) {
		t.Fatal(result, err)
	}

	withoutInputs := NewGiaManager()
	register, _ := withoutInputs.AppendCI()
	withoutInputs.AppendCO(register)
	withoutInputs.SetRegisterCount(1)
	result, err = withoutInputs.DiscoverInsertionInitialState(nil, GiaInsertionOptions{Frames: 2, Words: 1})
	if err != nil || len(result.Inputs) != 0 || !slices.Equal(result.Classes, []GiaInitialClass{GiaInitialKnownZero}) {
		t.Fatal(result, err)
	}
}

func TestGiaInsertionInitialStateErrorsLimitsCancellationAndDepth(t *testing.T) {
	graph := giaInsertionMachine(t)
	initial := []bool{false, false}
	tests := []struct {
		graph   *GiaManager
		initial []bool
		options GiaInsertionOptions
		want    error
	}{
		{nil, initial, GiaInsertionOptions{Words: 1}, ErrDimension},
		{graph, []bool{false}, GiaInsertionOptions{Words: 1}, ErrDimension},
		{graph, initial, GiaInsertionOptions{}, ErrDimension},
		{graph, initial, GiaInsertionOptions{Frames: -1, Words: 1}, ErrDimension},
		{graph, initial, GiaInsertionOptions{Words: 1, MaxFrames: -1}, ErrDimension},
		{graph, initial, GiaInsertionOptions{Words: 1, MaxWords: -1}, ErrDimension},
		{graph, initial, GiaInsertionOptions{Words: 1, MaxEntries: -1}, ErrDimension},
		{graph, initial, GiaInsertionOptions{Words: 1, MaxInputValues: -1}, ErrDimension},
		{graph, initial, GiaInsertionOptions{Words: 1, MaxSelectedPatterns: -1}, ErrDimension},
		{graph, initial, GiaInsertionOptions{Frames: 2, Words: 1, MaxFrames: 1}, ErrLimit},
		{graph, initial, GiaInsertionOptions{Words: 2, MaxWords: 1}, ErrLimit},
		{graph, initial, GiaInsertionOptions{Words: 1, MaxEntries: 1}, ErrLimit},
		{graph, initial, GiaInsertionOptions{Frames: 2, Words: 1, MaxInputValues: 1}, ErrLimit},
		{graph, initial, GiaInsertionOptions{Frames: 2, Words: 1, MaxSelectedPatterns: 1}, ErrLimit},
		{graph, initial, GiaInsertionOptions{Words: 1, Stop: func() bool { return true }}, ErrCancelled},
	}
	for index, test := range tests {
		result, err := test.graph.DiscoverInsertionInitialState(test.initial, test.options)
		if !reflect.DeepEqual(result, GiaInsertionResult{}) || !errors.Is(err, test.want) {
			t.Fatal(index, result, err, test.want)
		}
	}

	combinational := NewGiaManager()
	input, _ := combinational.AppendCI()
	combinational.AppendCO(input)
	if result, err := combinational.DiscoverInsertionInitialState(nil, GiaInsertionOptions{Words: 1}); !reflect.DeepEqual(result, GiaInsertionResult{}) || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	buffered := NewGiaManager()
	input, _ = buffered.AppendCI()
	register, _ := buffered.AppendCI()
	buffer, _ := buffered.AppendBuffer(input)
	buffered.AppendCO(buffer)
	buffered.AppendCO(register)
	buffered.SetRegisterCount(1)
	if result, err := buffered.DiscoverInsertionInitialState(nil, GiaInsertionOptions{Words: 1}); !reflect.DeepEqual(result, GiaInsertionResult{}) || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	deep := NewGiaManager()
	root, _ := deep.AppendCI()
	register, _ = deep.AppendCI()
	for index := 0; index < 10000; index++ {
		input, _ = deep.AppendCI()
		root, _ = deep.AppendAnd(root, input.NotIf(index&1 != 0))
	}
	deep.AppendCO(root)
	deep.SetRegisterCount(1)
	result, err := deep.DiscoverInsertionInitialState(nil, GiaInsertionOptions{Frames: 1, Words: 1, MaxEntries: 50000, MaxInputValues: 20000})
	if err != nil || len(result.Inputs) != deep.PrimaryInputCount() || len(result.Classes) != 1 {
		t.Fatal(len(result.Inputs), result.Classes, err)
	}
}

func FuzzGiaInsertionInitialState(f *testing.F) {
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
		for index := 0; index < len(data) && index < 12; index++ {
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
		initial := []bool{data[0]&16 != 0, data[len(data)-1]&32 != 0}
		frames := int(data[0]>>6) % 3
		got, err := graph.DiscoverInsertionInitialState(initial, GiaInsertionOptions{Frames: frames, Words: 1, MaxEntries: 10000, MaxInputValues: 100})
		if err != nil {
			t.Fatal(err)
		}
		want := giaInsertionReference(t, graph, initial, frames, 1)
		if !reflect.DeepEqual(got, want) {
			t.Fatal(got, want)
		}
	})
}
