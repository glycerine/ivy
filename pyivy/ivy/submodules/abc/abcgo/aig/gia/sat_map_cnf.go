package gia

import "github.com/glycerine/abcgo/misc/vec"

func Sbm_ManCreateCnf(p *Sbm_Man_t) int32 {
	if Sbm_ManCheckSol(p, p.vSolCuts) == 0 {
		return 0
	}
	if p.pSat.NumVars() != p.FirstVar {
		panic("gia: SAT mapping solver not reset")
	}
	total := p.FirstVar + int32(vec.Vec_WecSize(p.vCuts))
	for p.pSat.NumVars() < total {
		p.pSat.NewVariable()
	}
	add := func(lits ...int32) {
		if !p.pSat.AddClause(lits...) {
			panic("gia: inconsistent SAT mapping clauses")
		}
	}
	var previous []int32
	for i := 0; i < vec.Vec_WecSize(p.vObjCuts); i++ {
		lits := vec.Vec_IntArray(vec.Vec_WecEntry(p.vObjCuts, i))
		if len(lits) < 2 {
			panic("gia: empty SAT object cut list")
		}
		add(lits...)
		if i&1 != 0 {
			add(lits[len(lits)-1]^1, previous[len(previous)-1]^1)
		}
		previous = lits
	}
	for i := 0; i < vec.Vec_WecSize(p.vCuts); i++ {
		cut := vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, i))
		for k, lit := range cut {
			var variable int32
			if lit/2-1 < p.nInputs {
				if k == 0 {
					panic("gia: invalid input cut root")
				}
				variable = int32(vec.Vec_WecSize(p.vObjCuts)) + lit/2 - 1
			} else {
				variable = lit - p.LitShift
			}
			if variable < 0 {
				panic("gia: invalid SAT cut variable")
			}
			add(2*variable, 2*(p.FirstVar+int32(i))+1)
		}
	}
	return 1
}
