package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestStrashOutputClasses(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c, d := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, c, d))
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b^1))
	Gia_ManAppendCo(p, a)
	supp := Gia_ManCreateCoSupps(p, 0)
	var mapping *Vec_Int_t
	out := Gia_ManDupStrashReduce(p, supp, &mapping)
	if Gia_ManPoNum(out) != 3 || !reflect.DeepEqual(vec.Vec_IntArray(mapping), []int32{0, 0, 1, 2}) {
		t.Fatalf("mapping %v", vec.Vec_IntArray(mapping))
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		got, e := out.Evaluate(in)
		if e != nil || got[0] != (in[0] && in[1]) || got[1] != (in[0] && !in[1]) || got[2] != in[0] {
			t.Fatalf("reduced %v %v", got, e)
		}
	}
	var groups *Vec_Ptr_t
	kept := Gia_ManIsoStrashReduce(p, &groups, 0)
	if Gia_ManPoNum(kept) != 3 || vec.Vec_PtrSize(groups) != 3 || !reflect.DeepEqual(vec.Vec_IntArray(vec.Vec_PtrEntry(groups, 0).(*Vec_Int_t)), []int32{0, 1}) {
		t.Fatal("incorrect output classes")
	}
}
