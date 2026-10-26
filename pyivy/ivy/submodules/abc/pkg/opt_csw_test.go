package abc

// Go-only validation for src/opt/csw.

import (
	"errors"
	"slices"
	"testing"
)

func cswMustCI(t *testing.T, graph *AigManager) AigLit {
	t.Helper()
	literal, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func cswMustAnd(t *testing.T, graph *AigManager, first, second AigLit) AigLit {
	t.Helper()
	literal, err := graph.And(first, second)
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func cswMustCO(t *testing.T, graph *AigManager, driver AigLit) {
	t.Helper()
	if _, err := graph.CreateCO(driver); err != nil {
		t.Fatal(err)
	}
}

func cswScalarEvaluate(graph *AigManager, inputs []bool) ([]bool, error) {
	if graph == nil || len(inputs) != graph.InputCount() {
		return nil, ErrDimension
	}
	values := make([]bool, graph.ObjectCount())
	values[0] = true
	for index, id := range graph.Inputs() {
		values[id] = inputs[index]
	}
	read := func(literal AigLit) bool {
		value := values[literal.ID()]
		if literal.Complemented() {
			return !value
		}
		return value
	}
	for id := 1; id < graph.ObjectCount(); id++ {
		object, err := graph.Object(AigID(id))
		if errors.Is(err, ErrInvalidLiteral) {
			continue
		}
		if err != nil {
			return nil, err
		}
		switch object.Type {
		case AigCI:
		case AigAND:
			values[id] = read(object.Fanin0) && read(object.Fanin1)
		case AigBuffer, AigCO:
			values[id] = read(object.Fanin0)
		default:
			return nil, ErrUnsupported
		}
	}
	outputs := graph.Outputs()
	result := make([]bool, len(outputs))
	for index, id := range outputs {
		result[index] = values[id]
	}
	return result, nil
}

func cswEquivalent(t *testing.T, source, result *AigManager) {
	t.Helper()
	if source.InputCount() != result.InputCount() || source.OutputCount() != result.OutputCount() {
		t.Fatalf("terminal counts source=%d/%d result=%d/%d", source.InputCount(), source.OutputCount(), result.InputCount(), result.OutputCount())
	}
	if source.InputCount() > 12 {
		t.Fatal("test oracle input bound")
	}
	for assignment := 0; assignment < 1<<uint(source.InputCount()); assignment++ {
		inputs := make([]bool, source.InputCount())
		for index := range inputs {
			inputs[index] = assignment&(1<<uint(index)) != 0
		}
		want, wantErr := cswScalarEvaluate(source, inputs)
		got, gotErr := cswScalarEvaluate(result, inputs)
		if wantErr != nil || gotErr != nil || !slices.Equal(got, want) {
			t.Fatalf("assignment=%d got=%v want=%v errors=%v/%v", assignment, got, want, gotErr, wantErr)
		}
	}
}

func TestCswCutPrimitivesAndTruthOrdering(t *testing.T) {
	first := &cswCut{leaves: []AigID{1, 3, 7}}
	second := &cswCut{leaves: []AigID{2, 3, 8}}
	merged, okay := cswMergeLeaves(first.leaves, second.leaves, 5)
	if !okay || !slices.Equal(merged, []AigID{1, 2, 3, 7, 8}) {
		t.Fatal(merged, okay)
	}
	if merged, okay = cswMergeLeaves(first.leaves, second.leaves, 4); okay || merged != nil {
		t.Fatal(merged, okay)
	}
	if cswProjectAssignment([]AigID{1, 2, 3, 7}, first.leaves[:2], 0b1101) != 0b11 {
		t.Fatal("projection order")
	}
	cut := &cswCut{leaves: []AigID{4, 7, 9}, truth: []uint64{0}}
	for assignment := 0; assignment < 8; assignment++ {
		if assignment&1 != 0 && assignment&4 == 0 {
			cswTruthSet(cut.truth, assignment)
		}
	}
	if cswTruthSupport(cut.truth, 3) != 5 {
		t.Fatal("support")
	}
	cswMinimize(cut)
	if !slices.Equal(cut.leaves, []AigID{4, 9}) || len(cut.truth) != 1 || cut.truth[0]&15 != 2 {
		t.Fatal(cut.leaves, cut.truth)
	}
	if cswHash(&cswCut{leaves: []AigID{1, 2, 3}}) != uint32(1*1009^2*1049^3*1093) {
		t.Fatal("hash ordering")
	}
	graph := NewAigManager()
	left, right := cswMustCI(t, graph), cswMustCI(t, graph)
	_ = cswMustAnd(t, graph, left, right)
	manager := &cswManager{result: graph, table: make([]*cswCut, 3)}
	if cost, err := manager.cutCost2(&cswCut{leaves: []AigID{left.ID(), right.ID()}}); err != nil || cost != 0.5 {
		t.Fatal(cost, err)
	}
	manager.insert(&cswCut{leaves: []AigID{left.ID(), right.ID(), 3}})
	if manager.tableCount() != 1 {
		t.Fatal(manager.tableCount())
	}
}

func TestCswMergeAllocationSourceOrdering(t *testing.T) {
	// A signature rejection happens before allocation and preserves every
	// slot, even when all slots are occupied.
	kept := &cswCut{cost: 2, leaves: []AigID{7, 8}}
	worst := &cswCut{cost: 9, leaves: []AigID{9, 10}}
	slots := []*cswCut{kept, worst}
	first := &cswCut{leaves: []AigID{1, 2}, sign: cswCutSign(1) | cswCutSign(2)}
	second := &cswCut{leaves: []AigID{3, 4}, sign: cswCutSign(3) | cswCutSign(4)}
	if candidate, leaves, okay := cswAllocateMergedCut(slots, first, second, 3); candidate != nil || leaves != nil || okay {
		t.Fatal(candidate, leaves, okay)
	}
	if !slices.Equal(worst.leaves, []AigID{9, 10}) {
		t.Fatal("signature rejection evicted a cut")
	}

	// Signature collisions can pass the prefilter while the exact union is
	// infeasible. C allocates first, so this case evicts the worst cut.
	second = &cswCut{leaves: []AigID{33, 34}, sign: cswCutSign(33) | cswCutSign(34)}
	candidate, leaves, okay := cswAllocateMergedCut(slots, first, second, 3)
	if candidate != worst || leaves != nil || okay || worst.leaves != nil || worst.truth != nil {
		t.Fatalf("candidate=%p worst=%p leaves=%v okay=%v worst=%+v", candidate, worst, leaves, okay, worst)
	}
}

func TestCswIndependentCutModelCampaign(t *testing.T) {
	state := uint64(0x243f6a8885a308d3)
	next := func() uint64 {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		return state
	}
	for execution := 0; execution < 5000; execution++ {
		firstLeaves, secondLeaves := make([]AigID, 0, 8), make([]AigID, 0, 8)
		for id := AigID(0); id < 8; id++ {
			value := next()
			if value&1 != 0 {
				firstLeaves = append(firstLeaves, id)
			}
			if value&2 != 0 {
				secondLeaves = append(secondLeaves, id)
			}
		}
		if len(firstLeaves) == 0 {
			firstLeaves = append(firstLeaves, 0)
		}
		if len(secondLeaves) == 0 {
			secondLeaves = append(secondLeaves, 7)
		}
		maximum := 1 + int(next()%8)
		merged, okay := cswMergeLeaves(firstLeaves, secondLeaves, maximum)
		present := [8]bool{}
		for _, leaf := range firstLeaves {
			present[leaf] = true
		}
		for _, leaf := range secondLeaves {
			present[leaf] = true
		}
		wantMerged := make([]AigID, 0, 8)
		for leaf, exists := range present {
			if exists {
				wantMerged = append(wantMerged, AigID(leaf))
			}
		}
		wantOkay := len(wantMerged) <= maximum
		if okay != wantOkay || okay && !slices.Equal(merged, wantMerged) || !okay && merged != nil {
			t.Fatalf("execution=%d merge=%v/%v want=%v/%v", execution, merged, okay, wantMerged, wantOkay)
		}
		modelDominates := func(subset, superset []AigID) bool {
			for _, wanted := range subset {
				found := false
				for _, available := range superset {
					found = found || wanted == available
				}
				if !found {
					return false
				}
			}
			return true
		}
		if got, want := cswDominates(firstLeaves, secondLeaves), modelDominates(firstLeaves, secondLeaves); got != want {
			t.Fatalf("execution=%d dominance=%v want=%v", execution, got, want)
		}

		if !okay {
			continue
		}
		firstTruth := make([]uint64, max(1, (1<<uint(len(firstLeaves))+63)/64))
		secondTruth := make([]uint64, max(1, (1<<uint(len(secondLeaves))+63)/64))
		for i := range firstTruth {
			firstTruth[i] = next()
		}
		for i := range secondTruth {
			secondTruth[i] = next()
		}
		first := &cswCut{leaves: firstLeaves, truth: firstTruth}
		second := &cswCut{leaves: secondLeaves, truth: secondTruth}
		manager := &cswManager{options: CswOptions{MaxWork: 1 << 20, MaxBytes: 1 << 20}}
		complementFirst, complementSecond := next()&1 != 0, next()&1 != 0
		gotTruth, err := manager.mergeTruth(merged, first, second, complementFirst, complementSecond)
		if err != nil {
			t.Fatal(err)
		}
		var wantSupport uint32
		for assignment := 0; assignment < 1<<uint(len(merged)); assignment++ {
			project := func(leaves []AigID) int {
				value := 0
				for bit, leaf := range leaves {
					position := slices.Index(merged, leaf)
					if assignment&(1<<uint(position)) != 0 {
						value |= 1 << uint(bit)
					}
				}
				return value
			}
			left := cswTruthGet(firstTruth, project(firstLeaves))
			right := cswTruthGet(secondTruth, project(secondLeaves))
			want := (left != complementFirst) && (right != complementSecond)
			if cswTruthGet(gotTruth, assignment) != want {
				t.Fatalf("execution=%d assignment=%d truth mismatch", execution, assignment)
			}
		}
		for variable := range merged {
			bit := 1 << uint(variable)
			for assignment := 0; assignment < 1<<uint(len(merged)); assignment++ {
				if assignment&bit == 0 && cswTruthGet(gotTruth, assignment) != cswTruthGet(gotTruth, assignment|bit) {
					wantSupport |= 1 << uint(variable)
					break
				}
			}
		}
		if got := cswTruthSupport(gotTruth, len(merged)); got != wantSupport {
			t.Fatalf("execution=%d support=%x want=%x", execution, got, wantSupport)
		}
	}
}

func TestCswHashComplementAndReferenceCost(t *testing.T) {
	entry := &cswCut{node: 7, leaves: []AigID{1, 2, 3}, sign: cswCutSign(1) | cswCutSign(2) | cswCutSign(3), truth: []uint64{0x96}}
	manager := &cswManager{table: make([]*cswCut, 17), references: make([]uint64, 8)}
	manager.references[1], manager.references[2], manager.references[3] = 2, 5, 8
	manager.insert(entry)
	if got := manager.lookup(&cswCut{leaves: slices.Clone(entry.leaves), sign: entry.sign, truth: []uint64{0x96}}); got != AigLit(14) {
		t.Fatal("equal lookup", got)
	}
	if got := manager.lookup(&cswCut{leaves: slices.Clone(entry.leaves), sign: entry.sign, truth: []uint64{(^uint64(0x96)) & 0xff}}); got != AigLit(15) {
		t.Fatal("complement lookup", got)
	}
	if cost, err := manager.cutCost(entry); err != nil || cost != 500 {
		t.Fatal(cost, err)
	}
}

func TestCswSweepRewritesAndDeterminism(t *testing.T) {
	graph := NewAigManager()
	a, b, c := cswMustCI(t, graph), cswMustCI(t, graph), cswMustCI(t, graph)
	ab := cswMustAnd(t, graph, a, b)
	ac := cswMustAnd(t, graph, a, c)
	bc := cswMustAnd(t, graph, b, c)
	firstABC := cswMustAnd(t, graph, ac, bc)
	secondABC := cswMustAnd(t, graph, ab, c)
	notANotC := cswMustAnd(t, graph, a.Not(), c.Not())
	aOrC := notANotC.Not()
	redundantAB := cswMustAnd(t, graph, ab, aOrC)
	constant := cswMustAnd(t, graph, ab, cswMustAnd(t, graph, a, b.Not()))
	for _, output := range []AigLit{firstABC, secondABC, redundantAB, constant} {
		cswMustCO(t, graph, output)
	}
	beforeObjects, beforeAnds := graph.ObjectCount(), graph.AndCount()
	result, err := CswSweep(graph, CswOptions{CutsMax: 8, LeavesMax: 4})
	if err != nil {
		t.Fatal(err)
	}
	if graph.ObjectCount() != beforeObjects || graph.AndCount() != beforeAnds {
		t.Fatal("source mutated")
	}
	if result.Graph == graph || result.Stats.TrivialConstants == 0 || result.Stats.NodesRewritten+result.Stats.Trivial2 == 0 {
		t.Fatal(result.Stats)
	}
	if result.Graph.AndCount() >= graph.AndCount() {
		t.Fatalf("no reduction: %d >= %d", result.Graph.AndCount(), graph.AndCount())
	}
	cswEquivalent(t, graph, result.Graph)
	again, err := CswSweep(graph, CswOptions{CutsMax: 8, LeavesMax: 4})
	if err != nil || again.Stats != result.Stats || again.Graph.ObjectCount() != result.Graph.ObjectCount() {
		t.Fatal(err, result.Stats, again.Stats)
	}
	for id := 0; id < result.Graph.ObjectCount(); id++ {
		first, firstErr := result.Graph.Object(AigID(id))
		second, secondErr := again.Graph.Object(AigID(id))
		if !errors.Is(firstErr, secondErr) && firstErr != secondErr || first != second {
			t.Fatalf("object %d differs: %+v/%v %+v/%v", id, first, firstErr, second, secondErr)
		}
	}
}

func TestCswSequentialOrderingOwnershipAndErrors(t *testing.T) {
	graph := NewAigManager()
	input, state := cswMustCI(t, graph), cswMustCI(t, graph)
	next := cswMustAnd(t, graph, input, state.Not())
	cswMustCO(t, graph, next.Not())
	cswMustCO(t, graph, input)
	if err := graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err := graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	result, err := CswSweep(graph, CswOptions{})
	if err != nil || result.Graph.RegisterCount() != 1 || result.Graph.ConstraintCount() != 1 || result.Graph.Name != graph.Name || result.Graph.Spec != graph.Spec {
		t.Fatal(result, err)
	}
	cswEquivalent(t, graph, result.Graph)
	if _, err = result.Graph.CreateCI(); err != nil || graph.InputCount() != 2 || result.Graph.InputCount() != 3 {
		t.Fatal("result does not own its graph", err)
	}

	for _, row := range []struct {
		graph   *AigManager
		options CswOptions
		want    error
	}{
		{nil, CswOptions{}, ErrDimension},
		{graph, CswOptions{CutsMax: 1}, ErrDimension},
		{graph, CswOptions{LeavesMax: 17}, ErrDimension},
		{graph, CswOptions{MaxObjects: 1}, ErrLimit},
		{graph, CswOptions{MaxWork: 1}, ErrLimit},
		{graph, CswOptions{MaxBytes: 1}, ErrLimit},
		{graph, CswOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		got, sweepErr := CswSweep(row.graph, row.options)
		if got.Graph != nil || !errors.Is(sweepErr, row.want) {
			t.Fatalf("got=%v err=%v want=%v", got.Graph, sweepErr, row.want)
		}
	}
	buffered := NewAigManager()
	ci := cswMustCI(t, buffered)
	buffer, err := buffered.CreateBuffer(ci)
	if err != nil {
		t.Fatal(err)
	}
	cswMustCO(t, buffered, buffer)
	if got, sweepErr := CswSweep(buffered, CswOptions{}); got.Graph != nil || !errors.Is(sweepErr, ErrUnsupported) {
		t.Fatal(got.Graph, sweepErr)
	}
}

func cswGeneratedGraph(seed uint64) *AigManager {
	graph := NewAigManager()
	inputs := 1 + int(seed%4)
	literals := make([]AigLit, 0, inputs+12)
	for index := 0; index < inputs; index++ {
		literal, _ := graph.CreateCI()
		literals = append(literals, literal)
	}
	state := seed | 1
	for gate := 0; gate < 4+int(seed%9); gate++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		first := literals[int(state%uint64(len(literals)))]
		state = state*6364136223846793005 + 1442695040888963407
		second := literals[int(state%uint64(len(literals)))]
		first = first.NotIf(state&(1<<61) != 0)
		second = second.NotIf(state&(1<<62) != 0)
		literal, _ := graph.And(first, second)
		literals = append(literals, literal)
	}
	for output := 0; output < 1+int(seed%3); output++ {
		state = state*2862933555777941757 + 3037000493
		literal := literals[int(state%uint64(len(literals)))].NotIf(state>>63 != 0)
		_, _ = graph.CreateCO(literal)
	}
	return graph
}

func TestCswIndependentScalarCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state = state*6364136223846793005 + 1442695040888963407
		graph := cswGeneratedGraph(state)
		result, err := CswSweep(graph, CswOptions{CutsMax: 6, LeavesMax: 4, MaxWork: 1 << 20})
		if err != nil {
			t.Fatalf("execution %d seed %x: %v", execution, state, err)
		}
		cswEquivalent(t, graph, result.Graph)
	}
}

func FuzzCswSweep(f *testing.F) {
	f.Add(uint64(1), uint8(4), uint8(6))
	f.Add(uint64(0xdeadbeef), uint8(3), uint8(8))
	f.Fuzz(func(t *testing.T, seed uint64, rawLeaves, rawCuts uint8) {
		graph := cswGeneratedGraph(seed)
		options := CswOptions{LeavesMax: 2 + int(rawLeaves%5), CutsMax: 2 + int(rawCuts%9), MaxWork: 1 << 21}
		result, err := CswSweep(graph, options)
		if err != nil {
			t.Fatal(err)
		}
		cswEquivalent(t, graph, result.Graph)
	})
}
