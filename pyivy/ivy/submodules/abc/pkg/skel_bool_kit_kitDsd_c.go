package abc

import (
	"os"
	"strings"
)

// Hand-checked skeleton of src/bool/kit/kitDsd.c.

// Kit_DsdManAlloc
// C: Kit_DsdMan_t * Kit_DsdManAlloc( int nVars, int nNodes )
func KitDsdManAlloc(nVars int, nNodes int) *KitDsdMan {
	panic("TODO, stubbed")
}

// Kit_DsdManFree
// C: void Kit_DsdManFree( Kit_DsdMan_t * p )
func (p *KitDsdMan) KitDsdManFree() {
	panic("TODO, stubbed")
}

// Kit_DsdObjAlloc
// C: Kit_DsdObj_t * Kit_DsdObjAlloc( Kit_DsdNtk_t * pNtk, Kit_Dsd_t Type, int nFans )
func (pNtk *KitDsdNtk) KitDsdObjAlloc(Type KitDsd, nFans int) *KitDsdObj {
	panic("TODO, stubbed")
}

// Kit_DsdObjFree
// C: void Kit_DsdObjFree( Kit_DsdNtk_t * p, Kit_DsdObj_t * pObj )
func (p *KitDsdNtk) KitDsdObjFree(pObj *KitDsdObj) {
	panic("TODO, stubbed")
}

// Kit_DsdNtkAlloc
// C: Kit_DsdNtk_t * Kit_DsdNtkAlloc( int nVars )
func KitDsdNtkAlloc(nVars int) *KitDsdNtk {
	panic("TODO, stubbed")
}

// Kit_DsdNtkFree
// C: void Kit_DsdNtkFree( Kit_DsdNtk_t * pNtk )
func (pNtk *KitDsdNtk) KitDsdNtkFree() {
	panic("TODO, stubbed")
}

// Kit_DsdPrintHex
// C: void Kit_DsdPrintHex( FILE * pFile, unsigned * pTruth, int nFans )
func KitDsdPrintHex(pFile *os.File, pTruth []uint32, nFans int) {
	panic("TODO, stubbed")
}

// Kit_DsdWriteHex (pBuff is written through and the advanced pointer is returned; one builder serves both)
// C: char * Kit_DsdWriteHex( char * pBuff, unsigned * pTruth, int nFans )
func KitDsdWriteHex(pBuff *strings.Builder, pTruth []uint32, nFans int) *strings.Builder {
	panic("TODO, stubbed")
}

// Kit_DsdPrint2_rec
// C: void Kit_DsdPrint2_rec( FILE * pFile, Kit_DsdNtk_t * pNtk, int Id )
func KitDsdPrint2_rec(pFile *os.File, pNtk *KitDsdNtk, Id int) {
	panic("TODO, stubbed")
}

// Kit_DsdPrint2
// C: void Kit_DsdPrint2( FILE * pFile, Kit_DsdNtk_t * pNtk )
func KitDsdPrint2(pFile *os.File, pNtk *KitDsdNtk) {
	panic("TODO, stubbed")
}

// Kit_DsdPrint_rec
// C: void Kit_DsdPrint_rec( FILE * pFile, Kit_DsdNtk_t * pNtk, int Id )
func KitDsdPrint_rec(pFile *os.File, pNtk *KitDsdNtk, Id int) {
	panic("TODO, stubbed")
}

// Kit_DsdPrint
// C: void Kit_DsdPrint( FILE * pFile, Kit_DsdNtk_t * pNtk )
func KitDsdPrint(pFile *os.File, pNtk *KitDsdNtk) {
	panic("TODO, stubbed")
}

// Kit_DsdWrite_rec (pBuff is written through and the advanced pointer is returned; one builder serves both)
// C: char * Kit_DsdWrite_rec( char * pBuff, Kit_DsdNtk_t * pNtk, int Id )
func KitDsdWrite_rec(pBuff *strings.Builder, pNtk *KitDsdNtk, Id int) *strings.Builder {
	panic("TODO, stubbed")
}

// Kit_DsdWrite
// C: void Kit_DsdWrite( char * pBuff, Kit_DsdNtk_t * pNtk )
func KitDsdWrite(pBuff *strings.Builder, pNtk *KitDsdNtk) {
	panic("TODO, stubbed")
}

// Kit_DsdPrintExpanded
// C: void Kit_DsdPrintExpanded( Kit_DsdNtk_t * pNtk )
func (pNtk *KitDsdNtk) KitDsdPrintExpanded() {
	panic("TODO, stubbed")
}

