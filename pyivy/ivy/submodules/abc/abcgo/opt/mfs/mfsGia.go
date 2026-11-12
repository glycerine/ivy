package mfs

import "github.com/glycerine/abcgo/aig/aig"

// The optional GIA backend is excluded from upstream module.make. Its two-copy
// resubstitution miter has a native AIG representation here, using the same
// projection-XOR SAT contracts rather than a second pointer-tagged graph.
func Gia_ManCreateResubMiter(window *aig.Graph, divisors int) *aig.Graph {
	if divisors < 0 || window.OutputCount() < divisors+2 {
		panic("mfs: invalid resub miter interface")
	}
	g := aig.NewGraph()
	var copies [2]map[int]aig.Edge
	for k := 0; k < 2; k++ {
		memo := map[int]aig.Edge{0: g.Const1()}
		for _, e := range window.Inputs() {
			memo[e.NodeID()] = g.NewInput()
		}
		var rec func(aig.Edge) aig.Edge
		rec = func(e aig.Edge) aig.Edge {
			x, ok := memo[e.NodeID()]
			if !ok {
				n := window.Node(e)
				a := rec(n.Fanins[0])
				switch n.Kind {
				case aig.Buffer, aig.Output:
					x = a
				case aig.And:
					x = g.AND(a, rec(n.Fanins[1]))
				case aig.ExclusiveOr:
					x = g.XOR(a, rec(n.Fanins[1]))
				default:
					panic("mfs: invalid miter gate")
				}
				memo[e.NodeID()] = x
			}
			if e.Complemented() {
				x = x.Not()
			}
			return x
		}
		for _, e := range window.Outputs() {
			rec(e)
		}
		copies[k] = memo
	}
	outs := window.Outputs()
	read := func(k, i int) aig.Edge { return copies[k][outs[i].NodeID()] }
	care := g.Const1()
	firstDiv := len(outs) - divisors
	target := firstDiv - 1
	for i := 0; i < target; i++ {
		care = g.AND(care, g.AND(read(0, i), read(1, i)))
	}
	care = g.AND(care, g.AND(read(0, target), read(1, target).Not()))
	g.AddOutput(care)
	for i := firstDiv; i < len(outs); i++ {
		g.AddOutput(g.XOR(read(0, i), read(1, i)))
	}
	return g
}
func Abc_NtkMfsConstructGia(p *Mfs_Man_t) *aig.Graph {
	return Gia_ManCreateResubMiter(p.Window, len(p.Divisors))
}
func Abc_NtkMfsTryResubOnceGia(p *Mfs_Man_t, candidates []int) int {
	return Abc_NtkMfsTryResubOnce(p, candidates)
}
