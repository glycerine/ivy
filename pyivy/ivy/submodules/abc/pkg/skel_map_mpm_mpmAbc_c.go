package abc

// Hand-checked skeleton of src/map/mpm/mpmAbc.c.

// Mig_ManCreateChoices
// C: void Mig_ManCreateChoices( Mig_Man_t * pMig, Gia_Man_t * p )
func (pMig *MigMan) MigManCreateChoices(p *GiaManager) {
	panic("TODO, stubbed")
}

// Mig_ObjFanin0Copy (static inline)
// C: int Mig_ObjFanin0Copy( Gia_Obj_t * pObj )
func migObjFanin0Copy(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Mig_ObjFanin1Copy (static inline)
// C: int Mig_ObjFanin1Copy( Gia_Obj_t * pObj )
func migObjFanin1Copy(pObj *GiaObject) int {
	panic("TODO, stubbed")
}

// Mig_ManCreate (pGia is a void * that holds a Gia_Man_t *)
// C: Mig_Man_t * Mig_ManCreate( void * pGia )
func MigManCreate(pGia *GiaManager) *MigMan {
	panic("TODO, stubbed")
}

// Mpm_CutDataInt (static inline)
// C: unsigned Mpm_CutDataInt( Mpm_Cut_t * pCut )
func mpmCutDataInt(pCut *MpmCut) uint32 {
	panic("TODO, stubbed")
}

// Mpm_CutSetDataInt (static inline)
// C: void Mpm_CutSetDataInt( Mpm_Cut_t * pCut, int Data )
func mpmCutSetDataInt(pCut *MpmCut, Data int) {
	panic("TODO, stubbed")
}

// Mpm_ManNodeIfToGia_rec
// C: int Mpm_ManNodeIfToGia_rec( Gia_Man_t * pNew, Mpm_Man_t * pMan, Mig_Obj_t * pObj, Vec_Ptr_t * vVisited, int fHash )
func (pNew *GiaManager) MpmManNodeIfToGia_rec(pMan *MpmMan, pObj *MigObj, vVisited *PointerVector[MpmCut], fHash int) int {
	panic("TODO, stubbed")
}

// Mpm_ManNodeIfToGia
// C: int Mpm_ManNodeIfToGia( Gia_Man_t * pNew, Mpm_Man_t * pMan, Mig_Obj_t * pObj, Vec_Int_t * vLeaves, int fHash )
func (pNew *GiaManager) MpmManNodeIfToGia(pMan *MpmMan, pObj *MigObj, vLeaves *IntVector, fHash int) int {
	panic("TODO, stubbed")
}

// Mpm_ManFromIfLogic (returns a void *)
// C: void * Mpm_ManFromIfLogic( Mpm_Man_t * pMan )
func (pMan *MpmMan) MpmManFromIfLogic() *GiaManager {
	panic("TODO, stubbed")
}
