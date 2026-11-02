package abc

import (
	"fmt"
	"math"
)

// Port of src/sat/bsat2/SimpSolver.h and SimpSolver.cpp (namespace Minisat).
// SimpSolver derives from Solver: it embeds MinisatSolver and re-declares the
// overridden methods (NewVar, AddClause*, Solve*, ToDimacs*, GarbageCollect, ...).

// MinisatSimpSolverElimLt is SimpSolver::ElimLt.
type MinisatSimpSolverElimLt struct {
	n_occ *MinisatVec[int] // const vec<int>&
}

// NewMinisatSimpSolverElimLt is the explicit constructor ElimLt(const vec<int>& no).
// C: explicit ElimLt(const vec<int>& no)
func NewMinisatSimpSolverElimLt(no *MinisatVec[int]) MinisatSimpSolverElimLt {
	return MinisatSimpSolverElimLt{n_occ: no}
}

// cost is ElimLt::cost.
// C: uint64_t cost(Var x) const
func (e MinisatSimpSolverElimLt) cost(x MinisatVar) uint64 {
	return uint64(*e.n_occ.At(MinisatToIntLit(MinisatMkLit(x, false)))) * uint64(*e.n_occ.At(MinisatToIntLit(MinisatLitNot(MinisatMkLit(x, false)))))
}

// Call is ElimLt::operator().
// C: bool operator()(Var x, Var y) const
func (e MinisatSimpSolverElimLt) Call(x int, y int) bool { return e.cost(x) < e.cost(y) }

// MinisatSimpSolverClauseDeleted is SimpSolver::ClauseDeleted.
type MinisatSimpSolverClauseDeleted struct {
	ca *MinisatClauseAllocator // const ClauseAllocator&
}

// NewMinisatSimpSolverClauseDeleted is the explicit constructor ClauseDeleted(const ClauseAllocator& _ca).
// C: explicit ClauseDeleted(const ClauseAllocator& _ca)
func NewMinisatSimpSolverClauseDeleted(ca *MinisatClauseAllocator) MinisatSimpSolverClauseDeleted {
	return MinisatSimpSolverClauseDeleted{ca: ca}
}

// Call is ClauseDeleted::operator().
// C: bool operator()(const CRef& cr) const
func (d MinisatSimpSolverClauseDeleted) Call(cr MinisatCRef) bool { return d.ca.At(cr).Mark() == 1 }

// minisatSimpSolverCat is `static const char* _cat = "SIMP"`.
const minisatSimpSolverCat = "SIMP"

// Command-line options of the simplifier (static objects in C++).
var (
	opt_use_asymm_minisat         = NewMinisatBoolOption(minisatSimpSolverCat, "asymm", "Shrink clauses by asymmetric branching.", false)
	opt_use_rcheck_minisat        = NewMinisatBoolOption(minisatSimpSolverCat, "rcheck", "Check if a clause is already implied. (costly)", false)
	opt_use_elim_minisat          = NewMinisatBoolOption(minisatSimpSolverCat, "elim", "Perform variable elimination.", true)
	opt_grow_minisat              = NewMinisatIntOption(minisatSimpSolverCat, "grow", "Allow a variable elimination step to grow by a number of clauses.", 0, NewMinisatIntRange(math.MinInt32, math.MaxInt32))
	opt_clause_lim_minisat        = NewMinisatIntOption(minisatSimpSolverCat, "cl-lim", "Variables are not eliminated if it produces a resolvent with a length above this limit. -1 means no limit", 20, NewMinisatIntRange(-1, math.MaxInt32))
	opt_subsumption_lim_minisat   = NewMinisatIntOption(minisatSimpSolverCat, "sub-lim", "Do not check if subsumption against a clause larger than this. -1 means no limit.", 1000, NewMinisatIntRange(-1, math.MaxInt32))
	opt_simp_garbage_frac_minisat = NewMinisatDoubleOption(minisatSimpSolverCat, "simp-gc-frac", "The fraction of wasted memory allowed before a garbage collection is triggered during simplification.", 0.5, NewMinisatDoubleRange(0, false, math.Inf(1), false))
)

