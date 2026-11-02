package abc

import "os"

// Hand-checked skeleton of src/sat/glucose/Solver.h (namespace Gluco).
// Differs from the MiniSat Solver: LBD-based restarts and clause database reduction, a binary
// watch list, incremental mode, an external callback, certified-UNSAT output.
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

// GlucoSolverVarData is Solver::VarData.
type GlucoSolverVarData struct {
	reason GlucoCRef
	level  int
}

// glucoSolverMkVarData is Solver::mkVarData (protected static inline).
// C: static inline VarData mkVarData(CRef cr, int l)
func glucoSolverMkVarData(cr GlucoCRef, l int) GlucoSolverVarData {
	panic("TODO, stubbed")
}

// GlucoSolverWatcher is Solver::Watcher.
type GlucoSolverWatcher struct {
	cref    GlucoCRef
	blocker GlucoLit
}

// NewGlucoSolverWatcher is Watcher(CRef cr, Lit p).
// C: Watcher(CRef cr, Lit p)
func NewGlucoSolverWatcher(cr GlucoCRef, p GlucoLit) GlucoSolverWatcher {
	panic("TODO, stubbed")
}

// Eq is Watcher::operator==.
// C: bool operator==(const Watcher& w) const
func (w GlucoSolverWatcher) Eq(o GlucoSolverWatcher) bool {
	panic("TODO, stubbed")
}

// Ne is Watcher::operator!=.
// C: bool operator!=(const Watcher& w) const
func (w GlucoSolverWatcher) Ne(o GlucoSolverWatcher) bool {
	panic("TODO, stubbed")
}

// GlucoSolverWatcherDeleted is Solver::WatcherDeleted.
type GlucoSolverWatcherDeleted struct {
	ca *GlucoClauseAllocator // const ClauseAllocator&
}

// NewGlucoSolverWatcherDeleted is WatcherDeleted(const ClauseAllocator& _ca).
// C: WatcherDeleted(const ClauseAllocator& _ca)
func NewGlucoSolverWatcherDeleted(ca *GlucoClauseAllocator) GlucoSolverWatcherDeleted {
	panic("TODO, stubbed")
}

// Call is WatcherDeleted::operator().
// C: bool operator()(const Watcher& w) const
func (d GlucoSolverWatcherDeleted) Call(w GlucoSolverWatcher) bool {
	panic("TODO, stubbed")
}

// GlucoSolverVarOrderLt is Solver::VarOrderLt.
type GlucoSolverVarOrderLt struct {
	activity *GlucoVec[float64] // const vec<double>&
}

// NewGlucoSolverVarOrderLt is VarOrderLt(const vec<double>& act).
// C: VarOrderLt(const vec<double>& act)
func NewGlucoSolverVarOrderLt(act *GlucoVec[float64]) GlucoSolverVarOrderLt {
	panic("TODO, stubbed")
}

// Call is VarOrderLt::operator().
// C: bool operator()(Var x, Var y) const
func (l GlucoSolverVarOrderLt) Call(x int, y int) bool {
	panic("TODO, stubbed")
}

// GlucoSolver is Gluco::Solver, the CDCL SAT solver. Public data members are
// exported; protected ones keep the C++ spelling. SimpSolver embeds it.
type GlucoSolver struct {
	SolverType int // ABC identifies the solver's type as 0

	PCnfMan                any                       // void * pCnfMan: external CNF manager
	PCnfFunc               func(any, int, []int) int // int(*pCnfFunc)(void *p, int, int*): external callback. messages: 0: unsat; 1: sat; -1: still working
	NCallConfl             int                       // callback will be called every this number of conflicts
	Terminate_search_early bool                      // used to stop the solver early if it as instructed by an external caller
	Pstop                  *int                      // another callback
	NRuntimeLimit          uint64                    // runtime limit
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
}

// NewGlucoSolver is Solver::Solver().
// C: Solver::Solver()
func NewGlucoSolver() *GlucoSolver {
	panic("TODO, stubbed")
}

