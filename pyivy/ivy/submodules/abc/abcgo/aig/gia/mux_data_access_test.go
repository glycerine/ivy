package gia

import "testing"

func TestManHasMuxData(t *testing.T) {
	if Gia_ManHasMuxData(nil) {
		t.Fatal("nil manager has mux data")
	}
	p := Gia_ManStart(4)
	if Gia_ManHasMuxData(p) {
		t.Fatal("new manager has mux data")
	}
	p.PMuxes = make([]uint32, p.nObjs)
	if !Gia_ManHasMuxData(p) {
		t.Fatal("allocated mux data not detected")
	}
}
