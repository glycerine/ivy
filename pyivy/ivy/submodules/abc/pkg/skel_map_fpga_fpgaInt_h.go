package abc

// Hand-checked skeleton of src/map/fpga/fpgaInt.h.

// Hand-written skeleton of src/map/fpga/fpgaInt.h. FPGA_MAX_LEAVES is 6 and
// FPGA_MAX_LUTSIZE is 32. FPGA_ALLOCATE_FANOUT is commented out in the source,
// so the fanout members of Fpga_Node_t are not compiled. Go has no bitfields;
// each is a uint32 whose C width is in the comment. clock_t is int64.

// FpgaMan is Fpga_Man_t (struct Fpga_ManStruct_t_).
type FpgaMan struct {
	// the mapping graph
	pBins     []*FpgaNode  // Fpga_Node_t ** : the table of nodes hashed by their children
	nBins     int          // the size of the table
	pInputs   []*FpgaNode  // Fpga_Node_t ** : the array of inputs
	nInputs   int          // the number of inputs
	pOutputs  []*FpgaNode  // Fpga_Node_t ** : the array of outputs
	nOutputs  int          // the number of outputs
	nNodes    int          // the total number of nodes
	nLatches  int          // the number of latches in the circuit
	pConst1   *FpgaNode    // the constant 1 node
	vNodesAll *FpgaNodeVec // the nodes by number
	vAnds     *FpgaNodeVec // the nodes reachable from COs
	vMapping  *FpgaNodeVec // the nodes used in the current mapping
	// info about the original circuit
	pFileName      string    // char * : the file name
	ppOutputNames  []string  // char ** : the primary output names
	pInputArrivals []float32 // float * : the PI arrival times
	// mapping parameters
	nVarsMax      int     // the max number of variables
	fAreaRecovery int     // the flag to use area flow as the first parameter
	fVerbose      int     // the verbosiness flag
	fSwitching    int     // minimize the switching activity (instead of area)
	fLatchPaths   int     // optimize latch paths for delay, other paths for area
	nTravIds      int     // the counter of traversal IDs
	DelayTarget   float32 // the target required times
	// support of choice nodes
	nChoiceNodes int // the number of choice nodes
	nChoices     int // the number of all choices
	nCanons      int
	nMatches     int
	// the supergate library
	pLutLib *FpgaLutLib // the current LUT library
	// the memory managers
	mmNodes *ExtraFixedArena // Extra_MmFixed_t * : the memory manager for nodes
	mmCuts  *ExtraFixedArena // Extra_MmFixed_t * : the memory manager for cuts
	// resynthesis parameters
	fResynthesis   int     // the resynthesis flag
	fRequiredGlo   float32 // the global required times
	fRequiredShift float32 // the shift of the required times
	fRequiredStart float32 // the starting global required times
	fRequiredGain  float32 // the reduction in delay
	fAreaGlo       float32 // the total area
	fAreaGain      float32 // the reduction in area
	fEpsilon       float32 // the epsilon used to compare floats
	fDelayWindow   float32 // the delay window for delay-oriented resynthesis
	DelayLimit     float32 // for resynthesis
	AreaLimit      float32 // for resynthesis
	TimeLimit      float32 // for resynthesis
	// runtime statistics
	timeToMap   int64 // clock_t : time to transfer to the mapping structure
	timeCuts    int64 // clock_t : time to compute k-feasible cuts
	timeTruth   int64 // clock_t : time to compute the truth table for each cut
	timeMatch   int64 // clock_t : time to perform matching for each node
	timeRecover int64 // clock_t : time to perform area recovery
	timeToNet   int64 // clock_t : time to transfer back to the network
	timeTotal   int64 // clock_t : the total mapping time
	time1       int64 // clock_t : time to transfer to the mapping structure
	time2       int64 // clock_t : time to transfer to the mapping structure
}

// FpgaLutLib is Fpga_LutLib_t (struct Fpga_LutLibStruct_t_).
type FpgaLutLib struct {
	pName         string                  // char * : the name of the LUT library
	LutMax        int                     // the maximum LUT size
	fVarPinDelays int                     // set to 1 if variable pin delays are specified
	pLutAreas     [32 + 1]float32         // the areas of LUTs
	pLutDelays    [32 + 1][32 + 1]float32 // the delays of LUTs
}

