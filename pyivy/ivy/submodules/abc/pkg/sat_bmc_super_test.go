package abc

// Go-only tent/rank and bounded-check tests for src/sat/bmc/bmcBmcS.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func giaTentFixture(t testing.TB) (*GiaManager, GiaID) {
	t.Helper()
	graph := NewGiaManager()
	if err := graph.HashStart(); err != nil {
		t.Fatal(err)
	}
	input, _ := graph.AppendCI()
	register0, _ := graph.AppendCI()
	register1, _ := graph.AppendCI()
	output, err := graph.HashAnd(input, register0)
	if err != nil {
		t.Fatal(err)
	}
	next0, err := graph.HashAnd(input, register1)
	if err != nil {
		t.Fatal(err)
	}
	graph.AppendCO(output)
	graph.AppendCO(next0)
	graph.AppendCO(input)
	graph.HashStop()
	if err = graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	return graph, input.ID()
}

func tentOracle(graph *GiaManager) (int, []int, []bool) {
	first := make([]int, graph.ObjectCount())
	for index := range first {
		first[index] = -1
	}
	first[0] = 0
	candidates := make([]bool, graph.ObjectCount())
	roToRI := make(map[GiaID]GiaID)
	for register, ro := range graph.cis[graph.PrimaryInputCount():] {
		roToRI[ro] = graph.cos[graph.PrimaryOutputCount()+register]
	}
	roots := []GiaID{}
	for _, po := range graph.cos[:graph.PrimaryOutputCount()] {
		driver, _ := graph.objects[po].Fanin0(po)
		roots = append(roots, driver.ID())
	}
	begin, tent := 0, 0
	var visit func(GiaID)
	visit = func(id GiaID) {
		if first[id] >= 0 {
			if first[id] < tent {
				candidates[id] = true
			}
			return
		}
		first[id] = tent
		object := graph.objects[id]
		if object.IsAnd() {
			left, _ := object.Fanin0(id)
			right, _ := object.Fanin1(id)
			visit(left.ID())
			visit(right.ID())
		} else if ri, ok := roToRI[id]; ok {
			driver, _ := graph.objects[ri].Fanin0(ri)
			roots = append(roots, driver.ID())
		}
	}
	for begin < len(roots) {
		end := len(roots)
		for _, root := range roots[begin:end] {
			visit(root)
		}
		begin, tent = end, tent+1
	}
	candidates[0] = false
	return tent, first, candidates
}

