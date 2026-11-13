package live

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"strings"
)

// Signal edges belong to old; their source pData copies have been populated
// in new before this helper is called, as in createConstrained0LiveCone C.
func createConstrained0LiveCone(old, new *aig.Graph, signals []aig.Edge) aig.Edge {
	if len(signals) == 0 {
		panic("live: missing constrained-liveness target")
	}
	mapped := make([]aig.Edge, len(signals))
	for i, e := range signals {
		image := aig.Aig_ObjCopy(old, aig.Aig_Regular(e))
		if image == aig.InvalidEdge {
			panic("live: missing constrained-liveness signal copy")
		}
		mapped[i] = image ^ aig.Edge(e&1)
	}
	return aig.Aig_Or(new, all(new, mapped[:len(mapped)-1]).Not(), mapped[len(mapped)-1])
}
func collectCSSignals(n *Network, p *aig.Graph) []aig.Edge { return collectCSSignalsMode(n, p, true) }
func collectCSSignalsMode(n *Network, p *aig.Graph, source bool) []aig.Edge {
	signals := []aig.Edge{}
	target := aig.InvalidEdge
	for i := 0; i < outputCount(p, source); i++ {
		name := outputName(n, i)
		if strings.Contains(name, "csLiveConst_") {
			signals = append(signals, outputDriver(p, i))
		} else if strings.Contains(name, "csLiveTarget_") {
			target = outputDriver(p, i)
		}
	}
	if target == aig.InvalidEdge {
		panic("live: csLiveTarget_ output missing")
	}
	return append(signals, target)
}
func createNewAigWith0LivePo(p *aig.Graph, signals []aig.Edge, index *int32) *aig.Graph {
	return createNewAigWith0LivePoMode(p, signals, index, true)
}
func createNewAigWith0LivePoMode(p *aig.Graph, signals []aig.Edge, index *int32, source bool) *aig.Graph {
	c := newGraphCopy(p, "_0Live")
	if source {
		c = newSourceGraphCopy(p, "_0Live")
	}
	c.copyInputs()
	c.copySourceNodes()
	for _, e := range signals {
		c.signal(e)
	}
	for i := 0; i < c.outputCount(); i++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[i]))
	}
	if index != nil {
		*index = int32(c.outputCount())
	}
	aig.Aig_ObjCreateCo(c.destination, createConstrained0LiveCone(p, c.destination, signals))
	c.appendConstraints()
	for r := 0; r < registers(p); r++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[trueOutputs(p)+r]))
	}
	aig.Aig_ManSetRegNum(c.destination, int32(registers(p)))
	c.setConstraints()
	return c.destination
}
func generateWorkingAig(p *aig.Graph, n *Network, index *int32) *aig.Graph {
	return createNewAigWith0LivePo(p, collectCSSignals(n, p), index)
}

// checkMonotoneSignal and gatherMonotoneSignals are unfinished source entry
// points; the latter prints each internal object before returning NULL.
func checkMonotoneSignal() []aig.Edge { return nil }
func gatherMonotoneSignals(p *aig.Graph) []aig.Edge {
	for _, edge := range p.LiveEdges() {
		kind := p.Node(edge).Kind
		if kind == aig.And || kind == aig.ExclusiveOr {
			n := p.Node(edge)
			c0, c1 := " ", " "
			if n.Fanins[0].Complemented() {
				c0 = "'"
			}
			if n.Fanins[1].Complemented() {
				c1 = "'"
			}
			fmt.Printf("Node %4d : AND( %4d%s, %4d%s ) (refs = %3d)", edge.NodeID(), n.Fanins[0].NodeID(), c0, n.Fanins[1].NodeID(), c1, n.Refs)
			if equiv := aig.Aig_ObjEquiv(p, edge); equiv != aig.InvalidEdge {
				fmt.Printf("  { %4d ", edge.NodeID())
				for equiv != aig.InvalidEdge {
					phase := " "
					if aig.Aig_ObjPhase(p, equiv) != aig.Aig_ObjPhase(p, edge) {
						phase = "'"
					}
					fmt.Printf(" %4d%s", equiv.NodeID(), phase)
					equiv = aig.Aig_ObjEquiv(p, equiv)
				}
				fmt.Print(" }")
			} else if aig.Aig_ObjRepr(p, edge) != aig.InvalidEdge {
				fmt.Printf("  class of %d", edge.NodeID())
			}
			fmt.Println()
		}
	}
	return nil
}
