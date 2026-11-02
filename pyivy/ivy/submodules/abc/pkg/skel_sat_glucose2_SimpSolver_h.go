package abc

// Hand-checked skeleton of src/sat/glucose2/SimpSolver.h (namespace Gluco2).
// (Under CGLUCOSE_EXP the Solver additionally has circuit justification, see Solver.h.)
// SimpSolver derives from Solver: it embeds Gluco2Solver and re-declares the
// overridden methods (NewVar, AddClause*, Solve*, ToDimacs*, GarbageCollect, ...).

// Gluco2SimpSolverElimLt is SimpSolver::ElimLt.
type Gluco2SimpSolverElimLt struct {
	n_occ *Gluco2Vec[int] // const vec<int>&
}

// NewGluco2SimpSolverElimLt is the explicit constructor ElimLt(const vec<int>& no).
// C: explicit ElimLt(const vec<int>& no)
func NewGluco2SimpSolverElimLt(no *Gluco2Vec[int]) Gluco2SimpSolverElimLt {
	panic("TODO, stubbed")
}

// cost is ElimLt::cost.
// C: uint64_t cost(Var x) const
func (e Gluco2SimpSolverElimLt) cost(x Gluco2Var) uint64 {
	panic("TODO, stubbed")
}

// Call is ElimLt::operator().
// C: bool operator()(Var x, Var y) const
func (e Gluco2SimpSolverElimLt) Call(x int, y int) bool {
	panic("TODO, stubbed")
}

// Gluco2SimpSolverClauseDeleted is SimpSolver::ClauseDeleted.
type Gluco2SimpSolverClauseDeleted struct {
	ca *Gluco2ClauseAllocator // const ClauseAllocator&
}

// NewGluco2SimpSolverClauseDeleted is the explicit constructor ClauseDeleted(const ClauseAllocator& _ca).
// C: explicit ClauseDeleted(const ClauseAllocator& _ca)
func NewGluco2SimpSolverClauseDeleted(ca *Gluco2ClauseAllocator) Gluco2SimpSolverClauseDeleted {
	panic("TODO, stubbed")
}

// Call is ClauseDeleted::operator().
// C: bool operator()(const CRef& cr) const
func (d Gluco2SimpSolverClauseDeleted) Call(cr Gluco2CRef) bool {
	panic("TODO, stubbed")
}

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
	panic("TODO, stubbed")
}

// Destroy is SimpSolver::~SimpSolver().
// C: SimpSolver::~SimpSolver()
func (s *Gluco2SimpSolver) Destroy() {
	panic("TODO, stubbed")
}

// NewVar is SimpSolver::newVar (override; polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *Gluco2SimpSolver) NewVar(polarity bool, dvar bool) Gluco2Var {
	panic("TODO, stubbed")
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *Gluco2SimpSolver) AddClause(ps *Gluco2Vec[Gluco2Lit]) bool {
	panic("TODO, stubbed")
}

