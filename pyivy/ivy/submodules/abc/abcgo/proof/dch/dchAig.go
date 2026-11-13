package dch

import (
	"fmt"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// DeriveTotalAig combines alternative implementations using shared primary
// inputs. Each output cone is copied from every snapshot before the first
// snapshot's output is connected. Secondary alternatives deliberately remain
// dangling: cleaning them would destroy the structures used for choice finding.
func DeriveTotalAig(variants []*aig.Graph) (*aig.Graph, error) {
	if len(variants) == 0 || variants[0] == nil {
		return nil, fmt.Errorf("dch: no input snapshots")
	}
	first := variants[0]
	for i, g := range variants {
		if g == nil {
			return nil, fmt.Errorf("dch: nil snapshot %d", i)
		}
		if err := g.Check(); err != nil {
			return nil, fmt.Errorf("dch: snapshot %d: %w", i, err)
		}
		if _, err := graphOrder(g); err != nil {
			return nil, fmt.Errorf("dch: snapshot %d: %w", i, err)
		}
		if g.InputCount() != first.InputCount() || g.OutputCount() != first.OutputCount() || aig.Aig_ManRegNum(g) != aig.Aig_ManRegNum(first) || aig.Aig_ManConstrNum(g) != aig.Aig_ManConstrNum(first) {
			return nil, fmt.Errorf("dch: snapshot %d has a different interface", i)
		}
	}
	result := aig.NewGraph()
	inputs := make([]aig.Edge, first.InputCount())
	for i := range inputs {
		inputs[i] = result.NewInput()
	}
	copies := make([]map[int]aig.Edge, len(variants))
	for i, g := range variants {
		copies[i] = map[int]aig.Edge{0: result.Const1()}
		for j, e := range g.Inputs() {
			copies[i][e.NodeID()] = inputs[j]
		}
	}
	var copyEdge func(int, aig.Edge) (aig.Edge, error)
	copyEdge = func(i int, e aig.Edge) (aig.Edge, error) {
		regular := aig.Aig_Regular(e)
		if old, ok := copies[i][e.NodeID()]; ok {
			if e.Complemented() {
				old = old.Not()
			}
			return old, nil
		}
		node := variants[i].Node(regular)
		a, err := copyEdge(i, node.Fanins[0])
		if err != nil {
			return aig.InvalidEdge, err
		}
		var value aig.Edge
		switch node.Kind {
		case aig.Output:
			value = a
		case aig.And:
			b, err := copyEdge(i, node.Fanins[1])
			if err != nil {
				return aig.InvalidEdge, err
			}
			value = result.AND(a, b)
		default:
			return aig.InvalidEdge, fmt.Errorf("dch: unsupported snapshot node %d", e.NodeID())
		}
		copies[i][e.NodeID()] = value
		if e.Complemented() {
			value = value.Not()
		}
		return value, nil
	}
	outputs := make([][]aig.Edge, len(variants))
	for i, g := range variants {
		outputs[i] = g.Outputs()
	}
	for output := 0; output < first.OutputCount(); output++ {
		var driver aig.Edge
		for i := range variants {
			copied, err := copyEdge(i, outputs[i][output])
			if err != nil {
				return nil, err
			}
			if i == 0 {
				driver = copied
			}
		}
		result.AddOutput(driver)
	}
	if err := copyMetadata(result, first); err != nil {
		return nil, err
	}
	return result, nil
}
func Dch_DeriveTotalAig(variants []*aig.Graph) *aig.Graph {
	result, err := DeriveTotalAig(variants)
	if err != nil {
		panic(err)
	}
	return result
}
