package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Llb_AigMap selects AIG gates whose transitive CI support fits within a LUT
// of the requested size. This deterministic cut estimate keeps the LLB package
// independent of ABC's network mapper.
func Llb_AigMap(g *Aig_Man_t, lutSize, lutMin int32) *Vec_Int_t {
	result := vec.Vec_IntAlloc(32)
	if g == nil || lutSize < 1 {
		return result
	}
	nodes := g.Nodes()
	supports := make([]map[int]struct{}, len(nodes))
	for _, edge := range g.Inputs() {
		supports[edge.NodeID()] = map[int]struct{}{edge.NodeID(): {}}
	}
	outputs := g.Outputs()
	registers := int(g.RegisterCount())
	latchDrivers := map[int]bool{}
	for i := len(outputs) - registers; i < len(outputs); i++ {
		if i >= 0 {
			n := g.Node(outputs[i])
			latchDrivers[n.Fanins[0].NodeID()] = true
		}
	}
	for id, node := range nodes {
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			set := map[int]struct{}{}
			for _, fanin := range node.Fanins {
				for leaf := range supports[fanin.NodeID()] {
					set[leaf] = struct{}{}
				}
			}
			supports[id] = set
			if len(set) >= int(lutMin) && len(set) <= int(lutSize) && node.Refs > 0 && !latchDrivers[id] {
				vec.Vec_IntPush(result, int32(id))
			}
		case aig.Buffer, aig.Output:
			if node.Fanins[0] != aig.InvalidEdge {
				supports[id] = supports[node.Fanins[0].NodeID()]
			}
		}
	}
	return result
}
