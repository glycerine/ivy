package gia

import "testing"

func TestManMappingReturnsBorrowedVector(t *testing.T) {
	p := Gia_ManStart(10)
	if Gia_ManMapping(p) != nil {
		t.Fatal("new manager has a mapping")
	}
	p.vMapping = Vec_IntStart(2)
	Vec_IntWriteEntry(Gia_ManMapping(p), 1, 7)
	if Vec_IntEntry(p.vMapping, 1) != 7 {
		t.Fatal("accessor did not return manager's mapping")
	}
}
