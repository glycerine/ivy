package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMuxPathChangePreservesTruth(t *testing.T) {
	ntk := Str_NtkCreate(16, 24)
	defer Str_NtkDelete(ntk)
	var inputs [7]int32
	for i := range inputs {
		inputs[i] = Str_ObjCreate(ntk, STR_PI, nil)
	}
	m1 := Str_ObjCreate(ntk, STR_MUX, []int32{inputs[0], inputs[1], inputs[4]})
	m2 := Str_ObjCreate(ntk, STR_MUX, []int32{m1, inputs[2], inputs[5]})
	Str_ObjCreate(ntk, STR_MUX, []int32{m2, inputs[3], inputs[6]})
	dst := Gia_ManStart(30)
	dst.EnableMuxes()
	Gia_ManHashStart(dst)
	delays := vec.Vec_IntStart(512)
	vec.Vec_IntWriteEntry(delays, 0, 17)
	for _, lit := range inputs {
		copy := Gia_ManAppendCi(dst)
		Str_NtkObj(ntk, lit>>1).iCopy = copy
		vec.Vec_IntWriteEntry(delays, int(copy>>1), 17)
	}
	tree := make([]Str_Mux_t, 4)
	Str_MuxCreate(tree, ntk, m1>>1, 3, delays, 4)
	Str_MuxDelayEdge_rec(&tree[0], 0)
	var path []int32
	if Str_MuxFindPathEdge_rec(&tree[0], 0, &path) == 0 || len(path) < 3 {
		t.Fatalf("no path: %v", path)
	}
	var backup [3]Str_Mux_t
	Str_MuxChangeOnce(tree, path, 0, 2, backup[:], nil, nil)
	Str_MuxChangeUndo(tree, path, 0, 2, backup[:])
	Str_MuxChangeOnce(tree, path, 0, 2, nil, dst, delays)
	result := Str_MuxToGia_rec(dst, &tree[0], 0, delays)
	Gia_ManAppendCo(dst, result)
	Gia_ManHashStop(dst)
	for mask := 0; mask < 128; mask++ {
		v := make([]bool, 7)
		for i := range v {
			v[i] = mask&(1<<i) != 0
		}
		want := v[0]
		if v[4] {
			want = v[1]
		}
		if v[5] {
			want = v[2]
		}
		if v[6] {
			want = v[3]
		}
		got, err := dst.Evaluate(v)
		if err != nil || got[0] != want {
			t.Fatalf("mask %d got %v want %v: %v", mask, got, want, err)
		}
	}
}
