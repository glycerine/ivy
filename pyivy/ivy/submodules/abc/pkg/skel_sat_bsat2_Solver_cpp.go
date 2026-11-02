package abc

// Hand-checked skeleton of src/sat/bsat2/Solver.cpp (namespace Minisat).
//
// The member functions defined in this file (Solver::Solver, ~Solver, newVar,
// addClause_, attachClause, detachClause, removeClause, satisfied, cancelUntil,
// pickBranchLit, analyze, litRedundant, analyzeFinal, uncheckedEnqueue, propagate,
// reduceDB, removeSatisfied, rebuildOrderHeap, simplify, search, progressEstimate,
// solve_, toDimacs (2 overloads), relocAll, garbageCollect) are stubbed as
// methods of MinisatSolver in skel_sat_bsat2_Solver_h.go, next to the class
// declaration. Only the file-level statics are declared here.

// minisatSolverCat is `static const char* _cat = "CORE"`.
const minisatSolverCat = "CORE"

// Command-line options of the solver. In C++ each is a static object constructed
// at load time (the constructor registers it in Option::getOptionList()); the Go
// skeleton declares them zero-valued, to be constructed at port time:
//
//	opt_var_decay_minisat        = DoubleOption(_cat, "var-decay",    "The variable activity decay factor", 0.95, DoubleRange(0, false, 1, false))
//	opt_clause_decay_minisat     = DoubleOption(_cat, "cla-decay",    "The clause activity decay factor", 0.999, DoubleRange(0, false, 1, false))
//	opt_random_var_freq_minisat  = DoubleOption(_cat, "rnd-freq",     "...random variable...", 0, DoubleRange(0, true, 1, true))
//	opt_random_seed_minisat      = DoubleOption(_cat, "rnd-seed",     "Used by the random variable selection", 91648253, DoubleRange(0, false, HUGE_VAL, false))
//	opt_ccmin_mode_minisat       = IntOption(_cat, "ccmin-mode",      "...conflict clause minimization...", 2, IntRange(0, 2))
//	opt_phase_saving_minisat     = IntOption(_cat, "phase-saving",    "...phase saving...", 2, IntRange(0, 2))
//	opt_rnd_init_act_minisat     = BoolOption(_cat, "rnd-init",       "Randomize the initial activity", false)
//	opt_luby_restart_minisat     = BoolOption(_cat, "luby",           "Use the Luby restart sequence", true)
//	opt_restart_first_minisat    = IntOption(_cat, "rfirst",          "The base restart interval", 100, IntRange(1, INT32_MAX))
//	opt_restart_inc_minisat      = DoubleOption(_cat, "rinc",         "Restart interval increase factor", 2, DoubleRange(1, false, HUGE_VAL, false))
//	opt_garbage_frac_minisat     = DoubleOption(_cat, "gc-frac",      "...wasted memory...", 0.20, DoubleRange(0, false, HUGE_VAL, false))
var (
	opt_var_decay_minisat       MinisatDoubleOption
	opt_clause_decay_minisat    MinisatDoubleOption
	opt_random_var_freq_minisat MinisatDoubleOption
	opt_random_seed_minisat     MinisatDoubleOption
	opt_ccmin_mode_minisat      MinisatIntOption
	opt_phase_saving_minisat    MinisatIntOption
	opt_rnd_init_act_minisat    MinisatBoolOption
	opt_luby_restart_minisat    MinisatBoolOption
	opt_restart_first_minisat   MinisatIntOption
	opt_restart_inc_minisat     MinisatDoubleOption
	opt_garbage_frac_minisat    MinisatDoubleOption
)

// luby_minisat is `static double luby(double y, int x)` (finite subsequences of the Luby sequence).
// C: static double luby(double y, int x)
func luby_minisat(y float64, x int) float64 {
	panic("TODO, stubbed")
}

// mapVar_minisat is `static Var mapVar(Var x, vec<Var>& map, Var& max)`.
// C: static Var mapVar(Var x, vec<Var>& map, Var& max)
func mapVar_minisat(x MinisatVar, map_ *MinisatVec[MinisatVar], max *MinisatVar) MinisatVar {
	panic("TODO, stubbed")
}
