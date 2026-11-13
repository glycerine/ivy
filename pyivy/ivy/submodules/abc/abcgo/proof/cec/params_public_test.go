package cec

import "testing"

func TestPublicConflictLimit(t *testing.T) {
	var p Cec_ParCec_t
	p.SetConflictLimit(37)
	if p.nBTLimit != 37 {
		t.Fatalf("conflict limit = %d", p.nBTLimit)
	}
	p.SetConflictLimit(0)
	if p.nBTLimit != 0 {
		t.Fatalf("unlimited conflict limit = %d", p.nBTLimit)
	}
}
