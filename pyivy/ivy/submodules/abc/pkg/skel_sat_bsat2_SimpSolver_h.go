package abc

// Hand-checked skeleton of src/sat/bsat2/SimpSolver.h (namespace Minisat).
// SimpSolver derives from Solver: it embeds MinisatSolver and re-declares the
// overridden methods (NewVar, AddClause*, Solve*, ToDimacs*, GarbageCollect, ...).

// MinisatSimpSolverElimLt is SimpSolver::ElimLt.
type MinisatSimpSolverElimLt struct {
	n_occ *MinisatVec[int] // const vec<int>&
}

// NewMinisatSimpSolverElimLt is the explicit constructor ElimLt(const vec<int>& no).
// C: explicit ElimLt(const vec<int>& no)
func NewMinisatSimpSolverElimLt(no *MinisatVec[int]) MinisatSimpSolverElimLt {
	panic("TODO, stubbed")
}

// cost is ElimLt::cost.
// C: uint64_t cost(Var x) const
func (e MinisatSimpSolverElimLt) cost(x MinisatVar) uint64 {
	panic("TODO, stubbed")
}

// Call is ElimLt::operator().
// C: bool operator()(Var x, Var y) const
func (e MinisatSimpSolverElimLt) Call(x int, y int) bool {
	panic("TODO, stubbed")
}

// MinisatSimpSolverClauseDeleted is SimpSolver::ClauseDeleted.
type MinisatSimpSolverClauseDeleted struct {
	ca *MinisatClauseAllocator // const ClauseAllocator&
}

// NewMinisatSimpSolverClauseDeleted is the explicit constructor ClauseDeleted(const ClauseAllocator& _ca).
// C: explicit ClauseDeleted(const ClauseAllocator& _ca)
func NewMinisatSimpSolverClauseDeleted(ca *MinisatClauseAllocator) MinisatSimpSolverClauseDeleted {
	panic("TODO, stubbed")
}

// Call is ClauseDeleted::operator().
// C: bool operator()(const CRef& cr) const
func (d MinisatSimpSolverClauseDeleted) Call(cr MinisatCRef) bool {
	panic("TODO, stubbed")
}

// MinisatSimpSolver is Minisat::SimpSolver : public Solver.
type MinisatSimpSolver struct {
	MinisatSolver

	Grow              int     // allow a variable elimination step to grow by a number of clauses (default zero)
	Clause_lim        int     // variables are not eliminated if it produces a resolvent longer than this limit
	Subsumption_lim   int     // do not check subsumption against a clause larger than this (-1 = no limit)
	Simp_garbage_frac float64 // limit for when to issue a GC during simplification
	Use_asymm         bool    // shrink clauses by asymmetric branching
	Use_rcheck        bool    // check if a clause is already implied
	Use_elim          bool    // perform variable elimination
	Merges            int
	Asymm_lits        int
	Eliminated_vars   int

	// Solver state (protected):
	elimorder          int
	use_simplification bool
	elimclauses        MinisatVec[uint32]
	touched            MinisatVec[int8]
	occurs             MinisatOccLists[int, MinisatCRef, MinisatSimpSolverClauseDeleted]
	n_occ              MinisatVec[int]
	elim_heap          MinisatHeap[MinisatSimpSolverElimLt]
	subsumption_queue  MinisatQueue[MinisatCRef]
	frozen             MinisatVec[int8]
	eliminated         MinisatVec[int8]
	bwdsub_assigns     int
	n_touched          int
	bwdsub_tmpunit     MinisatCRef
}

// NewMinisatSimpSolver is SimpSolver::SimpSolver().
// C: SimpSolver::SimpSolver()
func NewMinisatSimpSolver() *MinisatSimpSolver {
	panic("TODO, stubbed")
}

// Destroy is SimpSolver::~SimpSolver().
// C: SimpSolver::~SimpSolver()
func (s *MinisatSimpSolver) Destroy() {
	panic("TODO, stubbed")
}

// NewVar is SimpSolver::newVar (override; polarity and dvar default to true).
// C: Var newVar(bool polarity = true, bool dvar = true)
func (s *MinisatSimpSolver) NewVar(polarity bool, dvar bool) MinisatVar {
	panic("TODO, stubbed")
}

// AddClause is the overload addClause(const vec<Lit>& ps).
// C: bool addClause(const vec<Lit>& ps)
func (s *MinisatSimpSolver) AddClause(ps *MinisatVec[MinisatLit]) bool {
	panic("TODO, stubbed")
}

