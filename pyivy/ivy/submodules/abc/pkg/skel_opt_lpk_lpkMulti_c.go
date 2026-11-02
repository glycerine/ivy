package abc

// Hand-checked skeleton of src/opt/lpk/lpkMulti.c.

// Lpk_CreateVarOrder
// C: void Lpk_CreateVarOrder( Kit_DsdNtk_t * pNtk, char pTable[][16] )
func LpkCreateVarOrder(pNtk *KitDsdNtk, pTable [][16]int8) {
	panic("TODO, stubbed")
}

// Lpk_CreateCommonOrder
// C: void Lpk_CreateCommonOrder( char pTable[][16], int piCofVar[], int nCBars, int pPrios[], int nVars, int fVerbose )
func LpkCreateCommonOrder(pTable [][16]int8, piCofVar []int, nCBars int, pPrios []int, nVars int, fVerbose int) {
	panic("TODO, stubbed")
}

// Lpk_FindHighest (pDecision is an out parameter)
// C: int Lpk_FindHighest( Kit_DsdNtk_t ** ppNtks, int * piLits, int nSize, int * pPrio, int * pDecision )
func LpkFindHighest(ppNtks []*KitDsdNtk, piLits []int, nSize int, pPrio []int, pDecision *int) int {
	panic("TODO, stubbed")
}

// Lpk_MapTreeMulti_rec
// C: If_Obj_t * Lpk_MapTreeMulti_rec( Lpk_Man_t * p, Kit_DsdNtk_t ** ppNtks, int * piLits, int * piCofVar, int nCBars, If_Obj_t ** ppLeaves, int nLeaves, int * pPrio )
func (p *LpkMan) LpkMapTreeMulti_rec(ppNtks []*KitDsdNtk, piLits []int, piCofVar []int, nCBars int, ppLeaves []*IfObj, nLeaves int, pPrio []int) *IfObj {
	panic("TODO, stubbed")
}

// Lpk_MapTreeMulti
// C: If_Obj_t * Lpk_MapTreeMulti( Lpk_Man_t * p, unsigned * pTruth, int nVars, If_Obj_t ** ppLeaves )
func (p *LpkMan) LpkMapTreeMulti(pTruth []uint32, nVars int, ppLeaves []*IfObj) *IfObj {
	panic("TODO, stubbed")
}
