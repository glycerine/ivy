package abc

// Standalone semantic, SAT, concurrency and fuzz tests for ivyFraig.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func ivyFraigFixture(t testing.TB) (*IvyManager, []IvyEdge, IvyEdge, IvyEdge) {
	t.Helper()
	m, err := NewIvyManager(IvyOptions{MaxObjects: 10000, MaxWork: 1000000})
	if err != nil {
		t.Fatal(err)
	}
	pis := make([]IvyEdge, 3)
	for i := range pis {
		pis[i], err = m.CreatePI()
		if err != nil {
			t.Fatal(err)
		}
	}
	ab, _ := m.And(pis[0], pis[1])
	left, _ := m.And(ab, pis[2])
	bc, _ := m.And(pis[1], pis[2])
	right, _ := m.And(pis[0], bc)
	if left == right {
		t.Fatal("fixture requires structurally distinct equivalent nodes")
	}
	notLeft, _ := m.CreateBuffer(left.Not())
	zero, _ := m.CreateBuffer(IvyConst0)
	for _, edge := range []IvyEdge{left, right, notLeft, zero} {
		if _, err = m.CreatePO(edge); err != nil {
			t.Fatal(err)
		}
	}
	if err = m.Check(); err != nil {
		t.Fatal(err)
	}
	return m, pis, left, right
}

func ivyFraigUnlimited(prove bool) IvyFraigOptions {
	p := DefaultIvyFraigParameters()
	p.SimulationWords = 4
	p.NodeBacktrackLimit = -1
	p.MiterBacktrackLimit = -1
	p.Prove = prove
	return IvyFraigOptions{Parameters: p, MaxObjects: 10000, MaxSATNodes: 100000, MaxCandidatePairs: 100000, MaxSATVisits: 1000000, MaxBytes: 16 << 20, MaxWork: 1000000}
}

func TestIvyFraigDefaultsReductionClassesAndOwnership(t *testing.T) {
	if len(IvyFraigDefinitions) != 87 || IvyFraigDefinitions[0] != "Ivy_FraigAddClass" || IvyFraigDefinitions[len(IvyFraigDefinitions)-1] != "Ivy_ObjSim" {
		t.Fatalf("definition inventory: %d %q %q", len(IvyFraigDefinitions), IvyFraigDefinitions[0], IvyFraigDefinitions[len(IvyFraigDefinitions)-1])
	}
	defaults := DefaultIvyFraigParameters()
	if defaults.SimulationWords != 32 || defaults.SimulationSaturation != 0.005 || defaults.PatternScores || defaults.MaximumPatternScore != 25 || !defaults.Sparse || defaults.ActivityConeRatio != 0.3 || defaults.ActivityConeBump != 10 || defaults.NodeBacktrackLimit != 100 || defaults.MiterBacktrackLimit != 500000 {
		t.Fatalf("defaults: %+v", defaults)
	}
	source, pis, left, right := ivyFraigFixture(t)
	beforeObjects := slices.Clone(source.objects)
	beforeInputs, beforeOutputs := source.Inputs(), source.Outputs()
	result, err := source.Fraig(ivyFraigUnlimited(false))
	if err != nil {
		t.Fatal(err)
	}
	if result.Graph == nil || result.Status != IvyFraigUnknown || result.Stats.Merges < 3 || result.Stats.ConstantMerges < 1 || result.Stats.ComplementMerges < 1 || result.Stats.SATProved < 1 || result.Stats.ObjectsAfter >= result.Stats.ObjectsBefore {
		t.Fatalf("result status=%v stats=%+v", result.Status, result.Stats)
	}
	if !result.MappingValid[left.ID()] || !result.MappingValid[right.ID()] || result.Mapping[left.ID()] != result.Mapping[right.ID()] {
		t.Fatalf("equivalent mapping left=%v/%v right=%v/%v", result.Mapping[left.ID()], result.MappingValid[left.ID()], result.Mapping[right.ID()], result.MappingValid[right.ID()])
	}
	if len(result.Classes) == 0 {
		t.Fatal("missing certified classes")
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, sourceErr := source.Evaluate(values, nil)
		got, targetErr := result.Graph.Evaluate(values, nil)
		if sourceErr != nil || targetErr != nil || !slices.Equal(got, want) {
			t.Fatalf("assignment=%d got=%v want=%v sourceErr=%v targetErr=%v", assignment, got, want, sourceErr, targetErr)
		}
	}
	if !slices.Equal(source.objects, beforeObjects) || !slices.Equal(source.Inputs(), beforeInputs) || !slices.Equal(source.Outputs(), beforeOutputs) {
		t.Fatal("source manager mutated")
	}
	result.Mapping[0] = 99
	result.MappingValid[0] = false
	result.Classes[0].Members[0] = 99
	again, err := source.Fraig(ivyFraigUnlimited(false))
	if err != nil || again.Mapping[0] != IvyConst1 || !again.MappingValid[0] || len(again.Classes) == 0 || again.Classes[0].Members[0] == 99 {
		t.Fatalf("result aliases retained state: result=%+v err=%v", again, err)
	}
	if len(pis) != len(result.Graph.Inputs()) {
		t.Fatal("input order/count changed")
	}
}

