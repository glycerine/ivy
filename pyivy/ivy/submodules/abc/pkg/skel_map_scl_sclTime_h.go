package abc

// Hand-checked skeleton of src/map/scl/sclTime.h.

// SCTime is SC_Time (struct SC_Time_).
type SCTime struct {
	pLib  *SCLib     // library
	vCis  *IntVector // comb inputs
	vCos  *IntVector // comb outputs
	nObjs int        // allocated size
	// get assignment
	vGates *IntVector // mapping of objId into gateId
	// timing information
	pLoads     []SCPair    // SC_Pair * : loads for each gate
	pLoads2    []SCPair    // SC_Pair * : loads for each gate
	pDepts     []SCPair    // SC_Pair * : departures for each gate
	pTimes     []SCPair    // SC_Pair * : arrivals for each gate
	pSlews     []SCPair    // SC_Pair * : slews for each gate
	pTimes2    []SCPair    // SC_Pair * : arrivals for each gate
	pSlews2    []SCPair    // SC_Pair * : slews for each gate
	pSlack     []float32   // float * : slacks for each gate
	pWLoadUsed *SCWireLoad // name of the used WireLoad model
	// optimization parameters
	SumArea   float32 // total area
	MaxDelay  float32 // max delay
	SumArea0  float32 // total area at the begining
	MaxDelay0 float32 // max delay at the begining
	BestDelay float32 // best delay in the middle
}

// Scl_ObjCell (static inline)
// C: SC_Cell * Scl_ObjCell( SC_Time * p, int i )
func (p *SCTime) sclObjCell(i int) *SCCell {
	panic("TODO, stubbed")
}

// Scl_ObjSetCell (static inline)
// C: void Scl_ObjSetCell( SC_Time * p, int i, SC_Cell * pCell )
func (p *SCTime) sclObjSetCell(i int, pCell *SCCell) {
	panic("TODO, stubbed")
}

// Scl_ObjLoad (static inline)
// C: SC_Pair * Scl_ObjLoad( SC_Time * p, int i )
func (p *SCTime) sclObjLoad(i int) *SCPair {
	panic("TODO, stubbed")
}

// Scl_ObjLoad2 (static inline)
// C: SC_Pair * Scl_ObjLoad2( SC_Time * p, int i )
func (p *SCTime) sclObjLoad2(i int) *SCPair {
	panic("TODO, stubbed")
}

// Scl_ObjDept (static inline)
// C: SC_Pair * Scl_ObjDept( SC_Time * p, int i )
func (p *SCTime) sclObjDept(i int) *SCPair {
	panic("TODO, stubbed")
}

// Scl_ObjTime (static inline)
// C: SC_Pair * Scl_ObjTime( SC_Time * p, int i )
func (p *SCTime) sclObjTime(i int) *SCPair {
	panic("TODO, stubbed")
}

// Scl_ObjSlew (static inline)
// C: SC_Pair * Scl_ObjSlew( SC_Time * p, int i )
func (p *SCTime) sclObjSlew(i int) *SCPair {
	panic("TODO, stubbed")
}

// Scl_ObjTime2 (static inline)
// C: SC_Pair * Scl_ObjTime2( SC_Time * p, int i )
func (p *SCTime) sclObjTime2(i int) *SCPair {
	panic("TODO, stubbed")
}

// Scl_ObjSlew2 (static inline)
// C: SC_Pair * Scl_ObjSlew2( SC_Time * p, int i )
func (p *SCTime) sclObjSlew2(i int) *SCPair {
	panic("TODO, stubbed")
}

