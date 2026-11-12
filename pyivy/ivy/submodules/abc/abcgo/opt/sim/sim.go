// Package sim computes circuit supports, variable symmetries and word-parallel
// combinational/sequential simulation. CI/CO order includes trailing registers.
package sim

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

type Info = [][]uint32

// Matrix is a symmetric relation on CI indices; diagonal entries are allowed.
type Matrix [][]uint32

func NewMatrix(n int) Matrix       { return Matrix(Sim_UtilInfoAlloc(n, SIM_NUM_WORDS(n), true)) }
func (m Matrix) Has(a, b int) bool { return Sim_HasBit(m[a], b) }
func (m Matrix) Set(a, b int)      { Sim_SetBit(m[a], b); Sim_SetBit(m[b], a) }

type Sim_Pat_t struct {
	Input, Output int
	Data          []uint32
}
type Sim_Man_t struct {
	Graph                  *aig.Graph
	Inputs, Outputs        []aig.Edge
	Order                  []aig.Edge
	Words                  int
	Sim0, Sim1             Info
	Structural, Functional Info
	Targets                [][]int
	Patterns               []*Sim_Pat_t
	Input                  int
	SATRuns, SAT, UNSAT    int
	rng                    *aig.AigRandomSource
}
type Sym_Man_t struct {
	Graph                             *aig.Graph
	Inputs, Outputs, Order            []aig.Edge
	Words                             int
	Sim                               Info
	Functional                        Info
	Supports                          [][]int
	Symmetric, NonSymmetric           []Matrix
	PairsTotal, PairsSym, PairsNonSym []int
	Output, Var1, Var2                int
	SATRuns, SAT, UNSAT               int
	rng                               *aig.AigRandomSource
}

func SIM_NUM_WORDS(n int) int {
	if n < 0 {
		panic("negative bit count")
	}
	words := n / 32
	if n&31 != 0 {
		words++
	}
	return words
}
func SIM_LAST_BITS(n int) int {
	if n&31 == 0 {
		return 32
	}
	return n & 31
}
func Sim_SetBit(p []uint32, i int)      { p[i>>5] |= uint32(1) << uint(i&31) }
func Sim_XorBit(p []uint32, i int)      { p[i>>5] ^= uint32(1) << uint(i&31) }
func Sim_HasBit(p []uint32, i int) bool { return p[i>>5]&(uint32(1)<<uint(i&31)) != 0 }

// graphOrder handles rewiring-created forward references and dangling gates.
// It never borrows graph marks, making independent analyses race safe.
func graphOrder(g *aig.Graph) []aig.Edge {
	if g == nil {
		panic("sim: nil graph")
	}
	if err := g.Check(); err != nil {
		panic(fmt.Errorf("sim: %w", err))
	}
	seen := make([]bool, int(aig.Aig_ManObjNumMax(g)))
	var order []aig.Edge
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		id := e.NodeID()
		if seen[id] {
			return
		}
		seen[id] = true
		n := g.Node(e)
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			visit(n.Fanins[0])
			visit(n.Fanins[1])
		case aig.Buffer, aig.Output:
			visit(n.Fanins[0])
		}
		order = append(order, aig.Edge(id<<1))
	}
	for _, e := range g.LiveEdges() {
		visit(e)
	}
	return order
}
func newRandom() *aig.AigRandomSource { return aig.NewAigRandomSource() }
func fillRandom(p []uint32, r *aig.AigRandomSource) {
	for i := range p {
		p[i] = r.Uint32()
	}
}
