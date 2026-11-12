package mfs

import "github.com/glycerine/abcgo/aig/aig"

func truthGraph(truth []uint32, vars int) *aig.Graph {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, vars)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	g.AddOutput(composeTruth(g, truth, inputs))
	return g
}
func composeTruth(g *aig.Graph, truth []uint32, inputs []aig.Edge) aig.Edge {
	var rec func(int, int) aig.Edge
	rec = func(level, base int) aig.Edge {
		if level == 0 {
			if truthBit(truth, base) {
				return g.Const1()
			}
			return g.Const0()
		}
		lo := rec(level-1, base)
		hi := rec(level-1, base|(1<<uint(level-1)))
		return g.Mux(inputs[level-1], hi, lo)
	}
	return rec(len(inputs), 0)
}

// Abc_NtkConstructAig constructs both original and target-flipped window copies.
// Output zero is observability; external care clauses follow, then the target
// and divisors (resubstitution) or target fanins (care enumeration).
func Abc_NtkConstructAig(p *Mfs_Man_t, target int) *aig.Graph {
	g := aig.NewGraph()
	base, flipped := map[int]aig.Edge{0: g.Const1()}, map[int]aig.Edge{0: g.Const1()}
	for _, id := range p.Support {
		e := g.NewInput()
		base[id] = e
		flipped[id] = e
	}
	var rec func(int, bool) aig.Edge
	rec = func(id int, flip bool) aig.Edge {
		memo := base
		if flip {
			memo = flipped
		}
		if e, ok := memo[id]; ok {
			return e
		}
		v := p.Network.Nodes[id]
		args := make([]aig.Edge, len(v.Fanins))
		for i, f := range v.Fanins {
			args[i] = rec(f, flip)
		}
		e := composeTruth(g, v.Truth, args)
		if flip && id == target {
			e = e.Not()
		}
		memo[id] = e
		return e
	}
	observe := g.Const0()
	for _, id := range p.Roots {
		observe = g.OR(observe, g.XOR(rec(id, false), rec(id, true)))
	}
	g.AddOutput(observe)
	p.CareOutputs = 0
	if care := p.Network.Care; care != nil {
		pi := map[int]aig.Edge{}
		positions := map[int]int{}
		for i, id := range p.Network.Inputs {
			positions[id] = i
		}
		cis := care.Inputs()
		for _, id := range p.Support {
			pi[cis[positions[id]].NodeID()] = base[id]
		}
		memo := map[int]aig.Edge{0: g.Const1()}
		available := map[int]bool{0: true}
		var copyCare func(aig.Edge) (aig.Edge, bool)
		copyCare = func(e aig.Edge) (aig.Edge, bool) {
			id := e.NodeID()
			if x, ok := pi[id]; ok {
				if e.Complemented() {
					x = x.Not()
				}
				return x, true
			}
			if x, ok := memo[id]; ok {
				if e.Complemented() {
					x = x.Not()
				}
				return x, available[id]
			}
			n := care.Node(e)
			if n.Kind == aig.Input {
				return aig.InvalidEdge, false
			}
			x, ok := copyCare(n.Fanins[0])
			if !ok {
				return aig.InvalidEdge, false
			}
			v := x
			if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
				y, ok := copyCare(n.Fanins[1])
				if !ok {
					return aig.InvalidEdge, false
				}
				if n.Kind == aig.And {
					v = g.AND(x, y)
				} else {
					v = g.XOR(x, y)
				}
			}
			memo[id] = v
			available[id] = true
			if e.Complemented() {
				v = v.Not()
			}
			return v, true
		}
		for _, co := range care.Outputs() {
			if e, ok := copyCare(co); ok {
				g.AddOutput(e)
				p.CareOutputs++
			}
		}
	}
	if p.Params.Resub {
		g.AddOutput(rec(target, false))
		for _, id := range p.Divisors {
			g.AddOutput(rec(id, false))
		}
	} else {
		for _, id := range p.Network.Nodes[target].Fanins {
			g.AddOutput(rec(id, false))
		}
	}
	p.Window, p.Values = g, base
	return g
}

// ToAIG preserves the LUT interface and register partition.
func (n *Network) ToAIG() (*aig.Graph, error) {
	order, err := n.topology()
	if err != nil {
		return nil, err
	}
	g := aig.NewGraph()
	g.Name, g.Spec = n.Name, n.Spec
	values := map[int]aig.Edge{0: g.Const1()}
	for _, id := range n.Inputs {
		values[id] = g.NewInput()
	}
	for _, id := range order {
		v := n.Nodes[id]
		if id == 0 || v.Input {
			continue
		}
		args := make([]aig.Edge, len(v.Fanins))
		for i, f := range v.Fanins {
			args[i] = values[f]
		}
		values[id] = composeTruth(g, v.Truth, args)
	}
	for _, e := range n.Outputs {
		x := values[e.ID()]
		if e&1 != 0 {
			x = x.Not()
		}
		g.AddOutput(x)
	}
	aig.Aig_ManSetRegNum(g, int32(n.Registers))
	return g, nil
}

// FromAIG retains original gates as local 2-input functions and unary buffers.
func FromAIG(g *aig.Graph) (*Network, error) {
	if err := g.Check(); err != nil {
		return nil, err
	}
	n := NewNetwork()
	n.Name, n.Spec = g.Name, g.Spec
	n.Registers = int(aig.Aig_ManRegNum(g))
	mapped := map[int]int{0: 0}
	for _, e := range g.Inputs() {
		mapped[e.NodeID()] = n.AddInput()
	}
	var rec func(aig.Edge) int
	rec = func(e aig.Edge) int {
		if id, ok := mapped[e.NodeID()]; ok {
			return id
		}
		v := g.Node(e)
		if v.Kind == aig.Output {
			e = v.Fanins[0]
			id := rec(e)
			truth := uint32(0xaaaaaaaa)
			if e.Complemented() {
				truth = ^truth
			}
			return n.AddNode([]int{id}, []uint32{truth})
		}
		fs := []aig.Edge{v.Fanins[0]}
		if v.Kind == aig.And || v.Kind == aig.ExclusiveOr {
			fs = append(fs, v.Fanins[1])
		}
		ids := make([]int, len(fs))
		for i, f := range fs {
			ids[i] = rec(f)
		}
		var truth uint32
		for a := 0; a < 32; a++ {
			x := a&1 != 0
			if fs[0].Complemented() {
				x = !x
			}
			value := x
			if len(fs) == 2 {
				y := a&2 != 0
				if fs[1].Complemented() {
					y = !y
				}
				if v.Kind == aig.And {
					value = x && y
				} else {
					value = x != y
				}
			}
			if value {
				truth |= 1 << uint(a)
			}
		}
		id := n.AddNode(ids, []uint32{truth})
		mapped[e.NodeID()] = id
		return id
	}
	// The source optimizes all internal objects, including output-unreachable
	// logic. Preserve these cones rather than silently sweeping the adapter.
	for _, e := range g.LiveEdges() {
		kind := g.Node(e).Kind
		if kind == aig.And || kind == aig.ExclusiveOr || kind == aig.Buffer {
			rec(e)
		}
	}
	for _, co := range g.Outputs() {
		e := g.Node(co).Fanins[0]
		id := rec(e)
		lit := Literal(id << 1)
		if e.Complemented() {
			lit = lit.Not()
		}
		n.AddOutput(lit)
	}
	return n, nil
}
