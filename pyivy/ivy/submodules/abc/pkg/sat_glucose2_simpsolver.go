package abc

import (
	"fmt"
	"math"
)

// Port of src/sat/glucose2/SimpSolver.h and SimpSolver.cpp (namespace Gluco2).
// SimpSolver derives from Solver: it embeds Gluco2Solver and re-declares the
// overridden methods (NewVar, AddClause*, Solve*, ToDimacs*, GarbageCollect, ...).

// Gluco2SimpSolverElimLt is SimpSolver::ElimLt.
type Gluco2SimpSolverElimLt struct {
	n_occ *Gluco2Vec[int] // const vec<int>&
}

// NewGluco2SimpSolverElimLt is the explicit constructor ElimLt(const vec<int>& no).
// C: explicit ElimLt(const vec<int>& no)
func NewGluco2SimpSolverElimLt(no *Gluco2Vec[int]) Gluco2SimpSolverElimLt {
	return Gluco2SimpSolverElimLt{n_occ: no}
}

// cost is ElimLt::cost.
// C: uint64_t cost(Var x) const
func (e Gluco2SimpSolverElimLt) cost(x Gluco2Var) uint64 {
	return uint64(*e.n_occ.At(Gluco2ToIntLit(Gluco2MkLit(x, false)))) * uint64(*e.n_occ.At(Gluco2ToIntLit(Gluco2LitNot(Gluco2MkLit(x, false)))))
}

// Call is ElimLt::operator().
// C: bool operator()(Var x, Var y) const
func (e Gluco2SimpSolverElimLt) Call(x int, y int) bool { return e.cost(x) < e.cost(y) }

// Gluco2SimpSolverClauseDeleted is SimpSolver::ClauseDeleted.
type Gluco2SimpSolverClauseDeleted struct {
	ca *Gluco2ClauseAllocator // const ClauseAllocator&
}

// NewGluco2SimpSolverClauseDeleted is the explicit constructor ClauseDeleted(const ClauseAllocator& _ca).
// C: explicit ClauseDeleted(const ClauseAllocator& _ca)
func NewGluco2SimpSolverClauseDeleted(ca *Gluco2ClauseAllocator) Gluco2SimpSolverClauseDeleted {
	return Gluco2SimpSolverClauseDeleted{ca: ca}
}

// Call is ClauseDeleted::operator().
// C: bool operator()(const CRef& cr) const
func (d Gluco2SimpSolverClauseDeleted) Call(cr Gluco2CRef) bool { return d.ca.At(cr).Mark() == 1 }

// gluco2SimpSolverCat is `static const char* _cat = "SIMP"`.
const gluco2SimpSolverCat = "SIMP"

// Command-line options of the simplifier (static objects in C++).
var (
	opt_use_asymm_glucose2         = NewGluco2BoolOption(gluco2SimpSolverCat, "asymm", "Shrink clauses by asymmetric branching.", false)
	opt_use_rcheck_glucose2        = NewGluco2BoolOption(gluco2SimpSolverCat, "rcheck", "Check if a clause is already implied. (costly)", false)
	opt_use_elim_glucose2          = NewGluco2BoolOption(gluco2SimpSolverCat, "elim", "Perform variable elimination.", true)
	opt_grow_glucose2              = NewGluco2IntOption(gluco2SimpSolverCat, "grow", "Allow a variable elimination step to grow by a number of clauses.", 0, NewGluco2IntRange(math.MinInt32, math.MaxInt32))
	opt_clause_lim_glucose2        = NewGluco2IntOption(gluco2SimpSolverCat, "cl-lim", "Variables are not eliminated if it produces a resolvent with a length above this limit. -1 means no limit", 20, NewGluco2IntRange(-1, math.MaxInt32))
	opt_subsumption_lim_glucose2   = NewGluco2IntOption(gluco2SimpSolverCat, "sub-lim", "Do not check if subsumption against a clause larger than this. -1 means no limit.", 1000, NewGluco2IntRange(-1, math.MaxInt32))
	opt_simp_garbage_frac_glucose2 = NewGluco2DoubleOption(gluco2SimpSolverCat, "simp-gc-frac", "The fraction of wasted memory allowed before a garbage collection is triggered during simplification.", 0.5, NewGluco2DoubleRange(0, false, math.Inf(1), false))
)

