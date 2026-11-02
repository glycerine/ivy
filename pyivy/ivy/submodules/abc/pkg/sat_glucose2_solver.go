package abc

import (
	"fmt"
	"math"
	"os"
)

// Port of src/sat/glucose/Solver.h and Glucose.cpp (namespace Gluco2).
// See sat_bsat2_vec.go for the C++ -> Go conventions. This Solver is the MiniSat
// solver extended with LBD based restarts and clause database reduction, a
// binary-clause watch list, an incremental mode, an external callback and
// certified-UNSAT output.

// Gluco2SolverVarData is Solver::VarData.
type Gluco2SolverVarData struct {
	reason Gluco2CRef
	level  int
}

// gluco2SolverMkVarData is Solver::mkVarData (protected static inline).
// C: static inline VarData mkVarData(CRef cr, int l)
func gluco2SolverMkVarData(cr Gluco2CRef, l int) Gluco2SolverVarData {
	return Gluco2SolverVarData{reason: cr, level: l}
}

// Gluco2SolverWatcher is Solver::Watcher.
type Gluco2SolverWatcher struct {
	cref    Gluco2CRef
	blocker Gluco2Lit
}

// NewGluco2SolverWatcher is Watcher(CRef cr, Lit p).
// C: Watcher(CRef cr, Lit p)
func NewGluco2SolverWatcher(cr Gluco2CRef, p Gluco2Lit) Gluco2SolverWatcher {
	return Gluco2SolverWatcher{cref: cr, blocker: p}
}

// Eq is Watcher::operator==.
// C: bool operator==(const Watcher& w) const
func (w Gluco2SolverWatcher) Eq(o Gluco2SolverWatcher) bool { return w.cref == o.cref }

// Ne is Watcher::operator!=.
// C: bool operator!=(const Watcher& w) const
func (w Gluco2SolverWatcher) Ne(o Gluco2SolverWatcher) bool { return w.cref != o.cref }

// Gluco2SolverWatcherDeleted is Solver::WatcherDeleted.
type Gluco2SolverWatcherDeleted struct {
	ca *Gluco2ClauseAllocator // const ClauseAllocator&
}

// NewGluco2SolverWatcherDeleted is WatcherDeleted(const ClauseAllocator& _ca).
// C: WatcherDeleted(const ClauseAllocator& _ca)
func NewGluco2SolverWatcherDeleted(ca *Gluco2ClauseAllocator) Gluco2SolverWatcherDeleted {
	return Gluco2SolverWatcherDeleted{ca: ca}
}

// Call is WatcherDeleted::operator().
// C: bool operator()(const Watcher& w) const
func (d Gluco2SolverWatcherDeleted) Call(w Gluco2SolverWatcher) bool {
	return d.ca.At(w.cref).Mark() == 1
}

// Gluco2SolverVarOrderLt is Solver::VarOrderLt.
type Gluco2SolverVarOrderLt struct {
	activity *Gluco2Vec[float64] // const vec<double>&
}

// NewGluco2SolverVarOrderLt is VarOrderLt(const vec<double>& act).
// C: VarOrderLt(const vec<double>& act)
func NewGluco2SolverVarOrderLt(act *Gluco2Vec[float64]) Gluco2SolverVarOrderLt {
	return Gluco2SolverVarOrderLt{activity: act}
}

// Call is VarOrderLt::operator().
// C: bool operator()(Var x, Var y) const
func (l Gluco2SolverVarOrderLt) Call(x int, y int) bool {
	return *l.activity.At(x) > *l.activity.At(y)
}

// File-level category strings of Glucose.cpp.
const (
	gluco2Cat       = "CORE"
	gluco2Cr        = "CORE -- RESTART"
	gluco2Cred      = "CORE -- REDUCE"
	gluco2Cm        = "CORE -- MINIMIZE"
	gluco2Certified = "CORE -- CERTIFIED UNSAT"
)

// Command-line options of the solver (static objects in C++: constructed when the
// program starts, registering themselves in Option::getOptionList()). opt_restart_first
// and opt_restart_inc are commented out in the C source.
var (
	opt_incremental_glucose2              = NewGluco2BoolOption(gluco2Cat, "incremental", "Use incremental SAT solving", false)
	opt_K_glucose2                        = NewGluco2DoubleOption(gluco2Cr, "K", "The constant used to force restart", 0.8, NewGluco2DoubleRange(0, false, 1, false))
	opt_R_glucose2                        = NewGluco2DoubleOption(gluco2Cr, "R", "The constant used to block restart", 1.4, NewGluco2DoubleRange(1, false, 5, false))
	opt_size_lbd_queue_glucose2           = NewGluco2IntOption(gluco2Cr, "szLBDQueue", "The size of moving average for LBD (restarts)", 50, NewGluco2IntRange(10, math.MaxInt32))
	opt_size_trail_queue_glucose2         = NewGluco2IntOption(gluco2Cr, "szTrailQueue", "The size of moving average for trail (block restarts)", 5000, NewGluco2IntRange(10, math.MaxInt32))
	opt_first_reduce_db_glucose2          = NewGluco2IntOption(gluco2Cred, "firstReduceDB", "The number of conflicts before the first reduce DB", 2000, NewGluco2IntRange(0, math.MaxInt32))
	opt_inc_reduce_db_glucose2            = NewGluco2IntOption(gluco2Cred, "incReduceDB", "Increment for reduce DB", 300, NewGluco2IntRange(0, math.MaxInt32))
	opt_spec_inc_reduce_db_glucose2       = NewGluco2IntOption(gluco2Cred, "specialIncReduceDB", "Special increment for reduce DB", 1000, NewGluco2IntRange(0, math.MaxInt32))
	opt_lb_lbd_frozen_clause_glucose2     = NewGluco2IntOption(gluco2Cred, "minLBDFrozenClause", "Protect clauses if their LBD decrease and is lower than (for one turn)", 30, NewGluco2IntRange(0, math.MaxInt32))
	opt_lb_size_minimzing_clause_glucose2 = NewGluco2IntOption(gluco2Cm, "minSizeMinimizingClause", "The min size required to minimize clause", 30, NewGluco2IntRange(3, math.MaxInt32))
	opt_lb_lbd_minimzing_clause_glucose2  = NewGluco2IntOption(gluco2Cm, "minLBDMinimizingClause", "The min LBD required to minimize clause", 6, NewGluco2IntRange(3, math.MaxInt32))
	opt_var_decay_glucose2                = NewGluco2DoubleOption(gluco2Cat, "var-decay", "The variable activity decay factor", 0.8, NewGluco2DoubleRange(0, false, 1, false))
	opt_clause_decay_glucose2             = NewGluco2DoubleOption(gluco2Cat, "cla-decay", "The clause activity decay factor", 0.999, NewGluco2DoubleRange(0, false, 1, false))
	opt_random_var_freq_glucose2          = NewGluco2DoubleOption(gluco2Cat, "rnd-freq", "The frequency with which the decision heuristic tries to choose a random variable", 0, NewGluco2DoubleRange(0, true, 1, true))
	opt_random_seed_glucose2              = NewGluco2DoubleOption(gluco2Cat, "rnd-seed", "Used by the random variable selection", 91648253, NewGluco2DoubleRange(0, false, math.Inf(1), false))
	opt_ccmin_mode_glucose2               = NewGluco2IntOption(gluco2Cat, "ccmin-mode", "Controls conflict clause minimization (0=none, 1=basic, 2=deep)", 2, NewGluco2IntRange(0, 2))
	opt_phase_saving_glucose2             = NewGluco2IntOption(gluco2Cat, "phase-saving", "Controls the level of phase saving (0=none, 1=limited, 2=full)", 2, NewGluco2IntRange(0, 2))
	opt_rnd_init_act_glucose2             = NewGluco2BoolOption(gluco2Cat, "rnd-init", "Randomize the initial activity", false)
	opt_garbage_frac_glucose2             = NewGluco2DoubleOption(gluco2Cat, "gc-frac", "The fraction of wasted memory allowed before a garbage collection is triggered", 0.20, NewGluco2DoubleRange(0, false, math.Inf(1), false))

	// opt_certified_ and opt_certified_file_ are not static in C (external linkage).
	opt_certified__glucose2      = NewGluco2BoolOption(gluco2Certified, "certified", "Certified UNSAT using DRUP format", false)
	opt_certified_file__glucose2 = NewGluco2StringOption(gluco2Certified, "certified-output", "Certified UNSAT output file", "NULL")
)

