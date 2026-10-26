package abc

// Stand-alone validation for src/opt/sim without a native oracle.

import (
	"errors"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"
)

func simTestCombinational(t *testing.T) (*AigManager, []AigLit) {
	t.Helper()
	graph := NewAigManager()
	inputs := make([]AigLit, 3)
	for i := range inputs {
		var err error
		inputs[i], err = graph.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	or, err := graph.Or(inputs[1], inputs[2])
	if err != nil {
		t.Fatal(err)
	}
	root, err := graph.And(inputs[0], or)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	return graph, inputs
}

func simTestSequential(t *testing.T) *AigManager {
	t.Helper()
	graph := NewAigManager()
	input, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	state, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(state); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(input); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return graph
}

func simTestBit(row []uint32, bit int) bool { return row[bit>>5]&(1<<uint(bit&31)) != 0 }

func TestSimUtilitiesAndOwnership(t *testing.T) {
	first := []uint32{0x80000001, 0x10}
	second := []uint32{1, 0x30}
	diffs, err := SimInfoDetectDiffs(first, second)
	if err != nil || !slices.Equal(diffs, []int{31, 37}) {
		t.Fatalf("diffs=%v err=%v", diffs, err)
	}
	news, err := SimInfoDetectNews(first, second)
	if err != nil || !slices.Equal(news, []int{37}) {
		t.Fatalf("news=%v err=%v", news, err)
	}
	if SimCountOnes(first) != 3 {
		t.Fatal("popcount")
	}
	if equal, _ := SimInfoEqual(first, slices.Clone(first)); !equal {
		t.Fatal("equal")
	}
	if implication, _ := SimInfoImplication([]uint32{1}, []uint32{3}); !implication {
		t.Fatal("implication")
	}
	if clause, _ := SimInfoClause([]uint32{0xaaaaaaaa}, []uint32{0x55555555}); !clause {
		t.Fatal("clause")
	}
	if SimSwitching([]uint32{0xffff0000}) != 0.5 {
		t.Fatal("switching")
	}
	info, err := NewSimInfo(2, 2, true, SimOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err = info.Write(1, first); err != nil {
		t.Fatal(err)
	}
	row, _ := info.Row(1)
	row[0] = 0
	owned, _ := info.Row(1)
	if owned[0] != first[0] {
		t.Fatal("borrowed row")
	}
	if err = info.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = info.Row(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	graph, _ := simTestCombinational(t)
	if err = SimulatePackedFrame(graph, info, 0, 1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestSimSupportSATStructuralAndSymmetry(t *testing.T) {
	graph, _ := simTestCombinational(t)
	structural, err := SimStructuralSupport(graph, SimOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(structural) != 1 || SimCountOnes(structural[0]) != 3 {
		t.Fatalf("structural=%v", structural)
	}
	functional, err := SimFunctionalSupport(graph, SimOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if SimCountOnes(functional.Supports[0]) != 3 || functional.Stats.SatRuns != 3 || functional.Stats.SatRunsSAT != 3 {
		t.Fatalf("functional=%v stats=%+v", functional.Supports, functional.Stats)
	}
	text, err := SimSupportString(functional.Supports, 3)
	if err != nil || text != "0: 0 1 2\n" {
		t.Fatalf("text=%q err=%v", text, err)
	}
	result, err := SimComputeTwoVariableSymmetries(graph, SimOptions{})
	if err != nil {
		t.Fatal(err)
	}
	bc, _ := result.Symmetric[0].Has(1, 2)
	ab, _ := result.NonSymmetric[0].Has(0, 1)
	if !bc || !ab || result.Stats.PairsTotal != 3 || result.Stats.PairsSymmetric != 1 || result.Stats.PairsStructurallySymmetric != 1 || result.Stats.PairsNonSymmetric != 2 {
		t.Fatalf("symmetry stats=%+v bc=%v ab=%v", result.Stats, bc, ab)
	}
	if disjoint, err := SimMatricesDisjoint(result.Symmetric, result.NonSymmetric); err != nil || !disjoint {
		t.Fatalf("disjoint=%v err=%v", disjoint, err)
	}
	witness, sat, err := SimSymmetryCounterexample(graph, 0, 0, 1, SimOptions{})
	if err != nil || !sat || len(witness) != 3 {
		t.Fatalf("witness=%v sat=%v err=%v", witness, sat, err)
	}
	left := slices.Clone(witness)
	right := slices.Clone(witness)
	left[0], left[1] = false, true
	right[0], right[1] = true, false
	a, _ := graph.Evaluate(left)
	b, _ := graph.Evaluate(right)
	if a[0] == b[0] {
		t.Fatal("invalid SAT witness")
	}
}

func TestSimAssociativeStructuralSymmetryAndCancelledSupport(t *testing.T) {
	graph := NewAigManager()
	inputs := make([]AigLit, 3)
	for i := range inputs {
		var err error
		inputs[i], err = graph.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	ab, _ := graph.And(inputs[0], inputs[1])
	abc, _ := graph.And(ab, inputs[2])
	if _, err := graph.CreateCO(abc); err != nil {
		t.Fatal(err)
	}
	support, err := SimStructuralSupport(graph, SimOptions{})
	if err != nil {
		t.Fatal(err)
	}
	matrices, count, err := SimStructuralSymmetries(graph, support, SimOptions{})
	if err != nil || count != 3 {
		t.Fatalf("associative structural symmetries=%d err=%v", count, err)
	}
	for a := 0; a < 3; a++ {
		for b := a + 1; b < 3; b++ {
			if symmetric, _ := matrices[0].Has(a, b); !symmetric {
				t.Fatalf("pair (%d,%d) missing", a, b)
			}
		}
	}

	cancel := NewAigManager()
	a, _ := cancel.CreateCI()
	b, _ := cancel.CreateCI()
	left, _ := cancel.And(a, b)
	right, _ := cancel.And(a, b.Not())
	root, _ := cancel.Or(left, right)
	_, _ = cancel.CreateCO(root)
	functional, err := SimFunctionalSupport(cancel, SimOptions{})
	if err != nil || !simTestBit(functional.Supports[0], 0) || simTestBit(functional.Supports[0], 1) {
		t.Fatalf("cancelled support=%v err=%v", functional.Supports, err)
	}
}

func TestSimSequentialSwitchingManagerLimitsAndCancellation(t *testing.T) {
	graph := simTestSequential(t)
	model := []bool{false, true, true, false}
	trace, err := SimulateSequentialModel(graph, 4, model, []SimInitial{SimInitZero}, SimOptions{})
	if err != nil {
		t.Fatal(err)
	}
	po := graph.Outputs()[0]
	want := []bool{false, false, true, true}
	for frame := range want {
		row, rowErr := trace.ObjectWords(po, frame)
		if rowErr != nil || (row[0] != 0) != want[frame] {
			t.Fatalf("frame=%d row=%x err=%v", frame, row, rowErr)
		}
	}
	if err = trace.Close(); err != nil {
		t.Fatal(err)
	}
	empty, err := SimulateSequentialModel(graph, 0, nil, []SimInitial{SimInitZero}, SimOptions{})
	if err != nil || empty.Frames != 0 {
		t.Fatalf("empty=%v err=%v", empty, err)
	}
	empty.Close()
	switching, err := SimNetworkSwitching(graph, 128, SimOptions{Seed: 1})
	if err != nil || len(switching) != graph.ObjectCount() {
		t.Fatalf("switching=%v err=%v", switching, err)
	}
	manager, err := NewSimManager(graph, false, SimOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if pattern, patternErr := manager.AllocPattern(); patternErr != nil || len(pattern.Data) != 1 {
		t.Fatalf("pattern=%v err=%v", pattern, patternErr)
	}
	statsText, _ := manager.StatsText()
	if !strings.Contains(statsText, "Total struct supps") {
		t.Fatal(statsText)
	}
	if err = manager.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Stats(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if _, err = SimStructuralSupport(graph, SimOptions{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = SimFunctionalSupport(graph, SimOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err = SimulateSequentialRandom(graph, 2, 2, []SimInitial{SimInitUnknown}, SimOptions{MaxWords: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = SimulateSequentialModel(graph, 1, []bool{false}, []SimInitial{99}, SimOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = SimSymmetryCounterexample(nil, 0, 0, 1, SimOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestSimFiveThousandScalarSequentialModels(t *testing.T) {
	graph := simTestSequential(t)
	random := rand.New(rand.NewSource(0x51a))
	for trial := 0; trial < 5000; trial++ {
		frames := 1 + random.Intn(8)
		model := make([]bool, frames)
		for i := range model {
			model[i] = random.Intn(2) != 0
		}
		initial := random.Intn(2) != 0
		initValue := SimInitZero
		if initial {
			initValue = SimInitOne
		}
		trace, err := SimulateSequentialModel(graph, frames, model, []SimInitial{initValue}, SimOptions{})
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		state := initial
		for frame := 0; frame < frames; frame++ {
			outputs, next, stepErr := graph.Step([]bool{model[frame]}, []bool{state})
			if stepErr != nil {
				t.Fatal(stepErr)
			}
			row, _ := trace.ObjectWords(graph.Outputs()[0], frame)
			if (row[0] == math.MaxUint32) != outputs[0] {
				t.Fatalf("trial=%d frame=%d got=%x want=%v", trial, frame, row[0], outputs[0])
			}
			state = next[0]
		}
		trace.Close()
	}
}

func TestSimFiveThousandPackedScalarModels(t *testing.T) {
	graph, _ := simTestCombinational(t)
	const patterns = 5000
	words := (patterns + 31) / 32
	info, err := NewSimInfo(graph.ObjectCount(), words, true, SimOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer info.Close()
	random := rand.New(rand.NewSource(0x51b))
	assignments := make([][]bool, patterns)
	for pattern := 0; pattern < patterns; pattern++ {
		assignment := make([]bool, graph.InputCount())
		for input, id := range graph.Inputs() {
			assignment[input] = random.Intn(2) != 0
			if assignment[input] {
				info.rows[id][pattern>>5] |= 1 << uint(pattern&31)
			}
		}
		assignments[pattern] = assignment
	}
	if err = SimulatePackedFrame(graph, info, 0, words); err != nil {
		t.Fatal(err)
	}
	output := graph.Outputs()[0]
	for pattern, assignment := range assignments {
		want, evaluateErr := graph.Evaluate(assignment)
		if evaluateErr != nil {
			t.Fatal(evaluateErr)
		}
		if simTestBit(info.rows[output], pattern) != want[0] {
			t.Fatalf("pattern=%d assignment=%v", pattern, assignment)
		}
	}
}

func FuzzSimScalarSequentialModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 0xabc, math.MaxUint64} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		graph := simTestSequential(t)
		random := rand.New(rand.NewSource(int64(seed)))
		frames := 1 + random.Intn(16)
		model := make([]bool, frames)
		for i := range model {
			model[i] = random.Intn(2) != 0
		}
		initial := random.Intn(2) != 0
		initValue := SimInitZero
		if initial {
			initValue = SimInitOne
		}
		trace, err := SimulateSequentialModel(graph, frames, model, []SimInitial{initValue}, SimOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer trace.Close()
		state := initial
		for frame := 0; frame < frames; frame++ {
			outputs, next, stepErr := graph.Step([]bool{model[frame]}, []bool{state})
			if stepErr != nil {
				t.Fatal(stepErr)
			}
			row, rowErr := trace.ObjectWords(graph.Outputs()[0], frame)
			if rowErr != nil || (row[0] == math.MaxUint32) != outputs[0] {
				t.Fatalf("frame=%d row=%x want=%v err=%v", frame, row, outputs[0], rowErr)
			}
			state = next[0]
		}
	})
}

func FuzzSimPackedScalarModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 0xabc, math.MaxUint64} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		graph, _ := simTestCombinational(t)
		random := rand.New(rand.NewSource(int64(seed)))
		patterns := 1 + random.Intn(160)
		words := (patterns + 31) / 32
		info, err := NewSimInfo(graph.ObjectCount(), words, true, SimOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer info.Close()
		assignments := make([][]bool, patterns)
		for pattern := 0; pattern < patterns; pattern++ {
			assignment := make([]bool, graph.InputCount())
			for input, id := range graph.Inputs() {
				assignment[input] = random.Intn(2) != 0
				if assignment[input] {
					info.rows[id][pattern>>5] |= 1 << uint(pattern&31)
				}
			}
			assignments[pattern] = assignment
		}
		if err = SimulatePackedFrame(graph, info, 0, words); err != nil {
			t.Fatal(err)
		}
		output := graph.Outputs()[0]
		for pattern, assignment := range assignments {
			want, evalErr := graph.Evaluate(assignment)
			if evalErr != nil || simTestBit(info.rows[output], pattern) != want[0] {
				t.Fatalf("pattern=%d assignment=%v want=%v err=%v", pattern, assignment, want, evalErr)
			}
		}
	})
}
