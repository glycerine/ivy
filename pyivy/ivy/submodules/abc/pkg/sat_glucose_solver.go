package abc

import (
	"fmt"
	"math"
	"os"
)

// Port of src/sat/glucose/Solver.h and Glucose.cpp (namespace Gluco).
// See sat_bsat2_vec.go for the C++ -> Go conventions. This Solver is the MiniSat
// solver extended with LBD based restarts and clause database reduction, a
// binary-clause watch list, an incremental mode, an external callback and
// certified-UNSAT output.

// GlucoSolverVarData is Solver::VarData.
type GlucoSolverVarData struct {
	reason GlucoCRef
	level  int
}

// glucoSolverMkVarData is Solver::mkVarData (protected static inline).
// C: static inline VarData mkVarData(CRef cr, int l)
func glucoSolverMkVarData(cr GlucoCRef, l int) GlucoSolverVarData {
	return GlucoSolverVarData{reason: cr, level: l}
}

// GlucoSolverWatcher is Solver::Watcher.
type GlucoSolverWatcher struct {
	cref    GlucoCRef
	blocker GlucoLit
}

// NewGlucoSolverWatcher is Watcher(CRef cr, Lit p).
// C: Watcher(CRef cr, Lit p)
func NewGlucoSolverWatcher(cr GlucoCRef, p GlucoLit) GlucoSolverWatcher {
	return GlucoSolverWatcher{cref: cr, blocker: p}
}

// Eq is Watcher::operator==.
// C: bool operator==(const Watcher& w) const
func (w GlucoSolverWatcher) Eq(o GlucoSolverWatcher) bool { return w.cref == o.cref }

// Ne is Watcher::operator!=.
// C: bool operator!=(const Watcher& w) const
func (w GlucoSolverWatcher) Ne(o GlucoSolverWatcher) bool { return w.cref != o.cref }

// GlucoSolverWatcherDeleted is Solver::WatcherDeleted.
type GlucoSolverWatcherDeleted struct {
	ca *GlucoClauseAllocator // const ClauseAllocator&
}

// NewGlucoSolverWatcherDeleted is WatcherDeleted(const ClauseAllocator& _ca).
// C: WatcherDeleted(const ClauseAllocator& _ca)
func NewGlucoSolverWatcherDeleted(ca *GlucoClauseAllocator) GlucoSolverWatcherDeleted {
	return GlucoSolverWatcherDeleted{ca: ca}
}

// Call is WatcherDeleted::operator().
// C: bool operator()(const Watcher& w) const
func (d GlucoSolverWatcherDeleted) Call(w GlucoSolverWatcher) bool {
	return d.ca.At(w.cref).Mark() == 1
}

// GlucoSolverVarOrderLt is Solver::VarOrderLt.
type GlucoSolverVarOrderLt struct {
	activity *GlucoVec[float64] // const vec<double>&
}

// NewGlucoSolverVarOrderLt is VarOrderLt(const vec<double>& act).
// C: VarOrderLt(const vec<double>& act)
func NewGlucoSolverVarOrderLt(act *GlucoVec[float64]) GlucoSolverVarOrderLt {
	return GlucoSolverVarOrderLt{activity: act}
}

// Call is VarOrderLt::operator().
// C: bool operator()(Var x, Var y) const
func (l GlucoSolverVarOrderLt) Call(x int, y int) bool {
	return *l.activity.At(x) > *l.activity.At(y)
}

// File-level category strings of Glucose.cpp.
const (
	glucoCat       = "CORE"
	glucoCr        = "CORE -- RESTART"
	glucoCred      = "CORE -- REDUCE"
	glucoCm        = "CORE -- MINIMIZE"
	glucoCertified = "CORE -- CERTIFIED UNSAT"
)

// Command-line options of the solver (static objects in C++: constructed when the
// program starts, registering themselves in Option::getOptionList()). opt_restart_first
// and opt_restart_inc are commented out in the C source.
var (
	opt_incremental_glucose              = NewGlucoBoolOption(glucoCat, "incremental", "Use incremental SAT solving", false)
	opt_K_glucose                        = NewGlucoDoubleOption(glucoCr, "K", "The constant used to force restart", 0.8, NewGlucoDoubleRange(0, false, 1, false))
	opt_R_glucose                        = NewGlucoDoubleOption(glucoCr, "R", "The constant used to block restart", 1.4, NewGlucoDoubleRange(1, false, 5, false))
	opt_size_lbd_queue_glucose           = NewGlucoIntOption(glucoCr, "szLBDQueue", "The size of moving average for LBD (restarts)", 50, NewGlucoIntRange(10, math.MaxInt32))
	opt_size_trail_queue_glucose         = NewGlucoIntOption(glucoCr, "szTrailQueue", "The size of moving average for trail (block restarts)", 5000, NewGlucoIntRange(10, math.MaxInt32))
	opt_first_reduce_db_glucose          = NewGlucoIntOption(glucoCred, "firstReduceDB", "The number of conflicts before the first reduce DB", 2000, NewGlucoIntRange(0, math.MaxInt32))
	opt_inc_reduce_db_glucose            = NewGlucoIntOption(glucoCred, "incReduceDB", "Increment for reduce DB", 300, NewGlucoIntRange(0, math.MaxInt32))
	opt_spec_inc_reduce_db_glucose       = NewGlucoIntOption(glucoCred, "specialIncReduceDB", "Special increment for reduce DB", 1000, NewGlucoIntRange(0, math.MaxInt32))
	opt_lb_lbd_frozen_clause_glucose     = NewGlucoIntOption(glucoCred, "minLBDFrozenClause", "Protect clauses if their LBD decrease and is lower than (for one turn)", 30, NewGlucoIntRange(0, math.MaxInt32))
	opt_lb_size_minimzing_clause_glucose = NewGlucoIntOption(glucoCm, "minSizeMinimizingClause", "The min size required to minimize clause", 30, NewGlucoIntRange(3, math.MaxInt32))
	opt_lb_lbd_minimzing_clause_glucose  = NewGlucoIntOption(glucoCm, "minLBDMinimizingClause", "The min LBD required to minimize clause", 6, NewGlucoIntRange(3, math.MaxInt32))
	opt_var_decay_glucose                = NewGlucoDoubleOption(glucoCat, "var-decay", "The variable activity decay factor", 0.8, NewGlucoDoubleRange(0, false, 1, false))
	opt_clause_decay_glucose             = NewGlucoDoubleOption(glucoCat, "cla-decay", "The clause activity decay factor", 0.999, NewGlucoDoubleRange(0, false, 1, false))
	opt_random_var_freq_glucose          = NewGlucoDoubleOption(glucoCat, "rnd-freq", "The frequency with which the decision heuristic tries to choose a random variable", 0, NewGlucoDoubleRange(0, true, 1, true))
	opt_random_seed_glucose              = NewGlucoDoubleOption(glucoCat, "rnd-seed", "Used by the random variable selection", 91648253, NewGlucoDoubleRange(0, false, math.Inf(1), false))
	opt_ccmin_mode_glucose               = NewGlucoIntOption(glucoCat, "ccmin-mode", "Controls conflict clause minimization (0=none, 1=basic, 2=deep)", 2, NewGlucoIntRange(0, 2))
	opt_phase_saving_glucose             = NewGlucoIntOption(glucoCat, "phase-saving", "Controls the level of phase saving (0=none, 1=limited, 2=full)", 2, NewGlucoIntRange(0, 2))
	opt_rnd_init_act_glucose             = NewGlucoBoolOption(glucoCat, "rnd-init", "Randomize the initial activity", false)
	opt_garbage_frac_glucose             = NewGlucoDoubleOption(glucoCat, "gc-frac", "The fraction of wasted memory allowed before a garbage collection is triggered", 0.20, NewGlucoDoubleRange(0, false, math.Inf(1), false))

	// opt_certified_ and opt_certified_file_ are not static in C (external linkage).
	opt_certified__glucose      = NewGlucoBoolOption(glucoCertified, "certified", "Certified UNSAT using DRUP format", false)
	opt_certified_file__glucose = NewGlucoStringOption(glucoCertified, "certified-output", "Certified UNSAT output file", "NULL")
)

