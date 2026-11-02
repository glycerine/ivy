package abc

// Hand-checked skeleton of src/aig/gia/giaFx.c.

// Gia_ManGraphToAig
// C: int Gia_ManGraphToAig( Gia_Man_t * p, Dec_Graph_t * pGraph )
func (p *GiaManager) GiaManGraphToAig(pGraph *DecGraph) int {
	panic("TODO, stubbed")
}

// Gia_ManSopToAig
// C: int Gia_ManSopToAig( Gia_Man_t * p, char * pSop, Vec_Int_t * vLeaves )
func (p *GiaManager) GiaManSopToAig(pSop string, vLeaves *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManFactorGraph
// C: int Gia_ManFactorGraph( Gia_Man_t * p, Dec_Graph_t * pFForm, Vec_Int_t * vLeaves )
func (p *GiaManager) GiaManFactorGraph(pFForm *DecGraph, vLeaves *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManFactorNode
// C: int Gia_ManFactorNode( Gia_Man_t * p, char * pSop, Vec_Int_t * vLeaves )
func (p *GiaManager) GiaManFactorNode(pSop string, vLeaves *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManComputeTruths
// C: Vec_Wrd_t * Gia_ManComputeTruths( Gia_Man_t * p, int nCutSize, int nLutNum, int fReverse )
func (p *GiaManager) GiaManComputeTruths(nCutSize int, nLutNum int, fReverse int) *WordVector {
	panic("TODO, stubbed")
}

// Gia_ManAssignNumbers
// C: int Gia_ManAssignNumbers( Gia_Man_t * p )
func (p *GiaManager) GiaManAssignNumbers() int {
	panic("TODO, stubbed")
}

// Gia_ManFxRetrieve
// C: Vec_Wec_t * Gia_ManFxRetrieve( Gia_Man_t * p, Vec_Str_t ** pvCompl, int fReverse )
func (p *GiaManager) GiaManFxRetrieve(pvCompl **ByteVector, fReverse int) *IntLevelVector {
	panic("TODO, stubbed")
}

// Gia_ManFxTopoOrder_rec
// C: void Gia_ManFxTopoOrder_rec( Vec_Wec_t * vCubes, Vec_Int_t * vFirst, Vec_Int_t * vCount, Vec_Int_t * vVisit, Vec_Int_t * vOrder, int iObj )
func GiaManFxTopoOrder_rec(vCubes *IntLevelVector, vFirst *IntVector, vCount *IntVector, vVisit *IntVector, vOrder *IntVector, iObj int) {
	panic("TODO, stubbed")
}

// Gia_ManFxTopoOrder
// C: Vec_Int_t * Gia_ManFxTopoOrder( Vec_Wec_t * vCubes, int nInputs, int nStart, Vec_Int_t ** pvFirst, Vec_Int_t ** pvCount )
func GiaManFxTopoOrder(vCubes *IntLevelVector, nInputs int, nStart int, pvFirst **IntVector, pvCount **IntVector) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManFxInsert
// C: Gia_Man_t * Gia_ManFxInsert( Gia_Man_t * p, Vec_Wec_t * vCubes, Vec_Str_t * vCompls )
func (p *GiaManager) GiaManFxInsert(vCubes *IntLevelVector, vCompls *ByteVector) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManPerformFx
// C: Gia_Man_t * Gia_ManPerformFx( Gia_Man_t * p, int nNewNodesMax, int LitCountMax, int fReverse, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManPerformFx(nNewNodesMax int, LitCountMax int, fReverse int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
