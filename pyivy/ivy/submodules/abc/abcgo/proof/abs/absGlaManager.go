package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

const GA2_BIG_NUM = 0x3FFFFFF0

// Ga2_Man_t contains the gate abstraction's cuts, frame literals and solver.
type Ga2_Man_t struct {
	pGia                                                        *gia.Gia_Man_t
	pPars                                                       *Abs_Par_t
	vCnfs                                                       []*vec.Vec_Int_t
	vIds, vProofIds, vAbs, vValues                              *vec.Vec_Int_t
	nProofIds, LimAbs, LimPpi, nMarked, fUseNewLine             int32
	pRnm                                                        *Rnm_Man_t
	vId2Lit                                                     []*vec.Vec_Int_t
	pSat                                                        *bsat.LabeledSolver2
	proofActive                                                 bool
	nSatVars, nCexes, nObjAdded, nPdrCalls                      int32
	pTable                                                      []int32
	nTable, nHashHit, nHashMiss, nHashOver                      int32
	vLits, vIsopMem                                             *vec.Vec_Int_t
	pSopSizes                                                   []byte
	pSops                                                       [][]byte
	timeStart, timeInit, timeSat, timeUnsat, timeCex, timeOther int64
	cpuDeadline                                                 int64
}

func Ga2_ObjId(p *Ga2_Man_t, o *gia.Gia_Obj_t) int32 {
	return vec.Vec_IntEntry(p.vIds, int(gia.Gia_ObjId(p.pGia, o)))
}
func Ga2_ObjSetId(p *Ga2_Man_t, o *gia.Gia_Obj_t, id int32) {
	vec.Vec_IntWriteEntry(p.vIds, int(gia.Gia_ObjId(p.pGia, o)), id)
}
func Ga2_ObjCnf0(p *Ga2_Man_t, o *gia.Gia_Obj_t) *vec.Vec_Int_t {
	id := Ga2_ObjId(p, o)
	if id < 0 {
		panic("abs: object has no abstraction ID")
	}
	return p.vCnfs[2*id]
}
func Ga2_ObjCnf1(p *Ga2_Man_t, o *gia.Gia_Obj_t) *vec.Vec_Int_t {
	id := Ga2_ObjId(p, o)
	if id < 0 {
		panic("abs: object has no abstraction ID")
	}
	return p.vCnfs[2*id+1]
}
func Ga2_ObjIsAbs0(p *Ga2_Man_t, o *gia.Gia_Obj_t) bool {
	id := Ga2_ObjId(p, o)
	if id < 0 {
		panic("abs: object has no abstraction ID")
	}
	return id < p.LimAbs
}
func Ga2_ObjIsLeaf0(p *Ga2_Man_t, o *gia.Gia_Obj_t) bool {
	id := Ga2_ObjId(p, o)
	if id < 0 {
		panic("abs: object has no abstraction ID")
	}
	return id >= p.LimAbs && id < p.LimPpi
}
func Ga2_ObjIsAbs(p *Ga2_Man_t, o *gia.Gia_Obj_t) bool {
	return Ga2_ObjId(p, o) >= 0 && Ga2_ObjCnf0(p, o) != nil
}
func Ga2_ObjIsLeaf(p *Ga2_Man_t, o *gia.Gia_Obj_t) bool {
	return Ga2_ObjId(p, o) >= 0 && Ga2_ObjCnf0(p, o) == nil
}
func Ga2_MapFrameMap(p *Ga2_Man_t, frame int32) *vec.Vec_Int_t { return p.vId2Lit[frame] }
func Ga2_ObjFindLit(p *Ga2_Man_t, o *gia.Gia_Obj_t, frame int32) int32 {
	id := Ga2_ObjId(p, o)
	if id < 0 || int(id) >= vec.Vec_IntSize(p.vValues) {
		panic("abs: invalid abstraction ID")
	}
	return vec.Vec_IntEntry(Ga2_MapFrameMap(p, frame), int(id))
}
func Ga2_ObjAddLit(p *Ga2_Man_t, o *gia.Gia_Obj_t, frame, lit int32) {
	if Ga2_ObjFindLit(p, o, frame) != -1 {
		panic("abs: duplicate object literal")
	}
	vec.Vec_IntSetEntry(Ga2_MapFrameMap(p, frame), int(Ga2_ObjId(p, o)), lit)
}
func Ga2_ObjFindOrAddLit(p *Ga2_Man_t, o *gia.Gia_Obj_t, frame int32) int32 {
	lit := Ga2_ObjFindLit(p, o, frame)
	if lit == -1 {
		lit = p.nSatVars << 1
		p.nSatVars++
		Ga2_ObjAddLit(p, o, frame, lit)
	}
	return lit
}
