package abc

// Tests for justification and ternary simulation from aig/aig/aigJust.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigJustifyFixture(t testing.TB, seed int) (*AigManager, []AigLit, AigID) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 6)
	leaf := make([]AigLit, 6)
	for i := range ci {
		ci[i], _ = g.CreateCI()
		leaf[i] = ci[i].NotIf(seed&(1<<uint(i%5)) != 0)
	}
	node := make([]AigLit, 7)
	node[0], _ = g.And(leaf[0], leaf[1])
	node[1], _ = g.And(leaf[0].Not(), leaf[2])
	node[2], _ = g.And(node[0], node[1])
	node[3], _ = g.And(leaf[3], leaf[4])
	node[4], _ = g.And(node[2].NotIf(seed&2 != 0), node[3].NotIf(seed&4 != 0))
	node[5], _ = g.And(node[0].NotIf(seed&8 != 0), node[3].NotIf(seed&16 != 0))
	node[6], _ = g.And(node[4].NotIf(seed&1 != 0), node[5].NotIf(seed&2 != 0))
	co, _ := g.CreateCO(node[6].NotIf(seed&4 != 0))
	return g, node, co
}

func TestAigJustificationNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_justify.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, Root, Desired, Success, SupportValue int
		Support, Pattern                           []SatLit
		Values                                     []AigTernary
		Objects                                    [][]int
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g, nodes, co := aigJustifyFixture(t, row.Seed)
		root := nodes[2].ID()
		if row.Seed&3 == 0 {
			root = nodes[2].ID()
		} else if row.Seed&3 == 1 {
			root = nodes[4].ID()
		} else if row.Seed&3 == 2 {
			root = nodes[6].ID()
		} else {
			root = co
		}
		if int(root) != row.Root {
			t.Fatal(row.Seed, root, row.Root)
		}
		if len(row.Objects) != len(g.objects) {
			t.Fatal(row.Seed, len(row.Objects), len(g.objects))
		}
		ciOrdinal := make(map[AigID]int, len(g.cis))
		for i, id := range g.cis {
			ciOrdinal[id] = i
		}
		for id, object := range g.objects {
			got := []int{int(object.Type), -1, -1, -1}
			if object.Type == AigAND || object.Type == AigCO {
				got[1] = int(object.Fanin0)
			}
			if object.Type == AigAND {
				got[2] = int(object.Fanin1)
			}
			if object.Type == AigCI {
				got[3] = ciOrdinal[AigID(id)]
			}
			if !slices.Equal(got, row.Objects[id]) {
				t.Fatal(row.Seed, id, got, row.Objects[id])
			}
		}
		support, success, err := g.FindJustification(root, row.Desired != 0, AigJustificationOptions{})
		if err != nil || success != (row.Success != 0) || !slices.Equal(support, row.Support) {
			t.Fatal(row.Seed, support, success, err, row.Support, row.Success)
		}
		value, err := g.TernarySimulate(root, support, AigJustificationOptions{})
		if err != nil || value != AigTernary(row.SupportValue) {
			t.Fatal(row.Seed, value, err, row.SupportValue)
		}
		for id, want := range row.Values {
			value, err = g.TernarySimulate(AigID(id), row.Pattern, AigJustificationOptions{})
			if err != nil || value != want {
				t.Fatal(row.Seed, id, value, err, want)
			}
		}
	}
}

