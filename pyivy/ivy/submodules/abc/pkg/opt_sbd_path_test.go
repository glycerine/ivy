package abc

// Standalone path, box, limit, concurrency and fuzz tests for sbdPath.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func sbdPathFixture(t testing.TB) (*GiaManager, *GiaLUTMapping, []GiaID) {
	t.Helper()
	graph := NewGiaManager()
	inputs := make([]GiaLit, 4)
	for index := range inputs {
		var err error
		inputs[index], err = graph.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	x, err := graph.AppendAnd(inputs[0], inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	y, err := graph.AppendAnd(x, inputs[2])
	if err != nil {
		t.Fatal(err)
	}
	w, err := graph.AppendAnd(y, inputs[3])
	if err != nil {
		t.Fatal(err)
	}
	z, err := graph.AppendAnd(inputs[1], inputs[2])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.AppendCO(w); err != nil {
		t.Fatal(err)
	}
	records := []GiaLUTRecord{
		{Root: x.ID(), Fanins: []GiaID{inputs[0].ID(), inputs[1].ID()}},
		{Root: y.ID(), Fanins: []GiaID{x.ID(), inputs[2].ID()}},
		{Root: w.ID(), Fanins: []GiaID{y.ID(), inputs[3].ID()}},
		{Root: z.ID(), Fanins: []GiaID{inputs[1].ID(), inputs[2].ID()}},
	}
	mapping, err := NewGiaLUTMapping(graph.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	return graph, mapping, []GiaID{inputs[0].ID(), inputs[1].ID(), inputs[2].ID(), inputs[3].ID(), x.ID(), y.ID(), w.ID(), z.ID()}
}

func TestSBDCriticalPathAndDelayTraceExact(t *testing.T) {
	graph, mapping, ids := sbdPathFixture(t)
	path, err := SBDCriticalPath(graph, mapping, nil, SBDPathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantPath := make([]bool, graph.ObjectCount())
	for _, id := range []GiaID{ids[0], ids[1], ids[2], ids[4], ids[5], ids[6]} {
		wantPath[id] = true
	}
	if !slices.Equal(path, wantPath) {
		t.Fatalf("critical=%v want=%v", path, wantPath)
	}
	expanded, err := SBDAddInternalToPath(graph, mapping, path, SBDPathOptions{})
	if err != nil || !slices.Equal(expanded, path) {
		t.Fatalf("expanded=%v path=%v err=%v", expanded, path, err)
	}
	expanded[ids[0]] = false
	if !path[ids[0]] {
		t.Fatal("expanded path aliases source")
	}

	trace, err := SBDTraceDelay(graph, mapping, SBDPathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	wantTracePath := make([]bool, graph.ObjectCount())
	for _, id := range []GiaID{ids[0], ids[1], ids[4], ids[5], ids[6]} {
		wantTracePath[id] = true
	}
	if trace.AIGNodes != 4 || trace.LUTs != 4 || trace.Levels != 3 || trace.PathNodes != 3 || trace.PathEdges != 4 || trace.AllPathEdges != 4 || !slices.Equal(trace.Path, wantTracePath) {
		t.Fatalf("trace=%+v wantPath=%v", trace, wantTracePath)
	}
}

func TestSBDCriticalPathTimingBox(t *testing.T) {
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	c, _ := graph.AppendCI()
	boxOutput, _ := graph.AppendCI()
	x, _ := graph.AppendAnd(a, c)
	y, _ := graph.AppendAnd(boxOutput, x)
	graph.AppendCO(x)
	graph.AppendCO(y)
	mapping, err := NewGiaLUTMapping(graph.ObjectCount(), []GiaLUTRecord{
		{Root: x.ID(), Fanins: []GiaID{a.ID(), c.ID()}},
		{Root: y.ID(), Fanins: []GiaID{boxOutput.ID(), x.ID()}},
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	timing, err := NewTimManager(3, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = timing.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{0}}, TimDelayOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err = timing.AddBox(TimBox{FirstInput: 0, Inputs: 1, FirstOutput: 2, Outputs: 1, DelayTable: 0}, TimBoxOptions{}); err != nil {
		t.Fatal(err)
	}
	path, err := SBDCriticalPath(graph, mapping, timing, SBDPathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := make([]bool, graph.ObjectCount())
	for _, id := range []GiaID{a.ID(), c.ID(), boxOutput.ID(), x.ID(), y.ID()} {
		want[id] = true
	}
	if !slices.Equal(path, want) {
		t.Fatalf("boxed path=%v want=%v", path, want)
	}
	if _, err = SBDCriticalPath(graph, mapping, NewTimManagerForTestMismatch(t), SBDPathOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("mismatched timing err=%v", err)
	}
}

func TestSBDCriticalPathPreservesOutputMajorFirstVisitSlack(t *testing.T) {
	graph := NewGiaManager()
	inputs := make([]GiaLit, 9)
	for index := range inputs {
		inputs[index], _ = graph.AppendCI()
	}
	add := func(left, right GiaLit) GiaLit {
		t.Helper()
		result, err := graph.AppendAnd(left, right)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	p := add(inputs[0], inputs[1])
	x := add(p, inputs[2])
	h1 := add(inputs[3], inputs[4])
	h2 := add(h1, inputs[5])
	h3 := add(h2, inputs[6])
	q1 := add(inputs[7], inputs[8])
	q2 := add(q1, inputs[3])
	y := add(x, q2)
	first := add(x, h3)
	second := add(y, inputs[0])
	graph.AppendCO(first)
	graph.AppendCO(second)
	records := []GiaLUTRecord{
		{Root: p.ID(), Fanins: []GiaID{inputs[0].ID(), inputs[1].ID()}},
		{Root: x.ID(), Fanins: []GiaID{p.ID(), inputs[2].ID()}},
		{Root: h1.ID(), Fanins: []GiaID{inputs[3].ID(), inputs[4].ID()}},
		{Root: h2.ID(), Fanins: []GiaID{h1.ID(), inputs[5].ID()}},
		{Root: h3.ID(), Fanins: []GiaID{h2.ID(), inputs[6].ID()}},
		{Root: q1.ID(), Fanins: []GiaID{inputs[7].ID(), inputs[8].ID()}},
		{Root: q2.ID(), Fanins: []GiaID{q1.ID(), inputs[3].ID()}},
		{Root: y.ID(), Fanins: []GiaID{x.ID(), q2.ID()}},
		{Root: first.ID(), Fanins: []GiaID{x.ID(), h3.ID()}},
		{Root: second.ID(), Fanins: []GiaID{y.ID(), inputs[0].ID()}},
	}
	mapping, err := NewGiaLUTMapping(graph.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	path, err := SBDCriticalPath(graph, mapping, nil, SBDPathOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !path[x.ID()] || path[inputs[2].ID()] {
		t.Fatalf("first output must visit x with zero slack before the second output: path=%v", path)
	}
}

func NewTimManagerForTestMismatch(t testing.TB) *TimManager {
	t.Helper()
	timing, err := NewTimManager(1, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	return timing
}

func TestSBDPathLimitsErrorsDepthAndConcurrency(t *testing.T) {
	graph, mapping, _ := sbdPathFixture(t)
	for _, row := range []struct {
		options SBDPathOptions
		want    error
	}{
		{SBDPathOptions{MaxObjects: -1}, ErrDimension},
		{SBDPathOptions{MaxVisits: -1}, ErrDimension},
		{SBDPathOptions{MaxObjects: 1}, ErrLimit},
		{SBDPathOptions{MaxVisits: 1}, ErrLimit},
		{SBDPathOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if path, err := SBDCriticalPath(graph, mapping, nil, row.options); path != nil || !errors.Is(err, row.want) {
			t.Fatalf("critical options=%+v path=%v err=%v want=%v", row.options, path, err, row.want)
		}
	}
	if path, err := SBDCriticalPath(nil, mapping, nil, SBDPathOptions{}); path != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil graph path=%v err=%v", path, err)
	}
	if path, err := SBDAddInternalToPath(graph, mapping, make([]bool, 1), SBDPathOptions{}); path != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("short path=%v err=%v", path, err)
	}
	wrong, _ := NewGiaLUTMapping(1, nil, 0)
	if trace, err := SBDTraceDelay(graph, wrong, SBDPathOptions{}); trace.Path != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("wrong mapping trace=%+v err=%v", trace, err)
	}

	deep := NewGiaManager()
	a, _ := deep.AppendCI()
	b, _ := deep.AppendCI()
	root := a
	records := make([]GiaLUTRecord, 0, 10000)
	for index := 0; index < 10000; index++ {
		next, err := deep.AppendAnd(root, b.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(err)
		}
		records = append(records, GiaLUTRecord{Root: next.ID(), Fanins: []GiaID{root.ID(), b.ID()}})
		root = next
	}
	deep.AppendCO(root)
	deepMapping, err := NewGiaLUTMapping(deep.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	path, err := SBDCriticalPath(deep, deepMapping, nil, SBDPathOptions{MaxObjects: 20000, MaxVisits: 100000})
	if err != nil || !path[root.ID()] || !path[a.ID()] || !path[b.ID()] {
		t.Fatalf("deep root/input path=%v/%v/%v err=%v", path[root.ID()], path[a.ID()], path[b.ID()], err)
	}

	baseline, _ := SBDCriticalPath(graph, mapping, nil, SBDPathOptions{})
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, err := SBDCriticalPath(graph, mapping, nil, SBDPathOptions{})
			trace, traceErr := SBDTraceDelay(graph, mapping, SBDPathOptions{})
			if err != nil || traceErr != nil || !slices.Equal(got, baseline) || trace.PathNodes != 3 {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func sbdPathFuzzGraph(t testing.TB, data []byte) (*GiaManager, *GiaLUTMapping, [][]GiaID, []int) {
	t.Helper()
	graph := NewGiaManager()
	pool := make([]GiaLit, 4)
	for index := range pool {
		pool[index], _ = graph.AppendCI()
	}
	gateCount := 1 + int(data[0]%12)
	records := make([]GiaLUTRecord, 0, gateCount)
	modelFanins := make([][]GiaID, graph.ObjectCount()+gateCount)
	levels := make([]int, graph.ObjectCount()+gateCount+1)
	for gate := 0; gate < gateCount; gate++ {
		first := int(data[(2*gate+1)%len(data)]) % len(pool)
		second := int(data[(2*gate+2)%len(data)]) % len(pool)
		if second == first {
			second = (second + 1) % len(pool)
		}
		node, err := graph.AppendAnd(pool[first], pool[second])
		if err != nil {
			t.Fatal(err)
		}
		fanins := []GiaID{pool[first].ID(), pool[second].ID()}
		records = append(records, GiaLUTRecord{Root: node.ID(), Fanins: fanins})
		modelFanins[node.ID()] = fanins
		levels[node.ID()] = max(levels[fanins[0]], levels[fanins[1]]) + 1
		pool = append(pool, node)
	}
	graph.AppendCO(pool[len(pool)-1])
	mapping, err := NewGiaLUTMapping(graph.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	return graph, mapping, modelFanins, levels[:graph.ObjectCount()]
}

func FuzzSBDCriticalPath(f *testing.F) {
	f.Add([]byte{4, 1, 2, 3, 4, 5})
	f.Add([]byte{11, 0, 1, 1, 2, 3, 0, 4})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		graph, mapping, fanins, levels := sbdPathFuzzGraph(t, data)
		rootID := graph.objects[graph.cos[0]].Diff0()
		root := graph.cos[0] - GiaID(rootID)
		wantCritical := make([]bool, graph.ObjectCount())
		type entry struct{ id, level, slack int }
		stack := []entry{{int(root), levels[root], 1}}
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if wantCritical[current.id] {
				continue
			}
			wantCritical[current.id] = true
			for index := len(fanins[current.id]) - 1; index >= 0; index-- {
				fanin := fanins[current.id][index]
				available := levels[fanin] + current.slack
				if available >= current.level-1 {
					stack = append(stack, entry{int(fanin), levels[fanin], max(0, available-(current.level-1))})
				}
			}
		}
		got, err := SBDCriticalPath(graph, mapping, nil, SBDPathOptions{MaxObjects: 128, MaxVisits: 4096})
		if err != nil || !slices.Equal(got, wantCritical) {
			t.Fatalf("critical got=%v want=%v err=%v data=%x", got, wantCritical, err, data)
		}

		wantDelay := make([]bool, graph.ObjectCount())
		wantDelay[root] = true
		wantNodes, wantEdges := 0, 0
		for id := len(fanins) - 1; id >= 0; id-- {
			if !wantDelay[id] || len(fanins[id]) == 0 {
				continue
			}
			wantNodes++
			for _, fanin := range fanins[id] {
				if levels[fanin]+1 == levels[id] {
					wantDelay[fanin] = true
					wantEdges++
				}
			}
		}
		trace, err := SBDTraceDelay(graph, mapping, SBDPathOptions{MaxObjects: 128, MaxVisits: 4096})
		if err != nil || trace.Levels != levels[root] || trace.PathNodes != wantNodes || trace.PathEdges != wantEdges || !slices.Equal(trace.Path, wantDelay) {
			t.Fatalf("trace=%+v want levels/nodes/edges/path=%d/%d/%d/%v err=%v data=%x", trace, levels[root], wantNodes, wantEdges, wantDelay, err, data)
		}
	})
}
