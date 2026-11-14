package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Jf_ManDeriveMapping attaches the selected structural cuts to the source GIA.
func Jf_ManDeriveMapping(p *Jf_Man_t) {
	if p.pPars.FCutMin != 0 {
		panic("JF structural mapping requires unminimized cuts")
	}
	mapping := vec.Vec_IntStart(int(p.pGia.nObjs))
	for _, obj := range p.pGia.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 || Gia_ObjRefNum(p.pGia, obj) == 0 {
			continue
		}
		cut := Jf_ObjCutBest(p, obj.id)
		vec.Vec_IntWriteEntry(mapping, int(obj.id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, Jf_CutSize(cut))
		for k := int32(1); k <= Jf_CutSize(cut); k++ {
			vec.Vec_IntPush(mapping, Jf_CutVar(cut, k))
		}
		vec.Vec_IntPush(mapping, obj.id)
	}
	p.pGia.vMapping = mapping
}
func Jf_ManSetDefaultPars(p *Jf_Par_t) {
	*p = Jf_Par_t{NLutSize: 6, NCutNum: 8, NRounds: 1, NVerbLimit: 5, DelayTarget: -1, FAreaOnly: 1, FOptEdge: 1, NLutSizeMax: JF_LEAF_MAX, NCutNumMax: JF_CUT_MAX}
}
func Jf_ManPrintStats(p *Jf_Man_t, title string) {
	if p.pPars.FVerbose == 0 {
		return
	}
	fmt.Printf("%s :  Level =%6d   Area =%9d   Edge =%9d   ", title, p.pPars.Delay, p.pPars.Area, p.pPars.Edge)
	if p.pPars.FGenCnf != 0 {
		fmt.Printf("Cnf =%9d   ", p.pPars.Clause)
	}
	fmt.Printf("Time = %s\n", time.Since(p.clkStart))
}

// Jf_ManDeriveMappingGia rebuilds the mapped network from minimized cuts.
func Jf_ManDeriveMappingGia(p *Jf_Man_t) *Gia_Man_t {
	if p.pPars.FCutMin == 0 {
		panic("JF minimized mapping requires cut minimization")
	}
	g := p.pGia
	result := Gia_ManStart(g.nObjs)
	result.pName = g.pName
	result.pSpec = g.pSpec
	copies := make([]int32, g.nObjs)
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	for _, id := range vec.Vec_IntArray(g.vCis) {
		copies[id] = Gia_ManAppendCi(result)
	}
	var literals, offsets, cover *Vec_Int_t
	if p.pPars.FGenCnf != 0 {
		literals = vec.Vec_IntAlloc(1000)
		offsets = vec.Vec_IntAlloc(1000)
		cover = vec.Vec_IntAlloc(100)
		vec.Vec_IntPush(offsets, 0)
		vec.Vec_IntPush(literals, 1)
	}
	records := []int32{0}
	mapped := make(map[int32]int32)
	var scratch []uint32
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 || Gia_ObjRefNum(g, obj) == 0 {
			continue
		}
		cut := Jf_ObjCutBest(p, obj.id)
		class := Jf_CutFuncClass(cut)
		size := Jf_CutSize(cut)
		if size == 0 {
			if class != 0 {
				panic("JF invalid constant cut")
			}
			copies[obj.id] = Jf_CutFunc(cut)
			continue
		}
		if size == 1 {
			if class != 1 {
				panic("JF invalid single-leaf cut")
			}
			lit := Jf_CutLit(cut, 1) ^ Jf_CutFuncCompl(cut)
			if copies[lit>>1] < 0 {
				panic("JF missing cut leaf")
			}
			copies[obj.id] = copies[lit>>1] ^ (lit & 1)
			continue
		}
		var truth []uint64
		if p.pPars.FFuncDsd != 0 {
			if extra.Sdm_ManReadDsdVarNum(p.pDsd, class) != size {
				panic("JF DSD support mismatch")
			}
			truth = []uint64{extra.Sdm_ManReadDsdTruth(p.pDsd, class)}
		} else {
			truth = vec.Vec_MemReadEntry(p.vTtMem, int(class))
		}
		leaves := make([]int32, 0, size)
		for k := int32(1); k <= size; k++ {
			lit := Jf_CutLit(cut, k)
			if lit == 0 {
				break
			}
			if copies[lit>>1] < 0 {
				panic("JF missing cut leaf")
			}
			leaves = append(leaves, copies[lit>>1]^(lit&1))
		}
		words := make([]uint32, 2*len(truth))
		for i, w := range truth {
			words[2*i] = uint32(w)
			words[2*i+1] = uint32(w >> 32)
		}
		lit := kit.Kit_TruthToGia(result, words, int32(len(leaves)), &scratch, leaves, 0)
		if p.pPars.FGenCnf != 0 {
			vLeaves := vec.Vec_IntAllocArrayCopy(leaves, len(leaves))
			Jf_ManGenCnf(truth[0], lit, vLeaves, literals, offsets, cover)
		}
		lit ^= Jf_CutFuncCompl(cut)
		copies[obj.id] = lit
		mapped[lit>>1] = int32(len(records))
		records = append(records, int32(len(leaves)))
		for _, leaf := range leaves {
			records = append(records, leaf>>1)
		}
		records = append(records, lit>>1)
	}
	for _, id := range vec.Vec_IntArray(g.vCos) {
		obj := Gia_ManObj(g, id)
		lit := copies[Gia_ObjFanin0(obj).id]
		if lit < 0 {
			panic("JF missing output driver")
		}
		lit ^= Gia_ObjFaninC0(obj)
		out := Gia_ManAppendCo(result, lit)
		if p.pPars.FGenCnf != 0 {
			leaves := vec.Vec_IntAllocArrayCopy([]int32{lit}, 1)
			Jf_ManGenCnf(0xaaaaaaaaaaaaaaaa, out, leaves, literals, offsets, cover)
		}
	}
	mapping := vec.Vec_IntStart(int(result.nObjs))
	for id, offset := range mapped {
		vec.Vec_IntWriteEntry(mapping, int(id), result.nObjs+offset)
	}
	for _, v := range records {
		vec.Vec_IntPush(mapping, v)
	}
	result.vMapping = mapping
	Gia_ManSetRegNum(result, Gia_ManRegNum(g))
	if p.pPars.FGenCnf != 0 {
		if p.pPars.fCnfObjIds != 0 {
			result.pData = Jf_ManCreateCnf(result, literals, offsets)
		} else {
			result.pData = Jf_ManCreateCnfRemap(result, literals, offsets, p.pPars.fAddOrCla)
		}
	}
	return result
}
