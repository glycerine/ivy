package abc

import (
	"fmt"
	"math"
	"os"
)

// Port of src/sat/bsat2/Solver.h and Solver.cpp (namespace Minisat).
// See sat_bsat2_vec.go for the C++ -> Go conventions.

// MinisatSolverVarData is Solver::VarData.
type MinisatSolverVarData struct {
	reason MinisatCRef
	level  int
}

// minisatSolverMkVarData is Solver::mkVarData (protected static inline).
// C: static inline VarData mkVarData(CRef cr, int l)
func minisatSolverMkVarData(cr MinisatCRef, l int) MinisatSolverVarData {
	return MinisatSolverVarData{reason: cr, level: l}
}

// MinisatSolverWatcher is Solver::Watcher.
type MinisatSolverWatcher struct {
	cref    MinisatCRef
	blocker MinisatLit
}

// NewMinisatSolverWatcher is Watcher(CRef cr, Lit p).
// C: Watcher(CRef cr, Lit p)
func NewMinisatSolverWatcher(cr MinisatCRef, p MinisatLit) MinisatSolverWatcher {
	return MinisatSolverWatcher{cref: cr, blocker: p}
}

// Eq is Watcher::operator==.
// C: bool operator==(const Watcher& w) const
func (w MinisatSolverWatcher) Eq(o MinisatSolverWatcher) bool { return w.cref == o.cref }

// Ne is Watcher::operator!=.
// C: bool operator!=(const Watcher& w) const
func (w MinisatSolverWatcher) Ne(o MinisatSolverWatcher) bool { return w.cref != o.cref }

// MinisatSolverWatcherDeleted is Solver::WatcherDeleted.
type MinisatSolverWatcherDeleted struct {
	ca *MinisatClauseAllocator // const ClauseAllocator&
}

// NewMinisatSolverWatcherDeleted is WatcherDeleted(const ClauseAllocator& _ca).
// C: WatcherDeleted(const ClauseAllocator& _ca)
func NewMinisatSolverWatcherDeleted(ca *MinisatClauseAllocator) MinisatSolverWatcherDeleted {
	return MinisatSolverWatcherDeleted{ca: ca}
}

// Call is WatcherDeleted::operator().
// C: bool operator()(const Watcher& w) const
func (d MinisatSolverWatcherDeleted) Call(w MinisatSolverWatcher) bool {
	return d.ca.At(w.cref).Mark() == 1
}

// MinisatSolverVarOrderLt is Solver::VarOrderLt.
type MinisatSolverVarOrderLt struct {
	activity *MinisatVec[float64] // const vec<double>&
}

// NewMinisatSolverVarOrderLt is VarOrderLt(const vec<double>& act).
// C: VarOrderLt(const vec<double>& act)
func NewMinisatSolverVarOrderLt(act *MinisatVec[float64]) MinisatSolverVarOrderLt {
	return MinisatSolverVarOrderLt{activity: act}
}

// Call is VarOrderLt::operator().
// C: bool operator()(Var x, Var y) const
func (l MinisatSolverVarOrderLt) Call(x int, y int) bool {
	return *l.activity.At(x) > *l.activity.At(y)
}

// minisatSolverCat is `static const char* _cat = "CORE"`.
const minisatSolverCat = "CORE"

// Command-line options of the solver (static objects in C++: constructed when the
// program starts, registering themselves in Option::getOptionList()).
var (
	opt_var_decay_minisat       = NewMinisatDoubleOption(minisatSolverCat, "var-decay", "The variable activity decay factor", 0.95, NewMinisatDoubleRange(0, false, 1, false))
	opt_clause_decay_minisat    = NewMinisatDoubleOption(minisatSolverCat, "cla-decay", "The clause activity decay factor", 0.999, NewMinisatDoubleRange(0, false, 1, false))
	opt_random_var_freq_minisat = NewMinisatDoubleOption(minisatSolverCat, "rnd-freq", "The frequency with which the decision heuristic tries to choose a random variable", 0, NewMinisatDoubleRange(0, true, 1, true))
	opt_random_seed_minisat     = NewMinisatDoubleOption(minisatSolverCat, "rnd-seed", "Used by the random variable selection", 91648253, NewMinisatDoubleRange(0, false, math.Inf(1), false))
	opt_ccmin_mode_minisat      = NewMinisatIntOption(minisatSolverCat, "ccmin-mode", "Controls conflict clause minimization (0=none, 1=basic, 2=deep)", 2, NewMinisatIntRange(0, 2))
	opt_phase_saving_minisat    = NewMinisatIntOption(minisatSolverCat, "phase-saving", "Controls the level of phase saving (0=none, 1=limited, 2=full)", 2, NewMinisatIntRange(0, 2))
	opt_rnd_init_act_minisat    = NewMinisatBoolOption(minisatSolverCat, "rnd-init", "Randomize the initial activity", false)
	opt_luby_restart_minisat    = NewMinisatBoolOption(minisatSolverCat, "luby", "Use the Luby restart sequence", true)
	opt_restart_first_minisat   = NewMinisatIntOption(minisatSolverCat, "rfirst", "The base restart interval", 100, NewMinisatIntRange(1, math.MaxInt32))
	opt_restart_inc_minisat     = NewMinisatDoubleOption(minisatSolverCat, "rinc", "Restart interval increase factor", 2, NewMinisatDoubleRange(1, false, math.Inf(1), false))
	opt_garbage_frac_minisat    = NewMinisatDoubleOption(minisatSolverCat, "gc-frac", "The fraction of wasted memory allowed before a garbage collection is triggered", 0.20, NewMinisatDoubleRange(0, false, math.Inf(1), false))
)

