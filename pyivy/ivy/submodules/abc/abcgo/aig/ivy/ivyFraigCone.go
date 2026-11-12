// Isolated cone extraction and exact checking ported from ivyFraig.c.
package ivy

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Ivy_FraigExtractCone_rec(p *Ivy_Man_t, n *Ivy_Obj_t, leaves, nodes *[]int32) {
	if n == nil || n.owner != p {
		panic("ivy: foreign cone node")
	}
	if n.fMarkB != 0 {
		return
	}
	n.fMarkB = 1
	if Ivy_ObjIsConst1(n) != 0 {
		return
	}
	if Ivy_ObjIsPi(n) != 0 {
		*leaves = append(*leaves, n.Id)
		return
	}
	if Ivy_ObjIsNode(n) == 0 {
		panic("ivy: combinational cone expected")
	}
	Ivy_FraigExtractCone_rec(p, Ivy_ObjFanin0(n), leaves, nodes)
	Ivy_FraigExtractCone_rec(p, Ivy_ObjFanin1(n), leaves, nodes)
	*nodes = append(*nodes, n.Id)
}
func Ivy_FraigExtractCone(p *Ivy_Man_t, a, b *Ivy_Obj_t, leaves *[]int32) *aig.Graph {
	if Ivy_ManIsAcyclic(p) == 0 {
		panic("ivy: cyclic cone")
	}
	for _, n := range p.vObjs {
		if n != nil && n.fMarkB != 0 {
			panic("ivy: cone extraction requires unmarked nodes")
		}
	}

	var nodes []int32
	*leaves = (*leaves)[:0]
	// Restore all marks on failure as well as success, without reusing pEquiv.
	defer func() {
		for _, n := range p.vObjs {
			if n != nil {
				n.fMarkB = 0
			}
		}
	}()
	Ivy_FraigExtractCone_rec(p, Ivy_Regular(a), leaves, &nodes)
	Ivy_FraigExtractCone_rec(p, Ivy_Regular(b), leaves, &nodes)
	result := aig.NewGraph()
	mapped := map[int32]aig.Edge{0: 0}
	for _, id := range *leaves {
		mapped[id] = result.NewInput()
	}
	edge := func(n *Ivy_Obj_t) aig.Edge {
		e := mapped[Ivy_Regular(n).Id]
		if Ivy_IsComplement(n) != 0 {
			return e.Not()
		}
		return e
	}
	for _, id := range nodes {
		n := p.vObjs[id]
		if n.Type == IVY_EXOR {
			mapped[id] = result.XOR(edge(n.pFanin0), edge(n.pFanin1))
		} else {
			mapped[id] = result.AND(edge(n.pFanin0), edge(n.pFanin1))
		}
	}
	root := result.XOR(edge(a), edge(b))
	if Ivy_ObjFaninPhase(a)^Ivy_ObjFaninPhase(b) != 0 {
		root = root.Not()
	}
	result.AddOutput(root)
	return result
}
func Ivy_FraigCheckCone(global *Ivy_FraigMan_t, p *Ivy_Man_t, a, b *Ivy_Obj_t, limit int32) int32 {
	var leaves []int32
	graph := Ivy_FraigExtractCone(p, a, b, &leaves)
	solver := bsat.NewSolver()
	nodes := graph.Nodes()
	for range nodes {
		solver.NewVariable()
	}
	solver.AddClause(bsat.Positive(0))
	literal := func(edge aig.Edge) bsat.Literal {
		if edge.Complemented() {
			return bsat.Negative(int32(edge.NodeID()))
		}
		return bsat.Positive(int32(edge.NodeID()))
	}
	for id, n := range nodes {
		output := bsat.Positive(int32(id))
		switch n.Kind {
		case aig.And:
			x, y := literal(n.Fanins[0]), literal(n.Fanins[1])
			solver.AddClause(x, output^1)
			solver.AddClause(y, output^1)
			solver.AddClause(x^1, y^1, output)
		case aig.Buffer, aig.Output:
			x := literal(n.Fanins[0])
			solver.AddClause(x, output^1)
			solver.AddClause(x^1, output)
		case aig.Constant, aig.Input:
		default:
			panic("ivy: unexpected extracted graph operator")
		}
	}
	solver.SetLimits(int64(max(limit, 0)), 0)
	status := solver.Solve(literal(graph.Outputs()[0]))
	if status == bsat.Unsatisfiable {
		return 1
	}
	if status != bsat.Satisfiable {
		return -1
	}
	clear(global.pPatWords)
	for i, input := range graph.Inputs() {
		value, known := solver.Value(int32(input.NodeID()))
		if known && value {
			index := -1
			for k, n := range p.vPis {
				if n.Id == leaves[i] {
					index = k
					break
				}
			}
			if index < 0 {
				panic("ivy: extracted leaf is not a primary input")
			}
			global.pPatWords[index/32] |= 1 << uint(index%32)
		}
	}
	return 0
}

