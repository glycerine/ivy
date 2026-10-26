package abc

// Tests for reconvergence-driven cuts from aig/aig/aigWin.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigWindowFixture(seed int) (*AigManager, AigID) {
	g := NewAigManager()
	ci := make([]AigLit, 6)
	leaf := make([]AigLit, 6)
	for i := range ci {
		ci[i], _ = g.CreateCI()
		leaf[i] = ci[i].NotIf(seed&(1<<uint(i%5)) != 0)
	}
	node := make([]AigLit, 7)
	node[0], _ = g.And(leaf[0], leaf[1])
	node[1], _ = g.And(leaf[2], leaf[3])
	node[2], _ = g.And(node[0], node[1])
	node[3], _ = g.And(node[0], leaf[4])
	node[4], _ = g.And(node[2], node[3])
	node[5], _ = g.And(node[1], leaf[5])
	node[6], _ = g.And(node[4], node[5])
	g.CreateCO(node[6])
	for bit, index := range []int{0, 1, 2, 3, 4} {
		if seed&(1<<uint(bit)) != 0 {
			g.CreateCO(node[index])
		}
	}
	return g, node[6].ID()
}

func TestAigWindowNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_window.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, SizeLimit, FanoutLimit int
		Front, Visited               []AigID
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g, root := aigWindowFixture(row.Seed)
		before := g.Clone()
		result, err := g.FindReconvergentCut(root, row.SizeLimit, row.FanoutLimit, AigWindowOptions{})
		if err != nil || !slices.Equal(result.Leaves, row.Front) || !slices.Equal(result.Visited, row.Visited) {
			t.Fatal(row.Seed, result, err, row)
		}
		if !reflect.DeepEqual(g, before) {
			t.Fatal(row.Seed, "source mutation")
		}
	}
}

func TestAigWindowErrorsOwnershipAndDeep(t *testing.T) {
	g, root := aigWindowFixture(3)
	for _, test := range []struct {
		manager      *AigManager
		root         AigID
		size, fanout int
		options      AigWindowOptions
		want         error
	}{
		{nil, root, 2, 1, AigWindowOptions{}, ErrDimension},
		{g, root, 1, 1, AigWindowOptions{}, ErrDimension},
		{g, root, 2, -1, AigWindowOptions{}, ErrDimension},
		{g, root, 2, 1, AigWindowOptions{MaxWork: -1}, ErrDimension},
		{g, InvalidAigID, 2, 1, AigWindowOptions{}, ErrInvalidLiteral},
		{g, g.cis[0], 2, 1, AigWindowOptions{}, ErrUnsupported},
		{g, root, 6, 3, AigWindowOptions{MaxWork: 1}, ErrLimit},
		{g, root, 6, 3, AigWindowOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		result, err := test.manager.FindReconvergentCut(test.root, test.size, test.fanout, test.options)
		if result.Leaves != nil || result.Visited != nil || !errors.Is(err, test.want) {
			t.Fatal(test, result, err)
		}
	}
	result, err := g.FindReconvergentCut(root, 6, 3, AigWindowOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result.Leaves[0] = 0
	result.Visited[0] = 0
	again, err := g.FindReconvergentCut(root, 6, 3, AigWindowOptions{})
	if err != nil || again.Leaves[0] == 0 || again.Visited[0] == 0 {
		t.Fatal(again, err)
	}

	buffered := NewAigManager()
	a, _ := buffered.CreateCI()
	b, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(a)
	bufferRoot, _ := buffered.And(buffer, b)
	buffered.CreateCO(bufferRoot)
	if result, err := buffered.FindReconvergentCut(bufferRoot.ID(), 3, 1, AigWindowOptions{}); result.Leaves != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}

	const depth = 10000
	deep := NewAigManager()
	a, _ = deep.CreateCI()
	b, _ = deep.CreateCI()
	chain, _ := deep.And(a, b)
	for i := 1; i < depth; i++ {
		chain, _ = deep.And(chain, b)
	}
	deep.CreateCO(chain)
	deepResult, err := deep.FindReconvergentCut(chain.ID(), 2, 0, AigWindowOptions{MaxWork: 30000})
	if err != nil || !slices.Equal(deepResult.Leaves, []AigID{b.ID(), a.ID()}) || len(deepResult.Visited) != depth+2 {
		t.Fatal(len(deepResult.Visited), deepResult.Leaves, err)
	}
}

func aigWindowReference(g *AigManager, root AigID, sizeLimit, fanoutLimit int) AigWindowCut {
	rootObject := g.objects[root]
	result := AigWindowCut{
		Leaves:  []AigID{rootObject.Fanin0.ID(), rootObject.Fanin1.ID()},
		Visited: []AigID{root, rootObject.Fanin0.ID(), rootObject.Fanin1.ID()},
	}
	marked := map[AigID]bool{root: true, result.Leaves[0]: true, result.Leaves[1]: true}
	for {
		best, bestCost := -1, 100
		for index, id := range result.Leaves {
			object := g.objects[id]
			if object.Type == AigCI {
				continue
			}
			cost := 0
			if !marked[object.Fanin0.ID()] {
				cost++
			}
			if !marked[object.Fanin1.ID()] {
				cost++
			}
			if cost == 2 && int(object.RefCount) > fanoutLimit {
				continue
			}
			if cost < bestCost || (cost == bestCost && g.objects[id].Level > g.objects[result.Leaves[best]].Level) {
				best, bestCost = index, cost
			}
			if cost == 0 {
				break
			}
		}
		if best < 0 || len(result.Leaves)-1+bestCost > sizeLimit {
			return result
		}
		object := g.objects[result.Leaves[best]]
		result.Leaves = append(result.Leaves[:best], result.Leaves[best+1:]...)
		for _, id := range []AigID{object.Fanin0.ID(), object.Fanin1.ID()} {
			if !marked[id] {
				marked[id] = true
				result.Leaves = append(result.Leaves, id)
				result.Visited = append(result.Visited, id)
			}
		}
	}
}

func FuzzAigWindow(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		signals := make([]AigLit, 6)
		for i := range signals {
			signals[i], _ = g.CreateCI()
		}
		var roots []AigLit
		for i := 0; i+1 < len(data) && len(signals) < 40; i += 2 {
			left := signals[int(data[i]&0x3f)%len(signals)].NotIf(data[i]&0x80 != 0)
			right := signals[int(data[i+1]&0x3f)%len(signals)].NotIf(data[i+1]&0x80 != 0)
			before := g.AndCount()
			literal, err := g.And(left, right)
			if err == nil {
				signals = append(signals, literal)
				if g.AndCount() > before {
					roots = append(roots, literal)
				}
			}
		}
		if len(roots) == 0 {
			return
		}
		root := roots[len(roots)-1]
		g.CreateCO(root)
		for i, literal := range roots[:len(roots)-1] {
			if len(data) != 0 && data[i%len(data)]&4 != 0 {
				g.CreateCO(literal)
			}
		}
		sizeLimit, fanoutLimit := 2, 0
		if len(data) != 0 {
			sizeLimit = 2 + int(data[0]&7)
			fanoutLimit = int(data[len(data)-1] & 3)
		}
		want := aigWindowReference(g, root.ID(), sizeLimit, fanoutLimit)
		got, err := g.FindReconvergentCut(root.ID(), sizeLimit, fanoutLimit, AigWindowOptions{})
		if err != nil || !slices.Equal(got.Leaves, want.Leaves) || !slices.Equal(got.Visited, want.Visited) || len(got.Leaves) > sizeLimit {
			t.Fatal(got, err, want)
		}
	})
}