// MinisatSolver is Minisat::Solver, the CDCL SAT solver. Public data members are
// exported; protected ones keep the C++ spelling. SimpSolver embeds it.
type MinisatSolver struct {
	// If problem is satisfiable, this vector contains the model (if any).
	Model MinisatVec[MinisatLbool]
	// If problem is unsatisfiable (possibly under assumptions), this vector represents the final conflict clause expressed in the assumptions.
	Conflict MinisatVec[MinisatLit]

	// Mode of operation:
	Verbosity                     int
	Var_decay                     float64
	Clause_decay                  float64
	Random_var_freq               float64
	Random_seed                   float64
	Luby_restart                  bool
	Ccmin_mode                    int     // conflict clause minimization (0=none, 1=basic, 2=deep)
	Phase_saving                  int     // phase saving (0=none, 1=limited, 2=full)
	Rnd_pol                       bool    // use random polarities for branching heuristics
	Rnd_init_act                  bool    // initialize variable activities with a small random value
	Garbage_frac                  float64 // fraction of wasted memory allowed before a garbage collection is triggered
	Restart_first                 int     // initial restart limit (default 100)
	Restart_inc                   float64 // restart limit multiplier (default 1.5)
	Learntsize_factor             float64 // initial limit for learnt clauses (default 1/3)
	Learntsize_inc                float64 // learnt-clause limit multiplier (default 1.1)
	Learntsize_adjust_start_confl int
	Learntsize_adjust_inc         float64

	// Statistics: (read-only member variables)
	Solves, Starts, Decisions, Rnd_decisions, Propagations, Conflicts        uint64
	Dec_vars, Clauses_literals, Learnts_literals, Max_literals, Tot_literals uint64

	// Solver state (protected):
	ok                      bool                                                                           // if FALSE, the constraints are already unsatisfiable
	clauses                 MinisatVec[MinisatCRef]                                                        // list of problem clauses
	learnts                 MinisatVec[MinisatCRef]                                                        // list of learnt clauses
	cla_inc                 float64                                                                        // amount to bump next clause with
	activity                MinisatVec[float64]                                                            // heuristic measurement of the activity of a variable
	var_inc                 float64                                                                        // amount to bump next variable with
	watches                 MinisatOccLists[MinisatLit, MinisatSolverWatcher, MinisatSolverWatcherDeleted] // 'watches[lit]' is a list of constraints watching 'lit'
	assigns                 MinisatVec[MinisatLbool]                                                       // the current assignments
	polarity                MinisatVec[int8]                                                               // preferred polarity of each variable
	decision                MinisatVec[int8]                                                               // declares if a variable is eligible for selection in the decision heuristic
	trail                   MinisatVec[MinisatLit]                                                         // assignment stack
	trail_lim               MinisatVec[int]                                                                // separator indices for different decision levels in 'trail'
	vardata                 MinisatVec[MinisatSolverVarData]                                               // reason and level for each variable
	qhead                   int                                                                            // head of queue (index into the trail)
	simpDB_assigns          int                                                                            // top-level assignments since last execution of 'simplify()'
	simpDB_props            int64                                                                          // remaining propagations before next 'simplify()'
	assumptions             MinisatVec[MinisatLit]                                                         // current set of assumptions provided to solve by the user
	order_heap              MinisatHeap[MinisatSolverVarOrderLt]                                           // priority queue of variables ordered by activity
	progress_estimate       float64                                                                        // set by 'search()'
	remove_satisfied        bool                                                                           // whether the linear scan for satisfied clauses is performed in 'simplify'
	ca                      MinisatClauseAllocator
	seen                    MinisatVec[int8]
	analyze_stack           MinisatVec[MinisatLit]
	analyze_toclear         MinisatVec[MinisatLit]
	add_tmp                 MinisatVec[MinisatLit]
	max_learnts             float64
	learntsize_adjust_confl float64
	learntsize_adjust_cnt   int
	conflict_budget         int64 // -1 means no budget
	propagation_budget      int64 // -1 means no budget
	asynch_interrupt        bool

	// virtGC models the virtual call garbageCollect(): a derived class that
	// overrides it (SimpSolver) sets this to its own method.
	virtGC func()
}

// NewMinisatSolver is Solver::Solver().
// C: Solver::Solver()
func NewMinisatSolver() *MinisatSolver {
	s := &MinisatSolver{}
	s.initMinisatSolver()
	return s
}

// initMinisatSolver is the body of the Solver constructor's member-initialiser
// list; it is also what the SimpSolver constructor runs for its base class.
func (s *MinisatSolver) initMinisatSolver() {
	// Parameters (user settable):
	s.Verbosity = 0
	s.Var_decay = opt_var_decay_minisat.Value()
	s.Clause_decay = opt_clause_decay_minisat.Value()
	s.Random_var_freq = opt_random_var_freq_minisat.Value()
	s.Random_seed = opt_random_seed_minisat.Value()
	s.Luby_restart = opt_luby_restart_minisat.Value()
	s.Ccmin_mode = int(opt_ccmin_mode_minisat.Value())
	s.Phase_saving = int(opt_phase_saving_minisat.Value())
	s.Rnd_pol = false
	s.Rnd_init_act = opt_rnd_init_act_minisat.Value()
	s.Garbage_frac = opt_garbage_frac_minisat.Value()
	s.Restart_first = int(opt_restart_first_minisat.Value())
	s.Restart_inc = opt_restart_inc_minisat.Value()
	// Parameters (the rest):
	s.Learntsize_factor = float64(1) / float64(3)
	s.Learntsize_inc = 1.1
	// Parameters (experimental):
	s.Learntsize_adjust_start_confl = 100
	s.Learntsize_adjust_inc = 1.5

	s.ok = true
	s.cla_inc = 1
	s.var_inc = 1
	s.ca = *NewMinisatClauseAllocator()
	s.watches = *NewMinisatOccLists[MinisatLit, MinisatSolverWatcher, MinisatSolverWatcherDeleted](NewMinisatSolverWatcherDeleted(&s.ca))
	s.qhead = 0
	s.simpDB_assigns = -1
	s.simpDB_props = 0
	s.order_heap = *NewMinisatHeap(NewMinisatSolverVarOrderLt(&s.activity))
	s.progress_estimate = 0
	s.remove_satisfied = true
	// Resource constraints:
	s.conflict_budget = -1
	s.propagation_budget = -1
	s.asynch_interrupt = false
}

// Destroy is Solver::~Solver() (virtual).
// C: virtual Solver::~Solver()
func (s *MinisatSolver) Destroy() {}

// NewVar is Solver::newVar (polarity and dvar default to true). Creates a new SAT
// variable; if 'dvar' is cleared, the variable will not be used as a decision variable.
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *MinisatSolver) NewVar(sign bool, dvar bool) MinisatVar {
	v := s.NVars()
	s.watches.Init(MinisatMkLit(v, false))
	s.watches.Init(MinisatMkLit(v, true))
	s.assigns.PushElem(MinisatLUndef)
	s.vardata.PushElem(minisatSolverMkVarData(MinisatCRefUndef, 0))
	if s.Rnd_init_act {
		s.activity.PushElem(minisatSolverDrand(&s.Random_seed) * 0.00001)
	} else {
		s.activity.PushElem(0)
	}
	s.seen.PushElem(0)
	if sign {
		s.polarity.PushElem(1)
	} else {
		s.polarity.PushElem(0)
	}
	s.decision.Push()
	s.trail.SetCapacity(v + 1)
	s.SetDecisionVar(v, dvar)
	return v
}

// minisatLitLess is LessThan_default<Lit> (Lit::operator<).
type minisatLitLess struct{}

func (minisatLitLess) Call(x MinisatLit, y MinisatLit) bool { return x.Lt(y) }

// AddClause_ is Solver::addClause_ (no superfluous internal copy; clobbers ps).
// C: bool addClause_(vec<Lit>& ps)
func (s *MinisatSolver) AddClause_(ps *MinisatVec[MinisatLit]) bool {
	if s.decisionLevel() != 0 {
		panic("MinisatSolver.AddClause_: decisionLevel() != 0")
	}
	if !s.ok {
		return false
	}
	// Check if clause is satisfied and remove false/duplicate literals:
	MinisatSortVec(ps, minisatLitLess{})
	var p MinisatLit
	var i, j int
	for i, j, p = 0, 0, MinisatLitUndef; i < ps.Size(); i++ {
		if s.ValueLit(*ps.At(i)) == MinisatLTrue || *ps.At(i) == MinisatLitNot(p) {
			return true
		} else if s.ValueLit(*ps.At(i)) != MinisatLFalse && *ps.At(i) != p {
			p = *ps.At(i)
			*ps.At(j) = p
			j++
		}
	}
	ps.Shrink(i - j)
	if ps.Size() == 0 {
		s.ok = false
		return false
	} else if ps.Size() == 1 {
		s.uncheckedEnqueue(*ps.At(0), MinisatCRefUndef)
		s.ok = s.propagate() == MinisatCRefUndef
		return s.ok
	} else {
		cr := s.ca.Alloc(ps, false)
		s.clauses.PushElem(cr)
		s.attachClause(cr)
	}
	return true
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *MinisatSolver) AddClause(ps *MinisatVec[MinisatLit]) bool {
	ps.CopyTo(&s.add_tmp)
	return s.AddClause_(&s.add_tmp)
}

