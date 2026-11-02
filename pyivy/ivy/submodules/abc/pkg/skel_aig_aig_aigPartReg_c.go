package abc

// Hand-checked skeleton of src/aig/aig/aigPartReg.c.

// AigManPre is Aig_ManPre_t.
type AigManPre struct {
	// input data
	pAig     *AigManager               // seq AIG manager
	vMatrix  *PointerVector[IntVector] // register dependency
	nRegsMax int                       // the max number of registers in the cluster
	// information about partitions
	vParts     *PointerVector[IntVector] // the partitions
	pfUsedRegs []int8                    // char * : the registers already included in the partitions
	// info about the current partition
	vRegs      *IntVector   // registers of this partition
	vUniques   *IntVector   // unique registers of this partition
	vFreeVars  *IntVector   // free variables of this partition
	vPartCost  *FloatVector // costs of adding each variable
	pfPartVars []int8       // char * : input/output registers of the partition
}

// Aig_ManRegManStart
// C: Aig_ManPre_t * Aig_ManRegManStart( Aig_Man_t * pAig, int nPartSize )
func (pAig *AigManager) AigManRegManStart(nPartSize int) *AigManPre {
	panic("TODO, stubbed")
}

// Aig_ManRegManStop
// C: void Aig_ManRegManStop( Aig_ManPre_t * p )
func (p *AigManPre) AigManRegManStop() {
	panic("TODO, stubbed")
}

// Aig_ManRegFindSeed
// C: int Aig_ManRegFindSeed( Aig_ManPre_t * p )
func (p *AigManPre) AigManRegFindSeed() int {
	panic("TODO, stubbed")
}

// Aig_ManRegFindBestVar
// C: int Aig_ManRegFindBestVar( Aig_ManPre_t * p )
func (p *AigManPre) AigManRegFindBestVar() int {
	panic("TODO, stubbed")
}

// Aig_ManRegPartitionAdd
// C: void Aig_ManRegPartitionAdd( Aig_ManPre_t * p, int iReg )
func (p *AigManPre) AigManRegPartitionAdd(iReg int) {
	panic("TODO, stubbed")
}

// Aig_ManRegProjectOnehots
// C: Vec_Ptr_t * Aig_ManRegProjectOnehots( Aig_Man_t * pAig, Aig_Man_t * pPart, Vec_Ptr_t * vOnehots, int fVerbose )
func (pAig *AigManager) AigManRegProjectOnehots(pPart *AigManager, vOnehots *PointerVector[IntVector], fVerbose int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManRegCreatePart (pnCountPis, pnCountRegs and ppMapBack are out parameters)
// C: Aig_Man_t * Aig_ManRegCreatePart( Aig_Man_t * pAig, Vec_Int_t * vPart, int * pnCountPis, int * pnCountRegs, int ** ppMapBack )
func (pAig *AigManager) AigManRegCreatePart(vPart *IntVector, pnCountPis *int, pnCountRegs *int, ppMapBack *[]int) *AigManager {
	panic("TODO, stubbed")
}

// Aig_ManRegPartitionSmart
// C: Vec_Ptr_t * Aig_ManRegPartitionSmart( Aig_Man_t * pAig, int nPartSize )
func (pAig *AigManager) AigManRegPartitionSmart(nPartSize int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManRegPartitionSimple
// C: Vec_Ptr_t * Aig_ManRegPartitionSimple( Aig_Man_t * pAig, int nPartSize, int nOverSize )
func (pAig *AigManager) AigManRegPartitionSimple(nPartSize int, nOverSize int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}

// Aig_ManPartDivide
// C: void Aig_ManPartDivide( Vec_Ptr_t * vResult, Vec_Int_t * vDomain, int nPartSize, int nOverSize )
func AigManPartDivide(vResult *PointerVector[IntVector], vDomain *IntVector, nPartSize int, nOverSize int) {
	panic("TODO, stubbed")
}

// Aig_ManRegPartitionTraverse_rec
// C: void Aig_ManRegPartitionTraverse_rec( Aig_Man_t * p, Aig_Obj_t * pObj, Vec_Ptr_t * vLos )
func (p *AigManager) AigManRegPartitionTraverse_rec(pObj *AigObject, vLos *PointerVector[AigObject]) {
	panic("TODO, stubbed")
}

// Aig_ManRegPartitionTraverse (returns the collected LOs)
// C: Vec_Ptr_t * Aig_ManRegPartitionTraverse( Aig_Man_t * p )
func (p *AigManager) AigManRegPartitionTraverse() *PointerVector[AigObject] {
	panic("TODO, stubbed")
}

// Aig_ManRegPartitionLinear
// C: Vec_Ptr_t * Aig_ManRegPartitionLinear( Aig_Man_t * pAig, int nPartSize )
func (pAig *AigManager) AigManRegPartitionLinear(nPartSize int) *PointerVector[IntVector] {
	panic("TODO, stubbed")
}