// AddEmptyClause is SimpSolver::addEmptyClause.
// C: bool addEmptyClause()
func (s *MinisatSimpSolver) AddEmptyClause() bool {
	panic("TODO, stubbed")
}

// AddClauseUnit is the overload addClause(Lit p).
// C: bool addClause(Lit p)
func (s *MinisatSimpSolver) AddClauseUnit(p MinisatLit) bool {
	panic("TODO, stubbed")
}

// AddClauseBinary is the overload addClause(Lit p, Lit q).
// C: bool addClause(Lit p, Lit q)
func (s *MinisatSimpSolver) AddClauseBinary(p MinisatLit, q MinisatLit) bool {
	panic("TODO, stubbed")
}

// AddClauseTernary is the overload addClause(Lit p, Lit q, Lit r).
// C: bool addClause(Lit p, Lit q, Lit r)
func (s *MinisatSimpSolver) AddClauseTernary(p MinisatLit, q MinisatLit, r MinisatLit) bool {
	panic("TODO, stubbed")
}

// AddClause_ is SimpSolver::addClause_.
// C: bool addClause_(vec<Lit>& ps)
func (s *MinisatSimpSolver) AddClause_(ps *MinisatVec[MinisatLit]) bool {
	panic("TODO, stubbed")
}

// Substitute is SimpSolver::substitute (replace all occurrences of v with x; may cause a contradiction).
// C: bool substitute(Var v, Lit x)
func (s *MinisatSimpSolver) Substitute(v MinisatVar, x MinisatLit) bool {
	panic("TODO, stubbed")
}

// SetFrozen is SimpSolver::setFrozen (a frozen variable is not eliminated).
// C: void setFrozen(Var v, bool b)
func (s *MinisatSimpSolver) SetFrozen(v MinisatVar, b bool) {
	panic("TODO, stubbed")
}

// IsEliminated is SimpSolver::isEliminated.
// C: bool isEliminated(Var v) const
func (s *MinisatSimpSolver) IsEliminated(v MinisatVar) bool {
	panic("TODO, stubbed")
}

// Solve is the overload solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) Solve(assumps *MinisatVec[MinisatLit], doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// SolveLimited is SimpSolver::solveLimited.
// C: lbool solveLimited(const vec<Lit>& assumps, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) SolveLimited(assumps *MinisatVec[MinisatLit], doSimp bool, turnOffSimp bool) MinisatLbool {
	panic("TODO, stubbed")
}

// SolveNoAssumps is the overload solve(bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) SolveNoAssumps(doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve1 is the overload solve(Lit p, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) Solve1(p MinisatLit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve2 is the overload solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) Solve2(p MinisatLit, q MinisatLit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Solve3 is the overload solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false).
// C: bool solve(Lit p, Lit q, Lit r, bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) Solve3(p MinisatLit, q MinisatLit, r MinisatLit, doSimp bool, turnOffSimp bool) bool {
	panic("TODO, stubbed")
}

// Eliminate is SimpSolver::eliminate (variable-elimination based simplification).
// C: bool eliminate(bool turn_off_elim = false)
func (s *MinisatSimpSolver) Eliminate(turnOffElim bool) bool {
	panic("TODO, stubbed")
}

// GarbageCollect is SimpSolver::garbageCollect (override).
// C: virtual void garbageCollect()
func (s *MinisatSimpSolver) GarbageCollect() {
	panic("TODO, stubbed")
}

// ToDimacsFile is the overload toDimacs(const char* file, const vec<Lit>& assumps).
// C: void toDimacs(const char* file, const vec<Lit>& assumps)
func (s *MinisatSimpSolver) ToDimacsFile(file string, assumps *MinisatVec[MinisatLit]) {
	panic("TODO, stubbed")
}

// ToDimacsFileNoAssumps is the overload toDimacs(const char* file).
// C: void toDimacs(const char* file)
func (s *MinisatSimpSolver) ToDimacsFileNoAssumps(file string) {
	panic("TODO, stubbed")
}

// ToDimacsFile1 is the overload toDimacs(const char* file, Lit p).
// C: void toDimacs(const char* file, Lit p)
func (s *MinisatSimpSolver) ToDimacsFile1(file string, p MinisatLit) {
	panic("TODO, stubbed")
}

