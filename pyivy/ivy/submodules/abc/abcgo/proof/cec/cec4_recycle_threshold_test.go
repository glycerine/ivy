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
			pars := &Cec_ParFra_t{JType: 0, NWords: 1, NRounds: 1, NItersMax: 10, NBTLimit: 1000, nSatVarMax: tc.maxVars, NCallsRecycle: tc.calls}
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

func runCec4PigeonholeRecycle(t *testing.T, conflictLimit int32) cec4SweepStats {
	t.Helper()
	const pigeons, holes = 6, 5
	g := gia.Gia_ManStart(1024)
	inputs := make([][]int32, pigeons)
	for i := range inputs {
		inputs[i] = make([]int32, holes)
		for j := range inputs[i] {
			inputs[i][j] = gia.Gia_ManAppendCi(g)
		}
	}
	x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManHashAlloc(g)
	formula := int32(1)
	for i := range inputs {
		clause := int32(0)
		for _, v := range inputs[i] {
			clause = gia.Gia_ManHashOr(g, clause, v)
		}
		formula = gia.Gia_ManHashAnd(g, formula, clause)
	}
	for h := 0; h < holes; h++ {
		for i := 0; i < pigeons; i++ {
			for j := 0; j < i; j++ {
				clause := gia.Gia_ManHashOr(g, inputs[i][h]^1, inputs[j][h]^1)
				formula = gia.Gia_ManHashAnd(g, formula, clause)
			}
		}
	}
	xy, yz := gia.Gia_ManHashAnd(g, x, y), gia.Gia_ManHashAnd(g, y, z)
	left, right := gia.Gia_ManHashAnd(g, xy, z), gia.Gia_ManHashAnd(g, x, yz)
	gia.Gia_ManAppendCo(g, formula)
	gia.Gia_ManAppendCo(g, left)
	gia.Gia_ManAppendCo(g, right)
	pars := &Cec_ParFra_t{JType: 0, NWords: 1, NRounds: 1, NItersMax: 10, NGenIters: 0, NBTLimit: conflictLimit, nSatVarMax: 1, NCallsRecycle: 0}
	manager := Cec4_ManCreate(g, pars)
	defer Cec4_ManDestroy(manager)
	out, cex, stats := cec4Sweep(g, pars, false, true, manager)
	if out == nil || cex != nil {
		t.Fatalf("out=%v cex=%v", out, cex)
	}
	if manager.nSatUndec != int32(stats.unknown) {
		t.Fatalf("manager unknown count %d differs from sweep count %d", manager.nSatUndec, stats.unknown)
	}
	return stats
}

func TestCec4RecycleStatusesAtConflictBudgets(t *testing.T) {
	// At the two-conflict setting both ports return the same counts. At one
	// conflict, the SAT backend cutoff classifies one query differently; both
	// remain sound, but this is a known low-budget status parity limit.
	for _, tc := range []struct {
		limit int32
		want  cec4SweepStats
	}{
		{limit: 1, want: cec4SweepStats{sat: 42, unknown: 14, recycles: 55}},
		{limit: 2, want: cec4SweepStats{sat: 45, unsat: 1, unknown: 10, recycles: 55}},
	} {
		t.Run(string(rune('0'+tc.limit))+"-conflicts", func(t *testing.T) {
			if got := runCec4PigeonholeRecycle(t, tc.limit); got != tc.want {
				t.Fatalf("counts = %+v, want %+v", got, tc.want)
			}
		})
	}
}
