package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

// projectAIGProperties copies a sequential AIG while retaining only the
// requested primary outputs. Register-input outputs remain in original order.
func projectAIGProperties(source *aig.Graph, properties []int) (*aig.Graph, error) {
	if source == nil {
		return nil, fmt.Errorf("bmc: nil source AIG")
	}
	regs := int(source.RegisterCount())
	inputs := source.Inputs()
	outputs := source.Outputs()
	if regs < 0 || regs > len(inputs) || regs > len(outputs) {
		return nil, fmt.Errorf("bmc: invalid sequential terminal counts")
	}
	poCount := len(outputs) - regs
	for _, property := range properties {
		if property < 0 || property >= poCount {
			return nil, fmt.Errorf("bmc: primary output index %d is out of range", property)
		}
	}

	nodes := source.Nodes()
	destination := aig.NewGraph()
	destination.Name, destination.Spec = source.Name, source.Spec
	mapping := make([]aig.Edge, len(nodes))
	for i := range mapping {
		mapping[i] = aig.InvalidEdge
	}
	mapping[0] = destination.Const1()
	for _, input := range inputs {
		if input.NodeID() < 0 || input.NodeID() >= len(mapping) {
			return nil, fmt.Errorf("bmc: input node is outside source graph")
		}
		mapping[input.NodeID()] = destination.NewInput()
	}
	mapEdge := func(edge aig.Edge) (aig.Edge, error) {
		id := edge.NodeID()
		if id < 0 || id >= len(mapping) || mapping[id] == aig.InvalidEdge {
			return aig.InvalidEdge, fmt.Errorf("bmc: source edge %d is not topologically mapped", edge)
		}
		mapped := mapping[id]
		if edge.Complemented() {
			mapped = mapped.Not()
		}
		return mapped, nil
	}
	for id, node := range nodes {
		if id == 0 || mapping[id] != aig.InvalidEdge {
			continue
		}
		switch node.Kind {
		case aig.Constant:
			mapping[id] = destination.Const1()
		case aig.And:
			a, err := mapEdge(node.Fanins[0])
			if err != nil {
				return nil, err
			}
			b, err := mapEdge(node.Fanins[1])
			if err != nil {
				return nil, err
			}
			mapping[id] = destination.AND(a, b)
		case aig.ExclusiveOr:
			a, err := mapEdge(node.Fanins[0])
			if err != nil {
				return nil, err
			}
			b, err := mapEdge(node.Fanins[1])
			if err != nil {
				return nil, err
			}
			mapping[id] = destination.NewXor(a, b)
		case aig.Buffer:
			value, err := mapEdge(node.Fanins[0])
			if err != nil {
				return nil, err
			}
			mapping[id] = value
		case aig.Output:
			// Outputs are re-created in the selected property order below.
		default:
			return nil, fmt.Errorf("bmc: unsupported AIG node kind %d", node.Kind)
		}
	}
	for _, property := range properties {
		driver, err := mapEdge(source.Node(outputs[property]).Fanins[0])
		if err != nil {
			return nil, err
		}
		destination.AddOutput(driver)
	}
	for _, registerInput := range outputs[poCount:] {
		driver, err := mapEdge(source.Node(registerInput).Fanins[0])
		if err != nil {
			return nil, err
		}
		destination.AddOutput(driver)
	}
	aig.Aig_ManSetRegNum(destination, int32(regs))
	return destination, nil
}
