package util

import (
	"math/bits"
	"math/rand"
	"reflect"
	"testing"
)

func isopTestStretch(t word, n int32) word {
	if n >= 6 {
		return t
	}
	width := uint(1 << uint(n))
	mask := word(1)<<width - 1
	t &= mask
	for shift := width; shift < 64; shift *= 2 {
		t |= t << shift
	}
	return t
}
func isopTestCoverage(t *testing.T, on, dc, result []word, n int32, cost word, cover []int32, xor bool) {
	t.Helper()
	count := Abc_CostCubes(cost)
	if count < 0 || int(count) > len(cover) {
		t.Fatal("bad cube count", cost)
	}
	cover = cover[:count]
	rebuilt := make([]word, isopWords(n))
	mode := int32(0)
	if xor {
		mode = 1
	}
	Abc_IsopBuildTruth(cover, n, rebuilt, mode, 0)
	for i := range rebuilt {
		if rebuilt[i]&on[i] != on[i] || rebuilt[i]&^dc[i] != 0 {
			t.Fatalf("coverage mismatch n=%d word=%d got=%x on=%x dc=%x cubes=%v", n, i, rebuilt[i], on[i], dc[i], cover)
		}
		if result != nil && rebuilt[i] != result[i] {
			t.Fatal("returned truth differs from cube truth")
		}
	}
	if Abc_IsopCountLits(cover, n) != Abc_CostLits(cost) {
		t.Fatal("cost literal count", cost, cover)
	}
}
func TestIsopExhaustiveThreeVariables(t *testing.T) {
	limit := Abc_Cube2Cost(1000)
	for table := word(0); table < 256; table++ {
		on := []word{isopTestStretch(table, 3)}
		result := make([]word, 1)
		cover := make([]int32, 100)
		cost := Abc_IsopCheck(on, on, result, 3, limit, cover)
		isopTestCoverage(t, on, on, result, 3, cost, cover, false)
		count := Abc_CostCubes(cost)
		for i := int32(0); i < count; i++ {
			without := append([]int32(nil), cover[:i]...)
			without = append(without, cover[i+1:count]...)
			truth := make([]word, 1)
			Abc_IsopBuildTruth(without, 3, truth, 0, 0)
			if truth[0] == on[0] {
				t.Fatalf("redundant cube for truth %x", on[0])
			}
		}
		esop := make([]int32, 100)
		cost = Abc_EsopCheck(on, 3, limit, esop)
		isopTestCoverage(t, on, on, nil, 3, cost, esop, true)
	}
}
func TestIsopDontCareCoverage(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for n := int32(1); n <= 6; n++ {
		for k := 0; k < 150; k++ {
			dc := isopTestStretch(r.Uint64(), n)
			on := isopTestStretch(r.Uint64(), n) & dc
			cover := make([]int32, 500)
			result := make([]word, 1)
			cost := Abc_Isop6Cover(on, dc, result, n, Abc_Cube2Cost(1000), cover)
			isopTestCoverage(t, []word{on}, []word{dc}, result, n, cost, cover, false)
			greedy := make([]int32, 500)
			cost = Abc_IsopNew([]word{on}, []word{dc}, result, n, Abc_Cube2Cost(1000), greedy)
			isopTestCoverage(t, []word{on}, []word{dc}, result, n, cost, greedy, false)
		}
	}
}
func TestIsopLargeSupport(t *testing.T) {
	wrappers := []FUNC_ISOP{Abc_Isop7Cover, Abc_Isop8Cover, Abc_Isop9Cover, Abc_Isop10Cover, Abc_Isop11Cover, Abc_Isop12Cover, Abc_Isop13Cover, Abc_Isop14Cover, Abc_Isop15Cover, Abc_Isop16Cover}
	for n := int32(7); n <= 16; n++ {
		on := make([]word, isopWords(n))
		for i := range on {
			on[i] = isopMask(0) & isopMask(1)
			if i&(1<<uint(n-7)) != 0 {
				on[i] = ^on[i]
			}
		}
		copyOn := append([]word(nil), on...)
		result := make([]word, len(on))
		cover := make([]int32, 100)
		cost := wrappers[n-7](on, on, result, Abc_Cube2Cost(1000), cover)
		isopTestCoverage(t, on, on, result, n, cost, cover, false)
		cost = Abc_EsopCheck(on, n, Abc_Cube2Cost(1000), cover)
		isopTestCoverage(t, on, on, nil, n, cost, cover, true)
		cost = Abc_IsopNew(on, on, result, n, Abc_Cube2Cost(1000), cover)
		isopTestCoverage(t, on, on, result, n, cost, cover, false)
		if !reflect.DeepEqual(on, copyOn) {
			t.Fatal("input mutated")
		}
	}
}
func TestIsopLimitsPolarityAndConstants(t *testing.T) {
	for _, constant := range []word{0, ^word(0)} {
		var cover []int32
		flag := Abc_Isop([]word{constant}, 0, 10, &cover, 1)
		result := make([]word, 1)
		Abc_IsopBuildTruth(cover, 0, result, 0, flag)
		if result[0] != constant {
			t.Fatal("constant result", flag, cover)
		}
	}
	on := []word{isopMask(0) | isopMask(1) | isopMask(2)}
	before := append([]word(nil), on...)
	var cover []int32
	flag := Abc_Isop(on, 3, 100, &cover, 1)
	if flag != 1 || len(cover) != 1 {
		t.Fatal("wrong preferred polarity", flag, cover)
	}
	result := make([]word, 1)
	Abc_IsopBuildTruth(cover, 3, result, 0, flag)
	if result[0] != on[0] || !reflect.DeepEqual(on, before) {
		t.Fatal("polarity/input mismatch")
	}
	if Abc_Isop(on, 3, 1, &cover, 0) != -1 {
		t.Fatal("cube limit ignored")
	}
	limit := Abc_Cube2Cost(1)
	if Abc_IsopCheck(on, on, result, 3, limit, nil) != limit || Abc_EsopCheck(on, 3, limit, nil) != limit {
		t.Fatal("packed cutoff not saturated")
	}
}
func TestIsopCnfAndCubeCost(t *testing.T) {
	on := []word{isopMask(0) & isopMask(1)}
	cover := make([]int32, 100)
	count := Abc_IsopCnf(on, 2, 100, cover)
	if count <= 0 {
		t.Fatal(count)
	}
	relation := make([]word, 1)
	Abc_IsopBuildTruth(cover[:count], 3, relation, 0, 0)
	expected := (on[0] &^ isopMask(2)) | (^on[0] & isopMask(2))
	if relation[0] != expected {
		t.Fatal("CNF polarity encoding")
	}
	for _, cube := range cover[:count] {
		if bits.OnesCount32(uint32(cube)) == 0 {
			t.Fatal("missing output literal")
		}
	}
}
