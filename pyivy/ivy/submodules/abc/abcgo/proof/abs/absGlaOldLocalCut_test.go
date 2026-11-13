package abs

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func assertSimpleExportPanic(t *testing.T) {
	t.Helper()
	if got := recover(); fmt.Sprint(got) != "gia: simple export changed object ID" {
		t.Fatalf("panic = %v, want simple-export identity assertion", got)
	}
}

// absGlaOld.c:844 calls Gia_ManToAigSimple, which asserts that structural
// hashing preserves object IDs. Duplicate and simplified gates violate it.
func TestOldGlaRejectsNonCanonicalAnds(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		t.Run(map[bool]string{false: "constant", true: "duplicate"}[duplicate], func(t *testing.T) {
			g := gia.Gia_ManStart(8)
			a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			var root int32
			if duplicate {
				gia.Gia_ManAppendAnd(g, a, b)
				root = gia.Gia_ManAppendAnd(g, a, b)
			} else {
				root = gia.Gia_ManAppendAnd(g, a, 0)
			}
			gia.Gia_ManAppendCo(g, root)
			defer assertSimpleExportPanic(t)
			_, _ = glaOldCnfForGia(g)
		})
	}
}
