package abc

// Standalone tests for src/opt/mfs/mfsWin.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func mfsWindowFixture(t *testing.T) (*ABCNetwork, int, []int) {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	start, _ := network.CreateNode()
	left, _ := network.CreateNode()
	right, _ := network.CreateNode()
	deep, _ := network.CreateNode()
	output, _ := network.CreatePO()
	for _, edge := range [][2]int{{start, left}, {start, right}, {right, deep}, {left, output}} {
		if err = network.AddFanin(edge[1], ABCObjectEdge{ID: edge[0]}); err != nil {
			t.Fatal(err)
		}
	}
	levels := make([]int, network.ObjectCount()+1)
	levels[start], levels[left], levels[right], levels[deep], levels[output] = 1, 2, 2, 3, 2
	return network, start, levels
}

func TestABCMFSWindowRootsSourceOrderAndCutoffs(t *testing.T) {
	network, start, levels := mfsWindowFixture(t)
	roots, err := ABCMFSWindowRoots(network, start, levels, 1, 8, ABCMFSWindowOptions{})
	if err != nil || !slices.Equal(roots, []int{2, 3}) {
		t.Fatalf("roots=%v err=%v", roots, err)
	}
	byFanout, err := ABCMFSWindowRoots(network, start, levels, 20, 1, ABCMFSWindowOptions{})
	if err != nil || !slices.Equal(byFanout, []int{start}) {
		t.Fatalf("fanout root=%v err=%v", byFanout, err)
	}
	strict, err := ABCMFSWindowRoots(network, start, levels, 2, 8, ABCMFSWindowOptions{})
	if err != nil || !slices.Equal(strict, []int{2}) {
		t.Fatalf("strict cutoff=%v err=%v", strict, err)
	}
	if levels[start] != 1 || network.objects[start].MarkA || network.objects[start].MarkB || network.objects[start].MarkC {
		t.Fatal("root collection mutated source state")
	}
}

func TestABCMFSWindowRootsValidationLimitsAndDepth(t *testing.T) {
	network, start, levels := mfsWindowFixture(t)
	cases := []struct {
		network *ABCNetwork
		start   int
		levels  []int
		window  int
		fanouts int
		options ABCMFSWindowOptions
		want    error
	}{
		{nil, start, levels, 1, 1, ABCMFSWindowOptions{}, ErrDimension},
		{network, -1, levels, 1, 1, ABCMFSWindowOptions{}, ErrDimension},
		{network, start, levels[:2], 1, 1, ABCMFSWindowOptions{}, ErrDimension},
		{network, start, levels, -1, 1, ABCMFSWindowOptions{}, ErrDimension},
		{network, start, levels, 1, -1, ABCMFSWindowOptions{}, ErrDimension},
		{network, start, levels, 1, 8, ABCMFSWindowOptions{MaxVisited: -1}, ErrDimension},
		{network, start, levels, 1, 8, ABCMFSWindowOptions{MaxRoots: -1}, ErrDimension},
		{network, start, levels, 1, 8, ABCMFSWindowOptions{MaxVisited: 1}, ErrLimit},
		{network, start, levels, 1, 8, ABCMFSWindowOptions{MaxRoots: 1}, ErrLimit},
		{network, start, levels, 1, 8, ABCMFSWindowOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	negative := slices.Clone(levels)
	negative[start] = -1
	cases = append(cases, struct {
		network *ABCNetwork
		start   int
		levels  []int
		window  int
		fanouts int
		options ABCMFSWindowOptions
		want    error
	}{network, start, negative, 1, 8, ABCMFSWindowOptions{}, ErrDimension})
	for index, row := range cases {
		roots, err := ABCMFSWindowRoots(row.network, row.start, row.levels, row.window, row.fanouts, row.options)
		if roots != nil || !errors.Is(err, row.want) {
			t.Fatalf("case %d: roots=%v err=%v want=%v", index, roots, err, row.want)
		}
	}

	deep, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: 20_010, MaxEdges: 20_010})
	first, _ := deep.CreateNode()
	previous := first
	deepLevels := make([]int, 10_003)
	for i := 1; i <= 10_000; i++ {
		next, _ := deep.CreateNode()
		if err := deep.AddFanin(next, ABCObjectEdge{ID: previous}); err != nil {
			t.Fatal(err)
		}
		deepLevels[next] = i
		previous = next
	}
	output, _ := deep.CreatePO()
	if err := deep.AddFanin(output, ABCObjectEdge{ID: previous}); err != nil {
		t.Fatal(err)
	}
	deepLevels[output] = 10_000
	if roots, err := ABCMFSWindowRoots(deep, first, deepLevels, 20_000, 2, ABCMFSWindowOptions{MaxVisited: 10_001}); err != nil || !slices.Equal(roots, []int{previous}) {
		t.Fatalf("deep roots=%v want=%d err=%v", roots, previous, err)
	}

	calls := 0
	if _, err := ABCMFSWindowRoots(network, start, levels, 1, 8, ABCMFSWindowOptions{Stop: func() bool { calls++; return false }}); err != nil {
		t.Fatal(err)
	}
	for stopAt := 1; stopAt <= calls; stopAt++ {
		count := 0
		roots, err := ABCMFSWindowRoots(network, start, levels, 1, 8, ABCMFSWindowOptions{Stop: func() bool { count++; return count == stopAt }})
		if roots != nil || !errors.Is(err, ErrCancelled) {
			t.Fatalf("stop %d/%d: roots=%v err=%v", stopAt, calls, roots, err)
		}
	}
}

