package abc

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigDFSFixture(t testing.TB, seed uint32) (*AigManager, AigID, []AigID, []AigID) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 6)
	for i := range ci {
		ci[i], _ = g.CreateCI()
	}
	previous, _ := g.And(ci[0], ci[1])
	middle := []AigLit{previous}
	for i := 2; i < 6; i++ {
		previous, _ = g.And(previous, ci[i])
		middle = append(middle, previous)
	}
	state := seed
	next := func() uint32 {
		state = state*1664525 + 1013904223
		return state
	}
	for i := 0; i < 12; i++ {
		a, b := int(next()%6), int(next()%6)
		if a == b {
			b = (b + 1) % 6
		}
		side, _ := g.And(ci[a], ci[b])
		previous, _ = g.And(previous, side)
	}
	g.CreateCO(previous.NotIf(seed&1 != 0))
	g.CreateCO(middle[int(seed)%5])
	g.CreateCO(ci[(int(seed)+2)%6].Not())
	cutLeaves := []AigID{middle[int(seed)%5].ID()}
	cutLeaves = append(cutLeaves, g.cis...)
	return g, previous.ID(), []AigID{g.cos[1], middle[(int(seed)+1)%5].ID()}, cutLeaves
}

func TestAigDFSNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/aig_dfs.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Seed                                      uint32
		Objects, Root, Topological                int
		DagSize, SupportSize, MaxLevel            int
		PostNodes, PostAll, PreNodes, PreAll, All []AigID
		Array, From, Support, Supports, Cut       []AigID
		Levels                                    [][]AigID
	}
	if e = json.Unmarshal(data, &rows); e != nil || len(rows) != 24 {
		t.Fatal(len(rows), e)
	}
	for _, row := range rows {
		g, root, explicit, cutLeaves := aigDFSFixture(t, row.Seed)
		before := slices.Clone(g.objects)
		postNodes, e := g.DFS(AigDFSOptions{NodesOnly: true})
		if e != nil || !slices.Equal(postNodes, row.PostNodes) {
			t.Fatal(row.Seed, postNodes, row.PostNodes, e)
		}
		postAll, e := g.DFS(AigDFSOptions{})
		if e != nil || !slices.Equal(postAll, row.PostAll) {
			t.Fatal(row.Seed, postAll, row.PostAll, e)
		}
		preNodes, e := g.DFS(AigDFSOptions{NodesOnly: true, Preorder: true})
		if e != nil || !slices.Equal(preNodes, row.PreNodes) {
			t.Fatal(row.Seed, preNodes, row.PreNodes, e)
		}
		preAll, e := g.DFS(AigDFSOptions{Preorder: true})
		if e != nil || !slices.Equal(preAll, row.PreAll) {
			t.Fatal(row.Seed, preAll, row.PreAll, e)
		}
		all, e := g.DFSAll(AigTraversalOptions{})
		if e != nil || !slices.Equal(all, row.All) {
			t.Fatal(row.Seed, all, row.All, e)
		}
		array, e := g.DFSArray(explicit, AigTraversalOptions{})
		if e != nil || !slices.Equal(array, row.Array) {
			t.Fatal(row.Seed, array, row.Array, e)
		}
		from, e := g.DFSNodes(explicit, AigTraversalOptions{})
		if e != nil || !slices.Equal(from, row.From) {
			t.Fatal(row.Seed, from, row.From, e)
		}
		support, e := g.Support([]AigID{root}, AigTraversalOptions{})
		if e != nil || !slices.Equal(support, row.Support) {
			t.Fatal(row.Seed, support, row.Support, e)
		}
		supports, e := g.Support(g.cos[:2], AigTraversalOptions{})
		if e != nil || !slices.Equal(supports, row.Supports) {
			t.Fatal(row.Seed, supports, row.Supports, e)
		}
		cut, e := g.CollectCut(root, cutLeaves, AigTraversalOptions{})
		if e != nil || !slices.Equal(cut, row.Cut) {
			t.Fatal(row.Seed, cut, row.Cut, e)
		}
		supportSize, e := g.SupportSize(root, AigTraversalOptions{})
		if e != nil || supportSize != row.SupportSize {
			t.Fatal(row.Seed, supportSize, row.SupportSize, e)
		}
		dagSize, e := g.DAGSize(root, AigTraversalOptions{})
		if e != nil || dagSize != row.DagSize {
			t.Fatal(row.Seed, dagSize, row.DagSize, e)
		}
		maximum, e := g.MaxOutputLevel()
		if e != nil || int(maximum) != row.MaxLevel {
			t.Fatal(row.Seed, maximum, row.MaxLevel, e)
		}
		levels, e := g.LevelBuckets(AigTraversalOptions{})
		if e != nil || !reflect.DeepEqual(levels, row.Levels) {
			t.Fatal(row.Seed, levels, row.Levels, e)
		}
		if e := g.VerifyTopologicalOrder(); e != nil || row.Topological != 1 || g.ObjectCount() != row.Objects || root != AigID(row.Root) {
			t.Fatal(row.Seed, root, e)
		}
		if !reflect.DeepEqual(before, g.objects) {
			t.Fatal("graph changed")
		}
	}
}

