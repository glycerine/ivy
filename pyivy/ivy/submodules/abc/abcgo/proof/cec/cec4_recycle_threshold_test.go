package cec

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestCec4SweepRecycleThresholdBoundary(t *testing.T) {
	for _, tc := range []struct {
		name         string
		maxVars      int32
		calls        int32
		wantRecycles int32
	}{
		{name: "disabled-by-variable-limit", maxVars: 1000, calls: 0, wantRecycles: 0},
		{name: "enabled-after-zero-call-threshold", maxVars: 1, calls: 0, wantRecycles: 1},
		{name: "strictly-over-call-threshold", maxVars: 1, calls: 2, wantRecycles: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gia.Gia_ManStart(20)
			x, y, z, w := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			xy, yz := gia.Gia_ManAppendAnd(g, x, y), gia.Gia_ManAppendAnd(g, y, z)
			left, right := gia.Gia_ManAppendAnd(g, xy, z), gia.Gia_ManAppendAnd(g, x, yz)
			yw, xw := gia.Gia_ManAppendAnd(g, y, w), gia.Gia_ManAppendAnd(g, x, w)
			left2, right2 := gia.Gia_ManAppendAnd(g, yw, x), gia.Gia_ManAppendAnd(g, y, xw)
			gia.Gia_ManAppendCo(g, left)
			gia.Gia_ManAppendCo(g, right)
			gia.Gia_ManAppendCo(g, left2)
			gia.Gia_ManAppendCo(g, right2)
			pars := &Cec_ParFra_t{jType: 0, nWords: 1, nRounds: 1, nItersMax: 10, nBTLimit: 1000, nSatVarMax: tc.maxVars, nCallsRecycle: tc.calls}
			manager := Cec4_ManCreate(g, pars)
			defer Cec4_ManDestroy(manager)
			out, cex, _ := cec4Sweep(g, pars, false, true, manager)
			if out == nil || cex != nil {
				t.Fatalf("out=%v cex=%v", out, cex)
			}
			if manager.nRecycles != tc.wantRecycles {
				t.Fatalf("recycles = %d, want %d (calls=%d vars=%d)", manager.nRecycles, tc.wantRecycles, manager.nCallsSince, len(manager.cnf.ids))
			}
			if manager.nSatUnsat == 0 {
				t.Fatal("associative-equivalence queries did not exercise UNSAT proof status")
			}
		})
	}
}
