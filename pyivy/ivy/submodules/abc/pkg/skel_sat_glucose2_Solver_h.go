package abc

import "os"

// Hand-checked skeleton of src/sat/glucose2/Solver.h (namespace Gluco2).
// Differs from the MiniSat Solver: LBD-based restarts and clause database reduction, a binary
// watch list, incremental mode, an external callback, certified-UNSAT output and circuit justification (CGLUCOSE_EXP).
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

// Gluco2SolverVarData is Solver::VarData.
type Gluco2SolverVarData struct {
	reason Gluco2CRef
	level  int
}

// gluco2SolverMkVarData is Solver::mkVarData (protected static inline).
// C: static inline VarData mkVarData(CRef cr, int l)
func gluco2SolverMkVarData(cr Gluco2CRef, l int) Gluco2SolverVarData {
	panic("TODO, stubbed")
}

// Gluco2SolverWatcher is Solver::Watcher.
type Gluco2SolverWatcher struct {
	cref    Gluco2CRef
	blocker Gluco2Lit
}

// NewGluco2SolverWatcher is Watcher(CRef cr, Lit p).
// C: Watcher(CRef cr, Lit p)
func NewGluco2SolverWatcher(cr Gluco2CRef, p Gluco2Lit) Gluco2SolverWatcher {
	panic("TODO, stubbed")
}

// Eq is Watcher::operator==.
// C: bool operator==(const Watcher& w) const
func (w Gluco2SolverWatcher) Eq(o Gluco2SolverWatcher) bool {
	panic("TODO, stubbed")
}

// Ne is Watcher::operator!=.
// C: bool operator!=(const Watcher& w) const
func (w Gluco2SolverWatcher) Ne(o Gluco2SolverWatcher) bool {
	panic("TODO, stubbed")
}

// Gluco2SolverWatcherDeleted is Solver::WatcherDeleted.
type Gluco2SolverWatcherDeleted struct {
	ca *Gluco2ClauseAllocator // const ClauseAllocator&
}

// NewGluco2SolverWatcherDeleted is WatcherDeleted(const ClauseAllocator& _ca).
// C: WatcherDeleted(const ClauseAllocator& _ca)
func NewGluco2SolverWatcherDeleted(ca *Gluco2ClauseAllocator) Gluco2SolverWatcherDeleted {
	panic("TODO, stubbed")
}

// Call is WatcherDeleted::operator().
// C: bool operator()(const Watcher& w) const
func (d Gluco2SolverWatcherDeleted) Call(w Gluco2SolverWatcher) bool {
	panic("TODO, stubbed")
}

// Gluco2SolverVarOrderLt is Solver::VarOrderLt.
type Gluco2SolverVarOrderLt struct {
	activity *Gluco2Vec[float64] // const vec<double>&
}

// NewGluco2SolverVarOrderLt is VarOrderLt(const vec<double>& act).
// C: VarOrderLt(const vec<double>& act)
func NewGluco2SolverVarOrderLt(act *Gluco2Vec[float64]) Gluco2SolverVarOrderLt {
	panic("TODO, stubbed")
}

// Call is VarOrderLt::operator().
// C: bool operator()(Var x, Var y) const
func (l Gluco2SolverVarOrderLt) Call(x int, y int) bool {
	panic("TODO, stubbed")
}

// Gluco2SolverJustData is Solver::JustData.
type Gluco2SolverJustData struct {
	now       uint32 // 1 bit
	act_fanin float64
}

// gluco2SolverMkJustData is Solver::mkJustData (protected static inline).
// C: static inline JustData mkJustData(bool n)
func gluco2SolverMkJustData(n bool) Gluco2SolverJustData {
	panic("TODO, stubbed")
}

// Gluco2SolverJustOrderLt is Solver::JustOrderLt.
type Gluco2SolverJustOrderLt struct {
	pS *Gluco2Solver // const Solver *
}

// NewNSSolverJustOrderLt is JustOrderLt(const Solver * _pS).
// C: JustOrderLt(const Solver * _pS)
func NewNSSolverJustOrderLt(pS *Gluco2Solver) Gluco2SolverJustOrderLt {
	panic("TODO, stubbed")
}

// Call is JustOrderLt::operator().
// C: bool operator()(Var x, Var y) const
func (l Gluco2SolverJustOrderLt) Call(x int, y int) bool {
	panic("TODO, stubbed")
}

// Gluco2SolverJustWatch is Solver::JustWatch.
type Gluco2SolverJustWatch struct {
	header_dir uint32 // header.dir: 1 bit (anonymous struct member `header`)
	head       Gluco2Var
	next       Gluco2Var
	prev       Gluco2Var
}

