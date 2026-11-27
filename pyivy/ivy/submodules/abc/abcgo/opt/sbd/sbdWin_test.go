package sbd

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// Sbd_ManExploreCut calls Sbd_ManSolve with up to nCutSize-1 = 9 divisors
// while it drops cut leaves one at a time. Sbd_ManSolve built a cube from
// s_Truths6[iVar] for every divisor in the final conflict, reading past the
// six-entry table for iVar >= 6 (ASan: global-buffer-overflow, sbdWin.c:222;
// CBUGS_FIXED.md #26). The truth table only matters for at most six
// divisors; with more, callers only test it against SBD_SAT_SAT/UNDEC.
func TestSbdManSolveMoreThanSixDivisors(t *testing.T) {
	const nDivs = 8
	pSat := sat_solver_new()
	// variables: 0..7 divisors, 8 pivot, 9 free
	sat_solver_setnvars(pSat, nDivs+2)
	pivot, free := int32(nDivs), int32(nDivs+1)
	// pivot == d7
	sat_solver_addclause(pSat, []int32{Abc_Var2Lit(pivot, 1), Abc_Var2Lit(7, 0)})
	sat_solver_addclause(pSat, []int32{Abc_Var2Lit(pivot, 0), Abc_Var2Lit(7, 1)})
	vDivSet := vec.Vec_IntAlloc(nDivs)
	for i := int32(0); i < nDivs; i++ {
		vec.Vec_IntPush(vDivSet, i)
	}
	vDivVars := vec.Vec_IntAlloc(0)
	vDivValues := vec.Vec_IntAlloc(0)
	vTemp := vec.Vec_IntAlloc(0)
	truth := Sbd_ManSolve(pSat, pivot, free, vDivSet, vDivVars, vDivValues, vTemp)
	if truth == SBD_SAT_SAT || truth == SBD_SAT_UNDEC {
		t.Fatalf("pivot is a function of the divisors; got %#x", truth)
	}
	// With six or fewer divisors the truth table is exact.
	pSat = sat_solver_new()
	sat_solver_setnvars(pSat, 8)
	pivot, free = 6, 7
	// pivot == d1 & !d4
	sat_solver_add_and(pSat, pivot, 1, 4, 0, 1, 0)
	vDivSet = vec.Vec_IntAlloc(6)
	for i := int32(0); i < 6; i++ {
		vec.Vec_IntPush(vDivSet, i)
	}
	truth = Sbd_ManSolve(pSat, pivot, free, vDivSet, vDivVars, vDivValues, vTemp)
	if want := s_Truths6[1] &^ s_Truths6[4]; truth != want {
		t.Fatalf("truth %#x, want %#x", truth, want)
	}
}
