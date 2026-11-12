package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

// nfExtractWindow exports the cut choices and current solution used by the
// SAT-based mapping optimizer. The input GIA must have CIs before its ANDs.
func (p *nfMatcher) nfExtractWindow(r *nfMappingRefs, roots *Vec_Int_t, cuts, objCuts *Vec_Wec_t, solution, gates *Vec_Int_t, areas *Vec_Wrd_t, invArea *word, startVar, nVars int32) int32 {
	g := p.cuts.graph
	inputs := Gia_ManCiNum(g)
	ands := Gia_ManAndNum(g)
	shift := 2*inputs + 2
	if 2*ands+inputs > nVars {
		fmt.Printf("The number of variables is too large: 2*%d + %d = %d > %d.\n", ands, inputs, 2*ands+inputs, nVars)
		return 0
	}
	*invArea = p.cuts.library.cells[3].AreaW
	vec.Vec_IntClear(roots)
	vec.Vec_WecClear(cuts)
	vec.Vec_WecClear(objCuts)
	vec.Vec_IntClear(solution)
	vec.Vec_IntClear(gates)
	vec.Vec_WrdClear(areas)
	for _, id := range vec.Vec_IntArray(g.vCos) {
		co := Gia_ManObj(g, id)
		fanin := Gia_ObjFaninId0(co, id)
		if Gia_ObjIsCi(Gia_ManObj(g, fanin)) != 0 {
			panic("gia: Nf window output has a primary-input driver")
		}
		vec.Vec_IntPush(roots, 2*fanin+Gia_ObjFaninC0(co)-shift)
	}
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		id := obj.id
		if int32(vec.Vec_WecSize(objCuts)) != 2*id-shift {
			panic("gia: Nf window objects are not ordered")
		}
		invIndex := [2]int{-1, -1}
		var selected [2]*nfMatch
		for phase := int32(0); phase < 2; phase++ {
			if r.refs[id][phase] == 0 {
				continue
			}
			kind := r.selected[id][phase]
			if kind < 0 {
				panic("gia: missing Nf window selection")
			}
			m := &p.objects[id][phase][kind]
			if m.inverted {
				invIndex[phase] = vec.Vec_IntSize(solution)
				vec.Vec_IntPush(solution, -1)
			} else {
				selected[phase] = m
			}
		}
		for phase := int32(0); phase < 2; phase++ {
			v := vec.Vec_WecPushLevel(objCuts)
			vec.Vec_IntPush(v, 2*(2*id+phase-shift)+1)
		}
		if p.cuts.sets[id] != 0 {
			data := p.cuts.storage.at(p.cuts.sets[id])
			offset := int32(1)
			for i := int32(0); i < data[0]; i++ {
				cut := data[offset:]
				count, function := Nf_CutSize(cut), Nf_CutFunc(cut)
				if int(function>>1) < len(p.cuts.library.matches) {
					leaves := Nf_CutLeaves(cut)
					for _, match := range p.cuts.library.matches[function>>1] {
						phase := int32(match.cfg.fCompl) ^ (function & 1)
						cfg := match.cfg
						cfg.fCompl = 0
						vec.Vec_IntPush(gates, match.gate)
						vec.Vec_WrdPush(areas, p.cuts.library.cells[match.gate].AreaW)
						for polarity := 0; polarity < 2; polarity++ {
							m := selected[polarity]
							if m != nil && m.cutOffset == offset && m.gate == match.gate && m.cfg == cfg {
								vec.Vec_IntPush(solution, int32(vec.Vec_WecSize(cuts)))
							}
						}
						cutLit := 2 * (startVar + int32(vec.Vec_WecSize(cuts)))
						v := vec.Vec_WecPushLevel(cuts)
						vec.Vec_IntPush(v, 2*id+phase)
						vec.Vec_IntPush(vec.Vec_WecEntry(objCuts, int(2*id+phase-shift)), cutLit)
						for pin := int32(0); pin < count; pin++ {
							fanin, polarity := leaves[Nf_CfgVar(cfg, pin)], Nf_CfgCompl(cfg, pin)
							if fanin >= inputs+1 || polarity != 0 {
								vec.Vec_IntPush(v, 2*fanin+polarity)
							}
						}
					}
				}
				offset += count + 1
			}
		}
		if invIndex[0] >= 0 && invIndex[1] >= 0 {
			panic("gia: circular Nf window inverters")
		}
		for phase := int32(0); phase < 2; phase++ {
			if invIndex[phase] >= 0 {
				vec.Vec_IntWriteEntry(solution, invIndex[phase], int32(vec.Vec_WecSize(cuts)))
			}
			vec.Vec_IntPush(vec.Vec_WecEntry(objCuts, int(2*id+phase-shift)), 2*(startVar+int32(vec.Vec_WecSize(cuts))))
			v := vec.Vec_WecPushLevel(cuts)
			vec.Vec_IntPush(v, 2*id+phase)
			vec.Vec_IntPush(v, 2*id+(phase^1))
			vec.Vec_IntPush(gates, 3)
			vec.Vec_WrdPush(areas, *invArea)
		}
	}
	for index, id := range vec.Vec_IntArray(g.vCis) {
		if r.refs[id][1] != 0 {
			vec.Vec_IntPush(solution, -(2*ands + int32(index)))
		}
	}
	return inputs
}
