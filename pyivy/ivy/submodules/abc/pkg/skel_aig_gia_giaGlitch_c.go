package abc

// Hand-checked skeleton of src/aig/gia/giaGlitch.c.

// GliObj is Gli_Obj_t.
type GliObj struct {
	fTerm    uint32   // unsigned : 1 : terminal node
	fPhase   uint32   // unsigned : 1 : value under 000 pattern
	fPhase2  uint32   // unsigned : 1 : value under 000 pattern
	fMark    uint32   // unsigned : 1 : user-controlled mark
	nFanins  uint32   // unsigned : 3 : the number of fanins
	nFanouts uint32   // unsigned : 25 : total number of fanouts
	Handle   uint32   // ID of the node
	pTruth   []uint64 // word * : truth table of the node
	uSimInfo uint32   // simulation info of the node
	// union { int iFanin; int nSwitches; }
	iFanin    int // the number of fanins added
	nSwitches int // the number of switches
	// union { int iFanout; int nGlitches; }
	iFanout   int   // the number of fanouts added
	nGlitches int   // the number of glitches ( nGlitches >= nSwitches )
	Fanios    []int // int Fanios[0] : the array of fanins/fanouts (flexible array member)
}

// GliMan is Gli_Man_t.
type GliMan struct {
	vCis         *IntVector // the vector of CIs (PIs + LOs)
	vCos         *IntVector // the vector of COs (POs + LIs)
	vCisChanged  *IntVector // the changed CIs
	vAffected    *IntVector // the affected nodes
	vFrontier    *IntVector // the fanouts of these nodes
	nObjs        int        // the number of objects
	nRegs        int        // the number of registers
	nTravIds     int        // traversal ID of the network
	iObjData     int        // pointer to the current data
	nObjData     int        // the size of array to store the logic network
	pObjData     []int      // int * : the internal nodes
	pSimInfoPrev []uint32   // unsigned * : previous values of the CIs
}

// Gli_ManCiNum (static inline)
// C: int Gli_ManCiNum( Gli_Man_t * p )
func (p *GliMan) gliManCiNum() int {
	panic("TODO, stubbed")
}

// Gli_ManCoNum (static inline)
// C: int Gli_ManCoNum( Gli_Man_t * p )
func (p *GliMan) gliManCoNum() int {
	panic("TODO, stubbed")
}

// Gli_ManPiNum (static inline)
// C: int Gli_ManPiNum( Gli_Man_t * p )
func (p *GliMan) gliManPiNum() int {
	panic("TODO, stubbed")
}

// Gli_ManPoNum (static inline)
// C: int Gli_ManPoNum( Gli_Man_t * p )
func (p *GliMan) gliManPoNum() int {
	panic("TODO, stubbed")
}

// Gli_ManRegNum (static inline)
// C: int Gli_ManRegNum( Gli_Man_t * p )
func (p *GliMan) gliManRegNum() int {
	panic("TODO, stubbed")
}

// Gli_ManObjNum (static inline)
// C: int Gli_ManObjNum( Gli_Man_t * p )
func (p *GliMan) gliManObjNum() int {
	panic("TODO, stubbed")
}

// Gli_ManNodeNum (static inline)
// C: int Gli_ManNodeNum( Gli_Man_t * p )
func (p *GliMan) gliManNodeNum() int {
	panic("TODO, stubbed")
}