// AddEmptyClause is Solver::addEmptyClause.
// C: bool addEmptyClause()
func (s *MinisatSolver) AddEmptyClause() bool {
	s.add_tmp.Clear(false)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *MinisatSolver) AddClauseUnit(p MinisatLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *MinisatSolver) AddClauseBinary(p MinisatLit, q MinisatLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *MinisatSolver) AddClauseTernary(p MinisatLit, q MinisatLit, r MinisatLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	s.add_tmp.PushElem(r)
	return s.AddClause_(&s.add_tmp)
}

// attachClause is Solver::attachClause (protected).
// C: void attachClause(CRef cr)
func (s *MinisatSolver) attachClause(cr MinisatCRef) {
	c := s.ca.At(cr)
	if c.Size() <= 1 {
		panic("MinisatSolver.attachClause: size <= 1")
	}
	s.watches.At(MinisatLitNot(*c.At(0))).PushElem(NewMinisatSolverWatcher(cr, *c.At(1)))
	s.watches.At(MinisatLitNot(*c.At(1))).PushElem(NewMinisatSolverWatcher(cr, *c.At(0)))
	if c.Learnt() {
		s.Learnts_literals += uint64(c.Size())
	} else {
		s.Clauses_literals += uint64(c.Size())
	}
}

// detachClause is Solver::detachClause (protected; strict defaults to false).
// C: void detachClause(CRef cr, bool strict = false)
func (s *MinisatSolver) detachClause(cr MinisatCRef, strict bool) {
	c := s.ca.At(cr)
	if c.Size() <= 1 {
		panic("MinisatSolver.detachClause: size <= 1")
	}
	if strict {
		minisatRemove[MinisatSolverWatcher](s.watches.At(MinisatLitNot(*c.At(0))), NewMinisatSolverWatcher(cr, *c.At(1)), MinisatSolverWatcher.Ne)
		minisatRemove[MinisatSolverWatcher](s.watches.At(MinisatLitNot(*c.At(1))), NewMinisatSolverWatcher(cr, *c.At(0)), MinisatSolverWatcher.Ne)
	} else {
		// Lazy detaching: (NOTE! Must clean all watcher lists before garbage collecting this clause)
		s.watches.Smudge(MinisatLitNot(*c.At(0)))
		s.watches.Smudge(MinisatLitNot(*c.At(1)))
	}
	if c.Learnt() {
		s.Learnts_literals -= uint64(c.Size())
	} else {
		s.Clauses_literals -= uint64(c.Size())
	}
}

// removeClause is Solver::removeClause (protected).
// C: void removeClause(CRef cr)
func (s *MinisatSolver) removeClause(cr MinisatCRef) {
	c := s.ca.At(cr)
	s.detachClause(cr, false)
	// Don't leave pointers to free'd memory!
	if s.locked(c) {
		s.vardata.At(MinisatLitVar(*c.At(0))).reason = MinisatCRefUndef
	}
	c.SetMark(1)
	s.ca.Free(cr)
}

// satisfied is Solver::satisfied (protected).
// C: bool satisfied(const Clause& c) const
func (s *MinisatSolver) satisfied(c *MinisatClause) bool {
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) == MinisatLTrue {
			return true
		}
	}
	return false
}

// cancelUntil is Solver::cancelUntil (protected): revert to the state at given
// level (keeping all assignment at 'level' but not beyond).
// C: void cancelUntil(int level)
func (s *MinisatSolver) cancelUntil(level int) {
	if s.decisionLevel() > level {
		for c := s.trail.Size() - 1; c >= *s.trail_lim.At(level); c-- {
			x := MinisatLitVar(*s.trail.At(c))
			*s.assigns.At(x) = MinisatLUndef
			if s.Phase_saving > 1 || (s.Phase_saving == 1) && c > *s.trail_lim.Last() {
				if MinisatSign(*s.trail.At(c)) {
					*s.polarity.At(x) = 1
				} else {
					*s.polarity.At(x) = 0
				}
			}
			s.insertVarOrder(x)
		}
		s.qhead = *s.trail_lim.At(level)
		s.trail.Shrink(s.trail.Size() - *s.trail_lim.At(level))
		s.trail_lim.Shrink(s.trail_lim.Size() - level)
	}
}

// pickBranchLit is Solver::pickBranchLit (protected).
// C: Lit pickBranchLit()
func (s *MinisatSolver) pickBranchLit() MinisatLit {
	next := -1 // var_Undef
	// Random decision:
	if minisatSolverDrand(&s.Random_seed) < s.Random_var_freq && !s.order_heap.Empty() {
		next = s.order_heap.At(minisatSolverIrand(&s.Random_seed, s.order_heap.Size()))
		if s.Value(next) == MinisatLUndef && *s.decision.At(next) != 0 {
			s.Rnd_decisions++
		}
	}
	// Activity based decision:
	for next == -1 || s.Value(next) != MinisatLUndef || *s.decision.At(next) == 0 {
		if s.order_heap.Empty() {
			next = -1
			break
		}
		next = s.order_heap.RemoveMin()
	}
	if next == -1 {
		return MinisatLitUndef
	}
	var sg bool
	if s.Rnd_pol {
		sg = minisatSolverDrand(&s.Random_seed) < 0.5
	} else {
		sg = *s.polarity.At(next) != 0
	}
	return MinisatMkLit(next, sg)
}

