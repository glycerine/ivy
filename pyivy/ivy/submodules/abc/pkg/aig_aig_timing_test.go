package abc

// Tests for owned direct/reverse timing from aig/aig/aigTiming.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigTimingOracleRow struct {
	Seed, Increase, Maximum int
	Forward, Reverse        []uint32
	Required                []int64
	Recomputed              []int
}

func aigTimingGraph(t testing.TB, seed int) *AigManager {
	t.Helper()
	graph := NewAigManager()
	v := aigObjectVariables(t, graph, 6)
	n0 := aigObjectAnd(t, graph, v[0].NotIf(seed&1 != 0), v[1].NotIf(seed&2 != 0))
	n1 := aigObjectAnd(t, graph, n0, v[2].NotIf(seed&4 != 0))
	n2 := aigObjectAnd(t, graph, n0, v[3].NotIf(seed&8 != 0))
	n3 := aigObjectAnd(t, graph, n1, n2)
	n4 := aigObjectAnd(t, graph, n3, v[4].NotIf(seed&16 != 0))
	dang0 := aigObjectAnd(t, graph, v[4].Not(), v[5])
	_ = aigObjectAnd(t, graph, dang0, v[0].Not())
	for _, driver := range []AigLit{n4.NotIf(seed&1 != 0), n2.NotIf(seed&2 != 0), v[5].NotIf(seed&4 != 0)} {
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	return graph
}

func TestAigTimingNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_timing.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []aigTimingOracleRow
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		graph := aigTimingGraph(t, row.Seed)
		timing, err := BuildAigTiming(graph, row.Increase, AigTimingOptions{})
		if err != nil || timing.MaximumLevel() != int64(row.Maximum) || !slices.Equal(timing.ForwardLevels(), row.Forward) || !slices.Equal(timing.ReverseLevels(), row.Reverse) {
			t.Fatal(row.Seed, err, timing.MaximumLevel(), row.Maximum, timing.ForwardLevels(), row.Forward, timing.ReverseLevels(), row.Reverse)
		}
		for id := range row.Forward {
			required, err := timing.RequiredLevel(AigID(id))
			if err != nil || required != row.Required[id] {
				t.Fatal(row.Seed, id, required, row.Required[id], err)
			}
			if row.Recomputed[id] >= 0 {
				level, err := timing.ReverseLevelNew(AigID(id))
				if err != nil || int(level) != row.Recomputed[id] {
					t.Fatal(row.Seed, id, level, row.Recomputed[id], err)
				}
			}
		}
		if err = timing.Verify(); err != nil {
			t.Fatal(row.Seed, err)
		}
	}
}

