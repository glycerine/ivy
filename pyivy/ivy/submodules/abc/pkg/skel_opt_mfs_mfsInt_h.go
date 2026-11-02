package abc

// Hand-written skeleton of the types and static inline functions in
// src/opt/mfs/mfsInt.h. MFS_FANIN_MAX is 12. The block comment in the middle of
// the C struct (pGia, pTas, vCex, vGiaLits) is commented out in the source.

// MfsMan is Mfs_Man_t.
type MfsMan struct {
	// input data
	pPars     *MfsPar
	pNtk      *ABCNetwork
	pCare     *AigManager
	vSuppsInv *PointerVector[IntVector]
	nFaninMax int
	// intermeditate data for the node
	vRoots       *PointerVector[ABCNetworkObject] // the roots of the window
	vSupp        *PointerVector[ABCNetworkObject] // the support of the window
	vNodes       *PointerVector[ABCNetworkObject] // the internal nodes of the window
	vDivs        *PointerVector[ABCNetworkObject] // the divisors of the node
	vDivLits     *IntVector                       // the SAT literals of divisor nodes
	vProjVarsCnf *IntVector                       // the projection variables
	vProjVarsSat *IntVector                       // the projection variables
	// intermediate simulation data
	vDivCexes *PointerVector[uint32] // the counter-example for dividors
	nDivWords int                    // the number of words
	nCexes    int                    // the numbe rof current counter-examples
	nSatCalls int
	nSatCexes int
	// used for bidecomposition
	vTruth            *IntVector
	pManDec           *BdcTable // Bdc_Man_t *
	nNodesDec         int
	nNodesGained      int
	nNodesGainedLevel int
	// solving data
	pAigWin       *AigManager                      // window AIG with constraints
	pCnf          *CnfDat                          // the CNF for the window
	pSat          *SatSolver                       // the SAT solver used
	pMan          *SatTruthInterpolator            // Int_Man_t * : interpolation manager
	vMem          *IntVector                       // memory for intermediate SOPs
	vLevels       *NestedVector[*ABCNetworkObject] // Vec_Vec_t * : levelized structure for updating
	vMfsFanins    *PointerVector[ABCNetworkObject] // the new set of fanins
	nTotConfLim   int                              // total conflict limit
	nTotConfLevel int                              // total conflicts on this level
	// switching activity
	vProbs *IntVector
	// the result of solving
	nFanins int         // the number of fanins
	nWords  int         // the number of words
	nCares  int         // the number of care minterms
	uCare   [128]uint32 // unsigned uCare[(MFS_FANIN_MAX<=5)?1:1<<(MFS_FANIN_MAX-5)] : the computed care-set
	// performance statistics
	nTryRemoves    int // number of fanin removals
	nTryResubs     int // number of resubstitutions
	nRemoves       int // number of fanin removals
	nResubs        int // number of resubstitutions
	nNodesTried    int
	nNodesResub    int
	nMintsCare     int
	nMintsTotal    int
	nNodesBad      int
	nTotalDivs     int
	nTimeOuts      int
	nTimeOutsLevel int
	nDcMints       int
	nMaxDivs       int
	dTotalRatios   float64
	// node/edge stats
	nTotalNodesBeg    int
	nTotalNodesEnd    int
	nTotalEdgesBeg    int
	nTotalEdgesEnd    int
	TotalSwitchingBeg float32
	TotalSwitchingEnd float32
	// statistics
	timeWin   int64 // abctime
	timeDiv   int64
	timeAig   int64
	timeGia   int64
	timeCnf   int64
	timeSat   int64
	timeInt   int64
	timeTotal int64
}

// Abc_MfsObjProb (static inline)
// C: float Abc_MfsObjProb( Mfs_Man_t * p, Abc_Obj_t * pObj )
func (p *MfsMan) abcMfsObjProb(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}
