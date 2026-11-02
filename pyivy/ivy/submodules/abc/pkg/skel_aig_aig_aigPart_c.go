package abc

// Hand-checked skeleton of src/aig/aig/aigPart.c.

// The `Vec_PtrForEachEntry( Vec_Int_t *, vSupports, vOne, i )` entry that
// ctags reports is a macro invocation, not a function.

// PartMan is Part_Man_t.
type PartMan struct {
	nChunkSize int                  // the size of one chunk of memory (~1 MB)
	nStepSize  int                  // the step size in saving memory (~64 bytes)
	pFreeBuf   []byte               // char * : the pointer to free memory
	nFreeSize  int                  // the size of remaining free memory
	vMemory    *PointerVector[byte] // the memory allocated (char * chunks)
	vFree      *PointerVector[byte] // the vector of free pieces of memory (char * list heads)
}

// PartOne is Part_One_t.
type PartOne struct {
	nRefs      int   // the number of references
	nOuts      int   // the number of outputs
	nOutsAlloc int   // the array size
	pOuts      []int // int pOuts[0] : flexible array member, the array of outputs
}

// Part_SizeType (static)
// C: int Part_SizeType( int nSize, int nStepSize )
func partSizeType(nSize int, nStepSize int) int {
	panic("TODO, stubbed")
}

// Part_OneNext (static; the next free piece is stored in the first bytes of the piece)
// C: char * Part_OneNext( char * pPart )
func partOneNext(pPart []byte) []byte {
	panic("TODO, stubbed")
}

// Part_OneSetNext (static)
// C: void Part_OneSetNext( char * pPart, char * pNext )
func partOneSetNext(pPart []byte, pNext []byte) {
	panic("TODO, stubbed")
}

// Part_ManStart
// C: Part_Man_t * Part_ManStart( int nChunkSize, int nStepSize )
func PartManStart(nChunkSize int, nStepSize int) *PartMan {
	panic("TODO, stubbed")
}

// Part_ManStop
// C: void Part_ManStop( Part_Man_t * p )
func (p *PartMan) PartManStop() {
	panic("TODO, stubbed")
}

// Part_ManFetch
// C: char * Part_ManFetch( Part_Man_t * p, int nSize )
func (p *PartMan) PartManFetch(nSize int) []byte {
	panic("TODO, stubbed")
}

// Part_ManRecycle
// C: void Part_ManRecycle( Part_Man_t * p, char * pMemory, int nSize )
func (p *PartMan) PartManRecycle(pMemory []byte, nSize int) {
	panic("TODO, stubbed")
}

// Part_ManFetchEntry (static)
// C: Part_One_t * Part_ManFetchEntry( Part_Man_t * p, int nWords, int nRefs )
func (p *PartMan) partManFetchEntry(nWords int, nRefs int) *PartOne {
	panic("TODO, stubbed")
}

// Part_ManRecycleEntry (static)
// C: void Part_ManRecycleEntry( Part_Man_t * p, Part_One_t * pEntry )
func (p *PartMan) partManRecycleEntry(pEntry *PartOne) {
	panic("TODO, stubbed")
}

// Part_ManMergeEntry
// C: Part_One_t * Part_ManMergeEntry( Part_Man_t * pMan, Part_One_t * p1, Part_One_t * p2, int nRefs )
func (pMan *PartMan) PartManMergeEntry(p1 *PartOne, p2 *PartOne, nRefs int) *PartOne {
	panic("TODO, stubbed")
}

// Part_ManTransferEntry
// C: Vec_Int_t * Part_ManTransferEntry( Part_One_t * p )
func PartManTransferEntry(p *PartOne) *IntVector {
	panic("TODO, stubbed")
}

// Aig_ManSupports (one Vec_Int_t * support per CO)
// C: Vec_Ptr_t * Aig_ManSupports( Aig_Man_t * pMan )
func (pMan *AigManager) AigManSupports() *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManSupportsTest
// C: void Aig_ManSupportsTest( Aig_Man_t * pMan )
func (pMan *AigManager) AigManSupportsTest() {
	panic("TODO, stubbed")
}

