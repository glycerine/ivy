package abc

// Hand-checked skeleton of src/opt/sfm/sfmTim.c.

// SfmTim is Sfm_Tim_t.
type SfmTim struct {
	// external
	pLib      *MioLibrary // library
	pExt      *SclCon     // external timing
	pNtk      *ABCNetwork // mapped network
	Delay     int         // the largest delay
	DeltaCrit int         // critical delay delta
	// timing info
	vTimArrs IntVector // Vec_Int_t (by value) : arrivals (rise/fall)
	vTimReqs IntVector // Vec_Int_t (by value) : required (rise/fall)
	// incremental timing
	vLevels IntLevelVector // Vec_Wec_t (by value) : levels
	// critical path
	vPath     IntVector  // Vec_Int_t (by value) : critical path
	vSortData WordVector // Vec_Wrd_t (by value) : node priority order
}

// Sfm_TimArrId (static inline; a pointer to the rise/fall pair)
// C: int * Sfm_TimArrId( Sfm_Tim_t * p, int Id )
func (p *SfmTim) sfmTimArrId(Id int) []int {
	panic("TODO, stubbed")
}

// Sfm_TimReqId (static inline; a pointer to the rise/fall pair)
// C: int * Sfm_TimReqId( Sfm_Tim_t * p, int Id )
func (p *SfmTim) sfmTimReqId(Id int) []int {
	panic("TODO, stubbed")
}

// Sfm_TimArr (static inline; a pointer to the rise/fall pair)
// C: int * Sfm_TimArr( Sfm_Tim_t * p, Abc_Obj_t * pNode )
func (p *SfmTim) sfmTimArr(pNode *ABCNetworkObject) []int {
	panic("TODO, stubbed")
}

// Sfm_TimReq (static inline; a pointer to the rise/fall pair)
// C: int * Sfm_TimReq( Sfm_Tim_t * p, Abc_Obj_t * pNode )
func (p *SfmTim) sfmTimReq(pNode *ABCNetworkObject) []int {
	panic("TODO, stubbed")
}

// Sfm_TimArrMaxId (static inline)
// C: int Sfm_TimArrMaxId( Sfm_Tim_t * p, int Id )
func (p *SfmTim) sfmTimArrMaxId(Id int) int {
	panic("TODO, stubbed")
}

