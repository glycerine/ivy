package abc

// Standalone tests for the remaining GIA replay, simulation, and cone subset.

import (
	"errors"
	"slices"
	"sync"
	"testing"
)

func giaRemainingSequential(t *testing.T) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	p, _ := m.AppendCI()
	r, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, r)
	bad, _ := m.AppendAnd(p, r)
	m.AppendCO(bad)
	m.AppendCO(x)
	if err := m.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestGiaRemainingTraceAndCounterexample(t *testing.T) {
	m := giaRemainingSequential(t)
	inputs := [][]bool{{true}, {false}, {true}}
	trace, err := m.SimulateTrace([]bool{false}, inputs, GiaTraceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantOutputs := [][]bool{{false}, {false}, {true}}
	wantStates := [][]bool{{false}, {true}, {true}, {false}}
	if !slices.EqualFunc(trace.Outputs, wantOutputs, slices.Equal[[]bool]) ||
		!slices.EqualFunc(trace.States, wantStates, slices.Equal[[]bool]) {
		t.Fatalf("trace outputs=%v states=%v", trace.Outputs, trace.States)
	}
	inputs[0][0] = false
	if !trace.States[1][0] || trace.Outputs[2][0] != true {
		t.Fatal("result aliases caller")
	}
	cex := GiaCounterexample{InitialState: []bool{false}, Inputs: [][]bool{{true}, {false}, {true}}, Output: 0, Frame: 2}
	if ok, e := m.VerifyCounterexample(cex, GiaTraceOptions{}); e != nil || !ok {
		t.Fatal(ok, e)
	}
	cex.Frame = 1
	cex.Inputs = cex.Inputs[:2]
	if ok, e := m.VerifyCounterexample(cex, GiaTraceOptions{}); e != nil || ok {
		t.Fatal(ok, e)
	}
	failure, ok, err := m.FindFirstFailure([]bool{false}, [][]bool{{true}, {false}, {true}}, 0, GiaTraceOptions{})
	if err != nil || !ok || failure != (GiaFailure{Output: 0, Frame: 2}) {
		t.Fatal(failure, ok, err)
	}
}

func TestGiaRemainingDualAndValidation(t *testing.T) {
	var nilManager *GiaManager
	if _, err := nilManager.SimulateTrace(nil, nil, GiaTraceOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := nilManager.SimulatePatterns(nil, GiaTraceOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := nilManager.PartitionOutputCones(GiaTraceOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	m := NewGiaManager()
	a, _ := m.AppendCI()
	m.AppendCO(a)
	m.AppendCO(a.Not())
	cex := GiaCounterexample{Inputs: [][]bool{{false}}, Frame: 0, Output: 0, DualOutput: true}
	if ok, err := m.VerifyCounterexample(cex, GiaTraceOptions{}); err != nil || !ok {
		t.Fatal(ok, err)
	}
	bad := []GiaCounterexample{
		{Inputs: nil, Frame: 0},
		{Inputs: [][]bool{{false}, {true}}, Frame: 0},
		{Inputs: [][]bool{{false}}, Frame: 0, Output: 1, DualOutput: true},
	}
	for _, c := range bad {
		if _, err := m.VerifyCounterexample(c, GiaTraceOptions{}); !errors.Is(err, ErrDimension) {
			t.Fatal(c, err)
		}
	}
	final := GiaCounterexample{Inputs: [][]bool{{true}, {false}}, Frame: 1, Output: 0}
	one := NewGiaManager()
	oneInput, _ := one.AppendCI()
	one.AppendCO(oneInput)
	if ok, err := one.VerifyCounterexample(final, GiaTraceOptions{}); err != nil || ok {
		t.Fatal("earlier assertion must not validate a final-frame CEX", ok, err)
	}
	paired := NewGiaManager()
	pairInput, _ := paired.AppendCI()
	paired.AppendCO(pairInput)
	paired.AppendCO(GiaConst0)
	dual := GiaCounterexample{Inputs: [][]bool{{true}, {false}}, Frame: 1, Output: 0, DualOutput: true}
	if ok, err := paired.VerifyCounterexample(dual, GiaTraceOptions{}); err != nil || ok {
		t.Fatal("dual comparison must use the supplied final frame", ok, err)
	}
	before := slices.Clone(m.objects)
	for _, options := range []GiaTraceOptions{
		{MaxFrames: 1},
		{MaxObjects: 2},
		{MaxWork: 1},
		{MaxBytes: 1},
		{MaxBytes: -1},
		{Stop: func() bool { return true }},
	} {
		if _, err := m.SimulateTrace(nil, [][]bool{{false}, {true}}, options); err == nil {
			t.Fatal("expected bounded failure", options)
		}
	}
	if !slices.Equal(before, m.objects) {
		t.Fatal("source mutated")
	}
}

func TestGiaRemainingEarliestFailureOrdering(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	m.AppendCO(a)
	m.AppendCO(b)
	inputs := [][]bool{{false, false}, {true, true}, {false, true}}
	failure, ok, err := m.FindFirstFailure(nil, inputs, 0, GiaTraceOptions{})
	if err != nil || !ok || failure != (GiaFailure{Output: 0, Frame: 1}) {
		t.Fatal(failure, ok, err)
	}
	failure, ok, err = m.FindFirstFailure(nil, inputs, 1, GiaTraceOptions{})
	if err != nil || !ok || failure != (GiaFailure{Output: 1, Frame: 1}) {
		t.Fatal(failure, ok, err)
	}
}

func TestGiaRemainingSimulationStatistics(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	m.AppendCO(x)
	patterns := [][]bool{{false, false}, {false, true}, {true, true}, {true, false}}
	report, err := m.SimulatePatterns(patterns, GiaTraceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	xs := report.Objects[x.ID()]
	if xs.Ones != 2 || xs.Samples != 4 || xs.Transitions != 3 || xs.TransitionSamples != 3 ||
		xs.ProbabilityOne != .5 || xs.SwitchingProbability != .5 || xs.TransitionProbability != 1 {
		t.Fatal(xs)
	}
	if report.Objects[0].Ones != 0 || report.Objects[0].Transitions != 0 {
		t.Fatal(report.Objects[0])
	}
	patterns[0][0] = true
	if xs.Ones != 2 {
		t.Fatal("report aliases patterns")
	}
	if _, err = m.SimulatePatterns([][]bool{{false}}, GiaTraceOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestGiaRemainingObjectKindsComplementsAndStatisticsModel(t *testing.T) {
	m := NewGiaManager()
	c, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	mux, _ := m.AppendMuxReal(c, x, a.Not())
	buffer, _ := m.AppendBuffer(mux.Not())
	and, _ := m.AppendAnd(buffer, b.Not())
	co, _ := m.AppendCO(and.Not())
	patterns := make([][]bool, 8)
	want := map[GiaID][]bool{x.ID(): {}, mux.ID(): {}, buffer.ID(): {}, and.ID(): {}, co.ID(): {}}
	for assignment := 0; assignment < 8; assignment++ {
		cv := assignment&1 != 0
		av := assignment&2 != 0
		bv := assignment&4 != 0
		patterns[assignment] = []bool{cv, av, bv}
		xv := av != bv
		muxv := !av
		if cv {
			muxv = xv
		}
		bufv := !muxv
		andv := bufv && !bv
		want[x.ID()] = append(want[x.ID()], xv)
		want[mux.ID()] = append(want[mux.ID()], muxv)
		want[buffer.ID()] = append(want[buffer.ID()], bufv)
		want[and.ID()] = append(want[and.ID()], andv)
		want[co.ID()] = append(want[co.ID()], !andv)
	}
	report, err := m.SimulatePatterns(patterns, GiaTraceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for id, values := range want {
		ones, transitions := 0, 0
		for i, value := range values {
			if value {
				ones++
			}
			if i > 0 && value != values[i-1] {
				transitions++
			}
		}
		entry := report.Objects[id]
		if entry.Ones != ones || entry.Samples != len(values) || entry.Transitions != transitions || entry.TransitionSamples != len(values)-1 {
			t.Fatal(id, entry, values)
		}
	}
}

func TestGiaRemainingConePartition(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	d, _ := m.AppendCI()
	ab, _ := m.AppendAnd(a, b)
	cd, _ := m.AppendAnd(c, d)
	ax, _ := m.AppendXorReal(a, b)
	m.AppendCO(ab)
	m.AppendCO(cd)
	m.AppendCO(ax)
	result, err := m.PartitionOutputCones(GiaTraceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{0, 2}, {1}}
	if !slices.EqualFunc(result.Groups, want, slices.Equal[[]int]) {
		t.Fatal(result.Groups)
	}
	result.Groups[0][0] = 9
	again, _ := m.PartitionOutputCones(GiaTraceOptions{})
	if again.Groups[0][0] != 0 {
		t.Fatal("result storage reused")
	}
	if _, err = m.PartitionOutputCones(GiaTraceOptions{MaxWork: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	calls := 0
	if _, err = m.PartitionOutputCones(GiaTraceOptions{Stop: func() bool { calls++; return calls > 1 }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func TestGiaRemainingConstantConesStaySeparate(t *testing.T) {
	m := NewGiaManager()
	m.AppendCO(GiaConst0)
	m.AppendCO(GiaConst1)
	partition, err := m.PartitionOutputCones(GiaTraceOptions{})
	if err != nil || !slices.EqualFunc(partition.Groups, [][]int{{0}, {1}}, slices.Equal[[]int]) {
		t.Fatal(partition, err)
	}
}

func TestGiaRemainingSequentialConeClosure(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	r, _ := m.AppendCI()
	o0, _ := m.AppendAnd(a, r)
	o1, _ := m.AppendXorReal(a, r)
	m.AppendCO(o0)
	m.AppendCO(o1)
	m.AppendCO(a)
	m.SetRegisterCount(1)
	result, err := m.PartitionOutputCones(GiaTraceOptions{})
	if err != nil || len(result.Groups) != 1 || !slices.Equal(result.Groups[0], []int{0, 1}) {
		t.Fatal(result, err)
	}
}

func TestGiaRemainingLimitsCancellationOwnershipAndConcurrentReaders(t *testing.T) {
	m := giaRemainingSequential(t)
	beforeObjects := slices.Clone(m.objects)
	beforeCIs, beforeCOs := slices.Clone(m.cis), slices.Clone(m.cos)
	patterns := [][]bool{{false, false}, {true, false}, {false, true}, {true, true}}
	for _, run := range []func(GiaTraceOptions) error{
		func(options GiaTraceOptions) error {
			result, err := m.SimulateTrace([]bool{false}, [][]bool{{true}}, options)
			if err != nil && (result.Outputs != nil || result.States != nil) {
				t.Fatal("partial trace", result)
			}
			return err
		},
		func(options GiaTraceOptions) error {
			result, err := m.SimulatePatterns(patterns, options)
			if err != nil && result.Objects != nil {
				t.Fatal("partial statistics", result)
			}
			return err
		},
		func(options GiaTraceOptions) error {
			result, err := m.PartitionOutputCones(options)
			if err != nil && result.Groups != nil {
				t.Fatal("partial partition", result)
			}
			return err
		},
	} {
		if err := run(GiaTraceOptions{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
			t.Fatal("byte limit", err)
		}
		polls := 0
		if err := run(GiaTraceOptions{Stop: func() bool { polls++; return polls == 5 }}); !errors.Is(err, ErrCancelled) || polls != 5 {
			t.Fatal("mid-operation cancellation", polls, err)
		}
	}
	if !slices.Equal(beforeObjects, m.objects) || !slices.Equal(beforeCIs, m.cis) || !slices.Equal(beforeCOs, m.cos) {
		t.Fatal("read-only operations mutated source")
	}
	const readers = 9
	var wg sync.WaitGroup
	errs := make(chan error, readers)
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(which int) {
			defer wg.Done()
			var err error
			switch which % 3 {
			case 0:
				_, err = m.SimulateTrace([]bool{false}, [][]bool{{true}, {false}}, GiaTraceOptions{})
			case 1:
				_, err = m.SimulatePatterns(patterns, GiaTraceOptions{})
			default:
				_, err = m.PartitionOutputCones(GiaTraceOptions{})
			}
			if err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func FuzzGiaRemainingTrace(f *testing.F) {
	f.Add([]byte{1, 0, 1, 1, 0})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 32 {
			data = data[:32]
		}
		m := giaRemainingSequential(t)
		initial := []bool{len(data) != 0 && data[0]&1 != 0}
		inputs := make([][]bool, len(data))
		for i := range data {
			inputs[i] = []bool{data[i]&2 != 0}
		}
		trace, err := m.SimulateTrace(initial, inputs, GiaTraceOptions{MaxFrames: 32, MaxObjects: 16, MaxWork: 512})
		if err != nil {
			t.Fatal(err)
		}
		state := initial[0]
		for frame := range inputs {
			primary := inputs[frame][0]
			wantOutput := primary && state
			if !slices.Equal(trace.Outputs[frame], []bool{wantOutput}) || !slices.Equal(trace.States[frame], []bool{state}) {
				t.Fatalf("frame %d trace=%v/%v direct=%v/%v", frame, trace.Outputs[frame], trace.States[frame], wantOutput, state)
			}
			state = primary != state
		}
		if !slices.Equal(trace.States[len(inputs)], []bool{state}) {
			t.Fatal("final state")
		}
	})
}
