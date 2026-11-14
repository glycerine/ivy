package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMfDeriveCnfs(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.FCutMin = 1
	pars.FGenCnf = 1
	pars.FCoarsen = 0
	p := Mf_ManAlloc(g, &pars)
	Mf_ManComputeCuts(p)
	var vars, clauses, lits int32
	cnfs := Mf_ManDeriveCnfs(p, &vars, &clauses, &lits)
	id := Mf_CutFunc(Mf_ObjCutBest(p, x>>1)) >> 1
	offset := vec.Vec_IntEntry(cnfs, int(id))
	if vars != 5 || clauses != 6 || lits != 12 || vec.Vec_IntEntry(cnfs, int(offset)) != 3 {
		t.Fatalf("vars %d clauses %d lits %d", vars, clauses, lits)
	}
}

func TestMfDeriveCnfSatisfiesAnd(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, x)
	var pars Jf_Par_t
	Mf_ManSetDefaultPars(&pars)
	pars.FCutMin = 1
	pars.FGenCnf = 1
	pars.fCnfMapping = 1
	p := Mf_ManAlloc(g, &pars)
	p.pGia0 = g
	Mf_ManComputeCuts(p)
	cnf := Mf_ManDeriveCnf(p, 0, 0)
	for mask := 0; mask < 1<<cnf.Variables; mask++ {
		valid := true
		for _, clause := range cnf.Clauses {
			satisfied := false
			for _, lit := range clause {
				value := (mask >> uint(lit>>1)) & 1
				if value^int(lit&1) == 1 {
					satisfied = true
					break
				}
			}
			if !satisfied {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		av := (mask >> uint(cnf.VarNums[a>>1])) & 1
		bv := (mask >> uint(cnf.VarNums[b>>1])) & 1
		ov := (mask >> uint(cnf.VarNums[g.pObjs[len(g.pObjs)-1].id])) & 1
		if ov != av&bv {
			t.Fatalf("invalid assignment %b", mask)
		}
	}
	if len(cnf.Mapping) == 0 {
		t.Fatal("missing CNF mapping")
	}
}
