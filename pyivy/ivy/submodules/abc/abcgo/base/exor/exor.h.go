// Package exor implements exclusive sum-of-products cover minimization.
package exor

import "fmt"

type VarValue uint8
type varvalue = VarValue

const (
	VAR_NEG VarValue = 1
	VAR_POS VarValue = 2
	VAR_ABS VarValue = 3
)

type CubeDist int
type cubedist = CubeDist

const (
	DIST2 CubeDist = iota
	DIST3
	DIST4
)
const (
	BPI              = 32
	LOGBPI           = 5
	BPIMASK          = 31
	ADDITIONAL_CUBES = 33
	CUBE_PAIR_FACTOR = 20
)

// Cube holds packed two-bit input value sets and one-bit output coefficients.
// XORing input value sets implements literal complementation/absence identities.
type Cube struct {
	Inputs, Outputs []uint32
	ID              uint64
	active          bool
}

func NewCube(inputs []VarValue, outputs []bool) *Cube {
	c := &Cube{Inputs: make([]uint32, (len(inputs)+15)/16), Outputs: make([]uint32, (len(outputs)+31)/32)}
	for i, v := range inputs {
		if v < 1 || v > 3 {
			panic("exor: invalid cube literal")
		}
		c.Inputs[i/16] |= uint32(v) << uint(2*(i%16))
	}
	for i, v := range outputs {
		if v {
			c.Outputs[i/32] |= 1 << uint(i%32)
		}
	}
	return c
}
func (c *Cube) Clone() *Cube {
	return &Cube{Inputs: append([]uint32(nil), c.Inputs...), Outputs: append([]uint32(nil), c.Outputs...)}
}
func (c *Cube) Output(i int) bool { return c.Outputs[i/32]&(1<<uint(i%32)) != 0 }
func (c *Cube) setVar(i int, v VarValue) {
	if v < 1 || v > 3 {
		panic("exor: invalid literal")
	}
	shift := uint(2 * (i % 16))
	c.Inputs[i/16] = c.Inputs[i/16]&^(3<<shift) | uint32(v)<<shift
}

type Options struct {
	Quality, Verbosity, MaxCubes int
	UseQCost                     bool
}
type Statistics struct{ CubesBefore, CubesAfter, LiteralsBefore, LiteralsAfter, QCostBefore, QCostAfter, Rewrites int }
type cubePair struct {
	a, b     *Cube
	distance int
}

// Manager owns its cubes, iterators, pair queues and free storage. Managers can
// be minimized concurrently; one manager is used by one goroutine at a time.
type Manager struct {
	Inputs, Outputs int
	Options         Options
	Stats           Statistics
	Cubes           []*Cube
	free            []*Cube
	nextID          uint64
	iter            int
	pairs           []cubePair
	pairIndex       int
	current         cubePair
	mark            int
	rangeStart      int
	undo            []*Cube
	linkGroups      [][]*Cube
	linkIndex       int
}
type cinfo = Statistics

func NewManager(inputs, outputs int, options Options) (*Manager, error) {
	if inputs < 0 || outputs <= 0 {
		return nil, fmt.Errorf("exor: invalid cover dimensions %d/%d", inputs, outputs)
	}
	if options.Quality < 0 || options.MaxCubes < 0 {
		return nil, fmt.Errorf("exor: negative minimization option")
	}
	return &Manager{Inputs: inputs, Outputs: outputs, Options: options, nextID: 1}, nil
}
func (m *Manager) validate(c *Cube) error {
	if c == nil || len(c.Inputs) != (m.Inputs+15)/16 || len(c.Outputs) != (m.Outputs+31)/32 {
		return fmt.Errorf("exor: cube dimensions do not match manager")
	}
	for i := 0; i < m.Inputs; i++ {
		if GetVar(c, i) == 0 {
			return fmt.Errorf("exor: empty input value set at %d", i)
		}
	}
	if m.Inputs%16 != 0 && len(c.Inputs) > 0 && c.Inputs[len(c.Inputs)-1]>>uint(2*(m.Inputs%16)) != 0 {
		return fmt.Errorf("exor: nonzero input padding")
	}
	if m.Outputs%32 != 0 && c.Outputs[len(c.Outputs)-1]>>uint(m.Outputs%32) != 0 {
		return fmt.Errorf("exor: nonzero output padding")
	}
	return nil
}
func (m *Manager) Add(c *Cube) error {
	if err := m.validate(c); err != nil {
		return err
	}
	if m.Options.MaxCubes > 0 && len(m.Cubes) >= m.Options.MaxCubes {
		return fmt.Errorf("exor: maximum cover size exceeded")
	}
	c = c.Clone()
	c.ID = m.nextID
	m.nextID++
	c.active = true
	m.Cubes = append(m.Cubes, c)
	return nil
}
func cloneCubes(a []*Cube) []*Cube {
	b := make([]*Cube, len(a))
	for i, c := range a {
		b[i] = c.Clone()
		b[i].ID = c.ID
		b[i].active = c.active
	}
	return b
}