// Destroy is Solver::~Solver() (virtual).
// C: virtual Solver::~Solver()
func (s *GlucoSolver) Destroy() {
	panic("TODO, stubbed")
}

// NewVar is Solver::newVar (polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *GlucoSolver) NewVar(polarity bool, dvar bool) GlucoVar {
	panic("TODO, stubbed")
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *GlucoSolver) AddClause(ps *GlucoVec[GlucoLit]) bool {
	panic("TODO, stubbed")
}

// AddEmptyClause is Solver::addEmptyClause.
// C: bool addEmptyClause()
func (s *GlucoSolver) AddEmptyClause() bool {
	panic("TODO, stubbed")
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *GlucoSolver) AddClauseUnit(p GlucoLit) bool {
	panic("TODO, stubbed")
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *GlucoSolver) AddClauseBinary(p GlucoLit, q GlucoLit) bool {
	panic("TODO, stubbed")
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *GlucoSolver) AddClauseTernary(p GlucoLit, q GlucoLit, r GlucoLit) bool {
	panic("TODO, stubbed")
}

// AddClause_ is Solver::addClause_ (no superfluous internal copy; clobbers ps).
// C: bool addClause_(vec<Lit>& ps)
func (s *GlucoSolver) AddClause_(ps *GlucoVec[GlucoLit]) bool {
	panic("TODO, stubbed")
}

// Simplify is Solver::simplify.
// C: bool simplify()
func (s *GlucoSolver) Simplify() bool {
	panic("TODO, stubbed")
}

// Solve is the overload solve(const vec<Lit>& assumps).
// C: bool solve(const vec<Lit>& assumps)
func (s *GlucoSolver) Solve(assumps *GlucoVec[GlucoLit]) bool {
	panic("TODO, stubbed")
}

// SolveLimited is Solver::solveLimited (with resource constraints).
// C: lbool solveLimited(const vec<Lit>& assumps)
func (s *GlucoSolver) SolveLimited(assumps *GlucoVec[GlucoLit]) GlucoLbool {
	panic("TODO, stubbed")
}

// SolveNoAssumps is the overload solve().
// C: bool solve()
func (s *GlucoSolver) SolveNoAssumps() bool {
	panic("TODO, stubbed")
}

// Solve1 is the overload solve(Lit p).
// C: bool solve(Lit p)
func (s *GlucoSolver) Solve1(p GlucoLit) bool {
	panic("TODO, stubbed")
}

// Solve2 is the overload solve(Lit p, Lit q).
// C: bool solve(Lit p, Lit q)
func (s *GlucoSolver) Solve2(p GlucoLit, q GlucoLit) bool {
	panic("TODO, stubbed")
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r).
// C: bool solve(Lit p, Lit q, Lit r)
func (s *GlucoSolver) Solve3(p GlucoLit, q GlucoLit, r GlucoLit) bool {
	panic("TODO, stubbed")
}

// Okay is Solver::okay.
// C: bool okay() const
func (s *GlucoSolver) Okay() bool {
	panic("TODO, stubbed")
}

// ToDimacsStream is the overload toDimacs(FILE* f, const vec<Lit>& assumps).
// C: void toDimacs(FILE* f, const vec<Lit>& assumps)
func (s *GlucoSolver) ToDimacsStream(f *os.File, assumps *GlucoVec[GlucoLit]) {
	panic("TODO, stubbed")
}

// ToDimacsFile is the overload toDimacs(const char *file, const vec<Lit>& assumps).
// C: void toDimacs(const char *file, const vec<Lit>& assumps)
func (s *GlucoSolver) ToDimacsFile(file string, assumps *GlucoVec[GlucoLit]) {
	panic("TODO, stubbed")
}

// ToDimacsClause is the overload toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max).
// C: void toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max)
func (s *GlucoSolver) ToDimacsClause(f *os.File, c *GlucoClause, map_ *GlucoVec[GlucoVar], max *GlucoVar) {
	panic("TODO, stubbed")
}