// Aig_ManSupportsInverse
// C: Vec_Ptr_t * Aig_ManSupportsInverse( Aig_Man_t * p )
func (p *AigManager) AigManSupportsInverse() *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManSupportsRegisters
// C: Vec_Ptr_t * Aig_ManSupportsRegisters( Aig_Man_t * p )
func (p *AigManager) AigManSupportsRegisters() *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManSuppCharStart
// C: unsigned * Aig_ManSuppCharStart( Vec_Int_t * vOne, int nPis )
func AigManSuppCharStart(vOne *IntVector, nPis int) []uint32 {
	panic("TODO, stubbed")
}

// Aig_ManSuppCharAdd
// C: void Aig_ManSuppCharAdd( unsigned * pBuffer, Vec_Int_t * vOne, int nPis )
func AigManSuppCharAdd(pBuffer []uint32, vOne *IntVector, nPis int) {
	panic("TODO, stubbed")
}

// Aig_ManSuppCharCommon
// C: int Aig_ManSuppCharCommon( unsigned * pBuffer, Vec_Int_t * vOne )
func AigManSuppCharCommon(pBuffer []uint32, vOne *IntVector) int {
	panic("TODO, stubbed")
}

// Aig_ManPartitionSmartFindPart
// C: int Aig_ManPartitionSmartFindPart( Vec_Ptr_t * vPartSuppsAll, Vec_Ptr_t * vPartsAll, Vec_Ptr_t * vPartSuppsBit, int nSuppSizeLimit, Vec_Int_t * vOne )
func AigManPartitionSmartFindPart(vPartSuppsAll *PointerVector[IntVector], vPartsAll *PointerVector[IntVector], vPartSuppsBit *PointerVector[uint32], nSuppSizeLimit int, vOne *IntVector) int {
	panic("TODO, stubbed")
}

// Aig_ManPartitionPrint
// C: void Aig_ManPartitionPrint( Aig_Man_t * p, Vec_Ptr_t * vPartsAll, Vec_Ptr_t * vPartSuppsAll )
func (p *AigManager) AigManPartitionPrint(vPartsAll *PointerVector[IntVector], vPartSuppsAll *PointerVector[IntVector]) {
	panic("TODO, stubbed")
}

// Aig_ManPartitionCompact
// C: void Aig_ManPartitionCompact( Vec_Ptr_t * vPartsAll, Vec_Ptr_t * vPartSuppsAll, int nSuppSizeLimit )
func AigManPartitionCompact(vPartsAll *PointerVector[IntVector], vPartSuppsAll *PointerVector[IntVector], nSuppSizeLimit int) {
	panic("TODO, stubbed")
}

