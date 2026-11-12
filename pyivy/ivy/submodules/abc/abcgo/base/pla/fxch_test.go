package pla

import (
	"math/rand"
	"reflect"
	"testing"
)

func evalLiteralCover(cubes [][]int32, values []bool) bool {
	for _, cube := range cubes {
		if cube == nil {
			continue
		}
		match := true
		for _, lit := range cube {
			if values[lit/2] == (lit&1 != 0) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
func evalExtracted(p *Fxch_Man_t, bits, n int) bool {
	values := make([]bool, len(p.Divs))
	for i := 0; i < n; i++ {
		values[i] = bits&(1<<i) != 0
	}
	for i := n; i < len(p.Divs); i++ {
		a, b := int32(p.Divs[i]), int32(p.Divs[i]>>32)
		values[i] = (values[a/2] != (a&1 != 0)) && (values[b/2] != (b&1 != 0))
	}
	return evalLiteralCover(p.Cubes, values)
}
func extractionCost(p *Fxch_Man_t, original int) int {
	cost := 2 * (len(p.Divs) - original)
	for _, cube := range p.Cubes {
		cost += len(cube)
	}
	return cost
}
func TestFastExtractionRandomEquivalenceAndCost(t *testing.T) {
	rng := rand.New(rand.NewSource(917))
	for trial := 0; trial < 300; trial++ {
		const n = 5
		cubes := make([][]int32, 10)
		for i := range cubes {
			cubes[i] = make([]int32, 0)
			for v := 0; v < n; v++ {
				choice := rng.Intn(3)
				if choice != 0 {
					cubes[i] = append(cubes[i], int32(v*2+choice-1))
				}
			}
		}
		p := Fxch_ManStart(cubes, make([][]int32, 2*n))
		before := extractionCost(p, n)
		Fxch_ManFastExtract(p, 0, 0)
		if after := extractionCost(p, n); after > before {
			t.Fatal("extraction grew literal cost", trial, before, after)
		}
		for bits := 0; bits < 1<<n; bits++ {
			values := make([]bool, n)
			for v := range values {
				values[v] = bits&(1<<v) != 0
			}
			if got, want := evalExtracted(p, bits, n), evalLiteralCover(cubes, values); got != want {
				t.Fatalf("trial%d bits%d got%v want%v original%v extracted%v defs%v", trial, bits, got, want, cubes, p.Cubes, p.Divs)
			}
		}
	}
}
func TestFastExtractionSelectsProfitableDivisor(t *testing.T) {
	cubes := [][]int32{{0, 2, 4}, {0, 2, 6}, {0, 2, 8}}
	p := Fxch_ManStart(cubes, make([][]int32, 10))
	Fxch_ManFastExtract(p, 0, 0)
	if len(p.Divs) <= 5 || extractionCost(p, 5) >= 9 {
		t.Fatal("common two-literal divisor was not extracted", p.Cubes, p.Divs)
	}
}
func TestConstantOneDumpAndFullLiteralPreservation(t *testing.T) {
	p := Pla_ManAlloc("one", 1, 1, 1)
	Pla_CubeSetLit(p.Cubes[0].Out, 0, PLA_LIT_ONE)
	Pla_ManConvertFromBits(p)
	name := t.TempDir() + "/one.pla"
	if err := Pla_ManDumpPla(p, name); err != nil {
		t.Fatal(err)
	}
	q, err := ReadFile(name)
	if err != nil || len(q.Cubes) != 1 || Pla_CubeGetLit(q.Cubes[0].In, 0) != PLA_LIT_DASH {
		t.Fatal("empty cube (constant one) was dropped", q, err)
	}
}

func TestFastExtractionOriginalCOracle(t *testing.T) {
	p := Fxch_ManStart([][]int32{{0, 2, 4}, {0, 2, 6}, {0, 2, 8}}, make([][]int32, 10))
	Fxch_ManFastExtract(p, 0, 0)
	if !reflect.DeepEqual(p.Divs, []uint64{0, 0, 0, 0, 0, 0x0000000200000000}) || !reflect.DeepEqual(p.Cubes, [][]int32{{4, 10}, {6, 10}, {8, 10}}) {
		t.Fatal("native FXCH mismatch", p.Divs, p.Cubes)
	}
}
