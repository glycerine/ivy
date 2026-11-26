package core

import "github.com/glycerine/abcgo/sat/bsat"

// SourceCNFNode exposes the live GIA metadata used by Cec_CnfNodeAddToSolver.
// And includes all source AND-encoded nodes; callers must supply an ordinary
// AND-inverter graph, expanding real XOR/MUX gates before SAT construction.
type SourceCNFNode struct {
	Fanins          [2]int32
	And, Input, Mux bool
	Phase, Value    int32
}

// SourceCNFMux records recognized AND-inverter MUX pins and the original
// grandchildren order used to allocate frontier variables in C.
type SourceCNFMux struct {
	Control, Then, Else int32
	Raw                 [4]int32
}

// SourceCNFGraph avoids a dependency on the GIA package. Node and Mux read
// the source graph's live metadata; its topology stays fixed while SAT runs.
type SourceCNFGraph struct {
	Nodes int
	Node  func(int32) SourceCNFNode
	Mux   func(int32) SourceCNFMux
}

// SourceCNF is the ported CEC frontier/supergate/MUX CNF builder, backed by
// BSAT. Variable zero is unused and variable one is constant zero.
type SourceCNF struct {
	graph     SourceCNFGraph
	polarFlip func() bool
	sat       *bsat.Solver
	ids       []int32
	next      int32 // nSatVars: the next SAT variable to assign
}

func NewSourceCNF(g SourceCNFGraph, polarFlip func() bool) *SourceCNF {
	return &SourceCNF{graph: g, polarFlip: polarFlip, ids: make([]int32, g.Nodes)}
}

// Solver and IDs expose borrowed source state for the source-shaped SAT
// wrappers and exact CNF tests. Reset invalidates the previous state.
func (s *SourceCNF) Solver() *bsat.Solver { return s.sat }
func (s *SourceCNF) IDs() []int32         { return s.ids }

// Reset is Cec_ManSatSolverRecycle: a solver sized for 1000 variables,
// variable 0 unused and variable 1 the constant-zero node.
func (s *SourceCNF) Reset() {
	s.sat = bsat.NewSolver()
	s.sat.SetNumVars(1000)
	if !s.sat.AddClause(3) {
		panic("cec: constant clause")
	}
	s.ids = make([]int32, s.graph.Nodes)
	s.ids[0] = 1
	s.next = 2
}

// NumSatVars is nSatVars, the number of SAT variables assigned so far.
func (s *SourceCNF) NumSatVars() int32 { return s.next }

func (s *SourceCNF) Literal(lit int32) int32 {
	id, sign := lit>>1, lit&1
	if s.polarFlip != nil && s.polarFlip() {
		sign ^= s.graph.Node(id).Phase
	}
	return 2*s.ids[id] | sign
}

func (s *SourceCNF) AddClause(lits ...int32) {
	if !s.sat.AddClause(lits...) {
		panic("cec: inconsistent source CNF")
	}
}

func (s *SourceCNF) AddNode(root int32) {
	if s.sat == nil {
		s.Reset()
	}
	if s.ids[root] != 0 {
		return
	}
	if s.graph.Node(root).Input {
		// as Cec_CnfNodeAddToSolver for a CI: number it and size the solver
		s.ids[root] = s.next
		s.next++
		s.sat.SetNumVars(s.next)
		return
	}
	var frontier []int32
	add := func(id int32) {
		if s.ids[id] != 0 {
			return
		}
		o := s.graph.Node(id)
		s.ids[id] = s.next
		s.next++
		if o.And {
			frontier = append(frontier, id)
		} else if !o.Input {
			panic("cec: source SAT requires AND-inverter graph")
		}
	}
	add(root)
	for i := 0; i < len(frontier); i++ {
		id := frontier[i]
		if s.graph.Node(id).Mux {
			mux := s.graph.Mux(id)
			for _, child := range mux.Raw {
				add(child)
			}
			x, y, z, f := s.Literal(mux.Control), s.Literal(mux.Then), s.Literal(mux.Else), s.Literal(id<<1)
			s.AddClause(x^1, y^1, f)
			s.AddClause(x^1, y, f^1)
			s.AddClause(x, z^1, f)
			s.AddClause(x, z, f^1)
			if mux.Then>>1 != mux.Else>>1 {
				s.AddClause(y, z, f^1)
				s.AddClause(y^1, z^1, f)
			}
		} else {
			var leaves []int32
			seen := make(map[int32]bool)
			var collect func(int32, bool)
			collect = func(lit int32, first bool) {
				oid := lit >> 1
				o := s.graph.Node(oid)
				if lit&1 != 0 || o.Input || (!first && o.Value > 1) || o.Mux {
					if !seen[lit] {
						seen[lit] = true
						leaves = append(leaves, lit)
					}
					return
				}
				if !o.And {
					panic("cec: invalid source supergate")
				}
				collect(o.Fanins[0], false)
				collect(o.Fanins[1], false)
			}
			collect(id<<1, true)
			for _, lit := range leaves {
				add(lit >> 1)
			}
			f := s.Literal(id << 1)
			long := make([]int32, 0, len(leaves)+1)
			for _, lit := range leaves {
				a := s.Literal(lit)
				s.AddClause(a, f^1)
				long = append(long, a^1)
			}
			s.AddClause(append(long, f)...)
		}
	}
}
