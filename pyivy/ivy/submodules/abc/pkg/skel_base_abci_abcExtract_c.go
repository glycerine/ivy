package abc

// Hand-checked skeleton of src/base/abci/abcExtract.c.

const SHARE_NUM = 2

// AbcShaMan is Abc_ShaMan_t.
type AbcShaMan struct {
	nMultiSize  int
	fVerbose    int
	pNtk        *ABCNetwork
	vBuckets    *PointerVector[IntVector]
	vObj2Lit    *IntVector
	nStartCols  int
	nCountGates int
	nFoundGates int
}

// Abc_NtkSharePack (static)
// C: word Abc_NtkSharePack( int Lev, int Id )
func abcNtkSharePack(Lev int, Id int) uint64 {
	panic("TODO, stubbed")
}

// Abc_NtkShareUnpackLev (static)
// C: int Abc_NtkShareUnpackLev( word Num )
func abcNtkShareUnpackLev(Num uint64) int {
	panic("TODO, stubbed")
}

// Abc_NtkShareUnpackId (static)
// C: int Abc_NtkShareUnpackId( word Num )
func abcNtkShareUnpackId(Num uint64) int {
	panic("TODO, stubbed")
}

// Abc_ShaManStart
// C: Abc_ShaMan_t * Abc_ShaManStart( Abc_Ntk_t * pNtk )
func (pNtk *ABCNetwork) AbcShaManStart() *AbcShaMan {
	panic("TODO, stubbed")
}

// Abc_ShaManStop
// C: void Abc_ShaManStop( Abc_ShaMan_t * p )
func (p *AbcShaMan) AbcShaManStop() {
	panic("TODO, stubbed")
}

// Abc_NtkShareSuperXor
// C: Vec_Wrd_t * Abc_NtkShareSuperXor( Abc_Obj_t * pObj, int * pfCompl, int * pCounter )
func AbcNtkShareSuperXor(pObj *ABCNetworkObject, pfCompl *int, pCounter *int) *WordVector {
	panic("TODO, stubbed")
}

// Abc_NtkShareSuperAnd
// C: Vec_Wrd_t * Abc_NtkShareSuperAnd( Abc_Obj_t * pObj, int * pCounter )
func AbcNtkShareSuperAnd(pObj *ABCNetworkObject, pCounter *int) *WordVector {
	panic("TODO, stubbed")
}

// Abc_NtkTraverseSupersXor_rec
// C: void Abc_NtkTraverseSupersXor_rec( Abc_ShaMan_t * p, Abc_Obj_t * pObj, Vec_Ptr_t * vInputs )
func (p *AbcShaMan) AbcNtkTraverseSupersXor_rec(pObj *ABCNetworkObject, vInputs *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkTraverseSupersAnd_rec
// C: void Abc_NtkTraverseSupersAnd_rec( Abc_ShaMan_t * p, Abc_Obj_t * pObj, Vec_Ptr_t * vInputs )
func (p *AbcShaMan) AbcNtkTraverseSupersAnd_rec(pObj *ABCNetworkObject, vInputs *PointerVector[ABCNetworkObject]) {
	panic("TODO, stubbed")
}

// Abc_NtkTraverseSupers
// C: void Abc_NtkTraverseSupers( Abc_ShaMan_t * p, int fAnd )
func (p *AbcShaMan) AbcNtkTraverseSupers(fAnd int) {
	panic("TODO, stubbed")
}

// Abc_NtkSharePrint
// C: void Abc_NtkSharePrint( Abc_ShaMan_t * p )
func (p *AbcShaMan) AbcNtkSharePrint() {
	panic("TODO, stubbed")
}

// Abc_NtkDumpBlif
// C: void Abc_NtkDumpBlif( Abc_Ntk_t * p )
func (p *ABCNetwork) AbcNtkDumpBlif() {
	panic("TODO, stubbed")
}

// Abc_NtkShareFindBestMatch
// C: void Abc_NtkShareFindBestMatch( Vec_Ptr_t * vBuckets, Vec_Int_t ** pvInput, Vec_Int_t ** pvInput2 )
func AbcNtkShareFindBestMatch(vBuckets *PointerVector[IntVector], pvInput **IntVector, pvInput2 **IntVector) {
	panic("TODO, stubbed")
}

// Abc_NtkShareOptimize
// C: void Abc_NtkShareOptimize( Abc_ShaMan_t * p, int fAnd )
func (p *AbcShaMan) AbcNtkShareOptimize(fAnd int) {
	panic("TODO, stubbed")
}

// Abc_NtkUpdateNetwork
// C: Abc_Ntk_t * Abc_NtkUpdateNetwork( Abc_ShaMan_t * p, int fAnd )
func (p *AbcShaMan) AbcNtkUpdateNetwork(fAnd int) *ABCNetwork {
	panic("TODO, stubbed")
}

// Abc_NtkShareXor
// C: Abc_Ntk_t * Abc_NtkShareXor( Abc_Ntk_t * pNtk, int nMultiSize, int fAnd, int fVerbose )
func (pNtk *ABCNetwork) AbcNtkShareXor(nMultiSize int, fAnd int, fVerbose int) *ABCNetwork {
	panic("TODO, stubbed")
}
