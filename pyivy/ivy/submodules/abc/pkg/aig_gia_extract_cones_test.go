package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func giaExtractTestGraph(t *testing.T) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	input, _ := g.AppendCI()
	r0, _ := g.AppendCI()
	r1, _ := g.AppendCI()
	r2, _ := g.AppendCI()
	n, _ := g.AppendAnd(input, r0)
	s, _ := g.AppendAnd(r1, r2)
	for _, l := range []GiaLit{n, s, r1, r0, GiaConst0} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.SetRegisterCount(3); err != nil {
		t.Fatal(err)
	}
	g.Name = "extract"
	g.Spec = "source.aig"
	return g
}
func TestGiaExtractConesNativeAndProjection(t *testing.T) {
	source := giaExtractTestGraph(t)
	before := slices.Clone(source.objects)
	raw, err := os.ReadFile("test_vectors/oracle/gia_extract_cones.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []struct {
		Inputs, Registers, Objects int
		Drivers                    []GiaLit
	}
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 8 {
		t.Fatal(err)
	}
	for k, selection := range [][]int{{0}, {1}, {1, 0}, nil} {
		for trim := 0; trim < 2; trim++ {
			extracted, err := source.ExtractOutputCones(selection, trim != 0, GiaCofactorOptions{})
			if err != nil {
				t.Fatal(err)
			}
			graph := extracted.Graph
			row := native[2*k+trim]
			if graph.InputCount() != row.Inputs || graph.RegisterCount() != row.Registers || graph.ObjectCount() != row.Objects || graph.Name != source.Name || graph.Spec != source.Spec {
				t.Fatal("native extracted shape", selection, trim, graph.InputCount(), graph.RegisterCount(), graph.ObjectCount(), row)
			}
			drivers := make([]GiaLit, 0, graph.OutputCount())
			for _, id := range graph.cos {
				l, err := graph.objects[id].Fanin0(id)
				if err != nil {
					t.Fatal(err)
				}
				drivers = append(drivers, l)
			}
			if !slices.Equal(drivers, row.Drivers) {
				t.Fatal("native driver/register order", drivers, row.Drivers)
			}
			if k == 1 && !slices.Equal(extracted.Registers, []int{1, 2, 0}) {
				t.Fatal("register discovery", extracted.Registers)
			}
			for assignment := 0; assignment < 16; assignment++ {
				full := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
				expected, err := source.Evaluate(full)
				if err != nil {
					t.Fatal(err)
				}
				projected := make([]bool, 0, graph.InputCount())
				for _, pi := range extracted.PrimaryInputs {
					projected = append(projected, full[pi])
				}
				for _, reg := range extracted.Registers {
					projected = append(projected, full[1+reg])
				}
				values, err := graph.Evaluate(projected)
				if err != nil {
					t.Fatal(err)
				}
				for i, output := range selection {
					if values[i] != expected[output] {
						t.Fatal("projected output", selection, trim, assignment)
					}
				}
				for i, reg := range extracted.Registers {
					if values[len(selection)+i] != expected[2+reg] {
						t.Fatal("projected transition", selection, trim, assignment)
					}
				}
			}
		}
	}
	if !slices.Equal(before, source.objects) {
		t.Fatal("extraction mutated scratch")
	}
}
func TestGiaExtractConesValidationAndConstraints(t *testing.T) {
	source := giaExtractTestGraph(t)
	for _, selection := range [][]int{{-1}, {2}} {
		if result, err := source.ExtractOutputCones(selection, true, GiaCofactorOptions{}); err == nil || result.Graph != nil {
			t.Fatal("invalid output")
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}} {
		if result, err := source.ExtractOutputCones([]int{1}, true, options); err == nil || result.Graph != nil {
			t.Fatal("invalid bound")
		}
	}
	for _, threshold := range []int{1, 3, 6} {
		calls := 0
		result, err := source.ExtractOutputCones([]int{1}, true, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || result.Graph != nil {
			t.Fatal("extraction cancellation", threshold, err)
		}
	}
	if err := source.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	result, err := source.ExtractOutputCones([]int{1}, true, GiaCofactorOptions{})
	if err != nil || result.Graph.ConstraintCount() != 1 {
		t.Fatal("selected constraint", err)
	}
	if _, err := source.ExtractOutputCones([]int{1, 0}, true, GiaCofactorOptions{}); err == nil {
		t.Fatal("constraint ordering")
	}
}

func TestGiaExtractConesNativeMuxAndDuplicateOutputs(t *testing.T) {
	source := NewGiaManager()
	input, _ := source.AppendCI()
	control, _ := source.AppendCI()
	data, _ := source.AppendCI()
	mux, err := source.AppendMuxReal(control, input, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{mux, control, GiaConst0} {
		if _, err = source.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	if err = source.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	extracted, err := source.ExtractOutputCones([]int{0, 0}, true, GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if extracted.Graph.MuxCount() != 1 || extracted.Graph.PrimaryOutputCount() != 2 || !slices.Equal(extracted.Registers, []int{1, 0}) {
		t.Fatal("native cone roles", extracted.Registers)
	}
	for assignment := 0; assignment < 8; assignment++ {
		full := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		projected := []bool{full[extracted.PrimaryInputs[0]], full[1+extracted.Registers[0]], full[1+extracted.Registers[1]]}
		want, err := source.Evaluate(full)
		if err != nil {
			t.Fatal(err)
		}
		got, err := extracted.Graph.Evaluate(projected)
		if err != nil || got[0] != want[0] || got[1] != want[0] || got[2] != want[1+extracted.Registers[0]] || got[3] != want[1+extracted.Registers[1]] {
			t.Fatal("native cone transition", assignment, got, want, err)
		}
	}
}