// Gluco2Solver is Gluco2::Solver, the CDCL SAT solver. Public data members are
// exported; protected ones keep the C++ spelling. SimpSolver embeds it.
type Gluco2Solver struct {
	SolverType int // ABC identifies the solver's type as 0

	PCnfMan                any                       // void * pCnfMan: external CNF manager
	PCnfFunc               func(any, int, []int) int // int(*pCnfFunc)(void *p, int, int*): external callback. messages: 0: unsat; 1: sat; -1: still working
	NCallConfl             int                       // callback will be called every this number of conflicts
	Terminate_search_early bool                      // used to stop the solver early if it as instructed by an external caller
	Pstop                  *int                      // another callback
	NRuntimeLimit          int64                     // runtime limit (abctime)
	User_vec               Gluco2Vec[int]
	User_lits              Gluco2Vec[Gluco2Lit]
	Jftr                   int

	// If problem is satisfiable, this vector contains the model (if any).
	Model Gluco2Vec[Gluco2Lbool]
	// If problem is unsatisfiable (possibly under assumptions), this vector represents the final conflict clause expressed in the assumptions.
	Conflict Gluco2Vec[Gluco2Lit]

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
	activity                              Gluco2Vec[float64]
	var_inc                               float64
	watches                               Gluco2OccLists[Gluco2Lit, Gluco2SolverWatcher, Gluco2SolverWatcherDeleted]
	watchesBin                            Gluco2OccLists[Gluco2Lit, Gluco2SolverWatcher, Gluco2SolverWatcherDeleted]
	clauses                               Gluco2Vec[Gluco2CRef]
	learnts                               Gluco2Vec[Gluco2CRef]
	assigns                               Gluco2Vec[Gluco2Lbool]
	polarity                              Gluco2Vec[int8]
	decision                              Gluco2Vec[int8]
	trail                                 Gluco2Vec[Gluco2Lit]
	nbpos                                 Gluco2Vec[int]
	trail_lim                             Gluco2Vec[int]
	vardata                               Gluco2Vec[Gluco2SolverVarData]
	qhead                                 int
	simpDB_assigns                        int
	simpDB_props                          int64
	assumptions                           Gluco2Vec[Gluco2Lit]
	order_heap                            Gluco2Heap[Gluco2SolverVarOrderLt]
	progress_estimate                     float64
	remove_satisfied                      bool
	permDiff                              Gluco2Vec[uint32]    // permDiff[var] contains the current conflict number... used to count the number of LBD
	lastDecisionLevel                     Gluco2Vec[Gluco2Lit] // #ifdef UPDATEVARACTIVITY (defined in Constants.h)
	ca                                    Gluco2ClauseAllocator
	nbclausesbeforereduce                 int                  // to know when it is time to reduce the clause database
	trailQueue                            Gluco2Bqueue[uint32] // bounded queues for restarts
	lbdQueue                              Gluco2Bqueue[uint32]
	sumLBD                                float32 // used to compute the global average of LBD
	sumAssumptions                        int
	seen                                  Gluco2Vec[int8]
	analyze_stack                         Gluco2Vec[Gluco2Lit]
	analyze_toclear                       Gluco2Vec[Gluco2Lit]
	add_tmp                               Gluco2Vec[Gluco2Lit]
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
	assumptionPositions, initialPositions Gluco2Vec[int]

	// Circuit justification (CGLUCOSE_EXP; protected unless noted):
	jdata                                 Gluco2Vec[Gluco2SolverJustData]
	jwatch                                Gluco2Vec[Gluco2SolverJustWatch]
	var2FaninLits                         Gluco2Vec[Gluco2Lit] // (~0): undefine
	var2TravId                            Gluco2Vec[uint32]
	var2Fanout0, var2FanoutN, var2FanoutP Gluco2Vec[Gluco2Lit]
	itpc                                  Gluco2CRef // the interpreted clause of a gate
	travId_prev, travId                   int64
	jheap                                 Gluco2Heap[Gluco2SolverJustOrderLt]
	jstack                                Gluco2Vec[Gluco2Var] // called by unchecked_enqueue, consumed by pickJustLit, cleaned by cancelUntil
	JustModel                             Gluco2Vec[Gluco2Lit] // model obtained by justification enabled (public)

	// virtGC models the virtual call garbageCollect() (and reset): a derived
	// class that overrides it (SimpSolver) sets this to its own method.
	virtGC func()
}

// NewGluco2Solver is Solver::Solver().
// C: Solver::Solver()
func NewGluco2Solver() *Gluco2Solver {
	s := &Gluco2Solver{}
	s.initGluco2Solver()
	return s
}

// initGluco2Solver is the body of the Solver constructor; the SimpSolver constructor runs it for its base class.
func (s *Gluco2Solver) initGluco2Solver() {
	s.SolverType = 0
	s.PCnfFunc = nil
	s.NCallConfl = 1000
	s.Terminate_search_early = false
	s.Pstop = nil
	s.NRuntimeLimit = 0

	s.Verbosity = 0
	s.VerbEveryConflicts = 10000
	s.ShowModel = 0
	s.K = opt_K_glucose2.Value()
	s.R = opt_R_glucose2.Value()
	s.SizeLBDQueue = float64(opt_size_lbd_queue_glucose2.Value())
	s.SizeTrailQueue = float64(opt_size_trail_queue_glucose2.Value())
	s.FirstReduceDB = int(opt_first_reduce_db_glucose2.Value())
	s.IncReduceDB = int(opt_inc_reduce_db_glucose2.Value())
	s.SpecialIncReduceDB = int(opt_spec_inc_reduce_db_glucose2.Value())
	s.LbLBDFrozenClause = uint32(opt_lb_lbd_frozen_clause_glucose2.Value())
	s.LbSizeMinimizingClause = int(opt_lb_size_minimzing_clause_glucose2.Value())
	s.LbLBDMinimizingClause = uint32(opt_lb_lbd_minimzing_clause_glucose2.Value())
	s.Var_decay = opt_var_decay_glucose2.Value()
	s.Clause_decay = opt_clause_decay_glucose2.Value()
	s.Random_var_freq = opt_random_var_freq_glucose2.Value()
	s.Random_seed = opt_random_seed_glucose2.Value()
	s.Ccmin_mode = int(opt_ccmin_mode_glucose2.Value())
	s.Phase_saving = int(opt_phase_saving_glucose2.Value())
	s.Rnd_pol = false
	s.Rnd_init_act = opt_rnd_init_act_glucose2.Value()
	s.Garbage_frac = opt_garbage_frac_glucose2.Value()
	s.CertifiedOutput = nil
	s.CertifiedUNSAT = opt_certified__glucose2.Value()
	s.curRestart = 1

	s.ok = true
	s.cla_inc = 1
	s.var_inc = 1
	s.ca = *NewGluco2ClauseAllocator()
	s.watches = *NewGluco2OccLists[Gluco2Lit, Gluco2SolverWatcher, Gluco2SolverWatcherDeleted](NewGluco2SolverWatcherDeleted(&s.ca))
	s.watchesBin = *NewGluco2OccLists[Gluco2Lit, Gluco2SolverWatcher, Gluco2SolverWatcherDeleted](NewGluco2SolverWatcherDeleted(&s.ca))
	s.qhead = 0
	s.simpDB_assigns = -1
	s.simpDB_props = 0
	s.order_heap = *NewGluco2Heap(NewGluco2SolverVarOrderLt(&s.activity))
	s.progress_estimate = 0
	s.remove_satisfied = true

	// Resource constraints:
	s.conflict_budget = -1
	s.propagation_budget = -1
	s.asynch_interrupt = false
	if opt_incremental_glucose2.Value() {
		s.incremental = 1
	}
	s.nbVarsInitialFormula = math.MaxInt32
	s.jheap = *NewGluco2Heap(NewGluco2SolverJustOrderLt(s))

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
		if opt_certified_file__glucose2.Value() == "NULL" {
			s.CertifiedOutput, err = os.OpenFile("/dev/stdout", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		} else {
			s.CertifiedOutput, err = os.Create(opt_certified_file__glucose2.Value())
		}
		_ = err // fopen returns NULL on failure; the C code does not check
	}

	s.Jftr = 0
	s.travId = 0
	s.travId_prev = 0
	// allocate space for clause interpretation
	var tmp Gluco2Vec[Gluco2Lit]
	tmp.GrowTo(3)
	s.itpc = s.ca.Alloc(&tmp, false)
	s.ca.At(s.itpc).Shrink(s.ca.At(s.itpc).Size())
}

// Destroy is Solver::~Solver() (virtual).
// C: virtual Solver::~Solver()
func (s *Gluco2Solver) Destroy() {}

// SetIncrementalMode is Solver::setIncrementalMode: sets the incremental mode to true.
// C: void setIncrementalMode()
func (s *Gluco2Solver) SetIncrementalMode() { s.incremental = 1 }

// InitNbInitialVars is Solver::initNbInitialVars: number of variables without selectors.
// C: void initNbInitialVars(int nb)
func (s *Gluco2Solver) InitNbInitialVars(nb int) { s.nbVarsInitialFormula = nb }

