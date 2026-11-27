// Package ioa reads and writes AIGER circuits using the Go standard library.
// The binary codecs originate in ABC's IOA package (Alan Mishchenko).
package ioa

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Literals use the AIGER convention: 0=false, 1=true, even literals name
// variables and the low bit complements their value.
type Latch struct{ Literal, Next, Init uint32 }
type And struct{ Literal, Left, Right uint32 }
type Symbol struct {
	Kind  byte
	Index uint32
}
type Model struct {
	MaxVar                     uint32
	Inputs                     []uint32
	Latches                    []Latch
	Outputs                    []uint32
	ANDs                       []And
	Bad, Constraints, Fairness []uint32
	Justice                    [][]uint32
	Symbols                    map[Symbol]string
	Name                       string
	Comment                    []byte // Payload after the marker line, excluding its newline.
}

// symbolKindOrder ranks symbol kinds in the order AIGER lists their sections.
const symbolKindOrder = "ilobcjf"

// OrderedSymbols returns the keys of m.Symbols by kind, in AIGER section
// order (i, l, o, b, c, j, f), then index: the order writers emit and
// readers such as ABC's meet them. Ranging over Symbols directly would
// follow Go's randomized map order.
func OrderedSymbols(symbols map[Symbol]string) []Symbol {
	return slices.SortedFunc(maps.Keys(symbols), func(a, b Symbol) int {
		ra, rb := symbolKindRank(a.Kind), symbolKindRank(b.Kind)
		if ra != rb {
			return cmp.Compare(ra, rb)
		}
		if a.Kind != b.Kind {
			return cmp.Compare(a.Kind, b.Kind)
		}
		return cmp.Compare(a.Index, b.Index)
	})
}

// symbolKindRank places unknown kinds after the AIGER sections.
func symbolKindRank(kind byte) int {
	if r := strings.IndexByte(symbolKindOrder, kind); r >= 0 {
		return r
	}
	return len(symbolKindOrder)
}

// OrderedSymbols is OrderedSymbols(m.Symbols).
func (m *Model) OrderedSymbols() []Symbol { return OrderedSymbols(m.Symbols) }

// AigerProvider breaks the dependency cycle between graph managers and I/O.
// A manager exports a snapshot; IOA never imports its implementation package.
type AigerProvider interface{ AigerModel() (*Model, error) }

func (m *Model) AigerModel() (*Model, error) {
	if m == nil {
		return nil, fmt.Errorf("nil AIGER model")
	}
	return m, nil
}
func (m *Model) Check() error {
	return m.check(m.HasDenseTopologicalOrder())
}

func (m *Model) check(dense bool) error {
	if m == nil {
		return fmt.Errorf("nil AIGER model")
	}
	if strings.IndexByte(m.Name, 0) >= 0 {
		return fmt.Errorf("model name contains NUL")
	}
	if m.MaxVar > 0x7fffffff {
		return fmt.Errorf("AIGER variable index exceeds literal range")
	}
	if dense {
		for _, l := range m.Latches {
			if l.Init != 0 && l.Init != 1 && l.Init != l.Literal {
				return fmt.Errorf("invalid reset literal %d", l.Init)
			}
		}
		if err := m.checkRootLiterals(func(lit uint32) error {
			if lit>>1 > m.MaxVar {
				return fmt.Errorf("undefined literal %d", lit)
			}
			return nil
		}); err != nil {
			return err
		}
	} else if err := m.checkUnorderedStructure(); err != nil {
		return err
	}
	for _, symbol := range m.OrderedSymbols() {
		name := m.Symbols[symbol]
		if len(name) == 0 {
			return fmt.Errorf("empty symbol name")
		}
		var n int
		switch symbol.Kind {
		case 'i':
			n = len(m.Inputs)
		case 'l':
			n = len(m.Latches)
		case 'o':
			n = len(m.Outputs)
		case 'b':
			n = len(m.Bad)
		case 'c':
			n = len(m.Constraints)
		case 'j':
			n = len(m.Justice)
		case 'f':
			n = len(m.Fairness)
		default:
			return fmt.Errorf("invalid symbol kind %q", symbol.Kind)
		}
		if uint64(symbol.Index) >= uint64(n) {
			return fmt.Errorf("symbol index out of range")
		}
		for _, c := range name {
			if c == '\n' || c == '\r' || c == 0 {
				return fmt.Errorf("invalid symbol name")
			}
		}
	}
	return nil
}

