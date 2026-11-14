package wlc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestBlastFastAdderExhaustive(t *testing.T) {
	for n := int32(1); n <= 4; n++ {
		for sign := int32(0); sign < 2; sign++ {
			for carry := int32(0); carry < 2; carry++ {
				g := gia.Gia_ManStart(1000)
				gia.Gia_ManHashAlloc(g)
				a, b := make([]int32, n+1), make([]int32, n)
				for i := int32(0); i < n; i++ {
					a[i] = gia.Gia_ManAppendCi(g)
				}
				for i := range b {
					b[i] = gia.Gia_ManAppendCi(g)
				}
				Wlc_BlastAdderFast(g, a, b, n, sign, carry)
				for _, lit := range a[:n] {
					gia.Gia_ManAppendCo(g, lit)
				}
				for x := int32(0); x < 1<<n; x++ {
					for y := int32(0); y < 1<<n; y++ {
						inputs := make([]bool, 2*n)
						for i := int32(0); i < n; i++ {
							inputs[i] = x&(1<<i) != 0
							inputs[n+i] = y&(1<<i) != 0
						}
						out, err := g.Evaluate(inputs)
						if err != nil {
							t.Fatal(err)
						}
						for i := int32(0); i < n; i++ {
							if out[i] != ((x+y+carry)&(1<<i) != 0) {
								t.Fatalf("width=%d sign=%d carry=%d x=%d y=%d bit=%d", n, sign, carry, x, y, i)
							}
						}
					}
				}
			}
		}
	}
}
