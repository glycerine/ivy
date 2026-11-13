package dch

import (
	"fmt"

	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/satoko"
)

func newManager(g *aig.Graph, parameters *Params) (*Manager, error) {
	if g == nil {
		return nil, fmt.Errorf("dch: nil AIG")
	}
	if err := g.Check(); err != nil {
		return nil, fmt.Errorf("dch: %w", err)
	}
	params := DefaultParams()
	if parameters != nil {
		params = *parameters
	}
	if params.Words <= 0 || params.ConflictLimit < 0 || params.SatVarMax < 0 || params.CallsRecycle < 0 {
		return nil, fmt.Errorf("dch: invalid simulation/SAT parameters")
	}
	if params.Words > int(^uint(0)>>1)/int(aig.Aig_ManObjNumMax(g)) {
		return nil, fmt.Errorf("dch: simulation allocation size overflows int")
	}
	order, err := graphOrder(g)
	if err != nil {
		return nil, err
	}
	aig.Aig_ManSetPhase(g)
	m := &Manager{
		graph: g, satGraph: g, params: params, order: order,
		proven: map[int]aig.Edge{}, blocked: map[int]bool{},
		fanouts: map[int][]int{}, ranks: map[int]int{},
		lastValues: make([]bool, int(aig.Aig_ManObjNumMax(g))),
	}
	for rank, edge := range order {
		id := edge.NodeID()
		n := g.Node(edge)
		m.ranks[id] = rank
		m.lastValues[id] = n.Phase
		count := 0
		switch n.Kind {
		case aig.Buffer, aig.Output:
			count = 1
		case aig.And, aig.ExclusiveOr:
			count = 2
		}
		for _, fanin := range n.Fanins[:count] {
			m.fanouts[fanin.NodeID()] = append(m.fanouts[fanin.NodeID()], id)
		}
	}
	return m, nil
}
func (m *Manager) recycleSolver() {
	m.solver = satoko.New()
	m.variables = map[int]int{}
	m.callsSince = 0
	m.stats.Recycles++
}
func graphOrder(g *aig.Graph) ([]aig.Edge, error) {
	colors := make(map[int]uint8)
	var order []aig.Edge
	var visit func(aig.Edge) error
	visit = func(edge aig.Edge) error {
		node := aig.Aig_Regular(edge)
		id := node.NodeID()
		if colors[id] == 2 {
			return nil
		}
		if colors[id] == 1 {
			return fmt.Errorf("dch: cyclic AIG at %d", id)
		}
		colors[id] = 1
		n := g.Node(node)
		switch n.Kind {
		case aig.Constant, aig.Input:
		case aig.Buffer, aig.Output:
			if err := visit(n.Fanins[0]); err != nil {
				return err
			}
		case aig.And, aig.ExclusiveOr:
			for _, child := range n.Fanins {
				if err := visit(child); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("dch: unsupported AIG node kind at %d", id)
		}
		colors[id] = 2
		order = append(order, node)
		return nil
	}
	for _, node := range g.LiveEdges() {
		if err := visit(node); err != nil {
			return nil, err
		}
	}
	return order, nil
}
