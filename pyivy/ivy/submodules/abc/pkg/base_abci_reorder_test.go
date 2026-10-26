package abc

// Standalone semantic, ownership, limit, concurrency and fuzz tests for abcReorder.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func abcBDDTruthFromFunction(t testing.TB, variables int, function func([]bool) bool) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(variables, false, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	values := make([]bool, variables)
	for assignment := 0; assignment < 1<<variables; assignment++ {
		for variable := range values {
			values[variable] = assignment&(1<<variable) != 0
		}
		if function(values) {
			if err = table.Set(uint64(assignment), true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return table
}

func abcBDDNodeValue(t testing.TB, node ABCBDDLocalNode, signals map[int]bool) bool {
	t.Helper()
	assignment := uint64(0)
	for variable, fanin := range node.Fanins {
		if signals[fanin] {
			assignment |= uint64(1) << variable
		}
	}
	value, err := node.Function.Get(assignment)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestABCNodeBDDReorderSemanticsCostAndOwnership(t *testing.T) {
	// A mux tree is deliberately presented in a poor data-before-select order.
	table := abcBDDTruthFromFunction(t, 5, func(x []bool) bool {
		if x[4] {
			if x[3] {
				return x[2]
			}
			return x[1]
		}
		return x[0]
	})
	node := ABCBDDLocalNode{Name: "mux", Fanins: []int{10, 11, 12, 13, 14}, Function: table}
	beforeWords := table.Words()
	beforeFanins := slices.Clone(node.Fanins)
	result, err := ABCNodeBDDReorder(node, ABCBDDReorderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Stats.NodesAfter >= result.Stats.NodesBefore || abcBDDIdentityOrder(result.Order) || len(result.Order) != 5 || result.Name != "mux" {
		t.Fatalf("result=%+v", result)
	}
	for assignment := 0; assignment < 32; assignment++ {
		signals := make(map[int]bool, 5)
		for index, fanin := range node.Fanins {
			signals[fanin] = assignment&(1<<index) != 0
		}
		got := abcBDDNodeValue(t, ABCBDDLocalNode{Fanins: result.Fanins, Function: result.Function}, signals)
		want := abcBDDNodeValue(t, node, signals)
		if got != want {
			t.Fatalf("assignment=%d got=%v want=%v order=%v", assignment, got, want, result.Order)
		}
	}
	if !slices.Equal(beforeWords, table.Words()) || !slices.Equal(beforeFanins, node.Fanins) {
		t.Fatal("source node was mutated")
	}
	result.Fanins[0] = 999
	result.Function.Set(0, !abcBDDNodeValue(t, node, map[int]bool{}))
	if !slices.Equal(beforeWords, table.Words()) || node.Fanins[0] != 10 {
		t.Fatal("result aliases source storage")
	}
}

func TestABCNodeBDDReorderDuplicateAndMinimumBase(t *testing.T) {
	table := abcBDDTruthFromFunction(t, 4, func(x []bool) bool { return x[0] && x[1] || x[2] })
	node := ABCBDDLocalNode{Name: "duplicate", Fanins: []int{7, 7, 9, 11}, Function: table}
	result, err := ABCNodeBDDReorder(node, ABCBDDReorderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(result.Fanins, []int{7, 9}) || result.Function.Variables() != 2 || result.Stats.FaninsBefore != 4 || result.Stats.FaninsAfter != 2 {
		t.Fatalf("result=%+v", result)
	}
	for bits := 0; bits < 8; bits++ {
		signals := map[int]bool{7: bits&1 != 0, 9: bits&2 != 0, 11: bits&4 != 0}
		if abcBDDNodeValue(t, node, signals) != abcBDDNodeValue(t, ABCBDDLocalNode{Fanins: result.Fanins, Function: result.Function}, signals) {
			t.Fatalf("signals=%v", signals)
		}
	}
	constant := ABCBDDLocalNode{Name: "constant", Fanins: []int{1, 1}, Function: abcBDDTruthFromFunction(t, 2, func(x []bool) bool { return x[0] != x[1] })}
	constantResult, err := ABCNodeBDDReorder(constant, ABCBDDReorderOptions{})
	if err != nil || len(constantResult.Fanins) != 0 || !constantResult.Function.IsConstant(false) {
		t.Fatalf("constant=%+v err=%v", constantResult, err)
	}
	wideTruth, err := NewVariableTruthTable(21, 0, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	wideFanins := make([]int, 21)
	for index := range wideFanins {
		wideFanins[index] = index
	}
	wideResult, err := ABCNodeBDDReorder(ABCBDDLocalNode{Name: "wide-small-support", Fanins: wideFanins, Function: wideTruth}, ABCBDDReorderOptions{})
	if err != nil || !slices.Equal(wideResult.Fanins, []int{0}) || wideResult.Function.Variables() != 1 {
		t.Fatalf("wide result=%+v err=%v", wideResult, err)
	}
}

func TestABCNetworkBDDReorderErrorsLimitsCancellationAndConcurrency(t *testing.T) {
	first := ABCBDDLocalNode{Name: "a", Fanins: []int{0, 1, 2}, Function: abcBDDTruthFromFunction(t, 3, func(x []bool) bool { return x[0] != x[1] && x[2] })}
	second := ABCBDDLocalNode{Name: "b", Fanins: []int{3}, Function: abcBDDTruthFromFunction(t, 1, func(x []bool) bool { return x[0] })}
	result, err := ABCNetworkBDDReorder([]ABCBDDLocalNode{first, second}, ABCBDDReorderOptions{})
	if err != nil || len(result.Nodes) != 2 || result.Nodes[1].Name != "b" {
		t.Fatalf("network=%+v err=%v", result, err)
	}
	badWidth := abcBDDTruthFromFunction(t, 2, func(x []bool) bool { return x[0] })
	for _, row := range []struct {
		nodes   []ABCBDDLocalNode
		options ABCBDDReorderOptions
		want    error
	}{
		{[]ABCBDDLocalNode{{Fanins: []int{1}, Function: badWidth}}, ABCBDDReorderOptions{}, ErrDimension},
		{[]ABCBDDLocalNode{{Fanins: []int{-1}, Function: second.Function}}, ABCBDDReorderOptions{}, ErrDimension},
		{[]ABCBDDLocalNode{first}, ABCBDDReorderOptions{MaxFanins: 2}, ErrLimit},
		{[]ABCBDDLocalNode{first, second}, ABCBDDReorderOptions{MaxNodes: 1}, ErrLimit},
		{[]ABCBDDLocalNode{first}, ABCBDDReorderOptions{MaxTruthWords: -1}, ErrDimension},
		{[]ABCBDDLocalNode{first}, ABCBDDReorderOptions{MaxEvaluations: 1}, ErrLimit},
		{[]ABCBDDLocalNode{first}, ABCBDDReorderOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		got, callErr := ABCNetworkBDDReorder(row.nodes, row.options)
		if !errors.Is(callErr, row.want) || got.Nodes != nil {
			t.Fatalf("options=%+v got=%+v err=%v want=%v", row.options, got, callErr, row.want)
		}
	}

	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, callErr := ABCNodeBDDReorder(first, ABCBDDReorderOptions{})
			if callErr != nil || got.Function == nil || got.Stats.NodesAfter > got.Stats.NodesBefore {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzABCNodeBDDReorder(f *testing.F) {
	f.Add(uint8(5), uint64(0x69969669), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(3), uint64(0xe8), []byte{7, 7, 9})
	f.Fuzz(func(t *testing.T, rawVariables uint8, bits uint64, rawFanins []byte) {
		variables := 1 + int(rawVariables%6)
		fanins := make([]int, variables)
		for index := range fanins {
			if len(rawFanins) == 0 {
				fanins[index] = index
			} else {
				fanins[index] = int(rawFanins[index%len(rawFanins)] % 5)
			}
		}
		table := abcBDDTruthFromFunction(t, variables, func(x []bool) bool {
			assignment := 0
			for index, value := range x {
				if value {
					assignment |= 1 << index
				}
			}
			return bits&(uint64(1)<<assignment) != 0
		})
		node := ABCBDDLocalNode{Fanins: fanins, Function: table}
		result, err := ABCNodeBDDReorder(node, ABCBDDReorderOptions{MaxEvaluations: 512})
		if err != nil {
			t.Fatal(err)
		}
		unique := make([]int, 0, 5)
		seen := make(map[int]bool)
		for _, fanin := range fanins {
			if !seen[fanin] {
				seen[fanin] = true
				unique = append(unique, fanin)
			}
		}
		for assignment := 0; assignment < 1<<len(unique); assignment++ {
			signals := make(map[int]bool, len(unique))
			for index, fanin := range unique {
				signals[fanin] = assignment&(1<<index) != 0
			}
			want := abcBDDNodeValue(t, node, signals)
			got := abcBDDNodeValue(t, ABCBDDLocalNode{Fanins: result.Fanins, Function: result.Function}, signals)
			if got != want {
				t.Fatalf("assignment=%d fanins=%v order=%v got=%v want=%v", assignment, fanins, result.Order, got, want)
			}
		}
		if result.Stats.NodesAfter > result.Stats.NodesBefore {
			t.Fatalf("cost increased: %+v", result.Stats)
		}
	})
}
