package abc

// Hand-checked skeleton of src/base/abci/abcPart.c.

// SuppMan is Supp_Man_t.
type SuppMan struct {
	nChunkSize int                  // the size of one chunk of memory (~1 MB)
	nStepSize  int                  // the step size in saving memory (~64 bytes)
	pFreeBuf   []byte               // char * : the pointer to free memory
	nFreeSize  int                  // the size of remaining free memory
	vMemory    *PointerVector[byte] // the memory allocated
	vFree      *PointerVector[byte] // the vector of free pieces of memory
}

// SuppOne is Supp_One_t.
type SuppOne struct {
	nRefs      int   // the number of references
	nOuts      int   // the number of outputs
	nOutsAlloc int   // the array size
	pOuts      []int // int pOuts[0] : the array of outputs (flexible array member)
}

// Supp_SizeType (static)
// C: int Supp_SizeType( int nSize, int nStepSize )
func suppSizeType(nSize int, nStepSize int) int {
	panic("TODO, stubbed")
}

// Supp_OneNext (static)
// C: char * Supp_OneNext( char * pPart )
func suppOneNext(pPart []byte) []byte {
	panic("TODO, stubbed")
}

// Supp_OneSetNext (static)
// C: void Supp_OneSetNext( char * pPart, char * pNext )
func suppOneSetNext(pPart []byte, pNext []byte) {
	panic("TODO, stubbed")
}

// Supp_ManStart
// C: Supp_Man_t * Supp_ManStart( int nChunkSize, int nStepSize )
func SuppManStart(nChunkSize int, nStepSize int) *SuppMan {
	panic("TODO, stubbed")
}

// Supp_ManStop
// C: void Supp_ManStop( Supp_Man_t * p )
func (p *SuppMan) SuppManStop() {
	panic("TODO, stubbed")
}

// Supp_ManFetch
// C: char * Supp_ManFetch( Supp_Man_t * p, int nSize )
func (p *SuppMan) SuppManFetch(nSize int) []byte {
	panic("TODO, stubbed")
}

// Supp_ManRecycle
// C: void Supp_ManRecycle( Supp_Man_t * p, char * pMemory, int nSize )
func (p *SuppMan) SuppManRecycle(pMemory []byte, nSize int) {
	panic("TODO, stubbed")
}

// Supp_ManFetchEntry (static)
// C: Supp_One_t * Supp_ManFetchEntry( Supp_Man_t * p, int nWords, int nRefs )
func (p *SuppMan) suppManFetchEntry(nWords int, nRefs int) *SuppOne {
	panic("TODO, stubbed")
}

// Supp_ManRecycleEntry (static)
// C: void Supp_ManRecycleEntry( Supp_Man_t * p, Supp_One_t * pEntry )
func (p *SuppMan) suppManRecycleEntry(pEntry *SuppOne) {
	panic("TODO, stubbed")
}

// Supp_ManMergeEntry
// C: Supp_One_t * Supp_ManMergeEntry( Supp_Man_t * pMan, Supp_One_t * p1, Supp_One_t * p2, int nRefs )
func (pMan *SuppMan) SuppManMergeEntry(p1 *SuppOne, p2 *SuppOne, nRefs int) *SuppOne {
	panic("TODO, stubbed")
}

// Supp_ManTransferEntry
// C: Vec_Int_t * Supp_ManTransferEntry( Supp_One_t * p )
func SuppManTransferEntry(p *SuppOne) *IntVector {
	panic("TODO, stubbed")
}

// Abc_NtkDfsNatural
// C: Vec_Ptr_t * Abc_NtkDfsNatural( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkDfsNatural() *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Abc_NtkComputeSupportsSmart
// C: Vec_Ptr_t * Abc_NtkComputeSupportsSmart( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkComputeSupportsSmart() *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Abc_NtkComputeSupportsNaive
// C: Vec_Ptr_t * Abc_NtkComputeSupportsNaive( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkComputeSupportsNaive() *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Abc_NtkSuppCharStart
// C: unsigned * Abc_NtkSuppCharStart( Vec_Int_t * vOne, int nPis )
func AbcNtkSuppCharStart(vOne *IntVector, nPis int) []uint32 {
	panic("TODO, stubbed")
}

// Abc_NtkSuppCharAdd
// C: void Abc_NtkSuppCharAdd( unsigned * pBuffer, Vec_Int_t * vOne, int nPis )
func AbcNtkSuppCharAdd(pBuffer []uint32, vOne *IntVector, nPis int) {
	panic("TODO, stubbed")
}

