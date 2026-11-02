package abc

// Hand-checked skeleton of src/aig/gia/giaEsop.c.

// EsoMan is Eso_Man_t.
type EsoMan struct {
	pGia   *GiaManager        // user's AIG
	nVars  int                // number of variables
	Cube1  int                // ID of const1 cube
	vEsops *IntLevelVector    // ESOP for each node
	pHash  *IntVectorInterner // Hsh_VecMan_t * : hash table for cubes
	vCubes *IntLevelVector    // cover during minimization
	vCube1 *IntVector         // first cube
	vCube2 *IntVector         // second cube
	vCube  *IntVector         // resulting cube
}

// Eso_ManCube (static inline)
// C: Vec_Int_t * Eso_ManCube( Eso_Man_t * p, int iCube )
func (p *EsoMan) esoManCube(iCube int) *IntVector {
	panic("TODO, stubbed")
}

// Eso_ManAlloc
// C: Eso_Man_t * Eso_ManAlloc( Gia_Man_t * pGia )
func (pGia *GiaManager) EsoManAlloc() *EsoMan {
	panic("TODO, stubbed")
}

// Eso_ManStop
// C: void Eso_ManStop( Eso_Man_t * p )
func (p *EsoMan) EsoManStop() {
	panic("TODO, stubbed")
}

// Eso_ManCoverPrint
// C: void Eso_ManCoverPrint( Eso_Man_t * p, Vec_Int_t * vEsop )
func (p *EsoMan) EsoManCoverPrint(vEsop *IntVector) {
	panic("TODO, stubbed")
}

// Eso_ManCoverDerive
// C: Vec_Wec_t * Eso_ManCoverDerive( Eso_Man_t * p, Vec_Ptr_t * vCover )
func (p *EsoMan) EsoManCoverDerive(vCover *PointerVector[IntVector]) *IntLevelVector {
	panic("TODO, stubbed")
}

// Eso_ManCoverConvert
// C: Gia_Man_t * Eso_ManCoverConvert( Eso_Man_t * p, Vec_Ptr_t * vCover )
func (p *EsoMan) EsoManCoverConvert(vCover *PointerVector[IntVector]) *GiaManager {
	panic("TODO, stubbed")
}

// Eso_ManFindDistOneLitEqual
// C: int Eso_ManFindDistOneLitEqual( int * pCube1, int * pCube2, int nLits )
func EsoManFindDistOneLitEqual(pCube1 []int, pCube2 []int, nLits int) int {
	panic("TODO, stubbed")
}

// Eso_ManFindDistOneLitNotEqual
// C: int Eso_ManFindDistOneLitNotEqual( int * pCube1, int * pCube2, int nLits )
func EsoManFindDistOneLitNotEqual(pCube1 []int, pCube2 []int, nLits int) int {
	panic("TODO, stubbed")
}

// Eso_ManMinimizeAdd
// C: void Eso_ManMinimizeAdd( Eso_Man_t * p, int Cube )
func (p *EsoMan) EsoManMinimizeAdd(Cube int) {
	panic("TODO, stubbed")
}

// Eso_ManMinimizeCopy
// C: void Eso_ManMinimizeCopy( Eso_Man_t * p, Vec_Int_t * vEsop )
func (p *EsoMan) EsoManMinimizeCopy(vEsop *IntVector) {
	panic("TODO, stubbed")
}

// Eso_ManComputeAnd
// C: int Eso_ManComputeAnd( Eso_Man_t * p, Vec_Int_t * vCube1, Vec_Int_t * vCube2, Vec_Int_t * vCube )
func (p *EsoMan) EsoManComputeAnd(vCube1 *IntVector, vCube2 *IntVector, vCube *IntVector) int {
	panic("TODO, stubbed")
}

// Eso_ManComputeOne
// C: void Eso_ManComputeOne( Eso_Man_t * p, Vec_Int_t * vEsop1, Vec_Int_t * vEsop2, Vec_Int_t * vEsop )
func (p *EsoMan) EsoManComputeOne(vEsop1 *IntVector, vEsop2 *IntVector, vEsop *IntVector) {
	panic("TODO, stubbed")
}

// Eso_ManTransformOne
// C: Vec_Int_t * Eso_ManTransformOne( Eso_Man_t * p, Vec_Int_t * vEsop, int fCompl, Vec_Int_t * vRes )
func (p *EsoMan) EsoManTransformOne(vEsop *IntVector, fCompl int, vRes *IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Eso_ManCompute
// C: Gia_Man_t * Eso_ManCompute( Gia_Man_t * pGia, int fVerbose, Vec_Wec_t ** pvRes )
func (pGia *GiaManager) EsoManCompute(fVerbose int, pvRes **IntLevelVector) *GiaManager {
	panic("TODO, stubbed")
}
