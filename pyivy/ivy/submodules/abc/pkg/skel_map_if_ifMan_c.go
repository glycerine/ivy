package abc

// Hand-checked skeleton of src/map/if/ifMan.c.

// If_ManCutSetRecycle
// C: void If_ManCutSetRecycle( If_Man_t * p, If_Set_t * pSet )
func (p *IfMan) ifManCutSetRecycle(pSet *IfSet) {
	panic("TODO, stubbed")
}

// If_ManCutSetFetch
// C: If_Set_t * If_ManCutSetFetch( If_Man_t * p )
func (p *IfMan) ifManCutSetFetch() *IfSet {
	panic("TODO, stubbed")
}

// If_ManStart
// C: If_Man_t * If_ManStart( If_Par_t * pPars )
func IfManStart(pPars *IfPar) *IfMan {
	panic("TODO, stubbed")
}

// If_ManRestart
// C: void If_ManRestart( If_Man_t * p )
func (p *IfMan) IfManRestart() {
	panic("TODO, stubbed")
}

// If_ManStop
// C: void If_ManStop( If_Man_t * p )
func (p *IfMan) IfManStop() {
	panic("TODO, stubbed")
}

// If_ManCreateCi
// C: If_Obj_t * If_ManCreateCi( If_Man_t * p )
func (p *IfMan) IfManCreateCi() *IfObj {
	panic("TODO, stubbed")
}

// If_ManCreateCo
// C: If_Obj_t * If_ManCreateCo( If_Man_t * p, If_Obj_t * pDriver )
func (p *IfMan) IfManCreateCo(pDriver *IfObj) *IfObj {
	panic("TODO, stubbed")
}

// If_ManCreateAnd
// C: If_Obj_t * If_ManCreateAnd( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1 )
func (p *IfMan) IfManCreateAnd(pFan0 *IfObj, pFan1 *IfObj) *IfObj {
	panic("TODO, stubbed")
}

// If_ManCreateXor
// C: If_Obj_t * If_ManCreateXor( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1 )
func (p *IfMan) IfManCreateXor(pFan0 *IfObj, pFan1 *IfObj) *IfObj {
	panic("TODO, stubbed")
}

// If_ManCreateMux
// C: If_Obj_t * If_ManCreateMux( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1, If_Obj_t * pCtrl )
func (p *IfMan) IfManCreateMux(pFan0 *IfObj, pFan1 *IfObj, pCtrl *IfObj) *IfObj {
	panic("TODO, stubbed")
}

// If_ManCreateChoice
// C: void If_ManCreateChoice( If_Man_t * p, If_Obj_t * pObj )
func (p *IfMan) IfManCreateChoice(pObj *IfObj) {
	panic("TODO, stubbed")
}

// If_ManSetupSet
// C: void If_ManSetupSet( If_Man_t * p, If_Set_t * pSet )
func (p *IfMan) IfManSetupSet(pSet *IfSet) {
	panic("TODO, stubbed")
}

// If_ManSetupCutTriv
// C: void If_ManSetupCutTriv( If_Man_t * p, If_Cut_t * pCut, int ObjId )
func (p *IfMan) IfManSetupCutTriv(pCut *IfCut, ObjId int) {
	panic("TODO, stubbed")
}

// If_ManSetupObj
// C: If_Obj_t * If_ManSetupObj( If_Man_t * p )
func (p *IfMan) ifManSetupObj() *IfObj {
	panic("TODO, stubbed")
}

// If_ManSetupCiCutSets
// C: void If_ManSetupCiCutSets( If_Man_t * p )
func (p *IfMan) IfManSetupCiCutSets() {
	panic("TODO, stubbed")
}

// If_ManSetupNodeCutSet
// C: If_Set_t * If_ManSetupNodeCutSet( If_Man_t * p, If_Obj_t * pObj )
func (p *IfMan) IfManSetupNodeCutSet(pObj *IfObj) *IfSet {
	panic("TODO, stubbed")
}

// If_ManDerefNodeCutSet
// C: void If_ManDerefNodeCutSet( If_Man_t * p, If_Obj_t * pObj )
func (p *IfMan) IfManDerefNodeCutSet(pObj *IfObj) {
	panic("TODO, stubbed")
}

// If_ManDerefChoiceCutSet
// C: void If_ManDerefChoiceCutSet( If_Man_t * p, If_Obj_t * pObj )
func (p *IfMan) IfManDerefChoiceCutSet(pObj *IfObj) {
	panic("TODO, stubbed")
}

// If_ManSetupSetAll
// C: void If_ManSetupSetAll( If_Man_t * p, int nCrossCut )
func (p *IfMan) IfManSetupSetAll(nCrossCut int) {
	panic("TODO, stubbed")
}
