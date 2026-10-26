package abc

// Standalone Go models for src/proof/abs abstraction and refinement.

import (
	"errors"
	"math/rand"
	"slices"
	"sort"
	"testing"
)

func absTestOptions() AbsOptions {
	return AbsOptions{
		MaxStateBits: 8, MaxInputBits: 4, MaxConcreteStates: 256,
		MaxAbstractStates: 256, MaxTransitions: 1 << 16, MaxFrames: 64,
		MaxRefinements: 16, MaxTraceStates: 65, MaxDepth: 65,
		MaxBytes: 8 << 20, MaxWork: 1 << 24,
	}
}

func absSystemFromFunction(t testing.TB, stateBits, inputBits int, initial []uint64, bad func(uint64) bool, next func(uint64, uint64) uint64) *AbsSystem {
	t.Helper()
	states, inputs := 1<<stateBits, 1<<inputBits
	table := make([]uint64, states*inputs)
	badTable := make([]bool, states)
	for state := 0; state < states; state++ {
		badTable[state] = bad(uint64(state))
		for input := 0; input < inputs; input++ {
			table[state*inputs+input] = next(uint64(state), uint64(input))
		}
	}
	system, err := NewAbsSystem(stateBits, inputBits, initial, table, badTable, absTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	return system
}

func absExplicitReach(system *AbsSystem) (bool, []uint64, []uint64) {
	inputs := uint64(1) << system.InputBits
	seen := make(map[uint64]bool)
	type item struct {
		state uint64
		path  []uint64
		ins   []uint64
	}
	queue := make([]item, 0)
	for _, state := range system.Initial {
		if !seen[state] {
			seen[state] = true
			queue = append(queue, item{state: state, path: []uint64{state}})
		}
	}
	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if system.Bad[current.state] {
			return true, current.path, current.ins
		}
		for input := uint64(0); input < inputs; input++ {
			next := system.Next[current.state*inputs+input]
			if seen[next] {
				continue
			}
			seen[next] = true
			path := append(slices.Clone(current.path), next)
			ins := append(slices.Clone(current.ins), input)
			queue = append(queue, item{state: next, path: path, ins: ins})
		}
	}
	return false, nil, nil
}

func TestAbsDefaultsInventoryAndConversions(t *testing.T) {
	defaults := DefaultAbsParameters()
	if defaults.FramesPast != 4 || defaults.LearnedMax != 1000 || defaults.LearnedStart != 1000 || defaults.LearnedDelta != 200 || defaults.LearnedPercent != 70 || defaults.RatioMax != 30 || !defaults.PropagateFanout || defaults.Frame != -1 || defaults.FrameProved != -1 || defaults.FramesNoChangeLimit != 2 {
		t.Fatalf("source defaults mismatch: %+v", defaults)
	}
	wantInventory := []string{"absDup.c", "absGla.c", "absGlaOld.c", "absIter.c", "absOldCex.c", "absOldRef.c", "absOldSat.c", "absOldSim.c", "absOut.c", "absPth.c", "absRef.c", "absRefSelect.c", "absRpm.c", "absRpmOld.c", "absVta.c", "absUtil.c"}
	if !slices.Equal(AbsBuildInventory, wantInventory) || AbsInventoryReport() == "" {
		t.Fatalf("build inventory mismatch: %v", AbsBuildInventory)
	}
	wantSources := []string{"abs.c", "abs.h", "absDup.c", "absGla.c", "absGlaOld.c", "absIter.c", "absOldCex.c", "absOldRef.c", "absOldSat.c", "absOldSim.c", "absOut.c", "absPth.c", "absRef.c", "absRef.h", "absRefJ.c", "absRefJ.h", "absRefSelect.c", "absRpm.c", "absRpmOld.c", "absUtil.c", "absVta.c", "module.make"}
	if !slices.Equal(AbsSourceInventory(), wantSources) {
		t.Fatalf("source inventory mismatch: %v", AbsSourceInventory())
	}
	frames := [][]int{{0, 2, 4}, {0, 4}, {4}}
	counts, err := AbsFramesToCounts(frames, 6, absTestOptions())
	if err != nil || !slices.Equal(counts, []int{2, 0, 1, 0, 3, 0}) {
		t.Fatalf("frames to counts: %v %v", counts, err)
	}
	round, err := AbsCountsToFrames(counts, absTestOptions())
	if err != nil || !slices.Equal(round[0], []int{0, 2, 4}) || !slices.Equal(round[1], []int{0, 4}) || !slices.Equal(round[2], []int{4}) {
		t.Fatalf("counts to frames: %v %v", round, err)
	}
}

