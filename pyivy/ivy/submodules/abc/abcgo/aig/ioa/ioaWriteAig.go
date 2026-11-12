package ioa

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
)

type WriteOptions struct{ ASCII, Compact, Symbols bool }

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
	if err := m.Check(); err != nil {
		return nil, err
	}
	result := &Model{Name: m.Name, Comment: append([]byte(nil), m.Comment...), Symbols: map[Symbol]string{}}
	for s, name := range m.Symbols {
		result.Symbols[s] = name
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
	m, err = m.Normalize()
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
		m, err = m.Normalize()
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
	var data bytes.Buffer
	format := "aig"
	if options.ASCII {
		format = "aag"
	}
	if options.Compact {
		format = "aig2"
	}
	counts := []uint32{m.MaxVar, uint32(len(m.Inputs)), uint32(len(m.Latches)), uint32(len(m.Outputs)), uint32(len(m.ANDs)), uint32(len(m.Bad)), uint32(len(m.Constraints)), uint32(len(m.Justice)), uint32(len(m.Fairness))}
	last := len(counts) - 1
	for last >= 5 && counts[last] == 0 {
		last--
	}
	fmt.Fprint(&data, format)
	for _, n := range counts[:last+1] {
		fmt.Fprintf(&data, " %d", n)
	}
	data.WriteByte('\n')
	if options.ASCII {
		for _, lit := range m.Inputs {
			fmt.Fprintln(&data, lit)
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
		data.Write(encoded)
	} else {
		for _, l := range m.Latches {
			if options.ASCII {
				fmt.Fprintf(&data, "%d ", l.Literal)
			}
			fmt.Fprint(&data, l.Next)
			if l.Init != 0 {
				fmt.Fprintf(&data, " %d", l.Init)
			}
			data.WriteByte('\n')
		}
		for _, list := range [][]uint32{m.Outputs, m.Bad, m.Constraints} {
			for _, lit := range list {
				fmt.Fprintln(&data, lit)
			}
		}
		for _, j := range m.Justice {
			fmt.Fprintln(&data, len(j))
		}
		for _, j := range m.Justice {
			for _, lit := range j {
				fmt.Fprintln(&data, lit)
			}
		}
		for _, lit := range m.Fairness {
			fmt.Fprintln(&data, lit)
		}
	}
	for _, a := range m.ANDs {
		if options.ASCII {
			fmt.Fprintf(&data, "%d %d %d\n", a.Literal, a.Left, a.Right)
		} else {
			left, right := a.Left, a.Right
			if left < right {
				left, right = right, left
			}
			var encoded []byte
			encoded = binary.AppendUvarint(encoded, uint64(a.Literal-left))
			encoded = binary.AppendUvarint(encoded, uint64(left-right))
			data.Write(encoded)
		}
	}
	if options.Symbols {
		keys := make([]Symbol, 0, len(m.Symbols))
		for key := range m.Symbols {
			keys = append(keys, key)
		}
		rank := func(k byte) int {
			switch k {
			case 'i':
				return 0
			case 'l':
				return 1
			case 'o':
				return 2
			case 'b':
				return 3
			case 'c':
				return 4
			case 'j':
				return 5
			case 'f':
				return 6
			}
			return 7
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].Kind != keys[j].Kind {
				return rank(keys[i].Kind) < rank(keys[j].Kind)
			}
			return keys[i].Index < keys[j].Index
		})
		for _, key := range keys {
			fmt.Fprintf(&data, "%c%d %s\n", key.Kind, key.Index, m.Symbols[key])
		}
	}
	if m.Name != "" || len(m.Comment) > 0 {
		data.WriteByte('c')
		if m.Name != "" {
			data.WriteByte('n')
			data.WriteString(m.Name)
			data.WriteByte(0)
		}
		data.WriteByte('\n')
		data.Write(m.Comment)
	}
	return data.Bytes(), nil
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