// MinisatSimpSolver is Minisat::SimpSolver : public Solver.
type MinisatSimpSolver struct {
	MinisatSolver

	Grow              int     // allow a variable elimination step to grow by a number of clauses (default zero)
	Clause_lim        int     // variables are not eliminated if it produces a resolvent longer than this limit
	Subsumption_lim   int     // do not check subsumption against a clause larger than this (-1 = no limit)
	Simp_garbage_frac float64 // limit for when to issue a GC during simplification
	Use_asymm         bool    // shrink clauses by asymmetric branching
	Use_rcheck        bool    // check if a clause is already implied
	Use_elim          bool    // perform variable elimination
	Merges            int
	Asymm_lits        int
	Eliminated_vars   int

	// Solver state (protected):
	elimorder          int
	use_simplification bool
	elimclauses        MinisatVec[uint32]
	touched            MinisatVec[int8]
	occurs             MinisatOccLists[int, MinisatCRef, MinisatSimpSolverClauseDeleted]
	n_occ              MinisatVec[int]
	elim_heap          MinisatHeap[MinisatSimpSolverElimLt]
	subsumption_queue  MinisatQueue[MinisatCRef]
	frozen             MinisatVec[int8]
	eliminated         MinisatVec[int8]
	bwdsub_assigns     int
	n_touched          int
	bwdsub_tmpunit     MinisatCRef
}

// NewMinisatSimpSolver is SimpSolver::SimpSolver().
// C: SimpSolver::SimpSolver()
func NewMinisatSimpSolver() *MinisatSimpSolver {
	s := &MinisatSimpSolver{}
	s.initMinisatSolver()
	s.virtGC = s.GarbageCollect // GarbageCollect is virtual
	s.Grow = int(opt_grow_minisat.Value())
	s.Clause_lim = int(opt_clause_lim_minisat.Value())
	s.Subsumption_lim = int(opt_subsumption_lim_minisat.Value())
	s.Simp_garbage_frac = opt_simp_garbage_frac_minisat.Value()
	s.Use_asymm = opt_use_asymm_minisat.Value()
	s.Use_rcheck = opt_use_rcheck_minisat.Value()
	s.Use_elim = opt_use_elim_minisat.Value()
	s.Merges = 0
	s.Asymm_lits = 0
	s.Eliminated_vars = 0
	s.elimorder = 1
	s.use_simplification = true
	s.occurs = *NewMinisatOccLists[int, MinisatCRef, MinisatSimpSolverClauseDeleted](NewMinisatSimpSolverClauseDeleted(&s.ca))
	s.elim_heap = *NewMinisatHeap(NewMinisatSimpSolverElimLt(&s.n_occ))
	s.bwdsub_assigns = 0
	s.n_touched = 0
	s.subsumption_queue = *NewMinisatQueue[MinisatCRef]()

	dummy := NewMinisatVecPad(1, MinisatLitUndef)
	s.ca.ExtraClauseField = true // NOTE: must happen before allocating the dummy clause below.
	s.bwdsub_tmpunit = s.ca.Alloc(dummy, false)
	s.remove_satisfied = false
	return s
}

// Destroy is SimpSolver::~SimpSolver().
// C: SimpSolver::~SimpSolver()
func (s *MinisatSimpSolver) Destroy() {}

// NewVar is SimpSolver::newVar (override; polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *MinisatSimpSolver) NewVar(sign bool, dvar bool) MinisatVar {
	v := s.MinisatSolver.NewVar(sign, dvar)

	s.frozen.PushElem(0)
	s.eliminated.PushElem(0)

	if s.use_simplification {
		s.n_occ.PushElem(0)
		s.n_occ.PushElem(0)
		s.occurs.Init(v)
		s.touched.PushElem(0)
		s.elim_heap.Insert(v)
	}
	return v
}

// IsEliminated is SimpSolver::isEliminated.
// C: bool isEliminated(Var v) const
func (s *MinisatSimpSolver) IsEliminated(v MinisatVar) bool { return *s.eliminated.At(v) != 0 }

// updateElimHeap is SimpSolver::updateElimHeap (protected).
// C: void updateElimHeap(Var v)
func (s *MinisatSimpSolver) updateElimHeap(v MinisatVar) {
	if !s.use_simplification {
		panic("MinisatSimpSolver.updateElimHeap: !use_simplification")
	}
	if s.elim_heap.InHeap(v) || (*s.frozen.At(v) == 0 && !s.IsEliminated(v) && s.Value(v) == MinisatLUndef) {
		s.elim_heap.Update(v)
	}
}

// SetFrozen is SimpSolver::setFrozen (a frozen variable is not eliminated).
// C: void setFrozen(Var v, bool b)
func (s *MinisatSimpSolver) SetFrozen(v MinisatVar, b bool) {
	if b {
		*s.frozen.At(v) = 1
	} else {
		*s.frozen.At(v) = 0
	}
	if s.use_simplification && !b {
		s.updateElimHeap(v)
	}
}

