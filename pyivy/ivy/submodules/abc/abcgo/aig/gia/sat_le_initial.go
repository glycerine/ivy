package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

// Sle_ManDeriveInit records node, cut, delay and edge variables of the current mapping.
func Sle_ManDeriveInit(p *Sle_Man_t) {
	g := p.pGia
	if Gia_ManHasMapping(g) == 0 {
		return
	}
	vec.Vec_IntClear(p.vPolars)
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(g, id)) == 0 || Gia_ObjIsLut(g, id) == 0 {
			continue
		}
		vec.Vec_IntPush(p.vPolars, id)
		fanins := Gia_ObjLutFanins(g, id)
		index := -1
		var chosen []int32
		for i, cut := range sleList(p.vCuts, id, false) {
			leaves := Sle_CutLeaves(cut)
			if Sle_ManCheckContained(leaves, int32(len(leaves)), fanins, int32(len(fanins))) != 0 {
				index = i
				chosen = cut
				break
			}
		}
		if index < 0 {
			fmt.Printf("Cannot find cut at node %d: %v\n", id, fanins)
			Sle_ManPrintCuts(g, p.vCuts, id)
			panic("gia: missing mapped LUT cut")
		}
		vec.Vec_IntPush(p.vPolars, vec.Vec_IntEntry(p.vCutFirst, int(id))+int32(index))
		if Sle_ManCutHasPisOnly(chosen, p.vMask) != 0 {
			vec.Vec_IntPush(p.vPolars, vec.Vec_IntEntry(p.vDelayFirst, int(id)))
		}
	}
	vec.Vec_IntSort(p.vPolars, 0)
	if g.VEdge1 == nil {
		return
	}
	edges := vec.Vec_IntArray(Gia_ManEdgeToArray(g))
	for i := 0; i < len(edges); i += 2 {
		fan, id := edges[i], edges[i+1]
		if fan >= id || Gia_ObjIsLut(g, fan) == 0 || Gia_ObjIsLut(g, id) == 0 || Gia_ObjIsAnd(Gia_ManObj(g, fan)) == 0 || Gia_ObjIsAnd(Gia_ManObj(g, id)) == 0 {
			panic("gia: invalid mapped edge")
		}
		index := vec.Vec_IntFind(vec.Vec_WecEntry(p.vCutFanins, int(id)), fan)
		if index >= 0 {
			vec.Vec_IntPush(p.vPolars, vec.Vec_IntEntry(p.vEdgeFirst, int(id))+int32(index))
		}
	}
}
