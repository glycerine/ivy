package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"strings"
)

func createConstrained0LiveCone(p *aig.Graph, signals []aig.Edge) aig.Edge {
	if len(signals) == 0 {
		panic("live: missing constrained-liveness target")
	}
	return aig.Aig_Or(p, all(p, signals[:len(signals)-1]).Not(), signals[len(signals)-1])
}
func collectCSSignals(n *Network, p *aig.Graph) []aig.Edge {
	signals := []aig.Edge{}
	target := aig.InvalidEdge
	for i := 0; i < propertyOutputs(p); i++ {
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
	c := newGraphCopy(p, "_0Live")
	c.copyInputs()
	mapped := make([]aig.Edge, len(signals))
	for i, e := range signals {
		mapped[i] = c.signal(e)
	}
	for i := 0; i < propertyOutputs(p); i++ {
		aig.Aig_ObjCreateCo(c.destination, c.signal(p.Outputs()[i]))
	}
	if index != nil {
		*index = int32(propertyOutputs(p))
	}
	aig.Aig_ObjCreateCo(c.destination, createConstrained0LiveCone(c.destination, mapped))
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
