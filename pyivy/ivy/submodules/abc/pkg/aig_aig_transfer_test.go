package abc

// Tests for cone transfer and CI composition from aig/aig/aigDfs.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigTransferComposeFixture(t testing.TB, seed int) (*AigManager, AigLit, AigLit) {
	t.Helper()
	g := NewAigManager()
	ci := make([]AigLit, 6)
	for i := range ci {
		var err error
		ci[i], err = g.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	n0, _ := g.And(ci[0].NotIf(seed&1 != 0), ci[1])
	n1, _ := g.And(ci[2], ci[3].NotIf(seed&2 != 0))
	n2, _ := g.And(ci[4].NotIf(seed&4 != 0), ci[5])
	n3, _ := g.And(n0.Not(), n1.Not())
	n4, _ := g.And(n3.Not(), n2)
	n5, _ := g.And(n0, n2.Not())
	root := n4
	if seed&8 != 0 {
		root = n5
	}
	root = root.NotIf(seed&16 != 0)
	return g, root, ci[(seed+2)%6].NotIf(seed&2 != 0)
}

func aigOracleObjects(g *AigManager) [][3]int {
	result := make([][3]int, 0, len(g.objects))
	for _, object := range g.objects {
		switch object.Type {
		case AigConstant1:
			result = append(result, [3]int{0, -1, -1})
		case AigCI:
			result = append(result, [3]int{1, -1, -1})
		case AigAND:
			result = append(result, [3]int{2, int(object.Fanin0), int(object.Fanin1)})
		default:
			result = append(result, [3]int{9, -1, -1})
		}
	}
	return result
}

func aigTestLiteralValue(t testing.TB, g *AigManager, literal AigLit, inputs []bool) bool {
	t.Helper()
	if len(inputs) != len(g.cis) || !g.validSignal(literal) {
		t.Fatal("invalid literal evaluation")
	}
	values := make([]bool, len(g.objects))
	values[0] = true
	for i, id := range g.cis {
		values[id] = inputs[i]
	}
	read := func(value AigLit) bool { return values[value.ID()] != value.Complemented() }
	for id, object := range g.objects {
		switch object.Type {
		case AigAND:
			values[id] = read(object.Fanin0) && read(object.Fanin1)
		case AigBuffer, AigCO:
			values[id] = read(object.Fanin0)
		}
	}
	return read(literal)
}

func TestAigTransferComposeNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_transfer_compose.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, Root, Function, Variable, TransferBefore, TransferLit, ComposeBefore, ComposeLit int
		TransferObjects, ComposeObjects                                                        [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		source, root, function := aigTransferComposeFixture(t, row.Seed)
		if int(root) != row.Root || int(function) != row.Function {
			t.Fatal(row.Seed, root, function)
		}
		sourceBefore := source.Clone()
		destination := NewAigManager()
		ci := make([]AigLit, 6)
		for i := range ci {
			ci[i], _ = destination.CreateCI()
		}
		destination.And(ci[0], ci[1].Not())
		if destination.ObjectCount() != row.TransferBefore {
			t.Fatal(row.Seed, destination.ObjectCount(), row.TransferBefore)
		}
		transferred, e := TransferAigCone(source, destination, root, 6, AigRewriteOptions{})
		if e != nil || int(transferred) != row.TransferLit || !slices.Equal(aigOracleObjects(destination), row.TransferObjects) {
			t.Fatal(row.Seed, transferred, row.TransferLit, aigOracleObjects(destination), row.TransferObjects, e)
		}
		if !reflect.DeepEqual(source, sourceBefore) {
			t.Fatal(row.Seed, "transfer changed source")
		}

		composedSource, composeRoot, composeFunction := aigTransferComposeFixture(t, row.Seed)
		if composedSource.ObjectCount() != row.ComposeBefore {
			t.Fatal(row.Seed, composedSource.ObjectCount(), row.ComposeBefore)
		}
		composed, e := composedSource.ComposeAigInput(composeRoot, composeFunction, row.Variable, AigRewriteOptions{})
		if e != nil || int(composed) != row.ComposeLit || !slices.Equal(aigOracleObjects(composedSource), row.ComposeObjects) {
			t.Fatal(row.Seed, composed, row.ComposeLit, aigOracleObjects(composedSource), row.ComposeObjects, e)
		}
		for assignment := 0; assignment < 64; assignment++ {
			inputs := make([]bool, 6)
			for i := range inputs {
				inputs[i] = assignment>>i&1 != 0
			}
			if aigTestLiteralValue(t, source, root, inputs) != aigTestLiteralValue(t, destination, transferred, inputs) {
				t.Fatal(row.Seed, assignment, "transfer semantics")
			}
			replaced := slices.Clone(inputs)
			replaced[row.Variable] = aigTestLiteralValue(t, source, function, inputs)
			if aigTestLiteralValue(t, source, root, replaced) != aigTestLiteralValue(t, composedSource, composed, inputs) {
				t.Fatal(row.Seed, assignment, "compose semantics")
			}
		}
	}
}

