package abc

// Go-only model and fuzz tests for src/sat/bmc/bmcCexMin1.c.

import (
	"errors"
	"slices"
	"testing"
)

func cexMin1SequentialAig(t testing.TB) *AigManager {
	t.Helper()
	g := NewAigManager()
	p, _ := g.CreateCI()
	q, _ := g.CreateCI()
	r, _ := g.CreateCI()
	and, err := g.And(p, q)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.CreateCO(r); err != nil {
		t.Fatal(err)
	}
	if _, err = g.CreateCO(and); err != nil {
		t.Fatal(err)
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = g.Check(); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestAigCexReasonSequentialTermsPrioritiesAndReasons(t *testing.T) {
	g := cexMin1SequentialAig(t)
	trace := cexMin2SequentialTrace(t)
	before := slices.Clone(g.objects)
	state, err := g.ComputeAigCounterexampleReasons(trace, false, AigCexReasonOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(state.FrameCIs[0], []AigID{g.cis[0], g.cis[1]}) || !slices.Equal(state.FrameCIs[1], []AigID{g.cis[2]}) {
		t.Fatal("frame terms", state.FrameCIs)
	}
	if len(state.Priorities[0]) != 2 || state.Priorities[0][0] != (AigCexPhasePriority{Priority: 2, Value: true}) || state.Priorities[0][1] != (AigCexPhasePriority{Priority: 1, Value: true}) {
		t.Fatal("frame-zero priorities", state.Priorities[0])
	}
	if len(state.Priorities[1]) != 1 || state.Priorities[1][0] != (AigCexPhasePriority{Priority: 1, Value: true}) {
		t.Fatal("inherited register priority", state.Priorities[1])
	}
	if len(state.Reasons[0]) != 0 || !slices.Equal(state.Reasons[1], []AigCexReasonLiteral{{Index: 0, Value: true}}) {
		t.Fatal("state reasons", state.Reasons)
	}
	pi, err := g.ComputeAigCounterexampleReasons(trace, true, AigCexReasonOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(pi.Reasons[0], []AigCexReasonLiteral{{Index: 0, Value: true}, {Index: 1, Value: true}}) || len(pi.Reasons[1]) != 0 {
		t.Fatal("PI reasons", pi.Reasons)
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("reason analysis mutated source")
	}
}

func TestAigCexReasonControllingZeroPriority(t *testing.T) {
	build := func(complement bool, assignment int) (*AigManager, *Counterexample) {
		g := NewAigManager()
		a, _ := g.CreateCI()
		b, _ := g.CreateCI()
		and, _ := g.And(a, b)
		g.CreateCO(and.NotIf(complement))
		trace, _ := NewCounterexample(0, 2, 1)
		trace.Property = 0
		trace.SetInput(0, 0, assignment&1 != 0)
		trace.SetInput(0, 1, assignment&2 != 0)
		return g, trace
	}
	for _, test := range []struct {
		complement bool
		assignment int
		want       []AigCexReasonLiteral
	}{
		{true, 0, []AigCexReasonLiteral{{Index: 0, Value: false}}},
		{true, 2, []AigCexReasonLiteral{{Index: 0, Value: false}}},
		{true, 1, []AigCexReasonLiteral{{Index: 1, Value: false}}},
		{false, 3, []AigCexReasonLiteral{{Index: 0, Value: true}, {Index: 1, Value: true}}},
	} {
		g, trace := build(test.complement, test.assignment)
		analysis, err := g.ComputeAigCounterexampleReasons(trace, true, AigCexReasonOptions{})
		if err != nil || !slices.Equal(analysis.Reasons[0], test.want) {
			t.Fatal(test, analysis.Reasons, err)
		}
	}
}

func TestAigCexReasonAlternateRegisterPriorityBand(t *testing.T) {
	g := NewAigManager()
	p, _ := g.CreateCI()
	q, _ := g.CreateCI()
	g.CreateCI()
	r, _ := g.CreateCI()
	node, _ := g.And(q, r)
	g.CreateCO(node.Not())
	g.CreateCO(p)
	g.CreateCO(p)
	g.SetRegisterCount(2)
	trace, _ := NewCounterexample(2, 2, 1)
	trace.Property = 0
	ordinaryPI, err := g.ComputeAigCounterexampleReasons(trace, true, AigCexReasonOptions{})
	if err != nil || !slices.Equal(ordinaryPI.Reasons[0], []AigCexReasonLiteral{{Index: 1, Value: false}}) {
		t.Fatal("registers-first band", ordinaryPI.Reasons, err)
	}
	highPI, err := g.ComputeAigCounterexampleReasons(trace, true, AigCexReasonOptions{RegistersHighPriority: true})
	if err != nil || len(highPI.Reasons[0]) != 0 {
		t.Fatal("high-register PI reasons", highPI.Reasons, err)
	}
	highRegisters, err := g.ComputeAigCounterexampleReasons(trace, false, AigCexReasonOptions{RegistersHighPriority: true})
	if err != nil || !slices.Equal(highRegisters.Reasons[0], []AigCexReasonLiteral{{Index: 1, Value: false}}) {
		t.Fatal("high-register state reason", highRegisters.Reasons, err)
	}
}

func TestAigCexReasonCubeGraph(t *testing.T) {
	g := cexMin1SequentialAig(t)
	trace := cexMin2SequentialTrace(t)
	result, err := g.MinimizeAigCounterexampleReasons(trace, AigCexReasonOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cube := result.CubeGraph
	if cube.Name != g.Name || cube.Spec != "" || cube.RegisterCount() != 1 || cube.PrimaryInputCount() != 2 || cube.PrimaryOutputCount() != 1 || cube.OutputCount() != 2 {
		t.Fatal("cube graph shape", cube.PrimaryInputCount(), cube.PrimaryOutputCount(), cube.OutputCount())
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		outputs, next, err := cube.Step(inputs, []bool{false})
		want := inputs[0] && inputs[1]
		if err != nil || outputs[0] != want || next[0] != want {
			t.Fatal(assignment, outputs, next, err)
		}
	}
	falseCube, err := g.DuplicateAigWithRegisterCubes([][]AigCexReasonLiteral{nil, {{Index: 0, Value: false}}}, AigCexReasonOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		outputs, next, err := falseCube.Step(inputs, []bool{false})
		wantNext := inputs[0] && inputs[1]
		if err != nil || outputs[0] != !wantNext || next[0] != wantNext {
			t.Fatal("negative register literal", assignment, outputs, next, err)
		}
	}
}

func TestAigCexReasonInitialStateAndBuffer(t *testing.T) {
	g := NewAigManager()
	r, _ := g.CreateCI()
	buffer, _ := g.CreateBuffer(r)
	g.CreateCO(buffer)
	g.CreateCO(r)
	g.SetRegisterCount(1)
	trace, _ := NewCounterexample(1, 0, 1)
	trace.Property = 0
	trace.SetInitial(0, true)
	analysis, err := g.ComputeAigCounterexampleReasons(trace, false, AigCexReasonOptions{})
	if err != nil || !slices.Equal(analysis.Reasons[0], []AigCexReasonLiteral{{Index: 0, Value: true}}) {
		t.Fatal(analysis.Reasons, err)
	}
}

func TestAigCexReasonValidationLimitsCancellationAndProvenance(t *testing.T) {
	g := cexMin1SequentialAig(t)
	trace := cexMin2SequentialTrace(t)
	for _, options := range []AigCexReasonOptions{{MaxObjects: -1}, {MaxVisits: -1}, {MaxFrames: -1}, {MaxEntries: -1}} {
		if _, err := g.ComputeAigCounterexampleReasons(trace, false, options); !errors.Is(err, ErrDimension) {
			t.Fatal(options, err)
		}
	}
	for _, options := range []AigCexReasonOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxFrames: 1}, {MaxEntries: 1}} {
		if _, err := g.ComputeAigCounterexampleReasons(trace, false, options); !errors.Is(err, ErrLimit) {
			t.Fatal(options, err)
		}
	}
	if _, err := g.ComputeAigCounterexampleReasons(trace, false, AigCexReasonOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("cancel", err)
	}
	bad := trace.Clone()
	bad.Property = 1
	if _, err := g.ComputeAigCounterexampleReasons(bad, false, AigCexReasonOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("property", err)
	}
	if _, err := g.DuplicateAigWithRegisterCubes([][]AigCexReasonLiteral{nil, {{Index: 1}}}, AigCexReasonOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("register literal", err)
	}
	constrained := g.Clone()
	constrained.SetConstraintCount(1)
	if _, err := constrained.DuplicateAigWithRegisterCubes([][]AigCexReasonLiteral{nil}, AigCexReasonOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("constraints", err)
	}
	analysis, err := g.ComputeAigCounterexampleReasons(trace, false, AigCexReasonOptions{})
	if err != nil {
		t.Fatal(err)
	}
	changed := g.Clone()
	changed.objects[changed.cos[0]].Fanin0 = AigConst0
	if _, err = changed.DuplicateAigWithReasonAnalysis(analysis, AigCexReasonOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal("stale analysis", err)
	}
	piAnalysis, err := g.ComputeAigCounterexampleReasons(trace, true, AigCexReasonOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.DuplicateAigWithReasonAnalysis(piAnalysis, AigCexReasonOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("PI reasons as register cubes", err)
	}
}

func TestAigCexReasonDeepIterativeCone(t *testing.T) {
	g := NewAigManager()
	a, _ := g.CreateCI()
	b, _ := g.CreateCI()
	node, _ := g.And(a, b)
	for index := 0; index < 10000; index++ {
		fanin := a
		if index&1 != 0 {
			fanin = b
		}
		var err error
		node, err = g.And(node, fanin)
		if err != nil {
			t.Fatal(index, err)
		}
	}
	g.CreateCO(node)
	trace, _ := NewCounterexample(0, 2, 1)
	trace.Property = 0
	trace.SetInput(0, 0, true)
	trace.SetInput(0, 1, true)
	analysis, err := g.ComputeAigCounterexampleReasons(trace, true, AigCexReasonOptions{})
	if err != nil || !slices.Equal(analysis.Reasons[0], []AigCexReasonLiteral{{Index: 1, Value: true}, {Index: 0, Value: true}}) {
		t.Fatal(analysis.Reasons, err)
	}
}

func FuzzAigCexReasonsForceFailure(f *testing.F) {
	f.Add(uint8(8), uint8(3))
	f.Add(uint8(0x96), uint8(2))
	f.Fuzz(func(t *testing.T, truth, rawAssignment uint8) {
		assignment := int(rawAssignment & 3)
		if truth>>uint(assignment)&1 == 0 {
			return
		}
		gia := ecoTruthGraph(t, truth)
		g, _, err := gia.ToAig()
		if err != nil {
			t.Fatal(err)
		}
		trace, err := NewCounterexample(0, 2, 1)
		if err != nil {
			t.Fatal(err)
		}
		trace.Property = 0
		trace.SetInput(0, 0, assignment&1 != 0)
		trace.SetInput(0, 1, assignment&2 != 0)
		analysis, err := g.ComputeAigCounterexampleReasons(trace, true, AigCexReasonOptions{})
		if err != nil {
			t.Fatal(truth, assignment, err)
		}
		for candidate := 0; candidate < 4; candidate++ {
			compatible := true
			for _, literal := range analysis.Reasons[0] {
				compatible = compatible && (candidate>>literal.Index&1 != 0) == literal.Value
			}
			if compatible && truth>>uint(candidate)&1 == 0 {
				t.Fatal("reason does not force failure", truth, assignment, candidate, analysis.Reasons[0])
			}
		}
	})
}
