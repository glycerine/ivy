package abc

import (
	"fmt"
	"math"
)

// Port of src/sat/glucose/SimpSolver.h and SimpSolver.cpp (namespace Gluco).
// SimpSolver derives from Solver: it embeds GlucoSolver and re-declares the
// overridden methods (NewVar, AddClause*, Solve*, ToDimacs*, GarbageCollect, ...).

// GlucoSimpSolverElimLt is SimpSolver::ElimLt.
type GlucoSimpSolverElimLt struct {
	n_occ *GlucoVec[int] // const vec<int>&
}

// NewGlucoSimpSolverElimLt is the explicit constructor ElimLt(const vec<int>& no).
// C: explicit ElimLt(const vec<int>& no)
func NewGlucoSimpSolverElimLt(no *GlucoVec[int]) GlucoSimpSolverElimLt {
	return GlucoSimpSolverElimLt{n_occ: no}
}

// cost is ElimLt::cost.
// C: uint64_t cost(Var x) const
func (e GlucoSimpSolverElimLt) cost(x GlucoVar) uint64 {
	return uint64(*e.n_occ.At(GlucoToIntLit(GlucoMkLit(x, false)))) * uint64(*e.n_occ.At(GlucoToIntLit(GlucoLitNot(GlucoMkLit(x, false)))))
}

// Call is ElimLt::operator().
// C: bool operator()(Var x, Var y) const
func (e GlucoSimpSolverElimLt) Call(x int, y int) bool { return e.cost(x) < e.cost(y) }

// GlucoSimpSolverClauseDeleted is SimpSolver::ClauseDeleted.
type GlucoSimpSolverClauseDeleted struct {
	ca *GlucoClauseAllocator // const ClauseAllocator&
}

// NewGlucoSimpSolverClauseDeleted is the explicit constructor ClauseDeleted(const ClauseAllocator& _ca).
// C: explicit ClauseDeleted(const ClauseAllocator& _ca)
func NewGlucoSimpSolverClauseDeleted(ca *GlucoClauseAllocator) GlucoSimpSolverClauseDeleted {
	return GlucoSimpSolverClauseDeleted{ca: ca}
}

// Call is ClauseDeleted::operator().
// C: bool operator()(const CRef& cr) const
func (d GlucoSimpSolverClauseDeleted) Call(cr GlucoCRef) bool { return d.ca.At(cr).Mark() == 1 }

// glucoSimpSolverCat is `static const char* _cat = "SIMP"`.
const glucoSimpSolverCat = "SIMP"

// Command-line options of the simplifier (static objects in C++).
var (
	opt_use_asymm_glucose         = NewGlucoBoolOption(glucoSimpSolverCat, "asymm", "Shrink clauses by asymmetric branching.", false)
	opt_use_rcheck_glucose        = NewGlucoBoolOption(glucoSimpSolverCat, "rcheck", "Check if a clause is already implied. (costly)", false)
	opt_use_elim_glucose          = NewGlucoBoolOption(glucoSimpSolverCat, "elim", "Perform variable elimination.", true)
	opt_grow_glucose              = NewGlucoIntOption(glucoSimpSolverCat, "grow", "Allow a variable elimination step to grow by a number of clauses.", 0, NewGlucoIntRange(math.MinInt32, math.MaxInt32))
	opt_clause_lim_glucose        = NewGlucoIntOption(glucoSimpSolverCat, "cl-lim", "Variables are not eliminated if it produces a resolvent with a length above this limit. -1 means no limit", 20, NewGlucoIntRange(-1, math.MaxInt32))
	opt_subsumption_lim_glucose   = NewGlucoIntOption(glucoSimpSolverCat, "sub-lim", "Do not check if subsumption against a clause larger than this. -1 means no limit.", 1000, NewGlucoIntRange(-1, math.MaxInt32))
	opt_simp_garbage_frac_glucose = NewGlucoDoubleOption(glucoSimpSolverCat, "simp-gc-frac", "The fraction of wasted memory allowed before a garbage collection is triggered during simplification.", 0.5, NewGlucoDoubleRange(0, false, math.Inf(1), false))
)

