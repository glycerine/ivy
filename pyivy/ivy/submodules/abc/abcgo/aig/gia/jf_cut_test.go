package gia

import (
	"reflect"
	"testing"
)

func TestJfCutHeaderAndSignatures(t *testing.T) {
	c := []int32{Jf_CutSetAll(0xabcdef, 7, 3), 2, 65, 132}
	if Jf_CutSize(c) != 3 || Jf_CutCost(c) != 7 || Jf_CutFunc(c) != 0xabcdef {
		t.Fatalf("header: %#x", uint32(c[0]))
	}
	Jf_CutSetSize(c, 2)
	Jf_CutSetCost(c, 4)
	Jf_CutSetFunc(c, 0x812345)
	if Jf_CutSize(c) != 2 || Jf_CutCost(c) != 4 || Jf_CutFunc(c) != 0x812345 {
		t.Fatalf("updated header: %#x", uint32(c[0]))
	}
	if Jf_CutGetSign32(c) != 3 || Jf_CutGetSign(c) != (word(1)<<1|word(1)<<32) {
		t.Fatalf("signatures: %#x %#x", Jf_CutGetSign32(c), Jf_CutGetSign(c))
	}
}
func TestJfCutMerges(t *testing.T) {
	a := []int32{2, 2, 6}
	b := []int32{2, 3, 8}
	out := make([]int32, 10)
	if Jf_CutMergeOrder(a, b, out, 4) != 1 || !reflect.DeepEqual(out[:5], []int32{4, 2, 3, 6, 8}) {
		t.Fatal(out)
	}
	if Jf_CutMergeOrder(a, b, out, 3) != 0 {
		t.Fatal("merge exceeded limit")
	}
	if Jf_CutMerge0(a, b, out, 4) != 1 || !reflect.DeepEqual(out[:5], []int32{4, 2, 6, 3, 8}) {
		t.Fatal(out)
	}
	if Jf_CutMerge1(a, b, out, 3) != 1 || !reflect.DeepEqual(out[:4], []int32{3, 2, 6, 8}) {
		t.Fatal(out)
	}
	config := Jf_CutMerge2(a, b, out, 3)
	if config == 0 || config&(1<<18) == 0 || !reflect.DeepEqual(out[:4], []int32{3, 2, 6, 8}) {
		t.Fatalf("config %#x cut %v", config, out)
	}
	if Jf_CutMerge2(a, b, out, 2) != 0 {
		t.Fatal("merge exceeded limit")
	}
	if !Jf_CutIsContainedOrder([]int32{3, 2, 6, 8}, []int32{2, 2, 8}) || Jf_CutIsContainedOrder([]int32{2, 2, 6}, []int32{1, 4}) {
		t.Fatal("ordered containment")
	}
	if !Jf_CutIsContained1(a, []int32{1, 3}) || Jf_CutIsContained0(a, []int32{1, 3}) {
		t.Fatal("polarity containment")
	}
}