// AddEmptyClause is SimpSolver::addEmptyClause.
// C: bool addEmptyClause()
func (s *Gluco2SimpSolver) AddEmptyClause() bool {
	panic("TODO, stubbed")
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *Gluco2SimpSolver) AddClauseUnit(p Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *Gluco2SimpSolver) AddClauseBinary(p Gluco2Lit, q Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *Gluco2SimpSolver) AddClauseTernary(p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// AddClause_ is SimpSolver::addClause_.
// C: bool addClause_(vec<Lit>& ps)
func (s *Gluco2SimpSolver) AddClause_(ps *Gluco2Vec[Gluco2Lit]) bool {
	panic("TODO, stubbed")
}

// AddVar is SimpSolver::addVar (adds enough variables to make sure there is variable v).
// C: void addVar(Var v)
func (s *Gluco2SimpSolver) AddVar(v Gluco2Var) {
	panic("TODO, stubbed")
}

// Substitute is SimpSolver::substitute (replace all occurrences of v with x; may cause a contradiction).
// C: bool substitute(Var v, Lit x)
func (s *Gluco2SimpSolver) Substitute(v Gluco2Var, x Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// SetFrozen is SimpSolver::setFrozen (a frozen variable is not eliminated).
// C: void setFrozen(Var v, bool b)
func (s *Gluco2SimpSolver) SetFrozen(v Gluco2Var, b bool) {
	panic("TODO, stubbed")
}

// IsEliminated is SimpSolver::isEliminated.
// C: bool isEliminated(Var v) const
func (s *Gluco2SimpSolver) IsEliminated(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// Solve is the overload solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) Solve(assumps *Gluco2Vec[Gluco2Lit], doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// SolveLimited is SimpSolver::solveLimited.
// C: lbool solveLimited(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) SolveLimited(assumps *Gluco2Vec[Gluco2Lit], doSimp bool, turnOffSimp bool) Gluco2Lbool {
	panic("TODO, stubbed")
}

// SolveNoAssumps is the overload solve(bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) SolveNoAssumps(doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve1 is the overload solve(Lit p, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) Solve1(p Gluco2Lit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve2 is the overload solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) Solve2(p Gluco2Lit, q Gluco2Lit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) Solve3(p Gluco2Lit, q Gluco2Lit, r Gluco2Lit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Reset is SimpSolver::reset (virtual override).
// C: virtual void reset()
func (s *Gluco2SimpSolver) Reset() {
	panic("TODO, stubbed")
}

// Eliminate is SimpSolver::eliminate (variable-elimination based simplification).
// C: bool eliminate(bool turn_off_elim = false)
func (s *Gluco2SimpSolver) Eliminate(turnOffElim bool) bool {
	panic("TODO, stubbed")
}

// GarbageCollect is SimpSolver::garbageCollect (override).
// C: virtual void garbageCollect()
func (s *Gluco2SimpSolver) GarbageCollect() {
	panic("TODO, stubbed")
}

// ToDimacsFile is the overload toDimacs(const char* file, const vec<Lit>& assumps).
// C: void toDimacs(const char* file, const vec<Lit>& assumps)
func (s *Gluco2SimpSolver) ToDimacsFile(file string, assumps *Gluco2Vec[Gluco2Lit]) {
	panic("TODO, stubbed")
}

// ToDimacsFileNoAssumps is the overload toDimacs(const char* file).
// C: void toDimacs(const char* file)
func (s *Gluco2SimpSolver) ToDimacsFileNoAssumps(file string) {
	panic("TODO, stubbed")
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *Gluco2SimpSolver) ToDimacsFile1(file string, p Gluco2Lit) {
	panic("TODO, stubbed")
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *Gluco2SimpSolver) ToDimacsFile2(file string, p Gluco2Lit, q Gluco2Lit) {
	panic("TODO, stubbed")
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *Gluco2SimpSolver) ToDimacsFile3(file string, p Gluco2Lit, q Gluco2Lit, r Gluco2Lit) {
	panic("TODO, stubbed")
}

// solve_ is SimpSolver::solve_ (protected; hides Solver::solve_).
// C: lbool solve_(bool do_simp = true, bool turn_off_simp = false)
func (s *Gluco2SimpSolver) solve_(doSimp bool, turnOffSimp bool) Gluco2Lbool {
	panic("TODO, stubbed")
}

// asymm is SimpSolver::asymm (protected).
// C: bool asymm(Var v, CRef cr)
func (s *Gluco2SimpSolver) asymm(v Gluco2Var, cr Gluco2CRef) bool {
	panic("TODO, stubbed")
}

// asymmVar is SimpSolver::asymmVar (protected).
// C: bool asymmVar(Var v)
func (s *Gluco2SimpSolver) asymmVar(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// updateElimHeap is SimpSolver::updateElimHeap (protected).
// C: void updateElimHeap(Var v)
func (s *Gluco2SimpSolver) updateElimHeap(v Gluco2Var) {
	panic("TODO, stubbed")
}

// gatherTouchedClauses is SimpSolver::gatherTouchedClauses (protected).
// C: void gatherTouchedClauses()
func (s *Gluco2SimpSolver) gatherTouchedClauses() {
	panic("TODO, stubbed")
}

// merge is the overload merge(..., vec<Lit>& out_clause) (protected).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, vec<Lit>& out_clause)
func (s *Gluco2SimpSolver) merge(_ps *Gluco2Clause, _qs *Gluco2Clause, v Gluco2Var, outClause *Gluco2Vec[Gluco2Lit]) bool {
	panic("TODO, stubbed")
}

// mergeSize is the overload merge(..., int& size) (protected).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, int& size)
func (s *Gluco2SimpSolver) mergeSize(_ps *Gluco2Clause, _qs *Gluco2Clause, v Gluco2Var, size *int) bool {
	panic("TODO, stubbed")
}

// backwardSubsumptionCheck is SimpSolver::backwardSubsumptionCheck (protected).
// C: bool backwardSubsumptionCheck(bool verbose = false)
func (s *Gluco2SimpSolver) backwardSubsumptionCheck(verbose bool) bool {
	panic("TODO, stubbed")
}

// eliminateVar is SimpSolver::eliminateVar (protected).
// C: bool eliminateVar(Var v)
func (s *Gluco2SimpSolver) eliminateVar(v Gluco2Var) bool {
	panic("TODO, stubbed")
}

// extendModel is SimpSolver::extendModel (protected).
// C: void extendModel()
func (s *Gluco2SimpSolver) extendModel() {
	panic("TODO, stubbed")
}

// removeClause is SimpSolver::removeClause (protected; hides Solver::removeClause).
// C: void removeClause(CRef cr)
func (s *Gluco2SimpSolver) removeClause(cr Gluco2CRef) {
	panic("TODO, stubbed")
}

// strengthenClause is SimpSolver::strengthenClause (protected).
// C: bool strengthenClause(CRef cr, Lit l)
func (s *Gluco2SimpSolver) strengthenClause(cr Gluco2CRef, l Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// cleanUpClauses is SimpSolver::cleanUpClauses (protected).
// C: void cleanUpClauses()
func (s *Gluco2SimpSolver) cleanUpClauses() {
	panic("TODO, stubbed")
}

// implied is SimpSolver::implied (protected).
// C: bool implied(const vec<Lit>& c)
func (s *Gluco2SimpSolver) implied(c *Gluco2Vec[Gluco2Lit]) bool {
	panic("TODO, stubbed")
}

// relocAll is SimpSolver::relocAll (protected; hides Solver::relocAll).
// C: void relocAll(ClauseAllocator& to)
func (s *Gluco2SimpSolver) relocAll(to *Gluco2ClauseAllocator) {
	panic("TODO, stubbed")
}
