package abc

// Hand-checked skeleton of src/map/if/ifDelay.c.

// If_CutMaxCubeSize
// C: int If_CutMaxCubeSize( Vec_Int_t * vCover, int nVars )
func ifCutMaxCubeSize(vCover *IntVector, nVars int) int {
	panic("TODO, stubbed")
}

// If_CutDelaySop
// C: int If_CutDelaySop( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) IfCutDelaySop(pCut *IfCut) int {
	panic("TODO, stubbed")
}

// If_CutSopBalancePinDelaysInt
// C: int If_CutSopBalancePinDelaysInt( Vec_Int_t * vCover, int * pTimes, word * pFaninRes, int nSuppAll, word * pRes )
func IfCutSopBalancePinDelaysInt(vCover *IntVector, pTimes []int, pFaninRes []uint64, nSuppAll int, pRes *uint64) int {
	panic("TODO, stubbed")
}

// If_CutSopBalancePinDelaysIntInt
// C: int If_CutSopBalancePinDelaysIntInt( Vec_Int_t * vCover, int * pTimes, int nSuppAll, char * pPerm )
func IfCutSopBalancePinDelaysIntInt(vCover *IntVector, pTimes []int, nSuppAll int, pPerm []int8) int {
	panic("TODO, stubbed")
}

// If_CutSopBalancePinDelays
// C: int If_CutSopBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm )
func (p *IfMan) IfCutSopBalancePinDelays(pCut *IfCut, pPerm []int8) int {
	panic("TODO, stubbed")
}

// If_CutSopBalanceEvalInt
// C: int If_CutSopBalanceEvalInt( Vec_Int_t * vCover, int * pTimes, int * pFaninLits, Vec_Int_t * vAig, int * piRes, int nSuppAll, int * pArea )
func IfCutSopBalanceEvalInt(vCover *IntVector, pTimes []int, pFaninLits []int, vAig *IntVector, piRes *int, nSuppAll int, pArea *int) int {
	panic("TODO, stubbed")
}

// If_CutSopBalanceEvalIntInt
// C: int If_CutSopBalanceEvalIntInt( Vec_Int_t * vCover, int nLeaves, int * pTimes, Vec_Int_t * vAig, int fCompl, int * pArea )
func IfCutSopBalanceEvalIntInt(vCover *IntVector, nLeaves int, pTimes []int, vAig *IntVector, fCompl int, pArea *int) int {
	panic("TODO, stubbed")
}

// If_CutSopBalanceEval
// C: int If_CutSopBalanceEval( If_Man_t * p, If_Cut_t * pCut, Vec_Int_t * vAig )
func (p *IfMan) IfCutSopBalanceEval(pCut *IfCut, vAig *IntVector) int {
	panic("TODO, stubbed")
}

// If_CutLutBalancePinDelays
// C: int If_CutLutBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm )
func (p *IfMan) IfCutLutBalancePinDelays(pCut *IfCut, pPerm []int8) int {
	panic("TODO, stubbed")
}

// If_CutLutBalanceEval
// C: int If_CutLutBalanceEval( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) IfCutLutBalanceEval(pCut *IfCut) int {
	panic("TODO, stubbed")
}