// Abc_NtkSuppCharCommon
// C: int Abc_NtkSuppCharCommon( unsigned * pBuffer, Vec_Int_t * vOne )
func AbcNtkSuppCharCommon(pBuffer []uint32, vOne *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkPartitionSmartFindPart
// C: int Abc_NtkPartitionSmartFindPart( Vec_Ptr_t * vPartSuppsAll, Vec_Ptr_t * vPartsAll, Vec_Ptr_t * vPartSuppsChar, int nSuppSizeLimit, Vec_Int_t * vOne )
func AbcNtkPartitionSmartFindPart(vPartSuppsAll *PointerVector[IntVector], vPartsAll *PointerVector[IntVector], vPartSuppsChar *PointerVector[uint32], nSuppSizeLimit int, vOne *IntVector) int {
	panic("TODO, stubbed")
}

// Abc_NtkPartitionPrint
// C: void Abc_NtkPartitionPrint( Abc_Ntk_t * pNtk, Vec_Ptr_t * vPartsAll, Vec_Ptr_t * vPartSuppsAll )
func (pNtk *ABCNetwork) AbcNtkPartitionPrint(vPartsAll *PointerVector[IntVector], vPartSuppsAll *PointerVector[IntVector]) {
	panic("TODO, stubbed")
}

// Abc_NtkPartitionCompact
// C: void Abc_NtkPartitionCompact( Vec_Ptr_t * vPartsAll, Vec_Ptr_t * vPartSuppsAll, int nSuppSizeLimit )
func AbcNtkPartitionCompact(vPartsAll *PointerVector[IntVector], vPartSuppsAll *PointerVector[IntVector], nSuppSizeLimit int) {
	panic("TODO, stubbed")
}

// Abc_NtkPartitionSmart
// C: Vec_Ptr_t * Abc_NtkPartitionSmart( Abc_Ntk_t * pNtk, int nSuppSizeLimit, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkPartitionSmart(nSuppSizeLimit int, fVerbose int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Abc_NtkPartitionNaive
// C: Vec_Ptr_t * Abc_NtkPartitionNaive( Abc_Ntk_t * pNtk, int nPartSize )
func (pNtk *ABCNetwork) AbcNtkPartitionNaive(nPartSize int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Abc_NtkConvertCos
// C: void Abc_NtkConvertCos( Abc_Ntk_t * pNtk, Vec_Int_t * vOuts, Vec_Ptr_t * vOutsPtr )
func (pNtk *ABCNetwork) AbcNtkConvertCos(vOuts *IntVector, vOutsPtr *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkPartStitchFindRepr_rec
// C: Abc_Obj_t * Abc_NtkPartStitchFindRepr_rec( Vec_Ptr_t * vEquiv, Abc_Obj_t * pObj )
func AbcNtkPartStitchFindRepr_rec(vEquiv *PointerVector[ABCNetworkObject], pObj *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkPartStitchCopy0 (static)
// C: Abc_Obj_t * Abc_NtkPartStitchCopy0( Vec_Ptr_t * vEquiv, Abc_Obj_t * pObj )
func abcNtkPartStitchCopy0(vEquiv *PointerVector[ABCNetworkObject], pObj *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NtkPartStitchCopy1 (static)
// C: Abc_Obj_t * Abc_NtkPartStitchCopy1( Vec_Ptr_t * vEquiv, Abc_Obj_t * pObj )
func abcNtkPartStitchCopy1(vEquiv *PointerVector[ABCNetworkObject], pObj *ABCNetworkObject) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Hop_ObjChild0Next (static)
// C: Hop_Obj_t * Hop_ObjChild0Next( Abc_Obj_t * pObj )
func hopObjChild0Next_abcPart(pObj *ABCNetworkObject) *HopObject {
	panic("TODO, stubbed")
}

// Hop_ObjChild1Next (static)
// C: Hop_Obj_t * Hop_ObjChild1Next( Abc_Obj_t * pObj )
func hopObjChild1Next_abcPart(pObj *ABCNetworkObject) *HopObject {
	panic("TODO, stubbed")
}

// Abc_NtkPartStartHop
// C: Hop_Man_t * Abc_NtkPartStartHop( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcNtkPartStartHop() *HopManager {
	panic("TODO, stubbed")
}

// Abc_NtkPartStitchChoices
// C: Abc_Ntk_t * Abc_NtkPartStitchChoices( Abc_Ntk_t * pNtk, Vec_Ptr_t * vParts )
func (pNtk *ABCNetwork) AbcNtkPartStitchChoices(vParts *PointerVector[ABCNetwork]) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFraigPartitioned (pParams is a void * that holds a Fraig_Params_t *)
// C: Abc_Ntk_t * Abc_NtkFraigPartitioned( Vec_Ptr_t * vStore, void * pParams )
func AbcNtkFraigPartitioned(vStore *PointerVector[ABCNetwork], pParams *FraigPar) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkFraigPartitionedTime (pParams is a void * that holds a Fraig_Params_t *)
// C: void Abc_NtkFraigPartitionedTime( Abc_Ntk_t * pNtk, void * pParams )
func (pNtk *ABCNetwork) AbcNtkFraigPartitionedTime(pParams *FraigPar) {
	panic("TODO, stubbed")
}
