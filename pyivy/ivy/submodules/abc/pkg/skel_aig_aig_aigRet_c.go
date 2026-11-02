package abc

// Hand-checked skeleton of src/aig/aig/aigRet.c.

// The iterator macros (Rtm_ManForEachObj, Rtm_ObjForEachFanin, ...) are macros,
// not functions, and are not represented. Go has no bitfields; the C widths
// are in the comments.

// RtmInit is the Rtm_Init_t enum: init values.
type RtmInit int

const (
	RTM_VAL_NONE RtmInit = iota // 0: non-existent value
	RTM_VAL_ZERO                // 1: initial value 0
	RTM_VAL_ONE                 // 2: initial value 1
	RTM_VAL_VOID                // 3: unused value
)

// RtmMan is Rtm_Man_t.
type RtmMan struct {
	// network representation
	vObjs *PointerVector[RtmObj] // retiming objects
	vPis  *PointerVector[RtmObj] // PIs only
	vPos  *PointerVector[RtmObj] // POs only
	pMem  *AigFlexArena          // Aig_MmFlex_t * : the memory manager
	// autonomous components after cutting off
	// storage for overflow latches
	pExtra      []uint32 // unsigned *
	nExtraCur   int
	nExtraAlloc int
}

// RtmEdg is Rtm_Edg_t.
type RtmEdg struct {
	nLats uint64 // unsigned long : 12, the number of latches
	LData uint64 // unsigned long : 20, the latches themselves
}

// RtmObj is Rtm_Obj_t.
type RtmObj struct {
	pCopy    any    // void * : the copy of this object
	Type     uint64 // unsigned long : 3, object type
	fMark    uint64 // unsigned long : 1, multipurpose mark
	fAuto    uint64 // unsigned long : 1, this object belongs to an autonomous component
	fCompl0  uint64 // unsigned long : 1, complemented attribute of the first edge
	fCompl1  uint64 // unsigned long : 1, complemented attribute of the second edge
	nFanins  uint64 // unsigned long : 8, the number of fanins
	Num      uint32 // unsigned : 17, the retiming number of this node
	Id       int    // ID of this object
	Temp     int    // temporary usage
	nFanouts int    // the number of fanouts
	pFanio   []any  // void * pFanio[0] : flexible array member, fanins and their edges (followed by fanouts and pointers to their edges)
}

// Rtm_ObjFanin (static)
// C: Rtm_Obj_t * Rtm_ObjFanin( Rtm_Obj_t * pObj, int i )
func rtmObjFanin(pObj *RtmObj, i int) *RtmObj {
	panic("TODO, stubbed")
}

// Rtm_ObjFanout (static)
// C: Rtm_Obj_t * Rtm_ObjFanout( Rtm_Obj_t * pObj, int i )
func rtmObjFanout(pObj *RtmObj, i int) *RtmObj {
	panic("TODO, stubbed")
}

// Rtm_ObjEdge (static)
// C: Rtm_Edg_t * Rtm_ObjEdge( Rtm_Obj_t * pObj, int i )
func rtmObjEdge(pObj *RtmObj, i int) *RtmEdg {
	panic("TODO, stubbed")
}

// Rtm_ObjFanoutEdge (static)
// C: Rtm_Edg_t * Rtm_ObjFanoutEdge( Rtm_Obj_t * pObj, int i )
func rtmObjFanoutEdge(pObj *RtmObj, i int) *RtmEdg {
	panic("TODO, stubbed")
}

// Rtm_InitNot (static)
// C: Rtm_Init_t Rtm_InitNot( Rtm_Init_t Val )
func rtmInitNot(Val RtmInit) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_InitNotCond (static)
// C: Rtm_Init_t Rtm_InitNotCond( Rtm_Init_t Val, int c )
func rtmInitNotCond(Val RtmInit, c int) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_InitAnd (static)
// C: Rtm_Init_t Rtm_InitAnd(Rtm_Init_t ValA, Rtm_Init_t ValB )
func rtmInitAnd(ValA RtmInit, ValB RtmInit) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_InitWordsNum (static)
// C: int Rtm_InitWordsNum( int nLats )
func rtmInitWordsNum(nLats int) int {
	panic("TODO, stubbed")
}

// Rtm_InitGetTwo (static)
// C: int Rtm_InitGetTwo( unsigned * p, int i )
func rtmInitGetTwo(p []uint32, i int) int {
	panic("TODO, stubbed")
}

// Rtm_InitSetTwo (static)
// C: void Rtm_InitSetTwo( unsigned * p, int i, int val )
func rtmInitSetTwo(p []uint32, i int, val int) {
	panic("TODO, stubbed")
}

// Rtm_InitXorTwo (static)
// C: void Rtm_InitXorTwo( unsigned * p, int i, int val )
func rtmInitXorTwo(p []uint32, i int, val int) {
	panic("TODO, stubbed")
}

