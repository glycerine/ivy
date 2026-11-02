package abc

import "os"

// Hand-checked skeleton of src/base/io/ioWriteBlifMv.c.

// Io_WriteBlifMv
// C: void Io_WriteBlifMv( Abc_Ntk_t * pNtk, char * FileName )
func IoWriteBlifMv(pNtk *ABCNetwork, FileName string) {
	panic("TODO, stubbed")
}

// Io_NtkWriteBlifMv (static)
// C: void Io_NtkWriteBlifMv( FILE * pFile, Abc_Ntk_t * pNtk )
func ioNtkWriteBlifMv(pFile *os.File, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Io_NtkWriteBlifMvOne (static)
// C: void Io_NtkWriteBlifMvOne( FILE * pFile, Abc_Ntk_t * pNtk )
func ioNtkWriteBlifMvOne(pFile *os.File, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Io_NtkWriteBlifMvPis (static)
// C: void Io_NtkWriteBlifMvPis( FILE * pFile, Abc_Ntk_t * pNtk )
func ioNtkWriteBlifMvPis(pFile *os.File, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Io_NtkWriteBlifMvPos (static)
// C: void Io_NtkWriteBlifMvPos( FILE * pFile, Abc_Ntk_t * pNtk )
func ioNtkWriteBlifMvPos(pFile *os.File, pNtk *ABCNetwork) {
	panic("TODO, stubbed")
}

// Io_NtkWriteBlifMvLatch (static)
// C: void Io_NtkWriteBlifMvLatch( FILE * pFile, Abc_Obj_t * pLatch )
func ioNtkWriteBlifMvLatch(pFile *os.File, pLatch *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Io_NtkWriteBlifMvSubckt (static)
// C: void Io_NtkWriteBlifMvSubckt( FILE * pFile, Abc_Obj_t * pNode )
func ioNtkWriteBlifMvSubckt(pFile *os.File, pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Io_NtkWriteBlifMvNode (static)
// C: void Io_NtkWriteBlifMvNode( FILE * pFile, Abc_Obj_t * pNode )
func ioNtkWriteBlifMvNode(pFile *os.File, pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}

// Io_NtkWriteBlifMvNodeFanins (static)
// C: void Io_NtkWriteBlifMvNodeFanins( FILE * pFile, Abc_Obj_t * pNode )
func ioNtkWriteBlifMvNodeFanins(pFile *os.File, pNode *ABCNetworkObject) {
	panic("TODO, stubbed")
}
