package abc

// Hand-checked skeleton of src/map/mapper/mapperCanon.c.

// Map_CanonComputeSlow
// C: int Map_CanonComputeSlow( unsigned uTruths[][2], int nVarsMax, int nVarsReal, unsigned uTruth[], unsigned char * puPhases, unsigned uTruthRes[] )
func MapCanonComputeSlow(uTruths [][2]uint32, nVarsMax int, nVarsReal int, uTruth []uint32, puPhases []uint8, uTruthRes []uint32) int {
	panic("TODO, stubbed")
}

// Map_CanonComputePhase (static)
// C: unsigned Map_CanonComputePhase( unsigned uTruths[][2], int nVars, unsigned uTruth, unsigned uPhase )
func mapCanonComputePhase(uTruths [][2]uint32, nVars int, uTruth uint32, uPhase uint32) uint32 {
	panic("TODO, stubbed")
}

// Map_CanonComputePhase6 (static)
// C: void Map_CanonComputePhase6( unsigned uTruths[][2], int nVars, unsigned uTruth[], unsigned uPhase, unsigned uTruthRes[] )
func mapCanonComputePhase6(uTruths [][2]uint32, nVars int, uTruth []uint32, uPhase uint32, uTruthRes []uint32) {
	panic("TODO, stubbed")
}

// Map_CanonComputeFast
// C: int Map_CanonComputeFast( Map_Man_t * p, int nVarsMax, int nVarsReal, unsigned uTruth[], unsigned char * puPhases, unsigned uTruthRes[] )
func (p *MapMan) MapCanonComputeFast(nVarsMax int, nVarsReal int, uTruth []uint32, puPhases []uint8, uTruthRes []uint32) int {
	panic("TODO, stubbed")
}