// solve_ is SimpSolver::solve_ (protected; hides Solver::solve_).
// C: lbool solve_(bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) solve_(doSimp bool, turnOffSimp bool) MinisatLbool {
	var extraFrozen MinisatVec[MinisatVar]
	result := MinisatLTrue

	doSimp = doSimp && s.use_simplification

	if doSimp {
		// Assumptions must be temporarily frozen to run variable elimination:
		for i := 0; i < s.assumptions.Size(); i++ {
			v := MinisatLitVar(*s.assumptions.At(i))

			// If an assumption has been eliminated, remember it.
			if s.IsEliminated(v) {
				panic("MinisatSimpSolver.solve_: assumption has been eliminated")
			}

			if *s.frozen.At(v) == 0 {
				// Freeze and store.
				s.SetFrozen(v, true)
				extraFrozen.PushElem(v)
			}
		}

		result = NewMinisatLboolB(s.Eliminate(turnOffSimp))
	}

	if result == MinisatLTrue {
		result = s.MinisatSolver.solve_()
	} else if s.Verbosity >= 1 {
		fmt.Printf("===============================================================================\n")
	}

	if result == MinisatLTrue {
		s.extendModel()
	}

	if doSimp {
		// Unfreeze the assumptions that were frozen:
		for i := 0; i < extraFrozen.Size(); i++ {
			s.SetFrozen(*extraFrozen.At(i), false)
		}
	}

	return result
}

// AddClause_ is SimpSolver::addClause_.
// C: bool addClause_(vec<Lit>& ps)
func (s *MinisatSimpSolver) AddClause_(ps *MinisatVec[MinisatLit]) bool {
	for i := 0; i < ps.Size(); i++ { // #ifndef NDEBUG
		if s.IsEliminated(MinisatLitVar(*ps.At(i))) {
			panic("MinisatSimpSolver.AddClause_: literal of an eliminated variable")
		}
	}

	nclauses := s.clauses.Size()

	if s.Use_rcheck && s.implied(ps) {
		return true
	}

	if !s.MinisatSolver.AddClause_(ps) {
		return false
	}

	if s.use_simplification && s.clauses.Size() == nclauses+1 {
		cr := *s.clauses.Last()
		c := s.ca.At(cr)

		// NOTE: the clause is added to the queue immediately and then
		// again during 'gatherTouchedClauses()'. If nothing happens
		// in between, it will only be checked once. Otherwise, it may
		// be checked twice unnecessarily. This is an unfortunate
		// consequence of how backward subsumption is used to mimic
		// forward subsumption.
		s.subsumption_queue.Insert(cr)
		for i := 0; i < c.Size(); i++ {
			s.occurs.At(MinisatLitVar(*c.At(i))).PushElem(cr)
			*s.n_occ.At(MinisatToIntLit(*c.At(i)))++
			*s.touched.At(MinisatLitVar(*c.At(i))) = 1
			s.n_touched++
			if s.elim_heap.InHeap(MinisatLitVar(*c.At(i))) {
				s.elim_heap.Increase(MinisatLitVar(*c.At(i)))
			}
		}
	}

	return true
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *MinisatSimpSolver) AddClause(ps *MinisatVec[MinisatLit]) bool {
	ps.CopyTo(&s.add_tmp)
	return s.AddClause_(&s.add_tmp)
}

// AddEmptyClause is SimpSolver::addEmptyClause.
// C: bool addEmptyClause()
func (s *MinisatSimpSolver) AddEmptyClause() bool {
	s.add_tmp.Clear(false)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *MinisatSimpSolver) AddClauseUnit(p MinisatLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *MinisatSimpSolver) AddClauseBinary(p MinisatLit, q MinisatLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *MinisatSimpSolver) AddClauseTernary(p MinisatLit, q MinisatLit, r MinisatLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	s.add_tmp.PushElem(r)
	return s.AddClause_(&s.add_tmp)
}

// removeClause is SimpSolver::removeClause (protected; hides Solver::removeClause).
// C: void removeClause(CRef cr)
func (s *MinisatSimpSolver) removeClause(cr MinisatCRef) {
	c := s.ca.At(cr)

	if s.use_simplification {
		for i := 0; i < c.Size(); i++ {
			*s.n_occ.At(MinisatToIntLit(*c.At(i)))--
			s.updateElimHeap(MinisatLitVar(*c.At(i)))
			s.occurs.Smudge(MinisatLitVar(*c.At(i)))
		}
	}

	s.MinisatSolver.removeClause(cr)
}

// strengthenClause is SimpSolver::strengthenClause (protected).
// C: bool strengthenClause(CRef cr, Lit l)
func (s *MinisatSimpSolver) strengthenClause(cr MinisatCRef, l MinisatLit) bool {
	c := s.ca.At(cr)
	if s.decisionLevel() != 0 || !s.use_simplification {
		panic("MinisatSimpSolver.strengthenClause: precondition")
	}

	// FIX: this is too inefficient but would be nice to have (properly implemented)
	// if (!find(subsumption_queue, &c))
	s.subsumption_queue.Insert(cr)

	if c.Size() == 2 {
		s.removeClause(cr)
		c.Strengthen(l)
	} else {
		s.detachClause(cr, true)
		c.Strengthen(l)
		s.attachClause(cr)
		minisatRemove[MinisatCRef](s.occurs.At(MinisatLitVar(l)), cr, minisatNe[MinisatCRef])
		*s.n_occ.At(MinisatToIntLit(l))--
		s.updateElimHeap(MinisatLitVar(l))
	}

	if c.Size() == 1 {
		return s.enqueue(*c.At(0), MinisatCRefUndef) && s.propagate() == MinisatCRefUndef
	}
	return true
}

