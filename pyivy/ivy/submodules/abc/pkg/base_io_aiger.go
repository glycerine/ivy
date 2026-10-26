package abc

// Derived from ABC src/base/io/ioReadAiger.c and ioWriteAiger.c.
// Original ABC algorithms by Alan Mishchenko.
// Checked Go parsing preserves raw AIGER constraint polarity separately from
// ABC's internal inverted-constraint convention. See PORT_STATUS.md.

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/glycerine/insdict"
)

var (
	ErrFormat      = errors.New("invalid file format")
	ErrUnsupported = errors.New("unsupported feature")
)

// LatchInit distinguishes known false, known true and unconstrained initial
// state. Unknown initialization must never be silently treated as false.
type LatchInit uint8

const (
	LatchInitZero LatchInit = iota
	LatchInitOne
	LatchInitUnknown
)

// AigerSymbol names one terminal in its own category (i,l,o,b,c).
type AigerSymbol struct {
	Kind  byte
	Index uint32
}
type AigerSymbolName struct {
	Symbol AigerSymbol
	Name   string
}

// AigerModel preserves the safety-related sections of binary AIGER. Graph
// POs are ordinary outputs, then bad-state properties, then raw positive
// constraints; latch inputs are the final COs. Justice/fairness are rejected
// until liveness support is implemented. The model must not be copied.
type AigerModel struct {
	graph                     *AigManager
	outputs, bad, constraints int
	initial                   []LatchInit
	symbols                   *insdict.Dict[AigerSymbol, string]
	comment                   []byte
}

func (m *AigerModel) Graph() *AigManager         { return m.graph }
func (m *AigerModel) OrdinaryOutputCount() int   { return m.outputs }
func (m *AigerModel) BadCount() int              { return m.bad }
func (m *AigerModel) ConstraintCount() int       { return m.constraints }
func (m *AigerModel) InitialValues() []LatchInit { return slices.Clone(m.initial) }
func (m *AigerModel) Comment() []byte            { return slices.Clone(m.comment) }
func (m *AigerModel) Symbols() []AigerSymbolName {
	result := make([]AigerSymbolName, 0, m.symbols.Len())
	for key, value := range m.symbols.All() {
		result = append(result, AigerSymbolName{key, value})
	}
	return result
}
func newAigerSymbols() *insdict.Dict[AigerSymbol, string] {
	return insdict.NewDictFunc[AigerSymbol, string](func(s AigerSymbol) uint64 { return insdict.Mix64(uint64(s.Index)<<8 | uint64(s.Kind)) })
}
func (m *AigerModel) symbolBound(kind byte) int {
	switch kind {
	case 'i':
		return m.graph.PrimaryInputCount()
	case 'l':
		return len(m.initial)
	case 'o':
		return m.outputs
	case 'b':
		return m.bad
	case 'c':
		return m.constraints
	}
	return -1
}
func (m *AigerModel) Validate() error { return m.validate(true) }
func (m *AigerModel) validate(checkGraphConstraints bool) error {
	if m == nil || m.graph == nil || m.symbols == nil {
		return ErrInvariant
	}
	if err := m.graph.Check(); err != nil {
		return err
	}
	if m.outputs < 0 || m.bad < 0 || m.constraints < 0 || len(m.initial) != m.graph.RegisterCount() {
		return ErrDimension
	}
	n, err := checkedSize(m.outputs, m.bad, 1, math.MaxInt32)
	if err != nil {
		return err
	}
	n, err = checkedSize(n, m.constraints, 1, math.MaxInt32)
	if err != nil {
		return err
	}
	if n != m.graph.PrimaryOutputCount() || (checkGraphConstraints && m.graph.ConstraintCount() != m.constraints) {
		return ErrDimension
	}
	for _, v := range m.initial {
		if v > LatchInitUnknown {
			return ErrDimension
		}
	}
	for key, value := range m.symbols.All() {
		bound := m.symbolBound(key.Kind)
		if bound < 0 || uint64(key.Index) >= uint64(bound) || value == "" || bytes.ContainsAny([]byte(value), "\x00\r\n") {
			return ErrDimension
		}
	}
	return nil
}

// AigerLimits bounds untrusted parsing before allocation. Zero fields select
// defaults; negative fields are invalid. MaxOutputs includes O+B+C.
type AigerLimits struct{ MaxBytes, MaxVariables, MaxOutputs int }

func (l AigerLimits) normalized() (AigerLimits, error) {
	if l.MaxBytes < 0 || l.MaxVariables < 0 || l.MaxOutputs < 0 {
		return l, ErrDimension
	}
	if l.MaxBytes == 0 {
		l.MaxBytes = 32 << 20
	}
	if l.MaxVariables == 0 {
		l.MaxVariables = 1 << 20
	}
	if l.MaxOutputs == 0 {
		l.MaxOutputs = 1 << 20
	}
	if uint64(l.MaxVariables) > uint64(MaxLiteralID) {
		return l, ErrLimit
	}
	return l, nil
}