// GlucoSolver is Gluco::Solver, the CDCL SAT solver. Public data members are
// exported; protected ones keep the C++ spelling. SimpSolver embeds it.
type GlucoSolver struct {
	SolverType int // ABC identifies the solver's type as 0

	PCnfMan                any                       // void * pCnfMan: external CNF manager
	PCnfFunc               func(any, int, []int) int // int(*pCnfFunc)(void *p, int, int*): external callback. messages: 0: unsat; 1: sat; -1: still working
	NCallConfl             int                       // callback will be called every this number of conflicts
	Terminate_search_early bool                      // used to stop the solver early if it as instructed by an external caller
	Pstop                  *int                      // another callback
	NRuntimeLimit          int64                     // runtime limit (abctime)
	User_vec               GlucoVec[int]
	User_lits              GlucoVec[GlucoLit]
	Jftr                   int

	// If problem is satisfiable, this vector contains the model (if any).
	Model GlucoVec[GlucoLbool]
	// If problem is unsatisfiable (possibly under assumptions), this vector represents the final conflict clause expressed in the assumptions.
	Conflict GlucoVec[GlucoLit]

	// Mode of operation:
	Verbosity              int
	VerbEveryConflicts     int
	ShowModel              int
	K                      float64
	R                      float64
	SizeLBDQueue           float64
	SizeTrailQueue         float64
	FirstReduceDB          int
	IncReduceDB            int
	SpecialIncReduceDB     int
	LbLBDFrozenClause      uint32
	LbSizeMinimizingClause int
	LbLBDMinimizingClause  uint32
	Var_decay              float64
	Clause_decay           float64
	Random_var_freq        float64
	Random_seed            float64
	Ccmin_mode             int      // conflict clause minimization (0=none, 1=basic, 2=deep)
	Phase_saving           int      // phase saving (0=none, 1=limited, 2=full)
	Rnd_pol                bool     // use random polarities for branching heuristics
	Rnd_init_act           bool     // initialize variable activities with a small random value
	Garbage_frac           float64  // fraction of wasted memory allowed before a garbage collection is triggered
	CertifiedOutput        *os.File // FILE*
	CertifiedUNSAT         bool

	// Statistics: (read-only member variables)
	NbRemovedClauses, NbReducedClauses, NbDL2, NbBin, NbUn, NbReduceDB, Solves, Starts, Decisions, Rnd_decisions, Propagations, Conflicts, ConflictsRestarts, Nbstopsrestarts, Nbstopsrestartssame, Lastblockatrestart int64
	Dec_vars, Clauses_literals, Learnts_literals, Max_literals, Tot_literals                                                                                                                                           int64

	// Solver state (protected):
	curRestart                            int64 // long
	lastIndexRed                          int
	ok                                    bool
	cla_inc                               float64
	activity                              GlucoVec[float64]
	var_inc                               float64
	watches                               GlucoOccLists[GlucoLit, GlucoSolverWatcher, GlucoSolverWatcherDeleted]
	watchesBin                            GlucoOccLists[GlucoLit, GlucoSolverWatcher, GlucoSolverWatcherDeleted]
	clauses                               GlucoVec[GlucoCRef]
	learnts                               GlucoVec[GlucoCRef]
	assigns                               GlucoVec[GlucoLbool]
	polarity                              GlucoVec[int8]
	decision                              GlucoVec[int8]
	trail                                 GlucoVec[GlucoLit]
	nbpos                                 GlucoVec[int]
	trail_lim                             GlucoVec[int]
	vardata                               GlucoVec[GlucoSolverVarData]
	qhead                                 int
	simpDB_assigns                        int
	simpDB_props                          int64
	assumptions                           GlucoVec[GlucoLit]
	order_heap                            GlucoHeap[GlucoSolverVarOrderLt]
	progress_estimate                     float64
	remove_satisfied                      bool
	permDiff                              GlucoVec[uint32]   // permDiff[var] contains the current conflict number... used to count the number of LBD
	lastDecisionLevel                     GlucoVec[GlucoLit] // #ifdef UPDATEVARACTIVITY (defined in Constants.h)
	ca                                    GlucoClauseAllocator
	nbclausesbeforereduce                 int                 // to know when it is time to reduce the clause database
	trailQueue                            GlucoBqueue[uint32] // bounded queues for restarts
	lbdQueue                              GlucoBqueue[uint32]
	sumLBD                                float32 // used to compute the global average of LBD
	sumAssumptions                        int
	seen                                  GlucoVec[int8]
	analyze_stack                         GlucoVec[GlucoLit]
	analyze_toclear                       GlucoVec[GlucoLit]
	add_tmp                               GlucoVec[GlucoLit]
	MYFLAG                                uint32
	max_learnts                           float64
	learntsize_adjust_confl               float64
	learntsize_adjust_cnt                 int
	conflict_budget                       int64 // -1 means no budget
	propagation_budget                    int64 // -1 means no budget
	asynch_interrupt                      bool
	incremental                           int // use incremental SAT solver
	nbVarsInitialFormula                  int // nb VAR in formula without assumptions (incremental SAT)
	totalTime4Sat, totalTime4Unsat        float64
	nbSatCalls, nbUnsatCalls              int
	assumptionPositions, initialPositions GlucoVec[int]

	// virtGC models the virtual call garbageCollect() (and reset): a derived
	// class that overrides it (SimpSolver) sets this to its own method.
	virtGC func()
}

// NewGlucoSolver is Solver::Solver().
// C: Solver::Solver()
func NewGlucoSolver() *GlucoSolver {
	s := &GlucoSolver{}
	s.initGlucoSolver()
	return s
}

