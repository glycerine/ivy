package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManCexMinDupWithCubes makes one output for each noninitial cube over
// register inputs and retains the original register transition functions.
func Saig_ManCexMinDupWithCubes(g *aig.Graph, cubes *vec.Vec_Vec_t) *aig.Graph {
	if g == nil || cubes == nil || aig.Aig_ManConstrNum(g) != 0 {
		return nil
	}
	q := aig.NewGraph()
	q.Name = g.Name
	mapping := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
	for _, e := range g.Inputs() {
		mapping[e] = q.NewInput()
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(e aig.Edge) aig.Edge {
		reg := aig.Aig_Regular(e)
		if c, ok := mapping[reg]; ok {
			return aig.Aig_NotCond(c, int32(e&1))
		}
		n := g.Node(reg)
		var c aig.Edge
		switch n.Kind {
		case aig.And:
			c = q.AND(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
		case aig.ExclusiveOr:
			c = q.NewXor(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
		case aig.Buffer, aig.Output:
			c = copyEdge(n.Fanins[0])
		default:
			panic("bmc: invalid cube input")
		}
		mapping[reg] = c
		return aig.Aig_NotCond(c, int32(e&1))
	}
	for level := 1; level < vec.Vec_VecSize(cubes); level++ {
		cube := q.Const1()
		for _, lit := range vec.Vec_IntArray(vec.Vec_VecEntryInt(cubes, level)) {
			if lit < 0 || lit>>1 >= aig.Aig_ManRegNum(g) {
				panic("bmc: invalid cube literal")
			}
			li := aig.Aig_ManLi(g, lit>>1)
			driver := copyEdge(aig.Aig_ObjChild0(g, li))
			cube = q.AND(cube, aig.Aig_NotCond(driver, lit&1))
		}
		q.AddOutput(cube)
	}
	for i := int32(0); i < aig.Aig_ManRegNum(g); i++ {
		q.AddOutput(copyEdge(aig.Aig_ObjChild0(g, aig.Aig_ManLi(g, i))))
	}
	aig.Aig_ManSetRegNum(q, aig.Aig_ManRegNum(g))
	aig.Aig_ManCleanup(q)
	return q
}
