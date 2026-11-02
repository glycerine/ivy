package abc

// Hand-written skeleton of the types in src/proof/ssw/sswInt.h.

// SswMan is Ssw_Man_t.
type SswMan struct {
	// parameters
	pPars   *SswPars // parameters
	nFrames int      // for quick lookup
	// AIGs used in the package
	pAig          *AigManager  // user-given AIG
	pFrames       *AigManager  // final AIG
	pNodeToFrames []*AigObject // Aig_Obj_t ** : mapping of AIG nodes into FRAIG nodes
	// equivalence classes
	ppClasses *SswCla // equivalence classes of nodes
	fRefined  int     // is set to 1 when refinement happens
	// SAT solving
	pMSatBmc *SswSat // SAT manager for base case
	pMSat    *SswSat // SAT manager for inductive case
	// SAT solving (latch corr only)
	vSimInfo       *PointerVector[uint32] // simulation information for the framed PIs
	nPatterns      int                    // the number of patterns saved
	nSimRounds     int                    // the number of simulation rounds performed
	nCallsCount    int                    // the number of calls in this round
	nCallsDelta    int                    // the number of calls to skip
	nCallsSat      int                    // the number of SAT calls in this round
	nCallsUnsat    int                    // the number of UNSAT calls in this round
	nRecycleCalls  int                    // the number of calls since last recycling
	nRecycles      int                    // the number of time SAT solver was recycled
	nRecyclesTotal int                    // the number of time SAT solver was recycled
	nVarsMax       int                    // the maximum variables in the solver
	nCallsMax      int                    // the maximum number of SAT calls
	// uniqueness
	vCommon        *PointerVector[AigObject] // the set of common variables in the logic cones
	iOutputLit     int                       // the output literal of the uniqueness constraint
	vDiffPairs     *IntVector                // is set to 1 if reg pair can be diff
	nUniques       int                       // the number of uniqueness constraints used
	nUniquesAdded  int                       // useful uniqueness constraints
	nUniquesUseful int                       // useful uniqueness constraints
	// dynamic constraint addition
	nSRMiterMaxId int                       // max ID after which the last frame begins
	vNewLos       *PointerVector[AigObject] // new time frame LOs of to constrain
	vNewPos       *IntVector                // new time frame POs of to add constraints
	pVisited      []int                     // int * : flags to label visited nodes in each frame
	nVisCounter   int                       // the traversal ID
	// sequential simulation
	pSml          *SswSml                   // the simulator
	iNodeStart    int                       // the first node considered
	iNodeLast     int                       // the last node considered
	vResimConsts  *PointerVector[AigObject] // resimulation constants
	vResimClasses *PointerVector[AigObject] // resimulation classes
	vInits        *IntVector                // the init values of primary inputs under constraints
	// counter example storage
	nPatWords int      // the number of words in the counter example
	pPatWords []uint32 // unsigned * : the counter example
	// constraints
	nConstrTotal   int // the number of total constraints
	nConstrReduced int // the number of reduced constraints
	nStrangers     int // the number of strange situations
	// SAT calls statistics
	nSatCalls      int // the number of SAT calls
	nSatProof      int // the number of proofs
	nSatFailsReal  int // the number of timeouts
	nSatCallsUnsat int // the number of unsat SAT calls
	nSatCallsSat   int // the number of sat SAT calls
	// node/register/lit statistics
	nLitsBeg  int
	nLitsEnd  int
	nNodesBeg int
	nNodesEnd int
	nRegsBeg  int
	nRegsEnd  int
	// equiv statistis
	nConesTotal   int
	nConesConstr  int
	nEquivsTotal  int
	nEquivsConstr int
	nNodesBegC    int
	nNodesEndC    int
	nRegsBegC     int
	nRegsEndC     int
	// runtime stats
	timeBmc       int64 // bounded model checking
	timeReduce    int64 // speculative reduction
	timeMarkCones int64 // marking the cones not to be refined
	timeSimSat    int64 // simulation of the counter-examples
	timeSat       int64 // solving SAT
	timeSatSat    int64 // sat
	timeSatUnsat  int64 // unsat
	timeSatUndec  int64 // undecided
	timeOther     int64 // other runtime
	timeTotal     int64 // total runtime
}

// SswSat is Ssw_Sat_t: internal SAT manager.
type SswSat struct {
	pAig         *AigManager               // the AIG manager
	fPolarFlip   int                       // flips polarity
	pSat         *SatSolver                // recyclable SAT solver
	nSatVars     int                       // the counter of SAT variables
	vSatVars     *IntVector                // mapping of each node into its SAT var
	vFanins      *PointerVector[AigObject] // fanins of the CNF node
	vUsedPis     *PointerVector[AigObject] // the PIs with SAT variables
	nSolverCalls int                       // the total number of SAT calls
}

// SswFrm is Ssw_Frm_t: internal frames manager.
type SswFrm struct {
	pAig     *AigManager               // user-given AIG
	nObjs    int                       // offset in terms of AIG nodes
	nFrames  int                       // the number of frames in current unrolling
	pFrames  *AigManager               // unrolled AIG
	vAig2Frm *PointerVector[AigObject] // mapping of AIG nodes into frame nodes
}
