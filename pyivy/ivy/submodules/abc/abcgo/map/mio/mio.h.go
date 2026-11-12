// Public gate and technology-cell representations from mio.h.
package mio

type Mio_PinPhase_t int32

const (
	MIO_PHASE_UNKNOWN Mio_PinPhase_t = iota
	MIO_PHASE_INV
	MIO_PHASE_NONINV
)

type Mio_Cell_t struct {
	pName       string
	Id, nFanins uint32
	Area        float32
	uTruth      uint64
	Delays      [6]float32
}
type Mio_Cell2_t struct {
	pName             string
	vExpr             *Vec_Int_t
	Id, Type, nFanins uint32
	AreaF             float32
	AreaW, uTruth     uint64
	iDelayAve         int32
	iDelays           [6]int32
	pMioGate          *Mio_Gate_t
}
type Mio_Cell_t_ = Mio_Cell_t
type Mio_Cell2_t_ = Mio_Cell2_t

// Technology cell accessors keep mapper consumers independent of storage layout.
func (c *Mio_Cell_t) Name() string            { return c.pName }
func (c *Mio_Cell_t) Fanins() int32           { return int32(c.nFanins) }
func (c *Mio_Cell_t) Truth() uint64           { return c.uTruth }
func (c *Mio_Cell2_t) Name() string           { return c.pName }
func (c *Mio_Cell2_t) Fanins() int32          { return int32(c.nFanins) }
func (c *Mio_Cell2_t) Truth() uint64          { return c.uTruth }
func (c *Mio_Cell2_t) Gate() *Mio_Gate_t      { return c.pMioGate }
func (c *Mio_Cell2_t) Expression() *Vec_Int_t { return c.vExpr }
