package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManDupWithCubes duplicates the next-state logic and replaces property
// outputs with conjunctions of the specified next-state register literals.
func Saig_ManDupWithCubes(g *aig.Graph, cubes *vec.Vec_Vec_t) *aig.Graph {
	if aig.Aig_ManConstrNum(g) != 0 {
		panic("cube duplication expects no constraints")
	}
	result := aig.NewGraph()
	result.Name = g.Name
	copies := make([]aig.Edge, len(g.Nodes()))
	copies[g.Const1().NodeID()] = result.Const1()
	for _, e := range g.Inputs() {
		copies[e.NodeID()] = result.NewInput()
	}
	copyEdge := func(e aig.Edge) aig.Edge {
		v := copies[e.NodeID()]
		if e.Complemented() {
			v = v.Not()
		}
		return v
	}
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		switch n.Kind {
		case aig.And:
			copies[e.NodeID()] = result.AND(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
		case aig.Buffer:
			copies[e.NodeID()] = copyEdge(n.Fanins[0])
		case aig.ExclusiveOr:
			panic("cube duplication expects an AND-inverter graph")
		}
	}
	for i := 0; i < vec.Vec_VecSize(cubes); i++ {
		cube := result.Const1()
		for _, lit := range vec.Vec_IntArray(vec.Vec_VecEntryInt(cubes, i)) {
			li := saig.Saig_ManLi(g, lit>>1)
			v := copyEdge(g.Node(li).Fanins[0])
			if lit&1 != 0 {
				v = v.Not()
			}
			cube = result.AND(cube, v)
		}
		result.AddOutput(cube)
	}
	for i := int32(0); i < saig.Saig_ManRegNum(g); i++ {
		li := saig.Saig_ManLi(g, i)
		result.AddOutput(copyEdge(g.Node(li).Fanins[0]))
	}
	aig.Aig_ManSetRegNum(result, saig.Saig_ManRegNum(g))
	return result
}
