package abc

import (
	"fmt"
	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/sat/satoko"
	"strings"
)

// The iso helper uses the importable AIG manager rather than importing GIA's
// command layer. Both retain ordered inputs and two compared output cones.
func Abc_NtkFunctionalIsoGia_rec(graph *aig.Graph, o *Object) aig.Edge {
	if dfsVisited(o) || dfsStop(o) {
		return aig.Edge(o.Temp)
	}
	dfsMark(o)
	args := make([]aig.Edge, len(o.fanins))
	for i, edge := range o.fanins {
		args[i] = Abc_NtkFunctionalIsoGia_rec(graph, edge.Object)
		if edge.Complement {
			args[i] = args[i].Not()
		}
	}
	var value aig.Edge
	switch {
	case o.Kind == ObjectNet || o.IsCO() || o.Barrier:
		if len(args) != 1 {
			panic("invalid transparent iso node")
		}
		value = args[0]
	case o.Data == aigANDMarker:
		if len(args) != 2 {
			panic("invalid AIG iso node")
		}
		value = graph.AND(args[0], args[1])
	default:
		if root, ok := o.Data.(*hop.Hop_Obj_t); ok {
			value = isoHop(graph, root, args)
		} else if root, ok := o.Data.(*core.DdNode); ok {
			value = isoBDD(graph, root, args)
		} else if gate, ok := o.Data.(*mio.Mio_Gate_t); ok {
			value = isoCover(graph, mio.Mio_GateReadSop(gate), args)
		} else if text, ok := o.Data.(string); ok {
			value = isoCover(graph, text, args)
		} else {
			fn := localFunction(o)
			values := make([]bool, len(args))
			var shannon func(int) aig.Edge
			shannon = func(index int) aig.Edge {
				if index == len(args) {
					if fn(values) {
						return 0
					}
					return 1
				}
				values[index] = false
				low := shannon(index + 1)
				values[index] = true
				high := shannon(index + 1)
				return graph.Mux(args[index], high, low)
			}
			value = shannon(0)
		}
	}
	o.Temp = int(value)
	return value
}
func isoCover(graph *aig.Graph, text string, args []aig.Edge) aig.Edge {
	result := aig.Edge(1)
	polarity := -1
	for _, line := range strings.Split(strings.TrimRight(text, "\x00"), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		pattern, output := "", ""
		if len(args) == 0 && len(fields) == 1 {
			output = fields[0]
		} else if len(fields) == 2 {
			pattern, output = fields[0], fields[1]
		} else {
			panic("invalid iso SOP")
		}
		if output == "x" || output == "n" {
			value := aig.Edge(1)
			for _, arg := range args {
				value = graph.XOR(value, arg)
			}
			if output == "n" {
				value = value.Not()
			}
			return value
		}
		if len(pattern) != len(args) || (output != "0" && output != "1") {
			panic("invalid iso SOP arity")
		}
		p := int(output[0] - '0')
		if polarity >= 0 && polarity != p {
			panic("mixed iso SOP polarity")
		}
		polarity = p
		cube := aig.Edge(0)
		for i, c := range pattern {
			switch c {
			case '-':
			case '0':
				cube = graph.AND(cube, args[i].Not())
			case '1':
				cube = graph.AND(cube, args[i])
			default:
				panic("invalid iso SOP literal")
			}
		}
		result = graph.OR(result, cube)
	}
	if polarity == 0 {
		result = result.Not()
	}
	return result
}
func isoReorderSupport(first, second []int32, common bool) ([]int32, []int32) {
	if !common {
		return first, second
	}
	a, b := make(map[int32]bool), make(map[int32]bool)
	for _, id := range first {
		a[id] = true
	}
	for _, id := range second {
		b[id] = true
	}
	var left, right, shared []int32
	for _, id := range first {
		if b[id] {
			shared = append(shared, id)
		} else {
			left = append(left, id)
		}
	}
	for _, id := range second {
		if !a[id] {
			right = append(right, id)
		}
	}
	return append(left, shared...), append(right, shared...)
}
func Abc_NtkFunctionalIsoGia(n *Network, first, second, common int32) *aig.Graph {
	if first < 0 || second < 0 || int(first) >= len(n.cos) || int(second) >= len(n.cos) {
		return nil
	}
	support1, support2 := Abc_NtkNodeSupportInt(n, first), Abc_NtkNodeSupportInt(n, second)
	if len(support1) != len(support2) {
		return nil
	}
	if _, err := n.Topological([]*Object{n.cos[first], n.cos[second]}); err != nil {
		return nil
	}
	support1, support2 = isoReorderSupport(support1, support2, common != 0)
	graph := aig.NewGraph()
	graph.Name = n.Name
	graph.Spec = n.Spec
	inputs := make([]aig.Edge, len(support1))
	for i, index := range support1 {
		inputs[i] = graph.NewInput()
		n.cis[index].Temp = int(inputs[i])
	}
	n.constant.Temp = 0
	Abc_NtkIncrementTravId(n)
	left := Abc_NtkFunctionalIsoGia_rec(graph, n.cos[first])
	for i, index := range support2 {
		n.cis[index].Temp = int(inputs[i])
	}
	Abc_NtkIncrementTravId(n)
	right := Abc_NtkFunctionalIsoGia_rec(graph, n.cos[second])
	graph.AddOutput(left)
	graph.AddOutput(right)
	return graph
}
func isoEquivalent(graph *aig.Graph) (bool, error) {
	model, err := graph.AigerModel()
	if err != nil {
		return false, err
	}
	if len(model.Outputs) != 2 {
		return false, fmt.Errorf("iso graph must have two outputs")
	}
	solver := satoko.New()
	solver.SetVariables(int(model.MaxVar) + 1)
	solver.AddClause([]int{1})
	for _, gate := range model.ANDs {
		z, a, b := int(gate.Literal), int(gate.Left), int(gate.Right)
		for _, clause := range [][]int{{z ^ 1, a}, {z ^ 1, b}, {z, a ^ 1, b ^ 1}} {
			if !solver.AddClause(clause) {
				return true, nil
			}
		}
	}
	left, right := int(model.Outputs[0]), int(model.Outputs[1])
	if !solver.AddClause([]int{left, right}) || !solver.AddClause([]int{left ^ 1, right ^ 1}) {
		return true, nil
	}
	return solver.Solve() == satoko.SATOKO_UNSAT, nil
}
func Abc_NtkFunctionalIsoInt(n *Network, first, second, common int32) int32 {
	graph := Abc_NtkFunctionalIsoGia(n, first, second, common)
	if graph == nil {
		return 0
	}
	equal, err := isoEquivalent(graph)
	return abcBool(err == nil && equal)
}
func Abc_NtkFunctionalIso(n *Network, first, second, common int32) int32 {
	return Abc_NtkFunctionalIsoInt(n, first, second, common)
}

