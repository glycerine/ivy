package abc

// Hand-checked skeleton of src/aig/gia/giaSpeedup.c.

// Gia_LutDelayTraceSortPins
// C: void Gia_LutDelayTraceSortPins( Gia_Man_t * p, int iObj, int * pPinPerm, float * pPinDelays )
func (p *GiaManager) GiaLutDelayTraceSortPins(iObj int, pPinPerm []int, pPinDelays []float32) {
	panic("TODO, stubbed")
}

// Gia_LutWhereIsPin
// C: int Gia_LutWhereIsPin( Gia_Man_t * p, int iFanout, int iFanin, int * pPinPerm )
func (p *GiaManager) GiaLutWhereIsPin(iFanout int, iFanin int, pPinPerm []int) int {
	panic("TODO, stubbed")
}

// Gia_ObjComputeArrival
// C: float Gia_ObjComputeArrival( Gia_Man_t * p, int iObj, int fUseSorting )
func (p *GiaManager) GiaObjComputeArrival(iObj int, fUseSorting int) float32 {
	panic("TODO, stubbed")
}

// Gia_ObjPropagateRequired
// C: float Gia_ObjPropagateRequired( Gia_Man_t * p, int iObj, int fUseSorting )
func (p *GiaManager) GiaObjPropagateRequired(iObj int, fUseSorting int) float32 {
	panic("TODO, stubbed")
}

// Gia_ManDelayTraceLut
// C: float Gia_ManDelayTraceLut( Gia_Man_t * p )
func (p *GiaManager) GiaManDelayTraceLut() float32 {
	panic("TODO, stubbed")
}

// Gia_ManDelayTraceLutPrint
// C: float Gia_ManDelayTraceLutPrint( Gia_Man_t * p, int fVerbose )
func (p *GiaManager) GiaManDelayTraceLutPrint(fVerbose int) float32 {
	panic("TODO, stubbed")
}

// Gia_LutDelayTraceTCEdges
// C: unsigned Gia_LutDelayTraceTCEdges( Gia_Man_t * p, int iObj, float tDelta )
func (p *GiaManager) GiaLutDelayTraceTCEdges(iObj int, tDelta float32) uint32 {
	panic("TODO, stubbed")
}

// Gia_ManSpeedupObj_rec
// C: int Gia_ManSpeedupObj_rec( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vNodes )
func (p *GiaManager) GiaManSpeedupObj_rec(pObj *GiaObject, vNodes *IntVector) int {
	panic("TODO, stubbed")
}

// Gia_ManSpeedupObj
// C: void Gia_ManSpeedupObj( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vLeaves, Vec_Int_t * vTimes )
func (pNew *GiaManager) GiaManSpeedupObj(p *GiaManager, pObj *GiaObject, vLeaves *IntVector, vTimes *IntVector) {
	panic("TODO, stubbed")
}

// Gia_ManSpeedup
// C: Gia_Man_t * Gia_ManSpeedup( Gia_Man_t * p, int Percentage, int Degree, int fVerbose, int fVeryVerbose )
func (p *GiaManager) GiaManSpeedup(Percentage int, Degree int, fVerbose int, fVeryVerbose int) *GiaManager {
	panic("TODO, stubbed")
}

// Gia_ObjComputeRequired
// C: float Gia_ObjComputeRequired( Gia_Man_t * p, int iObj, int fUseSorting )
func (p *GiaManager) GiaObjComputeRequired(iObj int, fUseSorting int) float32 {
	panic("TODO, stubbed")
}

// Gia_LutVerifyTiming
// C: int Gia_LutVerifyTiming(  Gia_Man_t * p )
func (p *GiaManager) GiaLutVerifyTiming() int {
	panic("TODO, stubbed")
}
