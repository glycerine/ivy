package bsat

import "github.com/glycerine/abcgo/aig/aig"

// AigInterpolationTest constructs two copies of a circuit's CNF, equates their
// inputs, and assumes unequal outputs. Its interpolant is the circuit relation
// over the original input variables and one additional output variable.
func AigInterpolationTest(graph *aig.Graph) *aig.Graph {
	if graph.OutputCount() != 1 || aig.Aig_ManRegNum(graph) != 0 {
		panic("interpolation test requires a combinational graph with one output")
	}
	if err := graph.Check(); err != nil {
		panic(err)
	}
	nodes := graph.Nodes()
	width := int32(len(nodes))
	s := sat_solver2_new()
	sat_solver2_setnvars(s, 2*width+1)
	globals := []int32{}
	for _, edge := range graph.Inputs() {
		globals = append(globals, int32(edge.NodeID()))
	}
	output := int32(graph.Outputs()[0].NodeID())
	globals = append(globals, output)
	for _, edge := range graph.LiveEdges() {
		v := int32(edge.NodeID())
		kind := graph.Node(edge).Kind
		if kind != aig.Input && kind != aig.Output {
			var_set_partA(s, v, 1)
		}
	}
	for copyIndex := int32(0); copyIndex < 2; copyIndex++ {
		shift := copyIndex * width
		mark := 1 - copyIndex
		for _, edge := range graph.LiveEdges() {
			v := int32(edge.NodeID())
			node := graph.Node(edge)
			switch node.Kind {
			case aig.Constant, aig.Input:
				continue
			case aig.And:
				a, b := node.Fanins[0], node.Fanins[1]
				sat_solver2_add_and(s, shift+v, shift+int32(a.NodeID()), shift+int32(b.NodeID()), int32(a&1), int32(b&1), mark, -1)
			case aig.ExclusiveOr:
				a, b := node.Fanins[0], node.Fanins[1]
				sat_solver2_add_xor(s, shift+v, shift+int32(a.NodeID()), shift+int32(b.NodeID()), int32((a^b)&1), mark, -1)
			case aig.Buffer, aig.Output:
				a := node.Fanins[0]
				sat_solver2_add_buffer(s, shift+v, shift+int32(a.NodeID()), int32(a&1), mark, -1)
			default:
				panic("unsupported AIG node in interpolation test")
			}
		}
	}
	for _, v := range globals[:len(globals)-1] {
		sat_solver2_add_buffer(s, v, width+v, 0, 0, -1)
	}
	sat_solver2_add_xor(s, output, width+output, 2*width, 0, 0, -1)
	s.pInt2 = Int2_ManStart(s, globals, int32(len(globals)))
	// Add constant constraints after all gate clauses so partition markings exist
	// before propagation creates any derived unit-clause annotations.
	sat_solver2_add_const(s, 0, 0, 1, -1)
	sat_solver2_add_const(s, width, 0, 0, -1)
	if sat_solver2_solve(s, []lit{toLit(2 * width)}, 0, 0, 0, 0) != l_False {
		panic("equal circuit copies did not refute unequal outputs")
	}
	return Int2_ManReadInterpolant(s)
}

// Gia_ManInterTest retains ABC's entry-point name; Go graphs use the shared AIG
// representation, avoiding a dependency on the separate GIA algorithm package.
func Gia_ManInterTest(graph *aig.Graph) *aig.Graph { return AigInterpolationTest(graph) }
