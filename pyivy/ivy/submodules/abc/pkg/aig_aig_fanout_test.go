package abc

// Tests for ordered fanout storage from aig/aig/aigFanout.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigFanoutFixture(t testing.TB, seed int) (*AigManager, AigID, AigID) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 6)
	leaf := make([]AigLit, 6)
	for i := range ci {
		ci[i], _ = g.CreateCI()
		leaf[i] = ci[i].NotIf(seed>>uint(i%5)&1 != 0)
	}
	ab, _ := g.And(leaf[0], leaf[1])
	ac, _ := g.And(leaf[0], leaf[2])
	root0, _ := g.And(ab, ac)
	root1, _ := g.And(ab, leaf[3])
	g.CreateCO(root0.NotIf(seed&1 != 0))
	g.CreateCO(root1.NotIf(seed>>1&1 != 0))
	g.CreateCO(ab.NotIf(seed>>2&1 != 0))
	g.CreateCO(ac.NotIf(seed>>3&1 != 0))
	return g, ab.ID(), root1.ID()
}

func aigFanoutPairs(t testing.TB, index *AigFanoutIndex) [][][2]int {
	t.Helper()
	result := make([][][2]int, index.ObjectCount())
	for id := range result {
		list, err := index.Fanouts(AigID(id))
		if err != nil {
			t.Fatal(err)
		}
		result[id] = make([][2]int, len(list))
		for i, edge := range list {
			result[id][i] = [2]int{int(edge.Object), int(edge.Pin)}
		}
	}
	return result
}

func TestAigFanoutNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_fanout.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, Objects int
		AB, Root1     AigID
		Initial       [][][2]int
		Removed       [][][2]int
		Restored      [][][2]int
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g, ab, root1 := aigFanoutFixture(t, row.Seed)
		index, buildErr := BuildAigFanoutIndex(g, AigTraversalOptions{})
		if buildErr != nil || index.ObjectCount() != row.Objects || ab != row.AB || root1 != row.Root1 || !reflect.DeepEqual(aigFanoutPairs(t, index), row.Initial) {
			t.Fatal(row.Seed, ab, root1, buildErr, aigFanoutPairs(t, index), row.Initial)
		}
		list, _ := index.Fanouts(ab)
		position := slices.IndexFunc(list, func(edge AigFanout) bool { return edge.Object == root1 })
		if position < 0 {
			t.Fatal(row.Seed, list)
		}
		if err = index.RemoveObject(g, ab, root1); err != nil || !reflect.DeepEqual(aigFanoutPairs(t, index), row.Removed) {
			t.Fatal(row.Seed, err, aigFanoutPairs(t, index), row.Removed)
		}
		if err = index.AddObject(g, ab, root1); err != nil || !reflect.DeepEqual(aigFanoutPairs(t, index), row.Restored) {
			t.Fatal(row.Seed, err, aigFanoutPairs(t, index), row.Restored)
		}
	}
}

func TestAigFanoutOwnershipErrorsAndLimits(t *testing.T) {
	g, source, fanout := aigFanoutFixture(t, 0)
	index, err := BuildAigFanoutIndex(g, AigTraversalOptions{})
	if err != nil || index.EdgeCount() != 12 {
		t.Fatal(index, err)
	}
	clone := index.Clone()
	list, _ := clone.Fanouts(source)
	list[0] = AigFanout{}
	original, _ := index.Fanouts(source)
	if original[0] == (AigFanout{}) {
		t.Fatal("query aliases index")
	}
	edge := slices.IndexFunc(original, func(edge AigFanout) bool { return edge.Object == fanout })
	if edge < 0 || clone.Remove(source, original[edge]) != nil || reflect.DeepEqual(clone, index) {
		t.Fatal(edge)
	}
	before := index.Clone()
	bad := []struct {
		source AigID
		edge   AigFanout
		add    bool
	}{
		{InvalidAigID, AigFanout{Object: fanout}, true},
		{source, AigFanout{Object: AigID(g.ObjectCount())}, true},
		{source, AigFanout{Object: fanout, Pin: 2}, true},
		{source, original[edge], true},
		{source, AigFanout{Object: fanout, Pin: 0}, false},
	}
	for _, row := range bad {
		if row.add {
			err = index.Add(row.source, row.edge)
		} else {
			err = index.Remove(row.source, row.edge)
		}
		if err == nil || !reflect.DeepEqual(index, before) {
			t.Fatal(row, err)
		}
	}
	if list, err := index.Fanouts(AigID(g.ObjectCount())); list != nil || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(list, err)
	}
	if list, err := (*AigFanoutIndex)(nil).Fanouts(0); list != nil || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(list, err)
	}
	for _, options := range []AigTraversalOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if result, err := BuildAigFanoutIndex(g, options); result != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	buffered.CreateCO(buffer)
	if result, err := BuildAigFanoutIndex(buffered, AigTraversalOptions{}); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
}

