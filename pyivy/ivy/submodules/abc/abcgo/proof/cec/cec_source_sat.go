package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
)

// cecSourceSAT ports the frontier allocation and supergate/MUX clauses in
// Cec_CnfNodeAddToSolver. Variable zero is unused and variable one is const0.
type cecSourceSAT struct {
	graph                                             *Gia_Man_t
	pars                                              *Cec_ParSat_t
	sat                                               *bsat.Solver
	ids                                               []int32
	since                                             int32
	calls, satCount, unsatCount, undecCount, recycles int32
}

func newCecSourceSAT(g *Gia_Man_t, p *Cec_ParSat_t) *cecSourceSAT {
	return &cecSourceSAT{graph: g, pars: p, ids: make([]int32, g.ObjectCount())}
}
func (s *cecSourceSAT) reset() {
	s.sat = bsat.NewSolver()
	s.sat.NewVariable()
	s.sat.NewVariable()
	if !s.sat.AddClause(3) {
		panic("cec: constant clause")
	}
	s.ids = make([]int32, s.graph.ObjectCount())
	s.ids[0] = 1
	s.since = 0
	s.recycles++
}
func (s *cecSourceSAT) literal(lit int32) int32 {
	id := lit >> 1
	sign := lit & 1
	if s.pars.fPolarFlip != 0 {
		sign ^= gia.Gia_ObjPhase(gia.Gia_ManObj(s.graph, id))
	}
	return 2*s.ids[id] | sign
}
func (s *cecSourceSAT) addClause(lits ...int32) {
	if !s.sat.AddClause(lits...) {
		panic("cec: inconsistent source CNF")
	}
}
func (s *cecSourceSAT) addNode(root int32) {
	if s.sat == nil {
		s.reset()
	}
	if s.ids[root] != 0 {
		return
	}
	var frontier []int32
	add := func(id int32) {
		if s.ids[id] != 0 {
			return
		}
		o := gia.Gia_ManObj(s.graph, id)
		s.ids[id] = s.sat.NewVariable()
		if gia.Gia_ObjIsAnd(o) != 0 {
			frontier = append(frontier, id)
		} else if gia.Gia_ObjIsCi(o) == 0 {
			panic("cec: source SAT requires AND-inverter graph")
		}
	}
	add(root)
	for i := 0; i < len(frontier); i++ {
		id := frontier[i]
		obj := gia.Gia_ManObj(s.graph, id)
		if gia.Gia_ObjIsMuxType(obj) != 0 {
			a, b := gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)
			raw := []*Gia_Obj_t{gia.Gia_ObjFanin0(a), gia.Gia_ObjFanin0(b), gia.Gia_ObjFanin1(a), gia.Gia_ObjFanin1(b)}
			for _, o := range raw {
				add(gia.Gia_ObjId(s.graph, o))
			}
			var t, e int32
			control := gia.Gia_ObjRecognizeMuxLits(s.graph, obj, &t, &e)
			x, y, z, f := s.literal(control), s.literal(t), s.literal(e), s.literal(id<<1)
			s.addClause(x^1, y^1, f)
			s.addClause(x^1, y, f^1)
			s.addClause(x, z^1, f)
			s.addClause(x, z, f^1)
			if t>>1 != e>>1 {
				s.addClause(y, z, f^1)
				s.addClause(y^1, z^1, f)
			}
		} else {
			var leaves []int32
			seen := make(map[int32]bool)
			var collect func(int32, bool)
			collect = func(lit int32, first bool) {
				oid := lit >> 1
				o := gia.Gia_ManObj(s.graph, oid)
				if lit&1 != 0 || gia.Gia_ObjIsCi(o) != 0 || (!first && gia.Gia_ObjValue(o) > 1) || gia.Gia_ObjIsMuxType(o) != 0 {
					if !seen[lit] {
						seen[lit] = true
						leaves = append(leaves, lit)
					}
					return
				}
				if gia.Gia_ObjIsAnd(o) == 0 {
					panic("cec: invalid source supergate")
				}
				collect(gia.Gia_ObjFaninLit0(o, oid), false)
				collect(gia.Gia_ObjFaninLit1(o, oid), false)
			}
			collect(id<<1, true)
			for _, lit := range leaves {
				add(lit >> 1)
			}
			f := s.literal(id << 1)
			long := make([]int32, 0, len(leaves)+1)
			for _, lit := range leaves {
				a := s.literal(lit)
				s.addClause(a, f^1)
				long = append(long, a^1)
			}
			s.addClause(append(long, f)...)
		}
	}
}