// Kit_DsdPrintFromTruth
// C: void Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars )
func KitDsdPrintFromTruth(pTruth []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_DsdPrintFromTruth2
// C: void Kit_DsdPrintFromTruth2( FILE * pFile, unsigned * pTruth, int nVars )
func KitDsdPrintFromTruth2(pFile *os.File, pTruth []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_DsdWriteFromTruth
// C: void Kit_DsdWriteFromTruth( char * pBuffer, unsigned * pTruth, int nVars )
func KitDsdWriteFromTruth(pBuffer *strings.Builder, pTruth []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_DsdTruthComputeNode_rec (returns a pointer into the manager's truth-table memory)
// C: unsigned * Kit_DsdTruthComputeNode_rec( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, int Id )
func (p *KitDsdMan) KitDsdTruthComputeNode_rec(pNtk *KitDsdNtk, Id int) []uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdTruthCompute (returns a pointer into the manager's truth-table memory)
// C: unsigned * Kit_DsdTruthCompute( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk )
func (p *KitDsdMan) KitDsdTruthCompute(pNtk *KitDsdNtk) []uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdTruthComputeNodeOne_rec (returns a pointer into the manager's truth-table memory)
// C: unsigned * Kit_DsdTruthComputeNodeOne_rec( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, int Id, unsigned uSupp )
func (p *KitDsdMan) KitDsdTruthComputeNodeOne_rec(pNtk *KitDsdNtk, Id int, uSupp uint32) []uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdTruthComputeOne (returns a pointer into the manager's truth-table memory)
// C: unsigned * Kit_DsdTruthComputeOne( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, unsigned uSupp )
func (p *KitDsdMan) KitDsdTruthComputeOne(pNtk *KitDsdNtk, uSupp uint32) []uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdTruthComputeNodeTwo_rec (returns a pointer into the manager's truth-table memory)
// C: unsigned * Kit_DsdTruthComputeNodeTwo_rec( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, int Id, unsigned uSupp, int iVar, unsigned * pTruthDec )
func (p *KitDsdMan) KitDsdTruthComputeNodeTwo_rec(pNtk *KitDsdNtk, Id int, uSupp uint32, iVar int, pTruthDec []uint32) []uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdTruthComputeTwo (returns a pointer into the manager's truth-table memory)
// C: unsigned * Kit_DsdTruthComputeTwo( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, unsigned uSupp, int iVar, unsigned * pTruthDec )
func (p *KitDsdMan) KitDsdTruthComputeTwo(pNtk *KitDsdNtk, uSupp uint32, iVar int, pTruthDec []uint32) []uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdTruth
// C: void Kit_DsdTruth( Kit_DsdNtk_t * pNtk, unsigned * pTruthRes )
func (pNtk *KitDsdNtk) KitDsdTruth(pTruthRes []uint32) {
	panic("TODO, stubbed")
}

// Kit_DsdTruthPartialTwo
// C: void Kit_DsdTruthPartialTwo( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, unsigned uSupp, int iVar, unsigned * pTruthCo, unsigned * pTruthDec )
func (p *KitDsdMan) KitDsdTruthPartialTwo(pNtk *KitDsdNtk, uSupp uint32, iVar int, pTruthCo []uint32, pTruthDec []uint32) {
	panic("TODO, stubbed")
}

// Kit_DsdTruthPartial
// C: void Kit_DsdTruthPartial( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, unsigned * pTruthRes, unsigned uSupp )
func (p *KitDsdMan) KitDsdTruthPartial(pNtk *KitDsdNtk, pTruthRes []uint32, uSupp uint32) {
	panic("TODO, stubbed")
}

// Kit_DsdCountLuts_rec (pCounter is an in/out counter)
// C: int Kit_DsdCountLuts_rec( Kit_DsdNtk_t * pNtk, int nLutSize, int Id, int * pCounter )
func (pNtk *KitDsdNtk) KitDsdCountLuts_rec(nLutSize int, Id int, pCounter *int) int {
	panic("TODO, stubbed")
}

// Kit_DsdCountLuts
// C: int Kit_DsdCountLuts( Kit_DsdNtk_t * pNtk, int nLutSize )
func (pNtk *KitDsdNtk) KitDsdCountLuts(nLutSize int) int {
	panic("TODO, stubbed")
}

// Kit_DsdNonDsdSizeMax
// C: int Kit_DsdNonDsdSizeMax( Kit_DsdNtk_t * pNtk )
func (pNtk *KitDsdNtk) KitDsdNonDsdSizeMax() int {
	panic("TODO, stubbed")
}

// Kit_DsdNonDsdPrimeMax
// C: Kit_DsdObj_t * Kit_DsdNonDsdPrimeMax( Kit_DsdNtk_t * pNtk )
func (pNtk *KitDsdNtk) KitDsdNonDsdPrimeMax() *KitDsdObj {
	panic("TODO, stubbed")
}

// Kit_DsdNonDsdSupports
// C: unsigned Kit_DsdNonDsdSupports( Kit_DsdNtk_t * pNtk )
func (pNtk *KitDsdNtk) KitDsdNonDsdSupports() uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdExpandCollectAnd_rec (nLitsNew is an in/out counter)
// C: void Kit_DsdExpandCollectAnd_rec( Kit_DsdNtk_t * p, unsigned iLit, unsigned * piLitsNew, int * nLitsNew )
func (p *KitDsdNtk) KitDsdExpandCollectAnd_rec(iLit uint32, piLitsNew []uint32, nLitsNew *int) {
	panic("TODO, stubbed")
}

// Kit_DsdExpandCollectXor_rec (nLitsNew is an in/out counter)
// C: void Kit_DsdExpandCollectXor_rec( Kit_DsdNtk_t * p, unsigned iLit, unsigned * piLitsNew, int * nLitsNew )
func (p *KitDsdNtk) KitDsdExpandCollectXor_rec(iLit uint32, piLitsNew []uint32, nLitsNew *int) {
	panic("TODO, stubbed")
}

// Kit_DsdExpandNode_rec
// C: int Kit_DsdExpandNode_rec( Kit_DsdNtk_t * pNew, Kit_DsdNtk_t * p, int iLit )
func (pNew *KitDsdNtk) KitDsdExpandNode_rec(p *KitDsdNtk, iLit int) int {
	panic("TODO, stubbed")
}

// Kit_DsdExpand
// C: Kit_DsdNtk_t * Kit_DsdExpand( Kit_DsdNtk_t * p )
func (p *KitDsdNtk) KitDsdExpand() *KitDsdNtk {
	panic("TODO, stubbed")
}

// Kit_DsdCompSort
// C: void Kit_DsdCompSort( int pPrios[], unsigned uSupps[], unsigned short * piLits, int nVars, unsigned piLitsRes[] )
func KitDsdCompSort(pPrios []int, uSupps []uint32, piLits []uint16, nVars int, piLitsRes []uint32) {
	panic("TODO, stubbed")
}

// Kit_DsdShrink_rec
// C: int Kit_DsdShrink_rec( Kit_DsdNtk_t * pNew, Kit_DsdNtk_t * p, int iLit, int pPrios[] )
func (pNew *KitDsdNtk) KitDsdShrink_rec(p *KitDsdNtk, iLit int, pPrios []int) int {
	panic("TODO, stubbed")
}

// Kit_DsdShrink
// C: Kit_DsdNtk_t * Kit_DsdShrink( Kit_DsdNtk_t * p, int pPrios[] )
func (p *KitDsdNtk) KitDsdShrink(pPrios []int) *KitDsdNtk {
	panic("TODO, stubbed")
}

// Kit_DsdRotate
// C: void Kit_DsdRotate( Kit_DsdNtk_t * p, int pFreqs[] )
func (p *KitDsdNtk) KitDsdRotate(pFreqs []int) {
	panic("TODO, stubbed")
}

// Kit_DsdGetSupports_rec
// C: unsigned Kit_DsdGetSupports_rec( Kit_DsdNtk_t * p, int iLit )
func (p *KitDsdNtk) KitDsdGetSupports_rec(iLit int) uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdGetSupports
// C: unsigned Kit_DsdGetSupports( Kit_DsdNtk_t * p )
func (p *KitDsdNtk) KitDsdGetSupports() uint32 {
	panic("TODO, stubbed")
}

// Kit_DsdFindLargeBox_rec
// C: int Kit_DsdFindLargeBox_rec( Kit_DsdNtk_t * pNtk, int Id, int Size )
func (pNtk *KitDsdNtk) KitDsdFindLargeBox_rec(Id int, Size int) int {
	panic("TODO, stubbed")
}

// Kit_DsdFindLargeBox
// C: int Kit_DsdFindLargeBox( Kit_DsdNtk_t * pNtk, int Size )
func (pNtk *KitDsdNtk) KitDsdFindLargeBox(Size int) int {
	panic("TODO, stubbed")
}

// Kit_DsdCountAigNodes_rec
// C: int Kit_DsdCountAigNodes_rec( Kit_DsdNtk_t * pNtk, int Id )
func (pNtk *KitDsdNtk) KitDsdCountAigNodes_rec(Id int) int {
	panic("TODO, stubbed")
}

// Kit_DsdCountAigNodes2
// C: int Kit_DsdCountAigNodes2( Kit_DsdNtk_t * pNtk )
func (pNtk *KitDsdNtk) KitDsdCountAigNodes2() int {
	panic("TODO, stubbed")
}

// Kit_DsdCountAigNodes
// C: int Kit_DsdCountAigNodes( Kit_DsdNtk_t * pNtk )
func (pNtk *KitDsdNtk) KitDsdCountAigNodes() int {
	panic("TODO, stubbed")
}

// Kit_DsdRootNodeHasCommonVars
// C: int Kit_DsdRootNodeHasCommonVars( Kit_DsdObj_t * pObj0, Kit_DsdObj_t * pObj1 )
func KitDsdRootNodeHasCommonVars(pObj0 *KitDsdObj, pObj1 *KitDsdObj) int {
	panic("TODO, stubbed")
}

// Kit_DsdCheckVar4Dec2
// C: int Kit_DsdCheckVar4Dec2( Kit_DsdNtk_t * pNtk0, Kit_DsdNtk_t * pNtk1 )
func (pNtk0 *KitDsdNtk) KitDsdCheckVar4Dec2(pNtk1 *KitDsdNtk) int {
	panic("TODO, stubbed")
}

// Kit_DsdDecompose_rec (pPar points at the parent literal slot that is written)
// C: void Kit_DsdDecompose_rec( Kit_DsdNtk_t * pNtk, Kit_DsdObj_t * pObj, unsigned uSupp, unsigned short * pPar, int nDecMux )
func (pNtk *KitDsdNtk) KitDsdDecompose_rec(pObj *KitDsdObj, uSupp uint32, pPar *uint16, nDecMux int) {
	panic("TODO, stubbed")
}

// Kit_DsdDecomposeInt
// C: Kit_DsdNtk_t * Kit_DsdDecomposeInt( unsigned * pTruth, int nVars, int nDecMux )
func KitDsdDecomposeInt(pTruth []uint32, nVars int, nDecMux int) *KitDsdNtk {
	panic("TODO, stubbed")
}

// Kit_DsdDecompose
// C: Kit_DsdNtk_t * Kit_DsdDecompose( unsigned * pTruth, int nVars )
func KitDsdDecompose(pTruth []uint32, nVars int) *KitDsdNtk {
	panic("TODO, stubbed")
}

// Kit_DsdDecomposeExpand
// C: Kit_DsdNtk_t * Kit_DsdDecomposeExpand( unsigned * pTruth, int nVars )
func KitDsdDecomposeExpand(pTruth []uint32, nVars int) *KitDsdNtk {
	panic("TODO, stubbed")
}

// Kit_DsdDecomposeMux
// C: Kit_DsdNtk_t * Kit_DsdDecomposeMux( unsigned * pTruth, int nVars, int nDecMux )
func KitDsdDecomposeMux(pTruth []uint32, nVars int, nDecMux int) *KitDsdNtk {
	panic("TODO, stubbed")
}

// Kit_DsdTestCofs
// C: int Kit_DsdTestCofs( Kit_DsdNtk_t * pNtk, unsigned * pTruthInit )
func (pNtk *KitDsdNtk) KitDsdTestCofs(pTruthInit []uint32) int {
	panic("TODO, stubbed")
}

// Kit_DsdEval
// C: int Kit_DsdEval( unsigned * pTruth, int nVars, int nLutSize )
func KitDsdEval(pTruth []uint32, nVars int, nLutSize int) int {
	panic("TODO, stubbed")
}

// Kit_DsdVerify
// C: void Kit_DsdVerify( Kit_DsdNtk_t * pNtk, unsigned * pTruth, int nVars )
func (pNtk *KitDsdNtk) KitDsdVerify(pTruth []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_DsdTest
// C: void Kit_DsdTest( unsigned * pTruth, int nVars )
func KitDsdTest(pTruth []uint32, nVars int) {
	panic("TODO, stubbed")
}

// Kit_DsdPrecompute4Vars
// C: void Kit_DsdPrecompute4Vars()
func KitDsdPrecompute4Vars() {
	panic("TODO, stubbed")
}

// Kit_DsdCofactoringGetVars
// C: int Kit_DsdCofactoringGetVars( Kit_DsdNtk_t ** ppNtk, int nSize, int * pVars )
func KitDsdCofactoringGetVars(ppNtk []*KitDsdNtk, nSize int, pVars []int) int {
	panic("TODO, stubbed")
}

// Kit_DsdCofactoring
// C: int Kit_DsdCofactoring( unsigned * pTruth, int nVars, int * pCofVars, int nLimit, int fVerbose )
func KitDsdCofactoring(pTruth []uint32, nVars int, pCofVars []int, nLimit int, fVerbose int) int {
	panic("TODO, stubbed")
}

// Kit_DsdPrintCofactors
// C: void Kit_DsdPrintCofactors( unsigned * pTruth, int nVars, int nCofLevel, int fVerbose )
func KitDsdPrintCofactors(pTruth []uint32, nVars int, nCofLevel int, fVerbose int) {
	panic("TODO, stubbed")
}

// Kit_DsdNpn4ClassNames (returns the array of class-name strings)
// C: char ** Kit_DsdNpn4ClassNames()
func KitDsdNpn4ClassNames() []string {
	panic("TODO, stubbed")
}
