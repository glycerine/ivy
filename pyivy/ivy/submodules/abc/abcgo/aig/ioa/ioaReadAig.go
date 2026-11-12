package ioa

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ParseError struct {
	Offset  int
	Message string
}

func (e *ParseError) Error() string { return fmt.Sprintf("AIGER byte %d: %s", e.Offset, e.Message) }
func Ioa_ReadAigerDecode(position *[]byte) (uint32, error) {
	data := *position
	for i := 0; i < 5; i++ {
		if i >= len(data) {
			return 0, fmt.Errorf("truncated AIGER integer")
		}
		if data[i]&0x80 == 0 {
			if i == 4 && data[i] > 15 {
				return 0, fmt.Errorf("AIGER integer overflow")
			}
			v, n := binary.Uvarint(data[:i+1])
			if n <= 0 {
				return 0, fmt.Errorf("invalid AIGER integer")
			}
			*position = data[i+1:]
			return uint32(v), nil
		}
	}
	return 0, fmt.Errorf("AIGER integer overflow")
}
func Ioa_WriteDecodeLiterals(position *[]byte, n int32) ([]uint32, error) {
	if n < 0 {
		return nil, fmt.Errorf("negative literal count")
	}
	result := make([]uint32, 0, n)
	if n == 0 {
		return result, nil
	}
	first, err := Ioa_ReadAigerDecode(position)
	if err != nil {
		return nil, err
	}
	result = append(result, first)
	for i := int32(1); i < n; i++ {
		delta, err := Ioa_ReadAigerDecode(position)
		if err != nil {
			return nil, err
		}
		value := int64(result[len(result)-1])
		if delta&1 != 0 {
			value -= int64(delta >> 1)
		} else {
			value += int64(delta >> 1)
		}
		if value < 0 || value > 0xffffffff {
			return nil, fmt.Errorf("compact literal delta overflow")
		}
		result = append(result, uint32(value))
	}
	return result, nil
}

type aigerReader struct {
	data []byte
	pos  int
}

func (r *aigerReader) fail(message string) error { return &ParseError{r.pos, message} }
func (r *aigerReader) line() ([]byte, error) {
	if r.pos >= len(r.data) {
		return nil, r.fail("unexpected end of file")
	}
	end := bytes.IndexByte(r.data[r.pos:], '\n')
	if end < 0 {
		return nil, r.fail("unterminated line")
	}
	line := r.data[r.pos : r.pos+end]
	r.pos += end + 1
	return bytes.TrimSuffix(line, []byte{'\r'}), nil
}
func (r *aigerReader) numbers(min, max int) ([]uint32, error) {
	line, err := r.line()
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(line))
	if len(fields) < min || len(fields) > max {
		return nil, r.fail("wrong number of fields")
	}
	nums := make([]uint32, len(fields))
	for i, f := range fields {
		v, err := strconv.ParseUint(f, 10, 32)
		if err != nil {
			return nil, r.fail("invalid unsigned integer")
		}
		nums[i] = uint32(v)
	}
	return nums, nil
}
func (r *aigerReader) decode() (uint32, error) {
	tail := r.data[r.pos:]
	v, err := Ioa_ReadAigerDecode(&tail)
	if err != nil {
		return 0, r.fail(err.Error())
	}
	r.pos = len(r.data) - len(tail)
	return v, nil
}
func (r *aigerReader) literals(n uint32) ([]uint32, error) {
	var result []uint32
	if n > 0 {
		result = make([]uint32, 0, n)
	}
	for i := uint32(0); i < n; i++ {
		line, err := r.numbers(1, 1)
		if err != nil {
			return nil, err
		}
		result = append(result, line[0])
	}
	return result, nil
}

