package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestFullAdderCutTruthClassification(t *testing.T) {
	for signs := 0; signs < 8; signs++ {
		p := gia.Gia_ManStart(64)
		var in [3]int32
		cut := []int32{3, 0, 0, 0}
		for i := range in {
			in[i] = gia.Gia_ManAppendCi(p) ^ int32(signs>>i&1)
			cut[i+1] = in[i] >> 1
		}
		var out [2]int32
		Acec_InsertFadd(p, in[:], out[:])
		var sum, carry int32
		if got := Dtc_ObjComputeTruth(p, out[0]>>1, cut, &sum); got != 1 {
			t.Fatalf("signs %d sum classification %d truth %x", signs, got, sum)
		}
		if got := Dtc_ObjComputeTruth(p, out[1]>>1, cut, &carry); got != 2 {
			t.Fatalf("signs %d carry classification %d truth %x", signs, got, carry)
		}
		for i := int32(1); i < gia.Gia_ManObjNum(p); i++ {
			if gia.Gia_ManObj(p, i).Value != 0 {
				t.Fatalf("node %d retains scratch value", i)
			}
		}
	}
}
