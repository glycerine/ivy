package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestStructuredMuxTreeRoundTrip(t *testing.T) {
	ntk := Str_NtkCreate(12, 16)
	defer Str_NtkDelete(ntk)
	var inputs [5]int32
	for i := range inputs {
		inputs[i] = Str_ObjCreate(ntk, STR_PI, nil)
	}
	child := Str_ObjCreate(ntk, STR_MUX, []int32{inputs[1], inputs[2], inputs[0]})
	Str_ObjCreate(ntk, STR_MUX, []int32{child, inputs[4], inputs[3]})
	dst := Gia_ManStart(20)
	dst.EnableMuxes()
	Gia_ManHashStart(dst)
	delays := vec.Vec_IntStart(256)
	for _, lit := range inputs {
		copy := Gia_ManAppendCi(dst)
		Str_NtkObj(ntk, lit>>1).iCopy = copy
		vec.Vec_IntWriteEntry(delays, int(copy>>1), 17)
	}
	vec.Vec_IntWriteEntry(delays, 0, 17)
	tree := make([]Str_Mux_t, 3)
	Str_MuxCreate(tree, ntk, child>>1, 2, delays, 4)
	if Str_MuxDelayEdge_rec(&tree[0], 0) == 0 {
		t.Fatal("missing tree delay")
	}
	result := Str_MuxToGia_rec(dst, &tree[0], 0, delays)
	Gia_ManAppendCo(dst, result)
	Gia_ManHashStop(dst)
	for mask := 0; mask < 32; mask++ {
		in := make([]bool, 5)
		for i := range in {
			in[i] = mask&(1<<i) != 0
		}
		expected := in[1]
		if in[0] {
			expected = in[2]
		}
		if in[3] {
			expected = in[4]
		}
		out, err := dst.Evaluate(in)
		if err != nil || out[0] != expected {
			t.Fatalf("mask %d: %v %v", mask, out, err)
		}
	}
}