func TestIvyFraigMiterProofCounterexampleAndEquivalent(t *testing.T) {
	safe, _, left, right := ivyFraigFixture(t)
	safeOnly, _ := NewIvyManager(IvyOptions{MaxObjects: 1000, MaxWork: 100000})
	inputs := make([]IvyEdge, 3)
	for i := range inputs {
		inputs[i], _ = safeOnly.CreatePI()
	}
	ab, _ := safeOnly.And(inputs[0], inputs[1])
	x, _ := safeOnly.And(ab, inputs[2])
	bc, _ := safeOnly.And(inputs[1], inputs[2])
	y, _ := safeOnly.And(inputs[0], bc)
	difference, _ := safeOnly.Xor(x, y)
	safeOnly.CreatePO(difference)
	proved, err := safeOnly.Fraig(ivyFraigUnlimited(true))
	if err != nil || proved.Status != IvyFraigProved || len(proved.Model) != 0 {
		t.Fatalf("proved status=%v model=%v err=%v", proved.Status, proved.Model, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		out, evalErr := proved.Graph.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}, nil)
		if evalErr != nil || len(out) != 1 || out[0] {
			t.Fatalf("proved assignment=%d out=%v err=%v", assignment, out, evalErr)
		}
	}

	unsafe, _ := NewIvyManager(IvyOptions{MaxObjects: 1000, MaxWork: 100000})
	a, _ := unsafe.CreatePI()
	b, _ := unsafe.CreatePI()
	bad, _ := unsafe.And(a, b)
	unsafe.CreatePO(bad)
	disproved, err := unsafe.Fraig(ivyFraigUnlimited(true))
	if err != nil || disproved.Status != IvyFraigCounterexample || len(disproved.Model) != 2 {
		t.Fatalf("disproved status=%v model=%v err=%v", disproved.Status, disproved.Model, err)
	}
	out, evalErr := unsafe.Evaluate(disproved.Model, nil)
	if evalErr != nil || len(out) != 1 || !out[0] {
		t.Fatalf("counterexample replay out=%v err=%v model=%v", out, evalErr, disproved.Model)
	}

	equivalent, err := safe.FraigEquivalent(left, right, ivyFraigUnlimited(false))
	if err != nil || equivalent.Status != FraigRelationProved {
		t.Fatalf("equivalent=%+v err=%v", equivalent, err)
	}
	notEquivalent, err := safe.FraigEquivalent(left, right.Not(), ivyFraigUnlimited(false))
	if err != nil || notEquivalent.Status != FraigRelationDisproved || len(notEquivalent.Model) != 3 {
		t.Fatalf("not equivalent=%+v err=%v", notEquivalent, err)
	}
	values := notEquivalent.Model
	leftValue, _ := safe.CutTruth(left, []uint32{1, 2, 3}, IvyExtendedOptions{MaxLeaves: 3, MaxTruth: 8})
	rightNotValue, _ := safe.CutTruth(right.Not(), []uint32{1, 2, 3}, IvyExtendedOptions{MaxLeaves: 3, MaxTruth: 8})
	assignment := 0
	for i, value := range values {
		if value {
			assignment |= 1 << uint(i)
		}
	}
	if ivyTruthBit(leftValue, assignment) == ivyTruthBit(rightNotValue, assignment) {
		t.Fatalf("equivalence counterexample does not replay: model=%v", values)
	}
}

