package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

// giaMfsBoxGroups maps MFS nodes for each timing box to its inputs and outputs.
// Group map entries encode the box index and whether the node is an output.
func giaMfsBoxGroups(p *Gia_Man_t, copies []int32, nodeCount int) ([][]int, []int) {
	groups := make([][]int, Gia_ManBoxNum(p))
	groupMap := make([]int, nodeCount)
	for i := range groupMap {
		groupMap[i] = -1
	}
	if len(groups) == 0 {
		return groups, groupMap
	}
	timing := p.PManTime.(*tim.Tim_Man_t)
	ci := tim.Tim_ManPiNum(timing)
	co := 0
	cis, cos := vec.Vec_IntArray(p.vCis), vec.Vec_IntArray(p.vCos)
	for box := range groups {
		ins, outs := tim.Tim_ManBoxInputNum(timing, box), tim.Tim_ManBoxOutputNum(timing, box)
		if co+ins > len(cos) || ci+outs > len(cis) {
			panic("gia: invalid box terminal counts")
		}
		for _, id := range cos[co : co+ins] {
			node := int(copies[id])
			if node <= 0 || node >= nodeCount {
				panic("gia: invalid MFS box input")
			}
			groups[box] = append(groups[box], node)
			groupMap[node] = box << 1
		}
		for _, id := range cis[ci : ci+outs] {
			node := int(copies[id])
			if node <= 0 || node >= nodeCount {
				panic("gia: invalid MFS box output")
			}
			groups[box] = append(groups[box], node)
			groupMap[node] = box<<1 | 1
		}
		co += ins
		ci += outs
	}
	if ci != len(cis) || co+tim.Tim_ManPoNum(timing) != len(cos) {
		panic("gia: inconsistent box terminals")
	}
	return groups, groupMap
}
