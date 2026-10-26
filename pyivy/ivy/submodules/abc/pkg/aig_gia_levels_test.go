package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaLevelsNativeKindsAndOrder(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	and, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(and)
	if err != nil {
		t.Fatal(err)
	}
	xor, err := g.AppendXorReal(a, b)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := g.AppendMuxReal(a, and, xor)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(buffer); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	levels, maximum, err := g.Levels(0)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_levels.json")
	if err != nil {
		t.Fatal(err)
	}
	var native struct {
		Max    uint32
		Levels []uint32
	}
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	if maximum != native.Max || !slices.Equal(levels, native.Levels) {
		t.Fatal("native levels", levels, maximum, native)
	}
	if !slices.Equal(levels, []uint32{0, 0, 0, 1, 1, 2, 4, 1, 4}) {
		t.Fatal("independent weighted depth", levels)
	}
	groups, err := g.Levelize(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 5 || !slices.Equal(groups[1], []GiaID{and.ID(), buffer.ID()}) || !slices.Equal(groups[2], []GiaID{xor.ID()}) || !slices.Equal(groups[4], []GiaID{mux.ID()}) {
		t.Fatal(groups)
	}
	order, err := g.BreadthFirstCrossCutOrder(0)
	if err != nil || !slices.Equal(order, []GiaID{and.ID(), buffer.ID(), xor.ID(), mux.ID()}) {
		t.Fatal(order, err)
	}
	levels[0] = 99
	groups[1][0] = 99
	order[0] = 99
	again, _, err := g.Levels(g.ObjectCount())
	if err != nil || again[0] != 0 {
		t.Fatal("levels alias manager")
	}
	reverse, reverseMax, err := g.ReverseLevels(0)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile("test_vectors/oracle/gia_reverse_levels.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	if reverseMax != native.Max || !slices.Equal(reverse, native.Levels) {
		t.Fatal("native reverse levels", reverse, reverseMax, native)
	}
	if !slices.Equal(reverse, []uint32{0, 3, 3, 2, 1, 2, 1, 0, 0}) {
		t.Fatal("independent reverse metric", reverse)
	}
	reverseGroups, err := g.ReverseLevelize(0)
	if err != nil {
		t.Fatal(err)
	}
	wantGroups := [][]GiaID{{7, 8}, {4, 6}, {3, 5}, {1, 2}}
	if len(reverseGroups) != len(wantGroups) {
		t.Fatal(reverseGroups)
	}
	for i, group := range wantGroups {
		if !slices.Equal(reverseGroups[i], group) {
			t.Fatal(reverseGroups)
		}
	}
	reverseOrder, err := g.ReverseLevelOrder(0)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile("test_vectors/oracle/gia_order.json")
	if err != nil {
		t.Fatal(err)
	}
	var nativeOrder []GiaID
	if err = json.Unmarshal(raw, &nativeOrder); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(reverseOrder, nativeOrder) || !slices.Equal(reverseOrder, []GiaID{8, 6, 5, 7, 3, 4, 0, 1, 2}) {
		t.Fatal("reverse level order", reverseOrder, nativeOrder)
	}
	seeded, seededMax, err := g.LevelsWithInputs([]uint32{3, 1}, 0)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile("test_vectors/oracle/gia_ci_levels.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	if seededMax != native.Max || !slices.Equal(seeded, native.Levels) || !slices.Equal(seeded, []uint32{0, 3, 1, 4, 4, 5, 7, 4, 7}) {
		t.Fatal("native CI levels", seeded, seededMax, native)
	}
	average, err := g.AverageOutputLevel([]uint32{3, 1}, 0)
	if err != nil || average != 5.5 {
		t.Fatal(average, err)
	}
	weighted, err := g.WeightedReverseLevels(0)
	if err != nil {
		t.Fatal(err)
	}
	required, err := g.RequiredLevels(0)
	if err != nil {
		t.Fatal(err)
	}
	slacks, err := g.LevelSlacks(0)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile("test_vectors/oracle/gia_required_levels.json")
	if err != nil {
		t.Fatal(err)
	}
	var nativeRequired struct {
		Reverse          []uint32
		Required, Slacks []int64
	}
	if err = json.Unmarshal(raw, &nativeRequired); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(weighted, nativeRequired.Reverse) || !slices.Equal(required, nativeRequired.Required) || !slices.Equal(slacks, nativeRequired.Slacks) {
		t.Fatal("native weighted metrics", weighted, required, slacks, nativeRequired)
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("levels mutated scratch")
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, _, err := g.Levels(cap); err == nil {
			t.Fatal("invalid cap", cap)
		}
	}
}
func TestGiaLevelsEmptyAndDangling(t *testing.T) {
	g := NewGiaManager()
	groups, err := g.Levelize(0)
	if err != nil || len(groups) != 1 || len(groups[0]) != 0 {
		t.Fatal(groups, err)
	}
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	chain := a
	for i := 0; i < 100; i++ {
		var err error
		chain, err = g.AppendAnd(chain, b)
		if err != nil {
			t.Fatal(err)
		}
	}
	// No CO drives this cone. It still contributes to ABC's maximum level.
	levels, maximum, err := g.Levels(0)
	if err != nil || maximum != 100 || levels[chain.ID()] != 100 {
		t.Fatal(maximum, err)
	}
}