// Sfm_TimArrMax (static inline)
// C: int Sfm_TimArrMax( Sfm_Tim_t * p, Abc_Obj_t * pNode )
func (p *SfmTim) sfmTimArrMax(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Sfm_TimSetReq (static inline)
// C: void Sfm_TimSetReq( Sfm_Tim_t * p, Abc_Obj_t * pNode, int t )
func (p *SfmTim) sfmTimSetReq(pNode *ABCNetworkObject, t int) {
	panic("TODO, stubbed")
}

// Sfm_TimSlack (static inline)
// C: int Sfm_TimSlack( Sfm_Tim_t * p, Abc_Obj_t * pNode )
func (p *SfmTim) sfmTimSlack(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Sfm_TimEdgeArrival (static inline)
// C: void Sfm_TimEdgeArrival( Sfm_Tim_t * p, Mio_Pin_t * pPin, int * pTimeIn, int * pTimeOut )
func (p *SfmTim) sfmTimEdgeArrival(pPin *MioPin, pTimeIn []int, pTimeOut []int) {
	panic("TODO, stubbed")
}

// Sfm_TimGateArrival (static inline)
// C: void Sfm_TimGateArrival( Sfm_Tim_t * p, Mio_Gate_t * pGate, int ** pTimesIn, int * pTimeOut )
func (p *SfmTim) sfmTimGateArrival(pGate *MioGate, pTimesIn [][]int, pTimeOut []int) {
	panic("TODO, stubbed")
}

// Sfm_TimNodeArrival (static inline)
// C: void Sfm_TimNodeArrival( Sfm_Tim_t * p, Abc_Obj_t * pNode )
func (p *SfmTim) sfmTimNodeArrival(pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Sfm_TimEdgeRequired (static inline)
// C: void Sfm_TimEdgeRequired( Sfm_Tim_t * p, Mio_Pin_t * pPin, int * pTimeIn, int * pTimeOut )
func (p *SfmTim) sfmTimEdgeRequired(pPin *MioPin, pTimeIn []int, pTimeOut []int) {
	panic("TODO, stubbed")
}

// Sfm_TimGateRequired (static inline)
// C: void Sfm_TimGateRequired( Sfm_Tim_t * p, Mio_Gate_t * pGate, int ** pTimesIn, int * pTimeOut )
func (p *SfmTim) sfmTimGateRequired(pGate *MioGate, pTimesIn [][]int, pTimeOut []int) {
	panic("TODO, stubbed")
}

// Sfm_TimNodeRequired
// C: void Sfm_TimNodeRequired( Sfm_Tim_t * p, Abc_Obj_t * pNode )
func (p *SfmTim) SfmTimNodeRequired(pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Sfm_TimCriticalPath_int
// C: void Sfm_TimCriticalPath_int( Sfm_Tim_t * p, Abc_Obj_t * pObj, Vec_Int_t * vPath, int SlackMax )
func (p *SfmTim) SfmTimCriticalPath_int(pObj *ABCNetworkObject, vPath *IntVector, SlackMax int) {
	panic("TODO, stubbed")
}

// Sfm_TimCriticalPath
// C: int Sfm_TimCriticalPath( Sfm_Tim_t * p, int Window )
func (p *SfmTim) SfmTimCriticalPath(Window int) int {
	panic("TODO, stubbed")
}

// Sfm_TimTrace
// C: int Sfm_TimTrace( Sfm_Tim_t * p )
func (p *SfmTim) SfmTimTrace() int {
	panic("TODO, stubbed")
}

// Sfm_TimStart
// C: Sfm_Tim_t * Sfm_TimStart( Mio_Library_t * pLib, Scl_Con_t * pExt, Abc_Ntk_t * pNtk, int DeltaCrit )
func (pLib *MioLibrary) SfmTimStart(pExt *SclCon, pNtk *ABCNetwork, DeltaCrit int) *SfmTim {
	panic("TODO, stubbed")
}

// Sfm_TimStop
// C: void Sfm_TimStop( Sfm_Tim_t * p )
func (p *SfmTim) SfmTimStop() {
	panic("TODO, stubbed")
}

// Sfm_TimReadNtkDelay
// C: int Sfm_TimReadNtkDelay( Sfm_Tim_t * p )
func (p *SfmTim) SfmTimReadNtkDelay() int {
	panic("TODO, stubbed")
}

// Sfm_TimReadObjDelay
// C: int Sfm_TimReadObjDelay( Sfm_Tim_t * p, int iObj )
func (p *SfmTim) SfmTimReadObjDelay(iObj int) int {
	panic("TODO, stubbed")
}

// Sfm_TimTest
// C: void Sfm_TimTest( Abc_Ntk_t * pNtk )
func SfmTimTest(pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Sfm_TimUpdateClean (static)
// C: void Sfm_TimUpdateClean( Sfm_Tim_t * p )
func (p *SfmTim) sfmTimUpdateClean() {
	panic("TODO, stubbed")
}

// Sfm_TimUpdateTiming
// C: void Sfm_TimUpdateTiming( Sfm_Tim_t * p, Vec_Int_t * vTimeNodes )
func (p *SfmTim) SfmTimUpdateTiming(vTimeNodes *IntVector) {
	panic("TODO, stubbed")
}

// Sfm_TimSortArrayByArrival
// C: int Sfm_TimSortArrayByArrival( Sfm_Tim_t * p, Vec_Int_t * vNodes, int iPivot )
func (p *SfmTim) SfmTimSortArrayByArrival(vNodes *IntVector, iPivot int) int {
	panic("TODO, stubbed")
}

// Sfm_TimPriorityNodes
// C: int Sfm_TimPriorityNodes( Sfm_Tim_t * p, Vec_Int_t * vCands, int Window )
func (p *SfmTim) SfmTimPriorityNodes(vCands *IntVector, Window int) int {
	panic("TODO, stubbed")
}

// Sfm_TimNodeIsNonCritical
// C: int Sfm_TimNodeIsNonCritical( Sfm_Tim_t * p, Abc_Obj_t * pPivot, Abc_Obj_t * pNode )
func (p *SfmTim) SfmTimNodeIsNonCritical(pPivot *ABCNetworkObject, pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Sfm_TimEvalRemapping
// C: int Sfm_TimEvalRemapping( Sfm_Tim_t * p, Vec_Int_t * vFanins, Vec_Int_t * vMap, Mio_Gate_t * pGate1, char * pFans1, Mio_Gate_t * pGate2, char * pFans2 )
func (p *SfmTim) SfmTimEvalRemapping(vFanins *IntVector, vMap *IntVector, pGate1 *MioGate, pFans1 []int8, pGate2 *MioGate, pFans2 []int8) int {
	panic("TODO, stubbed")
}
