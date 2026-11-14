package gia

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
)

type Jf_Man_t struct {
	pGia              *Gia_Man_t
	pPars             *Jf_Par_t
	pDsd              *extra.Sdm_Man_t
	vCnfs             *Vec_Int_t
	vTtMem            *vec.Vec_Mem_t
	vCuts, vArr, vDep Vec_Int_t
	vFlow, vRefs      Vec_Flt_t
	pMem              jfCutMemory
	vTemp             *Vec_Int_t
	pCutCmp           func(*Jf_Cut_t, *Jf_Cut_t) float32
	clkStart          time.Time
	CutCount          [4]word
	nCoarse           int32
}

func Jf_ObjCutH(p *Jf_Man_t, i int32) int32      { return vec.Vec_IntEntry(&p.vCuts, int(i)) }
func Jf_ObjCuts(p *Jf_Man_t, i int32) []int32    { return p.pMem.get(Jf_ObjCutH(p, i)) }
func Jf_ObjCutBest(p *Jf_Man_t, i int32) []int32 { return Jf_ObjCuts(p, i)[1:] }
func Jf_ObjArr(p *Jf_Man_t, i int32) int32       { return vec.Vec_IntEntry(&p.vArr, int(i)) }
func Jf_ObjDep(p *Jf_Man_t, i int32) int32       { return vec.Vec_IntEntry(&p.vDep, int(i)) }
func Jf_ObjFlow(p *Jf_Man_t, i int32) float32    { return vec.Vec_FltEntry(&p.vFlow, int(i)) }
func Jf_ObjRefs(p *Jf_Man_t, i int32) float32    { return vec.Vec_FltEntry(&p.vRefs, int(i)) }
func Jf_ObjLit(i, c int32) int32                 { return i<<1 | c }
func Jf_CutCnfSizeF(p *Jf_Man_t, f int32) int32  { return vec.Vec_IntEntry(p.vCnfs, int(f)) }
func Jf_CutCnfSize(p *Jf_Man_t, c []int32) int32 { return Jf_CutCnfSizeF(p, Jf_CutFuncClass(c)) }
func Jf_ObjFunc0(p *Gia_Obj_t, c []int32) int32  { return Jf_CutFunc(c) ^ Gia_ObjFaninC0(p) }
func Jf_ObjFunc1(p *Gia_Obj_t, c []int32) int32  { return Jf_CutFunc(c) ^ Gia_ObjFaninC1(p) }
func Jf_ManInitRefs(p *Jf_Man_t) []float32 {
	refs, coarse := jfInitRefs(p.pGia, p.pPars.FCoarsen != 0)
	p.nCoarse = coarse
	return refs
}
func Jf_ManAlloc(g *Gia_Man_t, pars *Jf_Par_t) *Jf_Man_t {
	if pars.NLutSize > JF_LEAF_MAX || pars.NCutNum > JF_CUT_MAX || pars.NLutSize < 0 || pars.NCutNum < 0 {
		panic("JF invalid mapping limits")
	}
	g.vMapping = nil
	p := &Jf_Man_t{pGia: g, pPars: pars, clkStart: time.Now()}
	if pars.FCutMin != 0 && pars.FFuncDsd == 0 {
		p.vTtMem = vec.Vec_MemAllocForTT(int(pars.NLutSize), 0)
	}
	if pars.FCutMin != 0 && pars.FFuncDsd != 0 {
		p.pDsd = extra.Sdm_ManRead()
		if pars.FGenCnf != 0 {
			p.vCnfs = vec.Vec_IntStart(595)
			extra.Sdm_ManReadCnfCosts(p.pDsd, vec.Vec_IntArray(p.vCnfs), 595)
		}
	}
	n := int(Gia_ManObjNum(g))
	vec.Vec_IntFill(&p.vCuts, n, 0)
	vec.Vec_IntFill(&p.vArr, n, 0)
	vec.Vec_IntFill(&p.vDep, n, 0)
	vec.Vec_FltFill(&p.vFlow, n, 0)
	*vec.Vec_FltArrayP(&p.vRefs) = Jf_ManInitRefs(p)
	p.vTemp = vec.Vec_IntAlloc(1000)
	return p
}
func Jf_ManProfileClasses(p *Jf_Man_t) {
	var counts, costs [595]int32
	var total, costTotal int32
	fmt.Printf("DSD classes that appear in more than %.1f %% of mapped nodes:\n", 0.1*float64(p.pPars.NVerbLimit))
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 || Gia_ObjRefNumId(p.pGia, obj.id) == 0 {
			continue
		}
		class := Jf_CutFuncClass(Jf_ObjCutBest(p, obj.id))
		if class < 0 || int(class) >= len(counts) {
			panic("JF invalid DSD class")
		}
		if p.pPars.FGenCnf != 0 {
			cost := Jf_CutCnfSizeF(p, class)
			costs[class] += cost
			costTotal += cost
		}
		counts[class]++
		total++
	}
	total = max(total, 1)
	costTotal = max(costTotal, 1)
	var other, otherCost int32
	for i, count := range counts {
		if count != 0 && 100*float64(count)/float64(total) >= 0.1*float64(p.pPars.NVerbLimit) {
			fmt.Printf("%5d  :  %-20s   %8d  %5.1f %%   %8d  %5.1f %%\n", i, extra.Sdm_ManReadDsdStr(p.pDsd, int32(i)), count, 100*float64(count)/float64(total), costs[i], 100*float64(costs[i])/float64(costTotal))
		} else {
			other += count
			otherCost += costs[i]
		}
	}
	fmt.Printf("Other  :  %-20s   %8d  %5.1f %%   %8d  %5.1f %%\n", "", other, 100*float64(other)/float64(total), otherCost, 100*float64(otherCost)/float64(costTotal))
}
func Jf_ManFree(p *Jf_Man_t) {
	if p == nil {
		return
	}
	if p.pPars.FVerbose != 0 && p.pDsd != nil {
		extra.Sdm_ManPrintDsdStats(p.pDsd, 0)
	}
	if p.pPars.FVerbose != 0 && p.vTtMem != nil {
		fmt.Printf("Unique truth tables = %d. Memory = %.2f MB   Time = %s\n", vec.Vec_MemEntryNum(p.vTtMem), vec.Vec_MemMemory(p.vTtMem)/(1<<20), time.Since(p.clkStart))
	}
	if p.pPars.FVeryVerbose != 0 && p.pPars.FCutMin != 0 && p.pPars.FFuncDsd != 0 {
		Jf_ManProfileClasses(p)
	}
	if p.pPars.FCoarsen != 0 {
		Gia_ManCleanMark0(p.pGia)
	}
	p.pGia.Refs = nil
	if p.vTtMem != nil {
		vec.Vec_MemHashFree(p.vTtMem)
		vec.Vec_MemFree(p.vTtMem)
	}
	p.vTtMem = nil
	p.vCnfs = nil
	p.vTemp = nil
	p.pMem.reset()
	p.vCuts = Vec_Int_t{}
	p.vArr = Vec_Int_t{}
	p.vDep = Vec_Int_t{}
	p.vFlow = Vec_Flt_t{}
	p.vRefs = Vec_Flt_t{}
}
