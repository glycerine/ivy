package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ssw_ManCountEquivs counts nodes with any representative, including const 1.
func Ssw_ManCountEquivs(p *Ssw_Man_t) int32 {
	var count int32
	for _, obj := range p.pAig.LiveEdges() {
		if aig.Aig_ObjRepr(p.pAig, obj) != aig.InvalidEdge {
			count++
		}
	}
	return count
}

// Ssw_ManCreate sets up the source AIG and the per-frame simulation state.
func Ssw_ManCreate(g *aig.Graph, pars *Ssw_Pars_t) *Ssw_Man_t {
	if g.RegisterCount() <= 0 {
		panic("ssw: sequential AIG required")
	}
	aig.Aig_ManFanoutStart(g)
	aig.Aig_ManSetCioIds(g)
	nFrames := pars.nFramesK + 1
	p := &Ssw_Man_t{pAig: g, pPars: pars, nFrames: nFrames, iOutputLit: -1,
		pNodeToFrames: make([]aig.Edge, len(g.Nodes())*int(nFrames)),
		vCommon:       vec.Vec_PtrAlloc(100), vNewLos: vec.Vec_PtrAlloc(100), vNewPos: vec.Vec_IntAlloc(100),
		vResimConsts: vec.Vec_PtrAlloc(100), vResimClasses: vec.Vec_PtrAlloc(100)}
	for i := range p.pNodeToFrames {
		p.pNodeToFrames[i] = aig.InvalidEdge
	}
	p.nPatWords = util.Abc_BitWordNum(saig.Saig_ManPiNum(g)*nFrames + g.RegisterCount())
	p.pPatWords = make([]uint32, p.nPatWords)
	return p
}
