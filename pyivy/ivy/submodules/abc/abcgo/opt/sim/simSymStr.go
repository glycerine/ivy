package sim

import "github.com/glycerine/abcgo/aig/aig"

type Pair struct{ A, B int }

func pair(a, b int) Pair {
	if a > b {
		a, b = b, a
	}
	return Pair{a, b}
}

type pairSet map[Pair]bool

// Sim_SymmsStructCompute propagates symmetries through implication supergates.
// Pairs survive only if every remaining child either omits both variables or
// already proves their symmetry, as in the source compatibility rules.
func Sim_SymmsStructCompute(g *aig.Graph, matrices []Matrix, functional Info) {
	structural := Sim_ComputeStrSupp(g)
	pi := Sim_SymmsCreateMap(g)
	known := make(map[int]pairSet)
	for _, e := range graphOrder(g) {
		n := g.Node(e)
		result := pairSet{}
		if n.Kind == aig.Buffer || n.Kind == aig.Output {
			for p := range known[n.Fanins[0].NodeID()] {
				result[p] = true
			}
		} else if n.Kind == aig.And {
			leaves := Sim_SymmsBalanceCollect_rec(g, e)
			var groups [2][]int
			var other []aig.Edge
			for _, leaf := range leaves {
				if index := pi[leaf.NodeID()]; index >= 0 {
					k := 0
					if leaf.Complemented() {
						k = 1
					}
					groups[k] = append(groups[k], index)
				} else {
					other = append(other, leaf)
				}
			}
			compatible := func(p Pair) bool {
				for _, child := range other {
					a, b := Sim_HasBit(structural[child.NodeID()], p.A), Sim_HasBit(structural[child.NodeID()], p.B)
					if a != b || a && !known[child.NodeID()][p] {
						return false
					}
				}
				return true
			}
			for _, group := range groups {
				for i, a := range group {
					for _, b := range group[i+1:] {
						p := pair(a, b)
						if compatible(p) {
							result[p] = true
						}
					}
				}
			}
			for _, child := range other {
				for p := range known[child.NodeID()] {
					if !compatible(p) {
						continue
					}
					ok := true
					for _, group := range groups {
						hasA, hasB := false, false
						for _, v := range group {
							hasA = hasA || v == p.A
							hasB = hasB || v == p.B
						}
						ok = ok && hasA == hasB
					}
					if ok {
						result[p] = true
					}
				}
			}
		}
		// XOR is a boundary for the original AND-only structural proof. SAT still
		// resolves all its functional symmetry pairs without a false structural claim.
		known[e.NodeID()] = result
	}
	for o, e := range g.Outputs() {
		for i := range g.Inputs() {
			matrices[o].Set(i, i)
		}
		for p := range known[e.NodeID()] {
			if Sim_HasBit(functional[o], p.A) && Sim_HasBit(functional[o], p.B) {
				matrices[o].Set(p.A, p.B)
			}
		}
	}
}
func Sim_SymmsBalanceCollect_rec(g *aig.Graph, root aig.Edge) []aig.Edge {
	var result []aig.Edge
	seen := map[aig.Edge]bool{}
	var rec func(aig.Edge)
	rec = func(e aig.Edge) {
		n := g.Node(e)
		if !e.Complemented() && n.Kind == aig.And {
			rec(n.Fanins[0])
			rec(n.Fanins[1])
			return
		}
		if !seen[e] {
			seen[e] = true
			result = append(result, e)
		}
	}
	rec(root)
	return result
}
func Sim_SymmsCreateMap(g *aig.Graph) []int {
	r := make([]int, int(aig.Aig_ManObjNumMax(g)))
	for i := range r {
		r[i] = -1
	}
	for i, e := range g.Inputs() {
		r[e.NodeID()] = i
	}
	return r
}