// NewVar is Solver::newVar (polarity and dvar default to true). Creates a new SAT
// variable; if 'dvar' is cleared, the variable will not be used as a decision variable.
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *Gluco2Solver) NewVar(sign bool, dvar bool) Gluco2Var {
	v := s.NVars()
	s.watches.Init(Gluco2MkLit(v, false))
	s.watches.Init(Gluco2MkLit(v, true))
	s.watchesBin.Init(Gluco2MkLit(v, false))
	s.watchesBin.Init(Gluco2MkLit(v, true))
	s.assigns.PushElem(Gluco2LUndef)
	s.vardata.PushElem(gluco2SolverMkVarData(Gluco2CRefUndef, 0))
	if s.Rnd_init_act {
		s.activity.PushElem(gluco2SolverDrand(&s.Random_seed) * 0.00001)
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
	if s.JustUsage() != 0 {
		s.jdata.PushElem(gluco2SolverMkJustData(false))
		s.jwatch.PushElem(gluco2SolverMkJustWatch())
		s.var2FanoutN.GrowToPad(s.NVars()<<1, gluco2LitNone)
		s.var2Fanout0.GrowToPad(s.NVars(), gluco2LitNone)
		s.var2TravId.GrowToPad(s.NVars(), 0)
	}
	return v
}

// AddVar is Solver::addVar (adds enough variables to make sure there is variable v).
// C: void addVar(Var v)
func (s *Gluco2Solver) AddVar(v Gluco2Var) {
	for v >= s.NVars() {
		s.NewVar(true, true)
	}
}

// gluco2LitLess is LessThan_default<Lit> (Lit::operator<).
type gluco2LitLess struct{}

func (gluco2LitLess) Call(x Gluco2Lit, y Gluco2Lit) bool { return x.Lt(y) }

// gluco2DimacsLit is the DRUP/DIMACS signed literal `(var(l) + 1) * (-2 * sign(l) + 1)`.
func gluco2DimacsLit(l Gluco2Lit) int {
	sg := 0
	if Gluco2Sign(l) {
		sg = 1
	}
	return (Gluco2LitVar(l) + 1) * (-2*sg + 1)
}

// AddClause_ is Solver::addClause_ (no superfluous internal copy; clobbers ps).
// C: bool addClause_(vec<Lit>& ps)
func (s *Gluco2Solver) AddClause_(ps *Gluco2Vec[Gluco2Lit]) bool {
	if s.decisionLevel() != 0 {
		panic("Gluco2Solver.AddClause_: decisionLevel() != 0")
	}
	if !s.ok {
		return false
	}

	// Check if clause is satisfied and remove false/duplicate literals:
	Gluco2SortVec(ps, gluco2LitLess{})

	var oc Gluco2Vec[Gluco2Lit]
	oc.Clear(false)

	var p Gluco2Lit
	var i, j int
	flag := 0
	if s.CertifiedUNSAT {
		for i, j, p = 0, 0, Gluco2LitUndef; i < ps.Size(); i++ {
			oc.PushElem(*ps.At(i))
			if s.ValueLit(*ps.At(i)) == Gluco2LTrue || *ps.At(i) == Gluco2LitNot(p) || s.ValueLit(*ps.At(i)) == Gluco2LFalse {
				flag = 1
			}
		}
	}

	for i, j, p = 0, 0, Gluco2LitUndef; i < ps.Size(); i++ {
		if s.ValueLit(*ps.At(i)) == Gluco2LTrue || *ps.At(i) == Gluco2LitNot(p) {
			return true
		} else if s.ValueLit(*ps.At(i)) != Gluco2LFalse && *ps.At(i) != p {
			p = *ps.At(i)
			*ps.At(j) = p
			j++
		}
	}
	ps.Shrink(i - j)

	if flag != 0 && s.CertifiedUNSAT {
		for i = 0; i < ps.Size(); i++ {
			fmt.Fprintf(s.CertifiedOutput, "%d ", gluco2DimacsLit(*ps.At(i)))
		}
		fmt.Fprintf(s.CertifiedOutput, "0\n")

		fmt.Fprintf(s.CertifiedOutput, "d ")
		for i = 0; i < oc.Size(); i++ {
			fmt.Fprintf(s.CertifiedOutput, "%d ", gluco2DimacsLit(*oc.At(i)))
		}
		fmt.Fprintf(s.CertifiedOutput, "0\n")
	}

	if ps.Size() == 0 {
		s.ok = false
		return false
	} else if ps.Size() == 1 {
		s.uncheckedEnqueue(*ps.At(0), Gluco2CRefUndef)
		s.ok = s.propagate() == Gluco2CRefUndef
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
func (s *Gluco2Solver) AddClause(ps *Gluco2Vec[Gluco2Lit]) bool {
	ps.CopyTo(&s.add_tmp)
	return s.AddClause_(&s.add_tmp)
}

// AddEmptyClause is Solver::addEmptyClause.
// C: bool addEmptyClause()
func (s *Gluco2Solver) AddEmptyClause() bool {
	s.add_tmp.Clear(false)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *Gluco2Solver) AddClauseUnit(p Gluco2Lit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *Gluco2Solver) AddClauseBinary(p Gluco2Lit, q Gluco2Lit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	return s.AddClause_(&s.add_tmp)
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *Gluco2Solver) AddClauseTernary(p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) bool {
	s.add_tmp.Clear(false)
	s.add_tmp.PushElem(p)
	s.add_tmp.PushElem(q)
	s.add_tmp.PushElem(r)
	return s.AddClause_(&s.add_tmp)
}

// attachClause is Solver::attachClause (protected).
// C: void attachClause(CRef cr)
func (s *Gluco2Solver) attachClause(cr Gluco2CRef) {
	c := s.ca.At(cr)
	if c.Size() <= 1 {
		panic("Gluco2Solver.attachClause: size <= 1")
	}
	if c.Size() == 2 {
		s.watchesBin.At(Gluco2LitNot(*c.At(0))).PushElem(NewGluco2SolverWatcher(cr, *c.At(1)))
		s.watchesBin.At(Gluco2LitNot(*c.At(1))).PushElem(NewGluco2SolverWatcher(cr, *c.At(0)))
	} else {
		s.watches.At(Gluco2LitNot(*c.At(0))).PushElem(NewGluco2SolverWatcher(cr, *c.At(1)))
		s.watches.At(Gluco2LitNot(*c.At(1))).PushElem(NewGluco2SolverWatcher(cr, *c.At(0)))
	}
	if c.Learnt() {
		s.Learnts_literals += int64(c.Size())
	} else {
		s.Clauses_literals += int64(c.Size())
	}
}

// detachClause is Solver::detachClause (protected; strict defaults to false).
// C: void detachClause(CRef cr, bool strict = false)
func (s *Gluco2Solver) detachClause(cr Gluco2CRef, strict bool) {
	c := s.ca.At(cr)
	if c.Size() <= 1 {
		panic("Gluco2Solver.detachClause: size <= 1")
	}
	lists := &s.watches
	if c.Size() == 2 {
		lists = &s.watchesBin
	}
	if strict {
		gluco2Remove[Gluco2SolverWatcher](lists.At(Gluco2LitNot(*c.At(0))), NewGluco2SolverWatcher(cr, *c.At(1)), Gluco2SolverWatcher.Ne)
		gluco2Remove[Gluco2SolverWatcher](lists.At(Gluco2LitNot(*c.At(1))), NewGluco2SolverWatcher(cr, *c.At(0)), Gluco2SolverWatcher.Ne)
	} else {
		// Lazy detaching: (NOTE! Must clean all watcher lists before garbage collecting this clause)
		lists.Smudge(Gluco2LitNot(*c.At(0)))
		lists.Smudge(Gluco2LitNot(*c.At(1)))
	}
	if c.Learnt() {
		s.Learnts_literals -= int64(c.Size())
	} else {
		s.Clauses_literals -= int64(c.Size())
	}
}

// removeClause is Solver::removeClause (protected).
// C: void removeClause(CRef cr)
func (s *Gluco2Solver) removeClause(cr Gluco2CRef) {
	c := s.ca.At(cr)

	if s.CertifiedUNSAT {
		fmt.Fprintf(s.CertifiedOutput, "d ")
		for i := 0; i < c.Size(); i++ {
			fmt.Fprintf(s.CertifiedOutput, "%d ", gluco2DimacsLit(*c.At(i)))
		}
		fmt.Fprintf(s.CertifiedOutput, "0\n")
	}

	s.detachClause(cr, false)
	// Don't leave pointers to free'd memory!
	if s.locked(c) {
		s.vardata.At(Gluco2LitVar(*c.At(0))).reason = Gluco2CRefUndef
	}
	c.SetMark(1)
	s.ca.Free_(cr)
}

// satisfied is Solver::satisfied (protected).
// C: bool satisfied(const Clause& c) const
func (s *Gluco2Solver) satisfied(c *Gluco2Clause) bool {
	if s.incremental != 0 { // Check clauses with many selectors is too time consuming
		return s.ValueLit(*c.At(0)) == Gluco2LTrue || s.ValueLit(*c.At(1)) == Gluco2LTrue
	}
	// Default mode.
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) == Gluco2LTrue {
			return true
		}
	}
	return false
}

// computeLBD is the overload computeLBD(const vec<Lit>& lits, int end = -1) (protected).
// C: unsigned int computeLBD(const vec<Lit>& lits, int end = -1)
func (s *Gluco2Solver) computeLBD(lits *Gluco2Vec[Gluco2Lit], end int) uint32 {
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
			if s.isSelector(Gluco2LitVar(*lits.At(i))) {
				continue
			}
			nbDone++
			l := s.level(Gluco2LitVar(*lits.At(i)))
			if *s.permDiff.At(l) != s.MYFLAG {
				*s.permDiff.At(l) = s.MYFLAG
				nblevels++
			}
		}
	} else { // -------- DEFAULT MODE. NOT A LOT OF DIFFERENCES... BUT EASIER TO READ
		for i := 0; i < lits.Size(); i++ {
			l := s.level(Gluco2LitVar(*lits.At(i)))
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
func (s *Gluco2Solver) computeLBDClause(c *Gluco2Clause) uint32 {
	nblevels := 0
	s.MYFLAG++

	if s.incremental != 0 { // ----------------- INCREMENTAL MODE
		nbDone := 0
		for i := 0; i < c.Size(); i++ {
			if uint32(nbDone) >= c.SizeWithoutSelectors() {
				break
			}
			if s.isSelector(Gluco2LitVar(*c.At(i))) {
				continue
			}
			nbDone++
			l := s.level(Gluco2LitVar(*c.At(i)))
			if *s.permDiff.At(l) != s.MYFLAG {
				*s.permDiff.At(l) = s.MYFLAG
				nblevels++
			}
		}
	} else { // -------- DEFAULT MODE. NOT A LOT OF DIFFERENCES... BUT EASIER TO READ
		for i := 0; i < c.Size(); i++ {
			l := s.level(Gluco2LitVar(*c.At(i)))
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
func (s *Gluco2Solver) isSelector(v Gluco2Var) bool {
	return s.incremental != 0 && v > s.nbVarsInitialFormula
}

// minimisationWithBinaryResolution is Solver::minimisationWithBinaryResolution (protected).
// C: void minimisationWithBinaryResolution(vec<Lit>& out_learnt)
func (s *Gluco2Solver) minimisationWithBinaryResolution(outLearnt *Gluco2Vec[Gluco2Lit]) {
	// Find the LBD measure
	lbd := s.computeLBD(outLearnt, -1)
	p := Gluco2LitNot(*outLearnt.At(0))

	if lbd <= s.LbLBDMinimizingClause {
		s.MYFLAG++

		for i := 1; i < outLearnt.Size(); i++ {
			*s.permDiff.At(Gluco2LitVar(*outLearnt.At(i))) = s.MYFLAG
		}

		wbin := s.watchesBin.At(p)
		nb := 0
		for k := 0; k < wbin.Size(); k++ {
			imp := wbin.At(k).blocker
			if *s.permDiff.At(Gluco2LitVar(imp)) == s.MYFLAG && s.ValueLit(imp) == Gluco2LTrue {
				nb++
				*s.permDiff.At(Gluco2LitVar(imp)) = s.MYFLAG - 1
			}
		}
		l := outLearnt.Size() - 1
		if nb > 0 {
			s.NbReducedClauses++
			for i := 1; i < outLearnt.Size()-nb; i++ {
				if *s.permDiff.At(Gluco2LitVar(*outLearnt.At(i))) != s.MYFLAG {
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
func (s *Gluco2Solver) cancelUntil(level int) {
	if s.decisionLevel() > level {
		for c := s.trail.Size() - 1; c >= *s.trail_lim.At(level); c-- {
			x := Gluco2LitVar(*s.trail.At(c))
			*s.assigns.At(x) = Gluco2LUndef
			if s.Phase_saving > 1 || ((s.Phase_saving == 1) && c > *s.trail_lim.Last()) {
				if Gluco2Sign(*s.trail.At(c)) {
					*s.polarity.At(x) = 1
				} else {
					*s.polarity.At(x) = 0
				}
			}
			s.insertVarOrder(x)
		}
		s.qhead = *s.trail_lim.At(level)
		if 0 < s.JustUsage() {
			for c := s.trail.Size() - 1; c >= *s.trail_lim.At(level); c-- {
				x := Gluco2LitVar(*s.trail.At(c))
				s.gateClearJwatch(x, level)
				s.jdata.At(x).now = 0
			}
		}
		s.trail.Shrink_(s.trail.Size() - *s.trail_lim.At(level))
		s.trail_lim.Shrink_(s.trail_lim.Size() - level)
		s.jstack.Shrink_(s.jstack.Size())
	}
}

// pickBranchLit is Solver::pickBranchLit (protected).
// C: Lit pickBranchLit()
func (s *Gluco2Solver) pickBranchLit() Gluco2Lit {
	next := Gluco2VarUndef
	// Random decision:
	if gluco2SolverDrand(&s.Random_seed) < s.Random_var_freq && !s.order_heap.Empty() {
		next = s.order_heap.At(gluco2SolverIrand(&s.Random_seed, s.order_heap.Size()))
		if s.Value(next) == Gluco2LUndef && *s.decision.At(next) != 0 {
			s.Rnd_decisions++
		}
	}
	// Activity based decision:
	for next == Gluco2VarUndef || s.Value(next) != Gluco2LUndef || *s.decision.At(next) == 0 {
		if s.order_heap.Empty() {
			next = Gluco2VarUndef
			break
		}
		next = s.order_heap.RemoveMin()
	}
	if next == Gluco2VarUndef {
		return Gluco2LitUndef
	}
	var sg bool
	if s.Rnd_pol {
		sg = gluco2SolverDrand(&s.Random_seed) < 0.5
	} else {
		sg = *s.polarity.At(next) != 0
	}
	return Gluco2MkLit(next, sg)
}

// analyze is Solver::analyze (protected): analyze conflict and produce a reason clause.
//
//	Pre-conditions: 'out_learnt' is assumed to be cleared; current decision level must be greater than root level.
//	Post-conditions: 'out_learnt[0]' is the asserting literal at level 'out_btlevel'; if out_learnt.size() > 1
//	then 'out_learnt[1]' has the greatest decision level of the rest of literals.
//
// C: void analyze(CRef confl, vec<Lit>& out_learnt, vec<Lit>& selectors, int& out_btlevel, unsigned int& lbd, unsigned int& szWithoutSelectors)
func (s *Gluco2Solver) analyze(confl Gluco2CRef, outLearnt *Gluco2Vec[Gluco2Lit], selectors *Gluco2Vec[Gluco2Lit], outBtlevel *int, lbd *uint32, szWithoutSelectors *uint32) {
	pathC := 0
	p := Gluco2LitUndef

	// Generate conflict clause:
	outLearnt.Push() // (leave room for the asserting literal)
	index := s.trail.Size() - 1

	for {
		if confl == Gluco2CRefUndef { // (otherwise should be UIP)
			panic("Gluco2Solver.analyze: confl == CRef_Undef")
		}
		var c *Gluco2Clause
		if Gluco2LitUndef != p {
			c = s.ca.At(s.castCRef(p))
		} else {
			c = s.ca.At(s.getConfClause(confl))
		}

		// Special case for binary clauses
		// The first one has to be SAT
		if p != Gluco2LitUndef && c.Size() == 2 && s.ValueLit(*c.At(0)) == Gluco2LFalse {
			if s.ValueLit(*c.At(1)) != Gluco2LTrue {
				panic("Gluco2Solver.analyze: binary reason without a true literal")
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
		if p == Gluco2LitUndef {
			j0 = 0
		}
		for j := j0; j < c.Size(); j++ {
			q := *c.At(j)

			if *s.seen.At(Gluco2LitVar(q)) == 0 && s.level(Gluco2LitVar(q)) > 0 {
				if !s.isSelector(Gluco2LitVar(q)) {
					s.varBumpActivity(Gluco2LitVar(q))
				}
				*s.seen.At(Gluco2LitVar(q)) = 1
				if s.level(Gluco2LitVar(q)) >= s.decisionLevel() {
					pathC++
					// UPDATEVARACTIVITY trick (see competition'09 companion paper)
					if !s.isSelector(Gluco2LitVar(q)) && (s.reason(Gluco2LitVar(q)) != Gluco2CRefUndef) && !s.isGateCRef(s.reason(Gluco2LitVar(q))) && s.ca.At(s.reason(Gluco2LitVar(q))).Learnt() {
						s.lastDecisionLevel.PushElem(q)
					}
				} else {
					if s.isSelector(Gluco2LitVar(q)) {
						if s.ValueLit(q) != Gluco2LFalse {
							panic("Gluco2Solver.analyze: selector literal is not false")
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
			v := Gluco2LitVar(*s.trail.At(index))
			index--
			if *s.seen.At(v) != 0 {
				break
			}
		}
		p = *s.trail.At(index + 1)
		confl = s.reason(Gluco2LitVar(p))
		*s.seen.At(Gluco2LitVar(p)) = 0
		pathC--

		if !(pathC > 0) {
			break
		}
	}
	*outLearnt.At(0) = Gluco2LitNot(p)

	// Simplify conflict clause:
	var i, j int
	for i = 0; i < selectors.Size(); i++ {
		outLearnt.PushElem(*selectors.At(i))
	}
	outLearnt.CopyTo_(&s.analyze_toclear)
	if s.Ccmin_mode == 2 {
		var abstractLevel uint32
		for i = 1; i < outLearnt.Size(); i++ {
			abstractLevel |= s.abstractLevel(Gluco2LitVar(*outLearnt.At(i))) // (maintain an abstraction of levels involved in conflict)
		}
		for i, j = 1, 1; i < outLearnt.Size(); i++ {
			if s.reason(Gluco2LitVar(*outLearnt.At(i))) == Gluco2CRefUndef || !s.litRedundant(*outLearnt.At(i), abstractLevel) {
				*outLearnt.At(j) = *outLearnt.At(i)
				j++
			}
		}
	} else if s.Ccmin_mode == 1 {
		for i, j = 1, 1; i < outLearnt.Size(); i++ {
			x := Gluco2LitVar(*outLearnt.At(i))

			if s.reason(x) == Gluco2CRefUndef {
				*outLearnt.At(j) = *outLearnt.At(i)
				j++
			} else {
				c := s.ca.At(s.castCRef(*outLearnt.At(i)))
				// Thanks to Siert Wieringa for this bug fix!
				k0 := 1
				if c.Size() == 2 {
					k0 = 0
				}
				for k := k0; k < c.Size(); k++ {
					if *s.seen.At(Gluco2LitVar(*c.At(k))) == 0 && s.level(Gluco2LitVar(*c.At(k))) > 0 {
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
			if s.level(Gluco2LitVar(*outLearnt.At(i))) > s.level(Gluco2LitVar(*outLearnt.At(maxI))) {
				maxI = i
			}
		}
		// Swap-in this literal at index 1:
		p := *outLearnt.At(maxI)
		*outLearnt.At(maxI) = *outLearnt.At(1)
		*outLearnt.At(1) = p
		*outBtlevel = s.level(Gluco2LitVar(p))
	}

	// Compute the size of the clause without selectors (incremental mode)
	if s.incremental != 0 {
		*szWithoutSelectors = 0
		for i := 0; i < outLearnt.Size(); i++ {
			if !s.isSelector(Gluco2LitVar(*outLearnt.At(i))) {
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
			tl := *s.lastDecisionLevel.At(i)
			if s.isGateCRef(s.reason(Gluco2LitVar(tl))) || s.ca.At(s.reason(Gluco2LitVar(tl))).Lbd() < *lbd {
				s.varBumpActivity(Gluco2LitVar(*s.lastDecisionLevel.At(i)))
			}
		}
		s.lastDecisionLevel.Clear(false)
	}

	for j = 0; j < s.analyze_toclear.Size(); j++ {
		*s.seen.At(Gluco2LitVar(*s.analyze_toclear.At(j))) = 0 // ('seen[]' is now cleared)
	}
	for j = 0; j < selectors.Size(); j++ {
		*s.seen.At(Gluco2LitVar(*selectors.At(j))) = 0
	}
}

// litRedundant is Solver::litRedundant (protected): check if 'p' can be removed.
// 'abstract_levels' is used to abort early if the algorithm is visiting literals
// at levels that cannot be removed later.
// C: bool litRedundant(Lit p, uint32_t abstract_levels)
func (s *Gluco2Solver) litRedundant(p Gluco2Lit, abstractLevels uint32) bool {
	s.analyze_stack.Clear(false)
	s.analyze_stack.PushElem(p)
	top := s.analyze_toclear.Size()
	for s.analyze_stack.Size() > 0 {
		if s.reason(Gluco2LitVar(*s.analyze_stack.Last())) == Gluco2CRefUndef {
			panic("Gluco2Solver.litRedundant: reason == CRef_Undef")
		}
		c := s.ca.At(s.castCRef(*s.analyze_stack.Last()))
		s.analyze_stack.Pop()
		if c.Size() == 2 && s.ValueLit(*c.At(0)) == Gluco2LFalse {
			if s.ValueLit(*c.At(1)) != Gluco2LTrue {
				panic("Gluco2Solver.litRedundant: binary reason without a true literal")
			}
			tmp := *c.At(0)
			*c.At(0) = *c.At(1)
			*c.At(1) = tmp
		}

		for i := 1; i < c.Size(); i++ {
			p := *c.At(i)
			if *s.seen.At(Gluco2LitVar(p)) == 0 && s.level(Gluco2LitVar(p)) > 0 {
				if s.reason(Gluco2LitVar(p)) != Gluco2CRefUndef && (s.abstractLevel(Gluco2LitVar(p))&abstractLevels) != 0 {
					*s.seen.At(Gluco2LitVar(p)) = 1
					s.analyze_stack.PushElem(p)
					s.analyze_toclear.PushElem(p)
				} else {
					for j := top; j < s.analyze_toclear.Size(); j++ {
						*s.seen.At(Gluco2LitVar(*s.analyze_toclear.At(j))) = 0
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
func (s *Gluco2Solver) analyzeFinal(p Gluco2Lit, outConflict *Gluco2Vec[Gluco2Lit]) {
	outConflict.Clear(false)
	outConflict.PushElem(p)

	if s.decisionLevel() == 0 {
		return
	}

	*s.seen.At(Gluco2LitVar(p)) = 1

	for i := s.trail.Size() - 1; i >= *s.trail_lim.At(0); i-- {
		x := Gluco2LitVar(*s.trail.At(i))
		if *s.seen.At(x) != 0 {
			if s.reason(x) == Gluco2CRefUndef {
				if s.level(x) <= 0 {
					panic("Gluco2Solver.analyzeFinal: level(x) <= 0")
				}
				outConflict.PushElem(Gluco2LitNot(*s.trail.At(i)))
			} else {
				c := s.ca.At(s.castCRef(*s.trail.At(i)))
				// Bug in case of assumptions due to special data structures for Binary.
				// Many thanks to Sam Bayless (sbayless@cs.ubc.ca) for discover this bug.
				j0 := 1
				if c.Size() == 2 {
					j0 = 0
				}
				for j := j0; j < c.Size(); j++ {
					if s.level(Gluco2LitVar(*c.At(j))) > 0 {
						*s.seen.At(Gluco2LitVar(*c.At(j))) = 1
					}
				}
			}
			*s.seen.At(x) = 0
		}
	}

	*s.seen.At(Gluco2LitVar(p)) = 0
}

// uncheckedEnqueue is Solver::uncheckedEnqueue (protected; from defaults to CRef_Undef).
// C: void uncheckedEnqueue(Lit p, CRef from = CRef_Undef)
func (s *Gluco2Solver) uncheckedEnqueue(p Gluco2Lit, from Gluco2CRef) {
	if s.ValueLit(p) != Gluco2LUndef {
		panic("Gluco2Solver.uncheckedEnqueue: literal already assigned")
	}
	*s.assigns.At(Gluco2LitVar(p)) = NewGluco2LboolB(!Gluco2Sign(p))
	*s.vardata.At(Gluco2LitVar(p)) = gluco2SolverMkVarData(from, s.decisionLevel())
	s.trail.Push_(p)
	if 0 < s.JustUsage() {
		*s.jdata.At(Gluco2LitVar(p)) = gluco2SolverMkJustData(s.isJReason(Gluco2LitVar(p)))
		if s.isJReason(Gluco2LitVar(p)) && Gluco2LUndef == s.Value(s.getFaninVar0(Gluco2LitVar(p))) && Gluco2LUndef == s.Value(s.getFaninVar1(Gluco2LitVar(p))) {
			s.jstack.PushElem(Gluco2LitVar(p))
		}
	}
}

// enqueue is Solver::enqueue (protected; from defaults to CRef_Undef): test if
// fact 'p' contradicts current state, enqueue otherwise.
// C: bool enqueue(Lit p, CRef from = CRef_Undef)
func (s *Gluco2Solver) enqueue(p Gluco2Lit, from Gluco2CRef) bool {
	if s.ValueLit(p) != Gluco2LUndef {
		return s.ValueLit(p) != Gluco2LFalse
	}
	s.uncheckedEnqueue(p, from)
	return true
}

// propagate is Solver::propagate (protected): propagates all enqueued facts. If a
// conflict arises, the conflicting clause is returned, otherwise CRef_Undef.
// Post-condition: the propagation queue is empty, even if there was a conflict.
// C: CRef propagate()
func (s *Gluco2Solver) propagate() Gluco2CRef {
	confl := Gluco2CRefUndef
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
		if 2 <= s.JustUsage() {
			if stats := s.gatePropagate(p); Gluco2CRefUndef != stats {
				confl = stats
				if Gluco2LTrue == s.Value(Gluco2LitVar(p)) {
					return confl
				}
			}
		}

		// First, Propagate binary clauses
		wbin := s.watchesBin.At(p)
		for k := 0; k < wbin.Size(); k++ {
			imp := wbin.At(k).blocker
			if s.ValueLit(imp) == Gluco2LFalse {
				return wbin.At(k).cref
			}
			if s.ValueLit(imp) == Gluco2LUndef {
				s.uncheckedEnqueue(imp, wbin.At(k).cref)
			}
		}

	NextWatcher:
		for i, j = 0, 0; i != end; {
			// Try to avoid inspecting the clause:
			blocker := wd[i].blocker
			if s.ValueLit(blocker) == Gluco2LTrue {
				wd[j] = wd[i]
				j++
				i++
				continue
			}

			// Make sure the false literal is data[1]:
			cr := wd[i].cref
			c := s.ca.At(cr)
			falseLit := Gluco2LitNot(p)
			if *c.At(0) == falseLit {
				*c.At(0) = *c.At(1)
				*c.At(1) = falseLit
			}
			if *c.At(1) != falseLit {
				panic("Gluco2Solver.propagate: c[1] != false_lit")
			}
			i++

			// If 0th watch is true, then clause is already satisfied.
			first := *c.At(0)
			w := NewGluco2SolverWatcher(cr, first)
			if first != blocker && s.ValueLit(first) == Gluco2LTrue {
				wd[j] = w
				j++
				continue
			}

			// Look for new watch:
			if s.incremental != 0 { // ----------------- INCREMENTAL MODE
				choosenPos := -1
				for k := 2; k < c.Size(); k++ {
					if s.ValueLit(*c.At(k)) != Gluco2LFalse {
						if s.decisionLevel() > s.assumptions.Size() {
							choosenPos = k
							break
						}
						choosenPos = k
						if s.ValueLit(*c.At(k)) == Gluco2LTrue || !s.isSelector(Gluco2LitVar(*c.At(k))) {
							break
						}
					}
				}
				if choosenPos != -1 {
					*c.At(1) = *c.At(choosenPos)
					*c.At(choosenPos) = falseLit
					s.watches.At(Gluco2LitNot(*c.At(1))).PushElem(w)
					continue NextWatcher
				}
			} else { // ----------------- DEFAULT  MODE (NOT INCREMENTAL)
				for k := 2; k < c.Size(); k++ {
					if s.ValueLit(*c.At(k)) != Gluco2LFalse {
						*c.At(1) = *c.At(k)
						*c.At(k) = falseLit
						s.watches.At(Gluco2LitNot(*c.At(1))).PushElem(w)
						continue NextWatcher
					}
				}
			}

			// Did not find watch -- clause is unit under assignment:
			wd[j] = w
			j++
			if s.ValueLit(first) == Gluco2LFalse {
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

// gluco2ReduceDBLt is the local struct reduceDB_lt of Glucose.cpp.
type gluco2ReduceDBLt struct {
	ca *Gluco2ClauseAllocator
}

// Call is reduceDB_lt::operator().
// C: bool operator()(CRef x, CRef y)
func (l gluco2ReduceDBLt) Call(x Gluco2CRef, y Gluco2CRef) bool {
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
func (s *Gluco2Solver) reduceDB() {
	var i, j int
	s.NbReduceDB++
	Gluco2SortVec(&s.learnts, gluco2ReduceDBLt{ca: &s.ca})

	// We have a lot of "good" clauses, it is difficult to compare them. Keep more !
	if s.ca.At(*s.learnts.At(s.learnts.Size() / gluco2RatioRemoveClauses)).Lbd() <= 3 {
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

// gluco2RatioRemoveClauses is the macro RATIOREMOVECLAUSES of Constants.h.
const gluco2RatioRemoveClauses = GlucoseRemoveClauseRatio

// removeSatisfied is Solver::removeSatisfied (protected).
// C: void removeSatisfied(vec<CRef>& cs)
func (s *Gluco2Solver) removeSatisfied(cs *Gluco2Vec[Gluco2CRef]) {
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
func (s *Gluco2Solver) rebuildOrderHeap() {
	var vs Gluco2Vec[Gluco2Var]
	for v := 0; v < s.NVars(); v++ {
		if *s.decision.At(v) != 0 && s.Value(v) == Gluco2LUndef {
			vs.PushElem(v)
		}
	}
	s.order_heap.Build(&vs)
}

// Simplify is Solver::simplify: simplify the clause database according to the
// current top-level assigment (currently only removes satisfied clauses).
// C: bool simplify()
func (s *Gluco2Solver) Simplify() bool {
	if s.decisionLevel() != 0 {
		panic("Gluco2Solver.Simplify: decisionLevel() != 0")
	}
	if !s.ok || s.propagate() != Gluco2CRefUndef {
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
func (s *Gluco2Solver) search(nofConflicts int) Gluco2Lbool {
	if !s.ok {
		panic("Gluco2Solver.search: !ok")
	}
	var backtrackLevel int
	conflictC := 0
	var learntClause, selectors Gluco2Vec[Gluco2Lit]
	var nblevels, szWoutSelectors uint32
	blocked := false
	s.Starts++

	for {
		confl := s.propagate()
		// exact conflict limit
		if !s.withinBudget() && confl != Gluco2CRefUndef {
			s.lbdQueue.Fastclear()
			s.cancelUntil(0)
			return Gluco2LUndef
		}
		if confl != Gluco2CRefUndef {
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
				fmt.Printf("c | %8d   %7d    %5d | %7d %8d %8d | %5d %8d   %6d %8d | %6.3f %% \n",
					int(s.Starts), int(s.Nbstopsrestarts), int(s.Conflicts/s.Starts),
					int(s.Dec_vars)-lim, s.NClauses(), int(s.Clauses_literals),
					int(s.NbReduceDB), s.NLearnts(), int(s.NbDL2), int(s.NbRemovedClauses), s.progressEstimate()*100)
			}
			if s.decisionLevel() == 0 {
				return Gluco2LFalse
			}

			s.trailQueue.Push(uint32(s.trail.Size()))
			// BLOCK RESTART (CP 2012 paper)
			if s.ConflictsRestarts > gluco2LowerBoundForBlockingRestart && s.lbdQueue.Isvalid() != 0 && float64(s.trail.Size()) > s.R*float64(s.trailQueue.Getavg()) {
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
					fmt.Fprintf(s.CertifiedOutput, "%d ", gluco2DimacsLit(*learntClause.At(i)))
				}
				fmt.Fprintf(s.CertifiedOutput, "0\n")
			}

			if learntClause.Size() == 1 {
				s.uncheckedEnqueue(*learntClause.At(0), Gluco2CRefUndef)
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
				return Gluco2LUndef
			}

			// Simplify the set of problem clauses:
			if s.decisionLevel() == 0 && !s.Simplify() {
				return Gluco2LFalse
			}
			// Perform clause database reduction !
			if s.Conflicts >= s.curRestart*int64(s.nbclausesbeforereduce) {
				if s.learnts.Size() <= 0 {
					panic("Gluco2Solver.search: no learnt clauses to reduce")
				}
				s.curRestart = (s.Conflicts / int64(s.nbclausesbeforereduce)) + 1
				s.reduceDB()
				s.nbclausesbeforereduce += s.IncReduceDB
			}

			next := Gluco2LitUndef
			for s.decisionLevel() < s.assumptions.Size() {
				// Perform user provided assumption:
				p := *s.assumptions.At(s.decisionLevel())
				if s.ValueLit(p) == Gluco2LTrue {
					// Dummy decision level:
					s.newDecisionLevel()
				} else if s.ValueLit(p) == Gluco2LFalse {
					s.analyzeFinal(Gluco2LitNot(p), &s.Conflict)
					return Gluco2LFalse
				} else {
					next = p
					break
				}
			}

			// pick from JustQueue
			jReason := -1
			if 0 < s.JustUsage() {
				if next == Gluco2LitUndef {
					s.Decisions++
					next = s.pickJustLit(&jReason)
					if next == Gluco2LitUndef {
						return Gluco2LTrue
					}
					s.addJwatch(Gluco2LitVar(next), jReason)
				}
			}

			if next == Gluco2LitUndef {
				// New variable decision:
				s.Decisions++
				next = s.pickBranchLit()
				if next == Gluco2LitUndef {
					// Model found:
					return Gluco2LTrue
				}
			}

			// Increase decision level and enqueue 'next'
			s.newDecisionLevel()
			s.uncheckedEnqueue(next, Gluco2CRefUndef)
		}
	}
}

// gluco2LowerBoundForBlockingRestart is the macro LOWER_BOUND_FOR_BLOCKING_RESTART of Constants.h.
const gluco2LowerBoundForBlockingRestart = GlucoseBlockingRestartLower

// progressEstimate is Solver::progressEstimate (protected).
// C: double progressEstimate() const
func (s *Gluco2Solver) progressEstimate() float64 {
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
func (s *Gluco2Solver) PrintIncrementalStats() {
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
func (s *Gluco2Solver) solve_() Gluco2Lbool {
	s.resetJustData(false)
	if 0 < s.JustUsage() {
		for i := 0; i < s.trail.Size(); i++ { // learnt unit clauses
			v := Gluco2LitVar(*s.trail.At(i))
			if s.isJReason(v) {
				s.jstack.PushElem(v)
			}
		}
	}
	if s.incremental != 0 && s.CertifiedUNSAT {
		fmt.Printf("Can not use incremental and certified unsat in the same time\n")
		os.Exit(255) // exit(-1)
	}

	s.JustModel.Shrink_(s.JustModel.Size())
	s.Model.Shrink_(s.Model.Size())
	s.Conflict.Shrink_(s.Conflict.Size())
	if !s.ok {
		s.travId_prev = s.travId
		return Gluco2LFalse
	}

	curTime := Gluco2CpuTime()

	s.Solves++

	status := Gluco2LUndef
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
		fmt.Printf("c |       NB   Blocked  Avg Cfc |    Vars  Clauses Literals |   Red   Learnts    LBD2  Removed |          | pol-inconsist\n")
		fmt.Printf("c =========================================================================================================\n")
	}

	// Search:
	currRestarts := 0
	for status == Gluco2LUndef {
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
		if status == Gluco2LFalse {
			fmt.Fprintf(s.CertifiedOutput, "0\n")
		}
		s.CertifiedOutput.Close()
	}

	if status == Gluco2LTrue {
		if s.JustUsage() != 0 {
			if !s.jheap.Empty() {
				panic("Gluco2Solver.solve_: jheap is not empty")
			}
			i, j := 0, 0
			s.JustModel.PushElem(Gluco2ToLit(0))
			for ; i < s.trail.Size(); i++ {
				if s.IsRoundWatch(Gluco2LitVar(*s.trail.At(i))) && !s.isTwoFanin(Gluco2LitVar(*s.trail.At(i))) {
					s.JustModel.PushElem(*s.trail.At(i))
					j++
				}
			}
			*s.JustModel.At(0) = Gluco2ToLit(j)
		} else {
			// Extend & copy model:
			s.Model.GrowTo(s.NVars())
			for i := 0; i < s.trail.Size(); i++ {
				*s.Model.At(Gluco2LitVar(*s.trail.At(i))) = s.Value(Gluco2LitVar(*s.trail.At(i)))
			}
		}
	} else if status == Gluco2LFalse && s.Conflict.Size() == 0 {
		s.ok = false
	}

	s.cancelUntil(0)

	finalTime := Gluco2CpuTime()
	if status == Gluco2LTrue {
		s.nbSatCalls++
		s.totalTime4Sat += finalTime - curTime
	}
	if status == Gluco2LFalse {
		s.nbUnsatCalls++
		s.totalTime4Unsat += finalTime - curTime
	}

	// ABC callback
	if s.PCnfFunc != nil && !s.Terminate_search_early { // hack to avoid calling callback twise if the solver was terminated early
		var pCex []int
		message := -1
		if status == Gluco2LTrue {
			message = 1
		} else if status == Gluco2LFalse {
			message = 0
		}
		if status == Gluco2LTrue {
			pCex = make([]int, s.NVars())
			for i := 0; i < s.NVars(); i++ {
				if *s.Model.At(i) == Gluco2LTrue {
					pCex[i] = 1
				}
			}
		}
		callbackResult := s.PCnfFunc(s.PCnfMan, message, pCex)
		if callbackResult != 0 {
			panic("Gluco2Solver.solve_: callback_result != 0")
		}
	} else if s.PCnfFunc != nil {
		s.Terminate_search_early = false // for next run
	}
	s.travId_prev = s.travId
	return status
}

// Solve is the overload solve(const vec<Lit>& assumps).
// C: bool solve(const vec<Lit>& assumps)
func (s *Gluco2Solver) Solve(assumps *Gluco2Vec[Gluco2Lit]) bool {
	s.BudgetOff()
	assumps.CopyTo(&s.assumptions)
	return s.solve_() == Gluco2LTrue
}

// SolveLimited is Solver::solveLimited (with resource constraints).
// C: lbool solveLimited(const vec<Lit>& assumps)
func (s *Gluco2Solver) SolveLimited(assumps *Gluco2Vec[Gluco2Lit]) Gluco2Lbool {
	assumps.CopyTo(&s.assumptions)
	return s.solve_()
}

// SolveNoAssumps is the overload solve().
// C: bool solve()
func (s *Gluco2Solver) SolveNoAssumps() bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	return s.solve_() == Gluco2LTrue
}

// Solve1 is the overload solve(Lit p).
// C: bool solve(Lit p)
func (s *Gluco2Solver) Solve1(p Gluco2Lit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	return s.solve_() == Gluco2LTrue
}

// Solve2 is the overload solve(Lit p, Lit q).
// C: bool solve(Lit p, Lit q)
func (s *Gluco2Solver) Solve2(p Gluco2Lit, q Gluco2Lit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	return s.solve_() == Gluco2LTrue
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r).
// C: bool solve(Lit p, Lit q, Lit r)
func (s *Gluco2Solver) Solve3(p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) bool {
	s.BudgetOff()
	s.assumptions.Clear(false)
	s.assumptions.PushElem(p)
	s.assumptions.PushElem(q)
	s.assumptions.PushElem(r)
	return s.solve_() == Gluco2LTrue
}

// Okay is Solver::okay.
// C: bool okay() const
func (s *Gluco2Solver) Okay() bool { return s.ok }

// mapVar_glucose2 is `static Var mapVar(Var x, vec<Var>& map, Var& max)`.
// C: static Var mapVar(Var x, vec<Var>& map, Var& max)
func mapVar_glucose2(x Gluco2Var, map_ *Gluco2Vec[Gluco2Var], max *Gluco2Var) Gluco2Var {
	if map_.Size() <= x || *map_.At(x) == -1 {
		map_.GrowToPad(x+1, -1)
		*map_.At(x) = *max
		*max++
	}
	return *map_.At(x)
}

// ToDimacsClause is the overload toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max).
// C: void toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max)
func (s *Gluco2Solver) ToDimacsClause(f *os.File, c *Gluco2Clause, map_ *Gluco2Vec[Gluco2Var], max *Gluco2Var) {
	if s.satisfied(c) {
		return
	}
	for i := 0; i < c.Size(); i++ {
		if s.ValueLit(*c.At(i)) != Gluco2LFalse {
			sg := ""
			if Gluco2Sign(*c.At(i)) {
				sg = "-"
			}
			fmt.Fprintf(f, "%s%d ", sg, mapVar_glucose2(Gluco2LitVar(*c.At(i)), map_, max)+1)
		}
	}
	fmt.Fprintf(f, "0\n")
}

// ToDimacsFile is the overload toDimacs(const char *file, const vec<Lit>& assumps).
// C: void toDimacs(const char *file, const vec<Lit>& assumps)
func (s *Gluco2Solver) ToDimacsFile(file string, assumps *Gluco2Vec[Gluco2Lit]) {
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
func (s *Gluco2Solver) ToDimacsStream(f *os.File, assumps *Gluco2Vec[Gluco2Lit]) {
	// Handle case when solver is in contradictory state:
	if !s.ok {
		fmt.Fprintf(f, "p cnf 1 2\n1 0\n-1 0\n")
		return
	}

	var map_ Gluco2Vec[Gluco2Var]
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
				if s.ValueLit(*c.At(j)) != Gluco2LFalse {
					mapVar_glucose2(Gluco2LitVar(*c.At(j)), &map_, &max)
				}
			}
		}
	}

	// Assumptions are added as unit clauses:
	cnt += s.assumptions.Size()

	fmt.Fprintf(f, "p cnf %d %d\n", max, cnt)

	for i = 0; i < s.assumptions.Size(); i++ {
		if s.ValueLit(*s.assumptions.At(i)) == Gluco2LFalse {
			panic("Gluco2Solver.ToDimacsStream: assumption is false")
		}
		sg := ""
		if Gluco2Sign(*s.assumptions.At(i)) {
			sg = "-"
		}
		fmt.Fprintf(f, "%s%d 0\n", sg, mapVar_glucose2(Gluco2LitVar(*s.assumptions.At(i)), &map_, &max)+1)
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
func (s *Gluco2Solver) ToDimacsFileNoAssumps(file string) {
	var as Gluco2Vec[Gluco2Lit]
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *Gluco2Solver) ToDimacsFile1(file string, p Gluco2Lit) {
	var as Gluco2Vec[Gluco2Lit]
	as.PushElem(p)
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *Gluco2Solver) ToDimacsFile2(file string, p Gluco2Lit, q Gluco2Lit) {
	var as Gluco2Vec[Gluco2Lit]
	as.PushElem(p)
	as.PushElem(q)
	s.ToDimacsFile(file, &as)
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *Gluco2Solver) ToDimacsFile3(file string, p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) {
	var as Gluco2Vec[Gluco2Lit]
	as.PushElem(p)
	as.PushElem(q)
	as.PushElem(r)
	s.ToDimacsFile(file, &as)
}

// PrintLit is Solver::printLit.
// C: void printLit(Lit l)
func (s *Gluco2Solver) PrintLit(l Gluco2Lit) {
	sg := ""
	if Gluco2Sign(l) {
		sg = "-"
	}
	c := byte('X')
	if s.ValueLit(l) == Gluco2LTrue {
		c = '1'
	} else if s.ValueLit(l) == Gluco2LFalse {
		c = '0'
	}
	fmt.Printf("%s%d:%c", sg, Gluco2LitVar(l)+1, c)
}

// PrintClause is Solver::printClause.
// C: void printClause(CRef c)
func (s *Gluco2Solver) PrintClause(cr Gluco2CRef) {
	c := s.ca.At(cr)
	for i := 0; i < c.Size(); i++ {
		s.PrintLit(*c.At(i))
		fmt.Printf(" ")
	}
}

// PrintInitialClause is Solver::printInitialClause.
// C: void printInitialClause(CRef c)
func (s *Gluco2Solver) PrintInitialClause(cr Gluco2CRef) {
	c := s.ca.At(cr)
	for i := 0; i < c.Size(); i++ {
		if !s.isSelector(Gluco2LitVar(*c.At(i))) {
			s.PrintLit(*c.At(i))
			fmt.Printf(" ")
		}
	}
}

// relocAll is Solver::relocAll (protected).
// C: void relocAll(ClauseAllocator& to)
func (s *Gluco2Solver) relocAll(to *Gluco2ClauseAllocator) {
	if Gluco2CRefUndef != s.itpc {
		s.setItpcSize(3)
		s.ca.Reloc(&s.itpc, to)
	}
	var v, sg, i, j int

	// All watchers:
	s.watches.CleanAll()
	s.watchesBin.CleanAll()
	for v = 0; v < s.NVars(); v++ {
		for sg = 0; sg < 2; sg++ {
			p := Gluco2MkLit(v, sg != 0)
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
		v := Gluco2LitVar(*s.trail.At(i))

		if s.isGateCRef(s.reason(v)) {
			continue
		}
		if s.reason(v) != Gluco2CRefUndef && (s.ca.At(s.reason(v)).Reloced() || s.locked(s.ca.At(s.reason(v)))) {
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
func (s *Gluco2Solver) GarbageCollect() {
	// Initialize the next region to a size corresponding to the estimated utilization degree. This
	// is not precise but should avoid some unnecessary reallocations for the new region:
	to := NewGluco2ClauseAllocatorCap(s.ca.Size() - s.ca.Wasted())

	s.relocAll(to)
	if s.Verbosity >= 2 {
		fmt.Printf("|  Garbage collection:   %12d bytes => %12d bytes             |\n",
			s.ca.Size()*Gluco2RegionAllocatorUnitSize, to.Size()*Gluco2RegionAllocatorUnitSize)
	}
	to.MoveTo(&s.ca)
}

// garbageCollectVirt is the virtual call garbageCollect().
func (s *Gluco2Solver) garbageCollectVirt() {
	if s.virtGC != nil {
		s.virtGC()
	} else {
		s.GarbageCollect()
	}
}

// Reset is Solver::reset (virtual): reset everything.
// C: virtual void reset()
func (s *Gluco2Solver) Reset() {
	s.ok = true
	s.K = opt_K_glucose2.Value()
	s.R = opt_R_glucose2.Value()
	s.FirstReduceDB = int(opt_first_reduce_db_glucose2.Value())
	s.Var_decay = opt_var_decay_glucose2.Value()
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
	s.Model.Shrink_(s.Model.Size())
	s.Conflict.Shrink_(s.Conflict.Size())
	s.activity.Shrink_(s.activity.Size())
	s.assigns.Shrink_(s.assigns.Size())
	s.polarity.Shrink_(s.polarity.Size())
	s.decision.Shrink_(s.decision.Size())
	s.trail.Shrink_(s.trail.Size())
	s.trail_lim.Shrink_(s.trail_lim.Size())
	s.vardata.Shrink_(s.vardata.Size())
	s.assumptions.Shrink_(s.assumptions.Size())
	s.nbpos.Shrink_(s.nbpos.Size())
	s.permDiff.Shrink_(s.permDiff.Size())
	s.lastDecisionLevel.Shrink_(s.lastDecisionLevel.Size())
	s.ca.Clear()
	s.seen.Shrink_(s.seen.Size())
	s.analyze_stack.Shrink_(s.analyze_stack.Size())
	s.analyze_toclear.Shrink_(s.analyze_toclear.Size())
	s.add_tmp.Clear(false)
	s.assumptionPositions.Clear(false)
	s.initialPositions.Clear(false)
	s.resetJustData(false)
	s.jwatch.Shrink_(s.jwatch.Size())
	s.jdata.Shrink_(s.jdata.Size())
	s.travId = 0
	s.travId_prev = 0
	s.var2TravId.Shrink_(s.var2TravId.Size())
	s.JustModel.Shrink_(s.JustModel.Size())
	s.var2FaninLits.Shrink_(s.var2FaninLits.Size())
	s.var2Fanout0.Shrink_(s.var2Fanout0.Size())
	s.var2FanoutN.Shrink_(s.var2FanoutN.Size())
	if Gluco2CRefUndef != s.itpc {
		s.itpc = Gluco2CRefUndef // clause allocator has been cleared, do not worry
		// allocate space for clause interpretation
		var tmp Gluco2Vec[Gluco2Lit]
		tmp.GrowTo(3)
		s.itpc = s.ca.Alloc(&tmp, false)
		s.ca.At(s.itpc).Shrink(s.ca.At(s.itpc).Size())
	}
}

// CheckGarbageFrac is the overload checkGarbage(double gf).
// C: void checkGarbage(double gf)
func (s *Gluco2Solver) CheckGarbageFrac(gf float64) {
	if float64(s.ca.Wasted()) > float64(s.ca.Size())*gf {
		s.garbageCollectVirt()
	}
}

// CheckGarbage is the overload checkGarbage().
// C: void checkGarbage()
func (s *Gluco2Solver) CheckGarbage() { s.CheckGarbageFrac(s.Garbage_frac) }

// gluco2SolverDrand is Solver::drand (protected static inline).
// C: static inline double drand(double& seed)
func gluco2SolverDrand(seed *float64) float64 {
	*seed *= 1389796
	q := int(*seed / 2147483647)
	*seed -= float64(q) * 2147483647
	return *seed / 2147483647
}

// gluco2SolverIrand is Solver::irand (protected static inline).
// C: static inline int irand(double& seed, int size)
func gluco2SolverIrand(seed *float64, size int) int {
	return int(gluco2SolverDrand(seed) * float64(size))
}

// insertVarOrder is Solver::insertVarOrder (protected).
// C: void insertVarOrder(Var x)
func (s *Gluco2Solver) insertVarOrder(x Gluco2Var) {
	if !s.order_heap.InHeap(x) && *s.decision.At(x) != 0 {
		s.order_heap.Insert(x)
	}
}

// newDecisionLevel is Solver::newDecisionLevel (protected).
// C: void newDecisionLevel()
func (s *Gluco2Solver) newDecisionLevel() { s.trail_lim.PushElem(s.trail.Size()) }

// varDecayActivity is Solver::varDecayActivity (protected).
// C: void varDecayActivity()
func (s *Gluco2Solver) varDecayActivity() { s.var_inc *= 1 / s.Var_decay }

// varBumpActivityInc is the overload varBumpActivity(Var v, double inc) (protected).
// C: void varBumpActivity(Var v, double inc)
func (s *Gluco2Solver) varBumpActivityInc(v Gluco2Var, inc float64) {
	*s.activity.At(v) += inc
	if *s.activity.At(v) > 1e100 {
		// Rescale:
		for i := 0; i < s.NVars(); i++ {
			*s.activity.At(i) *= 1e-100
		}
		s.var_inc *= 1e-100
		if s.JustUsage() != 0 {
			for i := 0; i < s.jheap.Size(); i++ {
				j := s.jheap.At(i)
				a0 := *s.activity.At(s.getFaninVar0(j))
				a1 := *s.activity.At(s.getFaninVar1(j))
				if a0 > a1 {
					s.jdata.At(j).act_fanin = a0
				} else {
					s.jdata.At(j).act_fanin = a1
				}
			}
		}
	}
	// Update order_heap with respect to new activity:
	if s.order_heap.InHeap(v) {
		s.order_heap.Decrease(v)
	}
	if s.JustUsage() != 0 {
		s.updateJustActivity(v)
	}
}

// varBumpActivity is the overload varBumpActivity(Var v) (protected).
// C: void varBumpActivity(Var v)
func (s *Gluco2Solver) varBumpActivity(v Gluco2Var) { s.varBumpActivityInc(v, s.var_inc) }

// claDecayActivity is Solver::claDecayActivity (protected).
// C: void claDecayActivity()
func (s *Gluco2Solver) claDecayActivity() { s.cla_inc *= 1 / s.Clause_decay }

// claBumpActivity is Solver::claBumpActivity (protected).
// C: void claBumpActivity(Clause& c)
func (s *Gluco2Solver) claBumpActivity(c *Gluco2Clause) {
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
func (s *Gluco2Solver) locked(c *Gluco2Clause) bool {
	ref := func(l Gluco2Lit) bool {
		r := s.reason(Gluco2LitVar(l))
		return s.ValueLit(l) == Gluco2LTrue && r != Gluco2CRefUndef && !s.isGateCRef(r) && r == s.ca.Ael(c)
	}
	if c.Size() > 2 {
		return ref(*c.At(0))
	}
	return ref(*c.At(0)) || ref(*c.At(1))
}

// decisionLevel is Solver::decisionLevel (protected).
// C: int decisionLevel() const
func (s *Gluco2Solver) decisionLevel() int { return s.trail_lim.Size() }

// abstractLevel is Solver::abstractLevel (protected): used to represent an
// abstraction of sets of decision levels.
// C: uint32_t abstractLevel(Var x) const
func (s *Gluco2Solver) abstractLevel(x Gluco2Var) uint32 { return 1 << (uint(s.level(x)) & 31) }

// reason is Solver::reason (protected).
// C: CRef reason(Var x) const
func (s *Gluco2Solver) reason(x Gluco2Var) Gluco2CRef { return s.vardata.At(x).reason }

// level is Solver::level (protected).
// C: int level(Var x) const
func (s *Gluco2Solver) level(x Gluco2Var) int { return s.vardata.At(x).level }

// Value is the overload value(Var x).
// C: lbool value(Var x) const
func (s *Gluco2Solver) Value(x Gluco2Var) Gluco2Lbool { return *s.assigns.At(x) }

// ValueLit is the overload value(Lit p).
// C: lbool value(Lit p) const
func (s *Gluco2Solver) ValueLit(p Gluco2Lit) Gluco2Lbool {
	return s.assigns.At(Gluco2LitVar(p)).Xor(Gluco2Sign(p))
}

// ModelValue is the overload modelValue(Var x).
// C: lbool modelValue(Var x) const
func (s *Gluco2Solver) ModelValue(x Gluco2Var) Gluco2Lbool { return *s.Model.At(x) }

// ModelValueLit is the overload modelValue(Lit p).
// C: lbool modelValue(Lit p) const
func (s *Gluco2Solver) ModelValueLit(p Gluco2Lit) Gluco2Lbool {
	return s.Model.At(Gluco2LitVar(p)).Xor(Gluco2Sign(p))
}

// NAssigns is Solver::nAssigns.
// C: int nAssigns() const
func (s *Gluco2Solver) NAssigns() int { return s.trail.Size() }

// NClauses is Solver::nClauses.
// C: int nClauses() const
func (s *Gluco2Solver) NClauses() int { return s.clauses.Size() }

// NLearnts is Solver::nLearnts.
// C: int nLearnts() const
func (s *Gluco2Solver) NLearnts() int { return s.learnts.Size() }

// NVars is Solver::nVars.
// C: int nVars() const
func (s *Gluco2Solver) NVars() int { return s.vardata.Size() }

// NFreeVars is Solver::nFreeVars.
// C: int nFreeVars() const
func (s *Gluco2Solver) NFreeVars() int {
	lim := s.trail.Size()
	if s.trail_lim.Size() != 0 {
		lim = *s.trail_lim.At(0)
	}
	return int(s.Dec_vars) - lim
}

// SetPolarity is Solver::setPolarity.
// C: void setPolarity(Var v, bool b)
func (s *Gluco2Solver) SetPolarity(v Gluco2Var, b bool) {
	if b {
		*s.polarity.At(v) = 1
	} else {
		*s.polarity.At(v) = 0
	}
}

// SetDecisionVar is Solver::setDecisionVar.
// C: void setDecisionVar(Var v, bool b)
func (s *Gluco2Solver) SetDecisionVar(v Gluco2Var, b bool) {
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
func (s *Gluco2Solver) SetConfBudget(x int64) { s.conflict_budget = s.Conflicts + x }

// SetPropBudget is Solver::setPropBudget.
// C: void setPropBudget(int64_t x)
func (s *Gluco2Solver) SetPropBudget(x int64) { s.propagation_budget = s.Propagations + x }

// BudgetOff is Solver::budgetOff.
// C: void budgetOff()
func (s *Gluco2Solver) BudgetOff() {
	s.conflict_budget = -1
	s.propagation_budget = -1
}

// Interrupt is Solver::interrupt (may be called asynchronously).
// C: void interrupt()
func (s *Gluco2Solver) Interrupt() { s.asynch_interrupt = true }

// ClearInterrupt is Solver::clearInterrupt.
// C: void clearInterrupt()
func (s *Gluco2Solver) ClearInterrupt() { s.asynch_interrupt = false }

// withinBudget is Solver::withinBudget (protected).
// C: bool withinBudget() const
func (s *Gluco2Solver) withinBudget() bool {
	return !s.asynch_interrupt &&
		(s.conflict_budget < 0 || uint64(s.Conflicts) < uint64(s.conflict_budget)) &&
		(s.propagation_budget < 0 || uint64(s.Propagations) < uint64(s.propagation_budget))
}
