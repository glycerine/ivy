package abc

// Hand-checked skeleton of src/map/mapper/mapperInt.h.

// Hand-written skeleton of the types in src/map/mapper/mapperInt.h.
// MAP_ALLOCATE_FANOUT is commented out in the source, so the fanout members of
// Map_Node_t are not compiled. Go has no bitfields; each is a uint32 whose C
// width is in the comment.

// MapMan is Map_Man_t (struct Map_ManStruct_t_).
type MapMan struct {
	// the mapping graph
	pBins       []*MapNode  // Map_Node_t ** : the table of nodes hashed by their children
	nBins       int         // the size of the table
	pInputs     []*MapNode  // Map_Node_t ** : the array of inputs
	nInputs     int         // the number of inputs
	pOutputs    []*MapNode  // Map_Node_t ** : the array of outputs
	nOutputs    int         // the number of outputs
	nNodes      int         // the total number of nodes
	pConst1     *MapNode    // the constant 1 node
	vMapObjs    *MapNodeVec // the array of all nodes
	vMapBufs    *MapNodeVec // the array of all nodes
	pNodeDelays []float32   // float * : the array of node delays
	// info about the original circuit
	ppOutputNames    []string  // char ** : the primary output names
	pInputArrivals   []MapTime // Map_Time_t * : the PI arrival times
	pOutputRequireds []MapTime // Map_Time_t * : the PI arrival times
	// mapping parameters
	nVarsMax          int     // the max number of variables
	fAreaRecovery     int     // the flag to enable area recovery
	fVerbose          int     // the verbosiness flag
	fMappingMode      int     // set to 1 when doing area
	fRequiredGlo      float32 // the global required times
	fEpsilon          float32 // the epsilon used to compare floats
	AreaBase          float32 // the area after delay-oriented mapping
	AreaFinal         float32 // the area after delay-oriented mapping
	nIterations       int     // How many matching passes to do
	fObeyFanoutLimits int     // Should mapper try to obey fanout limits or not
	DelayTarget       float32 // the required times set by the user
	nTravIds          int     // the traversal counter
	fSwitching        int     // use switching activity
	fSkipFanout       int     // skip large gates when mapping high-fanout nodes
	fUseProfile       int     // use standard-cell profile
	// the supergate library
	pSuperLib    *MapSuperLib   // the current supergate library
	uTruths      [6][2]uint32   // the elementary truth tables
	uTruthsLarge [10][32]uint32 // the elementary truth tables
	nCounts      [32]int        // the counter of minterms
	nCountsBest  [32]int        // the counter of minterms
	vVisited     *MapNodeVec    // the visited cuts during cut computation
	// the memory managers
	mmNodes *ExtraFixedArena // Extra_MmFixed_t * : the memory manager for nodes
	mmCuts  *ExtraFixedArena // Extra_MmFixed_t * : the memory manager for cuts
	// precomputed N-canonical forms
	uCanons   []uint16 // unsigned short * : N-canonical forms
	uPhases   [][]int8 // char ** : N-canonical phases
	pCounters []int8   // char * : counters of phases
	// various statistical variables
	nChoiceNodes      int // the number of choice nodes
	nChoices          int // the number of all choices
	nCanons           int // the number of times N-canonical form was computed
	nMatches          int // the number of times supergate matching was performed
	nPhases           int // the number of phases considered during matching
	nFanoutViolations int // the number of nodes in mapped circuit violating fanout
	// runtime statistics
	timeToMap int64 // abctime : time to transfer to the mapping structure
	timeCuts  int64 // abctime : time to compute k-feasible cuts
	timeTruth int64 // abctime : time to compute the truth table for each cut
	timeMatch int64 // abctime : time to perform matching for each node
	timeArea  int64 // abctime : time to recover area after delay oriented mapping
	timeSweep int64 // abctime : time to perform technology dependent sweep
	timeToNet int64 // abctime : time to transfer back to the network
	timeTotal int64 // abctime : the total mapping time
	time1     int64 // abctime
	time2     int64 // abctime
	time3     int64 // abctime
}

// MapSuperLib is Map_SuperLib_t (struct Map_SuperLibStruct_t_).
type MapSuperLib struct {
	// general info
	pName   string      // char * : the name of the supergate library
	pGenlib *MioLibrary // the generic library
	// other info
	nVarsMax    int // the max number of variables
	nSupersAll  int // the total number of supergates
	nSupersReal int // the total number of supergates
	nLines      int // the total number of lines in the supergate file
	fVerbose    int // the verbosity flag
	// hash tables
	ppSupers []*MapSuper   // Map_Super_t ** : the array of supergates
	tTableC  *MapHashTable // the table mapping N-canonical forms into supergates
	tTable   *MapHashTable // the table mapping truth tables into supergates
	// data structures for N-canonical form computation
	uTruths [6][2]uint32 // the elementary truth tables
	uMask   [2]uint32    // the mask for the truth table
	// the invertor
	pGateInv  *MioGate  // the pointer to the intertor gate
	tDelayInv MapTime   // the delay of the inverter
	AreaInv   float32   // the area of the inverter
	AreaBuf   float32   // the area of the buffer
	pSuperInv *MapSuper // the supergate representing the inverter
	// the memory manager for the internal table
	mmSupers  *ExtraFixedArena // Extra_MmFixed_t * : the mamory manager for supergates
	mmEntries *ExtraFixedArena // Extra_MmFixed_t * : the memory manager for the entries
	mmForms   *ExtraFlexArena  // Extra_MmFlex_t * : the memory manager for formulas
}

