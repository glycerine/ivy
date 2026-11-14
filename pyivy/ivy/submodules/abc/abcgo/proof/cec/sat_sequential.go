package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Save only the queried cone's CI assignments; untouched simulation bits keep
// their independent random values. Shared fanins are visited once per witness.
func Cec_ManSatSolveSeq_rec(sat *Cec_ManSat_t, p *Gia_Man_t, obj *Gia_Obj_t, info *Vec_Ptr_t, pattern, regs int32) {
	if sat.values == nil {
		panic("cec: sequential pattern requires SAT witness")
	}
	if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	if gia.Gia_ObjIsCi(obj) != 0 {
		value := uint32(Cec_ObjSatVarValue(sat, obj))
		cecWriteBit(cecInfo(info, regs+gia.Gia_ObjCioId(obj)), pattern, value)
		sat.nCexLits++
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("cec: unsupported sequential pattern node")
	}
	cecRequireAndGraph(p, obj)
	Cec_ManSatSolveSeq_rec(sat, p, gia.Gia_ObjFanin0(obj), info, pattern, regs)
	Cec_ManSatSolveSeq_rec(sat, p, gia.Gia_ObjFanin1(obj), info, pattern, regs)
}
func Cec_ManSatSolveSeq(info *Vec_Ptr_t, p *Gia_Man_t, pars *Cec_ParSat_t, regs int32, patterns *int32) *Vec_Str_t {
	gia.Gia_ManSetPhase(p)
	gia.Gia_ManLevelNum(p)
	gia.Gia_ManIncrementTravId(p)
	sat := Cec_ManSatCreate(p, pars)
	defer Cec_ManSatStop(sat)
	status := vec.Vec_StrAlloc(int(gia.Gia_ManCoNum(p)))
	initial := int32(32 * vec.Vec_PtrReadWordsSimInfo(info))
	capacity := initial
	if initial <= 0 {
		panic("cec: empty sequential pattern storage")
	}
	next := int32(0)
	for _, lit := range p.OutputLiterals() {
		if lit <= 1 {
			vec.Vec_StrPush(status, byte(1-lit))
			continue
		}
		result := sat.query(lit, nil)
		vec.Vec_StrPush(status, byte(result))
		if result != 0 || lit < 2 {
			continue
		}
		if next == capacity {
			oldWords := vec.Vec_PtrReadWordsSimInfo(info)
			vec.Vec_PtrReallocSimInfo(info)
			vec.Vec_PtrCleanSimInfo(info, oldWords, 2*oldWords)
			capacity = 2 * capacity
		}
		if next%initial == 0 {
			next++
		}
		gia.Gia_ManIncrementTravId(p)
		Cec_ManSatSolveSeq_rec(sat, p, gia.Gia_ManObj(p, lit>>1), info, next, regs)
		next++
	}
	if sat.pPars.FVerbose != 0 {
		Cec_ManSatPrintStats(sat)
	}
	if patterns != nil {
		*patterns = next - 1
	}
	return status
}
