package abc

// Hand-checked skeleton of src/sat/glucose/SimpSolver.h (namespace Gluco).
// SimpSolver derives from Solver: it embeds GlucoSolver and re-declares the
// overridden methods (NewVar, AddClause*, Solve*, ToDimacs*, GarbageCollect, ...).

// GlucoSimpSolverElimLt is SimpSolver::ElimLt.
type GlucoSimpSolverElimLt struct {
	n_occ *GlucoVec[int] // const vec<int>&
}

// NewGlucoSimpSolverElimLt is the explicit constructor ElimLt(const vec<int>& no).
// C: explicit ElimLt(const vec<int>& no)
func NewGlucoSimpSolverElimLt(no *GlucoVec[int]) GlucoSimpSolverElimLt {
	panic("TODO, stubbed")
}

// cost is ElimLt::cost.
// C: uint64_t cost(Var x) const
func (e GlucoSimpSolverElimLt) cost(x GlucoVar) uint64 {
	panic("TODO, stubbed")
}

// Call is ElimLt::operator().
// C: bool operator()(Var x, Var y) const
func (e GlucoSimpSolverElimLt) Call(x int, y int) bool {
	panic("TODO, stubbed")
}

// GlucoSimpSolverClauseDeleted is SimpSolver::ClauseDeleted.
type GlucoSimpSolverClauseDeleted struct {
	ca *GlucoClauseAllocator // const ClauseAllocator&
}

// NewGlucoSimpSolverClauseDeleted is the explicit constructor ClauseDeleted(const ClauseAllocator& _ca).
// C: explicit ClauseDeleted(const ClauseAllocator& _ca)
func NewGlucoSimpSolverClauseDeleted(ca *GlucoClauseAllocator) GlucoSimpSolverClauseDeleted {
	panic("TODO, stubbed")
}

// Call is ClauseDeleted::operator().
// C: bool operator()(const CRef& cr) const
func (d GlucoSimpSolverClauseDeleted) Call(cr GlucoCRef) bool {
	panic("TODO, stubbed")
}

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
	panic("TODO, stubbed")
}

// Destroy is SimpSolver::~SimpSolver().
// C: SimpSolver::~SimpSolver()
func (s *GlucoSimpSolver) Destroy() {
	panic("TODO, stubbed")
}

// NewVar is SimpSolver::newVar (override; polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *GlucoSimpSolver) NewVar(polarity bool, dvar bool) GlucoVar {
	panic("TODO, stubbed")
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *GlucoSimpSolver) AddClause(ps *GlucoVec[GlucoLit]) bool {
	panic("TODO, stubbed")
}

// AddEmptyClause is SimpSolver::addEmptyClause.
// C: bool addEmptyClause()
func (s *GlucoSimpSolver) AddEmptyClause() bool {
	panic("TODO, stubbed")
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *GlucoSimpSolver) AddClauseUnit(p GlucoLit) bool {
	panic("TODO, stubbed")
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *GlucoSimpSolver) AddClauseBinary(p GlucoLit, q GlucoLit) bool {
	panic("TODO, stubbed")
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *GlucoSimpSolver) AddClauseTernary(p GlucoLit, q GlucoLit, r GlucoLit) bool {
	panic("TODO, stubbed")
}

// AddClause_ is SimpSolver::addClause_.
// C: bool addClause_(vec<Lit>& ps)
func (s *GlucoSimpSolver) AddClause_(ps *GlucoVec[GlucoLit]) bool {
	panic("TODO, stubbed")
}

// AddVar is SimpSolver::addVar (adds enough variables to make sure there is variable v).
// C: void addVar(Var v)
func (s *GlucoSimpSolver) AddVar(v GlucoVar) {
	panic("TODO, stubbed")
}

// Substitute is SimpSolver::substitute (replace all occurrences of v with x; may cause a contradiction).
// C: bool substitute(Var v, Lit x)
func (s *GlucoSimpSolver) Substitute(v GlucoVar, x GlucoLit) bool {
	panic("TODO, stubbed")
}

// SetFrozen is SimpSolver::setFrozen (a frozen variable is not eliminated).
// C: void setFrozen(Var v, bool b)
func (s *GlucoSimpSolver) SetFrozen(v GlucoVar, b bool) {
	panic("TODO, stubbed")
}

// IsEliminated is SimpSolver::isEliminated.
// C: bool isEliminated(Var v) const
func (s *GlucoSimpSolver) IsEliminated(v GlucoVar) bool {
	panic("TODO, stubbed")
}

// Solve is the overload solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) Solve(assumps *GlucoVec[GlucoLit], doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// SolveLimited is SimpSolver::solveLimited.
// C: lbool solveLimited(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) SolveLimited(assumps *GlucoVec[GlucoLit], doSimp bool, turnOffSimp bool) GlucoLbool {
	panic("TODO, stubbed")
}

// SolveNoAssumps is the overload solve(bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) SolveNoAssumps(doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve1 is the overload solve(Lit p, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) Solve1(p GlucoLit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve2 is the overload solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) Solve2(p GlucoLit, q GlucoLit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) Solve3(p GlucoLit, q GlucoLit, r GlucoLit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Reset is SimpSolver::reset (virtual override).
// C: virtual void reset()
func (s *GlucoSimpSolver) Reset() {
	panic("TODO, stubbed")
}

