package abc

// Stand-alone tests for the lazy directional loader in sat/bmc/bmcLoad.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func lazyCNFGraph(t testing.TB) (*GiaManager, GiaLit, GiaLit, GiaLit) {
	t.Helper()
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	root, err := graph.AppendAnd(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	graph.AppendCO(root)
	return graph, a, b, root
}

func TestGiaLazyCNFDirectionalClausesAndLayout(t *testing.T) {
	graph, a, b, root := lazyCNFGraph(t)
	before := graph.Clone()
	manager, err := NewGiaLazyCNF(graph, GiaLazyCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if loaded, err := manager.LoadLiteral(root); err != nil || !loaded {
		t.Fatal(loaded, err)
	}
	want := [][]SatLit{{3}, {5, 7}, {5, 8}}
	if got := manager.Clauses(); !reflect.DeepEqual(got, want) {
		t.Fatal(got, want)
	}
	for _, test := range []struct {
		id       GiaID
		variable SatVar
	}{{0, 1}, {root.ID(), 2}, {a.ID(), 3}, {b.ID(), 4}} {
		if variable, found := manager.ObjectVariable(test.id); !found || variable != test.variable {
			t.Fatal(test, variable, found)
		}
	}
	if loaded, err := manager.LoadLiteral(root.Not()); err != nil || !loaded {
		t.Fatal(loaded, err)
	}
	want = append(want, []SatLit{4, 6, 9})
	if got := manager.Clauses(); !reflect.DeepEqual(got, want) {
		t.Fatal(got, want)
	}
	if loaded, err := manager.LoadLiteral(root); err != nil || loaded {
		t.Fatal(loaded, err)
	}
	stats := manager.Stats()
	if stats != (GiaLazyCNFStats{Variables: 5, Clauses: 4, Literals: 8, LoadRequests: 3, Loadings: 2}) {
		t.Fatal(stats)
	}
	copyClauses := manager.Clauses()
	copyClauses[0][0] = 0
	if manager.Clauses()[0][0] != 3 || !reflect.DeepEqual(graph, before) {
		t.Fatal("borrowed storage or source mutation")
	}
}

func giaLazyChain(t testing.TB, gates int, complement bool) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	root, _ := graph.AppendCI()
	for index := 0; index < gates; index++ {
		input, _ := graph.AppendCI()
		var err error
		root, err = graph.AppendAnd(root, input.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(err)
		}
	}
	graph.AppendCO(root.NotIf(complement))
	return graph
}

