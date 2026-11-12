package abc

import "github.com/glycerine/abcgo/aig/aig"

// Abc_NtkClpOneGia_rec copies a strashed cone into a structurally hashed graph.
func Abc_NtkClpOneGia_rec(g *aig.Graph, node *Object) aig.Edge {
	if Abc_NodeIsTravIdCurrent(node) != 0 || len(node.fanins) == 0 || node.IsCI() {
		return aig.Edge(node.Temp)
	}
	if node.Kind != ObjectNode || len(node.fanins) != 2 {
		panic("abc: expected strashed AND node")
	}
	Abc_NodeSetTravIdCurrent(node)
	a := Abc_NtkClpOneGia_rec(g, node.fanins[0].Object)
	b := Abc_NtkClpOneGia_rec(g, node.fanins[1].Object)
	if node.fanins[0].Complement {
		a = a.Not()
	}
	if node.fanins[1].Complement {
		b = b.Not()
	}
	result := g.AND(a, b)
	node.Temp = int(result)
	return result
}

// Abc_NtkStrashToGia copies output cones from a strashed ABC network.
func Abc_NtkStrashToGia(n *Network) *aig.Graph {
	if n == nil || Abc_NtkIsStrash(n) == 0 {
		panic("abc: expected strashed network")
	}
	for _, obj := range n.Objects() {
		obj.Temp = -1
	}
	g := aig.NewGraph()
	g.Name, g.Spec = n.Name, n.Spec
	n.constant.Temp = int(g.Const1())
	for _, ci := range n.cis {
		ci.Temp = int(g.NewInput())
	}
	Abc_NtkIncrementTravId(n)
	for _, co := range n.cos {
		if len(co.fanins) != 1 {
			panic("abc: invalid output")
		}
		edge := co.fanins[0]
		lit := Abc_NtkClpOneGia_rec(g, edge.Object)
		if edge.Complement {
			lit = lit.Not()
		}
		g.AddOutput(lit)
	}
	return g
}
