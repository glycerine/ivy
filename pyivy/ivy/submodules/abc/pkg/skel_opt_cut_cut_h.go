package abc

// Hand-written skeleton of the types in src/opt/cut/cut.h and cutInt.h.

// CutParams is Cut_Params_t (struct Cut_ParamsStruct_t_).
type CutParams struct {
	nVarsMax   int // the max cut size ("k" of the k-feasible cuts)
	nKeepMax   int // the max number of cuts kept at a node
	nIdsMax    int // the max number of IDs of cut objects
	nBitShift  int // the number of bits used for the latch counter of an edge
	nCutSet    int // the number of nodes in the cut set
	fTruth     int // compute truth tables
	fFilter    int // filter dominated cuts
	fSeq       int // compute sequential cuts
	fDrop      int // drop cuts on the fly
	fDag       int // compute only DAG cuts
	fTree      int // compute only tree cuts
	fGlobal    int // compute only global cuts
	fLocal     int // compute only local cuts
	fRecord    int // record the cut computation flow
	fRecordAig int // record the cut functions
	fFancy     int // perform fancy computations
	fMap       int // computes delay of FPGA mapping with cuts
	fAdjust    int // removed useless fanouts of XORs/MUXes
	fNpnSave   int // enables dumping 6-input truth tables
	fVerbose   int // the verbosiness flag
}

// CutCut is Cut_Cut_t (struct Cut_CutStruct_t_). Go has no bitfields; each
// bitfield is a uint32 whose C width is in the comment.
type CutCut struct {
	Num0     uint32  // unsigned : 11, temporary number
	Num1     uint32  // unsigned : 11, temporary number
	fSimul   uint32  // unsigned : 1, the value of cut's output at 000.. pattern
	fCompl   uint32  // unsigned : 1, the cut is complemented
	nVarsMax uint32  // unsigned : 4, the max number of vars [4-6]
	nLeaves  uint32  // unsigned : 4, the number of leaves [4-6]
	uSign    uint32  // the signature
	uCanon0  uint32  // the canonical form
	uCanon1  uint32  // the canonical form
	pNext    *CutCut // the next cut in the list
	pLeaves  []int   // int pLeaves[0] : flexible array member, the array of leaves
}

// CutMan is Cut_Man_t (struct Cut_ManStruct_t_).
type CutMan struct {
	// user preferences
	pParams    *CutParams // computation parameters
	vFanCounts *IntVector // the array of fanout counters
	vNodeAttrs *IntVector // node attributes (1 = global; 0 = local)
	// storage for cuts
	vCutsNew  *PointerVector[CutCut] // new cuts by node ID
	vCutsOld  *PointerVector[CutCut] // old cuts by node ID
	vCutsTemp *PointerVector[CutCut] // temp cuts for cutset nodes by cutset node number
	// memory management
	pMmCuts     *ExtraFixedArena // Extra_MmFixed_t *
	EntrySize   int
	nTruthWords int
	// temporary variables
	pReady      *CutCut
	vTemp       *PointerVector[CutCut]
	fCompl0     int
	fCompl1     int
	fSimul      int
	nNodeCuts   int
	pStore0     [2]*CutCut
	pStore1     [2]*CutCut
	pCompareOld *CutCut
	pCompareNew *CutCut
	puTemp      [4][]uint32 // unsigned * puTemp[4]
	// record of the cut computation
	vNodeCuts   *IntVector // the number of cuts for each node
	vNodeStarts *IntVector // the number of the starting cut of each node
	vCutPairs   *IntVector // the pairs of parent cuts for each cut
	// minimum delay mapping with the given cuts
	vCutsMax  *PointerVector[CutCut]
	vDelays   *IntVector
	vDelays2  *IntVector
	nDelayMin int
	// statistics
	nCutsCur     int
	nCutsAlloc   int
	nCutsDealloc int
	nCutsPeak    int
	nCutsTriv    int
	nCutsFilter  int
	nCutsLimit   int
	nNodes       int
	nNodesDag    int
	nNodesNoCuts int
	// runtime
	timeMerge  int64
	timeUnion  int64
	timeTruth  int64
	timeFilter int64
	timeHash   int64
	timeMap    int64
}
