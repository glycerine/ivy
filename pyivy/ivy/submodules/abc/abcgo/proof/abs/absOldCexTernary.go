package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
)

// Saig_ManCexVerifyUsingTernary checks whether the care bits force the
// counterexample's failing output even when every other input is unknown.
func Saig_ManCexVerifyUsingTernary(g *aig.Graph, cex, care *util.Abc_Cex_t) int32 {
	if cex.IFrame != care.IFrame || cex.NBits != care.NBits || cex.IPo < 0 || cex.IPo >= saig.Saig_ManPoNum(g) {
		panic("invalid ternary counterexample")
	}
	values := make([]int32, len(g.Nodes()))
	values[g.Const1().NodeID()] = SAIG_ONE
	for i := int32(0); i < saig.Saig_ManRegNum(g); i++ {
		if util.Abc_InfoHasBit(cex.PData, i) || util.Abc_InfoHasBit(care.PData, i) {
			panic("counterexample must start in the zero state")
		}
		values[saig.Saig_ManLo(g, i).NodeID()] = SAIG_ZER
	}
	get := func(e aig.Edge) int32 {
		v := values[e.NodeID()]
		if e.Complemented() {
			return Saig_ManSimInfoNot(v)
		}
		return v
	}
	bit := cex.NRegs
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		for i := int32(0); i < saig.Saig_ManPiNum(g); i++ {
			v := SAIG_UND
			if util.Abc_InfoHasBit(care.PData, bit) {
				v = SAIG_ZER
				if util.Abc_InfoHasBit(cex.PData, bit) {
					v = SAIG_ONE
				}
			}
			values[g.Inputs()[i].NodeID()] = v
			bit++
		}
		for _, e := range g.LiveEdges() {
			n := g.Node(e)
			switch n.Kind {
			case aig.And:
				values[e.NodeID()] = Saig_ManSimInfoAnd(get(n.Fanins[0]), get(n.Fanins[1]))
			case aig.Buffer:
				values[e.NodeID()] = get(n.Fanins[0])
			case aig.ExclusiveOr:
				panic("ternary counterexample verification expects an AND-inverter graph")
			}
		}
		for _, e := range g.Outputs() {
			values[e.NodeID()] = get(g.Node(e).Fanins[0])
		}
		for i := int32(0); i < saig.Saig_ManRegNum(g); i++ {
			values[saig.Saig_ManLo(g, i).NodeID()] = values[saig.Saig_ManLi(g, i).NodeID()]
		}
	}
	if bit != cex.NBits {
		panic("counterexample bit count does not match graph")
	}
	if values[g.Outputs()[cex.IPo].NodeID()] == SAIG_ONE {
		return 1
	}
	return 0
}
