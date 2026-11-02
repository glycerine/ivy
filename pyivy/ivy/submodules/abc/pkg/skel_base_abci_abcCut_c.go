package abc

// Hand-checked skeleton of src/base/abci/abcCut.c.

// Abc_NtkCutsSubtractFanunt
// C: void Abc_NtkCutsSubtractFanunt( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkCutsSubtractFanunt() {
	panic("TODO, stubbed")
}

// Abc_NtkCutsAddFanunt
// C: void Abc_NtkCutsAddFanunt( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkCutsAddFanunt() {
	panic("TODO, stubbed")
}

// Abc_NtkCuts
// C: Cut_Man_t * Abc_NtkCuts( Abc_Ntk_t * pNtk, Cut_Params_t * pParams )
func (pNtk *ABCNetwork) AbcNtkCuts(pParams *CutParams) *CutMan {
	panic("TODO, stubbed")
}

// Abc_NtkCutsOracle
// C: void Abc_NtkCutsOracle( Abc_Ntk_t * pNtk, Cut_Oracle_t * p )
func (pNtk *ABCNetwork) AbcNtkCutsOracle(p *CutOracle) {
	panic("TODO, stubbed")
}

// Abc_NtkSeqCuts
// C: Cut_Man_t * Abc_NtkSeqCuts( Abc_Ntk_t * pNtk, Cut_Params_t * pParams )
func (pNtk *ABCNetwork) AbcNtkSeqCuts(pParams *CutParams) *CutMan {
	panic("TODO, stubbed")
}

// Abc_NtkComputeArea (static)
// C: int Abc_NtkComputeArea( Abc_Ntk_t * pNtk, Cut_Man_t * p )
func (pNtk *ABCNetwork) abcNtkComputeArea(p *CutMan) int {
	panic("TODO, stubbed")
}

// Abc_NodeGetCutsRecursive (p is a void * that holds a Cut_Man_t *; returns a void * that holds a Cut_Cut_t *)
// C: void * Abc_NodeGetCutsRecursive( void * p, Abc_Obj_t * pObj, int fDag, int fTree )
func AbcNodeGetCutsRecursive(p *CutMan, pObj *ABCNetworkObject, fDag int, fTree int) *CutCut {
	panic("TODO, stubbed")
}

// Abc_NodeGetCuts (p is a void * that holds a Cut_Man_t *; returns a void * that holds a Cut_Cut_t *)
// C: void * Abc_NodeGetCuts( void * p, Abc_Obj_t * pObj, int fDag, int fTree )
func AbcNodeGetCuts(p *CutMan, pObj *ABCNetworkObject, fDag int, fTree int) *CutCut {
	panic("TODO, stubbed")
}

// Abc_NodeGetCutsSeq (p is a void * that holds a Cut_Man_t *)
// C: void Abc_NodeGetCutsSeq( void * p, Abc_Obj_t * pObj, int fTriv )
func AbcNodeGetCutsSeq(p *CutMan, pObj *ABCNetworkObject, fTriv int) {
	panic("TODO, stubbed")
}

// Abc_NodeReadCuts (p is a void * that holds a Cut_Man_t *)
// C: void * Abc_NodeReadCuts( void * p, Abc_Obj_t * pObj )
func AbcNodeReadCuts(p *CutMan, pObj *ABCNetworkObject) *CutCut {
	panic("TODO, stubbed")
}

// Abc_NodeFreeCuts (p is a void * that holds a Cut_Man_t *)
// C: void Abc_NodeFreeCuts( void * p, Abc_Obj_t * pObj )
func AbcNodeFreeCuts(p *CutMan, pObj *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintCuts (static; p is a void * that holds a Cut_Man_t *)
// C: void Abc_NtkPrintCuts( void * p, Abc_Ntk_t * pNtk, int fSeq )
func abcNtkPrintCuts(p *CutMan, pNtk *ABCNetwork, fSeq int) {
	panic("TODO, stubbed")
}

// Abc_NtkPrintCuts_ (static; p is a void * that holds a Cut_Man_t *)
// C: void Abc_NtkPrintCuts_( void * p, Abc_Ntk_t * pNtk, int fSeq )
func abcNtkPrintCuts_(p *CutMan, pNtk *ABCNetwork, fSeq int) {
	panic("TODO, stubbed")
}

// Abc_NtkGetNodeAttributes (static)
// C: Vec_Int_t * Abc_NtkGetNodeAttributes( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) abcNtkGetNodeAttributes() *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkSubDagSize_rec
// C: int Abc_NtkSubDagSize_rec( Abc_Obj_t * pObj, Vec_Int_t * vAttrs )
func AbcNtkSubDagSize_rec(pObj *ABCNetworkObject, vAttrs *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkGetNodeAttributes2
// C: Vec_Int_t * Abc_NtkGetNodeAttributes2( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkGetNodeAttributes2() *IntVector {
	panic("TODO, stubbed")
}