func TestABCMFSWindowRootsConcurrentFrozenNetwork(t *testing.T) {
	network, start, levels := mfsWindowFixture(t)
	var wait sync.WaitGroup
	var failures atomic.Int32
	for worker := 0; worker < 64; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			roots, err := ABCMFSWindowRoots(network, start, levels, 1, 8, ABCMFSWindowOptions{})
			if err != nil || !slices.Equal(roots, []int{2, 3}) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures: %d", failures.Load())
	}
}

func FuzzABCMFSWindowRoots(f *testing.F) {
	f.Add([]byte{6, 0, 1, 2, 3, 4, 5}, uint8(2), uint8(3))
	f.Add([]byte{2}, uint8(0), uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, windowByte, fanoutByte uint8) {
		nodeCount := 2
		if len(data) > 0 {
			nodeCount += int(data[0] % 23)
		}
		network, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
		nodes := make([]int, nodeCount)
		for i := range nodes {
			nodes[i], _ = network.CreateNode()
		}
		for i := 0; i+1 < nodeCount; i++ {
			if err := network.AddFanin(nodes[i+1], ABCObjectEdge{ID: nodes[i]}); err != nil {
				t.Fatal(err)
			}
		}
		if len(data) > 1 {
			for i, value := range data[1:] {
				if nodeCount < 3 {
					break
				}
				from := int(value) % (nodeCount - 1)
				span := nodeCount - from - 1
				to := from + 1 + (i+int(value)/7)%span
				if err := network.AddFanin(nodes[to], ABCObjectEdge{ID: nodes[from]}); err != nil {
					t.Fatal(err)
				}
			}
		}
		output, _ := network.CreatePO()
		if err := network.AddFanin(output, ABCObjectEdge{ID: nodes[nodeCount-1]}); err != nil {
			t.Fatal(err)
		}
		levels := make([]int, len(network.objects))
		for _, id := range nodes {
			for _, edge := range network.objects[id].Fanins {
				levels[id] = max(levels[id], levels[edge.ID]+1)
			}
		}
		levels[output] = levels[nodes[nodeCount-1]]
		window, fanoutLimit := int(windowByte%16), int(fanoutByte%8)
		maximum := levels[nodes[0]] + window
		visited := make(map[int]bool)
		want := make([]int, 0)
		var visit func(int)
		visit = func(id int) {
			if visited[id] {
				return
			}
			visited[id] = true
			object := network.objects[id]
			root := len(object.Fanouts) > fanoutLimit
			for _, fanout := range object.Fanouts {
				root = root || abcIsCO(network.objects[fanout].Type) || levels[fanout] > maximum
			}
			if root {
				want = append(want, id)
				return
			}
			for _, fanout := range object.Fanouts {
				visit(fanout)
			}
		}
		visit(nodes[0])
		got, err := ABCMFSWindowRoots(network, nodes[0], levels, window, fanoutLimit, ABCMFSWindowOptions{})
		if err != nil || !slices.Equal(got, want) {
			t.Fatalf("got=%v want=%v err=%v", got, want, err)
		}
	})
}
