package abc

// Hand-checked skeleton of src/opt/cut/cutPre22.c. CUT_CELL_MVAR is 9.
// Go has no bitfields; each is a uint32 with the C width in the comment.
// Cut_CellCanonicize is only forward-declared in the C source (never
// defined), so it has no stub.

// CutCell is Cut_Cell_t.
type CutCell struct {
	pNext         *CutCell             // pointer to the next cell in the table
	pNextVar      *CutCell             // pointer to the next cell of this support size
	pParent       *CutCell             // pointer to the cell used to derive this one
	nUsed         int                  // the number of times the cell is used
	Box           [4]int8              // char Box[4] : functions in the boxes
	nVars         uint32               // unsigned : 4, the number of variables
	CrossBar0     uint32               // unsigned : 4, the variable set equal
	CrossBar1     uint32               // unsigned : 4, the variable set equal
	CrossBarPhase uint32               // unsigned : 2, the phase of the cross bar (0, 1, or 2)
	CanonPhase    uint32               // unsigned : 18, the canonical phase
	CanonPerm     [9 + 3]int8          // char CanonPerm[CUT_CELL_MVAR+3] : semicanonical permutation
	Store         [2 * 9]int16         // short Store[2*CUT_CELL_MVAR] : minterm counts in the cofactors
	uTruth        [1 << (9 - 5)]uint32 // unsigned uTruth[1<<(CUT_CELL_MVAR-5)] : the current truth table
}

// CutCMan is Cut_CMan_t.
type CutCMan struct {
	// storage for canonical cells
	pMem     *ExtraFixedArena                // Extra_MmFixed_t *
	tTable   *StableTable[uintptr, *CutCell] // st__table * : keyed by the pointer-cast truth-table hash (st__ptrcmp, st__ptrhash), values are cell pointers
	pSameVar [9 + 1]*CutCell                 // Cut_Cell_t * pSameVar[CUT_CELL_MVAR+1]
	// elementary truth tables
	uInputs [9][1 << (9 - 5)]uint32 // unsigned uInputs[CUT_CELL_MVAR][1<<(CUT_CELL_MVAR-5)]
	// temporary truth tables
	uTemp1 [22][1 << (9 - 5)]uint32
	uTemp2 [22][1 << (9 - 5)]uint32
	uTemp3 [22][1 << (9 - 5)]uint32
	uFinal [1 << (9 - 5)]uint32
	puAux  [1 << (9 - 5)]uint32
	// statistical variables
	nTotal        int
	nGood         int
	nVarCounts    [9 + 1]int
	nSymGroups    [9 + 1]int
	nSymGroupsE   [9 + 1]int
	timeCanon     int64 // abctime
	timeSupp      int64 // abctime
	timeTable     int64 // abctime
	nCellFound    int
	nCellNotFound int
}

// Cut_CellLoad
// C: void Cut_CellLoad()
func CutCellLoad() {
	panic("TODO, stubbed")
}

// Cut_CellPrecompute
// C: void Cut_CellPrecompute()
func CutCellPrecompute() {
	panic("TODO, stubbed")
}

// Cut_CellTableLookup (declared static; defined without static)
// C: int Cut_CellTableLookup( Cut_CMan_t * p, Cut_Cell_t * pCell )
func (p *CutCMan) CutCellTableLookup(pCell *CutCell) int {
	panic("TODO, stubbed")
}

// Cut_CellSuppMin (declared static; defined without static)
// C: void Cut_CellSuppMin( Cut_Cell_t * pCell )
func (pCell *CutCell) CutCellSuppMin() {
	panic("TODO, stubbed")
}

// Cut_CellCrossBar (declared static; defined without static)
// C: void Cut_CellCrossBar( Cut_Cell_t * pCell )
func (pCell *CutCell) CutCellCrossBar() {
	panic("TODO, stubbed")
}

// Cut_CellTruthElem (declared static; defined without static)
// C: void Cut_CellTruthElem( unsigned * InA, unsigned * InB, unsigned * InC, unsigned * pOut, int nVars, int Type )
func CutCellTruthElem(InA []uint32, InB []uint32, InC []uint32, pOut []uint32, nVars int, Type int) {
	panic("TODO, stubbed")
}

// Cut_CManStart (declared static; defined without static)
// C: Cut_CMan_t * Cut_CManStart()
func CutCManStart() *CutCMan {
	panic("TODO, stubbed")
}

// Cut_CManStop (declared static; defined without static)
// C: void Cut_CManStop( Cut_CMan_t * p )
func (p *CutCMan) CutCManStop() {
	panic("TODO, stubbed")
}

// Cut_CellIsRunning
// C: int Cut_CellIsRunning()
func CutCellIsRunning() int {
	panic("TODO, stubbed")
}

// Cut_CellDumpToFile
// C: void Cut_CellDumpToFile()
func CutCellDumpToFile() {
	panic("TODO, stubbed")
}

// Cut_CellTruthLookup
// C: int Cut_CellTruthLookup( unsigned * pTruth, int nVars )
func CutCellTruthLookup(pTruth []uint32, nVars int) int {
	panic("TODO, stubbed")
}