// The optional CUDD checker uses an equivalent reduced ordered BDD here, without
// an external C library. Terminals are 0/1; unique/apply tables canonicalize nodes.
type ivyBddNode struct {
	variable  int32
	low, high uint32
}
type ivyBddKey struct {
	variable  int32
	low, high uint32
}
type ivyBddApplyKey struct {
	op   byte
	a, b uint32
}
type ivyBdd struct {
	nodes  []ivyBddNode
	unique map[ivyBddKey]uint32
	cache  map[ivyBddApplyKey]uint32
}

func (d *ivyBdd) node(variable int32, low, high uint32) uint32 {
	if low == high {
		return low
	}
	key := ivyBddKey{variable, low, high}
	if id, ok := d.unique[key]; ok {
		return id
	}
	id := uint32(len(d.nodes))
	d.nodes = append(d.nodes, ivyBddNode{variable, low, high})
	d.unique[key] = id
	return id
}
func (d *ivyBdd) apply(op byte, a, b uint32) uint32 {
	if a > b {
		a, b = b, a
	}
	if a < 2 && b < 2 {
		if op == '&' {
			return a & b
		}
		return a ^ b
	}
	key := ivyBddApplyKey{op, a, b}
	if value, ok := d.cache[key]; ok {
		return value
	}
	variable := int32(0x7fffffff)
	if a >= 2 {
		variable = d.nodes[a].variable
	}
	if b >= 2 {
		variable = min(variable, d.nodes[b].variable)
	}
	a0, a1, b0, b1 := a, a, b, b
	if a >= 2 && d.nodes[a].variable == variable {
		a0, a1 = d.nodes[a].low, d.nodes[a].high
	}
	if b >= 2 && d.nodes[b].variable == variable {
		b0, b1 = d.nodes[b].low, d.nodes[b].high
	}
	low, high := d.apply(op, a0, b0), d.apply(op, a1, b1)
	result := d.node(variable, low, high)
	d.cache[key] = result
	return result
}
func Ivy_FraigNodesAreEquivBdd(a, b *Ivy_Obj_t) int32 {
	if a == nil || b == nil || Ivy_Regular(a).owner != Ivy_Regular(b).owner {
		panic("ivy: BDD candidates require one manager")
	}
	p := Ivy_Regular(a).owner
	if Ivy_ManIsAcyclic(p) == 0 {
		panic("ivy: cyclic BDD cone")
	}
	d := ivyBdd{nodes: make([]ivyBddNode, 2), unique: map[ivyBddKey]uint32{}, cache: map[ivyBddApplyKey]uint32{}}
	memo := map[*Ivy_Obj_t]uint32{p.pConst1: 1}
	var build func(*Ivy_Obj_t) uint32
	build = func(edge *Ivy_Obj_t) uint32 {
		n := Ivy_Regular(edge)
		value, known := memo[n]
		if !known {
			switch n.Type {
			case IVY_PI:
				value = d.node(n.Id, 0, 1)
			case IVY_AND:
				value = d.apply('&', build(n.pFanin0), build(n.pFanin1))
			case IVY_EXOR:
				value = d.apply('^', build(n.pFanin0), build(n.pFanin1))
			default:
				panic("ivy: unsupported BDD cone")
			}
			memo[n] = value
		}
		if Ivy_IsComplement(edge) != 0 {
			return d.apply('^', value, 1)
		}
		return value
	}
	left, right := build(a), build(b)
	if Ivy_ObjFaninPhase(a)^Ivy_ObjFaninPhase(b) != 0 {
		right = d.apply('^', right, 1)
	}
	return ivyBool(left == right)
}
