package abc

// Tests for choice-chain traversal and levels from aig/aig/aigDfs.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigChoiceFixture(t testing.TB, seed int) (*AigManager, AigID, AigID, AigID) {
	t.Helper()
	g := NewAigManager()
	leaf := make([]AigLit, 6)
	for i := range leaf {
		leaf[i], _ = g.CreateCI()
		leaf[i] = leaf[i].NotIf(seed>>i&1 != 0)
	}
	ab, _ := g.And(leaf[0], leaf[1])
	cd, _ := g.And(leaf[2], leaf[3])
	head, _ := g.And(ab, cd)
	bcd, _ := g.And(leaf[1], cd)
	alt, _ := g.And(leaf[0], bcd)
	abc, _ := g.And(ab, leaf[2])
	third, _ := g.And(abc, leaf[3])
	ef, _ := g.And(leaf[4], leaf[5])
	root, _ := g.And(head, ef)
	g.CreateCO(root.NotIf(seed&1 != 0))
	if err := g.SetChoiceChains([][]AigID{{head.ID(), alt.ID(), third.ID()}}, AigTraversalOptions{}); err != nil {
		t.Fatal(err)
	}
	return g, head.ID(), alt.ID(), third.ID()
}

func TestAigChoicesNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_choices.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, Objects, Maximum, SmallSupport int
		Head, Alt, Third                     AigID
		PostNodes, PostAll                   []AigID
		PreNodes, PreAll                     []AigID
		Choices                              []AigID
		Levels                               []uint32
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g, head, alt, third := aigChoiceFixture(t, row.Seed)
		if g.ObjectCount() != row.Objects || head != row.Head || alt != row.Alt || third != row.Third || !reflect.DeepEqual(g.ChoiceChains(), [][]AigID{{head, alt, third}}) {
			t.Fatal(row.Seed, g.ObjectCount(), head, alt, third, g.ChoiceChains())
		}
		before := g.Clone()
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
		choices, e := g.DFSChoices(AigTraversalOptions{})
		if e != nil || !slices.Equal(choices, row.Choices) {
			t.Fatal(row.Seed, choices, row.Choices, e)
		}
		levels, e := g.ComputeChoiceLevels(AigTraversalOptions{})
		if e != nil || int(levels.Maximum) != row.Maximum || !slices.Equal(levels.Levels, row.Levels) {
			t.Fatal(row.Seed, levels, row.Maximum, row.Levels, e)
		}
		small, e := g.SmallSupportANDCount(AigTraversalOptions{})
		if e != nil || small != row.SmallSupport {
			t.Fatal(row.Seed, small, row.SmallSupport, e)
		}
		if !reflect.DeepEqual(g, before) {
			t.Fatal(row.Seed, "choice query changed graph")
		}
	}
}

