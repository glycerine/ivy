package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"os"
	"unsafe"
)

func Ga2_ManComputeTest(p *gia.Gia_Man_t) {
	start := util.Abc_Clock()
	Ga2_ManMarkup(p, 5, 0)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-start)
	var count int
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) != 0 && gia.Gia_ObjPhase(o) != 0 {
			count++
		}
	}
	util.Abc_Print(1, "Marked AND nodes = %6d.  ", count)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-start)
}
func Ga2_ManDumpStats(g *gia.Gia_Man_t, pars *Abs_Par_t, pSat *bsat.Sat_solver2, iFrame, fUseN int32) {
	name := "stats_gla"
	if fUseN != 0 {
		name += "n"
	}
	if pars.FUseFullProof != 0 {
		name += "p"
	}
	name += ".txt"
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s pi=%d ff=%d and=%d mem=%d bmc=%d", gia.Gia_ManName(g), gia.Gia_ManPiNum(g), gia.Gia_ManRegNum(g), gia.Gia_ManAndNum(g), int(1+bsat.Sat_solver2_memory_proof(pSat)/(1<<20)), iFrame)
	if classes := gia.Gia_ManGateClasses(g); classes != nil {
		fmt.Fprintf(f, " ff=%d and=%d", Gia_GlaCountFlops(g, classes), Gia_GlaCountNodes(g, classes))
	}
	fmt.Fprintf(f, "\n")
}

// abcPRMP is abc_global.h's ABC_PRMP.
func abcPRMP(a string, f, F float64) {
	util.Abc_Print(1, "%s =", a)
	percent := 0.0
	if F != 0 {
		percent = 100.0 * f / F
	}
	util.Abc_Print(1, "%10.3f MB (%6.2f %%)\n", f/(1<<20), percent)
}

func Ga2_ManReportMemory(p *Ga2_Man_t) {
	memAig := gia.Gia_ManMemory(p.pGia) + vec.Vec_IntMemory(gia.Gia_ManMapping(p.pGia))
	memSat := bsat.Sat_solver2_memory(p.pSat, 1)
	memPro := bsat.Sat_solver2_memory_proof(p.pSat)
	var memMap, memOth float64
	for _, v := range p.vId2Lit {
		memMap += vec.Vec_IntMemory(v)
	}
	memRef := Rnm_ManMemoryUsage(p.pRnm)
	memHash := float64(4 * 6 * int(p.nTable))
	memOth = float64(unsafe.Sizeof(*p))
	for _, v := range p.vCnfs {
		if v != nil {
			memOth += vec.Vec_IntMemory(v)
		}
	}
	memOth += vec.Vec_IntMemory(p.vIds)
	memOth += float64(4 * cap(p.vProofIds))
	memOth += vec.Vec_IntMemory(p.vAbs)
	memOth += vec.Vec_IntMemory(p.vValues)
	memOth += vec.Vec_IntMemory(p.vLits)
	memOth += vec.Vec_IntMemory(p.vIsopMem)
	memOth += 336450 + float64(1+unsafe.Sizeof(uintptr(0)))*65536
	memTot := memAig + memSat + memPro + memMap + memRef + memHash + memOth
	abcPRMP("Memory: AIG      ", memAig, memTot)
	abcPRMP("Memory: SAT      ", memSat, memTot)
	abcPRMP("Memory: Proof    ", memPro, memTot)
	abcPRMP("Memory: Map      ", memMap, memTot)
	abcPRMP("Memory: Refine   ", memRef, memTot)
	abcPRMP("Memory: Hash     ", memHash, memTot)
	abcPRMP("Memory: Other    ", memOth, memTot)
	abcPRMP("Memory: TOTAL    ", memTot, memTot)
}
