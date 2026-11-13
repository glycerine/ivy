package abs

import (
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gla_ManStop(p *Gla_Man_t) {
	if p == nil {
		return
	}
	if p.pPars.fVerbose != 0 && p.pSat != nil {
		s := p.pSat.Statistics()
		fmt.Printf("SAT solver: Var = %d Cla = %d Conf = %d Lrn = %d Reduce = %d Cex = %d Objs+ = %d\n", s.Variables, s.Clauses, s.Conflicts, s.Learned, p.pSat.Reductions(), p.nCexes, p.nObjAdded)
	}
	Rnm_ManStop(p.pRnm, 0)
	if p.pGia0 != nil {
		gia.Gia_ManStop(p.pGia)
	}
	p.pObjs = nil
	p.pObj2Obj = nil
	p.pvRefis = nil
	p.pSat = nil
	p.pCnf = nil
	p.vAbs, p.vTemp, p.vAddedNew, p.vObjCounts, p.vCoreCounts, p.vProofIds = nil, nil, nil, nil, nil, nil
}
func Gla_ManAbsPrintFrame(p *Gla_Man_t, coreSize, frames, conflicts, cexes int32, elapsed time.Duration) {
	if abcmain.Abc_FrameIsBatchMode() != 0 && coreSize <= 0 {
		return
	}
	pct := min(int32(100), 100*Gia_GlaAbsCount(p, 0, 0)/(p.nObjs-gia.Gia_ManPoNum(p.pGia)+gia.Gia_ManCoNum(p.pGia)+1))
	s := p.pSat.Statistics()
	memory, proof := p.pSat.Memory(false)
	fmt.Printf("%4d :%4d%6d%5d%5d%6d%8d", frames-1, pct, Gia_GlaAbsCount(p, 0, 0), Gla_ManCountPPis(p), Gia_GlaAbsCount(p, 1, 0), Gia_GlaAbsCount(p, 0, 1), conflicts)
	if cexes == 0 {
		fmt.Printf("%5s", "-")
	} else {
		fmt.Printf("%5d", cexes)
	}
	fmt.Printf(" %d %d %d %9.2f sec%5.0f MB", s.Variables, s.Clauses, s.Learned, elapsed.Seconds(), (memory+proof)/(1<<20))
	if coreSize > 0 && cexes > 0 {
		fmt.Println()
	} else {
		fmt.Print("\r")
	}
}
func Gla_ManReportMemory(p *Gla_Man_t) {
	aigMem := gia.Gia_ManMemory(p.pGia)
	satMem, proofMem := p.pSat.Memory(true)
	mapMem := float64(len(p.pObjs))*float64(unsafe.Sizeof(Gla_Obj_t{})) + float64(len(p.pObj2Obj))*4
	for i := range p.pObjs {
		mapMem += float64(vec.Vec_IntCap(&p.pObjs[i].vFrames)) * 4
	}
	refineMem := Rnm_ManMemoryUsage(p.pRnm)
	other := float64(unsafe.Sizeof(*p))
	for _, v := range []*vec.Vec_Int_t{p.vAddedNew, p.vTemp, p.vAbs} {
		if v != nil {
			other += float64(vec.Vec_IntCap(v)) * 4
		}
	}
	total := aigMem + satMem + proofMem + mapMem + refineMem + other
	for _, x := range []struct {
		name  string
		value float64
	}{{"AIG", aigMem}, {"SAT", satMem}, {"Proof", proofMem}, {"Map", mapMem}, {"Refine", refineMem}, {"Other", other}, {"TOTAL", total}} {
		pct := float64(0)
		if total != 0 {
			pct = 100 * x.value / total
		}
		fmt.Printf("Memory: %-9s = %10.3f MB (%6.2f %%)\n", x.name, x.value/(1<<20), pct)
	}
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
	name := p.pPars.pFileVabs
	if name == "" {
		name = "glabs.aig"
	}
	if verbose != 0 {
		fmt.Printf("Dumping abstracted model into file %q...\n", name)
	}
	abstract := Gia_ManDupAbsGates(p.pGia0, Gla_ManTranslate(p))
	defer gia.Gia_ManStop(abstract)
	if err := gia.Gia_AigerWrite(abstract, name, 0, 0, 0); err != nil {
		fmt.Fprintf(os.Stderr, "old GLA: %v\n", err)
	}
}
