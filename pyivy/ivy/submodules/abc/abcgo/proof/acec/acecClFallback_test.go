package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestSourceDetectAdditionalRejectsAmbiguousPolarity(t *testing.T) {
	p := gia.Gia_ManStart(32)
	gia.Gia_ManHashStart(p)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	root := gia.Gia_ManHashXor(p, gia.Gia_ManHashXor(p, a, b), c)
	gia.Gia_ManAppendCo(p, root)
	gia.Gia_ManHashStop(p)
	// The C implementation asserts on inconsistent leaf polarity here.
	defer func() {
		if got := recover(); got != "acec: inconsistent leaf polarity" {
			t.Fatalf("polarity panic = %v", got)
		}
	}()
	Acec_DetectAdditional(p, 0)
}
