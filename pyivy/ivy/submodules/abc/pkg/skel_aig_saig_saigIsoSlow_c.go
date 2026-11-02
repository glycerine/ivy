package abc

// Hand-checked skeleton of src/aig/saig/saigIsoSlow.c.

const ISO_NUM_INTS = 3

// IsoObj is Iso_Obj_t.
type IsoObj struct {
	// hashing entries (related to the parameter ISO_NUM_INTS!)
	Level     uint32 // unsigned : 30
	nFinNeg   uint32 // unsigned : 2
	FaninSig  uint32
	FanoutSig uint32
	// other data
	iNext  int // hash table entry
	iClass int // next one in class
	Id     int // unique ID
}

// IsoMan is Iso_Man_t.
type IsoMan struct {
	pAig      *AigManager            // user's AIG manager
	pObjs     []IsoObj               // Iso_Obj_t * : isomorphism objects
	nObjIds   int                    // counter of object IDs
	nClasses  int                    // total number of classes
	nEntries  int                    // total number of entries
	nSingles  int                    // total number of singletons
	nObjs     int                    // total objects
	nBins     int                    // hash table size
	pBins     []int                  // int * : hash table
	vSingles  *PointerVector[IsoObj] // singletons
	vClasses  *PointerVector[IsoObj] // other classes
	vTemp1    *PointerVector[any]    // other classes (holds Iso_Obj_t * and Aig_Obj_t * entries)
	vTemp2    *PointerVector[any]    // other classes (holds Iso_Obj_t * and Aig_Obj_t * entries)
	timeHash  int64                  // abctime
	timeFout  int64                  // abctime
	timeSort  int64                  // abctime
	timeOther int64                  // abctime
	timeTotal int64                  // abctime
}

// The static data tables of the prime numbers and the redefined ISO_MASK macros are not represented.

// Iso_ManObj (static inline)
// C: Iso_Obj_t * Iso_ManObj( Iso_Man_t * p, int i )
func (p *IsoMan) isoManObj(i int) *IsoObj {
	panic("TODO, stubbed")
}

// Iso_ObjId (static inline)
// C: int Iso_ObjId( Iso_Man_t * p, Iso_Obj_t * pObj )
func (p *IsoMan) isoObjId(pObj *IsoObj) int {
	panic("TODO, stubbed")
}

// Iso_AigObj (static inline)
// C: Aig_Obj_t * Iso_AigObj( Iso_Man_t * p, Iso_Obj_t * q )
func (p *IsoMan) isoAigObj(q *IsoObj) *AigObject {
	panic("TODO, stubbed")
}

// Iso_ReadPrimes
// C: void Iso_ReadPrimes( char * pFileName )
func IsoReadPrimes(pFileName string) {
	panic("TODO, stubbed")
}

// Iso_FindNumbers
// C: void Iso_FindNumbers()
func IsoFindNumbers() {
	panic("TODO, stubbed")
}

// Iso_ManObjCount_rec
// C: void Iso_ManObjCount_rec( Aig_Man_t * p, Aig_Obj_t * pObj, int * pnNodes, int * pnEdges )
func (p *AigManager) IsoManObjCount_rec(pObj *AigObject, pnNodes *int, pnEdges *int) {
	panic("TODO, stubbed")
}

// Iso_ManObjCount
// C: void Iso_ManObjCount( Aig_Man_t * p, Aig_Obj_t * pObj, int * pnNodes, int * pnEdges )
func (p *AigManager) IsoManObjCount(pObj *AigObject, pnNodes *int, pnEdges *int) {
	panic("TODO, stubbed")
}

// Iso_ManStart
// C: Iso_Man_t * Iso_ManStart( Aig_Man_t * pAig )
func (pAig *AigManager) IsoManStart() *IsoMan {
	panic("TODO, stubbed")
}

