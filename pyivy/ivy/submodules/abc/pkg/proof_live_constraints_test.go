package abc

// Tests for the Go port of ABC's liveness constraint and arena transforms.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func liveTestGraph(t testing.TB) (*AigManager, []string) {
	t.Helper()
	graph := NewAigManager()
	a := aigTestCI(t, graph)
	b := aigTestCI(t, graph)
	q := aigTestCI(t, graph)
	ab, err := graph.And(a, b)
	if err != nil {
		t.Fatal(err)
	}
	next, err := graph.Xor(a, q)
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range []AigLit{a, q, a, b, ab} {
		aigTestCO(t, graph, output)
	}
	aigTestCO(t, graph, next)
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	graph.Name, graph.Spec = "live", "source-spec"
	aigTestCheck(t, graph)
	return graph, []string{
		"csLiveConst_a", "csLiveTarget_q", "windowBegins_main",
		"withinWindow_main", "pendingSignal_main",
	}
}

func liveBools(mask, count int) []bool {
	result := make([]bool, count)
	for i := range result {
		result[i] = mask&(1<<i) != 0
	}
	return result
}

func TestLiveCombinationCountAndSourceOrder(t *testing.T) {
	for n := 0; n <= 18; n++ {
		for k := 0; k <= n; k++ {
			count, err := LiveCombinationCount(n, k)
			if err != nil {
				t.Fatal(n, k, err)
			}
			want := uint64(1)
			for i := 1; i <= k; i++ {
				want = want * uint64(n-k+i) / uint64(i)
			}
			if count != want {
				t.Fatal(n, k, count, want)
			}
		}
	}
	got, err := LiveCombinations(5, 3, LiveTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]int{{2, 1, 0}, {3, 1, 0}, {3, 2, 0}, {3, 2, 1}, {4, 1, 0}, {4, 2, 0}, {4, 2, 1}, {4, 3, 0}, {4, 3, 1}, {4, 3, 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	copy := CloneLiveCombinations(got)
	copy[0][0] = 99
	if got[0][0] == 99 {
		t.Fatal("combination clone aliases")
	}
	if _, err = LiveCombinationCount(2, 3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = LiveCombinationCount(1000, 500); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = LiveCombinations(8, 4, LiveTransformOptions{MaxCombinations: 10}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = LiveCombinations(3, 2, LiveTransformOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func TestLiveNamedConstrainedZeroLiveExhaustive(t *testing.T) {
	source, names := liveTestGraph(t)
	before := source.Clone()
	result, err := BuildNamedConstrainedZeroLive(source, names, LiveTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.OutputIndex != 5 || result.Graph.PrimaryOutputCount() != 6 || result.Graph.RegisterCount() != 1 || result.Graph.Name != "live_0Live" || result.Graph.Spec != "" {
		t.Fatal(result.OutputIndex, result.Graph.PrimaryOutputCount(), result.Graph.RegisterCount(), result.Graph.Name, result.Graph.Spec)
	}
	if !reflect.DeepEqual(source, before) {
		t.Fatal("source graph mutated")
	}
	for mask := 0; mask < 8; mask++ {
		inputs := liveBools(mask, 3)
		old, err := source.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := result.Graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got[:5], old[:5]) || got[6] != old[5] {
			t.Fatal(mask, old, got)
		}
		want := !old[0] || old[1]
		if got[result.OutputIndex] != want {
			t.Fatal(mask, got[result.OutputIndex], want)
		}
	}
	antecedents, consequent, err := CollectLiveConstraintOutputs(names)
	if err != nil || !slices.Equal(antecedents, []int{0}) || consequent != 1 {
		t.Fatal(antecedents, consequent, err)
	}
	for marker, want := range map[string]int{"windowBegins_": 2, "withinWindow_": 3, "pendingSignal": 4} {
		if got, findErr := FindLiveOutput(names, marker); findErr != nil || got != want {
			t.Fatal(marker, got, findErr)
		}
	}
	if _, _, err = CollectLiveConstraintOutputs([]string{"csLiveConst_x"}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	nodes, err := GatherLiveNodes(source)
	if err != nil || len(nodes) != source.AndCount() {
		t.Fatal(nodes, err)
	}
	if len(nodes) > 0 {
		nodes[0] = InvalidAigID
		again, againErr := GatherLiveNodes(source)
		if againErr != nil || again[0] == InvalidAigID {
			t.Fatal("node list aliases", againErr)
		}
	}
}

func TestLiveDisjunctiveTesterTraceModel(t *testing.T) {
	source, _ := liveTestGraph(t)
	result, err := BuildLiveDisjunctiveTester(source, []int{0, 1, 4}, 2, LiveTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantCombinations := [][]int{{0}, {1}, {2}, {1, 0}, {2, 0}, {2, 1}}
	if !reflect.DeepEqual(result.Combinations, wantCombinations) || result.FirstOutput != 5 {
		t.Fatal(result.Combinations, result.FirstOutput)
	}
	graph := result.Graph
	if graph.RegisterCount() != 7 || graph.PrimaryOutputCount() != 11 || graph.PrimaryInputCount() != 2 || graph.Check() != nil {
		t.Fatal(graph.RegisterCount(), graph.PrimaryOutputCount(), graph.PrimaryInputCount(), graph.Check())
	}
	for inputMask := 0; inputMask < 4; inputMask++ {
		primary := liveBools(inputMask, 2)
		for stateMask := 0; stateMask < 128; stateMask++ {
			state := liveBools(stateMask, 7)
			sourceOutputs, sourceNext, err := source.Step(primary, state[:1])
			if err != nil {
				t.Fatal(err)
			}
			outputs, next, err := graph.Step(primary, state)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(outputs[:5], sourceOutputs) || next[0] != sourceNext[0] {
				t.Fatal(inputMask, stateMask, sourceOutputs, outputs, sourceNext, next)
			}
			candidates := []bool{sourceOutputs[0], sourceOutputs[1], sourceOutputs[4]}
			for i, combination := range result.Combinations {
				disjunction := false
				for _, index := range combination {
					disjunction = disjunction || candidates[index]
				}
				want := !disjunction || state[1+i]
				if outputs[5+i] != want || next[1+i] != want {
					t.Fatal(inputMask, stateMask, i, outputs[5+i], next[1+i], want)
				}
			}
		}
	}
	general, err := BuildGeneralLiveDisjunctiveTester(source, 1, LiveTransformOptions{})
	if err != nil || general.Graph.RegisterCount() != 2 || !reflect.DeepEqual(general.Combinations, [][]int{{0}}) {
		t.Fatal(err, general.Combinations)
	}
}

func TestLiveArenaTransformExhaustiveTraceRelation(t *testing.T) {
	source, names := liveTestGraph(t)
	before := source.Clone()
	result, err := BuildArenaConstrainedZeroLive(source, names, [][]int{{0, 1}}, LiveTransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(source, before) {
		t.Fatal("arena transform mutated source")
	}
	if result.OutputIndex != 5 || !slices.Equal(result.BarrierRegisters, []int{1}) || result.ArenaViolationRegister != 2 || result.PendingRegister != 3 || !slices.Equal(result.MonotoneRegisters, []int{4}) {
		t.Fatal(result)
	}
	graph := result.Graph
	if graph.RegisterCount() != 5 || graph.PrimaryOutputCount() != 6 || graph.PrimaryInputCount() != 2 || graph.Check() != nil {
		t.Fatal(graph.RegisterCount(), graph.PrimaryOutputCount(), graph.PrimaryInputCount(), graph.Check())
	}
	for inputMask := 0; inputMask < 4; inputMask++ {
		primary := liveBools(inputMask, 2)
		for stateMask := 0; stateMask < 32; stateMask++ {
			state := liveBools(stateMask, 5)
			sourceOutputs, sourceNext, err := source.Step(primary, state[:1])
			if err != nil {
				t.Fatal(err)
			}
			outputs, next, err := graph.Step(primary, state)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(outputs[:5], sourceOutputs) || next[0] != sourceNext[0] {
				t.Fatal(inputMask, stateMask, sourceOutputs, outputs, sourceNext, next)
			}
			barrier := sourceOutputs[0] || sourceOutputs[1]
			window, within, pending := sourceOutputs[2], sourceOutputs[3], sourceOutputs[4]
			barrierNext := (barrier && window) || state[1]
			violation := (barrier != state[1]) && within
			violationNext := violation || state[2]
			monotone := !(pending && state[3] && state[4]) || barrier
			constrained := !sourceOutputs[0] || sourceOutputs[1]
			newOutput := (constrained || state[2]) && monotone
			wantNext := []bool{sourceNext[0], barrierNext, violationNext, pending, barrier}
			if outputs[5] != newOutput || !slices.Equal(next, wantNext) {
				t.Fatal(inputMask, stateMask, outputs[5], newOutput, next, wantNext)
			}
		}
	}
}

func TestLiveTransformValidationLimitsCancellationAndAtomicity(t *testing.T) {
	source, names := liveTestGraph(t)
	before := source.Clone()
	tests := []func() error{
		func() error {
			_, err := BuildConstrainedZeroLive(source, []int{99}, 1, LiveTransformOptions{})
			return err
		},
		func() error { _, err := BuildLiveDisjunctiveTester(source, nil, 1, LiveTransformOptions{}); return err },
		func() error {
			_, err := BuildGeneralLiveDisjunctiveTester(source, 2, LiveTransformOptions{})
			return err
		},
		func() error {
			_, err := BuildArenaConstrainedZeroLive(source, names, [][]int{{99}}, LiveTransformOptions{})
			return err
		},
		func() error {
			_, err := BuildNamedConstrainedZeroLive(source, names, LiveTransformOptions{MaxOperations: 1})
			return err
		},
		func() error {
			_, err := BuildNamedConstrainedZeroLive(source, names, LiveTransformOptions{Stop: func() bool { return true }})
			return err
		},
	}
	for i, run := range tests {
		if err := run(); err == nil {
			t.Fatal(i, "expected rejection")
		}
		if !reflect.DeepEqual(source, before) {
			t.Fatal(i, "failed transform mutated source")
		}
	}
	if _, err := BuildArenaConstrainedZeroLive(source, names[:4], [][]int{{0}}, LiveTransformOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func fuzzLiveDisjunctiveModel(t testing.TB, data []byte) {
	t.Helper()
	if len(data) < 2 {
		return
	}
	n := 1 + int(data[0]%4)
	k := 1 + int(data[1]%byte(n))
	graph := NewAigManager()
	variables := make([]AigLit, n)
	for i := range variables {
		variables[i] = aigTestCI(t, graph)
		aigTestCO(t, graph, variables[i].NotIf(i&1 != 0))
	}
	aigTestCheck(t, graph)
	candidates := make([]int, n)
	for i := range candidates {
		candidates[i] = i
	}
	result, err := BuildLiveDisjunctiveTester(graph, candidates, k, LiveTransformOptions{MaxCombinations: 32})
	if err != nil {
		t.Fatal(err)
	}
	inputs := make([]bool, n)
	for i := range inputs {
		inputs[i] = data[(2+i)%len(data)]&(1<<uint(i%8)) != 0
	}
	state := make([]bool, len(result.Combinations))
	for i := range state {
		state[i] = data[(2+n+i)%len(data)]&(1<<uint(i%8)) != 0
	}
	outputs, next, err := result.Graph.Step(inputs, state)
	if err != nil {
		t.Fatal(err)
	}
	for i, combination := range result.Combinations {
		disjunction := false
		for _, index := range combination {
			value := inputs[index]
			if index&1 != 0 {
				value = !value
			}
			disjunction = disjunction || value
		}
		want := !disjunction || state[i]
		if outputs[n+i] != want || next[i] != want {
			t.Fatal(n, k, i, outputs[n+i], next[i], want)
		}
	}
}

func TestLiveConstraintDeterministicModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		data := make([]byte, 12)
		for i := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[i] = byte(state)
		}
		fuzzLiveDisjunctiveModel(t, data)
	}
}

func FuzzLiveConstraints(f *testing.F) {
	for _, seed := range [][]byte{{1, 1}, {4, 3, 0xff}, {2, 2, 1, 2, 3}, {3, 1, 0, 0xff, 0x55}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		fuzzLiveDisjunctiveModel(t, data)
	})
}
