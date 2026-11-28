package abc

// Port of src/opt/sfm/sfmMit.c: temporary place holders (Sfm_MitStart
// returns NULL, so mfs3 always uses Sfm_Tim_t).

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
)

type Sfm_Mit_t struct {
	Value1 int32
	Value2 int32
	Value3 int32
	Value4 int32
}

func Sfm_MitStart(pLib *mio.Mio_Library_t, pScl any, pExt *scl.Scl_Con_t, pNtk *Network, DeltaCrit int32) *Sfm_Mit_t {
	return nil
}
func Sfm_MitStop(p *Sfm_Mit_t)                                       {}
func Sfm_MitReadNtkDelay(p *Sfm_Mit_t) int32                         { return 0 }
func Sfm_MitReadNtkMinSlack(p *Sfm_Mit_t) int32                      { return 0 }
func Sfm_MitReadObjDelay(p *Sfm_Mit_t, iObj int32) int32             { return 0 }
func Sfm_MitTransferLoad(p *Sfm_Mit_t, pNew, pOld *Object)           {}
func Sfm_MitTimingGrow(p *Sfm_Mit_t)                                 {}
func Sfm_MitUpdateLoad(p *Sfm_Mit_t, vTimeNodes []int32, fAdd int32) {}
func Sfm_MitUpdateTiming(p *Sfm_Mit_t, vTimeNodes []int32)           {}
func Sfm_MitSortArrayByArrival(p *Sfm_Mit_t, vNodes *[]int32, iPivot int32) int32 {
	return 0
}
func Sfm_MitPriorityNodes(p *Sfm_Mit_t, vCands *[]int32, Window int32) int32 { return 0 }
func Sfm_MitNodeIsNonCritical(p *Sfm_Mit_t, pPivot, pNode *Object) int32     { return 0 }
func Sfm_MitEvalRemapping(p *Sfm_Mit_t, vMffc []int32, pObj *Object, vFanins []int32, vMap []int32, pGate1 *mio.Mio_Gate_t, pFans1 []byte, pGate2 *mio.Mio_Gate_t, pFans2 []byte) int32 {
	return 0
}