type aigerCursor struct {
	data []byte
	pos  int
}

func (c *aigerCursor) bad(message string) error {
	return fmt.Errorf("%w at byte %d: %s", ErrFormat, c.pos, message)
}
func (c *aigerCursor) line() ([]byte, error) {
	n := bytes.IndexByte(c.data[c.pos:], '\n')
	if n < 0 {
		return nil, c.bad("unterminated line")
	}
	start := c.pos
	c.pos += n + 1
	return c.data[start : start+n], nil
}
func aigerDecimal(token []byte, limit uint32) (uint32, error) {
	if len(token) == 0 {
		return 0, ErrFormat
	}
	var value uint32
	for _, b := range token {
		if b < '0' || b > '9' {
			return 0, ErrFormat
		}
		digit := uint32(b - '0')
		if digit > limit || value > (limit-digit)/10 {
			return 0, ErrFormat
		}
		value = 10*value + digit
	}
	return value, nil
}
func (c *aigerCursor) numbers(minimum, maximum int, limit uint32) ([]uint32, error) {
	line, err := c.line()
	if err != nil {
		return nil, err
	}
	fields := bytes.SplitN(line, []byte{' '}, maximum+1)
	if len(fields) < minimum || len(fields) > maximum {
		return nil, c.bad("wrong field count")
	}
	result := make([]uint32, len(fields))
	for i, field := range fields {
		result[i], err = aigerDecimal(field, limit)
		if err != nil {
			return nil, c.bad("invalid unsigned decimal")
		}
	}
	return result, nil
}
func (c *aigerCursor) varint() (uint32, error) {
	var value uint32
	for i := 0; i < 5; i++ {
		if c.pos == len(c.data) {
			return 0, c.bad("truncated binary literal")
		}
		b := c.data[c.pos]
		c.pos++
		if i == 4 && b > 15 {
			return 0, c.bad("binary literal overflow")
		}
		value |= uint32(b&127) << uint(7*i)
		if b < 128 {
			return value, nil
		}
	}
	return 0, c.bad("binary literal overflow")
}

