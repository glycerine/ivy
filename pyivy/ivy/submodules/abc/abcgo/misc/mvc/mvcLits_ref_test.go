package mvc

import (
	"math/rand"
	"testing"
)

// The literal selectors must answer what mvcLits.c computes: counts of each
// literal over the cubes, restricted to the mask; Any = last literal seen
// twice, Best = first most frequent (if > 1), Worst = first least frequent
// among those seen at least twice, CountLiterals = total.
func TestLiteralSelectorsMatchReference(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for trial := 0; trial < 20000; trial++ {
		nBits := 1 + r.Intn(70)
		c := Mvc_CoverAlloc(Mvc_ManagerStart(), nBits)
		var cubes [][]bool
		for k := r.Intn(6); k > 0; k-- {
			x := Mvc_CubeAlloc(c)
			Mvc_CubeBitClean(x)
			row := make([]bool, nBits)
			for i := range row {
				if r.Intn(3) == 0 {
					row[i] = true
					Mvc_CubeBitInsert(x, i)
				}
			}
			cubes = append(cubes, row)
			Mvc_CoverAddCubeTail(c, x)
		}
		var mask *Mvc_Cube_t
		inMask := func(int) bool { return true }
		if r.Intn(2) == 0 {
			mask = Mvc_CubeAlloc(c)
			Mvc_CubeBitClean(mask)
			sel := make([]bool, nBits)
			for i := range sel {
				if r.Intn(2) == 0 {
					sel[i] = true
					Mvc_CubeBitInsert(mask, i)
				}
			}
			inMask = func(i int) bool { return sel[i] }
		}
		count := func(i int) int {
			n := 0
			for _, row := range cubes {
				if row[i] {
					n++
				}
			}
			return n
		}
		any, best, worst, total := -1, -1, -1, 0
		most, least := -1, 1000000
		for i := 0; i < nBits; i++ {
			total += count(i)
			if !inMask(i) {
				continue
			}
			k := count(i)
			if k > 1 {
				any = i
			}
			if most < k {
				best, most = i, k
			}
			if k >= 2 && least > k {
				worst, least = i, k
			}
		}
		if most <= 1 {
			best = -1
		}
		if got := Mvc_CoverAnyLiteral(c, mask); got != any {
			t.Fatalf("trial %d: Any = %d, want %d", trial, got, any)
		}
		if got := Mvc_CoverBestLiteral(c, mask); got != best {
			t.Fatalf("trial %d: Best = %d, want %d", trial, got, best)
		}
		if got := Mvc_CoverWorstLiteral(c, mask); got != worst {
			t.Fatalf("trial %d: Worst = %d, want %d", trial, got, worst)
		}
		if got := Mvc_CoverCountLiterals(c); got != total {
			t.Fatalf("trial %d: CountLiterals = %d, want %d", trial, got, total)
		}
	}
}
