package abc

// Hand-checked skeleton of src/map/mpm/mpmDsd.c.

// The static data table s_DsdClass6[595] is not represented.

// Mpm_ManGetTruthWithCnf
// C: Vec_Wrd_t * Mpm_ManGetTruthWithCnf( int Limit )
func MpmManGetTruthWithCnf(Limit int) *WordVector {
	panic("TODO, stubbed")
}

// Mpm_ManPrintDsdStats
// C: void Mpm_ManPrintDsdStats( Mpm_Man_t * p )
func (p *MpmMan) MpmManPrintDsdStats() {
	panic("TODO, stubbed")
}

// Mpm_ManBuildHashTable (Hsh_IntMan_t *; pvConfgRes is an out parameter)
// C: Hsh_IntMan_t * Mpm_ManBuildHashTable( Vec_Int_t ** pvConfgRes )
func MpmManBuildHashTable(pvConfgRes **IntVector) *FixedIntRecordInterner {
	panic("TODO, stubbed")
}

// Mpm_ManPrintPerm
// C: void Mpm_ManPrintPerm( unsigned s )
func MpmManPrintPerm(s uint32) {
	panic("TODO, stubbed")
}

// Mpm_ManPrecomputePerms
// C: void Mpm_ManPrecomputePerms( Mpm_Man_t * p )
func (p *MpmMan) MpmManPrecomputePerms() {
	panic("TODO, stubbed")
}

// Mpm_CutTruthFromDsd
// C: word Mpm_CutTruthFromDsd( Mpm_Man_t * pMan, Mpm_Cut_t * pCut, int iClass )
func (pMan *MpmMan) MpmCutTruthFromDsd(pCut *MpmCut, iClass int) uint64 {
	panic("TODO, stubbed")
}

// Mpm_CutCheckDsd6
// C: int Mpm_CutCheckDsd6( Mpm_Man_t * p, word t )
func (p *MpmMan) MpmCutCheckDsd6(t uint64) int {
	panic("TODO, stubbed")
}

// Mpm_CutComputeDsd6
// C: int Mpm_CutComputeDsd6( Mpm_Man_t * p, Mpm_Cut_t * pCut, Mpm_Cut_t * pCut0, Mpm_Cut_t * pCut1, Mpm_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, int Type )
func (p *MpmMan) MpmCutComputeDsd6(pCut *MpmCut, pCut0 *MpmCut, pCut1 *MpmCut, pCutC *MpmCut, fCompl0 int, fCompl1 int, fComplC int, Type int) int {
	panic("TODO, stubbed")
}