// ToDimacsFileNoAssumps is the overload toDimacs(const char* file).
// C: void toDimacs(const char* file)
func (s *GlucoSolver) ToDimacsFileNoAssumps(file string) {
	panic("TODO, stubbed")
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *GlucoSolver) ToDimacsFile1(file string, p GlucoLit) {
	panic("TODO, stubbed")
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *GlucoSolver) ToDimacsFile2(file string, p GlucoLit, q GlucoLit) {
	panic("TODO, stubbed")
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *GlucoSolver) ToDimacsFile3(file string, p GlucoLit, q GlucoLit, r GlucoLit) {
	panic("TODO, stubbed")
}

// Sat_solver_set_var_fanin_lit is Solver::sat_solver_set_var_fanin_lit.
// C: void sat_solver_set_var_fanin_lit(int, int, int)
func (s *GlucoSolver) Sat_solver_set_var_fanin_lit(v int, lit0 int, lit1 int) {
	panic("TODO, stubbed")
}

// Sat_solver_start_new_round is Solver::sat_solver_start_new_round.
// C: void sat_solver_start_new_round()
func (s *GlucoSolver) Sat_solver_start_new_round() {
	panic("TODO, stubbed")
}

// Sat_solver_mark_cone is Solver::sat_solver_mark_cone.
// C: void sat_solver_mark_cone(int)
func (s *GlucoSolver) Sat_solver_mark_cone(v int) {
	panic("TODO, stubbed")
}

// AddVar is Solver::addVar (adds enough variables to make sure there is variable v).
// C: void addVar(Var v)
func (s *GlucoSolver) AddVar(v GlucoVar) {
	panic("TODO, stubbed")
}

// PrintLit is Solver::printLit.
// C: void printLit(Lit l)
func (s *GlucoSolver) PrintLit(l GlucoLit) {
	panic("TODO, stubbed")
}

// PrintClause is Solver::printClause.
// C: void printClause(CRef c)
func (s *GlucoSolver) PrintClause(c GlucoCRef) {
	panic("TODO, stubbed")
}

// PrintInitialClause is Solver::printInitialClause.
// C: void printInitialClause(CRef c)
func (s *GlucoSolver) PrintInitialClause(c GlucoCRef) {
	panic("TODO, stubbed")
}

// GetCex is Solver::getCex.
// C: int * getCex() const
func (s *GlucoSolver) GetCex() []int {
	panic("TODO, stubbed")
}

// SetIncrementalMode is Solver::setIncrementalMode.
// C: void setIncrementalMode()
func (s *GlucoSolver) SetIncrementalMode() {
	panic("TODO, stubbed")
}

// InitNbInitialVars is Solver::initNbInitialVars.
// C: void initNbInitialVars(int nb)
func (s *GlucoSolver) InitNbInitialVars(nb int) {
	panic("TODO, stubbed")
}

// PrintIncrementalStats is Solver::printIncrementalStats.
// C: void printIncrementalStats()
func (s *GlucoSolver) PrintIncrementalStats() {
	panic("TODO, stubbed")
}

// Reset is Solver::reset (virtual).
// C: virtual void reset()
func (s *GlucoSolver) Reset() {
	panic("TODO, stubbed")
}

// SetPolarity is Solver::setPolarity.
// C: void setPolarity(Var v, bool b)
func (s *GlucoSolver) SetPolarity(v GlucoVar, b bool) {
	panic("TODO, stubbed")
}

// SetDecisionVar is Solver::setDecisionVar.
// C: void setDecisionVar(Var v, bool b)
func (s *GlucoSolver) SetDecisionVar(v GlucoVar, b bool) {
	panic("TODO, stubbed")
}

// Value is the overload value(Var x).
// C: lbool value(Var x) const
func (s *GlucoSolver) Value(x GlucoVar) GlucoLbool {
	panic("TODO, stubbed")
}

