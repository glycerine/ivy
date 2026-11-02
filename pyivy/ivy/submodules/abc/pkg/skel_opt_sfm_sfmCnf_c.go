package abc

// Hand-checked skeleton of src/opt/sfm/sfmCnf.c.

// Sfm_PrintCnf
// C: void Sfm_PrintCnf( Vec_Str_t * vCnf )
func SfmPrintCnf(vCnf *ByteVector) {
	panic("TODO, stubbed")
}

// Sfm_TruthToCnf
// C: int Sfm_TruthToCnf( word Truth, word * pTruth, int nVars, Vec_Int_t * vCover, Vec_Str_t * vCnf )
func SfmTruthToCnf(Truth uint64, pTruth []uint64, nVars int, vCover *IntVector, vCnf *ByteVector) int {
	panic("TODO, stubbed")
}

// Sfm_CreateCnf
// C: Vec_Wec_t * Sfm_CreateCnf( Sfm_Ntk_t * p )
func (p *SfmNtk) SfmCreateCnf() *IntLevelVector {
	panic("TODO, stubbed")
}

// Sfm_TranslateCnf
// C: void Sfm_TranslateCnf( Vec_Wec_t * vRes, Vec_Str_t * vCnf, Vec_Int_t * vFaninMap, int iPivotVar )
func SfmTranslateCnf(vRes *IntLevelVector, vCnf *ByteVector, vFaninMap *IntVector, iPivotVar int) {
	panic("TODO, stubbed")
}
