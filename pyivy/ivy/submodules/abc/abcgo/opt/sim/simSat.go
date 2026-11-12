package sim

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/satoko"
)

// cofactorDifference checks F(a=0) != F(a=1), or F(a=0,b=1) !=
// F(a=1,b=0). Unchanged inputs share SAT variables between the two copies.
// No resource limit is used: an unresolved query is never called UNSAT.
func cofactorDifference(g *aig.Graph, out, a, b int) (bool, []uint32) {
	inputs := g.Inputs()
	outputs := g.Outputs()
	if out < 0 || out >= len(outputs) || a < 0 || a >= len(inputs) || b < -1 || b >= len(inputs) || b == a {
		panic("invalid cofactor indices")
	}
	s := satoko.New()
	next := len(inputs) + 1
	s.SetVariables(next)
	s.AddClause([]int{0}) // variable 0 is true, matching AIG constant 1.
	pi := make(map[int]int, len(inputs))
	for i, e := range inputs {
		pi[e.NodeID()] = i
	}
	add := func(c ...int) { s.AddClause(c) }
	encode := func(second bool) int {
		memo := map[int]int{0: 0}
		var rec func(aig.Edge) int
		rec = func(e aig.Edge) int {
			id := e.NodeID()
			v, ok := memo[id]
			if !ok {
				n := g.Node(e)
				if n.Kind == aig.Input {
					i := pi[id]
					v = (i + 1) << 1
					if i == a {
						v = 1
						if second {
							v = 0
						}
					}
					if i == b {
						v = 0
						if second {
							v = 1
						}
					}
				} else {
					x := rec(n.Fanins[0])
					if n.Kind == aig.Buffer || n.Kind == aig.Output {
						v = x
					} else {
						y := rec(n.Fanins[1])
						v = next << 1
						next++
						s.SetVariables(next)
						switch n.Kind {
						case aig.And:
							add(x^1, y^1, v)
							add(x, v^1)
							add(y, v^1)
						case aig.ExclusiveOr:
							add(x, y, v^1)
							add(x^1, y^1, v^1)
							add(x, y^1, v)
							add(x^1, y, v)
						default:
							panic("unsupported cofactor gate")
						}
					}
				}
				memo[id] = v
			}
			if e.Complemented() {
				v ^= 1
			}
			return v
		}
		return rec(outputs[out])
	}
	x, y := encode(false), encode(true)
	add(x, y)
	add(x^1, y^1)
	status := s.SolveAssumptions(nil)
	if status == satoko.SATOKO_UNSAT {
		return false, nil
	}
	if status != satoko.SATOKO_SAT {
		panic("unbounded support/symmetry SAT query unresolved")
	}
	pattern := make([]uint32, SIM_NUM_WORDS(len(inputs)))
	for i := range inputs {
		if s.ModelValue(i+1) == 1 {
			Sim_SetBit(pattern, i)
		}
	}
	// Source symmetry simulation expects a base point with both pair bits one.
	Sim_SetBit(pattern, a)
	if b >= 0 {
		Sim_SetBit(pattern, b)
	}
	return true, pattern
}
