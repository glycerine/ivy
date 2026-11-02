package abc

// Hand-checked skeleton of src/map/if/ifSelect.c.

// If_ObjConePrint_rec
// C: void If_ObjConePrint_rec( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Ptr_t * vVisited )
func (pIfMan *IfMan) IfObjConePrint_rec(pIfObj *IfObj, vVisited *PointerVector[IfObj]) {
	panic("TODO, stubbed")
}

// If_ObjConePrint
// C: void If_ObjConePrint( If_Man_t * pIfMan, If_Obj_t * pIfObj )
func (pIfMan *IfMan) IfObjConePrint(pIfObj *IfObj) {
	panic("TODO, stubbed")
}

// If_ManNodeShapeMap_rec
// C: int If_ManNodeShapeMap_rec( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Ptr_t * vVisited, Vec_Int_t * vShape )
func (pIfMan *IfMan) IfManNodeShapeMap_rec(pIfObj *IfObj, vVisited *PointerVector[IfObj], vShape *IntVector) int {
	panic("TODO, stubbed")
}

// If_ManNodeShapeMap
// C: int If_ManNodeShapeMap( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Int_t * vShape )
func (pIfMan *IfMan) IfManNodeShapeMap(pIfObj *IfObj, vShape *IntVector) int {
	panic("TODO, stubbed")
}

// If_WordCountOnes (duplicate static name in ifMap.c)
// C: int If_WordCountOnes( unsigned uWord )
func ifWordCountOnes_ifSelect(uWord uint32) int {
	panic("TODO, stubbed")
}

// If_ManNodeShapeMap2_rec
// C: int If_ManNodeShapeMap2_rec( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Ptr_t * vVisited, Vec_Int_t * vShape )
func (pIfMan *IfMan) IfManNodeShapeMap2_rec(pIfObj *IfObj, vVisited *PointerVector[IfObj], vShape *IntVector) int {
	panic("TODO, stubbed")
}

// If_ManNodeShapeMap2
// C: int If_ManNodeShapeMap2( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Int_t * vShape )
func (pIfMan *IfMan) IfManNodeShapeMap2(pIfObj *IfObj, vShape *IntVector) int {
	panic("TODO, stubbed")
}

// If_ManConeCollect_rec
// C: int If_ManConeCollect_rec( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Ptr_t * vVisited, Vec_Ptr_t * vCone )
func (pIfMan *IfMan) IfManConeCollect_rec(pIfObj *IfObj, vVisited *PointerVector[IfObj], vCone *PointerVector[IfObj]) int {
	panic("TODO, stubbed")
}

// If_ManConeCollect
// C: Vec_Ptr_t * If_ManConeCollect( If_Man_t * pIfMan, If_Obj_t * pIfObj, If_Cut_t * pCut )
func (pIfMan *IfMan) IfManConeCollect(pIfObj *IfObj, pCut *IfCut) *PointerVector[IfObj] {
	panic("TODO, stubbed")
}

// sat_solver_add_choice
// C: void sat_solver_add_choice( sat_solver * pSat, int iVar, Vec_Int_t * vVars )
func (pSat *SatSolver) sat_solver_add_choice(iVar int, vVars *IntVector) {
	panic("TODO, stubbed")
}

// If_ObjSatVar
// C: int If_ObjSatVar( If_Obj_t * pIfObj )
func ifObjSatVar(pIfObj *IfObj) int {
	panic("TODO, stubbed")
}

// If_ObjSetSatVar
// C: void If_ObjSetSatVar( If_Obj_t * pIfObj, int v )
func ifObjSetSatVar(pIfObj *IfObj, v int) {
	panic("TODO, stubbed")
}

// If_ManNodeShape2_rec (the C first parameter is a sat_solver *)
// C: void If_ManNodeShape2_rec( sat_solver * pSat, If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Int_t * vShape )
func (pSat *SatSolver) IfManNodeShape2_rec(pIfMan *IfMan, pIfObj *IfObj, vShape *IntVector) {
	panic("TODO, stubbed")
}

// If_ManNodeShapeSat
// C: int If_ManNodeShapeSat( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Int_t * vShape )
func (pIfMan *IfMan) IfManNodeShapeSat(pIfObj *IfObj, vShape *IntVector) int {
	panic("TODO, stubbed")
}

// If_ManCheckShape
// C: int If_ManCheckShape( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Int_t * vShape )
func (pIfMan *IfMan) IfManCheckShape(pIfObj *IfObj, vShape *IntVector) int {
	panic("TODO, stubbed")
}

// If_ManNodeShape
// C: int If_ManNodeShape( If_Man_t * pIfMan, If_Obj_t * pIfObj, Vec_Int_t * vShape, int fExact )
func (pIfMan *IfMan) IfManNodeShape(pIfObj *IfObj, vShape *IntVector, fExact int) int {
	panic("TODO, stubbed")
}