// Scl_ObjTimeMax (static inline)
// C: float Scl_ObjTimeMax( SC_Time * p, int i )
func (p *SCTime) sclObjTimeMax(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ObjDepthMax (static inline)
// C: float Scl_ObjDepthMax( SC_Time * p, int i )
func (p *SCTime) sclObjDepthMax(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ObjGetSlack (static inline)
// C: float Scl_ObjGetSlack( SC_Time * p, int i, float D )
func (p *SCTime) sclObjGetSlack(i int, D float32) float32 {
	panic("TODO, stubbed")
}

// Scl_ObjGetSlackR (static inline)
// C: float Scl_ObjGetSlackR( SC_Time * p, int i, float D )
func (p *SCTime) sclObjGetSlackR(i int, D float32) float32 {
	panic("TODO, stubbed")
}

// Scl_ObjGetSlackF (static inline)
// C: float Scl_ObjGetSlackF( SC_Time * p, int i, float D )
func (p *SCTime) sclObjGetSlackF(i int, D float32) float32 {
	panic("TODO, stubbed")
}

// Scl_ObjSlack (static inline)
// C: float Scl_ObjSlack( SC_Time * p, int i )
func (p *SCTime) sclObjSlack(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ObjDupFanin (static inline)
// C: void Scl_ObjDupFanin( SC_Time * p, int i, int iFanin )
func (p *SCTime) sclObjDupFanin(i int, iFanin int) {
	panic("TODO, stubbed")
}

// Scl_ObjGain (static inline)
// C: float Scl_ObjGain( SC_Time * p, int i )
func (p *SCTime) sclObjGain(i int) float32 {
	panic("TODO, stubbed")
}

// Scl_ObjLegal (static inline)
// C: int Scl_ObjLegal( SC_Time * p, int i, float D )
func (p *SCTime) sclObjLegal(i int, D float32) int {
	panic("TODO, stubbed")
}

// Scl_ObjLoadFf (static inline)
// C: double Scl_ObjLoadFf( SC_Time * p, int i, int fRise )
func (p *SCTime) sclObjLoadFf(i int, fRise int) float64 {
	panic("TODO, stubbed")
}

// Scl_ObjTimePs (static inline)
// C: double Scl_ObjTimePs( SC_Time * p, int i, int fRise )
func (p *SCTime) sclObjTimePs(i int, fRise int) float64 {
	panic("TODO, stubbed")
}

// Scl_ObjSlewPs (static inline)
// C: double Scl_ObjSlewPs( SC_Time * p, int i, int fRise )
func (p *SCTime) sclObjSlewPs(i int, fRise int) float64 {
	panic("TODO, stubbed")
}

// Scl_ManAlloc (static inline)
// C: SC_Time * Scl_ManAlloc( SC_Lib * pLib, Vec_Int_t * vCis, Vec_Int_t * vCos, int nObjs )
func (pLib *SCLib) sclManAlloc(vCis *IntVector, vCos *IntVector, nObjs int) *SCTime {
	panic("TODO, stubbed")
}

// Scl_ManFree (static inline)
// C: void Scl_ManFree( SC_Time * p )
func (p *SCTime) sclManFree() {
	panic("TODO, stubbed")
}

// Scl_ManCleanTime (static inline)
// C: void Scl_ManCleanTime( SC_Time * p )
func (p *SCTime) sclManCleanTime() {
	panic("TODO, stubbed")
}

// Scl_ConeStore (static inline)
// C: void Scl_ConeStore( SC_Time * p, Vec_Int_t * vCone )
func (p *SCTime) sclConeStore(vCone *IntVector) {
	panic("TODO, stubbed")
}

// Scl_ConeRestore (static inline)
// C: void Scl_ConeRestore( SC_Time * p, Vec_Int_t * vCone )
func (p *SCTime) sclConeRestore(vCone *IntVector) {
	panic("TODO, stubbed")
}

// Scl_ConeClear (static inline)
// C: void Scl_ConeClear( SC_Time * p, Vec_Int_t * vCone )
func (p *SCTime) sclConeClear(vCone *IntVector) {
	panic("TODO, stubbed")
}

// Scl_PinTimeArrival (static inline)
// C: void Scl_PinTimeArrival( SC_Time * p, SC_Timing * pTime, int iObj, int iFanin )
func (p *SCTime) sclPinTimeArrival(pTime *SCTiming, iObj int, iFanin int) {
	panic("TODO, stubbed")
}

// Scl_PinTimeDeparture (static inline)
// C: void Scl_PinTimeDeparture( SC_Time * p, SC_Timing * pTime, int iObj, int iFanin )
func (p *SCTime) sclPinTimeDeparture(pTime *SCTiming, iObj int, iFanin int) {
	panic("TODO, stubbed")
}
