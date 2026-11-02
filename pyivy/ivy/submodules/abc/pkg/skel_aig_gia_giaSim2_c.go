package abc

// Hand-checked skeleton of src/aig/gia/giaSim2.c.

// GiaSim2 is Gia_Sim2_t.
type GiaSim2 struct {
	pAig      *GiaManager
	pPars     *GiaParSim
	nWords    int
	pDataSim  []uint32 // unsigned * : simulation data
	vClassOld *IntVector
	vClassNew *IntVector
}

// Gia_Sim2Data (static inline)
// C: unsigned * Gia_Sim2Data( Gia_Sim2_t * p, int i )
func (p *GiaSim2) giaSim2Data(i int) []uint32 {
	panic("TODO, stubbed")
}

// Gia_Sim2Delete
// C: void Gia_Sim2Delete( Gia_Sim2_t * p )
func (p *GiaSim2) GiaSim2Delete() {
	panic("TODO, stubbed")
}

// Gia_Sim2Create
// C: Gia_Sim2_t * Gia_Sim2Create( Gia_Man_t * pAig, Gia_ParSim_t * pPars )
func (pAig *GiaManager) GiaSim2Create(pPars *GiaParSim) *GiaSim2 {
	panic("TODO, stubbed")
}

// Gia_Sim2InfoRandom (static inline)
// C: void Gia_Sim2InfoRandom( Gia_Sim2_t * p, unsigned * pInfo )
func (p *GiaSim2) giaSim2InfoRandom(pInfo []uint32) {
	panic("TODO, stubbed")
}

// Gia_Sim2InfoZero (static inline)
// C: void Gia_Sim2InfoZero( Gia_Sim2_t * p, unsigned * pInfo )
func (p *GiaSim2) giaSim2InfoZero(pInfo []uint32) {
	panic("TODO, stubbed")
}

// Gia_Sim2InfoOne (static inline)
// C: void Gia_Sim2InfoOne( Gia_Sim2_t * p, unsigned * pInfo )
func (p *GiaSim2) giaSim2InfoOne(pInfo []uint32) {
	panic("TODO, stubbed")
}

// Gia_Sim2InfoCopy (static inline)
// C: void Gia_Sim2InfoCopy( Gia_Sim2_t * p, unsigned * pInfo, unsigned * pInfo0 )
func (p *GiaSim2) giaSim2InfoCopy(pInfo []uint32, pInfo0 []uint32) {
	panic("TODO, stubbed")
}

// Gia_Sim2SimulateCo (static inline)
// C: void Gia_Sim2SimulateCo( Gia_Sim2_t * p, Gia_Obj_t * pObj )
func (p *GiaSim2) giaSim2SimulateCo(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_Sim2SimulateNode (static inline)
// C: void Gia_Sim2SimulateNode( Gia_Sim2_t * p, Gia_Obj_t * pObj )
func (p *GiaSim2) giaSim2SimulateNode(pObj *GiaObject) {
	panic("TODO, stubbed")
}

// Gia_Sim2InfoTransfer (static inline)
// C: void Gia_Sim2InfoTransfer( Gia_Sim2_t * p )
func (p *GiaSim2) giaSim2InfoTransfer() {
	panic("TODO, stubbed")
}

// Gia_Sim2SimulateRound (static inline)
// C: void Gia_Sim2SimulateRound( Gia_Sim2_t * p )
func (p *GiaSim2) giaSim2SimulateRound() {
	panic("TODO, stubbed")
}

// Gia_Sim2CompareEqual
// C: int Gia_Sim2CompareEqual( unsigned * p0, unsigned * p1, int nWords, int fCompl )
func GiaSim2CompareEqual(p0 []uint32, p1 []uint32, nWords int, fCompl int) int {
	panic("TODO, stubbed")
}

// Gia_Sim2CompareZero
// C: int Gia_Sim2CompareZero( unsigned * p0, int nWords, int fCompl )
func GiaSim2CompareZero(p0 []uint32, nWords int, fCompl int) int {
	panic("TODO, stubbed")
}

// Gia_Sim2ClassCreate
// C: void Gia_Sim2ClassCreate( Gia_Man_t * p, Vec_Int_t * vClass )
func (p *GiaManager) GiaSim2ClassCreate(vClass *IntVector) {
	panic("TODO, stubbed")
}

// Gia_Sim2ClassRefineOne
// C: int Gia_Sim2ClassRefineOne( Gia_Sim2_t * p, int i )
func (p *GiaSim2) GiaSim2ClassRefineOne(i int) int {
	panic("TODO, stubbed")
}

// Gia_Sim2HashKey
// C: int Gia_Sim2HashKey( unsigned * pSim, int nWords, int nTableSize )
func GiaSim2HashKey(pSim []uint32, nWords int, nTableSize int) int {
	panic("TODO, stubbed")
}

// Gia_Sim2ProcessRefined
// C: void Gia_Sim2ProcessRefined( Gia_Sim2_t * p, Vec_Int_t * vRefined )
func (p *GiaSim2) GiaSim2ProcessRefined(vRefined *IntVector) {
	panic("TODO, stubbed")
}

// Gia_Sim2InfoRefineEquivs
// C: void Gia_Sim2InfoRefineEquivs( Gia_Sim2_t * p )
func (p *GiaSim2) GiaSim2InfoRefineEquivs() {
	panic("TODO, stubbed")
}

// Gia_Sim2InfoIsZero (static inline)
// C: int Gia_Sim2InfoIsZero( Gia_Sim2_t * p, unsigned * pInfo )
func (p *GiaSim2) giaSim2InfoIsZero(pInfo []uint32) int {
	panic("TODO, stubbed")
}

// Gia_Sim2CheckPos (static inline)
// C: int Gia_Sim2CheckPos( Gia_Sim2_t * p, int * piPo, int * piPat )
func (p *GiaSim2) giaSim2CheckPos(piPo *int, piPat *int) int {
	panic("TODO, stubbed")
}

// Gia_Sim2GenerateCounter
// C: Abc_Cex_t * Gia_Sim2GenerateCounter( Gia_Man_t * pAig, int iFrame, int iOut, int nWords, int iPat )
func (pAig *GiaManager) GiaSim2GenerateCounter(iFrame int, iOut int, nWords int, iPat int) *Counterexample {
	panic("TODO, stubbed")
}

// Gia_ManSimSimulateEquiv
// C: int Gia_ManSimSimulateEquiv( Gia_Man_t * pAig, Gia_ParSim_t * pPars )
func (pAig *GiaManager) GiaManSimSimulateEquiv(pPars *GiaParSim) int {
	panic("TODO, stubbed")
}
