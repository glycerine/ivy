package abc

// Hand-checked skeleton of src/sat/glucose2/Glucose2.cpp (namespace Gluco2).
//
// The member functions defined in this file (Solver::Solver, ~Solver, setIncrementalMode,
// initNbInitialVars, newVar, addClause_, attachClause, detachClause, removeClause, satisfied,
// computeLBD (2 overloads), minimisationWithBinaryResolution, cancelUntil, pickBranchLit,
// analyze, litRedundant, analyzeFinal, uncheckedEnqueue, propagate, reduceDB, removeSatisfied,
// rebuildOrderHeap, simplify, search, progressEstimate, printIncrementalStats, solve_,
// toDimacs (2 overloads), relocAll, garbageCollect, reset) are stubbed as methods of Gluco2Solver in
// skel_sat_glucose2_Solver_h.go, next to the class declaration. Only the file-level statics are
// declared here.

// File-level category strings: _cat = "CORE", _cr = "CORE -- RESTART", _cred = "CORE -- REDUCE",
// _cm = "CORE -- MINIMIZE", _certified = "CORE -- CERTIFIED UNSAT".
const (
	gluco2Cat       = "CORE"
	gluco2Cr        = "CORE -- RESTART"
	gluco2Cred      = "CORE -- REDUCE"
	gluco2Cm        = "CORE -- MINIMIZE"
	gluco2Certified = "CORE -- CERTIFIED UNSAT"
)

// Command-line option objects. In C++ each is a static object constructed at load
// time (the constructor registers it in Option::getOptionList()); the Go skeleton
// declares them zero-valued, to be constructed at port time. The C initialisers:
//
//	static BoolOption opt_incremental (_cat,"incremental", "Use incremental SAT solving", false);
//	static DoubleOption opt_K (_cr, "K", "The constant used to force restart", 0.8, DoubleRange(0, false, 1, false));
//	static DoubleOption opt_R (_cr, "R", "The constant used to block restart", 1.4, DoubleRange(1, false, 5, false));
//	static IntOption opt_size_lbd_queue (_cr, "szLBDQueue", "The size of moving average for LBD (restarts)", 50, IntRange(10, INT32_MAX));
//	static IntOption opt_size_trail_queue (_cr, "szTrailQueue", "The size of moving average for trail (block restarts)", 5000, IntRange(10, INT32_MAX));
//	static IntOption opt_first_reduce_db (_cred, "firstReduceDB", "The number of conflicts before the first reduce DB", 2000, IntRange(0, INT32_MAX));
//	static IntOption opt_inc_reduce_db (_cred, "incReduceDB", "Increment for reduce DB", 300, IntRange(0, INT32_MAX));
//	static IntOption opt_spec_inc_reduce_db (_cred, "specialIncReduceDB", "Special increment for reduce DB", 1000, IntRange(0, INT32_MAX));
//	static IntOption opt_lb_lbd_frozen_clause (_cred, "minLBDFrozenClause", "Protect clauses if their LBD decrease and is lower than (for one turn)", 30, IntRange(0, INT32_MAX));
//	static IntOption opt_lb_size_minimzing_clause (_cm, "minSizeMinimizingClause", "The min size required to minimize clause", 30, IntRange(3, INT32_MAX));
//	static IntOption opt_lb_lbd_minimzing_clause (_cm, "minLBDMinimizingClause", "The min LBD required to minimize clause", 6, IntRange(3, INT32_MAX));
//	static DoubleOption opt_var_decay (_cat, "var-decay", "The variable activity decay factor", 0.8, DoubleRange(0, false, 1, false));
//	static DoubleOption opt_clause_decay (_cat, "cla-decay", "The clause activity decay factor", 0.999, DoubleRange(0, false, 1, false));
//	static DoubleOption opt_random_var_freq (_cat, "rnd-freq", "The frequency with which the decision heuristic tries to choose a random variable", 0, DoubleRange(0, true, 1, true));
//	static DoubleOption opt_random_seed (_cat, "rnd-seed", "Used by the random variable selection", 91648253, DoubleRange(0, false, HUGE_VAL, false));
//	static IntOption opt_ccmin_mode (_cat, "ccmin-mode", "Controls conflict clause minimization (0=none, 1=basic, 2=deep)", 2, IntRange(0, 2));
//	static IntOption opt_phase_saving (_cat, "phase-saving", "Controls the level of phase saving (0=none, 1=limited, 2=full)", 2, IntRange(0, 2));
//	static BoolOption opt_rnd_init_act (_cat, "rnd-init", "Randomize the initial activity", false);
//	static IntOption opt_restart_first (_cat, "rfirst", "The base restart interval", 100, IntRange(1, INT32_MAX));
//	static DoubleOption opt_restart_inc (_cat, "rinc", "Restart interval increase factor", 2, DoubleRange(1, false, HUGE_VAL, false));
//	static DoubleOption opt_garbage_frac (_cat, "gc-frac", "The fraction of wasted memory allowed before a garbage collection is triggered", 0.20, DoubleRange(0, false, HUGE_VAL, false));
var (
	opt_incremental_glucose2              Gluco2BoolOption
	opt_K_glucose2                        Gluco2DoubleOption
	opt_R_glucose2                        Gluco2DoubleOption
	opt_size_lbd_queue_glucose2           Gluco2IntOption
	opt_size_trail_queue_glucose2         Gluco2IntOption
	opt_first_reduce_db_glucose2          Gluco2IntOption
	opt_inc_reduce_db_glucose2            Gluco2IntOption
	opt_spec_inc_reduce_db_glucose2       Gluco2IntOption
	opt_lb_lbd_frozen_clause_glucose2     Gluco2IntOption
	opt_lb_size_minimzing_clause_glucose2 Gluco2IntOption
	opt_lb_lbd_minimzing_clause_glucose2  Gluco2IntOption
	opt_var_decay_glucose2                Gluco2DoubleOption
	opt_clause_decay_glucose2             Gluco2DoubleOption
	opt_random_var_freq_glucose2          Gluco2DoubleOption
	opt_random_seed_glucose2              Gluco2DoubleOption
	opt_ccmin_mode_glucose2               Gluco2IntOption
	opt_phase_saving_glucose2             Gluco2IntOption
	opt_rnd_init_act_glucose2             Gluco2BoolOption
	opt_restart_first_glucose2            Gluco2IntOption
	opt_restart_inc_glucose2              Gluco2DoubleOption
	opt_garbage_frac_glucose2             Gluco2DoubleOption
)

// opt_certified_ and opt_certified_file_ are not static (external linkage):
//
//	BoolOption opt_certified_ (_certified, "certified", "Certified UNSAT using DRUP format", false);
//	StringOption opt_certified_file_ (_certified, "certified-output", "Certified UNSAT output file", "NULL");
var (
	opt_certified__glucose2      Gluco2BoolOption
	opt_certified_file__glucose2 Gluco2StringOption
)

// mapVar_glucose2 is `static Var mapVar(Var x, vec<Var>& map, Var& max)`.
// C: static Var mapVar(Var x, vec<Var>& map, Var& max)
func mapVar_glucose2(x Gluco2Var, map_ *Gluco2Vec[Gluco2Var], max *Gluco2Var) Gluco2Var {
	panic("TODO, stubbed")
}
