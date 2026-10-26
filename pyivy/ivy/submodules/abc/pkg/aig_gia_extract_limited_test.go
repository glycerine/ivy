package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaLimitedConesNativeAndBoundaryProjection(t *testing.T) {
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	c, _ := source.AppendCI()
	n, _ := source.AppendAnd(a, b)
	r, _ := source.AppendAnd(n, c)
	source.Name = "limited"
	source.Spec = "source.aig"
	before := slices.Clone(source.objects)
	raw, err := os.ReadFile("test_vectors/oracle/gia_extract_limited.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []struct {
		Objects int
		Inputs  []GiaID
		Drivers []GiaLit
	}
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 16 {
		t.Fatal(err)
	}
	for mode := 0; mode < 2; mode++ {
		for order := 0; order < 2; order++ {
			for limit := 0; limit < 4; limit++ {
				roots := []GiaID{r.ID(), n.ID()}
				if order != 0 {
					roots[0], roots[1] = roots[1], roots[0]
				}
				var result GiaCutExtraction
				if mode == 0 {
					result, err = source.ExtractLevelLimitedCones(roots, uint32(limit), GiaCofactorOptions{})
				} else {
					result, err = source.ExtractDepthLimitedCones(roots, limit, GiaCofactorOptions{})
				}
				if err != nil {
					t.Fatal(err)
				}
				row := native[mode*8+order*4+limit]
				graph := result.Graph
				drivers := make([]GiaLit, 0)
				for _, id := range graph.cos {
					l, err := graph.objects[id].Fanin0(id)
					if err != nil {
						t.Fatal(err)
					}
					drivers = append(drivers, l)
				}
				if graph.ObjectCount() != row.Objects || !slices.Equal(result.Inputs, row.Inputs) || !slices.Equal(drivers, row.Drivers) {
					t.Fatal("native limited cut", mode, order, limit, result.Inputs, drivers, row)
				}
				if graph.RegisterCount() != 0 || graph.ConstraintCount() != 0 || graph.Name != source.Name || graph.Spec != source.Spec {
					t.Fatal("cut metadata")
				}
				for assignment := 0; assignment < 8; assignment++ {
					bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
					values := []bool{false, bits[0], bits[1], bits[2], bits[0] && bits[1], bits[0] && bits[1] && bits[2]}
					projected := make([]bool, 0)
					for _, id := range result.Inputs {
						projected = append(projected, values[id])
					}
					outputs, err := graph.Evaluate(projected)
					if err != nil || outputs[0] != values[roots[0]] || outputs[1] != values[roots[1]] {
						t.Fatal("cut projection", mode, order, limit, assignment, outputs, err)
					}
				}
			}
		}
	}
	if !slices.Equal(before, source.objects) {
		t.Fatal("cut mutated source")
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}} {
		if result, err := source.ExtractDepthLimitedCones([]GiaID{r.ID()}, 2, options); err == nil || result.Graph != nil {
			t.Fatal("cut bounds")
		}
	}
	if _, err := source.ExtractDepthLimitedCones([]GiaID{r.ID()}, -1, GiaCofactorOptions{}); err == nil {
		t.Fatal("negative depth")
	}
	if _, err := source.ExtractLevelLimitedCones([]GiaID{GiaID(source.ObjectCount())}, 0, GiaCofactorOptions{}); err == nil {
		t.Fatal("invalid root")
	}
	for _, threshold := range []int{1, 3, 7} {
		calls := 0
		result, err := source.ExtractDepthLimitedCones([]GiaID{r.ID()}, 3, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || result.Graph != nil {
			t.Fatal("cut cancellation", threshold, err)
		}
	}
}

func TestGiaLimitedConesConstantsAndRegisterBoundaries(t *testing.T) {
	source := NewGiaManager()
	pi, _ := source.AppendCI()
	ro, _ := source.AppendCI()
	mux, err := source.AppendMuxReal(ro, pi, GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = source.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	if _, err = source.AppendCO(pi); err != nil {
		t.Fatal(err)
	}
	if err = source.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	result, err := source.ExtractDepthLimitedCones([]GiaID{mux.ID(), 0}, 1, GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Graph.RegisterCount() != 0 || result.Graph.MuxCount() != 1 || len(result.Inputs) != 2 {
		t.Fatal("native boundary roles", result.Inputs)
	}
	for assignment := 0; assignment < 4; assignment++ {
		full := []bool{assignment&1 != 0, assignment&2 != 0}
		projected := make([]bool, 0)
		for _, id := range result.Inputs {
			if id == pi.ID() {
				projected = append(projected, full[0])
			} else if id == ro.ID() {
				projected = append(projected, full[1])
			} else {
				t.Fatal("unexpected leaf", id)
			}
		}
		values, err := result.Graph.Evaluate(projected)
		if err != nil || values[0] != (full[0] && full[1]) || values[1] {
			t.Fatal(values, err)
		}
	}
	constant, err := source.ExtractDepthLimitedCones([]GiaID{0}, 0, GiaCofactorOptions{})
	if err != nil || len(constant.Inputs) != 0 || constant.Graph.OutputCount() != 1 {
		t.Fatal("constant boundary", err)
	}
}
