package cba

import "testing"

func TestTypePredicatesSourceBoundaries(t *testing.T) {
	unary := map[Cba_ObjType_t]bool{CBA_BOX_BUF: true, CBA_BOX_INV: true, CBA_BOX_LNOT: true, CBA_BOX_MIN: true, CBA_BOX_SQRT: true, CBA_BOX_ABS: true}
	mux := map[Cba_ObjType_t]bool{CBA_BOX_MUX: true, CBA_BOX_NMUX: true, CBA_BOX_SEL: true, CBA_BOX_PSEL: true}
	for kind := Cba_ObjType_t(-1); kind <= CBA_BOX_LAST; kind++ {
		wantBox := kind >= CBA_OBJ_BOX && kind < CBA_BOX_LAST
		wantSeq := kind >= CBA_BOX_RAM && kind <= CBA_BOX_DFFRS
		wantUnary := unary[kind] || kind >= CBA_BOX_RAND && kind <= CBA_BOX_RXNOR
		for _, tc := range []struct {
			name string
			got  int32
			want bool
		}{
			{"box", Cba_TypeIsBox(kind), wantBox},
			{"seq", Cba_TypeIsSeq(kind), wantSeq},
			{"unary", Cba_TypeIsUnary(kind), wantUnary},
			{"mux", Cba_TypeIsMux(kind), mux[kind]},
		} {
			if (tc.got != 0) != tc.want || tc.got < 0 || tc.got > 1 {
				t.Errorf("%s(%d)=%d, want %v", tc.name, kind, tc.got, tc.want)
			}
		}
	}
}
