package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestNormalizeWithoutArithmeticBox(t *testing.T) {
	p := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, a^1)
	// The C implementation passes the absent box directly to Acec_InsertBox.
	defer func() {
		if recover() == nil {
			t.Fatal("expected absent arithmetic box to violate source precondition")
		}
	}()
	Acec_Normalize(p, 0, 0)
}
