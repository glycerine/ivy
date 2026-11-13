package gia

import "testing"

func TestLutConfigMetadata(t *testing.T) {
	p := Gia_ManStart(4)
	v := Vec_IntStart(2)
	Gia_ManSetLutConfigs(p, v)
	if Gia_ManLutConfigs(p) != v {
		t.Fatal("LUT config metadata not retained")
	}
}