// gluco2SolverMkJustWatch is Solver::mkJustWatch (protected static inline).
// C: static inline JustWatch mkJustWatch()
func gluco2SolverMkJustWatch() Gluco2SolverJustWatch {
	panic("TODO, stubbed")
}

// Gluco2Solver is Gluco2::Solver, the CDCL SAT solver. Public data members are
// exported; protected ones keep the C++ spelling. SimpSolver embeds it.
type Gluco2Solver struct {
	SolverType int // ABC identifies the solver's type as 0

	PCnfMan                any                       // void * pCnfMan: external CNF manager
	PCnfFunc               func(any, int, []int) int // int(*pCnfFunc)(void *p, int, int*): external callback. messages: 0: unsat; 1: sat; -1: still working
	NCallConfl             int                       // callback will be called every this number of conflicts
	Terminate_search_early bool                      // used to stop the solver early if it as instructed by an external caller
	Pstop                  *int                      // another callback
	NRuntimeLimit          uint64                    // runtime limit
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

	// Circuit-based justification (protected):
	jdata                                 Gluco2Vec[Gluco2SolverJustData]
	jwatch                                Gluco2Vec[Gluco2SolverJustWatch]
	var2FaninLits                         Gluco2Vec[Gluco2Lit] // (~0): undefine
	var2TravId                            Gluco2Vec[uint32]
	var2Fanout0, var2FanoutN, var2FanoutP Gluco2Vec[Gluco2Lit]
	itpc                                  Gluco2CRef // the interpreted clause of a gate
	travId_prev, travId                   int64
	jheap                                 Gluco2Heap[Gluco2SolverJustOrderLt]
	jstack                                Gluco2Vec[Gluco2Var] // called by unchecked_enqueue, consumed by pickJustLit, cleaned by cancelUntil

	JustModel Gluco2Vec[Gluco2Lit] // model obtained by justification enabled (public)
}

// NewGluco2Solver is Solver::Solver().
// C: Solver::Solver()
func NewGluco2Solver() *Gluco2Solver {
	panic("TODO, stubbed")
}

// Destroy is Solver::~Solver() (virtual).
// C: virtual Solver::~Solver()
func (s *Gluco2Solver) Destroy() {
	panic("TODO, stubbed")
}

// NewVar is Solver::newVar (polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *Gluco2Solver) NewVar(polarity bool, dvar bool) Gluco2Var {
	panic("TODO, stubbed")
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *Gluco2Solver) AddClause(ps *Gluco2Vec[Gluco2Lit]) bool {
	panic("TODO, stubbed")
}

