package abc

// Tests for the complete operation surface from aig/aig/aigOper.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

type aigOperationObject [6]int

func aigOperationObjects(g *AigManager) []aigOperationObject {
	result := make([]aigOperationObject, len(g.objects))
	for id, object := range g.objects {
		fanin0, fanin1 := -1, -1
		if object.Fanin0.Valid() {
			fanin0 = int(object.Fanin0)
		}
		if object.Fanin1.Valid() {
			fanin1 = int(object.Fanin1)
		}
		phase := 0
		if object.Phase {
			phase = 1
		}
		result[id] = aigOperationObject{int(object.Type), fanin0, fanin1, int(object.RefCount), int(object.Level), phase}
	}
	return result
}

func TestAigOperationsNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_operations.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed, Nvars                                                 int
		OperAnd, OperXor, LookupExisting, LookupMissing             int
		Mux2, Mux, Majority, MultiAnd, MultiXor                     int
		MiterPairs, MiterTwo, CreateAnd, CreateOr, CreateXor, Check int
		Cis                                                         []AigID
		Objects                                                     []aigOperationObject
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g := NewAigManager()
		var variable [7]AigLit
		for i := range variable {
			variable[i], err = g.IthVar(i)
			if err != nil {
				t.Fatal(row.Seed, i, err)
			}
		}
		a := variable[0].NotIf(row.Seed&1 != 0)
		b := variable[1].NotIf(row.Seed&2 != 0)
		c := variable[2].NotIf(row.Seed&4 != 0)
		d := variable[3].NotIf(row.Seed&8 != 0)
		e := variable[4].NotIf(row.Seed&16 != 0)
		operAnd, _ := g.Operate(a, b, AigAND)
		operXor, _ := g.Operate(c, d, AigXOR)
		lookupExisting, found, err := g.LookupAnd(a, b)
		if err != nil || !found {
			t.Fatal(row.Seed, lookupExisting, found, err)
		}
		lookupMissing, missingFound, err := g.LookupAnd(a, c)
		if err != nil {
			t.Fatal(row.Seed, err)
		}
		mux2, _ := g.Mux2(c, operAnd, operXor)
		mux, _ := g.Mux(c, operAnd, operXor)
		majority, _ := g.Majority(a, b, e)
		arguments := []AigLit{a, b, c, d, e}
		multiAnd, _ := g.Multi(arguments, AigAND)
		multiXor, _ := g.Multi(arguments, AigXOR)
		miterPairs, err := g.MiterPairs([][2]AigLit{{a, b}, {c, d}, {e, variable[5]}}, AigOperationOptions{})
		if err != nil {
			t.Fatal(row.Seed, err)
		}
		left := []AigLit{a, c, e}
		right := []AigLit{variable[6], b, d}
		miterTwo, err := g.Miter(left, right, AigOperationOptions{})
		if err != nil {
			t.Fatal(row.Seed, err)
		}
		createAnd, _ := g.CreateAndFunction(row.Nvars, AigOperationOptions{})
		createOr, _ := g.CreateOrFunction(row.Nvars, AigOperationOptions{})
		createXor, _ := g.CreateXorFunction(row.Nvars, AigOperationOptions{})
		got := []int{
			int(operAnd), int(operXor), int(lookupExisting), -1,
			int(mux2), int(mux), int(majority), int(multiAnd), int(multiXor),
			int(miterPairs), int(miterTwo), int(createAnd), int(createOr), int(createXor),
		}
		if missingFound {
			got[3] = int(lookupMissing)
		}
		want := []int{
			row.OperAnd, row.OperXor, row.LookupExisting, row.LookupMissing,
			row.Mux2, row.Mux, row.Majority, row.MultiAnd, row.MultiXor,
			row.MiterPairs, row.MiterTwo, row.CreateAnd, row.CreateOr, row.CreateXor,
		}
		if !slices.Equal(got, want) || !slices.Equal(g.cis, row.Cis) || !slices.Equal(aigOperationObjects(g), row.Objects) || (g.Check() == nil) != (row.Check != 0) {
			t.Fatal(row.Seed, got, want, g.cis, row.Cis, aigOperationObjects(g), row.Objects, g.Check())
		}
		if !slices.Equal(arguments, []AigLit{a, b, c, d, e}) || !slices.Equal(left, []AigLit{a, c, e}) || !slices.Equal(right, []AigLit{variable[6], b, d}) {
			t.Fatal(row.Seed, "caller slice mutation")
		}
	}
}