// analyze is Solver::analyze (protected): analyze conflict and produce a reason clause.
//
//	Pre-conditions: 'out_learnt' is assumed to be cleared; current decision level must be greater than root level.
//	Post-conditions: 'out_learnt[0]' is the asserting literal at level 'out_btlevel'; if out_learnt.size() > 1
//	then 'out_learnt[1]' has the greatest decision level of the rest of literals.
//
// C: void analyze(CRef confl, vec<Lit>& out_learnt, int& out_btlevel)
func (s *MinisatSolver) analyze(confl MinisatCRef, outLearnt *MinisatVec[MinisatLit], outBtlevel *int) {
	pathC := 0
	p := MinisatLitUndef

	// Generate conflict clause:
	outLearnt.Push() // (leave room for the asserting literal)
	index := s.trail.Size() - 1

	for {
		if confl == MinisatCRefUndef { // (otherwise should be UIP)
			panic("MinisatSolver.analyze: confl == CRef_Undef")
		}
		c := s.ca.At(confl)

		if c.Learnt() {
			s.claBumpActivity(c)
		}

		j0 := 1
		if p == MinisatLitUndef {
			j0 = 0
		}
		for j := j0; j < c.Size(); j++ {
			q := *c.At(j)

			if *s.seen.At(MinisatLitVar(q)) == 0 && s.level(MinisatLitVar(q)) > 0 {
				s.varBumpActivity(MinisatLitVar(q))
				*s.seen.At(MinisatLitVar(q)) = 1
				if s.level(MinisatLitVar(q)) >= s.decisionLevel() {
					pathC++
				} else {
					outLearnt.PushElem(q)
				}
			}
		}

		// Select next clause to look at:
		for {
			v := MinisatLitVar(*s.trail.At(index))
			index--
			if *s.seen.At(v) != 0 {
				break
			}
		}
		p = *s.trail.At(index + 1)
		confl = s.reason(MinisatLitVar(p))
		*s.seen.At(MinisatLitVar(p)) = 0
		pathC--

		if !(pathC > 0) {
			break
		}
	}
	*outLearnt.At(0) = MinisatLitNot(p)

	// Simplify conflict clause:
	var i, j int
	outLearnt.CopyTo(&s.analyze_toclear)
	if s.Ccmin_mode == 2 {
		var abstractLevel uint32
		for i = 1; i < outLearnt.Size(); i++ {
			abstractLevel |= s.abstractLevel(MinisatLitVar(*outLearnt.At(i))) // (maintain an abstraction of levels involved in conflict)
		}
		for i, j = 1, 1; i < outLearnt.Size(); i++ {
			if s.reason(MinisatLitVar(*outLearnt.At(i))) == MinisatCRefUndef || !s.litRedundant(*outLearnt.At(i), abstractLevel) {
				*outLearnt.At(j) = *outLearnt.At(i)
				j++
			}
		}
	} else if s.Ccmin_mode == 1 {
		for i, j = 1, 1; i < outLearnt.Size(); i++ {
			x := MinisatLitVar(*outLearnt.At(i))

			if s.reason(x) == MinisatCRefUndef {
				*outLearnt.At(j) = *outLearnt.At(i)
				j++
			} else {
				c := s.ca.At(s.reason(MinisatLitVar(*outLearnt.At(i))))
				for k := 1; k < c.Size(); k++ {
					if *s.seen.At(MinisatLitVar(*c.At(k))) == 0 && s.level(MinisatLitVar(*c.At(k))) > 0 {
						*outLearnt.At(j) = *outLearnt.At(i)
						j++
						break
					}
				}
			}
		}
	} else {
		i = outLearnt.Size()
		j = i
	}

	s.Max_literals += uint64(outLearnt.Size())
	outLearnt.Shrink(i - j)
	s.Tot_literals += uint64(outLearnt.Size())

	// Find correct backtrack level:
	if outLearnt.Size() == 1 {
		*outBtlevel = 0
	} else {
		maxI := 1
		// Find the first literal assigned at the next-highest level:
		for i := 2; i < outLearnt.Size(); i++ {
			if s.level(MinisatLitVar(*outLearnt.At(i))) > s.level(MinisatLitVar(*outLearnt.At(maxI))) {
				maxI = i
			}
		}
		// Swap-in this literal at index 1:
		p := *outLearnt.At(maxI)
		*outLearnt.At(maxI) = *outLearnt.At(1)
		*outLearnt.At(1) = p
		*outBtlevel = s.level(MinisatLitVar(p))
	}

	for j = 0; j < s.analyze_toclear.Size(); j++ {
		*s.seen.At(MinisatLitVar(*s.analyze_toclear.At(j))) = 0 // ('seen[]' is now cleared)
	}
}

// litRedundant is Solver::litRedundant (protected): check if 'p' can be removed.
// 'abstract_levels' is used to abort early if the algorithm is visiting literals
// at levels that cannot be removed later.
// C: bool litRedundant(Lit p, uint32_t abstract_levels)
func (s *MinisatSolver) litRedundant(p MinisatLit, abstractLevels uint32) bool {
	s.analyze_stack.Clear(false)
	s.analyze_stack.PushElem(p)
	top := s.analyze_toclear.Size()
	for s.analyze_stack.Size() > 0 {
		if s.reason(MinisatLitVar(*s.analyze_stack.Last())) == MinisatCRefUndef {
			panic("MinisatSolver.litRedundant: reason == CRef_Undef")
		}
		c := s.ca.At(s.reason(MinisatLitVar(*s.analyze_stack.Last())))
		s.analyze_stack.Pop()

		for i := 1; i < c.Size(); i++ {
			p := *c.At(i)
			if *s.seen.At(MinisatLitVar(p)) == 0 && s.level(MinisatLitVar(p)) > 0 {
				if s.reason(MinisatLitVar(p)) != MinisatCRefUndef && (s.abstractLevel(MinisatLitVar(p))&abstractLevels) != 0 {
					*s.seen.At(MinisatLitVar(p)) = 1
					s.analyze_stack.PushElem(p)
					s.analyze_toclear.PushElem(p)
				} else {
					for j := top; j < s.analyze_toclear.Size(); j++ {
						*s.seen.At(MinisatLitVar(*s.analyze_toclear.At(j))) = 0
					}
					s.analyze_toclear.Shrink(s.analyze_toclear.Size() - top)
					return false
				}
			}
		}
	}
	return true
}

// analyzeFinal is Solver::analyzeFinal (protected): specialized analysis procedure
// to express the final conflict in terms of assumptions. Calculates the (possibly
// empty) set of assumptions that led to the assignment of 'p', and stores the
// result in 'out_conflict'.
// C: void analyzeFinal(Lit p, vec<Lit>& out_conflict)
func (s *MinisatSolver) analyzeFinal(p MinisatLit, outConflict *MinisatVec[MinisatLit]) {
	outConflict.Clear(false)
	outConflict.PushElem(p)

	if s.decisionLevel() == 0 {
		return
	}

	*s.seen.At(MinisatLitVar(p)) = 1

	for i := s.trail.Size() - 1; i >= *s.trail_lim.At(0); i-- {
		x := MinisatLitVar(*s.trail.At(i))
		if *s.seen.At(x) != 0 {
			if s.reason(x) == MinisatCRefUndef {
				if s.level(x) <= 0 {
					panic("MinisatSolver.analyzeFinal: level(x) <= 0")
				}
				outConflict.PushElem(MinisatLitNot(*s.trail.At(i)))
			} else {
				c := s.ca.At(s.reason(x))
				for j := 1; j < c.Size(); j++ {
					if s.level(MinisatLitVar(*c.At(j))) > 0 {
						*s.seen.At(MinisatLitVar(*c.At(j))) = 1
					}
				}
			}
			*s.seen.At(x) = 0
		}
	}

	*s.seen.At(MinisatLitVar(p)) = 0
}

// uncheckedEnqueue is Solver::uncheckedEnqueue (protected; from defaults to CRef_Undef).
// C: void uncheckedEnqueue(Lit p, CRef from = CRef_Undef)
func (s *MinisatSolver) uncheckedEnqueue(p MinisatLit, from MinisatCRef) {
	if s.ValueLit(p) != MinisatLUndef {
		panic("MinisatSolver.uncheckedEnqueue: literal already assigned")
	}
	*s.assigns.At(MinisatLitVar(p)) = NewMinisatLboolB(!MinisatSign(p))
	*s.vardata.At(MinisatLitVar(p)) = minisatSolverMkVarData(from, s.decisionLevel())
	s.trail.Push_(p)
}

// enqueue is Solver::enqueue (protected; from defaults to CRef_Undef): test if
// fact 'p' contradicts current state, enqueue otherwise.
// C: bool enqueue(Lit p, CRef from = CRef_Undef)
func (s *MinisatSolver) enqueue(p MinisatLit, from MinisatCRef) bool {
	if s.ValueLit(p) != MinisatLUndef {
		return s.ValueLit(p) != MinisatLFalse
	}
	s.uncheckedEnqueue(p, from)
	return true
}