// merge is the overload merge(..., vec<Lit>& out_clause) (protected): returns
// false if the clause is always satisfied ('out_clause' should not be used).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, vec<Lit>& out_clause)
func (s *MinisatSimpSolver) merge(_ps *MinisatClause, _qs *MinisatClause, v MinisatVar, outClause *MinisatVec[MinisatLit]) bool {
	s.Merges++
	outClause.Clear(false)

	psSmallest := _ps.Size() < _qs.Size()
	ps, qs := _ps, _qs
	if psSmallest {
		ps, qs = _qs, _ps
	}
	var i, j int

	for i = 0; i < qs.Size(); i++ {
		if MinisatLitVar(*qs.At(i)) != v {
			found := false
			for j = 0; j < ps.Size(); j++ {
				if MinisatLitVar(*ps.At(j)) == MinisatLitVar(*qs.At(i)) {
					if *ps.At(j) == MinisatLitNot(*qs.At(i)) {
						return false
					}
					found = true // goto next
					break
				}
			}
			if !found {
				outClause.PushElem(*qs.At(i))
			}
		}
	}

	for i = 0; i < ps.Size(); i++ {
		if MinisatLitVar(*ps.At(i)) != v {
			outClause.PushElem(*ps.At(i))
		}
	}

	return true
}

// mergeSize is the overload merge(..., int& size) (protected): returns false if
// the clause is always satisfied.
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, int& size)
func (s *MinisatSimpSolver) mergeSize(_ps *MinisatClause, _qs *MinisatClause, v MinisatVar, size *int) bool {
	s.Merges++

	psSmallest := _ps.Size() < _qs.Size()
	ps, qs := _ps, _qs
	if psSmallest {
		ps, qs = _qs, _ps
	}
	ps_ := ps.Lits()
	qs_ := qs.Lits()

	*size = ps.Size() - 1

	for i := 0; i < qs.Size(); i++ {
		if MinisatLitVar(qs_[i]) != v {
			found := false
			for j := 0; j < ps.Size(); j++ {
				if MinisatLitVar(ps_[j]) == MinisatLitVar(qs_[i]) {
					if ps_[j] == MinisatLitNot(qs_[i]) {
						return false
					}
					found = true // goto next
					break
				}
			}
			if !found {
				*size++
			}
		}
	}

	return true
}

// gatherTouchedClauses is SimpSolver::gatherTouchedClauses (protected).
// C: void gatherTouchedClauses()
func (s *MinisatSimpSolver) gatherTouchedClauses() {
	if s.n_touched == 0 {
		return
	}

	var i, j int
	for i = 0; i < s.subsumption_queue.Size(); i++ {
		if s.ca.At(*s.subsumption_queue.At(i)).Mark() == 0 {
			s.ca.At(*s.subsumption_queue.At(i)).SetMark(2)
		}
	}

	for i = 0; i < s.touched.Size(); i++ {
		if *s.touched.At(i) != 0 {
			cs := s.occurs.Lookup(i)
			for j = 0; j < cs.Size(); j++ {
				if s.ca.At(*cs.At(j)).Mark() == 0 {
					s.subsumption_queue.Insert(*cs.At(j))
					s.ca.At(*cs.At(j)).SetMark(2)
				}
			}
			*s.touched.At(i) = 0
		}
	}

	for i = 0; i < s.subsumption_queue.Size(); i++ {
		if s.ca.At(*s.subsumption_queue.At(i)).Mark() == 2 {
			s.ca.At(*s.subsumption_queue.At(i)).SetMark(0)
		}
	}

	s.n_touched = 0
}

// implied is SimpSolver::implied (protected).
// C: bool implied(const vec<Lit>& c)
func (s *MinisatSimpSolver) implied(c *MinisatVec[MinisatLit]) bool {
	if s.decisionLevel() != 0 {
		panic("MinisatSimpSolver.implied: decisionLevel() != 0")
	}

	s.trail_lim.PushElem(s.trail.Size())
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) == MinisatLTrue {
			s.cancelUntil(0)
			return false
		} else if s.ValueLit(*c.At(i)) != MinisatLFalse {
			if s.ValueLit(*c.At(i)) != MinisatLUndef {
				panic("MinisatSimpSolver.implied: value != l_Undef")
			}
			s.uncheckedEnqueue(MinisatLitNot(*c.At(i)), MinisatCRefUndef)
		}
	}

	result := s.propagate() != MinisatCRefUndef
	s.cancelUntil(0)
	return result
}