// initGlucoSolver is the body of the Solver constructor; the SimpSolver constructor runs it for its base class.
func (s *GlucoSolver) initGlucoSolver() {
	s.SolverType = 0
	s.PCnfFunc = nil
	s.NCallConfl = 1000
	s.Terminate_search_early = false
	s.Pstop = nil
	s.NRuntimeLimit = 0

	s.Verbosity = 0
	s.VerbEveryConflicts = 10000
	s.ShowModel = 0
	s.K = opt_K_glucose.Value()
	s.R = opt_R_glucose.Value()
	s.SizeLBDQueue = float64(opt_size_lbd_queue_glucose.Value())
	s.SizeTrailQueue = float64(opt_size_trail_queue_glucose.Value())
	s.FirstReduceDB = int(opt_first_reduce_db_glucose.Value())
	s.IncReduceDB = int(opt_inc_reduce_db_glucose.Value())
	s.SpecialIncReduceDB = int(opt_spec_inc_reduce_db_glucose.Value())
	s.LbLBDFrozenClause = uint32(opt_lb_lbd_frozen_clause_glucose.Value())
	s.LbSizeMinimizingClause = int(opt_lb_size_minimzing_clause_glucose.Value())
	s.LbLBDMinimizingClause = uint32(opt_lb_lbd_minimzing_clause_glucose.Value())
	s.Var_decay = opt_var_decay_glucose.Value()
	s.Clause_decay = opt_clause_decay_glucose.Value()
	s.Random_var_freq = opt_random_var_freq_glucose.Value()
	s.Random_seed = opt_random_seed_glucose.Value()
	s.Ccmin_mode = int(opt_ccmin_mode_glucose.Value())
	s.Phase_saving = int(opt_phase_saving_glucose.Value())
	s.Rnd_pol = false
	s.Rnd_init_act = opt_rnd_init_act_glucose.Value()
	s.Garbage_frac = opt_garbage_frac_glucose.Value()
	s.CertifiedOutput = nil
	s.CertifiedUNSAT = opt_certified__glucose.Value()
	s.curRestart = 1

	s.ok = true
	s.cla_inc = 1
	s.var_inc = 1
	s.ca = *NewGlucoClauseAllocator()
	s.watches = *NewGlucoOccLists[GlucoLit, GlucoSolverWatcher, GlucoSolverWatcherDeleted](NewGlucoSolverWatcherDeleted(&s.ca))
	s.watchesBin = *NewGlucoOccLists[GlucoLit, GlucoSolverWatcher, GlucoSolverWatcherDeleted](NewGlucoSolverWatcherDeleted(&s.ca))
	s.qhead = 0
	s.simpDB_assigns = -1
	s.simpDB_props = 0
	s.order_heap = *NewGlucoHeap(NewGlucoSolverVarOrderLt(&s.activity))
	s.progress_estimate = 0
	s.remove_satisfied = true

	// Resource constraints:
	s.conflict_budget = -1
	s.propagation_budget = -1
	s.asynch_interrupt = false
	if opt_incremental_glucose.Value() {
		s.incremental = 1
	}
	s.nbVarsInitialFormula = math.MaxInt32

	s.MYFLAG = 0
	// Initialize only first time. Useful for incremental solving, useless otherwise
	s.lbdQueue.InitSize(int(s.SizeLBDQueue))
	s.trailQueue.InitSize(int(s.SizeTrailQueue))
	s.sumLBD = 0
	s.nbclausesbeforereduce = s.FirstReduceDB
	s.totalTime4Sat = 0
	s.totalTime4Unsat = 0
	s.nbSatCalls = 0
	s.nbUnsatCalls = 0

	if s.CertifiedUNSAT {
		var err error
		if opt_certified_file__glucose.Value() == "NULL" {
			s.CertifiedOutput, err = os.OpenFile("/dev/stdout", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		} else {
			s.CertifiedOutput, err = os.Create(opt_certified_file__glucose.Value())
		}
		_ = err // fopen returns NULL on failure; the C code does not check
	}
}

// Destroy is Solver::~Solver() (virtual).
// C: virtual Solver::~Solver()
func (s *GlucoSolver) Destroy() {}

// SetIncrementalMode is Solver::setIncrementalMode: sets the incremental mode to true.
// C: void setIncrementalMode()
func (s *GlucoSolver) SetIncrementalMode() { s.incremental = 1 }

// InitNbInitialVars is Solver::initNbInitialVars: number of variables without selectors.
// C: void initNbInitialVars(int nb)
func (s *GlucoSolver) InitNbInitialVars(nb int) { s.nbVarsInitialFormula = nb }

// NewVar is Solver::newVar (polarity and dvar default to true). Creates a new SAT
// variable; if 'dvar' is cleared, the variable will not be used as a decision variable.
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *GlucoSolver) NewVar(sign bool, dvar bool) GlucoVar {
	v := s.NVars()
	s.watches.Init(GlucoMkLit(v, false))
	s.watches.Init(GlucoMkLit(v, true))
	s.watchesBin.Init(GlucoMkLit(v, false))
	s.watchesBin.Init(GlucoMkLit(v, true))
	s.assigns.PushElem(GlucoLUndef)
	s.vardata.PushElem(glucoSolverMkVarData(GlucoCRefUndef, 0))
	if s.Rnd_init_act {
		s.activity.PushElem(glucoSolverDrand(&s.Random_seed) * 0.00001)
	} else {
		s.activity.PushElem(0)
	}
	s.seen.PushElem(0)
	s.permDiff.PushElem(0)
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

// AddVar is Solver::addVar (adds enough variables to make sure there is variable v).
// C: void addVar(Var v)
func (s *GlucoSolver) AddVar(v GlucoVar) {
	for v >= s.NVars() {
		s.NewVar(true, true)
	}
}

// glucoLitLess is LessThan_default<Lit> (Lit::operator<).
type glucoLitLess struct{}

func (glucoLitLess) Call(x GlucoLit, y GlucoLit) bool { return x.Lt(y) }

// glucoDimacsLit is the DRUP/DIMACS signed literal `(var(l) + 1) * (-2 * sign(l) + 1)`.
func glucoDimacsLit(l GlucoLit) int {
	sg := 0
	if GlucoSign(l) {
		sg = 1
	}
	return (GlucoLitVar(l) + 1) * (-2*sg + 1)
}

// AddClause_ is Solver::addClause_ (no superfluous internal copy; clobbers ps).
// C: bool addClause_(vec<Lit>& ps)
func (s *GlucoSolver) AddClause_(ps *GlucoVec[GlucoLit]) bool {
	if s.decisionLevel() != 0 {
		panic("GlucoSolver.AddClause_: decisionLevel() != 0")
	}
	if !s.ok {
		return false
	}

	// Check if clause is satisfied and remove false/duplicate literals:
	GlucoSortVec(ps, glucoLitLess{})

	var oc GlucoVec[GlucoLit]
	oc.Clear(false)

	var p GlucoLit
	var i, j int
	flag := 0
	if s.CertifiedUNSAT {
		for i, j, p = 0, 0, GlucoLitUndef; i < ps.Size(); i++ {
			oc.PushElem(*ps.At(i))
			if s.ValueLit(*ps.At(i)) == GlucoLTrue || *ps.At(i) == GlucoLitNot(p) || s.ValueLit(*ps.At(i)) == GlucoLFalse {
				flag = 1
			}
		}
	}

	for i, j, p = 0, 0, GlucoLitUndef; i < ps.Size(); i++ {
		if s.ValueLit(*ps.At(i)) == GlucoLTrue || *ps.At(i) == GlucoLitNot(p) {
			return true
		} else if s.ValueLit(*ps.At(i)) != GlucoLFalse && *ps.At(i) != p {
			p = *ps.At(i)
			*ps.At(j) = p
			j++
		}
	}
	ps.Shrink(i - j)

	if flag != 0 && s.CertifiedUNSAT {
		for i = 0; i < ps.Size(); i++ {
			fmt.Fprintf(s.CertifiedOutput, "%d ", glucoDimacsLit(*ps.At(i)))
		}
		fmt.Fprintf(s.CertifiedOutput, "0\n")

		fmt.Fprintf(s.CertifiedOutput, "d ")
		for i = 0; i < oc.Size(); i++ {
			fmt.Fprintf(s.CertifiedOutput, "%d ", glucoDimacsLit(*oc.At(i)))
		}
		fmt.Fprintf(s.CertifiedOutput, "0\n")
	}

	if ps.Size() == 0 {
		s.ok = false
		return false
	} else if ps.Size() == 1 {
		s.uncheckedEnqueue(*ps.At(0), GlucoCRefUndef)
		s.ok = s.propagate() == GlucoCRefUndef
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
func (s *GlucoSolver) AddClause(ps *GlucoVec[GlucoLit]) bool {
	ps.CopyTo(&s.add_tmp)
	return s.AddClause_(&s.add_tmp)
}

// AddEmptyClause is Solver::addEmptyClause.
// C: bool addEmptyClause()
func (s *GlucoSolver) AddEmptyClause() bool {
	s.add_tmp.Clear(false)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *GlucoSolver) AddClauseUnit(p GlucoLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *GlucoSolver) AddClauseBinary(p GlucoLit, q GlucoLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *GlucoSolver) AddClauseTernary(p GlucoLit, q GlucoLit, r GlucoLit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	s.add_tmp.PushElem(r)
	return s.AddClause_(&s.add_tmp)
}

// attachClause is Solver::attachClause (protected).
// C: void attachClause(CRef cr)
func (s *GlucoSolver) attachClause(cr GlucoCRef) {
	c := s.ca.At(cr)
	if c.Size() <= 1 {
		panic("GlucoSolver.attachClause: size <= 1")
	}
	if c.Size() == 2 {
		s.watchesBin.At(GlucoLitNot(*c.At(0))).PushElem(NewGlucoSolverWatcher(cr, *c.At(1)))
		s.watchesBin.At(GlucoLitNot(*c.At(1))).PushElem(NewGlucoSolverWatcher(cr, *c.At(0)))
	} else {
		s.watches.At(GlucoLitNot(*c.At(0))).PushElem(NewGlucoSolverWatcher(cr, *c.At(1)))
		s.watches.At(GlucoLitNot(*c.At(1))).PushElem(NewGlucoSolverWatcher(cr, *c.At(0)))
	}
	if c.Learnt() {
		s.Learnts_literals += int64(c.Size())
	} else {
		s.Clauses_literals += int64(c.Size())
	}
}

// detachClause is Solver::detachClause (protected; strict defaults to false).
// C: void detachClause(CRef cr, bool strict = false)
func (s *GlucoSolver) detachClause(cr GlucoCRef, strict bool) {
	c := s.ca.At(cr)
	if c.Size() <= 1 {
		panic("GlucoSolver.detachClause: size <= 1")
	}
	lists := &s.watches
	if c.Size() == 2 {
		lists = &s.watchesBin
	}
	if strict {
		glucoRemove[GlucoSolverWatcher](lists.At(GlucoLitNot(*c.At(0))), NewGlucoSolverWatcher(cr, *c.At(1)), GlucoSolverWatcher.Ne)
		glucoRemove[GlucoSolverWatcher](lists.At(GlucoLitNot(*c.At(1))), NewGlucoSolverWatcher(cr, *c.At(0)), GlucoSolverWatcher.Ne)
	} else {
		// Lazy detaching: (NOTE! Must clean all watcher lists before garbage collecting this clause)
		lists.Smudge(GlucoLitNot(*c.At(0)))
		lists.Smudge(GlucoLitNot(*c.At(1)))
	}
	if c.Learnt() {
		s.Learnts_literals -= int64(c.Size())
	} else {
		s.Clauses_literals -= int64(c.Size())
	}
}

// removeClause is Solver::removeClause (protected).
// C: void removeClause(CRef cr)
func (s *GlucoSolver) removeClause(cr GlucoCRef) {
	c := s.ca.At(cr)

	if s.CertifiedUNSAT {
		fmt.Fprintf(s.CertifiedOutput, "d ")
		for i := 0; i < c.Size(); i++ {
			fmt.Fprintf(s.CertifiedOutput, "%d ", glucoDimacsLit(*c.At(i)))
		}
		fmt.Fprintf(s.CertifiedOutput, "0\n")
	}

	s.detachClause(cr, false)
	// Don't leave pointers to free'd memory!
	if s.locked(c) {
		s.vardata.At(GlucoLitVar(*c.At(0))).reason = GlucoCRefUndef
	}
	c.SetMark(1)
	s.ca.Free_(cr)
}

// satisfied is Solver::satisfied (protected).
// C: bool satisfied(const Clause& c) const
func (s *GlucoSolver) satisfied(c *GlucoClause) bool {
	if s.incremental != 0 { // Check clauses with many selectors is too time consuming
		return s.ValueLit(*c.At(0)) == GlucoLTrue || s.ValueLit(*c.At(1)) == GlucoLTrue
	}
	// Default mode.
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) == GlucoLTrue {
			return true
		}
	}
	return false
}

// computeLBD is the overload computeLBD(const vec<Lit>& lits, int end = -1) (protected).
// C: unsigned int computeLBD(const vec<Lit>& lits, int end = -1)
func (s *GlucoSolver) computeLBD(lits *GlucoVec[GlucoLit], end int) uint32 {
	nblevels := 0
	s.MYFLAG++

	if s.incremental != 0 { // ----------------- INCREMENTAL MODE
		if end == -1 {
			end = lits.Size()
		}
		var nbDone uint32
		for i := 0; i < lits.Size(); i++ {
			if nbDone >= uint32(end) {
				break
			}
			if s.isSelector(GlucoLitVar(*lits.At(i))) {
				continue
			}
			nbDone++
			l := s.level(GlucoLitVar(*lits.At(i)))
			if *s.permDiff.At(l) != s.MYFLAG {
				*s.permDiff.At(l) = s.MYFLAG
				nblevels++
			}
		}
	} else { // -------- DEFAULT MODE. NOT A LOT OF DIFFERENCES... BUT EASIER TO READ
		for i := 0; i < lits.Size(); i++ {
			l := s.level(GlucoLitVar(*lits.At(i)))
			if *s.permDiff.At(l) != s.MYFLAG {
				*s.permDiff.At(l) = s.MYFLAG
				nblevels++
			}
		}
	}
	return uint32(nblevels)
}

// computeLBDClause is the overload computeLBD(const Clause& c) (protected).
// C: unsigned int computeLBD(const Clause& c)
func (s *GlucoSolver) computeLBDClause(c *GlucoClause) uint32 {
	nblevels := 0
	s.MYFLAG++

	if s.incremental != 0 { // ----------------- INCREMENTAL MODE
		nbDone := 0
		for i := 0; i < c.Size(); i++ {
			if uint32(nbDone) >= c.SizeWithoutSelectors() {
				break
			}
			if s.isSelector(GlucoLitVar(*c.At(i))) {
				continue
			}
			nbDone++
			l := s.level(GlucoLitVar(*c.At(i)))
			if *s.permDiff.At(l) != s.MYFLAG {
				*s.permDiff.At(l) = s.MYFLAG
				nblevels++
			}
		}
	} else { // -------- DEFAULT MODE. NOT A LOT OF DIFFERENCES... BUT EASIER TO READ
		for i := 0; i < c.Size(); i++ {
			l := s.level(GlucoLitVar(*c.At(i)))
			if *s.permDiff.At(l) != s.MYFLAG {
				*s.permDiff.At(l) = s.MYFLAG
				nblevels++
			}
		}
	}
	return uint32(nblevels)
}

// isSelector is Solver::isSelector (protected inline).
// C: inline bool isSelector(Var v)
func (s *GlucoSolver) isSelector(v GlucoVar) bool {
	return s.incremental != 0 && v > s.nbVarsInitialFormula
}

// minimisationWithBinaryResolution is Solver::minimisationWithBinaryResolution (protected).
// C: void minimisationWithBinaryResolution(vec<Lit>& out_learnt)
func (s *GlucoSolver) minimisationWithBinaryResolution(outLearnt *GlucoVec[GlucoLit]) {
	// Find the LBD measure
	lbd := s.computeLBD(outLearnt, -1)
	p := GlucoLitNot(*outLearnt.At(0))

	if lbd <= s.LbLBDMinimizingClause {
		s.MYFLAG++

		for i := 1; i < outLearnt.Size(); i++ {
			*s.permDiff.At(GlucoLitVar(*outLearnt.At(i))) = s.MYFLAG
		}

		wbin := s.watchesBin.At(p)
		nb := 0
		for k := 0; k < wbin.Size(); k++ {
			imp := wbin.At(k).blocker
			if *s.permDiff.At(GlucoLitVar(imp)) == s.MYFLAG && s.ValueLit(imp) == GlucoLTrue {
				nb++
				*s.permDiff.At(GlucoLitVar(imp)) = s.MYFLAG - 1
			}
		}
		l := outLearnt.Size() - 1
		if nb > 0 {
			s.NbReducedClauses++
			for i := 1; i < outLearnt.Size()-nb; i++ {
				if *s.permDiff.At(GlucoLitVar(*outLearnt.At(i))) != s.MYFLAG {
					p := *outLearnt.At(l)
					*outLearnt.At(l) = *outLearnt.At(i)
					*outLearnt.At(i) = p
					l--
					i--
				}
			}
			outLearnt.Shrink(nb)
		}
	}
}

// cancelUntil is Solver::cancelUntil (protected): revert to the state at given
// level (keeping all assignment at 'level' but not beyond).
// C: void cancelUntil(int level)
func (s *GlucoSolver) cancelUntil(level int) {
	if s.decisionLevel() > level {
		for c := s.trail.Size() - 1; c >= *s.trail_lim.At(level); c-- {
			x := GlucoLitVar(*s.trail.At(c))
			*s.assigns.At(x) = GlucoLUndef
			if s.Phase_saving > 1 || ((s.Phase_saving == 1) && c > *s.trail_lim.Last()) {
				if GlucoSign(*s.trail.At(c)) {
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
func (s *GlucoSolver) pickBranchLit() GlucoLit {
	next := GlucoVarUndef
	// Random decision:
	if glucoSolverDrand(&s.Random_seed) < s.Random_var_freq && !s.order_heap.Empty() {
		next = s.order_heap.At(glucoSolverIrand(&s.Random_seed, s.order_heap.Size()))
		if s.Value(next) == GlucoLUndef && *s.decision.At(next) != 0 {
			s.Rnd_decisions++
		}
	}
	// Activity based decision:
	for next == GlucoVarUndef || s.Value(next) != GlucoLUndef || *s.decision.At(next) == 0 {
		if s.order_heap.Empty() {
			next = GlucoVarUndef
			break
		}
		next = s.order_heap.RemoveMin()
	}
	if next == GlucoVarUndef {
		return GlucoLitUndef
	}
	var sg bool
	if s.Rnd_pol {
		sg = glucoSolverDrand(&s.Random_seed) < 0.5
	} else {
		sg = *s.polarity.At(next) != 0
	}
	return GlucoMkLit(next, sg)
}

// analyze is Solver::analyze (protected): analyze conflict and produce a reason clause.
//
//	Pre-conditions: 'out_learnt' is assumed to be cleared; current decision level must be greater than root level.
//	Post-conditions: 'out_learnt[0]' is the asserting literal at level 'out_btlevel'; if out_learnt.size() > 1
//	then 'out_learnt[1]' has the greatest decision level of the rest of literals.
//
// C: void analyze(CRef confl, vec<Lit>& out_learnt, vec<Lit>& selectors, int& out_btlevel, unsigned int& lbd, unsigned int& szWithoutSelectors)
func (s *GlucoSolver) analyze(confl GlucoCRef, outLearnt *GlucoVec[GlucoLit], selectors *GlucoVec[GlucoLit], outBtlevel *int, lbd *uint32, szWithoutSelectors *uint32) {
	pathC := 0
	p := GlucoLitUndef

	// Generate conflict clause:
	outLearnt.Push() // (leave room for the asserting literal)
	index := s.trail.Size() - 1

	for {
		if confl == GlucoCRefUndef { // (otherwise should be UIP)
			panic("GlucoSolver.analyze: confl == CRef_Undef")
		}
		c := s.ca.At(confl)

		// Special case for binary clauses
		// The first one has to be SAT
		if p != GlucoLitUndef && c.Size() == 2 && s.ValueLit(*c.At(0)) == GlucoLFalse {
			if s.ValueLit(*c.At(1)) != GlucoLTrue {
				panic("GlucoSolver.analyze: binary reason without a true literal")
			}
			tmp := *c.At(0)
			*c.At(0) = *c.At(1)
			*c.At(1) = tmp
		}

		if c.Learnt() {
			s.claBumpActivity(c)
		}

		// DYNAMIC NBLEVEL trick (see competition'09 companion paper)
		if c.Learnt() && c.Lbd() > 2 {
			nblevels := s.computeLBDClause(c)
			if nblevels+1 < c.Lbd() { // improve the LBD
				if c.Lbd() <= s.LbLBDFrozenClause {
					c.SetCanBeDel(false)
				}
				// seems to be interesting : keep it for the next round
				c.SetLBD(int(nblevels)) // Update it
			}
		}

		j0 := 1
		if p == GlucoLitUndef {
			j0 = 0
		}
		for j := j0; j < c.Size(); j++ {
			q := *c.At(j)

			if *s.seen.At(GlucoLitVar(q)) == 0 && s.level(GlucoLitVar(q)) > 0 {
				if !s.isSelector(GlucoLitVar(q)) {
					s.varBumpActivity(GlucoLitVar(q))
				}
				*s.seen.At(GlucoLitVar(q)) = 1
				if s.level(GlucoLitVar(q)) >= s.decisionLevel() {
					pathC++
					// UPDATEVARACTIVITY trick (see competition'09 companion paper)
					if !s.isSelector(GlucoLitVar(q)) && (s.reason(GlucoLitVar(q)) != GlucoCRefUndef) && s.ca.At(s.reason(GlucoLitVar(q))).Learnt() {
						s.lastDecisionLevel.PushElem(q)
					}
				} else {
					if s.isSelector(GlucoLitVar(q)) {
						if s.ValueLit(q) != GlucoLFalse {
							panic("GlucoSolver.analyze: selector literal is not false")
						}
						selectors.PushElem(q)
					} else {
						outLearnt.PushElem(q)
					}
				}
			}
		}

		// Select next clause to look at:
		for {
			v := GlucoLitVar(*s.trail.At(index))
			index--
			if *s.seen.At(v) != 0 {
				break
			}
		}
		p = *s.trail.At(index + 1)
		confl = s.reason(GlucoLitVar(p))
		*s.seen.At(GlucoLitVar(p)) = 0
		pathC--

		if !(pathC > 0) {
			break
		}
	}
	*outLearnt.At(0) = GlucoLitNot(p)

	// Simplify conflict clause:
	var i, j int
	for i = 0; i < selectors.Size(); i++ {
		outLearnt.PushElem(*selectors.At(i))
	}
	outLearnt.CopyTo_(&s.analyze_toclear)
	if s.Ccmin_mode == 2 {
		var abstractLevel uint32
		for i = 1; i < outLearnt.Size(); i++ {
			abstractLevel |= s.abstractLevel(GlucoLitVar(*outLearnt.At(i))) // (maintain an abstraction of levels involved in conflict)
		}
		for i, j = 1, 1; i < outLearnt.Size(); i++ {
			if s.reason(GlucoLitVar(*outLearnt.At(i))) == GlucoCRefUndef || !s.litRedundant(*outLearnt.At(i), abstractLevel) {
				*outLearnt.At(j) = *outLearnt.At(i)
				j++
			}
		}
	} else if s.Ccmin_mode == 1 {
		for i, j = 1, 1; i < outLearnt.Size(); i++ {
			x := GlucoLitVar(*outLearnt.At(i))

			if s.reason(x) == GlucoCRefUndef {
				*outLearnt.At(j) = *outLearnt.At(i)
				j++
			} else {
				c := s.ca.At(s.reason(GlucoLitVar(*outLearnt.At(i))))
				// Thanks to Siert Wieringa for this bug fix!
				k0 := 1
				if c.Size() == 2 {
					k0 = 0
				}
				for k := k0; k < c.Size(); k++ {
					if *s.seen.At(GlucoLitVar(*c.At(k))) == 0 && s.level(GlucoLitVar(*c.At(k))) > 0 {
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
	s.Max_literals += int64(outLearnt.Size())
	outLearnt.Shrink(i - j)
	s.Tot_literals += int64(outLearnt.Size())

	/* ***************************************
	   Minimisation with binary clauses of the asserting clause
	   First of all : we look for small clauses
	   Then, we reduce clauses with small LBD.
	   Otherwise, this can be useless
	*/
	if s.incremental == 0 && outLearnt.Size() <= s.LbSizeMinimizingClause {
		s.minimisationWithBinaryResolution(outLearnt)
	}

	// Find correct backtrack level:
	if outLearnt.Size() == 1 {
		*outBtlevel = 0
	} else {
		maxI := 1
		// Find the first literal assigned at the next-highest level:
		for i := 2; i < outLearnt.Size(); i++ {
			if s.level(GlucoLitVar(*outLearnt.At(i))) > s.level(GlucoLitVar(*outLearnt.At(maxI))) {
				maxI = i
			}
		}
		// Swap-in this literal at index 1:
		p := *outLearnt.At(maxI)
		*outLearnt.At(maxI) = *outLearnt.At(1)
		*outLearnt.At(1) = p
		*outBtlevel = s.level(GlucoLitVar(p))
	}

	// Compute the size of the clause without selectors (incremental mode)
	if s.incremental != 0 {
		*szWithoutSelectors = 0
		for i := 0; i < outLearnt.Size(); i++ {
			if !s.isSelector(GlucoLitVar(*outLearnt.At(i))) {
				*szWithoutSelectors++
			} else if i > 0 {
				break
			}
		}
	} else {
		*szWithoutSelectors = uint32(outLearnt.Size())
	}

	// Compute LBD
	*lbd = s.computeLBD(outLearnt, outLearnt.Size()-selectors.Size())

	// UPDATEVARACTIVITY trick (see competition'09 companion paper)
	if s.lastDecisionLevel.Size() > 0 {
		for i := 0; i < s.lastDecisionLevel.Size(); i++ {
			if s.ca.At(s.reason(GlucoLitVar(*s.lastDecisionLevel.At(i)))).Lbd() < *lbd {
				s.varBumpActivity(GlucoLitVar(*s.lastDecisionLevel.At(i)))
			}
		}
		s.lastDecisionLevel.Clear(false)
	}

	for j = 0; j < s.analyze_toclear.Size(); j++ {
		*s.seen.At(GlucoLitVar(*s.analyze_toclear.At(j))) = 0 // ('seen[]' is now cleared)
	}
	for j = 0; j < selectors.Size(); j++ {
		*s.seen.At(GlucoLitVar(*selectors.At(j))) = 0
	}
}

// litRedundant is Solver::litRedundant (protected): check if 'p' can be removed.
// 'abstract_levels' is used to abort early if the algorithm is visiting literals
// at levels that cannot be removed later.
// C: bool litRedundant(Lit p, uint32_t abstract_levels)
func (s *GlucoSolver) litRedundant(p GlucoLit, abstractLevels uint32) bool {
	s.analyze_stack.Clear(false)
	s.analyze_stack.PushElem(p)
	top := s.analyze_toclear.Size()
	for s.analyze_stack.Size() > 0 {
		if s.reason(GlucoLitVar(*s.analyze_stack.Last())) == GlucoCRefUndef {
			panic("GlucoSolver.litRedundant: reason == CRef_Undef")
		}
		c := s.ca.At(s.reason(GlucoLitVar(*s.analyze_stack.Last())))
		s.analyze_stack.Pop()
		if c.Size() == 2 && s.ValueLit(*c.At(0)) == GlucoLFalse {
			if s.ValueLit(*c.At(1)) != GlucoLTrue {
				panic("GlucoSolver.litRedundant: binary reason without a true literal")
			}
			tmp := *c.At(0)
			*c.At(0) = *c.At(1)
			*c.At(1) = tmp
		}

		for i := 1; i < c.Size(); i++ {
			p := *c.At(i)
			if *s.seen.At(GlucoLitVar(p)) == 0 && s.level(GlucoLitVar(p)) > 0 {
				if s.reason(GlucoLitVar(p)) != GlucoCRefUndef && (s.abstractLevel(GlucoLitVar(p))&abstractLevels) != 0 {
					*s.seen.At(GlucoLitVar(p)) = 1
					s.analyze_stack.PushElem(p)
					s.analyze_toclear.PushElem(p)
				} else {
					for j := top; j < s.analyze_toclear.Size(); j++ {
						*s.seen.At(GlucoLitVar(*s.analyze_toclear.At(j))) = 0
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
// to express the final conflict in terms of assumptions.
// C: void analyzeFinal(Lit p, vec<Lit>& out_conflict)
func (s *GlucoSolver) analyzeFinal(p GlucoLit, outConflict *GlucoVec[GlucoLit]) {
	outConflict.Clear(false)
	outConflict.PushElem(p)

	if s.decisionLevel() == 0 {
		return
	}

	*s.seen.At(GlucoLitVar(p)) = 1

	for i := s.trail.Size() - 1; i >= *s.trail_lim.At(0); i-- {
		x := GlucoLitVar(*s.trail.At(i))
		if *s.seen.At(x) != 0 {
			if s.reason(x) == GlucoCRefUndef {
				if s.level(x) <= 0 {
					panic("GlucoSolver.analyzeFinal: level(x) <= 0")
				}
				outConflict.PushElem(GlucoLitNot(*s.trail.At(i)))
			} else {
				c := s.ca.At(s.reason(x))
				// Bug in case of assumptions due to special data structures for Binary.
				// Many thanks to Sam Bayless (sbayless@cs.ubc.ca) for discover this bug.
				j0 := 1
				if c.Size() == 2 {
					j0 = 0
				}
				for j := j0; j < c.Size(); j++ {
					if s.level(GlucoLitVar(*c.At(j))) > 0 {
						*s.seen.At(GlucoLitVar(*c.At(j))) = 1
					}
				}
			}
			*s.seen.At(x) = 0
		}
	}

	*s.seen.At(GlucoLitVar(p)) = 0
}

// uncheckedEnqueue is Solver::uncheckedEnqueue (protected; from defaults to CRef_Undef).
// C: void uncheckedEnqueue(Lit p, CRef from = CRef_Undef)
func (s *GlucoSolver) uncheckedEnqueue(p GlucoLit, from GlucoCRef) {
	if s.ValueLit(p) != GlucoLUndef {
		panic("GlucoSolver.uncheckedEnqueue: literal already assigned")
	}
	*s.assigns.At(GlucoLitVar(p)) = NewGlucoLboolB(!GlucoSign(p))
	*s.vardata.At(GlucoLitVar(p)) = glucoSolverMkVarData(from, s.decisionLevel())
	s.trail.Push_(p)
}

// enqueue is Solver::enqueue (protected; from defaults to CRef_Undef): test if
// fact 'p' contradicts current state, enqueue otherwise.
// C: bool enqueue(Lit p, CRef from = CRef_Undef)
func (s *GlucoSolver) enqueue(p GlucoLit, from GlucoCRef) bool {
	if s.ValueLit(p) != GlucoLUndef {
		return s.ValueLit(p) != GlucoLFalse
	}
	s.uncheckedEnqueue(p, from)
	return true
}

// propagate is Solver::propagate (protected): propagates all enqueued facts. If a
// conflict arises, the conflicting clause is returned, otherwise CRef_Undef.
// Post-condition: the propagation queue is empty, even if there was a conflict.
// C: CRef propagate()
func (s *GlucoSolver) propagate() GlucoCRef {
	confl := GlucoCRefUndef
	numProps := 0
	s.watches.CleanAll()
	s.watchesBin.CleanAll()

	for s.qhead < s.trail.Size() {
		p := *s.trail.At(s.qhead) // 'p' is enqueued fact to propagate.
		s.qhead++
		ws := s.watches.At(p)
		wd := ws.Data()
		var i, j int
		end := ws.Size()
		numProps++

		// First, Propagate binary clauses
		wbin := s.watchesBin.At(p)
		for k := 0; k < wbin.Size(); k++ {
			imp := wbin.At(k).blocker
			if s.ValueLit(imp) == GlucoLFalse {
				return wbin.At(k).cref
			}
			if s.ValueLit(imp) == GlucoLUndef {
				s.uncheckedEnqueue(imp, wbin.At(k).cref)
			}
		}

	NextWatcher:
		for i, j = 0, 0; i != end; {
			// Try to avoid inspecting the clause:
			blocker := wd[i].blocker
			if s.ValueLit(blocker) == GlucoLTrue {
				wd[j] = wd[i]
				j++
				i++
				continue
			}

			// Make sure the false literal is data[1]:
			cr := wd[i].cref
			c := s.ca.At(cr)
			falseLit := GlucoLitNot(p)
			if *c.At(0) == falseLit {
				*c.At(0) = *c.At(1)
				*c.At(1) = falseLit
			}
			if *c.At(1) != falseLit {
				panic("GlucoSolver.propagate: c[1] != false_lit")
			}
			i++

			// If 0th watch is true, then clause is already satisfied.
			first := *c.At(0)
			w := NewGlucoSolverWatcher(cr, first)
			if first != blocker && s.ValueLit(first) == GlucoLTrue {
				wd[j] = w
				j++
				continue
			}

			// Look for new watch:
			if s.incremental != 0 { // ----------------- INCREMENTAL MODE
				choosenPos := -1
				for k := 2; k < c.Size(); k++ {
					if s.ValueLit(*c.At(k)) != GlucoLFalse {
						if s.decisionLevel() > s.assumptions.Size() {
							choosenPos = k
							break
						}
						choosenPos = k
						if s.ValueLit(*c.At(k)) == GlucoLTrue || !s.isSelector(GlucoLitVar(*c.At(k))) {
							break
						}
					}
				}
				if choosenPos != -1 {
					*c.At(1) = *c.At(choosenPos)
					*c.At(choosenPos) = falseLit
					s.watches.At(GlucoLitNot(*c.At(1))).PushElem(w)
					continue NextWatcher
				}
			} else { // ----------------- DEFAULT  MODE (NOT INCREMENTAL)
				for k := 2; k < c.Size(); k++ {
					if s.ValueLit(*c.At(k)) != GlucoLFalse {
						*c.At(1) = *c.At(k)
						*c.At(k) = falseLit
						s.watches.At(GlucoLitNot(*c.At(1))).PushElem(w)
						continue NextWatcher
					}
				}
			}

			// Did not find watch -- clause is unit under assignment:
			wd[j] = w
			j++
			if s.ValueLit(first) == GlucoLFalse {
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
	s.Propagations += int64(numProps)
	s.simpDB_props -= int64(numProps)

	return confl
}

// glucoReduceDBLt is the local struct reduceDB_lt of Glucose.cpp.
type glucoReduceDBLt struct {
	ca *GlucoClauseAllocator
}

// Call is reduceDB_lt::operator().
// C: bool operator()(CRef x, CRef y)
func (l glucoReduceDBLt) Call(x GlucoCRef, y GlucoCRef) bool {
	cx, cy := l.ca.At(x), l.ca.At(y)
	// Main criteria... Like in MiniSat we keep all binary clauses
	if cx.Size() > 2 && cy.Size() == 2 {
		return true
	}
	if cy.Size() > 2 && cx.Size() == 2 {
		return false
	}
	if cx.Size() == 2 && cy.Size() == 2 {
		return false
	}
	// Second one based on literal block distance
	if cx.Lbd() > cy.Lbd() {
		return true
	}
	if cx.Lbd() < cy.Lbd() {
		return false
	}
	// Finally we can use old activity or size, we choose the last one
	return *cx.Activity() < *cy.Activity()
}

// reduceDB is Solver::reduceDB (protected): remove half of the learnt clauses,
// minus the clauses locked by the current assignment; binary clauses are never removed.
// C: void reduceDB()
func (s *GlucoSolver) reduceDB() {
	var i, j int
	s.NbReduceDB++
	GlucoSortVec(&s.learnts, glucoReduceDBLt{ca: &s.ca})

	// We have a lot of "good" clauses, it is difficult to compare them. Keep more !
	if s.ca.At(*s.learnts.At(s.learnts.Size() / glucoRatioRemoveClauses)).Lbd() <= 3 {
		s.nbclausesbeforereduce += s.SpecialIncReduceDB
	}
	// Useless :-)
	if s.ca.At(*s.learnts.Last()).Lbd() <= 5 {
		s.nbclausesbeforereduce += s.SpecialIncReduceDB
	}

	// Don't delete binary or locked clauses. From the rest, delete clauses from the first half
	// Keep clauses which seem to be usefull (their lbd was reduce during this sequence)
	limit := s.learnts.Size() / 2

	for i, j = 0, 0; i < s.learnts.Size(); i++ {
		c := s.ca.At(*s.learnts.At(i))
		if c.Lbd() > 2 && c.Size() > 2 && c.CanBeDel() && !s.locked(c) && (i < limit) {
			s.removeClause(*s.learnts.At(i))
			s.NbRemovedClauses++
		} else {
			if !c.CanBeDel() {
				limit++ // we keep c, so we can delete an other clause
			}
			c.SetCanBeDel(true) // At the next step, c can be delete
			*s.learnts.At(j) = *s.learnts.At(i)
			j++
		}
	}
	s.learnts.Shrink(i - j)
	s.CheckGarbage()
}

// glucoRatioRemoveClauses is the macro RATIOREMOVECLAUSES of Constants.h.
const glucoRatioRemoveClauses = GlucoseRemoveClauseRatio

// removeSatisfied is Solver::removeSatisfied (protected).
// C: void removeSatisfied(vec<CRef>& cs)
func (s *GlucoSolver) removeSatisfied(cs *GlucoVec[GlucoCRef]) {
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
func (s *GlucoSolver) rebuildOrderHeap() {
	var vs GlucoVec[GlucoVar]
	for v := 0; v < s.NVars(); v++ {
		if *s.decision.At(v) != 0 && s.Value(v) == GlucoLUndef {
			vs.PushElem(v)
		}
	}
	s.order_heap.Build(&vs)
}

// Simplify is Solver::simplify: simplify the clause database according to the
// current top-level assigment (currently only removes satisfied clauses).
// C: bool simplify()
func (s *GlucoSolver) Simplify() bool {
	if s.decisionLevel() != 0 {
		panic("GlucoSolver.Simplify: decisionLevel() != 0")
	}
	if !s.ok || s.propagate() != GlucoCRefUndef {
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
	s.simpDB_props = s.Clauses_literals + s.Learnts_literals // (shouldn't depend on stats really, but it will do for now)
	return true
}

// search is Solver::search (protected): search for a model. Use a negative value
// for 'nof_conflicts' to indicate infinity (the parameter is unused by Glucose).
// Returns l_True if a partial assignment consistent with the clause set is found,
// l_False if the clause set is unsatisfiable, l_Undef if the bound is reached.
// C: lbool search(int nof_conflicts)
func (s *GlucoSolver) search(nofConflicts int) GlucoLbool {
	if !s.ok {
		panic("GlucoSolver.search: !ok")
	}
	var backtrackLevel int
	conflictC := 0
	var learntClause, selectors GlucoVec[GlucoLit]
	var nblevels, szWoutSelectors uint32
	blocked := false
	s.Starts++

	for {
		confl := s.propagate()
		if confl != GlucoCRefUndef {
			// CONFLICT
			s.Conflicts++
			conflictC++
			s.ConflictsRestarts++
			if s.Conflicts%5000 == 0 && s.Var_decay < 0.95 {
				s.Var_decay += 0.01
			}

			if s.Verbosity >= 1 && s.Conflicts%int64(s.VerbEveryConflicts) == 0 {
				lim := s.trail.Size()
				if s.trail_lim.Size() != 0 {
					lim = *s.trail_lim.At(0)
				}
				fmt.Printf("c | %8d   %7d    %5d | %7d %8d %8d | %5d %8d   %6d %8d | %6.3f %% |\n",
					int(s.Starts), int(s.Nbstopsrestarts), int(s.Conflicts/s.Starts),
					int(s.Dec_vars)-lim, s.NClauses(), int(s.Clauses_literals),
					int(s.NbReduceDB), s.NLearnts(), int(s.NbDL2), int(s.NbRemovedClauses), s.progressEstimate()*100)
			}
			if s.decisionLevel() == 0 {
				return GlucoLFalse
			}

			s.trailQueue.Push(uint32(s.trail.Size()))
			// BLOCK RESTART (CP 2012 paper)
			if s.ConflictsRestarts > glucoLowerBoundForBlockingRestart && s.lbdQueue.Isvalid() != 0 && float64(s.trail.Size()) > s.R*float64(s.trailQueue.Getavg()) {
				s.lbdQueue.Fastclear()
				s.Nbstopsrestarts++
				if !blocked {
					s.Lastblockatrestart = s.Starts
					s.Nbstopsrestartssame++
					blocked = true
				}
			}

			learntClause.Clear(false)
			selectors.Clear(false)
			s.analyze(confl, &learntClause, &selectors, &backtrackLevel, &nblevels, &szWoutSelectors)

			s.lbdQueue.Push(nblevels)
			s.sumLBD += float32(nblevels)

			s.cancelUntil(backtrackLevel)

			if s.CertifiedUNSAT {
				for i := 0; i < learntClause.Size(); i++ {
					fmt.Fprintf(s.CertifiedOutput, "%d ", glucoDimacsLit(*learntClause.At(i)))
				}
				fmt.Fprintf(s.CertifiedOutput, "0\n")
			}

			if learntClause.Size() == 1 {
				s.uncheckedEnqueue(*learntClause.At(0), GlucoCRefUndef)
				s.NbUn++
			} else {
				cr := s.ca.Alloc(&learntClause, true)
				s.ca.At(cr).SetLBD(int(nblevels))
				s.ca.At(cr).SetSizeWithoutSelectors(szWoutSelectors)
				if nblevels <= 2 {
					s.NbDL2++ // stats
				}
				if s.ca.At(cr).Size() == 2 {
					s.NbBin++ // stats
				}
				s.learnts.PushElem(cr)
				s.attachClause(cr)

				s.claBumpActivity(s.ca.At(cr))
				s.uncheckedEnqueue(*learntClause.At(0), cr)
			}

			s.varDecayActivity()
			s.claDecayActivity()
		} else {
			// Our dynamic restart, see the SAT09 competition compagnion paper
			if (s.ConflictsRestarts != 0 && s.lbdQueue.Isvalid() != 0 && float64(s.lbdQueue.Getavg())*s.K > float64(s.sumLBD/float32(s.ConflictsRestarts))) || (s.Pstop != nil && *s.Pstop != 0) {
				s.lbdQueue.Fastclear()
				s.progress_estimate = s.progressEstimate()
				bt := 0
				if s.incremental != 0 { // DO NOT BACKTRACK UNTIL 0.. USELESS
					if s.decisionLevel() < s.assumptions.Size() {
						bt = s.decisionLevel()
					} else {
						bt = s.assumptions.Size()
					}
				}
				s.cancelUntil(bt)
				return GlucoLUndef
			}

			// Simplify the set of problem clauses:
			if s.decisionLevel() == 0 && !s.Simplify() {
				return GlucoLFalse
			}
			// Perform clause database reduction !
			if s.Conflicts >= s.curRestart*int64(s.nbclausesbeforereduce) {
				if s.learnts.Size() <= 0 {
					panic("GlucoSolver.search: no learnt clauses to reduce")
				}
				s.curRestart = (s.Conflicts / int64(s.nbclausesbeforereduce)) + 1
				s.reduceDB()
				s.nbclausesbeforereduce += s.IncReduceDB
			}

			next := GlucoLitUndef
			for s.decisionLevel() < s.assumptions.Size() {
				// Perform user provided assumption:
				p := *s.assumptions.At(s.decisionLevel())
				if s.ValueLit(p) == GlucoLTrue {
					// Dummy decision level:
					s.newDecisionLevel()
				} else if s.ValueLit(p) == GlucoLFalse {
					s.analyzeFinal(GlucoLitNot(p), &s.Conflict)
					return GlucoLFalse
				} else {
					next = p
					break
				}
			}

			if next == GlucoLitUndef {
				// New variable decision:
				s.Decisions++
				next = s.pickBranchLit()
				if next == GlucoLitUndef {
					// Model found:
					return GlucoLTrue
				}
			}

			// Increase decision level and enqueue 'next'
			s.newDecisionLevel()
			s.uncheckedEnqueue(next, GlucoCRefUndef)
		}
	}
}

// glucoLowerBoundForBlockingRestart is the macro LOWER_BOUND_FOR_BLOCKING_RESTART of Constants.h.
const glucoLowerBoundForBlockingRestart = GlucoseBlockingRestartLower

// progressEstimate is Solver::progressEstimate (protected).
// C: double progressEstimate() const
func (s *GlucoSolver) progressEstimate() float64 {
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

// PrintIncrementalStats is Solver::printIncrementalStats.
// C: void printIncrementalStats()
func (s *GlucoSolver) PrintIncrementalStats() {
	fmt.Printf("c---------- Glucose Stats -------------------------\n")
	fmt.Printf("c restarts              : %d\n", s.Starts)
	fmt.Printf("c nb ReduceDB           : %d\n", s.NbReduceDB)
	fmt.Printf("c nb removed Clauses    : %d\n", s.NbRemovedClauses)
	fmt.Printf("c nb learnts DL2        : %d\n", s.NbDL2)
	fmt.Printf("c nb learnts size 2     : %d\n", s.NbBin)
	fmt.Printf("c nb learnts size 1     : %d\n", s.NbUn)
	fmt.Printf("c conflicts             : %d\n", s.Conflicts)
	fmt.Printf("c decisions             : %d\n", s.Decisions)
	fmt.Printf("c propagations          : %d\n", s.Propagations)
	fmt.Printf("c SAT Calls             : %d in %.6g seconds\n", s.nbSatCalls, s.totalTime4Sat)
	fmt.Printf("c UNSAT Calls           : %d in %.6g seconds\n", s.nbUnsatCalls, s.totalTime4Unsat)
	fmt.Printf("c--------------------------------------------------\n")
}

// solve_ is Solver::solve_ (protected): main solve method (assumptions given in 'assumptions').
// C: lbool solve_()
func (s *GlucoSolver) solve_() GlucoLbool {
	if s.incremental != 0 && s.CertifiedUNSAT {
		fmt.Printf("Can not use incremental and certified unsat in the same time\n")
		os.Exit(255) // exit(-1)
	}

	s.Model.Clear(false)
	s.Conflict.Clear(false)
	if !s.ok {
		return GlucoLFalse
	}

	curTime := GlucoCpuTime()

	s.Solves++

	status := GlucoLUndef
	if s.incremental == 0 && s.Verbosity >= 1 {
		fmt.Printf("c ========================================[ MAGIC CONSTANTS ]==============================================\n")
		fmt.Printf("c | Constants are supposed to work well together :-)                                                      |\n")
		fmt.Printf("c | however, if you find better choices, please let us known...                                           |\n")
		fmt.Printf("c |-------------------------------------------------------------------------------------------------------|\n")
		fmt.Printf("c |                                |                                |                                     |\n")
		fmt.Printf("c | - Restarts:                    | - Reduce Clause DB:            | - Minimize Asserting:               |\n")
		fmt.Printf("c |   * LBD Queue    : %6d      |   * First     : %6d         |    * size < %3d                     |\n", s.lbdQueue.MaxSize(), s.nbclausesbeforereduce, s.LbSizeMinimizingClause)
		fmt.Printf("c |   * Trail  Queue : %6d      |   * Inc       : %6d         |    * lbd  < %3d                     |\n", s.trailQueue.MaxSize(), s.IncReduceDB, s.LbLBDMinimizingClause)
		fmt.Printf("c |   * K            : %6.2f      |   * Special   : %6d         |                                     |\n", s.K, s.SpecialIncReduceDB)
		fmt.Printf("c |   * R            : %6.2f      |   * Protected :  (lbd)< %2d     |                                     |\n", s.R, s.LbLBDFrozenClause)
		fmt.Printf("c |                                |                                |                                     |\n")
		fmt.Printf("c ==================================[ Search Statistics (every %6d conflicts) ]=========================\n", s.VerbEveryConflicts)
		fmt.Printf("c |                                                                                                       |\n")
		fmt.Printf("c |          RESTARTS           |          ORIGINAL         |              LEARNT              | Progress |\n")
		fmt.Printf("c |       NB   Blocked  Avg Cfc |    Vars  Clauses Literals |   Red   Learnts    LBD2  Removed |          |\n")
		fmt.Printf("c =========================================================================================================\n")
	}

	// Search:
	currRestarts := 0
	for status == GlucoLUndef {
		status = s.search(0) // the parameter is useless in glucose, kept to allow modifications
		if !s.withinBudget() || s.Terminate_search_early || (s.Pstop != nil && *s.Pstop != 0) {
			break
		}
		if s.NRuntimeLimit != 0 && AbcClock() > s.NRuntimeLimit {
			break
		}
		currRestarts++
	}

	if s.incremental == 0 && s.Verbosity >= 1 {
		fmt.Printf("c =========================================================================================================\n")
	}

	if s.CertifiedUNSAT { // Want certified output
		if status == GlucoLFalse {
			fmt.Fprintf(s.CertifiedOutput, "0\n")
		}
		s.CertifiedOutput.Close()
	}

	if status == GlucoLTrue {
		// Extend & copy model:
		s.Model.GrowTo(s.NVars())
		for i := 0; i < s.NVars(); i++ {
			*s.Model.At(i) = s.Value(i)
		}
	} else if status == GlucoLFalse && s.Conflict.Size() == 0 {
		s.ok = false
	}

	s.cancelUntil(0)

	finalTime := GlucoCpuTime()
	if status == GlucoLTrue {
		s.nbSatCalls++
		s.totalTime4Sat += finalTime - curTime
	}
	if status == GlucoLFalse {
		s.nbUnsatCalls++
		s.totalTime4Unsat += finalTime - curTime
	}

	// ABC callback
	if s.PCnfFunc != nil && !s.Terminate_search_early { // hack to avoid calling callback twise if the solver was terminated early
		var pCex []int
		message := -1
		if status == GlucoLTrue {
			message = 1
		} else if status == GlucoLFalse {
			message = 0
		}
		if status == GlucoLTrue {
			pCex = make([]int, s.NVars())
			for i := 0; i < s.NVars(); i++ {
				if *s.Model.At(i) == GlucoLTrue {
					pCex[i] = 1
				}
			}
		}
		callbackResult := s.PCnfFunc(s.PCnfMan, message, pCex)
		if callbackResult != 0 {
			panic("GlucoSolver.solve_: callback_result != 0")
		}
	} else if s.PCnfFunc != nil {
		s.Terminate_search_early = false // for next run
	}
	return status
}

// Solve is the overload solve(const vec<Lit>& assumps).
// C: bool solve(const vec<Lit>& assumps)
func (s *GlucoSolver) Solve(assumps *GlucoVec[GlucoLit]) bool {
	s.BudgetOff()
	assumps.CopyTo(&s.assumptions)
	return s.solve_() == GlucoLTrue
}

// SolveLimited is Solver::solveLimited (with resource constraints).
// C: lbool solveLimited(const vec<Lit>& assumps)
func (s *GlucoSolver) SolveLimited(assumps *GlucoVec[GlucoLit]) GlucoLbool {
	assumps.CopyTo(&s.assumptions)
	return s.solve_()
}

// SolveNoAssumps is the overload solve().
// C: bool solve()
func (s *GlucoSolver) SolveNoAssumps() bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	return s.solve_() == GlucoLTrue
}

// Solve1 is the overload solve(Lit p).
// C: bool solve(Lit p)
func (s *GlucoSolver) Solve1(p GlucoLit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	return s.solve_() == GlucoLTrue
}

// Solve2 is the overload solve(Lit p, Lit q).
// C: bool solve(Lit p, Lit q)
func (s *GlucoSolver) Solve2(p GlucoLit, q GlucoLit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	return s.solve_() == GlucoLTrue
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r).
// C: bool solve(Lit p, Lit q, Lit r)
func (s *GlucoSolver) Solve3(p GlucoLit, q GlucoLit, r GlucoLit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	s.assumptions.PushElem(r)
	return s.solve_() == GlucoLTrue
}

// Okay is Solver::okay.
// C: bool okay() const
func (s *GlucoSolver) Okay() bool { return s.ok }

// mapVar_glucose is `static Var mapVar(Var x, vec<Var>& map, Var& max)`.
// C: static Var mapVar(Var x, vec<Var>& map, Var& max)
func mapVar_glucose(x GlucoVar, map_ *GlucoVec[GlucoVar], max *GlucoVar) GlucoVar {
	if map_.Size() <= x || *map_.At(x) == -1 {
		map_.GrowToPad(x+1, -1)
		*map_.At(x) = *max
		*max++
	}
	return *map_.At(x)
}

// ToDimacsClause is the overload toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max).
// C: void toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max)
func (s *GlucoSolver) ToDimacsClause(f *os.File, c *GlucoClause, map_ *GlucoVec[GlucoVar], max *GlucoVar) {
	if s.satisfied(c) {
		return
	}
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) != GlucoLFalse {
			sg := ""
			if GlucoSign(*c.At(i)) {
				sg = "-"
			}
			fmt.Fprintf(f, "%s%d ", sg, mapVar_glucose(GlucoLitVar(*c.At(i)), map_, max)+1)
		}
	}
	fmt.Fprintf(f, "0\n")
}

// ToDimacsFile is the overload toDimacs(const char *file, const vec<Lit>& assumps).
// C: void toDimacs(const char *file, const vec<Lit>& assumps)
func (s *GlucoSolver) ToDimacsFile(file string, assumps *GlucoVec[GlucoLit]) {
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
func (s *GlucoSolver) ToDimacsStream(f *os.File, assumps *GlucoVec[GlucoLit]) {
	// Handle case when solver is in contradictory state:
	if !s.ok {
		fmt.Fprintf(f, "p cnf 1 2\n1 0\n-1 0\n")
		return
	}

	var map_ GlucoVec[GlucoVar]
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
				if s.ValueLit(*c.At(j)) != GlucoLFalse {
					mapVar_glucose(GlucoLitVar(*c.At(j)), &map_, &max)
				}
			}
		}
	}

	// Assumptions are added as unit clauses:
	cnt += s.assumptions.Size()

	fmt.Fprintf(f, "p cnf %d %d\n", max, cnt)

	for i = 0; i < s.assumptions.Size(); i++ {
		if s.ValueLit(*s.assumptions.At(i)) == GlucoLFalse {
			panic("GlucoSolver.ToDimacsStream: assumption is false")
		}
		sg := ""
		if GlucoSign(*s.assumptions.At(i)) {
			sg = "-"
		}
		fmt.Fprintf(f, "%s%d 0\n", sg, mapVar_glucose(GlucoLitVar(*s.assumptions.At(i)), &map_, &max)+1)
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
func (s *GlucoSolver) ToDimacsFileNoAssumps(file string) {
	var as GlucoVec[GlucoLit]
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *GlucoSolver) ToDimacsFile1(file string, p GlucoLit) {
	var as GlucoVec[GlucoLit]
	as.PushElem(p)
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *GlucoSolver) ToDimacsFile2(file string, p GlucoLit, q GlucoLit) {
	var as GlucoVec[GlucoLit]
	as.PushElem(p)
	as.PushElem(q)
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *GlucoSolver) ToDimacsFile3(file string, p GlucoLit, q GlucoLit, r GlucoLit) {
	var as GlucoVec[GlucoLit]
	as.PushElem(p)
	as.PushElem(q)
	as.PushElem(r)
	s.ToDimacsFile(file, &as)
}

// PrintLit is Solver::printLit.
// C: void printLit(Lit l)
func (s *GlucoSolver) PrintLit(l GlucoLit) {
	sg := ""
	if GlucoSign(l) {
		sg = "-"
	}
	c := byte('X')
	if s.ValueLit(l) == GlucoLTrue {
		c = '1'
	} else if s.ValueLit(l) == GlucoLFalse {
		c = '0'
	}
	fmt.Printf("%s%d:%c", sg, GlucoLitVar(l)+1, c)
}

// PrintClause is Solver::printClause.
// C: void printClause(CRef c)
func (s *GlucoSolver) PrintClause(cr GlucoCRef) {
	c := s.ca.At(cr)
	for i := 0; i < c.Size(); i++ {
		s.PrintLit(*c.At(i))
		fmt.Printf(" ")
	}
}

// PrintInitialClause is Solver::printInitialClause.
// C: void printInitialClause(CRef c)
func (s *GlucoSolver) PrintInitialClause(cr GlucoCRef) {
	c := s.ca.At(cr)
	for i := 0; i < c.Size(); i++ {
		if !s.isSelector(GlucoLitVar(*c.At(i))) {
			s.PrintLit(*c.At(i))
			fmt.Printf(" ")
		}
	}
}

// GetCex is Solver::getCex (the base solver has no counter-example: NULL).
// C: int * getCex() const
func (s *GlucoSolver) GetCex() []int { return nil }

// Sat_solver_set_var_fanin_lit is Solver::sat_solver_set_var_fanin_lit (empty in this solver).
// C: void sat_solver_set_var_fanin_lit(int, int, int)
func (s *GlucoSolver) Sat_solver_set_var_fanin_lit(v int, lit0 int, lit1 int) {}

// Sat_solver_start_new_round is Solver::sat_solver_start_new_round (empty in this solver).
// C: void sat_solver_start_new_round()
func (s *GlucoSolver) Sat_solver_start_new_round() {}

// Sat_solver_mark_cone is Solver::sat_solver_mark_cone (empty in this solver).
// C: void sat_solver_mark_cone(int)
func (s *GlucoSolver) Sat_solver_mark_cone(v int) {}

// relocAll is Solver::relocAll (protected).
// C: void relocAll(ClauseAllocator& to)
func (s *GlucoSolver) relocAll(to *GlucoClauseAllocator) {
	var v, sg, i, j int

	// All watchers:
	s.watches.CleanAll()
	s.watchesBin.CleanAll()
	for v = 0; v < s.NVars(); v++ {
		for sg = 0; sg < 2; sg++ {
			p := GlucoMkLit(v, sg != 0)
			ws := s.watches.At(p)
			for j = 0; j < ws.Size(); j++ {
				s.ca.Reloc(&ws.At(j).cref, to)
			}
			ws2 := s.watchesBin.At(p)
			for j = 0; j < ws2.Size(); j++ {
				s.ca.Reloc(&ws2.At(j).cref, to)
			}
		}
	}

	// All reasons:
	for i = 0; i < s.trail.Size(); i++ {
		v := GlucoLitVar(*s.trail.At(i))

		if s.reason(v) != GlucoCRefUndef && (s.ca.At(s.reason(v)).Reloced() || s.locked(s.ca.At(s.reason(v)))) {
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
func (s *GlucoSolver) GarbageCollect() {
	// Initialize the next region to a size corresponding to the estimated utilization degree. This
	// is not precise but should avoid some unnecessary reallocations for the new region:
	to := NewGlucoClauseAllocatorCap(s.ca.Size() - s.ca.Wasted())

	s.relocAll(to)
	if s.Verbosity >= 2 {
		fmt.Printf("|  Garbage collection:   %12d bytes => %12d bytes             |\n",
			s.ca.Size()*GlucoRegionAllocatorUnitSize, to.Size()*GlucoRegionAllocatorUnitSize)
	}
	to.MoveTo(&s.ca)
}

// garbageCollectVirt is the virtual call garbageCollect().
func (s *GlucoSolver) garbageCollectVirt() {
	if s.virtGC != nil {
		s.virtGC()
	} else {
		s.GarbageCollect()
	}
}

// Reset is Solver::reset (virtual): reset everything.
// C: virtual void reset()
func (s *GlucoSolver) Reset() {
	s.ok = true
	s.K = opt_K_glucose.Value()
	s.R = opt_R_glucose.Value()
	s.FirstReduceDB = int(opt_first_reduce_db_glucose.Value())
	s.Var_decay = opt_var_decay_glucose.Value()
	s.Solves, s.Starts, s.Decisions, s.Propagations, s.Conflicts, s.ConflictsRestarts = 0, 0, 0, 0, 0, 0
	s.curRestart = 1
	s.cla_inc, s.var_inc = 1, 1
	s.watches.Clear(false) // We don't free the memory, new calls should be of the same size order.
	s.watchesBin.Clear(false)
	s.qhead = 0
	s.simpDB_assigns = -1
	s.simpDB_props = 0
	s.order_heap.Clear(false)
	s.progress_estimate = 0
	s.conflict_budget = -1
	s.propagation_budget = -1
	s.nbVarsInitialFormula = math.MaxInt32
	s.totalTime4Sat = 0.
	s.totalTime4Unsat = 0.
	s.nbSatCalls, s.nbUnsatCalls = 0, 0
	s.MYFLAG = 0
	s.lbdQueue.Clear(false)
	s.lbdQueue.InitSize(int(s.SizeLBDQueue))
	s.trailQueue.Clear(false)
	s.trailQueue.InitSize(int(s.SizeTrailQueue))
	s.sumLBD = 0
	s.nbclausesbeforereduce = s.FirstReduceDB
	s.clauses.Clear(false)
	s.learnts.Clear(false)
	s.Model.Clear(false)
	s.Conflict.Clear(false)
	s.activity.Clear(false)
	s.assigns.Clear(false)
	s.polarity.Clear(false)
	s.decision.Clear(false)
	s.trail.Clear(false)
	s.nbpos.Clear(false)
	s.trail_lim.Clear(false)
	s.vardata.Clear(false)
	s.assumptions.Clear(false)
	s.permDiff.Clear(false)
	s.lastDecisionLevel.Clear(false)
	s.ca.Clear()
	s.seen.Clear(false)
	s.analyze_stack.Clear(false)
	s.analyze_toclear.Clear(false)
	s.add_tmp.Clear(false)
	s.assumptionPositions.Clear(false)
	s.initialPositions.Clear(false)
}

// CheckGarbageFrac is the overload checkGarbage(double gf).
// C: void checkGarbage(double gf)
func (s *GlucoSolver) CheckGarbageFrac(gf float64) {
	if float64(s.ca.Wasted()) > float64(s.ca.Size())*gf {
		s.garbageCollectVirt()
	}
}

// CheckGarbage is the overload checkGarbage().
// C: void checkGarbage()
func (s *GlucoSolver) CheckGarbage() { s.CheckGarbageFrac(s.Garbage_frac) }

// glucoSolverDrand is Solver::drand (protected static inline).
// C: static inline double drand(double& seed)
func glucoSolverDrand(seed *float64) float64 {
	*seed *= 1389796
	q := int(*seed / 2147483647)
	*seed -= float64(q) * 2147483647
	return *seed / 2147483647
}

// glucoSolverIrand is Solver::irand (protected static inline).
// C: static inline int irand(double& seed, int size)
func glucoSolverIrand(seed *float64, size int) int {
	return int(glucoSolverDrand(seed) * float64(size))
}

// insertVarOrder is Solver::insertVarOrder (protected).
// C: void insertVarOrder(Var x)
func (s *GlucoSolver) insertVarOrder(x GlucoVar) {
	if !s.order_heap.InHeap(x) && *s.decision.At(x) != 0 {
		s.order_heap.Insert(x)
	}
}

// newDecisionLevel is Solver::newDecisionLevel (protected).
// C: void newDecisionLevel()
func (s *GlucoSolver) newDecisionLevel() { s.trail_lim.PushElem(s.trail.Size()) }

// varDecayActivity is Solver::varDecayActivity (protected).
// C: void varDecayActivity()
func (s *GlucoSolver) varDecayActivity() { s.var_inc *= 1 / s.Var_decay }

// varBumpActivityInc is the overload varBumpActivity(Var v, double inc) (protected).
// C: void varBumpActivity(Var v, double inc)
func (s *GlucoSolver) varBumpActivityInc(v GlucoVar, inc float64) {
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
func (s *GlucoSolver) varBumpActivity(v GlucoVar) { s.varBumpActivityInc(v, s.var_inc) }

// claDecayActivity is Solver::claDecayActivity (protected).
// C: void claDecayActivity()
func (s *GlucoSolver) claDecayActivity() { s.cla_inc *= 1 / s.Clause_decay }

// claBumpActivity is Solver::claBumpActivity (protected).
// C: void claBumpActivity(Clause& c)
func (s *GlucoSolver) claBumpActivity(c *GlucoClause) {
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
// implication in the current state (a binary clause may have either literal as implied literal).
// C: bool locked(const Clause& c) const
func (s *GlucoSolver) locked(c *GlucoClause) bool {
	if c.Size() > 2 {
		return s.ValueLit(*c.At(0)) == GlucoLTrue && s.reason(GlucoLitVar(*c.At(0))) != GlucoCRefUndef &&
			s.reason(GlucoLitVar(*c.At(0))) == s.ca.Ael(c)
	}
	return (s.ValueLit(*c.At(0)) == GlucoLTrue && s.reason(GlucoLitVar(*c.At(0))) != GlucoCRefUndef && s.reason(GlucoLitVar(*c.At(0))) == s.ca.Ael(c)) ||
		(s.ValueLit(*c.At(1)) == GlucoLTrue && s.reason(GlucoLitVar(*c.At(1))) != GlucoCRefUndef && s.reason(GlucoLitVar(*c.At(1))) == s.ca.Ael(c))
}

// decisionLevel is Solver::decisionLevel (protected).
// C: int decisionLevel() const
func (s *GlucoSolver) decisionLevel() int { return s.trail_lim.Size() }

// abstractLevel is Solver::abstractLevel (protected): used to represent an
// abstraction of sets of decision levels.
// C: uint32_t abstractLevel(Var x) const
func (s *GlucoSolver) abstractLevel(x GlucoVar) uint32 { return 1 << (uint(s.level(x)) & 31) }

// reason is Solver::reason (protected).
// C: CRef reason(Var x) const
func (s *GlucoSolver) reason(x GlucoVar) GlucoCRef { return s.vardata.At(x).reason }

// level is Solver::level (protected).
// C: int level(Var x) const
func (s *GlucoSolver) level(x GlucoVar) int { return s.vardata.At(x).level }

// Value is the overload value(Var x).
// C: lbool value(Var x) const
func (s *GlucoSolver) Value(x GlucoVar) GlucoLbool { return *s.assigns.At(x) }

// ValueLit is the overload value(Lit p).
// C: lbool value(Lit p) const
func (s *GlucoSolver) ValueLit(p GlucoLit) GlucoLbool {
	return s.assigns.At(GlucoLitVar(p)).Xor(GlucoSign(p))
}

// ModelValue is the overload modelValue(Var x).
// C: lbool modelValue(Var x) const
func (s *GlucoSolver) ModelValue(x GlucoVar) GlucoLbool { return *s.Model.At(x) }

// ModelValueLit is the overload modelValue(Lit p).
// C: lbool modelValue(Lit p) const
func (s *GlucoSolver) ModelValueLit(p GlucoLit) GlucoLbool {
	return s.Model.At(GlucoLitVar(p)).Xor(GlucoSign(p))
}

// NAssigns is Solver::nAssigns.
// C: int nAssigns() const
func (s *GlucoSolver) NAssigns() int { return s.trail.Size() }

// NClauses is Solver::nClauses.
// C: int nClauses() const
func (s *GlucoSolver) NClauses() int { return s.clauses.Size() }

// NLearnts is Solver::nLearnts.
// C: int nLearnts() const
func (s *GlucoSolver) NLearnts() int { return s.learnts.Size() }

// NVars is Solver::nVars.
// C: int nVars() const
func (s *GlucoSolver) NVars() int { return s.vardata.Size() }

// NFreeVars is Solver::nFreeVars.
// C: int nFreeVars() const
func (s *GlucoSolver) NFreeVars() int {
	lim := s.trail.Size()
	if s.trail_lim.Size() != 0 {
		lim = *s.trail_lim.At(0)
	}
	return int(s.Dec_vars) - lim
}

// SetPolarity is Solver::setPolarity.
// C: void setPolarity(Var v, bool b)
func (s *GlucoSolver) SetPolarity(v GlucoVar, b bool) {
	if b {
		*s.polarity.At(v) = 1
	} else {
		*s.polarity.At(v) = 0
	}
}

// SetDecisionVar is Solver::setDecisionVar.
// C: void setDecisionVar(Var v, bool b)
func (s *GlucoSolver) SetDecisionVar(v GlucoVar, b bool) {
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
func (s *GlucoSolver) SetConfBudget(x int64) { s.conflict_budget = s.Conflicts + x }

// SetPropBudget is Solver::setPropBudget.
// C: void setPropBudget(int64_t x)
func (s *GlucoSolver) SetPropBudget(x int64) { s.propagation_budget = s.Propagations + x }

// BudgetOff is Solver::budgetOff.
// C: void budgetOff()
func (s *GlucoSolver) BudgetOff() {
	s.conflict_budget = -1
	s.propagation_budget = -1
}

// Interrupt is Solver::interrupt (may be called asynchronously).
// C: void interrupt()
func (s *GlucoSolver) Interrupt() { s.asynch_interrupt = true }

// ClearInterrupt is Solver::clearInterrupt.
// C: void clearInterrupt()
func (s *GlucoSolver) ClearInterrupt() { s.asynch_interrupt = false }

// withinBudget is Solver::withinBudget (protected).
// C: bool withinBudget() const
func (s *GlucoSolver) withinBudget() bool {
	return !s.asynch_interrupt &&
		(s.conflict_budget < 0 || uint64(s.Conflicts) < uint64(s.conflict_budget)) &&
		(s.propagation_budget < 0 || uint64(s.Propagations) < uint64(s.propagation_budget))
}
