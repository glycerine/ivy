package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Ga2_ManRestart reinitializes the solver around the current abstraction map.
func Ga2_ManRestart(p *Ga2_Man_t) {
	if p.pGia == nil || gia.Gia_ManGateClasses(p.pGia) == nil || p.pPars.IFrame >= 0 {
		panic("abs: invalid GLA restart state")
	}
	if gia.Gia_ManPiNum(p.pGia) > 0 && gia.Gia_ObjPhase(gia.Gia_ManPi(p.pGia, 0)) == 0 {
		panic("abs: missing GLA marks")
	}
	p.pSat = bsat.NewLabeledSolver2()
	p.pSat.SetLearnedPolicy(p.pPars.NLearnedStart, p.pPars.NLearnedDelta, p.pPars.NLearnedPerce)
	p.pSat.AddClause(-1, bsat.Literal(1))
	p.proofActive = false
	Ga2_ManShrinkAbs(p, 1, 1, 1)
	add := Ga2_ManAbsDerive(p.pGia)
	Ga2_ManAddToAbs(p, add)
	p.LimAbs = int32(vec.Vec_IntSize(p.vAbs))
	p.LimPpi = int32(vec.Vec_IntSize(p.vValues))
	if p.cpuDeadline != 0 {
		p.pSat.SetRuntimeDeadline(absCPUClock, p.cpuDeadline)
	}
	clear(p.pTable)
}