// propagate is Solver::propagate (protected): propagates all enqueued facts. If a
// conflict arises, the conflicting clause is returned, otherwise CRef_Undef.
// Post-condition: the propagation queue is empty, even if there was a conflict.
// C: CRef propagate()
func (s *MinisatSolver) propagate() MinisatCRef {
	confl := MinisatCRefUndef
	numProps := 0
	s.watches.CleanAll()

	for s.qhead < s.trail.Size() {
		p := *s.trail.At(s.qhead) // 'p' is enqueued fact to propagate.
		s.qhead++
		ws := s.watches.At(p)
		wd := ws.Data()
		var i, j int
		end := ws.Size()
		numProps++

	NextWatcher:
		for i, j = 0, 0; i != end; {
			// Try to avoid inspecting the clause:
			blocker := wd[i].blocker
			if s.ValueLit(blocker) == MinisatLTrue {
				wd[j] = wd[i]
				j++
				i++
				continue
			}

			// Make sure the false literal is data[1]:
			cr := wd[i].cref
			c := s.ca.At(cr)
			falseLit := MinisatLitNot(p)
			if *c.At(0) == falseLit {
				*c.At(0) = *c.At(1)
				*c.At(1) = falseLit
			}
			if *c.At(1) != falseLit {
				panic("MinisatSolver.propagate: c[1] != false_lit")
			}
			i++

			// If 0th watch is true, then clause is already satisfied.
			first := *c.At(0)
			w := NewMinisatSolverWatcher(cr, first)
			if first != blocker && s.ValueLit(first) == MinisatLTrue {
				wd[j] = w
				j++
				continue
			}

			// Look for new watch:
			for k := 2; k < c.Size(); k++ {
				if s.ValueLit(*c.At(k)) != MinisatLFalse {
					*c.At(1) = *c.At(k)
					*c.At(k) = falseLit
					s.watches.At(MinisatLitNot(*c.At(1))).PushElem(w)
					continue NextWatcher
				}
			}

			// Did not find watch -- clause is unit under assignment:
			wd[j] = w
			j++
			if s.ValueLit(first) == MinisatLFalse {
				confl = cr
				s.qhead = s.trail.Size()
				// Copy the remaining watches:
				for i < end {
					wd[j] = wd[i]
					j++
					i++
				}
			} else {
				s.uncheckedEnqueue(first, cr)
			}
		}
		ws.Shrink(i - j)
	}
	s.Propagations += uint64(numProps)
	s.simpDB_props -= int64(numProps)

	return confl
}

// minisatReduceDBLt is the local struct reduceDB_lt of Solver.cpp.
type minisatReduceDBLt struct {
	ca *MinisatClauseAllocator
}

// Call is reduceDB_lt::operator().
// C: bool operator()(CRef x, CRef y)
func (l minisatReduceDBLt) Call(x MinisatCRef, y MinisatCRef) bool {
	return l.ca.At(x).Size() > 2 && (l.ca.At(y).Size() == 2 || *l.ca.At(x).Activity() < *l.ca.At(y).Activity())
}

// reduceDB is Solver::reduceDB (protected): remove half of the learnt clauses,
// minus the clauses locked by the current assignment. Binary clauses are never removed.
// C: void reduceDB()
func (s *MinisatSolver) reduceDB() {
	var i, j int
	extraLim := s.cla_inc / float64(s.learnts.Size()) // Remove any clause below this activity

	MinisatSortVec(&s.learnts, minisatReduceDBLt{ca: &s.ca})
	// Don't delete binary or locked clauses. From the rest, delete clauses from the first half
	// and clauses with activity smaller than 'extra_lim':
	for i, j = 0, 0; i < s.learnts.Size(); i++ {
		c := s.ca.At(*s.learnts.At(i))
		if c.Size() > 2 && !s.locked(c) && (i < s.learnts.Size()/2 || float64(*c.Activity()) < extraLim) {
			s.removeClause(*s.learnts.At(i))
		} else {
			*s.learnts.At(j) = *s.learnts.At(i)
			j++
		}
	}
	s.learnts.Shrink(i - j)
	s.CheckGarbage()
}

// removeSatisfied is Solver::removeSatisfied (protected).
// C: void removeSatisfied(vec<CRef>& cs)
func (s *MinisatSolver) removeSatisfied(cs *MinisatVec[MinisatCRef]) {
	var i, j int
	for i, j = 0, 0; i < cs.Size(); i++ {
		c := s.ca.At(*cs.At(i))
		if s.satisfied(c) {
			s.removeClause(*cs.At(i))
		} else {
			*cs.At(j) = *cs.At(i)
			j++
		}
	}
	cs.Shrink(i - j)
}

// rebuildOrderHeap is Solver::rebuildOrderHeap (protected).
// C: void rebuildOrderHeap()
func (s *MinisatSolver) rebuildOrderHeap() {
	var vs MinisatVec[MinisatVar]
	for v := 0; v < s.NVars(); v++ {
		if *s.decision.At(v) != 0 && s.Value(v) == MinisatLUndef {
			vs.PushElem(v)
		}
	}
	s.order_heap.Build(&vs)
}

// Simplify is Solver::simplify: simplify the clause database according to the
// current top-level assigment (currently only removes satisfied clauses).
// C: bool simplify()
func (s *MinisatSolver) Simplify() bool {
	if s.decisionLevel() != 0 {
		panic("MinisatSolver.Simplify: decisionLevel() != 0")
	}
	if !s.ok || s.propagate() != MinisatCRefUndef {
		s.ok = false
		return false
	}
	if s.NAssigns() == s.simpDB_assigns || (s.simpDB_props > 0) {
		return true
	}
	// Remove satisfied clauses:
	s.removeSatisfied(&s.learnts)
	if s.remove_satisfied { // Can be turned off.
		s.removeSatisfied(&s.clauses)
	}
	s.CheckGarbage()
	s.rebuildOrderHeap()

	s.simpDB_assigns = s.NAssigns()
	s.simpDB_props = int64(s.Clauses_literals + s.Learnts_literals) // (shouldn't depend on stats really, but it will do for now)
	return true
}