// GlucoSimpSolver is Gluco::SimpSolver : public Solver.
type GlucoSimpSolver struct {
	GlucoSolver

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
	elimclauses        GlucoVec[uint32]
	touched            GlucoVec[int8]
	occurs             GlucoOccLists[int, GlucoCRef, GlucoSimpSolverClauseDeleted]
	n_occ              GlucoVec[int]
	elim_heap          GlucoHeap[GlucoSimpSolverElimLt]
	subsumption_queue  GlucoQueue[GlucoCRef]
	frozen             GlucoVec[int8]
	eliminated         GlucoVec[int8]
	bwdsub_assigns     int
	n_touched          int
	bwdsub_tmpunit     GlucoCRef
}

// NewGlucoSimpSolver is SimpSolver::SimpSolver().
// C: SimpSolver::SimpSolver()
func NewGlucoSimpSolver() *GlucoSimpSolver {
	s := &GlucoSimpSolver{}
	s.initGlucoSolver()
	s.virtGC = s.GarbageCollect // GarbageCollect is virtual
	s.Grow = int(opt_grow_glucose.Value())
	s.Clause_lim = int(opt_clause_lim_glucose.Value())
	s.Subsumption_lim = int(opt_subsumption_lim_glucose.Value())
	s.Simp_garbage_frac = opt_simp_garbage_frac_glucose.Value()
	s.Use_asymm = opt_use_asymm_glucose.Value()
	s.Use_rcheck = opt_use_rcheck_glucose.Value()
	s.Use_elim = opt_use_elim_glucose.Value()
	s.Merges = 0
	s.Asymm_lits = 0
	s.Eliminated_vars = 0
	s.elimorder = 1
	s.use_simplification = true
	s.occurs = *NewGlucoOccLists[int, GlucoCRef, GlucoSimpSolverClauseDeleted](NewGlucoSimpSolverClauseDeleted(&s.ca))
	s.elim_heap = *NewGlucoHeap(NewGlucoSimpSolverElimLt(&s.n_occ))
	s.bwdsub_assigns = 0
	s.n_touched = 0
	s.subsumption_queue = *NewGlucoQueue[GlucoCRef]()

	dummy := NewGlucoVecPad(1, GlucoLitUndef)
	s.ca.ExtraClauseField = true // NOTE: must happen before allocating the dummy clause below.
	s.bwdsub_tmpunit = s.ca.Alloc(dummy, false)
	s.remove_satisfied = false
	return s
}

// Destroy is SimpSolver::~SimpSolver().
// C: SimpSolver::~SimpSolver()
func (s *GlucoSimpSolver) Destroy() {}

// NewVar is SimpSolver::newVar (override; polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *GlucoSimpSolver) NewVar(sign bool, dvar bool) GlucoVar {
	v := s.GlucoSolver.NewVar(sign, dvar)

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
func (s *GlucoSimpSolver) IsEliminated(v GlucoVar) bool {
	if s.eliminated.Size() > 0 {
		return *s.eliminated.At(v) != 0
	}
	return false
}

// updateElimHeap is SimpSolver::updateElimHeap (protected).
// C: void updateElimHeap(Var v)
func (s *GlucoSimpSolver) updateElimHeap(v GlucoVar) {
	if !s.use_simplification {
		panic("GlucoSimpSolver.updateElimHeap: !use_simplification")
	}
	if s.elim_heap.InHeap(v) || (*s.frozen.At(v) == 0 && !s.IsEliminated(v) && s.Value(v) == GlucoLUndef) {
		s.elim_heap.Update(v)
	}
}

