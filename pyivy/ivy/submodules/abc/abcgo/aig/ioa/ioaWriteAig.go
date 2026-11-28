package ioa

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

type WriteOptions struct {
	ASCII, Compact, Symbols bool
	// Trailer, when non-nil, is written after the symbols in place of the
	// model's name and comment (Io_WriteAiger's "c" section).
	Trailer []byte
}

func Ioa_ObjMakeLit(variable, complement int32) uint32 {
	return uint32(variable)<<1 | uint32(complement)&1
}
func Ioa_WriteAigerEncode(buffer []byte, pos int32, value uint32) int32 {
	var encoded [5]byte
	n := binary.PutUvarint(encoded[:], uint64(value))
	if pos < 0 || int64(pos)+int64(n) > int64(len(buffer)) {
		panic("short AIGER encoding buffer")
	}
	copy(buffer[pos:], encoded[:n])
	return pos + int32(n)
}
func Ioa_WriteAigerEncodeStr(buffer *[]byte, value uint32) {
	*buffer = binary.AppendUvarint(*buffer, uint64(value))
}
func Ioa_WriteEncodeLiterals(literals []uint32) ([]byte, error) {
	if len(literals) == 0 {
		return nil, nil
	}
	data := binary.AppendUvarint(nil, uint64(literals[0]))
	previous := literals[0]
	for _, lit := range literals[1:] {
		delta := int64(lit) - int64(previous)
		sign := uint64(0)
		if delta < 0 {
			delta = -delta
			sign = 1
		}
		code := uint64(delta)<<1 | sign
		if code > 0xffffffff {
			return nil, fmt.Errorf("compact literal delta exceeds 32 bits")
		}
		data = binary.AppendUvarint(data, code)
		previous = lit
	}
	return data, nil
}

// Normalize returns a dense, topologically ordered copy for binary AIGER.
// Every original AND is retained, including gates not reachable from outputs.
func (m *Model) Normalize() (*Model, error) {
	return m.normalized(true)
}

