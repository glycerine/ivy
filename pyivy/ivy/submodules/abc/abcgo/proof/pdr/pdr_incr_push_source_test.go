package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// The C incremental push drops contained cubes, moves inductive cubes, and
// leaves a noninductive cube in its original frame.
func TestIncrementalPushSubsumesAndSeparatesCubes(t *testing.T) {
	for _, mono := range []int32{0, 1} {
		g := aig.NewGraph()
		x, q0, q1 := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.Const0())
		g.AddOutput(g.Const0()) // q0' = 0, so q0=1 is blocked
		g.AddOutput(x)          // q1' = x, so q1=1 is reachable
		_ = q0
		_ = q1
		aig.Aig_ManSetRegNum(g, 2)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.FMonoCnf = mono
		p := Pdr_ManStart(g, &pars, nil)
		Pdr_ManCreateSolver(p, 0)
		Pdr_ManCreateSolver(p, 1)
		push := testPdrSet([]int32{0}, nil)
		contained := testPdrSet([]int32{0, 2}, nil)
		stay := testPdrSet([]int32{2}, nil)
		vec.Vec_VecPush(p.vClauses, 1, contained)
		vec.Vec_VecPush(p.vClauses, 1, stay)
		vec.Vec_VecPush(p.vClauses, 1, push)
		next := IPdr_ManPushClausesK(p, 1)
		current := vec.Vec_VecEntry(p.vClauses, 1)
		if vec.Vec_PtrSize(next) != 1 || vec.Vec_PtrEntry(next, 0) != push || vec.Vec_PtrSize(current) != 1 || vec.Vec_PtrEntry(current, 0) != stay || contained.nRefs != 0 {
			t.Fatalf("mono=%d pushed=%d remaining=%d contained refs=%d", mono, vec.Vec_PtrSize(next), vec.Vec_PtrSize(current), contained.nRefs)
		}
		Pdr_SetDeref(push)
		vec.Vec_PtrFree(next)
		Pdr_ManStop(p)
	}
}

// ABC asserts when a last-frame push query returns unknown.
func TestIncrementalPushAssertsOnExpiredQuery(t *testing.T) {
	p, _ := pdrTestSATManager(t, false)
	Pdr_ManCreateSolver(p, 1)
	vec.Vec_VecPush(p.vClauses, 1, testPdrSet([]int32{0}, nil))
	p.timeToStop = util.Abc_Clock() - 1
	defer Pdr_ManStop(p)
	defer func() {
		if recover() == nil {
			t.Fatal("source assertion did not fire on an unknown push")
		}
	}()
	IPdr_ManPushClausesK(p, 1)
}
