package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/glucose2"
	"sync"
	"time"
)

func Cec4_ManCreate(g *Gia_Man_t, pars *Cec_ParFra_t) *Cec4_Man_t {
	p := &Cec4_Man_t{pAig: g, pPars: pars, timeStart: cecCPUClock()}
	p.pSat = glucose2.NewABCSolver()
	p.pSat.JustificationMode = int(pars.jType)
	p.vFrontier = vec.Vec_PtrAlloc(1000)
	p.vFanins = vec.Vec_PtrAlloc(100)
	p.vCexMin = vec.Vec_IntAlloc(100)
	p.vClassUpdates = vec.Vec_IntAlloc(100)
	p.vCexStamps = vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
	p.vCands = vec.Vec_IntAlloc(100)
	p.vVisit = vec.Vec_IntAlloc(100)
	p.vPat = vec.Vec_IntAlloc(100)
	p.vDisprPairs = vec.Vec_IntAlloc(100)
	p.vFails = vec.Vec_BitStart(int(gia.Gia_ManObjNum(g)))
	return p
}
func Cec4_ManDestroy(p *Cec4_Man_t) {
	if p.pPars.fVerbose != 0 {
		total := cecCPUClock() - p.timeStart
		sat := p.timeSatSat0 + p.timeSatSat + p.timeSatUnsat0 + p.timeSatUnsat + p.timeSatUndec
		entries := []struct {
			name string
			ns   int64
		}{
			{"SAT solving  ", sat}, {"  sat(easy)  ", p.timeSatSat0}, {"  sat        ", p.timeSatSat},
			{"  unsat(easy)", p.timeSatUnsat0}, {"  unsat      ", p.timeSatUnsat}, {"  fail       ", p.timeSatUndec},
			{"Generate CNF ", p.timeCnf}, {"Generate pats", p.timeGenPats}, {"Simulation   ", p.timeSim},
			{"Refinement   ", p.timeRefine}, {"Resim global ", p.timeResimGlo}, {"Resim local  ", p.timeResimLoc},
			{"Other        ", total - sat - p.timeSim - p.timeRefine - p.timeResimLoc - p.timeGenPats}, {"TOTAL        ", total},
		}
		for _, e := range entries {
			percent := float64(0)
			if total > 0 {
				percent = 100 * float64(e.ns) / float64(total)
			}
			fmt.Printf("%s =%9.2f sec (%6.2f %%)\n", e.name, float64(e.ns)/float64(time.Second), percent)
		}
	}
	gia.Gia_ManCecSimFreeAll(p.pAig)
	gia.Gia_ManCleanMark01(p.pAig)
	if p.pSat != nil {
		p.pSat.Reset()
	}
	gia.Gia_ManStopP(&p.pNew)
	p.pSat = nil
	p.vFrontier = nil
	p.vFanins = nil
	p.vCexMin = nil
	p.vClassUpdates = nil
	p.vCexStamps = nil
	p.vCands = nil
	p.vVisit = nil
	p.vPat = nil
	p.vDisprPairs = nil
	p.vFails = nil
	p.vRefClasses = nil
	p.vRefNodes = nil
	p.vRefBins = nil
	p.pTable = nil
	p.simState = nil
	p.packer = nil
	p.cnf = nil
	p.candidates = nil
	p.copies = nil
}

var cec4StatsClock struct {
	sync.Mutex
	last int64
}

func Cec4_ManPrintStats(g *Gia_Man_t, pars *Cec_ParFra_t, p *Cec4_Man_t, simulation int32) {
	if pars.fVerbose == 0 {
		return
	}
	now := cecCPUClock()
	cec4StatsClock.Lock()
	var elapsed time.Duration
	if p.nItersSim+p.nItersSat != 0 && cec4StatsClock.last != 0 {
		elapsed = time.Duration(now - cec4StatsClock.last)
	}
	cec4StatsClock.last = now
	cec4StatsClock.Unlock()
	var heads, constants, none int32
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		if gia.Gia_ObjIsHead(g, id) != 0 {
			heads++
		} else if gia.Gia_ObjIsConst(g, id) != 0 {
			constants++
		} else if gia.Gia_ObjIsNone(g, id) != 0 {
			none++
		}
	}
	literals := gia.Gia_ManObjNum(g) - heads - none
	if simulation != 0 {
		fmt.Printf("Sim %4d : %6.2f %%  ", p.nItersSim+p.nItersSat, 100*float64(literals)/float64(gia.Gia_ManCandNum(g)))
		p.nItersSim++
	} else {
		fmt.Printf("SAT %4d : %6.2f %%  ", p.nItersSim+p.nItersSat, 100*float64(p.nAndNodes)/float64(gia.Gia_ManAndNum(g)))
		p.nItersSat++
	}
	fmt.Printf("P =%7d  D =%7d  F =%8d  cst =%9d  cls =%8d  lit =%9d   Time = %6.2f sec\n", p.nSatUnsat, p.nSatSat, p.nSatUndec, constants, heads, literals, elapsed.Seconds())
}
func Cec4_ManPrintTfiConeStats(g *Gia_Man_t) {
	roots, nodes := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		if gia.Gia_ObjIsHead(g, id) == 0 || id%100 != 0 {
			continue
		}
		vec.Vec_IntClear(roots)
		vec.Vec_IntPush(roots, id)
		for member := g.Next(id); member > 0; member = g.Next(member) {
			vec.Vec_IntPush(roots, member)
		}
		gia.Gia_ManCollectTfi(g, roots, nodes)
		fmt.Printf("Class %6d : Roots = %6d  Nodes = %6d  \n", id, vec.Vec_IntSize(roots), vec.Vec_IntSize(nodes))
	}
}
func Cec4_ManPrintClasses2(g *Gia_Man_t) {
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		if gia.Gia_ObjIsHead(g, id) == 0 {
			continue
		}
		fmt.Printf("Class %d : ", id)
		for member := g.Next(id); member > 0; member = g.Next(member) {
			fmt.Printf("%d ", member)
		}
		fmt.Println()
	}
}
func Cec4_ManPrintClasses(g *Gia_Man_t) {
	count := 0
	for member := g.Next(0); member > 0; member = g.Next(member) {
		count++
	}
	fmt.Printf("Const0 class has %d entries.\n", count)
}
