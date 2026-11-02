package abc

// Hand-checked skeleton of src/aig/gia/giaFalse.c.

// Gia_ManFalseRebuildOne
// C: void Gia_ManFalseRebuildOne( Gia_Man_t * pNew, Gia_Man_t * p, Vec_Int_t * vHook, int fVerbose, int fVeryVerbose )
func (pNew *GiaManager) GiaManFalseRebuildOne(p *GiaManager, vHook *IntVector, fVerbose int, fVeryVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManFalseRebuild
// C: Gia_Man_t * Gia_ManFalseRebuild( Gia_Man_t * p, Vec_Wec_t * vHooks, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManFalseRebuild(vHooks *IntLevelVector, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCollectPath_rec
// C: void Gia_ManCollectPath_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vPath )
func (p *GiaManager) GiaManCollectPath_rec(pObj *GiaObject, vPath *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManCollectPath
// C: Vec_Int_t * Gia_ManCollectPath( Gia_Man_t * p, Gia_Obj_t * pObj )
func (p *GiaManager) GiaManCollectPath(pObj *GiaObject) *IntVector {
	panic("TODO, stubbed")
}

// Gia_ManCheckFalseOne
// C: void Gia_ManCheckFalseOne( Gia_Man_t * p, int iOut, int nTimeOut, Vec_Wec_t * vHooks, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManCheckFalseOne(iOut int, nTimeOut int, vHooks *IntLevelVector, fVerbose int, fVeryVerbose int) {
	panic("TODO, stubbed")
}

// Gia_ManCheckFalse2
// C: Gia_Man_t * Gia_ManCheckFalse2( Gia_Man_t * p, int nSlackMax, int nTimeOut, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManCheckFalse2(nSlackMax int, nTimeOut int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManFalseRebuildPath
// C: Gia_Man_t * Gia_ManFalseRebuildPath( Gia_Man_t * p, Vec_Int_t * vHooks, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManFalseRebuildPath(vHooks *IntVector, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCheckOne
// C: Gia_Man_t * Gia_ManCheckOne( Gia_Man_t * p, int iOut, int iObj, int nTimeOut, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManCheckOne(iOut int, iObj int, nTimeOut int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCheckFalseAll
// C: Gia_Man_t * Gia_ManCheckFalseAll( Gia_Man_t * p, int nSlackMax, int nTimeOut, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManCheckFalseAll(nSlackMax int, nTimeOut int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ManCheckFalse
// C: Gia_Man_t * Gia_ManCheckFalse( Gia_Man_t * p, int nSlackMax, int nTimeOut, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManCheckFalse(nSlackMax int, nTimeOut int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
