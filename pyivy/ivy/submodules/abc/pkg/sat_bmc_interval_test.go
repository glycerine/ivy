package abc

// Go-only exhaustive and fuzz tests for src/sat/bmc/bmcBmc2.c.

import (
	"errors"
	"reflect"
	"testing"
)

func aigIntervalTruthGraph(t testing.TB, output0, output1, next uint8) *AigManager {
	t.Helper()
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	register, _ := graph.CreateCI()
	variables := []AigLit{input, register}
	build := func(table uint8) AigLit {
		root := AigConst0
		for assignment := 0; assignment < 4; assignment++ {
			if table>>assignment&1 == 0 {
				continue
			}
			cube := AigConst1
			for variable, literal := range variables {
				var err error
				cube, err = graph.And(cube, literal.NotIf(assignment>>variable&1 == 0))
				if err != nil {
					t.Fatal(err)
				}
			}
			var err error
			root, err = graph.And(root.Not(), cube.Not())
			if err != nil {
				t.Fatal(err)
			}
			root = root.Not()
		}
		return root
	}
	graph.CreateCO(build(output0))
	graph.CreateCO(build(output1))
	graph.CreateCO(build(next))
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return graph
}

func intervalReference(output0, output1, next uint8, start, frames int) (int, int, bool) {
	eval := func(table uint8, input, state int) int { return int(table >> (input | state<<1) & 1) }
	outputs := [2]uint8{output0, output1}
	for frame := start; frame < frames; frame++ {
		for property, table := range outputs {
			for sequence := 0; sequence < 1<<(frame+1); sequence++ {
				state := 0
				for step := 0; step < frame; step++ {
					state = eval(next, sequence>>step&1, state)
				}
				if eval(table, sequence>>frame&1, state) != 0 {
					return frame, property, true
				}
			}
		}
	}
	return -1, -1, false
}

func ternaryPrefixReference(graph *AigManager, frames int) ([][]AigTernary, int, int) {
	state := make([]uint8, graph.RegisterCount())
	for register := range state {
		state[register] = 1 // possible zero
	}
	not := func(mask uint8) uint8 { return mask>>1&1 | mask<<1&2 }
	and := func(left, right uint8) uint8 {
		var result uint8
		for a := 0; a < 2; a++ {
			if left>>a&1 == 0 {
				continue
			}
			for b := 0; b < 2; b++ {
				if right>>b&1 != 0 {
					result |= 1 << (a & b)
				}
			}
		}
		return result
	}
	all := make([][]AigTernary, 0, frames)
	for frame := 0; frame < frames; frame++ {
		masks := make([]uint8, graph.ObjectCount())
		masks[0] = 2 // possible one
		for _, id := range graph.cis[:graph.PrimaryInputCount()] {
			masks[id] = 3
		}
		for register, id := range graph.cis[graph.PrimaryInputCount():] {
			masks[id] = state[register]
		}
		read := func(literal AigLit) uint8 {
			mask := masks[literal.ID()]
			if literal.Complemented() {
				mask = not(mask)
			}
			return mask
		}
		for id, object := range graph.objects {
			switch object.Type {
			case AigAND:
				masks[id] = and(read(object.Fanin0), read(object.Fanin1))
			case AigBuffer, AigCO:
				masks[id] = read(object.Fanin0)
			}
		}
		values := make([]AigTernary, len(masks))
		for id, mask := range masks {
			values[id] = AigTernary(mask)
		}
		all = append(all, values)
		for output, id := range graph.cos[:graph.PrimaryOutputCount()] {
			if masks[id] != 1 {
				return all, frame, output
			}
		}
		for register, id := range graph.cos[graph.PrimaryOutputCount():] {
			state[register] = masks[id]
		}
	}
	return all, -1, -1
}

func TestAigSequentialTernaryPrefix(t *testing.T) {
	graph := aigIntervalTruthGraph(t, 0x8, 0, 0xf) // r&i, 0, next=1
	before := graph.Clone()
	result, err := graph.SequentialTernaryPrefix(AigTernaryPrefixOptions{Frames: 4})
	if err != nil || result.Complete || result.FirstNonzeroFrame != 1 || result.FirstNonzeroOutput != 0 || len(result.Values) != 2 {
		t.Fatal(result, err)
	}
	first, second := result.Values[0][graph.cos[0]], result.Values[1][graph.cos[0]]
	if first != AigTernaryZero || second != AigTernaryUnknown {
		t.Fatal(first, second)
	}
	result.Values[0][0] = AigTernaryZero
	if !reflect.DeepEqual(graph, before) {
		t.Fatal("source mutated or result aliases source")
	}
	safe := aigIntervalTruthGraph(t, 0, 0, 0xa)
	complete, err := safe.SequentialTernaryPrefix(AigTernaryPrefixOptions{Frames: 3})
	if err != nil || !complete.Complete || len(complete.Values) != 3 || complete.FirstNonzeroFrame != -1 {
		t.Fatal(complete, err)
	}
}

