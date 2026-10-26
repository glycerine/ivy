package abc

// Standalone semantic, ownership, limit, concurrency and fuzz tests for abcBidec.c.

import (
	"errors"
	"math"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func abcBidecFixture(t testing.TB, variables int) (*HopManager, []HopEdge) {
	t.Helper()
	manager, err := NewHopManager(HopLimits{MaxObjects: 1 << 16, MaxVars: max(variables, 2)})
	if err != nil {
		t.Fatal(err)
	}
	inputs := make([]HopEdge, variables)
	for index := range inputs {
		inputs[index], err = manager.IthVar(index)
		if err != nil {
			t.Fatal(err)
		}
	}
	return manager, inputs
}

func abcBidecTruth(t testing.TB, variables int, bits uint64) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(variables, false, 1024)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 1<<variables; assignment++ {
		if bits&(uint64(1)<<assignment) != 0 {
			if err = table.Set(uint64(assignment), true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return table
}

func abcBidecCompare(t testing.TB, manager *HopManager, first, second HopEdge, variables int) {
	t.Helper()
	for assignment := 0; assignment < 1<<variables; assignment++ {
		inputs := make([]bool, variables)
		for variable := range inputs {
			inputs[variable] = assignment&(1<<variable) != 0
		}
		a, err := manager.Evaluate(first, inputs)
		if err != nil {
			t.Fatal(err)
		}
		b, err := manager.Evaluate(second, inputs)
		if err != nil || a != b {
			t.Fatalf("assignment=%d source=%v result=%v err=%v", assignment, a, b, err)
		}
	}
}

func TestABCBidecNodeExactSemanticsAndOwnership(t *testing.T) {
	manager, inputs := abcBidecFixture(t, 4)
	left, _ := manager.And(inputs[0], inputs[1])
	right, _ := manager.And(inputs[2], inputs[3].Not())
	root, _ := manager.Or(left, right)
	before, _ := manager.MemorySnapshot()
	result, err := ABCBidecResynthesizeNode(manager, ABCBidecNode{Root: root, Variables: 4, Probability: -1}, ABCBidecOptions{})
	if err != nil {
		t.Fatal(err)
	}
	abcBidecCompare(t, manager, root, result.Root, 4)
	if result.HopNodesBefore == 0 || result.HopNodesAfter == 0 || result.Gain != result.HopNodesBefore-result.HopNodesAfter || result.CompletedDontCares {
		t.Fatalf("result=%+v", result)
	}
	after, _ := manager.MemorySnapshot()
	if !slices.Equal(before, after[:len(before)]) {
		t.Fatal("resynthesis mutated pre-existing HOP objects")
	}
	for _, constant := range []HopEdge{manager.Const0(), manager.Const1()} {
		constantResult, constantErr := ABCBidecResynthesizeNode(manager, ABCBidecNode{Root: constant, Variables: 4, Probability: -1}, ABCBidecOptions{})
		if constantErr != nil {
			t.Fatal(constantErr)
		}
		abcBidecCompare(t, manager, constant, constantResult.Root, 4)
	}
}

func TestABCBidecCareAndPowerCompletion(t *testing.T) {
	manager, inputs := abcBidecFixture(t, 3)
	root, _ := manager.Xor(inputs[0], inputs[1])
	care := abcBidecTruth(t, 3, 0x3c)
	careBefore := care.Words()
	for _, test := range []struct {
		probability float64
		complete    int
	}{
		{-1, -1},
		{0.25, -1},
		{0.4, 0},
		{0.6, 1},
	} {
		result, err := ABCBidecResynthesizeNode(manager, ABCBidecNode{Root: root, Variables: 3, Care: care, Probability: test.probability}, ABCBidecOptions{})
		if err != nil {
			t.Fatal(test.probability, err)
		}
		if result.CompletedDontCares != (test.complete >= 0) {
			t.Fatalf("probability=%g completed=%v", test.probability, result.CompletedDontCares)
		}
		for assignment := 0; assignment < 8; assignment++ {
			values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			source, _ := manager.Evaluate(root, values)
			got, _ := manager.Evaluate(result.Root, values)
			cared, _ := care.Get(uint64(assignment))
			if test.complete < 0 && cared && got != source {
				t.Fatalf("probability=%g cared assignment=%d", test.probability, assignment)
			}
			if test.complete == 0 && got != source && cared || test.complete == 0 && got && !cared {
				t.Fatalf("zero completion probability=%g assignment=%d source=%v care=%v got=%v", test.probability, assignment, source, cared, got)
			}
			if test.complete == 1 && got != source && cared || test.complete == 1 && !got && !cared {
				t.Fatalf("one completion probability=%g assignment=%d source=%v care=%v got=%v", test.probability, assignment, source, cared, got)
			}
		}
	}
	if !slices.Equal(careBefore, care.Words()) {
		t.Fatal("care table was mutated")
	}
}

func TestABCBidecBatchSkipErrorsAtomicDepthAndConcurrency(t *testing.T) {
	manager, inputs := abcBidecFixture(t, 16)
	root, _ := manager.And(inputs[0], inputs[1])
	batch, err := ABCBidecResynthesize(manager, []ABCBidecNode{{Root: root, Variables: 2, Probability: -1}, {Root: inputs[15], Variables: 16, Probability: -1}}, ABCBidecOptions{})
	if err != nil || batch.SkippedWide != 1 || batch.Roots[1] != inputs[15] {
		t.Fatalf("batch=%+v err=%v", batch, err)
	}
	abcBidecCompare(t, manager, root, batch.Roots[0], 2)

	for _, row := range []struct {
		node    ABCBidecNode
		options ABCBidecOptions
		want    error
	}{
		{ABCBidecNode{Root: HopEdge(math.MaxUint32), Variables: 2, Probability: -1}, ABCBidecOptions{}, ErrDimension},
		{ABCBidecNode{Root: root, Variables: -1, Probability: -1}, ABCBidecOptions{}, ErrDimension},
		{ABCBidecNode{Root: root, Variables: 2, Probability: 2}, ABCBidecOptions{}, ErrDimension},
		{ABCBidecNode{Root: root, Variables: 2, Care: abcBidecTruth(t, 3, 1), Probability: -1}, ABCBidecOptions{}, ErrDimension},
		{ABCBidecNode{Root: root, Variables: 2, Probability: -1}, ABCBidecOptions{MaxBDCNodes: 2}, ErrLimit},
		{ABCBidecNode{Root: root, Variables: 2, Probability: -1}, ABCBidecOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		before, _ := manager.MemorySnapshot()
		got, callErr := ABCBidecResynthesizeNode(manager, row.node, row.options)
		after, _ := manager.MemorySnapshot()
		if !errors.Is(callErr, row.want) || got.Root != 0 || !slices.Equal(before, after) {
			t.Fatalf("node=%+v options=%+v got=%+v err=%v want=%v mutation=%v", row.node, row.options, got, callErr, row.want, !slices.Equal(before, after))
		}
	}
	if got, callErr := ABCBidecResynthesize(manager, []ABCBidecNode{{Root: root, Variables: 2}, {Root: root, Variables: 2}}, ABCBidecOptions{MaxRoots: 1}); !errors.Is(callErr, ErrLimit) || got.Roots != nil {
		t.Fatalf("root limit got=%+v err=%v", got, callErr)
	}
	beforeBatch, _ := manager.MemorySnapshot()
	invalid := ABCBidecNode{Root: HopEdge(math.MaxUint32), Variables: 2, Probability: -1}
	if got, callErr := ABCBidecResynthesize(manager, []ABCBidecNode{{Root: root, Variables: 2, Probability: -1}, invalid}, ABCBidecOptions{}); !errors.Is(callErr, ErrDimension) || got.Roots != nil {
		t.Fatalf("atomic batch got=%+v err=%v", got, callErr)
	}
	afterBatch, _ := manager.MemorySnapshot()
	if !slices.Equal(beforeBatch, afterBatch) {
		t.Fatal("failed batch retained nodes from its successful prefix")
	}

	deep, deepInputs := abcBidecFixture(t, 2)
	deepRoot := deepInputs[0]
	for index := 0; index < 10000; index++ {
		deepRoot, err = deep.And(deepRoot.NotIf(index%3 == 0), deepInputs[index&1].NotIf(index%5 == 0))
		if err != nil {
			t.Fatal(err)
		}
	}
	deepResult, err := ABCBidecResynthesizeNode(deep, ABCBidecNode{Root: deepRoot, Variables: 2, Probability: -1}, ABCBidecOptions{})
	if err != nil {
		t.Fatal(err)
	}
	abcBidecCompare(t, deep, deepRoot, deepResult.Root, 2)

	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			local, localErr := NewHopManager(HopLimits{MaxObjects: 1024, MaxVars: 3})
			if localErr != nil {
				failures.Add(1)
				return
			}
			localInputs := make([]HopEdge, 3)
			for index := range localInputs {
				localInputs[index], localErr = local.IthVar(index)
				if localErr != nil {
					failures.Add(1)
					return
				}
			}
			localRoot, localErr := local.Xor(localInputs[0], localInputs[1])
			if localErr != nil {
				failures.Add(1)
				return
			}
			result, localErr := ABCBidecResynthesizeNode(local, ABCBidecNode{Root: localRoot, Variables: 3, Probability: -1}, ABCBidecOptions{})
			if localErr != nil {
				failures.Add(1)
				return
			}
			for assignment := 0; assignment < 8; assignment++ {
				values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
				a, _ := local.Evaluate(localRoot, values)
				b, _ := local.Evaluate(result.Root, values)
				if a != b {
					failures.Add(1)
					return
				}
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent independent failures=%d", failures.Load())
	}
}

func FuzzABCBidecResynthesis(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint16(0xa55a), uint8(0))
	f.Add([]byte{0xff, 7, 19, 23}, uint16(0x0ff0), uint8(3))
	f.Fuzz(func(t *testing.T, data []byte, careBits uint16, mode uint8) {
		if len(data) == 0 {
			return
		}
		manager, inputs := abcBidecFixture(t, 4)
		pool := slices.Clone(inputs)
		gates := 1 + int(data[0]%12)
		for gate := 0; gate < gates; gate++ {
			left := pool[int(data[(2*gate+1)%len(data)])%len(pool)].NotIf(data[(gate+3)%len(data)]&1 != 0)
			right := pool[int(data[(2*gate+2)%len(data)])%len(pool)].NotIf(data[(gate+4)%len(data)]&2 != 0)
			var root HopEdge
			var err error
			if data[gate%len(data)]&4 != 0 {
				root, err = manager.Xor(left, right)
			} else {
				root, err = manager.And(left, right)
			}
			if err != nil {
				t.Fatal(err)
			}
			pool = append(pool, root)
		}
		root := pool[len(pool)-1]
		care := abcBidecTruth(t, 4, uint64(careBits))
		probabilities := [...]float64{-1, 0.25, 0.4, 0.6}
		probability := probabilities[mode&3]
		result, err := ABCBidecResynthesizeNode(manager, ABCBidecNode{Root: root, Variables: 4, Care: care, Probability: probability}, ABCBidecOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
			source, _ := manager.Evaluate(root, values)
			got, _ := manager.Evaluate(result.Root, values)
			cared, _ := care.Get(uint64(assignment))
			want, check := source, cared
			if probability == 0.4 {
				want, check = source && cared, true
			} else if probability == 0.6 {
				want, check = source || !cared, true
			}
			if check && got != want {
				t.Fatalf("assignment=%d probability=%g source=%v care=%v got=%v want=%v", assignment, probability, source, cared, got, want)
			}
		}
	})
}
