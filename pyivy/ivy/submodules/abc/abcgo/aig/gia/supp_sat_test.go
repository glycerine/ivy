package gia

import "testing"

func TestMin2SatSolveAndRollback(t *testing.T) {
	p := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	m := Gia_Man2SuppStart(p)
	defer Gia_Man2SuppStop(m)
	m.iLits = [2]int32{x, a ^ 1}
	Gia_Min2ManGatherSupp(m)
	Gia_Min2ManSimulate(m)
	if Gia_Min2ManSolve(m) != 1 {
		t.Fatal("expected unsatisfiable conjunction")
	}
	if m.pSat.VariableCount() != 0 {
		t.Fatal("SAT variables not rolled back")
	}
	for _, id := range Vec_IntArray(m.vSatVars) {
		if Gia_Min2ObjSatId(p, Gia_ManObj(p, id)) != -1 {
			t.Fatal("SAT ID not cleared")
		}
	}
	m.iLits = [2]int32{x, a}
	Gia_Min2ManGatherSupp(m)
	Vec_WrdWriteEntry(m.vSims, int(a>>1), 0)
	Vec_WrdWriteEntry(m.vSims, int(b>>1), 0)
	if Gia_Min2ManSimulate(m) != 1 || Gia_Min2ManSolve(m) != 0 {
		t.Fatal("expected satisfiable conjunction")
	}
	if Gia_Min2ManSimulate(m) != 0 || m.nSatSat != 1 || m.iPattern != 2 {
		t.Fatal("SAT witness was not retained for simulation")
	}
}
