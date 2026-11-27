package abs

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestOldGlaUnsatCoreForConstant(t *testing.T) {
	p := &Gla_Man_t{nObjs: 3, pObjs: make([]Gla_Obj_t, 3), nSatVars: 1, vAddedNew: vec.Vec_IntAlloc(0), vAbs: vec.Vec_IntAlloc(1), vTemp: vec.Vec_IntAlloc(3), pSat: bsat.Sat_solver2_new()}
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	p.pObjs[1].fConst = 1
	p.pObjs[1].fAbs = 1
	vec.Vec_IntPush(p.vAbs, 1)
	p.pObjs[2].fPo = 1
	p.pObjs[2].Fanins[0] = 1
	p.pObjRoot = &p.pObjs[2]
	Gia_GlaAddTimeFrame(p, 0)
	var status int32
	core := Gla_ManUnsatCore(p, 0, p.pSat, 0, 0, &status, nil)
	// The assumption contradicts a root unit with no reason clause, so, as in
	// satSolver2.c, no proof (and no core) is produced; only UNSAT is reported.
	if status != 1 || core != nil {
		t.Fatalf("status=%d core=%v", status, vec.Vec_IntArray(core))
	}
}

func TestOldGlaUnsatCoreDeadlineAndRetry(t *testing.T) {
	p := &Gla_Man_t{nObjs: 3, pObjs: make([]Gla_Obj_t, 3), pSat: bsat.Sat_solver2_new()}
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	p.pObjs[2].fPo = 1
	p.pObjs[2].Fanins[0] = 1
	p.pObjRoot = &p.pObjs[2]
	vec.Vec_IntPush(&p.pObjs[1].vFrames, 1)
	ga2NewVar(p.pSat)
	ga2NewVar(p.pSat)
	bsat.Sat_solver2_set_runtime_limit(p.pSat, nil, 1)
	var status, conflicts int32
	if core := Gla_ManUnsatCore(p, 0, p.pSat, 0, 0, &status, &conflicts); status != -1 || core != nil {
		t.Fatalf("expired deadline: status=%d core=%v", status, core)
	}
	bsat.Sat_solver2_set_runtime_limit(p.pSat, nil, 0)
	if core := Gla_ManUnsatCore(p, 0, p.pSat, 0, 0, &status, &conflicts); status != 0 || core != nil {
		t.Fatalf("retry: status=%d core=%v", status, core)
	}
}
func TestOldGlaRollbackClearsObjectsAndVariables(t *testing.T) {
	p := &Gla_Man_t{nObjs: 3, pObjs: make([]Gla_Obj_t, 3), nSatVars: 1, vAddedNew: vec.Vec_IntAlloc(4), vAbs: vec.Vec_IntAlloc(2), nAbsOld: 1}
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	p.pObjs[1].fAbs = 1
	p.pObjs[2].fAbs = 1
	vec.Vec_IntPush(p.vAbs, 1)
	vec.Vec_IntPush(p.vAbs, 2)
	Gla_ManGetVar(p, 2, 0)
	Gla_ManGetVar(p, 2, 1)
	Gla_ManRollBack(p)
	if p.pObjs[2].fAbs != 0 || vec.Vec_IntSize(p.vAbs) != 1 || Gla_ManCheckVar(p, 2, 0) != 0 || Gla_ManCheckVar(p, 2, 1) != 0 {
		t.Fatal("rollback failed")
	}
}
