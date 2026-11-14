package gia

import "github.com/glycerine/abcgo/misc/vec"

func sleValue(p *Sle_Man_t, variable int32) bool {
	value, ok := p.pSat.Value(variable)
	if !ok {
		panic("gia: SAT LUT model unavailable")
	}
	return value
}

// Sle_ManAddEdgeConstraints blocks model violations of the incident edge limit.
func Sle_ManAddEdgeConstraints(p *Sle_Man_t, maxEdges int32) int32 {
	if maxEdges != 1 && maxEdges != 2 {
		panic("gia: invalid SAT edge limit")
	}
	var clauses [][]int32
	var added int32
	for id := 0; id < vec.Vec_WecSize(p.vFanoutEdges); id++ {
		var selected []int32
		for _, edge := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vFanoutEdges, id)) {
			if sleValue(p, edge) {
				selected = append(selected, edge)
			}
		}
		if len(selected) <= int(maxEdges) {
			continue
		}
		added++
		for i := 0; i < len(selected); i++ {
			for j := i + 1; j < len(selected); j++ {
				if maxEdges == 1 {
					clauses = append(clauses, []int32{2*selected[i] + 1, 2*selected[j] + 1})
				} else {
					for k := j + 1; k < len(selected); k++ {
						clauses = append(clauses, []int32{2*selected[i] + 1, 2*selected[j] + 1, 2*selected[k] + 1})
					}
				}
			}
		}
	}
	for _, clause := range clauses {
		if !p.pSat.AddClause(clause...) {
			panic("gia: inconsistent dynamic edge constraints")
		}
		p.nEdgeClas2++
	}
	return added
}
func Sle_ManDeriveResult(p *Sle_Man_t, edges, mapping *Vec_Int_t) {
	g := p.pGia
	vec.Vec_IntFill(mapping, int(g.nObjs), 0)
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(g, id)) == 0 || !sleValue(p, id) {
			continue
		}
		base := vec.Vec_IntEntry(p.vCutFirst, int(id))
		var chosen []int32
		for i, cut := range sleList(p.vCuts, id, false) {
			if sleValue(p, base+int32(i)) {
				if chosen != nil {
					panic("gia: multiple selected SAT cuts")
				}
				chosen = cut
			}
		}
		if chosen == nil {
			panic("gia: missing selected SAT cut")
		}
		vec.Vec_IntWriteEntry(mapping, int(id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, Sle_CutSize(chosen))
		for _, leaf := range Sle_CutLeaves(chosen) {
			vec.Vec_IntPush(mapping, leaf)
		}
		vec.Vec_IntPush(mapping, id)
	}
	previous := g.VMapping
	g.VMapping = mapping
	defer func() { g.VMapping = previous }()
	vec.Vec_IntClear(edges)
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(g, id)) == 0 || !sleValue(p, id) {
			continue
		}
		base := vec.Vec_IntEntry(p.vEdgeFirst, int(id))
		fanins := Gia_ObjLutFanins(g, id)
		for i, fan := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vCutFanins, int(id))) {
			if !sleValue(p, fan) || !sleValue(p, base+int32(i)) {
				continue
			}
			for _, leaf := range fanins {
				if leaf == fan {
					vec.Vec_IntPushTwo(edges, fan, id)
					break
				}
			}
		}
	}
}
