package cec

import "testing"

func TestPublicConflictLimit(t *testing.T) {
	var p Cec_ParCec_t
	p.SetConflictLimit(37)
	if p.NBTLimit != 37 {
		t.Fatalf("conflict limit = %d", p.NBTLimit)
	}
	p.SetConflictLimit(0)
	if p.NBTLimit != 0 {
		t.Fatalf("unlimited conflict limit = %d", p.NBTLimit)
	}
}
