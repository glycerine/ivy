package dau

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"math/rand"
	"testing"
)

func dauDecBit(t []uint64, m int) uint64 { return t[m>>6] >> uint(m&63) & 1 }
func dauDecInputs(g, mask, bound int32) (int, int) {
	u, s, ui, si := 0, 0, 0, 0
	for v := int32(0); v < bound; v++ {
		if mask&(1<<uint(v)) != 0 {
			s |= int(g>>uint(v)&1) << uint(si)
			si++
		} else {
			u |= int(g>>uint(v)&1) << uint(ui)
			ui++
		}
	}
	return u, s
}
func dauDecCompose(comp, dec []uint64, n, free, shared, mask int32) []uint64 {
	p := make([]uint64, dauCountWords(n))
	bound := n - free
	for m := 0; m < 64*len(p); m++ {
		assignment := m & ((1 << uint(n)) - 1)
		u, s := dauDecInputs(int32(assignment)>>uint(free), mask, bound)
		d := dauDecBit(dec, u|s<<uint(bound-shared))
		c := (assignment & ((1 << uint(free)) - 1)) | s<<uint(free) | int(d)<<uint(free+shared)
		if dauDecBit(comp, c) != 0 {
			p[m>>6] |= 1 << uint(m&63)
		}
	}
	return p
}
func TestBoundSetThreeVariableExhaustive(t *testing.T) {
	for f := uint64(0); f < 256; f++ {
		p := []uint64{f * 0x0101010101010101}
		comp, dec := make([]uint64, 1), make([]uint64, 1)
		status := Dau_DecCheckSetTop(p, 3, 1, 2, 0, 0, dauDecSchedule(2), dec, comp)
		cofs := map[uint64]bool{}
		for g := int32(0); g < 4; g++ {
			cofs[dauDecCofactor(p, 1, g)[0]] = true
		}
		want := int32(0)
		if len(cofs) <= 2 {
			want = 1
		}
		if status != want {
			t.Fatal(f, status, want)
		}
		if status != 0 {
			got := dauDecCompose(comp, dec, 3, 1, 0, 0)
			if got[0] != p[0] {
				t.Fatal("composition mismatch", f)
			}
		}
	}
}
func TestBoundSetSharedMultiwordComposition(t *testing.T) {
	rng := rand.New(rand.NewSource(811))
	for n := int32(3); n <= 16; n++ {
		for trial := 0; trial < 12; trial++ {
			free := int32(1 + rng.Intn(int(n-2)))
			bound := n - free
			shared := int32(rng.Intn(int(min(3, bound-2)) + 1))
			mask := int32(0)
			for bits.OnesCount32(uint32(mask)) < int(shared) {
				mask |= 1 << uint(rng.Intn(int(bound)))
			}
			c, d := make([]uint64, dauCountWords(free+shared+1)), make([]uint64, dauCountWords(bound))
			for i := range c {
				c[i] = rng.Uint64()
			}
			for i := range d {
				d[i] = rng.Uint64()
			}
			Abc_TtNormalizeSmallTruth(c, free+shared+1)
			Abc_TtNormalizeSmallTruth(d, bound)
			input := dauDecCompose(c, d, n, free, shared, mask)
			comp, dec := make([]uint64, len(c)), make([]uint64, len(d))
			if Dau_DecCheckSetTop(input, n, free, bound, shared, mask, dauDecSchedule(bound), dec, comp) == 0 {
				t.Fatalf("valid decomposition rejected n%d f%d s%d", n, free, shared)
			}
			got := dauDecCompose(comp, dec, n, free, shared, mask)
			for i, v := range input {
				if got[i] != v {
					t.Fatal("multiword recomposition", n, free, shared, i)
				}
			}
		}
	}
}
func TestScalarDecompositionAndSetEncoding(t *testing.T) {
	rng := rand.New(rand.NewSource(812))
	if Dau_DecReadSet("aBc") != 1|3<<2|1<<4 {
		t.Fatal("uppercase shared-set encoding")
	}
	for n := int32(3); n <= 6; n++ {
		for trial := 0; trial < 80; trial++ {
			p := []uint64{rng.Uint64()}
			Abc_TtNormalizeSmallTruth(p, n)
			set := uint32(0)
			for v := int32(0); v < n; v++ {
				kind := rng.Intn(3)
				if kind == 2 {
					kind = 3
				}
				set |= uint32(kind) << uint(2*v)
			}
			var u, s, f int32
			Dau_DecSortSet(set, n, &u, &s, &f)
			if u < 2 || f < 1 {
				continue
			}
			c, d := []uint64{0}, []uint64{0}
			pc, pd := make([]int32, 16), make([]int32, 16)
			var nc, nd, ns int32
			if Dau_DecDecomposeSet(p, n, set, c, d, pc, pd, &nc, &nd, &ns) == 0 {
				continue
			}
			for m := 0; m < 1<<uint(n); m++ {
				di, ci := 0, 0
				for pos, v := range pd[:nd] {
					di |= ((m >> uint(v)) & 1) << uint(pos)
				}
				dv := dauDecBit(d, di)
				for pos, v := range pc[:nc] {
					if v == n {
						ci |= int(dv) << uint(pos)
					} else {
						ci |= ((m >> uint(v)) & 1) << uint(pos)
					}
				}
				if dauDecBit(c, ci) != dauDecBit(p, m) {
					t.Fatal("scalar composition", n, set, m)
				}
			}
		}
	}
	if vec.Vec_IntSize(Dau_DecFindSets([]uint64{0}, 0)) != 0 {
		t.Fatal("constant has a proper bound set")
	}
}
