package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
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