// SetFrozen is SimpSolver::setFrozen (a frozen variable is not eliminated).
// C: void setFrozen(Var v, bool b)
func (s *GlucoSimpSolver) SetFrozen(v GlucoVar, b bool) {
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
func (s *GlucoSimpSolver) solve_(doSimp bool, turnOffSimp bool) GlucoLbool {
	var extraFrozen GlucoVec[GlucoVar]
	result := GlucoLTrue

	doSimp = doSimp && s.use_simplification

	if doSimp {
		// Assumptions must be temporarily frozen to run variable elimination:
		for i := 0; i < s.assumptions.Size(); i++ {
			v := GlucoLitVar(*s.assumptions.At(i))

			// If an assumption has been eliminated, remember it.
			if s.IsEliminated(v) {
				panic("GlucoSimpSolver.solve_: assumption has been eliminated")
			}

			if *s.frozen.At(v) == 0 {
				// Freeze and store.
				s.SetFrozen(v, true)
				extraFrozen.PushElem(v)
			}
		}

		result = NewGlucoLboolB(s.Eliminate(turnOffSimp))
	}

	if result == GlucoLTrue {
		result = s.GlucoSolver.solve_()
	} else if s.Verbosity >= 1 {
		fmt.Printf("===============================================================================\n")
	}

	if result == GlucoLTrue {
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
func (s *GlucoSimpSolver) AddClause_(ps *GlucoVec[GlucoLit]) bool {
	for i := 0; i < ps.Size(); i++ { // #ifndef NDEBUG
		if s.IsEliminated(GlucoLitVar(*ps.At(i))) {
			panic("GlucoSimpSolver.AddClause_: literal of an eliminated variable")
		}
	}

	nclauses := s.clauses.Size()

	if s.Use_rcheck && s.implied(ps) {
		return true
	}

	if !s.GlucoSolver.AddClause_(ps) {
		return false
	}

	if s.Parsing == 0 && s.CertifiedUNSAT {
		for i := 0; i < ps.Size(); i++ {
			fmt.Fprintf(s.CertifiedOutput, "%d ", glucoDimacsLit(*ps.At(i)))
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
			s.occurs.At(GlucoLitVar(*c.At(i))).PushElem(cr)
			*s.n_occ.At(GlucoToIntLit(*c.At(i)))++
			*s.touched.At(GlucoLitVar(*c.At(i))) = 1
			s.n_touched++
			if s.elim_heap.InHeap(GlucoLitVar(*c.At(i))) {
				s.elim_heap.Increase(GlucoLitVar(*c.At(i)))
			}
		}
	}

	return true
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *GlucoSimpSolver) AddClause(ps *GlucoVec[GlucoLit]) bool {
	ps.CopyTo(&s.add_tmp)
	return s.AddClause_(&s.add_tmp)
}

// AddEmptyClause is SimpSolver::addEmptyClause.
// C: bool addEmptyClause()
func (s *GlucoSimpSolver) AddEmptyClause() bool {
	s.add_tmp.Clear(false)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *GlucoSimpSolver) AddClauseUnit(p GlucoLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *GlucoSimpSolver) AddClauseBinary(p GlucoLit, q GlucoLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *GlucoSimpSolver) AddClauseTernary(p GlucoLit, q GlucoLit, r GlucoLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	s.add_tmp.PushElem(r)
	return s.AddClause_(&s.add_tmp)
}

// removeClause is SimpSolver::removeClause (protected; hides Solver::removeClause).
// C: void removeClause(CRef cr)
func (s *GlucoSimpSolver) removeClause(cr GlucoCRef) {
	c := s.ca.At(cr)

	if s.use_simplification {
		for i := 0; i < c.Size(); i++ {
			*s.n_occ.At(GlucoToIntLit(*c.At(i)))--
			s.updateElimHeap(GlucoLitVar(*c.At(i)))
			s.occurs.Smudge(GlucoLitVar(*c.At(i)))
		}
	}

	s.GlucoSolver.removeClause(cr)
}

// strengthenClause is SimpSolver::strengthenClause (protected).
// C: bool strengthenClause(CRef cr, Lit l)
func (s *GlucoSimpSolver) strengthenClause(cr GlucoCRef, l GlucoLit) bool {
	c := s.ca.At(cr)
	if s.decisionLevel() != 0 || !s.use_simplification {
		panic("GlucoSimpSolver.strengthenClause: precondition")
	}

	// FIX: this is too inefficient but would be nice to have (properly implemented)
	// if (!find(subsumption_queue, &c))
	s.subsumption_queue.Insert(cr)

	if s.CertifiedUNSAT {
		for i := 0; i < c.Size(); i++ {
			if *c.At(i) != l {
				fmt.Fprintf(s.CertifiedOutput, "%d ", glucoDimacsLit(*c.At(i)))
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
				fmt.Fprintf(s.CertifiedOutput, "%d ", glucoDimacsLit(*c.At(i)))
			}
			fmt.Fprintf(s.CertifiedOutput, "0\n")
		}

		s.detachClause(cr, true)
		c.Strengthen(l)
		s.attachClause(cr)
		glucoRemove[GlucoCRef](s.occurs.At(GlucoLitVar(l)), cr, glucoNe[GlucoCRef])
		*s.n_occ.At(GlucoToIntLit(l))--
		s.updateElimHeap(GlucoLitVar(l))
	}

	if c.Size() == 1 {
		return s.enqueue(*c.At(0), GlucoCRefUndef) && s.propagate() == GlucoCRefUndef
	}
	return true
}

// merge is the overload merge(..., vec<Lit>& out_clause) (protected): returns
// false if the clause is always satisfied ('out_clause' should not be used).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, vec<Lit>& out_clause)
func (s *GlucoSimpSolver) merge(_ps *GlucoClause, _qs *GlucoClause, v GlucoVar, outClause *GlucoVec[GlucoLit]) bool {
	s.Merges++
	outClause.Clear(false)

	psSmallest := _ps.Size() < _qs.Size()
	ps, qs := _ps, _qs
	if psSmallest {
		ps, qs = _qs, _ps
	}
	var i, j int

	for i = 0; i < qs.Size(); i++ {
		if GlucoLitVar(*qs.At(i)) != v {
			found := false
			for j = 0; j < ps.Size(); j++ {
				if GlucoLitVar(*ps.At(j)) == GlucoLitVar(*qs.At(i)) {
					if *ps.At(j) == GlucoLitNot(*qs.At(i)) {
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
		if GlucoLitVar(*ps.At(i)) != v {
			outClause.PushElem(*ps.At(i))
		}
	}

	return true
}

// mergeSize is the overload merge(..., int& size) (protected): returns false if
// the clause is always satisfied.
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, int& size)
func (s *GlucoSimpSolver) mergeSize(_ps *GlucoClause, _qs *GlucoClause, v GlucoVar, size *int) bool {
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
		if GlucoLitVar(qs_[i]) != v {
			found := false
			for j := 0; j < ps.Size(); j++ {
				if GlucoLitVar(ps_[j]) == GlucoLitVar(qs_[i]) {
					if ps_[j] == GlucoLitNot(qs_[i]) {
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
func (s *GlucoSimpSolver) gatherTouchedClauses() {
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
func (s *GlucoSimpSolver) implied(c *GlucoVec[GlucoLit]) bool {
	if s.decisionLevel() != 0 {
		panic("GlucoSimpSolver.implied: decisionLevel() != 0")
	}

	s.trail_lim.PushElem(s.trail.Size())
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) == GlucoLTrue {
			s.cancelUntil(0)
			return false
		} else if s.ValueLit(*c.At(i)) != GlucoLFalse {
			if s.ValueLit(*c.At(i)) != GlucoLUndef {
				panic("GlucoSimpSolver.implied: value != l_Undef")
			}
			s.uncheckedEnqueue(GlucoLitNot(*c.At(i)), GlucoCRefUndef)
		}
	}

	result := s.propagate() != GlucoCRefUndef
	s.cancelUntil(0)
	return result
}

// backwardSubsumptionCheck is SimpSolver::backwardSubsumptionCheck (protected):
// backward subsumption + backward subsumption resolution.
// C: bool backwardSubsumptionCheck(bool verbose = false)
func (s *GlucoSimpSolver) backwardSubsumptionCheck(verbose bool) bool {
	cnt := 0
	subsumed := 0
	deletedLiterals := 0
	if s.decisionLevel() != 0 {
		panic("GlucoSimpSolver.backwardSubsumptionCheck: decisionLevel() != 0")
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

		if !(c.Size() > 1 || s.ValueLit(*c.At(0)) == GlucoLTrue) { // Unit-clauses should have been propagated before this point.
			panic("GlucoSimpSolver.backwardSubsumptionCheck: unit clause not propagated")
		}

		// Find best variable to scan:
		best := GlucoLitVar(*c.At(0))
		for i := 1; i < c.Size(); i++ {
			if s.occurs.At(GlucoLitVar(*c.At(i))).Size() < s.occurs.At(best).Size() {
				best = GlucoLitVar(*c.At(i))
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

				if l == GlucoLitUndef {
					subsumed++
					s.removeClause(cs[j])
				} else if l != GlucoLitError {
					deletedLiterals++

					if !s.strengthenClause(cs[j], GlucoLitNot(l)) {
						return false
					}

					// Did current candidate get deleted from cs? Then check candidate at index j again:
					if GlucoLitVar(l) == best {
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
func (s *GlucoSimpSolver) asymm(v GlucoVar, cr GlucoCRef) bool {
	c := s.ca.At(cr)
	if s.decisionLevel() != 0 {
		panic("GlucoSimpSolver.asymm: decisionLevel() != 0")
	}

	if c.Mark() != 0 || s.satisfied(c) {
		return true
	}

	s.trail_lim.PushElem(s.trail.Size())
	l := GlucoLitUndef
	for i := 0; i < c.Size(); i++ {
		if GlucoLitVar(*c.At(i)) != v && s.ValueLit(*c.At(i)) != GlucoLFalse {
			s.uncheckedEnqueue(GlucoLitNot(*c.At(i)), GlucoCRefUndef)
		} else {
			l = *c.At(i)
		}
	}

	if s.propagate() != GlucoCRefUndef {
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
func (s *GlucoSimpSolver) asymmVar(v GlucoVar) bool {
	if !s.use_simplification {
		panic("GlucoSimpSolver.asymmVar: !use_simplification")
	}

	cls := s.occurs.Lookup(v)

	if s.Value(v) != GlucoLUndef || cls.Size() == 0 {
		return true
	}

	for i := 0; i < cls.Size(); i++ {
		if !s.asymm(v, *cls.At(i)) {
			return false
		}
	}

	return s.backwardSubsumptionCheck(false)
}

// mkElimClauseLit_glucose is the overload mkElimClause(vec<uint32_t>&, Lit) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Lit x)
func mkElimClauseLit_glucose(elimclauses *GlucoVec[uint32], x GlucoLit) {
	elimclauses.PushElem(uint32(GlucoToIntLit(x)))
	elimclauses.PushElem(1)
}

// mkElimClauseVarClause_glucose is the overload mkElimClause(vec<uint32_t>&, Var, Clause&) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Var v, Clause& c)
func mkElimClauseVarClause_glucose(elimclauses *GlucoVec[uint32], v GlucoVar, c *GlucoClause) {
	first := elimclauses.Size()
	vPos := -1

	// Copy clause to elimclauses-vector. Remember position where the variable 'v' occurs:
	for i := 0; i < c.Size(); i++ {
		elimclauses.PushElem(uint32(GlucoToIntLit(*c.At(i))))
		if GlucoLitVar(*c.At(i)) == v {
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
func (s *GlucoSimpSolver) eliminateVar(v GlucoVar) bool {
	if *s.frozen.At(v) != 0 || s.IsEliminated(v) || s.Value(v) != GlucoLUndef {
		panic("GlucoSimpSolver.eliminateVar: precondition")
	}
	var i int

	// Split the occurrences into positive and negative:
	cls := s.occurs.Lookup(v)
	var pos, neg GlucoVec[GlucoCRef]
	for i = 0; i < cls.Size(); i++ {
		if glucoFind[GlucoLit](s.ca.At(*cls.At(i)), GlucoMkLit(v, false), glucoNe[GlucoLit]) {
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
			mkElimClauseVarClause_glucose(&s.elimclauses, v, s.ca.At(*neg.At(i)))
		}
		mkElimClauseLit_glucose(&s.elimclauses, GlucoMkLit(v, false))
		s.Eliminated_clauses += neg.Size()
	} else {
		for i := 0; i < pos.Size(); i++ {
			mkElimClauseVarClause_glucose(&s.elimclauses, v, s.ca.At(*pos.At(i)))
		}
		mkElimClauseLit_glucose(&s.elimclauses, GlucoLitNot(GlucoMkLit(v, false)))
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
	if s.watches.At(GlucoMkLit(v, false)).Size() == 0 {
		s.watches.At(GlucoMkLit(v, false)).Clear(true)
	}
	if s.watches.At(GlucoLitNot(GlucoMkLit(v, false))).Size() == 0 {
		s.watches.At(GlucoLitNot(GlucoMkLit(v, false))).Clear(true)
	}

	return s.backwardSubsumptionCheck(false)
}

// Substitute is SimpSolver::substitute (replace all occurences of v with x; may cause a contradiction).
// C: bool substitute(Var v, Lit x)
func (s *GlucoSimpSolver) Substitute(v GlucoVar, x GlucoLit) bool {
	if *s.frozen.At(v) != 0 || s.IsEliminated(v) || s.Value(v) != GlucoLUndef {
		panic("GlucoSimpSolver.Substitute: precondition")
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
			if GlucoLitVar(p) == v {
				substClause.PushElem(GlucoLitXor(x, GlucoSign(p)))
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
func (s *GlucoSimpSolver) extendModel() {
	var i, j int

	for i = s.elimclauses.Size() - 1; i > 0; i -= j {
		j = int(*s.elimclauses.At(i))
		i--
		skip := false
		for ; j > 1; j, i = j-1, i-1 {
			if s.ModelValueLit(GlucoToLit(int(*s.elimclauses.At(i)))) != GlucoLFalse {
				skip = true // goto next
				break
			}
		}
		if skip {
			continue
		}
		x := GlucoToLit(int(*s.elimclauses.At(i)))
		*s.Model.At(GlucoLitVar(x)) = NewGlucoLboolB(!GlucoSign(x))
	}
}

// Eliminate is SimpSolver::eliminate (variable-elimination based simplification).
// C: bool eliminate(bool turn_off_elim = false)
func (s *GlucoSimpSolver) Eliminate(turnOffElim bool) bool {
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
					panic("GlucoSimpSolver.Eliminate: queues not empty on interrupt")
				}
				s.elim_heap.Clear(false)
				return
			}

			for cnt := 0; !s.elim_heap.Empty(); cnt++ {
				elim := s.elim_heap.RemoveMin()

				if s.asynch_interrupt {
					break
				}

				if s.IsEliminated(elim) || s.Value(elim) != GlucoLUndef {
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
				if s.Use_elim && s.Value(elim) == GlucoLUndef && *s.frozen.At(elim) == 0 && !s.eliminateVar(elim) {
					s.ok = false
					return
				}

				s.CheckGarbageFrac(s.Simp_garbage_frac)
			}

			if s.subsumption_queue.Size() != 0 {
				panic("GlucoSimpSolver.Eliminate: subsumption queue not empty")
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
func (s *GlucoSimpSolver) cleanUpClauses() {
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
func (s *GlucoSimpSolver) relocAll(to *GlucoClauseAllocator) {
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
func (s *GlucoSimpSolver) GarbageCollect() {
	// Initialize the next region to a size corresponding to the estimated utilization degree. This
	// is not precise but should avoid some unnecessary reallocations for the new region:
	to := NewGlucoClauseAllocatorCap(s.ca.Size() - s.ca.Wasted())

	s.cleanUpClauses()
	to.ExtraClauseField = s.ca.ExtraClauseField // NOTE: this is important to keep (or lose) the extra fields.
	s.relocAll(to)
	s.GlucoSolver.relocAll(to)
	if s.Verbosity >= 2 {
		fmt.Printf("|  Garbage collection:   %12d bytes => %12d bytes             |\n",
			s.ca.Size()*GlucoRegionAllocatorUnitSize, to.Size()*GlucoRegionAllocatorUnitSize)
	}
	to.MoveTo(&s.ca)
}

// Solve is the overload solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) Solve(assumps *GlucoVec[GlucoLit], doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	assumps.CopyTo(&s.assumptions)
	return s.solve_(doSimp, turnOffSimp) == GlucoLTrue
}

// SolveLimited is SimpSolver::solveLimited.
// C: lbool solveLimited(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) SolveLimited(assumps *GlucoVec[GlucoLit], doSimp bool, turnOffSimp bool) GlucoLbool {
	assumps.CopyTo(&s.assumptions)
	return s.solve_(doSimp, turnOffSimp)
}

// SolveNoAssumps is the overload solve(bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) SolveNoAssumps(doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	return s.solve_(doSimp, turnOffSimp) == GlucoLTrue
}

// Solve1 is the overload solve(Lit p, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) Solve1(p GlucoLit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	return s.solve_(doSimp, turnOffSimp) == GlucoLTrue
}

// Solve2 is the overload solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) Solve2(p GlucoLit, q GlucoLit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	return s.solve_(doSimp, turnOffSimp) == GlucoLTrue
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) Solve3(p GlucoLit, q GlucoLit, r GlucoLit, doSimp bool, turnOffSimp bool) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	s.assumptions.PushElem(r)
	return s.solve_(doSimp, turnOffSimp) == GlucoLTrue
}

// The five toDimacs(...) overloads declared in SimpSolver.h are never defined in
// SimpSolver.cpp, so they are not ported (calling them is a link error in C++).

// AddVar is SimpSolver::addVar (adds enough variables to make sure there is variable v).
// C: void addVar(Var v)
func (s *GlucoSimpSolver) AddVar(v GlucoVar) {
	for v >= s.NVars() {
		s.NewVar(true, true)
	}
}

// Reset is SimpSolver::reset (virtual override).
// C: virtual void reset()
func (s *GlucoSimpSolver) Reset() {
	s.GlucoSolver.Reset()
	s.Grow = int(opt_grow_glucose.Value())
	s.Asymm_lits, s.Eliminated_vars, s.bwdsub_assigns, s.n_touched = 0, 0, 0, 0
	s.elimclauses.Clear(false)
	s.touched.Clear(false)
	s.occurs.Clear(false)
	s.n_occ.Clear(false)
	s.elim_heap.Clear(false)
	s.subsumption_queue.Clear(false)
	s.frozen.Clear(false)
	s.eliminated.Clear(false)
	dummy := NewGlucoVecPad(1, GlucoLitUndef)
	s.ca.ExtraClauseField = true // NOTE: must happen before allocating the dummy clause below.
	s.bwdsub_tmpunit = s.ca.Alloc(dummy, false)
	s.remove_satisfied = false
}
