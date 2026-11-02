package abc

// Hand-checked skeleton of src/map/scl/sclLoad.c.

// Abc_SclFindWireCaps
// C: Vec_Flt_t * Abc_SclFindWireCaps( SC_WireLoad * pWL, int nFanoutMax )
func AbcSclFindWireCaps(pWL *SCWireLoad, nFanoutMax int) *FloatVector {
	panic("TODO, stubbed")
}

// Abc_SclFindWireLoad
// C: float Abc_SclFindWireLoad( Vec_Flt_t * vWireCaps, int nFans )
func AbcSclFindWireLoad(vWireCaps *FloatVector, nFans int) float32 {
	panic("TODO, stubbed")
}

// Abc_SclAddWireLoad
// C: void Abc_SclAddWireLoad( SC_Man * p, Abc_Obj_t * pObj, int fSubtr )
func (p *SCMan) AbcSclAddWireLoad(pObj *ABCNetworkObject, fSubtr int) {
	panic("TODO, stubbed")
}

// Abc_SclComputeLoad
// C: void Abc_SclComputeLoad( SC_Man * p )
func (p *SCMan) AbcSclComputeLoad() {
	panic("TODO, stubbed")
}

// Abc_SclUpdateLoad
// C: void Abc_SclUpdateLoad( SC_Man * p, Abc_Obj_t * pObj, SC_Cell * pOld, SC_Cell * pNew )
func (p *SCMan) AbcSclUpdateLoad(pObj *ABCNetworkObject, pOld *SCCell, pNew *SCCell) {
	panic("TODO, stubbed")
}

// Abc_SclUpdateLoadSplit
// C: void Abc_SclUpdateLoadSplit( SC_Man * p, Abc_Obj_t * pBuffer, Abc_Obj_t * pFanout )
func (p *SCMan) AbcSclUpdateLoadSplit(pBuffer *ABCNetworkObject, pFanout *ABCNetworkObject) {
	panic("TODO, stubbed")
}