func TestAigChoicesOwnershipCleanupAndErrors(t *testing.T) {
	g, head, alt, third := aigChoiceFixture(t, 0)
	before := g.Clone()
	for _, chains := range [][][]AigID{
		{{head}},
		{{head, alt}, {third, alt}},
		{{head, AigID(g.ObjectCount())}},
		{{g.cis[0], alt}},
	} {
		if err := g.SetChoiceChains(chains, AigTraversalOptions{}); err == nil || !reflect.DeepEqual(g, before) {
			t.Fatal(chains, err)
		}
	}
	for _, options := range []AigTraversalOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		copy := before.Clone()
		if err := copy.SetChoiceChains([][]AigID{{head, alt, third}}, options); err == nil || !reflect.DeepEqual(copy, before) {
			t.Fatal(options, err)
		}
		if result, err := g.DFSChoices(options); result != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(result, err)
		}
		if result, err := g.ComputeChoiceLevels(options); result != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(result, err)
		}
	}
	if err := (*AigManager)(nil).SetChoiceChains(nil, AigTraversalOptions{}); err == nil {
		t.Fatal(err)
	}
	if result, err := (*AigManager)(nil).DFSChoices(AigTraversalOptions{}); result != nil || err == nil {
		t.Fatal(result, err)
	}

	clone := g.Clone()
	if err := clone.SetChoiceChains(nil, AigTraversalOptions{}); err != nil || clone.ChoiceChains() != nil || len(g.ChoiceChains()) != 1 {
		t.Fatal(err, clone.ChoiceChains(), g.ChoiceChains())
	}
	if result, err := clone.DFSChoices(AigTraversalOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	newInput, _ := g.CreateCI()
	newNode, _ := g.And(AigLit(head)<<1, newInput)
	if newNode == InvalidAigLit || g.Check() != nil || !reflect.DeepEqual(g.ChoiceChains(), [][]AigID{{head, alt, third}}) {
		t.Fatal(newNode, g.Check(), g.ChoiceChains())
	}
	if removed := g.Cleanup(); removed != 1 {
		t.Fatal("choice cleanup removed", removed)
	}
	if _, err := g.Object(alt); err != nil {
		t.Fatal("choice alternative removed", err)
	}
	if _, err := g.Object(third); err != nil {
		t.Fatal("choice alternative removed", err)
	}
	if g.Check() != nil {
		t.Fatal(g.Check())
	}

	corrupt := before.Clone()
	corrupt.equivs[third] = head
	if err := corrupt.Check(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	corrupt = before.Clone()
	corrupt.equivs[third] = alt
	if err := corrupt.Check(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestAigChoiceLevelsDependencyCycleAndDeep(t *testing.T) {
	g, head, alt, _ := aigChoiceFixture(t, 0)
	cycle, _ := g.And(AigLit(head)<<1, AigLit(alt)<<1)
	links := make([]AigID, g.ObjectCount())
	for i := range links {
		links[i] = InvalidAigID
	}
	links[head] = cycle.ID()
	g.equivs = links
	if _, err := g.ComputeChoiceLevels(AigTraversalOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}

	deep := NewAigManager()
	leaves := make([]AigLit, 10002)
	for i := range leaves {
		leaves[i], _ = deep.CreateCI()
	}
	headLit, _ := deep.And(leaves[0], leaves[1])
	chain := []AigID{headLit.ID()}
	for i := 2; i < len(leaves); i++ {
		next, _ := deep.And(leaves[0], leaves[i])
		chain = append(chain, next.ID())
	}
	deep.CreateCO(headLit)
	if err := deep.SetChoiceChains([][]AigID{chain}, AigTraversalOptions{MaxObjects: 30000, MaxVisits: 100000}); err != nil {
		t.Fatal(err)
	}
	order, err := deep.DFSChoices(AigTraversalOptions{MaxObjects: 30000, MaxVisits: 100000})
	if err != nil || len(order) != len(chain) {
		t.Fatal(len(order), len(chain), err)
	}
	levels, err := deep.ComputeChoiceLevels(AigTraversalOptions{MaxObjects: 30000, MaxVisits: 100000})
	if err != nil || levels.Maximum != 1 {
		t.Fatal(levels, err)
	}
}

func FuzzAigChoices(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(31))
	f.Fuzz(func(t *testing.T, code uint8) {
		g, head, alt, third := aigChoiceFixture(t, int(code&31))
		order, err := g.DFSChoices(AigTraversalOptions{})
		if err != nil {
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
		for _, id := range order {
			object := g.objects[id]
			for _, fanin := range []AigLit{object.Fanin0, object.Fanin1} {
				if g.objects[fanin.ID()].Type == AigAND && position[fanin.ID()] >= position[id] {
					t.Fatal("fanin order", fanin.ID(), id)
				}
			}
			if next := g.choiceNext(id); next != InvalidAigID && position[next] >= position[id] {
				t.Fatal("choice order", next, id)
			}
		}
		levels, err := g.ComputeChoiceLevels(AigTraversalOptions{})
		if err != nil || levels.Maximum != 4 || levels.Levels[head] != 3 || levels.Levels[alt] != 3 || levels.Levels[third] != 3 {
			t.Fatal(levels, err)
		}
		for assignment := 0; assignment < 64; assignment++ {
			inputs := make([]bool, 6)
			for i := range inputs {
				inputs[i] = assignment>>i&1 != 0
			}
			value := aigTestLiteralValue(t, g, AigLit(head)<<1, inputs)
			if value != aigTestLiteralValue(t, g, AigLit(alt)<<1, inputs) || value != aigTestLiteralValue(t, g, AigLit(third)<<1, inputs) {
				t.Fatal("non-equivalent choice", assignment)
			}
		}
	})
}
