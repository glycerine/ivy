package gia

import "testing"

func TestCofNetworkCones(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	z := Gia_ManAppendAnd(p, x, b^1)
	o := Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, z)
	network := Cof_ManCreateLogicSimple(p)
	if Cof_ManNodeNum(network) != 3 || Cof_ObjFanoutNum(Cof_ManObj(network, uint32(x>>1))) != 2 {
		t.Fatal("incorrect connectivity")
	}
	output := Cof_ManObj(network, uint32(o>>1))
	if count := Cof_ManTfiSize(network, []*Cof_Obj_t{output}, 1); count != 2 {
		t.Fatalf("TFI %d", count)
	}
	if count := Cof_ManSuppSize(network, []*Cof_Obj_t{output}, 1); count != 3 {
		t.Fatalf("support %d", count)
	}
	if count := Cof_ManTfoSize(network, []*Cof_Obj_t{Cof_ManObj(network, uint32(a>>1))}, 1); count != 3 {
		t.Fatalf("TFO %d", count)
	}
	if !network.objects[z>>1].complements[0] && !network.objects[z>>1].complements[1] {
		t.Fatal("lost complemented edge")
	}
	Cof_ManResetTravId(network)
	if count := Cof_ManSuppSize(network, []*Cof_Obj_t{output}, 1); count != 3 {
		t.Fatal("traversal reset")
	}
}
