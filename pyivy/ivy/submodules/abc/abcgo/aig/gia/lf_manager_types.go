package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

type Lf_Man_t struct {
	pGia                                                                             *Gia_Man_t
	pPars                                                                            *Jf_Par_t
	nCutWords, nSetWords                                                             int32
	pObjBests                                                                        []Lf_Bst_t
	vMemSets                                                                         Vec_Ptr_t
	vFreeSets                                                                        Vec_Int_t
	vTtMem                                                                           *vec.Vec_Mem_t
	vFreePages                                                                       Vec_Ptr_t
	vStoreOld, vStoreNew                                                             Lf_Mem_t
	vOffsets, vRequired, vCutSets                                                    Vec_Int_t
	vFlowRefs                                                                        Vec_Flt_t
	vMapRefs                                                                         Vec_Int_t
	vSwitches                                                                        Vec_Flt_t
	vCiArrivals                                                                      Vec_Int_t
	clkStart                                                                         time.Time
	CutCount                                                                         [4]float64
	Switches                                                                         float64
	nFrontMax, nCoDrivers, nInverters, nTimeFails, Iter, fUseEla, nCutMux, nCutEqual int32
	nCutCounts                                                                       [LF_LEAF_MAX + 1]int32
}

func Lf_CutTruth(p *Lf_Man_t, c *Lf_Cut_t) []word {
	return vec.Vec_MemReadEntry(p.vTtMem, int(c.iFunc>>1))
}
func Lf_ObjOff(p *Lf_Man_t, i int32) int32          { return vec.Vec_IntEntry(&p.vOffsets, int(i)) }
func Lf_ObjRequired(p *Lf_Man_t, i int32) int32     { return vec.Vec_IntEntry(&p.vRequired, int(i)) }
func Lf_ObjSetRequired(p *Lf_Man_t, i, t int32)     { vec.Vec_IntDowndateEntry(&p.vRequired, int(i), t) }
func Lf_ObjReadBest(p *Lf_Man_t, i int32) *Lf_Bst_t { return &p.pObjBests[Lf_ObjOff(p, i)] }
func Lf_ObjFlowRefs(p *Lf_Man_t, i int32) float32 {
	return vec.Vec_FltEntry(&p.vFlowRefs, int(Lf_ObjOff(p, i)))
}
func Lf_ObjMapRefNum(p *Lf_Man_t, i int32) int32 {
	return vec.Vec_IntEntry(&p.vMapRefs, int(Lf_ObjOff(p, i)))
}
func Lf_ObjMapRefInc(p *Lf_Man_t, i int32) int32 {
	ptr := vec.Vec_IntEntryP(&p.vMapRefs, int(Lf_ObjOff(p, i)))
	old := *ptr
	*ptr++
	return old
}
func Lf_ObjMapRefDec(p *Lf_Man_t, i int32) int32 {
	ptr := vec.Vec_IntEntryP(&p.vMapRefs, int(Lf_ObjOff(p, i)))
	*ptr--
	return *ptr
}
func Lf_ObjSwitches(p *Lf_Man_t, i int32) float32 { return vec.Vec_FltEntry(&p.vSwitches, int(i)) }
