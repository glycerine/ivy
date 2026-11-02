package abc

// Hand-checked skeleton of src/opt/mfs/mfsStrash.c.

// Abc_MfsConvertAigToHop_rec
// C: void Abc_MfsConvertAigToHop_rec( Aig_Obj_t * pObj, Hop_Man_t * pHop )
func AbcMfsConvertAigToHop_rec(pObj *AigObject, pHop *HopManager) {
	panic("TODO, stubbed")
}

// Abc_MfsConvertAigToHop
// C: Hop_Obj_t * Abc_MfsConvertAigToHop( Aig_Man_t * pMan, Hop_Man_t * pHop )
func (pMan *AigManager) AbcMfsConvertAigToHop(pHop *HopManager) *HopObject {
	panic("TODO, stubbed")
}

// Abc_MfsConvertHopToAig_rec
// C: void Abc_MfsConvertHopToAig_rec( Hop_Obj_t * pObj, Aig_Man_t * pMan )
func AbcMfsConvertHopToAig_rec(pObj *HopObject, pMan *AigManager) {
	panic("TODO, stubbed")
}

// Abc_MfsConvertHopToAig
// C: void Abc_MfsConvertHopToAig( Abc_Obj_t * pObjOld, Aig_Man_t * pMan )
func AbcMfsConvertHopToAig(pObjOld *ABCNetworkObject, pMan *AigManager) {
	panic("TODO, stubbed")
}

// Abc_NtkConstructAig_rec
// C: Aig_Obj_t * Abc_NtkConstructAig_rec( Mfs_Man_t * p, Abc_Obj_t * pNode, Aig_Man_t * pMan )
func (p *MfsMan) AbcNtkConstructAig_rec(pNode *ABCNetworkObject, pMan *AigManager) *AigObject {
	panic("TODO, stubbed")
}

// Abc_NtkConstructCare_rec
// C: Aig_Obj_t * Abc_NtkConstructCare_rec( Aig_Man_t * pCare, Aig_Obj_t * pObj, Aig_Man_t * pMan )
func (pCare *AigManager) AbcNtkConstructCare_rec(pObj *AigObject, pMan *AigManager) *AigObject {
	panic("TODO, stubbed")
}

// Abc_NtkConstructAig
// C: Aig_Man_t * Abc_NtkConstructAig( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkConstructAig(pNode *ABCNetworkObject) *AigManager {
	panic("TODO, stubbed")
}

// Abc_NtkAigForConstraints
// C: Aig_Man_t * Abc_NtkAigForConstraints( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkAigForConstraints(pNode *ABCNetworkObject) *AigManager {
	panic("TODO, stubbed")
}

// Abc_NtkConstraintRatio
// C: double Abc_NtkConstraintRatio( Mfs_Man_t * p, Abc_Obj_t * pNode )
func (p *MfsMan) AbcNtkConstraintRatio(pNode *ABCNetworkObject) float64 {
	panic("TODO, stubbed")
}
