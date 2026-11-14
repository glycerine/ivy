package gia

// Gia_Man_t retains the original manager fields. Object pointers remain stable
// as the Go slice grows; object IDs replace C pointer subtraction.
type Gia_Man_t struct {
	aigerComment    []byte             // standard AIGER comment tail
	aigerExtensions map[byte][]byte    // uninterpreted AIGER extension payloads
	aigerLatchInits []int32            // standard latch states 0/1/-1, separate from vRegInits
	aigerSymbols    map[byte][]string  // input/latch/output/property symbol names
	aigerBadOutputs int32              // bad outputs within the CO ordering
	literalHash     map[[2]int32]int32 // Go bridge structural hash, initialized from existing ANDs
	store           *objectStore
	pName           string       // name of the AIG
	pSpec           string       // name of the input file
	nRegs           int32        // number of registers
	nRegsAlloc      int32        // number of allocated registers
	nObjs           int32        // number of objects
	nObjsAlloc      int32        // number of allocated objects
	pObjs           []*Gia_Obj_t // the array of objects
	pMuxes          []uint32     // control signals of MUXes
	nXors           int32        // the number of XORs
	nMuxes          int32        // the number of MUXes
	nBufs           int32        // the number of buffers
	vCis            *Vec_Int_t   // the vector of CIs (PIs + LOs)
	vCos            *Vec_Int_t   // the vector of COs (POs + LIs)
	vHash           Vec_Int_t    // hash links
	vHTable         Vec_Int_t    // hash table
	fAddStrash      int32        // performs additional structural hashing
	fSweeper        int32        // sweeper is running
	fGiaSimple      int32        // simple mode (no const-propagation and strashing)
	vRefs           Vec_Int_t    // the reference count
	Refs            []int32      // the reference count
	pLutRefs        []int32      // the reference count
	vLevels         *Vec_Int_t   // levels of the nodes
	nLevels         int32        // the mamixum level
	nConstrs        int32        // the number of constraints
	nTravIds        int32        // the current traversal ID
	nFront          int32        // frontier size
	pReprsOld       []int32      // representatives (for CIs and ANDs)
	pReprs          []Gia_Rpr_t  // representatives (for CIs and ANDs)
	pNexts          []int32      // next nodes in the equivalence classes
	pSibls          []int32      // next nodes in the choice nodes
	pIso            []int32      // pairs of structurally isomorphic nodes
	nTerLoop        int32        // the state where loop begins
	nTerStates      int32        // the total number of ternary states
	pFanData        []int32      // the database to store fanout information
	nFansAlloc      int32        // the size of fanout representation
	vFanoutNums     *Vec_Int_t   // static fanout
	vFanout         *Vec_Int_t   // static fanout
	vMapping        *Vec_Int_t   // mapping for each node
	vMapping2       *Vec_Wec_t   // mapping for each node
	vFanouts2       *Vec_Wec_t   // mapping fanouts
	vCellMapping    *Vec_Int_t   // mapping for each node
	pSatlutWinman   any          // windowing for SAT-based mapping
	vPacking        *Vec_Int_t   // packing information
	vConfigs        *Vec_Int_t   // cell configurations
	pCellStr        string       // cell description
	vLutConfigs     *Vec_Int_t   // LUT configurations
	vEdgeDelay      *Vec_Int_t   // special edge information
	vEdgeDelayR     *Vec_Int_t   // special edge information
	vEdge1          *Vec_Int_t   // special edge information
	vEdge2          *Vec_Int_t   // special edge information
	pCexComb        *Abc_Cex_t   // combinational counter-example
	pCexSeq         *Abc_Cex_t   // sequential counter-example
	vSeqModelVec    *Vec_Ptr_t   // sequential counter-examples
	vCopies         Vec_Int_t    // intermediate copies
	vCopies2        Vec_Int_t    // intermediate copies
	vVar2Obj        *Vec_Int_t   // mapping of variables into objects
	vTruths         *Vec_Int_t   // used for truth table computation
	vFlopClasses    *Vec_Int_t   // classes of flops for retiming/merging/etc
	vGateClasses    *Vec_Int_t   // classes of gates for abstraction
	vObjClasses     *Vec_Int_t   // classes of objects for abstraction
	vInitClasses    *Vec_Int_t   // classes of flops for retiming/merging/etc
	vRegClasses     *Vec_Int_t   // classes of registers for sequential synthesis
	vRegInits       *Vec_Int_t   // initial state
	vDoms           *Vec_Int_t   // dominators
	vBarBufs        *Vec_Int_t   // barrier buffers
	vXors           *Vec_Int_t   // temporary XORs
	pSwitching      []byte       // switching activity for each object
	pPlacement      []Gia_Plc_t  // placement of the objects
	pAigExtra       *Gia_Man_t   // combinational logic of holes
	vInArrs         *Vec_Flt_t   // PI arrival times
	vOutReqs        *Vec_Flt_t   // PO required times
	vCiArrs         *Vec_Int_t   // CI arrival times
	vCoReqs         *Vec_Int_t   // CO required times
	vCoArrs         *Vec_Int_t   // CO arrival times
	vCoAttrs        *Vec_Int_t   // CO attributes
	And2Delay       int32        // delay of the AND gate
	DefInArrs       float32      // default PI arrival times
	DefOutReqs      float32      // default PO required times
	vSwitching      *Vec_Int_t   // switching activity
	pTravIds        []int32      // separate traversal ID representation
	nTravIdsAlloc   int32        // the number of trav IDs allocated
	vNamesIn        *Vec_Ptr_t   // the input names
	vNamesOut       *Vec_Ptr_t   // the output names
	vUserPiIds      *Vec_Int_t   // numbers assigned to PIs by the user
	vUserPoIds      *Vec_Int_t   // numbers assigned to POs by the user
	vUserFfIds      *Vec_Int_t   // numbers assigned to FFs by the user
	vCiNumsOrig     *Vec_Int_t   // original CI names
	vCoNumsOrig     *Vec_Int_t   // original CO names
	vIdsOrig        *Vec_Int_t   // original object IDs
	vIdsEquiv       *Vec_Int_t   // original object IDs proved equivalent
	vCofVars        *Vec_Int_t   // cofactoring variables
	vClockDoms      *Vec_Vec_t   // clock domains
	vTiming         *Vec_Flt_t   // arrival/required/slack
	pManTime        any          // the timing manager
	pLutLib         any          // LUT library
	nHashHit        word         // hash table hit
	nHashMiss       word         // hash table miss
	pData           any          // various user data
	cexValues       [][]bool     // per-frame values while probing a counterexample
	pData2          *uint32      // various user data
	iData           int32        // various user data
	iData2          int32        // various user data
	nAnd2Delay      int32        // AND2 delay scaled to match delay numbers used
	fVerbose        int32        // verbose reports
	MappedArea      int32        // area after mapping
	MappedDelay     int32        // delay after mapping
	// bit-parallel simulation
	fBuiltInSim  int32
	iPatsPi      int32
	nSimWords    int32
	nSimWordsT   int32
	iPastPiMax   int32
	nSimWordsMax int32
	vSims        *Vec_Wrd_t
	vSimsT       *Vec_Wrd_t
	vSimsPi      *Vec_Wrd_t
	vSimsPo      *Vec_Wrd_t
	vClassOld    *Vec_Int_t
	vClassNew    *Vec_Int_t
	// incremental simulation
	fIncrSim    int32
	iNextPi     int32
	iTimeStamp  int32
	vTimeStamps *Vec_Int_t
	// truth table computation for small functions
	nTtVars   int32      // truth table variables
	nTtWords  int32      // truth table words
	vTtNums   *Vec_Int_t // object numbers
	vTtNodes  *Vec_Int_t // internal nodes
	vTtInputs *Vec_Ptr_t // truth tables for constant and primary inputs
	vTtMemory *Vec_Wrd_t // truth tables for internal nodes
	// balancing
	vSuper *Vec_Int_t // supergate
	vStore *Vec_Int_t // node storage
	// existential quantification
	iSuppPi    int32      // the number of support variables
	nSuppWords int32      // the number of support words
	vSuppWords *Vec_Wrd_t // support information
	vCopiesTwo Vec_Int_t  // intermediate copies
	vSuppVars  Vec_Int_t  // used variables
	vVarMap    Vec_Int_t  // used variables
	pUData     *Gia_Dat_t
}

// Gia_ManBarBufs exposes the barrier-buffer map used by clients which build
// and later consume barrier buffers on a manager.
func Gia_ManBarBufs(p *Gia_Man_t) *Vec_Int_t {
	if p.vBarBufs == nil {
		p.vBarBufs = Vec_IntAlloc(0)
	}
	return p.vBarBufs
}