// search is Solver::search (protected): search for a model the specified number
// of conflicts. Use a negative value for 'nof_conflicts' to indicate infinity.
// Returns l_True if a partial assignment consistent with the clause set is
// found, l_False if the clause set is unsatisfiable, l_Undef if the bound on the
// number of conflicts is reached.
// C: lbool search(int nof_conflicts)
func (s *MinisatSolver) search(nofConflicts int) MinisatLbool {
	if !s.ok {
		panic("MinisatSolver.search: !ok")
	}
	var backtrackLevel int
	conflictC := 0
	var learntClause MinisatVec[MinisatLit]
	s.Starts++

	for {
		confl := s.propagate()
		if confl != MinisatCRefUndef {
			// CONFLICT
			s.Conflicts++
			conflictC++
			if s.decisionLevel() == 0 {
				return MinisatLFalse
			}

			learntClause.Clear(false)
			s.analyze(confl, &learntClause, &backtrackLevel)
			s.cancelUntil(backtrackLevel)

			if learntClause.Size() == 1 {
				s.uncheckedEnqueue(*learntClause.At(0), MinisatCRefUndef)
			} else {
				cr := s.ca.Alloc(&learntClause, true)
				s.learnts.PushElem(cr)
				s.attachClause(cr)
				s.claBumpActivity(s.ca.At(cr))
				s.uncheckedEnqueue(*learntClause.At(0), cr)
			}

			s.varDecayActivity()
			s.claDecayActivity()

			s.learntsize_adjust_cnt--
			if s.learntsize_adjust_cnt == 0 {
				s.learntsize_adjust_confl *= s.Learntsize_adjust_inc
				s.learntsize_adjust_cnt = int(s.learntsize_adjust_confl)
				s.max_learnts *= s.Learntsize_inc

				if s.Verbosity >= 1 {
					lim := s.trail.Size()
					if s.trail_lim.Size() != 0 {
						lim = *s.trail_lim.At(0)
					}
					fmt.Printf("| %9d | %7d %8d %8d | %8d %8d %6.0f | %6.3f %% |\n",
						int(s.Conflicts),
						int(s.Dec_vars)-lim, s.NClauses(), int(s.Clauses_literals),
						int(s.max_learnts), s.NLearnts(), float64(int64(s.Learnts_literals))/float64(s.NLearnts()), s.progressEstimate()*100)
				}
			}
		} else {
			// NO CONFLICT
			if nofConflicts >= 0 && conflictC >= nofConflicts || !s.withinBudget() {
				// Reached bound on number of conflicts:
				s.progress_estimate = s.progressEstimate()
				s.cancelUntil(0)
				return MinisatLUndef
			}

			// Simplify the set of problem clauses:
			if s.decisionLevel() == 0 && !s.Simplify() {
				return MinisatLFalse
			}

			if float64(s.learnts.Size()-s.NAssigns()) >= s.max_learnts {
				// Reduce the set of learnt clauses:
				s.reduceDB()
			}

			next := MinisatLitUndef
			for s.decisionLevel() < s.assumptions.Size() {
				// Perform user provided assumption:
				p := *s.assumptions.At(s.decisionLevel())
				if s.ValueLit(p) == MinisatLTrue {
					// Dummy decision level:
					s.newDecisionLevel()
				} else if s.ValueLit(p) == MinisatLFalse {
					s.analyzeFinal(MinisatLitNot(p), &s.Conflict)
					return MinisatLFalse
				} else {
					next = p
					break
				}
			}

			if next == MinisatLitUndef {
				// New variable decision:
				s.Decisions++
				next = s.pickBranchLit()

				if next == MinisatLitUndef {
					// Model found:
					return MinisatLTrue
				}
			}

			// Increase decision level and enqueue 'next'
			s.newDecisionLevel()
			s.uncheckedEnqueue(next, MinisatCRefUndef)
		}
	}
}

// progressEstimate is Solver::progressEstimate (protected).
// C: double progressEstimate() const
func (s *MinisatSolver) progressEstimate() float64 {
	progress := float64(0)
	F := 1.0 / float64(s.NVars())

	for i := 0; i <= s.decisionLevel(); i++ {
		beg := 0
		if i != 0 {
			beg = *s.trail_lim.At(i - 1)
		}
		end := s.trail.Size()
		if i != s.decisionLevel() {
			end = *s.trail_lim.At(i)
		}
		progress += math.Pow(F, float64(i)) * float64(end-beg)
	}
	return progress / float64(s.NVars())
}

// luby_minisat is `static double luby(double y, int x)`: finite subsequences of the Luby sequence.
// C: static double luby(double y, int x)
func luby_minisat(y float64, x int) float64 {
	// Find the finite subsequence that contains index 'x', and the size of that subsequence:
	var size, seq int
	for size, seq = 1, 0; size < x+1; seq, size = seq+1, 2*size+1 {
	}
	for size-1 != x {
		size = (size - 1) >> 1
		seq--
		x = x % size
	}
	return math.Pow(y, float64(seq))
}

// solve_ is Solver::solve_ (protected): main solve method (assumptions given in 'assumptions').
// C: lbool solve_()
func (s *MinisatSolver) solve_() MinisatLbool {
	s.Model.Clear(false)
	s.Conflict.Clear(false)
	if !s.ok {
		return MinisatLFalse
	}

	s.Solves++

	s.max_learnts = float64(s.NClauses()) * s.Learntsize_factor
	s.learntsize_adjust_confl = float64(s.Learntsize_adjust_start_confl)
	s.learntsize_adjust_cnt = int(s.learntsize_adjust_confl)
	status := MinisatLUndef

	if s.Verbosity >= 1 {
		fmt.Printf("============================[ Search Statistics ]==============================\n")
		fmt.Printf("| Conflicts |          ORIGINAL         |          LEARNT          | Progress |\n")
		fmt.Printf("|           |    Vars  Clauses Literals |    Limit  Clauses Lit/Cl |          |\n")
		fmt.Printf("===============================================================================\n")
	}

	// Search:
	currRestarts := 0
	for status == MinisatLUndef {
		var restBase float64
		if s.Luby_restart {
			restBase = luby_minisat(s.Restart_inc, currRestarts)
		} else {
			restBase = math.Pow(s.Restart_inc, float64(currRestarts))
		}
		status = s.search(int(restBase * float64(s.Restart_first)))
		if !s.withinBudget() {
			break
		}
		currRestarts++
	}

	if s.Verbosity >= 1 {
		fmt.Printf("===============================================================================\n")
	}

	if status == MinisatLTrue {
		// Extend & copy model:
		s.Model.GrowTo(s.NVars())
		for i := 0; i < s.NVars(); i++ {
			*s.Model.At(i) = s.Value(i)
		}
	} else if status == MinisatLFalse && s.Conflict.Size() == 0 {
		s.ok = false
	}

	s.cancelUntil(0)
	return status
}

// Solve is the overload solve(const vec<Lit>& assumps).
// C: bool solve(const vec<Lit>& assumps)
func (s *MinisatSolver) Solve(assumps *MinisatVec[MinisatLit]) bool {
	s.BudgetOff()
	assumps.CopyTo(&s.assumptions)
	return s.solve_() == MinisatLTrue
}

// SolveLimited is Solver::solveLimited (with resource constraints).
// C: lbool solveLimited(const vec<Lit>& assumps)
func (s *MinisatSolver) SolveLimited(assumps *MinisatVec[MinisatLit]) MinisatLbool {
	assumps.CopyTo(&s.assumptions)
	return s.solve_()
}

// SolveNoAssumps is the overload solve().
// C: bool solve()
func (s *MinisatSolver) SolveNoAssumps() bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	return s.solve_() == MinisatLTrue
}

// Solve1 is the overload solve(Lit p).
// C: bool solve(Lit p)
func (s *MinisatSolver) Solve1(p MinisatLit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	return s.solve_() == MinisatLTrue
}

// Solve2 is the overload solve(Lit p, Lit q).
// C: bool solve(Lit p, Lit q)
func (s *MinisatSolver) Solve2(p MinisatLit, q MinisatLit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	return s.solve_() == MinisatLTrue
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r).
// C: bool solve(Lit p, Lit q, Lit r)
func (s *MinisatSolver) Solve3(p MinisatLit, q MinisatLit, r MinisatLit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	s.assumptions.PushElem(r)
	return s.solve_() == MinisatLTrue
}

// Okay is Solver::okay.
// C: bool okay() const
func (s *MinisatSolver) Okay() bool { return s.ok }

// mapVar_minisat is `static Var mapVar(Var x, vec<Var>& map, Var& max)`.
// C: static Var mapVar(Var x, vec<Var>& map, Var& max)
func mapVar_minisat(x MinisatVar, map_ *MinisatVec[MinisatVar], max *MinisatVar) MinisatVar {
	if map_.Size() <= x || *map_.At(x) == -1 {
		map_.GrowToPad(x+1, -1)
		*map_.At(x) = *max
		*max++
	}
	return *map_.At(x)
}