// backwardSubsumptionCheck is SimpSolver::backwardSubsumptionCheck (protected):
// backward subsumption + backward subsumption resolution.
// C: bool backwardSubsumptionCheck(bool verbose = false)
func (s *MinisatSimpSolver) backwardSubsumptionCheck(verbose bool) bool {
	cnt := 0
	subsumed := 0
	deletedLiterals := 0
	if s.decisionLevel() != 0 {
		panic("MinisatSimpSolver.backwardSubsumptionCheck: decisionLevel() != 0")
	}

	for s.subsumption_queue.Size() > 0 || s.bwdsub_assigns < s.trail.Size() {

		// Empty subsumption queue and return immediately on user-interrupt:
		if s.asynch_interrupt {
			s.subsumption_queue.Clear(false)
			s.bwdsub_assigns = s.trail.Size()
			break
		}

		// Check top-level assignments by creating a dummy clause and placing it in the queue:
		if s.subsumption_queue.Size() == 0 && s.bwdsub_assigns < s.trail.Size() {
			l := *s.trail.At(s.bwdsub_assigns)
			s.bwdsub_assigns++
			*s.ca.At(s.bwdsub_tmpunit).At(0) = l
			s.ca.At(s.bwdsub_tmpunit).CalcAbstraction()
			s.subsumption_queue.Insert(s.bwdsub_tmpunit)
		}

		cr := s.subsumption_queue.Peek()
		s.subsumption_queue.Pop()
		c := s.ca.At(cr)

		if c.Mark() != 0 {
			continue
		}

		if verbose && s.Verbosity >= 2 {
			if cnt%1000 == 0 {
				fmt.Printf("subsumption left: %10d (%10d subsumed, %10d deleted literals)\r", s.subsumption_queue.Size(), subsumed, deletedLiterals)
			}
			cnt++
		}

		if !(c.Size() > 1 || s.ValueLit(*c.At(0)) == MinisatLTrue) { // Unit-clauses should have been propagated before this point.
			panic("MinisatSimpSolver.backwardSubsumptionCheck: unit clause not propagated")
		}

		// Find best variable to scan:
		best := MinisatLitVar(*c.At(0))
		for i := 1; i < c.Size(); i++ {
			if s.occurs.At(MinisatLitVar(*c.At(i))).Size() < s.occurs.At(best).Size() {
				best = MinisatLitVar(*c.At(i))
			}
		}

		// Search all candidates:
		_cs := s.occurs.Lookup(best)
		cs := _cs.Data()

		for j := 0; j < _cs.Size(); j++ {
			if c.Mark() != 0 {
				break
			} else if s.ca.At(cs[j]).Mark() == 0 && cs[j] != cr && (s.Subsumption_lim == -1 || s.ca.At(cs[j]).Size() < s.Subsumption_lim) {
				l := c.Subsumes(s.ca.At(cs[j]))

				if l == MinisatLitUndef {
					subsumed++
					s.removeClause(cs[j])
				} else if l != MinisatLitError {
					deletedLiterals++

					if !s.strengthenClause(cs[j], MinisatLitNot(l)) {
						return false
					}

					// Did current candidate get deleted from cs? Then check candidate at index j again:
					if MinisatLitVar(l) == best {
						j--
					}
				}
			}
		}
	}

	return true
}

// asymm is SimpSolver::asymm (protected).
// C: bool asymm(Var v, CRef cr)
func (s *MinisatSimpSolver) asymm(v MinisatVar, cr MinisatCRef) bool {
	c := s.ca.At(cr)
	if s.decisionLevel() != 0 {
		panic("MinisatSimpSolver.asymm: decisionLevel() != 0")
	}

	if c.Mark() != 0 || s.satisfied(c) {
		return true
	}

	s.trail_lim.PushElem(s.trail.Size())
	l := MinisatLitUndef
	for i := 0; i < c.Size(); i++ {
		if MinisatLitVar(*c.At(i)) != v && s.ValueLit(*c.At(i)) != MinisatLFalse {
			s.uncheckedEnqueue(MinisatLitNot(*c.At(i)), MinisatCRefUndef)
		} else {
			l = *c.At(i)
		}
	}

	if s.propagate() != MinisatCRefUndef {
		s.cancelUntil(0)
		s.Asymm_lits++
		if !s.strengthenClause(cr, l) {
			return false
		}
	} else {
		s.cancelUntil(0)
	}

	return true
}

