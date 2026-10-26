package abc

// Stand-alone tests for src/opt/cgt.

import (
	"errors"
	"slices"
	"testing"
)

func cgtSequentialFixture(t testing.TB, variant byte) (*AigManager, AigLit, AigLit) {
	t.Helper()
	graph := NewAigManager()
	enable, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	data, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	state, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	update := data
	switch variant % 4 {
	case 1:
		update, err = graph.Xor(data, state)
	case 2:
		update, err = graph.And(data, state.Not())
	case 3:
		update, err = graph.Or(data, state)
	}
	if err != nil {
		t.Fatal(err)
	}
	next, err := graph.Mux(enable, update, state)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(next); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = graph.Check(); err != nil {
		t.Fatal(err)
	}
	return graph, enable, state
}

func TestClockGatingDefaultsAndManagerOwnership(t *testing.T) {
	parameters := DefaultClockGatingParameters()
	if parameters.LevelMax != 25 || parameters.CandidateMax != 1000 || parameters.ODCMax != 0 || parameters.ConflictMax != 10 || parameters.VariablesMin != 1000 || parameters.FlopsMin != 10 {
		t.Fatalf("defaults %+v", parameters)
	}
	graph, enable, _ := cgtSequentialFixture(t, 0)
	manager, err := NewClockGatingManager(graph, nil, parameters, ClockGatingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	useful := make([]bool, graph.ObjectCount())
	useful[enable.ID()] = true
	if err = manager.SetUseful(useful); err != nil {
		t.Fatal(err)
	}
	manager.candidates[0] = []AigLit{enable}
	useful[enable.ID()] = false
	if !manager.useful[enable.ID()] {
		t.Fatal("manager aliases useful mask")
	}
	if err = manager.Clean(); err != nil {
		t.Fatal(err)
	}
	stats, err := manager.Stats()
	stored, storedErr := manager.Candidates()
	if err != nil || storedErr != nil || stats.Recycles != 0 || !slices.Equal(stored[0], []AigLit{enable}) {
		t.Fatal(stats, stored, err, storedErr)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Gates(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestClockGatingTraversalAndMiter(t *testing.T) {
	graph, enable, _ := cgtSequentialFixture(t, 1)
	driver := graph.objects[graph.cos[0]].Fanin0
	candidates, err := DetectClockGateCandidates(graph, driver, nil, 100, 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(candidates, enable) {
		t.Fatalf("missing enable in %v", candidates)
	}
	positions := make(map[AigID]int)
	for i, literal := range candidates {
		positions[literal.ID()] = i
		object := graph.objects[literal.ID()]
		if object.Type == AigAND {
			if positions[object.Fanin0.ID()] >= i || positions[object.Fanin1.ID()] >= i {
				t.Fatalf("not postorder: %v", candidates)
			}
		}
	}
	visited, err := CollectClockGateVisited(graph, []AigID{driver.ID()}, nil)
	if err != nil || len(visited) == 0 || visited[len(visited)-1] != driver.ID() {
		t.Fatal(visited, err)
	}
	miter, err := DeriveClockGateMiterAIG(graph, ClockGatingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if miter.InputCount() != graph.InputCount() || miter.OutputCount() != 1 || miter.RegisterCount() != 0 {
		t.Fatalf("miter terminals %d %d %d", miter.InputCount(), miter.OutputCount(), miter.RegisterCount())
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		outputs, evaluateErr := miter.Evaluate(inputs)
		if evaluateErr != nil {
			t.Fatal(evaluateErr)
		}
		want := inputs[0] && inputs[1]
		if outputs[0] != want {
			t.Fatalf("inputs %v: got %v want %v", inputs, outputs[0], want)
		}
	}
}

func TestClockGatingCandidateDecisionAndDerivation(t *testing.T) {
	graph, enable, _ := cgtSequentialFixture(t, 3)
	manager, err := NewClockGatingManager(graph, nil, DefaultClockGatingParameters(), ClockGatingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	useful := make([]bool, graph.ObjectCount())
	useful[enable.ID()] = true
	if err = manager.SetUseful(useful); err != nil {
		t.Fatal(err)
	}
	candidates, err := manager.FindCandidates()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(candidates[0], enable.Not()) {
		t.Fatalf("expected !enable candidate: %v", candidates)
	}
	gates, err := manager.DecideSimple()
	if err != nil {
		t.Fatal(err)
	}
	if len(gates[0]) != 1 || gates[0][0] != enable.Not() {
		t.Fatalf("selected %v", gates)
	}
	before := graph.ObjectCount()
	gated, used, err := DeriveClockGatedAIG(graph, gates, ClockGatingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if graph.ObjectCount() != before || used < 0 || gated.RegisterCount() != 1 {
		t.Fatal(before, graph.ObjectCount(), used, gated.RegisterCount())
	}
	for assignment := 0; assignment < 8; assignment++ {
		primary := []bool{assignment&1 != 0, assignment&2 != 0}
		state := []bool{assignment&4 != 0}
		_, originalNext, originalErr := graph.Step(primary, state)
		_, gatedNext, gatedErr := gated.Step(primary, state)
		if originalErr != nil || gatedErr != nil || !slices.Equal(originalNext, gatedNext) {
			t.Fatalf("%v: original=%v gated=%v errors=%v/%v", assignment, originalNext, gatedNext, originalErr, gatedErr)
		}
	}
}

func TestClockGatingCareODCAndAreaDecision(t *testing.T) {
	graph, enable, _ := cgtSequentialFixture(t, 0)
	complete, err := CompleteClockGates(graph, [][]AigLit{{enable.Not()}}, nil)
	if err != nil || !slices.Equal(complete, []AigLit{enable.Not()}) {
		t.Fatalf("complete gates %v %v", complete, err)
	}
	manager, err := NewClockGatingManager(graph, nil, DefaultClockGatingParameters(), ClockGatingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	manager.candidates = [][]AigLit{{enable.Not()}}
	area, err := manager.DecideArea()
	if err != nil || len(area[0]) != 1 || area[0][0] != enable.Not() {
		t.Fatalf("area decision %v %v", area, err)
	}
	_ = manager.Close()

	care := NewAigManager()
	for i := 0; i < graph.InputCount(); i++ {
		if _, err = care.CreateCI(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = care.CreateCO(AigConst0); err != nil {
		t.Fatal(err)
	}
	careManager, err := NewClockGatingManager(graph, care, DefaultClockGatingParameters(), ClockGatingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	useful := make([]bool, graph.ObjectCount())
	useful[enable.ID()] = true
	if err = careManager.SetUseful(useful); err != nil {
		t.Fatal(err)
	}
	careCandidates, err := careManager.FindCandidates()
	if err != nil || !slices.Equal(careCandidates[0], []AigLit{enable}) {
		t.Fatalf("unsatisfiable care candidates %v %v", careCandidates, err)
	}
	careStats, err := careManager.Stats()
	if err != nil || careStats.Calls != 1 || careStats.Unsatisfiable != 1 {
		t.Fatal(careStats, err)
	}
	_ = careManager.Close()

	odcParameters := DefaultClockGatingParameters()
	odcParameters.ODCMax = 1
	odcManager, err := NewClockGatingManager(graph, nil, odcParameters, ClockGatingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if err = odcManager.SetUseful(useful); err != nil {
		t.Fatal(err)
	}
	if _, err = odcManager.FindCandidates(); err != nil {
		t.Fatal(err)
	}
	_ = odcManager.Close()
}

func TestClockGatingSecondFrameODCAndReducedDerivation(t *testing.T) {
	graph := NewAigManager()
	primary, _ := graph.CreateCI()
	firstState, _ := graph.CreateCI()
	secondState, _ := graph.CreateCI()
	observed, err := graph.And(firstState, secondState)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(observed); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(AigConst0); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(primary); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	work := graph.Clone()
	careLiteral, err := cgtODCCare(work, graph, firstState.ID(), 100, ClockGatingLimits{MaxObjects: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = work.CreateCO(careLiteral); err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		outputs, evaluateErr := work.Evaluate(inputs)
		if evaluateErr != nil || outputs[len(outputs)-1] != inputs[0] {
			t.Fatalf("assignment=%d inputs=%v outputs=%v err=%v", assignment, inputs, outputs, evaluateErr)
		}
	}

	reduce := NewAigManager()
	a, _ := reduce.CreateCI()
	b, _ := reduce.CreateCI()
	data, _ := reduce.CreateCI()
	state, _ := reduce.CreateCI()
	gate, err := reduce.And(a, b)
	if err != nil {
		t.Fatal(err)
	}
	next, err := reduce.Mux(gate, state, data)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = reduce.CreateCO(gate)
	_, _ = reduce.CreateCO(next)
	if err = reduce.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	reduced, _, err := DeriveClockGatedAIGReduced(reduce, [][]AigLit{{gate}}, ClockGatingLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		primaryInputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		current := []bool{assignment&8 != 0}
		_, wantNext, originalErr := reduce.Step(primaryInputs, current)
		outputs, gotNext, reducedErr := reduced.Step(primaryInputs, current)
		if originalErr != nil || reducedErr != nil || outputs[0] || !slices.Equal(gotNext, wantNext) {
			t.Fatalf("assignment=%d outputs=%v next=%v/%v errors=%v/%v", assignment, outputs, gotNext, wantNext, originalErr, reducedErr)
		}
	}
	if result, _, conflictErr := DeriveClockGatedAIGReduced(reduce, [][]AigLit{{gate, gate.Not()}}, ClockGatingLimits{}); result != nil || !errors.Is(conflictErr, ErrInvariant) {
		t.Fatal(result, conflictErr)
	}
}

func TestClockGatingInactiveCandidateThreshold(t *testing.T) {
	graph, _, _ := cgtSequentialFixture(t, 1)
	parameters := DefaultClockGatingParameters()
	parameters.CandidateMax = 1
	manager, err := NewClockGatingManager(graph, nil, parameters, ClockGatingLimits{MaxCandidates: 100})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	if _, err = manager.FindCandidates(); errors.Is(err, ErrLimit) {
		t.Fatalf("inactive source nCandMax was incorrectly enforced: %v", err)
	} else if err != nil {
		t.Fatal(err)
	}
}

func TestClockGatingLimitsCancellationAndAtomicity(t *testing.T) {
	graph, enable, _ := cgtSequentialFixture(t, 0)
	before := graph.Clone()
	if _, _, err := DeriveClockGatedAIG(graph, [][]AigLit{{enable.Not()}}, ClockGatingLimits{MaxObjects: 3}); !errors.Is(err, ErrLimit) {
		t.Fatalf("object limit %v", err)
	}
	if graph.ObjectCount() != before.ObjectCount() || graph.AndCount() != before.AndCount() {
		t.Fatal("failed transaction changed source")
	}
	calls := 0
	stop := func() bool { calls++; return calls == 3 }
	if _, _, err := DeriveClockGatedAIG(graph, [][]AigLit{{enable.Not()}}, ClockGatingLimits{Stop: stop}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel %v", err)
	}
	manager, err := NewClockGatingManager(graph, nil, DefaultClockGatingParameters(), ClockGatingLimits{MaxCandidates: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = manager.FindCandidates(); !errors.Is(err, ErrLimit) {
		t.Fatalf("candidate limit %v", err)
	}
	stored, _ := manager.Candidates()
	if len(stored[0]) != 0 {
		t.Fatal("failed candidate transaction committed")
	}
}

func cgtIndependentModel(data []byte) error {
	variant := byte(0)
	if len(data) != 0 {
		variant = data[0]
	}
	graph := NewAigManager()
	enable, err := graph.CreateCI()
	if err != nil {
		return err
	}
	dataInput, err := graph.CreateCI()
	if err != nil {
		return err
	}
	state, err := graph.CreateCI()
	if err != nil {
		return err
	}
	update := dataInput
	if variant&1 != 0 {
		update, err = graph.Xor(dataInput, state)
	} else if variant&2 != 0 {
		update, err = graph.And(dataInput.NotIf(variant&4 != 0), state.NotIf(variant&8 != 0))
	}
	if err != nil {
		return err
	}
	next, err := graph.Mux(enable, update, state)
	if err != nil {
		return err
	}
	if _, err = graph.CreateCO(next); err != nil {
		return err
	}
	if err = graph.SetRegisterCount(1); err != nil {
		return err
	}
	gated, _, err := DeriveClockGatedAIG(graph, [][]AigLit{{enable.Not()}}, ClockGatingLimits{MaxObjects: 1000})
	if err != nil {
		return err
	}
	for assignment := 0; assignment < 8; assignment++ {
		primary := []bool{assignment&1 != 0, assignment&2 != 0}
		current := []bool{assignment&4 != 0}
		_, want, e := graph.Step(primary, current)
		if e != nil {
			return e
		}
		_, got, e := gated.Step(primary, current)
		if e != nil {
			return e
		}
		if !slices.Equal(got, want) {
			return ErrInvariant
		}
	}
	area := NewAigManager()
	control, err := area.CreateCI()
	if err != nil {
		return err
	}
	mask, err := area.CreateCI()
	if err != nil {
		return err
	}
	areaData, err := area.CreateCI()
	if err != nil {
		return err
	}
	areaState, err := area.CreateCI()
	if err != nil {
		return err
	}
	internalGate, err := area.And(control, mask.NotIf(variant&16 != 0))
	if err != nil {
		return err
	}
	areaUpdate := areaData.NotIf(variant&32 != 0)
	areaNext, err := area.Mux(internalGate, areaState, areaUpdate)
	if err != nil {
		return err
	}
	if _, err = area.CreateCO(areaNext); err != nil {
		return err
	}
	if err = area.SetRegisterCount(1); err != nil {
		return err
	}
	areaGated, _, err := DeriveClockGatedAIGReduced(area, [][]AigLit{{internalGate}}, ClockGatingLimits{MaxObjects: 1000})
	if err != nil {
		return err
	}
	for assignment := 0; assignment < 16; assignment++ {
		primary := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		current := []bool{assignment&8 != 0}
		_, want, stepErr := area.Step(primary, current)
		if stepErr != nil {
			return stepErr
		}
		_, got, stepErr := areaGated.Step(primary, current)
		if stepErr != nil {
			return stepErr
		}
		if !slices.Equal(got, want) {
			return ErrInvariant
		}
	}
	return nil
}

func TestClockGatingIndependentModelCampaign(t *testing.T) {
	state := uint64(0xd1b54a32d192ed03)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		if err := cgtIndependentModel([]byte{byte(state)}); err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
	}
}

func FuzzClockGatingIndependentModel(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{1})
	f.Add([]byte{14})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			return
		}
		if err := cgtIndependentModel(data); err != nil {
			t.Fatal(err)
		}
	})
}
