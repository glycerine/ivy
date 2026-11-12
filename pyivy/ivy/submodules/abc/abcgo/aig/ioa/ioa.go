// Package ioa reads and writes AIGER circuits using the Go standard library.
// The binary codecs originate in ABC's IOA package (Alan Mishchenko).
package ioa

import (
	"fmt"
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
	if m == nil {
		return fmt.Errorf("nil AIGER model")
	}
	if strings.IndexByte(m.Name, 0) >= 0 {
		return fmt.Errorf("model name contains NUL")
	}
	if m.MaxVar > 0x7fffffff {
		return fmt.Errorf("AIGER variable index exceeds literal range")
	}
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
	for v := range gates {
		if err := visit(v); err != nil {
			return err
		}
	}
	for symbol, name := range m.Symbols {
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
