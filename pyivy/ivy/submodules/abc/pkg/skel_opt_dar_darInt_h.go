package abc

// Hand-written skeleton of the types and static inline functions in
// src/opt/dar/darInt.h. The iterator macros (Dar_ObjForEachCutAll,
// Dar_ObjForEachCut, Dar_CutForEachLeaf) are macros and are not represented.
// Go has no bitfields; each is a uint32 whose C width is in the comment.

// DarCut is Dar_Cut_t: the AIG 4-cut (6 words).
type DarCut struct {
	uSign   uint32 // cut signature
	uTruth  uint32 // unsigned : 16, the truth table of the cut function
	Value   uint32 // unsigned : 11, the value of the cut
	fBest   uint32 // unsigned : 1, marks the best cut
	fUsed   uint32 // unsigned : 1, marks the cut currently in use
	nLeaves uint32 // unsigned : 3, the number of leaves
	pLeaves [4]int // the array of leaves
}

// DarMan is Dar_Man_t: the AIG manager.
type DarMan struct {
	// input data
	pPars *DarRwrPar  // rewriting parameters
	pAig  *AigManager // AIG manager
	// various data members
	pMemCuts  *AigFixedArena            // Aig_MmFixed_t * : memory manager for cuts
	pManCnf   *CnfMan                   // void * : CNF managers
	vCutNodes *PointerVector[AigObject] // the nodes with cuts allocated
	// current rewriting step
	vLeavesBest *PointerVector[AigObject] // the best set of leaves
	OutBest     int                       // the best output (in the library)
	OutNumBest  int                       // the best number of the output
	GainBest    int                       // the best gain
	LevelBest   int                       // the level of node with the best gain
	ClassBest   int                       // the equivalence class of the best replacement
	// function statistics
	nTotalSubgs int      // the total number of subgraphs tried
	ClassTimes  [222]int // the runtimes for each class
	ClassGains  [222]int // the gains for each class
	ClassSubgs  [222]int // the graphs for each class
	nCutMemUsed int      // memory used for cuts
	// rewriting statistics
	nNodesInit   int // the original number of nodes
	nNodesTried  int // the number of nodes attempted
	nCutsAll     int // all cut pairs
	nCutsTried   int // computed cuts
	nCutsUsed    int // used cuts
	nCutsBad     int // bad cuts due to absent fanin
	nCutsGood    int // good cuts
	nCutsSkipped int // skipped bad cuts
	// timing statistics
	timeCuts  int64 // abctime
	timeEval  int64
	timeOther int64
	timeTotal int64
	time1     int64
	time2     int64
}

// Dar_ObjCuts (static inline)
// C: Dar_Cut_t * Dar_ObjCuts( Aig_Obj_t * pObj )
// Returns the array of cuts stored in pObj->pData.
func darObjCuts(pObj *AigObject) []DarCut {
	panic("TODO, stubbed")
}

// Dar_ObjSetCuts (static inline)
// C: void Dar_ObjSetCuts( Aig_Obj_t * pObj, Dar_Cut_t * pCuts )
func darObjSetCuts(pObj *AigObject, pCuts []DarCut) {
	panic("TODO, stubbed")
}
