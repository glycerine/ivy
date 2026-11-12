package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

type Sbl_Man_t struct {
	pSat                                                                                            *bsat.Solver
	vCardVars                                                                                       *Vec_Int_t
	nVars, LogN, Power2, FirstVar                                                                   int32
	nTried, nImproved, nRuns, nHashWins, nSmallWins, nLargeWins, nIterOuts                          int32
	LutSize, nBTLimit, DelayMax, nEdges, fDelay, fReverse, fVerbose, fVeryVerbose, fVeryVeryVerbose int32
	pGia                                                                                            *Gia_Man_t
	vLeaves, vAnds, vNodes, vRoots, vRootVars                                                       *Vec_Int_t
	pHash                                                                                           *vec.Hsh_VecMan_t
	vArrs, vReqs                                                                                    *Vec_Int_t
	vWindow                                                                                         *Vec_Wec_t
	vPath, vEdges                                                                                   *Vec_Int_t
	vCutsI1, vCutsI2, vCutsN1, vCutsN2                                                              *Vec_Wrd_t
	vCutsNum, vCutsStart, vCutsObj                                                                  *Vec_Int_t
	vTempI1, vTempI2, vTempN1, vTempN2                                                              *Vec_Wrd_t
	vSolInit, vSolCur, vSolBest, vLits, vAssump, vPolar                                             *Vec_Int_t
	timeWin, timeCut, timeSat, timeSatSat, timeSatUns, timeSatUnd, timeTime, timeTotal, timeOther   time.Duration
	timeStart                                                                                       time.Time
}

func Sbl_ManAlloc(g *Gia_Man_t, number int32) *Sbl_Man_t {
	if number < 1 || number > 128 {
		panic("gia: invalid SAT window size")
	}
	p := &Sbl_Man_t{pGia: g, nVars: number, LogN: util.Abc_Base2Log(uint32(number))}
	p.Power2 = 1 << p.LogN
	p.pSat = Sbm_AddCardinSolver(p.LogN, &p.vCardVars)
	p.FirstVar = p.pSat.NumVars()
	p.vLeaves = vec.Vec_IntAlloc(int(number))
	p.vAnds = vec.Vec_IntAlloc(int(number))
	p.vNodes = vec.Vec_IntAlloc(int(number))
	p.vRoots = vec.Vec_IntAlloc(int(number))
	p.vRootVars = vec.Vec_IntAlloc(int(number))
	p.pHash = vec.Hsh_VecManStart(1000)
	p.vArrs = vec.Vec_IntAlloc(0)
	p.vReqs = vec.Vec_IntAlloc(0)
	p.vWindow = vec.Vec_WecAlloc(128)
	p.vPath = vec.Vec_IntAlloc(32)
	p.vEdges = vec.Vec_IntAlloc(32)
	p.vCutsI1 = vec.Vec_WrdAlloc(1000)
	p.vCutsI2 = vec.Vec_WrdAlloc(1000)
	p.vCutsN1 = vec.Vec_WrdAlloc(1000)
	p.vCutsN2 = vec.Vec_WrdAlloc(1000)
	p.vCutsNum = vec.Vec_IntAlloc(64)
	p.vCutsStart = vec.Vec_IntAlloc(64)
	p.vCutsObj = vec.Vec_IntAlloc(1000)
	p.vTempI1 = vec.Vec_WrdAlloc(32)
	p.vTempI2 = vec.Vec_WrdAlloc(32)
	p.vTempN1 = vec.Vec_WrdAlloc(32)
	p.vTempN2 = vec.Vec_WrdAlloc(32)
	p.vSolInit = vec.Vec_IntAlloc(64)
	p.vSolCur = vec.Vec_IntAlloc(64)
	p.vSolBest = vec.Vec_IntAlloc(64)
	p.vLits = vec.Vec_IntAlloc(64)
	p.vAssump = vec.Vec_IntAlloc(64)
	p.vPolar = vec.Vec_IntAlloc(1000)
	Gia_ManFillValue(g)
	return p
}
func Sbl_ManClean(p *Sbl_Man_t) {
	p.timeStart = time.Now()
	p.pSat = Sbm_AddCardinSolver(p.LogN, &p.vCardVars)
	for _, v := range []*Vec_Int_t{p.vLeaves, p.vAnds, p.vNodes, p.vRoots, p.vRootVars, p.vArrs, p.vReqs, p.vPath, p.vEdges, p.vCutsNum, p.vCutsStart, p.vCutsObj, p.vSolInit, p.vSolCur, p.vSolBest, p.vLits, p.vAssump, p.vPolar} {
		vec.Vec_IntClear(v)
	}
	for _, v := range []*Vec_Wrd_t{p.vCutsI1, p.vCutsI2, p.vCutsN1, p.vCutsN2, p.vTempI1, p.vTempI2, p.vTempN1, p.vTempN2} {
		vec.Vec_WrdClear(v)
	}
	vec.Vec_WecClear(p.vWindow)
	Gia_ManFillValue(p.pGia)
}
func Sbl_ManStop(p *Sbl_Man_t) { p.pSat = nil }