func TestAigDFSLimitsDanglingBuffersAndDeep(t *testing.T) {
	g, root, roots, cutLeaves := aigDFSFixture(t, 1)
	for _, o := range []AigTraversalOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if r, e := g.DFSNodes(roots, o); r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
		if r, e := g.Support([]AigID{root}, o); r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
		if n, e := g.DAGSize(root, o); n != 0 || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(n, e)
		}
		if r, e := g.CollectCut(root, cutLeaves, o); r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	if r, e := g.CollectCut(root, []AigID{cutLeaves[0], cutLeaves[0]}, AigTraversalOptions{}); r != nil || e != ErrDimension {
		t.Fatal(r, e)
	}
	if r, e := g.CollectCut(root, nil, AigTraversalOptions{}); r != nil || e != ErrDimension {
		t.Fatal(r, e)
	}
	for _, bad := range []AigID{InvalidAigID, AigID(g.ObjectCount())} {
		if r, e := g.DFSNodes([]AigID{bad}, AigTraversalOptions{}); r != nil || e != ErrInvalidLiteral {
			t.Fatal(r, e)
		}
		if r, e := g.Support([]AigID{bad}, AigTraversalOptions{}); r != nil || e != ErrInvalidLiteral {
			t.Fatal(r, e)
		}
	}

	dangling := NewAigManager()
	a, _ := dangling.CreateCI()
	b, _ := dangling.CreateCI()
	c, _ := dangling.CreateCI()
	dangling.And(a, b)
	dangling.CreateCO(c)
	if r, e := dangling.DFS(AigDFSOptions{NodesOnly: true}); r != nil || e != ErrInvariant {
		t.Fatal(r, e)
	}
	if r, e := dangling.DFSAll(AigTraversalOptions{}); r != nil || e != ErrInvariant {
		t.Fatal(r, e)
	}

	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	co, _ := buffered.CreateCO(buffer)
	from, e := buffered.DFSNodes([]AigID{co}, AigTraversalOptions{})
	if e != nil || !slices.Equal(from, []AigID{input.ID(), buffer.ID()}) {
		t.Fatal(from, e)
	}
	support, e := buffered.Support([]AigID{buffer.ID()}, AigTraversalOptions{})
	if e != nil || !slices.Equal(support, []AigID{input.ID()}) {
		t.Fatal(support, e)
	}
	if n, e := buffered.DAGSize(buffer.ID(), AigTraversalOptions{}); e != nil || n != 0 {
		t.Fatal(n, e)
	}

	deep := NewAigManager()
	chain, _ := deep.CreateCI()
	for i := 0; i < 10000; i++ {
		input, _ := deep.CreateCI()
		chain, e = deep.And(chain, input)
		if e != nil {
			t.Fatal(e)
		}
	}
	deep.CreateCO(chain)
	order, e := deep.DFS(AigDFSOptions{NodesOnly: true})
	if e != nil || len(order) != 10000 || order[len(order)-1] != chain.ID() {
		t.Fatal(len(order), e)
	}
	if n, e := deep.DAGSize(chain.ID(), AigTraversalOptions{}); e != nil || n != 10000 {
		t.Fatal(n, e)
	}
	if support, e := deep.Support([]AigID{chain.ID()}, AigTraversalOptions{}); e != nil || len(support) != 10001 {
		t.Fatal(len(support), e)
	}
}