// asymmVar is SimpSolver::asymmVar (protected).
// C: bool asymmVar(Var v)
func (s *MinisatSimpSolver) asymmVar(v MinisatVar) bool {
	if !s.use_simplification {
		panic("MinisatSimpSolver.asymmVar: !use_simplification")
	}

	cls := s.occurs.Lookup(v)

	if s.Value(v) != MinisatLUndef || cls.Size() == 0 {
		return true
	}

	for i := 0; i < cls.Size(); i++ {
		if !s.asymm(v, *cls.At(i)) {
			return false
		}
	}

	return s.backwardSubsumptionCheck(false)
}

// mkElimClauseLit_minisat is the overload mkElimClause(vec<uint32_t>&, Lit) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Lit x)
func mkElimClauseLit_minisat(elimclauses *MinisatVec[uint32], x MinisatLit) {
	elimclauses.PushElem(uint32(MinisatToIntLit(x)))
	elimclauses.PushElem(1)
}

// mkElimClauseVarClause_minisat is the overload mkElimClause(vec<uint32_t>&, Var, Clause&) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Var v, Clause& c)
func mkElimClauseVarClause_minisat(elimclauses *MinisatVec[uint32], v MinisatVar, c *MinisatClause) {
	first := elimclauses.Size()
	vPos := -1

	// Copy clause to elimclauses-vector. Remember position where the variable 'v' occurs:
	for i := 0; i < c.Size(); i++ {
		elimclauses.PushElem(uint32(MinisatToIntLit(*c.At(i))))
		if MinisatLitVar(*c.At(i)) == v {
			vPos = i + first
		}
	}
	if vPos == -1 {
		panic("mkElimClauseVarClause: variable not in clause")
	}

	// Swap the first literal with the 'v' literal, so that the literal
	// containing 'v' will occur first in the clause:
	tmp := *elimclauses.At(vPos)
	*elimclauses.At(vPos) = *elimclauses.At(first)
	*elimclauses.At(first) = tmp

	// Store the length of the clause last:
	elimclauses.PushElem(uint32(c.Size()))
}

// eliminateVar is SimpSolver::eliminateVar (protected).
// C: bool eliminateVar(Var v)
func (s *MinisatSimpSolver) eliminateVar(v MinisatVar) bool {
	if *s.frozen.At(v) != 0 || s.IsEliminated(v) || s.Value(v) != MinisatLUndef {
		panic("MinisatSimpSolver.eliminateVar: precondition")
	}
	var i int

	// Split the occurrences into positive and negative:
	cls := s.occurs.Lookup(v)
	var pos, neg MinisatVec[MinisatCRef]
	for i = 0; i < cls.Size(); i++ {
		if minisatFind[MinisatLit](s.ca.At(*cls.At(i)), MinisatMkLit(v, false), minisatNe[MinisatLit]) {
			pos.PushElem(*cls.At(i))
		} else {
			neg.PushElem(*cls.At(i))
		}
	}

	// Check wether the increase in number of clauses stays within the allowed ('grow'). Moreover, no
	// clause must exceed the limit on the maximal clause size (if it is set):
	cnt := 0
	clauseSize := 0

	for i = 0; i < pos.Size(); i++ {
		for j := 0; j < neg.Size(); j++ {
			if s.mergeSize(s.ca.At(*pos.At(i)), s.ca.At(*neg.At(j)), v, &clauseSize) {
				cnt++
				if cnt > cls.Size()+s.Grow || (s.Clause_lim != -1 && clauseSize > s.Clause_lim) {
					return true
				}
			}
		}
	}

	// Delete and store old clauses:
	*s.eliminated.At(v) = 1
	s.SetDecisionVar(v, false)
	s.Eliminated_vars++

	if pos.Size() > neg.Size() {
		for i := 0; i < neg.Size(); i++ {
			mkElimClauseVarClause_minisat(&s.elimclauses, v, s.ca.At(*neg.At(i)))
		}
		mkElimClauseLit_minisat(&s.elimclauses, MinisatMkLit(v, false))
	} else {
		for i := 0; i < pos.Size(); i++ {
			mkElimClauseVarClause_minisat(&s.elimclauses, v, s.ca.At(*pos.At(i)))
		}
		mkElimClauseLit_minisat(&s.elimclauses, MinisatLitNot(MinisatMkLit(v, false)))
	}

	for i = 0; i < cls.Size(); i++ {
		s.removeClause(*cls.At(i))
	}

	// Produce clauses in cross product:
	resolvent := &s.add_tmp
	for i = 0; i < pos.Size(); i++ {
		for j := 0; j < neg.Size(); j++ {
			if s.merge(s.ca.At(*pos.At(i)), s.ca.At(*neg.At(j)), v, resolvent) && !s.AddClause_(resolvent) {
				return false
			}
		}
	}

	// Free occurs list for this variable:
	s.occurs.At(v).Clear(true)

	// Free watchers lists for this variable, if possible:
	if s.watches.At(MinisatMkLit(v, false)).Size() == 0 {
		s.watches.At(MinisatMkLit(v, false)).Clear(true)
	}
	if s.watches.At(MinisatLitNot(MinisatMkLit(v, false))).Size() == 0 {
		s.watches.At(MinisatLitNot(MinisatMkLit(v, false))).Clear(true)
	}

	return s.backwardSubsumptionCheck(false)
}