// Rtm_ObjGetFirst1 (static)
// C: Rtm_Init_t Rtm_ObjGetFirst1( Rtm_Edg_t * pEdge )
func rtmObjGetFirst1(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjGetLast1 (static)
// C: Rtm_Init_t Rtm_ObjGetLast1( Rtm_Edg_t * pEdge )
func rtmObjGetLast1(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjGetOne1 (static)
// C: Rtm_Init_t Rtm_ObjGetOne1( Rtm_Edg_t * pEdge, int i )
func rtmObjGetOne1(pEdge *RtmEdg, i int) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjRemFirst1 (static)
// C: Rtm_Init_t Rtm_ObjRemFirst1( Rtm_Edg_t * pEdge )
func rtmObjRemFirst1(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjRemLast1 (static)
// C: Rtm_Init_t Rtm_ObjRemLast1( Rtm_Edg_t * pEdge )
func rtmObjRemLast1(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjAddFirst1 (static)
// C: void Rtm_ObjAddFirst1( Rtm_Edg_t * pEdge, Rtm_Init_t Val )
func rtmObjAddFirst1(pEdge *RtmEdg, Val RtmInit) {
	panic("TODO, stubbed")
}

// Rtm_ObjAddLast1 (static)
// C: void Rtm_ObjAddLast1( Rtm_Edg_t * pEdge, Rtm_Init_t Val )
func rtmObjAddLast1(pEdge *RtmEdg, Val RtmInit) {
	panic("TODO, stubbed")
}

// Rtm_ObjGetFirst2 (static)
// C: Rtm_Init_t Rtm_ObjGetFirst2( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjGetFirst2(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjGetLast2 (static)
// C: Rtm_Init_t Rtm_ObjGetLast2( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjGetLast2(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjGetOne2 (static)
// C: Rtm_Init_t Rtm_ObjGetOne2( Rtm_Man_t * p, Rtm_Edg_t * pEdge, int i )
func (p *RtmMan) rtmObjGetOne2(pEdge *RtmEdg, i int) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjRemLast2 (static)
// C: Rtm_Init_t Rtm_ObjRemLast2( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjRemLast2(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjAddLast2 (static)
// C: void Rtm_ObjAddLast2( Rtm_Man_t * p, Rtm_Edg_t * pEdge, Rtm_Init_t Val )
func (p *RtmMan) rtmObjAddLast2(pEdge *RtmEdg, Val RtmInit) {
	panic("TODO, stubbed")
}

// Rtm_ObjGetFirst (static)
// C: Rtm_Init_t Rtm_ObjGetFirst( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjGetFirst(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjGetLast (static)
// C: Rtm_Init_t Rtm_ObjGetLast( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjGetLast(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjGetOne (static)
// C: Rtm_Init_t Rtm_ObjGetOne( Rtm_Man_t * p, Rtm_Edg_t * pEdge, int i )
func (p *RtmMan) rtmObjGetOne(pEdge *RtmEdg, i int) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjRemFirst (static)
// C: Rtm_Init_t Rtm_ObjRemFirst( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjRemFirst(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjRemLast (static)
// C: Rtm_Init_t Rtm_ObjRemLast( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjRemLast(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjAddFirst (static)
// C: void Rtm_ObjAddFirst( Rtm_Man_t * p, Rtm_Edg_t * pEdge, Rtm_Init_t Val )
func (p *RtmMan) rtmObjAddFirst(pEdge *RtmEdg, Val RtmInit) {
	panic("TODO, stubbed")
}

// Rtm_ObjAddLast (static)
// C: void Rtm_ObjAddLast( Rtm_Man_t * p, Rtm_Edg_t * pEdge, Rtm_Init_t Val )
func (p *RtmMan) rtmObjAddLast(pEdge *RtmEdg, Val RtmInit) {
	panic("TODO, stubbed")
}

// Rtm_ObjTransferToSmall (static)
// C: void Rtm_ObjTransferToSmall( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjTransferToSmall(pEdge *RtmEdg) {
	panic("TODO, stubbed")
}

// Rtm_ObjTransferToBig (static)
// C: void Rtm_ObjTransferToBig( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjTransferToBig(pEdge *RtmEdg) {
	panic("TODO, stubbed")
}

// Rtm_ObjTransferToBigger (static)
// C: void Rtm_ObjTransferToBigger( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjTransferToBigger(pEdge *RtmEdg) {
	panic("TODO, stubbed")
}

// Rtm_ObjRemFirst2 (static)
// C: Rtm_Init_t Rtm_ObjRemFirst2( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) rtmObjRemFirst2(pEdge *RtmEdg) RtmInit {
	panic("TODO, stubbed")
}

// Rtm_ObjAddFirst2 (static)
// C: void Rtm_ObjAddFirst2( Rtm_Man_t * p, Rtm_Edg_t * pEdge, Rtm_Init_t Val )
func (p *RtmMan) rtmObjAddFirst2(pEdge *RtmEdg, Val RtmInit) {
	panic("TODO, stubbed")
}

// Rtm_PrintEdge
// C: void Rtm_PrintEdge( Rtm_Man_t * p, Rtm_Edg_t * pEdge )
func (p *RtmMan) RtmPrintEdge(pEdge *RtmEdg) {
	panic("TODO, stubbed")
}

