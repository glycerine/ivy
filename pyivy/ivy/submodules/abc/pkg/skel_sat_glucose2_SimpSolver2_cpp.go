package abc

// Hand-checked skeleton of src/sat/glucose2/SimpSolver2.cpp (namespace Gluco2).
//
// The member functions defined in this file (SimpSolver::SimpSolver, ~SimpSolver, newVar, solve_,
// addClause_, removeClause, strengthenClause, merge (2 overloads), gatherTouchedClauses, implied,
// backwardSubsumptionCheck, asymm, asymmVar, eliminateVar, substitute, extendModel, eliminate,
// cleanUpClauses, relocAll, garbageCollect, reset) are stubbed as methods of Gluco2SimpSolver in
// skel_sat_glucose2_SimpSolver_h.go, next to the class declaration. Only the file-level statics are
// declared here.

// _cat = "SIMP"
const gluco2SimpCat = "SIMP"

// Command-line option objects. In C++ each is a static object constructed at load
// time (the constructor registers it in Option::getOptionList()); the Go skeleton
// declares them zero-valued, to be constructed at port time. The C initialisers:
//
//	static BoolOption opt_use_asymm (_cat, "asymm", "Shrink clauses by asymmetric branching.", false);
//	static BoolOption opt_use_rcheck (_cat, "rcheck", "Check if a clause is already implied. (costly)", false);
//	static BoolOption opt_use_elim (_cat, "elim", "Perform variable elimination.", true);
//	static IntOption opt_grow (_cat, "grow", "Allow a variable elimination step to grow by a number of clauses.", 0);
//	static IntOption opt_clause_lim (_cat, "cl-lim", "Variables are not eliminated if it produces a resolvent with a length above this limit. -1 means no limit", 20, IntRange(-1, INT32_MAX));
//	static IntOption opt_subsumption_lim (_cat, "sub-lim", "Do not check if subsumption against a clause larger than this. -1 means no limit.", 1000, IntRange(-1, INT32_MAX));
//	static DoubleOption opt_simp_garbage_frac(_cat, "simp-gc-frac", "The fraction of wasted memory allowed before a garbage collection is triggered during simplification.", 0.5, DoubleRange(0, false, HUGE_VAL, false));
var (
	opt_use_asymm_glucose2         Gluco2BoolOption
	opt_use_rcheck_glucose2        Gluco2BoolOption
	opt_use_elim_glucose2          Gluco2BoolOption
	opt_grow_glucose2              Gluco2IntOption
	opt_clause_lim_glucose2        Gluco2IntOption
	opt_subsumption_lim_glucose2   Gluco2IntOption
	opt_simp_garbage_frac_glucose2 Gluco2DoubleOption
)

// mkElimClauseLit_glucose2 is the overload mkElimClause(vec<uint32_t>&, Lit) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Lit x)
func mkElimClauseLit_glucose2(elimclauses *Gluco2Vec[uint32], x Gluco2Lit) {
	panic("TODO, stubbed")
}

// mkElimClauseVarClause_glucose2 is the overload mkElimClause(vec<uint32_t>&, Var, Clause&) (static).
// C: static void mkElimClause(vec<uint32_t>& elimclauses, Var v, Clause& c)
func mkElimClauseVarClause_glucose2(elimclauses *Gluco2Vec[uint32], v Gluco2Var, c *Gluco2Clause) {
	panic("TODO, stubbed")
}
