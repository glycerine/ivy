package abc

// Hand-checked skeleton of src/bdd/llb/llbInt.h.

// LlbMan is Llb_Man_t.
type LlbMan struct {
	pPars    *GiaParLlb             // parameters
	pAigGlo  *AigManager            // initial AIG manager (owned by the caller)
	pAig     *AigManager            // derived AIG manager (created in this package)
	dd       *CuddFoundation        // BDD manager
	ddG      *CuddFoundation        // BDD manager
	ddR      *CuddFoundation        // BDD manager
	vObj2Var *IntVector             // mapping AIG ObjId into BDD var index
	vVar2Obj *IntVector             // mapping BDD var index into AIG ObjId
	vGroups  *PointerVector[LlbGrp] // group Id into group pointer
	pMatrix  *LlbMtr                // dependency matrix
	// image computation
	vRings   *PointerVector[DdNode] // onion rings
	vVarBegs *IntVector             // the first group where the var appears
	vVarEnds *IntVector             // the last group where the var appears
	// variable mapping
	vNs2Glo *IntVector // next state variables into global variables
	vCs2Glo *IntVector // next state variables into global variables
	vGlo2Cs *IntVector // global variables into current state variables
	vGlo2Ns *IntVector // global variables into current state variables
}

// LlbMtr is Llb_Mtr_t.
type LlbMtr struct {
	nPis     int       // number of primary inputs
	nFfs     int       // number of flip-flops
	nRows    int       // number of rows
	nCols    int       // number of columns
	pColSums []int     // int * : sum of values in a column
	pColGrps []*LlbGrp // Llb_Grp_t ** : group structure for each col
	pRowSums []int     // int * : sum of values in a row
	pMatrix  [][]int8  // char ** : dependency matrix
	pMan     *LlbMan   // manager
	// partial product
	pProdVars []int8 // char * : variables in the partial product
	pProdNums []int  // int * : var counts in the remaining partitions
}

// LlbGrp is Llb_Grp_t.
type LlbGrp struct {
	Id     int                       // group ID
	vIns   *PointerVector[AigObject] // input AIG objs
	vOuts  *PointerVector[AigObject] // output AIG objs
	vNodes *PointerVector[AigObject] // internal AIG objs
	pMan   *LlbMan                   // manager
	pPrev  *LlbGrp                   // previous group
	pNext  *LlbGrp                   // next group
}

// Llb_ObjBddVar (static inline)
// C: int Llb_ObjBddVar( Vec_Int_t * vOrder, Aig_Obj_t * pObj )
func llbObjBddVar(vOrder *IntVector, pObj *AigObject) int {
	panic("TODO, stubbed")
}
