package abc

// Hand-checked skeleton of src/opt/lpk/lpkMap.c.

// Lpk_MapPrimeInternal
// C: If_Obj_t * Lpk_MapPrimeInternal( If_Man_t * pIfMan, Kit_Graph_t * pGraph )
func (pIfMan *IfMan) LpkMapPrimeInternal(pGraph *KitGraph) *IfObj {
	panic("TODO, stubbed")
}

// Lpk_MapPrime
// C: If_Obj_t * Lpk_MapPrime( Lpk_Man_t * p, unsigned * pTruth, int nVars, If_Obj_t ** ppLeaves )
func (p *LpkMan) LpkMapPrime(pTruth []uint32, nVars int, ppLeaves []*IfObj) *IfObj {
	panic("TODO, stubbed")
}

// Lpk_MapTree_rec
// C: If_Obj_t * Lpk_MapTree_rec( Lpk_Man_t * p, Kit_DsdNtk_t * pNtk, If_Obj_t ** ppLeaves, int iLit, If_Obj_t * pResult )
func (p *LpkMan) LpkMapTree_rec(pNtk *KitDsdNtk, ppLeaves []*IfObj, iLit int, pResult *IfObj) *IfObj {
	panic("TODO, stubbed")
}
