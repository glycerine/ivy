package abc

// Hand-checked skeleton of src/base/abci/abcMiter.c.

// AddFrameMapping is typedef void (*AddFrameMapping)( Abc_Obj_t*, Abc_Obj_t*, int, void*).
type AddFrameMapping = func(*ABCNetworkObject, *ABCNetworkObject, int, any)

// Abc_NtkMiter
// C: Abc_Ntk_t * Abc_NtkMiter( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int fComb, int nPartSize, int fImplic, int fMulti )
func (pNtk1 *ABCNetwork) AbcNtkMiter(pNtk2 *ABCNetwork, fComb int, nPartSize int, fImplic int, fMulti int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMiterInt (static)
// C: Abc_Ntk_t * Abc_NtkMiterInt( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int fComb, int nPartSize, int fImplic, int fMulti )
func (pNtk1 *ABCNetwork) abcNtkMiterInt(pNtk2 *ABCNetwork, fComb int, nPartSize int, fImplic int, fMulti int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMiterPrepare (static)
// C: void Abc_NtkMiterPrepare( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Abc_Ntk_t * pNtkMiter, int fComb, int nPartSize, int fMulti )
func (pNtk1 *ABCNetwork) abcNtkMiterPrepare(pNtk2 *ABCNetwork, pNtkMiter *ABCNetwork, fComb int, nPartSize int, fMulti int) {
	panic("TODO, stubbed")
}

// Abc_NtkMiterAddOne (static)
// C: void Abc_NtkMiterAddOne( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkMiter )
func (pNtk *ABCNetwork) abcNtkMiterAddOne(pNtkMiter *ABCNetwork) {
	panic("TODO, stubbed")
}

// Abc_NtkMiterAddCone
// C: void Abc_NtkMiterAddCone( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkMiter, Abc_Obj_t * pRoot )
func (pNtk *ABCNetwork) AbcNtkMiterAddCone(pNtkMiter *ABCNetwork, pRoot *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_NtkMiterFinalize (static)
// C: void Abc_NtkMiterFinalize( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, Abc_Ntk_t * pNtkMiter, int fComb, int nPartSize, int fImplic, int fMulti )
func (pNtk1 *ABCNetwork) abcNtkMiterFinalize(pNtk2 *ABCNetwork, pNtkMiter *ABCNetwork, fComb int, nPartSize int, fImplic int, fMulti int) {
	panic("TODO, stubbed")
}

// Abc_NtkMiterAnd
// C: Abc_Ntk_t * Abc_NtkMiterAnd( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int fOr, int fCompl2 )
func (pNtk1 *ABCNetwork) AbcNtkMiterAnd(pNtk2 *ABCNetwork, fOr int, fCompl2 int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMiterCofactor
// C: Abc_Ntk_t * Abc_NtkMiterCofactor( Abc_Ntk_t * pNtk, Vec_Int_t * vPiValues )
func (pNtk *ABCNetwork) AbcNtkMiterCofactor(vPiValues *IntVector) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMiterForCofactors
// C: Abc_Ntk_t * Abc_NtkMiterForCofactors( Abc_Ntk_t * pNtk, int Out, int In1, int In2 )
func (pNtk *ABCNetwork) AbcNtkMiterForCofactors(Out int, In1 int, In2 int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMiterQuantify
// C: Abc_Ntk_t * Abc_NtkMiterQuantify( Abc_Ntk_t * pNtk, int In, int fExist )
func (pNtk *ABCNetwork) AbcNtkMiterQuantify(In int, fExist int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMiterQuantifyPis
// C: Abc_Ntk_t * Abc_NtkMiterQuantifyPis( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkMiterQuantifyPis() *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkMiterIsConstant
// C: int Abc_NtkMiterIsConstant( Abc_Ntk_t * pMiter )
func (pMiter *ABCNetwork) AbcNtkMiterIsConstant() int {
	panic("TODO, stubbed")
}

// Abc_NtkMiterReport
// C: void Abc_NtkMiterReport( Abc_Ntk_t * pMiter )
func (pMiter *ABCNetwork) AbcNtkMiterReport() {
	panic("TODO, stubbed")
}

// Abc_NtkFrames
// C: Abc_Ntk_t * Abc_NtkFrames( Abc_Ntk_t * pNtk, int nFrames, int fInitial, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkFrames(nFrames int, fInitial int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkAddFrame (static)
// C: void Abc_NtkAddFrame( Abc_Ntk_t * pNtkFrames, Abc_Ntk_t * pNtk, int iFrame )
func (pNtkFrames *ABCNetwork) abcNtkAddFrame(pNtk *ABCNetwork, iFrame int) {
	panic("TODO, stubbed")
}

// Abc_NtkFrames2 (arg is a void *)
// C: Abc_Ntk_t * Abc_NtkFrames2( Abc_Ntk_t * pNtk, int nFrames, int fInitial, AddFrameMapping addFrameMapping, void* arg )
func (pNtk *ABCNetwork) AbcNtkFrames2(nFrames int, fInitial int, addFrameMapping AddFrameMapping, arg any) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkAddFrame2 (static; arg is a void *)
// C: void Abc_NtkAddFrame2( Abc_Ntk_t * pNtkFrames, Abc_Ntk_t * pNtk, int iFrame, Vec_Ptr_t * vNodes, AddFrameMapping addFrameMapping, void* arg )
func (pNtkFrames *ABCNetwork) abcNtkAddFrame2(pNtk *ABCNetwork, iFrame int, vNodes *PointerVector[ABCNetworkObject], addFrameMapping AddFrameMapping, arg any) {
	panic("TODO, stubbed")
}

// Abc_NtkDemiter
// C: int Abc_NtkDemiter( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDemiter() int {
	panic("TODO, stubbed")
}

// Abc_NtkCombinePos
// C: int Abc_NtkCombinePos( Abc_Ntk_t * pNtk, int fAnd, int fXor )
func (pNtk *ABCNetwork) AbcNtkCombinePos(fAnd int, fXor int) int {
	panic("TODO, stubbed")
}

// Abc_NtkTryNewMiter
// C: Vec_Ptr_t * Abc_NtkTryNewMiter( char * pFileName0, char * pFileName1 )
func AbcNtkTryNewMiter(pFileName0 string, pFileName1 string) *PointerVector[Counterexample] {
	panic("TODO, stubbed")
}

// Abc_NtkReadNodeNames
// C: Vec_Ptr_t * Abc_NtkReadNodeNames( Abc_Ntk_t * pNtk, char * pFileName )
func (pNtk *ABCNetwork) AbcNtkReadNodeNames(pFileName string) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkSpecialMuxTree_rec
// C: Abc_Obj_t * Abc_NtkSpecialMuxTree_rec( Abc_Ntk_t * pNew, Abc_Obj_t ** pCtrl, int nCtrl, Abc_Obj_t ** pData, int Shift )
func (pNew *ABCNetwork) AbcNtkSpecialMuxTree_rec(pCtrl []*ABCNetworkObject, nCtrl int, pData []*ABCNetworkObject, Shift int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkSpecialMiter
// C: Abc_Ntk_t * Abc_NtkSpecialMiter( Abc_Ntk_t * pNtk, Vec_Ptr_t * vNodes )
func (pNtk *ABCNetwork) AbcNtkSpecialMiter(vNodes *PointerVector[ABCNetworkObject]) *ABCNetwork {
	panic("TODO, stubbed")
}
