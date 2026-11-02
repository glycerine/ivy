package abc

// Hand-checked skeleton of src/base/abci/abcIfif.c.

const IFIF_MAX_LEAVES = 6

// AbcIffObj is Abc_IffObj_t.
type AbcIffObj struct {
	Delay [IFIF_MAX_LEAVES + 1]float32 // separate delay
}

// AbcIffMan is Abc_IffMan_t.
type AbcIffMan struct {
	pNtk  *ABCNetwork
	pPars *IfifPar
	nObjs int
	pObjs []AbcIffObj // Abc_IffObj_t *
}

// Abc_IffObj (static)
// C: Abc_IffObj_t * Abc_IffObj( Abc_IffMan_t * p, int i )
func (p *AbcIffMan) abcIffObj(i int) *AbcIffObj {
	panic("TODO, stubbed")
}

// Abc_IffDelay (static)
// C: float Abc_IffDelay( Abc_IffMan_t * p, Abc_Obj_t * pObj, int fDelay1 )
func (p *AbcIffMan) abcIffDelay(pObj *ABCNetworkObject, fDelay1 int) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkIfifStart
// C: Abc_IffMan_t * Abc_NtkIfifStart( Abc_Ntk_t * pNtk, Ifif_Par_t * pPars )
func (pNtk *ABCNetwork) AbcNtkIfifStart(pPars *IfifPar) *AbcIffMan {
	panic("TODO, stubbed")
}

// Abc_NtkIfifStop
// C: void Abc_NtkIfifStop( Abc_IffMan_t * p )
func (p *AbcIffMan) AbcNtkIfifStop() {
	panic("TODO, stubbed")
}

// Abc_ObjSortByDelay
// C: void Abc_ObjSortByDelay( Abc_IffMan_t * p, Abc_Obj_t * pObj, int fDelay1, Abc_Obj_t ** ppNodes )
func (p *AbcIffMan) AbcObjSortByDelay(pObj *ABCNetworkObject, fDelay1 int, ppNodes []*ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Abc_ObjDelay0
// C: float Abc_ObjDelay0( Abc_IffMan_t * p, Abc_Obj_t * pObj )
func (p *AbcIffMan) AbcObjDelay0(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_ObjDelay1
// C: float Abc_ObjDelay1( Abc_IffMan_t * p, Abc_Obj_t * pObj )
func (p *AbcIffMan) AbcObjDelay1(pObj *ABCNetworkObject) float32 {
	panic("TODO, stubbed")
}

// Abc_ObjDelayDegree
// C: float Abc_ObjDelayDegree( Abc_IffMan_t * p, Abc_Obj_t * pObj, int d )
func (p *AbcIffMan) AbcObjDelayDegree(pObj *ABCNetworkObject, d int) float32 {
	panic("TODO, stubbed")
}

// Abc_NtkPerformIfif
// C: void Abc_NtkPerformIfif( Abc_Ntk_t * pNtk, Ifif_Par_t * pPars )
func (pNtk *ABCNetwork) AbcNtkPerformIfif(pPars *IfifPar) {
	panic("TODO, stubbed")
}
