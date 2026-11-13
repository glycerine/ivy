package acec

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestBoothRecognitionAgainstOriginalC(t *testing.T) {
	f, err := os.Open("test_vectors/acec_booth_results.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	in := bufio.NewReader(f)
	for mode := 0; mode < 3; mode++ {
		for mask := 0; mask < 32; mask++ {
			p := gia.Gia_ManStart(128)
			gia.Gia_ManHashStart(p)
			var x [6]int32
			for i := range x {
				x[i] = gia.Gia_ManAppendCi(p)
			}
			a, b, c, d, e := x[0]^int32(mask&1), x[1]^int32((mask>>1)&1), x[2]^int32((mask>>2)&1), x[3]^int32((mask>>3)&1), x[4]^int32((mask>>4)&1)
			inner := gia.Gia_ManHashXor(p, a, b)
			var root int32
			if mode == 2 {
				root = gia.Gia_ManHashXor(p, gia.Gia_ManHashAnd(p, inner, d), c)
			} else {
				both := gia.Gia_ManHashAnd(p, a, b)
				neither := gia.Gia_ManHashAnd(p, a^1, b^1)
				mux := gia.Gia_ManHashMux(p, c, both, neither)
				product := gia.Gia_ManHashOr(p, gia.Gia_ManHashAnd(p, mux, d), gia.Gia_ManHashAnd(p, inner, e))
				control := c
				if mode == 1 {
					control = x[5]
				}
				root = gia.Gia_ManHashXor(p, product, control)
			}
			oneIns, twoIns := [5]int32{-99, -99, -99, -99, -99}, [5]int32{-99, -99, -99, -99, -99}
			one := Acec_DetectBoothOne(p, gia.Gia_ManObj(p, root>>1), oneIns[:])
			two := Acec_DetectBoothTwo(p, gia.Gia_ManObj(p, root>>1), twoIns[:])
			var wantOne, wantTwo int32
			var w1, w2 [5]int32
			if _, err = fmt.Fscan(in, &wantOne, &wantTwo, &w1[0], &w1[1], &w1[2], &w1[3], &w1[4], &w2[0], &w2[1], &w2[2], &w2[3], &w2[4]); err != nil {
				t.Fatalf("mode=%d mask=%d source fixture: %v", mode, mask, err)
			}
			if one != wantOne || two != wantTwo || oneIns != w1 || twoIns != w2 {
				t.Fatalf("mode=%d mask=%d Go=%d,%d %v %v C=%d,%d %v %v", mode, mask, one, two, oneIns, twoIns, wantOne, wantTwo, w1, w2)
			}
		}
	}
}
