package abc

import "os"

// Hand-checked skeleton of src/sat/bsat2/Solver.h (namespace Minisat).
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

// MinisatSolverVarData is Solver::VarData.
type MinisatSolverVarData struct {
	reason MinisatCRef
	level  int
}

// minisatSolverMkVarData is Solver::mkVarData (protected static inline).
// C: static inline VarData mkVarData(CRef cr, int l)
func minisatSolverMkVarData(cr MinisatCRef, l int) MinisatSolverVarData {
	panic("TODO, stubbed")
}

// MinisatSolverWatcher is Solver::Watcher.
type MinisatSolverWatcher struct {
	cref    MinisatCRef
	blocker MinisatLit
}

// NewMinisatSolverWatcher is Watcher(CRef cr, Lit p).
// C: Watcher(CRef cr, Lit p)
func NewMinisatSolverWatcher(cr MinisatCRef, p MinisatLit) MinisatSolverWatcher {
	panic("TODO, stubbed")
}

// Eq is Watcher::operator==.
// C: bool operator==(const Watcher& w) const
func (w MinisatSolverWatcher) Eq(o MinisatSolverWatcher) bool {
	panic("TODO, stubbed")
}

// Ne is Watcher::operator!=.
// C: bool operator!=(const Watcher& w) const
func (w MinisatSolverWatcher) Ne(o MinisatSolverWatcher) bool {
	panic("TODO, stubbed")
}

// MinisatSolverWatcherDeleted is Solver::WatcherDeleted.
type MinisatSolverWatcherDeleted struct {
	ca *MinisatClauseAllocator // const ClauseAllocator&
}

// NewMinisatSolverWatcherDeleted is WatcherDeleted(const ClauseAllocator& _ca).
// C: WatcherDeleted(const ClauseAllocator& _ca)
func NewMinisatSolverWatcherDeleted(ca *MinisatClauseAllocator) MinisatSolverWatcherDeleted {
	panic("TODO, stubbed")
}

// Call is WatcherDeleted::operator().
// C: bool operator()(const Watcher& w) const
func (d MinisatSolverWatcherDeleted) Call(w MinisatSolverWatcher) bool {
	panic("TODO, stubbed")
}

// MinisatSolverVarOrderLt is Solver::VarOrderLt.
type MinisatSolverVarOrderLt struct {
	activity *MinisatVec[float64] // const vec<double>&
}

// NewMinisatSolverVarOrderLt is VarOrderLt(const vec<double>& act).
// C: VarOrderLt(const vec<double>& act)
func NewMinisatSolverVarOrderLt(act *MinisatVec[float64]) MinisatSolverVarOrderLt {
	panic("TODO, stubbed")
}

// Call is VarOrderLt::operator().
// C: bool operator()(Var x, Var y) const
func (l MinisatSolverVarOrderLt) Call(x int, y int) bool {
	panic("TODO, stubbed")
}

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
}

// NewMinisatSolver is Solver::Solver().
// C: Solver::Solver()
func NewMinisatSolver() *MinisatSolver {
	panic("TODO, stubbed")
}

// Destroy is Solver::~Solver() (virtual).
// C: virtual Solver::~Solver()
func (s *MinisatSolver) Destroy() {
	panic("TODO, stubbed")
}

// NewVar is Solver::newVar (polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *MinisatSolver) NewVar(polarity bool, dvar bool) MinisatVar {
	panic("TODO, stubbed")
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *MinisatSolver) AddClause(ps *MinisatVec[MinisatLit]) bool {
	panic("TODO, stubbed")
}

// AddEmptyClause is Solver::addEmptyClause.
// C: bool addEmptyClause()
func (s *MinisatSolver) AddEmptyClause() bool {
	panic("TODO, stubbed")
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *MinisatSolver) AddClauseUnit(p MinisatLit) bool {
	panic("TODO, stubbed")
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *MinisatSolver) AddClauseBinary(p MinisatLit, q MinisatLit) bool {
	panic("TODO, stubbed")
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *MinisatSolver) AddClauseTernary(p MinisatLit, q MinisatLit, r MinisatLit) bool {
	panic("TODO, stubbed")
}

// AddClause_ is Solver::addClause_ (no superfluous internal copy; clobbers ps).
// C: bool addClause_(vec<Lit>& ps)
func (s *MinisatSolver) AddClause_(ps *MinisatVec[MinisatLit]) bool {
	panic("TODO, stubbed")
}

// Simplify is Solver::simplify.
// C: bool simplify()
func (s *MinisatSolver) Simplify() bool {
	panic("TODO, stubbed")
}

// Solve is the overload solve(const vec<Lit>& assumps).
// C: bool solve(const vec<Lit>& assumps)
func (s *MinisatSolver) Solve(assumps *MinisatVec[MinisatLit]) bool {
	panic("TODO, stubbed")
}