func TestAbsRealCounterexampleReplayAndOwnership(t *testing.T) {
	system := absSystemFromFunction(t, 2, 1, []uint64{0}, func(state uint64) bool { return state == 3 }, func(state, input uint64) uint64 {
		if input == 0 {
			return state
		}
		return (state + 1) & 3
	})
	manager, err := NewAbsManager(system, nil, absTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	if manager.Phase() != AbsReady {
		t.Fatalf("initial phase: %v", manager.Phase())
	}
	// The manager owns its transition system.
	system.Next[1] = 0
	system.Bad[3] = false
	result, err := manager.Analyze()
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != AbsCounterexample || result.Trace == nil || result.Stats.SpuriousTraces == 0 || !slices.Equal(result.Trace.States, []uint64{0, 1, 2, 3}) || !slices.Equal(result.Trace.Inputs, []uint64{1, 1, 1}) {
		t.Fatalf("unexpected result: %+v", result)
	}
	if manager.Phase() != AbsFinished || result.Stats.Nodes != 4 || result.Stats.Clauses != 8 || result.Stats.States == 0 || result.Stats.Frames == 0 {
		t.Fatalf("lifecycle/stats: phase=%v stats=%+v", manager.Phase(), result.Stats)
	}
	if valid, err := manager.Verify(result); err != nil || !valid {
		t.Fatalf("counterexample verification: %v %v", valid, err)
	}
	tampered := result
	tamperedTrace := result.Trace.Clone()
	tamperedTrace.States[1] = 3
	tampered.Trace = &tamperedTrace
	if valid, err := manager.Verify(tampered); err != nil || valid {
		t.Fatalf("tampered concrete trace verified: %v %v", valid, err)
	}
	replayed, real, err := manager.Replay(result.Trace.Clone())
	if err != nil || !real || !slices.Equal(replayed.States, result.Trace.States) {
		t.Fatalf("replay failed: %+v %v %v", replayed, real, err)
	}
	trace := result.Trace.Clone()
	trace.Inputs[0] = 0
	if _, real, err = manager.Replay(trace); err != nil || real {
		t.Fatalf("invalid trace replayed: %v %v", real, err)
	}
	result.Trace.States[0] = 99
	if replayed.States[0] != 0 {
		t.Fatal("trace ownership alias")
	}
	manager.Close()
	manager.Close()
	if manager.Phase() != AbsClosed {
		t.Fatalf("closed phase: %v", manager.Phase())
	}
	if _, err := manager.Analyze(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed manager: %v", err)
	}
}

func TestAbsSafeRefinementFixedPoint(t *testing.T) {
	system := absSystemFromFunction(t, 3, 1, []uint64{0}, func(state uint64) bool { return state == 7 }, func(state, input uint64) uint64 {
		if input == 0 {
			return state & 1
		}
		return (state ^ 1) & 1
	})
	manager, err := NewAbsManager(system, nil, absTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	result, err := manager.Analyze()
	if err != nil || result.Status != AbsProved || result.Reason != AbsFixedPoint || result.Trace != nil || result.Stats.SpuriousTraces == 0 {
		t.Fatalf("safe result: %+v %v", result, err)
	}
	if valid, verifyErr := manager.Verify(result); verifyErr != nil || !valid {
		t.Fatalf("proof verification: %v %v", valid, verifyErr)
	}
	if len(result.VisibleBits) == 0 || !sort.IntsAreSorted(result.VisibleBits) {
		t.Fatalf("non-monotone visible set: %v", result.VisibleBits)
	}
}

func TestAbsLimitsCancellationAndAtomicRefinement(t *testing.T) {
	system := absSystemFromFunction(t, 3, 1, []uint64{0}, func(state uint64) bool { return state == 7 }, func(state, input uint64) uint64 { return (state + input) & 7 })
	manager, err := NewAbsManager(system, []int{2}, absTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	before := manager.VisibleBits()
	if err := manager.Refine([]int{0, 99}); !errors.Is(err, ErrDimension) || !slices.Equal(before, manager.VisibleBits()) {
		t.Fatalf("failed refinement mutated: %v %v", err, manager.VisibleBits())
	}
	if err := manager.Refine([]int{0}); err != nil || !slices.Equal(manager.VisibleBits(), []int{0, 2}) {
		t.Fatalf("valid refinement: %v %v", manager.VisibleBits(), err)
	}
	if err := manager.Refine([]int{0}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("duplicate refinement: %v", err)
	}
	cancel := absTestOptions()
	stopped := false
	cancel.Stop = func() bool { return stopped }
	cancelled, err := NewAbsManager(system, nil, cancel)
	if err != nil {
		t.Fatal(err)
	}
	stopped = true
	if _, err := cancelled.Analyze(); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancellation: %v", err)
	}
	limited := absTestOptions()
	// Construction needs exactly 25 work units; analysis then gets an
	// independent budget and must exhaust it transactionally.
	limited.MaxWork = 25
	limitedManager, err := NewAbsManager(system, nil, limited)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := limitedManager.Analyze(); !errors.Is(err, ErrLimit) {
		t.Fatalf("work limit: %v", err)
	}
	frameLimit := absTestOptions()
	frameLimit.MaxFrames = 1
	frameManager, err := NewAbsManager(system, []int{0, 1, 2}, frameLimit)
	if err != nil {
		t.Fatal(err)
	}
	result, err := frameManager.Analyze()
	if err != nil || result.Status != AbsUnknown || result.Reason != AbsFrameLimit {
		t.Fatalf("frame limit: %+v %v", result, err)
	}
}

func TestAbsConstructorRefinementAndConversionResourceBounds(t *testing.T) {
	o := absTestOptions()
	next := make([]uint64, 8)
	bad := make([]bool, 4)
	workLimited := o
	workLimited.MaxWork = 12 // 1 initial + 8 transitions + 4 state slots.
	if _, err := NewAbsSystem(2, 1, []uint64{0}, next, bad, workLimited); !errors.Is(err, ErrLimit) {
		t.Fatalf("constructor work limit: %v", err)
	}
	cancelled := o
	cancelled.Stop = func() bool { return true }
	if _, err := NewAbsSystem(2, 1, []uint64{0}, next, bad, cancelled); !errors.Is(err, ErrCancelled) {
		t.Fatalf("constructor cancellation: %v", err)
	}
	system, err := NewAbsSystem(2, 1, []uint64{0}, next, bad, o)
	if err != nil {
		t.Fatal(err)
	}
	oneRefinement := o
	oneRefinement.MaxRefinements = 1
	manager, err := NewAbsManager(system, nil, oneRefinement)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.Refine([]int{0}); err != nil {
		t.Fatal(err)
	}
	before := manager.VisibleBits()
	if err := manager.Refine([]int{1}); !errors.Is(err, ErrLimit) || !slices.Equal(before, manager.VisibleBits()) {
		t.Fatalf("refinement limit/atomicity: %v %v", err, manager.VisibleBits())
	}
	frameBound := o
	frameBound.MaxFrames = 1
	if _, err := AbsFramesToCounts([][]int{{0}, {0}}, 1, frameBound); !errors.Is(err, ErrDimension) {
		t.Fatalf("frame conversion bound: %v", err)
	}
	objectBound := o
	objectBound.MaxConcreteStates = 1
	if _, err := AbsCountsToFrames([]int{0, 0}, objectBound); !errors.Is(err, ErrLimit) {
		t.Fatalf("count object bound: %v", err)
	}
}

func TestAbsMalformedSystemsAndConversionLimits(t *testing.T) {
	o := absTestOptions()
	if _, err := NewAbsSystem(2, 1, nil, make([]uint64, 8), make([]bool, 4), o); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty initial: %v", err)
	}
	if _, err := NewAbsSystem(2, 1, []uint64{4}, make([]uint64, 8), make([]bool, 4), o); !errors.Is(err, ErrInvariant) {
		t.Fatalf("bad initial: %v", err)
	}
	badNext := make([]uint64, 8)
	badNext[3] = 4
	if _, err := NewAbsSystem(2, 1, []uint64{0}, badNext, make([]bool, 4), o); !errors.Is(err, ErrInvariant) {
		t.Fatalf("bad next: %v", err)
	}
	if _, err := NewAbsSystem(2, 1, []uint64{0}, make([]uint64, 7), make([]bool, 4), o); !errors.Is(err, ErrDimension) {
		t.Fatalf("table shape: %v", err)
	}
	duplicateInitial := make([]uint64, 128)
	boundedInitial, err := NewAbsSystem(2, 1, duplicateInitial, make([]uint64, 8), make([]bool, 4), o)
	if err != nil || len(boundedInitial.Initial) != 1 || cap(boundedInitial.Initial) > 4 {
		t.Fatalf("canonical initial allocation: %+v %v", boundedInitial, err)
	}
	tooSmall := o
	tooSmall.MaxBytes = 4
	if _, err := NewAbsSystem(2, 1, []uint64{0}, make([]uint64, 8), make([]bool, 4), tooSmall); !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: %v", err)
	}
	nodeLimit := o
	nodeLimit.MaxNodes = 3
	if _, err := NewAbsSystem(2, 1, []uint64{0}, make([]uint64, 8), make([]bool, 4), nodeLimit); !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit: %v", err)
	}
	clauseLimit := o
	clauseLimit.MaxClauses = 7
	if _, err := NewAbsSystem(2, 1, []uint64{0}, make([]uint64, 8), make([]bool, 4), clauseLimit); !errors.Is(err, ErrLimit) {
		t.Fatalf("clause limit: %v", err)
	}
	if _, err := AbsFramesToCounts([][]int{{3}}, 3, o); !errors.Is(err, ErrDimension) {
		t.Fatalf("frame object: %v", err)
	}
	countsLimit := o
	countsLimit.MaxFrames = 1
	if _, err := AbsCountsToFrames([]int{2}, countsLimit); !errors.Is(err, ErrLimit) {
		t.Fatalf("count frame limit: %v", err)
	}
	cancel := o
	cancel.Stop = func() bool { return true }
	if _, err := AbsFramesToCounts([][]int{{0}}, 1, cancel); !errors.Is(err, ErrCancelled) {
		t.Fatalf("conversion cancellation: %v", err)
	}
}

