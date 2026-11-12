package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestStructuredMultiBalance(t *testing.T) {
	original := Gia_ManStart(16)
	ntk := Str_NtkCreate(8, 8)
	defer Str_NtkDelete(ntk)
	leaves := make([]int32, 4)
	for i := range leaves {
		leaves[i] = Str_ObjCreate(ntk, STR_PI, nil)
	}
	gate := Str_ObjCreate(ntk, STR_AND, leaves)
	obj := Str_NtkObj(ntk, gate>>1)
	dst := Gia_ManStart(16)
	Gia_ManHashStart(dst)
	for _, lit := range leaves {
		id := Str_ObjId(ntk, Str_NtkObj(ntk, lit>>1))
		Str_NtkObj(ntk, id).iCopy = Gia_ManAppendCi(dst)
	}
	vec.Vec_IntFill(&dst.vCopies, 256, -1)
	delays := vec.Vec_IntStart(256)
	for id := int32(0); id < dst.nObjs; id++ {
		vec.Vec_IntWriteEntry(delays, int(id), 17)
	}
	Str_NtkBalanceMulti(dst, ntk, obj, delays, 4)
	Gia_ManAppendCo(dst, obj.iCopy)
	Gia_ManHashStop(dst)
	for mask := 0; mask < 16; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		out, err := dst.Evaluate(in)
		if err != nil || out[0] != (mask == 15) {
			t.Fatalf("mask %d gives %v: %v", mask, out, err)
		}
	}
	_ = original
}
