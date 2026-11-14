package res

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
)

type resNetworkMeta struct {
	name, spec  string
	inputNames  []string
	outputNames []string
}

// ResubstituteNetwork converts a strashed ABC network to the graph resub path,
// runs exact cut-local rewriting, and returns a rebuilt ABC AIG. The input
// network is not mutated.
func ResubstituteNetwork(n *abc.Network, params GraphParams) (*abc.Network, GraphStats, error) {
	var stats GraphStats
	g, err := ResGraphFromNetwork(n)
	if err != nil {
		return nil, stats, err
	}
	stats, err = ResubstituteGraph(g, params)
	if err != nil {
		return nil, stats, err
	}
	out, err := ResNetworkFromGraph(g)
	if err != nil {
		return nil, stats, err
	}
	return out, stats, nil
}

// ResGraphFromNetwork copies a combinational, structurally hashed ABC AIG into
// aig.Graph while preserving CI/CO order, names, network name, and spec.
func ResGraphFromNetwork(n *abc.Network) (*aig.Graph, error) {
	if n == nil {
		return nil, fmt.Errorf("res: nil ABC network")
	}
	if n.Function != uint8(abc.ABC_FUNC_AIG) {
		return nil, fmt.Errorf("res: expected an AIG network, function kind is %d", n.Function)
	}
	if n.Type != uint8(abc.ABC_NTK_STRASH) {
		return nil, fmt.Errorf("res: expected a strashed network, type kind is %d", n.Type)
	}
	manager, ok := n.FunctionManager.(*abc.Abc_Aig_t)
	if !ok || abc.Abc_AigCheck(manager) == 0 {
		return nil, fmt.Errorf("res: input network has no valid AIG hash manager")
	}
	if len(n.ObjectsOfKind(abc.ObjectLatch)) != 0 {
		return nil, fmt.Errorf("res: sequential AIG networks are not supported")
	}
	graph := aig.NewGraph()
	graph.Name, graph.Spec = n.Name, n.Spec
	meta := &resNetworkMeta{name: n.Name, spec: n.Spec}
	graph.Data = meta
	edges := make(map[*abc.Object]aig.Edge, len(n.Objects()))
	constant := n.One().Object
	edges[constant] = graph.Const1()
	inputs := n.Inputs()
	for _, input := range inputs {
		if input == nil || input == constant || input.Kind != abc.ObjectPI {
			return nil, fmt.Errorf("res: invalid ABC input object")
		}
		edges[input] = graph.NewInput()
		meta.inputNames = append(meta.inputNames, input.Name())
	}

	objects, err := n.Topological(n.Objects())
	if err != nil {
		return nil, fmt.Errorf("res: ABC network is not acyclic: %w", err)
	}
	for _, object := range objects {
		if object == nil || object == constant || object.IsCI() || object.IsCO() {
			continue
		}
		if object.Kind != abc.ObjectNode || abc.Abc_AigNodeIsAnd(object) == 0 {
			return nil, fmt.Errorf("res: unsupported object %d kind=%d in AIG network", object.ID, object.Kind)
		}
		fanins := object.Fanins()
		if len(fanins) != 2 {
			return nil, fmt.Errorf("res: AIG node %d has %d fanins", object.ID, len(fanins))
		}
		a, ok := resGraphEdge(edges, fanins[0])
		if !ok {
			return nil, fmt.Errorf("res: missing graph copy for node %d fanin 0", object.ID)
		}
		b, ok := resGraphEdge(edges, fanins[1])
		if !ok {
			return nil, fmt.Errorf("res: missing graph copy for node %d fanin 1", object.ID)
		}
		edges[object] = graph.AND(a, b)
	}

	for _, output := range n.Outputs() {
		if output.Kind != abc.ObjectPO {
			return nil, fmt.Errorf("res: only primary outputs are supported, output %d kind is %d", output.ID, output.Kind)
		}
		fanins := output.Fanins()
		if len(fanins) != 1 {
			return nil, fmt.Errorf("res: ABC output %d has %d fanins", output.ID, len(fanins))
		}
		driver, ok := resGraphEdge(edges, fanins[0])
		if !ok {
			return nil, fmt.Errorf("res: missing graph copy for output %d", output.ID)
		}
		graph.AddOutput(driver)
		meta.outputNames = append(meta.outputNames, output.Name())
	}
	if err := graph.Check(); err != nil {
		return nil, fmt.Errorf("res: generated graph is invalid: %w", err)
	}
	return graph, nil
}