// ReadAiger parses standard binary AIGER and ABC's compact aig2 variant.
// ASCII AAG, compressed input and liveness are not yet implemented and are
// rejected explicitly. The caller retains ownership of data; it is not kept.
func ReadAiger(data []byte, limits AigerLimits) (*AigerModel, error) {
	limits, err := limits.normalized()
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	cur := aigerCursor{data: data}
	header, err := cur.line()
	if err != nil {
		return nil, err
	}
	fields := bytes.SplitN(header, []byte{' '}, 11)
	if len(fields) < 6 || len(fields) > 10 {
		return nil, cur.bad("invalid header field count")
	}
	compact := bytes.Equal(fields[0], []byte("aig2"))
	if !compact && !bytes.Equal(fields[0], []byte("aig")) {
		return nil, fmt.Errorf("%w: expected binary aig or aig2", ErrUnsupported)
	}
	var counts [9]uint32
	for i, field := range fields[1:] {
		counts[i], err = aigerDecimal(field, math.MaxInt32)
		if err != nil {
			return nil, cur.bad("invalid header number")
		}
	}
	maximum, inputs, latches, outputs, ands, bad, constraints, justice, fairness := counts[0], counts[1], counts[2], counts[3], counts[4], counts[5], counts[6], counts[7], counts[8]
	if justice != 0 || fairness != 0 {
		return nil, fmt.Errorf("%w: AIGER justice/fairness", ErrUnsupported)
	}
	if uint64(inputs)+uint64(latches)+uint64(ands) != uint64(maximum) {
		return nil, cur.bad("M != I + L + A")
	}
	outCount := uint64(outputs) + uint64(bad) + uint64(constraints)
	if uint64(maximum) > uint64(limits.MaxVariables) || outCount > uint64(limits.MaxOutputs) {
		return nil, ErrLimit
	}
	// Counts cannot cause large allocations from a tiny malicious header.
	// Every driver consumes at least one binary byte or two ASCII bytes;
	// each AND consumes at least two binary bytes.
	minBytes := uint64(latches) + outCount
	if !compact {
		minBytes *= 2
	}
	minBytes += 2 * uint64(ands)
	if minBytes > uint64(len(data)-cur.pos) {
		return nil, cur.bad("declared sections exceed remaining bytes")
	}
	objectBound := uint64(maximum) + outCount + uint64(latches) + 1
	if objectBound > uint64(MaxLiteralID)+1 {
		return nil, ErrLimit
	}
	literalBound := 2*maximum + 1
	drivers := make([]uint32, int(latches)+int(outCount))
	model := &AigerModel{graph: NewAigManager(), outputs: int(outputs), bad: int(bad), constraints: int(constraints), initial: make([]LatchInit, int(latches)), symbols: newAigerSymbols()}
	if compact {
		var previous int64
		for i := range drivers {
			value, err := cur.varint()
			if err != nil {
				return nil, err
			}
			if i == 0 {
				previous = int64(value)
			} else {
				delta := int64(value >> 1)
				if value&1 != 0 {
					delta = -delta
				}
				previous += delta
			}
			if previous < 0 || previous > int64(literalBound) {
				return nil, cur.bad("compact driver outside literal range")
			}
			drivers[i] = uint32(previous)
		}
	} else {
		for i := range drivers {
			maxFields := 1
			if i < int(latches) {
				maxFields = 2
			}
			values, err := cur.numbers(1, maxFields, literalBound)
			if err != nil {
				return nil, err
			}
			drivers[i] = values[0]
			if len(values) == 2 {
				switch values[1] {
				case 0:
					model.initial[i] = LatchInitZero
				case 1:
					model.initial[i] = LatchInitOne
				default:
					if values[1] != 2*(inputs+uint32(i)+1) {
						return nil, cur.bad("invalid latch initialization")
					}
					model.initial[i] = LatchInitUnknown
				}
			}
		}
	}
	mapping := make([]AigLit, int(maximum)+1)
	mapping[0] = AigConst0
	for i := uint32(1); i <= inputs+latches; i++ {
		l, err := model.graph.CreateCI()
		if err != nil {
			return nil, err
		}
		mapping[i] = l
	}
	convert := func(raw uint32) AigLit { return mapping[raw>>1].NotIf(raw&1 != 0) }
	for i := uint32(0); i < ands; i++ {
		lhs := 2 * (inputs + latches + i + 1)
		delta0, err := cur.varint()
		if err != nil {
			return nil, err
		}
		if delta0 > lhs {
			return nil, cur.bad("first AND delta underflow")
		}
		rhs0 := lhs - delta0
		delta1, err := cur.varint()
		if err != nil {
			return nil, err
		}
		if delta1 > rhs0 {
			return nil, cur.bad("second AND delta underflow")
		}
		rhs1 := rhs0 - delta1
		if rhs0>>1 >= lhs>>1 || rhs1>>1 >= lhs>>1 {
			return nil, cur.bad("non-topological AND fanin")
		}
		l, err := model.graph.And(convert(rhs0), convert(rhs1))
		if err != nil {
			return nil, err
		}
		mapping[lhs>>1] = l
	}
	// COs must be ordinary/bad/constraint outputs followed by latch inputs,
	// even though the file stores latch drivers before output drivers.
	for _, driver := range drivers[int(latches):] {
		if _, err := model.graph.CreateCO(convert(driver)); err != nil {
			return nil, err
		}
	}
	for _, driver := range drivers[:int(latches)] {
		if _, err := model.graph.CreateCO(convert(driver)); err != nil {
			return nil, err
		}
	}
	if err := model.graph.SetRegisterCount(int(latches)); err != nil {
		return nil, err
	}
	if err := model.graph.SetConstraintCount(int(constraints)); err != nil {
		return nil, err
	}
	for cur.pos < len(data) {
		line, err := cur.line()
		if err != nil {
			return nil, err
		}
		if bytes.Equal(line, []byte("c")) {
			model.comment = slices.Clone(data[cur.pos:])
			cur.pos = len(data)
			break
		}
		if len(line) < 3 {
			return nil, cur.bad("invalid symbol line")
		}
		separator := bytes.IndexByte(line, ' ')
		if separator < 2 || separator == len(line)-1 {
			return nil, cur.bad("invalid symbol name")
		}
		bound := model.symbolBound(line[0])
		if bound < 0 {
			return nil, cur.bad("unknown symbol kind")
		}
		index, err := aigerDecimal(line[1:separator], math.MaxInt32)
		if err != nil || uint64(index) >= uint64(bound) {
			return nil, cur.bad("symbol index outside category")
		}
		key := AigerSymbol{line[0], index}
		if _, exists := model.symbols.Get2(key); exists {
			return nil, cur.bad("duplicate symbol")
		}
		name := line[separator+1:]
		if bytes.ContainsAny(name, "\x00\r") {
			return nil, cur.bad("invalid symbol name")
		}
		model.symbols.Put(key, string(name))
	}
	if err := model.Validate(); err != nil {
		return nil, err
	}
	return model, nil
}

