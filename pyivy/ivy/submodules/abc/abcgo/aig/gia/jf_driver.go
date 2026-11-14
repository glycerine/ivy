package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
	"time"
)

// Jf_ManDeriveGia reconstructs an unmapped AIG from the selected cuts.
func Jf_ManDeriveGia(p *Jf_Man_t) *Gia_Man_t {
	g := p.pGia
	dst := Gia_ManStart(g.nObjs)
	dst.pName = g.pName
	dst.pSpec = g.pSpec
	copies := make([]int32, g.nObjs)
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	for _, id := range vec.Vec_IntArray(g.vCis) {
		copies[id] = Gia_ManAppendCi(dst)
	}
	Gia_ManHashStart(dst)
	var scratch []uint32
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 || Gia_ObjRefNum(g, obj) == 0 {
			continue
		}
		cut := Jf_ObjCutBest(p, obj.id)
		size := Jf_CutSize(cut)
		class := Jf_CutFuncClass(cut)
		if p.pPars.FCutMin != 0 && size == 0 {
			if class != 0 {
				panic("JF invalid constant cut")
			}
			copies[obj.id] = Jf_CutFunc(cut)
			continue
		}
		if p.pPars.FCutMin != 0 && size == 1 {
			if class != 1 {
				panic("JF invalid single-leaf cut")
			}
			lit := Jf_CutLit(cut, 1) ^ Jf_CutFuncCompl(cut)
			copies[obj.id] = copies[lit>>1] ^ (lit & 1)
			continue
		}
		ids, leaves := make([]int32, 0, size), make([]int32, 0, size)
		for k := int32(1); k <= size; k++ {
			lit := Jf_CutLit(cut, k)
			if lit == 0 {
				break
			}
			ids = append(ids, lit>>1)
			if copies[lit>>1] < 0 {
				panic("JF missing cut leaf")
			}
			leaves = append(leaves, copies[lit>>1]^(lit&1))
		}
		var truth []uint64
		if p.pPars.FCutMin == 0 {
			truth = giaNewTruthEval(g, p.pPars.NLutSize, ids).node(obj.id)
		} else if p.pPars.FFuncDsd != 0 {
			if extra.Sdm_ManReadDsdVarNum(p.pDsd, class) != size {
				panic("JF DSD support mismatch")
			}
			truth = []uint64{extra.Sdm_ManReadDsdTruth(p.pDsd, class)}
		} else {
			truth = vec.Vec_MemReadEntry(p.vTtMem, int(class))
		}
		words := make([]uint32, 2*len(truth))
		for i, w := range truth {
			words[2*i], words[2*i+1] = uint32(w), uint32(w>>32)
		}
		lit := kit.Kit_TruthToGia(dst, words, int32(len(leaves)), &scratch, leaves, 1)
		if p.pPars.FCutMin != 0 {
			lit ^= Jf_CutFuncCompl(cut)
		}
		copies[obj.id] = lit
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		obj := Gia_ManObj(g, id)
		lit := copies[Gia_ObjFanin0(obj).id]
		if lit < 0 {
			panic("JF missing output driver")
		}
		Gia_ManAppendCo(dst, lit^Gia_ObjFaninC0(obj))
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(g))
	if p.pPars.FCutMin == 0 {
		dst = Gia_ManCleanup(dst)
	}
	return dst
}
func Jf_ManPerformMapping(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if Gia_ManBufNum(g) != 0 {
		panic("JF mapping does not support buffers")
	}
	if pars.FCutMin != 0 && pars.FFuncDsd != 0 && pars.NLutSize > 6 {
		panic("JF DSD supports at most six leaves")
	}
	if pars.FGenCnf != 0 {
		pars.FCutMin = 1
		pars.FFuncDsd = 1
		pars.FOptEdge = 0
	}
	if pars.FCutMin != 0 && pars.FFuncDsd == 0 {
		pars.FCoarsen = 0
	}
	p := Jf_ManAlloc(g, pars)
	defer Jf_ManFree(p)
	if pars.FAreaOnly != 0 {
		p.pCutCmp = Jf_CutCompareArea
	} else {
		p.pCutCmp = Jf_CutCompareDelay
	}
	Jf_ManComputeCuts(p, 0)
	Jf_ManComputeRefs(p)
	Jf_ManPrintStats(p, "Start")
	for i := int32(0); i < pars.NRounds; i++ {
		if pars.FGenCnf == 0 {
			Jf_ManPropagateFlow(p, pars.FOptEdge)
			Jf_ManPrintStats(p, "Flow ")
		}
		Jf_ManPropagateEla(p, 0)
		Jf_ManPrintStats(p, "Area ")
		Jf_ManPropagateEla(p, 1)
		Jf_ManPrintStats(p, "Edge ")
	}
	if pars.FVeryVerbose != 0 && pars.FCutMin != 0 && pars.FFuncDsd == 0 {
		vec.Vec_MemDumpTruthTables(p.vTtMem, g.pName, int(pars.NLutSize))
	}
	if pars.FPureAig != 0 {
		return Jf_ManDeriveGia(p)
	}
	if pars.FCutMin != 0 {
		return Jf_ManDeriveMappingGia(p)
	}
	Jf_ManDeriveMapping(p)
	return g
}
func Jf_ManDeriveCnf(p *Gia_Man_t, objIDs int32) *Gia_Man_t {
	var pars Jf_Par_t
	Jf_ManSetDefaultPars(&pars)
	pars.FGenCnf = 1
	pars.fCnfObjIds = objIDs
	return Jf_ManPerformMapping(p, &pars)
}
func Jf_ManDeriveCnfMiter(p *Gia_Man_t, verbose int32) *Gia_Man_t {
	var pars Jf_Par_t
	Jf_ManSetDefaultPars(&pars)
	pars.FGenCnf = 1
	pars.fAddOrCla = 1
	pars.FVerbose = verbose
	return Jf_ManPerformMapping(p, &pars)
}
func Jf_ManDumpCnf(p *Gia_Man_t, filename string, verbose int32) {
	start := time.Now()
	result := Jf_ManDeriveCnfMiter(p, verbose)
	data := result.pData.(*cnf.CNF)
	result.pData = nil
	cnf.Cnf_DataWriteIntoFile(data, filename, 0, nil, nil)
	count := 0
	for _, clause := range data.Clauses {
		count += len(clause)
	}
	fmt.Printf("CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d. Time = %s\n", data.Variables, len(data.Clauses), count, time.Since(start))
	cnf.Cnf_DataFree(data)
	Gia_ManStop(result)
}
func Jf_ManTestCnf(p *Gia_Man_t) {
	result := Jf_ManDeriveCnf(p, 1)
	data := result.pData.(*cnf.CNF)
	result.pData = nil
	cnf.Cnf_DataWriteIntoFile(data, "test.cnf", 0, nil, nil)
	for i := 0; i < data.Variables; i++ {
		fmt.Printf("%d : %d %d\n", i, data.Obj2Count[i], data.Obj2Clause[i])
	}
	Gia_ManStop(result)
	cnf.Cnf_DataFree(data)
}
