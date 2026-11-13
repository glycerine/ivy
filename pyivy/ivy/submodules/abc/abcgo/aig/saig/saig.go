// Package saig provides sequential AIG helpers.
package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// C spellings remain aliases while the wider ABC port is being migrated.
type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Vec_t = vec.Vec_Vec_t
type Vec_Wrd_t = vec.Vec_Wrd_t

type Sec_MtrStatus_t struct {
	nInputs, nNodes, nOutputs int32
	nUnsat, nSat, nUndec      int32
	iOut                      int32
}

type Saig_ParBbr_t struct {
	TimeLimit, NBddMax, NIterMax        int32
	FPartition, FReorder, FReorderImage int32
	FVerbose, FSilent, FSkipOutCheck    int32
	IFrame                              int32
}

func Saig_ManPiNum(p *aig.Graph) int32          { return p.TruePiCount() }
func Saig_ManPoNum(p *aig.Graph) int32          { return p.TruePoCount() }
func Saig_ManCiNum(p *aig.Graph) int32          { return p.TruePiCount() + p.RegisterCount() }
func Saig_ManCoNum(p *aig.Graph) int32          { return p.TruePoCount() + p.RegisterCount() }
func Saig_ManRegNum(p *aig.Graph) int32         { return aig.Aig_ManRegNum(p) }
func Saig_ManConstrNum(p *aig.Graph) int32      { return aig.Aig_ManConstrNum(p) }
func Saig_ManLo(p *aig.Graph, i int32) aig.Edge { return aig.Aig_ManCi(p, p.TruePiCount()+i) }
func Saig_ManLi(p *aig.Graph, i int32) aig.Edge { return aig.Aig_ManCo(p, p.TruePoCount()+i) }

func Saig_ObjIsPi(p *aig.Graph, e aig.Edge) bool {
	e = aig.Aig_Regular(e)
	return aig.Aig_ObjIsCi(p, e) && aig.Aig_ObjCioId(p, e) < Saig_ManPiNum(p)
}
func Saig_ObjIsPo(p *aig.Graph, e aig.Edge) bool {
	e = aig.Aig_Regular(e)
	return aig.Aig_ObjIsCo(p, e) && aig.Aig_ObjCioId(p, e) < Saig_ManPoNum(p)
}
func Saig_ObjIsLo(p *aig.Graph, e aig.Edge) bool {
	e = aig.Aig_Regular(e)
	return aig.Aig_ObjIsCi(p, e) && aig.Aig_ObjCioId(p, e) >= Saig_ManPiNum(p)
}
func Saig_ObjIsLi(p *aig.Graph, e aig.Edge) bool {
	e = aig.Aig_Regular(e)
	return aig.Aig_ObjIsCo(p, e) && aig.Aig_ObjCioId(p, e) >= Saig_ManPoNum(p)
}
func Saig_ObjLoToLi(p *aig.Graph, e aig.Edge) aig.Edge {
	if !Saig_ObjIsLo(p, e) {
		panic("saig: object is not a latch output")
	}
	return Saig_ManLi(p, Saig_ObjRegId(p, e))
}
func Saig_ObjLiToLo(p *aig.Graph, e aig.Edge) aig.Edge {
	if !Saig_ObjIsLi(p, e) {
		panic("saig: object is not a latch input")
	}
	return Saig_ManLo(p, Saig_ObjRegId(p, e))
}
func Saig_ObjRegId(p *aig.Graph, e aig.Edge) int32 {
	e = aig.Aig_Regular(e)
	if Saig_ObjIsLo(p, e) {
		return aig.Aig_ObjCioId(p, e) - Saig_ManPiNum(p)
	}
	if Saig_ObjIsLi(p, e) {
		return aig.Aig_ObjCioId(p, e) - Saig_ManPoNum(p)
	}
	panic("saig: object is not a register terminal")
}
