package gia

import (
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
)

func TestMiniEquivalenceMiterAndReduction(t *testing.T) {
	p := miniaig.Mini_AigStart()
	a, b := miniaig.Mini_AigCreatePi(p), miniaig.Mini_AigCreatePi(p)
	x := miniaig.Mini_AigAnd(p, a, b)
	miniaig.Mini_AigCreatePo(p, x)
	equiv := make([]int32, miniaig.Mini_AigNodeNum(p))
	for i := range equiv {
		equiv[i] = -1
	}
	equiv[x>>1] = int32(a)
	miter := Gia_MiniAigMiter(p, equiv)
	reduced := Gia_MiniAigReduce(p, equiv)
	for i := 0; i < 4; i++ {
		inputs := []bool{i&1 != 0, i&2 != 0}
		out, err := miter.Evaluate(inputs)
		if err != nil || len(out) != 1 || out[0] != (i == 1) {
			t.Fatalf("miter %v: %v %v", inputs, out, err)
		}
		out, err = reduced.Evaluate(inputs)
		if err != nil || len(out) != 1 || out[0] != (i&1 != 0) {
			t.Fatalf("reduced %v: %v %v", inputs, out, err)
		}
	}
}
