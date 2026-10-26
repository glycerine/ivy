package abc

// Minimal AIG, LUT, operator, and NDR representations from src/aig/miniaig.

import (
	"encoding/binary"
	"math"
	"slices"
	"strconv"
	"strings"
)

const (
	MiniAIGNull  = math.MaxInt32
	MiniLUTNull  = math.MaxInt32
	MiniLUTNull2 = math.MaxInt32 - 1
)

type MiniNetworkLimits struct {
	MaxNodes   int
	MaxRecords int
	Stop       func() bool
}

func normalizeMiniLimits(l MiniNetworkLimits) (MiniNetworkLimits, error) {
	if l.MaxNodes < 0 || l.MaxRecords < 0 {
		return l, ErrDimension
	}
	if l.MaxNodes == 0 {
		l.MaxNodes = 1 << 24
	}
	if l.MaxRecords == 0 {
		l.MaxRecords = 1 << 26
	}
	return l, nil
}

type AbcOperator int

const (
	AbcOperNone AbcOperator = iota
	AbcOperPI
	AbcOperPO
	AbcOperCI
	AbcOperCO
	AbcOperFON
	AbcOperBox
	AbcOperConstF
	AbcOperConstT
	AbcOperConstX
	AbcOperConstZ
	AbcOperBitBuf
	AbcOperBitInv
	AbcOperBitAnd
	AbcOperBitNand
	AbcOperBitOr
	AbcOperBitNor
	AbcOperBitXor
	AbcOperBitNxor
	AbcOperBitSharp
	AbcOperBitSharpL
	AbcOperBitMux
	AbcOperBitMaj
	AbcOperABC
	AbcOperBA
	AbcOperBO
	AbcOperBX
	AbcOperBN
	AbcOperBAO
	AbcOperBOA
	AbcOperRedAnd
	AbcOperRedNand
	AbcOperRedOr
	AbcOperRedNor
	AbcOperRedXor
	AbcOperRedNxor
	AbcOperLogicNot
	AbcOperLogicAnd
	AbcOperLogicNand
	AbcOperLogicOr
	AbcOperLogicNor
	AbcOperLogicXor
	AbcOperLogicXnor
	AbcOperSelNMux
	AbcOperSelSel
	AbcOperSelPSel
	AbcOperSelEnc
	AbcOperSelPEnc
	AbcOperSelDec
	AbcOperSelEDec
	AbcOperAriAdd
	AbcOperAriSub
	AbcOperAriMul
	AbcOperAriSMul
	AbcOperAriDiv
	AbcOperAriMod
	AbcOperAriRem
	AbcOperAriPow
	AbcOperAriMin
	AbcOperAriSqrt
	AbcOperAriAbs
	AbcOperCompSLess
	AbcOperCompLess
	AbcOperCompLessEqual
	AbcOperCompMoreEqual
	AbcOperCompMore
	AbcOperCompEqual
	AbcOperCompNotEqual
	AbcOperShiftL
	AbcOperShiftR
	AbcOperShiftLA
	AbcOperShiftRA
	AbcOperShiftRotL
	AbcOperShiftRotR
	AbcOperNode
	AbcOperLUT
	AbcOperGate
	AbcOperTable
	AbcOperTri
	AbcOperRAM
	AbcOperRAMR
	AbcOperRAMW
	AbcOperRAMWC
	AbcOperRAML
	AbcOperRAMS
	AbcOperRAMBox
	AbcOperLatch
	AbcOperLatchRS
	AbcOperDFF
	AbcOperDFFRSE
	AbcOperDFFLast
	AbcOperSlice
	AbcOperConcat
	AbcOperZeroPad
	AbcOperSignExt
	AbcOperLogicImpl
	AbcOperAriSquare
	AbcOperConst
	AbcOperAriAddSub
	AbcOperLast
)

func AbcOperatorName(op AbcOperator) (string, bool) {
	names := map[AbcOperator]string{
		AbcOperPI: "pi", AbcOperPO: "po", AbcOperCI: "ci", AbcOperCO: "co", AbcOperFON: "fon", AbcOperBox: "box",
		AbcOperBitBuf: "buf", AbcOperBitInv: "~", AbcOperBitAnd: "&", AbcOperBitNand: "~&", AbcOperBitOr: "|", AbcOperBitNor: "~|", AbcOperBitXor: "^", AbcOperBitNxor: "~^", AbcOperBitMux: "mux", AbcOperBitMaj: "maj",
		AbcOperRedAnd: "&", AbcOperRedNand: "~&", AbcOperRedOr: "|", AbcOperRedNor: "~|", AbcOperRedXor: "^", AbcOperRedNxor: "~^",
		AbcOperLogicNot: "!", AbcOperLogicImpl: "=>", AbcOperLogicAnd: "&&", AbcOperLogicOr: "||", AbcOperLogicXor: "^^",
		AbcOperAriAdd: "+", AbcOperAriSub: "-", AbcOperAriMul: "*", AbcOperAriSMul: "*", AbcOperAriDiv: "/", AbcOperAriRem: "%", AbcOperAriMod: "mod", AbcOperAriPow: "**", AbcOperAriMin: "-", AbcOperAriSqrt: "sqrt", AbcOperAriSquare: "squar",
		AbcOperCompEqual: "==", AbcOperCompNotEqual: "!=", AbcOperCompLess: "<", AbcOperCompMore: ">", AbcOperCompLessEqual: "<=", AbcOperCompMoreEqual: ">=",
		AbcOperShiftL: "<<", AbcOperShiftR: ">>", AbcOperShiftLA: "<<<", AbcOperShiftRA: ">>>", AbcOperShiftRotL: "rotL", AbcOperShiftRotR: "rotR",
		AbcOperDFFRSE: "DFFRSE", AbcOperSlice: "[:]", AbcOperConcat: "{}", AbcOperZeroPad: "zPad", AbcOperSignExt: "sExt", AbcOperConst: "const", AbcOperTable: "table", AbcOperLUT: "lut",
	}
	name, ok := names[op]
	return name, ok
}

type AIGLiteral int
type AIGNodeID int
type miniAIGNode struct{ fanin0, fanin1 int }

func AIGVarLiteral(variable AIGNodeID, complement bool) AIGLiteral {
	value := int(variable) << 1
	if complement {
		value++
	}
	return AIGLiteral(value)
}
func (l AIGLiteral) Variable() AIGNodeID { return AIGNodeID(int(l) >> 1) }
func (l AIGLiteral) Complemented() bool  { return int(l)&1 != 0 }
func (l AIGLiteral) Not() AIGLiteral     { return l ^ 1 }
func (l AIGLiteral) NotIf(condition bool) AIGLiteral {
	if condition {
		return l.Not()
	}
	return l
}
func (l AIGLiteral) Regular() AIGLiteral { return l &^ 1 }

type MiniAIG struct {
	nodes     []miniAIGNode
	registers int
	limits    MiniNetworkLimits
	closed    bool
}

