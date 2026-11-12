package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

type Sle_Man_t struct {
	pGia                                                        *Gia_Man_t
	nLevels, fVerbose, nSatCalls                                int32
	nNodeVars, nCutVars, nEdgeVars, nDelayVars, nVarsTotal      int32
	nCutClas, nEdgeClas, nEdgeClas2, nDelayClas                 int32
	pSat                                                        *bsat.Solver
	vMask                                                       *Vec_Bit_t
	vCuts                                                       *Vec_Int_t
	vCutFanins, vFanoutEdges, vEdgeCuts                         *Vec_Wec_t
	vObjMap, vCutFirst, vEdgeFirst, vDelayFirst, vPolars, vLits *Vec_Int_t
	timeStart                                                   time.Time
}

func Sle_ManList(p *Sle_Man_t, id int32) []int32 {
	data := vec.Vec_IntArray(p.vCuts)
	return data[data[id]:]
}
func Sle_ManAlloc(g *Gia_Man_t, levels, verbose int32) *Sle_Man_t {
	p := &Sle_Man_t{pGia: g, nLevels: levels, fVerbose: verbose}
	p.vMask = Sle_ManInternalNodeMask(g)
	p.vCuts = Sle_ManComputeCuts(g, 4, verbose)
	p.vCutFanins = Sle_ManCollectCutFanins(g, p.vCuts, p.vMask)
	p.vFanoutEdges = vec.Vec_WecStart(int(g.nObjs))
	p.vEdgeCuts = vec.Vec_WecAlloc(100)
	p.vObjMap = vec.Vec_IntStartFull(int(g.nObjs))
	p.vCutFirst = vec.Vec_IntStartFull(int(g.nObjs))
	p.vEdgeFirst = vec.Vec_IntStartFull(int(g.nObjs))
	p.vDelayFirst = vec.Vec_IntStartFull(int(g.nObjs))
	p.vPolars = vec.Vec_IntAlloc(100)
	p.vLits = vec.Vec_IntAlloc(100)
	p.nLevels = Sle_ManComputeDelay(g, p.vCuts)
	return p
}
func Sle_ManStop(p *Sle_Man_t) { p.pSat = nil }
func Sle_ManMarkupVariables(p *Sle_Man_t) {
	counter := p.pGia.nObjs
	p.nNodeVars = counter
	for id := int32(1); id < p.pGia.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, id)) != 0 {
			vec.Vec_IntWriteEntry(p.vCutFirst, int(id), counter)
			counter += Sle_ListCutNum(Sle_ManList(p, id))
		}
	}
	p.nCutVars = counter - p.nNodeVars
	for id := int32(1); id < p.pGia.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, id)) != 0 {
			vec.Vec_IntWriteEntry(p.vEdgeFirst, int(id), counter)
			counter += int32(vec.Vec_IntSize(vec.Vec_WecEntry(p.vCutFanins, int(id))))
		}
	}
	p.nEdgeVars = counter - p.nNodeVars - p.nCutVars
	for id := int32(1); id < p.pGia.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p.pGia, id)) != 0 {
			vec.Vec_IntWriteEntry(p.vDelayFirst, int(id), counter)
			counter += p.nLevels
		}
	}
	p.nDelayVars = counter - p.nNodeVars - p.nCutVars - p.nEdgeVars
	p.nVarsTotal = counter
}
func Sle_ManCheckContained(cut []int32, nCut int32, lut []int32, nLut int32) int32 {
	if nCut < 0 || nLut < 0 || int(nCut) > len(cut) || int(nLut) > len(lut) {
		panic("gia: invalid cut lengths")
	}
	if nCut > nLut {
		return 0
	}
	for _, x := range cut[:nCut] {
		found := false
		for _, y := range lut[:nLut] {
			if x == y {
				found = true
				break
			}
		}
		if !found {
			return 0
		}
	}
	return 1
}