func TestAigIntervalBoundedCheckOrderStartAndReplay(t *testing.T) {
	graph := aigIntervalTruthGraph(t, 0xc, 0xa, 0xa) // r, i, next=i
	if err := graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	before := graph.Clone()
	first, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{Frames: 3, NodeBatch: 1})
	if err != nil || first.Status != BmcCounterexample || first.Frame != 0 || first.Property != 1 || first.CheckedTargets != 2 || first.Counterexample == nil {
		t.Fatal("first", first, err)
	}
	if err = graph.ReplayCounterexample(first.Counterexample); err != nil {
		t.Fatal("replay", err)
	}
	started, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{StartFrame: 1, Frames: 3})
	if err != nil || started.Status != BmcCounterexample || started.Frame != 1 || started.Property != 0 || started.SkippedTargets != 2 || started.CheckedTargets != 1 {
		t.Fatal("started", started, err)
	}
	if err = graph.ReplayCounterexample(started.Counterexample); err != nil {
		t.Fatal("started replay", err)
	}
	if !reflect.DeepEqual(graph, before) {
		t.Fatal("source graph or constraint role mutated")
	}
}

func TestAigIntervalBoundedCheckSafeCallbacksAndOwnership(t *testing.T) {
	graph := aigIntervalTruthGraph(t, 0, 0, 0xa)
	type event struct {
		frame, property int
		status          SatStatus
	}
	events := []event{}
	result, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{Frames: 2, OnTargetDone: func(frame, property int, status SatStatus) {
		events = append(events, event{frame, property, status})
	}})
	if err != nil || result.Status != BmcNoCounterexample || result.Reason != BmcBoundReached || result.CheckedFrames != 2 || result.CheckedTargets != 4 || result.Counterexample != nil {
		t.Fatal(result, err)
	}
	want := []event{{0, 0, SatUnsatisfiable}, {0, 1, SatUnsatisfiable}, {1, 0, SatUnsatisfiable}, {1, 1, SatUnsatisfiable}}
	if !reflect.DeepEqual(events, want) {
		t.Fatal(events)
	}
	skipped, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{StartFrame: 2, Frames: 2})
	if err != nil || skipped.Status != BmcNoCounterexample || skipped.CheckedTargets != 0 || skipped.SkippedTargets != 4 || skipped.CheckedFrames != 0 {
		t.Fatal(skipped, err)
	}
}

func TestAigIntervalErrorsLimitsUnknownAndCancellation(t *testing.T) {
	graph := aigIntervalTruthGraph(t, 0xc, 0xa, 0xa)
	for _, options := range []AigIntervalBMCOptions{
		{},
		{Frames: 1, StartFrame: 2},
		{Frames: 1, NodeBatch: -1},
		{Frames: 1, MaxObjects: -1},
	} {
		if _, err := graph.IntervalBoundedCheck(options); !errors.Is(err, ErrDimension) {
			t.Fatal(options, err)
		}
	}
	limitGraph := aigIntervalTruthGraph(t, 0x8, 0, 0xa)
	for _, options := range []AigIntervalBMCOptions{
		{Frames: 2, MaxObjects: 1},
		{Frames: 2, MaxMappingEntries: 1},
		{Frames: 2, MaxVariables: 1},
		{Frames: 2, MaxClauses: 1},
		{Frames: 2, MaxLiterals: 1},
	} {
		result, err := limitGraph.IntervalBoundedCheck(options)
		if err != nil || result.Status != BmcUnknown || result.Reason != BmcResourceLimit {
			t.Fatal(options, result, err)
		}
	}
	cancelled, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || cancelled.Status != BmcUnknown || cancelled.Reason != BmcCancelled || cancelled.SatReason != SatCancelled {
		t.Fatal("cancelled", cancelled, err)
	}
	unknown, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{Frames: 1, SAT: SatLimits{Propagations: 1}})
	if err != nil || unknown.Status != BmcUnknown || unknown.Reason != BmcSatLimit || unknown.SatReason != SatPropagationLimit {
		t.Fatal("unknown", unknown, err)
	}
	if _, err := (*AigManager)(nil).IntervalBoundedCheck(AigIntervalBMCOptions{Frames: 1}); !errors.Is(err, ErrDimension) {
		t.Fatal("nil", err)
	}
	noOutputs := NewAigManager()
	noOutputs.CreateCI()
	if _, err := noOutputs.IntervalBoundedCheck(AigIntervalBMCOptions{Frames: 1}); !errors.Is(err, ErrDimension) {
		t.Fatal("no outputs", err)
	}
	for _, options := range []AigTernaryPrefixOptions{{}, {Frames: 1, MaxObjects: -1}, {Frames: 1, MaxEntries: -1}, {Frames: 1, MaxObjects: 1}, {Frames: 1, MaxEntries: 1}, {Frames: 1, Stop: func() bool { return true }}} {
		_, err := graph.SequentialTernaryPrefix(options)
		if err == nil {
			t.Fatal("ternary option accepted", options)
		}
	}
}

