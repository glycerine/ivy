package abc

// Hand-checked skeleton of src/aig/saig/saigSimFast.c.

// FaigMan is Faig_Man_t: the fast AIG manager.
type FaigMan struct {
	// parameters
	nPis int
	nPos int
	nCis int
	nCos int
	nFfs int
	nNos int
	// offsets
	nPis1       int
	nCis1       int
	nCisNos1    int
	nCisNosPos1 int
	nObjs       int
	// allocated data
	nWords int
	pObjs  []int // int pObjs[0] : flexible array member
}

// Faig_CheckIdPi (static inline)
// C: int Faig_CheckIdPi( Faig_Man_t * p, int i )
func (p *FaigMan) faigCheckIdPi(i int) int {
	panic("TODO, stubbed")
}

// Faig_CheckIdLo (static inline)
// C: int Faig_CheckIdLo( Faig_Man_t * p, int i )
func (p *FaigMan) faigCheckIdLo(i int) int {
	panic("TODO, stubbed")
}

// Faig_CheckIdNo (static inline)
// C: int Faig_CheckIdNo( Faig_Man_t * p, int i )
func (p *FaigMan) faigCheckIdNo(i int) int {
	panic("TODO, stubbed")
}

// Faig_CheckIdPo (static inline)
// C: int Faig_CheckIdPo( Faig_Man_t * p, int i )
func (p *FaigMan) faigCheckIdPo(i int) int {
	panic("TODO, stubbed")
}

// Faig_CheckIdLi (static inline)
// C: int Faig_CheckIdLi( Faig_Man_t * p, int i )
func (p *FaigMan) faigCheckIdLi(i int) int {
	panic("TODO, stubbed")
}

// Faig_CheckIdCo (static inline)
// C: int Faig_CheckIdCo( Faig_Man_t * p, int i )
func (p *FaigMan) faigCheckIdCo(i int) int {
	panic("TODO, stubbed")
}

// Faig_CheckIdObj (static inline)
// C: int Faig_CheckIdObj( Faig_Man_t * p, int i )
func (p *FaigMan) faigCheckIdObj(i int) int {
	panic("TODO, stubbed")
}

// Faig_ObjIdToNumPi (static inline)
// C: int Faig_ObjIdToNumPi( Faig_Man_t * p, int i )
func (p *FaigMan) faigObjIdToNumPi(i int) int {
	panic("TODO, stubbed")
}

// Faig_ObjIdToNumLo (static inline)
// C: int Faig_ObjIdToNumLo( Faig_Man_t * p, int i )
func (p *FaigMan) faigObjIdToNumLo(i int) int {
	panic("TODO, stubbed")
}

// Faig_ObjIdToNumNo (static inline)
// C: int Faig_ObjIdToNumNo( Faig_Man_t * p, int i )
func (p *FaigMan) faigObjIdToNumNo(i int) int {
	panic("TODO, stubbed")
}

// Faig_ObjIdToNumPo (static inline)
// C: int Faig_ObjIdToNumPo( Faig_Man_t * p, int i )
func (p *FaigMan) faigObjIdToNumPo(i int) int {
	panic("TODO, stubbed")
}

// Faig_ObjIdToNumLi (static inline)
// C: int Faig_ObjIdToNumLi( Faig_Man_t * p, int i )
func (p *FaigMan) faigObjIdToNumLi(i int) int {
	panic("TODO, stubbed")
}

// Faig_ObjIdToNumCo (static inline)
// C: int Faig_ObjIdToNumCo( Faig_Man_t * p, int i )
func (p *FaigMan) faigObjIdToNumCo(i int) int {
	panic("TODO, stubbed")
}

// Faig_ObjLoToLi (static inline)
// C: int Faig_ObjLoToLi( Faig_Man_t * p, int i )
func (p *FaigMan) faigObjLoToLi(i int) int {
	panic("TODO, stubbed")
}

// Faig_ObjLiToLo (static inline)
// C: int Faig_ObjLiToLo( Faig_Man_t * p, int i )
func (p *FaigMan) faigObjLiToLo(i int) int {
	panic("TODO, stubbed")
}

