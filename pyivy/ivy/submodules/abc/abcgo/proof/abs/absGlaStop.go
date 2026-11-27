package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Ga2_ManStop prints the solver statistics under -v and frees the manager
// (absGla.c).
func Ga2_ManStop(p *Ga2_Man_t) {
	gia.Gia_ManSetMapping(p.pGia, nil)
	gia.Gia_ManSetPhase(p.pGia)
	if p.pPars.FVerbose != 0 {
		util.Abc_Print(1, "SAT solver:  Var = %d  Cla = %d  Conf = %d  Lrn = %d  Reduce = %d  Cex = %d  ObjsAdded = %d\n",
			bsat.Sat_solver2_nvars(p.pSat), bsat.Sat_solver2_nclauses(p.pSat),
			bsat.Sat_solver2_nconflicts(p.pSat), bsat.Sat_solver2_nlearnts(p.pSat),
			bsat.Sat_solver2_NDBreduces(p.pSat), p.nCexes, p.nObjAdded)
	}
	if p.pPars.FVerbose != 0 {
		util.Abc_Print(1, "Hash hits = %d.  Hash misses = %d.  Hash overs = %d.  Concurrent calls = %d.\n",
			p.nHashHit, p.nHashMiss, p.nHashOver, p.nPdrCalls)
	}
	if p.pSat != nil {
		bsat.Sat_solver2_delete(p.pSat)
	}
	p.vCnfs = nil
	p.vId2Lit = nil
	p.vIds = nil
	p.vProofIds = nil
	p.vAbs = nil
	p.vValues = nil
	p.vLits = nil
	p.vIsopMem = nil
	Rnm_ManStop(p.pRnm, 0)
	p.pTable = nil
	p.pSopSizes = nil
	p.pSops = nil
}