func TestAigIntervalGlobalConflictLimit(t *testing.T) {
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	or := func(left, right AigLit) AigLit {
		bothFalse, err := graph.And(left.Not(), right.Not())
		if err != nil {
			t.Fatal(err)
		}
		return bothFalse.Not()
	}
	clauses := []AigLit{or(a, b), or(a, b.Not()), or(a.Not(), b), or(a.Not(), b.Not())}
	root := AigConst1
	for _, clause := range clauses {
		var err error
		root, err = graph.And(root, clause)
		if err != nil {
			t.Fatal(err)
		}
	}
	graph.CreateCO(root)
	result, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{Frames: 1, GlobalConflicts: 1})
	if err != nil || result.Status != BmcUnknown || result.Reason != BmcSatLimit || result.SatReason != SatConflictLimit {
		t.Fatal(result, err)
	}
}

func FuzzAigSequentialTernaryPrefix(f *testing.F) {
	for _, seed := range [][4]byte{{0, 0, 0, 1}, {0x8, 0, 0xf, 4}, {0xa, 0xc, 0x6, 3}, {0xf, 0, 0, 2}} {
		f.Add(seed[0], seed[1], seed[2], seed[3])
	}
	f.Fuzz(func(t *testing.T, output0, output1, next, framesByte byte) {
		output0, output1, next = output0&0xf, output1&0xf, next&0xf
		frames := 1 + int(framesByte%4)
		graph := aigIntervalTruthGraph(t, output0, output1, next)
		result, err := graph.SequentialTernaryPrefix(AigTernaryPrefixOptions{Frames: frames})
		if err != nil {
			t.Fatal(err)
		}
		wantValues, wantFrame, wantOutput := ternaryPrefixReference(graph, frames)
		if len(result.Values) != len(wantValues) {
			t.Fatal(output0, output1, next, frames, len(result.Values), len(wantValues))
		}
		for frame := range wantValues {
			if !reflect.DeepEqual(result.Values[frame], wantValues[frame]) {
				t.Fatal(output0, output1, next, frame, result.Values[frame], wantValues[frame])
			}
		}
		wantComplete := wantFrame < 0
		if result.FirstNonzeroFrame != wantFrame || result.FirstNonzeroOutput != wantOutput || result.Complete != wantComplete {
			t.Fatal(output0, output1, next, frames, result, wantFrame, wantOutput, wantComplete)
		}
	})
}

func TestAigIntervalDeepIterativeCone(t *testing.T) {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	register, _ := graph.CreateCI()
	root := register
	for index := 0; index < 10_000; index++ {
		var err error
		root, err = graph.And(root, input.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	graph.CreateCO(root)
	graph.CreateCO(register)
	graph.SetRegisterCount(1)
	ternary, err := graph.SequentialTernaryPrefix(AigTernaryPrefixOptions{Frames: 1, MaxObjects: graph.ObjectCount()})
	if err != nil || !ternary.Complete {
		t.Fatal(ternary, err)
	}
	result, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{Frames: 1})
	if err != nil || result.Status != BmcNoCounterexample {
		t.Fatal(result, err)
	}
}

func FuzzAigIntervalBoundedCheck(f *testing.F) {
	for _, seed := range [][5]byte{{0, 0, 0, 0, 1}, {0xc, 0xa, 0xa, 0, 2}, {0x8, 0x6, 0x9, 1, 2}, {0xf, 0, 0, 0, 1}} {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4])
	}
	f.Fuzz(func(t *testing.T, output0, output1, next, startByte, framesByte byte) {
		output0, output1, next = output0&0xf, output1&0xf, next&0xf
		frames := 1 + int(framesByte&1)
		start := int(startByte) % (frames + 1)
		graph := aigIntervalTruthGraph(t, output0, output1, next)
		result, err := graph.IntervalBoundedCheck(AigIntervalBMCOptions{StartFrame: start, Frames: frames})
		if err != nil {
			t.Fatal(err)
		}
		wantFrame, wantProperty, found := intervalReference(output0, output1, next, start, frames)
		if found {
			if result.Status != BmcCounterexample || result.Frame != wantFrame || result.Property != wantProperty || result.Counterexample == nil {
				t.Fatal(output0, output1, next, start, frames, result, wantFrame, wantProperty)
			}
			if err = graph.ReplayCounterexample(result.Counterexample); err != nil {
				t.Fatal("replay", err)
			}
		} else if result.Status != BmcNoCounterexample || result.Counterexample != nil {
			t.Fatal(output0, output1, next, start, frames, result)
		}
	})
}
