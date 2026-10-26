package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaPiCofactorNativeAndTruth(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	r, _ := g.AppendAnd(n, c)
	for _, l := range []GiaLit{n, r} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	g.Name = "pi-cofactor"
	g.Spec = "source.aig"
	before := slices.Clone(g.objects)
	raw, err := os.ReadFile("test_vectors/oracle/gia_cofactor_pi.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]uint32
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 6 {
		t.Fatal(err)
	}
	for index := 0; index < 3; index++ {
		for fixed := 0; fixed < 2; fixed++ {
			graph, err := g.CofactorPrimaryInput(index, fixed != 0, GiaCofactorOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if graph.InputCount() != 3 || graph.OutputCount() != 2 || graph.ConstraintCount() != 1 || graph.Name != g.Name || graph.Spec != g.Spec {
				t.Fatal("PI cofactor metadata")
			}
			masks := make([]uint32, 2)
			for assignment := 0; assignment < 8; assignment++ {
				bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
				values, err := graph.Evaluate(bits)
				if err != nil {
					t.Fatal(err)
				}
				bits[index] = fixed != 0
				if values[0] != (bits[0] && bits[1]) || values[1] != (bits[0] && bits[1] && bits[2]) {
					t.Fatal("fixed PI truth", index, fixed, assignment, values)
				}
				for output, value := range values {
					if value {
						masks[output] |= 1 << assignment
					}
				}
			}
			if !slices.Equal(masks, native[2*index+fixed]) {
				t.Fatal("native PI cofactor", index, fixed, masks, native[2*index+fixed])
			}
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("PI cofactor mutated source")
	}
	for _, index := range []int{-1, 3} {
		if _, err := g.CofactorPrimaryInput(index, false, GiaCofactorOptions{}); err == nil {
			t.Fatal("invalid PI index")
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}} {
		if _, err := g.CofactorPrimaryInput(0, false, options); err == nil {
			t.Fatal("invalid caps")
		}
	}
	for _, threshold := range []int{1, 5, 9} {
		calls := 0
		graph, err := g.CofactorPrimaryInput(0, true, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || graph != nil {
			t.Fatal("PI cofactor cancellation", threshold, err)
		}
	}
}
func TestGiaPiCofactorSequentialNativeMux(t *testing.T) {
	g := NewGiaManager()
	control, _ := g.AppendCI()
	data, _ := g.AppendCI()
	state, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(control, data, state)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{mux, data} {
		if _, err = g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = g.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	graph, err := g.CofactorPrimaryInput(0, true, GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if graph.RegisterCount() != 1 || graph.ConstraintCount() != 1 {
		t.Fatal("lost sequential metadata")
	}
	for assignment := 0; assignment < 8; assignment++ {
		values, err := graph.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if err != nil || values[0] != (assignment&2 != 0) || values[1] != (assignment&2 != 0) {
			t.Fatal(values, err)
		}
	}
	if _, err := g.CofactorPrimaryInput(2, false, GiaCofactorOptions{}); err == nil {
		t.Fatal("register output accepted as PI")
	}
}
