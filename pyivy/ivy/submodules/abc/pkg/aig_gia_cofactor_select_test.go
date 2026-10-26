package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaFanoutCofactorSelectionNative(t *testing.T) {
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
	dangling, err := g.AppendAnd(b, c)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_cofactor_select.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []struct {
		Signals []GiaID
		Truth   []uint32
	}
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 4 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for limit := 0; limit < 4; limit++ {
		selected, err := g.CofactorCandidates(limit, 0)
		if err != nil || !slices.Equal(selected, native[limit].Signals) {
			t.Fatal("native inclusive selection", limit, selected, native[limit], err)
		}
		if limit == 0 && !slices.Contains(selected, dangling.ID()) {
			t.Fatal("zero threshold omitted dangling gate")
		}
		graph, err := g.CofactorRebuildByFanout(limit, GiaMultiCofactorOptions{})
		if err != nil || graph == g {
			t.Fatal(graph, err)
		}
		masks := make([]uint32, 2)
		for assignment := 0; assignment < 8; assignment++ {
			bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			values, err := graph.Evaluate(bits)
			if err != nil {
				t.Fatal(err)
			}
			if values[0] != (bits[0] && bits[1]) || values[1] != (bits[0] && bits[1] && bits[2]) {
				t.Fatal("fanout reconstruction", limit, assignment)
			}
			for output, value := range values {
				if value {
					masks[output] |= 1 << assignment
				}
			}
		}
		if !slices.Equal(masks, native[limit].Truth) {
			t.Fatal(masks, native[limit])
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source scratch mutated")
	}
	if _, err := g.CofactorCandidates(-1, 0); err == nil {
		t.Fatal("negative threshold")
	}
	if _, err := g.CofactorRebuildByFanout(0, GiaMultiCofactorOptions{GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}}); err == nil {
		t.Fatal("object cap")
	}
	graph, err := g.CofactorRebuildByFanout(0, GiaMultiCofactorOptions{GiaCofactorOptions: GiaCofactorOptions{Stop: func() bool { return true }}})
	if !errors.Is(err, ErrCancelled) || graph != nil {
		t.Fatal("cancelled selection", err)
	}
}
