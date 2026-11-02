package abc

// Hand-written skeleton of the types in src/proof/pdr/pdrInt.h.

// PdrSet is Pdr_Set_t.
type PdrSet struct {
	Sign   uint64 // word : signature
	nRefs  int    // ref counter
	nTotal int    // total literals
	nLits  int    // num flop literals
	Lits   []int  // int Lits[0] : flexible array member
}

// PdrObl is Pdr_Obl_t.
type PdrObl struct {
	iFrame int     // time frame
	prio   int     // priority
	nRefs  int     // reference counter
	pState *PdrSet // state cube
	pNext  *PdrObl // next one
	pLink  *PdrObl // queue link
}

// PdrMan is Pdr_Man_t.
type PdrMan struct {
	// input problem
	pPars *PdrPar     // parameters
	pAig  *AigManager // user's AIG
	pGia  *GiaManager // user's AIG
	// static CNF representation
	pCnfMan  *CnfMan    // CNF manager
	pCnf1    *CnfDat    // CNF for this AIG
	vVar2Reg *IntVector // mapping of SAT var into registers
	// dynamic CNF representation
	pCnf2     *CnfDat                  // CNF for this AIG
	pvId2Vars []IntVector              // Vec_Int_t * : array indexed by ObjId; for each used ObjId, maps frame into SAT var
	vVar2Ids  PointerVector[IntVector] // Vec_Ptr_t (by value) : for each used frame, maps SAT var into ObjId
	vVLits    *IntLevelVector          // Vec_Wec_t * : CNF literals
	// data representation
	iOutCur     int                            // current output
	nPrioShift  int                            // priority shift
	vCexes      *PointerVector[Counterexample] // counter-examples for each output
	vSolvers    *PointerVector[SatSolver]      // SAT solvers
	vClauses    *NestedVector[*PdrSet]         // Vec_Vec_t * : clauses by timeframe
	pQueue      *PdrObl                        // proof obligations
	pOrder      []int                          // int * : ordering of the lits
	vActVars    *IntVector                     // the counter of activation variables
	iUseFrame   int                            // the first used frame
	nAbsFlops   int                            // the number of flops used
	vAbsFlops   *IntVector                     // flops currently used
	vMapFf2Ppi  *IntVector
	vMapPpi2Ff  *IntVector
	nCexes      int
	nCexesTotal int
	// terminary simulation
	pTxs3 *Txs3Man
	// internal use
	vPrio      *IntVector             // priority flops
	vLits      *IntVector             // array of literals
	vCiObjs    *IntVector             // cone leaves
	vCoObjs    *IntVector             // cone roots
	vCiVals    *IntVector             // cone leaf values
	vCoVals    *IntVector             // cone root values
	vNodes     *IntVector             // cone nodes
	vUndo      *IntVector             // cone undos
	vVisits    *IntVector             // intermediate
	vCi2Rem    *IntVector             // CIs to be removed
	vRes       *IntVector             // final result
	pTime4Outs []int64                // abctime * : timeout per output
	vInfCubes  *PointerVector[PdrSet] // infinity clauses/cubes
	// statistics
	nBlocks   int // the number of times blockState was called
	nObligs   int // the number of proof obligations derived
	nCubes    int // the number of cubes derived
	nCalls    int // the number of SAT calls
	nCallsS   int // the number of SAT calls (sat)
	nCallsU   int // the number of SAT calls (unsat)
	nStarts   int // the number of SAT solver restarts
	nFrames   int // frames explored
	nCasesSS  int
	nCasesSU  int
	nCasesUS  int
	nCasesUU  int
	nQueCur   int
	nQueMax   int
	nQueLim   int
	nXsimRuns int
	nXsimLits int
	// runtime
	timeToStop    int64
	timeToStopOne int64
	// time stats
	tSat      int64
	tSatSat   int64
	tSatUnsat int64
	tGeneral  int64
	tPush     int64
	tTsim     int64
	tContain  int64
	tCnf      int64
	tAbs      int64
	tTotal    int64
}
