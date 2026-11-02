package abc

// Hand-written skeleton of the type in src/proof/ssw/sswClass.c.

// SswCla is Ssw_Cla_t.
type SswCla struct {
	// class information
	pAig        *AigManager    // original AIG manager
	pId2Class   [][]*AigObject // Aig_Obj_t *** : non-const classes by ID of repr node
	pClassSizes []int          // int * : sizes of each equivalence class
	fConstCorr  int
	// statistics
	nClasses int // the total number of non-const classes
	nCands1  int // the total number of const candidates
	nLits    int // the number of literals in all classes
	// memory
	pMemClasses     []*AigObject // Aig_Obj_t ** : memory allocated for equivalence classes
	pMemClassesFree []*AigObject // Aig_Obj_t ** : memory allocated for equivalence classes to be used
	// temporary data
	vClassOld *PointerVector[AigObject] // old equivalence class after splitting
	vClassNew *PointerVector[AigObject] // new equivalence class(es) after splitting
	vRefined  *PointerVector[AigObject] // the nodes refined since the last iteration
	// procedures used for class refinement
	pManData           any                                   // void *
	pFuncNodeHash      func(any, *AigObject) uint32          // returns hash key of the node
	pFuncNodeIsConst   func(any, *AigObject) int             // returns 1 if the node is a constant
	pFuncNodesAreEqual func(any, *AigObject, *AigObject) int // returns 1 if nodes are equal up to a complement
}