func TestAigTransferComposeErrorsAtomicAndSpecial(t *testing.T) {
	source, root, function := aigTransferComposeFixture(t, 1)
	destination := NewAigManager()
	destinationBefore := destination.Clone()
	for _, call := range []func() (AigLit, error){
		func() (AigLit, error) { return TransferAigCone(nil, destination, root, 6, AigRewriteOptions{}) },
		func() (AigLit, error) { return TransferAigCone(source, nil, root, 6, AigRewriteOptions{}) },
		func() (AigLit, error) { return TransferAigCone(source, destination, root, -1, AigRewriteOptions{}) },
		func() (AigLit, error) {
			return TransferAigCone(source, destination, InvalidAigLit, 6, AigRewriteOptions{})
		},
		func() (AigLit, error) { return TransferAigCone(source, destination, root, 1, AigRewriteOptions{}) },
		func() (AigLit, error) {
			return TransferAigCone(source, destination, root, 6, AigRewriteOptions{MaxObjects: 1})
		},
		func() (AigLit, error) {
			return TransferAigCone(source, destination, root, 6, AigRewriteOptions{MaxVisits: 1})
		},
		func() (AigLit, error) {
			return TransferAigCone(source, destination, root, 6, AigRewriteOptions{Stop: func() bool { return true }})
		},
	} {
		if got, e := call(); got != InvalidAigLit || e == nil {
			t.Fatal(got, e)
		}
		if !reflect.DeepEqual(destination, destinationBefore) {
			t.Fatal("failed transfer mutated destination")
		}
	}
	if got, e := TransferAigCone(source, source, root, 0, AigRewriteOptions{}); e != nil || got != root {
		t.Fatal(got, e)
	}
	if got, e := TransferAigCone(source, destination, AigConst0, 0, AigRewriteOptions{}); e != nil || got != AigConst0 || !reflect.DeepEqual(destination, destinationBefore) {
		t.Fatal(got, e)
	}
	if got, e := TransferAigCone(source, destination, root, 6, AigRewriteOptions{}); e != nil || !destination.validSignal(got) || destination.InputCount() != 6 {
		t.Fatal(got, destination.InputCount(), e)
	}

	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	before := destination.Clone()
	if got, e := TransferAigCone(buffered, destination, buffer, 1, AigRewriteOptions{}); got != InvalidAigLit || !errors.Is(e, ErrDimension) || !reflect.DeepEqual(destination, before) {
		t.Fatal(got, e)
	}

	for _, call := range []func(*AigManager) (AigLit, error){
		func(g *AigManager) (AigLit, error) { return g.ComposeAigInput(root, function, -1, AigRewriteOptions{}) },
		func(g *AigManager) (AigLit, error) { return g.ComposeAigInput(root, function, 6, AigRewriteOptions{}) },
		func(g *AigManager) (AigLit, error) {
			return g.ComposeAigInput(InvalidAigLit, function, 0, AigRewriteOptions{})
		},
		func(g *AigManager) (AigLit, error) {
			return g.ComposeAigInput(root, InvalidAigLit, 0, AigRewriteOptions{})
		},
		func(g *AigManager) (AigLit, error) {
			return g.ComposeAigInput(root, function, 0, AigRewriteOptions{MaxObjects: 1})
		},
		func(g *AigManager) (AigLit, error) {
			return g.ComposeAigInput(root, function, 0, AigRewriteOptions{MaxVisits: 1})
		},
		func(g *AigManager) (AigLit, error) {
			return g.ComposeAigInput(root, function, 0, AigRewriteOptions{Stop: func() bool { return true }})
		},
	} {
		g := source.Clone()
		before = g.Clone()
		if got, e := call(g); got != InvalidAigLit || e == nil || !reflect.DeepEqual(g, before) {
			t.Fatal(got, e)
		}
	}
	if got, e := (*AigManager)(nil).ComposeAigInput(root, function, 0, AigRewriteOptions{}); got != InvalidAigLit || e == nil {
		t.Fatal(got, e)
	}
}