// FpgaNode is Fpga_Node_t (struct Fpga_NodeStruct_t_).
type FpgaNode struct {
	// general information about the node
	pNext  *FpgaNode // the next node in the hash table
	pLevel *FpgaNode // the next node in the linked list by level
	Num    int       // the unique number of this node
	NumA   int       // the unique number of this node
	Num2   int       // the temporary number of this node
	nRefs  int       // the number of references (fanouts) of the given node
	fMark0 uint32    // unsigned : 1, the mark used for traversals
	fMark1 uint32    // unsigned : 1, the mark used for traversals
	fInv   uint32    // unsigned : 1, the complemented attribute for the equivalent nodes
	Value  uint32    // unsigned : 2, the value of the nodes
	fUsed  uint32    // unsigned : 1, the flag indicating that the node is used in the mapping
	fTemp  uint32    // unsigned : 1, unused
	Level  uint32    // unsigned : 11, the level of the given node
	uData  uint32    // unsigned : 14, used to mark the fanins, for which resynthesis was tried
	TravId int
	// the successors of this node
	p1     *FpgaNode // the first child
	p2     *FpgaNode // the second child
	pNextE *FpgaNode // the next functionally equivalent node
	pRepr  *FpgaNode // the representative of the functionally equivalent class
	// the delay information
	tRequired   float32 // the best area flow
	aEstFanouts float32 // the fanout estimation
	Switching   float32 // the probability of switching
	LValue      int     // the l-value of the node
	nLatches1   int16   // the number of latches on the first edge
	nLatches2   int16   // the number of latches on the second edge
	// cut information
	pCutBest *FpgaCut // the best mapping
	pCutOld  *FpgaCut // the old mapping
	pCuts    *FpgaCut // mapping choices for the node (elementary comes first)
	pCutsN   *FpgaCut // mapping choices for the node (elementary comes first)
	// misc information
	pData0 any // char * : temporary storage for the corresponding network node (a pointer in disguise)
}

// FpgaCut is Fpga_Cut_t (struct Fpga_CutStruct_t_).
type FpgaCut struct {
	pOne     *FpgaCut         // the father of this cut
	pTwo     *FpgaCut         // the mother of this cut
	pRoot    *FpgaNode        // the root of the cut
	ppLeaves [6 + 1]*FpgaNode // the leaves of this cut
	fLevel   float32          // the average level of the fanins
	uSign    uint32           // signature for quick comparison
	fMark    int8             // char : the mark to denote visited cut
	Phase    int8             // char : the mark to denote complemented cut
	nLeaves  int8             // char : the number of leaves of this cut
	nVolume  int8             // char : the volume of this cut
	tArrival float32          // the arrival time
	aFlow    float32          // the area flow of the cut
	pNext    *FpgaCut         // the pointer to the next cut in the list
}

// FpgaNodeVec is Fpga_NodeVec_t (struct Fpga_NodeVecStruct_t_).
type FpgaNodeVec struct {
	pArray []*FpgaNode // Fpga_Node_t ** : the array of nodes
	nSize  int         // the number of entries in the array
	nCap   int         // the number of allocated entries
}

// Fpga_FloatMoreThan (static inline)
// C: int Fpga_FloatMoreThan( Fpga_Man_t * p, float Arg1, float Arg2 )
func (p *FpgaMan) fpgaFloatMoreThan(Arg1 float32, Arg2 float32) int {
	panic("TODO, stubbed")
}

// Fpga_FloatLessThan (static inline)
// C: int Fpga_FloatLessThan( Fpga_Man_t * p, float Arg1, float Arg2 )
func (p *FpgaMan) fpgaFloatLessThan(Arg1 float32, Arg2 float32) int {
	panic("TODO, stubbed")
}

// Fpga_FloatEqual (static inline)
// C: int Fpga_FloatEqual( Fpga_Man_t * p, float Arg1, float Arg2 )
func (p *FpgaMan) fpgaFloatEqual(Arg1 float32, Arg2 float32) int {
	panic("TODO, stubbed")
}
