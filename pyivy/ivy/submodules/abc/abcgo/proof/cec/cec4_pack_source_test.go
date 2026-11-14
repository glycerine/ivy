package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// cecSatG2.c asserts the current pattern cursor, not the requested bit.
func TestCec4PackTryChecksSourcePatternCursor(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCi(g)
	p := Cec4_ManCreate(g, &Cec_ParFra_t{NWords: 1})
	Cec4_ManSimAlloc(p, 1)
	lits := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(lits, 2)
	for _, cursor := range []int{0, 64} {
		p.pattern = cursor
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("cursor %d: expected source assertion", cursor)
				}
			}()
			Cec4_ManPackAddPatternTry(p, 1, lits)
		}()
	}
	p.pattern = 1
	if got := Cec4_ManPackAddPatternTry(p, 1, lits); got != 1 {
		t.Fatalf("valid cursor result %d", got)
	}
}
