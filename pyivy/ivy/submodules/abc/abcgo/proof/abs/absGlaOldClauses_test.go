package abs

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func TestOldGlaCutClauses(t *testing.T) {
	for assignment := 0; assignment < 8; assignment++ {
		p := &Gla_Man_t{nObjs: 4, pObjs: make([]Gla_Obj_t, 4), nSatVars: 1, vAddedNew: vec.Vec_IntAlloc(0), pSat: bsat.NewLabeledSolver2()}
		for i := range p.pObjs {
			p.pObjs[i].id = int32(i)
		}
		p.pObjs[1].fPi = 1
		p.pObjs[2].fPi = 1
		p.pObjs[3].fAnd = 1
		p.pObjs[3].iGiaObj = 3
		p.pCnf = cnf.NewMappedCNF(4, [][]int32{{7, 2}, {7, 4}, {6, 3, 5}}, nil, []int32{-1, -1, -1, 0}, []int32{-1, 0, 0, 3})
		Gla_ManAddClauses(p, 3, 0, vec.Vec_IntAlloc(4))
		var assumptions []bsat.Literal
		for id := int32(1); id <= 3; id++ {
			v := Gla_ManGetVar(p, id, 0)
			lit := v * 2
			if assignment&(1<<uint(id-1)) == 0 {
				lit++
			}
			assumptions = append(assumptions, bsat.Literal(lit))
		}
		got := p.pSat.Solve(assumptions...)
		want := (assignment&4 != 0) == (assignment&3 == 3)
		if (got == bsat.Satisfiable) != want {
			t.Fatalf("assignment=%d status=%v want sat=%v", assignment, got, want)
		}
	}
}
func TestOldGlaRegisterInitialAndNextFrame(t *testing.T) {
	p := &Gla_Man_t{nObjs: 3, pObjs: make([]Gla_Obj_t, 3), nSatVars: 1, vAddedNew: vec.Vec_IntAlloc(0), pSat: bsat.NewLabeledSolver2()}
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	p.pObjs[1].fPi = 1
	p.pObjs[2].fRo = 1
	p.pObjs[2].nFanins = 1
	p.pObjs[2].Fanins[0] = 1
	p.pObjs[2].fCompl0 = 1
	Gla_ManAddClauses(p, 2, 0, vec.Vec_IntAlloc(0))
	Gla_ManAddClauses(p, 2, 1, vec.Vec_IntAlloc(0))
	q0, q1, pi := Gla_ManGetVar(p, 2, 0), Gla_ManGetVar(p, 2, 1), Gla_ManGetVar(p, 1, 0)
	if p.pSat.Solve(bsat.Literal(q0*2)) != bsat.Unsatisfiable {
		t.Fatal("initial flop can be one")
	}
	if p.pSat.Solve(bsat.Literal(pi*2+1), bsat.Literal(q1*2+1)) != bsat.Unsatisfiable {
		t.Fatal("inverted next state incorrect")
	}
}