func TestAigSupportThresholdCount(t *testing.T) {
	g := NewAigManager()
	root, _ := g.CreateCI()
	for i := 0; i < 20; i++ {
		input, _ := g.CreateCI()
		root, _ = g.And(root, input)
	}
	if got, err := g.SmallSupportANDCount(AigTraversalOptions{}); err != nil || got != 15 {
		t.Fatal(got, err)
	}
	for _, row := range [][2]int{{0, 0}, {1, 0}, {2, 1}, {7, 6}, {16, 15}, {21, 20}, {100, 20}} {
		got, err := g.CountANDsWithSupportAtMost(row[0], AigTraversalOptions{})
		if err != nil || got != row[1] {
			t.Fatal(row[0], got, row[1], err)
		}
	}
	if got, err := g.CountANDsWithSupportAtMost(-1, AigTraversalOptions{}); got != 0 || !errors.Is(err, ErrDimension) {
		t.Fatal(got, err)
	}
	for _, options := range []AigTraversalOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if got, err := g.SmallSupportANDCount(options); got != 0 || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(got, err)
		}
	}
}

func FuzzAigDFS(f *testing.F) {
	f.Add([]byte{0, 2, 4, 0, 1, 6, 8, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		var ops [][4]uint32
		for i := 0; i+3 < len(data); i += 4 {
			ops = append(ops, [4]uint32{uint32(data[i]), uint32(data[i+1]), uint32(data[i+2]), uint32(data[i+3])})
		}
		g := fastMarkGraph(t, ops, []uint32{uint32(2 * (6 + len(ops)))})
		root := g.objects[g.cos[0]].Fanin0.ID()
		before := slices.Clone(g.objects)
		order, e := g.DFSNodes([]AigID{g.cos[0]}, AigTraversalOptions{})
		if e != nil {
			t.Fatal(e)
		}
		support, e := g.Support([]AigID{root}, AigTraversalOptions{})
		if e != nil {
			t.Fatal(e)
		}
		dag, e := g.DAGSize(root, AigTraversalOptions{})
		if e != nil {
			t.Fatal(e)
		}
		positions := make([]int, len(g.objects))
		for i := range positions {
			positions[i] = -1
		}
		ands, cis := 0, 0
		var orderedAnds []AigID
		for i, id := range order {
			if positions[id] >= 0 {
				t.Fatal("duplicate", id)
			}
			positions[id] = i
			obj := g.objects[id]
			switch obj.Type {
			case AigCI:
				cis++
			case AigAND:
				ands++
				orderedAnds = append(orderedAnds, id)
				for _, fanin := range []AigLit{obj.Fanin0, obj.Fanin1} {
					if fanin.ID() != 0 && positions[fanin.ID()] < 0 {
						t.Fatal("not postorder", id, fanin)
					}
				}
			default:
				t.Fatal(obj.Type)
			}
		}
		if ands != dag || cis != len(support) || !reflect.DeepEqual(before, g.objects) {
			t.Fatal(ands, dag, cis, support)
		}
		wantSmall := 0
		for id, object := range g.objects {
			if object.Type != AigAND {
				continue
			}
			size, err := g.SupportSize(AigID(id), AigTraversalOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if size <= 16 {
				wantSmall++
			}
		}
		if small, err := g.SmallSupportANDCount(AigTraversalOptions{}); err != nil || small != wantSmall {
			t.Fatal(small, wantSmall, err)
		}
		if g.objects[root].Type == AigAND {
			cut, e := g.CollectCut(root, support, AigTraversalOptions{})
			if e != nil || !slices.Equal(cut, orderedAnds) {
				t.Fatal(cut, orderedAnds, e)
			}
		}
	})
}
