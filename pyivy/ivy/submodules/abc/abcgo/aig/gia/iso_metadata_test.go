package gia

import "testing"

func TestIsomorphismMetadataOwnershipAndValidation(t *testing.T) {
	p := Gia_ManStart(8)
	Gia_ManAppendCi(p)
	pairs := []int32{0, 1}
	p.SetIsomorphism(pairs)
	pairs[1] = 0
	got := p.Isomorphism()
	if got[1] != 1 {
		t.Fatal("setter retained caller-owned storage")
	}
	got[1] = 0
	if p.Isomorphism()[1] != 1 {
		t.Fatal("getter exposed manager-owned storage")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for wrong pair count")
		}
	}()
	p.SetIsomorphism([]int32{0})
}
