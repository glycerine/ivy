package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Ga2_ManRestart creates a new SAT solver and reloads the abstraction from
// the GIA's gate classes (absGla.c).
func Ga2_ManRestart(p *Ga2_Man_t) {
	var vToAdd *vec.Vec_Int_t
	Lit := []int32{1}
	assert(p.pGia != nil && gia.Gia_ManGateClasses(p.pGia) != nil, "p->pGia != NULL && p->pGia->vGateClasses != NULL")
	assert(gia.Gia_ManPiNum(p.pGia) == 0 || gia.Gia_ObjPhase(gia.Gia_ManPi(p.pGia, 0)) != 0, "Gia_ManPi(p->pGia, 0)->fPhase") // marks are set
	// clear SAT variable numbers (begin with 1)
	if p.pSat != nil {
		bsat.Sat_solver2_delete(p.pSat)
	}
	p.pSat = bsat.Sat_solver2_new()
	bsat.Sat_solver2_SetLearntLimits(p.pSat, p.pPars.NLearnedStart, p.pPars.NLearnedDelta, p.pPars.NLearnedPerce)
	// add clause x0 = 0  (lit0 = 1; lit1 = 0)
	bsat.Sat_solver2_addclause(p.pSat, Lit, -1)
	// remove previous abstraction
	Ga2_ManShrinkAbs(p, 1, 1, 1)
	// start new abstraction
	vToAdd = Ga2_ManAbsDerive(p.pGia)
	assert(bsat.Sat_solver2_Prf2(p.pSat) == nil, "p->pSat->pPrf2 == NULL")
	assert(p.pPars.IFrame < 0, "p->pPars->iFrame < 0")
	Ga2_ManAddToAbs(p, vToAdd)
	p.LimAbs = int32(vec.Vec_IntSize(p.vAbs))
	p.LimPpi = int32(vec.Vec_IntSize(p.vValues))
	// set runtime limit
	if p.pPars.NTimeOut != 0 {
		bsat.Sat_solver2_set_runtime_limit(p.pSat, absCPUClock, int64(p.pPars.NTimeOut)*util.CLOCKS_PER_SEC+p.cpuStart)
	}
	// clean the hash table
	clear(p.pTable)
}
