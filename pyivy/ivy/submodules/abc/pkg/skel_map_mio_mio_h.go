package abc

// Hand-written skeleton of the types in src/map/mio/mio.h. Go has no
// bitfields; each is a uint32 whose C width is in the comment.

// MioPinPhase is the Mio_PinPhase_t enum.
type MioPinPhase int

const (
	MIO_PHASE_UNKNOWN MioPinPhase = iota
	MIO_PHASE_INV
	MIO_PHASE_NONINV
)

// MioCell is Mio_Cell_t.
type MioCell struct {
	pName   string     // char * : name
	Id      uint32     // unsigned : 28, gate ID
	nFanins uint32     // unsigned : 4, gate fanins
	Area    float32    // area
	uTruth  uint64     // word : truth table
	Delays  [6]float32 // delay
}

// MioCell2 is Mio_Cell2_t.
type MioCell2 struct {
	pName     string     // char * : name
	vExpr     *IntVector // expression
	Id        uint32     // unsigned : 26, gate ID
	Type      uint32     // unsigned : 2, gate type
	nFanins   uint32     // unsigned : 4, gate fanins
	AreaF     float32    // area
	AreaW     uint64     // word : area
	uTruth    uint64     // word : truth table
	iDelayAve int        // average delay
	iDelays   [6]int     // delay
	pMioGate  *MioGate   // void * : gate pointer
}