// SolveLimited is Solver::solveLimited (with resource constraints).
// C: lbool solveLimited(const vec<Lit>& assumps)
func (s *MinisatSolver) SolveLimited(assumps *MinisatVec[MinisatLit]) MinisatLbool {
	panic("TODO, stubbed")
}

// SolveNoAssumps is the overload solve().
// C: bool solve()
func (s *MinisatSolver) SolveNoAssumps() bool {
	panic("TODO, stubbed")
}

// Solve1 is the overload solve(Lit p).
// C: bool solve(Lit p)
func (s *MinisatSolver) Solve1(p MinisatLit) bool {
	panic("TODO, stubbed")
}

// Solve2 is the overload solve(Lit p, Lit q).
// C: bool solve(Lit p, Lit q)
func (s *MinisatSolver) Solve2(p MinisatLit, q MinisatLit) bool {
	panic("TODO, stubbed")
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r).
// C: bool solve(Lit p, Lit q, Lit r)
func (s *MinisatSolver) Solve3(p MinisatLit, q MinisatLit, r MinisatLit) bool {
	panic("TODO, stubbed")
}

// Okay is Solver::okay.
// C: bool okay() const
func (s *MinisatSolver) Okay() bool {
	panic("TODO, stubbed")
}

// ToDimacsStream is the overload toDimacs(FILE* f, const vec<Lit>& assumps).
// C: void toDimacs(FILE* f, const vec<Lit>& assumps)
func (s *MinisatSolver) ToDimacsStream(f *os.File, assumps *MinisatVec[MinisatLit]) {
	panic("TODO, stubbed")
}

// ToDimacsFile is the overload toDimacs(const char *file, const vec<Lit>& assumps).
// C: void toDimacs(const char *file, const vec<Lit>& assumps)
func (s *MinisatSolver) ToDimacsFile(file string, assumps *MinisatVec[MinisatLit]) {
	panic("TODO, stubbed")
}

// ToDimacsClause is the overload toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max).
// C: void toDimacs(FILE* f, Clause& c, vec<Var>& map, Var& max)
func (s *MinisatSolver) ToDimacsClause(f *os.File, c *MinisatClause, map_ *MinisatVec[MinisatVar], max *MinisatVar) {
	panic("TODO, stubbed")
}

// ToDimacsFileNoAssumps is the overload toDimacs(const char* file).
// C: void toDimacs(const char* file)
func (s *MinisatSolver) ToDimacsFileNoAssumps(file string) {
	panic("TODO, stubbed")
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *MinisatSolver) ToDimacsFile1(file string, p MinisatLit) {
	panic("TODO, stubbed")
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *MinisatSolver) ToDimacsFile2(file string, p MinisatLit, q MinisatLit) {
	panic("TODO, stubbed")
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *MinisatSolver) ToDimacsFile3(file string, p MinisatLit, q MinisatLit, r MinisatLit) {
	panic("TODO, stubbed")
}

// SetPolarity is Solver::setPolarity.
// C: void setPolarity(Var v, bool b)
func (s *MinisatSolver) SetPolarity(v MinisatVar, b bool) {
	panic("TODO, stubbed")
}

// SetDecisionVar is Solver::setDecisionVar.
// C: void setDecisionVar(Var v, bool b)
func (s *MinisatSolver) SetDecisionVar(v MinisatVar, b bool) {
	panic("TODO, stubbed")
}

// Value is the overload value(Var x).
// C: lbool value(Var x) const
func (s *MinisatSolver) Value(x MinisatVar) MinisatLbool {
	panic("TODO, stubbed")
}

// ValueLit is the overload value(Lit p).
// C: lbool value(Lit p) const
func (s *MinisatSolver) ValueLit(p MinisatLit) MinisatLbool {
	panic("TODO, stubbed")
}

// ModelValue is the overload modelValue(Var x).
// C: lbool modelValue(Var x) const
func (s *MinisatSolver) ModelValue(x MinisatVar) MinisatLbool {
	panic("TODO, stubbed")
}

// ModelValueLit is the overload modelValue(Lit p).
// C: lbool modelValue(Lit p) const
func (s *MinisatSolver) ModelValueLit(p MinisatLit) MinisatLbool {
	panic("TODO, stubbed")
}

// NAssigns is Solver::nAssigns.
// C: int nAssigns() const
func (s *MinisatSolver) NAssigns() int {
	panic("TODO, stubbed")
}

// NClauses is Solver::nClauses.
// C: int nClauses() const
func (s *MinisatSolver) NClauses() int {
	panic("TODO, stubbed")
}