// Substitute is SimpSolver::substitute (replace all occurences of v with x; may cause a contradiction).
// C: bool substitute(Var v, Lit x)
func (s *MinisatSimpSolver) Substitute(v MinisatVar, x MinisatLit) bool {
	if *s.frozen.At(v) != 0 || s.IsEliminated(v) || s.Value(v) != MinisatLUndef {
		panic("MinisatSimpSolver.Substitute: precondition")
	}

	if !s.ok {
		return false
	}

	*s.eliminated.At(v) = 1
	s.SetDecisionVar(v, false)
	cls := s.occurs.Lookup(v)

	substClause := &s.add_tmp
	for i := 0; i < cls.Size(); i++ {
		c := s.ca.At(*cls.At(i))

		substClause.Clear(false)
		for j := 0; j < c.Size(); j++ {
			p := *c.At(j)
			if MinisatLitVar(p) == v {
				substClause.PushElem(MinisatLitXor(x, MinisatSign(p)))
			} else {
				substClause.PushElem(p)
			}
		}

		s.removeClause(*cls.At(i))

		if !s.AddClause_(substClause) {
			s.ok = false
			return false
		}
	}

	return true
}

// extendModel is SimpSolver::extendModel (protected).
// C: void extendModel()
func (s *MinisatSimpSolver) extendModel() {
	var i, j int

	for i = s.elimclauses.Size() - 1; i > 0; i -= j {
		j = int(*s.elimclauses.At(i))
		i--
		skip := false
		for ; j > 1; j, i = j-1, i-1 {
			if s.ModelValueLit(MinisatToLit(int(*s.elimclauses.At(i)))) != MinisatLFalse {
				skip = true // goto next
				break
			}
		}
		if skip {
			continue
		}
		x := MinisatToLit(int(*s.elimclauses.At(i)))
		*s.Model.At(MinisatLitVar(x)) = NewMinisatLboolB(!MinisatSign(x))
	}
}

// Eliminate is SimpSolver::eliminate (variable-elimination based simplification).
// C: bool eliminate(bool turn_off_elim = false)
func (s *MinisatSimpSolver) Eliminate(turnOffElim bool) bool {
	if !s.Simplify() {
		return false
	} else if !s.use_simplification {
		return true
	}

	// Main simplification loop (the C `goto cleanup` exits become returns of the closure):
	func() {
		for s.n_touched > 0 || s.bwdsub_assigns < s.trail.Size() || s.elim_heap.Size() > 0 {

			s.gatherTouchedClauses()
			if (s.subsumption_queue.Size() > 0 || s.bwdsub_assigns < s.trail.Size()) &&
				!s.backwardSubsumptionCheck(true) {
				s.ok = false
				return
			}

			// Empty elim_heap and return immediately on user-interrupt:
			if s.asynch_interrupt {
				if s.bwdsub_assigns != s.trail.Size() || s.subsumption_queue.Size() != 0 || s.n_touched != 0 {
					panic("MinisatSimpSolver.Eliminate: queues not empty on interrupt")
				}
				s.elim_heap.Clear(false)
				return
			}

			for cnt := 0; !s.elim_heap.Empty(); cnt++ {
				elim := s.elim_heap.RemoveMin()

				if s.asynch_interrupt {
					break
				}

				if s.IsEliminated(elim) || s.Value(elim) != MinisatLUndef {
					continue
				}

				if s.Verbosity >= 2 && cnt%100 == 0 {
					fmt.Printf("elimination left: %10d\r", s.elim_heap.Size())
				}

				if s.Use_asymm {
					// Temporarily freeze variable. Otherwise, it would immediately end up on the queue again:
					wasFrozen := *s.frozen.At(elim)
					*s.frozen.At(elim) = 1
					if !s.asymmVar(elim) {
						s.ok = false
						return
					}
					*s.frozen.At(elim) = wasFrozen
				}

				// At this point, the variable may have been set by assymetric branching, so check it
				// again. Also, don't eliminate frozen variables:
				if s.Use_elim && s.Value(elim) == MinisatLUndef && *s.frozen.At(elim) == 0 && !s.eliminateVar(elim) {
					s.ok = false
					return
				}

				s.CheckGarbageFrac(s.Simp_garbage_frac)
			}

			if s.subsumption_queue.Size() != 0 {
				panic("MinisatSimpSolver.Eliminate: subsumption queue not empty")
			}
		}
	}()

	// cleanup:
	// If no more simplification is needed, free all simplification-related data structures:
	if turnOffElim {
		s.touched.Clear(true)
		s.occurs.Clear(true)
		s.n_occ.Clear(true)
		s.elim_heap.Clear(true)
		s.subsumption_queue.Clear(true)

		s.use_simplification = false
		s.remove_satisfied = true
		s.ca.ExtraClauseField = false

		// Force full cleanup (this is safe and desirable since it only happens once):
		s.rebuildOrderHeap()
		s.GarbageCollect()
	} else {
		// Cheaper cleanup:
		s.cleanUpClauses() // TODO: can we make 'cleanUpClauses()' not be linear in the problem size somehow?
		s.CheckGarbage()
	}

	if s.Verbosity >= 1 && s.elimclauses.Size() > 0 {
		fmt.Printf("|  Eliminated clauses:     %10.2f Mb                                      |\n",
			float64(s.elimclauses.Size()*4)/(1024*1024))
	}

	return s.ok
}