// Gli_ManObj (static inline)
// C: Gli_Obj_t * Gli_ManObj( Gli_Man_t * p, int v )
func (p *GliMan) gliManObj(v int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ManCi (static inline)
// C: Gli_Obj_t * Gli_ManCi( Gli_Man_t * p, int v )
func (p *GliMan) gliManCi(v int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ManCo (static inline)
// C: Gli_Obj_t * Gli_ManCo( Gli_Man_t * p, int v )
func (p *GliMan) gliManCo(v int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ManPi (static inline)
// C: Gli_Obj_t * Gli_ManPi( Gli_Man_t * p, int v )
func (p *GliMan) gliManPi(v int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ManPo (static inline)
// C: Gli_Obj_t * Gli_ManPo( Gli_Man_t * p, int v )
func (p *GliMan) gliManPo(v int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ManRo (static inline)
// C: Gli_Obj_t * Gli_ManRo( Gli_Man_t * p, int v )
func (p *GliMan) gliManRo(v int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ManRi (static inline)
// C: Gli_Obj_t * Gli_ManRi( Gli_Man_t * p, int v )
func (p *GliMan) gliManRi(v int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ObjIsTerm (static inline)
// C: int Gli_ObjIsTerm( Gli_Obj_t * pObj )
func gliObjIsTerm(pObj *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_ObjIsCi (static inline)
// C: int Gli_ObjIsCi( Gli_Obj_t * pObj )
func gliObjIsCi(pObj *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_ObjIsCo (static inline)
// C: int Gli_ObjIsCo( Gli_Obj_t * pObj )
func gliObjIsCo(pObj *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_ObjIsNode (static inline)
// C: int Gli_ObjIsNode( Gli_Obj_t * pObj )
func gliObjIsNode(pObj *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_ObjFaninNum (static inline)
// C: int Gli_ObjFaninNum( Gli_Obj_t * pObj )
func gliObjFaninNum(pObj *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_ObjFanoutNum (static inline)
// C: int Gli_ObjFanoutNum( Gli_Obj_t * pObj )
func gliObjFanoutNum(pObj *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_ObjSize (static inline)
// C: int Gli_ObjSize( Gli_Obj_t * pObj )
func gliObjSize(pObj *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_ObjFanin (static inline)
// C: Gli_Obj_t * Gli_ObjFanin( Gli_Obj_t * pObj, int i )
func gliObjFanin(pObj *GliObj, i int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ObjFanout (static inline)
// C: Gli_Obj_t * Gli_ObjFanout( Gli_Obj_t * pObj, int i )
func gliObjFanout(pObj *GliObj, i int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ManAlloc
// C: Gli_Man_t * Gli_ManAlloc( int nObjs, int nRegs, int nFanioPairs )
func GliManAlloc(nObjs int, nRegs int, nFanioPairs int) *GliMan {
	panic("TODO, stubbed")
}

// Gli_ManStop
// C: void Gli_ManStop( Gli_Man_t * p )
func (p *GliMan) GliManStop() {
	panic("TODO, stubbed")
}

// Gli_ManPrintObjects
// C: void Gli_ManPrintObjects( Gli_Man_t * p )
func (p *GliMan) GliManPrintObjects() {
	panic("TODO, stubbed")
}

// Gli_ManFinalize
// C: void Gli_ManFinalize( Gli_Man_t * p )
func (p *GliMan) GliManFinalize() {
	panic("TODO, stubbed")
}

// Gli_ObjAddFanin
// C: void Gli_ObjAddFanin( Gli_Obj_t * pObj, Gli_Obj_t * pFanin )
func GliObjAddFanin(pObj *GliObj, pFanin *GliObj) {
	panic("TODO, stubbed")
}

// Gli_ObjAlloc
// C: Gli_Obj_t * Gli_ObjAlloc( Gli_Man_t * p, int nFanins, int nFanouts )
func (p *GliMan) GliObjAlloc(nFanins int, nFanouts int) *GliObj {
	panic("TODO, stubbed")
}

// Gli_ManCreateCi
// C: int Gli_ManCreateCi( Gli_Man_t * p, int nFanouts )
func (p *GliMan) GliManCreateCi(nFanouts int) int {
	panic("TODO, stubbed")
}

// Gli_ManCreateCo
// C: int Gli_ManCreateCo( Gli_Man_t * p, int iFanin )
func (p *GliMan) GliManCreateCo(iFanin int) int {
	panic("TODO, stubbed")
}

// Gli_NodeComputeValue (static inline)
// C: int Gli_NodeComputeValue( Gli_Obj_t * pNode )
func gliNodeComputeValue(pNode *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_NodeComputeValue2 (static inline)
// C: int Gli_NodeComputeValue2( Gli_Obj_t * pNode )
func gliNodeComputeValue2(pNode *GliObj) int {
	panic("TODO, stubbed")
}

// Gli_ManCreateNode
// C: int Gli_ManCreateNode( Gli_Man_t * p, Vec_Int_t * vFanins, int nFanouts, word * pGateTruth )
func (p *GliMan) GliManCreateNode(vFanins *IntVector, nFanouts int, pGateTruth []uint64) int {
	panic("TODO, stubbed")
}

// Gli_ObjNumSwitches
// C: int Gli_ObjNumSwitches( Gli_Man_t * p, int iNode )
func (p *GliMan) GliObjNumSwitches(iNode int) int {
	panic("TODO, stubbed")
}

// Gli_ObjNumGlitches
// C: int Gli_ObjNumGlitches( Gli_Man_t * p, int iNode )
func (p *GliMan) GliObjNumGlitches(iNode int) int {
	panic("TODO, stubbed")
}

// Gli_ManSetPiRandom
// C: void Gli_ManSetPiRandom( Gli_Man_t * p, float PiTransProb )
func (p *GliMan) GliManSetPiRandom(PiTransProb float32) {
	panic("TODO, stubbed")
}

// Gli_ManSetPiFromSaved
// C: void Gli_ManSetPiFromSaved( Gli_Man_t * p, int iBit )
func (p *GliMan) GliManSetPiFromSaved(iBit int) {
	panic("TODO, stubbed")
}

// Gli_ManSwitching
// C: void Gli_ManSwitching( Gli_Man_t * p )
func (p *GliMan) GliManSwitching() {
	panic("TODO, stubbed")
}

// Gli_ManGlitching
// C: void Gli_ManGlitching( Gli_Man_t * p )
func (p *GliMan) GliManGlitching() {
	panic("TODO, stubbed")
}

// Gli_ManVerify
// C: void Gli_ManVerify( Gli_Man_t * p )
func (p *GliMan) GliManVerify() {
	panic("TODO, stubbed")
}

// Gli_ManSimulateSeqNode
// C: unsigned Gli_ManSimulateSeqNode( Gli_Man_t * p, Gli_Obj_t * pNode )
func (p *GliMan) GliManSimulateSeqNode(pNode *GliObj) uint32 {
	panic("TODO, stubbed")
}

// Gli_ManUpdateRandomInput (static inline)
// C: unsigned Gli_ManUpdateRandomInput( unsigned uInfo, float PiTransProb )
func gliManUpdateRandomInput(uInfo uint32, PiTransProb float32) uint32 {
	panic("TODO, stubbed")
}

// Gli_ManSimulateSeqPref
// C: void Gli_ManSimulateSeqPref( Gli_Man_t * p, int nPref )
func (p *GliMan) GliManSimulateSeqPref(nPref int) {
	panic("TODO, stubbed")
}

// Gli_ManSetDataSaved
// C: void Gli_ManSetDataSaved( Gli_Man_t * p, int iBit )
func (p *GliMan) GliManSetDataSaved(iBit int) {
	panic("TODO, stubbed")
}

// Gli_ManSetPiRandomSeq
// C: void Gli_ManSetPiRandomSeq( Gli_Man_t * p, float PiTransProb )
func (p *GliMan) GliManSetPiRandomSeq(PiTransProb float32) {
	panic("TODO, stubbed")
}

// Gli_ManSwitchesAndGlitches
// C: void Gli_ManSwitchesAndGlitches( Gli_Man_t * p, int nPatterns, float PiTransProb, int fVerbose )
func (p *GliMan) GliManSwitchesAndGlitches(nPatterns int, PiTransProb float32, fVerbose int) {
	panic("TODO, stubbed")
}
