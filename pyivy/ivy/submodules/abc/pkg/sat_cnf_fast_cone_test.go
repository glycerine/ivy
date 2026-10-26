package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func checkFastConeTruth(t testing.TB, g *AigManager, root AigID, leaves []AigID, truth uint64) {
	t.Helper()
	for assignment := 0; assignment < 64; assignment++ {
		values := make([]bool, len(g.objects))
		values[0] = true
		for i, id := range g.cis {
			values[id] = assignment>>uint(i)&1 != 0
		}
		for id, obj := range g.objects {
			if obj.Type == AigAND {
				values[id] = (values[obj.Fanin0.ID()] != obj.Fanin0.Complemented()) && (values[obj.Fanin1.ID()] != obj.Fanin1.Complemented())
			}
		}
		index := 0
		for i, id := range leaves {
			if values[id] {
				index |= 1 << uint(i)
			}
		}
		if (truth>>uint(index)&1 != 0) != values[root] {
			t.Fatal(root, leaves, assignment, truth)
		}
	}
}
func TestFastCNFConeNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_fast_cone.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Root                 AigID
		Ops                  [][2]uint32
		Marked               []AigID
		Leaves, SignedLeaves []AigLit
		Nodes                []AigID
		Truth                *uint64
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		g := darBuildGraph(t, row.Ops, nil)
		g.CreateCO(AigLit(row.Root) << 1)
		before := slices.Clone(g.objects)
		regular, e := CollectFastCNFLeaves(g, row.Root, row.Marked, false, FastCNFConeOptions{})
		if e != nil || !slices.Equal(regular, row.Leaves) {
			t.Fatal(regular, row.Leaves, e)
		}
		signed, e := CollectFastCNFLeaves(g, row.Root, row.Marked, true, FastCNFConeOptions{})
		if e != nil || !slices.Equal(signed, row.SignedLeaves) {
			t.Fatal(signed, row.SignedLeaves, e)
		}
		var leaves []AigID
		for _, l := range regular {
			leaves = append(leaves, l.ID())
		}
		nodes, e := CollectFastCNFVolume(g, row.Root, leaves, FastCNFConeOptions{})
		if e != nil || !slices.Equal(nodes, row.Nodes) {
			t.Fatal(nodes, row.Nodes, e)
		}
		truth, e := DeriveFastCNFCutTruth(g, leaves, nodes, FastCNFConeOptions{})
		if row.Truth == nil {
			if e != ErrDimension {
				t.Fatal(e)
			}
		} else {
			if e != nil || truth != *row.Truth {
				t.Fatal(truth, *row.Truth, e)
			}
			checkFastConeTruth(t, g, row.Root, leaves, truth)
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("mutated graph")
		}
	}
}
func TestFastCNFConeBoundsAndDeep(t *testing.T) {
	g := mappedCNFFixture(t, 0, 0)
	marks := []AigID{0, 1, 2, 3, 4, 7}
	leaves, e := CollectFastCNFLeaves(g, 7, marks, false, FastCNFConeOptions{})
	if e != nil || len(leaves) != 4 {
		t.Fatal(leaves, e)
	}
	for _, o := range []FastCNFConeOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		r, e := CollectFastCNFLeaves(g, 7, marks, false, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	if r, e := CollectFastCNFLeaves(g, 7, nil, false, FastCNFConeOptions{}); r != nil || e != ErrDimension {
		t.Fatal(r, e)
	}
	if _, e = DeriveFastCNFCutTruth(g, []AigID{1, 1}, []AigID{5}, FastCNFConeOptions{}); e != ErrDimension {
		t.Fatal(e)
	}
	if _, e = DeriveFastCNFCutTruth(g, []AigID{1}, []AigID{5}, FastCNFConeOptions{}); e != ErrDimension {
		t.Fatal(e)
	}
	deep := NewAigManager()
	first, _ := deep.CreateCI()
	root := first
	marked := []AigID{0, first.ID()}
	for i := 0; i < 10000; i++ {
		input, _ := deep.CreateCI()
		marked = append(marked, input.ID())
		root, _ = deep.And(root, input)
	}
	collected, e := CollectFastCNFLeaves(deep, root.ID(), marked, false, FastCNFConeOptions{})
	if e != nil || len(collected) != 10001 {
		t.Fatal(len(collected), e)
	}
	ids := make([]AigID, len(collected))
	for i, l := range collected {
		ids[i] = l.ID()
	}
	nodes, e := CollectFastCNFVolume(deep, root.ID(), ids, FastCNFConeOptions{})
	if e != nil || len(nodes) != 10000 {
		t.Fatal(len(nodes), e)
	}
}
func FuzzFastCNFCone(f *testing.F) {
	f.Add([]byte{2, 4, 6, 8, 14, 16})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 80 {
			data = data[:80]
		}
		var ops [][2]uint32
		for i := 0; i+1 < len(data); i += 2 {
			ops = append(ops, [2]uint32{uint32(data[i]), uint32(data[i+1])})
		}
		g := darBuildGraph(t, ops, nil)
		var root AigID
		for id, obj := range g.objects {
			if obj.Type == AigAND {
				root = AigID(id)
			}
		}
		if root == 0 {
			return
		}
		g.CreateCO(AigLit(root) << 1)
		var marked []AigID
		for id, obj := range g.objects {
			if obj.Type == AigCI || obj.Type == AigConstant1 || obj.Type == AigAND && (id%3 == 0 || obj.RefCount > 1 || AigID(id) == root) {
				marked = append(marked, AigID(id))
			}
		}
		leaves, e := CollectFastCNFLeaves(g, root, marked, false, FastCNFConeOptions{})
		if e != nil {
			t.Fatal(e)
		}
		ids := make([]AigID, len(leaves))
		for i, l := range leaves {
			ids[i] = l.ID()
		}
		nodes, e := CollectFastCNFVolume(g, root, ids, FastCNFConeOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if len(ids) > 6 {
			return
		}
		truth, e := DeriveFastCNFCutTruth(g, ids, nodes, FastCNFConeOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkFastConeTruth(t, g, root, ids, truth)
	})
}