// AddEmptyClause is Solver::addEmptyClause.
// C: bool addEmptyClause()
func (s *Gluco2Solver) AddEmptyClause() bool {
	panic("TODO, stubbed")
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *Gluco2Solver) AddClauseUnit(p Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *Gluco2Solver) AddClauseBinary(p Gluco2Lit, q Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *Gluco2Solver) AddClauseTernary(p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// AddClause_ is Solver::addClause_ (no superfluous internal copy; clobbers ps).
// C: bool addClause_(vec<Lit>& ps)
func (s *Gluco2Solver) AddClause_(ps *Gluco2Vec[Gluco2Lit]) bool {
	panic("TODO, stubbed")
}

// Simplify is Solver::simplify.
// C: bool simplify()
func (s *Gluco2Solver) Simplify() bool {
	panic("TODO, stubbed")
}

// Solve is the overload solve(const vec<Lit>& assumps).
// C: bool solve(const vec<Lit>& assumps)
func (s *Gluco2Solver) Solve(assumps *Gluco2Vec[Gluco2Lit]) bool {
	panic("TODO, stubbed")
}

// SolveLimited is Solver::solveLimited (with resource constraints).
// C: lbool solveLimited(const vec<Lit>& assumps)
func (s *Gluco2Solver) SolveLimited(assumps *Gluco2Vec[Gluco2Lit]) Gluco2Lbool {
	panic("TODO, stubbed")
}

// SolveNoAssumps is the overload solve().
// C: bool solve()
func (s *Gluco2Solver) SolveNoAssumps() bool {
	panic("TODO, stubbed")
}

// Solve1 is the overload solve(Lit p).
// C: bool solve(Lit p)
func (s *Gluco2Solver) Solve1(p Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// Solve2 is the overload solve(Lit p, Lit q).
// C: bool solve(Lit p, Lit q)
func (s *Gluco2Solver) Solve2(p Gluco2Lit, q Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r).
// C: bool solve(Lit p, Lit q, Lit r)
func (s *Gluco2Solver) Solve3(p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// Okay is Solver::okay.
// C: bool okay() const
func (s *Gluco2Solver) Okay() bool {
	panic("TODO, stubbed")
}

// ToDimacsStream is the overload toDimacs(FILE* f, const vec<Lit>& assumps).
// C: void toDimacs(FILE* f, const vec<Lit>& assumps)
func (s *Gluco2Solver) ToDimacsStream(f *os.File, assumps *Gluco2Vec[Gluco2Lit]) {
	panic("TODO, stubbed")
}

// ToDimacsFile is the overload toDimacs(const char *file, const vec<Lit>& assumps).
// C: void toDimacs(const char *file, const vec<Lit>& assumps)
func (s *Gluco2Solver) ToDimacsFile(file string, assumps *Gluco2Vec[Gluco2Lit]) {
	panic("TODO, stubbed")
}

// ToDimacsClause is the overload toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max).
// C: void toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max)
func (s *Gluco2Solver) ToDimacsClause(f *os.File, c *Gluco2Clause, map_ *Gluco2Vec[Gluco2Var], max *Gluco2Var) {
	panic("TODO, stubbed")
}

// ToDimacsFileNoAssumps is the overload toDimacs(const char* file).
// C: void toDimacs(const char* file)
func (s *Gluco2Solver) ToDimacsFileNoAssumps(file string) {
	panic("TODO, stubbed")
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *Gluco2Solver) ToDimacsFile1(file string, p Gluco2Lit) {
	panic("TODO, stubbed")
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *Gluco2Solver) ToDimacsFile2(file string, p Gluco2Lit, q Gluco2Lit) {
	panic("TODO, stubbed")
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *Gluco2Solver) ToDimacsFile3(file string, p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) {
	panic("TODO, stubbed")
}

// Sat_solver_set_var_fanin_lit is Solver::sat_solver_set_var_fanin_lit.
// C: void sat_solver_set_var_fanin_lit(int, int, int)
func (s *Gluco2Solver) Sat_solver_set_var_fanin_lit(v int, lit0 int, lit1 int) {
	panic("TODO, stubbed")
}

// Sat_solver_start_new_round is Solver::sat_solver_start_new_round.
// C: void sat_solver_start_new_round()
func (s *Gluco2Solver) Sat_solver_start_new_round() {
	panic("TODO, stubbed")
}

// Sat_solver_mark_cone is Solver::sat_solver_mark_cone.
// C: void sat_solver_mark_cone(int)
func (s *Gluco2Solver) Sat_solver_mark_cone(v int) {
	panic("TODO, stubbed")
}

// AddVar is Solver::addVar (adds enough variables to make sure there is variable v).
// C: void addVar(Var v)
func (s *Gluco2Solver) AddVar(v Gluco2Var) {
	panic("TODO, stubbed")
}

// PrintLit is Solver::printLit.
// C: void printLit(Lit l)
func (s *Gluco2Solver) PrintLit(l Gluco2Lit) {
	panic("TODO, stubbed")
}

// PrintClause is Solver::printClause.
// C: void printClause(CRef c)
func (s *Gluco2Solver) PrintClause(c Gluco2CRef) {
	panic("TODO, stubbed")
}

// PrintInitialClause is Solver::printInitialClause.
// C: void printInitialClause(CRef c)
func (s *Gluco2Solver) PrintInitialClause(c Gluco2CRef) {
	panic("TODO, stubbed")
}

// GetCex is Solver::getCex.
// C: int * getCex() const
func (s *Gluco2Solver) GetCex() []int {
	panic("TODO, stubbed")
}

// SetIncrementalMode is Solver::setIncrementalMode.
// C: void setIncrementalMode()
func (s *Gluco2Solver) SetIncrementalMode() {
	panic("TODO, stubbed")
}

// InitNbInitialVars is Solver::initNbInitialVars.
// C: void initNbInitialVars(int nb)
func (s *Gluco2Solver) InitNbInitialVars(nb int) {
	panic("TODO, stubbed")
}

// PrintIncrementalStats is Solver::printIncrementalStats.
// C: void printIncrementalStats()
func (s *Gluco2Solver) PrintIncrementalStats() {
	panic("TODO, stubbed")
}

// Reset is Solver::reset (virtual).
// C: virtual void reset()
func (s *Gluco2Solver) Reset() {
	panic("TODO, stubbed")
}

// SetPolarity is Solver::setPolarity.
// C: void setPolarity(Var v, bool b)
func (s *Gluco2Solver) SetPolarity(v Gluco2Var, b bool) {
	panic("TODO, stubbed")
}

// SetDecisionVar is Solver::setDecisionVar.
// C: void setDecisionVar(Var v, bool b)
func (s *Gluco2Solver) SetDecisionVar(v Gluco2Var, b bool) {
	panic("TODO, stubbed")
}

// Value is the overload value(Var x).
// C: lbool value(Var x) const
func (s *Gluco2Solver) Value(x Gluco2Var) Gluco2Lbool {
	panic("TODO, stubbed")
}

// ValueLit is the overload value(Lit p).
// C: lbool value(Lit p) const
func (s *Gluco2Solver) ValueLit(p Gluco2Lit) Gluco2Lbool {
	panic("TODO, stubbed")
}

// ModelValue is the overload modelValue(Var x).
// C: lbool modelValue(Var x) const
func (s *Gluco2Solver) ModelValue(x Gluco2Var) Gluco2Lbool {
	panic("TODO, stubbed")
}

// ModelValueLit is the overload modelValue(Lit p).
// C: lbool modelValue(Lit p) const
func (s *Gluco2Solver) ModelValueLit(p Gluco2Lit) Gluco2Lbool {
	panic("TODO, stubbed")
}

// NAssigns is Solver::nAssigns.
// C: int nAssigns() const
func (s *Gluco2Solver) NAssigns() int {
	panic("TODO, stubbed")
}

// NClauses is Solver::nClauses.
// C: int nClauses() const
func (s *Gluco2Solver) NClauses() int {
	panic("TODO, stubbed")
}

// NLearnts is Solver::nLearnts.
// C: int nLearnts() const
func (s *Gluco2Solver) NLearnts() int {
	panic("TODO, stubbed")
}

// NVars is Solver::nVars.
// C: int nVars() const
func (s *Gluco2Solver) NVars() int {
	panic("TODO, stubbed")
}

// NFreeVars is Solver::nFreeVars.
// C: int nFreeVars() const
func (s *Gluco2Solver) NFreeVars() int {
	panic("TODO, stubbed")
}

// SetConfBudget is Solver::setConfBudget.
// C: void setConfBudget(int64_t x)
func (s *Gluco2Solver) SetConfBudget(x int64) {
	panic("TODO, stubbed")
}

// SetPropBudget is Solver::setPropBudget.
// C: void setPropBudget(int64_t x)
func (s *Gluco2Solver) SetPropBudget(x int64) {
	panic("TODO, stubbed")
}

// BudgetOff is Solver::budgetOff.
// C: void budgetOff()
func (s *Gluco2Solver) BudgetOff() {
	panic("TODO, stubbed")
}

// Interrupt is Solver::interrupt (may be called asynchronously).
// C: void interrupt()
func (s *Gluco2Solver) Interrupt() {
	panic("TODO, stubbed")
}

// ClearInterrupt is Solver::clearInterrupt.
// C: void clearInterrupt()
func (s *Gluco2Solver) ClearInterrupt() {
	panic("TODO, stubbed")
}

// GarbageCollect is Solver::garbageCollect (virtual).
// C: virtual void garbageCollect()
func (s *Gluco2Solver) GarbageCollect() {
	panic("TODO, stubbed")
}

// CheckGarbageFrac is the overload checkGarbage(double gf).
// C: void checkGarbage(double gf)
func (s *Gluco2Solver) CheckGarbageFrac(gf float64) {
	panic("TODO, stubbed")
}

// CheckGarbage is the overload checkGarbage().
// C: void checkGarbage()
func (s *Gluco2Solver) CheckGarbage() {
	panic("TODO, stubbed")
}

// gluco2SolverDrand is Solver::drand (protected static inline).
// C: static inline double drand(double& seed)
func gluco2SolverDrand(seed *float64) float64 {
	panic("TODO, stubbed")
}

// gluco2SolverIrand is Solver::irand (protected static inline).
// C: static inline int irand(double& seed, int size)
func gluco2SolverIrand(seed *float64, size int) int {
	panic("TODO, stubbed")
}

// insertVarOrder is Solver::insertVarOrder (protected).
// C: void insertVarOrder(Var x)
func (s *Gluco2Solver) insertVarOrder(x Gluco2Var) {
	panic("TODO, stubbed")
}

// pickBranchLit is Solver::pickBranchLit (protected).
// C: Lit pickBranchLit()
func (s *Gluco2Solver) pickBranchLit() Gluco2Lit {
	panic("TODO, stubbed")
}

// newDecisionLevel is Solver::newDecisionLevel (protected).
// C: void newDecisionLevel()
func (s *Gluco2Solver) newDecisionLevel() {
	panic("TODO, stubbed")
}

// uncheckedEnqueue is Solver::uncheckedEnqueue (protected; from defaults to CRef_Undef).
// C: void uncheckedEnqueue(Lit p, CRef from = CRef_Undef)
func (s *Gluco2Solver) uncheckedEnqueue(p Gluco2Lit, from Gluco2CRef) {
	panic("TODO, stubbed")
}

// enqueue is Solver::enqueue (protected; from defaults to CRef_Undef).
// C: bool enqueue(Lit p, CRef from = CRef_Undef)
func (s *Gluco2Solver) enqueue(p Gluco2Lit, from Gluco2CRef) bool {
	panic("TODO, stubbed")
}

// propagate is Solver::propagate (protected).
// C: CRef propagate()
func (s *Gluco2Solver) propagate() Gluco2CRef {
	panic("TODO, stubbed")
}

// cancelUntil is Solver::cancelUntil (protected).
// C: void cancelUntil(int level)
func (s *Gluco2Solver) cancelUntil(level int) {
	panic("TODO, stubbed")
}

// analyze is Solver::analyze (protected).
// C: void analyze(CRef confl, vec<Lit>& out_learnt, vec<Lit> & selectors, int& out_btlevel, unsigned int &nblevels, unsigned int &szWithoutSelectors)
func (s *Gluco2Solver) analyze(confl Gluco2CRef, outLearnt *Gluco2Vec[Gluco2Lit], selectors *Gluco2Vec[Gluco2Lit], outBtlevel *int, nblevels *uint32, szWithoutSelectors *uint32) {
	panic("TODO, stubbed")
}

// analyzeFinal is Solver::analyzeFinal (protected).
// C: void analyzeFinal(Lit p, vec<Lit>& out_conflict)
func (s *Gluco2Solver) analyzeFinal(p Gluco2Lit, outConflict *Gluco2Vec[Gluco2Lit]) {
	panic("TODO, stubbed")
}

// litRedundant is Solver::litRedundant (protected).
// C: bool litRedundant(Lit p, uint32_t abstract_levels)
func (s *Gluco2Solver) litRedundant(p Gluco2Lit, abstractLevels uint32) bool {
	panic("TODO, stubbed")
}

// search is Solver::search (protected).
// C: lbool search(int nof_conflicts)
func (s *Gluco2Solver) search(nofConflicts int) Gluco2Lbool {
	panic("TODO, stubbed")
}

// solve_ is Solver::solve_ (protected).
// C: lbool solve_()
func (s *Gluco2Solver) solve_() Gluco2Lbool {
	panic("TODO, stubbed")
}

// reduceDB is Solver::reduceDB (protected).
// C: void reduceDB()
func (s *Gluco2Solver) reduceDB() {
	panic("TODO, stubbed")
}

// removeSatisfied is Solver::removeSatisfied (protected).
// C: void removeSatisfied(vec<CRef>& cs)
func (s *Gluco2Solver) removeSatisfied(cs *Gluco2Vec[Gluco2CRef]) {
	panic("TODO, stubbed")
}

// rebuildOrderHeap is Solver::rebuildOrderHeap (protected).
// C: void rebuildOrderHeap()
func (s *Gluco2Solver) rebuildOrderHeap() {
	panic("TODO, stubbed")
}

// varDecayActivity is Solver::varDecayActivity (protected).
// C: void varDecayActivity()
func (s *Gluco2Solver) varDecayActivity() {
	panic("TODO, stubbed")
}

// varBumpActivityInc is the overload varBumpActivity(Var v, double inc) (protected).
// C: void varBumpActivity(Var v, double inc)
func (s *Gluco2Solver) varBumpActivityInc(v Gluco2Var, inc float64) {
	panic("TODO, stubbed")
}

// varBumpActivity is the overload varBumpActivity(Var v) (protected).
// C: void varBumpActivity(Var v)
func (s *Gluco2Solver) varBumpActivity(v Gluco2Var) {
	panic("TODO, stubbed")
}

// claDecayActivity is Solver::claDecayActivity (protected).
// C: void claDecayActivity()
func (s *Gluco2Solver) claDecayActivity() {
	panic("TODO, stubbed")
}

// claBumpActivity is Solver::claBumpActivity (protected).
// C: void claBumpActivity(Clause& c)
func (s *Gluco2Solver) claBumpActivity(c *Gluco2Clause) {
	panic("TODO, stubbed")
}

// attachClause is Solver::attachClause (protected).
// C: void attachClause(CRef cr)
func (s *Gluco2Solver) attachClause(cr Gluco2CRef) {
	panic("TODO, stubbed")
}

// detachClause is Solver::detachClause (protected; strict defaults to false).
// C: void detachClause(CRef cr, bool strict = false)
func (s *Gluco2Solver) detachClause(cr Gluco2CRef, strict bool) {
	panic("TODO, stubbed")
}

// removeClause is Solver::removeClause (protected).
// C: void removeClause(CRef cr)
func (s *Gluco2Solver) removeClause(cr Gluco2CRef) {
	panic("TODO, stubbed")
}

// locked is Solver::locked (protected).
// C: bool locked(const Clause& c) const
func (s *Gluco2Solver) locked(c *Gluco2Clause) bool {
	panic("TODO, stubbed")
}

// satisfied is Solver::satisfied (protected).
// C: bool satisfied(const Clause& c) const
func (s *Gluco2Solver) satisfied(c *Gluco2Clause) bool {
	panic("TODO, stubbed")
}

// computeLBD is the overload computeLBD(const vec<Lit>& lits, int end = -1) (protected).
// C: unsigned int computeLBD(const vec<Lit>& lits, int end = -1)
func (s *Gluco2Solver) computeLBD(lits *Gluco2Vec[Gluco2Lit], end int) uint32 {
	panic("TODO, stubbed")
}

// computeLBDClause is the overload computeLBD(const Clause& c) (protected).
// C: unsigned int computeLBD(const Clause& c)
func (s *Gluco2Solver) computeLBDClause(c *Gluco2Clause) uint32 {
	panic("TODO, stubbed")
}

// minimisationWithBinaryResolution is Solver::minimisationWithBinaryResolution (protected).
// C: void minimisationWithBinaryResolution(vec<Lit>& out_learnt)
func (s *Gluco2Solver) minimisationWithBinaryResolution(outLearnt *Gluco2Vec[Gluco2Lit]) {
	panic("TODO, stubbed")
}

// isSelector is Solver::isSelector (protected inline).
// C: inline bool isSelector(Var v)
func (s *Gluco2Solver) isSelector(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// relocAll is Solver::relocAll (protected).
// C: void relocAll(ClauseAllocator& to)
func (s *Gluco2Solver) relocAll(to *Gluco2ClauseAllocator) {
	panic("TODO, stubbed")
}

// decisionLevel is Solver::decisionLevel (protected).
// C: int decisionLevel() const
func (s *Gluco2Solver) decisionLevel() int {
	panic("TODO, stubbed")
}

// abstractLevel is Solver::abstractLevel (protected).
// C: uint32_t abstractLevel(Var x) const
func (s *Gluco2Solver) abstractLevel(x Gluco2Var) uint32 {
	panic("TODO, stubbed")
}

// reason is Solver::reason (protected).
// C: CRef reason(Var x) const
func (s *Gluco2Solver) reason(x Gluco2Var) Gluco2CRef {
	panic("TODO, stubbed")
}

// Level is Solver::level (public in src/sat/glucose2, moved to compile JustOrderLt).
// C: int level(Var x) const
func (s *Gluco2Solver) Level(x Gluco2Var) int {
	panic("TODO, stubbed")
}

// progressEstimate is Solver::progressEstimate (protected).
// C: double progressEstimate() const
func (s *Gluco2Solver) progressEstimate() float64 {
	panic("TODO, stubbed")
}

// withinBudget is Solver::withinBudget (protected).
// C: bool withinBudget() const
func (s *Gluco2Solver) withinBudget() bool {
	panic("TODO, stubbed")
}

// justification support (src/sat/glucose2; protected unless noted)

// addJwatch is Solver::addJwatch.
// C: void addJwatch(Var host, Var member)
func (s *Gluco2Solver) addJwatch(host Gluco2Var, member Gluco2Var) {
	panic("TODO, stubbed")
}

// delJwatch is Solver::delJwatch.
// C: void delJwatch(Var member)
func (s *Gluco2Solver) delJwatch(member Gluco2Var) {
	panic("TODO, stubbed")
}

// isTwoFanin is Solver::isTwoFanin.
// C: bool isTwoFanin(Var v) const
func (s *Gluco2Solver) isTwoFanin(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// isAND is Solver::isAND.
// C: bool isAND(Var v) const
func (s *Gluco2Solver) isAND(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// isJReason is Solver::isJReason.
// C: bool isJReason(Var v) const
func (s *Gluco2Solver) isJReason(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// getFaninLit0 is Solver::getFaninLit0.
// C: Lit getFaninLit0(Var v) const
func (s *Gluco2Solver) getFaninLit0(v Gluco2Var) Gluco2Lit {
	panic("TODO, stubbed")
}

// getFaninLit1 is Solver::getFaninLit1.
// C: Lit getFaninLit1(Var v) const
func (s *Gluco2Solver) getFaninLit1(v Gluco2Var) Gluco2Lit {
	panic("TODO, stubbed")
}

// getFaninC0 is Solver::getFaninC0.
// C: bool getFaninC0(Var v) const
func (s *Gluco2Solver) getFaninC0(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// getFaninC1 is Solver::getFaninC1.
// C: bool getFaninC1(Var v) const
func (s *Gluco2Solver) getFaninC1(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// getFaninVar0 is Solver::getFaninVar0.
// C: Var getFaninVar0(Var v) const
func (s *Gluco2Solver) getFaninVar0(v Gluco2Var) Gluco2Var {
	panic("TODO, stubbed")
}

// getFaninVar1 is Solver::getFaninVar1.
// C: Var getFaninVar1(Var v) const
func (s *Gluco2Solver) getFaninVar1(v Gluco2Var) Gluco2Var {
	panic("TODO, stubbed")
}

// getFaninPlt0 is Solver::getFaninPlt0.
// C: Lit getFaninPlt0(Var v) const
func (s *Gluco2Solver) getFaninPlt0(v Gluco2Var) Gluco2Lit {
	panic("TODO, stubbed")
}

// getFaninPlt1 is Solver::getFaninPlt1.
// C: Lit getFaninPlt1(Var v) const
func (s *Gluco2Solver) getFaninPlt1(v Gluco2Var) Gluco2Lit {
	panic("TODO, stubbed")
}

// maxActiveLit is Solver::maxActiveLit.
// C: Lit maxActiveLit(Lit lit0, Lit lit1) const
func (s *Gluco2Solver) maxActiveLit(lit0 Gluco2Lit, lit1 Gluco2Lit) Gluco2Lit {
	panic("TODO, stubbed")
}

// gateJustFanin is Solver::gateJustFanin (l_Undef=satisfied, 0/1 = fanin0/fanin1 requires justify).
// C: Lit gateJustFanin(Var v) const
func (s *Gluco2Solver) gateJustFanin(v Gluco2Var) Gluco2Lit {
	panic("TODO, stubbed")
}

// gateAddJwatch is Solver::gateAddJwatch.
// C: void gateAddJwatch(Var v)
func (s *Gluco2Solver) gateAddJwatch(v Gluco2Var) {
	panic("TODO, stubbed")
}

// gatePropagateCheck is Solver::gatePropagateCheck.
// C: CRef gatePropagateCheck(Var v, Var t)
func (s *Gluco2Solver) gatePropagateCheck(v Gluco2Var, t Gluco2Var) Gluco2CRef {
	panic("TODO, stubbed")
}

// gatePropagateCheckThis is Solver::gatePropagateCheckThis.
// C: CRef gatePropagateCheckThis(Var v)
func (s *Gluco2Solver) gatePropagateCheckThis(v Gluco2Var) Gluco2CRef {
	panic("TODO, stubbed")
}

// gatePropagateCheckFanout is Solver::gatePropagateCheckFanout.
// C: CRef gatePropagateCheckFanout(Var v, Lit lfo)
func (s *Gluco2Solver) gatePropagateCheckFanout(v Gluco2Var, lfo Gluco2Lit) Gluco2CRef {
	panic("TODO, stubbed")
}

// setItpcSize is Solver::setItpcSize (sz <= 3).
// C: void setItpcSize(int sz)
func (s *Gluco2Solver) setItpcSize(sz int) {
	panic("TODO, stubbed")
}

// updateJustActivity is Solver::updateJustActivity.
// C: void updateJustActivity(Var v)
func (s *Gluco2Solver) updateJustActivity(v Gluco2Var) {
	panic("TODO, stubbed")
}

// ResetJustData is Solver::ResetJustData.
// C: void ResetJustData(bool fCleanMemory)
func (s *Gluco2Solver) resetJustData(fCleanMemory bool) {
	panic("TODO, stubbed")
}

// pickJustLit is Solver::pickJustLit.
// C: Lit pickJustLit(Var& j_reason)
func (s *Gluco2Solver) pickJustLit(jReason *Gluco2Var) Gluco2Lit {
	panic("TODO, stubbed")
}

// justCheck is Solver::justCheck.
// C: void justCheck()
func (s *Gluco2Solver) justCheck() {
	panic("TODO, stubbed")
}

// pushJustQueue is Solver::pushJustQueue.
// C: void pushJustQueue(Var v)
func (s *Gluco2Solver) pushJustQueue(v Gluco2Var) {
	panic("TODO, stubbed")
}

// restoreJustQueue is Solver::restoreJustQueue (call with cancelUntil).
// C: void restoreJustQueue(int level)
func (s *Gluco2Solver) restoreJustQueue(level int) {
	panic("TODO, stubbed")
}

// gateClearJwatch is Solver::gateClearJwatch.
// C: void gateClearJwatch(Var v, int backtrack_level)
func (s *Gluco2Solver) gateClearJwatch(v Gluco2Var, backtrackLevel int) {
	panic("TODO, stubbed")
}

// gatePropagate is Solver::gatePropagate.
// C: CRef gatePropagate(Lit p)
func (s *Gluco2Solver) gatePropagate(p Gluco2Lit) Gluco2CRef {
	panic("TODO, stubbed")
}

// interpret is Solver::interpret.
// C: CRef interpret(Var v, Var t)
func (s *Gluco2Solver) interpret(v Gluco2Var, t Gluco2Var) Gluco2CRef {
	panic("TODO, stubbed")
}

// castCRef is Solver::castCRef (interpret a gate into a clause).
// C: CRef castCRef(Lit p)
func (s *Gluco2Solver) castCRef(p Gluco2Lit) Gluco2CRef {
	panic("TODO, stubbed")
}

// getConfClause is Solver::getConfClause.
// C: CRef getConfClause(CRef r)
func (s *Gluco2Solver) getConfClause(r Gluco2CRef) Gluco2CRef {
	panic("TODO, stubbed")
}

// Var2CRef is Solver::Var2CRef.
// C: CRef Var2CRef(Var v) const
func (s *Gluco2Solver) Var2CRef(v Gluco2Var) Gluco2CRef {
	panic("TODO, stubbed")
}

// CRef2Var is Solver::CRef2Var.
// C: Var CRef2Var(CRef cr) const
func (s *Gluco2Solver) CRef2Var(cr Gluco2CRef) Gluco2Var {
	panic("TODO, stubbed")
}

// isGateCRef is Solver::isGateCRef.
// C: bool isGateCRef(CRef cr) const
func (s *Gluco2Solver) isGateCRef(cr Gluco2CRef) bool {
	panic("TODO, stubbed")
}

// SetVarFaninLits is Solver::setVarFaninLits (public).
// C: void setVarFaninLits(Var v, Lit lit1, Lit lit2)
func (s *Gluco2Solver) SetVarFaninLits(v Gluco2Var, lit1 Gluco2Lit, lit2 Gluco2Lit) {
	panic("TODO, stubbed")
}

// SetNewRound is Solver::setNewRound (public).
// C: int setNewRound()
func (s *Gluco2Solver) SetNewRound() int {
	panic("TODO, stubbed")
}

// MarkCone is Solver::markCone (public).
// C: void markCone(Var v)
func (s *Gluco2Solver) MarkCone(v Gluco2Var) {
	panic("TODO, stubbed")
}

// IsRoundWatch is Solver::isRoundWatch (public).
// C: bool isRoundWatch(Var v) const
func (s *Gluco2Solver) IsRoundWatch(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// GetJustData is Solver::getJustData (public).
// C: const JustData& getJustData(int v) const
func (s *Gluco2Solver) GetJustData(v int) *Gluco2SolverJustData {
	panic("TODO, stubbed")
}

// VarActivity is Solver::varActivity (public).
// C: double varActivity(int v) const
func (s *Gluco2Solver) VarActivity(v int) float64 {
	panic("TODO, stubbed")
}

// JustActivity is Solver::justActivity (public).
// C: double justActivity(int v) const
func (s *Gluco2Solver) JustActivity(v int) float64 {
	panic("TODO, stubbed")
}

// VarPolarity is Solver::varPolarity (public).
// C: int varPolarity(int v)
func (s *Gluco2Solver) VarPolarity(v int) int {
	panic("TODO, stubbed")
}

// JustUsage is Solver::justUsage (public).
// C: int justUsage() const
func (s *Gluco2Solver) JustUsage() int {
	panic("TODO, stubbed")
}
