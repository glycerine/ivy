package abc

// Hand-checked skeleton of src/aig/gia/giaCCof.c.

// CcfMan is Ccf_Man_t: the manager.
type CcfMan struct {
	// user data
	pGia      *GiaManager // single-output AIG manager
	nFrameMax int         // maximum number of frames
	nConfMax  int         // maximum number of conflicts
	nTimeMax  int         // maximum runtime in seconds
	fVerbose  int         // verbose flag
	// internal data
	pUnr    *GiaManUnr  // void * : unrolling manager
	pFrames *GiaManager // unrolled timeframes
	vCopies *IntVector  // copy pointers of the AIG
	pSat    *SatSolver  // SAT solver
}

// Ccf_ManStart
// C: Ccf_Man_t * Ccf_ManStart( Gia_Man_t * pGia, int nFrameMax, int nConfMax, int nTimeMax, int fVerbose )
func (pGia *GiaManager) CcfManStart(nFrameMax int, nConfMax int, nTimeMax int, fVerbose int) *CcfMan {
	panic("TODO, stubbed")
}

// Ccf_ManStop
// C: void Ccf_ManStop( Ccf_Man_t * p )
func (p *CcfMan) CcfManStop() {
	panic("TODO, stubbed")
}

// Gia_ManCofExtendSolver
// C: void Gia_ManCofExtendSolver( Ccf_Man_t * p )
func (p *CcfMan) GiaManCofExtendSolver() {
	panic("TODO, stubbed")
}

// Gia_Obj0Copy (static inline)
// C: int Gia_Obj0Copy( Vec_Int_t * vCopies, int Fan0, int fCompl0 )
func giaObj0Copy(vCopies *IntVector, Fan0 int, fCompl0 int) int {
	panic("TODO, stubbed")
}

// Gia_Obj1Copy (static inline)
// C: int Gia_Obj1Copy( Vec_Int_t * vCopies, int Fan1, int fCompl1 )
func giaObj1Copy(vCopies *IntVector, Fan1 int, fCompl1 int) int {
	panic("TODO, stubbed")
}

// Gia_ManCofOneDerive_rec
// C: void Gia_ManCofOneDerive_rec( Ccf_Man_t * p, int Id )
func (p *CcfMan) GiaManCofOneDerive_rec(Id int) {
	panic("TODO, stubbed")
}

// Gia_ManCofOneDerive
// C: int Gia_ManCofOneDerive( Ccf_Man_t * p, int LitProp )
func (p *CcfMan) GiaManCofOneDerive(LitProp int) int {
	panic("TODO, stubbed")
}

// Gia_ManCofGetReachable
// C: int Gia_ManCofGetReachable( Ccf_Man_t * p, int Lit )
func (p *CcfMan) GiaManCofGetReachable(Lit int) int {
	panic("TODO, stubbed")
}

// Gia_ManCofTest
// C: Gia_Man_t * Gia_ManCofTest( Gia_Man_t * pGia, int nFrameMax, int nConfMax, int nTimeMax, int fVerbose )
func (pGia *GiaManager) GiaManCofTest(nFrameMax int, nConfMax int, nTimeMax int, fVerbose int) *GiaManager {
	panic("TODO, stubbed")
}
