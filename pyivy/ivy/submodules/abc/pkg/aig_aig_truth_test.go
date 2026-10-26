package abc

// Tests for cut truth evaluation from aig/aig/aigTruth.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigTruthFixture(t testing.TB, seed int) (*AigManager, AigLit) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 10)
	leaf := make([]AigLit, 10)
	for i := range ci {
		ci[i], _ = g.CreateCI()
		leaf[i] = ci[i].NotIf(seed>>uint(i%5)&1 != 0)
	}
	node := make([]AigLit, 9)
	node[0], _ = g.And(leaf[0], leaf[1])
	node[1], _ = g.And(leaf[2], leaf[3])
	node[2], _ = g.And(node[0], node[1])
	node[3], _ = g.And(leaf[4], leaf[5])
	node[4], _ = g.And(leaf[6], leaf[7])
	node[5], _ = g.And(node[3], node[4])
	node[6], _ = g.And(node[2], node[5])
	node[7], _ = g.And(leaf[8], leaf[9])
	node[8], _ = g.And(node[6], node[7])
	return g, node[8]
}

func TestAigCutTruthNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_truth.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cuts []struct {
			Seed          int
			Root          AigID
			Leaves, Nodes []AigID
			Truth         []uint64
		}
		One []struct {
			Operation                         AigType
			ComplementFirst, ComplementSecond int
			First, Second, Result             []uint64
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil || len(fixture.Cuts) != 32 || len(fixture.One) != 5 {
		t.Fatal(len(fixture.Cuts), len(fixture.One), err)
	}
	for _, row := range fixture.Cuts {
		g, root := aigTruthFixture(t, row.Seed)
		before := g.Clone()
		result, cutErr := g.CutTruth(root, row.Leaves, AigCutTruthOptions{})
		if cutErr != nil || root.ID() != row.Root || !slices.Equal(result.Nodes, row.Nodes) || !slices.Equal(result.Table.Words(), row.Truth) {
			t.Fatal(row.Seed, root, result.Nodes, row.Nodes, result.Table, row.Truth, cutErr)
		}
		if !reflect.DeepEqual(g, before) {
			t.Fatal(row.Seed, "source changed")
		}
	}
	for _, row := range fixture.One {
		first, firstErr := TruthTableFromWords(7, row.First, 0)
		second, secondErr := TruthTableFromWords(7, row.Second, 0)
		result, oneErr := AigCutTruthOne(first, second, row.Operation, row.ComplementFirst != 0, row.ComplementSecond != 0)
		if firstErr != nil || secondErr != nil || oneErr != nil || !slices.Equal(result.Words(), row.Result) {
			t.Fatal(row, firstErr, secondErr, result, oneErr)
		}
	}
}

func TestAigCutTruthBoundariesLimitsAndCancellation(t *testing.T) {
	g, root := aigTruthFixture(t, 0)
	leaves := slices.Clone(g.cis)
	for _, row := range []struct {
		literal AigLit
		value   uint64
	}{{AigConst1, ^uint64(0)}, {AigConst0, 0}} {
		result, err := g.CutTruth(row.literal, nil, AigCutTruthOptions{})
		if err != nil || len(result.Nodes) != 0 || !slices.Equal(result.Table.Words(), []uint64{row.value}) {
			t.Fatal(row, result, err)
		}
	}
	result, err := g.CutTruth((AigLit(g.cis[1]) << 1).Not(), []AigID{g.cis[0], g.cis[1], g.cis[1]}, AigCutTruthOptions{})
	if err != nil || len(result.Nodes) != 0 || !slices.Equal(result.Table.Words(), []uint64{^truthVariablePattern(1)}) {
		t.Fatal(result, err)
	}
	if result, err := g.CutTruth(AigLit(g.cis[0])<<1, []AigID{g.cis[1]}, AigCutTruthOptions{}); result.Table != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	co, _ := g.CreateCO(root)
	buffer, _ := g.CreateBuffer(root)
	for _, badLeaves := range [][]AigID{{0}, {co}, {buffer.ID()}, {AigID(g.ObjectCount())}} {
		if result, err := g.CutTruth(root, badLeaves, AigCutTruthOptions{}); result.Table != nil || err == nil {
			t.Fatal(badLeaves, result, err)
		}
	}
	if result, err := g.CutTruth(buffer, leaves, AigCutTruthOptions{}); result.Table != nil || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(result, err)
	}
	for _, options := range []AigCutTruthOptions{
		{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxWords: -1},
		{MaxObjects: 8}, {MaxMappingEntries: g.ObjectCount() - 1}, {MaxWords: 16},
		{Stop: func() bool { return true }},
	} {
		if result, err := g.CutTruth(root, leaves, options); result.Table != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, result, err)
		}
	}
	calls := 0
	if result, err := g.CutTruth(root, leaves, AigCutTruthOptions{Stop: func() bool { calls++; return calls == 10 }}); result.Table != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(calls, result, err)
	}
	if result, err := (*AigManager)(nil).CutTruth(root, leaves, AigCutTruthOptions{}); result.Table != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if table, err := AigCutTruthOne(nil, nil, AigAND, false, false); table != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(table, err)
	}
	a, _ := NewVariableTruthTable(2, 0, 0)
	b, _ := NewVariableTruthTable(2, 1, 0)
	for _, row := range []struct {
		op            AigType
		first, second bool
	}{{AigCI, false, false}, {AigXOR, true, false}, {AigXOR, false, true}} {
		if table, err := AigCutTruthOne(a, b, row.op, row.first, row.second); table != nil || err == nil {
			t.Fatal(row, table, err)
		}
	}
}

func FuzzAigCutTruth(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		signals := make([]AigLit, 6)
		for i := range signals {
			signals[i], _ = g.CreateCI()
		}
		for i := 0; i+1 < len(data) && len(signals) < 70; i += 2 {
			a := signals[int(data[i]&0x7f)%len(signals)].NotIf(data[i]&0x80 != 0)
			b := signals[int(data[i+1]&0x7f)%len(signals)].NotIf(data[i+1]&0x80 != 0)
			if value, err := g.And(a, b); err == nil {
				signals = append(signals, value)
			}
		}
		root := signals[len(signals)-1].NotIf(len(data)&1 != 0)
		result, err := g.CutTruth(root, g.cis, AigCutTruthOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 64; assignment++ {
			inputs := make([]bool, 6)
			for i := range inputs {
				inputs[i] = assignment>>i&1 != 0
			}
			want := aigTestLiteralValue(t, g, root, inputs)
			got, err := result.Table.Get(uint64(assignment))
			if err != nil || got != want {
				t.Fatal(assignment, got, want, err, result.Table.Words())
			}
		}
		fresh, err := g.CutTruth(root, g.cis, AigCutTruthOptions{})
		if err != nil || !reflect.DeepEqual(result, fresh) {
			t.Fatal(result, fresh, err)
		}
	})
}
