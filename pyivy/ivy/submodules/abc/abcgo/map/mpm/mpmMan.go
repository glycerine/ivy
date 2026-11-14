package mpm

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"time"
)

// Mpm_ManStart initializes mapping storage and the selected truth or DSD mode.
func Mpm_ManStart(mig *Mig_Man_t, pars *Mpm_Par_t) *Mpm_Man_t {
	assert(pars.nNumCuts <= MPM_CUT_MAX, "cut count")
	assert(pars.fUseTruth == 0 || pars.pLib.LutMax <= 16, "truth LUT size")
	assert(pars.fUseDsd == 0 || pars.pLib.LutMax <= 6, "DSD LUT size")
	Mig_ManSetRefs(mig)
	p := &Mpm_Man_t{pMig: mig, pPars: pars, pLibLut: pars.pLib, nLutSize: pars.pLib.LutMax, nNumCuts: pars.nNumCuts}
	if pars.fUseTruth != 0 {
		p.nTruWords = util.Abc_Truth6WordNum(p.nLutSize)
	}
	assert(Mpm_CutWordNum(32) < 32, "cut size bits")
	p.pManCuts = mem.Mmr_StepStart(13, util.Abc_Base2Log(uint32(Mpm_CutWordNum(p.nLutSize)+1)))
	vec.Vec_PtrGrow(&p.vFreeUnits, int(p.nNumCuts)+1)
	for i := p.nNumCuts; i >= 0; i-- {
		p.pCutUnits[i].pCut.pLeaves = make([]int32, MPM_VAR_MAX)
		vec.Vec_PtrPush(&p.vFreeUnits, &p.pCutUnits[i])
	}
	p.vTemp = vec.Vec_PtrAlloc(1000)
	n := int(Mig_ManObjNum(mig))
	for _, v := range []*vec.Vec_Int_t{&p.vCutBests, &p.vCutLists, &p.vMigRefs, &p.vMapRefs, &p.vEstRefs, &p.vTimes, &p.vAreas, &p.vEdges} {
		vec.Vec_IntFill(v, n, 0)
	}
	vec.Vec_IntFill(&p.vRequireds, n, util.ABC_INFINITY)
	assert(pars.fUseTruth == 0 || pars.fUseDsd == 0, "truth and DSD are exclusive")
	if pars.fUseTruth != 0 {
		p.vTtMem = vec.Vec_MemAlloc(int(p.nTruWords), 12)
		vec.Vec_MemHashAlloc(p.vTtMem, 10000)
		p.funcCst0 = int32(vec.Vec_MemHashInsert(p.vTtMem, p.Truth[:]))
		util.Abc_TtUnit(p.Truth[:], p.nTruWords, 0)
		p.funcVar0 = int32(vec.Vec_MemHashInsert(p.vTtMem, p.Truth[:]))
	} else if pars.fUseDsd != 0 {
		Mpm_ManPrecomputePerms(p)
		p.funcVar0 = 1
	}
	p.timeTotal = util.Abc_Clock()
	mig.pMan = p
	return p
}
func Mpm_ManStop(p *Mpm_Man_t) {
	if p.pPars.fUseTruth != 0 && p.pPars.fVeryVerbose != 0 {
		file, err := os.Create("truths.txt")
		if err != nil {
			panic(err)
		}
		vec.Vec_MemDump(file, p.vTtMem)
		if err = file.Close(); err != nil {
			panic(err)
		}
		count := vec.Vec_MemEntryNum(p.vTtMem)
		fmt.Printf("Dumped %d %d-var truth tables into file \"truths.txt\" (%.2f MB).\n", count, p.nLutSize, (16.0*float64(p.nTruWords)+1)*float64(count)/(1<<20))
	}
	if p.pPars.fUseDsd != 0 && p.pPars.fVerbose != 0 {
		Mpm_ManPrintDsdStats(p)
	}
	if p.vTtMem != nil {
		vec.Vec_MemHashFree(p.vTtMem)
		vec.Vec_MemFree(p.vTtMem)
	}
	if p.pHash != nil {
		vec.Vec_WrdFree(p.vPerm6)
		vec.Vec_IntFree(p.vMap2Perm)
		vec.Vec_IntFree(p.vConfgRes)
		vec.Hsh_IntManStop(p.pHash)
	}
	vec.Vec_WecFreeP(&p.vNpnConfigs)
	vec.Vec_PtrFree(p.vTemp)
	mem.Mmr_StepStop(p.pManCuts)
	*p = Mpm_Man_t{}
}
func Mpm_ManPrintStatsInit(p *Mpm_Man_t) {
	fmt.Printf("K = %d.  C = %d.  Cand = %d. XOR = %d. MUX = %d. Choice = %d.  CutMin = %d. Truth = %d. DSD = %d.\n", p.nLutSize, p.nNumCuts, Mig_ManCandNum(p.pMig), Mig_ManXorNum(p.pMig), Mig_ManMuxNum(p.pMig), p.pMig.nChoices, p.pPars.fCutMin, p.pPars.fUseTruth, p.pPars.fUseDsd)
}
func Mpm_ManPrintStats(p *Mpm_Man_t) {
	mig := float64(Mig_ManObjNum(p.pMig)) * 16 / (1 << 20)
	mapping := float64(Mig_ManObjNum(p.pMig)) * 48 / (1 << 20)
	cuts := float64(mem.Mmr_StepMemory(p.pManCuts)) / (1 << 17)
	fmt.Printf("Memory usage:  Mig = %.2f MB  Map = %.2f MB  Cut = %.2f MB    Total = %.2f MB.  ", mig, mapping, cuts, mig+mapping+cuts)
	if p.timeDerive == 0 {
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-p.timeTotal)
		return
	}
	fmt.Println()
	p.timeTotal = util.Abc_Clock() - p.timeTotal
	p.timeOther = p.timeTotal - p.timeDerive
	fmt.Println("Runtime breakdown:")
	for _, r := range []struct {
		name string
		t    int64
	}{{"Complete cut computation   ", p.timeDerive}, {"- Merging cuts             ", p.timeMerge}, {"- Evaluting cut parameters ", p.timeEval}, {"- Checking cut containment ", p.timeCompare}, {"- Adding cuts to storage   ", p.timeStore}, {"Other                      ", p.timeOther}, {"TOTAL                      ", p.timeTotal}} {
		var percent float64
		if p.timeTotal != 0 {
			percent = 100 * float64(r.t) / float64(p.timeTotal)
		}
		fmt.Printf("%s =%9.2f sec (%6.2f %%)\n", r.name, float64(r.t)/float64(time.Second), percent)
	}
}