// Faig_NodeChild0 (static inline)
// C: int Faig_NodeChild0( Faig_Man_t * p, int n )
func (p *FaigMan) faigNodeChild0(n int) int {
	panic("TODO, stubbed")
}

// Faig_NodeChild1 (static inline)
// C: int Faig_NodeChild1( Faig_Man_t * p, int n )
func (p *FaigMan) faigNodeChild1(n int) int {
	panic("TODO, stubbed")
}

// Faig_CoChild0 (static inline)
// C: int Faig_CoChild0( Faig_Man_t * p, int n )
func (p *FaigMan) faigCoChild0(n int) int {
	panic("TODO, stubbed")
}

// Faig_ObjFaninC (static inline)
// C: int Faig_ObjFaninC( int iFan )
func faigObjFaninC(iFan int) int {
	panic("TODO, stubbed")
}

// Faig_ObjFanin (static inline)
// C: int Faig_ObjFanin( int iFan )
func faigObjFanin(iFan int) int {
	panic("TODO, stubbed")
}

// Faig_ManIsCorrect
// C: int Faig_ManIsCorrect( Aig_Man_t * pAig )
func (pAig *AigManager) FaigManIsCorrect() int {
	panic("TODO, stubbed")
}

// Faig_ManAlloc
// C: Faig_Man_t * Faig_ManAlloc( Aig_Man_t * pAig )
func (pAig *AigManager) FaigManAlloc() *FaigMan {
	panic("TODO, stubbed")
}

// Faig_ManCreate
// C: Faig_Man_t * Faig_ManCreate( Aig_Man_t * pAig )
func (pAig *AigManager) FaigManCreate() *FaigMan {
	panic("TODO, stubbed")
}

// Faig_SimulateNode (static inline)
// C: unsigned Faig_SimulateNode( Faig_Man_t * p, int Id, unsigned * pSimInfo )
func (p *FaigMan) faigSimulateNode(Id int, pSimInfo []uint32) uint32 {
	panic("TODO, stubbed")
}

// Faig_SimulateCo (static inline)
// C: unsigned Faig_SimulateCo( Faig_Man_t * p, int Id, unsigned * pSimInfo )
func (p *FaigMan) faigSimulateCo(Id int, pSimInfo []uint32) uint32 {
	panic("TODO, stubbed")
}

// Faig_SimulateRandomShift (static inline)
// C: unsigned Faig_SimulateRandomShift( unsigned uOld )
func faigSimulateRandomShift(uOld uint32) uint32 {
	panic("TODO, stubbed")
}

// Faig_SimulateTransferShift (static inline)
// C: unsigned Faig_SimulateTransferShift( unsigned uOld, unsigned uNew )
func faigSimulateTransferShift(uOld uint32, uNew uint32) uint32 {
	panic("TODO, stubbed")
}

// Faig_ManSimulateFrames
// C: int * Faig_ManSimulateFrames( Faig_Man_t * p, int nFrames, int nPref, int fTrans )
func (p *FaigMan) FaigManSimulateFrames(nFrames int, nPref int, fTrans int) []int {
	panic("TODO, stubbed")
}

// Faig_ManComputeSwitching
// C: float Faig_ManComputeSwitching( int nOnes, int nSimWords )
func FaigManComputeSwitching(nOnes int, nSimWords int) float32 {
	panic("TODO, stubbed")
}

// Faig_ManComputeProbOne
// C: float Faig_ManComputeProbOne( int nOnes, int nSimWords )
func FaigManComputeProbOne(nOnes int, nSimWords int) float32 {
	panic("TODO, stubbed")
}

// Faig_ManComputeSwitchProbs4
// C: Vec_Int_t * Faig_ManComputeSwitchProbs4( Aig_Man_t * p, int nFrames, int nPref, int fProbOne )
func (p *AigManager) FaigManComputeSwitchProbs4(nFrames int, nPref int, fProbOne int) *IntVector {
	panic("TODO, stubbed")
}

// Saig_ManComputeSwitchProb3s
// C: Vec_Int_t * Saig_ManComputeSwitchProb3s( Aig_Man_t * p, int nFrames, int nPref, int fProbOne )
func (p *AigManager) SaigManComputeSwitchProb3s(nFrames int, nPref int, fProbOne int) *IntVector {
	panic("TODO, stubbed")
}
