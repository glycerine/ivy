package kit

import (
	"math/rand"
	"testing"
)

func kitCoverValue(cover []uint32, x int) bool {
	for _, c := range cover {
		yes := true
		for v := 0; v < 16; v++ {
			lit := c >> uint(2*v) & 3
			if lit == 1 && x&(1<<uint(v)) != 0 || lit == 2 && x&(1<<uint(v)) == 0 || lit == 3 {
				yes = false
				break
			}
		}
		if yes {
			return true
		}
	}
	return false
}
func TestIsopFactoringAllThreeVariableFunctions(t *testing.T) {
	for f := uint32(0); f < 256; f++ {
		truth := f * 0x01010101
		for phase := int32(0); phase < 2; phase++ {
			var cover, memory []uint32
			compl := Kit_TruthIsop([]uint32{truth}, 3, &cover, phase)
			if compl < 0 {
				t.Fatal("ISOP failed", f)
			}
			g := Kit_SopFactor(cover, compl, 3, &memory)
			got := Kit_GraphToTruth(g)
			if got != truth {
				t.Fatalf("f=%x phase=%d cover=%x graph=%x", truth, compl, cover, got)
			}
		}
	}
}
func TestIsopMultiwordAndDontCare(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	for nVars := int32(5); nVars <= 9; nVars++ {
		for trial := 0; trial < 10; trial++ {
			truth := make([]uint32, Kit_TruthWordNum(nVars))
			for i := range truth {
				truth[i] = rng.Uint32()
			}
			var cover []uint32
			phase := Kit_TruthIsop(truth, nVars, &cover, 1)
			if phase < 0 {
				t.Fatal("ISOP failed")
			}
			for x := 0; x < 1<<uint(nVars); x++ {
				want := truth[x>>5]&(1<<uint(x&31)) != 0
				got := kitCoverValue(cover, x)
				if phase != 0 {
					got = !got
				}
				if got != want {
					t.Fatal(nVars, trial, x)
				}
			}
			off, on := make([]uint32, len(truth)), make([]uint32, len(truth))
			for i := range truth {
				care := rng.Uint32()
				on[i] = truth[i] & care
				off[i] = ^truth[i] & care
			}
			phase = Kit_TruthIsop2(off, on, nVars, &cover, 1, 0)
			if phase < 0 {
				t.Fatal("ISOP2 failed")
			}
			for x := 0; x < 1<<uint(nVars); x++ {
				got := kitCoverValue(cover, x)
				if phase != 0 {
					got = !got
				}
				bit := uint32(1) << uint(x&31)
				if on[x>>5]&bit != 0 && !got || off[x>>5]&bit != 0 && got {
					t.Fatal("care violation", nVars, trial, x)
				}
			}
		}
	}
}
func TestAlgebraicDivisionPreservesCover(t *testing.T) {
	// (a+b)(c+d)+e, using one positive literal bit per variable.
	sop := Kit_Sop_t{pCubes: []uint32{1 | 4, 1 | 8, 2 | 4, 2 | 8, 16}}
	div := Kit_Sop_t{pCubes: []uint32{1, 2}}
	var q, r Kit_Sop_t
	Kit_SopDivideInternal(&sop, &div, &q, &r, nil)
	if q.cubeNum() != 2 || r.cubeNum() != 1 || r.pCubes[0] != 16 {
		t.Fatal(q, r)
	}
	if sop.pCubes[0] != 5 {
		t.Fatal("source cover mutated")
	}
}
