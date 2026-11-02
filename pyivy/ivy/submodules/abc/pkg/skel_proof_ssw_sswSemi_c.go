package abc

// Hand-checked skeleton of src/proof/ssw/sswSemi.c.

// SswSem is Ssw_Sem_t: BMC manager.
type SswSem struct {
	// parameters
	nConfMaxStart int // the starting conflict limit
	nConfMax      int // the intermediate conflict limit
	nFramesSweep  int // the number of frames to sweep
	fVerbose      int // prints output statistics
	// equivalences considered
	pMan     *SswMan                   // SAT sweeping manager
	vTargets *PointerVector[AigObject] // the nodes that are watched
	// storage for patterns
	nPatternsAlloc int                    // the max number of interesting states
	nPatterns      int                    // the number of patterns
	vPatterns      *PointerVector[uint32] // storage for the interesting states (unsigned * sim-info rows)
	vHistory       *IntVector             // what state and how many steps
}

// Ssw_SemManStart
// C: Ssw_Sem_t * Ssw_SemManStart( Ssw_Man_t * pMan, int nConfMax, int fVerbose )
func (pMan *SswMan) SswSemManStart(nConfMax int, fVerbose int) *SswSem {
	panic("TODO, stubbed")
}

// Ssw_SemManStop
// C: void Ssw_SemManStop( Ssw_Sem_t * p )
func (p *SswSem) SswSemManStop() {
	panic("TODO, stubbed")
}

// Ssw_SemCheckTargets
// C: int Ssw_SemCheckTargets( Ssw_Sem_t * p )
func (p *SswSem) SswSemCheckTargets() int {
	panic("TODO, stubbed")
}

// Ssw_ManFilterBmcSavePattern
// C: void Ssw_ManFilterBmcSavePattern( Ssw_Sem_t * p )
func (p *SswSem) SswManFilterBmcSavePattern() {
	panic("TODO, stubbed")
}

// Ssw_ManFilterBmc
// C: int Ssw_ManFilterBmc( Ssw_Sem_t * pBmc, int iPat, int fCheckTargets )
func (pBmc *SswSem) SswManFilterBmc(iPat int, fCheckTargets int) int {
	panic("TODO, stubbed")
}

// Ssw_FilterUsingSemi
// C: int Ssw_FilterUsingSemi( Ssw_Man_t * pMan, int fCheckTargets, int nConfMax, int fVerbose )
func (pMan *SswMan) SswFilterUsingSemi(fCheckTargets int, nConfMax int, fVerbose int) int {
	panic("TODO, stubbed")
}