// NLearnts is Solver::nLearnts.
// C: int nLearnts() const
func (s *MinisatSolver) NLearnts() int {
	panic("TODO, stubbed")
}

// NVars is Solver::nVars.
// C: int nVars() const
func (s *MinisatSolver) NVars() int {
	panic("TODO, stubbed")
}

// NFreeVars is Solver::nFreeVars.
// C: int nFreeVars() const
func (s *MinisatSolver) NFreeVars() int {
	panic("TODO, stubbed")
}

// SetConfBudget is Solver::setConfBudget.
// C: void setConfBudget(int64_t x)
func (s *MinisatSolver) SetConfBudget(x int64) {
	panic("TODO, stubbed")
}

// SetPropBudget is Solver::setPropBudget.
// C: void setPropBudget(int64_t x)
func (s *MinisatSolver) SetPropBudget(x int64) {
	panic("TODO, stubbed")
}

// BudgetOff is Solver::budgetOff.
// C: void budgetOff()
func (s *MinisatSolver) BudgetOff() {
	panic("TODO, stubbed")
}

// Interrupt is Solver::interrupt (may be called asynchronously).
// C: void interrupt()
func (s *MinisatSolver) Interrupt() {
	panic("TODO, stubbed")
}

// ClearInterrupt is Solver::clearInterrupt.
// C: void clearInterrupt()
func (s *MinisatSolver) ClearInterrupt() {
	panic("TODO, stubbed")
}

// GarbageCollect is Solver::garbageCollect (virtual).
// C: virtual void garbageCollect()
func (s *MinisatSolver) GarbageCollect() {
	panic("TODO, stubbed")
}

// CheckGarbageFrac is the overload checkGarbage(double gf).
// C: void checkGarbage(double gf)
func (s *MinisatSolver) CheckGarbageFrac(gf float64) {
	panic("TODO, stubbed")
}

// CheckGarbage is the overload checkGarbage().
// C: void checkGarbage()
func (s *MinisatSolver) CheckGarbage() {
	panic("TODO, stubbed")
}

// minisatSolverDrand is Solver::drand (protected static inline).
// C: static inline double drand(double& seed)
func minisatSolverDrand(seed *float64) float64 {
	panic("TODO, stubbed")
}

// minisatSolverIrand is Solver::irand (protected static inline).
// C: static inline int irand(double& seed, int size)
func minisatSolverIrand(seed *float64, size int) int {
	panic("TODO, stubbed")
}

// insertVarOrder is Solver::insertVarOrder (protected).
// C: void insertVarOrder(Var x)
func (s *MinisatSolver) insertVarOrder(x MinisatVar) {
	panic("TODO, stubbed")
}

// pickBranchLit is Solver::pickBranchLit (protected).
// C: Lit pickBranchLit()
func (s *MinisatSolver) pickBranchLit() MinisatLit {
	panic("TODO, stubbed")
}

// newDecisionLevel is Solver::newDecisionLevel (protected).
// C: void newDecisionLevel()
func (s *MinisatSolver) newDecisionLevel() {
	panic("TODO, stubbed")
}

// uncheckedEnqueue is Solver::uncheckedEnqueue (protected; from defaults to CRef_Undef).
// C: void uncheckedEnqueue(Lit p, CRef from = CRef_Undef)
func (s *MinisatSolver) uncheckedEnqueue(p MinisatLit, from MinisatCRef) {
	panic("TODO, stubbed")
}

// enqueue is Solver::enqueue (protected; from defaults to CRef_Undef).
// C: bool enqueue(Lit p, CRef from = CRef_Undef)
func (s *MinisatSolver) enqueue(p MinisatLit, from MinisatCRef) bool {
	panic("TODO, stubbed")
}

// propagate is Solver::propagate (protected).
// C: CRef propagate()
func (s *MinisatSolver) propagate() MinisatCRef {
	panic("TODO, stubbed")
}

// cancelUntil is Solver::cancelUntil (protected).
// C: void cancelUntil(int level)
func (s *MinisatSolver) cancelUntil(level int) {
	panic("TODO, stubbed")
}

// analyze is Solver::analyze (protected).
// C: void analyze(CRef confl, vec<Lit>& out_learnt, int& out_btlevel)
func (s *MinisatSolver) analyze(confl MinisatCRef, outLearnt *MinisatVec[MinisatLit], outBtlevel *int) {
	panic("TODO, stubbed")
}

// analyzeFinal is Solver::analyzeFinal (protected).
// C: void analyzeFinal(Lit p, vec<Lit>& out_conflict)
func (s *MinisatSolver) analyzeFinal(p MinisatLit, outConflict *MinisatVec[MinisatLit]) {
	panic("TODO, stubbed")
}

