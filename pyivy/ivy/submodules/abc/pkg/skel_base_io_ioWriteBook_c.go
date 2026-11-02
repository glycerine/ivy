package abc

import "os"

// Hand-checked skeleton of src/base/io/ioWriteBook.c.

// Abc_Rint (static inline)
// C: double Abc_Rint( double x )
func abcRint(x float64) float64 {
	panic("TODO, stubbed")
}

// Io_WriteBookLogic
// C: void Io_WriteBookLogic( Abc_Ntk_t * pNtk, char * FileName )
func IoWriteBookLogic(pNtk *ABCNetwork, FileName string) {
	panic("TODO, stubbed")
}

// Io_WriteBook
// C: void Io_WriteBook( Abc_Ntk_t * pNtk, char * FileName )
func IoWriteBook(pNtk *ABCNetwork, FileName string) {
	panic("TODO, stubbed")
}

// Io_NtkWriteNodes (static)
// C: unsigned Io_NtkWriteNodes( FILE * pFile, Abc_Ntk_t * pNtk )
func ioNtkWriteNodes(pFile *os.File, pNtk *ABCNetwork) uint32 {
	panic("TODO, stubbed")
}

// Io_NtkWritePiPoNodes (static)
// C: void Io_NtkWritePiPoNodes( FILE * pFile, Abc_Ntk_t * pNtk )
func ioNtkWritePiPoNodes(pFile *os.File, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Io_NtkWriteLatchNode (static)
// C: void Io_NtkWriteLatchNode( FILE * pFile, Abc_Obj_t * pLatch, int NodesOrPl )
func ioNtkWriteLatchNode(pFile *os.File, pLatch *ABCNetworkObject, NodesOrPl int) {
	panic("TODO, stubbed")
}

// Io_NtkWriteIntNode (static)
// C: unsigned Io_NtkWriteIntNode( FILE * pFile, Abc_Obj_t * pNode, int NodesOrPl )
func ioNtkWriteIntNode(pFile *os.File, pNode *ABCNetworkObject, NodesOrPl int) uint32 {
	panic("TODO, stubbed")
}

// Io_NtkWriteNodeGate (static)
// C: unsigned Io_NtkWriteNodeGate( FILE * pFile, Abc_Obj_t * pNode )
func ioNtkWriteNodeGate(pFile *os.File, pNode *ABCNetworkObject) uint32 {
	panic("TODO, stubbed")
}

// Io_NtkWriteNets (static)
// C: void Io_NtkWriteNets( FILE * pFile, Abc_Ntk_t * pNtk )
func ioNtkWriteNets(pFile *os.File, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Io_NtkWriteIntNet (static)
// C: void Io_NtkWriteIntNet( FILE * pFile, Abc_Obj_t * pNet )
func ioNtkWriteIntNet(pFile *os.File, pNet *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Io_NtkBuildLayout (static; the forward declaration names the file parameters pFile1, pFile2)
// C: void Io_NtkBuildLayout( FILE * pFileScl, FILE * pFilePl, Abc_Ntk_t * pNtk, double aspectRatio, double whiteSpace, unsigned coreCellArea )
func ioNtkBuildLayout(pFileScl *os.File, pFilePl *os.File, pNtk *ABCNetwork, aspectRatio float64, whiteSpace float64, coreCellArea uint32) {
	panic("TODO, stubbed")
}

// Io_NtkWriteScl (static)
// C: void Io_NtkWriteScl( FILE * pFile, unsigned numCoreRows, double layoutWidth )
func ioNtkWriteScl(pFile *os.File, numCoreRows uint32, layoutWidth float64) {
	panic("TODO, stubbed")
}

// Io_NtkWritePl (static; the forward declaration lists layoutHeight before layoutWidth, the definition (followed here) lists layoutWidth first)
// C: void Io_NtkWritePl( FILE * pFile, Abc_Ntk_t * pNtk, unsigned numTerms, double layoutWidth, double layoutHeight )
func ioNtkWritePl(pFile *os.File, pNtk *ABCNetwork, numTerms uint32, layoutWidth float64, layoutHeight float64) {
	panic("TODO, stubbed")
}

// Io_NtkOrderingPads (static)
// C: Vec_Ptr_t * Io_NtkOrderingPads( Abc_Ntk_t * pNtk, Vec_Ptr_t * vTerms )
func ioNtkOrderingPads(pNtk *ABCNetwork, vTerms *PointerVector[ABCNetworkObject]) *PointerVector[ABCNetworkObject] {
	panic("TODO, stubbed")
}

// Io_NtkBfsPads (static; the forward declaration names pTerm as pCurrEntry)
// C: Abc_Obj_t * Io_NtkBfsPads( Abc_Ntk_t * pNtk, Abc_Obj_t * pTerm, unsigned numTerms, int * pOrdered )
func ioNtkBfsPads(pNtk *ABCNetwork, pTerm *ABCNetworkObject, numTerms uint32, pOrdered []int) *ABCNetworkObject {
	panic("TODO, stubbed")
}

// Abc_NodeIsNand2 (static)
// C: int Abc_NodeIsNand2( Abc_Obj_t * pNode )
func abcNodeIsNand2(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeIsNor2 (static)
// C: int Abc_NodeIsNor2( Abc_Obj_t * pNode )
func abcNodeIsNor2(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeIsAnd2 (static)
// C: int Abc_NodeIsAnd2( Abc_Obj_t * pNode )
func abcNodeIsAnd2(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeIsOr2 (static)
// C: int Abc_NodeIsOr2( Abc_Obj_t * pNode )
func abcNodeIsOr2(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeIsXor2 (static)
// C: int Abc_NodeIsXor2( Abc_Obj_t * pNode )
func abcNodeIsXor2(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}

// Abc_NodeIsXnor2 (static)
// C: int Abc_NodeIsXnor2( Abc_Obj_t * pNode )
func abcNodeIsXnor2(pNode *ABCNetworkObject) int {
	panic("TODO, stubbed")
}
