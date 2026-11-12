package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSbmManagerCheckSol(t *testing.T) {
	p := Sbm_ManAlloc(3)
	p.nInputs = 1
	p.LitShift = 4
	vec.Vec_WecPush(p.vObjCuts, 0, 0)
	vec.Vec_WecPush(p.vObjCuts, 1, 0)
	vec.Vec_WecPush(p.vCuts, 0, 4)
	vec.Vec_IntPush(p.vSolCuts, 0)
	vec.Vec_IntPush(p.vRoots, 0)
	if Sbm_ManCheckSol(p, p.vSolCuts) != 1 {
		t.Fatal("valid solution rejected")
	}
	vec.Vec_IntPush(p.vRoots, 1)
	if Sbm_ManCheckSol(p, p.vSolCuts) != 0 {
		t.Fatal("unmapped root accepted")
	}
	Sbm_ManStop(p)
}

func TestSbmMappingCNF(t *testing.T) {
	p := Sbm_ManAlloc(3)
	p.nInputs = 1
	p.LitShift = 4
	vec.Vec_WecPush(p.vCuts, 0, 4)
	vec.Vec_WecPush(p.vCuts, 1, 5)
	vec.Vec_WecPush(p.vObjCuts, 0, 1)
	vec.Vec_WecPush(p.vObjCuts, 0, 2*p.FirstVar)
	vec.Vec_WecPush(p.vObjCuts, 1, 3)
	vec.Vec_WecPush(p.vObjCuts, 1, 2*(p.FirstVar+1))
	vec.Vec_IntPush(p.vSolCuts, 0)
	vec.Vec_IntPush(p.vRoots, 0)
	if Sbm_ManCreateCnf(p) != 1 {
		t.Fatal("CNF rejected")
	}
	if p.pSat.Solve(bsat.Positive(0)) != bsat.Satisfiable {
		t.Fatal("mapped root should be satisfiable")
	}
	if p.pSat.Solve(bsat.Positive(0), bsat.Negative(p.FirstVar)) != bsat.Unsatisfiable {
		t.Fatal("root must select cut")
	}
	if p.pSat.Solve(bsat.Positive(p.FirstVar), bsat.Positive(p.FirstVar+1)) != bsat.Unsatisfiable {
		t.Fatal("opposite inverter cuts must be exclusive")
	}
}
