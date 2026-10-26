package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaMuxCofactorsNativeAndReconstruction(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	if err := g.HashStart(); err != nil {
		t.Fatal(err)
	}
	root, err := g.HashOr(n, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(root); err != nil {
		t.Fatal(err)
	}
	g.Name = "mux-cofactors"
	g.Spec = "source.aig"
	raw, err := os.ReadFile("test_vectors/oracle/gia_mux_cofactors.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []uint32
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 3 {
		t.Fatal(err)
	}
	for pivot := 0; pivot < 3; pivot++ {
		negative, err := g.CofactorPrimaryInput(pivot, false, GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		positive, err := g.CofactorPrimaryInput(pivot, true, GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		beforePositive, beforeNegative := slices.Clone(positive.objects), slices.Clone(negative.objects)
		combined, err := GiaMuxCofactors(pivot, positive, negative, GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if combined.Name != positive.Name || combined.Spec != positive.Spec || combined.InputCount() != 3 || combined.OutputCount() != 1 {
			t.Fatal("composition metadata")
		}
		mask := uint32(0)
		for assignment := 0; assignment < 8; assignment++ {
			bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			values, err := combined.Evaluate(bits)
			if err != nil {
				t.Fatal(err)
			}
			if values[0] != ((bits[0] && bits[1]) || bits[2]) {
				t.Fatal("cofactor mux truth", pivot, assignment)
			}
			if values[0] {
				mask |= 1 << assignment
			}
		}
		if mask != native[pivot] {
			t.Fatal(mask, native[pivot])
		}
		if !slices.Equal(beforePositive, positive.objects) || !slices.Equal(beforeNegative, negative.objects) {
			t.Fatal("source cofactors mutated")
		}
		for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}} {
			if _, err := GiaMuxCofactors(pivot, positive, negative, options); err == nil {
				t.Fatal("invalid cap")
			}
		}
		calls := 0
		result, err := GiaMuxCofactors(pivot, positive, negative, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= 5 }})
		if !errors.Is(err, ErrCancelled) || result != nil {
			t.Fatal("cancelled composition", err)
		}
	}
	if _, err := GiaMuxCofactors(-1, g, g, GiaCofactorOptions{}); err == nil {
		t.Fatal("invalid control PI")
	}
	if _, err := GiaMuxCofactors(3, g, g, GiaCofactorOptions{}); err == nil {
		t.Fatal("out-of-range control")
	}
	if err := g.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	combined, err := GiaMuxCofactors(0, g, g, GiaCofactorOptions{})
	if err != nil || combined.ConstraintCount() != 1 {
		t.Fatal("constraint preservation", err)
	}
	empty := NewGiaManager()
	if _, err := GiaMuxCofactors(0, g, empty, GiaCofactorOptions{}); err == nil {
		t.Fatal("incompatible cofactors")
	}
}
