package sbd

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func sbcAddInternal(p *Network, id int32, path *vec.Vec_Bit_t, seen map[int32]bool) int32 {
	if seen[id] {
		return int32(vec.Vec_BitEntry(path, int(id)))
	}
	seen[id] = true
	node := p.object(id)
	if node.Kind == aig.Input {
		return int32(vec.Vec_BitEntry(path, int(id)))
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("critical LUT path contains non-gate")
	}
	value := int32(0)
	for _, fanin := range p.cut(id) {
		value |= sbcAddInternal(p, fanin, path, seen)
	}
	if value != 0 {
		vec.Vec_BitWriteEntry(path, int(id), 1)
	}
	return value
}
func Sbc_ManAddInternalToPath_rec(p *Network, id int32, path *vec.Vec_Bit_t) int32 {
	return sbcAddInternal(p, id, path, map[int32]bool{})
}
func Sbc_ManAddInternalToPath(p *Network, path *vec.Vec_Bit_t) {
	for _, edge := range p.Graph.LiveEdges() {
		id := int32(edge.NodeID())
		if _, ok := p.LUTs[id]; !ok || vec.Vec_BitEntry(path, int(id)) == 0 {
			continue
		}
		seen := map[int32]bool{}
		for _, fanin := range p.cut(id) {
			seen[fanin] = true
		}
		sbcAddInternal(p, id, path, seen)
	}
}
func sbcCriticalPath(p *Network, levels []int32, id, levelFan int32, path *vec.Vec_Bit_t, slack int32, seen map[int32]bool) {
	if seen[id] {
		return
	}
	seen[id] = true
	vec.Vec_BitWriteEntry(path, int(id), 1)
	node := p.object(id)
	if node.Kind == aig.Input {
		for _, co := range p.BoxInputs[id] {
			driver := p.outputDriver(co)
			if driver != 0 && levels[driver]+slack >= levelFan-1 {
				sbcCriticalPath(p, levels, driver, levels[driver], path, max(0, levels[driver]+slack-(levelFan-1)), seen)
			}
		}
		return
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("critical path contains non-LUT gate")
	}
	for _, fanin := range p.cut(id) {
		if levels[fanin]+slack >= levelFan-1 {
			sbcCriticalPath(p, levels, fanin, levels[fanin], path, max(0, levels[fanin]+slack-(levelFan-1)), seen)
		}
	}
}
func Sbc_ManCriticalPath_rec(p *Network, levels []int32, id, level int32, path *vec.Vec_Bit_t, slack int32) {
	sbcCriticalPath(p, levels, id, level, path, slack, map[int32]bool{})
}
func Sbc_ManCriticalPath(p *Network) *vec.Vec_Bit_t {
	maximum := p.deriveLevels()
	path := vec.Vec_BitStart(len(p.Graph.Nodes()))
	seen := map[int32]bool{}
	for _, co := range p.Graph.Outputs() {
		driver := p.outputDriver(int32(co.NodeID()))
		if driver != 0 && p.Levels[driver] == maximum {
			sbcCriticalPath(p, p.Levels, driver, p.Levels[driver], path, 1, seen)
		}
	}
	Sbc_ManAddInternalToPath(p, path)
	return path
}
func Sbc_ManDelayTrace(p *Network) {
	if len(p.LUTs) == 0 {
		fmt.Println("No mapping is available.")
		return
	}
	maximum := p.deriveLevels()
	path := vec.Vec_BitStart(len(p.Graph.Nodes()))
	for _, co := range p.Graph.Outputs() {
		driver := p.outputDriver(int32(co.NodeID()))
		if p.Levels[driver] == maximum {
			vec.Vec_BitWriteEntry(path, int(driver), 1)
		}
	}
	gates := p.Graph.LiveEdges()
	nodes, edges, all := 0, 0, 0
	for i := len(gates) - 1; i >= 0; i-- {
		id := int32(gates[i].NodeID())
		cut, ok := p.LUTs[id]
		if !ok || vec.Vec_BitEntry(path, int(id)) == 0 {
			continue
		}
		nodes++
		for _, fanin := range cut.Leaves {
			if p.Levels[fanin]+1 < p.Levels[id] {
				continue
			}
			vec.Vec_BitWriteEntry(path, int(fanin), 1)
			edges++
		}
	}
	for id, cut := range p.LUTs {
		for _, fanin := range cut.Leaves {
			if vec.Vec_BitEntry(path, int(id)) != 0 && vec.Vec_BitEntry(path, int(fanin)) != 0 {
				all++
			}
		}
	}
	fmt.Printf("AIG = %d. LUT = %d. Lev = %d. Path nodes = %d. Path edges = %d. (%d.)\n", len(p.LUTs), len(p.LUTs), maximum, nodes, edges, all)
}
