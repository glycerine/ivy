package abc

// Hand-checked skeleton of src/aig/saig/saigStrSim.c.

const SAIG_WORDS = 16

// Saig_ObjNext (static inline)
// C: Aig_Obj_t * Saig_ObjNext( Aig_Obj_t ** ppNexts, Aig_Obj_t * pObj )
func saigObjNext(ppNexts []*AigObject, pObj *AigObject) *AigObject {
	panic("TODO, stubbed")
}

// Saig_ObjSetNext (static inline)
// C: void Saig_ObjSetNext( Aig_Obj_t ** ppNexts, Aig_Obj_t * pObj, Aig_Obj_t * pNext )
func saigObjSetNext(ppNexts []*AigObject, pObj *AigObject, pNext *AigObject) {
	panic("TODO, stubbed")
}

// Saig_StrSimHash
// C: unsigned Saig_StrSimHash( Aig_Obj_t * pObj )
func SaigStrSimHash(pObj *AigObject) uint32 {
	panic("TODO, stubbed")
}

// Saig_StrSimIsEqual
// C: int Saig_StrSimIsEqual( Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 )
func SaigStrSimIsEqual(pObj0 *AigObject, pObj1 *AigObject) int {
	panic("TODO, stubbed")
}

// Saig_StrSimIsZero
// C: int Saig_StrSimIsZero( Aig_Obj_t * pObj )
func SaigStrSimIsZero(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Saig_StrSimIsOne
// C: int Saig_StrSimIsOne( Aig_Obj_t * pObj )
func SaigStrSimIsOne(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Saig_StrSimAssignRandom
// C: void Saig_StrSimAssignRandom( Aig_Obj_t * pObj )
func SaigStrSimAssignRandom(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_StrSimAssignOne
// C: void Saig_StrSimAssignOne( Aig_Obj_t * pObj )
func SaigStrSimAssignOne(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_StrSimAssignZeroInit
// C: void Saig_StrSimAssignZeroInit( Aig_Obj_t * pObj )
func SaigStrSimAssignZeroInit(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_StrSimulateNode
// C: void Saig_StrSimulateNode( Aig_Obj_t * pObj, int i )
func SaigStrSimulateNode(pObj *AigObject, i int) {
	panic("TODO, stubbed")
}

// Saig_StrSimSaveOutput
// C: void Saig_StrSimSaveOutput( Aig_Obj_t * pObj, int i )
func SaigStrSimSaveOutput(pObj *AigObject, i int) {
	panic("TODO, stubbed")
}

// Saig_StrSimTransfer
// C: void Saig_StrSimTransfer( Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 )
func SaigStrSimTransfer(pObj0 *AigObject, pObj1 *AigObject) {
	panic("TODO, stubbed")
}

// Saig_StrSimTransferNext
// C: void Saig_StrSimTransferNext( Aig_Obj_t * pObj0, Aig_Obj_t * pObj1, int i )
func SaigStrSimTransferNext(pObj0 *AigObject, pObj1 *AigObject, i int) {
	panic("TODO, stubbed")
}

// Saig_StrSimulateRound
// C: void Saig_StrSimulateRound( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SaigStrSimulateRound(p1 *AigManager) {
	panic("TODO, stubbed")
}

// Saig_StrSimTableLookup
// C: Aig_Obj_t * Saig_StrSimTableLookup( Aig_Obj_t ** ppTable, Aig_Obj_t ** ppNexts, int nTableSize, Aig_Obj_t * pObj )
func SaigStrSimTableLookup(ppTable []*AigObject, ppNexts []*AigObject, nTableSize int, pObj *AigObject) *AigObject {
	panic("TODO, stubbed")
}

// Saig_StrSimTableInsert
// C: void Saig_StrSimTableInsert( Aig_Obj_t ** ppTable, Aig_Obj_t ** ppNexts, int nTableSize, Aig_Obj_t * pObj )
func SaigStrSimTableInsert(ppTable []*AigObject, ppNexts []*AigObject, nTableSize int, pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_StrSimDetectUnique
// C: int Saig_StrSimDetectUnique( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SaigStrSimDetectUnique(p1 *AigManager) int {
	panic("TODO, stubbed")
}

// Saig_StrSimCountMatchedFlops
// C: int Saig_StrSimCountMatchedFlops( Aig_Man_t * p )
func (p *AigManager) SaigStrSimCountMatchedFlops() int {
	panic("TODO, stubbed")
}

// Saig_StrSimCountMatchedNodes
// C: int Saig_StrSimCountMatchedNodes( Aig_Man_t * p )
func (p *AigManager) SaigStrSimCountMatchedNodes() int {
	panic("TODO, stubbed")
}

// Saig_StrSimPrepareAig
// C: void Saig_StrSimPrepareAig( Aig_Man_t * p )
func (p *AigManager) SaigStrSimPrepareAig() {
	panic("TODO, stubbed")
}

// Saig_StrSimSetInitMatching
// C: void Saig_StrSimSetInitMatching( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SaigStrSimSetInitMatching(p1 *AigManager) {
	panic("TODO, stubbed")
}

// Saig_StrSimSetFinalMatching
// C: void Saig_StrSimSetFinalMatching( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SaigStrSimSetFinalMatching(p1 *AigManager) {
	panic("TODO, stubbed")
}

// Saig_StrSimSetContiguousMatching_rec
// C: void Saig_StrSimSetContiguousMatching_rec( Aig_Man_t * p, Aig_Obj_t * pObj )
func (p *AigManager) SaigStrSimSetContiguousMatching_rec(pObj *AigObject) {
	panic("TODO, stubbed")
}

// Saig_StrSimSetContiguousMatching
// C: void Saig_StrSimSetContiguousMatching( Aig_Man_t * p0, Aig_Man_t * p1 )
func (p0 *AigManager) SaigStrSimSetContiguousMatching(p1 *AigManager) {
	panic("TODO, stubbed")
}

// Ssw_StrSimMatchingExtendOne
// C: void Ssw_StrSimMatchingExtendOne( Aig_Man_t * p, Vec_Ptr_t * vNodes )
func (p *AigManager) SswStrSimMatchingExtendOne(vNodes *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Ssw_StrSimMatchingCountUnmached
// C: int Ssw_StrSimMatchingCountUnmached( Aig_Man_t * p )
func (p *AigManager) SswStrSimMatchingCountUnmached() int {
	panic("TODO, stubbed")
}

// Ssw_StrSimMatchingExtend
// C: void Ssw_StrSimMatchingExtend( Aig_Man_t * p0, Aig_Man_t * p1, int nDist, int fVerbose )
func (p0 *AigManager) SswStrSimMatchingExtend(p1 *AigManager, nDist int, fVerbose int) {
	panic("TODO, stubbed")
}

// Saig_StrSimPerformMatching
// C: Vec_Int_t * Saig_StrSimPerformMatching( Aig_Man_t * p0, Aig_Man_t * p1, int nDist, int fVerbose, Aig_Man_t ** ppMiter )
func (p0 *AigManager) SaigStrSimPerformMatching(p1 *AigManager, nDist int, fVerbose int, ppMiter **AigManager) *IntVector {
	panic("TODO, stubbed")
}