// The encoder consumes dense definitions directly and orders each pair of
// fanins while emitting it. Normalize must return an independent, sorted copy.
func (m *Model) normalized(copyDense bool) (*Model, error) {
	dense := m.HasDenseTopologicalOrder()
	if err := m.check(dense); err != nil {
		return nil, err
	}
	if dense && !copyDense {
		return m, nil
	}
	result := &Model{Name: m.Name, Comment: append([]byte(nil), m.Comment...), Symbols: map[Symbol]string{}}
	for s, name := range m.Symbols {
		result.Symbols[s] = name
	}
	if dense {
		result.MaxVar = m.MaxVar
		result.Inputs = append([]uint32(nil), m.Inputs...)
		result.Latches = append([]Latch(nil), m.Latches...)
		result.ANDs = append([]And(nil), m.ANDs...)
		for i := range result.ANDs {
			a := &result.ANDs[i]
			if a.Left < a.Right {
				a.Left, a.Right = a.Right, a.Left
			}
		}
		result.Outputs = append([]uint32(nil), m.Outputs...)
		result.Bad = append([]uint32(nil), m.Bad...)
		result.Constraints = append([]uint32(nil), m.Constraints...)
		result.Fairness = append([]uint32(nil), m.Fairness...)
		for _, group := range m.Justice {
			result.Justice = append(result.Justice, append([]uint32(nil), group...))
		}
		return result, nil
	}
	mapped := map[uint32]uint32{0: 0}
	next := uint32(1)
	for _, lit := range m.Inputs {
		mapped[lit>>1] = next << 1
		result.Inputs = append(result.Inputs, next<<1)
		next++
	}
	for _, l := range m.Latches {
		mapped[l.Literal>>1] = next << 1
		result.Latches = append(result.Latches, Latch{Literal: next << 1})
		next++
	}
	gates := map[uint32]And{}
	for _, a := range m.ANDs {
		gates[a.Literal>>1] = a
	}
	var translate func(uint32) uint32
	translate = func(lit uint32) uint32 {
		v := lit >> 1
		if value, ok := mapped[v]; ok {
			return value ^ (lit & 1)
		}
		a := gates[v]
		left, right := translate(a.Left), translate(a.Right)
		lhs := next << 1
		next++
		mapped[v] = lhs
		if left < right {
			left, right = right, left
		}
		result.ANDs = append(result.ANDs, And{lhs, left, right})
		return lhs ^ (lit & 1)
	}
	for _, a := range m.ANDs {
		translate(a.Literal)
	}
	for i, l := range m.Latches {
		result.Latches[i].Next = translate(l.Next)
		result.Latches[i].Init = l.Init
		if l.Init == l.Literal {
			result.Latches[i].Init = result.Latches[i].Literal
		}
	}
	list := func(values []uint32) []uint32 {
		var r []uint32
		for _, lit := range values {
			r = append(r, translate(lit))
		}
		return r
	}
	result.Outputs = list(m.Outputs)
	result.Bad = list(m.Bad)
	result.Constraints = list(m.Constraints)
	result.Fairness = list(m.Fairness)
	for _, j := range m.Justice {
		result.Justice = append(result.Justice, list(j))
	}
	result.MaxVar = next - 1
	return result, nil
}
func getModel(provider AigerProvider) (*Model, error) {
	if provider == nil {
		return nil, fmt.Errorf("nil AIGER provider")
	}
	m, err := provider.AigerModel()
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("nil AIGER model")
	}
	return m, nil
}
func Ioa_WriteAigerLiterals(provider AigerProvider) ([]uint32, error) {
	m, err := getModel(provider)
	if err != nil {
		return nil, err
	}
	m, err = m.normalized(false)
	if err != nil {
		return nil, err
	}
	result := make([]uint32, 0, len(m.Latches)+len(m.Outputs)+len(m.Bad)+len(m.Constraints))
	for _, l := range m.Latches {
		result = append(result, l.Next)
	}
	result = append(result, m.Outputs...)
	result = append(result, m.Bad...)
	result = append(result, m.Constraints...)
	return result, nil
}
func Marshal(provider AigerProvider, options WriteOptions) ([]byte, error) {
	m, err := getModel(provider)
	if err != nil {
		return nil, err
	}
	if options.ASCII && options.Compact {
		return nil, fmt.Errorf("ASCII and compact encodings are mutually exclusive")
	}
	if options.ASCII {
		if err = m.Check(); err != nil {
			return nil, err
		}
	} else {
		m, err = m.normalized(false)
		if err != nil {
			return nil, err
		}
	}
	if options.Compact {
		if len(m.Justice) > 0 || len(m.Fairness) > 0 {
			return nil, fmt.Errorf("compact aig2 does not support justice/fairness")
		}
		for _, l := range m.Latches {
			if l.Init != 0 {
				return nil, fmt.Errorf("compact aig2 cannot preserve nonzero/uninitialized latch reset")
			}
		}
	}
	// Each binary AND contributes two varints. Reserve four bytes per gate
	// and, when requested, room for typical symbol records.
	capacity := 4 * len(m.ANDs)
	if options.Symbols {
		capacity += 12 * len(m.Symbols)
	}
	data := make([]byte, 0, capacity)
	number := func(n uint32) { data = strconv.AppendUint(data, uint64(n), 10) }
	line := func(n uint32) {
		number(n)
		data = append(data, '\n')
	}
	format := "aig"
	if options.ASCII {
		format = "aag"
	}
	if options.Compact {
		format = "aig2"
	}
	counts := []uint32{m.MaxVar, uint32(len(m.Inputs)), uint32(len(m.Latches)), uint32(len(m.Outputs)), uint32(len(m.ANDs)), uint32(len(m.Bad)), uint32(len(m.Constraints)), uint32(len(m.Justice)), uint32(len(m.Fairness))}
	last := len(counts) - 1
	minLast := 4
	if m.ExtendedHeader {
		minLast = 6 // B and C are written
	}
	for last > minLast && counts[last] == 0 {
		last--
	}
	data = append(data, format...)
	for _, n := range counts[:last+1] {
		data = append(data, ' ')
		number(n)
	}
	data = append(data, '\n')
	if options.ASCII {
		for _, lit := range m.Inputs {
			line(lit)
		}
	}
	if options.Compact {
		literals := make([]uint32, 0)
		for _, l := range m.Latches {
			literals = append(literals, l.Next)
		}
		literals = append(literals, m.Outputs...)
		literals = append(literals, m.Bad...)
		literals = append(literals, m.Constraints...)
		encoded, err := Ioa_WriteEncodeLiterals(literals)
		if err != nil {
			return nil, err
		}
		data = append(data, encoded...)
	} else {
		for _, l := range m.Latches {
			if options.ASCII {
				number(l.Literal)
				data = append(data, ' ')
			}
			number(l.Next)
			if l.Init != 0 {
				data = append(data, ' ')
				number(l.Init)
			}
			data = append(data, '\n')
		}
		for _, list := range [][]uint32{m.Outputs, m.Bad, m.Constraints} {
			for _, lit := range list {
				line(lit)
			}
		}
		for _, j := range m.Justice {
			data = strconv.AppendInt(data, int64(len(j)), 10)
			data = append(data, '\n')
		}
		for _, j := range m.Justice {
			for _, lit := range j {
				line(lit)
			}
		}
		for _, lit := range m.Fairness {
			line(lit)
		}
	}
	for _, a := range m.ANDs {
		if options.ASCII {
			number(a.Literal)
			data = append(data, ' ')
			number(a.Left)
			data = append(data, ' ')
			line(a.Right)
		} else {
			left, right := a.Left, a.Right
			if left < right {
				left, right = right, left
			}
			data = binary.AppendUvarint(data, uint64(a.Literal-left))
			data = binary.AppendUvarint(data, uint64(left-right))
		}
	}
	if options.Symbols && len(m.Symbols) != 0 {
		const kinds = "ilobcjf"
		counts := [...]int{len(m.Inputs), len(m.Latches), len(m.Outputs), len(m.Bad), len(m.Constraints), len(m.Justice), len(m.Fairness)}
		writeSymbol := func(key Symbol, name string) {
			data = append(data, key.Kind)
			number(key.Index)
			data = append(data, ' ')
			data = append(data, name...)
			data = append(data, '\n')
		}
		slots := 0
		for _, n := range counts {
			slots += n
		}
		if len(m.Symbols) >= slots/2 {
			for rank, n := range counts {
				for i := range n {
					key := Symbol{kinds[rank], uint32(i)}
					if name, ok := m.Symbols[key]; ok {
						writeSymbol(key, name)
					}
				}
			}
		} else {
			keys := make([]uint64, 0, len(m.Symbols))
			for key := range m.Symbols {
				rank := strings.IndexByte(kinds, key.Kind)
				keys = append(keys, uint64(rank)<<32|uint64(key.Index))
			}
			slices.Sort(keys)
			for _, ordered := range keys {
				key := Symbol{kinds[ordered>>32], uint32(ordered)}
				writeSymbol(key, m.Symbols[key])
			}
		}
	}
	if options.Trailer != nil {
		return append(data, options.Trailer...), nil
	}
	if m.Name != "" || len(m.Comment) > 0 {
		data = append(data, 'c')
		if m.Name != "" {
			data = append(data, 'n')
			data = append(data, m.Name...)
			data = append(data, 0)
		}
		data = append(data, '\n')
		data = append(data, m.Comment...)
	}
	return data, nil
}
func Write(writer io.Writer, provider AigerProvider, options WriteOptions) error {
	data, err := Marshal(provider, options)
	if err != nil {
		return err
	}
	n, err := writer.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	return err
}
func Ioa_WriteAigerIntoMemoryStr(provider AigerProvider) ([]byte, error) {
	return Marshal(provider, WriteOptions{})
}
func Ioa_WriteAigerIntoMemory(provider AigerProvider, size *int32) ([]byte, error) {
	data, err := Marshal(provider, WriteOptions{})
	if err != nil {
		return nil, err
	}
	if uint64(len(data)) > 0x7fffffff {
		return nil, fmt.Errorf("AIGER buffer exceeds int32 size")
	}
	if size != nil {
		*size = int32(len(data))
	}
	return data, nil
}
func Ioa_WriteAiger(provider AigerProvider, path string, symbols, compact int32) error {
	data, err := Marshal(provider, WriteOptions{Symbols: symbols != 0, Compact: compact != 0})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
func Ioa_WriteAigerBufferTest(provider AigerProvider, path string, symbols, compact int32) error {
	return Ioa_WriteAiger(provider, path, symbols, compact)
}
