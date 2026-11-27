package abs

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Gla_ManStop(p *Gla_Man_t) {
	if p.pPars.FVerbose != 0 {
		util.Abc_Print(1, "SAT solver:  Var = %d  Cla = %d  Conf = %d  Lrn = %d  Reduce = %d  Cex = %d  Objs+ = %d\n",
			bsat.Sat_solver2_nvars(p.pSat), bsat.Sat_solver2_nclauses(p.pSat), bsat.Sat_solver2_nconflicts(p.pSat),
			bsat.Sat_solver2_nlearnts(p.pSat), bsat.Sat_solver2_NDBreduces(p.pSat), p.nCexes, p.nObjAdded)
	}
	// stop the refinement manager
	Rnm_ManStop(p.pRnm, 0)
	p.pvRefis = nil
	p.pCnf = nil
	if p.pGia0 != nil {
		gia.Gia_ManStop(p.pGia)
	}
	bsat.Sat_solver2_delete(p.pSat)
	p.vObjCounts, p.vAddedNew, p.vCoreCounts, p.vProofIds, p.vTemp, p.vAbs = nil, nil, nil, nil, nil, nil
	p.pObj2Obj = nil
	p.pObjs = nil
}
func Gla_ManAbsPrintFrame(p *Gla_Man_t, nCoreSize, nFrames, nConfls, nCexes int32, Time int64) {
	if abcmain.Abc_FrameIsBatchMode() != 0 && nCoreSize <= 0 {
		return
	}
	util.Abc_Print(1, "%4d :", nFrames-1)
	util.Abc_Print(1, "%4d", min(100, 100*Gia_GlaAbsCount(p, 0, 0)/(p.nObjs-gia.Gia_ManPoNum(p.pGia)+gia.Gia_ManCoNum(p.pGia)+1)))
	util.Abc_Print(1, "%6d", Gia_GlaAbsCount(p, 0, 0))
	util.Abc_Print(1, "%5d", Gla_ManCountPPis(p))
	util.Abc_Print(1, "%5d", Gia_GlaAbsCount(p, 1, 0))
	util.Abc_Print(1, "%6d", Gia_GlaAbsCount(p, 0, 1))
	util.Abc_Print(1, "%8d", nConfls)
	if nCexes == 0 {
		util.Abc_Print(1, "%5c", '-')
	} else {
		util.Abc_Print(1, "%5d", nCexes)
	}
	util.Abc_PrintInt(bsat.Sat_solver2_nvars(p.pSat))
	util.Abc_PrintInt(bsat.Sat_solver2_nclauses(p.pSat))
	util.Abc_PrintInt(bsat.Sat_solver2_nlearnts(p.pSat))
	util.Abc_Print(1, "%9.2f sec", float64(Time)/float64(util.CLOCKS_PER_SEC))
	util.Abc_Print(1, "%5.0f MB", (bsat.Sat_solver2_memory_proof(p.pSat)+bsat.Sat_solver2_memory(p.pSat, 0))/(1<<20))
	if nCoreSize > 0 && nCexes > 0 {
		util.Abc_Print(1, "%s", "\n")
	} else {
		util.Abc_Print(1, "%s", "\r")
	}
}
func Gla_ManReportMemory(p *Gla_Man_t) {
	memAig := float64(gia.Gia_ManObjNum(p.pGia)) * 12
	memSat := bsat.Sat_solver2_memory(p.pSat, 1)
	memPro := bsat.Sat_solver2_memory_proof(p.pSat)
	memMap := float64(p.nObjs)*float64(unsafe.Sizeof(Gla_Obj_t{})) + float64(gia.Gia_ManObjNum(p.pGia))*4
	memRef := Rnm_ManMemoryUsage(p.pRnm)
	memOth := float64(unsafe.Sizeof(*p))
	for i := range p.pObjs {
		memMap += float64(vec.Vec_IntCap(&p.pObjs[i].vFrames)) * 4
	}
	memOth += float64(vec.Vec_IntCap(p.vAddedNew)) * 4
	memOth += float64(vec.Vec_IntCap(p.vTemp)) * 4
	memOth += float64(vec.Vec_IntCap(p.vAbs)) * 4
	memTot := memAig + memSat + memPro + memMap + memRef + memOth
	abcPRMP("Memory: AIG      ", memAig, memTot)
	abcPRMP("Memory: SAT      ", memSat, memTot)
	abcPRMP("Memory: Proof    ", memPro, memTot)
	abcPRMP("Memory: Map      ", memMap, memTot)
	abcPRMP("Memory: Refine   ", memRef, memTot)
	abcPRMP("Memory: Other    ", memOth, memTot)
	abcPRMP("Memory: TOTAL    ", memTot, memTot)
}
func Gia_GlaSendAbsracted(p *Gla_Man_t, verbose int32) {
	_ = verbose
	if abcmain.Abc_FrameIsBridgeMode() == 0 {
		panic("old GLA bridge mode is not active")
	}
	abstract := Gia_ManDupAbsGates(p.pGia0, Gla_ManTranslate(p))
	defer gia.Gia_ManStop(abstract)
	net, err := gia.Gia_ManToBridgeNetlist(abstract)
	if err != nil {
		panic(err)
	}
	util.Gia_ManToBridgeAbsNetlist(os.Stdout, net, 107)
}
func Gia_GlaSendCancel(p *Gla_Man_t, verbose int32) {
	_, _ = p, verbose
	if abcmain.Abc_FrameIsBridgeMode() == 0 {
		panic("old GLA bridge mode is not active")
	}
	util.Gia_ManToBridgeBadAbs(os.Stdout)
}
func Gia_GlaDumpAbsracted(p *Gla_Man_t, verbose int32) {
	name := p.pPars.PFileVabs
	if name == "" {
		name = "glabs.aig"
	}
	if verbose != 0 {
		fmt.Printf("Dumping abstracted model into file \"%s\"...\n", name)
	}
	abstract := Gia_ManDupAbsGates(p.pGia0, Gla_ManTranslate(p))
	defer gia.Gia_ManStop(abstract)
	if err := gia.Gia_AigerWrite(abstract, name, 0, 0, 0); err != nil {
		fmt.Fprintf(os.Stderr, "old GLA: %v\n", err)
	}
}
