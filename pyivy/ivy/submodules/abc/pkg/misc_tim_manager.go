package abc

// Timing terminal storage/traversal from misc/tim/timMan.c, timTime.c,
// timTrav.c and timInt.h. Box propagation is a separate pending layer.

import "slices"

const TimEternity float32 = 1000000000
const timTraversalLimit = (1 << 30) - 1

type TimTerminal struct {
	ID, TravID, Box, Position int
	Arrival, Required         float32
}

// TimManager owns float32 terminal timing and native traversal epochs. This
// box associations distinguish main PIs/POs from box terminals. It is single-owner;
// getters update traversal state and must not be called concurrently.
type TimManager struct {
	cis, cos     []TimTerminal
	boxes        []TimBox
	boxPins      int
	tables       []TimDelayTable
	tableValues  int
	epoch        int
	useTraversal bool
}

func NewTimManager(cis, cos, maxTerminals int) (*TimManager, error) {
	if cis < 0 || cos < 0 || maxTerminals < 0 {
		return nil, ErrDimension
	}
	if maxTerminals == 0 {
		maxTerminals = 1 << 20
	}
	if _, err := checkedSize(cis, 1, cos, min(maxTerminals, int(^uint32(0)>>1))); err != nil {
		return nil, err
	}
	m := &TimManager{cis: make([]TimTerminal, cis), cos: make([]TimTerminal, cos), useTraversal: true}
	for i := range m.cis {
		m.cis[i] = TimTerminal{ID: i, Box: -1, Position: -1, Required: TimEternity}
	}
	for i := range m.cos {
		m.cos[i] = TimTerminal{ID: i, Box: -1, Position: -1, Required: TimEternity}
	}
	return m, nil
}
func (m *TimManager) Clone() *TimManager {
	if m == nil {
		return nil
	}
	copy := *m
	copy.cis = slices.Clone(m.cis)
	copy.cos = slices.Clone(m.cos)
	copy.boxes = slices.Clone(m.boxes)
	copy.tables = slices.Clone(m.tables)
	for i := range copy.tables {
		copy.tables[i].Delays = slices.Clone(m.tables[i].Delays)
	}
	return &copy
}
func (m *TimManager) Counts() (int, int) {
	if m == nil {
		return 0, 0
	}
	return len(m.cis), len(m.cos)
}
func (m *TimManager) TraversalID() int {
	if m == nil {
		return 0
	}
	return m.epoch
}
func (m *TimManager) SetTraversalEnabled(enabled bool) error {
	if m == nil {
		return ErrDimension
	}
	m.useTraversal = enabled
	return nil
}
func (m *TimManager) IncrementTraversal() error {
	if m == nil {
		return ErrDimension
	}
	if m.epoch >= timTraversalLimit {
		m.epoch = 0
		for i := range m.cis {
			m.cis[i].TravID = 0
		}
		for i := range m.cos {
			m.cos[i].TravID = 0
		}
	}
	m.epoch++
	return nil
}
func (m *TimManager) terminal(ci bool, index int) (*TimTerminal, error) {
	if m == nil || index < 0 {
		return nil, ErrDimension
	}
	objects := m.cos
	if ci {
		objects = m.cis
	}
	if index >= len(objects) {
		return nil, ErrDimension
	}
	return &objects[index], nil
}
func (m *TimManager) CI(index int) (TimTerminal, error) {
	p, err := m.terminal(true, index)
	if err != nil {
		return TimTerminal{}, err
	}
	return *p, nil
}
func (m *TimManager) CO(index int) (TimTerminal, error) {
	p, err := m.terminal(false, index)
	if err != nil {
		return TimTerminal{}, err
	}
	return *p, nil
}
func (m *TimManager) IsCurrent(ci bool, index int) (bool, error) {
	p, err := m.terminal(ci, index)
	if err != nil {
		return false, err
	}
	if !m.useTraversal {
		return false, ErrInvariant
	}
	return p.TravID == m.epoch, nil
}
func (m *TimManager) InitPIArrival(index int, value float32) error {
	p, err := m.terminal(true, index)
	if err != nil {
		return err
	}
	p.Arrival = value
	return nil
}
func (m *TimManager) InitPORequired(index int, value float32) error {
	p, err := m.terminal(false, index)
	if err != nil {
		return err
	}
	p.Required = value
	return nil
}
func (m *TimManager) InitPIArrivalAll(value float32) error {
	if m == nil {
		return ErrDimension
	}
	for i := range m.cis {
		if m.cis[i].Box < 0 {
			m.cis[i].Arrival = value
		}
	}
	return nil
}

// InitPORequiredAll follows C's SetCoRequired-based bulk initializer. Go
// preflights every traversal assertion so failure leaves all terminals intact.
func (m *TimManager) InitPORequiredAll(value float32) error {
	if m == nil {
		return ErrDimension
	}
	for _, p := range m.cos {
		if p.Box < 0 && m.useTraversal && m.epoch != 0 && p.TravID == m.epoch {
			return ErrInvariant
		}
	}
	for i := range m.cos {
		if m.cos[i].Box < 0 {
			m.cos[i].Required = value
			m.cos[i].TravID = m.epoch
		}
	}
	return nil
}
func (m *TimManager) SetCOArrival(index int, value float32) error {
	p, err := m.terminal(false, index)
	if err != nil {
		return err
	}
	if m.useTraversal && p.TravID == m.epoch {
		return ErrInvariant
	}
	p.Arrival = value
	p.TravID = m.epoch
	return nil
}
func (m *TimManager) SetCIRequired(index int, value float32) error {
	p, err := m.terminal(true, index)
	if err != nil {
		return err
	}
	if m.useTraversal && p.TravID == m.epoch {
		return ErrInvariant
	}
	p.Required = value
	p.TravID = m.epoch
	return nil
}
func (m *TimManager) SetCORequired(index int, value float32) error {
	p, err := m.terminal(false, index)
	if err != nil {
		return err
	}
	if m.useTraversal && m.epoch != 0 && p.TravID == m.epoch {
		return ErrInvariant
	}
	p.Required = value
	p.TravID = m.epoch
	return nil
}
func (m *TimManager) CIArrival(index int) (float32, error) {
	p, err := m.terminal(true, index)
	if err != nil {
		return 0, err
	}
	if m.useTraversal && p.TravID == m.epoch {
		return p.Arrival, nil
	}
	if p.Box < 0 {
		p.TravID = m.epoch
		return p.Arrival, nil
	}
	return m.boxArrival(index)
}

// CORequired returns cached or propagated required time. Box propagation
// guards the actual input arc, correcting C's output-indexed disabled-arc bug.
func (m *TimManager) CORequired(index int) (float32, error) {
	p, err := m.terminal(false, index)
	if err != nil {
		return 0, err
	}
	if m.useTraversal && p.TravID == m.epoch {
		return p.Required, nil
	}
	if p.Box < 0 {
		p.TravID = m.epoch
		return p.Required, nil
	}
	return m.boxRequired(index)
}
