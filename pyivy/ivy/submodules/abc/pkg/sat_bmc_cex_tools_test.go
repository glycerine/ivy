package abc

// Go-only exhaustive and model tests for src/sat/bmc/bmcCexTools.c.

import (
	"encoding/binary"
	"errors"
	"slices"
	"strings"
	"testing"
)

func cexToolSequentialGraph(t testing.TB) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	p, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(r); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(p); err != nil {
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

func cexToolSequentialTrace(t testing.TB) *Counterexample {
	t.Helper()
	trace, err := NewCounterexample(1, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	trace.Property = 0
	if err = trace.SetInput(0, 0, true); err != nil {
		t.Fatal(err)
	}
	return trace
}

func TestCounterexampleInnerCareEssentialPipeline(t *testing.T) {
	g := cexToolSequentialGraph(t)
	trace := cexToolSequentialTrace(t)
	before := slices.Clone(g.objects)
	analysis, err := g.AnalyzeCounterexample(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !analysis.CareVerified || !analysis.MinimizedVerified || !slices.Equal(before, g.objects) {
		t.Fatal("verification/source ownership")
	}
	wantState := [][]bool{{true, false}, {false, true}}
	wantImplication := [][]bool{{false, true}, {false, false}}
	for frame := 0; frame < 2; frame++ {
		for input := 0; input < 2; input++ {
			state, _ := analysis.States.Input(frame, input)
			implied, _ := analysis.Implications.Input(frame, input)
			if state != wantState[frame][input] || implied != wantImplication[frame][input] {
				t.Fatal("expanded trace", frame, input, state, implied)
			}
		}
	}
	for _, mask := range []*Counterexample{analysis.Care, analysis.Essential, analysis.Minimized} {
		pi0, _ := mask.Input(0, 0)
		ro1, _ := mask.Input(1, 1)
		if !pi0 || !ro1 {
			t.Fatal("missing causal PI/state bit", mask.PackedWords())
		}
	}
	if analysis.Statistics.TotalInputBits != 2 || analysis.Statistics.CareBits != 1 || analysis.Statistics.EssentialBits != 1 || analysis.Statistics.MinimizedBits != 1 {
		t.Fatal("statistics", analysis.Statistics)
	}
}

func TestCounterexampleUnrollAndMonotoneNetwork(t *testing.T) {
	g := cexToolSequentialGraph(t)
	trace := cexToolSequentialTrace(t)
	unrolled, err := g.UnrollCounterexampleProperty(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if unrolled.RegisterCount() != 0 || unrolled.PrimaryInputCount() != 2 || unrolled.PrimaryOutputCount() != 1 {
		t.Fatal("unroll shape")
	}
	network, err := g.BuildCounterexampleNetwork(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if network.Name != "unate" || network.PrimaryInputCount() != 2 || network.PrimaryOutputCount() != 1 {
		t.Fatal("network shape")
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		unrolledOutputs, err := unrolled.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		networkOutputs, err := network.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if unrolledOutputs[0] != inputs[0] || networkOutputs[0] != inputs[0] {
			t.Fatal(assignment, unrolledOutputs, networkOutputs)
		}
	}
	extended := NewGiaManager()
	a, _ := extended.AppendCI()
	b, _ := extended.AppendCI()
	x, _ := extended.AppendXorReal(a, b)
	extended.AppendCO(x)
	extendedTrace, _ := NewCounterexample(0, 2, 1)
	extendedTrace.Property = 0
	extendedTrace.SetInput(0, 0, true)
	if graph, err := extended.BuildCounterexampleNetwork(extendedTrace, CounterexampleToolOptions{}); graph != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended source recurrence", err)
	}
}

func TestCounterexampleMonotoneNetworkANDRecurrences(t *testing.T) {
	build := func(complement bool, assignment int) (*GiaManager, *Counterexample) {
		g := NewGiaManager()
		a, _ := g.AppendCI()
		b, _ := g.AppendCI()
		and, _ := g.AppendAnd(a, b)
		g.AppendCO(and.NotIf(complement))
		trace, _ := NewCounterexample(0, 2, 1)
		trace.Property = 0
		trace.SetInput(0, 0, assignment&1 != 0)
		trace.SetInput(0, 1, assignment&2 != 0)
		return g, trace
	}
	for _, test := range []struct {
		complement bool
		assignment int
		want       func(bool, bool) bool
	}{
		{false, 3, func(a, b bool) bool { return a && b }},
		{true, 0, func(a, b bool) bool { return a || b }},
		{true, 2, func(a, _ bool) bool { return a }},
	} {
		g, trace := build(test.complement, test.assignment)
		network, err := g.BuildCounterexampleNetwork(trace, CounterexampleToolOptions{})
		if err != nil {
			t.Fatal(test.assignment, err)
		}
		for mask := 0; mask < 4; mask++ {
			a, b := mask&1 != 0, mask&2 != 0
			outputs, err := network.Evaluate([]bool{a, b})
			if err != nil || outputs[0] != test.want(a, b) {
				t.Fatal(test.assignment, mask, outputs, err)
			}
		}
	}
}

func TestCounterexampleAnalysisExhaustiveTwoInputFunctions(t *testing.T) {
	for truth := 1; truth < 256; truth++ {
		g := ecoTruthGraph(t, uint8(truth))
		for assignment := 0; assignment < 4; assignment++ {
			if truth>>uint(assignment)&1 == 0 {
				continue
			}
			trace, err := NewCounterexample(0, 2, 1)
			if err != nil {
				t.Fatal(err)
			}
			trace.Property = 0
			trace.SetInput(0, 0, assignment&1 != 0)
			trace.SetInput(0, 1, assignment&2 != 0)
			analysis, err := g.AnalyzeCounterexample(trace, CounterexampleToolOptions{})
			if err != nil {
				t.Fatal(truth, assignment, err)
			}
			for input := 0; input < 2; input++ {
				cared, _ := analysis.Care.Input(0, input)
				essential, _ := analysis.Essential.Input(0, input)
				flippedAssignment := assignment ^ (1 << input)
				wantEssential := truth>>uint(flippedAssignment)&1 == 0
				if essential != (cared && wantEssential) {
					t.Fatal("essential bit", truth, assignment, input, cared, essential, wantEssential)
				}
			}
		}
	}
}

func TestCounterexampleCareVerificationDetectsInsufficientMask(t *testing.T) {
	g := ecoTruthGraph(t, 8) // a & b
	trace, _ := NewCounterexample(0, 2, 1)
	trace.Property = 0
	trace.SetInput(0, 0, true)
	trace.SetInput(0, 1, true)
	care, _ := NewCounterexample(0, 2, 1)
	care.Property = 0
	ok, err := g.VerifyCounterexampleCare(trace, care, CounterexampleToolOptions{})
	if err != nil || ok {
		t.Fatal(ok, err)
	}
	care.SetInput(0, 0, true)
	care.SetInput(0, 1, true)
	ok, err = g.VerifyCounterexampleCare(trace, care, CounterexampleToolOptions{})
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
}

func TestCounterexampleAnalysisNativeXorMux(t *testing.T) {
	g := NewGiaManager()
	control, _ := g.AppendCI()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	xor, _ := g.AppendXorReal(a, b)
	mux, _ := g.AppendMuxReal(control, xor, a)
	g.AppendCO(mux)
	trace, _ := NewCounterexample(0, 3, 1)
	trace.Property = 0
	trace.SetInput(0, 0, true)
	trace.SetInput(0, 1, true)
	analysis, err := g.AnalyzeCounterexample(trace, CounterexampleToolOptions{})
	if err != nil || !analysis.CareVerified || !analysis.MinimizedVerified {
		t.Fatal(analysis, err)
	}
	controlCare, _ := analysis.Care.Input(0, 0)
	aCare, _ := analysis.Care.Input(0, 1)
	bCare, _ := analysis.Care.Input(0, 2)
	if controlCare || !aCare || !bCare {
		t.Fatal("mux/XOR local care", controlCare, aCare, bCare)
	}
	for input, want := range []bool{false, true, true} {
		essential, _ := analysis.Essential.Input(0, input)
		if essential != want {
			t.Fatal("extended essential", input, essential, want)
		}
	}
}

func TestCounterexampleAnalysisDeepIterativeCone(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	node, _ := g.AppendAnd(a, b)
	for index := 0; index < 10000; index++ {
		fanin := a
		if index&1 != 0 {
			fanin = b
		}
		var err error
		node, err = g.AppendAnd(node, fanin)
		if err != nil {
			t.Fatal(index, err)
		}
	}
	g.AppendCO(node)
	trace, _ := NewCounterexample(0, 2, 1)
	trace.Property = 0
	trace.SetInput(0, 0, true)
	trace.SetInput(0, 1, true)
	analysis, err := g.AnalyzeCounterexample(trace, CounterexampleToolOptions{})
	if err != nil || !analysis.MinimizedVerified {
		t.Fatal(err)
	}
	for input := 0; input < 2; input++ {
		essential, _ := analysis.Essential.Input(0, input)
		if !essential {
			t.Fatal("deep essential", input)
		}
	}
}

func TestCounterexampleDiagnosticsAndCounts(t *testing.T) {
	g := cexToolSequentialGraph(t)
	g.Name = "counter"
	trace := cexToolSequentialTrace(t)
	if count, err := CounterexampleInputOneCount(trace, 1); err != nil || count != 1 {
		t.Fatal(count, err)
	}
	formatted, err := FormatCounterexampleInputs(trace, 1, true)
	if err != nil || !strings.Contains(formatted, "  0 : 1   1") || !strings.Contains(formatted, "  1 : 0   0") {
		t.Fatal(formatted, err)
	}
	short, err := FormatCounterexampleInputs(trace, 1, false)
	if err != nil || strings.Contains(short, "  0 :") {
		t.Fatal(short, err)
	}
	stats, err := g.CounterexampleAOGStatistics(0)
	if err != nil || stats.Name != "counter" || stats.PrimaryInputs != 1 || stats.Gates != 0 || stats.Levels != 0 {
		t.Fatal(stats, err)
	}
}

func TestGiaCareNodeStatisticsIndependentPatterns(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendAnd(a, b)
	d, _ := g.AppendAnd(c, a.Not())
	g.AppendCO(c)
	g.AppendCO(d)
	patterns := [][]bool{{false, false}, {true, false}, {false, true}, {true, true}}
	stats, err := g.CareNodeStatistics(patterns, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if stats.GateCount != 2 || len(stats.PatternCounts) != 4 {
		t.Fatal(stats)
	}
	total := 0
	for index, count := range stats.PatternCounts {
		if count < 1 || count > 2 {
			t.Fatal(index, count)
		}
		total += count
	}
	if stats.Average != float64(total)/4 {
		t.Fatal(stats.Average, total)
	}
}

func TestDecodeMNISTImagesOwnedAndBounded(t *testing.T) {
	data := make([]byte, 20)
	binary.BigEndian.PutUint32(data[0:4], 2051)
	binary.BigEndian.PutUint32(data[4:8], 2)
	binary.BigEndian.PutUint32(data[8:12], 1)
	binary.BigEndian.PutUint32(data[12:16], 2)
	data[16], data[17], data[18], data[19] = 1, 128, 3, 0
	patterns, err := DecodeMNISTImages(data, 0, 0)
	if err != nil || len(patterns) != 2 || len(patterns[0]) != 16 || !patterns[0][0] || !patterns[0][15] || !patterns[1][0] || !patterns[1][1] {
		t.Fatal(patterns, err)
	}
	patterns[0][0] = false
	if !patterns[1][0] {
		t.Fatal("rows alias")
	}
	one, err := DecodeMNISTImages(data, 1, 16)
	if err != nil || len(one) != 1 {
		t.Fatal(len(one), err)
	}
	for _, bad := range [][]byte{nil, data[:19], append([]byte{}, data...)} {
		if len(bad) == len(data) {
			bad[3] = 0
		}
		if _, err := DecodeMNISTImages(bad, 0, 0); err == nil {
			t.Fatal("accepted malformed MNIST")
		}
	}
	if _, err := DecodeMNISTImages(data, 0, 31); !errors.Is(err, ErrLimit) {
		t.Fatal("bit limit", err)
	}
}

func TestCounterexampleToolsLimitsCancellationAndDimensions(t *testing.T) {
	g := cexToolSequentialGraph(t)
	trace := cexToolSequentialTrace(t)
	for _, options := range []CounterexampleToolOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxFrames: -1}, {MaxBits: -1}} {
		if _, _, err := g.CounterexampleInnerStates(trace, options); !errors.Is(err, ErrDimension) {
			t.Fatal(options, err)
		}
	}
	for _, options := range []CounterexampleToolOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxFrames: 1}, {MaxBits: 1}} {
		if _, _, err := g.CounterexampleInnerStates(trace, options); !errors.Is(err, ErrLimit) {
			t.Fatal(options, err)
		}
	}
	if _, _, err := g.CounterexampleInnerStates(trace, CounterexampleToolOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("cancel", err)
	}
	states, implications, err := g.CounterexampleInnerStates(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.CounterexampleCareBits(states, implications, nil, true, CounterexampleToolOptions{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("care object limit", err)
	}
	if _, err = g.CounterexampleCareBits(states, implications, nil, true, CounterexampleToolOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("care cancellation", err)
	}
	if _, err = g.CounterexampleEssentialBits(states, states, CounterexampleToolOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("essential cancellation", err)
	}
	if _, err = g.VerifyCounterexampleCare(trace, states, CounterexampleToolOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("verify cancellation", err)
	}
	if _, err = g.BuildCounterexampleNetwork(trace, CounterexampleToolOptions{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("network object limit", err)
	}
	if _, err = g.UnrollCounterexampleProperty(trace, CounterexampleToolOptions{MaxMappingEntries: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("unroll mapping limit", err)
	}
	bad, _ := NewCounterexample(0, 1, 2)
	bad.Property = 0
	if _, _, err := g.CounterexampleInnerStates(bad, CounterexampleToolOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("dimensions", err)
	}
}

func FuzzCounterexampleCareAgainstTruthTable(f *testing.F) {
	f.Add(uint8(8), uint8(3))
	f.Add(uint8(0x96), uint8(2))
	f.Fuzz(func(t *testing.T, truth, rawAssignment uint8) {
		assignment := int(rawAssignment & 3)
		if truth>>uint(assignment)&1 == 0 {
			return
		}
		g := ecoTruthGraph(t, truth)
		trace, err := NewCounterexample(0, 2, 1)
		if err != nil {
			t.Fatal(err)
		}
		trace.Property = 0
		trace.SetInput(0, 0, assignment&1 != 0)
		trace.SetInput(0, 1, assignment&2 != 0)
		analysis, err := g.AnalyzeCounterexample(trace, CounterexampleToolOptions{})
		if err != nil {
			t.Fatal(truth, assignment, err)
		}
		for _, mask := range []*Counterexample{analysis.Care, analysis.Minimized} {
			for candidate := 0; candidate < 4; candidate++ {
				compatible := true
				for input := 0; input < 2; input++ {
					known, _ := mask.Input(0, input)
					compatible = compatible && (!known || candidate>>input&1 == assignment>>input&1)
				}
				if compatible && truth>>uint(candidate)&1 == 0 {
					t.Fatal("mask permits non-failure", truth, assignment, candidate, mask.PackedWords())
				}
			}
		}
	})
}
