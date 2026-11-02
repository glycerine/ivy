package abc

// Hand-checked skeleton of src/aig/gia/giaIso2.c.

// GiaIso2Man is Gia_Iso2Man_t.
type GiaIso2Man struct {
	pGia       *GiaManager
	nObjs      int
	nUniques   int
	vUniques   *IntVector                // unique numbers
	vTied      *IntVector                // tied objects
	vTable     *IntVector                // hash table
	vPlaces    *IntVector                // used places in the table
	vSingles   *PointerVector[GiaObject] // singleton objects
	vVec0      *IntVector                // isomorphism map
	vVec1      *IntVector                // isomorphism map
	vMap0      *IntVector                // isomorphism map
	vMap1      *IntVector                // isomorphism map
	nIters     int
	timeStart  int64 // abctime
	timeSim    int64 // abctime
	timeRefine int64 // abctime
	timeSort   int64 // abctime
	timeOther  int64 // abctime
	timeTotal  int64 // abctime
}

// Gia_Iso2ManCollectTies
// C: Vec_Int_t * Gia_Iso2ManCollectTies( Gia_Man_t * p )
func (p *GiaManager) GiaIso2ManCollectTies() *IntVector {
	panic("TODO, stubbed")
}

// Gia_Iso2ManPrepare
// C: void Gia_Iso2ManPrepare( Gia_Man_t * p )
func (p *GiaManager) GiaIso2ManPrepare() {
	panic("TODO, stubbed")
}

