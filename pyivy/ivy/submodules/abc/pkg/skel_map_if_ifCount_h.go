package abc

// Hand-checked skeleton of src/map/if/ifCount.h.

// If_LogCreateAnd (static inline)
// C: int If_LogCreateAnd( Vec_Int_t * vAig, int iLit0, int iLit1, int nSuppAll )
func ifLogCreateAnd(vAig *IntVector, iLit0 int, iLit1 int, nSuppAll int) int {
	panic("TODO, stubbed")
}

// If_LogCreateMux (static inline)
// C: int If_LogCreateMux( Vec_Int_t * vAig, int iLitC, int iLit1, int iLit0, int nSuppAll )
func ifLogCreateMux(vAig *IntVector, iLitC int, iLit1 int, iLit0 int, nSuppAll int) int {
	panic("TODO, stubbed")
}

// If_LogCreateXor (static inline)
// C: int If_LogCreateXor( Vec_Int_t * vAig, int iLit0, int iLit1, int nSuppAll )
func ifLogCreateXor(vAig *IntVector, iLit0 int, iLit1 int, nSuppAll int) int {
	panic("TODO, stubbed")
}

// If_LogCreateAndXor (static inline)
// C: int If_LogCreateAndXor( Vec_Int_t * vAig, int iLit0, int iLit1, int nSuppAll, int fXor )
func ifLogCreateAndXor(vAig *IntVector, iLit0 int, iLit1 int, nSuppAll int, fXor int) int {
	panic("TODO, stubbed")
}

// If_LogCreateAndXorMulti (static inline)
// C: int If_LogCreateAndXorMulti( Vec_Int_t * vAig, int * pFaninLits, int nFanins, int nSuppAll, int fXor )
func ifLogCreateAndXorMulti(vAig *IntVector, pFaninLits []int, nFanins int, nSuppAll int, fXor int) int {
	panic("TODO, stubbed")
}

// If_LogCounterAddAig (static inline)
// C: int If_LogCounterAddAig( int * pTimes, int * pnTimes, int * pFaninLits, int Num, int iLit, Vec_Int_t * vAig, int nSuppAll, int fXor, int fXorFunc )
func ifLogCounterAddAig(pTimes []int, pnTimes *int, pFaninLits []int, Num int, iLit int, vAig *IntVector, nSuppAll int, fXor int, fXorFunc int) int {
	panic("TODO, stubbed")
}

// If_LogCounterDelayXor (static inline)
// C: int If_LogCounterDelayXor( int * pTimes, int nTimes )
func ifLogCounterDelayXor(pTimes []int, nTimes int) int {
	panic("TODO, stubbed")
}

// If_CutPinDelayGet (static inline)
// C: int If_CutPinDelayGet( word D, int v )
func ifCutPinDelayGet(D uint64, v int) int {
	panic("TODO, stubbed")
}

// If_CutPinDelaySet (static inline)
// C: void If_CutPinDelaySet( word * pD, int v, int d )
func ifCutPinDelaySet(pD *uint64, v int, d int) {
	panic("TODO, stubbed")
}

// If_CutPinDelayInit (static inline)
// C: word If_CutPinDelayInit( int v )
func ifCutPinDelayInit(v int) uint64 {
	panic("TODO, stubbed")
}

// If_CutPinDelayMax (static inline)
// C: word If_CutPinDelayMax( word D1, word D2, int nVars, int AddOn )
func ifCutPinDelayMax(D1 uint64, D2 uint64, nVars int, AddOn int) uint64 {
	panic("TODO, stubbed")
}

// If_CutPinDelayDecrement (static inline)
// C: word If_CutPinDelayDecrement( word D1, int nVars )
func ifCutPinDelayDecrement(D1 uint64, nVars int) uint64 {
	panic("TODO, stubbed")
}

// If_CutPinDelayEqual (static inline)
// C: int If_CutPinDelayEqual( word D1, word D2, int nVars )
func ifCutPinDelayEqual(D1 uint64, D2 uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CutPinDelayDom (static inline)
// C: int If_CutPinDelayDom( word D1, word D2, int nVars )
func ifCutPinDelayDom(D1 uint64, D2 uint64, nVars int) int {
	panic("TODO, stubbed")
}

// If_CutPinDelayTranslate (static inline)
// C: void If_CutPinDelayTranslate( word D, int nVars, char * pPerm )
func ifCutPinDelayTranslate(D uint64, nVars int, pPerm []int8) {
	panic("TODO, stubbed")
}

// If_CutPinDelayPrint (static inline)
// C: void If_CutPinDelayPrint( word D, int nVars )
func ifCutPinDelayPrint(D uint64, nVars int) {
	panic("TODO, stubbed")
}

// If_LogCounterPinDelays (static inline)
// C: int If_LogCounterPinDelays( int * pTimes, int * pnTimes, word * pPinDels, int Num, word PinDel, int nSuppAll, int fXor )
func ifLogCounterPinDelays(pTimes []int, pnTimes *int, pPinDels []uint64, Num int, PinDel uint64, nSuppAll int, fXor int) int {
	panic("TODO, stubbed")
}

// If_LogPinDelaysMulti (static inline)
// C: word If_LogPinDelaysMulti( word * pPinDels, int nFanins, int nSuppAll, int fXor )
func ifLogPinDelaysMulti(pPinDels []uint64, nFanins int, nSuppAll int, fXor int) uint64 {
	panic("TODO, stubbed")
}

// If_AigVerifyArray (static inline)
// C: word If_AigVerifyArray( Vec_Int_t * vAig, int nLeaves )
func ifAigVerifyArray(vAig *IntVector, nLeaves int) uint64 {
	panic("TODO, stubbed")
}

// If_AigPrintArray (static inline)
// C: void If_AigPrintArray( Vec_Int_t * vAig, int nLeaves )
func ifAigPrintArray(vAig *IntVector, nLeaves int) {
	panic("TODO, stubbed")
}

// If_LogCounter64Eval (static inline)
// C: int If_LogCounter64Eval( word Count )
func ifLogCounter64Eval(Count uint64) int {
	panic("TODO, stubbed")
}

// If_LogCounter64Add (static inline)
// C: word If_LogCounter64Add( word Count, int Num )
func ifLogCounter64Add(Count uint64, Num int) uint64 {
	panic("TODO, stubbed")
}

// If_LogCounter32Eval (static inline)
// C: int If_LogCounter32Eval( unsigned Count, int Start )
func ifLogCounter32Eval(Count uint32, Start int) int {
	panic("TODO, stubbed")
}

// If_LogCounter32Add (static inline)
// C: unsigned If_LogCounter32Add( unsigned Count, int * pStart, int Num )
func ifLogCounter32Add(Count uint32, pStart *int, Num int) uint32 {
	panic("TODO, stubbed")
}

// If_LogCounterAdd (static inline)
// C: int If_LogCounterAdd( int * pTimes, int * pnTimes, int Num, int fXor )
func ifLogCounterAdd(pTimes []int, pnTimes *int, Num int, fXor int) int {
	panic("TODO, stubbed")
}
