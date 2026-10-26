package abc

// Port of src/proof/pdr/pdrUtil.c: Pdr_Set* cube operations. Allocation-only
// reference counts are replaced by immutable owned slices and Go lifetimes.

import (
	"fmt"
	"math"
	"slices"
)

// PdrLit names a register or PI in its cube's own namespace, not a SAT
// variable. The low bit is polarity: 2*i means bit i is one, 2*i+1 means zero.
type PdrLit uint32

const InvalidPdrLit = PdrLit(math.MaxUint32)

func MakePdrLit(index uint32, negative bool) (PdrLit, error) {
	if index > MaxLiteralID {
		return InvalidPdrLit, ErrInvalidLiteral
	}
	l := PdrLit(index) << 1
	if negative {
		l |= 1
	}
	return l, nil
}
func (l PdrLit) Valid() bool    { return l <= math.MaxInt32 }
func (l PdrLit) Index() uint32  { return uint32(l >> 1) }
func (l PdrLit) Negative() bool { return l&1 != 0 }
func (l PdrLit) Not() PdrLit    { return l ^ 1 }

// PdrCube is a conjunction of sorted register literals with separately owned
// PI assignments retained in caller order. Its signature is C's modulo-63
// filter; a matching signature alone never establishes containment.
// Cubes are immutable and can be shared among proof obligations.
type PdrCube struct {
	signature     uint64
	state, inputs []PdrLit
}

func pdrNormalize(lits []PdrLit, sort bool) ([]PdrLit, error) {
	result := make([]PdrLit, 0, len(lits))
	seen := make(map[uint32]PdrLit, len(lits))
	for _, l := range lits {
		if !l.Valid() {
			return nil, ErrInvalidLiteral
		}
		if old, yes := seen[l.Index()]; yes {
			if old != l {
				return nil, fmt.Errorf("%w: contradictory cube literals", ErrDimension)
			}
			continue
		}
		seen[l.Index()] = l
		result = append(result, l)
	}
	if sort {
		slices.Sort(result)
	}
	return result, nil
}
func NewPdrCube(state, inputs []PdrLit) (*PdrCube, error) {
	if _, err := checkedSize(len(state), len(inputs), 1, (1<<30)-1); err != nil {
		return nil, err
	}
	normalized, err := pdrNormalize(state, true)
	if err != nil {
		return nil, err
	}
	pis, err := pdrNormalize(inputs, false)
	if err != nil {
		return nil, err
	}
	cube := &PdrCube{state: normalized, inputs: pis}
	cube.setSignature()
	return cube, nil
}
func (c *PdrCube) setSignature() {
	c.signature = 0
	for _, l := range c.state {
		c.signature |= uint64(1) << uint(l%63)
	}
}
func (c *PdrCube) Signature() uint64       { return c.signature }
func (c *PdrCube) StateCount() int         { return len(c.state) }
func (c *PdrCube) LiteralCount() int       { return len(c.state) + len(c.inputs) }
func (c *PdrCube) StateLiterals() []PdrLit { return slices.Clone(c.state) }
func (c *PdrCube) InputLiterals() []PdrLit { return slices.Clone(c.inputs) }
func (c *PdrCube) Clone() *PdrCube {
	return &PdrCube{signature: c.signature, state: slices.Clone(c.state), inputs: slices.Clone(c.inputs)}
}
func (c *PdrCube) Check() error {
	if c == nil {
		return ErrDimension
	}
	var signature uint64
	for i, l := range c.state {
		if !l.Valid() || (i > 0 && c.state[i-1].Index() >= l.Index()) {
			return ErrInvariant
		}
		signature |= uint64(1) << uint(l%63)
	}
	if signature != c.signature {
		return ErrInvariant
	}
	seen := make(map[uint32]bool, len(c.inputs))
	for _, l := range c.inputs {
		if !l.Valid() || seen[l.Index()] {
			return ErrInvariant
		}
		seen[l.Index()] = true
	}
	return nil
}

// Without implements Pdr_SetCreateFrom, preserving the PI tail.
func (c *PdrCube) Without(index int) (*PdrCube, error) {
	if c == nil || index < 0 || index >= len(c.state) {
		return nil, ErrDimension
	}
	state := make([]PdrLit, 0, len(c.state)-1)
	state = append(state, c.state[:index]...)
	state = append(state, c.state[index+1:]...)
	result := &PdrCube{state: state, inputs: slices.Clone(c.inputs)}
	result.setSignature()
	return result, nil
}

