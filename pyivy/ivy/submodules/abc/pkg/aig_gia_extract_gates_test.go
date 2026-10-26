package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaExtractGatesNativeShapesAndTruth(t *testing.T) {
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	c, _ := source.AppendCI()
	n, _ := source.AppendAnd(a, b)
	r, _ := source.AppendAnd(n, c)
	for _, l := range []GiaLit{n, r} {
		if _, err := source.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	source.Name = "gate-cones"
	source.Spec = "source.aig"
	before := slices.Clone(source.objects)
	raw, err := os.ReadFile("test_vectors/oracle/gia_extract_gates.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []struct {
		Inputs, Objects int
		Drivers         []GiaLit
	}
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 10 {
		t.Fatal(err)
	}
	for k, roots := range [][]GiaID{{n.ID()}, {r.ID(), n.ID()}, {a.ID()}, {0}, nil} {
		for trim := 0; trim < 2; trim++ {
			result, err := source.ExtractGateCones(roots, trim != 0, GiaCofactorOptions{})
			if err != nil {
				t.Fatal(err)
			}
			graph := result.Graph
			row := native[2*k+trim]
			if graph.InputCount() != row.Inputs || graph.ObjectCount() != row.Objects || graph.OutputCount() != len(roots) || graph.Name != source.Name || graph.Spec != source.Spec || len(result.Registers) != 0 {
				t.Fatal("native gate cone shape", roots, trim, row)
			}
			drivers := make([]GiaLit, 0)
			for _, id := range graph.cos {
				l, err := graph.objects[id].Fanin0(id)
				if err != nil {
					t.Fatal(err)
				}
				drivers = append(drivers, l)
			}
			if !slices.Equal(drivers, row.Drivers) {
				t.Fatal("native extracted drivers", drivers, row.Drivers)
			}
			for assignment := 0; assignment < 8; assignment++ {
				bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
				expected := []bool{false, bits[0], bits[1], bits[2], bits[0] && bits[1], bits[0] && bits[1] && bits[2]}
				projected := make([]bool, 0)
				for _, pi := range result.PrimaryInputs {
					projected = append(projected, bits[pi])
				}
				values, err := graph.Evaluate(projected)
				if err != nil {
					t.Fatal(err)
				}
				for i, root := range roots {
					if values[i] != expected[root] {
						t.Fatal("extracted gate truth", roots, trim, assignment)
					}
				}
			}
		}
	}
	if !slices.Equal(before, source.objects) {
		t.Fatal("source mutated")
	}
	for _, roots := range [][]GiaID{{source.cos[0]}, {GiaID(source.ObjectCount())}} {
		if _, err := source.ExtractGateCones(roots, true, GiaCofactorOptions{}); err == nil {
			t.Fatal("invalid gate root")
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}} {
		if _, err := source.ExtractGateCones([]GiaID{n.ID()}, true, options); err == nil {
			t.Fatal("invalid bound")
		}
	}
	for _, threshold := range []int{1, 3, 5} {
		calls := 0
		result, err := source.ExtractGateCones([]GiaID{r.ID()}, true, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || result.Graph != nil {
			t.Fatal("gate extraction cancellation", threshold, err)
		}
	}
	if err := source.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	result, err := source.ExtractGateCones([]GiaID{r.ID()}, true, GiaCofactorOptions{})
	if err != nil || result.Graph.ConstraintCount() != 0 {
		t.Fatal("arbitrary gate output inherited constraint", err)
	}
	if err := source.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if _, err := source.ExtractGateCones([]GiaID{n.ID()}, true, GiaCofactorOptions{}); err == nil {
		t.Fatal("sequential extraction accepted")
	}
}
