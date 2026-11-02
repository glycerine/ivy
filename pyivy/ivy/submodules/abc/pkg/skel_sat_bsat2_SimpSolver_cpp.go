package abc

// Hand-checked skeleton of src/sat/bsat2/SimpSolver.cpp (namespace Minisat).
//
// The member functions defined in this file (SimpSolver::SimpSolver, ~SimpSolver,
// newVar, solve_, addClause_, removeClause, strengthenClause, merge (2 overloads),
// gatherTouchedClauses, implied, backwardSubsumptionCheck, asymm, asymmVar,
// eliminateVar, substitute, extendModel, eliminate, cleanUpClauses, relocAll,
// garbageCollect) are stubbed as methods of MinisatSimpSolver in
// skel_sat_bsat2_SimpSolver_h.go, next to the class declaration. Only the
// file-level statics are declared here.

// minisatSimpSolverCat is `static const char* _cat = "SIMP"`.
const minisatSimpSolverCat = "SIMP"

// Command-line options of the simplifier (static objects in C++; declared
// zero-valued here, to be constructed at port time):
//
//	opt_use_asymm_minisat         = BoolOption(_cat, "asymm", "Shrink clauses by asymmetric branching.", false)
//	opt_use_rcheck_minisat        = BoolOption(_cat, "rcheck", "Check if a clause is already implied. (costly)", false)
//	opt_use_elim_minisat          = BoolOption(_cat, "elim", "Perform variable elimination.", true)
//	opt_grow_minisat              = IntOption(_cat, "grow", "Allow a variable elimination step to grow by a number of clauses.", 0)
//	opt_clause_lim_minisat        = IntOption(_cat, "cl-lim", "...resolvent with a length above this limit. -1 means no limit", 20, IntRange(-1, INT32_MAX))
//	opt_subsumption_lim_minisat   = IntOption(_cat, "sub-lim", "...subsumption against a clause larger than this. -1 means no limit.", 1000, IntRange(-1, INT32_MAX))
//	opt_simp_garbage_frac_minisat = DoubleOption(_cat, "simp-gc-frac", "...garbage collection is triggered during simplification.", 0.5, DoubleRange(0, false, HUGE_VAL, false))
var (
	opt_use_asymm_minisat         MinisatBoolOption
	opt_use_rcheck_minisat        MinisatBoolOption
	opt_use_elim_minisat          MinisatBoolOption
	opt_grow_minisat              MinisatIntOption
	opt_clause_lim_minisat        MinisatIntOption
	opt_subsumption_lim_minisat   MinisatIntOption
	opt_simp_garbage_frac_minisat MinisatDoubleOption
)

// mkElimClauseLit is the overload mkElimClause(vec<uint32_t>&, Lit) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Lit x)
func mkElimClauseLit(elimclauses *MinisatVec[uint32], x MinisatLit) {
	panic("TODO, stubbed")
}

// mkElimClauseVarClause is the overload mkElimClause(vec<uint32_t>&, Var, Clause&) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Var v, Clause& c)
func mkElimClauseVarClause(elimclauses *MinisatVec[uint32], v MinisatVar, c *MinisatClause) {
	panic("TODO, stubbed")
}
