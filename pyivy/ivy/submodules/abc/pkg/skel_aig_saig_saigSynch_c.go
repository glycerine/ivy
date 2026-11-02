package abc

// Hand-checked skeleton of src/aig/saig/saigSynch.c.

// Saig_SynchNot (static inline)
// C: unsigned Saig_SynchNot( unsigned w )
func saigSynchNot(w uint32) uint32 {
	panic("TODO, stubbed")
}

// Saig_SynchAnd (static inline)
// C: unsigned Saig_SynchAnd( unsigned u, unsigned w )
func saigSynchAnd(u uint32, w uint32) uint32 {
	panic("TODO, stubbed")
}

// Saig_SynchRandomBinary (static inline)
// C: unsigned Saig_SynchRandomBinary()
func saigSynchRandomBinary() uint32 {
	panic("TODO, stubbed")
}

// Saig_SynchRandomTernary (static inline)
// C: unsigned Saig_SynchRandomTernary()
func saigSynchRandomTernary() uint32 {
	panic("TODO, stubbed")
}

// Saig_SynchTernary (static inline)
// C: unsigned Saig_SynchTernary( int v )
func saigSynchTernary(v int) uint32 {
	panic("TODO, stubbed")
}

// Saig_SynchSetConstant1
// C: void Saig_SynchSetConstant1( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords )
func (pAig *AigManager) SaigSynchSetConstant1(vSimInfo *PointerVector[uint32], nWords int) {
	panic("TODO, stubbed")
}

// Saig_SynchInitRegsTernary
// C: void Saig_SynchInitRegsTernary( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords )
func (pAig *AigManager) SaigSynchInitRegsTernary(vSimInfo *PointerVector[uint32], nWords int) {
	panic("TODO, stubbed")
}

// Saig_SynchInitRegsBinary
// C: void Saig_SynchInitRegsBinary( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords )
func (pAig *AigManager) SaigSynchInitRegsBinary(vSimInfo *PointerVector[uint32], nWords int) {
	panic("TODO, stubbed")
}

// Saig_SynchInitPisRandom
// C: void Saig_SynchInitPisRandom( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords )
func (pAig *AigManager) SaigSynchInitPisRandom(vSimInfo *PointerVector[uint32], nWords int) {
	panic("TODO, stubbed")
}

// Saig_SynchInitPisGiven
// C: void Saig_SynchInitPisGiven( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords, char * pValues )
func (pAig *AigManager) SaigSynchInitPisGiven(vSimInfo *PointerVector[uint32], nWords int, pValues []int8) {
	panic("TODO, stubbed")
}

// Saig_SynchTernarySimulate
// C: void Saig_SynchTernarySimulate( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords )
func (pAig *AigManager) SaigSynchTernarySimulate(vSimInfo *PointerVector[uint32], nWords int) {
	panic("TODO, stubbed")
}

// Saig_SynchTernaryTransferState
// C: void Saig_SynchTernaryTransferState( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords )
func (pAig *AigManager) SaigSynchTernaryTransferState(vSimInfo *PointerVector[uint32], nWords int) {
	panic("TODO, stubbed")
}

// Saig_SynchCountX
// C: int Saig_SynchCountX( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords, int * piPat )
func (pAig *AigManager) SaigSynchCountX(vSimInfo *PointerVector[uint32], nWords int, piPat *int) int {
	panic("TODO, stubbed")
}

// Saig_SynchSavePattern
// C: int Saig_SynchSavePattern( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, int nWords, int iPat, Vec_Str_t * vSequence )
func (pAig *AigManager) SaigSynchSavePattern(vSimInfo *PointerVector[uint32], nWords int, iPat int, vSequence *ByteVector) int {
	panic("TODO, stubbed")
}

// Saig_SynchSequenceRun
// C: int Saig_SynchSequenceRun( Aig_Man_t * pAig, Vec_Ptr_t * vSimInfo, Vec_Str_t * vSequence, int fTernary )
func (pAig *AigManager) SaigSynchSequenceRun(vSimInfo *PointerVector[uint32], vSequence *ByteVector, fTernary int) int {
	panic("TODO, stubbed")
}

// Saig_SynchSequence
// C: Vec_Str_t * Saig_SynchSequence( Aig_Man_t * pAig, int nWords )
func (pAig *AigManager) SaigSynchSequence(nWords int) *ByteVector {
	panic("TODO, stubbed")
}

// Saig_ManDupInitZero
// C: Aig_Man_t * Saig_ManDupInitZero( Aig_Man_t * p )
func (p *AigManager) SaigManDupInitZero() *AigManager {
	panic("TODO, stubbed")
}

// Saig_SynchSequenceApply
// C: Aig_Man_t * Saig_SynchSequenceApply( Aig_Man_t * pAig, int nWords, int fVerbose )
func (pAig *AigManager) SaigSynchSequenceApply(nWords int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_Synchronize
// C: Aig_Man_t * Saig_Synchronize( Aig_Man_t * pAig1, Aig_Man_t * pAig2, int nWords, int fVerbose )
func (pAig1 *AigManager) SaigSynchronize(pAig2 *AigManager, nWords int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
