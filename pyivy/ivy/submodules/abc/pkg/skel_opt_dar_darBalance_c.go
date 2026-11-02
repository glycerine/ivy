package abc

// Hand-checked skeleton of src/opt/dar/darBalance.c.

// Dar_ObjCompareLits
// C: int Dar_ObjCompareLits( Aig_Obj_t ** pp1, Aig_Obj_t ** pp2 )
func DarObjCompareLits(pp1 **AigObject, pp2 **AigObject) int {
	panic("TODO, stubbed")
}

// Dar_BalanceUniqify
// C: void Dar_BalanceUniqify( Aig_Obj_t * pObj, Vec_Ptr_t * vNodes, int fExor )
func DarBalanceUniqify(pObj *AigObject, vNodes *PointerVector[AigObject], fExor int) {
	panic("TODO, stubbed")
}

// Dar_BalanceCone_rec
// C: void Dar_BalanceCone_rec( Aig_Obj_t * pRoot, Aig_Obj_t * pObj, Vec_Ptr_t * vSuper )
func DarBalanceCone_rec(pRoot *AigObject, pObj *AigObject, vSuper *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Dar_BalanceCone
// C: Vec_Ptr_t * Dar_BalanceCone( Aig_Obj_t * pObj, Vec_Vec_t * vStore, int Level )
func DarBalanceCone(pObj *AigObject, vStore *NestedVector[*AigObject], Level int) *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Dar_BalanceFindLeft
// C: int Dar_BalanceFindLeft( Vec_Ptr_t * vSuper )
func DarBalanceFindLeft(vSuper *PointerVector[AigObject]) int {
	panic("TODO, stubbed")
}

// Dar_BalancePermute
// C: void Dar_BalancePermute( Aig_Man_t * p, Vec_Ptr_t * vSuper, int LeftBound, int fExor )
func (p *AigManager) DarBalancePermute(vSuper *PointerVector[AigObject], LeftBound int, fExor int) {
	panic("TODO, stubbed")
}

// Aig_NodeCompareLevelsDecrease
// C: int Aig_NodeCompareLevelsDecrease( Aig_Obj_t ** pp1, Aig_Obj_t ** pp2 )
func AigNodeCompareLevelsDecrease(pp1 **AigObject, pp2 **AigObject) int {
	panic("TODO, stubbed")
}

// Dar_BalancePushUniqueOrderByLevel
// C: void Dar_BalancePushUniqueOrderByLevel( Vec_Ptr_t * vStore, Aig_Obj_t * pObj, int fExor )
func DarBalancePushUniqueOrderByLevel(vStore *PointerVector[AigObject], pObj *AigObject, fExor int) {
	panic("TODO, stubbed")
}

// Dar_BalanceBuildSuper
// C: Aig_Obj_t * Dar_BalanceBuildSuper( Aig_Man_t * p, Vec_Ptr_t * vSuper, Aig_Type_t Type, int fUpdateLevel )
func (p *AigManager) DarBalanceBuildSuper(vSuper *PointerVector[AigObject], Type AigType, fUpdateLevel int) *AigObject {
	panic("TODO, stubbed")
}

// Aig_BaseSize
// C: int Aig_BaseSize( Aig_Man_t * p, Aig_Obj_t * pObj, int nLutSize )
func (p *AigManager) AigBaseSize(pObj *AigObject, nLutSize int) int {
	panic("TODO, stubbed")
}

// Dar_BalanceBuildSuperTop
// C: Aig_Obj_t * Dar_BalanceBuildSuperTop( Aig_Man_t * p, Vec_Ptr_t * vSuper, Aig_Type_t Type, int fUpdateLevel, int nLutSize )
func (p *AigManager) DarBalanceBuildSuperTop(vSuper *PointerVector[AigObject], Type AigType, fUpdateLevel int, nLutSize int) *AigObject {
	panic("TODO, stubbed")
}

// Dar_Balance_rec
// C: Aig_Obj_t * Dar_Balance_rec( Aig_Man_t * pNew, Aig_Obj_t * pObjOld, Vec_Vec_t * vStore, int Level, int fUpdateLevel )
func (pNew *AigManager) DarBalance_rec(pObjOld *AigObject, vStore *NestedVector[*AigObject], Level int, fUpdateLevel int) *AigObject {
	panic("TODO, stubbed")
}

// Dar_ManBalance
// C: Aig_Man_t * Dar_ManBalance( Aig_Man_t * p, int fUpdateLevel )
func (p *AigManager) DarManBalance(fUpdateLevel int) *AigManager {
	panic("TODO, stubbed")
}

// Dar_ManBalanceXor
// C: Aig_Man_t * Dar_ManBalanceXor( Aig_Man_t * pAig, int fExor, int fUpdateLevel, int fVerbose )
func (pAig *AigManager) DarManBalanceXor(fExor int, fUpdateLevel int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Dar_BalancePrintStats
// C: void Dar_BalancePrintStats( Aig_Man_t * p )
func (p *AigManager) DarBalancePrintStats() {
	panic("TODO, stubbed")
}
