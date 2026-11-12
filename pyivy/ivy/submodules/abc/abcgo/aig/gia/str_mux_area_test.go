package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMuxAreaRestructurePreservesTruth(t *testing.T) {
	ntk := Str_NtkCreate(16, 24)
	defer Str_NtkDelete(ntk)
	var in [7]int32
	for i := range in {
		in[i] = Str_ObjCreate(ntk, STR_PI, nil)
	}
	left := Str_ObjCreate(ntk, STR_MUX, []int32{in[0], in[1], in[4]})
	right := Str_ObjCreate(ntk, STR_MUX, []int32{in[2], in[3], in[5]})
	Str_ObjCreate(ntk, STR_MUX, []int32{left, right, in[6]})
	dst := Gia_ManStart(32)
	dst.EnableMuxes()
	Gia_ManHashStart(dst)
	delays := vec.Vec_IntStart(512)
	vec.Vec_IntWriteEntry(delays, 0, 17)
	for _, lit := range in {
		copy := Gia_ManAppendCi(dst)
		Str_NtkObj(ntk, lit>>1).iCopy = copy
		vec.Vec_IntWriteEntry(delays, int(copy>>1), 17)
	}
	result := Str_MuxRestructureArea(dst, ntk, left>>1, 3, delays, 4, 0)
	Gia_ManAppendCo(dst, result)
	Gia_ManHashStop(dst)
	for mask := 0; mask < 128; mask++ {
		v := make([]bool, 7)
		for i := range v {
			v[i] = mask&(1<<i) != 0
		}
		a, b := v[0], v[2]
		if v[4] {
			a = v[1]
		}
		if v[5] {
			b = v[3]
		}
		want := a
		if v[6] {
			want = b
		}
		got, err := dst.Evaluate(v)
		if err != nil || got[0] != want {
			t.Fatalf("mask %d got %v want %v: %v", mask, got, want, err)
		}
	}
}
