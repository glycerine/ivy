package cba

import "testing"

func TestSourceFonEncoding(t *testing.T) {
	if Cba_FonIsReal(0) != 0 || Cba_FonIsConst(0) != 0 {
		t.Fatal("zero must remain undriven")
	}
	for _, id := range []int32{0, 1, 10, 1 << 20} {
		f := Cba_FonFromConst(id)
		if Cba_FonIsConst(f) != 1 || Cba_FonIsReal(f) != 0 || Cba_FonConst(f) != id {
			t.Fatalf("id %d fon %d", id, f)
		}
	}
	if Cba_FonIsReal(1) != 1 || Cba_FonIsConst(1) != 0 {
		t.Fatal("positive fon")
	}
}
func TestSourceFonEncodingPreconditions(t *testing.T) {
	for _, fn := range []func(){func() { Cba_FonConst(0) }, func() { Cba_FonFromConst(-1) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("missing source assertion")
				}
			}()
			fn()
		}()
	}
}