// MapNode is Map_Node_t (struct Map_NodeStruct_t_), the mapping node.
type MapNode struct {
	// general information about the node
	p         *MapMan    // the mapping manager
	pNext     *MapNode   // the next node in the hash table
	Num       int        // the unique number of this node
	TravId    int        // the traversal ID (use to avoid cleaning marks)
	nRefs     int        // the number of references (fanouts) of the given node
	fMark0    uint32     // unsigned : 1, the mark used for traversals
	fMark1    uint32     // unsigned : 1, the mark used for traversals
	fUsed     uint32     // unsigned : 1, the mark to mark the node or its fanins
	fInv      uint32     // unsigned : 1, the complemented attribute for the equivalent nodes
	fInvert   uint32     // unsigned : 1, the flag to denote the use of interter
	Level     uint32     // unsigned : 16, the level of the given node
	NumTemp   uint32     // unsigned : 10, the level of the given node
	nRefAct   [3]int     // estimated fanout for current covering phase, neg and pos and sum
	nRefEst   [3]float32 // actual fanout for previous covering phase, neg and pos and sum
	Switching float32    // the probability of switching
	// connectivity
	p1     *MapNode // the first child
	p2     *MapNode // the second child
	pNextE *MapNode // the next functionally equivalent node
	pRepr  *MapNode // the representative of the functionally equivalent class
	// the delay information
	tArrival  [2]MapTime // the best arrival time of the neg (0) and pos (1) phases
	tRequired [2]MapTime // the required time of the neg (0) and pos (1) phases
	// misc information
	pCutBest [2]*MapCut // the best mapping for neg and pos phase
	pCuts    *MapCut    // mapping choices for the node (elementary comes first)
	pData0   any        // char * : temporary storage for the corresponding network node (a pointer in disguise)
	pData1   any        // char * : temporary storage for the corresponding network node (a pointer in disguise)
}

// MapMatch is Map_Match_t (struct Map_MatchStruct_t_): the match of the cut.
type MapMatch struct {
	// information used for matching
	pSupers *MapSuper
	uPhase  uint32
	// information about the best selected match
	uPhaseBest uint32    // the best phase (the EXOR of match's phase and gate's phase)
	pSuperBest *MapSuper // the best supergate matched
	// the parameters of the match
	tArrive  MapTime // the arrival time of this match
	AreaFlow float32 // the area flow or area of this match
}

// MapCut is Map_Cut_t (struct Map_CutStruct_t_): the cuts used for matching.
type MapCut struct {
	pNext    *MapCut     // the pointer to the next cut in the list
	pOne     *MapCut     // the father of this cut
	pTwo     *MapCut     // the mother of this cut
	ppLeaves [6]*MapNode // the leaves of this cut
	uTruth   uint32      // truth table for five-input cuts
	nLeaves  int8        // char : the number of leaves
	nVolume  int8        // char : the volume of this cut
	fMark    int8        // char : the mark to denote visited cut
	Phase    int8        // char : the mark to denote complemented cut
	M        [2]MapMatch // the matches for positive/negative phase
}

// MapSuper is Map_Super_t (struct Map_SuperStruct_t_): the supergate
// internally represented.
type MapSuper struct {
	Num       int          // the ID of the supergate
	fSuper    uint32       // unsigned : 1, the flag to distinquish a real super from a fake one
	fExclude  uint32       // unsigned : 1, the flag if set causes gate to be excluded from being used for mapping
	nFanins   uint32       // unsigned : 3, the number of inputs
	nGates    uint32       // unsigned : 3, the number of gates inside this supergate
	nFanLimit uint32       // unsigned : 4, the max number of fanout count
	nSupers   uint32       // unsigned : 16, the number of supergates in the list
	nPhases   uint32       // unsigned : 4, the number of phases for matching with canonical form
	uPhases   [4]uint8     // unsigned char uPhases[4] : the maximum of 4 phases for matching with canonical form
	nUsed     int          // the number of times the supergate is used
	pFanins   [6]*MapSuper // the fanins of the gate
	pRoot     *MioGate     // the root gate
	uTruth    [2]uint32    // the truth table
	tDelaysR  [6]MapTime   // the pin-to-pin delay constraints for the rise of the output
	tDelaysF  [6]MapTime   // the pin-to-pin delay constraints for the rise of the output
	tDelayMax MapTime      // the maximum delay
	Area      float32      // the area
	pFormula  string       // char * : the symbolic formula
	pNext     *MapSuper    // the pointer to the next super in the list
}

// MapNodeVec is Map_NodeVec_t (struct Map_NodeVecStruct_t_): the vector of nodes.
type MapNodeVec struct {
	pArray []*MapNode // Map_Node_t ** : the array of nodes
	nSize  int        // the number of entries in the array
	nCap   int        // the number of allocated entries
}

// MapHashTable is Map_HashTable_t (struct Map_HashTableStruct_t_): the hash table.
type MapHashTable struct {
	pBins    []*MapHashEntry  // Map_HashEntry_t ** : the table bins
	nBins    int              // the size of the table
	nEntries int              // the total number of entries in the table
	mmMan    *ExtraFixedArena // Extra_MmFixed_t * : the memory manager for entries
}

// MapHashEntry is Map_HashEntry_t (struct Map_HashEntryStruct_t_): the entry
// in the hash table.
type MapHashEntry struct {
	uTruth [2]uint32     // the truth table for 6-var function
	uPhase uint32        // the phase to tranform it into the canonical form
	pGates *MapSuper     // the linked list of matching supergates
	pNext  *MapHashEntry // the next entry in the hash table
}