func TestAigTimingErrorsOwnershipRefreshAndDeep(t *testing.T) {
	graph := aigTimingGraph(t, 7)
	for _, test := range []struct {
		graph    *AigManager
		increase int
		options  AigTimingOptions
		want     error
	}{
		{nil, 0, AigTimingOptions{}, ErrDimension},
		{graph, -1, AigTimingOptions{}, ErrDimension},
		{graph, 0, AigTimingOptions{MaxObjects: -1}, ErrDimension},
		{graph, 0, AigTimingOptions{MaxObjects: graph.ObjectCount() - 1}, ErrLimit},
		{graph, 0, AigTimingOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := BuildAigTiming(test.graph, test.increase, test.options); result != nil || !errors.Is(err, test.want) {
			t.Fatal(result, err, test.want)
		}
	}
	buffered := NewAigManager()
	input := aigObjectVariables(t, buffered, 1)[0]
	if _, err := buffered.CreateBuffer(input); err != nil {
		t.Fatal(err)
	}
	if result, err := BuildAigTiming(buffered, 0, AigTimingOptions{}); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}

	timing, err := BuildAigTiming(graph, 2, AigTimingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	forward, reverse := timing.ForwardLevels(), timing.ReverseLevels()
	forward[7], reverse[7] = 99, 99
	if timing.ForwardLevels()[7] == 99 || timing.ReverseLevels()[7] == 99 {
		t.Fatal("returned slices alias timing")
	}
	before := *timing
	before.objects = slices.Clone(timing.objects)
	before.forward = slices.Clone(timing.forward)
	before.reverse = slices.Clone(timing.reverse)
	if err = timing.Refresh(graph, AigTimingOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(*timing, before) {
		t.Fatal(err, reflect.DeepEqual(*timing, before))
	}
	if err = timing.ClearReverseLevel(7); err != nil || !errors.Is(timing.Verify(), ErrInvariant) {
		t.Fatal(err, timing.Verify())
	}
	if err = timing.Refresh(graph, AigTimingOptions{}); err != nil || timing.Verify() != nil {
		t.Fatal(err, timing.Verify())
	}
	extra, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	oldDriver := graph.objects[graph.cos[0]].Fanin0
	newDriver := aigObjectAnd(t, graph, oldDriver, extra)
	if _, err = graph.CreateCO(newDriver); err != nil {
		t.Fatal(err)
	}
	if err = timing.Refresh(graph, AigTimingOptions{}); err != nil || timing.ObjectCount() != graph.ObjectCount() || timing.MaximumLevel() != 7 || timing.reverse[newDriver.ID()] != 1 || timing.reverse[oldDriver.ID()] != 2 || timing.Verify() != nil {
		t.Fatal(err, timing.ObjectCount(), graph.ObjectCount(), timing.MaximumLevel(), timing.reverse[newDriver.ID()], timing.reverse[oldDriver.ID()], timing.Verify())
	}
	invalid := AigID(graph.ObjectCount() + 1)
	if _, err = timing.ReverseLevel(invalid); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if _, err = timing.RequiredLevel(invalid); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	if _, err = timing.ReverseLevelNew(graph.cis[0]); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if err = (*AigTiming)(nil).Refresh(graph, AigTimingOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}

	const depth = 10000
	deep := NewAigManager()
	v := aigObjectVariables(t, deep, depth)
	root := v[0]
	for index := 1; index < depth; index++ {
		root = aigObjectAnd(t, deep, root, v[index])
	}
	if _, err = deep.CreateCO(root); err != nil {
		t.Fatal(err)
	}
	deepTiming, err := BuildAigTiming(deep, 3, AigTimingOptions{MaxObjects: deep.ObjectCount()})
	firstAND := AigID(depth + 1)
	if err != nil || deepTiming.MaximumLevel() != depth+2 || deepTiming.reverse[root.ID()] != 1 || deepTiming.reverse[firstAND] != uint32(depth-1) || deepTiming.Verify() != nil {
		t.Fatal(err, deepTiming.MaximumLevel(), deepTiming.reverse[root.ID()], deepTiming.reverse[firstAND], deepTiming.Verify())
	}
}

func FuzzAigTiming(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x123456789abcdef0, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		graph := NewAigManager()
		count := 2 + int(seed%10)
		v := aigObjectVariables(t, graph, count)
		nodes := slices.Clone(v)
		for index := 0; index < count*2; index++ {
			left := nodes[int((seed>>uint(index%32))%uint64(len(nodes)))]
			right := v[(index+1)%len(v)].NotIf(seed&(1<<uint(index%63)) != 0)
			node := aigObjectAnd(t, graph, left, right)
			nodes = append(nodes, node)
		}
		for index := 0; index < 3; index++ {
			driver := nodes[int((seed>>uint((index*7)%32))%uint64(len(nodes)))]
			if _, err := graph.CreateCO(driver); err != nil {
				t.Fatal(err)
			}
		}
		timing, err := BuildAigTiming(graph, int(seed%5), AigTimingOptions{})
		if err != nil || timing.Verify() != nil {
			t.Fatal(seed, err, timing.Verify())
		}
		for id, object := range graph.objects {
			if object.Type != AigAND {
				continue
			}
			want := uint32(1)
			for fanoutID, fanout := range graph.objects {
				if (fanout.Type == AigAND && (fanout.Fanin0.ID() == AigID(id) || fanout.Fanin1.ID() == AigID(id))) || (fanout.Type == AigCO && fanout.Fanin0.ID() == AigID(id)) {
					want = max(want, timing.reverse[fanoutID]+1)
				}
			}
			if timing.reverse[id] != want {
				t.Fatal(seed, id, timing.reverse[id], want)
			}
		}
	})
}
