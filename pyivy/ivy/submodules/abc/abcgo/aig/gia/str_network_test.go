package gia

import "testing"

func TestStructuredNetworkToGia(t *testing.T) {
	original := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(original), Gia_ManAppendCi(original)
	x := Gia_ManAppendAnd(original, a, b^1)
	Gia_ManAppendCo(original, x)
	ntk := Str_NtkCreate(8, 8)
	defer Str_NtkDelete(ntk)
	sa := Str_ObjCreate(ntk, STR_PI, nil)
	sb := Str_ObjCreate(ntk, STR_PI, nil)
	sx := Str_ObjCreate(ntk, STR_AND, []int32{sa, sb ^ 1})
	Str_ObjCreate(ntk, STR_PO, []int32{sx})
	out := Str_NtkToGia(original, ntk)
	if Gia_ManVerifyTwoTruths(original, out) != 1 {
		t.Fatal("structured conversion changed function")
	}
	if Str_ObjFaninId(ntk, Str_NtkObj(ntk, sx>>1), 1) != sb>>1 || Str_ObjFaninC(ntk, Str_NtkObj(ntk, sx>>1), 1) != 1 {
		t.Fatal("structured fanin lost phase")
	}
}