// Gia_Iso2ManPropagate
// C: void Gia_Iso2ManPropagate( Gia_Man_t * p )
func (p *GiaManager) GiaIso2ManPropagate() {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCone_rec
// C: unsigned Gia_Iso2ManCone_rec( Gia_Man_t * p, int Id, int Level )
func (p *GiaManager) GiaIso2ManCone_rec(Id int, Level int) uint32 {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCone
// C: unsigned Gia_Iso2ManCone( Gia_Man_t * p, int Id, int Level )
func (p *GiaManager) GiaIso2ManCone(Id int, Level int) uint32 {
	panic("TODO, stubbed")
}

// Gia_Iso2ManUpdate
// C: void Gia_Iso2ManUpdate( Gia_Iso2Man_t * p, int Level )
func (p *GiaIso2Man) GiaIso2ManUpdate(Level int) {
	panic("TODO, stubbed")
}

// Gia_Iso2ManStart
// C: Gia_Iso2Man_t * Gia_Iso2ManStart( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaIso2ManStart() *GiaIso2Man {
	panic("TODO, stubbed")
}

// Gia_Iso2ManStop
// C: void Gia_Iso2ManStop( Gia_Iso2Man_t * p )
func (p *GiaIso2Man) GiaIso2ManStop() {
	panic("TODO, stubbed")
}

// Gia_Iso2ManPrint
// C: void Gia_Iso2ManPrint( Gia_Iso2Man_t * p, abctime Time, int fVerbose )
func (p *GiaIso2Man) GiaIso2ManPrint(Time int64, fVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ObjCompareByValue2
// C: int Gia_ObjCompareByValue2( Gia_Obj_t ** pp1, Gia_Obj_t ** pp2 )
func GiaObjCompareByValue2(pp1 **GiaObject, pp2 **GiaObject) int {
	panic("TODO, stubbed")
}

// Gia_Iso2ManUniqify
// C: int Gia_Iso2ManUniqify( Gia_Iso2Man_t * p )
func (p *GiaIso2Man) GiaIso2ManUniqify() int {
	panic("TODO, stubbed")
}

// Gia_Iso2ManDerivePoClasses
// C: Vec_Wec_t * Gia_Iso2ManDerivePoClasses( Gia_Man_t * pGia )
func (pGia *GiaManager) GiaIso2ManDerivePoClasses() *IntLevelVector {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCollectOrder2_rec
// C: void Gia_Iso2ManCollectOrder2_rec( Gia_Man_t * p, int Id, Vec_Int_t * vVec )
func (p *GiaManager) GiaIso2ManCollectOrder2_rec(Id int, vVec *IntVector) {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCollectOrder2
// C: Vec_Int_t * Gia_Iso2ManCollectOrder2( Gia_Man_t * pGia, int * pPos, int nPos )
func (pGia *GiaManager) GiaIso2ManCollectOrder2(pPos []int, nPos int) *IntVector {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCollectOrder_rec
// C: void Gia_Iso2ManCollectOrder_rec( Gia_Man_t * p, int Id, Vec_Int_t * vRoots, Vec_Int_t * vVec, Vec_Int_t * vMap )
func (p *GiaManager) GiaIso2ManCollectOrder_rec(Id int, vRoots *IntVector, vVec *IntVector, vMap *IntVector) {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCollectOrder
// C: void Gia_Iso2ManCollectOrder( Gia_Man_t * pGia, int * pPos, int nPos, Vec_Int_t * vRoots, Vec_Int_t * vVec, Vec_Int_t * vMap )
func (pGia *GiaManager) GiaIso2ManCollectOrder(pPos []int, nPos int, vRoots *IntVector, vVec *IntVector, vMap *IntVector) {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCheckIsoPair
// C: int Gia_Iso2ManCheckIsoPair( Gia_Man_t * p, Vec_Int_t * vVec0, Vec_Int_t * vVec1, Vec_Int_t * vMap0, Vec_Int_t * vMap1 )
func (p *GiaManager) GiaIso2ManCheckIsoPair(vVec0 *IntVector, vVec1 *IntVector, vMap0 *IntVector, vMap1 *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCheckIsoClassOneSkip
// C: int Gia_Iso2ManCheckIsoClassOneSkip( Gia_Man_t * p, Vec_Int_t * vClass, Vec_Int_t * vRoots, Vec_Int_t * vVec0, Vec_Int_t * vVec1, Vec_Int_t * vMap0, Vec_Int_t * vMap1 )
func (p *GiaManager) GiaIso2ManCheckIsoClassOneSkip(vClass *IntVector, vRoots *IntVector, vVec0 *IntVector, vVec1 *IntVector, vMap0 *IntVector, vMap1 *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCheckIsoClassesSkip
// C: Vec_Wec_t * Gia_Iso2ManCheckIsoClassesSkip( Gia_Man_t * p, Vec_Wec_t * vEquivs )
func (p *GiaManager) GiaIso2ManCheckIsoClassesSkip(vEquivs *IntLevelVector) *IntLevelVector {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCheckIsoClassOne
// C: void Gia_Iso2ManCheckIsoClassOne( Gia_Man_t * p, Vec_Int_t * vClass, Vec_Int_t * vRoots, Vec_Int_t * vVec0, Vec_Int_t * vVec1, Vec_Int_t * vMap0, Vec_Int_t * vMap1, Vec_Int_t * vNewClass )
func (p *GiaManager) GiaIso2ManCheckIsoClassOne(vClass *IntVector, vRoots *IntVector, vVec0 *IntVector, vVec1 *IntVector, vMap0 *IntVector, vMap1 *IntVector, vNewClass *IntVector) {
	panic("TODO, stubbed")
}

// Gia_Iso2ManCheckIsoClasses
// C: Vec_Wec_t * Gia_Iso2ManCheckIsoClasses( Gia_Man_t * p, Vec_Wec_t * vEquivs )
func (p *GiaManager) GiaIso2ManCheckIsoClasses(vEquivs *IntLevelVector) *IntLevelVector {
	panic("TODO, stubbed")
}

// Gia_Iso2ManPerform
// C: Vec_Wec_t * Gia_Iso2ManPerform( Gia_Man_t * pGia, int fVerbose )
func (pGia *GiaManager) GiaIso2ManPerform(fVerbose int) *IntLevelVector {
	panic("TODO, stubbed")
}

// Gia_ManIsoReduce2
// C: Gia_Man_t * Gia_ManIsoReduce2( Gia_Man_t * pGia, Vec_Ptr_t ** pvPosEquivs, Vec_Ptr_t ** pvPiPerms, int fEstimate, int fBetterQual, int fDualOut, int fVerbose, int fVeryVerbose )
func (pGia *GiaManager) GiaManIsoReduce2(pvPosEquivs **PointerVector[IntVector], pvPiPerms **PointerVector[IntVector], fEstimate int, fBetterQual int, fDualOut int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