// litRedundant is Solver::litRedundant (protected).
// C: bool litRedundant(Lit p, uint32_t abstract_levels)
func (s *MinisatSolver) litRedundant(p MinisatLit, abstractLevels uint32) bool {
	panic("TODO, stubbed")
}

// search is Solver::search (protected).
// C: lbool search(int nof_conflicts)
func (s *MinisatSolver) search(nofConflicts int) MinisatLbool {
	panic("TODO, stubbed")
}

// solve_ is Solver::solve_ (protected).
// C: lbool solve_()
func (s *MinisatSolver) solve_() MinisatLbool {
	panic("TODO, stubbed")
}

// reduceDB is Solver::reduceDB (protected).
// C: void reduceDB()
func (s *MinisatSolver) reduceDB() {
	panic("TODO, stubbed")
}

// removeSatisfied is Solver::removeSatisfied (protected).
// C: void removeSatisfied(vec<CRef>& cs)
func (s *MinisatSolver) removeSatisfied(cs *MinisatVec[MinisatCRef]) {
	panic("TODO, stubbed")
}

// rebuildOrderHeap is Solver::rebuildOrderHeap (protected).
// C: void rebuildOrderHeap()
func (s *MinisatSolver) rebuildOrderHeap() {
	panic("TODO, stubbed")
}

// varDecayActivity is Solver::varDecayActivity (protected).
// C: void varDecayActivity()
func (s *MinisatSolver) varDecayActivity() {
	panic("TODO, stubbed")
}

// varBumpActivityInc is the overload varBumpActivity(Var v, double inc) (protected).
// C: void varBumpActivity(Var v, double inc)
func (s *MinisatSolver) varBumpActivityInc(v MinisatVar, inc float64) {
	panic("TODO, stubbed")
}

// varBumpActivity is the overload varBumpActivity(Var v) (protected).
// C: void varBumpActivity(Var v)
func (s *MinisatSolver) varBumpActivity(v MinisatVar) {
	panic("TODO, stubbed")
}

// claDecayActivity is Solver::claDecayActivity (protected).
// C: void claDecayActivity()
func (s *MinisatSolver) claDecayActivity() {
	panic("TODO, stubbed")
}

// claBumpActivity is Solver::claBumpActivity (protected).
// C: void claBumpActivity(Clause& c)
func (s *MinisatSolver) claBumpActivity(c *MinisatClause) {
	panic("TODO, stubbed")
}

// attachClause is Solver::attachClause (protected).
// C: void attachClause(CRef cr)
func (s *MinisatSolver) attachClause(cr MinisatCRef) {
	panic("TODO, stubbed")
}

// detachClause is Solver::detachClause (protected; strict defaults to false).
// C: void detachClause(CRef cr, bool strict = false)
func (s *MinisatSolver) detachClause(cr MinisatCRef, strict bool) {
	panic("TODO, stubbed")
}

// removeClause is Solver::removeClause (protected).
// C: void removeClause(CRef cr)
func (s *MinisatSolver) removeClause(cr MinisatCRef) {
	panic("TODO, stubbed")
}

// locked is Solver::locked (protected).
// C: bool locked(const Clause& c) const
func (s *MinisatSolver) locked(c *MinisatClause) bool {
	panic("TODO, stubbed")
}

// satisfied is Solver::satisfied (protected).
// C: bool satisfied(const Clause& c) const
func (s *MinisatSolver) satisfied(c *MinisatClause) bool {
	panic("TODO, stubbed")
}

// relocAll is Solver::relocAll (protected).
// C: void relocAll(ClauseAllocator& to)
func (s *MinisatSolver) relocAll(to *MinisatClauseAllocator) {
	panic("TODO, stubbed")
}

// decisionLevel is Solver::decisionLevel (protected).
// C: int decisionLevel() const
func (s *MinisatSolver) decisionLevel() int {
	panic("TODO, stubbed")
}

// abstractLevel is Solver::abstractLevel (protected).
// C: uint32_t abstractLevel(Var x) const
func (s *MinisatSolver) abstractLevel(x MinisatVar) uint32 {
	panic("TODO, stubbed")
}

// reason is Solver::reason (protected).
// C: CRef reason(Var x) const
func (s *MinisatSolver) reason(x MinisatVar) MinisatCRef {
	panic("TODO, stubbed")
}

// level is Solver::level (protected).
// C: int level(Var x) const
func (s *MinisatSolver) level(x MinisatVar) int {
	panic("TODO, stubbed")
}

// progressEstimate is Solver::progressEstimate (protected).
// C: double progressEstimate() const
func (s *MinisatSolver) progressEstimate() float64 {
	panic("TODO, stubbed")
}

// withinBudget is Solver::withinBudget (protected).
// C: bool withinBudget() const
func (s *MinisatSolver) withinBudget() bool {
	panic("TODO, stubbed")
}