func resGraphEdge(copies map[*abc.Object]aig.Edge, edge abc.Edge) (aig.Edge, bool) {
	value, ok := copies[edge.Object]
	if edge.Complement {
		value = value.Not()
	}
	return value, ok
}

// ResNetworkFromGraph builds a fresh strashed ABC network from graph nodes.
func ResNetworkFromGraph(graph *aig.Graph) (*abc.Network, error) {
	if graph == nil {
		return nil, fmt.Errorf("res: nil AIG graph")
	}
	if err := graph.Check(); err != nil {
		return nil, fmt.Errorf("res: invalid AIG graph: %w", err)
	}
	meta, _ := graph.Data.(*resNetworkMeta)
	name, spec := graph.Name, graph.Spec
	if meta != nil {
		name, spec = meta.name, meta.spec
	}
	network := abc.NewNetwork(name)
	network.Type = uint8(abc.ABC_NTK_STRASH)
	network.Function = uint8(abc.ABC_FUNC_AIG)
	network.Spec = spec
	manager := abc.Abc_AigAlloc(network)
	network.FunctionManager = manager

	nodes := graph.Nodes()
	copies := make([]abc.Edge, len(nodes))
	copies[0] = network.One()
	for i, input := range graph.Inputs() {
		pi := network.NewObject(abc.ObjectPI)
		if meta != nil && i < len(meta.inputNames) && meta.inputNames[i] != "" {
			pi.SetName(meta.inputNames[i])
		}
		copies[input.NodeID()] = pi.Signal()
	}
	visiting := make([]bool, len(nodes))
	var convert func(aig.Edge) (abc.Edge, error)
	convert = func(edge aig.Edge) (abc.Edge, error) {
		id := edge.NodeID()
		if id < 0 || id >= len(nodes) || !graph.ValidEdge(aig.Edge(id<<1)) {
			return abc.Edge{}, fmt.Errorf("res: invalid graph edge %d", id)
		}
		if copies[id].Object != nil {
			result := copies[id]
			if edge.Complemented() {
				result = result.Not()
			}
			return result, nil
		}
		if visiting[id] {
			return abc.Edge{}, fmt.Errorf("res: graph cycle at node %d", id)
		}
		visiting[id] = true
		node := nodes[id]
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			a, err := convert(node.Fanins[0])
			if err != nil {
				return abc.Edge{}, err
			}
			b, err := convert(node.Fanins[1])
			if err != nil {
				return abc.Edge{}, err
			}
			if node.Kind == aig.ExclusiveOr {
				copies[id] = abc.Abc_AigXor(manager, a, b)
			} else {
				copies[id] = abc.Abc_AigAnd(manager, a, b)
			}
		case aig.Buffer:
			value, err := convert(node.Fanins[0])
			if err != nil {
				return abc.Edge{}, err
			}
			copies[id] = value
		default:
			return abc.Edge{}, fmt.Errorf("res: unsupported graph node kind %d at %d", node.Kind, id)
		}
		visiting[id] = false
		result := copies[id]
		if edge.Complemented() {
			result = result.Not()
		}
		return result, nil
	}

	for i, output := range graph.Outputs() {
		driver, err := convert(graph.Node(output).Fanins[0])
		if err != nil {
			return nil, fmt.Errorf("res: output %d: %w", i, err)
		}
		po := network.NewObject(abc.ObjectPO)
		if err := network.AddFanin(po, driver); err != nil {
			return nil, fmt.Errorf("res: add output %d: %w", i, err)
		}
		if meta != nil && i < len(meta.outputNames) && meta.outputNames[i] != "" {
			po.SetName(meta.outputNames[i])
		}
	}
	if err := network.Check(); err != nil {
		return nil, fmt.Errorf("res: rebuilt ABC network is invalid: %w", err)
	}
	if abc.Abc_AigCheck(manager) == 0 {
		return nil, fmt.Errorf("res: rebuilt ABC network violates strashing invariants")
	}
	return network, nil
}