func TestAbsAPILifecycleAndValidation(t *testing.T) {
	if (*AbsSystem)(nil).Clone() != nil {
		t.Fatal("nil clone")
	}
	if (*AbsManager)(nil).Phase() != AbsClosed || (*AbsManager)(nil).VisibleBits() != nil || (*AbsManager)(nil).Stats() != (AbsStats{}) {
		t.Fatal("nil manager accessors")
	}
	system := absSystemFromFunction(t, 2, 1, []uint64{0, 0}, func(state uint64) bool { return state == 3 }, func(state, input uint64) uint64 { return (state + input) & 3 })
	clone := system.Clone()
	clone.Initial[0] = 2
	if system.Initial[0] != 0 {
		t.Fatal("system clone aliases")
	}
	if next, err := system.NextState(1, 1); err != nil || next != 2 {
		t.Fatalf("next state: %d %v", next, err)
	}
	if _, err := system.NextState(4, 0); !errors.Is(err, ErrDimension) {
		t.Fatalf("invalid state: %v", err)
	}
	if _, err := (*AbsSystem)(nil).NextState(0, 0); !errors.Is(err, ErrInvariant) {
		t.Fatalf("nil system: %v", err)
	}
	if _, err := NewAbsManager(nil, nil, absTestOptions()); !errors.Is(err, ErrInvariant) {
		t.Fatalf("nil manager system: %v", err)
	}
	if _, err := NewAbsManager(system, []int{0, 0}, absTestOptions()); !errors.Is(err, ErrDimension) {
		t.Fatalf("duplicate visible bit: %v", err)
	}
	manager, err := NewAbsManager(system, []int{0, 1}, absTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	if manager.Stats().Nodes != 4 || !slices.Equal(manager.VisibleBits(), []int{0, 1}) {
		t.Fatalf("manager accessors: %+v %v", manager.Stats(), manager.VisibleBits())
	}
	if valid, err := manager.Verify(AbsResult{Status: AbsUnknown}); err != nil || valid {
		t.Fatalf("unknown verified: %v %v", valid, err)
	}
	if valid, err := manager.Verify(AbsResult{Status: AbsCounterexample}); !errors.Is(err, ErrDimension) || valid {
		t.Fatalf("missing trace verified: %v %v", valid, err)
	}
	manager.Close()
	if manager.VisibleBits() != nil {
		t.Fatal("closed visible bits")
	}

	// Zero-valued options exercise bounded defaults without external state.
	minimal, err := NewAbsSystem(0, 0, []uint64{0}, []uint64{0}, []bool{false}, AbsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	minimalManager, err := NewAbsManager(minimal, nil, AbsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := minimalManager.Analyze()
	if err != nil || result.Status != AbsProved {
		t.Fatalf("zero-bit system: %+v %v", result, err)
	}
}

func TestAbsIndependentModel5000(t *testing.T) {
	rng := rand.New(rand.NewSource(0xab515))
	for trial := 0; trial < 5000; trial++ {
		stateBits := 1 + rng.Intn(4)
		inputBits := rng.Intn(3)
		states, inputs := 1<<stateBits, 1<<inputBits
		next := make([]uint64, states*inputs)
		for i := range next {
			next[i] = uint64(rng.Intn(states))
		}
		bad := make([]bool, states)
		for i := range bad {
			bad[i] = rng.Intn(7) == 0
		}
		initial := []uint64{uint64(rng.Intn(states))}
		if rng.Intn(4) == 0 {
			initial = append(initial, uint64(rng.Intn(states)))
		}
		system, err := NewAbsSystem(stateBits, inputBits, initial, next, bad, absTestOptions())
		if err != nil {
			t.Fatalf("trial %d system: %v", trial, err)
		}
		wantBad, _, _ := absExplicitReach(system)
		manager, err := NewAbsManager(system, nil, absTestOptions())
		if err != nil {
			t.Fatal(err)
		}
		result, err := manager.Analyze()
		if err != nil {
			t.Fatalf("trial %d analyze: %v", trial, err)
		}
		if (result.Status == AbsCounterexample) != wantBad || (result.Status == AbsProved) == wantBad {
			t.Fatalf("trial %d status %v want bad %v", trial, result.Status, wantBad)
		}
		if wantBad {
			if result.Trace == nil {
				t.Fatalf("trial %d missing trace", trial)
			}
			trace, real, err := manager.Replay(result.Trace.Clone())
			if err != nil || !real || len(trace.States) != len(trace.Inputs)+1 || !system.Bad[trace.States[len(trace.States)-1]] {
				t.Fatalf("trial %d replay: %+v %v %v", trial, trace, real, err)
			}
		}
	}
}

func FuzzAbsAbstractionRefinement(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 7, 15, 31, 0x12345678} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		rng := rand.New(rand.NewSource(int64(seed)))
		stateBits := 1 + rng.Intn(4)
		inputBits := rng.Intn(3)
		states, inputs := 1<<stateBits, 1<<inputBits
		next := make([]uint64, states*inputs)
		for i := range next {
			next[i] = uint64(rng.Intn(states))
		}
		bad := make([]bool, states)
		for i := range bad {
			bad[i] = rng.Intn(5) == 0
		}
		system, err := NewAbsSystem(stateBits, inputBits, []uint64{uint64(rng.Intn(states))}, next, bad, absTestOptions())
		if err != nil {
			t.Fatal(err)
		}
		want, _, _ := absExplicitReach(system)
		manager, err := NewAbsManager(system, nil, absTestOptions())
		if err != nil {
			t.Fatal(err)
		}
		result, err := manager.Analyze()
		if err != nil || (result.Status == AbsCounterexample) != want {
			t.Fatalf("status=%v want=%v err=%v", result.Status, want, err)
		}
	})
}
