package abc

// Hand-written skeleton of types in src/sat/cnf/cnf.h (declarations only).

// CnfDat is Cnf_Dat_t: the CNF asserting outputs of AIG to be 1.
type CnfDat struct {
	pMan        *AigManager // the AIG manager, for which CNF is computed
	nVars       int         // the number of variables
	nLiterals   int         // the number of CNF literals
	nClauses    int         // the number of CNF clauses
	pClauses    [][]int     // int ** : the CNF clauses
	pVarNums    []int       // int * : the number of CNF variable for each node ID (-1 if unused)
	pObj2Clause []int       // int * : the mapping of objects into clauses
	pObj2Count  []int       // int * : the mapping of objects into clause number
	pClaPols    []uint8     // unsigned char * : polarity of input literals in each clause
	vMapping    *IntVector  // mapping of internal nodes
}

// CnfMan is Cnf_Man_t: the CNF computation manager.
type CnfMan struct {
	pManAig     *AigManager   // the underlying AIG manager
	pSopSizes   []int8        // char * : sizes of SOPs for 4-variable functions
	pSops       []string      // char ** : the SOPs for 4-variable functions
	aArea       int           // the area of the mapping
	pMemCuts    *AigFlexArena // Aig_MmFlex_t * : memory manager for cuts
	nMergeLimit int           // the limit on the size of merged cut
	pTruths     [4][]uint32   // unsigned * pTruths[4] : temporary truth tables
	vMemory     *IntVector    // memory for intermediate ISOP representation
	timeCuts    int64
	timeMap     int64
	timeSave    int64
}

// CnfCut is Cnf_Cut_t: the cut used to represent node in the AIG.
type CnfCut struct {
	nFanins int8          // the number of leaves
	Cost    int8          // the cost of this cut
	nWords  int16         // the number of words in truth table
	vIsop   [2]*IntVector // neg/pos ISOPs
	pFanins []int         // int pFanins[0] : flexible array member, the fanins (followed by the truth table)
}
