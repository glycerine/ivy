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
	seen := make(map[int32]bool)
	var visit func(int32)
	visit = func(id int32) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsCi(o) != 0 {
			value := uint32(0)
			if sat.values[id] {
				value = 1
			}
			cecWriteBit(cecInfo(info, regs+gia.Gia_ObjCioId(o)), pattern, value)
			return
		}
		a, b, c, kind := p.Gate(id)
		visit(a >> 1)
		if kind == 2 || kind == 3 || kind == 4 {
			visit(b >> 1)
		}
		if kind == 4 {
			visit(c >> 1)
		}
	}
	visit(gia.Gia_ObjId(p, obj))
}
func Cec_ManSatSolveSeq(info *Vec_Ptr_t, p *Gia_Man_t, pars *Cec_ParSat_t, regs int32, patterns *int32) *Vec_Str_t {
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
		Cec_ManSatSolveSeq_rec(sat, p, gia.Gia_ManObj(p, lit>>1), info, next, regs)
		next++
	}
	if patterns != nil {
		*patterns = next - 1
	}
	return status
}
