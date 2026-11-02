package abc

// Hand-checked skeleton of src/opt/sbd/sbdCut2.c.

// The Sbd_Cut_t struct is the same as in sbdCut.c (SbdCut).

// SbdSrv is Sbd_Srv_t.
type SbdSrv struct {
	nLutSize int
	nCutSize int
	nCutNum  int
	fVerbose int
	pGia     *GiaManager  // user's AIG manager (will be modified by adding nodes)
	vMirrors *IntVector   // mirrors for each node
	vLutLevs *IntVector   // delays for each node
	vLevs    *IntVector   // levels for each node
	vRefs    *IntVector   // refs for each node
	pCuts    [501]SbdCut  // temporary cuts
	ppCuts   [501]*SbdCut // temporary cut pointers
	clkStart int64        // abctime : starting time
	vCut0    *IntVector   // current cut
	vCut     *IntVector   // current cut
	vCutTop  *IntVector   // current cut
	vCutBot  *IntVector   // current cut
}

// Sbd_ManCutServerStart
// C: Sbd_Srv_t * Sbd_ManCutServerStart( Gia_Man_t * pGia, Vec_Int_t * vMirrors, Vec_Int_t * vLutLevs, Vec_Int_t * vLevs, Vec_Int_t * vRefs, int nLutSize, int nCutSize, int nCutNum, int fVerbose )
func (pGia *GiaManager) SbdManCutServerStart(vMirrors *IntVector, vLutLevs *IntVector, vLevs *IntVector, vRefs *IntVector, nLutSize int, nCutSize int, nCutNum int, fVerbose int) *SbdSrv {
	panic("TODO, stubbed")
}

// Sbd_ManCutServerStop
// C: void Sbd_ManCutServerStop( Sbd_Srv_t * p )
func (p *SbdSrv) SbdManCutServerStop() {
	panic("TODO, stubbed")
}

// Sbd_ManCutIsTopo_rec
// C: int Sbd_ManCutIsTopo_rec( Gia_Man_t * p, Vec_Int_t * vMirrors, int iObj )
func (p *GiaManager) SbdManCutIsTopo_rec(vMirrors *IntVector, iObj int) int {
	panic("TODO, stubbed")
}

// Sbd_ManCutIsTopo
// C: int Sbd_ManCutIsTopo( Gia_Man_t * p, Vec_Int_t * vMirrors, Vec_Int_t * vCut, int iObj )
func (p *GiaManager) SbdManCutIsTopo(vMirrors *IntVector, vCut *IntVector, iObj int) int {
	panic("TODO, stubbed")
}

// Sbd_ManCutExpandOne (static)
// C: int Sbd_ManCutExpandOne( Gia_Man_t * p, Vec_Int_t * vMirrors, Vec_Int_t * vLutLevs, Vec_Int_t * vCut, int iThis, int iObj )
func (p *GiaManager) sbdManCutExpandOne(vMirrors *IntVector, vLutLevs *IntVector, vCut *IntVector, iThis int, iObj int) int {
	panic("TODO, stubbed")
}

// Vec_IntIsOrdered (defined in sbdCut2.c, not static)
// C: void Vec_IntIsOrdered( Vec_Int_t * vCut )
func VecIntIsOrdered(vCut *IntVector) {
	panic("TODO, stubbed")
}

// Sbd_ManCutExpand
// C: void Sbd_ManCutExpand( Gia_Man_t * p, Vec_Int_t * vMirrors, Vec_Int_t * vLutLevs, Vec_Int_t * vCut )
func (p *GiaManager) SbdManCutExpand(vMirrors *IntVector, vLutLevs *IntVector, vCut *IntVector) {
	panic("TODO, stubbed")
}

// Sbd_ManCutReload
// C: void Sbd_ManCutReload( Vec_Int_t * vMirrors, Vec_Int_t * vLutLevs, int LevStop, Vec_Int_t * vCut, Vec_Int_t * vCutTop, Vec_Int_t * vCutBot )
func SbdManCutReload(vMirrors *IntVector, vLutLevs *IntVector, LevStop int, vCut *IntVector, vCutTop *IntVector, vCutBot *IntVector) {
	panic("TODO, stubbed")
}

// Sbd_ManCutCollect_rec
// C: int Sbd_ManCutCollect_rec( Gia_Man_t * p, Vec_Int_t * vMirrors, int iObj, int LevStop, Vec_Int_t * vLutLevs, Vec_Int_t * vCut )
func (p *GiaManager) SbdManCutCollect_rec(vMirrors *IntVector, iObj int, LevStop int, vLutLevs *IntVector, vCut *IntVector) int {
	panic("TODO, stubbed")
}

// Sbd_ManCutReduceTop
// C: int Sbd_ManCutReduceTop( Gia_Man_t * p, Vec_Int_t * vMirrors, int iObj, Vec_Int_t * vLutLevs, Vec_Int_t * vCut, Vec_Int_t * vCutTop, int nCutSize )
func (p *GiaManager) SbdManCutReduceTop(vMirrors *IntVector, iObj int, vLutLevs *IntVector, vCut *IntVector, vCutTop *IntVector, nCutSize int) int {
	panic("TODO, stubbed")
}

// Sbd_ManCutServerFirst
// C: int Sbd_ManCutServerFirst( Sbd_Srv_t * p, int iObj, int * pLeaves )
func (p *SbdSrv) SbdManCutServerFirst(iObj int, pLeaves []int) int {
	panic("TODO, stubbed")
}
