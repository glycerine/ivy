package abc

// Hand-checked skeleton of src/base/wlc/wlcUif.c.

// Wlc_NtkCollectBoxes
// C: void Wlc_NtkCollectBoxes( Wlc_Ntk_t * p, Vec_Int_t * vBoxIds )
func WlcNtkCollectBoxes(p *WLCNetwork, vBoxIds *IntVector) {
	panic("TODO, stubbed")
}

// Wlc_NtkCollectAddMult (pCountA and pCountM are out parameters)
// C: Vec_Int_t * Wlc_NtkCollectAddMult( Wlc_Ntk_t * p, Wlc_BstPar_t * pPar, int * pCountA, int * pCountM )
func WlcNtkCollectAddMult(p *WLCNetwork, pPar *WlcBstPar, pCountA *int, pCountM *int) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkPairIsUifable
// C: int Wlc_NtkPairIsUifable( Wlc_Ntk_t * p, Wlc_Obj_t * pObj, Wlc_Obj_t * pObj2 )
func WlcNtkPairIsUifable(p *WLCNetwork, pObj *WLCObject, pObj2 *WLCObject) int {
	panic("TODO, stubbed")
}

// Wlc_NtkCollectMultipliers
// C: Vec_Int_t * Wlc_NtkCollectMultipliers( Wlc_Ntk_t * p )
func WlcNtkCollectMultipliers(p *WLCNetwork) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkFindUifableMultiplierPairs
// C: Vec_Int_t * Wlc_NtkFindUifableMultiplierPairs( Wlc_Ntk_t * p )
func WlcNtkFindUifableMultiplierPairs(p *WLCNetwork) *IntVector {
	panic("TODO, stubbed")
}

// Wlc_NtkAbstractNodes
// C: Wlc_Ntk_t * Wlc_NtkAbstractNodes( Wlc_Ntk_t * p, Vec_Int_t * vNodesInit )
func WlcNtkAbstractNodes(p *WLCNetwork, vNodesInit *IntVector) *WLCNetwork {
	panic("TODO, stubbed")
}

// Wlc_NtkUifNodePairs
// C: Wlc_Ntk_t * Wlc_NtkUifNodePairs( Wlc_Ntk_t * p, Vec_Int_t * vPairsInit )
func WlcNtkUifNodePairs(p *WLCNetwork, vPairsInit *IntVector) *WLCNetwork {
	panic("TODO, stubbed")
}