func TestAigJustificationErrorsLimitsOwnershipAndDeep(t *testing.T) {
	g, nodes, _ := aigJustifyFixture(t, 3)
	before := g.Clone()
	for _, options := range []AigJustificationOptions{{MaxVisits: -1}, {MaxLiterals: -1}, {MaxVisits: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }}} {
		literals, _, err := g.FindJustification(nodes[6].ID(), true, options)
		if literals != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, literals, err)
		}
	}
	for _, options := range []AigJustificationOptions{{MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if value, err := g.TernarySimulate(nodes[6].ID(), nil, options); value != 0 || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, value, err)
		}
	}
	if !reflect.DeepEqual(g, before) {
		t.Fatal("source mutation")
	}
	for _, root := range []AigID{AigID(g.ObjectCount()), InvalidAigID} {
		if got, _, err := g.FindJustification(root, true, AigJustificationOptions{}); got != nil || !errors.Is(err, ErrInvalidLiteral) {
			t.Fatal(root, got, err)
		}
		if _, err := g.TernarySimulate(root, nil, AigJustificationOptions{}); !errors.Is(err, ErrInvalidLiteral) {
			t.Fatal(root, err)
		}
	}
	if got, _, err := (*AigManager)(nil).FindJustification(0, true, AigJustificationOptions{}); got != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(got, err)
	}
	if value, err := g.TernarySimulate(nodes[6].ID(), []SatLit{SatLit(99) << 1}, AigJustificationOptions{}); value != 0 || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(value, err)
	}
	// Later duplicate assignments replace earlier ones.
	value, err := g.TernarySimulate(g.cis[0], []SatLit{0, 1}, AigJustificationOptions{})
	if err != nil || value != AigTernaryZero {
		t.Fatal(value, err)
	}
	if value, err = g.TernarySimulate(0, nil, AigJustificationOptions{}); err != nil || value != AigTernaryOne {
		t.Fatal(value, err)
	}
	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	buffered.CreateCO(input)
	// A disconnected buffer is harmless; a reached one is explicitly unsupported.
	if _, ok, err := buffered.FindJustification(input.ID(), true, AigJustificationOptions{}); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if got, _, err := buffered.FindJustification(buffer.ID(), true, AigJustificationOptions{}); got != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(got, err)
	}

	const depth = 10000
	deep := NewAigManager()
	root, _ := deep.CreateCI()
	assignment := []SatLit{0}
	for i := 0; i < depth; i++ {
		input, _ := deep.CreateCI()
		root, _ = deep.And(root, input)
		literal, _ := MakeSatLit(SatVar(i+1), false)
		assignment = append(assignment, literal)
	}
	literals, ok, err := deep.FindJustification(root.ID(), true, AigJustificationOptions{MaxVisits: 30000, MaxLiterals: 20000})
	if err != nil || !ok || len(literals) != depth+1 {
		t.Fatal(len(literals), ok, err)
	}
	if value, err = deep.TernarySimulate(root.ID(), assignment, AigJustificationOptions{MaxVisits: 30000}); err != nil || value != AigTernaryOne {
		t.Fatal(value, err)
	}
}

func FuzzAigJustification(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		signals := make([]AigLit, 6)
		for i := range signals {
			signals[i], _ = g.CreateCI()
		}
		for i := 0; i+1 < len(data) && len(signals) < 70; i += 2 {
			left := signals[int(data[i]&0x3f)%len(signals)].NotIf(data[i]&0x80 != 0)
			right := signals[int(data[i+1]&0x3f)%len(signals)].NotIf(data[i+1]&0x80 != 0)
			if literal, err := g.And(left, right); err == nil {
				signals = append(signals, literal)
			}
		}
		root := signals[len(signals)-1].ID()
		pattern := make([]SatLit, 0, 6)
		for i := 0; i < 6 && i < len(data); i++ {
			if data[i]&2 == 0 {
				literal, _ := MakeSatLit(SatVar(i), data[i]&1 != 0)
				pattern = append(pattern, literal)
			}
		}
		value, err := g.TernarySimulate(root, pattern, AigJustificationOptions{})
		if err != nil {
			t.Fatal(err)
		}
		known := make([]bool, len(g.objects))
		truth := make([]bool, len(g.objects))
		known[0], truth[0] = true, true
		for _, literal := range pattern {
			id := g.cis[literal.Var()]
			known[id], truth[id] = true, !literal.Negative()
		}
		read := func(literal AigLit) (bool, bool) {
			return known[literal.ID()], truth[literal.ID()] != literal.Complemented()
		}
		for id, object := range g.objects {
			if object.Type != AigAND {
				continue
			}
			ka, a := read(object.Fanin0)
			kb, b := read(object.Fanin1)
			if (ka && !a) || (kb && !b) {
				known[id], truth[id] = true, false
			} else if ka && kb {
				known[id], truth[id] = true, true
			}
		}
		want := AigTernaryUnknown
		if known[root] {
			if truth[root] {
				want = AigTernaryOne
			} else {
				want = AigTernaryZero
			}
		}
		if value != want {
			t.Fatal(value, want)
		}
		for _, desired := range []bool{false, true} {
			support, success, err := g.FindJustification(root, desired, AigJustificationOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if success && root != 0 {
				forced, err := g.TernarySimulate(root, support, AigJustificationOptions{})
				if err != nil || forced != map[bool]AigTernary{false: AigTernaryZero, true: AigTernaryOne}[desired] {
					t.Fatal(desired, support, forced, err)
				}
			}
		}
	})
}
