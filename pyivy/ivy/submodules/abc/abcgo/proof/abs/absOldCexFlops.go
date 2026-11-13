package abs

import (
	"sort"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManCbaFilterFlops selects the candidate abstracted registers whose
// observed values disagree most often with their simulated next-state logic.
// Candidate entries index the zero entries in classes, not the original flops.
func Saig_ManCbaFilterFlops(g *aig.Graph, cex *util.Abc_Cex_t, classes, candidates *vec.Vec_Int_t, count int32) *vec.Vec_Int_t {
	if int(saig.Saig_ManRegNum(g)) != vec.Vec_IntSize(classes) || count < 0 || vec.Vec_IntSize(candidates) <= int(count) {
		panic("invalid flop-filter arguments")
	}
	var mapping []int32
	for i, class := range vec.Vec_IntArray(classes) {
		if class == 0 {
			mapping = append(mapping, int32(i))
		}
	}
	piCount := saig.Saig_ManPiNum(g)
	if piCount+int32(len(mapping)) != cex.NPis {
		panic("counterexample input count does not match abstraction")
	}
	values := make([]bool, len(g.Nodes()))
	values[g.Const1().NodeID()] = true
	costs := make([]int, len(mapping))
	value := func(e aig.Edge) bool { return values[e.NodeID()] != e.Complemented() }
	for frame := int32(0); frame < cex.IFrame; frame++ {
		base := cex.NRegs + frame*cex.NPis
		// The C implementation omitted these real PIs, incorrectly treating them as zero.
		for i := int32(0); i < piCount; i++ {
			values[g.Inputs()[i].NodeID()] = util.Abc_InfoHasBit(cex.PData, base+i)
		}
		for i, reg := range mapping {
			values[saig.Saig_ManLo(g, reg).NodeID()] = util.Abc_InfoHasBit(cex.PData, base+piCount+int32(i))
		}
		for _, e := range g.LiveEdges() {
			n := g.Node(e)
			switch n.Kind {
			case aig.And:
				values[e.NodeID()] = value(n.Fanins[0]) && value(n.Fanins[1])
			case aig.ExclusiveOr:
				panic("flop filtering expects an AND-inverter graph")
			case aig.Buffer:
				values[e.NodeID()] = value(n.Fanins[0])
			}
		}
		for _, e := range g.Outputs() {
			values[e.NodeID()] = value(g.Node(e).Fanins[0])
		}
		for i := int32(0); i < saig.Saig_ManRegNum(g); i++ {
			values[saig.Saig_ManLo(g, i).NodeID()] = values[saig.Saig_ManLi(g, i).NodeID()]
		}
		next := base + cex.NPis + piCount
		for i, reg := range mapping {
			if values[saig.Saig_ManLi(g, reg).NodeID()] != util.Abc_InfoHasBit(cex.PData, next+int32(i)) {
				costs[i]++
			}
		}
	}
	ordered := append([]int32(nil), vec.Vec_IntArray(candidates)...)
	sort.SliceStable(ordered, func(i, j int) bool { return costs[ordered[i]] > costs[ordered[j]] })
	result := vec.Vec_IntAlloc(int(count))
	for _, entry := range ordered[:count] {
		vec.Vec_IntPush(result, entry)
	}
	return result
}