// ValueLit is the overload value(Lit p).
// C: lbool value(Lit p) const
func (s *GlucoSolver) ValueLit(p GlucoLit) GlucoLbool {
	panic("TODO, stubbed")
}

// ModelValue is the overload modelValue(Var x).
// C: lbool modelValue(Var x) const
func (s *GlucoSolver) ModelValue(x GlucoVar) GlucoLbool {
	panic("TODO, stubbed")
}

// ModelValueLit is the overload modelValue(Lit p).
// C: lbool modelValue(Lit p) const
func (s *GlucoSolver) ModelValueLit(p GlucoLit) GlucoLbool {
	panic("TODO, stubbed")
}

// NAssigns is Solver::nAssigns.
// C: int nAssigns() const
func (s *GlucoSolver) NAssigns() int {
	panic("TODO, stubbed")
}

// NClauses is Solver::nClauses.
// C: int nClauses() const
func (s *GlucoSolver) NClauses() int {
	panic("TODO, stubbed")
}

// NLearnts is Solver::nLearnts.
// C: int nLearnts() const
func (s *GlucoSolver) NLearnts() int {
	panic("TODO, stubbed")
}

// NVars is Solver::nVars.
// C: int nVars() const
func (s *GlucoSolver) NVars() int {
	panic("TODO, stubbed")
}

// NFreeVars is Solver::nFreeVars.
// C: int nFreeVars() const
func (s *GlucoSolver) NFreeVars() int {
	panic("TODO, stubbed")
}

// SetConfBudget is Solver::setConfBudget.
// C: void setConfBudget(int64_t x)
func (s *GlucoSolver) SetConfBudget(x int64) {
	panic("TODO, stubbed")
}

// SetPropBudget is Solver::setPropBudget.
// C: void setPropBudget(int64_t x)
func (s *GlucoSolver) SetPropBudget(x int64) {
	panic("TODO, stubbed")
}

// BudgetOff is Solver::budgetOff.
// C: void budgetOff()
func (s *GlucoSolver) BudgetOff() {
	panic("TODO, stubbed")
}

// Interrupt is Solver::interrupt (may be called asynchronously).
// C: void interrupt()
func (s *GlucoSolver) Interrupt() {
	panic("TODO, stubbed")
}

// ClearInterrupt is Solver::clearInterrupt.
// C: void clearInterrupt()
func (s *GlucoSolver) ClearInterrupt() {
	panic("TODO, stubbed")
}

// GarbageCollect is Solver::garbageCollect (virtual).
// C: virtual void garbageCollect()
func (s *GlucoSolver) GarbageCollect() {
	panic("TODO, stubbed")
}

// CheckGarbageFrac is the overload checkGarbage(double gf).
// C: void checkGarbage(double gf)
func (s *GlucoSolver) CheckGarbageFrac(gf float64) {
	panic("TODO, stubbed")
}

// CheckGarbage is the overload checkGarbage().
// C: void checkGarbage()
func (s *GlucoSolver) CheckGarbage() {
	panic("TODO, stubbed")
}

// glucoSolverDrand is Solver::drand (protected static inline).
// C: static inline double drand(double& seed)
func glucoSolverDrand(seed *float64) float64 {
	panic("TODO, stubbed")
}

// glucoSolverIrand is Solver::irand (protected static inline).
// C: static inline int irand(double& seed, int size)
func glucoSolverIrand(seed *float64, size int) int {
	panic("TODO, stubbed")
}

// insertVarOrder is Solver::insertVarOrder (protected).
// C: void insertVarOrder(Var x)
func (s *GlucoSolver) insertVarOrder(x GlucoVar) {
	panic("TODO, stubbed")
}

// pickBranchLit is Solver::pickBranchLit (protected).
// C: Lit pickBranchLit()
func (s *GlucoSolver) pickBranchLit() GlucoLit {
	panic("TODO, stubbed")
}

// newDecisionLevel is Solver::newDecisionLevel (protected).
// C: void newDecisionLevel()
func (s *GlucoSolver) newDecisionLevel() {
	panic("TODO, stubbed")
}

