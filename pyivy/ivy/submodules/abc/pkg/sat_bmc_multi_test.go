package abc

// Go-only model, ownership and fuzz tests for src/sat/bmc/bmcMulti.c.

import (
	"errors"
	"slices"
	"testing"
)

func multiProveTestModel(t testing.TB) *AigerModel {
	t.Helper()
	graph := NewAigManager()
	input, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	latch, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	for _, driver := range []AigLit{input, AigConst0, latch, AigConst1} {
		if _, err = graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	model, err := NewAigerModel(graph, 3, 0, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func TestProcessMultiProveOutputsOwnershipAndAtomicErrors(t *testing.T) {
	first, _ := NewCounterexample(0, 1, 1)
	first.Property = 0
	second, _ := NewCounterexample(0, 1, 1)
	second.Property = 1
	input := []*Counterexample{nil, first, nil, second}
	output := make([]*Counterexample, 6)
	left, compacted, err := ProcessMultiProveOutputs(input, output, []int{4, 1, 5, 0})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(left, []int{0, 2}) || !slices.Equal(compacted, []int{4, 5}) {
		t.Fatal(left, compacted)
	}
	if output[1] != first || output[0] != second || input[1] != nil || input[3] != nil || first.Property != 1 || second.Property != 0 {
		t.Fatal("counterexample ownership transfer")
	}
	badIn := []*Counterexample{first, nil}
	badOut := []*Counterexample{nil, nil}
	if _, _, err = ProcessMultiProveOutputs(badIn, badOut, []int{0, 0}); !errors.Is(err, ErrDimension) || badIn[0] != first || badOut[0] != nil {
		t.Fatal("duplicate map must fail atomically", err)
	}
	if _, _, err = ProcessMultiProveOutputs(nil, nil, []int{0}); !errors.Is(err, ErrDimension) {
		t.Fatal("dimension error", err)
	}
}

func TestMultiProveConstantCountsAndReport(t *testing.T) {
	model := multiProveTestModel(t)
	aigCount, err := CountAigConstantZeroPrimaryOutputs(model.graph)
	if err != nil || aigCount != 1 {
		t.Fatal(aigCount, err)
	}
	gia, _, err := GiaFromAig(model.graph)
	if err != nil {
		t.Fatal(err)
	}
	giaCount, err := CountGiaConstantZeroPrimaryOutputs(gia)
	if err != nil || giaCount != 1 {
		t.Fatal(giaCount, err)
	}
	report, err := NewMultiProveReport(model.graph, "BMC", 3, 5, 2*model.graph.ObjectCount())
	if err != nil {
		t.Fatal(err)
	}
	if report.Stage != "BMC" || report.Iteration != 3 || report.PrimaryInputs != 1 || report.PrimaryOutputs != 3 || report.Registers != 1 || report.Solved != 2 || report.SolvedPercent != 40 || report.SizePercent != 50 {
		t.Fatal(report)
	}
	if _, err = CountAigConstantZeroPrimaryOutputs(nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = CountGiaConstantZeroPrimaryOutputs(nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestMultiProveSynthesisPreservesSequentialSemantics(t *testing.T) {
	model := multiProveTestModel(t)
	originalObjects := model.graph.ObjectCount()
	// Add an unreachable cone so cleanup has observable work to do.
	a, _ := model.graph.IthVar(0)
	b, _ := model.graph.IthVar(1)
	if _, err := model.graph.And(a, b); err != nil {
		t.Fatal(err)
	}
	result, err := model.graph.MultiProveSynthesize(GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.ObjectCount() > originalObjects || result.RegisterCount() != model.graph.RegisterCount() || result.PrimaryOutputCount() != model.graph.PrimaryOutputCount() {
		t.Fatal("synthesis shape", originalObjects, result.ObjectCount())
	}
	for initial := 0; initial < 2; initial++ {
		for input := 0; input < 2; input++ {
			state := []bool{initial != 0}
			primary := []bool{input != 0}
			wantOutputs, wantNext, e := model.graph.Step(primary, state)
			if e != nil {
				t.Fatal(e)
			}
			gotOutputs, gotNext, e := result.Step(primary, state)
			if e != nil || !slices.Equal(gotOutputs, wantOutputs) || !slices.Equal(gotNext, wantNext) {
				t.Fatal(initial, input, gotOutputs, wantOutputs, gotNext, wantNext, e)
			}
		}
	}
	if _, err = model.graph.MultiProveSynthesize(GiaCofactorOptions{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("object limit", err)
	}
	if _, err = model.graph.MultiProveSynthesize(GiaCofactorOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("cancellation", err)
	}
}

func TestMultiProveFindsReplayableOriginalCounterexamples(t *testing.T) {
	model := multiProveTestModel(t)
	result, err := model.MultiProve(MultiProveOptions{MaxIterations: 1, SimulationFrames: 1, SimulationWords: 1, InitialBMCFrames: 2, MaxSimulation: 10000, UseSynthesis: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Reason != MultiProveAllResolved || result.Iterations != 1 || result.ConstantZero != 1 || result.FinalOutputs != 0 || len(result.Outputs) != 3 {
		t.Fatal(result.Reason, result.Iterations, result.ConstantZero, result.FinalOutputs, result.Outputs)
	}
	if result.Outputs[0].Status != MultiProveCounterexample || result.Outputs[1].Status != MultiProveConstantZero || result.Outputs[2].Status != MultiProveCounterexample {
		t.Fatal(result.Outputs)
	}
	for property, output := range result.Outputs {
		if output.Counterexample == nil {
			continue
		}
		if output.Counterexample.Property != property || output.Counterexample.Registers() != 1 || output.Counterexample.Inputs() != 1 {
			t.Fatal("original coordinates", property, output.Counterexample)
		}
		if err := model.ValidateCounterexample(output.Counterexample); err != nil {
			t.Fatal(property, err)
		}
	}
	if result.Outputs[2].Counterexample.Frames() != 2 {
		t.Fatal("latch property must fail in frame one", result.Outputs[2].Counterexample.Frames())
	}
	stages := make([]string, len(result.Reports))
	for index, report := range result.Reports {
		stages[index] = report.Stage
	}
	if !slices.Equal(stages, []string{"SIM", "BMC", "SYN"}) {
		t.Fatal(stages)
	}
}

func TestMultiProveLimitsCancellationAndGiaEntry(t *testing.T) {
	model := multiProveTestModel(t)
	beforeObjects := slices.Clone(model.graph.objects)
	if result, err := model.MultiProve(MultiProveOptions{Stop: func() bool { return true }}); err != nil || result.Reason != MultiProveCancelled {
		t.Fatal(result.Reason, err)
	}
	if _, err := model.MultiProve(MultiProveOptions{MaxIterations: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := model.MultiProve(MultiProveOptions{MaxIterations: 1, SimulationWords: 2, MaxSimulation: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	combinational, err := NewAigerModel(NewAigManager(), 0, 0, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = combinational.MultiProve(MultiProveOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	gia, _, err := GiaFromAig(model.graph)
	if err != nil {
		t.Fatal(err)
	}
	result, err := gia.MultiProve(MultiProveOptions{MaxIterations: 1, SimulationFrames: 1, SimulationWords: 1, InitialBMCFrames: 2, MaxSimulation: 10000})
	if err != nil || result.FinalOutputs != 0 {
		t.Fatal(result.FinalOutputs, err)
	}
	if !slices.Equal(beforeObjects, model.graph.objects) {
		t.Fatal("multi-prove mutated its source graph")
	}
}

func TestMultiProveGapAndConstraintPrefixes(t *testing.T) {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	latch, _ := graph.CreateCI()
	graph.CreateCO(input)
	graph.CreateCO(input.Not()) // raw positive AIGER constraint
	graph.CreateCO(latch)
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	model, err := NewAigerModel(graph, 1, 0, 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := model.MultiProve(MultiProveOptions{MaxIterations: 4, GapIterations: 1, SimulationFrames: 2, SimulationWords: 1, InitialBMCFrames: 2, MaxSimulation: 10000})
	if err != nil {
		t.Fatal(err)
	}
	if result.Reason != MultiProveGapLimit || result.Iterations != 1 || result.FinalOutputs != 1 || result.Outputs[0].Status != MultiProveUnresolved {
		t.Fatal(result.Reason, result.Iterations, result.FinalOutputs, result.Outputs)
	}
	for _, report := range result.Reports {
		if report.TotalOutputs != 2 || report.PrimaryOutputs != 1 || report.Solved != 0 {
			t.Fatal("constraint/report accounting", report)
		}
	}
}

func FuzzProcessMultiProveOutputs(f *testing.F) {
	for _, seed := range []struct{ data []byte }{{nil}, {[]byte{1, 0, 1}}, {[]byte{0, 1, 1, 0, 1}}} {
		f.Add(seed.data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		count := len(data) % 32
		input := make([]*Counterexample, count)
		output := make([]*Counterexample, count)
		mapping := make([]int, count)
		for index := range mapping {
			mapping[index] = count - 1 - index
			if data[index]&1 != 0 {
				input[index], _ = NewCounterexample(0, 0, 1)
				input[index].Property = index
			}
		}
		left, compacted, err := ProcessMultiProveOutputs(input, output, mapping)
		if err != nil {
			t.Fatal(err)
		}
		if len(left) != len(compacted) {
			t.Fatal(left, compacted)
		}
		for local, original := range mapping {
			found := data[local]&1 != 0
			if found != (output[original] != nil) || found && input[local] != nil || !found && input[local] != nil {
				t.Fatal(local, original, found)
			}
		}
		for index, local := range left {
			if compacted[index] != mapping[local] || data[local]&1 != 0 {
				t.Fatal(index, local, compacted[index])
			}
		}
	})
}

func FuzzMultiProveReplay(f *testing.F) {
	f.Add(byte(1), byte(1), byte(2))
	f.Add(byte(2), byte(3), byte(4))
	f.Fuzz(func(t *testing.T, words, simulationFrames, bmcFrames byte) {
		model := multiProveTestModel(t)
		options := MultiProveOptions{MaxIterations: 1, SimulationWords: int(words%2) + 1, SimulationFrames: int(simulationFrames%3) + 1, InitialBMCFrames: int(bmcFrames%3) + 1, MaxSimulation: 100000}
		result, err := model.MultiProve(options)
		if err != nil {
			t.Fatal(err)
		}
		for property, output := range result.Outputs {
			switch output.Status {
			case MultiProveCounterexample:
				if output.Counterexample == nil || output.Counterexample.Property != property {
					t.Fatal(property, output)
				}
				if err := model.ValidateCounterexample(output.Counterexample); err != nil {
					t.Fatal(property, err)
				}
			case MultiProveConstantZero:
				if model.graph.objects[model.graph.cos[property]].Fanin0 != AigConst0 {
					t.Fatal("false constant-zero claim", property)
				}
			case MultiProveUnresolved:
			default:
				t.Fatal("invalid status", output.Status)
			}
		}
	})
}
