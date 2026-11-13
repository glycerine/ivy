package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestFxDrivers(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	gia.Gia_ManAppendCo(p, gia.Gia_ManAppendAnd(p, a, b))
	if got := Bmc_FxCompute(p); got != 1 {
		t.Fatalf("dual cover: %d", got)
	}
	if got := Bmc_FxComputeOne(p, 2, 2); got != 1 {
		t.Fatalf("divisor extraction: %d", got)
	}
}