// uncheckedEnqueue is Solver::uncheckedEnqueue (protected; from defaults to CRef_Undef).
// C: void uncheckedEnqueue(Lit p, CRef from = CRef_Undef)
func (s *GlucoSolver) uncheckedEnqueue(p GlucoLit, from GlucoCRef) {
	panic("TODO, stubbed")
}

// enqueue is Solver::enqueue (protected; from defaults to CRef_Undef).
// C: bool enqueue(Lit p, CRef from = CRef_Undef)
func (s *GlucoSolver) enqueue(p GlucoLit, from GlucoCRef) bool {
	panic("TODO, stubbed")
}

// propagate is Solver::propagate (protected).
// C: CRef propagate()
func (s *GlucoSolver) propagate() GlucoCRef {
	panic("TODO, stubbed")
}

// cancelUntil is Solver::cancelUntil (protected).
// C: void cancelUntil(int level)
func (s *GlucoSolver) cancelUntil(level int) {
	panic("TODO, stubbed")
}

// analyze is Solver::analyze (protected).
// C: void analyze(CRef confl, vec<Lit>& out_learnt, vec<Lit> & selectors, int& out_btlevel, unsigned int &nblevels, unsigned int &szWithoutSelectors)
func (s *GlucoSolver) analyze(confl GlucoCRef, outLearnt *GlucoVec[GlucoLit], selectors *GlucoVec[GlucoLit], outBtlevel *int, nblevels *uint32, szWithoutSelectors *uint32) {
	panic("TODO, stubbed")
}

// analyzeFinal is Solver::analyzeFinal (protected).
// C: void analyzeFinal(Lit p, vec<Lit>& out_conflict)
func (s *GlucoSolver) analyzeFinal(p GlucoLit, outConflict *GlucoVec[GlucoLit]) {
	panic("TODO, stubbed")
}

// litRedundant is Solver::litRedundant (protected).
// C: bool litRedundant(Lit p, uint32_t abstract_levels)
func (s *GlucoSolver) litRedundant(p GlucoLit, abstractLevels uint32) bool {
	panic("TODO, stubbed")
}

// search is Solver::search (protected).
// C: lbool search(int nof_conflicts)
func (s *GlucoSolver) search(nofConflicts int) GlucoLbool {
	panic("TODO, stubbed")
}

// solve_ is Solver::solve_ (protected).
// C: lbool solve_()
func (s *GlucoSolver) solve_() GlucoLbool {
	panic("TODO, stubbed")
}

// reduceDB is Solver::reduceDB (protected).
// C: void reduceDB()
func (s *GlucoSolver) reduceDB() {
	panic("TODO, stubbed")
}

// removeSatisfied is Solver::removeSatisfied (protected).
// C: void removeSatisfied(vec<CRef>& cs)
func (s *GlucoSolver) removeSatisfied(cs *GlucoVec[GlucoCRef]) {
	panic("TODO, stubbed")
}

// rebuildOrderHeap is Solver::rebuildOrderHeap (protected).
// C: void rebuildOrderHeap()
func (s *GlucoSolver) rebuildOrderHeap() {
	panic("TODO, stubbed")
}

// varDecayActivity is Solver::varDecayActivity (protected).
// C: void varDecayActivity()
func (s *GlucoSolver) varDecayActivity() {
	panic("TODO, stubbed")
}

// varBumpActivityInc is the overload varBumpActivity(Var v, double inc) (protected).
// C: void varBumpActivity(Var v, double inc)
func (s *GlucoSolver) varBumpActivityInc(v GlucoVar, inc float64) {
	panic("TODO, stubbed")
}

// varBumpActivity is the overload varBumpActivity(Var v) (protected).
// C: void varBumpActivity(Var v)
func (s *GlucoSolver) varBumpActivity(v GlucoVar) {
	panic("TODO, stubbed")
}