// ToDimacsFile2 is the overload toDimacs(const char* file, Lit p, Lit q).
// C: void toDimacs(const char* file, Lit p, Lit q)
func (s *MinisatSimpSolver) ToDimacsFile2(file string, p MinisatLit, q MinisatLit) {
	panic("TODO, stubbed")
}

// ToDimacsFile3 is the overload toDimacs(const char* file, Lit p, Lit q, Lit r).
// C: void toDimacs(const char* file, Lit p, Lit q, Lit r)
func (s *MinisatSimpSolver) ToDimacsFile3(file string, p MinisatLit, q MinisatLit, r MinisatLit) {
	panic("TODO, stubbed")
}

// solve_ is SimpSolver::solve_ (protected; hides Solver::solve_).
// C: lbool solve_(bool do_simp = true, bool turn_off_simp = false)
func (s *MinisatSimpSolver) solve_(doSimp bool, turnOffSimp bool) MinisatLbool {
	panic("TODO, stubbed")
}

// asymm is SimpSolver::asymm (protected).
// C: bool asymm(Var v, CRef cr)
func (s *MinisatSimpSolver) asymm(v MinisatVar, cr MinisatCRef) bool {
	panic("TODO, stubbed")
}

// asymmVar is SimpSolver::asymmVar (protected).
// C: bool asymmVar(Var v)
func (s *MinisatSimpSolver) asymmVar(v MinisatVar) bool {
	panic("TODO, stubbed")
}

// updateElimHeap is SimpSolver::updateElimHeap (protected).
// C: void updateElimHeap(Var v)
func (s *MinisatSimpSolver) updateElimHeap(v MinisatVar) {
	panic("TODO, stubbed")
}

// gatherTouchedClauses is SimpSolver::gatherTouchedClauses (protected).
// C: void gatherTouchedClauses()
func (s *MinisatSimpSolver) gatherTouchedClauses() {
	panic("TODO, stubbed")
}

// merge is the overload merge(..., vec<Lit>& out_clause) (protected).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, vec<Lit>& out_clause)
func (s *MinisatSimpSolver) merge(_ps *MinisatClause, _qs *MinisatClause, v MinisatVar, outClause *MinisatVec[MinisatLit]) bool {
	panic("TODO, stubbed")
}

// mergeSize is the overload merge(..., int& size) (protected).
// C: bool merge(const Clause& _ps, const Clause& _qs, Var v, int& size)
func (s *MinisatSimpSolver) mergeSize(_ps *MinisatClause, _qs *MinisatClause, v MinisatVar, size *int) bool {
	panic("TODO, stubbed")
}

// backwardSubsumptionCheck is SimpSolver::backwardSubsumptionCheck (protected).
// C: bool backwardSubsumptionCheck(bool verbose = false)
func (s *MinisatSimpSolver) backwardSubsumptionCheck(verbose bool) bool {
	panic("TODO, stubbed")
}

// eliminateVar is SimpSolver::eliminateVar (protected).
// C: bool eliminateVar(Var v)
func (s *MinisatSimpSolver) eliminateVar(v MinisatVar) bool {
	panic("TODO, stubbed")
}

// extendModel is SimpSolver::extendModel (protected).
// C: void extendModel()
func (s *MinisatSimpSolver) extendModel() {
	panic("TODO, stubbed")
}

// removeClause is SimpSolver::removeClause (protected; hides Solver::removeClause).
// C: void removeClause(CRef cr)
func (s *MinisatSimpSolver) removeClause(cr MinisatCRef) {
	panic("TODO, stubbed")
}

// strengthenClause is SimpSolver::strengthenClause (protected).
// C: bool strengthenClause(CRef cr, Lit l)
func (s *MinisatSimpSolver) strengthenClause(cr MinisatCRef, l MinisatLit) bool {
	panic("TODO, stubbed")
}

// cleanUpClauses is SimpSolver::cleanUpClauses (protected).
// C: void cleanUpClauses()
func (s *MinisatSimpSolver) cleanUpClauses() {
	panic("TODO, stubbed")
}

// implied is SimpSolver::implied (protected).
// C: bool implied(const vec<Lit>& c)
func (s *MinisatSimpSolver) implied(c *MinisatVec[MinisatLit]) bool {
	panic("TODO, stubbed")
}

// relocAll is SimpSolver::relocAll (protected; hides Solver::relocAll).
// C: void relocAll(ClauseAllocator& to)
func (s *MinisatSimpSolver) relocAll(to *MinisatClauseAllocator) {
	panic("TODO, stubbed")
}
