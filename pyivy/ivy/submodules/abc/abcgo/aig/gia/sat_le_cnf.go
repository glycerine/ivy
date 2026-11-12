package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Sle_ManDeriveCnf creates cover, edge and delay constraints for LUT mapping.
func Sle_ManDeriveCnf(p *Sle_Man_t, conflictLimit, dynamic int32) {
	p.pSat = bsat.NewSolver()
	for p.pSat.NumVars() < p.nVarsTotal {
		p.pSat.NewVariable()
	}
	p.pSat.SetLimits(int64(conflictLimit), 0)
	add := func(lits ...int32) {
		if !p.pSat.AddClause(lits...) {
			panic("gia: inconsistent SAT LUT constraints")
		}
	}
	pos := func(v int32) int32 { return 2 * v }
	neg := func(v int32) int32 { return 2*v + 1 }
	g := p.pGia
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		driver := Gia_ObjFaninId0p(g, Gia_ManCo(g, i))
		if vec.Vec_BitEntry(p.vMask, int(driver)) != 0 {
			add(pos(driver))
		}
	}
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(g, id)) == 0 {
			continue
		}
		cutBase := vec.Vec_IntEntry(p.vCutFirst, int(id))
		edgeBase := vec.Vec_IntEntry(p.vEdgeFirst, int(id))
		cuts := sleList(p.vCuts, id, false)
		fans := vec.Vec_IntArray(vec.Vec_WecEntry(p.vCutFanins, int(id)))
		clause := []int32{neg(id)}
		for i := range cuts {
			clause = append(clause, pos(cutBase+int32(i)))
		}
		add(clause...)
		for i := range cuts {
			for j := i + 1; j < len(cuts); j++ {
				add(neg(cutBase+int32(i)), neg(cutBase+int32(j)))
			}
		}
		edgeCuts := make([][]int32, len(fans))
		for i, cut := range cuts {
			cutVar := cutBase + int32(i)
			for _, leaf := range Sle_CutLeaves(cut) {
				if vec.Vec_BitEntry(p.vMask, int(leaf)) == 0 {
					continue
				}
				add(neg(cutVar), pos(leaf))
				p.nCutClas++
				index := vec.Vec_IntFind(vec.Vec_WecEntry(p.vCutFanins, int(id)), leaf)
				if index < 0 {
					panic("gia: missing SAT cut edge")
				}
				edgeCuts[index] = append(edgeCuts[index], cutVar)
			}
			add(neg(cutVar), pos(id))
		}
		for i, covering := range edgeCuts {
			if len(covering) == 0 {
				panic("gia: uncovered SAT cut edge")
			}
			vec.Vec_WecPush(p.vFanoutEdges, int(fans[i]), edgeBase+int32(i))
			clause := []int32{neg(edgeBase + int32(i))}
			for _, v := range covering {
				clause = append(clause, pos(v))
			}
			add(clause...)
			p.nEdgeClas++
		}
	}
	for id := 0; id < vec.Vec_WecSize(p.vFanoutEdges); id++ {
		edges := vec.Vec_WecEntry(p.vFanoutEdges, id)
		fans := vec.Vec_WecEntry(p.vCutFanins, id)
		base := vec.Vec_IntEntry(p.vEdgeFirst, id)
		for i := 0; i < vec.Vec_IntSize(fans); i++ {
			vec.Vec_IntPush(edges, base+int32(i))
		}
		if dynamic == 0 {
			list := vec.Vec_IntArray(edges)
			for i := range list {
				for j := i + 1; j < len(list); j++ {
					add(neg(list[i]), neg(list[j]))
					p.nEdgeClas2++
				}
			}
		}
	}
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(g, id)) == 0 {
			continue
		}
		edgeBase := vec.Vec_IntEntry(p.vEdgeFirst, int(id))
		delayBase := vec.Vec_IntEntry(p.vDelayFirst, int(id))
		for _, cut := range sleList(p.vCuts, id, false) {
			if Sle_ManCutHasPisOnly(cut, p.vMask) != 0 {
				add(pos(delayBase))
				break
			}
		}
		for e, fan := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vCutFanins, int(id))) {
			in := vec.Vec_IntEntry(p.vDelayFirst, int(fan))
			edge := edgeBase + int32(e)
			for d := int32(0); d < p.nLevels; d++ {
				clause := []int32{neg(id), neg(fan), neg(in + d), pos(edge)}
				if d < p.nLevels-1 {
					clause = append(clause, pos(delayBase+d+1))
				}
				add(clause...)
				clause = []int32{neg(id), neg(fan), neg(in + d)}
				if d < p.nLevels-1 {
					clause = append(clause, neg(edge))
				}
				clause = append(clause, pos(delayBase+d))
				add(clause...)
				p.nDelayClas += 2
			}
		}
	}
}
