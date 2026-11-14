package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

type Seg_Man_t struct {
	pSat                                                   *bsat.Solver
	nVars, LogN, Power2, FirstVar                          int32
	nBTLimit, DelayMax, nEdges, fDelay, fReverse, fVerbose int32
	pGia                                                   *Gia_Man_t
	vPolars, vToSkip, vEdges, vFirsts, vNvars, vLits       *Vec_Int_t
	pLevels                                                []int32
	timeStart                                              time.Time
}

func Seg_ManCountIntLevels(p *Seg_Man_t, first int32) int32 {
	g := p.pGia
	vec.Vec_IntFill(p.vFirsts, int(g.nObjs), -1)
	vec.Vec_IntFill(p.vNvars, int(g.nObjs), 0)
	if g.PManTime != nil {
		p.DelayMax = Gia_ManLutLevelWithBoxes(g)
		p.pLevels = append(p.pLevels[:0], vec.Vec_IntArray(g.vLevels)...)
		g.vLevels = nil
	} else {
		p.DelayMax = Gia_ManLutLevel(g, &p.pLevels)
	}
	for id := int32(1); id < g.nObjs; id++ {
		obj := Gia_ManObj(g, id)
		if Gia_ObjIsCo(obj) != 0 || (Gia_ObjIsAnd(obj) != 0 && Gia_ObjIsLut(g, id) == 0) {
			continue
		}
		if Gia_ObjIsCi(obj) == 0 && Gia_ObjIsLut(g, id) == 0 {
			panic("gia: invalid SAT edge level object")
		}
		vec.Vec_IntWriteEntry(p.vFirsts, int(id), first)
		n := p.pLevels[id]
		if n < 2 {
			n = 0
		}
		vec.Vec_IntWriteEntry(p.vNvars, int(id), n)
		first += n
	}
	return first
}
func Seg_ManAlloc(g *Gia_Man_t, fanouts int32) *Seg_Man_t {
	p := &Seg_Man_t{pGia: g, vPolars: vec.Vec_IntAlloc(1000), vToSkip: vec.Vec_IntAlloc(1000), vFirsts: vec.Vec_IntAlloc(0), vNvars: vec.Vec_IntAlloc(0), vLits: vec.Vec_IntAlloc(0), pSat: bsat.NewSolver()}
	p.vEdges = Seg_ManCountIntEdges(g, p.vPolars, p.vToSkip, fanouts)
	p.nVars = int32(vec.Vec_IntSize(p.vEdges) / 2)
	p.LogN = util.Abc_Base2Log(uint32(p.nVars))
	p.Power2 = 1 << p.LogN
	p.FirstVar = p.nVars
	all := Seg_ManCountIntLevels(p, p.FirstVar)
	for p.pSat.NumVars() < all {
		p.pSat.NewVariable()
	}
	Gia_ManFillValue(g)
	return p
}
func Seg_ManClean(p *Seg_Man_t) {
	p.timeStart = time.Now()
	p.pSat = bsat.NewSolver()
	for p.pSat.NumVars() < p.nVars {
		p.pSat.NewVariable()
	}
	for _, v := range []*Vec_Int_t{p.vEdges, p.vFirsts, p.vNvars, p.vLits, p.vPolars, p.vToSkip} {
		vec.Vec_IntClear(v)
	}
	Gia_ManFillValue(p.pGia)
}
func Seg_ManStop(p *Seg_Man_t) { p.pSat = nil; p.pLevels = nil }
