package abc

// Derived from ABC src/base/cba/cba.h, cbaTypes.h, cbaPrs.h,
// cbaNtk.c, cbaBlast.c, cbaReadBlif.c, cbaReadVer.c,
// cbaWriteBlif.c, cbaWriteVer.c, cbaCom.c, and module.make.

import (
	"bufio"
	"fmt"
	"math/bits"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

// CBAObjectType preserves the source object and primitive ordering.
type CBAObjectType uint8

const (
	CBAObjectNone CBAObjectType = iota
	CBAObjectPI
	CBAObjectPO
	CBAObjectBox
	CBABoxConstFalse
	CBABoxConstTrue
	CBABoxConstX
	CBABoxConstZ
	CBABoxBuf
	CBABoxInv
	CBABoxAnd
	CBABoxNand
	CBABoxOr
	CBABoxNor
	CBABoxXor
	CBABoxXnor
	CBABoxSharp
	CBABoxSharpLeft
	CBABoxMux
	CBABoxMajority
	CBABoxAndCarry
	CBABoxBorrow
	CBABoxBorrowOut
	CBABoxXorCarry
	CBABoxCarryN
	CBABoxAndOr
	CBABoxOrAnd
	CBABoxReduceAnd
	CBABoxReduceNand
	CBABoxReduceOr
	CBABoxReduceNor
	CBABoxReduceXor
	CBABoxReduceXnor
	CBABoxLogicalNot
	CBABoxLogicalAnd
	CBABoxLogicalNand
	CBABoxLogicalOr
	CBABoxLogicalNor
	CBABoxLogicalXor
	CBABoxLogicalXnor
	CBABoxNMux
	CBABoxSelect
	CBABoxPrioritySelect
	CBABoxEncode
	CBABoxPriorityEncode
	CBABoxDecode
	CBABoxEnableDecode
	CBABoxAdd
	CBABoxSub
	CBABoxMul
	CBABoxSignedMul
	CBABoxDiv
	CBABoxMod
	CBABoxRem
	CBABoxPow
	CBABoxMin
	CBABoxSqrt
	CBABoxAbs
	CBABoxSignedLess
	CBABoxLess
	CBABoxLessEqual
	CBABoxMoreEqual
	CBABoxMore
	CBABoxEqual
	CBABoxNotEqual
	CBABoxShiftLeft
	CBABoxShiftRight
	CBABoxArithmeticShiftLeft
	CBABoxArithmeticShiftRight
	CBABoxRotateLeft
	CBABoxRotateRight
	CBABoxNode
	CBABoxLUT
	CBABoxGate
	CBABoxTable
	CBABoxTri
	CBABoxRAM
	CBABoxRAMRead
	CBABoxRAMWrite
	CBABoxRAMWriteClock
	CBABoxRAMLatch
	CBABoxRAMSequential
	CBABoxRAMBox
	CBABoxLatch
	CBABoxLatchResetSet
	CBABoxDFF
	CBABoxDFFComplement
	CBABoxDFFResetSet
	CBABoxSlice
	CBABoxConcat
	CBAObjectTypeCount
)

// CBATypeNames is the deterministic primitive spelling used by reports.
var CBATypeNames = [...]string{
	"none", "pi", "po", "box", "0", "1", "x", "z", "buf", "inv",
	"and", "nand", "or", "nor", "xor", "xnor", "sharp", "sharpl", "mux", "maj",
	"abc", "ba", "bo", "bx", "bn", "bao", "boa", "rand", "rnand", "ror",
	"rnor", "rxor", "rxnor", "lnot", "land", "lnand", "lor", "lnor", "lxor", "lxnor",
	"nmux", "sel", "psel", "enc", "penc", "dec", "edec", "add", "sub", "mul",
	"smul", "div", "mod", "rem", "pow", "min", "sqrt", "abs", "slt", "lt",
	"le", "ge", "gt", "eq", "ne", "shl", "shr", "ashl", "ashr", "rotl",
	"rotr", "node", "lut", "gate", "table", "tri", "ram", "ramr", "ramw", "ramwc",
	"raml", "rams", "rambox", "latch", "latchrs", "dff", "dffcpl", "dffrs", "slice", "concat",
}

// CBALimits bounds all retained and temporary work.
type CBALimits struct {
	MaxModules    int
	MaxObjects    int
	MaxEdges      int
	MaxBits       int
	MaxOperations int
	MaxBytes      int
	MaxTextBytes  int
	MaxDepth      int
	Stop          func() bool
}

func normalizeCBALimits(l CBALimits) (CBALimits, error) {
	values := []*int{&l.MaxModules, &l.MaxObjects, &l.MaxEdges, &l.MaxBits, &l.MaxOperations, &l.MaxBytes, &l.MaxTextBytes, &l.MaxDepth}
	defaults := []int{1 << 14, 1 << 20, 1 << 22, 1 << 24, 1 << 27, 1 << 29, 1 << 28, 1 << 14}
	for i := range values {
		if *values[i] < 0 {
			return l, ErrDimension
		}
		if *values[i] == 0 {
			*values[i] = defaults[i]
		}
	}
	return l, nil
}

type cbaBudget struct {
	limits CBALimits
	ops    int
}

func (b *cbaBudget) use(n int) error {
	if n < 0 || b.ops > b.limits.MaxOperations-n {
		return ErrLimit
	}
	b.ops += n
	if b.limits.Stop != nil && b.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

// CBARange preserves declared direction as well as width.
type CBARange struct{ Left, Right int }

func (r CBARange) Width() (int, error) {
	var d uint64
	if r.Left >= r.Right {
		d = uint64(r.Left) - uint64(r.Right)
	} else {
		d = uint64(r.Right) - uint64(r.Left)
	}
	if d >= uint64(^uint(0)>>1) {
		return 0, ErrLimit
	}
	return int(d + 1), nil
}

// CBAObject is one owned word-level object. Object IDs are stable and one-based.
type CBAObject struct {
	ID       int
	Type     CBAObjectType
	Name     string
	Width    int
	Signed   bool
	Inputs   []int
	Value    uint64
	ParamA   int
	ParamB   int
	Table    uint64
	Module   string
	Instance string
}

func (o CBAObject) clone() CBAObject {
	o.Inputs = slices.Clone(o.Inputs)
	return o
}

// CBANetwork is an owned, topologically ordered word-level network.
type CBANetwork struct {
	ID      int
	Name    string
	objects []CBAObject
	inputs  []int
	outputs []int
	names   map[string]int
	edges   int
	bits    int
	bytes   int
	limits  CBALimits
	owner   *CBADesign
	closed  bool
}

const cbaNetworkBaseBytes = 64
const cbaObjectBaseBytes = 96

func NewCBANetwork(name string, limits CBALimits) (*CBANetwork, error) {
	limits, err := normalizeCBALimits(limits)
	if err != nil {
		return nil, err
	}
	if !cbaLegalName(name) || len(name) > limits.MaxBytes {
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if len(name) > limits.MaxBytes-cbaNetworkBaseBytes {
		return nil, ErrLimit
	}
	return &CBANetwork{Name: name, objects: []CBAObject{{}}, names: make(map[string]int), bytes: cbaNetworkBaseBytes + len(name), limits: limits}, nil
}

func (n *CBANetwork) check() error {
	if n == nil || n.closed || len(n.objects) == 0 || n.names == nil {
		return ErrInvariant
	}
	return nil
}

func cbaLegalName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if !(r == '_' || unicode.IsLetter(r) || i > 0 && unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

func (n *CBANetwork) addObject(obj CBAObject) (int, error) {
	if err := n.check(); err != nil {
		return 0, err
	}
	if n.limits.Stop != nil && n.limits.Stop() {
		return 0, ErrCancelled
	}
	if obj.Width < 1 || obj.Width > 64 || !cbaLegalName(obj.Name) || obj.Type <= CBAObjectNone || obj.Type >= CBAObjectTypeCount {
		return 0, ErrDimension
	}
	if _, exists := n.names[obj.Name]; exists {
		return 0, ErrInvariant
	}
	objectBytes := cbaObjectBaseBytes + len(obj.Name)
	if len(obj.Inputs) > (n.limits.MaxBytes-objectBytes)/8 {
		return 0, ErrLimit
	}
	objectBytes += 8 * len(obj.Inputs)
	if len(n.objects) >= n.limits.MaxObjects || n.edges > n.limits.MaxEdges-len(obj.Inputs) || n.bits > n.limits.MaxBits-obj.Width || n.bytes > n.limits.MaxBytes-objectBytes {
		return 0, ErrLimit
	}
	for _, id := range obj.Inputs {
		if id <= 0 || id >= len(n.objects) {
			return 0, ErrInvariant
		}
	}
	obj.ID = len(n.objects)
	obj.Inputs = slices.Clone(obj.Inputs)
	n.objects = append(n.objects, obj)
	n.names[obj.Name] = obj.ID
	n.edges += len(obj.Inputs)
	n.bits += obj.Width
	n.bytes += objectBytes
	return obj.ID, nil
}

func (n *CBANetwork) AddInput(name string, width int, signed bool) (int, error) {
	id, err := n.addObject(CBAObject{Type: CBAObjectPI, Name: name, Width: width, Signed: signed})
	if err == nil {
		n.inputs = append(n.inputs, id)
	}
	return id, err
}

func (n *CBANetwork) AddConstant(name string, width int, value uint64) (int, error) {
	if width < 1 || width > 64 {
		return 0, ErrDimension
	}
	t := CBABoxBuf
	if value == 0 {
		t = CBABoxConstFalse
	} else if value == cbaMask(width) {
		t = CBABoxConstTrue
	}
	return n.addObject(CBAObject{Type: t, Name: name, Width: width, Value: value & cbaMask(width)})
}

func (n *CBANetwork) AddNode(name string, typ CBAObjectType, width int, signed bool, inputs []int) (int, error) {
	if err := n.check(); err != nil {
		return 0, err
	}
	for _, id := range inputs {
		if id <= 0 || id >= len(n.objects) {
			return 0, ErrInvariant
		}
	}
	if err := n.checkNodeShape(typ, width, inputs); err != nil {
		return 0, err
	}
	paramA := 0
	if len(inputs) > 0 && inputs[0] > 0 && inputs[0] < len(n.objects) {
		paramA = n.objects[inputs[0]].Width
	}
	return n.addObject(CBAObject{Type: typ, Name: name, Width: width, Signed: signed, Inputs: inputs, ParamA: paramA})
}

func (n *CBANetwork) checkNodeShape(typ CBAObjectType, width int, inputs []int) error {
	if width < 1 || width > 64 {
		return ErrDimension
	}
	need := -1
	switch typ {
	case CBABoxInv, CBABoxBuf, CBABoxNode, CBABoxGate, CBABoxReduceAnd, CBABoxReduceNand, CBABoxReduceOr, CBABoxReduceNor, CBABoxReduceXor, CBABoxReduceXnor, CBABoxLogicalNot, CBABoxMin, CBABoxSqrt, CBABoxAbs:
		need = 1
	case CBABoxAnd, CBABoxNand, CBABoxOr, CBABoxNor, CBABoxXor, CBABoxXnor, CBABoxLogicalAnd, CBABoxLogicalNand, CBABoxLogicalOr, CBABoxLogicalNor, CBABoxLogicalXor, CBABoxLogicalXnor, CBABoxAdd, CBABoxSub, CBABoxMul, CBABoxSignedMul, CBABoxDiv, CBABoxMod, CBABoxRem, CBABoxPow, CBABoxSignedLess, CBABoxLess, CBABoxLessEqual, CBABoxMoreEqual, CBABoxMore, CBABoxEqual, CBABoxNotEqual, CBABoxShiftLeft, CBABoxShiftRight, CBABoxArithmeticShiftLeft, CBABoxArithmeticShiftRight, CBABoxRotateLeft, CBABoxRotateRight:
		need = 2
	case CBABoxMux, CBABoxMajority:
		need = 3
	case CBABoxConcat:
		if len(inputs) == 0 {
			return ErrDimension
		}
		total := 0
		for _, id := range inputs {
			if total > 64-n.objects[id].Width {
				return ErrLimit
			}
			total += n.objects[id].Width
		}
		if total != width {
			return ErrDimension
		}
	}
	if need >= 0 && len(inputs) != need {
		return ErrDimension
	}
	if typ >= CBABoxReduceAnd && typ <= CBABoxReduceXnor || typ >= CBABoxLogicalNot && typ <= CBABoxLogicalXnor || typ >= CBABoxSignedLess && typ <= CBABoxNotEqual {
		if width != 1 {
			return ErrDimension
		}
	}
	if typ == CBABoxMux && (n.objects[inputs[1]].Width != width || n.objects[inputs[2]].Width != width) {
		return ErrDimension
	}
	if typ == CBABoxMajority {
		for _, id := range inputs {
			if n.objects[id].Width != width {
				return ErrDimension
			}
		}
	}
	if typ == CBABoxInv || typ == CBABoxBuf || typ == CBABoxNode || typ == CBABoxGate || typ == CBABoxMin || typ == CBABoxSqrt || typ == CBABoxAbs {
		if n.objects[inputs[0]].Width != width {
			return ErrDimension
		}
	}
	if typ == CBABoxAnd || typ == CBABoxNand || typ == CBABoxOr || typ == CBABoxNor || typ == CBABoxXor || typ == CBABoxXnor || typ == CBABoxAdd || typ == CBABoxSub || typ == CBABoxMul || typ == CBABoxSignedMul || typ == CBABoxDiv || typ == CBABoxMod || typ == CBABoxRem {
		if n.objects[inputs[0]].Width != width || n.objects[inputs[1]].Width != width {
			return ErrDimension
		}
	}
	if typ >= CBABoxSignedLess && typ <= CBABoxNotEqual {
		if n.objects[inputs[0]].Width != n.objects[inputs[1]].Width {
			return ErrDimension
		}
	}
	if typ == CBABoxShiftLeft || typ == CBABoxShiftRight || typ == CBABoxArithmeticShiftLeft || typ == CBABoxArithmeticShiftRight || typ == CBABoxRotateLeft || typ == CBABoxRotateRight || typ == CBABoxPow {
		if n.objects[inputs[0]].Width != width {
			return ErrDimension
		}
	}
	return nil
}

func (n *CBANetwork) AddSlice(name string, source, left, right int) (int, error) {
	if err := n.check(); err != nil {
		return 0, err
	}
	if source <= 0 || source >= len(n.objects) || left < 0 || right < 0 || left >= n.objects[source].Width || right >= n.objects[source].Width {
		return 0, ErrDimension
	}
	w := left - right
	if w < 0 {
		w = -w
	}
	return n.addObject(CBAObject{Type: CBABoxSlice, Name: name, Width: w + 1, Inputs: []int{source}, ParamA: left, ParamB: right})
}

func (n *CBANetwork) AddTable(name string, inputs []int, table uint64) (int, error) {
	if err := n.check(); err != nil {
		return 0, err
	}
	if len(inputs) > 6 {
		return 0, ErrLimit
	}
	return n.addObject(CBAObject{Type: CBABoxTable, Name: name, Width: 1, Inputs: inputs, Table: table})
}

func (n *CBANetwork) AddOutput(name string, width int, source int) (int, error) {
	if err := n.check(); err != nil {
		return 0, err
	}
	if source <= 0 || source >= len(n.objects) || n.objects[source].Width != width {
		return 0, ErrDimension
	}
	id, err := n.addObject(CBAObject{Type: CBAObjectPO, Name: name, Width: width, Inputs: []int{source}})
	if err == nil {
		n.outputs = append(n.outputs, id)
	}
	return id, err
}

func (n *CBANetwork) Object(id int) (CBAObject, error) {
	if err := n.check(); err != nil {
		return CBAObject{}, err
	}
	if id <= 0 || id >= len(n.objects) {
		return CBAObject{}, ErrDimension
	}
	return n.objects[id].clone(), nil
}

func (n *CBANetwork) Objects() ([]CBAObject, error) {
	if err := n.check(); err != nil {
		return nil, err
	}
	r := make([]CBAObject, len(n.objects)-1)
	for i := 1; i < len(n.objects); i++ {
		r[i-1] = n.objects[i].clone()
	}
	return r, nil
}

func (n *CBANetwork) Inputs() []int {
	if n == nil || n.closed {
		return nil
	}
	return slices.Clone(n.inputs)
}
func (n *CBANetwork) Outputs() []int {
	if n == nil || n.closed {
		return nil
	}
	return slices.Clone(n.outputs)
}

func (n *CBANetwork) Validate() error {
	if err := n.check(); err != nil {
		return err
	}
	if len(n.objects) > n.limits.MaxObjects || n.edges > n.limits.MaxEdges || n.bits > n.limits.MaxBits || n.bytes > n.limits.MaxBytes {
		return ErrLimit
	}
	seenNames := make(map[string]int, len(n.objects))
	levels := make([]int, len(n.objects))
	budget := cbaBudget{limits: n.limits}
	edges, width, retainedBytes := 0, 0, cbaNetworkBaseBytes+len(n.Name)
	for id := 1; id < len(n.objects); id++ {
		if n.limits.Stop != nil && n.limits.Stop() {
			return ErrCancelled
		}
		o := n.objects[id]
		if err := budget.use(1 + len(o.Inputs)); err != nil {
			return err
		}
		if o.ID != id || o.Width < 1 || o.Width > 64 || !cbaLegalName(o.Name) || o.Type <= CBAObjectNone || o.Type >= CBAObjectTypeCount {
			return ErrInvariant
		}
		if _, ok := seenNames[o.Name]; ok {
			return ErrInvariant
		}
		seenNames[o.Name] = id
		if n.names[o.Name] != id {
			return ErrInvariant
		}
		level := 1
		for _, in := range o.Inputs {
			if in <= 0 || in >= id {
				return ErrInvariant
			}
			if levels[in] >= level {
				level = levels[in] + 1
			}
		}
		if err := n.validateStoredShape(o); err != nil {
			return err
		}
		if level > n.limits.MaxDepth {
			return ErrLimit
		}
		levels[id] = level
		edges += len(o.Inputs)
		width += o.Width
		retainedBytes += cbaObjectBaseBytes + len(o.Name) + 8*len(o.Inputs)
	}
	if edges != n.edges || width != n.bits || retainedBytes != n.bytes || len(seenNames) != len(n.names) {
		return ErrInvariant
	}
	listed := make(map[int]bool, len(n.inputs)+len(n.outputs))
	for _, id := range n.inputs {
		if id <= 0 || id >= len(n.objects) || n.objects[id].Type != CBAObjectPI || listed[id] {
			return ErrInvariant
		}
		listed[id] = true
	}
	for _, id := range n.outputs {
		if id <= 0 || id >= len(n.objects) || n.objects[id].Type != CBAObjectPO || listed[id] {
			return ErrInvariant
		}
		listed[id] = true
	}
	for id := 1; id < len(n.objects); id++ {
		if (n.objects[id].Type == CBAObjectPI || n.objects[id].Type == CBAObjectPO) != listed[id] {
			return ErrInvariant
		}
	}
	return nil
}

func (n *CBANetwork) validateStoredShape(o CBAObject) error {
	switch o.Type {
	case CBAObjectPI:
		if len(o.Inputs) != 0 {
			return ErrInvariant
		}
		return nil
	case CBAObjectPO:
		if len(o.Inputs) != 1 || n.objects[o.Inputs[0]].Width != o.Width {
			return ErrInvariant
		}
		return nil
	case CBABoxConstFalse, CBABoxConstTrue, CBABoxConstX, CBABoxConstZ:
		if len(o.Inputs) != 0 {
			return ErrInvariant
		}
		return nil
	case CBABoxBuf:
		if len(o.Inputs) == 0 {
			return nil
		}
	case CBABoxSlice:
		if len(o.Inputs) != 1 || o.ParamA < 0 || o.ParamB < 0 || o.ParamA >= n.objects[o.Inputs[0]].Width || o.ParamB >= n.objects[o.Inputs[0]].Width {
			return ErrInvariant
		}
		w := o.ParamA - o.ParamB
		if w < 0 {
			w = -w
		}
		if o.Width != w+1 {
			return ErrInvariant
		}
		return nil
	case CBABoxTable, CBABoxLUT:
		if o.Width != 1 || len(o.Inputs) > 6 {
			return ErrInvariant
		}
		return nil
	}
	if err := n.checkNodeShape(o.Type, o.Width, o.Inputs); err != nil {
		return ErrInvariant
	}
	return nil
}

func cbaMask(width int) uint64 {
	if width >= 64 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(width) - 1
}

func cbaSigned(v uint64, width int) int64 {
	v &= cbaMask(width)
	if width == 64 {
		return int64(v)
	}
	shift := uint(64 - width)
	return int64(v<<shift) >> shift
}

func cbaRotateLeft(v uint64, shift, width int) uint64 {
	if width == 0 {
		return 0
	}
	shift %= width
	mask := cbaMask(width)
	if shift == 0 {
		return v & mask
	}
	return ((v << uint(shift)) | (v >> uint(width-shift))) & mask
}

func cbaIntegerSqrt(v uint64) uint64 {
	var root, bit uint64 = 0, uint64(1) << 62
	for bit > v {
		bit >>= 2
	}
	for bit != 0 {
		if v >= root+bit {
			v -= root + bit
			root = (root >> 1) + bit
		} else {
			root >>= 1
		}
		bit >>= 2
	}
	return root
}

func cbaEvalObject(o CBAObject, values []uint64) (uint64, error) {
	args := make([]uint64, len(o.Inputs))
	for i, id := range o.Inputs {
		args[i] = values[id]
	}
	need := func(n int) error {
		if len(args) != n {
			return ErrDimension
		}
		return nil
	}
	mask := cbaMask(o.Width)
	var v uint64
	switch o.Type {
	case CBABoxConstFalse:
		return 0, nil
	case CBABoxConstTrue:
		return mask, nil
	case CBABoxBuf, CBABoxNode, CBABoxGate:
		if len(args) == 0 {
			return o.Value & mask, nil
		}
		if err := need(1); err != nil {
			return 0, err
		}
		v = args[0]
	case CBAObjectPO:
		if err := need(1); err != nil {
			return 0, err
		}
		v = args[0]
	case CBABoxInv:
		if err := need(1); err != nil {
			return 0, err
		}
		v = ^args[0]
	case CBABoxAnd, CBABoxNand, CBABoxOr, CBABoxNor, CBABoxXor, CBABoxXnor:
		if err := need(2); err != nil {
			return 0, err
		}
		switch o.Type {
		case CBABoxAnd:
			v = args[0] & args[1]
		case CBABoxNand:
			v = ^(args[0] & args[1])
		case CBABoxOr:
			v = args[0] | args[1]
		case CBABoxNor:
			v = ^(args[0] | args[1])
		case CBABoxXor:
			v = args[0] ^ args[1]
		case CBABoxXnor:
			v = ^(args[0] ^ args[1])
		}
	case CBABoxReduceAnd, CBABoxReduceNand, CBABoxReduceOr, CBABoxReduceNor, CBABoxReduceXor, CBABoxReduceXnor:
		if err := need(1); err != nil {
			return 0, err
		}
		inputWidth := o.ParamA
		if inputWidth < 1 || inputWidth > 64 {
			inputWidth = 64
		}
		inputMask := cbaMask(inputWidth)
		v0 := args[0] & inputMask
		switch o.Type {
		case CBABoxReduceAnd:
			v = cbaBoolWord(v0 == inputMask)
		case CBABoxReduceNand:
			v = cbaBoolWord(v0 != inputMask)
		case CBABoxReduceOr:
			v = cbaBoolWord(v0 != 0)
		case CBABoxReduceNor:
			v = cbaBoolWord(v0 == 0)
		case CBABoxReduceXor:
			v = uint64(bits.OnesCount64(v0) & 1)
		case CBABoxReduceXnor:
			v = uint64(1 - bits.OnesCount64(v0)&1)
		}
	case CBABoxLogicalNot:
		if err := need(1); err != nil {
			return 0, err
		}
		v = cbaBoolWord(args[0] == 0)
	case CBABoxLogicalAnd, CBABoxLogicalNand, CBABoxLogicalOr, CBABoxLogicalNor, CBABoxLogicalXor, CBABoxLogicalXnor:
		if err := need(2); err != nil {
			return 0, err
		}
		a, b := args[0] != 0, args[1] != 0
		switch o.Type {
		case CBABoxLogicalAnd:
			v = cbaBoolWord(a && b)
		case CBABoxLogicalNand:
			v = cbaBoolWord(!(a && b))
		case CBABoxLogicalOr:
			v = cbaBoolWord(a || b)
		case CBABoxLogicalNor:
			v = cbaBoolWord(!(a || b))
		case CBABoxLogicalXor:
			v = cbaBoolWord(a != b)
		case CBABoxLogicalXnor:
			v = cbaBoolWord(a == b)
		}
	case CBABoxMux:
		if err := need(3); err != nil {
			return 0, err
		}
		if args[0] != 0 {
			v = args[1]
		} else {
			v = args[2]
		}
	case CBABoxMajority:
		if err := need(3); err != nil {
			return 0, err
		}
		v = args[0]&args[1] | args[0]&args[2] | args[1]&args[2]
	case CBABoxAdd, CBABoxSub, CBABoxMul, CBABoxSignedMul:
		if err := need(2); err != nil {
			return 0, err
		}
		if o.Type == CBABoxAdd {
			v = args[0] + args[1]
		} else if o.Type == CBABoxSub {
			v = args[0] - args[1]
		} else {
			v = args[0] * args[1]
		}
	case CBABoxDiv, CBABoxMod, CBABoxRem:
		if err := need(2); err != nil {
			return 0, err
		}
		if args[1] == 0 {
			return 0, nil
		}
		if o.Signed {
			a, b := cbaSigned(args[0], o.Width), cbaSigned(args[1], o.Width)
			if o.Type == CBABoxDiv {
				v = uint64(a / b)
			} else {
				v = uint64(a % b)
			}
		} else if o.Type == CBABoxDiv {
			v = args[0] / args[1]
		} else {
			v = args[0] % args[1]
		}
	case CBABoxPow:
		if err := need(2); err != nil {
			return 0, err
		}
		v = 1
		for base, exponent := args[0]&mask, args[1]; exponent != 0; exponent >>= 1 {
			if exponent&1 != 0 {
				v = v * base & mask
			}
			base = base * base & mask
		}
	case CBABoxMin:
		if err := need(1); err != nil {
			return 0, err
		}
		v = -args[0]
	case CBABoxSqrt:
		if err := need(1); err != nil {
			return 0, err
		}
		v = cbaIntegerSqrt(args[0] & cbaMask(o.ParamA))
	case CBABoxAbs:
		if err := need(1); err != nil {
			return 0, err
		}
		inputWidth := o.ParamA
		if inputWidth < 1 || inputWidth > 64 {
			inputWidth = o.Width
		}
		s := cbaSigned(args[0], inputWidth)
		if s < 0 {
			v = uint64(-s)
		} else {
			v = uint64(s)
		}
	case CBABoxLess, CBABoxLessEqual, CBABoxMoreEqual, CBABoxMore, CBABoxEqual, CBABoxNotEqual, CBABoxSignedLess:
		if err := need(2); err != nil {
			return 0, err
		}
		switch o.Type {
		case CBABoxLess:
			v = cbaBoolWord(args[0] < args[1])
		case CBABoxLessEqual:
			v = cbaBoolWord(args[0] <= args[1])
		case CBABoxMoreEqual:
			v = cbaBoolWord(args[0] >= args[1])
		case CBABoxMore:
			v = cbaBoolWord(args[0] > args[1])
		case CBABoxEqual:
			v = cbaBoolWord(args[0] == args[1])
		case CBABoxNotEqual:
			v = cbaBoolWord(args[0] != args[1])
		case CBABoxSignedLess:
			inputWidth := o.ParamA
			if inputWidth < 1 || inputWidth > 64 {
				inputWidth = o.Width
			}
			v = cbaBoolWord(cbaSigned(args[0], inputWidth) < cbaSigned(args[1], inputWidth))
		}
	case CBABoxShiftLeft, CBABoxArithmeticShiftLeft:
		if err := need(2); err != nil {
			return 0, err
		}
		if args[1] >= uint64(o.Width) {
			v = 0
		} else {
			v = args[0] << uint(args[1])
		}
	case CBABoxShiftRight:
		if err := need(2); err != nil {
			return 0, err
		}
		if args[1] >= uint64(o.Width) {
			v = 0
		} else {
			v = args[0] >> uint(args[1])
		}
	case CBABoxArithmeticShiftRight:
		if err := need(2); err != nil {
			return 0, err
		}
		if args[1] >= uint64(o.Width) {
			if cbaSigned(args[0], o.Width) < 0 {
				v = mask
			}
		} else {
			v = uint64(cbaSigned(args[0], o.Width) >> uint(args[1]))
		}
	case CBABoxRotateLeft:
		if err := need(2); err != nil {
			return 0, err
		}
		v = cbaRotateLeft(args[0], int(args[1]%uint64(o.Width)), o.Width)
	case CBABoxRotateRight:
		if err := need(2); err != nil {
			return 0, err
		}
		s := int(args[1] % uint64(o.Width))
		v = cbaRotateLeft(args[0], (o.Width-s)%o.Width, o.Width)
	case CBABoxSlice:
		if err := need(1); err != nil {
			return 0, err
		}
		lo := o.ParamA
		if o.ParamB < lo {
			lo = o.ParamB
		}
		v = args[0] >> uint(lo)
	case CBABoxConcat:
		return 0, ErrUnsupported
	case CBABoxTable, CBABoxLUT:
		if len(args) > 6 {
			return 0, ErrLimit
		}
		index := 0
		for i, a := range args {
			if a&1 != 0 {
				index |= 1 << uint(i)
			}
		}
		v = (o.Table >> uint(index)) & 1
	default:
		return 0, ErrUnsupported
	}
	return v & mask, nil
}

func cbaBoolWord(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}

// Evaluate computes all named outputs without mutating retained network state.
func (n *CBANetwork) Evaluate(inputs map[string]uint64) (map[string]uint64, error) {
	if err := n.Validate(); err != nil {
		return nil, err
	}
	b := cbaBudget{limits: n.limits}
	values := make([]uint64, len(n.objects))
	for id := 1; id < len(n.objects); id++ {
		if err := b.use(1 + len(n.objects[id].Inputs)); err != nil {
			return nil, err
		}
		o := n.objects[id]
		if o.Type == CBAObjectPI {
			value, ok := inputs[o.Name]
			if !ok {
				return nil, ErrInvariant
			}
			values[id] = value & cbaMask(o.Width)
			continue
		}
		var v uint64
		var err error
		if o.Type == CBABoxConcat {
			used := 0
			for i := len(o.Inputs) - 1; i >= 0; i-- {
				in := n.objects[o.Inputs[i]]
				if used > 64-in.Width {
					return nil, ErrLimit
				}
				v |= (values[in.ID] & cbaMask(in.Width)) << uint(used)
				used += in.Width
			}
		} else {
			v, err = cbaEvalObject(o, values)
		}
		if err != nil {
			return nil, err
		}
		values[id] = v
	}
	result := make(map[string]uint64, len(n.outputs))
	for _, id := range n.outputs {
		result[n.objects[id].Name] = values[id]
	}
	return result, nil
}

// CBAStats is an owned source-shaped structural summary.
type CBAStats struct {
	Modules, Objects, Inputs, Outputs, Edges, Bits int
	TypeCounts                                     [CBAObjectTypeCount]int
}

func (n *CBANetwork) Stats() (CBAStats, error) {
	if err := n.Validate(); err != nil {
		return CBAStats{}, err
	}
	s := CBAStats{Modules: 1, Objects: len(n.objects) - 1, Inputs: len(n.inputs), Outputs: len(n.outputs), Edges: n.edges, Bits: n.bits}
	for i := 1; i < len(n.objects); i++ {
		s.TypeCounts[n.objects[i].Type]++
	}
	return s, nil
}

func (n *CBANetwork) Close() {
	if n == nil {
		return
	}
	n.closed = true
	n.objects = nil
	n.inputs = nil
	n.outputs = nil
	n.names = nil
}

// CBADesign owns modules and root selection.
type CBADesign struct {
	Name     string
	networks []*CBANetwork
	byName   map[string]int
	root     int
	limits   CBALimits
	closed   bool
}

func NewCBADesign(name string, limits CBALimits) (*CBADesign, error) {
	limits, err := normalizeCBALimits(limits)
	if err != nil {
		return nil, err
	}
	if !cbaLegalName(name) || len(name) > limits.MaxBytes {
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &CBADesign{Name: name, networks: []*CBANetwork{nil}, byName: make(map[string]int), limits: limits}, nil
}

func (d *CBADesign) AddNetwork(n *CBANetwork) (int, error) {
	if d == nil || d.closed || n == nil {
		return 0, ErrInvariant
	}
	if len(d.networks) > d.limits.MaxModules {
		return 0, ErrLimit
	}
	if n.owner != nil || n.ID != 0 {
		return 0, ErrInvariant
	}
	if _, ok := d.byName[n.Name]; ok {
		return 0, ErrInvariant
	}
	if err := n.Validate(); err != nil {
		return 0, err
	}
	objects, edges, bitCount, byteCount := len(n.objects)-1, n.edges, n.bits, len(d.Name)+n.bytes
	for _, old := range d.networks[1:] {
		objects += len(old.objects) - 1
		edges += old.edges
		bitCount += old.bits
		if byteCount > d.limits.MaxBytes-old.bytes {
			return 0, ErrLimit
		}
		byteCount += old.bytes
	}
	if objects > d.limits.MaxObjects || edges > d.limits.MaxEdges || bitCount > d.limits.MaxBits || byteCount > d.limits.MaxBytes {
		return 0, ErrLimit
	}
	id := len(d.networks)
	n.ID = id
	n.owner = d
	d.networks = append(d.networks, n)
	d.byName[n.Name] = id
	if d.root == 0 {
		d.root = id
	}
	return id, nil
}
func (d *CBADesign) SetRoot(name string) error {
	if d == nil || d.closed {
		return ErrInvariant
	}
	id, ok := d.byName[name]
	if !ok {
		return ErrInvariant
	}
	d.root = id
	return nil
}
func (d *CBADesign) Root() (*CBANetwork, error) {
	if d == nil || d.closed || d.root <= 0 || d.root >= len(d.networks) {
		return nil, ErrInvariant
	}
	return d.networks[d.root], nil
}
func (d *CBADesign) Networks() ([]*CBANetwork, error) {
	if d == nil || d.closed {
		return nil, ErrInvariant
	}
	return slices.Clone(d.networks[1:]), nil
}
func (d *CBADesign) Validate() error {
	if d == nil || d.closed || d.root <= 0 || d.root >= len(d.networks) || len(d.byName) != len(d.networks)-1 {
		return ErrInvariant
	}
	objects, edges, bits, byteCount := 0, 0, 0, len(d.Name)
	for id := 1; id < len(d.networks); id++ {
		n := d.networks[id]
		if n == nil || n.ID != id || d.byName[n.Name] != id {
			return ErrInvariant
		}
		if err := n.Validate(); err != nil {
			return err
		}
		objects += len(n.objects) - 1
		edges += n.edges
		bits += n.bits
		if byteCount > d.limits.MaxBytes-n.bytes {
			return ErrLimit
		}
		byteCount += n.bytes
		if objects > d.limits.MaxObjects || edges > d.limits.MaxEdges || bits > d.limits.MaxBits {
			return ErrLimit
		}
	}
	return nil
}
func (d *CBADesign) Close() {
	if d == nil {
		return
	}
	for _, n := range d.networks[1:] {
		n.Close()
	}
	d.closed = true
	d.networks = nil
	d.byName = nil
}

// CBAEvaluatePrimitive exposes the reviewed word-level primitive semantics.
func CBAEvaluatePrimitive(typ CBAObjectType, width int, signed bool, args []uint64, paramA, paramB int) (uint64, error) {
	if width < 1 || width > 64 {
		return 0, ErrDimension
	}
	o := CBAObject{Type: typ, Width: width, Signed: signed, Inputs: make([]int, len(args)), ParamA: paramA, ParamB: paramB}
	values := make([]uint64, len(args)+1)
	for i, v := range args {
		o.Inputs[i] = i + 1
		values[i+1] = v
	}
	return cbaEvalObject(o, values)
}

// CBAWriteVerilog emits a deterministic structural subset accepted by CBAParseVerilog.
func CBAWriteVerilog(d *CBADesign, limits CBALimits) (string, error) {
	limits, err := normalizeCBALimits(limits)
	if err != nil {
		return "", err
	}
	if err = d.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	budget := cbaBudget{limits: limits}
	for _, n := range d.networks[1:] {
		if err = budget.use(len(n.objects)); err != nil {
			return "", err
		}
		ports := make([]string, 0, len(n.inputs)+len(n.outputs))
		for _, id := range n.inputs {
			ports = append(ports, n.objects[id].Name)
		}
		for _, id := range n.outputs {
			ports = append(ports, n.objects[id].Name)
		}
		fmt.Fprintf(&b, "module %s(%s);\n", n.Name, strings.Join(ports, ", "))
		for _, id := range n.inputs {
			cbaWriteDecl(&b, "input", n.objects[id])
		}
		for _, id := range n.outputs {
			cbaWriteDecl(&b, "output", n.objects[id])
		}
		outputSet := make(map[int]bool, len(n.outputs))
		for _, id := range n.outputs {
			outputSet[id] = true
		}
		for id := 1; id < len(n.objects); id++ {
			o := n.objects[id]
			if o.Type != CBAObjectPI && o.Type != CBAObjectPO {
				cbaWriteDecl(&b, "wire", o)
			}
		}
		for id := 1; id < len(n.objects); id++ {
			o := n.objects[id]
			if o.Type == CBAObjectPI || o.Type == CBAObjectPO {
				continue
			}
			expr, e := cbaVerilogExpr(n, o)
			if e != nil {
				return "", e
			}
			fmt.Fprintf(&b, "assign %s = %s;\n", o.Name, expr)
		}
		for _, id := range n.outputs {
			o := n.objects[id]
			fmt.Fprintf(&b, "assign %s = %s;\n", o.Name, n.objects[o.Inputs[0]].Name)
		}
		b.WriteString("endmodule\n\n")
		if b.Len() > limits.MaxTextBytes {
			return "", ErrLimit
		}
	}
	return b.String(), nil
}

func cbaWriteDecl(b *strings.Builder, kind string, o CBAObject) {
	fmt.Fprintf(b, "%s ", kind)
	if o.Signed {
		b.WriteString("signed ")
	}
	if o.Width > 1 {
		fmt.Fprintf(b, "[%d:0] ", o.Width-1)
	}
	fmt.Fprintf(b, "%s;\n", o.Name)
}

func cbaVerilogExpr(n *CBANetwork, o CBAObject) (string, error) {
	name := func(i int) string { return n.objects[o.Inputs[i]].Name }
	switch o.Type {
	case CBABoxConstFalse, CBABoxConstTrue, CBABoxBuf:
		if len(o.Inputs) == 0 {
			return fmt.Sprintf("%d'd%d", o.Width, o.Value&cbaMask(o.Width)), nil
		}
		return name(0), nil
	case CBABoxInv:
		return "~" + name(0), nil
	case CBABoxAnd:
		return "(" + name(0) + " & " + name(1) + ")", nil
	case CBABoxNand:
		return "~(" + name(0) + " & " + name(1) + ")", nil
	case CBABoxOr:
		return "(" + name(0) + " | " + name(1) + ")", nil
	case CBABoxNor:
		return "~(" + name(0) + " | " + name(1) + ")", nil
	case CBABoxXor:
		return "(" + name(0) + " ^ " + name(1) + ")", nil
	case CBABoxXnor:
		return "~(" + name(0) + " ^ " + name(1) + ")", nil
	case CBABoxMux:
		return "(" + name(0) + " ? " + name(1) + " : " + name(2) + ")", nil
	case CBABoxAdd:
		return "(" + name(0) + " + " + name(1) + ")", nil
	case CBABoxSub:
		return "(" + name(0) + " - " + name(1) + ")", nil
	case CBABoxMul:
		return "(" + name(0) + " * " + name(1) + ")", nil
	case CBABoxDiv:
		return "(" + name(0) + " / " + name(1) + ")", nil
	case CBABoxMod:
		return "(" + name(0) + " % " + name(1) + ")", nil
	case CBABoxLess:
		return "(" + name(0) + " < " + name(1) + ")", nil
	case CBABoxEqual:
		return "(" + name(0) + " == " + name(1) + ")", nil
	case CBABoxNotEqual:
		return "(" + name(0) + " != " + name(1) + ")", nil
	case CBABoxShiftLeft:
		return "(" + name(0) + " << " + name(1) + ")", nil
	case CBABoxShiftRight:
		return "(" + name(0) + " >> " + name(1) + ")", nil
	case CBABoxSlice:
		return fmt.Sprintf("%s[%d:%d]", name(0), o.ParamA, o.ParamB), nil
	default:
		return "", ErrUnsupported
	}
}

// CBAParseVerilog accepts deterministic module/declaration/assign syntax.
func CBAParseVerilog(text string, limits CBALimits) (*CBADesign, error) {
	limits, err := normalizeCBALimits(limits)
	if err != nil {
		return nil, err
	}
	if len(text) > limits.MaxTextBytes {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	clean := cbaStripLineComments(text)
	design, err := NewCBADesign("cba", limits)
	if err != nil {
		return nil, err
	}
	rest := clean
	for {
		rest = strings.TrimSpace(rest)
		if rest == "" {
			break
		}
		if !strings.HasPrefix(rest, "module ") {
			return nil, ErrFormat
		}
		rest = rest[len("module "):]
		end := strings.Index(rest, "endmodule")
		if end < 0 {
			return nil, ErrFormat
		}
		block := rest[:end]
		rest = rest[end+len("endmodule"):]
		n, e := cbaParseVerilogModule(block, limits)
		if e != nil {
			return nil, e
		}
		if _, e = design.AddNetwork(n); e != nil {
			return nil, e
		}
		if len(design.networks)-1 > limits.MaxModules {
			return nil, ErrLimit
		}
	}
	if len(design.networks) == 1 {
		return nil, ErrFormat
	}
	design.Name = design.networks[1].Name
	if err = design.SetRoot(design.networks[len(design.networks)-1].Name); err != nil {
		return nil, err
	}
	if err = design.Validate(); err != nil {
		return nil, err
	}
	return design, nil
}

func cbaStripLineComments(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if p := strings.Index(line, "//"); p >= 0 {
			lines[i] = line[:p]
		}
	}
	return strings.Join(lines, "\n")
}

func cbaParseVerilogModule(block string, limits CBALimits) (*CBANetwork, error) {
	semi := strings.IndexByte(block, ';')
	if semi < 0 {
		return nil, ErrFormat
	}
	head := strings.TrimSpace(block[:semi])
	p := strings.IndexByte(head, '(')
	name := head
	if p >= 0 {
		name = strings.TrimSpace(head[:p])
	}
	n, err := NewCBANetwork(name, limits)
	if err != nil {
		return nil, err
	}
	widths := make(map[string]int)
	signed := make(map[string]bool)
	outputs := make(map[string]bool)
	declarations := make(map[string]string)
	statements := strings.Split(block[semi+1:], ";")
	budget := cbaBudget{limits: limits}
	for _, raw := range statements {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		if err = budget.use(1); err != nil {
			return nil, err
		}
		kind := ""
		for _, k := range []string{"input", "output", "wire"} {
			if strings.HasPrefix(s, k+" ") {
				kind = k
				s = strings.TrimSpace(s[len(k):])
				break
			}
		}
		if kind != "" {
			sg := false
			if strings.HasPrefix(s, "signed ") {
				sg = true
				s = strings.TrimSpace(s[7:])
			}
			w := 1
			if strings.HasPrefix(s, "[") {
				q := strings.IndexByte(s, ']')
				if q < 0 {
					return nil, ErrFormat
				}
				parts := strings.Split(s[1:q], ":")
				if len(parts) != 2 {
					return nil, ErrFormat
				}
				a, e1 := strconv.Atoi(strings.TrimSpace(parts[0]))
				b, e2 := strconv.Atoi(strings.TrimSpace(parts[1]))
				if e1 != nil || e2 != nil {
					return nil, ErrFormat
				}
				w, err = (CBARange{Left: a, Right: b}).Width()
				if err != nil {
					return nil, err
				}
				s = strings.TrimSpace(s[q+1:])
			}
			for _, nm := range strings.Split(s, ",") {
				nm = strings.TrimSpace(nm)
				if !cbaLegalName(nm) {
					return nil, ErrFormat
				}
				if _, ok := declarations[nm]; ok {
					return nil, ErrInvariant
				}
				declarations[nm] = kind
				widths[nm] = w
				signed[nm] = sg
				if kind == "input" {
					if _, e := n.AddInput(nm, w, sg); e != nil {
						return nil, e
					}
				} else if kind == "output" {
					outputs[nm] = true
				}
			}
			continue
		}
		if strings.HasPrefix(s, "assign ") {
			eq := strings.IndexByte(s, '=')
			if eq < 0 {
				return nil, ErrFormat
			}
			lhs := strings.TrimSpace(s[len("assign "):eq])
			rhs := strings.TrimSpace(s[eq+1:])
			w := widths[lhs]
			if w == 0 {
				return nil, ErrInvariant
			}
			driverName := lhs
			if outputs[lhs] {
				driverName = cbaUniqueDriverName(n, lhs)
			}
			source, e := cbaParseCanonicalExpr(n, driverName, rhs, w, signed[lhs])
			if e != nil {
				return nil, e
			}
			if outputs[lhs] {
				if _, e = n.AddOutput(lhs, w, source); e != nil {
					return nil, e
				}
			}
			continue
		}
		return nil, ErrUnsupported
	}
	for out := range outputs {
		if _, ok := n.names[out]; !ok {
			return nil, ErrInvariant
		}
	}
	return n, nil
}

func cbaUniqueDriverName(n *CBANetwork, output string) string {
	base := "_cba_drv_" + output
	name := base
	for i := 1; ; i++ {
		if _, exists := n.names[name]; !exists {
			return name
		}
		name = base + "_" + strconv.Itoa(i)
	}
}

func cbaParseCanonicalExpr(n *CBANetwork, lhs, rhs string, width int, sg bool) (int, error) {
	if id, ok := n.names[rhs]; ok {
		return id, nil
	}
	if strings.HasPrefix(rhs, "~") {
		inner := strings.TrimSpace(strings.TrimPrefix(rhs, "~"))
		inner = strings.Trim(inner, "()")
		if id, ok := n.names[inner]; ok {
			return n.AddNode(lhs, CBABoxInv, width, sg, []int{id})
		}
		for _, op := range []struct {
			s string
			t CBAObjectType
		}{{" & ", CBABoxNand}, {" | ", CBABoxNor}, {" ^ ", CBABoxXnor}} {
			if q := strings.Index(inner, op.s); q >= 0 {
				ids, ok := cbaLookupNames(n, []string{strings.TrimSpace(inner[:q]), strings.TrimSpace(inner[q+len(op.s):])})
				if !ok {
					return 0, ErrInvariant
				}
				return n.AddNode(lhs, op.t, width, sg, ids)
			}
		}
	}
	if q := strings.Index(rhs, "'d"); q > 0 {
		w, e1 := strconv.Atoi(rhs[:q])
		v, e2 := strconv.ParseUint(rhs[q+2:], 10, 64)
		if e1 != nil || e2 != nil || w != width {
			return 0, ErrFormat
		}
		return n.AddConstant(lhs, width, v)
	}
	if lb := strings.IndexByte(rhs, '['); lb > 0 && strings.HasSuffix(rhs, "]") {
		id, ok := n.names[strings.TrimSpace(rhs[:lb])]
		if !ok {
			return 0, ErrInvariant
		}
		parts := strings.Split(rhs[lb+1:len(rhs)-1], ":")
		if len(parts) != 2 {
			return 0, ErrFormat
		}
		a, e1 := strconv.Atoi(parts[0])
		b, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil {
			return 0, ErrFormat
		}
		return n.AddSlice(lhs, id, a, b)
	}
	expr := strings.TrimSpace(rhs)
	for strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		expr = strings.TrimSpace(expr[1 : len(expr)-1])
	}
	if q := strings.Index(expr, " ? "); q >= 0 {
		c := strings.TrimSpace(expr[:q])
		tail := expr[q+3:]
		r := strings.Index(tail, " : ")
		if r < 0 {
			return 0, ErrFormat
		}
		a, b := strings.TrimSpace(tail[:r]), strings.TrimSpace(tail[r+3:])
		ids, ok := cbaLookupNames(n, []string{c, a, b})
		if !ok {
			return 0, ErrInvariant
		}
		return n.AddNode(lhs, CBABoxMux, width, sg, ids)
	}
	ops := []struct {
		s string
		t CBAObjectType
	}{{" == ", CBABoxEqual}, {" != ", CBABoxNotEqual}, {" << ", CBABoxShiftLeft}, {" >> ", CBABoxShiftRight}, {" & ", CBABoxAnd}, {" | ", CBABoxOr}, {" ^ ", CBABoxXor}, {" + ", CBABoxAdd}, {" - ", CBABoxSub}, {" * ", CBABoxMul}, {" / ", CBABoxDiv}, {" % ", CBABoxMod}, {" < ", CBABoxLess}}
	for _, op := range ops {
		if q := strings.Index(expr, op.s); q >= 0 {
			a, b := strings.TrimSpace(expr[:q]), strings.TrimSpace(expr[q+len(op.s):])
			ids, ok := cbaLookupNames(n, []string{a, b})
			if !ok {
				return 0, ErrInvariant
			}
			return n.AddNode(lhs, op.t, width, sg, ids)
		}
	}
	return 0, ErrUnsupported
}
func cbaLookupNames(n *CBANetwork, names []string) ([]int, bool) {
	ids := make([]int, len(names))
	for i, name := range names {
		var ok bool
		ids[i], ok = n.names[name]
		if !ok {
			return nil, false
		}
	}
	return ids, true
}

// CBAWriteBLIF emits Boolean networks with source-shaped .names tables.
func CBAWriteBLIF(d *CBADesign, limits CBALimits) (string, error) {
	limits, err := normalizeCBALimits(limits)
	if err != nil {
		return "", err
	}
	if err = d.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	budget := cbaBudget{limits: limits}
	for _, n := range d.networks[1:] {
		for id := 1; id < len(n.objects); id++ {
			if n.objects[id].Width != 1 {
				return "", ErrUnsupported
			}
		}
		fmt.Fprintf(&b, ".model %s\n", n.Name)
		b.WriteString(".inputs")
		for _, id := range n.inputs {
			fmt.Fprintf(&b, " %s", n.objects[id].Name)
		}
		b.WriteByte('\n')
		b.WriteString(".outputs")
		for _, id := range n.outputs {
			fmt.Fprintf(&b, " %s", n.objects[id].Name)
		}
		b.WriteByte('\n')
		for id := 1; id < len(n.objects); id++ {
			o := n.objects[id]
			if o.Type == CBAObjectPI || o.Type == CBAObjectPO {
				continue
			}
			if len(o.Inputs) > 6 {
				return "", ErrLimit
			}
			fmt.Fprint(&b, ".names")
			for _, in := range o.Inputs {
				fmt.Fprintf(&b, " %s", n.objects[in].Name)
			}
			fmt.Fprintf(&b, " %s\n", o.Name)
			table, e := cbaTruthTable(o)
			if e != nil {
				return "", e
			}
			for m := 0; m < 1<<uint(len(o.Inputs)); m++ {
				if err = budget.use(1); err != nil {
					return "", err
				}
				if table>>uint(m)&1 == 0 {
					continue
				}
				for i := 0; i < len(o.Inputs); i++ {
					if m>>uint(i)&1 != 0 {
						b.WriteByte('1')
					} else {
						b.WriteByte('0')
					}
				}
				if len(o.Inputs) > 0 {
					b.WriteByte(' ')
				}
				b.WriteString("1\n")
			}
		}
		for _, id := range n.outputs {
			o := n.objects[id]
			fmt.Fprintf(&b, ".names %s %s\n1 1\n", n.objects[o.Inputs[0]].Name, o.Name)
		}
		b.WriteString(".end\n\n")
		if b.Len() > limits.MaxTextBytes {
			return "", ErrLimit
		}
	}
	return b.String(), nil
}

func cbaTruthTable(o CBAObject) (uint64, error) {
	if o.Type == CBABoxTable || o.Type == CBABoxLUT {
		return o.Table, nil
	}
	if len(o.Inputs) > 6 {
		return 0, ErrLimit
	}
	var table uint64
	values := make([]uint64, len(o.Inputs)+1)
	for m := 0; m < 1<<uint(len(o.Inputs)); m++ {
		copyO := o
		copyO.Inputs = make([]int, len(o.Inputs))
		for i := range o.Inputs {
			copyO.Inputs[i] = i + 1
			values[i+1] = uint64(m >> uint(i) & 1)
		}
		v, err := cbaEvalObject(copyO, values)
		if err != nil {
			return 0, err
		}
		if v&1 != 0 {
			table |= uint64(1) << uint(m)
		}
	}
	return table, nil
}

// CBAParseBLIF accepts .model/.inputs/.outputs/.names/.end with on-set cubes.
func CBAParseBLIF(text string, limits CBALimits) (*CBADesign, error) {
	limits, err := normalizeCBALimits(limits)
	if err != nil {
		return nil, err
	}
	if len(text) > limits.MaxTextBytes {
		return nil, ErrLimit
	}
	design, err := NewCBADesign("cba", limits)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 1024), limits.MaxTextBytes)
	var n *CBANetwork
	outputs := make(map[string]bool)
	var names []string
	var table uint64
	cubeCount := 0
	budget := cbaBudget{limits: limits}
	flush := func() error {
		if names == nil {
			return nil
		}
		if len(names) < 1 {
			return ErrFormat
		}
		inputNames := names[:len(names)-1]
		out := names[len(names)-1]
		ids, ok := cbaLookupNames(n, inputNames)
		if !ok {
			return ErrInvariant
		}
		var id int
		var e error
		if len(ids) == 1 && table == 2 {
			id = ids[0]
		} else {
			objectName := out
			if outputs[out] {
				objectName = cbaUniqueDriverName(n, out)
			}
			id, e = n.AddTable(objectName, ids, table)
			if e != nil {
				return e
			}
		}
		if outputs[out] {
			_, e = n.AddOutput(out, 1, id)
			if e != nil {
				return e
			}
		}
		names = nil
		table = 0
		cubeCount = 0
		return nil
	}
	for scanner.Scan() {
		if err = budget.use(1); err != nil {
			return nil, err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, ".") {
			if err = flush(); err != nil {
				return nil, err
			}
			fields := strings.Fields(line)
			switch fields[0] {
			case ".model":
				if len(fields) != 2 || n != nil {
					return nil, ErrFormat
				}
				n, err = NewCBANetwork(fields[1], limits)
				if err != nil {
					return nil, err
				}
			case ".inputs":
				if n == nil {
					return nil, ErrFormat
				}
				for _, name := range fields[1:] {
					if _, err = n.AddInput(name, 1, false); err != nil {
						return nil, err
					}
				}
			case ".outputs":
				if n == nil {
					return nil, ErrFormat
				}
				for _, name := range fields[1:] {
					if !cbaLegalName(name) {
						return nil, ErrFormat
					}
					outputs[name] = true
				}
			case ".names":
				if n == nil || len(fields) < 2 {
					return nil, ErrFormat
				}
				names = slices.Clone(fields[1:])
			case ".end":
				if n == nil {
					return nil, ErrFormat
				}
				for name := range outputs {
					id, ok := n.names[name]
					if !ok || n.objects[id].Type != CBAObjectPO {
						return nil, ErrInvariant
					}
				}
				if _, err = design.AddNetwork(n); err != nil {
					return nil, err
				}
				n = nil
				outputs = make(map[string]bool)
			default:
				return nil, ErrUnsupported
			}
			continue
		}
		if names == nil {
			return nil, ErrFormat
		}
		fields := strings.Fields(line)
		if len(fields) < 1 || len(fields) > 2 {
			return nil, ErrFormat
		}
		pattern := fields[0]
		out := "1"
		if len(fields) == 2 {
			out = fields[1]
		} else if len(names) == 1 && fields[0] == "1" {
			pattern = ""
		}
		if out != "1" || len(pattern) != len(names)-1 {
			return nil, ErrUnsupported
		}
		if len(pattern) > 6 {
			return nil, ErrLimit
		}
		for _, c := range pattern {
			if c != '0' && c != '1' && c != '-' {
				return nil, ErrFormat
			}
		}
		for m := 0; m < 1<<uint(len(pattern)); m++ {
			if err = budget.use(1); err != nil {
				return nil, err
			}
			match := true
			for i, c := range pattern {
				if c != '-' && int(c-'0') != m>>uint(i)&1 {
					match = false
					break
				}
			}
			if match {
				table |= uint64(1) << uint(m)
			}
		}
		cubeCount++
		if cubeCount > limits.MaxOperations {
			return nil, ErrLimit
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	if err = flush(); err != nil {
		return nil, err
	}
	if n != nil || len(design.networks) == 1 {
		return nil, ErrFormat
	}
	design.Name = design.networks[1].Name
	if err = design.SetRoot(design.networks[len(design.networks)-1].Name); err != nil {
		return nil, err
	}
	if err = design.Validate(); err != nil {
		return nil, err
	}
	return design, nil
}

// CBABuildInventory is the exact active module.make source order.
var CBABuildInventory = [...]string{"cbaBlast.c", "cbaCba.c", "cbaCom.c", "cbaNtk.c", "cbaReadBlif.c", "cbaReadVer.c", "cbaWriteBlif.c", "cbaWriteVer.c"}

// Coverage inventory: the concrete model covers owned design/network/object
// storage, stable IDs, source primitive numbering, ranges, topology/cycle
// checking, structural statistics, deterministic ordering, word-level primitive
// evaluation, bounded bit-vector semantics, Boolean BLIF round trips, and a
// deterministic structural Verilog subset. Native ABC frame commands, complete
// SystemVerilog grammar, mapped-cell libraries, sequential blasting, GIA/ABC
// insertion, arbitrary attributes, multi-output primitive layout, collapse and
// group rewriting remain outside this standalone adaptation.
