package abc

// Hand-checked skeleton of src/opt/sfm/sfmNtk.c.

// Sfm_CheckConsistency
// C: void Sfm_CheckConsistency( Vec_Wec_t * vFanins, int nPis, int nPos, Vec_Str_t * vFixed )
func SfmCheckConsistency(vFanins *IntLevelVector, nPis int, nPos int, vFixed *ByteVector) {
	panic("TODO, stubbed")
}

// Sfm_CreateFanout
// C: void Sfm_CreateFanout( Vec_Wec_t * vFanins, Vec_Wec_t * vFanouts )
func SfmCreateFanout(vFanins *IntLevelVector, vFanouts *IntLevelVector) {
	panic("TODO, stubbed")
}

// Sfm_ObjLevelNew (static)
// C: int Sfm_ObjLevelNew( Vec_Int_t * vArray, Vec_Int_t * vLevels, int fAddLevel )
func sfmObjLevelNew(vArray *IntVector, vLevels *IntVector, fAddLevel int) int {
	panic("TODO, stubbed")
}

// Sfm_CreateLevel
// C: void Sfm_CreateLevel( Vec_Wec_t * vFanins, Vec_Int_t * vLevels, Vec_Str_t * vEmpty )
func SfmCreateLevel(vFanins *IntLevelVector, vLevels *IntVector, vEmpty *ByteVector) {
	panic("TODO, stubbed")
}

// Sfm_ObjLevelNewR (static)
// C: int Sfm_ObjLevelNewR( Vec_Int_t * vArray, Vec_Int_t * vLevelsR, int fAddLevel )
func sfmObjLevelNewR(vArray *IntVector, vLevelsR *IntVector, fAddLevel int) int {
	panic("TODO, stubbed")
}

// Sfm_CreateLevelR
// C: void Sfm_CreateLevelR( Vec_Wec_t * vFanouts, Vec_Int_t * vLevelsR, Vec_Str_t * vEmpty )
func SfmCreateLevelR(vFanouts *IntLevelVector, vLevelsR *IntVector, vEmpty *ByteVector) {
	panic("TODO, stubbed")
}

// Sfm_NtkConstruct
// C: Sfm_Ntk_t * Sfm_NtkConstruct( Vec_Wec_t * vFanins, int nPis, int nPos, Vec_Str_t * vFixed, Vec_Str_t * vEmpty, Vec_Wrd_t * vTruths, Vec_Int_t * vStarts, Vec_Wrd_t * vTruths2 )
func SfmNtkConstruct(vFanins *IntLevelVector, nPis int, nPos int, vFixed *ByteVector, vEmpty *ByteVector, vTruths *WordVector, vStarts *IntVector, vTruths2 *WordVector) *SfmNtk {
	panic("TODO, stubbed")
}

// Sfm_NtkPrepare
// C: void Sfm_NtkPrepare( Sfm_Ntk_t * p )
func (p *SfmNtk) SfmNtkPrepare() {
	panic("TODO, stubbed")
}

// Sfm_NtkFree
// C: void Sfm_NtkFree( Sfm_Ntk_t * p )
func (p *SfmNtk) SfmNtkFree() {
	panic("TODO, stubbed")
}

// Sfm_NtkRemoveFanin
// C: void Sfm_NtkRemoveFanin( Sfm_Ntk_t * p, int iNode, int iFanin )
func (p *SfmNtk) SfmNtkRemoveFanin(iNode int, iFanin int) {
	panic("TODO, stubbed")
}

// Sfm_NtkAddFanin
// C: void Sfm_NtkAddFanin( Sfm_Ntk_t * p, int iNode, int iFanin )
func (p *SfmNtk) SfmNtkAddFanin(iNode int, iFanin int) {
	panic("TODO, stubbed")
}

// Sfm_NtkDeleteObj_rec
// C: void Sfm_NtkDeleteObj_rec( Sfm_Ntk_t * p, int iNode )
func (p *SfmNtk) SfmNtkDeleteObj_rec(iNode int) {
	panic("TODO, stubbed")
}

// Sfm_NtkUpdateLevel_rec
// C: void Sfm_NtkUpdateLevel_rec( Sfm_Ntk_t * p, int iNode )
func (p *SfmNtk) SfmNtkUpdateLevel_rec(iNode int) {
	panic("TODO, stubbed")
}

// Sfm_NtkUpdateLevelR_rec
// C: void Sfm_NtkUpdateLevelR_rec( Sfm_Ntk_t * p, int iNode )
func (p *SfmNtk) SfmNtkUpdateLevelR_rec(iNode int) {
	panic("TODO, stubbed")
}

// Sfm_NtkUpdate
// C: void Sfm_NtkUpdate( Sfm_Ntk_t * p, int iNode, int f, int iFaninNew, word uTruth )
func (p *SfmNtk) SfmNtkUpdate(iNode int, f int, iFaninNew int, uTruth uint64) {
	panic("TODO, stubbed")
}

// Sfm_NodeReadFanins
// C: Vec_Int_t * Sfm_NodeReadFanins( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) SfmNodeReadFanins(i int) *IntVector {
	panic("TODO, stubbed")
}

// Sfm_NodeReadTruth (a pointer into the truth-table storage)
// C: word * Sfm_NodeReadTruth( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) SfmNodeReadTruth(i int) []uint64 {
	panic("TODO, stubbed")
}

// Sfm_NodeReadFixed
// C: int Sfm_NodeReadFixed( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) SfmNodeReadFixed(i int) int {
	panic("TODO, stubbed")
}

// Sfm_NodeReadUsed
// C: int Sfm_NodeReadUsed( Sfm_Ntk_t * p, int i )
func (p *SfmNtk) SfmNodeReadUsed(i int) int {
	panic("TODO, stubbed")
}
