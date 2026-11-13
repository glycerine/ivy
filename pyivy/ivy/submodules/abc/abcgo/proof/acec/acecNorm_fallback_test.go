package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"runtime"
	"testing"
)

func TestNormalizeWithoutArithmeticBox(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a^1)
	if box := Acec_DeriveBox(p, nil, 0, 0, 0); box != nil {
		t.Fatal("expected no arithmetic box")
	}
	// Compiled C dereferences the absent box in Acec_InsertBox.
	defer func() {
		if _, ok := recover().(runtime.Error); !ok {
			t.Fatal("expected nil dereference of absent arithmetic box")
		}
	}()
	Acec_Normalize(p, 0, 0)
}
