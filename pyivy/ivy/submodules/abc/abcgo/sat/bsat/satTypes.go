package bsat

import "os"

type Status int32

const (
	Unknown       Status = l_Undef
	Satisfiable   Status = l_True
	Unsatisfiable Status = l_False
)

type word = uint64
type ABC_INT64_T = int64
type abctime = int64
type sat_solver = sat_solver_t
type varinfo = varinfo_t
type varinfo_t struct {
	val uint32 /* C bitfield: 2 bits; review packing and truncation */  // variable value
	pol uint32 /* C bitfield: 1 bits; review packing and truncation */  // last polarity
	tag uint32 /* C bitfield: 1 bits; review packing and truncation */  // conflict analysis tag
	lev uint32 /* C bitfield: 28 bits; review packing and truncation */ // variable level
}
type sat_solver_t struct {
	lastStatus                        Status
	limitConflicts, limitPropagations int64
	satAssignments
	size int32 // nof variables
	cap  int32 // size of varmaps

	// clauses
	Mem      Sat_Mem_t
	hLearnts int32 // the first learnt clause
	hBinary  int32 // the special binary clause
	binary   *clause
	wlists   []veci // watcher lists

	// rollback
	iVarPivot   int32 // the pivot for variables
	iTrailPivot int32 // the pivot for trail
	hProofPivot int32 // the pivot for proof records

	// activities
	VarActType int32
	ClaActType int32
	var_inc    word     // Amount to bump next variable with.
	var_inc2   word     // Amount to bump next variable with.
	var_decay  word     // INVERSE decay factor for variable activity: stores 1/decay.
	activity   []uint64 // A heuristic measurement of the activity of a variable.
	activity2  []uint64 // backup variable activity
	cla_inc    uint32   // Amount to bump next clause with.
	cla_decay  uint32   // INVERSE decay factor for clause activity: stores 1/decay.
	act_clas   veci     // contain clause activities

	pFreqs   []uint8 // how many times this variable was assigned a value
	nVarUsed int32

	//    varinfo *   vi;            // variable information
	loads []int32 //

	orderpos []int32 // Index in variable order.
	stack    veci    // (contains: var)

	order veci // Variable order. (heap) (contains: var)
	//    veci        model;         // If problem is solved, this vector contains the model (contains: lbool).
	model      []int32 // If problem is solved, this vector contains the model (contains: lbool).
	conf_final veci    // If problem is unsatisfiable (possibly under assumptions),
	// this vector represent the final conflict clause expressed in the assumptions.

	root_level        int32 // Level of first proper decision.
	simpdb_assigns    int32 // Number of top-level assignments at last 'simplifyDB()'.
	simpdb_props      int32 // Number of propagations before next 'simplifyDB()'.
	random_seed       float64
	progress_estimate float64
	verbosity         int32 // Verbosity level. 0=silent, 1=some progress report, 2=everything
	fVerbose          int32
	fPrintClause      int32

	stats        stats_t
	nLearntMax   int32 // max number of learned clauses
	nLearntStart int32 // starting learned clause limit
	nLearntDelta int32 // delta of learned clause limit
	nLearntRatio int32 // ratio percentage of learned clauses
	nDBreduces   int32 // number of DB reductions

	nConfLimit    ABC_INT64_T  // external limit on the number of conflicts
	nInsLimit     ABC_INT64_T  // external limit on the number of implications
	nRuntimeLimit abctime      // external limit on runtime
	runtimeClock  func() int64 // nil selects the existing wall clock

	act_vars             veci      // variables whose activity has changed
	factors              []float64 // the activity factors
	activityFactorActive map[int32]bool
	nRestarts            int32 // the number of local restarts
	nCalls               int32 // the number of local restarts
	nCalls2              int32 // the number of local restarts
	unit_lits            veci  // variables whose activity has changed
	pivot_vars           veci  // pivot variables

	fSkipSimplify int32 // set to one to skip simplification of the clause database
	fNotUseRandom int32 // do not allow random decisions with a fixed probability
	fNoRestarts   int32 // disables periodic restarts

	pGlobalVars []int32 // for experiments with global vars during interpolation
	// clause store
	pStore  any
	fSolved int32

	// trace recording
	pFile    *os.File
	nClauses int32
	nRoots   int32

	temp_clause veci // temporary storage for a CNF clause

	// assignment storage
	user_vars   veci // variable IDs
	user_values veci // values of these variables

	// CNF loading
	pCnfMan  any                    // external CNF manager
	pCnfFunc func(any, int32) int32 // external callback

	// termination callback
	RunId     int32             // SAT id in this run
	pFuncStop func(int32) int32 // callback to terminate
}
