package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
	"time"
)

func Cec2_ManCreate(g *gia.Gia_Man_t, pars *Cec2_Par_t) *Cec2_Man_t {
	p := &Cec2_Man_t{pAig: g, pPars: pars, timeStart: time.Now().UnixNano()}
	p.pNew = gia.Gia_ManStart(gia.Gia_ManObjNum(g))
	gia.Gia_ManFillValue(g)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(g), 0)
	for i := int32(0); i < gia.Gia_ManCiNum(g); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManCi(g, i), gia.Gia_ManAppendCi(p.pNew))
	}
	gia.Gia_ManHashAlloc(p.pNew)
	gia.Gia_ManExtendCopy2Array(p.pNew)
	p.pSat = satoko.New()
	p.pSat.Options().ConflictLimit = int64(pars.nConfLimit)
	if pars.fUseCones == 0 {
		p.pSat.Bookmark()
	}
	p.vFrontier = vec.Vec_PtrAlloc(1000)
	p.vFanins = vec.Vec_PtrAlloc(100)
	p.vNodesNew = vec.Vec_IntAlloc(100)
	p.vSatVars = vec.Vec_IntAlloc(100)
	p.vObjSatPairs = vec.Vec_IntAlloc(100)
	p.vCexTriples = vec.Vec_IntAlloc(100)
	g.SetData(p.pSat)
	return p
}
func Cec2_ManDestroy(p *Cec2_Man_t) {
	if p.pPars.fVerbose != 0 {
		total := time.Now().UnixNano() - p.timeStart
		sat := p.timeSatSat + p.timeSatUnsat + p.timeSatUndec
		entries := []struct {
			name string
			ns   int64
		}{
			{"SAT solving", sat}, {"  sat      ", p.timeSatSat}, {"  unsat    ", p.timeSatUnsat},
			{"  fail     ", p.timeSatUndec}, {"Simulation ", p.timeSim}, {"Refinement ", p.timeRefine},
			{"Rollback   ", p.timeExtra}, {"Other      ", total - sat - p.timeSim - p.timeRefine - p.timeExtra}, {"TOTAL      ", total},
		}
		for _, e := range entries {
			percent := float64(0)
			if total != 0 {
				percent = 100 * float64(e.ns) / float64(total)
			}
			fmt.Printf("%s =%9.2f sec (%6.2f %%)\n", e.name, float64(e.ns)/float64(time.Second), percent)
		}
	}
	gia.Gia_ManCecSimFree(p.pAig)
	gia.Gia_ManCleanMark01(p.pAig)
	// The C manager's borrowed pData points to the solver being destroyed.
	if p.pAig.Data() == p.pSat {
		p.pAig.SetData(nil)
	}
	gia.Gia_ManStopP(&p.pNew)
	p.pSat = nil
	p.vFrontier = nil
	p.vFanins = nil
	p.vNodesNew = nil
	p.vSatVars = nil
	p.vObjSatPairs = nil
	p.vCexTriples = nil
}
