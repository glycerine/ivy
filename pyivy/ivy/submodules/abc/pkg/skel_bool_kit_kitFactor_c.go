package abc

// Hand-checked skeleton of src/bool/kit/kitFactor.c.

// Kit_SopFactor
// C: Kit_Graph_t * Kit_SopFactor( Vec_Int_t * vCover, int fCompl, int nVars, Vec_Int_t * vMemory )
func KitSopFactor(vCover *IntVector, fCompl int, nVars int, vMemory *IntVector) *KitGraph {
	panic("TODO, stubbed")
}

// Kit_SopFactor_rec (static)
// C: Kit_Edge_t Kit_SopFactor_rec( Kit_Graph_t * pFForm, Kit_Sop_t * cSop, int nLits, Vec_Int_t * vMemory )
func kitSopFactor_rec(pFForm *KitGraph, cSop *KitSop, nLits int, vMemory *IntVector) KitEdge {
	panic("TODO, stubbed")
}

// Kit_SopFactorLF_rec (static)
// C: Kit_Edge_t Kit_SopFactorLF_rec( Kit_Graph_t * pFForm, Kit_Sop_t * cSop, Kit_Sop_t * cSimple, int nLits, Vec_Int_t * vMemory )
func kitSopFactorLF_rec(pFForm *KitGraph, cSop *KitSop, cSimple *KitSop, nLits int, vMemory *IntVector) KitEdge {
	panic("TODO, stubbed")
}

// Kit_SopFactorTrivialCube_rec
// C: Kit_Edge_t Kit_SopFactorTrivialCube_rec( Kit_Graph_t * pFForm, unsigned uCube, int nStart, int nFinish )
func KitSopFactorTrivialCube_rec(pFForm *KitGraph, uCube uint32, nStart int, nFinish int) KitEdge {
	panic("TODO, stubbed")
}

// Kit_SopFactorTrivialCube (static)
// C: Kit_Edge_t Kit_SopFactorTrivialCube( Kit_Graph_t * pFForm, unsigned uCube, int nLits )
func kitSopFactorTrivialCube(pFForm *KitGraph, uCube uint32, nLits int) KitEdge {
	panic("TODO, stubbed")
}

// Kit_SopFactorTrivial_rec
// C: Kit_Edge_t Kit_SopFactorTrivial_rec( Kit_Graph_t * pFForm, unsigned * pCubes, int nCubes, int nLits )
func KitSopFactorTrivial_rec(pFForm *KitGraph, pCubes []uint32, nCubes int, nLits int) KitEdge {
	panic("TODO, stubbed")
}

// Kit_SopFactorTrivial (static)
// C: Kit_Edge_t Kit_SopFactorTrivial( Kit_Graph_t * pFForm, Kit_Sop_t * cSop, int nLits )
func kitSopFactorTrivial(pFForm *KitGraph, cSop *KitSop, nLits int) KitEdge {
	panic("TODO, stubbed")
}

// Kit_FactorTest
// C: void Kit_FactorTest( unsigned * pTruth, int nVars )
func KitFactorTest(pTruth []uint32, nVars int) {
	panic("TODO, stubbed")
}