func TestGiaAnalyzeTentsExactOwnershipAndLimits(t *testing.T) {
	graph, input := giaTentFixture(t)
	before := graph.Clone()
	result, err := graph.AnalyzeTents(GiaRankedUnrollOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantTents, wantFirst, wantCandidates := tentOracle(graph)
	if result.Tents != wantTents || result.Tents != 3 || !slices.Equal(result.Candidates, []GiaID{input}) || result.MaxRank < 1 || result.Reachable == 0 {
		t.Fatal(result, wantTents)
	}
	for id, tent := range wantFirst {
		if tent >= 0 && !graph.objects[id].IsCO() && result.TentByObject[id] != tent {
			t.Fatal("tent", id, result.TentByObject[id], tent)
		}
		if wantCandidates[id] != slices.Contains(result.Candidates, GiaID(id)) {
			t.Fatal("candidate", id)
		}
	}
	if len(result.Boundaries) == 0 || !result.Boundaries[0].PrimaryInput || !result.Boundaries[0].Candidate {
		t.Fatal("boundary metadata", result.Boundaries)
	}
	result.TentByObject[input] = 99
	result.Candidates[0] = InvalidGiaID
	if !reflect.DeepEqual(graph, before) {
		t.Fatal("analysis aliases or mutates source")
	}
	for _, options := range []GiaRankedUnrollOptions{{MaxRank: 1}, {MaxMappingEntries: 1}, {Stop: func() bool { return true }}} {
		if _, err = graph.AnalyzeTents(options); err == nil {
			t.Fatal("limit accepted", options)
		}
	}
	withoutOutputs := NewGiaManager()
	withoutOutputs.AppendCI()
	if _, err = withoutOutputs.AnalyzeTents(GiaRankedUnrollOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("no outputs", err)
	}
	extended := NewGiaManager()
	a, _ := extended.AppendCI()
	b, _ := extended.AppendCI()
	xor, _ := extended.AppendXorReal(a, b)
	extended.AppendCO(xor)
	if _, err = extended.AnalyzeTents(GiaRankedUnrollOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended", err)
	}
}

func TestGiaSuperBoundedCheckOrderWorkersAndReplay(t *testing.T) {
	aig := aigIntervalTruthGraph(t, 0xc, 0xa, 0xa)
	aig.SetConstraintCount(1)
	graph, _, err := GiaFromAig(aig)
	if err != nil {
		t.Fatal(err)
	}
	before := graph.Clone()
	type event struct {
		frame, property int
		failed          bool
	}
	for _, workers := range []int{1, 4} {
		events := []event{}
		result, checkErr := graph.SuperBoundedCheck(GiaSuperBMCOptions{Frames: 3, FrameBatch: 2, Workers: workers, OnFrameDone: func(frame, property int, failed bool) {
			events = append(events, event{frame, property, failed})
		}})
		if checkErr != nil || result.Status != BmcCounterexample || result.Frame != 0 || result.Property != 1 || result.Counterexample == nil {
			t.Fatal(workers, result, checkErr)
		}
		if !reflect.DeepEqual(events, []event{{0, 0, false}, {0, 1, true}}) {
			t.Fatal(workers, events)
		}
		if err = aig.ReplayCounterexample(result.Counterexample); err != nil {
			t.Fatal("replay", err)
		}
	}
	if !reflect.DeepEqual(graph, before) {
		t.Fatal("source mutated")
	}
}

func TestGiaSuperBoundedCheckSafeLimitsAndCancellation(t *testing.T) {
	aig := aigIntervalTruthGraph(t, 0, 0, 0xa)
	graph, _, _ := GiaFromAig(aig)
	result, err := graph.SuperBoundedCheck(GiaSuperBMCOptions{Frames: 2, Workers: 99})
	if err != nil || result.Status != BmcNoCounterexample || result.CheckedFrames != 2 {
		t.Fatal(result, err)
	}
	for _, options := range []GiaSuperBMCOptions{{}, {Frames: 1, FrameBatch: -1}, {Frames: 1, Workers: 100}, {Frames: 1, MaxObjects: -1}} {
		if _, err = graph.SuperBoundedCheck(options); !errors.Is(err, ErrDimension) {
			t.Fatal(options, err)
		}
	}
	for _, options := range []GiaSuperBMCOptions{{Frames: 2, MaxObjects: 1}, {Frames: 2, MaxMappingEntries: 1}, {Frames: 2, MaxVariables: 1}, {Frames: 2, MaxClauses: 1}, {Frames: 2, MaxLiterals: 1}} {
		limited, limitErr := graph.SuperBoundedCheck(options)
		if limitErr != nil || limited.Status != BmcUnknown || limited.Reason != BmcResourceLimit {
			t.Fatal(options, limited, limitErr)
		}
	}
	cancelled, err := graph.SuperBoundedCheck(GiaSuperBMCOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || cancelled.Status != BmcUnknown || cancelled.Reason != BmcCancelled {
		t.Fatal(cancelled, err)
	}
	unknownEvents := []bool{}
	unknownGraph, _, _ := GiaFromAig(aigIntervalTruthGraph(t, 0xc, 0xa, 0xa))
	unknown, err := unknownGraph.SuperBoundedCheck(GiaSuperBMCOptions{Frames: 1, SAT: SatLimits{Propagations: 1}, OnFrameDone: func(_ int, _ int, failed bool) { unknownEvents = append(unknownEvents, failed) }})
	if err != nil || unknown.Status != BmcUnknown || !slices.Equal(unknownEvents, []bool{false}) {
		t.Fatal("unknown callback", unknown, unknownEvents, err)
	}
	extended := NewGiaManager()
	pi, _ := extended.AppendCI()
	ro, _ := extended.AppendCI()
	xor, _ := extended.AppendXorReal(pi, ro)
	extended.AppendCO(xor)
	extended.AppendCO(ro)
	extended.SetRegisterCount(1)
	if _, err = extended.SuperBoundedCheck(GiaSuperBMCOptions{Frames: 1}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended", err)
	}
}

func TestGiaSuperBoundedCheckDeepIterativeCone(t *testing.T) {
	graph := NewGiaManager()
	if err := graph.HashStart(); err != nil {
		t.Fatal(err)
	}
	input, _ := graph.AppendCI()
	register, _ := graph.AppendCI()
	root := register
	for index := 0; index < 10_000; index++ {
		var err error
		root, err = graph.HashAnd(root, input.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	graph.AppendCO(root)
	graph.AppendCO(register)
	graph.HashStop()
	graph.SetRegisterCount(1)
	result, err := graph.SuperBoundedCheck(GiaSuperBMCOptions{Frames: 1})
	if err != nil || result.Status != BmcNoCounterexample {
		t.Fatal(result, err)
	}
}

func FuzzGiaAnalyzeTents(f *testing.F) {
	for _, seed := range [][]byte{{}, {0, 1, 2}, {3, 7, 9, 11, 2}, {255, 1, 254, 2}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 24 {
			return
		}
		graph := NewGiaManager()
		graph.HashStart()
		input, _ := graph.AppendCI()
		register0, _ := graph.AppendCI()
		register1, _ := graph.AppendCI()
		roots := []GiaLit{GiaConst0, GiaConst1, input, register0, register1}
		for index, value := range data {
			left := roots[int(value)%len(roots)].NotIf(value&0x40 != 0)
			right := roots[(index+int(value>>3))%len(roots)].NotIf(value&0x80 != 0)
			root, err := graph.HashAnd(left, right)
			if err != nil {
				t.Fatal(err)
			}
			roots = append(roots, root)
		}
		graph.AppendCO(roots[len(roots)-1])
		graph.AppendCO(roots[len(roots)/2])
		graph.AppendCO(roots[len(roots)/3])
		graph.HashStop()
		graph.SetRegisterCount(2)
		result, err := graph.AnalyzeTents(GiaRankedUnrollOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantTents, wantFirst, wantCandidates := tentOracle(graph)
		if result.Tents != wantTents {
			t.Fatal(result.Tents, wantTents)
		}
		for id := range wantFirst {
			if wantFirst[id] >= 0 && !graph.objects[id].IsCO() && result.TentByObject[id] != wantFirst[id] {
				t.Fatal("tent", id, result.TentByObject[id], wantFirst[id])
			}
			if wantCandidates[id] != slices.Contains(result.Candidates, GiaID(id)) {
				t.Fatal("candidate", id)
			}
		}
	})
}

func FuzzGiaSuperBoundedCheck(f *testing.F) {
	for _, seed := range [][6]byte{{0, 0, 0, 1, 1, 0}, {0xc, 0xa, 0xa, 2, 3, 1}, {0x8, 0x6, 0x9, 3, 2, 0}} {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4], seed[5])
	}
	f.Fuzz(func(t *testing.T, output0, output1, next, workersByte, framesByte, constraintByte byte) {
		output0, output1, next = output0&0xf, output1&0xf, next&0xf
		frames := 1 + int(framesByte&1)
		aig := aigIntervalTruthGraph(t, output0, output1, next)
		if constraintByte&1 != 0 {
			aig.SetConstraintCount(1)
		}
		graph, _, err := GiaFromAig(aig)
		if err != nil {
			t.Fatal(err)
		}
		result, err := graph.SuperBoundedCheck(GiaSuperBMCOptions{Frames: frames, FrameBatch: 1 + int(workersByte%3), Workers: 1 + int(workersByte%4)})
		if err != nil {
			t.Fatal(err)
		}
		wantFrame, wantProperty, found := intervalReference(output0, output1, next, 0, frames)
		if found {
			if result.Status != BmcCounterexample || result.Frame != wantFrame || result.Property != wantProperty || result.Counterexample == nil {
				t.Fatal(output0, output1, next, frames, result, wantFrame, wantProperty)
			}
			if err = aig.ReplayCounterexample(result.Counterexample); err != nil {
				t.Fatal("replay", err)
			}
		} else if result.Status != BmcNoCounterexample {
			t.Fatal(output0, output1, next, frames, result)
		}
	})
}
