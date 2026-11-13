package abs

import (
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Rf2_Obj_t struct {
	Value, fVisit, fPPi, Prio uint32
}

type Rf2_Man_t_ struct {
	pGia                                            *gia.Gia_Man_t
	pCex                                            *util.Abc_Cex_t
	vMap                                            *vec.Vec_Int_t
	fPropFanout, fVerbose                           int32
	vObjs, vFanins                                  *vec.Vec_Int_t
	pvVecs                                          []vec.Vec_Int_t
	vGrp2Ppi                                        *vec.Vec_Vec_t
	nMapWords                                       int32
	pObjs                                           []Rf2_Obj_t
	nObjs, nObjsAlloc, nObjsFrame, nCalls, nRefines int32
	timeFwd, timeBwd, timeVer, timeTotal            time.Duration
}
