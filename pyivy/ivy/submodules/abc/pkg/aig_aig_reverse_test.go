package abc

// Tests for reverse fanout DFS from aig/aig/aigDfs.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigReverseFixture(t testing.TB, seed uint32) *AigManager {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 6)
	for i := range ci {
		ci[i], _ = g.CreateCI()
	}
	nodes := make([]AigLit, 20)
	nodes[0], _ = g.And(ci[0], ci[1].NotIf(seed&1 != 0))
	state := seed
	next := func() uint32 {
		state = state*1664525 + 1013904223
		return state
	}
	for i := 1; i < len(nodes); i++ {
		a, b := int(next()%uint32(6+i)), int(next()%uint32(6+i))
		if a == b {
			b = (b + 1) % (6 + i)
		}
		left, right := ci[0], ci[0]
		if a < 6 {
			left = ci[a]
		} else {
			left = nodes[a-6]
		}
		if b < 6 {
			right = ci[b]
		} else {
			right = nodes[b-6]
		}
		nodes[i], _ = g.And(left.NotIf(state>>4&1 != 0), right.NotIf(state>>9&1 != 0))
	}
	g.And(ci[(int(seed)+1)%6].Not(), ci[(int(seed)+4)%6])
	g.And(ci[(int(seed)+2)%6].Not(), nodes[(int(seed)*3)%20])
	g.CreateCO(nodes[19].NotIf(seed&1 != 0))
	g.CreateCO(nodes[(int(seed)*7)%20])
	g.CreateCO(ci[(int(seed)+3)%6].Not())
	return g
}

func TestAigReverseDFSNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_dfs_reverse.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed    uint32
		Objects int
		Count   int
		Root    AigID
		Reverse []AigID
		Marks   []AigID
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g := aigReverseFixture(t, row.Seed)
		before := g.Clone()
		got, e := g.ReverseDFS(AigTraversalOptions{})
		if e != nil || g.ObjectCount() != row.Objects || !slices.Equal(got, row.Reverse) {
			t.Fatal(row.Seed, g.ObjectCount(), row.Objects, got, row.Reverse, e)
		}
		mask, e := g.ConeANDMask(row.Root, AigTraversalOptions{})
		if e != nil {
			t.Fatal(row.Seed, e)
		}
		var marks []AigID
		for id, marked := range mask {
			if marked {
				marks = append(marks, AigID(id))
			}
		}
		if !slices.Equal(marks, row.Marks) || len(marks) != row.Count {
			t.Fatal(row.Seed, marks, row.Marks, row.Count)
		}
		if !reflect.DeepEqual(g, before) {
			t.Fatal(row.Seed, "graph changed")
		}
	}
}

func TestAigReverseDFSBuffersLimitsAndDeep(t *testing.T) {
	g := NewAigManager()
	a, _ := g.CreateCI()
	b, _ := g.CreateCI()
	n0, _ := g.And(a, b)
	buffer, _ := g.CreateBuffer(n0)
	n1, _ := g.And(a.Not(), buffer)
	g.CreateCO(n1)
	g.CreateCO(n0)
	if got, err := g.ReverseDFS(AigTraversalOptions{}); err != nil || !slices.Equal(got, []AigID{n1.ID(), buffer.ID(), n0.ID()}) {
		t.Fatal(got, err)
	}
	for _, options := range []AigTraversalOptions{
		{MaxObjects: 1},
		{MaxVisits: 1},
		{Stop: func() bool { return true }},
	} {
		if got, err := g.ReverseDFS(options); got != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(got, err)
		}
		if got, err := g.ConeANDMask(n1.ID(), options); got != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(got, err)
		}
	}
	if got, err := (*AigManager)(nil).ReverseDFS(AigTraversalOptions{}); got != nil || err == nil {
		t.Fatal(got, err)
	}
	if got, err := g.ConeANDMask(AigID(g.ObjectCount()), AigTraversalOptions{}); got != nil || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(got, err)
	}
	if mask, err := g.ConeANDMask(buffer.ID(), AigTraversalOptions{}); err != nil || slices.Contains(mask, true) {
		t.Fatal(mask, err)
	}

	deep := NewAigManager()
	root, _ := deep.CreateCI()
	for i := 0; i < 10000; i++ {
		input, _ := deep.CreateCI()
		root, _ = deep.And(root, input)
	}
	deep.CreateCO(root)
	got, err := deep.ReverseDFS(AigTraversalOptions{MaxObjects: 30000, MaxVisits: 100000})
	if err != nil || len(got) != 10000 || got[0] != root.ID() {
		t.Fatal(len(got), err)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] <= got[i] {
			t.Fatal(i, got[i-1], got[i])
		}
	}
}

func FuzzAigReverseDFS(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		g := NewAigManager()
		literals := make([]AigLit, 4)
		for i := range literals {
			literals[i], _ = g.CreateCI()
		}
		for i := 0; i+2 < len(data); i += 3 {
			a := literals[int(data[i])%len(literals)].NotIf(data[i+2]&1 != 0)
			b := literals[int(data[i+1])%len(literals)].NotIf(data[i+2]&2 != 0)
			next, err := g.And(a, b)
			if err != nil {
				t.Fatal(err)
			}
			if data[i+2]&4 != 0 {
				next, err = g.CreateBuffer(next)
				if err != nil {
					t.Fatal(err)
				}
			}
			literals = append(literals, next)
		}
		if len(data)&1 != 0 {
			g.CreateCO(literals[len(literals)-1])
		}
		before := g.Clone()
		order, err := g.ReverseDFS(AigTraversalOptions{})
		if err != nil || !reflect.DeepEqual(g, before) {
			t.Fatal(err)
		}
		position := make([]int, g.ObjectCount())
		for i := range position {
			position[i] = -1
		}
		for i, id := range order {
			if position[id] != -1 {
				t.Fatal("duplicate", id)
			}
			position[id] = i
		}
		for id, object := range g.objects {
			if object.Type == AigAND || object.Type == AigBuffer {
				if position[id] < 0 {
					t.Fatal("missing", id)
				}
			}
			for _, fanin := range []AigLit{object.Fanin0, object.Fanin1} {
				if !fanin.Valid() || int(fanin.ID()) >= len(g.objects) {
					continue
				}
				if (object.Type == AigAND || object.Type == AigBuffer) && (g.objects[fanin.ID()].Type == AigAND || g.objects[fanin.ID()].Type == AigBuffer) && position[AigID(id)] >= position[fanin.ID()] {
					t.Fatal("fanout order", fanin.ID(), id)
				}
			}
		}
	})
}