// HasDenseTopologicalOrder reports whether inputs, latches and ANDs define
// consecutive variables and every AND follows both of its fanins.
func (m *Model) HasDenseTopologicalOrder() bool {
	if m == nil || m.MaxVar > 0x7fffffff || uint64(m.MaxVar) != uint64(len(m.Inputs))+uint64(len(m.Latches))+uint64(len(m.ANDs)) {
		return false
	}
	next := uint32(2)
	for _, lit := range m.Inputs {
		if lit != next {
			return false
		}
		next += 2
	}
	for _, latch := range m.Latches {
		if latch.Literal != next {
			return false
		}
		next += 2
	}
	for _, gate := range m.ANDs {
		if gate.Literal != next || gate.Left >= next || gate.Right >= next {
			return false
		}
		next += 2
	}
	return true
}

func (m *Model) checkRootLiterals(check func(uint32) error) error {
	for _, l := range m.Latches {
		if err := check(l.Next); err != nil {
			return err
		}
	}
	for _, list := range [][]uint32{m.Outputs, m.Bad, m.Constraints, m.Fairness} {
		for _, lit := range list {
			if err := check(lit); err != nil {
				return err
			}
		}
	}
	for _, list := range m.Justice {
		for _, lit := range list {
			if err := check(lit); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Model) checkUnorderedStructure() error {
	defined := map[uint32]bool{0: true}
	gates := map[uint32]And{}
	add := func(lit uint32) error {
		if lit == 0 || lit&1 != 0 || lit>>1 > m.MaxVar {
			return fmt.Errorf("invalid defining literal %d", lit)
		}
		if defined[lit>>1] {
			return fmt.Errorf("duplicate variable %d", lit>>1)
		}
		defined[lit>>1] = true
		return nil
	}
	for _, lit := range m.Inputs {
		if err := add(lit); err != nil {
			return err
		}
	}
	for _, l := range m.Latches {
		if err := add(l.Literal); err != nil {
			return err
		}
		if l.Init != 0 && l.Init != 1 && l.Init != l.Literal {
			return fmt.Errorf("invalid reset literal %d", l.Init)
		}
	}
	for _, a := range m.ANDs {
		if err := add(a.Literal); err != nil {
			return err
		}
		gates[a.Literal>>1] = a
	}
	check := func(lit uint32) error {
		if lit>>1 > m.MaxVar || !defined[lit>>1] {
			return fmt.Errorf("undefined literal %d", lit)
		}
		return nil
	}
	if err := m.checkRootLiterals(check); err != nil {
		return err
	}
	for _, a := range m.ANDs {
		if err := check(a.Left); err != nil {
			return err
		}
		if err := check(a.Right); err != nil {
			return err
		}
	}
	state := map[uint32]byte{}
	var visit func(uint32) error
	visit = func(v uint32) error {
		a, ok := gates[v]
		if !ok {
			return nil
		}
		if state[v] == 1 {
			return fmt.Errorf("combinational cycle at variable %d", v)
		}
		if state[v] == 2 {
			return nil
		}
		state[v] = 1
		if err := visit(a.Left >> 1); err != nil {
			return err
		}
		if err := visit(a.Right >> 1); err != nil {
			return err
		}
		state[v] = 2
		return nil
	}
	// Visit in variable order so a cyclic model always reports the same cycle.
	for _, v := range slices.Sorted(maps.Keys(gates)) {
		if err := visit(v); err != nil {
			return err
		}
	}
	return nil
}
