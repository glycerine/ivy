package abc

// Hand-checked skeleton of src/bdd/llb/llb1Core.c.

// Llb_ManSetDefaultParams
// C: void Llb_ManSetDefaultParams( Gia_ParLlb_t * p )
func LlbManSetDefaultParams(p *GiaParLlb) {
	panic("TODO, stubbed")
}

// Llb_ManPrintAig
// C: void Llb_ManPrintAig( Llb_Man_t * p )
func (p *LlbMan) LlbManPrintAig() {
	panic("TODO, stubbed")
}

// Llb_ManModelCheckAig
// C: int Llb_ManModelCheckAig( Aig_Man_t * pAigGlo, Gia_ParLlb_t * pPars, Vec_Int_t * vHints, DdManager ** pddGlo )
func (pAigGlo *AigManager) LlbManModelCheckAig(pPars *GiaParLlb, vHints *IntVector, pddGlo **CuddFoundation) int {
	panic("TODO, stubbed")
}

// Llb_ManModelCheckGia
// C: int Llb_ManModelCheckGia( Gia_Man_t * pGia, Gia_ParLlb_t * pPars )
func (pGia *GiaManager) LlbManModelCheckGia(pPars *GiaParLlb) int {
	panic("TODO, stubbed")
}