func TestAigTransferComposeDeepAndInternalFunction(t *testing.T) {
	g := NewAigManager()
	inputs := make([]AigLit, 10001)
	for i := range inputs {
		inputs[i], _ = g.CreateCI()
	}
	root := inputs[0]
	for i := 1; i < len(inputs); i++ {
		root, _ = g.And(root, inputs[i])
	}
	destination := NewAigManager()
	transferred, err := TransferAigCone(g, destination, root, len(inputs), AigRewriteOptions{MaxObjects: 30000, MaxVisits: 100000})
	if err != nil || transferred == InvalidAigLit || destination.AndCount() != 10000 {
		t.Fatal(transferred, destination.AndCount(), err)
	}

	small, smallRoot, _ := aigTransferComposeFixture(t, 7)
	internal := AigLit(9) << 1
	original := small.Clone()
	composed, err := small.ComposeAigInput(smallRoot, internal, 0, AigRewriteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 64; assignment++ {
		values := make([]bool, 6)
		for i := range values {
			values[i] = assignment>>i&1 != 0
		}
		replaced := slices.Clone(values)
		replaced[0] = aigTestLiteralValue(t, original, internal, values)
		if aigTestLiteralValue(t, original, smallRoot, replaced) != aigTestLiteralValue(t, small, composed, values) {
			t.Fatal(assignment)
		}
	}
}

func FuzzAigTransferCompose(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 1, 4, 2, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 96 {
			data = data[:96]
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
			literals = append(literals, next)
		}
		root := literals[len(literals)-1].NotIf(len(data)&1 != 0)
		function := literals[(len(literals)*3)/5].NotIf(len(data)&2 != 0)
		destination := NewAigManager()
		transferred, err := TransferAigCone(g, destination, root, 4, AigRewriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		composedGraph := g.Clone()
		variable := len(data) % 4
		composed, err := composedGraph.ComposeAigInput(root, function, variable, AigRewriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			inputs := make([]bool, 4)
			for i := range inputs {
				inputs[i] = assignment>>i&1 != 0
			}
			if aigTestLiteralValue(t, g, root, inputs) != aigTestLiteralValue(t, destination, transferred, inputs[:destination.InputCount()]) {
				t.Fatal("transfer", assignment)
			}
			replaced := slices.Clone(inputs)
			replaced[variable] = aigTestLiteralValue(t, g, function, inputs)
			if aigTestLiteralValue(t, g, root, replaced) != aigTestLiteralValue(t, composedGraph, composed, inputs) {
				t.Fatal("compose", assignment)
			}
		}
	})
}
