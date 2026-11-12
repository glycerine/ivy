package dec

// Network is the boundary between decomposition and a concrete ABC, Hop or Ivy
// network. Each implementation owns its signal representation and hashing.
type Network interface {
	ConstantOne() any
	Complement(any) any
	And(any, any) any
}
type VariableNetwork interface {
	Network
	Variable(int) any
}
type UnhashedNetwork interface {
	Network
	CreateAnd(any, any) any
}
type ReplacementNetwork interface {
	Network
	Replace(old, replacement any, updateLevels bool)
}
type CountingNetwork interface {
	Network
	LookupAnd(any, any) any
	SameRegular(any, any) bool
	Level(any) int
	InDeletionCone(any) bool
}

func decNetworkEdge(n Network, value any, compl uint32) any {
	if compl != 0 {
		return n.Complement(value)
	}
	return value
}
func graphToNetwork(n Network, g *Dec_Graph_t, and func(any, any) any) any {
	if g.fConst != 0 {
		return decNetworkEdge(n, n.ConstantOne(), g.eRoot.fCompl)
	}
	if Dec_GraphIsVar(g) != 0 {
		return decNetworkEdge(n, g.pNodes[g.eRoot.Node].pFunc, g.eRoot.fCompl)
	}
	for i := g.nLeaves; i < g.nSize; i++ {
		p := g.pNodes[i]
		a := decNetworkEdge(n, g.pNodes[p.eEdge0.Node].pFunc, p.eEdge0.fCompl)
		b := decNetworkEdge(n, g.pNodes[p.eEdge1.Node].pFunc, p.eEdge1.fCompl)
		p.pFunc = and(a, b)
	}
	return decNetworkEdge(n, g.pNodes[g.eRoot.Node].pFunc, g.eRoot.fCompl)
}
func Dec_GraphToNetwork(n Network, g *Dec_Graph_t) any { return graphToNetwork(n, g, n.And) }
func Dec_GraphToNetworkNoStrash(n UnhashedNetwork, g *Dec_Graph_t) any {
	return graphToNetwork(n, g, n.CreateAnd)
}
func Dec_GraphToNetworkAig(n Network, g *Dec_Graph_t) any { return Dec_GraphToNetwork(n, g) }
func Dec_GraphToNetworkIvy(n Network, g *Dec_Graph_t) any { return Dec_GraphToNetwork(n, g) }
func Dec_GraphToAig(n Network, g *Dec_Graph_t, fanins []any) any {
	if len(fanins) != int(g.nLeaves) {
		panic("dec: fanin count")
	}
	for i, v := range fanins {
		g.pNodes[i].pFunc = v
	}
	return Dec_GraphToNetwork(n, g)
}
func Dec_SopToAig(n Network, sop string, fanins []any) any {
	g := Dec_Factor(sop)
	return Dec_GraphToAig(n, g, fanins)
}
func Dec_GraphFactorSop(n VariableNetwork, sop string) any {
	g := Dec_Factor(sop)
	for i := int32(0); i < g.nLeaves; i++ {
		g.pNodes[i].pFunc = n.Variable(int(i))
	}
	return Dec_GraphToNetwork(n, g)
}
func Dec_GraphUpdateNetwork(n ReplacementNetwork, root any, g *Dec_Graph_t, updateLevels int32, gain int32) {
	n.Replace(root, Dec_GraphToNetwork(n, g), updateLevels != 0)
}

// Count prospective added nodes without modifying the target network.
func Dec_GraphToNetworkCount(n CountingNetwork, root any, g *Dec_Graph_t, nodeMax, levelMax int32) int32 {
	if g.fConst != 0 || Dec_GraphIsVar(g) != 0 {
		return 0
	}
	for i := int32(0); i < g.nLeaves; i++ {
		g.pNodes[i].Level = uint32(n.Level(g.pNodes[i].pFunc))
	}
	count := int32(0)
	for i := g.nLeaves; i < g.nSize; i++ {
		p := g.pNodes[i]
		a, b := g.pNodes[p.eEdge0.Node], g.pNodes[p.eEdge1.Node]
		var signal any
		av, bv := a.pFunc, b.pFunc
		if av != nil && bv != nil {
			av = decNetworkEdge(n, av, p.eEdge0.fCompl)
			bv = decNetworkEdge(n, bv, p.eEdge1.fCompl)
			signal = n.LookupAnd(av, bv)
			if signal != nil && n.SameRegular(signal, root) {
				return -1
			}
		}
		if signal == nil || n.InDeletionCone(signal) {
			count++
			if count > nodeMax {
				return -1
			}
		}
		level := int(a.Level)
		if int(b.Level) > level {
			level = int(b.Level)
		}
		level++
		if signal != nil {
			if n.SameRegular(signal, n.ConstantOne()) {
				level = 0
			} else if n.SameRegular(signal, av) {
				level = n.Level(av)
			} else if n.SameRegular(signal, bv) {
				level = n.Level(bv)
			}
		}
		if level > int(levelMax) {
			return -1
		}
		p.Level = uint32(level)
		p.pFunc = signal
	}
	return count
}
