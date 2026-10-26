package abc

// Derived from ABC src/base/bac/bac.h, bacPrs.h, bacBac.c, bacBlast.c,
// bacCom.c, bacNtk.c, bacOper.c, bacPrsBuild.c, bacPrsTrans.c, bacPtr.c,
// bacPtrAbc.c, bacReadBlif.c, bacReadVer.c, bacWriteBlif.c, and bacWriteVer.c.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
)

var ErrBACClosed = errors.New("BAC design is closed")

// BACType is the owned counterpart of Bac_ObjType_t. The numeric order keeps
// the source primitive families contiguous.
type BACType uint8

const (
	BACNone BACType = iota
	BACInput
	BACConst0
	BACConst1
	BACConstX
	BACConstZ
	BACBuffer
	BACInvert
	BACAnd
	BACNand
	BACOr
	BACNor
	BACXor
	BACXnor
	BACMux
	BACMajority
	BACReduceAnd
	BACReduceNand
	BACReduceOr
	BACReduceNor
	BACReduceXor
	BACReduceXnor
	BACLogicAnd
	BACLogicNand
	BACLogicOr
	BACLogicNor
	BACLogicXor
	BACLogicXnor
	BACAdd
	BACSubtract
	BACMultiply
	BACDivide
	BACModulo
	BACRemainder
	BACPower
	BACMinimum
	BACAbsolute
	BACLess
	BACLessEqual
	BACGreaterEqual
	BACGreater
	BACEqual
	BACNotEqual
	BACShiftLeft
	BACShiftRight
	BACRotateLeft
	BACRotateRight
	BACUser
	BACUnknown
)

var bacTypeNames = [...]string{
	"none", "input", "c0", "c1", "cx", "cz", "buf", "inv", "and", "nand",
	"or", "nor", "xor", "xnor", "mux", "maj", "rand", "rnand", "ror", "rnor",
	"rxor", "rxnor", "land", "lnand", "lor", "lnor", "lxor", "lxnor", "add",
	"sub", "mul", "div", "mod", "rem", "pow", "min", "abs", "lt", "le", "ge",
	"gt", "eq", "ne", "shl", "shr", "rotl", "rotr", "user", "unknown",
}

func (t BACType) String() string {
	if int(t) < len(bacTypeNames) {
		return bacTypeNames[t]
	}
	return "unknown"
}

func bacTypeByName(s string) (BACType, bool) {
	for i := 1; i < len(bacTypeNames); i++ {
		if bacTypeNames[i] == s {
			return BACType(i), true
		}
	}
	return BACNone, false
}

// BACLimits bounds aggregate retained state and work. Zero selects a safe
// default. Widths are capped at 64 because evaluation uses explicit words.
type BACLimits struct {
	MaxModules    int
	MaxObjects    int
	MaxEdges      int
	MaxWidth      int
	MaxBytes      int
	MaxTextBytes  int
	MaxOperations int
	MaxDepth      int
	Stop          func() bool
}

func (l BACLimits) normalized() BACLimits {
	if l.MaxModules <= 0 {
		l.MaxModules = 1024
	}
	if l.MaxObjects <= 0 {
		l.MaxObjects = 1 << 20
	}
	if l.MaxEdges <= 0 {
		l.MaxEdges = 1 << 22
	}
	if l.MaxWidth <= 0 || l.MaxWidth > 64 {
		l.MaxWidth = 64
	}
	if l.MaxBytes <= 0 {
		l.MaxBytes = 1 << 26
	}
	if l.MaxTextBytes <= 0 {
		l.MaxTextBytes = 1 << 24
	}
	if l.MaxOperations <= 0 {
		l.MaxOperations = 1 << 25
	}
	if l.MaxDepth <= 0 {
		l.MaxDepth = 256
	}
	return l
}

type bacBudget struct {
	limits            BACLimits
	operations, bytes int
}