func TestGiaLazyCNFSolveEagerLazyAndConstants(t *testing.T) {
	const gates = 12
	graph := giaLazyChain(t, gates, false)
	lazy, err := NewGiaLazyCNF(graph, GiaLazyCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	lazyResult, err := lazy.SolveOutput(0, false, SatLimits{})
	if err != nil || lazyResult.Result.Status != SatSatisfiable || len(lazyResult.Inputs) != graph.InputCount() || lazyResult.Iterations < 2 {
		t.Fatal(lazyResult, err)
	}
	outputs, err := graph.Evaluate(lazyResult.Inputs)
	if err != nil || !outputs[0] {
		t.Fatal(outputs, err)
	}
	eager, _ := NewGiaLazyCNF(graph, GiaLazyCNFOptions{})
	eagerResult, err := eager.SolveOutput(0, true, SatLimits{})
	if err != nil || eagerResult.Result.Status != SatSatisfiable || eagerResult.Iterations != 1 {
		t.Fatal(eagerResult, err)
	}
	if lazyResult.Stats.Loadings != gates || eagerResult.Stats.Loadings != 2*gates || lazyResult.Stats.Clauses != 1+2*gates || eagerResult.Stats.Clauses != 1+3*gates {
		t.Fatal(lazyResult.Stats, eagerResult.Stats)
	}

	for _, test := range []struct {
		literal GiaLit
		want    SatStatus
	}{{GiaConst0, SatUnsatisfiable}, {GiaConst1, SatSatisfiable}} {
		literal, want := test.literal, test.want
		constant := NewGiaManager()
		constant.AppendCO(literal)
		manager, newErr := NewGiaLazyCNF(constant, GiaLazyCNFOptions{})
		if newErr != nil {
			t.Fatal(newErr)
		}
		answer, solveErr := manager.SolveOutput(0, false, SatLimits{})
		if solveErr != nil || answer.Result.Status != want {
			t.Fatal(literal, answer, solveErr)
		}
	}
}

func giaLazyOutputSatisfiable(t testing.TB, graph *GiaManager, output int) bool {
	t.Helper()
	if graph.InputCount() > 12 {
		t.Fatal("brute-force width")
	}
	inputs := make([]bool, graph.InputCount())
	for assignment := 0; assignment < 1<<len(inputs); assignment++ {
		for index := range inputs {
			inputs[index] = assignment>>index&1 != 0
		}
		outputs, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if outputs[output] {
			return true
		}
	}
	return false
}

func TestGiaLazyCNFExhaustiveFunctions(t *testing.T) {
	variables := 3
	for truth := 0; truth < 1<<(1<<variables); truth++ {
		graph := NewGiaManager()
		inputs := make([]GiaLit, variables)
		for index := range inputs {
			inputs[index], _ = graph.AppendCI()
		}
		root := GiaConst0
		for assignment := 0; assignment < 1<<variables; assignment++ {
			if truth>>assignment&1 == 0 {
				continue
			}
			cube := GiaConst1
			for variable, input := range inputs {
				literal := input.NotIf(assignment>>variable&1 == 0)
				if cube == GiaConst1 {
					cube = literal
				} else {
					cube, _ = graph.AppendAnd(cube, literal)
				}
			}
			if root == GiaConst0 {
				root = cube
			} else {
				bothFalse, _ := graph.AppendAnd(root.Not(), cube.Not())
				root = bothFalse.Not()
			}
		}
		graph.AppendCO(root)
		manager, err := NewGiaLazyCNF(graph, GiaLazyCNFOptions{})
		if err != nil {
			t.Fatal(truth, err)
		}
		answer, err := manager.SolveOutput(0, truth&1 != 0, SatLimits{})
		want := giaLazyOutputSatisfiable(t, graph, 0)
		if err != nil || (answer.Result.Status == SatSatisfiable) != want {
			t.Fatal(truth, answer.Result.Status, want, err)
		}
		if answer.Result.Status == SatSatisfiable {
			outputs, evalErr := graph.Evaluate(answer.Inputs)
			if evalErr != nil || !outputs[0] {
				t.Fatal(truth, outputs, evalErr)
			}
		}
	}
}

func TestGiaLazyCNFLimitsCancellationAtomicityAndLifecycle(t *testing.T) {
	graph, _, _, root := lazyCNFGraph(t)
	if manager, err := NewGiaLazyCNF(nil, GiaLazyCNFOptions{}); manager != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(manager, err)
	}
	for _, options := range []GiaLazyCNFOptions{
		{MaxVariables: -1}, {MaxClauses: -1}, {MaxLiterals: -1}, {MaxIterations: -1},
		{MaxVariables: 1}, {MaxClauses: 0, MaxLiterals: 0, MaxIterations: 0, Stop: func() bool { return true }},
	} {
		if manager, err := NewGiaLazyCNF(graph, options); manager != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, manager, err)
		}
	}
	for _, options := range []GiaLazyCNFOptions{{MaxVariables: 3}, {MaxClauses: 2}, {MaxLiterals: 4}} {
		manager, err := NewGiaLazyCNF(graph, options)
		if err != nil {
			t.Fatal(options, err)
		}
		beforeStats, beforeClauses := manager.Stats(), manager.Clauses()
		if loaded, loadErr := manager.LoadLiteral(root); loaded || !errors.Is(loadErr, ErrLimit) {
			t.Fatal(options, loaded, loadErr)
		}
		if manager.Stats() != beforeStats || !reflect.DeepEqual(manager.Clauses(), beforeClauses) {
			t.Fatal("failed load committed", options)
		}
	}
	limited, err := NewGiaLazyCNF(giaLazyChain(t, 4, false), GiaLazyCNFOptions{MaxIterations: 1})
	if err != nil {
		t.Fatal(err)
	}
	before := limited.Stats()
	if answer, solveErr := limited.SolveOutput(0, false, SatLimits{}); !errors.Is(solveErr, ErrLimit) || answer.Result.Status != SatSatisfiable || limited.Stats() != before {
		t.Fatal(answer, solveErr, limited.Stats(), before)
	}

	for threshold := 2; threshold <= 4; threshold++ {
		calls := 0
		manager, newErr := NewGiaLazyCNF(graph, GiaLazyCNFOptions{Stop: func() bool {
			calls++
			return calls == threshold
		}})
		if newErr != nil {
			if !errors.Is(newErr, ErrCancelled) {
				t.Fatal(threshold, newErr)
			}
			continue
		}
		beforeStats := manager.Stats()
		_, loadErr := manager.LoadLiteral(root)
		if !errors.Is(loadErr, ErrCancelled) || manager.Stats() != beforeStats {
			t.Fatal(threshold, calls, loadErr, manager.Stats(), beforeStats)
		}
	}

	buffered := NewGiaManager()
	input, _ := buffered.AppendCI()
	buffer, _ := buffered.AppendBuffer(input)
	buffered.AppendCO(buffer)
	if manager, err := NewGiaLazyCNF(buffered, GiaLazyCNFOptions{}); manager != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(manager, err)
	}
	manager, _ := NewGiaLazyCNF(graph, GiaLazyCNFOptions{})
	manager.Close()
	manager.Close()
	if manager.graph != nil {
		t.Fatal("close retained graph")
	}
	if _, err := manager.LoadLiteral(root); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func FuzzGiaLazyCNF(f *testing.F) {
	f.Add([]byte{1, 3, 5, 7, 9, 11, 13})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		graph := NewGiaManager()
		width := 1 + int(data[0]%6)
		signals := make([]GiaLit, width)
		for index := range signals {
			signals[index], _ = graph.AppendCI()
		}
		for index := 1; index < len(data) && len(signals) < 48; index++ {
			leftIndex := int(data[index]) % len(signals)
			rightIndex := (leftIndex + 1 + int(data[index]>>4)) % len(signals)
			if signals[leftIndex].ID() == signals[rightIndex].ID() {
				continue
			}
			left := signals[leftIndex].NotIf(data[index]&1 != 0)
			right := signals[rightIndex].NotIf(data[index]&2 != 0)
			literal, err := graph.AppendAnd(left, right)
			if err == nil {
				signals = append(signals, literal)
			}
		}
		root := signals[int(data[len(data)-1])%len(signals)].NotIf(data[len(data)-1]&0x80 != 0)
		graph.AppendCO(root)
		manager, err := NewGiaLazyCNF(graph, GiaLazyCNFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		answer, err := manager.SolveOutput(0, data[0]&0x80 != 0, SatLimits{})
		want := giaLazyOutputSatisfiable(t, graph, 0)
		if err != nil || (answer.Result.Status == SatSatisfiable) != want {
			t.Fatal(answer.Result.Status, want, err)
		}
		if answer.Result.Status == SatSatisfiable {
			outputs, evalErr := graph.Evaluate(answer.Inputs)
			if evalErr != nil || !slices.Equal(outputs, []bool{true}) {
				t.Fatal(outputs, evalErr)
			}
		}
	})
}
