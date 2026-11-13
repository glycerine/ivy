package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// ABC's Acec_ProduceBox finds a two-adder box even with one output.
func TestSourceProduceBoxSingleOutput(t *testing.T) {
	p := gia.Gia_ManStart(200)
	in := make([]int32, 7)
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(p)
	}
	gia.Gia_ManHashStart(p)
	product := gia.Gia_ManHashAnd(p, in[0], in[1])
	digit := gia.Gia_ManHashXor(p, in[2], in[3])
	pp := gia.Gia_ManHashXor(p, gia.Gia_ManHashAnd(p, digit, in[4]), in[5])
	gia.Gia_ManHashStop(p)
	var lo, hi [2]int32
	Acec_InsertFadd(p, []int32{product, pp, in[6]}, lo[:])
	Acec_InsertFadd(p, []int32{lo[1], in[0], in[5]}, hi[:])
	gia.Gia_ManAppendCo(p, hi[0])
	if Acec_ProduceBox(p, 0) == nil {
		t.Fatal("ABC produced an arithmetic box for this graph")
	}
}