func TestAigFanoutDeepAndWide(t *testing.T) {
	const width = 10000
	g := NewAigManager()
	common, _ := g.CreateCI()
	for i := 0; i < width; i++ {
		other, _ := g.CreateCI()
		g.And(common, other)
	}
	index, err := BuildAigFanoutIndex(g, AigTraversalOptions{MaxObjects: 30000, MaxVisits: 100000})
	if err != nil {
		t.Fatal(err)
	}
	list, err := index.Fanouts(common.ID())
	if err != nil || len(list) != width || index.EdgeCount() != 2*width {
		t.Fatal(len(list), index.EdgeCount(), err)
	}
	for i, edge := range list {
		if edge.Object != AigID(3+2*i) || edge.Pin != 0 {
			t.Fatal(i, edge)
		}
	}
}

func FuzzAigFanout(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		signals := make([]AigLit, 8)
		for i := range signals {
			signals[i], _ = g.CreateCI()
		}
		for i := 0; i+1 < len(data) && len(signals) < 72; i += 2 {
			a := signals[int(data[i])%len(signals)].NotIf(data[i]&0x80 != 0)
			b := signals[int(data[i+1])%len(signals)].NotIf(data[i+1]&0x80 != 0)
			if result, err := g.And(a, b); err == nil {
				signals = append(signals, result)
			}
		}
		for i := 0; i < len(signals) && i < 4; i++ {
			g.CreateCO(signals[len(signals)-1-i])
		}
		index, err := BuildAigFanoutIndex(g, AigTraversalOptions{})
		if err != nil {
			t.Fatal(err)
		}
		expected := make([][]AigFanout, g.ObjectCount())
		for id, object := range g.objects {
			switch object.Type {
			case AigAND:
				expected[object.Fanin0.ID()] = append(expected[object.Fanin0.ID()], AigFanout{Object: AigID(id), Pin: 0})
				expected[object.Fanin1.ID()] = append(expected[object.Fanin1.ID()], AigFanout{Object: AigID(id), Pin: 1})
			case AigCO:
				expected[object.Fanin0.ID()] = append(expected[object.Fanin0.ID()], AigFanout{Object: AigID(id), Pin: 0})
			}
		}
		edges := 0
		for id := range expected {
			actual, err := index.Fanouts(AigID(id))
			if err != nil || !slices.Equal(actual, expected[id]) || len(actual) != int(g.objects[id].RefCount) {
				t.Fatal(id, actual, expected[id], g.objects[id].RefCount, err)
			}
			edges += len(actual)
		}
		if edges != index.EdgeCount() {
			t.Fatal(edges, index.EdgeCount())
		}
		for source, list := range expected {
			if len(list) == 0 {
				continue
			}
			edge := list[len(list)/2]
			if index.Remove(AigID(source), edge) != nil || index.Add(AigID(source), edge) != nil {
				t.Fatal(source, edge)
			}
			actual, _ := index.Fanouts(AigID(source))
			want := append(slices.Clone(list[:len(list)/2]), list[len(list)/2+1:]...)
			want = append(want, edge)
			if !slices.Equal(actual, want) {
				t.Fatal(actual, want)
			}
			break
		}
	})
}
