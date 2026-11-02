package abc

// Hand-written skeleton of the types in src/map/cov/cov.h and covInt.h.

// MinCube is Min_Cube_t.
type MinCube struct {
	pNext  *MinCube // the pointer to the next cube in the cover
	nVars  uint32   // unsigned : 10, the number of variables
	nWords uint32   // unsigned : 12, the number of machine words
	nLits  uint32   // unsigned : 10, the number of literals in the cube
	uData  []uint32 // unsigned uData[1] : the bit-data for the cube
}

// MinMan is Min_Man_t.
type MinMan struct {
	nVars   int              // the number of vars
	nWords  int              // the number of words
	pMemMan *ExtraFixedArena // Extra_MmFixed_t * : memory manager for cubes
	// temporary cubes
	pOne0   *MinCube    // tautology cube
	pOne1   *MinCube    // tautology cube
	pTriv0  [2]*MinCube // trivial cube
	pTriv1  [2]*MinCube // trivial cube
	pTemp   *MinCube    // cube for computing the distance
	pBubble *MinCube    // cube used as a separator
	// temporary storage for the new cover
	nCubes  int        // the number of cubes
	ppStore []*MinCube // Min_Cube_t ** : storage for cubes by number of literals
}

// CovObj is Cov_Obj_t: storage for node information.
type CovObj struct {
	pCover [3]*MinCube // pos/neg/esop
	vSupp  *IntVector  // computed support (all nodes except CIs)
}

// CovMan is Cov_Man_t: storage for additional information.
type CovMan struct {
	// general characteristics
	nFaninMax  int                    // the number of vars
	nCubesMax  int                    // the limit on the number of cubes in the intermediate covers
	nWords     int                    // the number of words
	vFanCounts *IntVector             // fanout counts
	vObjStrs   *PointerVector[CovObj] // object structures
	pMemory    []CovObj               // void * : memory for the internal data strctures
	pManMin    *MinMan                // the cube manager
	fUseEsop   int                    // enables ESOPs
	fUseSop    int                    // enables SOPs
	// arrays to map local variables
	vComTo0 *IntVector // mapping of common variables into first fanin
	vComTo1 *IntVector // mapping of common variables into second fanin
	vPairs0 *IntVector // the first var in each pair of common vars
	vPairs1 *IntVector // the second var in each pair of common vars
	vTriv0  *IntVector // trival support of the first node
	vTriv1  *IntVector // trival support of the second node
	// statistics
	nSupps    int // supports created
	nSuppsMax int // the maximum number of supports
	nBoundary int // the boundary size
	nNodes    int // the number of nodes processed
}
