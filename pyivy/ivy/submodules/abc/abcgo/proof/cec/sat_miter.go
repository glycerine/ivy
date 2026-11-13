package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
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

// Cec_ManSavePattern records only the inputs in the queried cones, retaining
// the original cube literal convention: even requires true, odd requires false.
func Cec_ManSavePattern(p *Cec_ManSat_t, a, b *Gia_Obj_t) {
	if p.values == nil {
		panic("cec: cannot save pattern without a SAT witness")
	}
	p.vCex = vec.Vec_IntAlloc(16)
	indices := make(map[int32]int)
	for i, lit := range p.pAig.InputLiterals() {
		indices[lit>>1] = i
	}
	visited := make(map[int32]bool)
	var visit func(int32)
	visit = func(id int32) {
		if visited[id] {
			return
		}
		visited[id] = true
		x, y, z, kind := p.pAig.Gate(id)
		if kind == core.Input {
			lit := int32(indices[id]) << 1
			if !p.values[id] {
				lit |= 1
			}
			vec.Vec_IntPush(p.vCex, lit)
			return
		}
		switch kind {
		case core.And, core.Xor:
			visit(x >> 1)
			visit(y >> 1)
		case core.Mux:
			visit(x >> 1)
			visit(y >> 1)
			visit(z >> 1)
		case core.Buffer, core.Output:
			visit(x >> 1)
		}
	}
	if a != nil {
		visit(gia.Gia_ObjId(p.pAig, gia.Gia_Regular(a)))
	}
	if b != nil {
		visit(gia.Gia_ObjId(p.pAig, gia.Gia_Regular(b)))
	}
}

func Cec_ManSatSolveMiter(graph *Gia_Man_t, params *Cec_ParSat_t, statusOut **Vec_Str_t) *Vec_Int_t {
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
