package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Ssw_Cla_t_ stores classes by representative node ID.
type Ssw_Cla_t_ struct {
	pAig                           *aig.Graph
	pId2Class                      [][]aig.Edge
	pClassSizes                    []int32
	fConstCorr                     int32
	nClasses, nCands1, nLits       int32
	pMemClasses                    []aig.Edge
	pMemClassesFree                int
	vClassOld, vClassNew, vRefined *vec.Vec_Ptr_t
	pManData                       any
	pFuncNodeHash                  func(any, aig.Edge) uint32
	pFuncNodeIsConst               func(any, aig.Edge) int32
	pFuncNodesAreEqual             func(any, aig.Edge, aig.Edge) int32
}

func Ssw_ClassesStart(g *aig.Graph) *Ssw_Cla_t {
	n := len(g.Nodes())
	if !aig.Aig_ManHasReprs(g) {
		aig.Aig_ManReprStart(g, int32(n))
	}
	return &Ssw_Cla_t_{pAig: g, pId2Class: make([][]aig.Edge, n), pClassSizes: make([]int32, n), vClassOld: vec.Vec_PtrAlloc(100), vClassNew: vec.Vec_PtrAlloc(100), vRefined: vec.Vec_PtrAlloc(1000)}
}
func Ssw_ClassesSetData(p *Ssw_Cla_t, data any, hash func(any, aig.Edge) uint32, isConst func(any, aig.Edge) int32, equal func(any, aig.Edge, aig.Edge) int32) {
	p.pManData, p.pFuncNodeHash, p.pFuncNodeIsConst, p.pFuncNodesAreEqual = data, hash, isConst, equal
}
func Ssw_ClassesStop(p *Ssw_Cla_t) {
	if p != nil {
		*p = Ssw_Cla_t_{}
	}
}
func Ssw_ClassesReadAig(p *Ssw_Cla_t) *aig.Graph        { return p.pAig }
func Ssw_ClassesGetRefined(p *Ssw_Cla_t) *vec.Vec_Ptr_t { return p.vRefined }
func Ssw_ClassesClearRefined(p *Ssw_Cla_t)              { vec.Vec_PtrClear(p.vRefined) }
func Ssw_ClassesCand1Num(p *Ssw_Cla_t) int32            { return p.nCands1 }
func Ssw_ClassesClassNum(p *Ssw_Cla_t) int32            { return p.nClasses }
func Ssw_ClassesLitNum(p *Ssw_Cla_t) int32              { return p.nLits }
