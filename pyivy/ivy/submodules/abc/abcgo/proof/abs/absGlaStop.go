package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
)

// Ga2_ManStop restores GIA scratch state and releases manager references.
func Ga2_ManStop(p *Ga2_Man_t) {
	if p == nil {
		return
	}
	gia.Gia_ManSetMapping(p.pGia, nil)
	gia.Gia_ManSetPhase(p.pGia)
	if p.pPars.fVerbose != 0 {
		if p.pSat != nil {
			stats := p.pSat.Statistics()
			fmt.Printf("SAT solver:  Var = %d  Cla = %d  Conf = %d  Lrn = %d  Reduce = %d  Cex = %d  ObjsAdded = %d\n", stats.Variables, stats.Clauses, stats.Conflicts, stats.Learned, p.pSat.Reductions(), p.nCexes, p.nObjAdded)
		}
		fmt.Printf("Hash hits = %d.  Hash misses = %d.  Hash overs = %d.  Concurrent calls = %d.\n", p.nHashHit, p.nHashMiss, p.nHashOver, p.nPdrCalls)
	}
	Rnm_ManStop(p.pRnm, 0)
	p.pSat = nil
	p.vCnfs = nil
	p.vId2Lit = nil
	p.vIds = nil
	p.vProofIds = nil
	p.vAbs = nil
	p.vValues = nil
	p.vLits = nil
	p.vIsopMem = nil
	p.pTable = nil
	p.pSopSizes = nil
	p.pSops = nil
}
