package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Cec_ManSatReadCex(p *Cec_ManSat_t) *Vec_Int_t { return p.vCex }
func Cec_ManSatAddToStore(store, pattern *Vec_Int_t, output int32) {
	vec.Vec_IntPush(store, output)
	if pattern == nil {
		vec.Vec_IntPush(store, -1)
		return
	}
	vec.Vec_IntPush(store, int32(vec.Vec_IntSize(pattern)))
	for _, lit := range vec.Vec_IntArray(pattern) {
		vec.Vec_IntPush(store, lit)
	}
}

// Cec_ManSatSolveMiter_rec records CI values in source cone traversal order.
// It shares the manager's current traversal ID across both queried roots.
func Cec_ManSatSolveMiter_rec(p *Cec_ManSat_t, graph *Gia_Man_t, obj *Gia_Obj_t) {
	if gia.Gia_ObjIsTravIdCurrent(graph, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(graph, obj)
	if gia.Gia_ObjIsCi(obj) != 0 {
		p.nCexLits++
		lit := gia.Gia_ObjCioId(obj) << 1
		if Cec_ObjSatVarValue(p, obj) == 0 {
			lit |= 1
		}
		vec.Vec_IntPush(p.vCex, lit)
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("cec: pattern cone requires AND or CI")
	}
	cecRequireAndGraph(graph, obj)
	Cec_ManSatSolveMiter_rec(p, graph, gia.Gia_ObjFanin0(obj))
	Cec_ManSatSolveMiter_rec(p, graph, gia.Gia_ObjFanin1(obj))
}

// Cec_ManSavePattern clears and reuses the source pattern vector.
func Cec_ManSavePattern(p *Cec_ManSat_t, a, b *Gia_Obj_t) {
	if p.values == nil {
		panic("cec: cannot save pattern without a SAT witness")
	}
	vec.Vec_IntClear(p.vCex)
	gia.Gia_ManIncrementTravId(p.pAig)
	Cec_ManSatSolveMiter_rec(p, p.pAig, gia.Gia_Regular(a))
	if b != nil {
		Cec_ManSatSolveMiter_rec(p, p.pAig, gia.Gia_Regular(b))
	}
}

func Cec_ManSatSolveMiter(graph *Gia_Man_t, params *Cec_ParSat_t, statusOut **Vec_Str_t) *Vec_Int_t {
	gia.Gia_ManSetPhase(graph)
	gia.Gia_ManLevelNum(graph)
	gia.Gia_ManIncrementTravId(graph)
	p := Cec_ManSatCreate(graph, params)
	defer Cec_ManSatStop(p)
	outputs := graph.OutputLiterals()
	status := vec.Vec_StrAlloc(len(outputs))
	store := vec.Vec_IntAlloc(16)
	for i, lit := range outputs {
		// C handles constant outputs before calling Cec_ManSatCheckNode.
		if lit <= 1 {
			if lit == 0 {
				vec.Vec_StrPush(status, 1)
			} else {
				vec.Vec_StrPush(status, 0)
				Cec_ManSatAddToStore(store, vec.Vec_IntAlloc(0), int32(i))
			}
			continue
		}
		if !p.pPars.deadline.IsZero() && !time.Now().Before(p.pPars.deadline) {
			vec.Vec_StrPush(status, byte(255))
			Cec_ManSatAddToStore(store, nil, int32(i))
			continue
		}
		result := p.query(lit, nil)
		vec.Vec_StrPush(status, byte(result))
		switch result {
		case -1:
			Cec_ManSatAddToStore(store, nil, int32(i))
		case 0:
			Cec_ManSavePattern(p, gia.Gia_ManObj(graph, lit>>1), nil)
			Cec_ManSatAddToStore(store, p.vCex, int32(i))
		}
	}
	if statusOut != nil {
		*statusOut = status
	}
	return store
}

func Cex_ManGenSimple(p *Cec_ManSat_t, output int32) *Abc_Cex_t {
	cex := util.Abc_CexAlloc(0, int32(len(p.pAig.InputLiterals())), 1)
	cex.IPo = output
	return cex
}
func Cex_ManGenCex(p *Cec_ManSat_t, output int32) *Abc_Cex_t {
	if p.values == nil {
		panic("cec: counterexample requested without SAT witness")
	}
	cex := Cex_ManGenSimple(p, output)
	for i, lit := range p.pAig.InputLiterals() {
		if p.values[lit>>1] {
			cex.PData[i>>5] |= uint32(1) << uint(i&31)
		}
	}
	return cex
}