// ToDimacsClause is the overload toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max).
// C: void toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max)
func (s *MinisatSolver) ToDimacsClause(f *os.File, c *MinisatClause, map_ *MinisatVec[MinisatVar], max *MinisatVar) {
	if s.satisfied(c) {
		return
	}
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) != MinisatLFalse {
			sg := ""
			if MinisatSign(*c.At(i)) {
				sg = "-"
			}
			fmt.Fprintf(f, "%s%d ", sg, mapVar_minisat(MinisatLitVar(*c.At(i)), map_, max)+1)
		}
	}
	fmt.Fprintf(f, "0\n")
}

// ToDimacsFile is the overload toDimacs(const char *file, const vec<Lit>& assumps).
// C: void toDimacs(const char *file, const vec<Lit>& assumps)
func (s *MinisatSolver) ToDimacsFile(file string, assumps *MinisatVec[MinisatLit]) {
	f, err := os.Create(file) // fopen(file, "wr")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open file %s\n", file)
		os.Exit(1)
	}
	s.ToDimacsStream(f, assumps)
	f.Close()
}

// ToDimacsStream is the overload toDimacs(FILE* f, const vec<Lit>& assumps).
// C: void toDimacs(FILE* f, const vec<Lit>& assumps)
func (s *MinisatSolver) ToDimacsStream(f *os.File, assumps *MinisatVec[MinisatLit]) {
	// Handle case when solver is in contradictory state:
	if !s.ok {
		fmt.Fprintf(f, "p cnf 1 2\n1 0\n-1 0\n")
		return
	}

	var map_ MinisatVec[MinisatVar]
	max := 0

	// Cannot use removeClauses here because it is not safe
	// to deallocate them at this point. Could be improved.
	var i int
	cnt := 0
	for i = 0; i < s.clauses.Size(); i++ {
		if !s.satisfied(s.ca.At(*s.clauses.At(i))) {
			cnt++
		}
	}

	for i = 0; i < s.clauses.Size(); i++ {
		if !s.satisfied(s.ca.At(*s.clauses.At(i))) {
			c := s.ca.At(*s.clauses.At(i))
			for j := 0; j < c.Size(); j++ {
				if s.ValueLit(*c.At(j)) != MinisatLFalse {
					mapVar_minisat(MinisatLitVar(*c.At(j)), &map_, &max)
				}
			}
		}
	}

	// Assumptions are added as unit clauses:
	cnt += s.assumptions.Size()

	fmt.Fprintf(f, "p cnf %d %d\n", max, cnt)

	for i = 0; i < s.assumptions.Size(); i++ {
		if s.ValueLit(*s.assumptions.At(i)) == MinisatLFalse {
			panic("MinisatSolver.ToDimacsStream: assumption is false")
		}
		sg := ""
		if MinisatSign(*s.assumptions.At(i)) {
			sg = "-"
		}
		fmt.Fprintf(f, "%s%d 0\n", sg, mapVar_minisat(MinisatLitVar(*s.assumptions.At(i)), &map_, &max)+1)
	}

	for i = 0; i < s.clauses.Size(); i++ {
		s.ToDimacsClause(f, s.ca.At(*s.clauses.At(i)), &map_, &max)
	}

	if s.Verbosity > 0 {
		fmt.Printf("Wrote %d clauses with %d variables.\n", cnt, max)
	}
}

// ToDimacsFileNoAssumps is the overload toDimacs(const char* file).
// C: void toDimacs(const char* file)
func (s *MinisatSolver) ToDimacsFileNoAssumps(file string) {
	var as MinisatVec[MinisatLit]
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *MinisatSolver) ToDimacsFile1(file string, p MinisatLit) {
	var as MinisatVec[MinisatLit]
	as.PushElem(p)
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *MinisatSolver) ToDimacsFile2(file string, p MinisatLit, q MinisatLit) {
	var as MinisatVec[MinisatLit]
	as.PushElem(p)
	as.PushElem(q)
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *MinisatSolver) ToDimacsFile3(file string, p MinisatLit, q MinisatLit, r MinisatLit) {
	var as MinisatVec[MinisatLit]
	as.PushElem(p)
	as.PushElem(q)
	as.PushElem(r)
	s.ToDimacsFile(file, &as)
}

// relocAll is Solver::relocAll (protected).
// C: void relocAll(ClauseAllocator& to)
func (s *MinisatSolver) relocAll(to *MinisatClauseAllocator) {
	// All watchers:
	s.watches.CleanAll()
	for v := 0; v < s.NVars(); v++ {
		for sg := 0; sg < 2; sg++ {
			p := MinisatMkLit(v, sg != 0)
			ws := s.watches.At(p)
			for j := 0; j < ws.Size(); j++ {
				s.ca.Reloc(&ws.At(j).cref, to)
			}
		}
	}

	// All reasons:
	var i int
	for i = 0; i < s.trail.Size(); i++ {
		v := MinisatLitVar(*s.trail.At(i))

		if s.reason(v) != MinisatCRefUndef && (s.ca.At(s.reason(v)).Reloced() || s.locked(s.ca.At(s.reason(v)))) {
			s.ca.Reloc(&s.vardata.At(v).reason, to)
		}
	}

	// All learnt:
	for i = 0; i < s.learnts.Size(); i++ {
		s.ca.Reloc(s.learnts.At(i), to)
	}

	// All original:
	for i = 0; i < s.clauses.Size(); i++ {
		s.ca.Reloc(s.clauses.At(i), to)
	}
}

// GarbageCollect is Solver::garbageCollect (virtual).
// C: virtual void garbageCollect()
func (s *MinisatSolver) GarbageCollect() {
	// Initialize the next region to a size corresponding to the estimated utilization degree. This
	// is not precise but should avoid some unnecessary reallocations for the new region:
	to := NewMinisatClauseAllocatorCap(s.ca.Size() - s.ca.Wasted())

	s.relocAll(to)
	if s.Verbosity >= 2 {
		fmt.Printf("|  Garbage collection:   %12d bytes => %12d bytes             |\n",
			s.ca.Size()*MinisatRegionAllocatorUnitSize, to.Size()*MinisatRegionAllocatorUnitSize)
	}
	to.MoveTo(&s.ca)
}

// garbageCollectVirt is the virtual call garbageCollect().
func (s *MinisatSolver) garbageCollectVirt() {
	if s.virtGC != nil {
		s.virtGC()
	} else {
		s.GarbageCollect()
	}
}

// CheckGarbageFrac is the overload checkGarbage(double gf).
// C: void checkGarbage(double gf)
func (s *MinisatSolver) CheckGarbageFrac(gf float64) {
	if float64(s.ca.Wasted()) > float64(s.ca.Size())*gf {
		s.garbageCollectVirt()
	}
}

// CheckGarbage is the overload checkGarbage().
// C: void checkGarbage()
func (s *MinisatSolver) CheckGarbage() { s.CheckGarbageFrac(s.Garbage_frac) }

// minisatSolverDrand is Solver::drand (protected static inline).
// C: static inline double drand(double& seed)
func minisatSolverDrand(seed *float64) float64 {
	*seed *= 1389796
	q := int(*seed / 2147483647)
	*seed -= float64(q) * 2147483647
	return *seed / 2147483647
}

// minisatSolverIrand is Solver::irand (protected static inline).
// C: static inline int irand(double& seed, int size)
func minisatSolverIrand(seed *float64, size int) int {
	return int(minisatSolverDrand(seed) * float64(size))
}

// insertVarOrder is Solver::insertVarOrder (protected).
// C: void insertVarOrder(Var x)
func (s *MinisatSolver) insertVarOrder(x MinisatVar) {
	if !s.order_heap.InHeap(x) && *s.decision.At(x) != 0 {
		s.order_heap.Insert(x)
	}
}