// Aig_ManPartitionSmart (pvPartSupps is an out parameter (nullable))
// C: Vec_Ptr_t * Aig_ManPartitionSmart( Aig_Man_t * p, int nSuppSizeLimit, int fVerbose, Vec_Ptr_t ** pvPartSupps )
func (p *AigManager) AigManPartitionSmart(nSuppSizeLimit int, fVerbose int, pvPartSupps **PointerVector[IntVector]) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManPartitionSmartRegisters
// C: Vec_Ptr_t * Aig_ManPartitionSmartRegisters( Aig_Man_t * pAig, int nSuppSizeLimit, int fVerbose )
func (pAig *AigManager) AigManPartitionSmartRegisters(nSuppSizeLimit int, fVerbose int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManPartitionNaive
// C: Vec_Ptr_t * Aig_ManPartitionNaive( Aig_Man_t * p, int nPartSize )
func (p *AigManager) AigManPartitionNaive(nPartSize int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManDupPart_rec
// C: Aig_Obj_t * Aig_ManDupPart_rec( Aig_Man_t * pNew, Aig_Man_t * pOld, Aig_Obj_t * pObj, Vec_Int_t * vSuppMap )
func (pNew *AigManager) AigManDupPart_rec(pOld *AigManager, pObj *AigObject, vSuppMap *IntVector) *AigObject {
	panic("TODO, stubbed")
}

// Aig_ManDupPart (returns the (possibly complemented) child copies)
// C: Vec_Ptr_t * Aig_ManDupPart( Aig_Man_t * pNew, Aig_Man_t * pOld, Vec_Int_t * vPart, Vec_Int_t * vSuppMap, int fInverse )
func (pNew *AigManager) AigManDupPart(pOld *AigManager, vPart *IntVector, vSuppMap *IntVector, fInverse int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Aig_ManDupPartAll_rec
// C: void Aig_ManDupPartAll_rec( Aig_Man_t * pNew, Aig_Man_t * pOld, Aig_Obj_t * pObj )
func (pNew *AigManager) AigManDupPartAll_rec(pOld *AigManager, pObj *AigObject) {
	panic("TODO, stubbed")
}

// Aig_ManDupPartAll
// C: Aig_Man_t * Aig_ManDupPartAll( Aig_Man_t * pOld, Vec_Int_t * vPart )
func (pOld *AigManager) AigManDupPartAll(vPart *IntVector) *AigManager {
	panic("TODO, stubbed")
}

// Aig_ManSupportNodes_rec
// C: void Aig_ManSupportNodes_rec( Aig_Man_t * p, Aig_Obj_t * pObj, Vec_Int_t * vSupport )
func (p *AigManager) AigManSupportNodes_rec(pObj *AigObject, vSupport *IntVector) {
	panic("TODO, stubbed")
}

// Aig_ManSupportNodes
// C: Vec_Ptr_t * Aig_ManSupportNodes( Aig_Man_t * p, Vec_Ptr_t * vParts )
func (p *AigManager) AigManSupportNodes(vParts *PointerVector[IntVector]) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManMiterPartitioned
// C: Vec_Ptr_t * Aig_ManMiterPartitioned( Aig_Man_t * p1, Aig_Man_t * p2, int nPartSize, int fSmart )
func (p1 *AigManager) AigManMiterPartitioned(p2 *AigManager, nPartSize int, fSmart int) *PointerVector[AigManager] {
	panic("TODO, stubbed")
}

// Aig_ManChoicePartitioned
// C: Aig_Man_t * Aig_ManChoicePartitioned( Vec_Ptr_t * vAigs, int nPartSize, int nConfMax, int nLevelMax, int fVerbose )
func AigManChoicePartitioned(vAigs *PointerVector[AigManager], nPartSize int, nConfMax int, nLevelMax int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Aig_ManFraigPartitioned
// C: Aig_Man_t * Aig_ManFraigPartitioned( Aig_Man_t * pAig, int nPartSize, int nConfMax, int nLevelMax, int fVerbose )
func (pAig *AigManager) AigManFraigPartitioned(nPartSize int, nConfMax int, nLevelMax int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Aig_ObjSetRepr_ (static)
// C: void Aig_ObjSetRepr_( Aig_Man_t * p, Aig_Obj_t * pNode1, Aig_Obj_t * pNode2 )
func (p *AigManager) aigObjSetRepr_(pNode1 *AigObject, pNode2 *AigObject) {
	panic("TODO, stubbed")
}

// Aig_ManChoiceConstructiveOne
// C: void Aig_ManChoiceConstructiveOne( Aig_Man_t * pNew, Aig_Man_t * pPrev, Aig_Man_t * pThis )
func (pNew *AigManager) AigManChoiceConstructiveOne(pPrev *AigManager, pThis *AigManager) {
	panic("TODO, stubbed")
}

// Aig_ManChoiceEval
// C: void Aig_ManChoiceEval( Aig_Man_t * p )
func (p *AigManager) AigManChoiceEval() {
	panic("TODO, stubbed")
}

// Aig_ManChoiceConstructive
// C: Aig_Man_t * Aig_ManChoiceConstructive( Vec_Ptr_t * vAigs, int fVerbose )
func AigManChoiceConstructive(vAigs *PointerVector[AigManager], fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