func TestIvyFraigLimitsCancellationUnsupportedAndConcurrency(t *testing.T) {
	source, _, _, _ := ivyFraigFixture(t)
	badParameters := []IvyFraigParameters{
		{SimulationWords: -1},
		{SimulationWords: 129},
		{SimulationWords: 1, SimulationSaturation: 2, ActivityConeRatio: 0.3},
		{SimulationWords: 1, ActivityConeRatio: 1},
		{SimulationWords: 1, ActivityConeRatio: 0.3, NodeBacktrackLimit: -2},
	}
	for _, parameters := range badParameters {
		if result, err := source.Fraig(IvyFraigOptions{Parameters: parameters}); result.Graph != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("parameters=%+v result=%+v err=%v", parameters, result, err)
		}
	}
	for _, options := range []IvyFraigOptions{
		{Parameters: DefaultIvyFraigParameters(), MaxObjects: 1},
		{Parameters: DefaultIvyFraigParameters(), MaxSATNodes: 1},
		{Parameters: DefaultIvyFraigParameters(), MaxCandidatePairs: 1},
		{Parameters: DefaultIvyFraigParameters(), MaxBytes: 1},
		{Parameters: DefaultIvyFraigParameters(), MaxWork: 1},
	} {
		if result, err := source.Fraig(options); result.Graph != nil || !errors.Is(err, ErrLimit) {
			t.Fatalf("limited options=%+v result=%+v err=%v", options, result, err)
		}
	}
	if result, err := (*IvyManager)(nil).Fraig(ivyFraigUnlimited(false)); result.Graph != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil result=%+v err=%v", result, err)
	}
	sequential, _ := NewIvyManager(IvyOptions{})
	pi, _ := sequential.CreatePI()
	sequential.Latch(pi, IvyInit0)
	if result, err := sequential.Fraig(ivyFraigUnlimited(false)); result.Graph != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("sequential result=%+v err=%v", result, err)
	}
	if result, err := source.Fraig(IvyFraigOptions{Parameters: DefaultIvyFraigParameters(), Stop: func() bool { return true }}); result.Graph != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancelled result=%+v err=%v", result, err)
	}
	before := slices.Clone(source.objects)
	var wait sync.WaitGroup
	var failures atomic.Int32
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := source.Fraig(ivyFraigUnlimited(false))
			if err != nil || result.Graph == nil || result.Stats.Merges < 3 {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 || !slices.Equal(source.objects, before) {
		t.Fatalf("concurrent failures=%d source changed=%v", failures.Load(), !slices.Equal(source.objects, before))
	}
}

func FuzzIvyFraig(f *testing.F) {
	f.Add([]byte{3, 0, 1, 2, 5, 7, 11, 13}, uint8(2))
	f.Add([]byte{}, uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, outputByte uint8) {
		inputs := 2
		if len(data) > 0 {
			inputs += int(data[0] % 4)
		}
		m, _ := NewIvyManager(IvyOptions{MaxObjects: 1000, MaxWork: 100000})
		edges := make([]IvyEdge, inputs, inputs+16)
		for i := range edges {
			edges[i], _ = m.CreatePI()
		}
		for offset := 1; offset+2 < len(data) && len(edges) < inputs+16; offset += 3 {
			available := len(edges)
			a := edges[int(data[offset])%available].NotIf(data[offset+2]&1 != 0)
			b := edges[int(data[offset+1])%available].NotIf(data[offset+2]&2 != 0)
			var edge IvyEdge
			var err error
			if data[offset+2]&4 != 0 {
				edge, err = m.Xor(a, b)
			} else {
				edge, err = m.And(a, b)
			}
			if err != nil {
				t.Fatal(err)
			}
			edges = append(edges, edge)
		}
		outputs := 1 + int(outputByte%3)
		for i := 0; i < outputs; i++ {
			edge := edges[(i+int(outputByte))%len(edges)]
			if i&1 != 0 {
				edge = edge.Not()
			}
			if _, err := m.CreatePO(edge); err != nil {
				t.Fatal(err)
			}
		}
		result, err := m.Fraig(ivyFraigUnlimited(false))
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<uint(inputs); assignment++ {
			values := make([]bool, inputs)
			for i := range values {
				values[i] = assignment&(1<<uint(i)) != 0
			}
			want, sourceErr := m.Evaluate(values, nil)
			got, targetErr := result.Graph.Evaluate(values, nil)
			if sourceErr != nil || targetErr != nil || !slices.Equal(got, want) {
				t.Fatalf("assignment=%d got=%v want=%v sourceErr=%v targetErr=%v", assignment, got, want, sourceErr, targetErr)
			}
		}
	})
}
