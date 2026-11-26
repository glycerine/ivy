package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestNormalizeWithoutArithmeticBox(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a^1)
	if box := Acec_DeriveBox(p, nil, 0, 0, 0); box != nil {
		t.Fatal("expected no arithmetic box")
	}
	// ABC dereferenced the absent box in Acec_InsertBox; now it reports it
	// and returns a copy (CBUGS_FIXED.md #11).
	q := Acec_Normalize(p, 0, 0)
	if q == nil || gia.Gia_ManCiNum(q) != 1 || gia.Gia_ManCoNum(q) != 1 || gia.Gia_ManAndNum(q) != 0 {
		t.Fatal("expected a copy of the AIG")
	}
}
