package abc

// Stand-alone tests for timeframe construction from sat/bmc/bmcBmc.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func bmcFramesMachine(t testing.TB) *AigManager {
	t.Helper()
	graph := NewAigManager()
	pi0, _ := graph.CreateCI()
	pi1, _ := graph.CreateCI()
	ro0, _ := graph.CreateCI()
	ro1, _ := graph.CreateCI()
	property, _ := graph.And(pi0, ro0.Not())
	constraint, _ := graph.Or(pi1, ro1)
	next0, _ := graph.Or(pi0, ro1)
	next1, _ := graph.And(pi1.Not(), ro0)
	graph.CreateCO(property)
	graph.CreateCO(constraint)
	graph.CreateCO(next0)
	graph.CreateCO(next1)
	if err := graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	return graph
}

func bmcFramesIndependentTrace(t testing.TB, source *AigManager, frames int, primaryInputs []bool) []bool {
	t.Helper()
	piCount := source.PrimaryInputCount()
	poCount := source.PrimaryOutputCount()
	state := make([]bool, source.RegisterCount())
	result := make([]bool, 0, frames*poCount)
	for frame := 0; frame < frames; frame++ {
		assignment := append(slices.Clone(primaryInputs[frame*piCount:(frame+1)*piCount]), state...)
		outputs, err := source.Evaluate(assignment)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, outputs[:poCount]...)
		state = slices.Clone(outputs[poCount:])
	}
	return result
}

func TestBmcFramesSemanticsOwnershipAndConstraints(t *testing.T) {
	source := bmcFramesMachine(t)
	before := source.Clone()
	const frames = 3
	expansion, err := source.BuildBmcFrames(BmcFramesOptions{Frames: frames})
	if err != nil {
		t.Fatal(err)
	}
	graph := expansion.Graph()
	if expansion.FramesBuilt() != frames || expansion.ReachableANDs() == 0 || graph == source || graph.RegisterCount() != 0 || graph.ConstraintCount() != 0 || graph.InputCount() != frames*source.PrimaryInputCount() || graph.OutputCount() != frames*source.PrimaryOutputCount() || graph.Check() != nil {
		t.Fatal(expansion.FramesBuilt(), expansion.ReachableANDs(), graph.InputCount(), graph.OutputCount(), graph.Check())
	}
	if !reflect.DeepEqual(source, before) {
		t.Fatal("source mutated")
	}
	for bits := 0; bits < 1<<(frames*source.PrimaryInputCount()); bits++ {
		inputs := make([]bool, frames*source.PrimaryInputCount())
		for index := range inputs {
			inputs[index] = bits>>index&1 != 0
		}
		got, err := graph.Evaluate(inputs)
		want := bmcFramesIndependentTrace(t, source, frames, inputs)
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(bits, got, want, err)
		}
	}
}

func TestBmcFramesCompletedFrameThreshold(t *testing.T) {
	source := NewAigManager()
	pi, _ := source.CreateCI()
	ro, _ := source.CreateCI()
	root, _ := source.And(pi, ro.Not())
	source.CreateCO(root)
	source.CreateCO(root)
	source.CreateCO(pi)
	if err := source.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ threshold, wantFrames int }{{1, 2}, {2, 3}, {3, 4}} {
		threshold, wantFrames := test.threshold, test.wantFrames
		expansion, err := source.BuildBmcFrames(BmcFramesOptions{Frames: 6, MaxReachableANDs: threshold})
		if err != nil || expansion.FramesBuilt() != wantFrames || expansion.ReachableANDs() != threshold || expansion.Graph().OutputCount() != wantFrames*2 {
			t.Fatal(threshold, expansion, err)
		}
	}
	full, err := source.BuildBmcFrames(BmcFramesOptions{Frames: 6})
	if err != nil || full.FramesBuilt() != 6 || full.ReachableANDs() != 5 {
		t.Fatal(full, err)
	}
}

func TestBmcFramesErrorsLimitsCancellationAndDepth(t *testing.T) {
	source := bmcFramesMachine(t)
	for _, options := range []BmcFramesOptions{
		{}, {Frames: -1}, {Frames: 1, MaxReachableANDs: -1}, {Frames: 1, MaxObjects: -1},
		{Frames: 1, MaxMappingEntries: -1}, {Frames: 1, MaxObjects: 1},
		{Frames: 1, MaxMappingEntries: 1}, {Frames: 1, Stop: func() bool { return true }},
	} {
		result, err := source.BuildBmcFrames(options)
		if result != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := (*AigManager)(nil).BuildBmcFrames(BmcFramesOptions{Frames: 1}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	combinational := NewAigManager()
	input, _ := combinational.CreateCI()
	combinational.CreateCO(input)
	if result, err := combinational.BuildBmcFrames(BmcFramesOptions{Frames: 1}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	buffered := NewAigManager()
	input, _ = buffered.CreateCI()
	register, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	buffered.CreateCO(buffer)
	buffered.CreateCO(register)
	buffered.SetRegisterCount(1)
	if result, err := buffered.BuildBmcFrames(BmcFramesOptions{Frames: 1}); result != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	for trigger := 2; trigger < 25; trigger++ {
		calls := 0
		result, err := source.BuildBmcFrames(BmcFramesOptions{Frames: 4, Stop: func() bool {
			calls++
			return calls == trigger
		}})
		if result != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(trigger, calls, result, err)
		}
	}

	deep := NewAigManager()
	root, _ := deep.CreateCI()
	for index := 0; index < 10000; index++ {
		input, _ := deep.CreateCI()
		root, _ = deep.And(root, input.NotIf(index&1 != 0))
	}
	register, _ = deep.CreateCI()
	deep.CreateCO(root)
	deep.CreateCO(register)
	deep.SetRegisterCount(1)
	expansion, err := deep.BuildBmcFrames(BmcFramesOptions{Frames: 1, MaxObjects: 30000, MaxMappingEntries: 30000})
	if err != nil || expansion.Graph().AndCount() != 10000 || expansion.ReachableANDs() != 10000 || expansion.Graph().Check() != nil {
		t.Fatal(expansion, err)
	}
}

func FuzzBmcFrames(f *testing.F) {
	f.Add([]byte{1, 3, 5, 7, 9, 11, 13, 15})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		graph := NewAigManager()
		pi0, _ := graph.CreateCI()
		pi1, _ := graph.CreateCI()
		register, _ := graph.CreateCI()
		signals := []AigLit{pi0, pi1, register, AigConst0, AigConst1}
		for index := 0; index+1 < len(data) && len(signals) < 40; index += 2 {
			left := signals[int(data[index])%len(signals)].NotIf(data[index]&0x80 != 0)
			right := signals[int(data[index+1])%len(signals)].NotIf(data[index+1]&0x80 != 0)
			literal, err := graph.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			signals = append(signals, literal)
		}
		graph.CreateCO(signals[int(data[0])%len(signals)].NotIf(data[0]&1 != 0))
		graph.CreateCO(signals[int(data[len(data)-1])%len(signals)].NotIf(data[len(data)-1]&1 != 0))
		graph.SetRegisterCount(1)
		expansion, err := graph.BuildBmcFrames(BmcFramesOptions{Frames: 3})
		if err != nil {
			t.Fatal(err)
		}
		inputs := make([]bool, 6)
		for index := range inputs {
			inputs[index] = data[index%len(data)]>>uint(index&7)&1 != 0
		}
		got, err := expansion.Graph().Evaluate(inputs)
		want := bmcFramesIndependentTrace(t, graph, 3, inputs)
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, err)
		}
	})
}
