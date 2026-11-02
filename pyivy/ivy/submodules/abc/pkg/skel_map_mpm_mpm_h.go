package abc

// Hand-written skeleton of the types in src/map/mpm/mpm.h, mpmInt.h and
// mpmMig.h. MPM_VAR_MAX is 12 and MPM_CUT_MAX is 32. Go has no bitfields;
// each is a uint32 whose C width is in the comment. A C trailing array used as
// a flexible array member is a slice.

// MpmLibLut is Mpm_LibLut_t.
type MpmLibLut struct {
	pName         string              // char * : the name of the LUT library
	LutMax        int                 // the maximum LUT size
	fVarPinDelays int                 // set to 1 if variable pin delays are specified
	pLutAreas     [12 + 1]int         // the areas of LUTs
	pLutDelays    [12 + 1][12 + 1]int // the delays of LUTs
}

// MpmPar is Mpm_Par_t.
type MpmPar struct {
	pLib         *MpmLibLut
	pScl         *SCLib // void *
	nNumCuts     int
	DelayTarget  int
	fUseGates    int
	fUseTruth    int
	fUseDsd      int
	fCutMin      int
	fOneRound    int
	fDeriveLuts  int
	fMap4Cnf     int
	fMap4Aig     int
	fMap4Gates   int
	fVerbose     int
	fVeryVerbose int
}

// MigFan is Mig_Fan_t.
type MigFan struct {
	fCompl uint32 // unsigned : 1, the complemented attribute
	Id     uint32 // unsigned : 31, fanin ID
}

// MigObj is Mig_Obj_t.
type MigObj struct {
	pFans [4]MigFan // fanins
}

// MigMan is Mig_Man_t.
type MigMan struct {
	pName    string                // char * : name
	nObjs    int                   // number of objects
	nRegs    int                   // number of flops
	nChoices int                   // number of choices
	vPages   PointerVector[MigObj] // Vec_Ptr_t (by value) : memory pages
	vCis     IntVector             // Vec_Int_t (by value) : CI IDs
	vCos     IntVector             // Vec_Int_t (by value) : CO IDs
	// object iterator
	pPage *MigObj // current page
	iPage int     // current page index
	// attributes
	nTravIds int       // traversal ID counter
	vTravIds IntVector // Vec_Int_t (by value) : traversal IDs
	vLevels  IntVector // Vec_Int_t (by value) : levels
	vSibls   IntVector // Vec_Int_t (by value) : choice nodes
	vRefs    IntVector // Vec_Int_t (by value) : ref counters
	vCopies  IntVector // Vec_Int_t (by value) : copies
	pMan     any       // void * : mapping manager
}

// MpmCut is Mpm_Cut_t (8 bytes + NLeaves * 4 bytes).
type MpmCut struct {
	hNext    int    // next cut
	iFunc    uint32 // unsigned : 25, function
	fCompl   uint32 // unsigned : 1
	fUseless uint32 // unsigned : 1, internal flag
	nLeaves  uint32 // unsigned : 5, leaves
	pLeaves  []int  // int pLeaves[1] : flexible array member, leaves
}

// MpmUni is Mpm_Uni_t (48 bytes).
type MpmUni struct {
	mTime    int         // arrival time
	mArea    int         // area (flow)
	mEdge    int         // edge (flow)
	mAveRefs int         // area references
	uSign    uint64      // word : cut signature
	Cost     int         // user cost
	pCut     MpmCut      // new cut
	Data     [12 - 1]int // padding
}

// MpmDsd is Mpm_Dsd_t.
type MpmDsd struct {
	nVars    int    // support size
	nAnds    int    // the number of AND gates
	nClauses int    // the number of CNF clauses
	uTruth   uint64 // word : truth table
	pStr     string // char * : description
}

// MpmMan is Mpm_Man_t.
type MpmMan struct {
	pMig  *MigMan // AIG manager
	pPars *MpmPar // mapping parameters
	// mapping parameters
	nLutSize  int        // LUT size
	nNumCuts  int        // cut count
	nTruWords int        // words in the truth table
	pLibLut   *MpmLibLut // LUT library
	// mapping attributes
	fMainRun    int    // after preprocessing is finished
	GloRequired int    // global arrival time
	GloArea     uint64 // word : total area
	GloEdge     uint64 // word : total edge
	// cut management
	pManCuts *MmrStepArena // Mmr_Step_t * : cut memory
	// temporary cut storage
	nCutStore  int                    // number of cuts in storage
	pCutStore  [32 + 1]*MpmUni        // storage for cuts
	pCutUnits  [32 + 1]MpmUni         // cut info units
	vFreeUnits PointerVector[MpmUni]  // Vec_Ptr_t (by value) : free cut info units
	vTemp      *PointerVector[MpmUni] // storage for cuts
	// cut comparison
	pCutCmp func(*MpmUni, *MpmUni) int // procedure to compare cuts
	// fanin cuts/signatures
	nCuts  [3]int             // fanin cut counts
	pCuts  [3][32 + 1]*MpmCut // fanin cuts
	pSigns [3][32 + 1]uint64  // word : fanin cut signatures
	// truth tables
	vTtMem   *WordMemory // Vec_Mem_t * : truth table memory and hash table
	funcCst0 int         // constant 0
	funcVar0 int         // variable 0
	Truth0   [1 << (12 - 6)]uint64
	Truth1   [1 << (12 - 6)]uint64
	TruthC   [1 << (12 - 6)]uint64
	Truth    [1 << (12 - 6)]uint64
	// DSD
	pDsd6       *MpmDsd                 // NPN class information
	pHash       *FixedIntRecordInterner // Hsh_IntMan_t * : maps DSD functions into NPN classes
	vConfgRes   *IntVector              // configurations
	vPerm6      *WordVector             // permutations of DSD classes
	Perm6       [720][6]int8            // char Perm6[720][6] : permutations
	vMap2Perm   *IntVector              // maps number into its permutation
	uPermMask   [3]uint32
	uComplMask  [3]uint32
	vNpnConfigs *IntLevelVector // Vec_Wec_t *
	// mapping attributes
	vCutBests  IntVector // Vec_Int_t (by value) : cut best
	vCutLists  IntVector // Vec_Int_t (by value) : cut list
	vMigRefs   IntVector // Vec_Int_t (by value) : original references
	vMapRefs   IntVector // Vec_Int_t (by value) : exact mapping references
	vEstRefs   IntVector // Vec_Int_t (by value) : estimated mapping references
	vRequireds IntVector // Vec_Int_t (by value) : required time
	vTimes     IntVector // Vec_Int_t (by value) : arrival time
	vAreas     IntVector // Vec_Int_t (by value) : area
	vEdges     IntVector // Vec_Int_t (by value) : edge
	nCountDsd  [600]int
	nNonDsd    int
	nNoMatch   int
	// statistics
	nCutsMerged    int
	nCutsMergedAll int
	nSmallSupp     int
	timeDerive     int64 // abctime
	timeMerge      int64
	timeEval       int64
	timeCompare    int64
	timeStore      int64
	timeOther      int64
	timeTotal      int64
}