// Eliminate is SimpSolver::eliminate (variable-elimination based simplification).
// C: bool eliminate(bool turn_off_elim = false)
func (s *GlucoSimpSolver) Eliminate(turnOffElim bool) bool {
	panic("TODO, stubbed")
}

// GarbageCollect is SimpSolver::garbageCollect (override).
// C: virtual void garbageCollect()
func (s *GlucoSimpSolver) GarbageCollect() {
	panic("TODO, stubbed")
}

// ToDimacsFile is the overload toDimacs(const char* file, const vec<Lit>& assumps).
// C: void toDimacs(const char* file, const vec<Lit>& assumps)
func (s *GlucoSimpSolver) ToDimacsFile(file string, assumps *GlucoVec[GlucoLit]) {
	panic("TODO, stubbed")
}

// ToDimacsFileNoAssumps is the overload toDimacs(const char* file).
// C: void toDimacs(const char* file)
func (s *GlucoSimpSolver) ToDimacsFileNoAssumps(file string) {
	panic("TODO, stubbed")
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *GlucoSimpSolver) ToDimacsFile1(file string, p GlucoLit) {
	panic("TODO, stubbed")
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *GlucoSimpSolver) ToDimacsFile2(file string, p GlucoLit, q GlucoLit) {
	panic("TODO, stubbed")
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *GlucoSimpSolver) ToDimacsFile3(file string, p GlucoLit, q GlucoLit, r GlucoLit) {
	panic("TODO, stubbed")
}

// solve_ is SimpSolver::solve_ (protected; hides Solver::solve_).
// C: lbool solve_(bool do_simp = true, bool turn_off_simp = false)
func (s *GlucoSimpSolver) solve_(doSimp bool, turnOffSimp bool) GlucoLbool {
	panic("TODO, stubbed")
}

// asymm is SimpSolver::asymm (protected).
// C: bool asymm(Var v, CRef cr)
func (s *GlucoSimpSolver) asymm(v GlucoVar, cr GlucoCRef) bool {
	panic("TODO, stubbed")
}

// asymmVar is SimpSolver::asymmVar (protected).
// C: bool asymmVar(Var v)
func (s *GlucoSimpSolver) asymmVar(v GlucoVar) bool {
	panic("TODO, stubbed")
}

// updateElimHeap is SimpSolver::updateElimHeap (protected).
// C: void updateElimHeap(Var v)
func (s *GlucoSimpSolver) updateElimHeap(v GlucoVar) {
	panic("TODO, stubbed")
}

// gatherTouchedClauses is SimpSolver::gatherTouchedClauses (protected).
// C: void gatherTouchedClauses()
func (s *GlucoSimpSolver) gatherTouchedClauses() {
	panic("TODO, stubbed")
}

// merge is the overload merge(..., vec<Lit>& out_clause) (protected).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, vec<Lit>& out_clause)
func (s *GlucoSimpSolver) merge(_ps *GlucoClause, _qs *GlucoClause, v GlucoVar, outClause *GlucoVec[GlucoLit]) bool {
	panic("TODO, stubbed")
}

// mergeSize is the overload merge(..., int& size) (protected).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, int& size)
func (s *GlucoSimpSolver) mergeSize(_ps *GlucoClause, _qs *GlucoClause, v GlucoVar, size *int) bool {
	panic("TODO, stubbed")
}

// backwardSubsumptionCheck is SimpSolver::backwardSubsumptionCheck (protected).
// C: bool backwardSubsumptionCheck(bool verbose = false)
func (s *GlucoSimpSolver) backwardSubsumptionCheck(verbose bool) bool {
	panic("TODO, stubbed")
}

// eliminateVar is SimpSolver::eliminateVar (protected).
// C: bool eliminateVar(Var v)
func (s *GlucoSimpSolver) eliminateVar(v GlucoVar) bool {
	panic("TODO, stubbed")
}

// extendModel is SimpSolver::extendModel (protected).
// C: void extendModel()
func (s *GlucoSimpSolver) extendModel() {
	panic("TODO, stubbed")
}

// removeClause is SimpSolver::removeClause (protected; hides Solver::removeClause).
// C: void removeClause(CRef cr)
func (s *GlucoSimpSolver) removeClause(cr GlucoCRef) {
	panic("TODO, stubbed")
}

// strengthenClause is SimpSolver::strengthenClause (protected).
// C: bool strengthenClause(CRef cr, Lit l)
func (s *GlucoSimpSolver) strengthenClause(cr GlucoCRef, l GlucoLit) bool {
	panic("TODO, stubbed")
}

// cleanUpClauses is SimpSolver::cleanUpClauses (protected).
// C: void cleanUpClauses()
func (s *GlucoSimpSolver) cleanUpClauses() {
	panic("TODO, stubbed")
}

// implied is SimpSolver::implied (protected).
// C: bool implied(const vec<Lit>& c)
func (s *GlucoSimpSolver) implied(c *GlucoVec[GlucoLit]) bool {
	panic("TODO, stubbed")
}

// relocAll is SimpSolver::relocAll (protected; hides Solver::relocAll).
// C: void relocAll(ClauseAllocator& to)
func (s *GlucoSimpSolver) relocAll(to *GlucoClauseAllocator) {
	panic("TODO, stubbed")
}
