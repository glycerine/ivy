package abc

// Hand-checked skeleton of src/aig/saig/saigTrans.c.

// Saig_ManStartMap1 (static inline)
// C: void Saig_ManStartMap1( Aig_Man_t * p, int nFrames )
func (p *AigManager) saigManStartMap1(nFrames int) {
	panic("TODO, stubbed")
}

// Saig_ManStopMap1 (static inline)
// C: void Saig_ManStopMap1( Aig_Man_t * p )
func (p *AigManager) saigManStopMap1() {
	panic("TODO, stubbed")
}

// Saig_ManHasMap1 (static inline)
// C: int Saig_ManHasMap1( Aig_Man_t * p )
func (p *AigManager) saigManHasMap1() int {
	panic("TODO, stubbed")
}

// Saig_ManSetMap1 (static inline)
// C: void Saig_ManSetMap1( Aig_Man_t * p, Aig_Obj_t * pOld, int f1, Aig_Obj_t * pNew )
func (p *AigManager) saigManSetMap1(pOld *AigObject, f1 int, pNew *AigObject) {
	panic("TODO, stubbed")
}

// Saig_ManGetMap1 (static inline)
// C: int Saig_ManGetMap1( Aig_Man_t * p, Aig_Obj_t * pOld, int f1 )
func (p *AigManager) saigManGetMap1(pOld *AigObject, f1 int) int {
	panic("TODO, stubbed")
}

// Saig_ManStartMap2 (static inline)
// C: void Saig_ManStartMap2( Aig_Man_t * p, int nFrames )
func (p *AigManager) saigManStartMap2(nFrames int) {
	panic("TODO, stubbed")
}

// Saig_ManStopMap2 (static inline)
// C: void Saig_ManStopMap2( Aig_Man_t * p )
func (p *AigManager) saigManStopMap2() {
	panic("TODO, stubbed")
}

// Saig_ManHasMap2 (static inline)
// C: int Saig_ManHasMap2( Aig_Man_t * p )
func (p *AigManager) saigManHasMap2() int {
	panic("TODO, stubbed")
}

// Saig_ManSetMap2 (static inline)
// C: void Saig_ManSetMap2( Aig_Man_t * p, Aig_Obj_t * pOld, int f1, Aig_Obj_t * pNew, int f2 )
func (p *AigManager) saigManSetMap2(pOld *AigObject, f1 int, pNew *AigObject, f2 int) {
	panic("TODO, stubbed")
}

// Saig_ManGetMap2 (static inline)
// C: int Saig_ManGetMap2( Aig_Man_t * p, Aig_Obj_t * pOld, int f1, int * pf2 )
func (p *AigManager) saigManGetMap2(pOld *AigObject, f1 int, pf2 *int) int {
	panic("TODO, stubbed")
}

// Saig_ManCreateMapping
// C: void Saig_ManCreateMapping( Aig_Man_t * pAig, Aig_Man_t * pFrames, int nFrames )
func (pAig *AigManager) SaigManCreateMapping(pFrames *AigManager, nFrames int) {
	panic("TODO, stubbed")
}

// Saig_ManFramesNonInitial
// C: Aig_Man_t * Saig_ManFramesNonInitial( Aig_Man_t * pAig, int nFrames )
func (pAig *AigManager) SaigManFramesNonInitial(nFrames int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManFramesInitialMapped
// C: Aig_Man_t * Saig_ManFramesInitialMapped( Aig_Man_t * pAig, int nFrames, int nFramesMax, int fInit )
func (pAig *AigManager) SaigManFramesInitialMapped(nFrames int, nFramesMax int, fInit int) *AigManager {
	panic("TODO, stubbed")
}

// Saig_ManTimeframeSimplify
// C: Aig_Man_t * Saig_ManTimeframeSimplify( Aig_Man_t * pAig, int nFrames, int nFramesMax, int fInit, int fVerbose )
func (pAig *AigManager) SaigManTimeframeSimplify(nFrames int, nFramesMax int, fInit int, fVerbose int) *AigManager {
	panic("TODO, stubbed")
}
