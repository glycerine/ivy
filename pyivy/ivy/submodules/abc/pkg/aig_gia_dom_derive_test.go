package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaDeriveDominatedConeCofactors(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	root, _ := g.AppendAnd(n, c)
	g.Name = "cone"
	g.Spec = "source.aig"
	raw, err := os.ReadFile("test_vectors/oracle/gia_dom_derive.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]uint32
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 3 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for variables := 0; variables <= 2; variables++ {
		result, err := g.DeriveDominatedCone(root.ID(), []GiaID{n.ID(), c.ID()}, variables, 0)
		if err != nil {
			t.Fatal(err)
		}
		if result.InputCount() != 2 || result.OutputCount() != 1<<variables || result.Name != g.Name || result.Spec != g.Spec {
			t.Fatal("cofactor metadata")
		}
		masks := make([]uint32, result.OutputCount())
		for input := 0; input < 4; input++ {
			values, err := result.Evaluate([]bool{input&1 != 0, input&2 != 0})
			if err != nil {
				t.Fatal(err)
			}
			for assignment, value := range values {
				if value {
					masks[assignment] |= 1 << input
				}
				left, right := input&1 != 0, input&2 != 0
				if variables >= 1 {
					left = assignment&1 != 0
				}
				if variables == 2 {
					right = assignment&2 != 0
				}
				if value != (left && right) {
					t.Fatal("cofactor semantics", variables, input, assignment)
				}
			}
		}
		if !slices.Equal(masks, native[variables]) {
			t.Fatal("native cofactor truth", masks, native[variables])
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("derivation mutated source")
	}
	for _, support := range [][]GiaID{{a.ID()}, {n.ID(), n.ID()}, {0, c.ID()}, {GiaID(g.ObjectCount())}} {
		if _, err := g.DeriveDominatedCone(root.ID(), support, 0, 0); err == nil {
			t.Fatal("invalid or incomplete support", support)
		}
	}
	for _, variables := range []int{-1, 3, 6} {
		if _, err := g.DeriveDominatedCone(root.ID(), []GiaID{n.ID(), c.ID()}, variables, 0); err == nil {
			t.Fatal("invalid cofactor count")
		}
	}
	if _, err := g.DeriveDominatedCone(root.ID(), []GiaID{n.ID(), c.ID()}, 0, 1); err == nil {
		t.Fatal("object cap")
	}
}
func TestGiaDeriveDominatedConeNativeMux(t *testing.T) {
	g := NewGiaManager()
	control, _ := g.AppendCI()
	yes, _ := g.AppendCI()
	no, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(control, yes, no)
	if err != nil {
		t.Fatal(err)
	}
	derived, err := g.DeriveDominatedCone(mux.ID(), []GiaID{control.ID(), yes.ID(), no.ID()}, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		values, err := derived.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
		if values[0] != (assignment&4 != 0) || values[1] != (assignment&2 != 0) {
			t.Fatal("mux cofactor", values)
		}
	}
}

func TestGiaDominatedConeFiveVariableLimit(t *testing.T) {
	g := NewGiaManager()
	inputs := make([]GiaLit, 5)
	support := make([]GiaID, 5)
	for i := range inputs {
		var err error
		inputs[i], err = g.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
		support[i] = inputs[i].ID()
	}
	root := inputs[0]
	for _, input := range inputs[1:] {
		var err error
		root, err = g.AppendAnd(root, input)
		if err != nil {
			t.Fatal(err)
		}
	}
	derived, err := g.DeriveDominatedCone(root.ID(), support, 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if derived.InputCount() != 5 || derived.OutputCount() != 32 {
		t.Fatal("five-variable dimensions")
	}
	values, err := derived.Evaluate([]bool{false, false, false, false, false})
	if err != nil {
		t.Fatal(err)
	}
	for output, value := range values {
		if value != (output == 31) {
			t.Fatal("five-variable cofactor", output, value)
		}
	}
	if partial, err := g.DeriveDominatedCone(root.ID(), support, 5, g.ObjectCount()); err == nil || partial != nil {
		t.Fatal("output allocation cap ignored")
	}
}
