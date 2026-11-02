package abc

// Hand-written skeleton of the types in src/map/mio/mioInt.h.

// MioLibrary is Mio_Library_t (struct Mio_LibraryStruct_t_).
type MioLibrary struct {
	pName       string                         // char * : the name of the library
	nGates      int                            // the number of the gates
	ppGates0    []*MioGate                     // Mio_Gate_t ** : the array of gates in the original order
	ppGatesName []*MioGate                     // Mio_Gate_t ** : the array of gates sorted by name
	pGates      *MioGate                       // the linked list of all gates in no particular order
	pGate0      *MioGate                       // the constant zero gate
	pGate1      *MioGate                       // the constant one gate
	pGateBuf    *MioGate                       // the buffer
	pGateInv    *MioGate                       // the inverter
	pGateNand2  *MioGate                       // the NAND2 gate
	pGateAnd2   *MioGate                       // the AND2 gate
	pGateNor2   *MioGate                       // the NOR2 gate
	pGateOr2    *MioGate                       // the OR2 gate
	tName2Gate  *StableTable[string, *MioGate] // st__table * : the mapping of gate names into their pointer
	pMmFlex     *ExtraFlexArena                // Mem_Flex_t * : the memory manaqer for SOPs
	vCube       *ByteVector                    // temporary cube
	// matching
	fPinFilter int             // pin filtering
	fPinPerm   int             // pin permutation
	fPinQuick  int             // pin permutation
	vTtMem     *WordMemory     // truth tables
	vTt2Match  *IntLevelVector // matches for truth tables
	pCells     []MioCell2      // Mio_Cell2_t * : library gates (an array of nCells)
	nCells     int             // library gate count
	vNames     *PointerVector[string]
	vTruths    *WordVector
	vTt2Match4 *IntVector
	vConfigs   *IntVector
	vTtMem2    [3]*WordMemory
	vTt2Match2 [3]*IntVector
}

// MioGate is Mio_Gate_t (struct Mio_GateStruct_t_). The anonymous union
// { word uTruth; word * pTruth; } is represented by both members.
type MioGate struct {
	// information derived from the genlib file
	pName    string  // char * : the name of the gate
	dArea    float64 // the area of the gate
	pForm    string  // char * : the formula describing functionality of the gate
	pPins    *MioPin // the linked list of all pins (one pin if info is the same)
	pOutName string  // char * : the name of the output pin
	// the library to which this gate belongs
	pLib *MioLibrary
	// the next gate in the list
	pNext *MioGate
	pTwin *MioGate
	// the derived information
	Cell      int        // cell id
	nInputs   int        // the number of inputs
	Profile   int        // the number of occurrences
	Profile2  int        // the number of occurrences
	dDelayMax float64    // the maximum delay
	pSop      string     // char * : sum-of-products
	vExpr     *IntVector // boolean expression
	uTruth    uint64     // union member (word) : truth table
	pTruth    []uint64   // union member (word *) : pointer to the truth table
	Value     int        // user's information
}

// MioPin is Mio_Pin_t (struct Mio_PinStruct_t_).
type MioPin struct {
	pName            string      // char *
	Phase            MioPinPhase // Mio_PinPhase_t
	dLoadInput       float64
	dLoadMax         float64
	dDelayBlockRise  float64
	dDelayFanoutRise float64
	dDelayBlockFall  float64
	dDelayFanoutFall float64
	dDelayBlockMax   float64
	pNext            *MioPin
}
