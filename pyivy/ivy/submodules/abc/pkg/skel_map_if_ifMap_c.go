package abc

// Hand-checked skeleton of src/map/if/ifMap.c.

// The static local table Pin2Pin[2][3] in If_CutDelaySpecial is not represented.

// If_ManCutAigDelay_rec
// C: int If_ManCutAigDelay_rec( If_Man_t * p, If_Obj_t * pObj, Vec_Ptr_t * vVisited )
func (p *IfMan) IfManCutAigDelay_rec(pObj *IfObj, vVisited *PointerVector[IfObj]) int {
	panic("TODO, stubbed")
}

// If_ManCutAigDelay
// C: int If_ManCutAigDelay( If_Man_t * p, If_Obj_t * pObj, If_Cut_t * pCut )
func (p *IfMan) IfManCutAigDelay(pObj *IfObj, pCut *IfCut) int {
	panic("TODO, stubbed")
}

// If_WordCountOnes (duplicate static name in ifSelect.c)
// C: int If_WordCountOnes( unsigned uWord )
func ifWordCountOnes_ifMap(uWord uint32) int {
	panic("TODO, stubbed")
}

// If_CutDelaySpecial
// C: float If_CutDelaySpecial( If_Man_t * p, If_Cut_t * pCut, int fCarry )
func (p *IfMan) IfCutDelaySpecial(pCut *IfCut, fCarry int) float32 {
	panic("TODO, stubbed")
}

// If_CutArrTimeProfile (returns p->pArrTimeProfile)
// C: int * If_CutArrTimeProfile( If_Man_t * p, If_Cut_t * pCut )
func (p *IfMan) IfCutArrTimeProfile(pCut *IfCut) []int {
	panic("TODO, stubbed")
}

// If_ObjPerformMappingAnd
// C: void If_ObjPerformMappingAnd( If_Man_t * p, If_Obj_t * pObj, int Mode, int fPreprocess, int fFirst )
func (p *IfMan) IfObjPerformMappingAnd(pObj *IfObj, Mode int, fPreprocess int, fFirst int) {
	panic("TODO, stubbed")
}

// If_ObjPerformMappingChoice
// C: void If_ObjPerformMappingChoice( If_Man_t * p, If_Obj_t * pObj, int Mode, int fPreprocess )
func (p *IfMan) IfObjPerformMappingChoice(pObj *IfObj, Mode int, fPreprocess int) {
	panic("TODO, stubbed")
}

// If_ManPerformMappingRound
// C: int If_ManPerformMappingRound( If_Man_t * p, int nCutsUsed, int Mode, int fPreprocess, int fFirst, char * pLabel )
func (p *IfMan) IfManPerformMappingRound(nCutsUsed int, Mode int, fPreprocess int, fFirst int, pLabel string) int {
	panic("TODO, stubbed")
}