// claDecayActivity is Solver::claDecayActivity (protected).
// C: void claDecayActivity()
func (s *GlucoSolver) claDecayActivity() {
	panic("TODO, stubbed")
}

// claBumpActivity is Solver::claBumpActivity (protected).
// C: void claBumpActivity(Clause& c)
func (s *GlucoSolver) claBumpActivity(c *GlucoClause) {
	panic("TODO, stubbed")
}

// attachClause is Solver::attachClause (protected).
// C: void attachClause(CRef cr)
func (s *GlucoSolver) attachClause(cr GlucoCRef) {
	panic("TODO, stubbed")
}

// detachClause is Solver::detachClause (protected; strict defaults to false).
// C: void detachClause(CRef cr, bool strict = false)
func (s *GlucoSolver) detachClause(cr GlucoCRef, strict bool) {
	panic("TODO, stubbed")
}

// removeClause is Solver::removeClause (protected).
// C: void removeClause(CRef cr)
func (s *GlucoSolver) removeClause(cr GlucoCRef) {
	panic("TODO, stubbed")
}

// locked is Solver::locked (protected).
// C: bool locked(const Clause& c) const
func (s *GlucoSolver) locked(c *GlucoClause) bool {
	panic("TODO, stubbed")
}

// satisfied is Solver::satisfied (protected).
// C: bool satisfied(const Clause& c) const
func (s *GlucoSolver) satisfied(c *GlucoClause) bool {
	panic("TODO, stubbed")
}

// computeLBD is the overload computeLBD(const vec<Lit>& lits, int end = -1) (protected).
// C: unsigned int computeLBD(const vec<Lit>& lits, int end = -1)
func (s *GlucoSolver) computeLBD(lits *GlucoVec[GlucoLit], end int) uint32 {
	panic("TODO, stubbed")
}

// computeLBDClause is the overload computeLBD(const Clause& c) (protected).
// C: unsigned int computeLBD(const Clause& c)
func (s *GlucoSolver) computeLBDClause(c *GlucoClause) uint32 {
	panic("TODO, stubbed")
}

// minimisationWithBinaryResolution is Solver::minimisationWithBinaryResolution (protected).
// C: void minimisationWithBinaryResolution(vec<Lit>& out_learnt)
func (s *GlucoSolver) minimisationWithBinaryResolution(outLearnt *GlucoVec[GlucoLit]) {
	panic("TODO, stubbed")
}

// isSelector is Solver::isSelector (protected inline).
// C: inline bool isSelector(Var v)
func (s *GlucoSolver) isSelector(v GlucoVar) bool {
	panic("TODO, stubbed")
}

// relocAll is Solver::relocAll (protected).
// C: void relocAll(ClauseAllocator& to)
func (s *GlucoSolver) relocAll(to *GlucoClauseAllocator) {
	panic("TODO, stubbed")
}

// decisionLevel is Solver::decisionLevel (protected).
// C: int decisionLevel() const
func (s *GlucoSolver) decisionLevel() int {
	panic("TODO, stubbed")
}

// abstractLevel is Solver::abstractLevel (protected).
// C: uint32_t abstractLevel(Var x) const
func (s *GlucoSolver) abstractLevel(x GlucoVar) uint32 {
	panic("TODO, stubbed")
}

// reason is Solver::reason (protected).
// C: CRef reason(Var x) const
func (s *GlucoSolver) reason(x GlucoVar) GlucoCRef {
	panic("TODO, stubbed")
}

// level is Solver::level (protected).
// C: int level(Var x) const
func (s *GlucoSolver) level(x GlucoVar) int {
	panic("TODO, stubbed")
}

// progressEstimate is Solver::progressEstimate (protected).
// C: double progressEstimate() const
func (s *GlucoSolver) progressEstimate() float64 {
	panic("TODO, stubbed")
}

// withinBudget is Solver::withinBudget (protected).
// C: bool withinBudget() const
func (s *GlucoSolver) withinBudget() bool {
	panic("TODO, stubbed")
}
