package wlc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestBlastArithmeticExhaustive(t *testing.T) {
	for n := int32(1); n <= 4; n++ {
		g := gia.Gia_ManStart(2000)
		gia.Gia_ManHashAlloc(g)
		a, b := make([]int32, n), make([]int32, n)
		for i := int32(0); i < n; i++ {
			a[i] = gia.Gia_ManAppendCi(g)
		}
		for i := int32(0); i < n; i++ {
			b[i] = gia.Gia_ManAppendCi(g)
		}
		less, less2, signed := Wlc_BlastLess(g, a, b, n), Wlc_BlastLess2(g, a, b, n), Wlc_BlastLessSigned(g, a, b, n)
		sum, diff := append([]int32(nil), a...), append([]int32(nil), a...)
		Wlc_BlastAdder(g, sum, b, n, 0)
		Wlc_BlastSubtract(g, diff, b, n, 1)
		for _, out := range append([]int32{less, less2, signed}, append(sum, diff...)...) {
			gia.Gia_ManAppendCo(g, out)
		}
		limit := int32(1) << n
		for x := int32(0); x < limit; x++ {
			for y := int32(0); y < limit; y++ {
				inputs := make([]bool, 2*n)
				for i := int32(0); i < n; i++ {
					inputs[i] = x&(1<<i) != 0
					inputs[n+i] = y&(1<<i) != 0
				}
				out, err := g.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				sx, sy := x, y
				if x&(limit>>1) != 0 {
					sx -= limit
				}
				if y&(limit>>1) != 0 {
					sy -= limit
				}
				if out[0] != (x < y) || out[1] != (x < y) || out[2] != (sx < sy) {
					t.Fatalf("width=%d x=%d y=%d compare=%v", n, x, y, out[:3])
				}
				for i := int32(0); i < n; i++ {
					if out[3+i] != ((x+y)&(1<<i) != 0) || out[3+n+i] != ((x-y)&(1<<i) != 0) {
						t.Fatalf("width=%d x=%d y=%d bit=%d arithmetic=%v", n, x, y, i, out)
					}
				}
			}
		}
	}
}