// Gluco2SimpSolver is Gluco2::SimpSolver : public Solver.
type Gluco2SimpSolver struct {
	Gluco2Solver

	Grow               int     // allow a variable elimination step to grow by a number of clauses (default zero)
	Clause_lim         int     // variables are not eliminated if it produces a resolvent longer than this limit
	Subsumption_lim    int     // do not check subsumption against a clause larger than this (-1 = no limit)
	Simp_garbage_frac  float64 // limit for when to issue a GC during simplification
	Use_asymm          bool    // shrink clauses by asymmetric branching
	Use_rcheck         bool    // check if a clause is already implied
	Use_elim           bool    // perform variable elimination
	Parsing            int
	Merges             int
	Asymm_lits         int
	Eliminated_vars    int
	Eliminated_clauses int

	// Solver state (protected):
	elimorder          int
	use_simplification bool
	elimclauses        Gluco2Vec[uint32]
	touched            Gluco2Vec[int8]
	occurs             Gluco2OccLists[int, Gluco2CRef, Gluco2SimpSolverClauseDeleted]
	n_occ              Gluco2Vec[int]
	elim_heap          Gluco2Heap[Gluco2SimpSolverElimLt]
	subsumption_queue  Gluco2Queue[Gluco2CRef]
	frozen             Gluco2Vec[int8]
	eliminated         Gluco2Vec[int8]
	bwdsub_assigns     int
	n_touched          int
	bwdsub_tmpunit     Gluco2CRef
}

// NewGluco2SimpSolver is SimpSolver::SimpSolver().
// C: SimpSolver::SimpSolver()
func NewGluco2SimpSolver() *Gluco2SimpSolver {
	s := &Gluco2SimpSolver{}
	s.initGluco2Solver()
	s.virtGC = s.GarbageCollect // GarbageCollect is virtual
	s.Grow = int(opt_grow_glucose2.Value())
	s.Clause_lim = int(opt_clause_lim_glucose2.Value())
	s.Subsumption_lim = int(opt_subsumption_lim_glucose2.Value())
	s.Simp_garbage_frac = opt_simp_garbage_frac_glucose2.Value()
	s.Use_asymm = opt_use_asymm_glucose2.Value()
	s.Use_rcheck = opt_use_rcheck_glucose2.Value()
	s.Use_elim = opt_use_elim_glucose2.Value()
	s.Merges = 0
	s.Asymm_lits = 0
	s.Eliminated_vars = 0
	s.elimorder = 1
	s.use_simplification = true
	s.occurs = *NewGluco2OccLists[int, Gluco2CRef, Gluco2SimpSolverClauseDeleted](NewGluco2SimpSolverClauseDeleted(&s.ca))
	s.elim_heap = *NewGluco2Heap(NewGluco2SimpSolverElimLt(&s.n_occ))
	s.bwdsub_assigns = 0
	s.n_touched = 0
	s.subsumption_queue = *NewGluco2Queue[Gluco2CRef]()

	dummy := NewGluco2VecPad(1, Gluco2LitUndef)
	s.ca.ExtraClauseField = true // NOTE: must happen before allocating the dummy clause below.
	s.bwdsub_tmpunit = s.ca.Alloc(dummy, false)
	s.remove_satisfied = false
	return s
}

// Destroy is SimpSolver::~SimpSolver().
// C: SimpSolver::~SimpSolver()
func (s *Gluco2SimpSolver) Destroy() {}

