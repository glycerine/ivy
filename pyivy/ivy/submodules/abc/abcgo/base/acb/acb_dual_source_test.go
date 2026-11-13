package acb

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// The source uses two literals (value, unknown) to encode 0, 1, and X.
// Exercise both possible value bits for X, because the second bit alone
// determines whether the value is unknown.
func TestSourceDualRailOperations(t *testing.T) {
	pairs := [][2]int32{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
	val := func(v [2]int32) int {
		if v[1] != 0 {
			return 2
		}
		return int(v[0])
	}
	for _, force := range []int32{0, 1} {
		t.Run(fmt.Sprintf("force_zero_%d", force), func(t *testing.T) {
			old := fForceZero
			fForceZero = force
			defer func() { fForceZero = old }()
			p := gia.Gia_ManStart(100)
			gia.Gia_ManHashAlloc(p)
			for _, a := range pairs {
				out := [2]int32{}
				Gia_ManDualNot(p, a[:], out[:])
				expected := val(a)
				if expected != 2 {
					expected ^= 1
				}
				if val(out) != expected {
					t.Fatalf("not %v = %v", a, out)
				}
				for _, b := range pairs {
					va, vb := val(a), val(b)
					Gia_ManDualAnd2(p, a[:], b[:], out[:])
					want := 2
					if va == 0 || vb == 0 {
						want = 0
					} else if va == 1 && vb == 1 {
						want = 1
					}
					if val(out) != want {
						t.Fatalf("and %v %v = %v want %d", a, b, out, want)
					}
					packed := []int32{a[0], a[1], b[0], b[1]}
					Gia_ManDualAndN(p, packed, 2, out[:])
					if val(out) != want {
						t.Fatalf("andN %v %v = %v want %d", a, b, out, want)
					}
					Gia_ManDualXor2(p, a[:], b[:], out[:])
					want = 2
					if va != 2 && vb != 2 {
						want = va ^ vb
					}
					if val(out) != want {
						t.Fatalf("xor %v %v = %v want %d", a, b, out, want)
					}
					Gia_ManDualXorN(p, packed, 2, out[:])
					if val(out) != want {
						t.Fatalf("xorN %v %v = %v want %d", a, b, out, want)
					}
					Gia_ManDualDc(p, a[:], b[:], out[:])
					want = va
					if vb != 0 {
						want = 2
					}
					if val(out) != want {
						t.Fatalf("dc %v %v = %v want %d", a, b, out, want)
					}
					for _, c := range pairs {
						Gia_ManDualMux(p, a[:], b[:], c[:], out[:])
						vc := val(c)
						want = 2
						if va == 0 {
							want = vc
						} else if va == 1 {
							want = vb
						} else if vb == vc && vb != 2 {
							want = vb
						}
						if val(out) != want {
							t.Fatalf("mux %v %v %v = %v want %d", a, b, c, out, want)
						}
					}
				}
			}
		})
	}
}