func aigerAppendNumber(data []byte, n uint32) []byte { return strconv.AppendUint(data, uint64(n), 10) }
func aigerAppendVarint(data []byte, n uint32) []byte {
	for n >= 128 {
		data = append(data, byte(n&127)|128)
		n >>= 7
	}
	return append(data, byte(n))
}

// MarshalBinary writes standard binary AIGER with terminal names and comments.
// ABC's compact variant is accepted by ReadAiger but is not emitted here.
// Buffer objects are bypassed; tombstones do not consume external variables.
func (m *AigerModel) MarshalBinary() ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	g := m.graph
	nInputs, nRegs := g.PrimaryInputCount(), g.RegisterCount()
	raw := make([]uint32, len(g.objects))
	raw[0] = 1 // internal constant one -> external true
	for index, id := range g.cis {
		raw[id] = 2 * uint32(index+1)
	}
	convert := func(l AigLit) uint32 { return raw[l.ID()] ^ uint32(l&1) }
	binary := make([]byte, 0)
	next := uint32(len(g.cis) + 1)
	for i, obj := range g.objects {
		switch obj.Type {
		case AigBuffer:
			raw[i] = convert(obj.Fanin0)
		case AigAND:
			lhs := 2 * next
			next++
			a, b := convert(obj.Fanin0), convert(obj.Fanin1)
			if a < b {
				a, b = b, a
			}
			if a >= lhs {
				return nil, ErrInvariant
			}
			binary = aigerAppendVarint(binary, lhs-a)
			binary = aigerAppendVarint(binary, a-b)
			raw[i] = lhs
		}
	}
	header := []uint32{next - 1, uint32(nInputs), uint32(nRegs), uint32(m.outputs), uint32(g.AndCount())}
	if m.bad != 0 || m.constraints != 0 {
		header = append(header, uint32(m.bad), uint32(m.constraints))
	}
	result := []byte("aig")
	for _, value := range header {
		result = append(result, ' ')
		result = aigerAppendNumber(result, value)
	}
	result = append(result, '\n')
	for i := 0; i < nRegs; i++ {
		result = aigerAppendNumber(result, convert(g.objects[g.cos[g.PrimaryOutputCount()+i]].Fanin0))
		if m.initial[i] != LatchInitZero {
			result = append(result, ' ')
			initial := uint32(1)
			if m.initial[i] == LatchInitUnknown {
				initial = 2 * uint32(nInputs+i+1)
			}
			result = aigerAppendNumber(result, initial)
		}
		result = append(result, '\n')
	}
	for _, id := range g.cos[:g.PrimaryOutputCount()] {
		result = aigerAppendNumber(result, convert(g.objects[id].Fanin0))
		result = append(result, '\n')
	}
	result = append(result, binary...)
	for key, name := range m.symbols.All() {
		result = append(result, key.Kind)
		result = aigerAppendNumber(result, key.Index)
		result = append(result, ' ')
		result = append(result, name...)
		result = append(result, '\n')
	}
	if m.comment != nil {
		result = append(result, 'c', '\n')
		result = append(result, m.comment...)
	}
	return result, nil
}

// NewAigerModel attaches safety section metadata to a concrete graph. The
// caller transfers exclusive mutation ownership of graph to the model. A nil
// initial slice means every latch is initialized to zero; other slices must
// specify exactly one LatchInit per register and are copied.
func NewAigerModel(graph *AigManager, outputs, bad, constraints int, initial []LatchInit) (*AigerModel, error) {
	if graph == nil {
		return nil, ErrDimension
	}
	if initial == nil {
		initial = make([]LatchInit, graph.RegisterCount())
	}
	model := &AigerModel{graph: graph, outputs: outputs, bad: bad, constraints: constraints, initial: slices.Clone(initial), symbols: newAigerSymbols()}
	if err := model.validate(false); err != nil {
		return nil, err
	}
	if err := graph.SetConstraintCount(constraints); err != nil {
		return nil, err
	}
	return model, nil
}

// SetSymbol updates a terminal name without changing its insertion position.
func (m *AigerModel) SetSymbol(symbol AigerSymbol, name string) error {
	bound := m.symbolBound(symbol.Kind)
	if bound < 0 || uint64(symbol.Index) >= uint64(bound) || name == "" || bytes.ContainsAny([]byte(name), "\x00\r\n") {
		return ErrDimension
	}
	m.symbols.Put(symbol, name)
	return nil
}

// SetComment copies arbitrary comment bytes. Nil removes the comment section;
// an empty nonnil slice requests an empty section.
func (m *AigerModel) SetComment(comment []byte) { m.comment = slices.Clone(comment) }
