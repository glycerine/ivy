package abc

// Manager ownership from src/bool/rsb/rsbMan.c.

import "slices"

const (
	RsbMaxLeaves   = 20
	RsbMaxDivisors = 200
)

type RsbManagerConfig struct {
	LeafMaximum, DivisorMaximum, DecompositionMaximum int
	Verbose                                           bool
}

type RsbManagerStats struct {
	Config                 RsbManagerConfig
	CounterexampleCapacity int
	PatternWordCapacity    int
	Fanins, OriginalFanins int
	IntermediateTryEntries int
}

// RsbManager owns the resubstitution decomposition scratch. It is concrete and
// noncopyable; later RSB algorithms in the same package can use its typed
// storage directly without an interface or a C-style untyped arena.
type RsbManager struct {
	config          RsbManagerConfig
	counterexamples []uint64
	patterns        []uint32
	fanins          []int
	originalFanins  []int
	tries           []int
	closed          bool
}

func NewRsbManager(config RsbManagerConfig) (*RsbManager, error) {
	if config.LeafMaximum < 0 || config.LeafMaximum > RsbMaxLeaves ||
		config.DivisorMaximum < 0 || config.DivisorMaximum > RsbMaxDivisors ||
		config.DecompositionMaximum < 0 {
		return nil, ErrDimension
	}
	words := 1
	if config.LeafMaximum > 6 {
		words = 1 << uint(config.LeafMaximum-6)
	}
	return &RsbManager{
		config:          config,
		counterexamples: make([]uint64, 0, config.DivisorMaximum+150),
		patterns:        make([]uint32, 0, words),
		fanins:          make([]int, 0, 10),
		originalFanins:  make([]int, 0, 10),
		tries:           make([]int, 0, 10),
	}, nil
}

func (m *RsbManager) Config() (RsbManagerConfig, error) {
	if m == nil || m.closed {
		return RsbManagerConfig{}, ErrInvariant
	}
	return m.config, nil
}

// Fanins and OriginalFanins return owned snapshots. The source accessors expose
// mutable vectors; explicit replacement below preserves that use without
// allowing callers to alias manager storage.
func (m *RsbManager) Fanins() []int {
	if m == nil || m.closed {
		return nil
	}
	return slices.Clone(m.fanins)
}

func (m *RsbManager) OriginalFanins() []int {
	if m == nil || m.closed {
		return nil
	}
	return slices.Clone(m.originalFanins)
}

func (m *RsbManager) replaceFanins(values []int, original bool) error {
	if m == nil || m.closed {
		return ErrInvariant
	}
	if len(values) > m.config.DivisorMaximum {
		return ErrLimit
	}
	for _, value := range values {
		if value < 0 || value >= m.config.DivisorMaximum {
			return ErrDimension
		}
	}
	copyValues := slices.Clone(values)
	if original {
		m.originalFanins = copyValues
	} else {
		m.fanins = copyValues
	}
	return nil
}

func (m *RsbManager) SetFanins(values []int) error {
	return m.replaceFanins(values, false)
}

func (m *RsbManager) SetOriginalFanins(values []int) error {
	return m.replaceFanins(values, true)
}

func (m *RsbManager) Statistics() RsbManagerStats {
	if m == nil || m.closed {
		return RsbManagerStats{}
	}
	return RsbManagerStats{
		Config:                 m.config,
		CounterexampleCapacity: cap(m.counterexamples),
		PatternWordCapacity:    cap(m.patterns),
		Fanins:                 len(m.fanins),
		OriginalFanins:         len(m.originalFanins),
		IntermediateTryEntries: len(m.tries),
	}
}

func (m *RsbManager) Close() {
	if m == nil || m.closed {
		return
	}
	m.counterexamples = nil
	m.patterns = nil
	m.fanins = nil
	m.originalFanins = nil
	m.tries = nil
	m.closed = true
}