// Iso_ManStop
// C: void Iso_ManStop( Iso_Man_t * p, int fVerbose )
func (p *IsoMan) IsoManStop(fVerbose int) {
	panic("TODO, stubbed")
}

// Iso_ObjCompare
// C: int Iso_ObjCompare( Iso_Obj_t ** pp1, Iso_Obj_t ** pp2 )
func IsoObjCompare(pp1 **IsoObj, pp2 **IsoObj) int {
	panic("TODO, stubbed")
}

// Iso_ObjCompareByData
// C: int Iso_ObjCompareByData( Aig_Obj_t ** pp1, Aig_Obj_t ** pp2 )
func IsoObjCompareByData(pp1 **AigObject, pp2 **AigObject) int {
	panic("TODO, stubbed")
}

// Iso_ObjHash (static inline)
// C: int Iso_ObjHash( Iso_Obj_t * pIso, int nBins )
func isoObjHash(pIso *IsoObj, nBins int) int {
	panic("TODO, stubbed")
}

// Iso_ObjHashAdd (static inline)
// C: int Iso_ObjHashAdd( Iso_Man_t * p, Iso_Obj_t * pIso )
func (p *IsoMan) isoObjHashAdd(pIso *IsoObj) int {
	panic("TODO, stubbed")
}

// Iso_ManCollectClasses
// C: void Iso_ManCollectClasses( Iso_Man_t * p )
func (p *IsoMan) IsoManCollectClasses() {
	panic("TODO, stubbed")
}

// Iso_ManCreate
// C: Iso_Man_t * Iso_ManCreate( Aig_Man_t * pAig )
func (pAig *AigManager) IsoManCreate() *IsoMan {
	panic("TODO, stubbed")
}

// Iso_ManAssignAdjacency
// C: void Iso_ManAssignAdjacency( Iso_Man_t * p )
func (p *IsoMan) IsoManAssignAdjacency() {
	panic("TODO, stubbed")
}

// Iso_ManPrintClasseSizes
// C: void Iso_ManPrintClasseSizes( Iso_Man_t * p )
func (p *IsoMan) IsoManPrintClasseSizes() {
	panic("TODO, stubbed")
}

// Iso_ManPrintClasses
// C: void Iso_ManPrintClasses( Iso_Man_t * p, int fVerbose, int fVeryVerbose )
func (p *IsoMan) IsoManPrintClasses(fVerbose int, fVeryVerbose int) {
	panic("TODO, stubbed")
}

// Iso_ManRehashClassNodes
// C: void Iso_ManRehashClassNodes( Iso_Man_t * p )
func (p *IsoMan) IsoManRehashClassNodes() {
	panic("TODO, stubbed")
}

// Iso_ManFindBestObj
// C: Iso_Obj_t * Iso_ManFindBestObj( Iso_Man_t * p, Iso_Obj_t * pIso )
func (p *IsoMan) IsoManFindBestObj(pIso *IsoObj) *IsoObj {
	panic("TODO, stubbed")
}

// Iso_ManBreakTies
// C: void Iso_ManBreakTies( Iso_Man_t * p, int fVerbose )
func (p *IsoMan) IsoManBreakTies(fVerbose int) {
	panic("TODO, stubbed")
}

// Iso_ManFinalize
// C: Vec_Int_t * Iso_ManFinalize( Iso_Man_t * p )
func (p *IsoMan) IsoManFinalize() *IntVector {
	panic("TODO, stubbed")
}

// Iso_ManDumpOneClass
// C: void Iso_ManDumpOneClass( Iso_Man_t * p )
func (p *IsoMan) IsoManDumpOneClass() {
	panic("TODO, stubbed")
}

// Saig_ManFindIsoPerm
// C: Vec_Int_t * Saig_ManFindIsoPerm( Aig_Man_t * pAig, int fVerbose )
func (pAig *AigManager) SaigManFindIsoPerm(fVerbose int) *IntVector {
	panic("TODO, stubbed")
}