func NewMiniAIG(limits MiniNetworkLimits) (*MiniAIG, error) {
	limits, err := normalizeMiniLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.MaxNodes < 1 {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &MiniAIG{nodes: []miniAIGNode{{MiniAIGNull, MiniAIGNull}}, limits: limits}, nil
}
func (a *MiniAIG) check() error {
	if a == nil {
		return ErrDimension
	}
	if a.closed || len(a.nodes) == 0 || len(a.nodes) > a.limits.MaxNodes || a.registers < 0 {
		return ErrInvariant
	}
	return nil
}
func (a *MiniAIG) NodeCount() int {
	if a == nil {
		return 0
	}
	return len(a.nodes)
}
func (a *MiniAIG) RegisterCount() int {
	if a == nil {
		return 0
	}
	return a.registers
}
func (a *MiniAIG) SetRegisterCount(n int) error {
	if err := a.check(); err != nil {
		return err
	}
	if n < 0 {
		return ErrDimension
	}
	if a.limits.Stop != nil && a.limits.Stop() {
		return ErrCancelled
	}
	a.registers = n
	return nil
}
func (a *MiniAIG) Fanins(id AIGNodeID) (AIGLiteral, AIGLiteral, error) {
	if err := a.check(); err != nil {
		return 0, 0, err
	}
	i := int(id)
	if i < 0 || i >= len(a.nodes) {
		return 0, 0, ErrDimension
	}
	n := a.nodes[i]
	return AIGLiteral(n.fanin0), AIGLiteral(n.fanin1), nil
}
func (a *MiniAIG) IsConst(id AIGNodeID) bool { return id == 0 }
func (a *MiniAIG) IsPI(id AIGNodeID) bool {
	i := int(id)
	return i > 0 && i < len(a.nodes) && a.nodes[i].fanin0 == MiniAIGNull
}
func (a *MiniAIG) IsPO(id AIGNodeID) bool {
	i := int(id)
	return i > 0 && i < len(a.nodes) && a.nodes[i].fanin0 != MiniAIGNull && a.nodes[i].fanin1 == MiniAIGNull
}
func (a *MiniAIG) IsAND(id AIGNodeID) bool {
	i := int(id)
	return i > 0 && i < len(a.nodes) && a.nodes[i].fanin0 != MiniAIGNull && a.nodes[i].fanin1 != MiniAIGNull
}
func (a *MiniAIG) IDs(kind byte) []AIGNodeID {
	result := make([]AIGNodeID, 0)
	for i := 1; i < len(a.nodes); i++ {
		id := AIGNodeID(i)
		if kind == 'i' && a.IsPI(id) || kind == 'o' && a.IsPO(id) || kind == 'a' && a.IsAND(id) {
			result = append(result, id)
		}
	}
	return result
}
func (a *MiniAIG) validateLiteral(l AIGLiteral) bool { return l >= 0 && int(l) < 2*len(a.nodes) }
func (a *MiniAIG) reserve(count int) error {
	if err := a.check(); err != nil {
		return err
	}
	if count < 0 || len(a.nodes) > a.limits.MaxNodes-count {
		return ErrLimit
	}
	if a.limits.Stop != nil && a.limits.Stop() {
		return ErrCancelled
	}
	return nil
}
func (a *MiniAIG) CreatePI() (AIGLiteral, error) {
	if err := a.reserve(1); err != nil {
		return 0, err
	}
	lit := AIGLiteral(2 * len(a.nodes))
	a.nodes = append(a.nodes, miniAIGNode{MiniAIGNull, MiniAIGNull})
	return lit, nil
}
func (a *MiniAIG) CreatePO(l AIGLiteral) (AIGLiteral, error) {
	if !a.validateLiteral(l) {
		return 0, ErrDimension
	}
	if err := a.reserve(1); err != nil {
		return 0, err
	}
	out := AIGLiteral(2 * len(a.nodes))
	a.nodes = append(a.nodes, miniAIGNode{int(l), MiniAIGNull})
	return out, nil
}
func (a *MiniAIG) appendAND(x, y AIGLiteral) AIGLiteral {
	lit := AIGLiteral(2 * len(a.nodes))
	a.nodes = append(a.nodes, miniAIGNode{int(x), int(y)})
	return lit
}
func (a *MiniAIG) AND(x, y AIGLiteral) (AIGLiteral, error) {
	if !a.validateLiteral(x) || !a.validateLiteral(y) {
		return 0, ErrDimension
	}
	if err := a.reserve(1); err != nil {
		return 0, err
	}
	return a.appendAND(x, y), nil
}
func (a *MiniAIG) OR(x, y AIGLiteral) (AIGLiteral, error) {
	if !a.validateLiteral(x) || !a.validateLiteral(y) {
		return 0, ErrDimension
	}
	if err := a.reserve(1); err != nil {
		return 0, err
	}
	return a.appendAND(x.Not(), y.Not()).Not(), nil
}
func (a *MiniAIG) MUX(c, one, zero AIGLiteral) (AIGLiteral, error) {
	if !a.validateLiteral(c) || !a.validateLiteral(one) || !a.validateLiteral(zero) {
		return 0, ErrDimension
	}
	if err := a.reserve(3); err != nil {
		return 0, err
	}
	x := a.appendAND(c, one)
	y := a.appendAND(c.Not(), zero)
	return a.appendAND(x.Not(), y.Not()).Not(), nil
}
func (a *MiniAIG) XOR(x, y AIGLiteral) (AIGLiteral, error) { return a.MUX(x, y.Not(), y) }
func (a *MiniAIG) ANDProp(x, y AIGLiteral) (AIGLiteral, error) {
	if !a.validateLiteral(x) || !a.validateLiteral(y) {
		return 0, ErrDimension
	}
	if x < 2 {
		if x == 1 {
			return y, nil
		}
		return 0, nil
	}
	if y < 2 {
		if y == 1 {
			return x, nil
		}
		return 0, nil
	}
	if x == y {
		return x, nil
	}
	if x == y.Not() {
		return 0, nil
	}
	return a.AND(x, y)
}
func (a *MiniAIG) MUXProp(c, one, zero AIGLiteral) (AIGLiteral, error) {
	clone := a.Clone()
	x, e := clone.ANDProp(c.Not(), zero)
	if e != nil {
		return 0, e
	}
	y, e := clone.ANDProp(c, one)
	if e != nil {
		return 0, e
	}
	z, e := clone.ANDProp(x.Not(), y.Not())
	if e != nil {
		return 0, e
	}
	a.nodes = clone.nodes
	return z.Not(), nil
}
func (a *MiniAIG) Clone() *MiniAIG {
	if a == nil {
		return nil
	}
	result := *a
	result.nodes = slices.Clone(a.nodes)
	return &result
}

var miniTruth5 = [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}

func MiniAIGTruthHasVariable(t uint32, v int) bool {
	return v >= 0 && v < 5 && ((t<<uint(1<<v))&miniTruth5[v]) != (t&miniTruth5[v])
}
func MiniAIGTruthCofactor(t uint32, v int, one bool) (uint32, error) {
	if v < 0 || v >= 5 {
		return 0, ErrDimension
	}
	m := miniTruth5[v]
	if one {
		return (t & m) | ((t & m) >> uint(1<<v)), nil
	}
	return (t &^ m) | ((t &^ m) << uint(1<<v)), nil
}
func (a *MiniAIG) Truth(vars []AIGLiteral, truth uint32) (AIGLiteral, error) {
	if len(vars) > 5 {
		return 0, ErrDimension
	}
	for _, v := range vars {
		if !a.validateLiteral(v) {
			return 0, ErrDimension
		}
	}
	work := a.Clone()
	var build func(int, uint32) (AIGLiteral, error)
	build = func(n int, t uint32) (AIGLiteral, error) {
		if t == 0 {
			return 0, nil
		}
		if t == math.MaxUint32 {
			return 1, nil
		}
		v := n - 1
		for v >= 0 && !MiniAIGTruthHasVariable(t, v) {
			v--
		}
		if v < 0 {
			return 0, ErrInvariant
		}
		t0, _ := MiniAIGTruthCofactor(t, v, false)
		t1, _ := MiniAIGTruthCofactor(t, v, true)
		l0, e := build(v, t0)
		if e != nil {
			return 0, e
		}
		l1, e := build(v, t1)
		if e != nil {
			return 0, e
		}
		return work.MUXProp(vars[v], l1, l0)
	}
	lit, e := build(len(vars), truth)
	if e != nil {
		return 0, e
	}
	a.nodes = work.nodes
	return lit, nil
}
func (a *MiniAIG) Check() error {
	if err := a.check(); err != nil {
		return err
	}
	if a.nodes[0].fanin0 != MiniAIGNull || a.nodes[0].fanin1 != MiniAIGNull {
		return ErrInvariant
	}
	for i, n := range a.nodes {
		if i == 0 {
			continue
		}
		if n.fanin0 == MiniAIGNull && n.fanin1 != MiniAIGNull {
			return ErrInvariant
		}
		for _, f := range []int{n.fanin0, n.fanin1} {
			if f != MiniAIGNull && (f < 0 || f >= 2*i) {
				return ErrInvariant
			}
		}
	}
	return nil
}
func (a *MiniAIG) Counts() (pis, pos, ands int) {
	for i := 1; i < len(a.nodes); i++ {
		id := AIGNodeID(i)
		if a.IsPI(id) {
			pis++
		} else if a.IsPO(id) {
			pos++
		} else if a.IsAND(id) {
			ands++
		}
	}
	return
}
func (a *MiniAIG) Stats() string {
	pis, pos, ands := a.Counts()
	return "PI = " + strconv.Itoa(pis) + ". PO = " + strconv.Itoa(pos) + ". Node = " + strconv.Itoa(ands) + "."
}
func (a *MiniAIG) Verilog(module string) (string, error) {
	if err := a.Check(); err != nil {
		return "", err
	}
	if module == "" {
		return "", ErrDimension
	}
	pis := a.IDs('i')
	pos := a.IDs('o')
	var b strings.Builder
	b.WriteString("module " + module + " (\n")
	for i, id := range pis {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("i" + strconv.Itoa(int(id)))
	}
	for _, id := range pos {
		if len(pis) > 0 || id != pos[0] {
			b.WriteString(", ")
		}
		b.WriteString("o" + strconv.Itoa(int(id)))
	}
	b.WriteString("\n);\n")
	if len(pis) > 0 {
		b.WriteString("  input ")
		for i, id := range pis {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString("i" + strconv.Itoa(int(id)))
		}
		b.WriteString(";\n")
	}
	if len(pos) > 0 {
		b.WriteString("  output ")
		for i, id := range pos {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString("o" + strconv.Itoa(int(id)))
		}
		b.WriteString(";\n")
	}
	name := func(l int) string {
		prefix := "n"
		v := l >> 1
		if v == 0 {
			if l&1 != 0 {
				return "1'b1"
			}
			return "1'b0"
		}
		if a.IsPI(AIGNodeID(v)) {
			prefix = "i"
		}
		s := prefix + strconv.Itoa(v)
		if l&1 != 0 {
			return "~" + s
		}
		return s
	}
	for _, id := range a.IDs('a') {
		n := a.nodes[id]
		b.WriteString("  wire n" + strconv.Itoa(int(id)) + " = " + name(n.fanin0) + " & " + name(n.fanin1) + ";\n")
	}
	for _, id := range pos {
		b.WriteString("  assign o" + strconv.Itoa(int(id)) + " = " + name(a.nodes[id].fanin0) + ";\n")
	}
	b.WriteString("endmodule\n")
	return b.String(), nil
}
func (a *MiniAIG) Evaluate(inputs []bool) ([]bool, error) {
	if err := a.Check(); err != nil {
		return nil, err
	}
	pis, _, _ := a.Counts()
	if len(inputs) != pis {
		return nil, ErrDimension
	}
	values := make([]bool, len(a.nodes))
	pi := 0
	outs := make([]bool, 0)
	eval := func(l int) bool {
		v := values[l>>1]
		if l&1 != 0 {
			return !v
		}
		return v
	}
	for i := 1; i < len(a.nodes); i++ {
		n := a.nodes[i]
		if n.fanin0 == MiniAIGNull {
			values[i] = inputs[pi]
			pi++
		} else if n.fanin1 == MiniAIGNull {
			outs = append(outs, eval(n.fanin0))
		} else {
			values[i] = eval(n.fanin0) && eval(n.fanin1)
		}
	}
	return outs, nil
}
func (a *MiniAIG) MarshalBinary() ([]byte, error) {
	if err := a.Check(); err != nil {
		return nil, err
	}
	b := make([]byte, 8+8*len(a.nodes))
	binary.LittleEndian.PutUint32(b, uint32(2*len(a.nodes)))
	binary.LittleEndian.PutUint32(b[4:], uint32(a.registers))
	for i, n := range a.nodes {
		binary.LittleEndian.PutUint32(b[8+8*i:], uint32(n.fanin0))
		binary.LittleEndian.PutUint32(b[12+8*i:], uint32(n.fanin1))
	}
	return b, nil
}
func ParseMiniAIG(b []byte, limits MiniNetworkLimits) (*MiniAIG, error) {
	if len(b) < 8 || len(b)%8 != 0 {
		return nil, ErrDimension
	}
	size := int(binary.LittleEndian.Uint32(b))
	if size < 2 || size&1 != 0 || len(b) != 8+4*size {
		return nil, ErrDimension
	}
	a, e := NewMiniAIG(limits)
	if e != nil {
		return nil, e
	}
	if size/2 > a.limits.MaxNodes {
		return nil, ErrLimit
	}
	a.nodes = make([]miniAIGNode, size/2)
	a.registers = int(int32(binary.LittleEndian.Uint32(b[4:])))
	for i := range a.nodes {
		a.nodes[i] = miniAIGNode{int(int32(binary.LittleEndian.Uint32(b[8+8*i:]))), int(int32(binary.LittleEndian.Uint32(b[12+8*i:])))}
	}
	if e = a.Check(); e != nil {
		return nil, e
	}
	return a, nil
}
func (a *MiniAIG) Close() {
	if a != nil {
		a.nodes = nil
		a.closed = true
	}
}

type LUTNodeID int
type miniLUTKind uint8

const (
	lutConst miniLUTKind = iota
	lutPI
	lutPO
	lutNode
)

type miniLUTNode struct {
	kind   miniLUTKind
	fanins []LUTNodeID
	truth  []uint32
}
type MiniLUT struct {
	size      int
	nodes     []miniLUTNode
	registers int
	limits    MiniNetworkLimits
	closed    bool
}

func MiniLUTWordCount(size int) (int, error) {
	if size < 2 || size > 16 {
		return 0, ErrDimension
	}
	if size > 5 {
		return 1 << (size - 5), nil
	}
	return 1, nil
}
func NewMiniLUT(size int, limits MiniNetworkLimits) (*MiniLUT, error) {
	words, err := MiniLUTWordCount(size)
	if err != nil {
		return nil, err
	}
	limits, err = normalizeMiniLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.MaxNodes < 2 {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	ones := make([]uint32, words)
	for i := range ones {
		ones[i] = math.MaxUint32
	}
	return &MiniLUT{size: size, limits: limits, nodes: []miniLUTNode{
		{kind: lutConst, truth: make([]uint32, words)},
		{kind: lutConst, truth: ones},
	}}, nil
}
func (l *MiniLUT) check() error {
	if l == nil {
		return ErrDimension
	}
	if l.closed || l.size < 2 || l.size > 16 || len(l.nodes) < 2 || len(l.nodes) > l.limits.MaxNodes {
		return ErrInvariant
	}
	return nil
}
func (l *MiniLUT) Size() int {
	if l == nil {
		return 0
	}
	return l.size
}
func (l *MiniLUT) NodeCount() int {
	if l == nil {
		return 0
	}
	return len(l.nodes)
}
func (l *MiniLUT) RegisterCount() int {
	if l == nil {
		return 0
	}
	return l.registers
}
func (l *MiniLUT) SetRegisterCount(n int) error {
	if e := l.check(); e != nil {
		return e
	}
	if n < 0 {
		return ErrDimension
	}
	if l.limits.Stop != nil && l.limits.Stop() {
		return ErrCancelled
	}
	l.registers = n
	return nil
}
func (l *MiniLUT) reserve() error {
	if e := l.check(); e != nil {
		return e
	}
	if len(l.nodes) >= l.limits.MaxNodes {
		return ErrLimit
	}
	if l.limits.Stop != nil && l.limits.Stop() {
		return ErrCancelled
	}
	return nil
}
func (l *MiniLUT) CreatePI() (LUTNodeID, error) {
	if e := l.reserve(); e != nil {
		return 0, e
	}
	id := LUTNodeID(len(l.nodes))
	l.nodes = append(l.nodes, miniLUTNode{kind: lutPI})
	return id, nil
}
func (l *MiniLUT) CreatePO(f LUTNodeID) (LUTNodeID, error) {
	if f < 0 || int(f) >= len(l.nodes) {
		return 0, ErrDimension
	}
	if e := l.reserve(); e != nil {
		return 0, e
	}
	id := LUTNodeID(len(l.nodes))
	l.nodes = append(l.nodes, miniLUTNode{kind: lutPO, fanins: []LUTNodeID{f}})
	return id, nil
}
func (l *MiniLUT) CreateNode(f []LUTNodeID, truth []uint32) (LUTNodeID, error) {
	if len(f) > l.size {
		return 0, ErrDimension
	}
	for _, v := range f {
		if v < 0 || int(v) >= len(l.nodes) {
			return 0, ErrDimension
		}
	}
	words, _ := MiniLUTWordCount(l.size)
	if len(truth) != words {
		return 0, ErrDimension
	}
	if e := l.reserve(); e != nil {
		return 0, e
	}
	id := LUTNodeID(len(l.nodes))
	kind := lutNode
	if len(f) == 0 {
		// Mini_LutCreateNode with no fanins has the same all-NULL encoding as
		// Mini_LutCreatePi and is classified as a PI by the source macros.
		kind = lutPI
	}
	l.nodes = append(l.nodes, miniLUTNode{kind: kind, fanins: slices.Clone(f), truth: slices.Clone(truth)})
	return id, nil
}
func (l *MiniLUT) IsConst(id LUTNodeID) bool { return id == 0 || id == 1 }
func (l *MiniLUT) IsPI(id LUTNodeID) bool {
	i := int(id)
	return i > 1 && i < len(l.nodes) && l.nodes[i].kind == lutPI
}
func (l *MiniLUT) IsPO(id LUTNodeID) bool {
	i := int(id)
	return i > 1 && i < len(l.nodes) && l.nodes[i].kind == lutPO
}
func (l *MiniLUT) IsNode(id LUTNodeID) bool {
	i := int(id)
	return i > 1 && i < len(l.nodes) && l.nodes[i].kind == lutNode
}
func (l *MiniLUT) Fanins(id LUTNodeID) ([]LUTNodeID, error) {
	if e := l.check(); e != nil {
		return nil, e
	}
	i := int(id)
	if i < 0 || i >= len(l.nodes) {
		return nil, ErrDimension
	}
	return slices.Clone(l.nodes[i].fanins), nil
}
func (l *MiniLUT) TruthWords(id LUTNodeID) ([]uint32, error) {
	if e := l.check(); e != nil {
		return nil, e
	}
	i := int(id)
	if i < 0 || i >= len(l.nodes) {
		return nil, ErrDimension
	}
	return slices.Clone(l.nodes[i].truth), nil
}
func (l *MiniLUT) Check() error {
	if e := l.check(); e != nil {
		return e
	}
	words, _ := MiniLUTWordCount(l.size)
	for i, n := range l.nodes {
		if len(n.truth) != 0 && len(n.truth) != words {
			return ErrInvariant
		}
		for _, f := range n.fanins {
			if f < 0 || int(f) >= i {
				return ErrInvariant
			}
		}
	}
	return nil
}
func (l *MiniLUT) Counts() (pis, pos, nodes int) {
	for i := 2; i < len(l.nodes); i++ {
		switch l.nodes[i].kind {
		case lutPI:
			pis++
		case lutPO:
			pos++
		case lutNode:
			nodes++
		}
	}
	return
}
func (l *MiniLUT) Stats() string {
	pis, pos, nodes := l.Counts()
	return "PI = " + strconv.Itoa(pis) + ". PO = " + strconv.Itoa(pos) + ". LUT = " + strconv.Itoa(nodes) + "."
}
func (l *MiniLUT) Evaluate(inputs []bool) ([]bool, error) {
	if e := l.Check(); e != nil {
		return nil, e
	}
	pis, _, _ := l.Counts()
	if len(inputs) != pis {
		return nil, ErrDimension
	}
	vals := make([]bool, len(l.nodes))
	vals[1] = true
	pi := 0
	outs := make([]bool, 0)
	for i := 2; i < len(l.nodes); i++ {
		n := l.nodes[i]
		switch n.kind {
		case lutPI:
			vals[i] = inputs[pi]
			pi++
		case lutPO:
			outs = append(outs, vals[n.fanins[0]])
		case lutNode:
			index := 0
			for k, f := range n.fanins {
				if vals[f] {
					index |= 1 << k
				}
			}
			vals[i] = (n.truth[index>>5]>>uint(index&31))&1 != 0
		}
	}
	return outs, nil
}
func (l *MiniLUT) MarshalBinary() ([]byte, error) {
	if e := l.Check(); e != nil {
		return nil, e
	}
	words, _ := MiniLUTWordCount(l.size)
	stride := l.size + words
	b := make([]byte, 12+4*stride*len(l.nodes))
	binary.LittleEndian.PutUint32(b, uint32(len(l.nodes)))
	binary.LittleEndian.PutUint32(b[4:], uint32(l.registers))
	binary.LittleEndian.PutUint32(b[8:], uint32(l.size))
	faninOffset := 12
	truthOffset := 12 + 4*l.size*len(l.nodes)
	for _, n := range l.nodes {
		fans := make([]int, l.size)
		for i := range fans {
			fans[i] = MiniLUTNull
		}
		copyFan := n.fanins
		copy(fans, func() []int {
			x := make([]int, len(copyFan))
			for i, v := range copyFan {
				x[i] = int(v)
			}
			return x
		}())
		if n.kind == lutPO && l.size > 1 {
			fans[1] = MiniLUTNull2
		}
		for _, v := range fans {
			binary.LittleEndian.PutUint32(b[faninOffset:], uint32(v))
			faninOffset += 4
		}
		truth := n.truth
		if len(truth) == 0 {
			truth = make([]uint32, words)
		}
		for _, v := range truth {
			binary.LittleEndian.PutUint32(b[truthOffset:], v)
			truthOffset += 4
		}
	}
	return b, nil
}
func ParseMiniLUT(b []byte, limits MiniNetworkLimits) (*MiniLUT, error) {
	if len(b) < 12 {
		return nil, ErrDimension
	}
	count := int(binary.LittleEndian.Uint32(b))
	size := int(binary.LittleEndian.Uint32(b[8:]))
	words, e := MiniLUTWordCount(size)
	if e != nil {
		return nil, e
	}
	if count < 2 || len(b) != 12+4*count*(size+words) {
		return nil, ErrDimension
	}
	l, e := NewMiniLUT(size, limits)
	if e != nil {
		return nil, e
	}
	if count > l.limits.MaxNodes {
		return nil, ErrLimit
	}
	l.nodes = make([]miniLUTNode, count)
	l.registers = int(int32(binary.LittleEndian.Uint32(b[4:])))
	faninOffset := 12
	truthOffset := 12 + 4*size*count
	for i := 0; i < count; i++ {
		raw := make([]int, size)
		for k := range raw {
			raw[k] = int(int32(binary.LittleEndian.Uint32(b[faninOffset:])))
			faninOffset += 4
		}
		truth := make([]uint32, words)
		for k := range truth {
			truth[k] = binary.LittleEndian.Uint32(b[truthOffset:])
			truthOffset += 4
		}
		kind := lutNode
		fans := make([]LUTNodeID, 0)
		if i < 2 {
			for _, value := range raw {
				if value != MiniLUTNull {
					return nil, ErrInvariant
				}
			}
			want := uint32(0)
			if i == 1 {
				want = math.MaxUint32
			}
			for _, value := range truth {
				if value != want {
					return nil, ErrInvariant
				}
			}
			kind = lutConst
		} else if raw[0] == MiniLUTNull {
			for _, value := range raw[1:] {
				if value != MiniLUTNull {
					return nil, ErrInvariant
				}
			}
			kind = lutPI
		} else if size > 1 && raw[1] == MiniLUTNull2 {
			for _, value := range raw[2:] {
				if value != MiniLUTNull {
					return nil, ErrInvariant
				}
			}
			kind = lutPO
			fans = []LUTNodeID{LUTNodeID(raw[0])}
		} else {
			ended := false
			for _, v := range raw {
				if v >= MiniLUTNull2 {
					if v != MiniLUTNull {
						return nil, ErrInvariant
					}
					ended = true
					continue
				}
				if ended {
					return nil, ErrInvariant
				}
				fans = append(fans, LUTNodeID(v))
			}
		}
		l.nodes[i] = miniLUTNode{kind: kind, fanins: fans, truth: truth}
	}
	if e = l.Check(); e != nil {
		return nil, e
	}
	return l, nil
}
func (l *MiniLUT) Close() {
	if l != nil {
		l.nodes = nil
		l.closed = true
	}
}

// MiniAIGToLUT converts each AND into a two-input LUT. Complemented edges are
// absorbed into the LUT truth table, so the result remains an edge-regular
// MiniLUT network.
func MiniAIGToLUT(a *MiniAIG, limits MiniNetworkLimits) (*MiniLUT, error) {
	if err := a.Check(); err != nil {
		return nil, err
	}
	lut, err := NewMiniLUT(2, limits)
	if err != nil {
		return nil, err
	}
	mapping := make([]LUTNodeID, len(a.nodes))
	mapping[0] = 0
	for i := 1; i < len(a.nodes); i++ {
		id := AIGNodeID(i)
		n := a.nodes[i]
		if a.IsPI(id) {
			mapping[i], err = lut.CreatePI()
		} else if a.IsAND(id) {
			f0, f1 := AIGLiteral(n.fanin0), AIGLiteral(n.fanin1)
			var truth uint32
			for assignment := 0; assignment < 32; assignment++ {
				x := assignment&1 != 0
				y := assignment&2 != 0
				if f0.Complemented() {
					x = !x
				}
				if f1.Complemented() {
					y = !y
				}
				if x && y {
					truth |= 1 << assignment
				}
			}
			mapping[i], err = lut.CreateNode([]LUTNodeID{mapping[f0.Variable()], mapping[f1.Variable()]}, []uint32{truth})
		} else {
			f := AIGLiteral(n.fanin0)
			source := mapping[f.Variable()]
			if f.Complemented() {
				source, err = lut.CreateNode([]LUTNodeID{source}, []uint32{0x55555555})
			}
			if err == nil {
				_, err = lut.CreatePO(source)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	if err = lut.SetRegisterCount(a.registers); err != nil {
		return nil, err
	}
	return lut, nil
}

// MiniLUTToAIG synthesizes LUT truth tables with at most five fanins using
// Shannon decomposition, preserving node and output order.
func MiniLUTToAIG(lut *MiniLUT, limits MiniNetworkLimits) (*MiniAIG, error) {
	if err := lut.Check(); err != nil {
		return nil, err
	}
	aig, err := NewMiniAIG(limits)
	if err != nil {
		return nil, err
	}
	mapping := make([]AIGLiteral, len(lut.nodes))
	mapping[1] = 1
	for i := 2; i < len(lut.nodes); i++ {
		n := lut.nodes[i]
		switch n.kind {
		case lutPI:
			mapping[i], err = aig.CreatePI()
		case lutPO:
			_, err = aig.CreatePO(mapping[n.fanins[0]])
		case lutNode:
			if len(n.fanins) > 5 {
				return nil, ErrDimension
			}
			vars := make([]AIGLiteral, len(n.fanins))
			for k, fanin := range n.fanins {
				vars[k] = mapping[fanin]
			}
			truth := n.truth[0]
			if len(n.fanins) < 5 {
				bits := 1 << uint(len(n.fanins))
				mask := uint32(1<<uint(bits)) - 1
				pattern := truth & mask
				truth = 0
				for shift := 0; shift < 32; shift += bits {
					truth |= pattern << uint(shift)
				}
			}
			mapping[i], err = aig.Truth(vars, truth)
		default:
			return nil, ErrInvariant
		}
		if err != nil {
			return nil, err
		}
	}
	if err = aig.SetRegisterCount(lut.registers); err != nil {
		return nil, err
	}
	return aig, nil
}

type NDRRecordType uint8

const (
	NDRNone NDRRecordType = iota
	NDRDesign
	NDRModule
	NDRObject
	NDRInput
	NDROutput
	NDROperatorType
	NDRName
	NDRRange
	NDRFunction
	NDRTarget
	NDRUnknown
)

type NDRObjectSpec struct {
	Operator     AbcOperator
	InstanceName int
	RangeLeft    int
	RangeRight   int
	Signed       bool
	Inputs       []int
	Outputs      []int
	Function     []byte
}
type NDRDesignData struct {
	head   []NDRRecordType
	body   []uint32
	limits MiniNetworkLimits
	closed bool
}

func NewNDRDesign(name int, limits MiniNetworkLimits) (*NDRDesignData, error) {
	if name <= 0 {
		return nil, ErrDimension
	}
	limits, e := normalizeMiniLimits(limits)
	if e != nil {
		return nil, e
	}
	if limits.MaxRecords < 2 {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	d := &NDRDesignData{limits: limits}
	d.push(NDRDesign, 0)
	d.push(NDRName, name)
	d.body[0] = 2
	return d, nil
}
func (d *NDRDesignData) check() error {
	if d == nil {
		return ErrDimension
	}
	if d.closed || len(d.head) != len(d.body) || len(d.head) < 2 || len(d.head) > d.limits.MaxRecords || d.head[0] != NDRDesign || int(d.body[0]) != len(d.body) {
		return ErrInvariant
	}
	return nil
}
func (d *NDRDesignData) push(t NDRRecordType, v int) {
	d.head = append(d.head, t)
	d.body = append(d.body, uint32(v))
}
func (d *NDRDesignData) reserve(n int) error {
	if e := d.check(); e != nil {
		return e
	}
	if n < 0 || len(d.head) > d.limits.MaxRecords-n {
		return ErrLimit
	}
	if d.limits.Stop != nil && d.limits.Stop() {
		return ErrCancelled
	}
	return nil
}
func (d *NDRDesignData) RecordCount() int {
	if d == nil {
		return 0
	}
	return len(d.head)
}
func (d *NDRDesignData) DataType(i int) (NDRRecordType, error) {
	if e := d.check(); e != nil {
		return 0, e
	}
	if i < 0 || i >= len(d.head) || d.head[i] == NDRNone {
		return 0, ErrDimension
	}
	return d.head[i], nil
}
func (d *NDRDesignData) DataSize(i int) (int, error) {
	t, e := d.DataType(i)
	if e != nil {
		return 0, e
	}
	if t > NDRObject {
		return 1, nil
	}
	return int(d.body[i]), nil
}
func (d *NDRDesignData) DataEntry(i int) (int, error) {
	if _, e := d.DataType(i); e != nil {
		return 0, e
	}
	return int(int32(d.body[i])), nil
}
func (d *NDRDesignData) AddModule(name int) (int, error) {
	if name <= 0 {
		return 0, ErrDimension
	}
	if e := d.reserve(2); e != nil {
		return 0, e
	}
	m := len(d.head)
	d.push(NDRModule, 2)
	d.push(NDRName, name)
	d.body[0] += 2
	return m + 256, nil
}
func ndrRangeRecords(s NDRObjectSpec) int {
	if s.Signed {
		return 3
	}
	if s.RangeLeft == 0 && s.RangeRight == 0 {
		return 0
	}
	if s.RangeLeft == s.RangeRight {
		return 1
	}
	return 2
}
func (d *NDRDesignData) AddObject(moduleID int, s NDRObjectSpec) error {
	m := moduleID - 256
	if m < 0 || m >= len(d.head) || d.head[m] != NDRModule || s.Operator == AbcOperNone {
		return ErrDimension
	}
	if m+int(d.body[m]) != len(d.head) {
		return ErrDimension
	}
	if s.Operator == AbcOperLUT && s.Function != nil && len(s.Function) != 8 {
		return ErrDimension
	}
	function := s.Function
	if s.Operator != AbcOperLUT {
		for index, value := range function {
			if value == 0 {
				function = function[:index]
				break
			}
		}
	}
	fnWords := 0
	if s.Function != nil {
		if s.Operator == AbcOperLUT {
			fnWords = 2
		} else {
			// Ndr_DataPushString reserves and writes a terminating zero byte,
			// including a full extra word when the byte length is divisible by 4.
			fnWords = (len(function) + 4) / 4
		}
	}
	n := 2 + ndrRangeRecords(s)
	if s.InstanceName != 0 {
		n++
	}
	n += len(s.Inputs) + len(s.Outputs) + fnWords
	if e := d.reserve(n); e != nil {
		return e
	}
	o := len(d.head)
	d.push(NDRObject, 0)
	d.push(NDROperatorType, int(s.Operator))
	if s.Signed {
		d.push(NDRRange, s.RangeLeft)
		d.push(NDRRange, s.RangeRight)
		d.push(NDRRange, 1)
	} else if s.RangeLeft != 0 || s.RangeRight != 0 {
		d.push(NDRRange, s.RangeLeft)
		if s.RangeLeft != s.RangeRight {
			d.push(NDRRange, s.RangeRight)
		}
	}
	if s.InstanceName != 0 {
		d.push(NDRName, s.InstanceName)
	}
	for _, v := range s.Inputs {
		d.push(NDRInput, v)
	}
	for _, v := range s.Outputs {
		d.push(NDROutput, v)
	}
	for i := 0; i < fnWords; i++ {
		var w [4]byte
		start := min(len(function), 4*i)
		copy(w[:], function[start:min(len(function), 4*i+4)])
		d.push(NDRFunction, int(binary.LittleEndian.Uint32(w[:])))
	}
	added := len(d.head) - o
	d.body[o] = uint32(added)
	d.body[m] += uint32(added)
	d.body[0] += uint32(added)
	return nil
}
func (d *NDRDesignData) Modules() ([]int, error) {
	if e := d.check(); e != nil {
		return nil, e
	}
	r := make([]int, 0)
	for i := 1; i < len(d.head); {
		size, _ := d.DataSize(i)
		if size <= 0 || i > len(d.head)-size {
			return nil, ErrInvariant
		}
		if d.head[i] == NDRModule {
			r = append(r, i+256)
		}
		i += size
	}
	return r, nil
}
func (d *NDRDesignData) Objects(moduleID int) ([]int, error) {
	if e := d.check(); e != nil {
		return nil, e
	}
	m := moduleID - 256
	if m < 0 || m >= len(d.head) || d.head[m] != NDRModule {
		return nil, ErrDimension
	}
	r := make([]int, 0)
	for i := m + 1; i < m+int(d.body[m]); {
		size, _ := d.DataSize(i)
		if size <= 0 || i > len(d.head)-size || i+size > m+int(d.body[m]) {
			return nil, ErrInvariant
		}
		if d.head[i] == NDRObject {
			r = append(r, i)
		}
		i += size
	}
	return r, nil
}

// Targets is the owned counterpart of Ndr_ModForEachTarget. Returned values
// are record indexes, preserving module order.
func (d *NDRDesignData) Targets(moduleID int) ([]int, error) {
	if e := d.check(); e != nil {
		return nil, e
	}
	m := moduleID - 256
	if m < 0 || m >= len(d.head) || d.head[m] != NDRModule {
		return nil, ErrDimension
	}
	result := make([]int, 0)
	end := m + int(d.body[m])
	for record := m + 1; record < end; {
		size, err := d.DataSize(record)
		if err != nil || size <= 0 || record > end-size {
			return nil, ErrInvariant
		}
		if d.head[record] == NDRTarget {
			result = append(result, record)
		}
		record += size
	}
	return result, nil
}
func (d *NDRDesignData) ObjectEntries(obj int, t NDRRecordType) ([]int, error) {
	if e := d.check(); e != nil {
		return nil, e
	}
	if obj < 0 || obj >= len(d.head) || d.head[obj] != NDRObject {
		return nil, ErrDimension
	}
	return d.recordEntries(obj, t)
}
func (d *NDRDesignData) recordEntries(container int, t NDRRecordType) ([]int, error) {
	if container < 0 || container >= len(d.head) || d.head[container] > NDRObject {
		return nil, ErrDimension
	}
	r := make([]int, 0)
	end := container + int(d.body[container])
	if end > len(d.head) || end <= container {
		return nil, ErrInvariant
	}
	for i := container + 1; i < end; {
		if d.head[i] == t {
			r = append(r, int(int32(d.body[i])))
		}
		size, err := d.DataSize(i)
		if err != nil || size <= 0 || i > end-size {
			return nil, ErrInvariant
		}
		i += size
	}
	return r, nil
}
func (d *NDRDesignData) ObjectIsType(obj int, op AbcOperator) (bool, error) {
	v, e := d.ObjectEntries(obj, NDROperatorType)
	if e != nil {
		return false, e
	}
	return len(v) > 0 && v[0] == int(op), nil
}
func (d *NDRDesignData) Counts(moduleID int) (cis, cos, nodes int, err error) {
	objects, e := d.Objects(moduleID)
	if e != nil {
		err = e
		return
	}
	for _, o := range objects {
		ops, _ := d.ObjectEntries(o, NDROperatorType)
		if len(ops) == 0 {
			continue
		}
		if AbcOperator(ops[0]) == AbcOperCI {
			cis++
		} else if AbcOperator(ops[0]) == AbcOperCO {
			cos++
		} else {
			nodes++
		}
	}
	return
}
func (d *NDRDesignData) MarshalBinary() ([]byte, error) {
	if e := d.check(); e != nil {
		return nil, e
	}
	b := make([]byte, 5*len(d.head))
	for i, v := range d.body {
		binary.LittleEndian.PutUint32(b[4*i:], v)
	}
	off := 4 * len(d.head)
	for i, t := range d.head {
		b[off+i] = byte(t)
	}
	return b, nil
}
func ParseNDRDesign(b []byte, limits MiniNetworkLimits) (*NDRDesignData, error) {
	if len(b) == 0 || len(b)%5 != 0 {
		return nil, ErrDimension
	}
	limits, e := normalizeMiniLimits(limits)
	if e != nil {
		return nil, e
	}
	n := len(b) / 5
	if n > limits.MaxRecords {
		return nil, ErrLimit
	}
	d := &NDRDesignData{head: make([]NDRRecordType, n), body: make([]uint32, n), limits: limits}
	for i := range d.body {
		d.body[i] = binary.LittleEndian.Uint32(b[4*i:])
		d.head[i] = NDRRecordType(b[4*n+i])
		if d.head[i] > NDRUnknown {
			return nil, ErrDimension
		}
	}
	if e = d.check(); e != nil {
		return nil, e
	}
	mods, e := d.Modules()
	if e != nil {
		return nil, e
	}
	for _, m := range mods {
		if _, e = d.Objects(m); e != nil {
			return nil, e
		}
	}
	return d, nil
}
func (d *NDRDesignData) Close() {
	if d != nil {
		d.head = nil
		d.body = nil
		d.closed = true
	}
}

func NDRRangeText(values []int, skipBinary bool) string {
	if skipBinary && (len(values) == 0 || len(values) == 1) {
		return ""
	}
	if len(values) == 3 && skipBinary {
		return "signed "
	}
	if len(values) == 0 {
		if skipBinary {
			return "[0:0]"
		}
		return "[0]"
	}
	if len(values) == 1 {
		if skipBinary {
			return "[" + strconv.Itoa(values[0]) + ":" + strconv.Itoa(values[0]) + "]"
		}
		return "[" + strconv.Itoa(values[0]) + "]"
	}
	return "[" + strconv.Itoa(values[0]) + ":" + strconv.Itoa(values[1]) + "]"
}
func (d *NDRDesignData) Summary() string {
	mods, _ := d.Modules()
	var b strings.Builder
	b.WriteString("NDR design with ")
	b.WriteString(strconv.Itoa(len(mods)))
	b.WriteString(" modules")
	return b.String()
}

func ndrName(names []string, id int) (string, error) {
	if id < 0 || id >= len(names) || names[id] == "" {
		return "", ErrDimension
	}
	return names[id], nil
}
func (d *NDRDesignData) objectRange(obj int) []int {
	values, _ := d.ObjectEntries(obj, NDRRange)
	return values
}
func (d *NDRDesignData) objectOperator(obj int) (AbcOperator, error) {
	v, e := d.ObjectEntries(obj, NDROperatorType)
	if e != nil || len(v) != 1 {
		if e != nil {
			return 0, e
		}
		return 0, ErrInvariant
	}
	return AbcOperator(v[0]), nil
}
func (d *NDRDesignData) Verilog(names []string) (string, error) {
	mods, err := d.Modules()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, moduleID := range mods {
		m := moduleID - 256
		moduleNames, e := d.recordEntries(m, NDRName)
		if e != nil || len(moduleNames) != 1 {
			if e != nil {
				return "", e
			}
			return "", ErrInvariant
		}
		moduleName, e := ndrName(names, moduleNames[0])
		if e != nil {
			return "", e
		}
		objects, e := d.Objects(moduleID)
		if e != nil {
			return "", e
		}
		inputs := make([]string, 0)
		outputs := make([]string, 0)
		for _, obj := range objects {
			op, e := d.objectOperator(obj)
			if e != nil {
				return "", e
			}
			if op == AbcOperCI {
				ids, _ := d.ObjectEntries(obj, NDROutput)
				if len(ids) != 1 {
					return "", ErrInvariant
				}
				name, e := ndrName(names, ids[0])
				if e != nil {
					return "", e
				}
				inputs = append(inputs, name)
			} else if op == AbcOperCO {
				ids, _ := d.ObjectEntries(obj, NDRInput)
				if len(ids) != 1 {
					return "", ErrInvariant
				}
				name, e := ndrName(names, ids[0])
				if e != nil {
					return "", e
				}
				outputs = append(outputs, name)
			}
		}
		b.WriteString("module " + moduleName + " (")
		all := append(slices.Clone(inputs), outputs...)
		b.WriteString(strings.Join(all, ", "))
		b.WriteString(");\n")
		for _, obj := range objects {
			op, _ := d.objectOperator(obj)
			ins, _ := d.ObjectEntries(obj, NDRInput)
			outs, _ := d.ObjectEntries(obj, NDROutput)
			if op == AbcOperCI {
				b.WriteString("  input " + NDRRangeText(d.objectRange(obj), true))
				if len(d.objectRange(obj)) > 0 {
					b.WriteByte(' ')
				}
				name, _ := ndrName(names, outs[0])
				b.WriteString(name + ";\n")
				continue
			}
			if op == AbcOperCO {
				b.WriteString("  output " + NDRRangeText(d.objectRange(obj), true))
				if len(d.objectRange(obj)) > 0 {
					b.WriteByte(' ')
				}
				name, _ := ndrName(names, ins[0])
				b.WriteString(name + ";\n")
				continue
			}
			inNames := make([]string, len(ins))
			for i, id := range ins {
				inNames[i], e = ndrName(names, id)
				if e != nil {
					return "", e
				}
			}
			if op >= 256 {
				moduleNames, e := d.recordEntries(int(op)-256, NDRName)
				if e != nil || len(moduleNames) != 1 {
					if e != nil {
						return "", e
					}
					return "", ErrInvariant
				}
				referenced, e := ndrName(names, moduleNames[0])
				if e != nil {
					return "", e
				}
				instance := ""
				instanceIDs, _ := d.ObjectEntries(obj, NDRName)
				if len(instanceIDs) > 0 {
					instance, e = ndrName(names, instanceIDs[0])
					if e != nil {
						return "", e
					}
					instance = " " + instance
				}
				b.WriteString("  " + referenced + instance + " (" + strings.Join(inNames, ", ") + ");\n")
				continue
			}
			if len(outs) != 1 {
				return "", ErrDimension
			}
			out, e := ndrName(names, outs[0])
			if e != nil {
				return "", e
			}
			instance := ""
			instanceIDs, _ := d.ObjectEntries(obj, NDRName)
			if len(instanceIDs) > 0 {
				instance, e = ndrName(names, instanceIDs[0])
				if e != nil {
					return "", e
				}
				instance = " " + instance
			}
			switch op {
			case AbcOperDFF:
				if len(inNames) != 2 {
					return "", ErrDimension
				}
				b.WriteString("  ABC_DFF" + instance + " (.q(" + out + "), .d(" + inNames[0] + "), .init(" + inNames[1] + "));\n")
				continue
			case AbcOperDFFRSE:
				if len(inNames) != 8 {
					return "", ErrDimension
				}
				ports := []string{"d", "clk", "reset", "set", "enable", "async", "sre", "init"}
				parts := []string{".q(" + out + ")"}
				for index, port := range ports {
					parts = append(parts, "."+port+"("+inNames[index]+")")
				}
				b.WriteString("  ABC_DFFRSE" + instance + " (" + strings.Join(parts, ", ") + ");\n")
				continue
			case AbcOperRAMR:
				if len(inNames) != 2 {
					return "", ErrDimension
				}
				b.WriteString("  ABC_READ" + instance + " (.data(" + out + "), .mem_in(" + inNames[0] + "), .addr(" + inNames[1] + "));\n")
				continue
			case AbcOperRAMW:
				if len(inNames) != 3 {
					return "", ErrDimension
				}
				b.WriteString("  ABC_WRITE" + instance + " (.mem_out(" + out + "), .mem_in(" + inNames[0] + "), .addr(" + inNames[1] + "), .data(" + inNames[2] + "));\n")
				continue
			}
			operator, ok := AbcOperatorName(op)
			if !ok {
				return "", ErrDimension
			}
			b.WriteString("  assign " + out + " = ")
			switch {
			case op == AbcOperBitBuf && len(inNames) == 1:
				b.WriteString(inNames[0])
			case op == AbcOperSlice && len(inNames) == 1:
				b.WriteString(inNames[0] + NDRRangeText(d.objectRange(obj), false))
			case op == AbcOperConcat && len(inNames) > 0:
				b.WriteString("{" + strings.Join(inNames, ", ") + "}")
			case op == AbcOperAriAdd && len(inNames) == 3:
				b.WriteString(strings.Join(inNames, " + "))
			case len(inNames) == 0:
				raw, _ := d.ObjectEntries(obj, NDRFunction)
				bytes := make([]byte, 4*len(raw))
				for i, v := range raw {
					binary.LittleEndian.PutUint32(bytes[4*i:], uint32(v))
				}
				b.WriteString(strings.TrimRight(string(bytes), "\x00"))
			case len(inNames) == 1:
				b.WriteString(operator + " " + inNames[0])
			case len(inNames) == 2:
				b.WriteString(inNames[0] + " " + operator + " " + inNames[1])
			case op == AbcOperBitMux && len(inNames) == 3:
				b.WriteString(inNames[0] + " ? " + inNames[1] + " : " + inNames[2])
			default:
				return "", ErrDimension
			}
			b.WriteString(";\n")
		}
		b.WriteString("endmodule\n")
	}
	return b.String(), nil
}