// cleanUpClauses is SimpSolver::cleanUpClauses (protected).
// C: void cleanUpClauses()
func (s *MinisatSimpSolver) cleanUpClauses() {
	s.occurs.CleanAll()
	var i, j int
	for i, j = 0, 0; i < s.clauses.Size(); i++ {
		if s.ca.At(*s.clauses.At(i)).Mark() == 0 {
			*s.clauses.At(j) = *s.clauses.At(i)
			j++
		}
	}
	s.clauses.Shrink(i - j)
}

// relocAll is SimpSolver::relocAll (protected; hides Solver::relocAll).
// C: void relocAll(ClauseAllocator& to)
func (s *MinisatSimpSolver) relocAll(to *MinisatClauseAllocator) {
	if !s.use_simplification {
		return
	}

	// All occurs lists:
	var i int
	for i = 0; i < s.NVars(); i++ {
		cs := s.occurs.At(i)
		for j := 0; j < cs.Size(); j++ {
			s.ca.Reloc(cs.At(j), to)
		}
	}

	// Subsumption queue:
	for i = 0; i < s.subsumption_queue.Size(); i++ {
		s.ca.Reloc(s.subsumption_queue.At(i), to)
	}

	// Temporary clause:
	s.ca.Reloc(&s.bwdsub_tmpunit, to)
}

// GarbageCollect is SimpSolver::garbageCollect (override).
// C: virtual void garbageCollect()
func (s *MinisatSimpSolver) GarbageCollect() {
	// Initialize the next region to a size corresponding to the estimated utilization degree. This
	// is not precise but should avoid some unnecessary reallocations for the new region:
	to := NewMinisatClauseAllocatorCap(s.ca.Size() - s.ca.Wasted())

	s.cleanUpClauses()
	to.ExtraClauseField = s.ca.ExtraClauseField // NOTE: this is important to keep (or lose) the extra fields.
	s.relocAll(to)
	s.MinisatSolver.relocAll(to)
	if s.Verbosity >= 2 {
		fmt.Printf("|  Garbage collection:   %12d bytes => %12d bytes             |\n",
			s.ca.Size()*MinisatRegionAllocatorUnitSize, to.Size()*MinisatRegionAllocatorUnitSize)
	}
	to.MoveTo(&s.ca)
}

// Solve is the overload solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) Solve(assumps *MinisatVec[MinisatLit], doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	assumps.CopyTo(&s.assumptions)
	return s.solve_(doSimp, turnOffSimp) == MinisatLTrue
}

// SolveLimited is SimpSolver::solveLimited.
// C: lbool solveLimited(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) SolveLimited(assumps *MinisatVec[MinisatLit], doSimp bool, turnOffSimp bool) MinisatLbool {
	assumps.CopyTo(&s.assumptions)
	return s.solve_(doSimp, turnOffSimp)
}

// SolveNoAssumps is the overload solve(bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) SolveNoAssumps(doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	return s.solve_(doSimp, turnOffSimp) == MinisatLTrue
}

// Solve1 is the overload solve(Lit p, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) Solve1(p MinisatLit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	return s.solve_(doSimp, turnOffSimp) == MinisatLTrue
}

// Solve2 is the overload solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) Solve2(p MinisatLit, q MinisatLit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	return s.solve_(doSimp, turnOffSimp) == MinisatLTrue
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) Solve3(p MinisatLit, q MinisatLit, r MinisatLit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	s.assumptions.PushElem(r)
	return s.solve_(doSimp, turnOffSimp) == MinisatLTrue
}

// The five toDimacs(...) overloads declared in SimpSolver.h are never defined in
// SimpSolver.cpp, so they are not ported (calling them is a link error in C++).
