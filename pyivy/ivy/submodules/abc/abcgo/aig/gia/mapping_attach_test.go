package gia

import "testing"

func TestSetMappedDataValidatesBeforeAttaching(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, root)
	mapping := Vec_IntStart(int(p.nObjs))
	Vec_IntWriteEntry(mapping, int(root>>1), int32(Vec_IntSize(mapping)))
	for _, v := range []int32{2, a >> 1, b >> 1, -(root >> 1)} {
		Vec_IntPush(mapping, v)
	}
	packing := Vec_IntAlloc(3)
	for _, v := range []int32{1, 1, root >> 1} {
		Vec_IntPush(packing, v)
	}
	configs := Vec_IntAlloc(3)
	for _, v := range []int32{1, 1, 9} {
		Vec_IntPush(configs, v)
	}
	if err := p.SetMappedData(mapping, packing, configs, ""); err == nil {
		t.Fatal("accepted missing cell description")
	}
	if p.VMapping != nil || p.VPacking != nil || p.VConfigs != nil {
		t.Fatal("failed call changed manager")
	}
	if err := p.SetMappedData(mapping, packing, configs, "{ab}"); err != nil {
		t.Fatal(err)
	}
	if Gia_ManHasMapping(p) == 0 || Gia_ObjLutIsMux(p, root>>1) == 0 {
		t.Fatal("mapping or mux marker lost")
	}
	if cell, got := p.CellConfiguration(); cell != "{ab}" || got != configs {
		t.Fatal("cell configuration lost")
	}
	if err := p.SetMappedData(mapping, nil, nil, ""); err == nil {
		t.Fatal("overwrote existing metadata")
	}
}