func (b *bacBudget) take(operations, byteCount int) error {
	if operations < 0 || byteCount < 0 || operations > b.limits.MaxOperations-b.operations || byteCount > b.limits.MaxBytes-b.bytes {
		return ErrLimit
	}
	b.operations += operations
	b.bytes += byteCount
	if b.limits.Stop != nil && b.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func bacCheckedAdd(total *int, values ...int) error {
	for _, value := range values {
		if value < 0 || *total > int(^uint(0)>>1)-value {
			return ErrLimit
		}
		*total += value
	}
	return nil
}

// BACObject is an immutable-input, stable-ID word-level object.
type BACObject struct {
	ID     int
	Type   BACType
	Name   string
	Width  int
	Inputs []int
	Value  uint64
	Module string
}

func (o BACObject) clone() BACObject { o.Inputs = append([]int(nil), o.Inputs...); return o }

type BACOutput struct {
	Name   string
	Driver int
	Width  int
}

// BACNetwork owns one topologically ordered module.
type BACNetwork struct {
	Name         string
	objects      []BACObject
	inputs       []int
	outputs      []BACOutput
	nameToObject map[string]int
}

func newBACNetwork(name string) *BACNetwork {
	return &BACNetwork{Name: name, nameToObject: make(map[string]int)}
}

func (n *BACNetwork) clone() *BACNetwork {
	r := newBACNetwork(n.Name)
	r.objects = make([]BACObject, len(n.objects))
	for i := range n.objects {
		r.objects[i] = n.objects[i].clone()
	}
	r.inputs = append([]int(nil), n.inputs...)
	r.outputs = append([]BACOutput(nil), n.outputs...)
	for k, v := range n.nameToObject {
		r.nameToObject[k] = v
	}
	return r
}

func (n *BACNetwork) Objects() []BACObject {
	r := make([]BACObject, len(n.objects))
	for i := range n.objects {
		r[i] = n.objects[i].clone()
	}
	return r
}
func (n *BACNetwork) Inputs() []int        { return append([]int(nil), n.inputs...) }
func (n *BACNetwork) Outputs() []BACOutput { return append([]BACOutput(nil), n.outputs...) }

// BACDesign is an owned hierarchical design. Module order is deterministic.
type BACDesign struct {
	Name                      string
	Spec                      string
	Root                      string
	modules                   []*BACNetwork
	moduleByName              map[string]int
	limits                    BACLimits
	objects, edges, nameBytes int
	closed                    bool
}

func NewBACDesign(name string, limits BACLimits) (*BACDesign, error) {
	limits = limits.normalized()
	if name == "" || len(name) > limits.MaxTextBytes {
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &BACDesign{Name: name, limits: limits, moduleByName: make(map[string]int), nameBytes: len(name)}, nil
}

func (d *BACDesign) check() error {
	if d == nil || d.closed {
		return ErrBACClosed
	}
	if d.limits.Stop != nil && d.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func (d *BACDesign) Close() {
	if d != nil {
		d.closed = true
		d.modules = nil
		d.moduleByName = nil
		d.objects = 0
		d.edges = 0
	}
}

func (d *BACDesign) ModuleNames() []string {
	if d == nil || d.closed {
		return nil
	}
	r := make([]string, len(d.modules))
	for i := range d.modules {
		r[i] = d.modules[i].Name
	}
	return r
}

func (d *BACDesign) Module(name string) (*BACNetwork, bool) {
	if d == nil || d.closed {
		return nil, false
	}
	i, ok := d.moduleByName[name]
	if !ok {
		return nil, false
	}
	return d.modules[i].clone(), true
}

func (d *BACDesign) addModuleOwned(n *BACNetwork) error {
	if err := d.check(); err != nil {
		return err
	}
	if n == nil || n.Name == "" {
		return ErrDimension
	}
	if _, ok := d.moduleByName[n.Name]; ok {
		return ErrInvariant
	}
	if len(d.modules) >= d.limits.MaxModules || d.nameBytes > d.limits.MaxTextBytes-len(n.Name) {
		return ErrLimit
	}
	if err := validateBACNetwork(n, d.limits); err != nil {
		return err
	}
	budget := bacBudget{limits: d.limits}
	edges := len(n.outputs)
	names := len(n.Name)
	for _, o := range n.objects {
		if err := budget.take(1+len(o.Inputs), 0); err != nil {
			return err
		}
		edges += len(o.Inputs)
		names += len(o.Name)
	}
	for _, o := range n.outputs {
		if err := budget.take(1, 0); err != nil {
			return err
		}
		names += len(o.Name)
	}
	if len(n.objects) > d.limits.MaxObjects-d.objects || edges > d.limits.MaxEdges-d.edges || names > d.limits.MaxTextBytes-d.nameBytes {
		return ErrLimit
	}
	d.moduleByName[n.Name] = len(d.modules)
	d.modules = append(d.modules, n.clone())
	d.objects += len(n.objects)
	d.edges += edges
	d.nameBytes += names
	if d.Root == "" {
		d.Root = n.Name
	}
	return nil
}

func (d *BACDesign) AddNetwork(name string) (*BACBuilder, error) {
	if err := d.check(); err != nil {
		return nil, err
	}
	if name == "" || len(name) > d.limits.MaxTextBytes {
		return nil, ErrDimension
	}
	if _, ok := d.moduleByName[name]; ok {
		return nil, ErrInvariant
	}
	return &BACBuilder{design: d, network: newBACNetwork(name), nameBytes: len(name)}, nil
}

// BACBuilder accumulates one module and commits it atomically.
type BACBuilder struct {
	design    *BACDesign
	network   *BACNetwork
	edges     int
	nameBytes int
	committed bool
}

func (b *BACBuilder) addObject(kind BACType, name string, width int, inputs []int, value uint64, module string) (int, error) {
	if b == nil || b.design == nil || b.network == nil || b.committed {
		return -1, ErrInvariant
	}
	if err := b.design.check(); err != nil {
		return -1, err
	}
	if name == "" || width < 1 || width > b.design.limits.MaxWidth || kind <= BACNone || kind >= BACUnknown {
		return -1, ErrDimension
	}
	if _, exists := b.network.nameToObject[name]; exists {
		return -1, ErrInvariant
	}
	if len(inputs) > b.design.limits.MaxEdges || len(b.network.objects)+b.design.objects >= b.design.limits.MaxObjects || b.edges > b.design.limits.MaxEdges-b.design.edges-len(inputs) {
		return -1, ErrLimit
	}
	if len(name) > b.design.limits.MaxTextBytes-b.design.nameBytes-b.nameBytes {
		return -1, ErrLimit
	}
	for _, id := range inputs {
		if id < 0 || id >= len(b.network.objects) {
			return -1, ErrInvariant
		}
	}
	o := BACObject{ID: len(b.network.objects), Type: kind, Name: name, Width: width, Inputs: append([]int(nil), inputs...), Value: value, Module: module}
	if err := validateBACObject(o, b.network.objects); err != nil {
		return -1, err
	}
	b.network.nameToObject[name] = o.ID
	b.network.objects = append(b.network.objects, o)
	b.edges += len(inputs)
	b.nameBytes += len(name)
	if kind == BACInput {
		b.network.inputs = append(b.network.inputs, o.ID)
	}
	return o.ID, nil
}

func (b *BACBuilder) AddInput(name string, width int) (int, error) {
	return b.addObject(BACInput, name, width, nil, 0, "")
}
func (b *BACBuilder) AddConstant(name string, width int, value uint64) (int, error) {
	return b.addObject(BACConst0, name, width, nil, value, "")
}
func (b *BACBuilder) AddNode(kind BACType, name string, width int, inputs ...int) (int, error) {
	return b.addObject(kind, name, width, inputs, 0, "")
}
func (b *BACBuilder) AddInstance(module, name string, width int, inputs ...int) (int, error) {
	return b.addObject(BACUser, name, width, inputs, 0, module)
}

func (b *BACBuilder) AddOutput(name string, driver int) error {
	if b == nil || b.network == nil || b.committed {
		return ErrInvariant
	}
	if err := b.design.check(); err != nil {
		return err
	}
	if name == "" || driver < 0 || driver >= len(b.network.objects) {
		return ErrDimension
	}
	for _, p := range b.network.outputs {
		if p.Name == name {
			return ErrInvariant
		}
	}
	if b.edges >= b.design.limits.MaxEdges-b.design.edges {
		return ErrLimit
	}
	if len(name) > b.design.limits.MaxTextBytes-b.design.nameBytes-b.nameBytes {
		return ErrLimit
	}
	b.network.outputs = append(b.network.outputs, BACOutput{Name: name, Driver: driver, Width: b.network.objects[driver].Width})
	b.edges++
	b.nameBytes += len(name)
	return nil
}

func (b *BACBuilder) Commit() error {
	if b == nil || b.committed {
		return ErrInvariant
	}
	if err := b.design.addModuleOwned(b.network); err != nil {
		return err
	}
	b.committed = true
	return nil
}

func validateBACObject(o BACObject, prior []BACObject) error {
	if o.ID != len(prior) || o.Name == "" || o.Width < 1 || o.Width > 64 {
		return ErrInvariant
	}
	width := func(i int) int { return prior[o.Inputs[i]].Width }
	for _, id := range o.Inputs {
		if id < 0 || id >= len(prior) {
			return ErrInvariant
		}
	}
	same := func() bool {
		for i := range o.Inputs {
			if width(i) != o.Width {
				return false
			}
		}
		return true
	}
	switch o.Type {
	case BACInput, BACConst0, BACConst1:
		if len(o.Inputs) != 0 {
			return ErrDimension
		}
	case BACConstX, BACConstZ:
		return ErrDimension
	case BACBuffer, BACInvert, BACAbsolute:
		if len(o.Inputs) != 1 || width(0) != o.Width {
			return ErrDimension
		}
	case BACAnd, BACNand, BACOr, BACNor, BACXor, BACXnor, BACAdd, BACSubtract, BACMultiply, BACMinimum:
		if len(o.Inputs) != 2 || !same() {
			return ErrDimension
		}
	case BACDivide, BACModulo, BACRemainder, BACPower, BACShiftLeft, BACShiftRight, BACRotateLeft, BACRotateRight:
		if len(o.Inputs) != 2 || width(0) != o.Width {
			return ErrDimension
		}
	case BACMux:
		if len(o.Inputs) != 3 || width(0) != 1 || width(1) != o.Width || width(2) != o.Width {
			return ErrDimension
		}
	case BACMajority:
		if len(o.Inputs) != 3 || !same() {
			return ErrDimension
		}
	case BACReduceAnd, BACReduceNand, BACReduceOr, BACReduceNor, BACReduceXor, BACReduceXnor:
		if len(o.Inputs) != 1 || o.Width != 1 {
			return ErrDimension
		}
	case BACLogicAnd, BACLogicNand, BACLogicOr, BACLogicNor, BACLogicXor, BACLogicXnor:
		if len(o.Inputs) != 2 || o.Width != 1 {
			return ErrDimension
		}
	case BACLess, BACLessEqual, BACGreaterEqual, BACGreater, BACEqual, BACNotEqual:
		if len(o.Inputs) != 2 || width(0) != width(1) || o.Width != 1 {
			return ErrDimension
		}
	case BACUser:
		if o.Module == "" {
			return ErrDimension
		}
	default:
		return ErrDimension
	}
	return nil
}

func validateBACNetwork(n *BACNetwork, limits BACLimits) error {
	if n == nil || n.Name == "" || len(n.objects) > limits.MaxObjects {
		return ErrInvariant
	}
	if len(n.nameToObject) != len(n.objects) {
		return ErrInvariant
	}
	edges := len(n.outputs)
	inputSeen := make(map[int]bool, len(n.inputs))
	for _, id := range n.inputs {
		if id < 0 || id >= len(n.objects) || n.objects[id].Type != BACInput || inputSeen[id] {
			return ErrInvariant
		}
		inputSeen[id] = true
	}
	outputSeen := make(map[string]bool, len(n.outputs))
	for i, o := range n.objects {
		id, ok := n.nameToObject[o.Name]
		if o.ID != i || !ok || id != i || (o.Type == BACInput) != inputSeen[i] {
			return ErrInvariant
		}
		if err := validateBACObject(o, n.objects[:i]); err != nil {
			return err
		}
		edges += len(o.Inputs)
		if edges > limits.MaxEdges {
			return ErrLimit
		}
	}
	for _, p := range n.outputs {
		if p.Driver < 0 || p.Driver >= len(n.objects) || p.Width != n.objects[p.Driver].Width || p.Name == "" || outputSeen[p.Name] {
			return ErrInvariant
		}
		outputSeen[p.Name] = true
	}
	return nil
}

func bacMask(width int) uint64 {
	if width >= 64 {
		return math.MaxUint64
	}
	return (uint64(1) << width) - 1
}

func evalBACObject(o BACObject, objects []BACObject, values []uint64, budget *bacBudget) (uint64, error) {
	m := bacMask(o.Width)
	in := func(i int) uint64 { return values[o.Inputs[i]] }
	var v uint64
	switch o.Type {
	case BACConst0:
		v = o.Value
	case BACConst1:
		v = m
	case BACBuffer:
		v = in(0)
	case BACInvert:
		v = ^in(0)
	case BACAnd:
		v = in(0) & in(1)
	case BACNand:
		v = ^(in(0) & in(1))
	case BACOr:
		v = in(0) | in(1)
	case BACNor:
		v = ^(in(0) | in(1))
	case BACXor:
		v = in(0) ^ in(1)
	case BACXnor:
		v = ^(in(0) ^ in(1))
	case BACMux:
		if in(0)&1 == 0 {
			v = in(1)
		} else {
			v = in(2)
		}
	case BACMajority:
		v = (in(0) & in(1)) | (in(0) & in(2)) | (in(1) & in(2))
	case BACReduceAnd:
		if in(0)&bacMask(objects[o.Inputs[0]].Width) == bacMask(objects[o.Inputs[0]].Width) {
			v = 1
		}
	case BACReduceNand:
		if in(0)&bacMask(objects[o.Inputs[0]].Width) != bacMask(objects[o.Inputs[0]].Width) {
			v = 1
		}
	case BACReduceOr:
		if in(0) != 0 {
			v = 1
		}
	case BACReduceNor:
		if in(0) == 0 {
			v = 1
		}
	case BACReduceXor:
		v = uint64(bits.OnesCount64(in(0)) & 1)
	case BACReduceXnor:
		v = uint64(1 - bits.OnesCount64(in(0))&1)
	case BACLogicAnd:
		if in(0) != 0 && in(1) != 0 {
			v = 1
		}
	case BACLogicNand:
		if in(0) == 0 || in(1) == 0 {
			v = 1
		}
	case BACLogicOr:
		if in(0) != 0 || in(1) != 0 {
			v = 1
		}
	case BACLogicNor:
		if in(0) == 0 && in(1) == 0 {
			v = 1
		}
	case BACLogicXor:
		if (in(0) != 0) != (in(1) != 0) {
			v = 1
		}
	case BACLogicXnor:
		if (in(0) != 0) == (in(1) != 0) {
			v = 1
		}
	case BACAdd:
		v = in(0) + in(1)
	case BACSubtract:
		v = in(0) - in(1)
	case BACMultiply:
		v = in(0) * in(1)
	case BACDivide:
		if in(1) == 0 {
			return 0, ErrDimension
		}
		v = in(0) / in(1)
	case BACModulo, BACRemainder:
		if in(1) == 0 {
			return 0, ErrDimension
		}
		v = in(0) % in(1)
	case BACPower:
		v = 1
		base, exp := in(0), in(1)
		for exp != 0 {
			if err := budget.take(1, 0); err != nil {
				return 0, err
			}
			if exp&1 != 0 {
				v *= base
			}
			base *= base
			exp >>= 1
		}
	case BACMinimum:
		if in(0) < in(1) {
			v = in(0)
		} else {
			v = in(1)
		}
	case BACAbsolute:
		x := in(0) & m
		sign := uint64(1) << (o.Width - 1)
		if x&sign != 0 {
			v = (^x + 1) & m
		} else {
			v = x
		}
	case BACLess:
		if in(0) < in(1) {
			v = 1
		}
	case BACLessEqual:
		if in(0) <= in(1) {
			v = 1
		}
	case BACGreaterEqual:
		if in(0) >= in(1) {
			v = 1
		}
	case BACGreater:
		if in(0) > in(1) {
			v = 1
		}
	case BACEqual:
		if in(0) == in(1) {
			v = 1
		}
	case BACNotEqual:
		if in(0) != in(1) {
			v = 1
		}
	case BACShiftLeft:
		if in(1) < 64 {
			v = in(0) << in(1)
		}
	case BACShiftRight:
		if in(1) < 64 {
			v = in(0) >> in(1)
		}
	case BACRotateLeft:
		if o.Width == 64 {
			v = bits.RotateLeft64(in(0), int(in(1)&63))
		} else {
			s := int(in(1) % uint64(o.Width))
			v = (in(0) << s) | (in(0) >> ((o.Width - s) % o.Width))
		}
	case BACRotateRight:
		if o.Width == 64 {
			v = bits.RotateLeft64(in(0), -int(in(1)&63))
		} else {
			s := int(in(1) % uint64(o.Width))
			v = (in(0) >> s) | (in(0) << ((o.Width - s) % o.Width))
		}
	case BACUser:
		return 0, ErrInvariant
	default:
		return 0, ErrInvariant
	}
	return v & m, nil
}

func (d *BACDesign) Evaluate(module string, inputs map[string]uint64) (map[string]uint64, error) {
	if err := d.check(); err != nil {
		return nil, err
	}
	i, ok := d.moduleByName[module]
	if !ok {
		return nil, ErrDimension
	}
	n := d.modules[i]
	if err := validateBACNetwork(n, d.limits); err != nil {
		return nil, err
	}
	b := bacBudget{limits: d.limits}
	values := make([]uint64, len(n.objects))
	if err := b.take(0, len(values)*8); err != nil {
		return nil, err
	}
	for _, o := range n.objects {
		if err := b.take(1+len(o.Inputs), 0); err != nil {
			return nil, err
		}
		if o.Type == BACInput {
			v, found := inputs[o.Name]
			if !found {
				return nil, ErrDimension
			}
			values[o.ID] = v & bacMask(o.Width)
			continue
		}
		v, err := evalBACObject(o, n.objects, values, &b)
		if err != nil {
			return nil, err
		}
		values[o.ID] = v
	}
	r := make(map[string]uint64, len(n.outputs))
	for _, p := range n.outputs {
		if err := b.take(1, len(p.Name)+8); err != nil {
			return nil, err
		}
		r[p.Name] = values[p.Driver] & bacMask(p.Width)
	}
	return r, nil
}

// Validate verifies every module and all user-module references without
// mutating the design.
func (d *BACDesign) Validate() error {
	if d == nil {
		return ErrBACClosed
	}
	b := bacBudget{limits: d.limits}
	return d.validateWithBudget(&b)
}

func (d *BACDesign) validateWithBudget(b *bacBudget) error {
	if err := d.check(); err != nil {
		return err
	}
	if len(d.modules) == 0 || d.Root == "" {
		return ErrInvariant
	}
	if d.Name == "" || len(d.Name) > d.limits.MaxTextBytes || len(d.Spec) > d.limits.MaxTextBytes || len(d.Root) > d.limits.MaxTextBytes {
		return ErrLimit
	}
	if len(d.moduleByName) != len(d.modules) || len(d.modules) > d.limits.MaxModules {
		return ErrInvariant
	}
	if _, ok := d.moduleByName[d.Root]; !ok {
		return ErrInvariant
	}
	objects, edges, names := 0, 0, len(d.Name)
	if err := bacCheckedAdd(&names, len(d.Spec), len(d.Root)); err != nil {
		return err
	}
	for moduleID, n := range d.modules {
		if n == nil || d.moduleByName[n.Name] != moduleID {
			return ErrInvariant
		}
		work := len(n.outputs)
		edges += len(n.outputs)
		if err := bacCheckedAdd(&objects, len(n.objects)); err != nil || objects > d.limits.MaxObjects {
			return ErrLimit
		}
		if err := bacCheckedAdd(&names, len(n.Name)); err != nil {
			return err
		}
		for _, o := range n.objects {
			work += 1 + len(o.Inputs)
			if err := bacCheckedAdd(&edges, len(o.Inputs)); err != nil || edges > d.limits.MaxEdges {
				return ErrLimit
			}
			if err := bacCheckedAdd(&names, len(o.Name)); err != nil {
				return err
			}
		}
		for _, output := range n.outputs {
			if err := bacCheckedAdd(&names, len(output.Name)); err != nil {
				return err
			}
		}
		if names > d.limits.MaxTextBytes {
			return ErrLimit
		}
		if err := b.take(work, 0); err != nil {
			return err
		}
		if err := validateBACNetwork(n, d.limits); err != nil {
			return err
		}
		for _, o := range n.objects {
			if o.Type == BACUser {
				targetID, ok := d.moduleByName[o.Module]
				if !ok {
					return ErrInvariant
				}
				target := d.modules[targetID]
				if len(target.outputs) != 1 || len(o.Inputs) != len(target.inputs) || o.Width != target.outputs[0].Width {
					return ErrDimension
				}
				for j, id := range o.Inputs {
					if n.objects[id].Width != target.objects[target.inputs[j]].Width {
						return ErrDimension
					}
				}
			}
		}
	}
	if objects != d.objects || edges != d.edges || names != d.nameBytes+len(d.Spec)+len(d.Root) {
		return ErrInvariant
	}
	if _, err := d.bacBinarySize(); err != nil {
		return err
	}
	if err := b.take(0, len(d.modules)); err != nil {
		return err
	}
	state := make([]uint8, len(d.modules))
	var visit func(int, int) error
	visit = func(id, depth int) error {
		if depth > d.limits.MaxDepth {
			return ErrLimit
		}
		if state[id] == 1 {
			return ErrInvariant
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, o := range d.modules[id].objects {
			if o.Type == BACUser {
				if err := b.take(1, 0); err != nil {
					return err
				}
				if err := visit(d.moduleByName[o.Module], depth+1); err != nil {
					return err
				}
			}
		}
		state[id] = 2
		return nil
	}
	for i := range d.modules {
		if err := visit(i, 1); err != nil {
			return err
		}
	}
	return nil
}

// Clone replaces the pointer-vector conversion layer with explicit ownership.
func (d *BACDesign) Clone() (*BACDesign, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	r, err := NewBACDesign(d.Name, d.limits)
	if err != nil {
		return nil, err
	}
	r.Spec, r.Root = d.Spec, d.Root
	for _, n := range d.modules {
		if err := r.addModuleOwned(n); err != nil {
			return nil, err
		}
	}
	return r, nil
}

const bacBinaryMagic = "BACG1\x00"

func appendBACU32(dst []byte, v int) ([]byte, error) {
	if v < 0 || uint64(v) > math.MaxUint32 {
		return nil, ErrLimit
	}
	return binary.LittleEndian.AppendUint32(dst, uint32(v)), nil
}
func appendBACString(dst []byte, s string) ([]byte, error) {
	var err error
	dst, err = appendBACU32(dst, len(s))
	if err != nil {
		return nil, err
	}
	return append(dst, s...), nil
}

func (d *BACDesign) bacBinarySize() (int, error) {
	total := len(bacBinaryMagic)
	addString := func(s string) error { return bacCheckedAdd(&total, 4, len(s)) }
	for _, s := range []string{d.Name, d.Spec, d.Root} {
		if err := addString(s); err != nil {
			return 0, err
		}
	}
	if err := bacCheckedAdd(&total, 4); err != nil {
		return 0, err
	}
	for _, n := range d.modules {
		if err := addString(n.Name); err != nil {
			return 0, err
		}
		if err := bacCheckedAdd(&total, 4); err != nil {
			return 0, err
		}
		for _, o := range n.objects {
			if err := bacCheckedAdd(&total, 2); err != nil {
				return 0, err
			}
			if err := addString(o.Name); err != nil {
				return 0, err
			}
			if err := bacCheckedAdd(&total, 8); err != nil {
				return 0, err
			}
			if err := addString(o.Module); err != nil {
				return 0, err
			}
			if len(o.Inputs) > (int(^uint(0)>>1)-total-4)/4 {
				return 0, ErrLimit
			}
			if err := bacCheckedAdd(&total, 4, 4*len(o.Inputs)); err != nil {
				return 0, err
			}
		}
		if err := bacCheckedAdd(&total, 4); err != nil {
			return 0, err
		}
		for _, output := range n.outputs {
			if err := addString(output.Name); err != nil {
				return 0, err
			}
			if err := bacCheckedAdd(&total, 4); err != nil {
				return 0, err
			}
		}
	}
	if total > d.limits.MaxBytes {
		return 0, ErrLimit
	}
	return total, nil
}

// MarshalBAC is a deterministic, portable counterpart of Bac_ManWriteBacInt.
func (d *BACDesign) MarshalBAC() ([]byte, error) {
	if d == nil {
		return nil, ErrBACClosed
	}
	budget := bacBudget{limits: d.limits}
	if err := d.validateWithBudget(&budget); err != nil {
		return nil, err
	}
	size, err := d.bacBinarySize()
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, size)
	out = append(out, bacBinaryMagic...)
	for _, s := range []string{d.Name, d.Spec, d.Root} {
		out, err = appendBACString(out, s)
		if err != nil {
			return nil, err
		}
	}
	out, err = appendBACU32(out, len(d.modules))
	if err != nil {
		return nil, err
	}
	for _, n := range d.modules {
		if err := budget.take(1, 0); err != nil {
			return nil, err
		}
		out, err = appendBACString(out, n.Name)
		if err != nil {
			return nil, err
		}
		out, err = appendBACU32(out, len(n.objects))
		if err != nil {
			return nil, err
		}
		for _, o := range n.objects {
			if err := budget.take(1+len(o.Inputs), 0); err != nil {
				return nil, err
			}
			out = append(out, byte(o.Type), byte(o.Width))
			out, err = appendBACString(out, o.Name)
			if err != nil {
				return nil, err
			}
			out = binary.LittleEndian.AppendUint64(out, o.Value)
			out, err = appendBACString(out, o.Module)
			if err != nil {
				return nil, err
			}
			out, err = appendBACU32(out, len(o.Inputs))
			if err != nil {
				return nil, err
			}
			for _, id := range o.Inputs {
				out, err = appendBACU32(out, id)
				if err != nil {
					return nil, err
				}
			}
		}
		out, err = appendBACU32(out, len(n.outputs))
		if err != nil {
			return nil, err
		}
		for _, p := range n.outputs {
			if err := budget.take(1, 0); err != nil {
				return nil, err
			}
			out, err = appendBACString(out, p.Name)
			if err != nil {
				return nil, err
			}
			out, err = appendBACU32(out, p.Driver)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(out) != size {
		return nil, ErrInvariant
	}
	return append([]byte(nil), out...), nil
}

type bacDecoder struct {
	data   []byte
	pos    int
	budget bacBudget
}

func (r *bacDecoder) take(n int) ([]byte, error) {
	if n < 0 || n > len(r.data)-r.pos {
		return nil, ErrDimension
	}
	if err := r.budget.take(1, n); err != nil {
		return nil, err
	}
	p := r.data[r.pos : r.pos+n]
	r.pos += n
	return p, nil
}
func (r *bacDecoder) u32() (int, error) {
	p, e := r.take(4)
	if e != nil {
		return 0, e
	}
	return int(binary.LittleEndian.Uint32(p)), nil
}
func (r *bacDecoder) str() (string, error) {
	n, e := r.u32()
	if e != nil {
		return "", e
	}
	p, e := r.take(n)
	if e != nil {
		return "", e
	}
	return string(p), nil
}

// UnmarshalBAC parses owned bytes atomically and rejects noncanonical graphs.
func UnmarshalBAC(data []byte, limits BACLimits) (*BACDesign, error) {
	limits = limits.normalized()
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	r := bacDecoder{data: data, budget: bacBudget{limits: limits}}
	p, e := r.take(len(bacBinaryMagic))
	if e != nil || string(p) != bacBinaryMagic {
		return nil, ErrDimension
	}
	name, e := r.str()
	if e != nil {
		return nil, e
	}
	spec, e := r.str()
	if e != nil {
		return nil, e
	}
	root, e := r.str()
	if e != nil {
		return nil, e
	}
	d, e := NewBACDesign(name, limits)
	if e != nil {
		return nil, e
	}
	d.Spec = spec
	nm, e := r.u32()
	if e != nil {
		return nil, e
	}
	if nm < 1 || nm > limits.MaxModules {
		return nil, ErrLimit
	}
	for mi := 0; mi < nm; mi++ {
		mn, e := r.str()
		if e != nil {
			return nil, e
		}
		b, e := d.AddNetwork(mn)
		if e != nil {
			return nil, e
		}
		no, e := r.u32()
		if e != nil {
			return nil, e
		}
		if no > limits.MaxObjects-d.objects {
			return nil, ErrLimit
		}
		for oi := 0; oi < no; oi++ {
			h, e := r.take(2)
			if e != nil {
				return nil, e
			}
			kind, width := BACType(h[0]), int(h[1])
			on, e := r.str()
			if e != nil {
				return nil, e
			}
			q, e := r.take(8)
			if e != nil {
				return nil, e
			}
			value := binary.LittleEndian.Uint64(q)
			mod, e := r.str()
			if e != nil {
				return nil, e
			}
			ni, e := r.u32()
			if e != nil {
				return nil, e
			}
			if ni > limits.MaxEdges {
				return nil, ErrLimit
			}
			ins := make([]int, ni)
			for j := range ins {
				ins[j], e = r.u32()
				if e != nil {
					return nil, e
				}
			}
			id, e := b.addObject(kind, on, width, ins, value, mod)
			if e != nil || id != oi {
				if e == nil {
					e = ErrInvariant
				}
				return nil, e
			}
		}
		np, e := r.u32()
		if e != nil {
			return nil, e
		}
		for j := 0; j < np; j++ {
			pn, e := r.str()
			if e != nil {
				return nil, e
			}
			driver, e := r.u32()
			if e != nil {
				return nil, e
			}
			if e = b.AddOutput(pn, driver); e != nil {
				return nil, e
			}
		}
		if e = b.Commit(); e != nil {
			return nil, e
		}
	}
	if r.pos != len(data) {
		return nil, ErrDimension
	}
	d.Root = root
	if e = d.Validate(); e != nil {
		return nil, e
	}
	return d, nil
}

func bacIdentifier(s string) bool {
	if s == "" || !((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z') || s[0] == '_') {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

func bacWidthDecl(width int) string {
	if width == 1 {
		return ""
	}
	return fmt.Sprintf("[%d:0] ", width-1)
}

func bacAppendText(dst *strings.Builder, limit int, text string) error {
	if limit < 0 || len(text) > limit-dst.Len() {
		return ErrLimit
	}
	dst.WriteString(text)
	return nil
}

func bacVerilogExpr(o BACObject, n *BACNetwork) (string, error) {
	name := func(i int) string { return n.objects[o.Inputs[i]].Name }
	bin := map[BACType]string{BACAnd: "&", BACOr: "|", BACXor: "^", BACAdd: "+", BACSubtract: "-", BACMultiply: "*", BACDivide: "/", BACModulo: "%", BACLess: "<", BACLessEqual: "<=", BACGreaterEqual: ">=", BACGreater: ">", BACEqual: "==", BACNotEqual: "!=", BACShiftLeft: "<<", BACShiftRight: ">>"}
	switch o.Type {
	case BACConst0:
		return fmt.Sprintf("%d'd%d", o.Width, o.Value&bacMask(o.Width)), nil
	case BACConst1:
		return fmt.Sprintf("%d'd%d", o.Width, bacMask(o.Width)), nil
	case BACBuffer:
		return name(0), nil
	case BACInvert:
		return "~" + name(0), nil
	case BACNand:
		return fmt.Sprintf("~(%s & %s)", name(0), name(1)), nil
	case BACNor:
		return fmt.Sprintf("~(%s | %s)", name(0), name(1)), nil
	case BACXnor:
		return fmt.Sprintf("~(%s ^ %s)", name(0), name(1)), nil
	case BACMux:
		return fmt.Sprintf("(%s ? %s : %s)", name(0), name(2), name(1)), nil
	}
	if op, ok := bin[o.Type]; ok {
		return fmt.Sprintf("(%s %s %s)", name(0), op, name(1)), nil
	}
	return "", ErrDimension
}

// Verilog emits the deterministic synthesizable subset consumed by ParseBACVerilog.
func (d *BACDesign) Verilog(module string) (string, error) {
	if d == nil {
		return "", ErrBACClosed
	}
	budget := bacBudget{limits: d.limits}
	if e := d.validateWithBudget(&budget); e != nil {
		return "", e
	}
	i, ok := d.moduleByName[module]
	if !ok {
		return "", ErrDimension
	}
	n := d.modules[i]
	if !bacIdentifier(n.Name) {
		return "", ErrDimension
	}
	for _, id := range n.inputs {
		if !bacIdentifier(n.objects[id].Name) {
			return "", ErrDimension
		}
	}
	for _, output := range n.outputs {
		if !bacIdentifier(output.Name) {
			return "", ErrDimension
		}
		if _, exists := n.nameToObject[output.Name]; exists {
			return "", ErrInvariant
		}
	}
	var s strings.Builder
	ports := make([]string, 0, len(n.inputs)+len(n.outputs))
	for _, id := range n.inputs {
		if e := budget.take(1, 0); e != nil {
			return "", e
		}
		ports = append(ports, n.objects[id].Name)
	}
	for _, p := range n.outputs {
		if e := budget.take(1, 0); e != nil {
			return "", e
		}
		ports = append(ports, p.Name)
	}
	if e := bacAppendText(&s, d.limits.MaxTextBytes, fmt.Sprintf("module %s(%s);\n", n.Name, strings.Join(ports, ", "))); e != nil {
		return "", e
	}
	for _, id := range n.inputs {
		o := n.objects[id]
		if e := bacAppendText(&s, d.limits.MaxTextBytes, fmt.Sprintf("  input %s%s;\n", bacWidthDecl(o.Width), o.Name)); e != nil {
			return "", e
		}
	}
	for _, p := range n.outputs {
		if e := bacAppendText(&s, d.limits.MaxTextBytes, fmt.Sprintf("  output %s%s;\n", bacWidthDecl(p.Width), p.Name)); e != nil {
			return "", e
		}
	}
	for _, o := range n.objects {
		if e := budget.take(1+len(o.Inputs), 0); e != nil {
			return "", e
		}
		if o.Type == BACInput {
			continue
		}
		if !bacIdentifier(o.Name) {
			return "", ErrDimension
		}
		if e := bacAppendText(&s, d.limits.MaxTextBytes, fmt.Sprintf("  wire %s%s;\n", bacWidthDecl(o.Width), o.Name)); e != nil {
			return "", e
		}
		ex, e := bacVerilogExpr(o, n)
		if e != nil {
			return "", e
		}
		if e := bacAppendText(&s, d.limits.MaxTextBytes, fmt.Sprintf("  assign %s = %s;\n", o.Name, ex)); e != nil {
			return "", e
		}
	}
	for _, p := range n.outputs {
		if e := budget.take(1, 0); e != nil {
			return "", e
		}
		if e := bacAppendText(&s, d.limits.MaxTextBytes, fmt.Sprintf("  assign %s = %s;\n", p.Name, n.objects[p.Driver].Name)); e != nil {
			return "", e
		}
	}
	if e := bacAppendText(&s, d.limits.MaxTextBytes, "endmodule\n"); e != nil {
		return "", e
	}
	return s.String(), nil
}

func parseBACDecl(line, keyword string) (int, string, error) {
	x := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, keyword)), ";")
	width := 1
	if strings.HasPrefix(x, "[") {
		j := strings.IndexByte(x, ']')
		if j < 0 {
			return 0, "", ErrDimension
		}
		r := strings.Split(x[1:j], ":")
		if len(r) != 2 {
			return 0, "", ErrDimension
		}
		hi, e := strconv.Atoi(r[0])
		if e != nil || r[1] != "0" || hi < 0 || hi >= 64 {
			return 0, "", ErrDimension
		}
		width = hi + 1
		x = strings.TrimSpace(x[j+1:])
	}
	if !bacIdentifier(x) {
		return 0, "", ErrDimension
	}
	return width, x, nil
}

func stripOuter(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		return strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}

// ParseBACVerilog accepts the bounded structural subset emitted by Verilog.
func ParseBACVerilog(text string, limits BACLimits) (*BACDesign, error) {
	limits = limits.normalized()
	if len(text) > limits.MaxTextBytes || len(text) > limits.MaxBytes {
		return nil, ErrLimit
	}
	lines := strings.Split(text, "\n")
	var name string
	var b *BACBuilder
	var d *BACDesign
	ids := make(map[string]int)
	outWidths := make(map[string]int)
	wireWidths := make(map[string]int)
	declared := make(map[string]bool)
	assigned := make(map[string]bool)
	headerPorts := []string{}
	inputOrder := []string{}
	outOrder := []string{}
	sawModule, sawEnd := false, false
	budget := bacBudget{limits: limits}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if e := budget.take(1, len(line)); e != nil {
			return nil, e
		}
		if sawEnd {
			return nil, ErrDimension
		}
		switch {
		case strings.HasPrefix(line, "module "):
			if sawModule || !strings.HasSuffix(line, ");") {
				return nil, ErrDimension
			}
			h := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			j := strings.IndexByte(h, '(')
			if j < 1 || !strings.HasSuffix(h, ");") {
				return nil, ErrDimension
			}
			name = strings.TrimSpace(h[:j])
			if !bacIdentifier(name) {
				return nil, ErrDimension
			}
			var e error
			d, e = NewBACDesign(name, limits)
			if e != nil {
				return nil, e
			}
			b, e = d.AddNetwork(name)
			if e != nil {
				return nil, e
			}
			portText := strings.TrimSpace(h[j+1 : len(h)-2])
			if portText != "" {
				seen := make(map[string]bool)
				for _, port := range strings.Split(portText, ",") {
					port = strings.TrimSpace(port)
					if !bacIdentifier(port) || seen[port] {
						return nil, ErrDimension
					}
					seen[port] = true
					headerPorts = append(headerPorts, port)
				}
			}
			sawModule = true
		case strings.HasPrefix(line, "input "):
			if b == nil {
				return nil, ErrDimension
			}
			w, n, e := parseBACDecl(line, "input ")
			if e != nil {
				return nil, e
			}
			if declared[n] {
				return nil, ErrInvariant
			}
			id, e := b.AddInput(n, w)
			if e != nil {
				return nil, e
			}
			declared[n] = true
			ids[n] = id
			inputOrder = append(inputOrder, n)
		case strings.HasPrefix(line, "output "):
			if b == nil {
				return nil, ErrDimension
			}
			w, n, e := parseBACDecl(line, "output ")
			if e != nil {
				return nil, e
			}
			if declared[n] {
				return nil, ErrInvariant
			}
			declared[n] = true
			outWidths[n] = w
			outOrder = append(outOrder, n)
		case strings.HasPrefix(line, "wire "):
			if b == nil {
				return nil, ErrDimension
			}
			w, n, e := parseBACDecl(line, "wire ")
			if e != nil {
				return nil, e
			}
			if declared[n] {
				return nil, ErrInvariant
			}
			declared[n] = true
			wireWidths[n] = w
		case strings.HasPrefix(line, "assign "):
			if b == nil {
				return nil, ErrDimension
			}
			x := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "assign ")), ";")
			parts := strings.SplitN(x, "=", 2)
			if len(parts) != 2 {
				return nil, ErrDimension
			}
			lhs, expr := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if !bacIdentifier(lhs) || assigned[lhs] {
				return nil, ErrDimension
			}
			declaredWidth, isOutput := outWidths[lhs]
			if !isOutput {
				var exists bool
				declaredWidth, exists = wireWidths[lhs]
				if !exists {
					return nil, ErrInvariant
				}
			}
			driver := -1
			if id, ok := ids[expr]; ok {
				driver = id
				if b.network.objects[id].Width != declaredWidth {
					return nil, ErrDimension
				}
			} else {
				target := lhs
				if isOutput {
					target = "$drv_" + lhs
				}
				expr = stripOuter(expr)
				kind := BACNone
				ins := []int{}
				width := declaredWidth
				if strings.HasPrefix(expr, "~(") && strings.HasSuffix(expr, ")") {
					q := strings.Split(strings.TrimSpace(expr[2:len(expr)-1]), " ")
					if len(q) != 3 {
						return nil, ErrDimension
					}
					ops := map[string]BACType{"&": BACNand, "|": BACNor, "^": BACXnor}
					var ok bool
					kind, ok = ops[q[1]]
					if !ok {
						return nil, ErrDimension
					}
					for _, v := range []string{q[0], q[2]} {
						id, yes := ids[v]
						if !yes {
							return nil, ErrDimension
						}
						ins = append(ins, id)
					}
					if width == 0 {
						width = b.network.objects[ins[0]].Width
					}
				} else if strings.HasPrefix(expr, "~") {
					kind = BACInvert
					id, ok := ids[strings.TrimSpace(expr[1:])]
					if !ok {
						return nil, ErrDimension
					}
					ins = []int{id}
					if width == 0 {
						width = b.network.objects[id].Width
					}
				} else if q := strings.Split(expr, " "); len(q) == 5 && q[1] == "?" && q[3] == ":" {
					kind = BACMux
					for _, k := range []int{0, 4, 2} {
						id, ok := ids[q[k]]
						if !ok {
							return nil, ErrDimension
						}
						ins = append(ins, id)
					}
					if width == 0 {
						width = b.network.objects[ins[1]].Width
					}
				} else if q := strings.Split(expr, " "); len(q) == 3 {
					ops := map[string]BACType{"&": BACAnd, "~&": BACNand, "|": BACOr, "~|": BACNor, "^": BACXor, "~^": BACXnor, "+": BACAdd, "-": BACSubtract, "*": BACMultiply, "/": BACDivide, "%": BACModulo, "<": BACLess, "<=": BACLessEqual, ">=": BACGreaterEqual, ">": BACGreater, "==": BACEqual, "!=": BACNotEqual, "<<": BACShiftLeft, ">>": BACShiftRight}
					var ok bool
					kind, ok = ops[q[1]]
					if !ok {
						return nil, ErrDimension
					}
					for _, v := range []string{q[0], q[2]} {
						id, yes := ids[v]
						if !yes {
							return nil, ErrDimension
						}
						ins = append(ins, id)
					}
					if width == 0 {
						width = b.network.objects[ins[0]].Width
					}
					if kind >= BACLess && kind <= BACNotEqual {
						width = 1
					}
				} else if k := strings.Index(expr, "'d"); k > 0 {
					w, e := strconv.Atoi(expr[:k])
					v, e2 := strconv.ParseUint(expr[k+2:], 10, 64)
					if e != nil || e2 != nil || w < 1 || w > 64 {
						return nil, ErrDimension
					}
					kind = BACConst0
					width = w
					id, e := b.AddConstant(target, w, v)
					if e != nil {
						return nil, e
					}
					driver = id
					ids[lhs] = id
				} else {
					return nil, ErrDimension
				}
				if driver < 0 {
					id, e := b.AddNode(kind, target, width, ins...)
					if e != nil {
						return nil, e
					}
					driver = id
					ids[lhs] = id
				}
			}
			if _, isOut := outWidths[lhs]; !isOut {
				ids[lhs] = driver
			} else {
				ids[lhs] = driver
			}
			assigned[lhs] = true
		case line == "endmodule":
			if !sawModule {
				return nil, ErrDimension
			}
			sawEnd = true
		case strings.HasPrefix(line, "//"):
		default:
			return nil, ErrDimension
		}
	}
	if d == nil || b == nil || !sawModule || !sawEnd {
		return nil, ErrDimension
	}
	declaredPorts := append(append([]string(nil), inputOrder...), outOrder...)
	if len(headerPorts) != len(declaredPorts) {
		return nil, ErrDimension
	}
	for i := range headerPorts {
		if headerPorts[i] != declaredPorts[i] {
			return nil, ErrDimension
		}
	}
	for _, n := range outOrder {
		id, ok := ids[n]
		if !ok {
			return nil, ErrInvariant
		}
		if b.network.objects[id].Width != outWidths[n] {
			return nil, ErrDimension
		}
		if e := b.AddOutput(n, id); e != nil {
			return nil, e
		}
	}
	if e := b.Commit(); e != nil {
		return nil, e
	}
	d.Root = name
	return d, nil
}

func bacBLIFCubes(kind BACType) ([]string, bool) {
	switch kind {
	case BACConst0:
		return nil, true
	case BACConst1:
		return []string{"1"}, true
	case BACBuffer:
		return []string{"1 1"}, true
	case BACInvert:
		return []string{"0 1"}, true
	case BACAnd:
		return []string{"11 1"}, true
	case BACNand:
		return []string{"0- 1", "-0 1"}, true
	case BACOr:
		return []string{"1- 1", "-1 1"}, true
	case BACNor:
		return []string{"00 1"}, true
	case BACXor:
		return []string{"10 1", "01 1"}, true
	case BACXnor:
		return []string{"00 1", "11 1"}, true
	case BACMux:
		return []string{"01- 1", "1-1 1"}, true
	}
	return nil, false
}

// BLIF emits a deterministic scalar subset and rejects word-level modules.
func (d *BACDesign) BLIF(module string) (string, error) {
	if d == nil {
		return "", ErrBACClosed
	}
	budget := bacBudget{limits: d.limits}
	if e := d.validateWithBudget(&budget); e != nil {
		return "", e
	}
	i, ok := d.moduleByName[module]
	if !ok {
		return "", ErrDimension
	}
	n := d.modules[i]
	if !bacIdentifier(n.Name) {
		return "", ErrDimension
	}
	for _, o := range n.objects {
		if e := budget.take(1+len(o.Inputs), 0); e != nil {
			return "", e
		}
		if o.Width != 1 {
			return "", ErrDimension
		}
		if !bacIdentifier(o.Name) {
			return "", ErrDimension
		}
	}
	for _, output := range n.outputs {
		if !bacIdentifier(output.Name) {
			return "", ErrDimension
		}
		if _, exists := n.nameToObject[output.Name]; exists {
			return "", ErrInvariant
		}
	}
	var s strings.Builder
	if e := bacAppendText(&s, d.limits.MaxTextBytes, fmt.Sprintf(".model %s\n", n.Name)); e != nil {
		return "", e
	}
	if e := bacAppendText(&s, d.limits.MaxTextBytes, ".inputs"); e != nil {
		return "", e
	}
	for _, id := range n.inputs {
		if e := bacAppendText(&s, d.limits.MaxTextBytes, " "+n.objects[id].Name); e != nil {
			return "", e
		}
	}
	if e := bacAppendText(&s, d.limits.MaxTextBytes, "\n.outputs"); e != nil {
		return "", e
	}
	for _, p := range n.outputs {
		if e := bacAppendText(&s, d.limits.MaxTextBytes, " "+p.Name); e != nil {
			return "", e
		}
	}
	if e := bacAppendText(&s, d.limits.MaxTextBytes, "\n"); e != nil {
		return "", e
	}
	for _, o := range n.objects {
		if e := budget.take(1+len(o.Inputs), 0); e != nil {
			return "", e
		}
		if o.Type == BACInput {
			continue
		}
		cubes, ok := bacBLIFCubes(o.Type)
		if !ok {
			return "", ErrDimension
		}
		if e := bacAppendText(&s, d.limits.MaxTextBytes, ".names"); e != nil {
			return "", e
		}
		for _, id := range o.Inputs {
			if e := bacAppendText(&s, d.limits.MaxTextBytes, " "+n.objects[id].Name); e != nil {
				return "", e
			}
		}
		if e := bacAppendText(&s, d.limits.MaxTextBytes, " "+o.Name+"\n"); e != nil {
			return "", e
		}
		for _, c := range cubes {
			if e := bacAppendText(&s, d.limits.MaxTextBytes, c+"\n"); e != nil {
				return "", e
			}
		}
	}
	for _, p := range n.outputs {
		if e := budget.take(1, 0); e != nil {
			return "", e
		}
		if e := bacAppendText(&s, d.limits.MaxTextBytes, fmt.Sprintf(".names %s %s\n1 1\n", n.objects[p.Driver].Name, p.Name)); e != nil {
			return "", e
		}
	}
	if e := bacAppendText(&s, d.limits.MaxTextBytes, ".end\n"); e != nil {
		return "", e
	}
	return s.String(), nil
}

func matchBACBLIF(inputs int, cubes []string) (BACType, bool) {
	key := strings.Join(cubes, "\n")
	switch {
	case inputs == 0 && key == "":
		return BACConst0, true
	case inputs == 0 && key == "1":
		return BACConst1, true
	case inputs == 1 && key == "1 1":
		return BACBuffer, true
	case inputs == 1 && key == "0 1":
		return BACInvert, true
	case inputs == 2 && key == "11 1":
		return BACAnd, true
	case inputs == 2 && key == "0- 1\n-0 1":
		return BACNand, true
	case inputs == 2 && key == "1- 1\n-1 1":
		return BACOr, true
	case inputs == 2 && key == "00 1":
		return BACNor, true
	case inputs == 2 && key == "10 1\n01 1":
		return BACXor, true
	case inputs == 2 && key == "00 1\n11 1":
		return BACXnor, true
	case inputs == 3 && key == "01- 1\n1-1 1":
		return BACMux, true
	}
	return BACNone, false
}

// ParseBACBLIF accepts the exact scalar structural subset emitted by BLIF.
func ParseBACBLIF(text string, limits BACLimits) (*BACDesign, error) {
	limits = limits.normalized()
	if len(text) > limits.MaxTextBytes || len(text) > limits.MaxBytes {
		return nil, ErrLimit
	}
	lines := strings.Split(text, "\n")
	var model string
	inputs, outputs := []string{}, []string{}
	type table struct{ names, cubes []string }
	tabs := []table{}
	sawModel, sawEnd := false, false
	budget := bacBudget{limits: limits}
	for i := 0; i < len(lines); {
		line := strings.TrimSpace(lines[i])
		i++
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if e := budget.take(1, len(line)); e != nil {
			return nil, e
		}
		if sawEnd {
			return nil, ErrDimension
		}
		f := strings.Fields(line)
		switch f[0] {
		case ".model":
			if len(f) != 2 || sawModel {
				return nil, ErrDimension
			}
			model = f[1]
			sawModel = true
		case ".inputs":
			inputs = append(inputs, f[1:]...)
		case ".outputs":
			outputs = append(outputs, f[1:]...)
		case ".names":
			if len(f) < 2 {
				return nil, ErrDimension
			}
			t := table{names: append([]string(nil), f[1:]...)}
			for i < len(lines) {
				c := strings.TrimSpace(lines[i])
				if c == "" {
					i++
					continue
				}
				if strings.HasPrefix(c, ".") {
					break
				}
				if e := budget.take(1, len(c)); e != nil {
					return nil, e
				}
				t.cubes = append(t.cubes, c)
				i++
			}
			tabs = append(tabs, t)
		case ".end":
			if len(f) != 1 || !sawModel {
				return nil, ErrDimension
			}
			sawEnd = true
		case ".blackbox":
		default:
			return nil, ErrDimension
		}
	}
	if !sawModel || !sawEnd || !bacIdentifier(model) {
		return nil, ErrDimension
	}
	d, e := NewBACDesign(model, limits)
	if e != nil {
		return nil, e
	}
	b, e := d.AddNetwork(model)
	if e != nil {
		return nil, e
	}
	ids := make(map[string]int)
	outSet := make(map[string]bool)
	for _, n := range outputs {
		if !bacIdentifier(n) || outSet[n] {
			return nil, ErrInvariant
		}
		outSet[n] = true
	}
	for _, n := range inputs {
		if !bacIdentifier(n) {
			return nil, ErrDimension
		}
		id, e := b.AddInput(n, 1)
		if e != nil {
			return nil, e
		}
		ids[n] = id
	}
	for ti, t := range tabs {
		out := t.names[len(t.names)-1]
		inNames := t.names[:len(t.names)-1]
		ins := make([]int, len(inNames))
		for j, n := range inNames {
			id, ok := ids[n]
			if !ok {
				return nil, ErrInvariant
			}
			ins[j] = id
		}
		kind, ok := matchBACBLIF(len(ins), t.cubes)
		if !ok {
			return nil, ErrDimension
		}
		if outSet[out] && kind == BACBuffer && len(ins) == 1 {
			ids[out] = ins[0]
			continue
		}
		name := out
		if outSet[out] {
			name = fmt.Sprintf("$po_%d_%s", ti, out)
		}
		id, e := b.AddNode(kind, name, 1, ins...)
		if e != nil {
			return nil, e
		}
		ids[out] = id
	}
	for _, n := range outputs {
		id, ok := ids[n]
		if !ok {
			return nil, ErrInvariant
		}
		if e := b.AddOutput(n, id); e != nil {
			return nil, e
		}
	}
	if e := b.Commit(); e != nil {
		return nil, e
	}
	d.Root = model
	return d, nil
}

// Collapse returns a deep owned copy for primitive-only designs and rejects
// unresolved user modules. This preserves atomicity while hierarchy insertion
// and source scratch-array algorithms remain explicit gaps.
func (d *BACDesign) Collapse() (*BACDesign, error) {
	if e := d.Validate(); e != nil {
		return nil, e
	}
	for _, n := range d.modules {
		for _, o := range n.objects {
			if o.Type == BACUser {
				return nil, ErrDimension
			}
		}
	}
	return d.Clone()
}

// BACSourceInventory records the assigned source reconciliation. Static access,
// iterator, command, pointer-vector, parser, writer, and clipping helpers are
// absorbed into the concrete owned operations above or remain partial as noted.
var BACSourceInventory = []string{
	"bac.h: partial concrete owned manager/network/object API; native pointer ABI and complete macro surface absent",
	"bacBac.c: deterministic bounded portable binary round trip",
	"bacBlast.c: partial direct word evaluation replaces complete GIA bit blasting",
	"bacCom.c: command registration and process frame integration absent",
	"bacNtk.c: stable objects, fanin topology, names, validation, cloning, and primitive collapse",
	"bacOper.c: broad primitive semantics; source clipping is partial",
	"bacPrs.h: parser ownership, limits, names, ports, and error returns represented",
	"bacPrsBuild.c: partial direct owned construction replaces packed parser handles",
	"bacPrsTrans.c: partial direct translation replaces pointer-vector mutation",
	"bacPtr.c: owned clone and deterministic text replace pointer-vector dumps",
	"bacPtrAbc.c: native ABC frame conversion absent",
	"bacReadBlif.c: bounded scalar structural subset",
	"bacReadVer.c: bounded structural subset emitted by this port",
	"bacWriteBlif.c: deterministic scalar structural subset",
	"bacWriteVer.c: deterministic word-level structural subset",
	"module.make: exact fifteen-unit inventory represented",
}

var BACBuildInventory = []string{
	"bacBlast.c", "bacBac.c", "bacCom.c", "bacLib.c", "bacNtk.c", "bacPrsBuild.c", "bacPrsTrans.c", "bacPtr.c", "bacPtrAbc.c", "bacReadBlif.c", "bacReadSmt.c", "bacReadVer.c", "bacWriteBlif.c", "bacWriteSmt.c", "bacWriteVer.c",
}
