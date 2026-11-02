package abc

// Hand-checked skeleton of src/proof/fra/fraHot.c.

// Fra_RegToLit (static inline)
// C: int Fra_RegToLit( int n, int c )
func fraRegToLit(n int, c int) int {
	panic("TODO, stubbed")
}

// Fra_LitReg (static inline)
// C: int Fra_LitReg( int n )
func fraLitReg(n int) int {
	panic("TODO, stubbed")
}

// Fra_LitSign (static inline)
// C: int Fra_LitSign( int n )
func fraLitSign(n int) int {
	panic("TODO, stubbed")
}

// Fra_OneHotNodeIsConst
// C: int Fra_OneHotNodeIsConst( Fra_Sml_t * pSeq, Aig_Obj_t * pObj )
func (pSeq *FraSml) FraOneHotNodeIsConst(pObj *AigObject) int {
	panic("TODO, stubbed")
}

// Fra_OneHotNodesAreEqual
// C: int Fra_OneHotNodesAreEqual( Fra_Sml_t * pSeq, Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 )
func (pSeq *FraSml) FraOneHotNodesAreEqual(pObj0 *AigObject, pObj1 *AigObject) int {
	panic("TODO, stubbed")
}

// Fra_OneHotNodesAreClause
// C: int Fra_OneHotNodesAreClause( Fra_Sml_t * pSeq, Aig_Obj_t * pObj1, Aig_Obj_t * pObj2, int fCompl1, int fCompl2 )
func (pSeq *FraSml) FraOneHotNodesAreClause(pObj1 *AigObject, pObj2 *AigObject, fCompl1 int, fCompl2 int) int {
	panic("TODO, stubbed")
}

// Fra_OneHotCompute
// C: Vec_Int_t * Fra_OneHotCompute( Fra_Man_t * p, Fra_Sml_t * pSim )
func (p *FraMan) FraOneHotCompute(pSim *FraSml) *IntVector {
	panic("TODO, stubbed")
}

// Fra_OneHotAssume
// C: void Fra_OneHotAssume( Fra_Man_t * p, Vec_Int_t * vOneHots )
func (p *FraMan) FraOneHotAssume(vOneHots *IntVector) {
	panic("TODO, stubbed")
}

// Fra_OneHotCheck
// C: void Fra_OneHotCheck( Fra_Man_t * p, Vec_Int_t * vOneHots )
func (p *FraMan) FraOneHotCheck(vOneHots *IntVector) {
	panic("TODO, stubbed")
}

// Fra_OneHotRefineUsingCex
// C: int Fra_OneHotRefineUsingCex( Fra_Man_t * p, Vec_Int_t * vOneHots )
func (p *FraMan) FraOneHotRefineUsingCex(vOneHots *IntVector) int {
	panic("TODO, stubbed")
}

// Fra_OneHotCount
// C: int Fra_OneHotCount( Fra_Man_t * p, Vec_Int_t * vOneHots )
func (p *FraMan) FraOneHotCount(vOneHots *IntVector) int {
	panic("TODO, stubbed")
}

// Fra_OneHotEstimateCoverage
// C: void Fra_OneHotEstimateCoverage( Fra_Man_t * p, Vec_Int_t * vOneHots )
func (p *FraMan) FraOneHotEstimateCoverage(vOneHots *IntVector) {
	panic("TODO, stubbed")
}

// Fra_OneHotCreateExdc
// C: Aig_Man_t * Fra_OneHotCreateExdc( Fra_Man_t * p, Vec_Int_t * vOneHots )
func (p *FraMan) FraOneHotCreateExdc(vOneHots *IntVector) *AigManager {
	panic("TODO, stubbed")
}

// Fra_OneHotAddKnownConstraint
// C: void Fra_OneHotAddKnownConstraint( Fra_Man_t * p, Vec_Ptr_t * vOnehots )
// (the Vec_Ptr_t holds Vec_Int_t * groups)
func (p *FraMan) FraOneHotAddKnownConstraint(vOnehots *PointerVector[IntVector]) {
	panic("TODO, stubbed")
}