// Rtm_ManAlloc
// C: Rtm_Man_t * Rtm_ManAlloc( Aig_Man_t * p )
func (p *AigManager) RtmManAlloc() *RtmMan {
	panic("TODO, stubbed")
}

// Rtm_ManFree
// C: void Rtm_ManFree( Rtm_Man_t * p )
func (p *RtmMan) RtmManFree() {
	panic("TODO, stubbed")
}

// Rtm_ManLatchMax
// C: int Rtm_ManLatchMax( Rtm_Man_t * p )
func (p *RtmMan) RtmManLatchMax() int {
	panic("TODO, stubbed")
}

// Rtm_ObjAlloc
// C: Rtm_Obj_t * Rtm_ObjAlloc( Rtm_Man_t * pRtm, int nFanins, int nFanouts )
func (pRtm *RtmMan) RtmObjAlloc(nFanins int, nFanouts int) *RtmObj {
	panic("TODO, stubbed")
}

// Rtm_ObjAddFanin
// C: void Rtm_ObjAddFanin( Rtm_Obj_t * pObj, Rtm_Obj_t * pFanin, int fCompl )
func (pObj *RtmObj) RtmObjAddFanin(pFanin *RtmObj, fCompl int) {
	panic("TODO, stubbed")
}

// Rtm_ObjCheckRetimeFwd
// C: int Rtm_ObjCheckRetimeFwd( Rtm_Obj_t * pObj )
func (pObj *RtmObj) RtmObjCheckRetimeFwd() int {
	panic("TODO, stubbed")
}

// Rtm_ObjCheckRetimeBwd
// C: int Rtm_ObjCheckRetimeBwd( Rtm_Obj_t * pObj )
func (pObj *RtmObj) RtmObjCheckRetimeBwd() int {
	panic("TODO, stubbed")
}

// Rtm_ObjGetDegreeFwd
// C: int Rtm_ObjGetDegreeFwd( Rtm_Obj_t * pObj )
func (pObj *RtmObj) RtmObjGetDegreeFwd() int {
	panic("TODO, stubbed")
}

// Rtm_ObjGetDegreeBwd
// C: int Rtm_ObjGetDegreeBwd( Rtm_Obj_t * pObj )
func (pObj *RtmObj) RtmObjGetDegreeBwd() int {
	panic("TODO, stubbed")
}

// Rtm_ObjRetimeFwd
// C: void Rtm_ObjRetimeFwd( Rtm_Man_t * pRtm, Rtm_Obj_t * pObj )
func (pRtm *RtmMan) RtmObjRetimeFwd(pObj *RtmObj) {
	panic("TODO, stubbed")
}

// Rtm_ObjRetimeBwd
// C: void Rtm_ObjRetimeBwd( Rtm_Man_t * pRtm, Rtm_Obj_t * pObj )
func (pRtm *RtmMan) RtmObjRetimeBwd(pObj *RtmObj) {
	panic("TODO, stubbed")
}

// Rtm_ObjMarkAutoFwd_rec
// C: void Rtm_ObjMarkAutoFwd_rec( Rtm_Obj_t * pObj )
func (pObj *RtmObj) RtmObjMarkAutoFwd_rec() {
	panic("TODO, stubbed")
}

// Rtm_ManMarkAutoFwd
// C: int Rtm_ManMarkAutoFwd( Rtm_Man_t * pRtm )
func (pRtm *RtmMan) RtmManMarkAutoFwd() int {
	panic("TODO, stubbed")
}

// Rtm_ObjMarkAutoBwd_rec
// C: void Rtm_ObjMarkAutoBwd_rec( Rtm_Obj_t * pObj )
func (pObj *RtmObj) RtmObjMarkAutoBwd_rec() {
	panic("TODO, stubbed")
}

// Rtm_ManMarkAutoBwd
// C: int Rtm_ManMarkAutoBwd( Rtm_Man_t * pRtm )
func (pRtm *RtmMan) RtmManMarkAutoBwd() int {
	panic("TODO, stubbed")
}

// Rtm_ManFromAig
// C: Rtm_Man_t * Rtm_ManFromAig( Aig_Man_t * p )
func (p *AigManager) RtmManFromAig() *RtmMan {
	panic("TODO, stubbed")
}

// Rtm_ManToAig_rec (pLatches points at the latch count of the current edge)
// C: Aig_Obj_t * Rtm_ManToAig_rec( Aig_Man_t * pNew, Rtm_Man_t * pRtm, Rtm_Obj_t * pObjRtm, int * pLatches )
func (pNew *AigManager) RtmManToAig_rec(pRtm *RtmMan, pObjRtm *RtmObj, pLatches *int) *AigObject {
	panic("TODO, stubbed")
}

// Rtm_ManToAig
// C: Aig_Man_t * Rtm_ManToAig( Rtm_Man_t * pRtm )
func (pRtm *RtmMan) RtmManToAig() *AigManager {
	panic("TODO, stubbed")
}

// Rtm_ManRetime
// C: Aig_Man_t * Rtm_ManRetime( Aig_Man_t * p, int fForward, int nStepsMax, int fVerbose )
func (p *AigManager) RtmManRetime(fForward int, nStepsMax int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
