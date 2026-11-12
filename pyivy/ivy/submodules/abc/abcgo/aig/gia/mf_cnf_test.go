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
	pars.fCutMin = 1
	pars.fGenCnf = 1
	pars.fCoarsen = 0
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