// Read parses standard binary/ASCII AIGER 1.9 and ABC's compact aig2 variant.
// Data is copied into the model; the caller retains ownership of its buffer.
func Read(data []byte) (*Model, error) {
	r := &aigerReader{data: data}
	header, err := r.line()
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(header))
	if len(fields) < 6 || len(fields) > 10 {
		return nil, r.fail("invalid AIGER header")
	}
	format := fields[0]
	if format != "aig" && format != "aag" && format != "aig2" {
		return nil, r.fail("unknown AIGER format")
	}
	var sizes [9]uint32
	for i, f := range fields[1:] {
		v, err := strconv.ParseUint(f, 10, 32)
		if err != nil {
			return nil, r.fail("invalid header count")
		}
		sizes[i] = uint32(v)
	}
	max, inputs, latches, outputs, ands, bad, constraints, justice, fairness := sizes[0], sizes[1], sizes[2], sizes[3], sizes[4], sizes[5], sizes[6], sizes[7], sizes[8]
	definitions := uint64(inputs) + uint64(latches) + uint64(ands)
	if max > 0x7fffffff || definitions > uint64(max) {
		return nil, r.fail("inconsistent maximum variable/counts")
	}
	if format != "aag" && definitions != uint64(max) {
		return nil, r.fail("binary AIGER requires contiguous variable IDs")
	}
	// Bound header-directed allocation independently of malformed input sizes.
	if definitions+uint64(outputs)+uint64(bad)+uint64(constraints)+uint64(justice)+uint64(fairness) > 1<<24 {
		return nil, r.fail("AIGER object limit exceeded")
	}
	if format == "aig2" && (justice != 0 || fairness != 0) {
		return nil, r.fail("compact aig2 does not encode justice/fairness sections")
	}
	m := &Model{MaxVar: max, Symbols: map[Symbol]string{}}
	for i := uint32(0); i < inputs; i++ {
		lit := 2 * (i + 1)
		if format == "aag" {
			line, e := r.numbers(1, 1)
			if e != nil {
				return nil, e
			}
			lit = line[0]
		}
		m.Inputs = append(m.Inputs, lit)
	}
	if format == "aig2" {
		tail := r.data[r.pos:]
		drivers, e := Ioa_WriteDecodeLiterals(&tail, int32(latches+outputs+bad+constraints))
		if e != nil {
			return nil, r.fail(e.Error())
		}
		r.pos = len(data) - len(tail)
		for i := uint32(0); i < latches; i++ {
			m.Latches = append(m.Latches, Latch{Literal: 2 * (inputs + i + 1), Next: drivers[i]})
		}
		at := latches
		m.Outputs = append(m.Outputs, drivers[at:at+outputs]...)
		at += outputs
		m.Bad = append(m.Bad, drivers[at:at+bad]...)
		at += bad
		m.Constraints = append(m.Constraints, drivers[at:at+constraints]...)
	} else {
		for i := uint32(0); i < latches; i++ {
			min, maxFields := 1, 2
			if format == "aag" {
				min, maxFields = 2, 3
			}
			line, e := r.numbers(min, maxFields)
			if e != nil {
				return nil, e
			}
			l := Latch{Literal: 2 * (inputs + i + 1)}
			at := 0
			if format == "aag" {
				l.Literal = line[0]
				at = 1
			}
			l.Next = line[at]
			if len(line) > at+1 {
				l.Init = line[at+1]
			}
			m.Latches = append(m.Latches, l)
		}
		m.Outputs, err = r.literals(outputs)
		if err != nil {
			return nil, err
		}
		m.Bad, err = r.literals(bad)
		if err != nil {
			return nil, err
		}
		m.Constraints, err = r.literals(constraints)
		if err != nil {
			return nil, err
		}
		counts, err := r.literals(justice)
		if err != nil {
			return nil, err
		}
		var total uint64
		for _, n := range counts {
			total += uint64(n)
		}
		if total > uint64(len(data)) {
			return nil, r.fail("justice sizes exceed input")
		}
		for _, n := range counts {
			list, err := r.literals(n)
			if err != nil {
				return nil, err
			}
			m.Justice = append(m.Justice, list)
		}
		m.Fairness, err = r.literals(fairness)
		if err != nil {
			return nil, err
		}
	}
	for i := uint32(0); i < ands; i++ {
		a := And{Literal: 2 * (inputs + latches + i + 1)}
		if format == "aag" {
			line, e := r.numbers(3, 3)
			if e != nil {
				return nil, e
			}
			a.Literal, a.Left, a.Right = line[0], line[1], line[2]
		} else {
			d0, e := r.decode()
			if e != nil {
				return nil, e
			}
			if d0 > a.Literal {
				return nil, r.fail("invalid first AND delta")
			}
			a.Left = a.Literal - d0
			d1, e := r.decode()
			if e != nil {
				return nil, e
			}
			if d1 > a.Left {
				return nil, r.fail("invalid second AND delta")
			}
			a.Right = a.Left - d1
			if a.Left >= a.Literal || a.Right >= a.Literal {
				return nil, r.fail("non-topological binary AND")
			}
		}
		m.ANDs = append(m.ANDs, a)
	}
	for r.pos < len(data) {
		if data[r.pos] == 'c' && (r.pos+1 == len(data) || data[r.pos+1] == '\n' || data[r.pos+1] == '\r' || data[r.pos+1] == 'n') {
			comment := data[r.pos+1:]
			if len(comment) > 0 && comment[0] == 'n' {
				end := bytes.IndexByte(comment, 0)
				if end < 0 {
					return nil, r.fail("unterminated ABC model name")
				}
				m.Name = string(comment[1:end])
				comment = comment[end+1:]
			}
			if bytes.HasPrefix(comment, []byte("\r\n")) {
				comment = comment[2:]
			} else if len(comment) > 0 && comment[0] == '\n' {
				comment = comment[1:]
			}
			m.Comment = append([]byte(nil), comment...)
			r.pos = len(data)
			break
		}
		line, e := r.line()
		if e != nil {
			return nil, e
		}
		if len(line) == 0 {
			return nil, r.fail("empty symbol line")
		}
		space := bytes.IndexByte(line, ' ')
		if space < 2 {
			return nil, r.fail("invalid symbol line")
		}
		index, e := strconv.ParseUint(string(line[1:space]), 10, 32)
		if e != nil {
			return nil, r.fail("invalid symbol index")
		}
		symbol := Symbol{Kind: line[0], Index: uint32(index)}
		if _, exists := m.Symbols[symbol]; exists {
			return nil, r.fail("duplicate symbol")
		}
		m.Symbols[symbol] = string(line[space+1:])
	}
	if err := m.Check(); err != nil {
		return nil, r.fail(err.Error())
	}
	return m, nil
}
func Ioa_ReadAigerFromMemory(data []byte, n int32, check int32) (*Model, error) {
	if n < 0 || int64(n) > int64(len(data)) {
		return nil, fmt.Errorf("invalid AIGER buffer length")
	}
	return Read(data[:n])
}
func Ioa_ReadAiger(path string, check int32) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Read(data)
}