// NewVar is SimpSolver::newVar (override; polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *Gluco2SimpSolver) NewVar(sign bool, dvar bool) Gluco2Var {
	v := s.Gluco2Solver.NewVar(sign, dvar)

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
func (s *Gluco2SimpSolver) IsEliminated(v Gluco2Var) bool {
	if s.eliminated.Size() > 0 {
		return *s.eliminated.At(v) != 0
	}
	return false
}

// updateElimHeap is SimpSolver::updateElimHeap (protected).
// C: void updateElimHeap(Var v)
func (s *Gluco2SimpSolver) updateElimHeap(v Gluco2Var) {
	if !s.use_simplification {
		panic("Gluco2SimpSolver.updateElimHeap: !use_simplification")
	}
	if s.elim_heap.InHeap(v) || (*s.frozen.At(v) == 0 && !s.IsEliminated(v) && s.Value(v) == Gluco2LUndef) {
		s.elim_heap.Update(v)
	}
}

// SetFrozen is SimpSolver::setFrozen (a frozen variable is not eliminated).
// C: void setFrozen(Var v, bool b)
func (s *Gluco2SimpSolver) SetFrozen(v Gluco2Var, b bool) {
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
func (s *Gluco2SimpSolver) solve_(doSimp bool, turnOffSimp bool) Gluco2Lbool {
	var extraFrozen Gluco2Vec[Gluco2Var]
	result := Gluco2LTrue

	doSimp = doSimp && s.use_simplification

	if doSimp {
		// Assumptions must be temporarily frozen to run variable elimination:
		for i := 0; i < s.assumptions.Size(); i++ {
			v := Gluco2LitVar(*s.assumptions.At(i))

			// If an assumption has been eliminated, remember it.
			if s.IsEliminated(v) {
				panic("Gluco2SimpSolver.solve_: assumption has been eliminated")
			}

			if *s.frozen.At(v) == 0 {
				// Freeze and store.
				s.SetFrozen(v, true)
				extraFrozen.PushElem(v)
			}
		}

		result = NewGluco2LboolB(s.Eliminate(turnOffSimp))
	}

	if result == Gluco2LTrue {
		result = s.Gluco2Solver.solve_()
	} else if s.Verbosity >= 1 {
		fmt.Printf("===============================================================================\n")
	}

	if result == Gluco2LTrue {
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
func (s *Gluco2SimpSolver) AddClause_(ps *Gluco2Vec[Gluco2Lit]) bool {
	for i := 0; i < ps.Size(); i++ { // #ifndef NDEBUG
		if s.IsEliminated(Gluco2LitVar(*ps.At(i))) {
			panic("Gluco2SimpSolver.AddClause_: literal of an eliminated variable")
		}
	}

	nclauses := s.clauses.Size()

	if s.Use_rcheck && s.implied(ps) {
		return true
	}

	if !s.Gluco2Solver.AddClause_(ps) {
		return false
	}

	if s.Parsing == 0 && s.CertifiedUNSAT {
		for i := 0; i < ps.Size(); i++ {
			fmt.Fprintf(s.CertifiedOutput, "%d ", gluco2DimacsLit(*ps.At(i)))
		}
		fmt.Fprintf(s.CertifiedOutput, "0\n")
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
			s.occurs.At(Gluco2LitVar(*c.At(i))).PushElem(cr)
			*s.n_occ.At(Gluco2ToIntLit(*c.At(i)))++
			*s.touched.At(Gluco2LitVar(*c.At(i))) = 1
			s.n_touched++
			if s.elim_heap.InHeap(Gluco2LitVar(*c.At(i))) {
				s.elim_heap.Increase(Gluco2LitVar(*c.At(i)))
			}
		}
	}

	return true
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *Gluco2SimpSolver) AddClause(ps *Gluco2Vec[Gluco2Lit]) bool {
	ps.CopyTo(&s.add_tmp)
	return s.AddClause_(&s.add_tmp)
}

// AddEmptyClause is SimpSolver::addEmptyClause.
// C: bool addEmptyClause()
func (s *Gluco2SimpSolver) AddEmptyClause() bool {
	s.add_tmp.Clear(false)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *Gluco2SimpSolver) AddClauseUnit(p Gluco2Lit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *Gluco2SimpSolver) AddClauseBinary(p Gluco2Lit, q Gluco2Lit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *Gluco2SimpSolver) AddClauseTernary(p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	s.add_tmp.PushElem(r)
	return s.AddClause_(&s.add_tmp)
}

// removeClause is SimpSolver::removeClause (protected; hides Solver::removeClause).
// C: void removeClause(CRef cr)
func (s *Gluco2SimpSolver) removeClause(cr Gluco2CRef) {
	c := s.ca.At(cr)

	if s.use_simplification {
		for i := 0; i < c.Size(); i++ {
			*s.n_occ.At(Gluco2ToIntLit(*c.At(i)))--
			s.updateElimHeap(Gluco2LitVar(*c.At(i)))
			s.occurs.Smudge(Gluco2LitVar(*c.At(i)))
		}
	}

	s.Gluco2Solver.removeClause(cr)
}

// strengthenClause is SimpSolver::strengthenClause (protected).
// C: bool strengthenClause(CRef cr, Lit l)
func (s *Gluco2SimpSolver) strengthenClause(cr Gluco2CRef, l Gluco2Lit) bool {
	c := s.ca.At(cr)
	if s.decisionLevel() != 0 || !s.use_simplification {
		panic("Gluco2SimpSolver.strengthenClause: precondition")
	}

	// FIX: this is too inefficient but would be nice to have (properly implemented)
	// if (!find(subsumption_queue, &c))
	s.subsumption_queue.Insert(cr)

	if s.CertifiedUNSAT {
		for i := 0; i < c.Size(); i++ {
			if *c.At(i) != l {
				fmt.Fprintf(s.CertifiedOutput, "%d ", gluco2DimacsLit(*c.At(i)))
			}
		}
		fmt.Fprintf(s.CertifiedOutput, "0\n")
	}

	if c.Size() == 2 {
		s.removeClause(cr)
		c.Strengthen(l)
	} else {
		if s.CertifiedUNSAT {
			fmt.Fprintf(s.CertifiedOutput, "d ")
			for i := 0; i < c.Size(); i++ {
				fmt.Fprintf(s.CertifiedOutput, "%d ", gluco2DimacsLit(*c.At(i)))
			}
			fmt.Fprintf(s.CertifiedOutput, "0\n")
		}

		s.detachClause(cr, true)
		c.Strengthen(l)
		s.attachClause(cr)
		gluco2Remove[Gluco2CRef](s.occurs.At(Gluco2LitVar(l)), cr, gluco2Ne[Gluco2CRef])
		*s.n_occ.At(Gluco2ToIntLit(l))--
		s.updateElimHeap(Gluco2LitVar(l))
	}

	if c.Size() == 1 {
		return s.enqueue(*c.At(0), Gluco2CRefUndef) && s.propagate() == Gluco2CRefUndef
	}
	return true
}

// merge is the overload merge(..., vec<Lit>& out_clause) (protected): returns
// false if the clause is always satisfied ('out_clause' should not be used).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, vec<Lit>& out_clause)
func (s *Gluco2SimpSolver) merge(_ps *Gluco2Clause, _qs *Gluco2Clause, v Gluco2Var, outClause *Gluco2Vec[Gluco2Lit]) bool {
	s.Merges++
	outClause.Clear(false)

	psSmallest := _ps.Size() < _qs.Size()
	ps, qs := _ps, _qs
	if psSmallest {
		ps, qs = _qs, _ps
	}
	var i, j int

	for i = 0; i < qs.Size(); i++ {
		if Gluco2LitVar(*qs.At(i)) != v {
			found := false
			for j = 0; j < ps.Size(); j++ {
				if Gluco2LitVar(*ps.At(j)) == Gluco2LitVar(*qs.At(i)) {
					if *ps.At(j) == Gluco2LitNot(*qs.At(i)) {
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
		if Gluco2LitVar(*ps.At(i)) != v {
			outClause.PushElem(*ps.At(i))
		}
	}

	return true
}

// mergeSize is the overload merge(..., int& size) (protected): returns false if
// the clause is always satisfied.
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, int& size)
func (s *Gluco2SimpSolver) mergeSize(_ps *Gluco2Clause, _qs *Gluco2Clause, v Gluco2Var, size *int) bool {
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
		if Gluco2LitVar(qs_[i]) != v {
			found := false
			for j := 0; j < ps.Size(); j++ {
				if Gluco2LitVar(ps_[j]) == Gluco2LitVar(qs_[i]) {
					if ps_[j] == Gluco2LitNot(qs_[i]) {
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
func (s *Gluco2SimpSolver) gatherTouchedClauses() {
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
func (s *Gluco2SimpSolver) implied(c *Gluco2Vec[Gluco2Lit]) bool {
	if s.decisionLevel() != 0 {
		panic("Gluco2SimpSolver.implied: decisionLevel() != 0")
	}

	s.trail_lim.PushElem(s.trail.Size())
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) == Gluco2LTrue {
			s.cancelUntil(0)
			return false
		} else if s.ValueLit(*c.At(i)) != Gluco2LFalse {
			if s.ValueLit(*c.At(i)) != Gluco2LUndef {
				panic("Gluco2SimpSolver.implied: value != l_Undef")
			}
			s.uncheckedEnqueue(Gluco2LitNot(*c.At(i)), Gluco2CRefUndef)
		}
	}

	result := s.propagate() != Gluco2CRefUndef
	s.cancelUntil(0)
	return result
}

// backwardSubsumptionCheck is SimpSolver::backwardSubsumptionCheck (protected):
// backward subsumption + backward subsumption resolution.
// C: bool backwardSubsumptionCheck(bool verbose = false)
func (s *Gluco2SimpSolver) backwardSubsumptionCheck(verbose bool) bool {
	cnt := 0
	subsumed := 0
	deletedLiterals := 0
	if s.decisionLevel() != 0 {
		panic("Gluco2SimpSolver.backwardSubsumptionCheck: decisionLevel() != 0")
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

		if !(c.Size() > 1 || s.ValueLit(*c.At(0)) == Gluco2LTrue) { // Unit-clauses should have been propagated before this point.
			panic("Gluco2SimpSolver.backwardSubsumptionCheck: unit clause not propagated")
		}

		// Find best variable to scan:
		best := Gluco2LitVar(*c.At(0))
		for i := 1; i < c.Size(); i++ {
			if s.occurs.At(Gluco2LitVar(*c.At(i))).Size() < s.occurs.At(best).Size() {
				best = Gluco2LitVar(*c.At(i))
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

				if l == Gluco2LitUndef {
					subsumed++
					s.removeClause(cs[j])
				} else if l != Gluco2LitError {
					deletedLiterals++

					if !s.strengthenClause(cs[j], Gluco2LitNot(l)) {
						return false
					}

					// Did current candidate get deleted from cs? Then check candidate at index j again:
					if Gluco2LitVar(l) == best {
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
func (s *Gluco2SimpSolver) asymm(v Gluco2Var, cr Gluco2CRef) bool {
	c := s.ca.At(cr)
	if s.decisionLevel() != 0 {
		panic("Gluco2SimpSolver.asymm: decisionLevel() != 0")
	}

	if c.Mark() != 0 || s.satisfied(c) {
		return true
	}

	s.trail_lim.PushElem(s.trail.Size())
	l := Gluco2LitUndef
	for i := 0; i < c.Size(); i++ {
		if Gluco2LitVar(*c.At(i)) != v && s.ValueLit(*c.At(i)) != Gluco2LFalse {
			s.uncheckedEnqueue(Gluco2LitNot(*c.At(i)), Gluco2CRefUndef)
		} else {
			l = *c.At(i)
		}
	}

	if s.propagate() != Gluco2CRefUndef {
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
func (s *Gluco2SimpSolver) asymmVar(v Gluco2Var) bool {
	if !s.use_simplification {
		panic("Gluco2SimpSolver.asymmVar: !use_simplification")
	}

	cls := s.occurs.Lookup(v)

	if s.Value(v) != Gluco2LUndef || cls.Size() == 0 {
		return true
	}

	for i := 0; i < cls.Size(); i++ {
		if !s.asymm(v, *cls.At(i)) {
			return false
		}
	}

	return s.backwardSubsumptionCheck(false)
}

// mkElimClauseLit_glucose2 is the overload mkElimClause(vec<uint32_t>&, Lit) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Lit x)
func mkElimClauseLit_glucose2(elimclauses *Gluco2Vec[uint32], x Gluco2Lit) {
	elimclauses.PushElem(uint32(Gluco2ToIntLit(x)))
	elimclauses.PushElem(1)
}

// mkElimClauseVarClause_glucose2 is the overload mkElimClause(vec<uint32_t>&, Var, Clause&) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Var v, Clause& c)
func mkElimClauseVarClause_glucose2(elimclauses *Gluco2Vec[uint32], v Gluco2Var, c *Gluco2Clause) {
	first := elimclauses.Size()
	vPos := -1

	// Copy clause to elimclauses-vector. Remember position where the variable 'v' occurs:
	for i := 0; i < c.Size(); i++ {
		elimclauses.PushElem(uint32(Gluco2ToIntLit(*c.At(i))))
		if Gluco2LitVar(*c.At(i)) == v {
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
func (s *Gluco2SimpSolver) eliminateVar(v Gluco2Var) bool {
	if *s.frozen.At(v) != 0 || s.IsEliminated(v) || s.Value(v) != Gluco2LUndef {
		panic("Gluco2SimpSolver.eliminateVar: precondition")
	}
	var i int

	// Split the occurrences into positive and negative:
	cls := s.occurs.Lookup(v)
	var pos, neg Gluco2Vec[Gluco2CRef]
	for i = 0; i < cls.Size(); i++ {
		if gluco2Find[Gluco2Lit](s.ca.At(*cls.At(i)), Gluco2MkLit(v, false), gluco2Ne[Gluco2Lit]) {
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
			mkElimClauseVarClause_glucose2(&s.elimclauses, v, s.ca.At(*neg.At(i)))
		}
		mkElimClauseLit_glucose2(&s.elimclauses, Gluco2MkLit(v, false))
		s.Eliminated_clauses += neg.Size()
	} else {
		for i := 0; i < pos.Size(); i++ {
			mkElimClauseVarClause_glucose2(&s.elimclauses, v, s.ca.At(*pos.At(i)))
		}
		mkElimClauseLit_glucose2(&s.elimclauses, Gluco2LitNot(Gluco2MkLit(v, false)))
		s.Eliminated_clauses += pos.Size()
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

	for i = 0; i < cls.Size(); i++ {
		s.removeClause(*cls.At(i))
	}

	// Free occurs list for this variable:
	s.occurs.At(v).Clear(true)

	// Free watchers lists for this variable, if possible:
	if s.watches.At(Gluco2MkLit(v, false)).Size() == 0 {
		s.watches.At(Gluco2MkLit(v, false)).Clear(true)
	}
	if s.watches.At(Gluco2LitNot(Gluco2MkLit(v, false))).Size() == 0 {
		s.watches.At(Gluco2LitNot(Gluco2MkLit(v, false))).Clear(true)
	}

	return s.backwardSubsumptionCheck(false)
}

// Substitute is SimpSolver::substitute (replace all occurences of v with x; may cause a contradiction).
// C: bool substitute(Var v, Lit x)
func (s *Gluco2SimpSolver) Substitute(v Gluco2Var, x Gluco2Lit) bool {
	if *s.frozen.At(v) != 0 || s.IsEliminated(v) || s.Value(v) != Gluco2LUndef {
		panic("Gluco2SimpSolver.Substitute: precondition")
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
			if Gluco2LitVar(p) == v {
				substClause.PushElem(Gluco2LitXor(x, Gluco2Sign(p)))
			} else {
				substClause.PushElem(p)
			}
		}

		if !s.AddClause_(substClause) {
			s.ok = false
			return false
		}

		s.removeClause(*cls.At(i))
	}

	return true
}

// extendModel is SimpSolver::extendModel (protected).
// C: void extendModel()
func (s *Gluco2SimpSolver) extendModel() {
	var i, j int

	for i = s.elimclauses.Size() - 1; i > 0; i -= j {
		j = int(*s.elimclauses.At(i))
		i--
		skip := false
		for ; j > 1; j, i = j-1, i-1 {
			if s.ModelValueLit(Gluco2ToLit(int(*s.elimclauses.At(i)))) != Gluco2LFalse {
				skip = true // goto next
				break
			}
		}
		if skip {
			continue
		}
		x := Gluco2ToLit(int(*s.elimclauses.At(i)))
		*s.Model.At(Gluco2LitVar(x)) = NewGluco2LboolB(!Gluco2Sign(x))
	}
}

// Eliminate is SimpSolver::eliminate (variable-elimination based simplification).
// C: bool eliminate(bool turn_off_elim = false)
func (s *Gluco2SimpSolver) Eliminate(turnOffElim bool) bool {
	if !s.Simplify() {
		return false
	} else if !s.use_simplification {
		return true
	}

	toPerform := s.clauses.Size() <= 4800000
	if !toPerform {
		fmt.Printf("c Too many clauses... No preprocessing\n")
	}

	// Main simplification loop (the C `goto cleanup` exits become returns of the closure):
	func() {
		for toPerform && (s.n_touched > 0 || s.bwdsub_assigns < s.trail.Size() || s.elim_heap.Size() > 0) {

			s.gatherTouchedClauses()
			if (s.subsumption_queue.Size() > 0 || s.bwdsub_assigns < s.trail.Size()) &&
				!s.backwardSubsumptionCheck(true) {
				s.ok = false
				return
			}

			// Empty elim_heap and return immediately on user-interrupt:
			if s.asynch_interrupt {
				if s.bwdsub_assigns != s.trail.Size() || s.subsumption_queue.Size() != 0 || s.n_touched != 0 {
					panic("Gluco2SimpSolver.Eliminate: queues not empty on interrupt")
				}
				s.elim_heap.Clear(false)
				return
			}

			for cnt := 0; !s.elim_heap.Empty(); cnt++ {
				elim := s.elim_heap.RemoveMin()

				if s.asynch_interrupt {
					break
				}

				if s.IsEliminated(elim) || s.Value(elim) != Gluco2LUndef {
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
				if s.Use_elim && s.Value(elim) == Gluco2LUndef && *s.frozen.At(elim) == 0 && !s.eliminateVar(elim) {
					s.ok = false
					return
				}

				s.CheckGarbageFrac(s.Simp_garbage_frac)
			}

			if s.subsumption_queue.Size() != 0 {
				panic("Gluco2SimpSolver.Eliminate: subsumption queue not empty")
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
		fmt.Printf("c |  Eliminated clauses:     %10.2f Mb                                                                |\n",
			float64(s.elimclauses.Size()*4)/(1024*1024))
	}

	return s.ok
}

// cleanUpClauses is SimpSolver::cleanUpClauses (protected).
// C: void cleanUpClauses()
func (s *Gluco2SimpSolver) cleanUpClauses() {
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
func (s *Gluco2SimpSolver) relocAll(to *Gluco2ClauseAllocator) {
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
func (s *Gluco2SimpSolver) GarbageCollect() {
	// Initialize the next region to a size corresponding to the estimated utilization degree. This
	// is not precise but should avoid some unnecessary reallocations for the new region:
	to := NewGluco2ClauseAllocatorCap(s.ca.Size() - s.ca.Wasted())

	s.cleanUpClauses()
	to.ExtraClauseField = s.ca.ExtraClauseField // NOTE: this is important to keep (or lose) the extra fields.
	s.relocAll(to)
	s.Gluco2Solver.relocAll(to)
	if s.Verbosity >= 2 {
		fmt.Printf("|  Garbage collection:   %12d bytes => %12d bytes             |\n",
			s.ca.Size()*Gluco2RegionAllocatorUnitSize, to.Size()*Gluco2RegionAllocatorUnitSize)
	}
	to.MoveTo(&s.ca)
}

// Solve is the overload solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) Solve(assumps *Gluco2Vec[Gluco2Lit], doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	assumps.CopyTo(&s.assumptions)
	return s.solve_(doSimp, turnOffSimp) == Gluco2LTrue
}

// SolveLimited is SimpSolver::solveLimited.
// C: lbool solveLimited(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) SolveLimited(assumps *Gluco2Vec[Gluco2Lit], doSimp bool, turnOffSimp bool) Gluco2Lbool {
	assumps.CopyTo(&s.assumptions)
	return s.solve_(doSimp, turnOffSimp)
}

// SolveNoAssumps is the overload solve(bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) SolveNoAssumps(doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	return s.solve_(doSimp, turnOffSimp) == Gluco2LTrue
}

// Solve1 is the overload solve(Lit p, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) Solve1(p Gluco2Lit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	return s.solve_(doSimp, turnOffSimp) == Gluco2LTrue
}

// Solve2 is the overload solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) Solve2(p Gluco2Lit, q Gluco2Lit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	return s.solve_(doSimp, turnOffSimp) == Gluco2LTrue
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) Solve3(p Gluco2Lit, q Gluco2Lit, r Gluco2Lit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	s.assumptions.PushElem(r)
	return s.solve_(doSimp, turnOffSimp) == Gluco2LTrue
}

// The five toDimacs(...) overloads declared in SimpSolver.h are never defined in
// SimpSolver.cpp, so they are not ported (calling them is a link error in C++).

// AddVar is SimpSolver::addVar (adds enough variables to make sure there is variable v).
// C: void addVar(Var v)
func (s *Gluco2SimpSolver) AddVar(v Gluco2Var) {
	for v >= s.NVars() {
		s.NewVar(true, true)
	}
}

// Reset is SimpSolver::reset (virtual override).
// C: virtual void reset()
func (s *Gluco2SimpSolver) Reset() {
	s.Gluco2Solver.Reset()
	s.Grow = int(opt_grow_glucose2.Value())
	s.Asymm_lits, s.Eliminated_vars, s.bwdsub_assigns, s.n_touched = 0, 0, 0, 0
	s.elimclauses.Clear(false)
	s.touched.Clear(false)
	s.occurs.Clear(false)
	s.n_occ.Clear(false)
	s.elim_heap.Clear(false)
	s.subsumption_queue.Clear(false)
	s.frozen.Clear(false)
	s.eliminated.Clear(false)
	dummy := NewGluco2VecPad(1, Gluco2LitUndef)
	s.ca.ExtraClauseField = true // NOTE: must happen before allocating the dummy clause below.
	s.bwdsub_tmpunit = s.ca.Alloc(dummy, false)
	s.remove_satisfied = false
}