// Subset implements Pdr_SetCreateSubset and additionally verifies that the
// supplied literals actually belong to this cube. Duplicate equal literals
// are canonicalized; contradictory assignments are rejected.
func (c *PdrCube) Subset(state []PdrLit) (*PdrCube, error) {
	if c == nil {
		return nil, ErrDimension
	}
	subset, err := NewPdrCube(state, c.inputs)
	if err != nil {
		return nil, err
	}
	if !c.Contains(subset) {
		return nil, ErrDimension
	}
	return subset, nil
}

// Contains follows Pdr_SetContains: receiver's register-literal set is a
// superset of other's. PI assignments do not participate. Empty sets are
// handled mathematically instead of asserting as the C helper does.
func (c *PdrCube) Contains(other *PdrCube) bool {
	if c == nil || other == nil || len(c.state) < len(other.state) || c.signature&other.signature != other.signature {
		return false
	}
	i, j := 0, 0
	for i < len(c.state) && j < len(other.state) {
		if c.state[i] < other.state[j] {
			i++
		} else if c.state[i] == other.state[j] {
			i++
			j++
		} else {
			return false
		}
	}
	return j == len(other.state)
}

// IntersectsZero implements Pdr_SetIsInit; remove=-1 retains every literal.
func (c *PdrCube) IntersectsZero(remove int) (bool, error) {
	if c == nil || remove < -1 || remove >= len(c.state) {
		return false, ErrDimension
	}
	for i, l := range c.state {
		if i != remove && !l.Negative() {
			return false, nil
		}
	}
	return true, nil
}

// IntersectsInitial extends the zero-init test to known one/unknown reset bits.
func (c *PdrCube) IntersectsInitial(initial []LatchInit) (bool, error) {
	if c == nil {
		return false, ErrDimension
	}
	for _, v := range initial {
		if v > LatchInitUnknown {
			return false, ErrDimension
		}
	}
	for _, l := range c.state {
		if uint64(l.Index()) >= uint64(len(initial)) {
			return false, ErrDimension
		}
	}
	for _, l := range c.state {
		init := initial[l.Index()]
		if init != LatchInitUnknown && (init == LatchInitOne) == l.Negative() {
			return false, nil
		}
	}
	return true, nil
}
func (c *PdrCube) Matches(state []bool) (bool, error) {
	if c == nil {
		return false, ErrDimension
	}
	for _, l := range c.state {
		if uint64(l.Index()) >= uint64(len(state)) {
			return false, ErrDimension
		}
	}
	for _, l := range c.state {
		if state[l.Index()] == l.Negative() {
			return false, nil
		}
	}
	return true, nil
}

// Compare follows Pdr_SetCompare: larger differing literals sort first, with
// the shorter cube first when one is a prefix. Nil sorts before a real cube.
func (c *PdrCube) Compare(other *PdrCube) int {
	if c == nil {
		if other == nil {
			return 0
		}
		return -1
	}
	if other == nil {
		return 1
	}
	for i := 0; i < min(len(c.state), len(other.state)); i++ {
		if c.state[i] > other.state[i] {
			return -1
		}
		if c.state[i] < other.state[i] {
			return 1
		}
	}
	if len(c.state) < len(other.state) {
		return -1
	}
	if len(c.state) > len(other.state) {
		return 1
	}
	return 0
}

// Intersection ports ZPdr_SetIntersection. PI assignments are discarded.
// A protected literal present in either cube must remain in the intersection;
// this check includes tails that the surveyed C loop can fail to inspect.
func (c *PdrCube) Intersection(other *PdrCube, protected []PdrLit) (*PdrCube, bool, error) {
	if c == nil || other == nil {
		return nil, false, ErrDimension
	}
	for _, l := range protected {
		if !l.Valid() {
			return nil, false, ErrInvalidLiteral
		}
		_, a := slices.BinarySearch(c.state, l)
		_, b := slices.BinarySearch(other.state, l)
		if a != b {
			return nil, false, nil
		}
	}
	common := make([]PdrLit, 0, min(len(c.state), len(other.state)))
	for i, j := 0, 0; i < len(c.state) && j < len(other.state); {
		if c.state[i] < other.state[j] {
			i++
		} else if c.state[i] > other.state[j] {
			j++
		} else {
			common = append(common, c.state[i])
			i++
			j++
		}
	}
	cube, err := NewPdrCube(common, nil)
	return cube, err == nil, err
}

// StateString implements the full-register display of Pdr_SetPrint.
func (c *PdrCube) StateString(registers int) (string, error) {
	if c == nil || registers < 0 {
		return "", ErrDimension
	}
	for _, l := range c.state {
		if uint64(l.Index()) >= uint64(registers) {
			return "", ErrDimension
		}
	}
	result := make([]byte, registers)
	for i := range result {
		result[i] = '-'
	}
	for _, l := range c.state {
		value := byte('1')
		if l.Negative() {
			value = '0'
		}
		result[l.Index()] = value
	}
	return string(result), nil
}
