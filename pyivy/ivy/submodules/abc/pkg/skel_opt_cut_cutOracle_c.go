package abc

// Hand-checked skeleton of src/opt/cut/cutOracle.c.

// CutOracle is Cut_Oracle_t (struct Cut_OracleStruct_t_).
type CutOracle struct {
	// cut comptupatation parameters
	pParams    *CutParams
	vFanCounts *IntVector
	fSimul     int
	// storage for cuts
	vCutsNew *PointerVector[CutCut]
	vCuts0   *PointerVector[CutCut]
	vCuts1   *PointerVector[CutCut]
	// oracle info
	vNodeCuts   *IntVector
	vNodeStarts *IntVector
	vCutPairs   *IntVector
	// memory management
	pMmCuts     *ExtraFixedArena // Extra_MmFixed_t *
	EntrySize   int
	nTruthWords int
	// stats
	timeTotal int64 // abctime
	nCuts     int
	nCutsTriv int
}

// Cut_OracleStart
// C: Cut_Oracle_t * Cut_OracleStart( Cut_Man_t * pMan )
func (pMan *CutMan) CutOracleStart() *CutOracle {
	panic("TODO, stubbed")
}

// Cut_OracleStop
// C: void Cut_OracleStop( Cut_Oracle_t * p )
func (p *CutOracle) CutOracleStop() {
	panic("TODO, stubbed")
}

// Cut_OracleSetFanoutCounts
// C: void Cut_OracleSetFanoutCounts( Cut_Oracle_t * p, Vec_Int_t * vFanCounts )
func (p *CutOracle) CutOracleSetFanoutCounts(vFanCounts *IntVector) {
	panic("TODO, stubbed")
}

// Cut_OracleReadDrop
// C: int Cut_OracleReadDrop( Cut_Oracle_t * p )
func (p *CutOracle) CutOracleReadDrop() int {
	panic("TODO, stubbed")
}

// Cut_OracleNodeSetTriv
// C: void Cut_OracleNodeSetTriv( Cut_Oracle_t * p, int Node )
func (p *CutOracle) CutOracleNodeSetTriv(Node int) {
	panic("TODO, stubbed")
}

// Cut_CutStart (static)
// C: Cut_Cut_t * Cut_CutStart( Cut_Oracle_t * p )
func (p *CutOracle) cutCutStart() *CutCut {
	panic("TODO, stubbed")
}

// Cut_CutTriv (static)
// C: Cut_Cut_t * Cut_CutTriv( Cut_Oracle_t * p, int Node )
func (p *CutOracle) cutCutTriv(Node int) *CutCut {
	panic("TODO, stubbed")
}

// Cut_CutMerge (static)
// C: Cut_Cut_t * Cut_CutMerge( Cut_Oracle_t * p, Cut_Cut_t * pCut0, Cut_Cut_t * pCut1 )
func (p *CutOracle) cutCutMerge(pCut0 *CutCut, pCut1 *CutCut) *CutCut {
	panic("TODO, stubbed")
}

// Cut_OracleComputeCuts
// C: Cut_Cut_t * Cut_OracleComputeCuts( Cut_Oracle_t * p, int Node, int Node0, int Node1, int fCompl0, int fCompl1 )
func (p *CutOracle) CutOracleComputeCuts(Node int, Node0 int, Node1 int, fCompl0 int, fCompl1 int) *CutCut {
	panic("TODO, stubbed")
}

// Cut_OracleFreeCuts
// C: void Cut_OracleFreeCuts( Cut_Oracle_t * p, int Node )
func (p *CutOracle) CutOracleFreeCuts(Node int) {
	panic("TODO, stubbed")
}

// Cut_OracleTryDroppingCuts
// C: void Cut_OracleTryDroppingCuts( Cut_Oracle_t * p, int Node )
func (p *CutOracle) CutOracleTryDroppingCuts(Node int) {
	panic("TODO, stubbed")
}