// newDecisionLevel is Solver::newDecisionLevel (protected).
// C: void newDecisionLevel()
func (s *MinisatSolver) newDecisionLevel() { s.trail_lim.PushElem(s.trail.Size()) }

// varDecayActivity is Solver::varDecayActivity (protected).
// C: void varDecayActivity()
func (s *MinisatSolver) varDecayActivity() { s.var_inc *= 1 / s.Var_decay }

// varBumpActivityInc is the overload varBumpActivity(Var v, double inc) (protected).
// C: void varBumpActivity(Var v, double inc)
func (s *MinisatSolver) varBumpActivityInc(v MinisatVar, inc float64) {
	*s.activity.At(v) += inc
	if *s.activity.At(v) > 1e100 {
		// Rescale:
		for i := 0; i < s.NVars(); i++ {
			*s.activity.At(i) *= 1e-100
		}
		s.var_inc *= 1e-100
	}
	// Update order_heap with respect to new activity:
	if s.order_heap.InHeap(v) {
		s.order_heap.Decrease(v)
	}
}

// varBumpActivity is the overload varBumpActivity(Var v) (protected).
// C: void varBumpActivity(Var v)
func (s *MinisatSolver) varBumpActivity(v MinisatVar) { s.varBumpActivityInc(v, s.var_inc) }

// claDecayActivity is Solver::claDecayActivity (protected).
// C: void claDecayActivity()
func (s *MinisatSolver) claDecayActivity() { s.cla_inc *= 1 / s.Clause_decay }

// claBumpActivity is Solver::claBumpActivity (protected).
// C: void claBumpActivity(Clause& c)
func (s *MinisatSolver) claBumpActivity(c *MinisatClause) {
	a := c.Activity()
	*a = float32(float64(*a) + s.cla_inc)
	if float64(*a) > 1e20 {
		// Rescale:
		for i := 0; i < s.learnts.Size(); i++ {
			la := s.ca.At(*s.learnts.At(i)).Activity()
			*la *= float32(1e-20)
		}
		s.cla_inc *= 1e-20
	}
}

// locked is Solver::locked (protected): true if a clause is the reason for some
// implication in the current state.
// C: bool locked(const Clause& c) const
func (s *MinisatSolver) locked(c *MinisatClause) bool {
	return s.ValueLit(*c.At(0)) == MinisatLTrue && s.reason(MinisatLitVar(*c.At(0))) != MinisatCRefUndef &&
		s.reason(MinisatLitVar(*c.At(0))) == s.ca.Ael(c)
}

// decisionLevel is Solver::decisionLevel (protected).
// C: int decisionLevel() const
func (s *MinisatSolver) decisionLevel() int { return s.trail_lim.Size() }

// abstractLevel is Solver::abstractLevel (protected): used to represent an
// abstraction of sets of decision levels.
// C: uint32_t abstractLevel(Var x) const
func (s *MinisatSolver) abstractLevel(x MinisatVar) uint32 { return 1 << (uint(s.level(x)) & 31) }

// reason is Solver::reason (protected).
// C: CRef reason(Var x) const
func (s *MinisatSolver) reason(x MinisatVar) MinisatCRef { return s.vardata.At(x).reason }

// level is Solver::level (protected).
// C: int level(Var x) const
func (s *MinisatSolver) level(x MinisatVar) int { return s.vardata.At(x).level }

// Value is the overload value(Var x).
// C: lbool value(Var x) const
func (s *MinisatSolver) Value(x MinisatVar) MinisatLbool { return *s.assigns.At(x) }

// ValueLit is the overload value(Lit p).
// C: lbool value(Lit p) const
func (s *MinisatSolver) ValueLit(p MinisatLit) MinisatLbool {
	return s.assigns.At(MinisatLitVar(p)).Xor(MinisatSign(p))
}

// ModelValue is the overload modelValue(Var x).
// C: lbool modelValue(Var x) const
func (s *MinisatSolver) ModelValue(x MinisatVar) MinisatLbool { return *s.Model.At(x) }

// ModelValueLit is the overload modelValue(Lit p).
// C: lbool modelValue(Lit p) const
func (s *MinisatSolver) ModelValueLit(p MinisatLit) MinisatLbool {
	return s.Model.At(MinisatLitVar(p)).Xor(MinisatSign(p))
}

// NAssigns is Solver::nAssigns.
// C: int nAssigns() const
func (s *MinisatSolver) NAssigns() int { return s.trail.Size() }

// NClauses is Solver::nClauses.
// C: int nClauses() const
func (s *MinisatSolver) NClauses() int { return s.clauses.Size() }

// NLearnts is Solver::nLearnts.
// C: int nLearnts() const
func (s *MinisatSolver) NLearnts() int { return s.learnts.Size() }

// NVars is Solver::nVars.
// C: int nVars() const
func (s *MinisatSolver) NVars() int { return s.vardata.Size() }

// NFreeVars is Solver::nFreeVars.
// C: int nFreeVars() const
func (s *MinisatSolver) NFreeVars() int {
	lim := s.trail.Size()
	if s.trail_lim.Size() != 0 {
		lim = *s.trail_lim.At(0)
	}
	return int(s.Dec_vars) - lim
}

// SetPolarity is Solver::setPolarity.
// C: void setPolarity(Var v, bool b)
func (s *MinisatSolver) SetPolarity(v MinisatVar, b bool) {
	if b {
		*s.polarity.At(v) = 1
	} else {
		*s.polarity.At(v) = 0
	}
}

// SetDecisionVar is Solver::setDecisionVar.
// C: void setDecisionVar(Var v, bool b)
func (s *MinisatSolver) SetDecisionVar(v MinisatVar, b bool) {
	if b && *s.decision.At(v) == 0 {
		s.Dec_vars++
	} else if !b && *s.decision.At(v) != 0 {
		s.Dec_vars--
	}
	if b {
		*s.decision.At(v) = 1
	} else {
		*s.decision.At(v) = 0
	}
	s.insertVarOrder(v)
}

// SetConfBudget is Solver::setConfBudget.
// C: void setConfBudget(int64_t x)
func (s *MinisatSolver) SetConfBudget(x int64) { s.conflict_budget = int64(s.Conflicts) + x }

// SetPropBudget is Solver::setPropBudget.
// C: void setPropBudget(int64_t x)
func (s *MinisatSolver) SetPropBudget(x int64) { s.propagation_budget = int64(s.Propagations) + x }

// BudgetOff is Solver::budgetOff.
// C: void budgetOff()
func (s *MinisatSolver) BudgetOff() {
	s.conflict_budget = -1
	s.propagation_budget = -1
}

// Interrupt is Solver::interrupt (may be called asynchronously).
// C: void interrupt()
func (s *MinisatSolver) Interrupt() { s.asynch_interrupt = true }

// ClearInterrupt is Solver::clearInterrupt.
// C: void clearInterrupt()
func (s *MinisatSolver) ClearInterrupt() { s.asynch_interrupt = false }

// withinBudget is Solver::withinBudget (protected).
// C: bool withinBudget() const
func (s *MinisatSolver) withinBudget() bool {
	return !s.asynch_interrupt &&
		(s.conflict_budget < 0 || s.Conflicts < uint64(s.conflict_budget)) &&
		(s.propagation_budget < 0 || s.Propagations < uint64(s.propagation_budget))
}
