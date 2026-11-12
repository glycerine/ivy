package gia

import "testing"

func TestOfCutStorage(t *testing.T) {
	a, b := ofUnitCut(4), ofUnitCut(9)
	b.size = 2
	b.leaves[1] = 12
	b.function = 0x7ffffff
	s := newOfCutStorage()
	h := s.save([]*ofCut{&a, &b})
	data := s.at(h)
	if data[0] != 2 || ofStoredSize(data[1:]) != 1 || ofStoredFunction(data[1:]) != 2 {
		t.Fatal("invalid first cut")
	}
	second := data[6:]
	if ofStoredFunction(second) != 0x7ffffff || ofStoredDelay1(second) != -1 || ofStoredDelay2(second) != -1 || ofStoredFlow(second) != -1 {
		t.Fatal("invalid second cut")
	}
	ofStoredSetDelay1(second, 10)
	ofStoredSetDelay2(second, 20)
	ofStoredSetFlow(second, 30)
	if ofStoredDelay1(second) != 10 || ofStoredDelay2(second) != 20 || ofStoredFlow(second) != 30 {
		t.Fatal("invalid annotations")
	}
	s.lift(h)
	if ofStoredVar(second, 1) != 12 || ofStoredFlag(second, 1) != 0 {
		t.Fatal("incorrect lifted leaf")
	}
	ofStoredSetFlag(second, 1, 1)
	if ofStoredVar(second, 1) != 12 || ofStoredFlag(second, 1) != 1 {
		t.Fatal("incorrect leaf flag")
	}
}

func TestOfCutStoragePageBoundary(t *testing.T) {
	s := newOfCutStorage()
	s.next = 0xfffe
	c := ofUnitCut(1)
	h := s.save([]*ofCut{&c})
	if h != 1<<16 || s.at(h)[0] != 1 {
		t.Fatalf("page handle %d", h)
	}
}