func isoHop(graph *aig.Graph, root *hop.Hop_Obj_t, args []aig.Edge) aig.Edge {
	memo := map[*hop.Hop_Obj_t]aig.Edge{}
	var visit func(*hop.Hop_Obj_t) aig.Edge
	visit = func(edge *hop.Hop_Obj_t) aig.Edge {
		node := hop.Hop_Regular(edge)
		value, ok := memo[node]
		if !ok {
			switch {
			case hop.Hop_ObjIsConst1(node) != 0:
				value = 0
			case hop.Hop_ObjIsPi(node) != 0:
				value = args[node.PioNum]
			default:
				a, b := visit(hop.Hop_ObjChild0(node)), visit(hop.Hop_ObjChild1(node))
				if hop.Hop_ObjIsExor(node) != 0 {
					value = graph.XOR(a, b)
				} else {
					value = graph.AND(a, b)
				}
			}
			memo[node] = value
		}
		if hop.Hop_IsComplement(edge) != 0 {
			value = value.Not()
		}
		return value
	}
	return visit(root)
}
func isoBDD(graph *aig.Graph, root *core.DdNode, args []aig.Edge) aig.Edge {
	memo := map[*core.DdNode]aig.Edge{}
	var visit func(*core.DdNode) aig.Edge
	visit = func(edge *core.DdNode) aig.Edge {
		node := core.Cudd_Regular(edge)
		value, ok := memo[node]
		if !ok {
			if core.Cudd_IsConstant(node) {
				value = 0
			} else {
				value = graph.Mux(args[node.Index], visit(node.T), visit(node.E))
			}
			memo[node] = value
		}
		if core.Cudd_IsComplement(edge) {
			value = value.Not()
		}
		return value
	}
	return visit(root)
}