func TestAigOperationsErrorsLimitsAndDeep(t *testing.T) {
	g := NewAigManager()
	a, _ := g.CreateCI()
	b, _ := g.CreateCI()
	before := g.Clone()
	if result, err := (*AigManager)(nil).Operate(a, b, AigAND); result != InvalidAigLit || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(result, err)
	}
	if result, err := g.Operate(a, b, AigCO); result != InvalidAigLit || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, found, err := g.LookupAnd(a, b); result != InvalidAigLit || found || err != nil {
		t.Fatal(result, found, err)
	}
	and, _ := g.And(a, b)
	if result, found, err := g.LookupAnd(b, a); err != nil || !found || result != and {
		t.Fatal(result, found, err, and)
	}
	if !reflect.DeepEqual(g, before) {
		// Account for the one deliberate AND, then preserve that checkpoint.
		before = g.Clone()
	}
	invalid := AigLit(g.ObjectCount()+10) << 1
	for _, test := range []struct {
		left, right []AigLit
		options     AigOperationOptions
		want        error
	}{
		{nil, nil, AigOperationOptions{}, ErrDimension},
		{[]AigLit{a}, nil, AigOperationOptions{}, ErrDimension},
		{[]AigLit{invalid}, []AigLit{b}, AigOperationOptions{}, ErrInvalidLiteral},
		{[]AigLit{a, b, a}, []AigLit{b, a, b}, AigOperationOptions{MaxOperations: -1}, ErrDimension},
		{[]AigLit{a, b, a}, []AigLit{b, a, b}, AigOperationOptions{MaxOperations: 4}, ErrLimit},
		{[]AigLit{a}, []AigLit{b}, AigOperationOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		checkpoint := g.Clone()
		result, err := g.Miter(test.left, test.right, test.options)
		if result != InvalidAigLit || !errors.Is(err, test.want) || !reflect.DeepEqual(g, checkpoint) {
			t.Fatal(test, result, err)
		}
	}
	expected, _ := g.Xor(a, b)
	if result, err := g.Miter([]AigLit{a, b, a}, []AigLit{b, a, b}, AigOperationOptions{MaxOperations: 5}); err != nil || result != expected {
		t.Fatal(result, err, expected)
	}
	for _, operation := range []func(int, AigOperationOptions) (AigLit, error){
		g.CreateAndFunction, g.CreateOrFunction, g.CreateXorFunction,
	} {
		for _, test := range []struct {
			variables int
			options   AigOperationOptions
			want      error
		}{
			{-1, AigOperationOptions{}, ErrDimension},
			{2, AigOperationOptions{MaxOperations: -1}, ErrDimension},
			{3, AigOperationOptions{MaxOperations: 2}, ErrLimit},
			{3, AigOperationOptions{Stop: func() bool { return true }}, ErrCancelled},
		} {
			checkpoint := g.Clone()
			result, err := operation(test.variables, test.options)
			if result != InvalidAigLit || !errors.Is(err, test.want) || !reflect.DeepEqual(g, checkpoint) {
				t.Fatal(test, result, err)
			}
		}
	}

	const depth = 10000
	deep := NewAigManager()
	root, err := deep.CreateAndFunction(depth, AigOperationOptions{MaxOperations: depth})
	if err != nil || deep.InputCount() != depth || deep.objects[root.ID()].Level != depth-1 || deep.Check() != nil {
		t.Fatal(root, deep.InputCount(), deep.objects[root.ID()].Level, err, deep.Check())
	}
	pairs := make([][2]AigLit, depth)
	for i := range pairs {
		pairs[i] = [2]AigLit{AigLit(deep.cis[i]) << 1, AigLit(deep.cis[i]) << 1}
	}
	result, err := deep.MiterPairs(pairs, AigOperationOptions{MaxOperations: 2*depth - 1})
	if err != nil || result != AigConst0 {
		t.Fatal(result, err)
	}
}

func TestAigCreateFunctionsSemantics(t *testing.T) {
	for variables := 0; variables <= 7; variables++ {
		for kind, create := range []func(*AigManager, int) (AigLit, error){
			func(g *AigManager, n int) (AigLit, error) { return g.CreateAndFunction(n, AigOperationOptions{}) },
			func(g *AigManager, n int) (AigLit, error) { return g.CreateOrFunction(n, AigOperationOptions{}) },
			func(g *AigManager, n int) (AigLit, error) { return g.CreateXorFunction(n, AigOperationOptions{}) },
		} {
			g := NewAigManager()
			root, err := create(g, variables)
			if err != nil {
				t.Fatal(kind, variables, err)
			}
			g.CreateCO(root)
			for assignment := 0; assignment < 1<<variables; assignment++ {
				inputs := make([]bool, variables)
				ones := 0
				for i := range inputs {
					inputs[i] = assignment&(1<<i) != 0
					if inputs[i] {
						ones++
					}
				}
				output, err := g.Evaluate(inputs)
				want := ones == variables
				if kind == 1 {
					want = ones != 0
				} else if kind == 2 {
					want = ones&1 != 0
				}
				if err != nil || output[0] != want {
					t.Fatal(kind, variables, assignment, output, want, err)
				}
			}
		}
	}
}

func FuzzAigOperationMiter(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		ci := make([]AigLit, 4)
		for i := range ci {
			ci[i], _ = g.CreateCI()
		}
		count := max(1, min(len(data), 16))
		left, right := make([]AigLit, count), make([]AigLit, count)
		for i := 0; i < count; i++ {
			value := byte(i)
			if len(data) != 0 {
				value = data[i%len(data)]
			}
			left[i] = ci[int(value&3)].NotIf(value&0x10 != 0)
			right[i] = ci[int((value>>2)&3)].NotIf(value&0x20 != 0)
		}
		root, err := g.Miter(left, right, AigOperationOptions{})
		if err != nil {
			t.Fatal(err)
		}
		g.CreateCO(root)
		for assignment := 0; assignment < 16; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
			want := false
			for i := range left {
				lv := inputs[left[i].ID()-1] != left[i].Complemented()
				rv := inputs[right[i].ID()-1] != right[i].Complemented()
				want = want || lv != rv
			}
			output, err := g.Evaluate(inputs)
			if err != nil || output[0] != want {
				t.Fatal(assignment, output, want, err)
			}
		}
	})
}
