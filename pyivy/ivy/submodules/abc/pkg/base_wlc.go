package abc

// Derived from ABC src/base/wlc/wlc.h, wlcNtk.c, wlcBlast.c, wlcWin.c,
// wlcReadVer.c, wlcWriteVer.c, and module.make. The remaining WLC sources are
// inventoried below and retain explicit integration gaps.

import (
	"fmt"
	"math/bits"
	"slices"
	"sort"
	"strings"
)

// WLCObjectType preserves the source object-type numbering.
type WLCObjectType uint8

const (
	WLCObjectNone WLCObjectType = iota
	WLCObjectPI
	WLCObjectPO
	WLCObjectFO
	WLCObjectFI
	WLCObjectFF
	WLCObjectConst
	WLCObjectBuf
	WLCObjectMux
	WLCObjectShiftRight
	WLCObjectArithmeticShiftRight
	WLCObjectShiftLeft
	WLCObjectArithmeticShiftLeft
	WLCObjectRotateRight
	WLCObjectRotateLeft
	WLCObjectBitNot
	WLCObjectBitAnd
	WLCObjectBitOr
	WLCObjectBitXor
	WLCObjectBitNand
	WLCObjectBitNor
	WLCObjectBitXnor
	WLCObjectBitSelect
	WLCObjectBitConcat
	WLCObjectZeroPad
	WLCObjectSignExtend
	WLCObjectLogicNot
	WLCObjectLogicImplies
	WLCObjectLogicAnd
	WLCObjectLogicOr
	WLCObjectLogicXor
	WLCObjectEqual
	WLCObjectNotEqual
	WLCObjectLess
	WLCObjectGreater
	WLCObjectLessEqual
	WLCObjectGreaterEqual
	WLCObjectReduceAnd
	WLCObjectReduceOr
	WLCObjectReduceXor
	WLCObjectReduceNand
	WLCObjectReduceNor
	WLCObjectReduceXnor
	WLCObjectAdd
	WLCObjectSub
	WLCObjectMultiply
	WLCObjectDivide
	WLCObjectRemainder
	WLCObjectModulus
	WLCObjectPower
	WLCObjectMinus
	WLCObjectSqrt
	WLCObjectSquare
	WLCObjectTable
	WLCObjectRead
	WLCObjectWrite
	WLCObjectAddSub
	WLCObjectSelect
	WLCObjectDecoder
	WLCObjectLUT
	WLCObjectTypeCount
)

var wlcTypeNames = [...]string{
	"none", "pi", "po", "ff", "bi", "ff", "const", "buf", "mux",
	">>", ">>>", "<<", "<<<", "rotR", "rotL", "~", "&", "|", "^",
	"~&", "~|", "~^", "[:]", "{,}", "zPad", "sExt", "!", "=>", "&&",
	"||", "^^", "==", "!=", "<", ">", "<=", ">=", "&", "|", "^",
	"~&", "~|", "~^", "+", "-", "*", "/", "%", "mod", "**", "-",
	"sqrt", "squar", "table", "READ", "WRITE", "addsub", "sel", "dec", "LUT",
}

func (t WLCObjectType) String() string {
	if int(t) >= len(wlcTypeNames) {
		return "out_of_bound"
	}
	return wlcTypeNames[t]
}

// WLCBuildUnits is the exact default-build source list from module.make.
var WLCBuildUnits = [...]string{
	"wlcAbs.c", "wlcAbs2.c", "wlcAbc.c", "wlcPth.c", "wlcBlast.c",
	"wlcCom.c", "wlcGraft.c", "wlcJson.c", "wlcMem.c", "wlcNdr.c",
	"wlcNtk.c", "wlcReadSmt.c", "wlcReadVer.c", "wlcSim.c", "wlcShow.c",
	"wlcStdin.c", "wlcUif.c", "wlcWin.c", "wlcWriteVer.c",
}

// WLCSourceInventory records the reviewed boundaries of the cluster.
var WLCSourceInventory = [...]string{
	"module.make", "wlc.c", "wlc.h", "wlcAbc.c", "wlcAbs.c", "wlcAbs2.c",
	"wlcBlast.c", "wlcCom.c", "wlcGraft.c", "wlcJson.c", "wlcMem.c", "wlcNdr.c",
	"wlcNtk.c", "wlcPth.c", "wlcReadSmt.c", "wlcReadVer.c", "wlcShow.c",
	"wlcSim.c", "wlcStdin.c", "wlcUif.c", "wlcWin.c", "wlcWriteVer.c",
}

// WLCLimits bounds retained state and temporary work.
type WLCLimits struct {
	MaxObjects    int
	MaxEdges      int
	MaxBits       int
	MaxBytes      int
	MaxOperations int
	MaxTextBytes  int
	MaxDepth      int
	MaxFanins     int
	Stop          func() bool
}

func normalizeWLCLimits(l WLCLimits) (WLCLimits, error) {
	vals := []*int{&l.MaxObjects, &l.MaxEdges, &l.MaxBits, &l.MaxBytes, &l.MaxOperations, &l.MaxTextBytes, &l.MaxDepth, &l.MaxFanins}
	defs := []int{1 << 20, 1 << 22, 1 << 24, 1 << 29, 1 << 27, 1 << 28, 1 << 14, 1 << 16}
	for i := range vals {
		if *vals[i] < 0 {
			return l, ErrDimension
		}
		if *vals[i] == 0 {
			*vals[i] = defs[i]
		}
	}
	return l, nil
}

type wlcBudget struct {
	limits WLCLimits
	ops    int
}

func (n *WLCNetwork) checkWork(passes, extra int) error {
	if n == nil || passes < 0 || extra < 0 {
		return ErrInvariant
	}
	work := len(n.objects) - 1
	if work < 0 || n.edges < 0 || n.edges > n.limits.MaxOperations || work > n.limits.MaxOperations-n.edges {
		return ErrLimit
	}
	work += n.edges
	if extra > n.limits.MaxOperations || passes != 0 && work > (n.limits.MaxOperations-extra)/passes {
		return ErrLimit
	}
	return nil
}

func (n *WLCNetwork) checkTempBytes(bytes int) error {
	if n == nil || bytes < 0 || n.bytes > n.limits.MaxBytes || bytes > n.limits.MaxBytes-n.bytes {
		return ErrLimit
	}
	return nil
}

func (b *wlcBudget) use(n int) error {
	if n < 0 || b.ops > b.limits.MaxOperations-n {
		return ErrLimit
	}
	b.ops += n
	if b.limits.Stop != nil && b.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

// WLCRange retains declared bit direction.
type WLCRange struct{ End, Begin int }

func (r WLCRange) Width() (int, error) {
	var d uint64
	if r.End >= r.Begin {
		d = uint64(r.End) - uint64(r.Begin)
	} else {
		d = uint64(r.Begin) - uint64(r.End)
	}
	if d >= uint64(^uint(0)>>1) || d+1 > 64 {
		return 0, ErrLimit
	}
	return int(d + 1), nil
}

// WLCParameters represents the source abstraction defaults without callbacks.
type WLCParameters struct {
	BitsAdd, BitsMul, BitsMux, BitsFlop int
	IterationLimit, SignalLimit         int
	XorOutput, CheckClauses, MFFC       bool
	LoadTrace, Hybrid                   bool
	Verbose, PDRVerbose                 bool
}

func DefaultWLCParameters() WLCParameters {
	max := int(^uint(0) >> 1)
	return WLCParameters{BitsAdd: max, BitsMul: max, BitsMux: max, BitsFlop: max, IterationLimit: 1000, SignalLimit: max, XorOutput: true, CheckClauses: true, MFFC: true, LoadTrace: true, Hybrid: true}
}

// WLCObject is an owned, stable-ID word-level node.
type WLCObject struct {
	ID       int
	Type     WLCObjectType
	Name     string
	Range    WLCRange
	Signed   bool
	Fanins   []int
	Value    uint64
	ParamA   int
	ParamB   int
	Table    uint64
	IsOutput bool
	IsFI     bool
	XValue   bool
}

func (o WLCObject) clone() WLCObject {
	o.Fanins = slices.Clone(o.Fanins)
	return o
}

// WLCFlop binds a flop-output node to its next-state source.
type WLCFlop struct {
	Object  int
	Next    int
	Initial uint64
}

// WLCStep is one caller-owned evaluation result.
type WLCStep struct {
	Values    []uint64
	Outputs   []uint64
	NextState []uint64
}

// WLCStats summarizes an owned network.
type WLCStats struct {
	Objects, Edges, Bits, Inputs, Outputs, Flops, MaxLevel int
	TypeCounts                                             [WLCObjectTypeCount]int
}

// WLCNetwork is an owned, topologically ordered word-level circuit.
type WLCNetwork struct {
	Name    string
	objects []WLCObject
	inputs  []int
	outputs []int
	flops   []WLCFlop
	names   map[string]int
	edges   int
	bits    int
	bytes   int
	limits  WLCLimits
	closed  bool
}

func wlcLegalName(s string) bool {
	if s == "" || !((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z') || s[0] == '_') {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '$') {
			return false
		}
	}
	return true
}

func NewWLCNetwork(name string, limits WLCLimits) (*WLCNetwork, error) {
	limits, err := normalizeWLCLimits(limits)
	if err != nil {
		return nil, err
	}
	if !wlcLegalName(name) || len(name) > limits.MaxBytes {
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &WLCNetwork{Name: name, objects: []WLCObject{{}}, names: make(map[string]int), bytes: len(name), limits: limits}, nil
}

func wlcShape(t WLCObjectType, fanins int) bool {
	switch t {
	case WLCObjectPI, WLCObjectFO, WLCObjectConst:
		return fanins == 0
	case WLCObjectBuf, WLCObjectBitNot, WLCObjectZeroPad, WLCObjectSignExtend, WLCObjectLogicNot,
		WLCObjectReduceAnd, WLCObjectReduceOr, WLCObjectReduceXor, WLCObjectReduceNand,
		WLCObjectReduceNor, WLCObjectReduceXnor, WLCObjectMinus, WLCObjectSqrt,
		WLCObjectSquare, WLCObjectBitSelect, WLCObjectDecoder, WLCObjectLUT:
		return fanins == 1
	case WLCObjectShiftRight, WLCObjectArithmeticShiftRight, WLCObjectShiftLeft,
		WLCObjectArithmeticShiftLeft, WLCObjectRotateRight, WLCObjectRotateLeft,
		WLCObjectBitAnd, WLCObjectBitOr, WLCObjectBitXor, WLCObjectBitNand,
		WLCObjectBitNor, WLCObjectBitXnor, WLCObjectLogicImplies, WLCObjectLogicAnd,
		WLCObjectLogicOr, WLCObjectLogicXor, WLCObjectEqual, WLCObjectNotEqual,
		WLCObjectLess, WLCObjectGreater, WLCObjectLessEqual, WLCObjectGreaterEqual,
		WLCObjectMultiply, WLCObjectDivide, WLCObjectRemainder, WLCObjectModulus,
		WLCObjectPower:
		return fanins == 2
	case WLCObjectAdd, WLCObjectSub:
		return fanins == 2 || fanins == 3
	case WLCObjectMux:
		return fanins >= 3 && fanins <= 65
	case WLCObjectBitConcat:
		return fanins > 0
	case WLCObjectAddSub:
		return fanins == 4
	case WLCObjectSelect:
		return fanins >= 2
	default:
		return false
	}
}

func wlcObjectSemantics(objects []WLCObject, o WLCObject) error {
	w, err := o.Range.Width()
	if err != nil {
		return err
	}
	width := func(id int) (int, error) {
		if id <= 0 || id >= len(objects) {
			return 0, ErrInvariant
		}
		return objects[id].Range.Width()
	}
	switch o.Type {
	case WLCObjectMux:
		sw, err := width(o.Fanins[0])
		if err != nil {
			return err
		}
		if sw >= 63 || sw > 6 || len(o.Fanins) != 1+(1<<sw) {
			return ErrDimension
		}
	case WLCObjectSelect:
		sw, err := width(o.Fanins[0])
		if err != nil {
			return err
		}
		if sw != len(o.Fanins)-1 {
			return ErrDimension
		}
		for _, id := range o.Fanins[1:] {
			fw, err := width(id)
			if err != nil || fw != w {
				return ErrDimension
			}
		}
	case WLCObjectBitConcat:
		total := 0
		for _, id := range o.Fanins {
			fw, err := width(id)
			if err != nil || total > 64-fw {
				return ErrDimension
			}
			total += fw
		}
		if total != w {
			return ErrDimension
		}
	case WLCObjectZeroPad, WLCObjectSignExtend:
		fw, err := width(o.Fanins[0])
		if err != nil || fw > w {
			return ErrDimension
		}
	case WLCObjectRotateLeft, WLCObjectRotateRight:
		fw, err := width(o.Fanins[0])
		if err != nil || fw != w {
			return ErrDimension
		}
	case WLCObjectBitSelect:
		fw, err := width(o.Fanins[0])
		if err != nil {
			return err
		}
		fr := objects[o.Fanins[0]].Range
		lo, hi := fr.Begin, fr.End
		if lo > hi {
			lo, hi = hi, lo
		}
		selectLo, selectHi := o.ParamA, o.ParamB
		if selectLo > selectHi {
			selectLo, selectHi = selectHi, selectLo
		}
		if selectLo < lo || selectHi > hi || selectHi-selectLo+1 != w || fw < w {
			return ErrDimension
		}
	}
	return nil
}

func (n *WLCNetwork) addObject(o WLCObject) (int, error) {
	if n == nil || n.closed {
		return 0, ErrInvariant
	}
	if o.Type <= WLCObjectNone || o.Type >= WLCObjectTypeCount || !wlcLegalName(o.Name) || !wlcShape(o.Type, len(o.Fanins)) {
		return 0, ErrDimension
	}
	w, err := o.Range.Width()
	if err != nil {
		return 0, err
	}
	if len(o.Fanins) > n.limits.MaxFanins || len(n.objects) > n.limits.MaxObjects || n.edges > n.limits.MaxEdges-len(o.Fanins) || n.bits > n.limits.MaxBits-w {
		return 0, ErrLimit
	}
	if _, ok := n.names[o.Name]; ok {
		return 0, ErrInvariant
	}
	for _, f := range o.Fanins {
		if f <= 0 || f >= len(n.objects) {
			return 0, ErrInvariant
		}
	}
	if err := wlcObjectSemantics(n.objects, o); err != nil {
		return 0, err
	}
	addBytes := len(o.Name) + 48 + 8*len(o.Fanins)
	if o.Type == WLCObjectPI {
		addBytes += 8
	}
	if o.IsOutput {
		addBytes += 8
	}
	if addBytes < len(o.Name) || n.bytes > n.limits.MaxBytes-addBytes {
		return 0, ErrLimit
	}
	if n.limits.Stop != nil && n.limits.Stop() {
		return 0, ErrCancelled
	}
	o.ID = len(n.objects)
	o.Value &= wlcMask(w)
	o.Fanins = slices.Clone(o.Fanins)
	n.objects = append(n.objects, o)
	n.names[o.Name] = o.ID
	n.edges += len(o.Fanins)
	n.bits += w
	n.bytes += addBytes
	if o.Type == WLCObjectPI {
		n.inputs = append(n.inputs, o.ID)
	}
	if o.IsOutput {
		n.outputs = append(n.outputs, o.ID)
	}
	return o.ID, nil
}

func (n *WLCNetwork) AddInput(name string, r WLCRange, signed bool) (int, error) {
	return n.addObject(WLCObject{Type: WLCObjectPI, Name: name, Range: r, Signed: signed})
}

func (n *WLCNetwork) AddConstant(name string, r WLCRange, value uint64, xValue bool) (int, error) {
	return n.addObject(WLCObject{Type: WLCObjectConst, Name: name, Range: r, Value: value, XValue: xValue})
}

func (n *WLCNetwork) AddNode(name string, typ WLCObjectType, r WLCRange, signed bool, fanins []int) (int, error) {
	if typ == WLCObjectPI || typ == WLCObjectFO || typ == WLCObjectConst || typ == WLCObjectBitSelect || typ == WLCObjectLUT {
		return 0, ErrDimension
	}
	return n.addObject(WLCObject{Type: typ, Name: name, Range: r, Signed: signed, Fanins: fanins})
}

func (n *WLCNetwork) AddSlice(name string, source, end, begin int, signed bool) (int, error) {
	if n == nil || n.closed || source <= 0 || source >= len(n.objects) {
		return 0, ErrInvariant
	}
	sr := n.objects[source].Range
	minBit, maxBit := sr.Begin, sr.End
	if minBit > maxBit {
		minBit, maxBit = maxBit, minBit
	}
	if end < minBit || end > maxBit || begin < minBit || begin > maxBit {
		return 0, ErrDimension
	}
	r := WLCRange{End: end, Begin: begin}
	id, err := n.addObject(WLCObject{Type: WLCObjectBitSelect, Name: name, Range: r, Signed: signed, Fanins: []int{source}, ParamA: end, ParamB: begin})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (n *WLCNetwork) AddLUT(name string, source int, outWidth int, truth uint64) (int, error) {
	if outWidth <= 0 || outWidth > 64 || n == nil || n.closed || source <= 0 || source >= len(n.objects) {
		return 0, ErrDimension
	}
	inWidth, err := n.objects[source].Range.Width()
	if err != nil || inWidth > 6 {
		return 0, ErrDimension
	}
	return n.addObject(WLCObject{Type: WLCObjectLUT, Name: name, Range: WLCRange{End: outWidth - 1}, Fanins: []int{source}, Table: truth})
}

func (n *WLCNetwork) rollbackLast(id int) {
	if n == nil || id != len(n.objects)-1 {
		return
	}
	o := n.objects[id]
	w, _ := o.Range.Width()
	delete(n.names, o.Name)
	n.edges -= len(o.Fanins)
	n.bits -= w
	dropBytes := len(o.Name) + 48 + 8*len(o.Fanins)
	if o.Type == WLCObjectPI {
		dropBytes += 8
	}
	if o.IsOutput {
		dropBytes += 8
	}
	n.bytes -= dropBytes
	n.objects = n.objects[:id]
}

func (n *WLCNetwork) MarkOutput(id int) error {
	if n == nil || n.closed || id <= 0 || id >= len(n.objects) || n.objects[id].IsOutput {
		return ErrInvariant
	}
	if n.limits.Stop != nil && n.limits.Stop() {
		return ErrCancelled
	}
	if n.bytes > n.limits.MaxBytes-8 {
		return ErrLimit
	}
	n.objects[id].IsOutput = true
	n.outputs = append(n.outputs, id)
	n.bytes += 8
	return nil
}

func (n *WLCNetwork) AddFlop(name string, r WLCRange, signed bool, next int, initial uint64) (int, error) {
	if next <= 0 || n == nil || next >= len(n.objects) {
		return 0, ErrInvariant
	}
	id, err := n.AddFlopOutput(name, r, signed, initial)
	if err != nil {
		return 0, err
	}
	if err := n.SetFlopNext(id, next); err != nil {
		n.flops = n.flops[:len(n.flops)-1]
		n.bytes -= 24
		n.rollbackLast(id)
		return 0, err
	}
	return id, nil
}

// AddFlopOutput creates a state input before its next-state cone is built.
func (n *WLCNetwork) AddFlopOutput(name string, r WLCRange, signed bool, initial uint64) (int, error) {
	id, err := n.addObject(WLCObject{Type: WLCObjectFO, Name: name, Range: r, Signed: signed})
	if err != nil {
		return 0, err
	}
	w, _ := r.Width()
	if len(n.flops) >= n.limits.MaxObjects || n.bytes > n.limits.MaxBytes-24 {
		n.rollbackLast(id)
		return 0, ErrLimit
	}
	n.flops = append(n.flops, WLCFlop{Object: id, Initial: initial & wlcMask(w)})
	n.bytes += 24
	return id, nil
}

// SetFlopNext completes a previously created flop output transactionally.
func (n *WLCNetwork) SetFlopNext(object, next int) error {
	if n == nil || n.closed || object <= 0 || object >= len(n.objects) || next <= 0 || next >= len(n.objects) {
		return ErrInvariant
	}
	if n.limits.Stop != nil && n.limits.Stop() {
		return ErrCancelled
	}
	for i := range n.flops {
		if n.flops[i].Object != object {
			continue
		}
		if n.flops[i].Next != 0 {
			return ErrInvariant
		}
		ow, err1 := n.objects[object].Range.Width()
		nw, err2 := n.objects[next].Range.Width()
		if err1 != nil || err2 != nil || ow != nw || n.objects[next].IsFI {
			return ErrDimension
		}
		n.flops[i].Next = next
		n.objects[next].IsFI = true
		return nil
	}
	return ErrInvariant
}

func (n *WLCNetwork) Object(id int) (WLCObject, error) {
	if n == nil || n.closed || id <= 0 || id >= len(n.objects) {
		return WLCObject{}, ErrInvariant
	}
	if len(n.objects[id].Fanins) > (n.limits.MaxBytes-n.bytes)/8 {
		return WLCObject{}, ErrLimit
	}
	if n.limits.Stop != nil && n.limits.Stop() {
		return WLCObject{}, ErrCancelled
	}
	return n.objects[id].clone(), nil
}

func (n *WLCNetwork) Objects() ([]WLCObject, error) {
	if n == nil || n.closed {
		return nil, ErrInvariant
	}
	temp := 0
	for i := 1; i < len(n.objects); i++ {
		if temp > n.limits.MaxBytes-48-8*len(n.objects[i].Fanins) {
			return nil, ErrLimit
		}
		temp += 48 + 8*len(n.objects[i].Fanins)
	}
	if temp > n.limits.MaxBytes-n.bytes {
		return nil, ErrLimit
	}
	r := make([]WLCObject, len(n.objects)-1)
	for i := 1; i < len(n.objects); i++ {
		if n.limits.Stop != nil && n.limits.Stop() {
			return nil, ErrCancelled
		}
		r[i-1] = n.objects[i].clone()
	}
	return r, nil
}

func (n *WLCNetwork) Inputs() []int {
	if n == nil || n.closed {
		return nil
	}
	return slices.Clone(n.inputs)
}
func (n *WLCNetwork) Outputs() []int {
	if n == nil || n.closed {
		return nil
	}
	return slices.Clone(n.outputs)
}
func (n *WLCNetwork) Flops() []WLCFlop {
	if n == nil || n.closed {
		return nil
	}
	return slices.Clone(n.flops)
}

func wlcMask(width int) uint64 {
	if width >= 64 {
		return ^uint64(0)
	}
	if width <= 0 {
		return 0
	}
	return uint64(1)<<width - 1
}

func wlcSigned(v uint64, width int) int64 {
	v &= wlcMask(width)
	if width == 64 {
		return int64(v)
	}
	if width > 0 && v&(uint64(1)<<(width-1)) != 0 {
		v |= ^wlcMask(width)
	}
	return int64(v)
}

func wlcExtend(v uint64, from, to int, signed bool) uint64 {
	if signed {
		return uint64(wlcSigned(v, from)) & wlcMask(to)
	}
	return v & wlcMask(to)
}

func wlcISqrt(v uint64) uint64 {
	var r uint64
	bit := uint64(1) << 62
	for bit > v {
		bit >>= 2
	}
	for ; bit != 0; bit >>= 2 {
		if v >= r+bit {
			v -= r + bit
			r = (r >> 1) + bit
		} else {
			r >>= 1
		}
	}
	return r
}

func (n *WLCNetwork) evalObject(o WLCObject, values []uint64, b *wlcBudget) (uint64, error) {
	w, _ := o.Range.Width()
	if err := b.use(1 + len(o.Fanins)); err != nil {
		return 0, err
	}
	arg := func(i int) uint64 { return values[o.Fanins[i]] }
	argWidth := func(i int) int { x, _ := n.objects[o.Fanins[i]].Range.Width(); return x }
	mask := wlcMask(w)
	switch o.Type {
	case WLCObjectConst:
		if o.XValue {
			return 0, ErrUnsupported
		}
		return o.Value & mask, nil
	case WLCObjectBuf, WLCObjectZeroPad:
		return arg(0) & mask, nil
	case WLCObjectSignExtend:
		return wlcExtend(arg(0), argWidth(0), w, true), nil
	case WLCObjectBitNot:
		return ^wlcExtend(arg(0), argWidth(0), w, o.Signed) & mask, nil
	case WLCObjectBitAnd, WLCObjectBitOr, WLCObjectBitXor, WLCObjectBitNand, WLCObjectBitNor, WLCObjectBitXnor:
		a := wlcExtend(arg(0), argWidth(0), w, o.Signed)
		c := wlcExtend(arg(1), argWidth(1), w, o.Signed)
		var v uint64
		switch o.Type {
		case WLCObjectBitAnd, WLCObjectBitNand:
			v = a & c
		case WLCObjectBitOr, WLCObjectBitNor:
			v = a | c
		default:
			v = a ^ c
		}
		if o.Type == WLCObjectBitNand || o.Type == WLCObjectBitNor || o.Type == WLCObjectBitXnor {
			v = ^v
		}
		return v & mask, nil
	case WLCObjectMux:
		sel := arg(0)
		idx := 0
		if sel < uint64(len(o.Fanins)-1) {
			idx = int(sel)
		} else {
			idx = 0
		}
		return wlcExtend(arg(idx+1), argWidth(idx+1), w, n.objects[o.Fanins[idx+1]].Signed), nil
	case WLCObjectShiftLeft, WLCObjectArithmeticShiftLeft:
		s := arg(1)
		if s >= uint64(w) {
			return 0, nil
		}
		return arg(0) << s & mask, nil
	case WLCObjectShiftRight:
		s := arg(1)
		if s >= uint64(w) {
			return 0, nil
		}
		return arg(0) >> s & mask, nil
	case WLCObjectArithmeticShiftRight:
		s := arg(1)
		if s >= uint64(w) {
			if wlcSigned(arg(0), argWidth(0)) < 0 {
				return mask, nil
			}
			return 0, nil
		}
		return uint64(wlcSigned(arg(0), argWidth(0))>>s) & mask, nil
	case WLCObjectRotateLeft, WLCObjectRotateRight:
		if w == 0 {
			return 0, ErrDimension
		}
		s := int(arg(1) % uint64(w))
		v := arg(0) & mask
		if s == 0 {
			return v, nil
		}
		if o.Type == WLCObjectRotateLeft {
			return (v<<s | v>>(w-s)) & mask, nil
		}
		return (v>>s | v<<(w-s)) & mask, nil
	case WLCObjectBitSelect:
		lo := o.ParamA
		hi := o.ParamB
		if lo > hi {
			lo, hi = hi, lo
		}
		fr := n.objects[o.Fanins[0]].Range
		base, top := fr.Begin, fr.End
		if base > top {
			base, top = top, base
		}
		if lo < base || hi > top {
			return 0, ErrDimension
		}
		return arg(0) >> (lo - base) & mask, nil
	case WLCObjectBitConcat:
		var v uint64
		used := 0
		for i := len(o.Fanins) - 1; i >= 0; i-- {
			fw := argWidth(i)
			if used > 64-fw {
				return 0, ErrLimit
			}
			v |= (arg(i) & wlcMask(fw)) << used
			used += fw
		}
		return v & mask, nil
	case WLCObjectLogicNot:
		if arg(0) == 0 {
			return 1, nil
		}
		return 0, nil
	case WLCObjectLogicImplies:
		if arg(0) == 0 || arg(1) != 0 {
			return 1, nil
		}
		return 0, nil
	case WLCObjectLogicAnd:
		if arg(0) != 0 && arg(1) != 0 {
			return 1, nil
		}
		return 0, nil
	case WLCObjectLogicOr:
		if arg(0) != 0 || arg(1) != 0 {
			return 1, nil
		}
		return 0, nil
	case WLCObjectLogicXor:
		if (arg(0) != 0) != (arg(1) != 0) {
			return 1, nil
		}
		return 0, nil
	case WLCObjectEqual, WLCObjectNotEqual:
		eq := arg(0) == arg(1)
		if (o.Type == WLCObjectEqual) == eq {
			return 1, nil
		}
		return 0, nil
	case WLCObjectLess, WLCObjectGreater, WLCObjectLessEqual, WLCObjectGreaterEqual:
		var less, equal bool
		if o.Signed {
			a, c := wlcSigned(arg(0), argWidth(0)), wlcSigned(arg(1), argWidth(1))
			less = a < c
			equal = a == c
		} else {
			less = arg(0) < arg(1)
			equal = arg(0) == arg(1)
		}
		ok := less
		if o.Type == WLCObjectGreater {
			ok = !less && !equal
		}
		if o.Type == WLCObjectLessEqual {
			ok = less || equal
		}
		if o.Type == WLCObjectGreaterEqual {
			ok = !less || equal
		}
		if ok {
			return 1, nil
		}
		return 0, nil
	case WLCObjectReduceAnd, WLCObjectReduceNand:
		v := arg(0) & wlcMask(argWidth(0))
		ok := v == wlcMask(argWidth(0))
		if o.Type == WLCObjectReduceNand {
			ok = !ok
		}
		if ok {
			return 1, nil
		}
		return 0, nil
	case WLCObjectReduceOr, WLCObjectReduceNor:
		ok := arg(0) != 0
		if o.Type == WLCObjectReduceNor {
			ok = !ok
		}
		if ok {
			return 1, nil
		}
		return 0, nil
	case WLCObjectReduceXor, WLCObjectReduceXnor:
		ok := bits.OnesCount64(arg(0)&wlcMask(argWidth(0)))&1 == 1
		if o.Type == WLCObjectReduceXnor {
			ok = !ok
		}
		if ok {
			return 1, nil
		}
		return 0, nil
	case WLCObjectAdd:
		v := arg(0) + arg(1)
		if len(o.Fanins) == 3 {
			v += arg(2) & 1
		}
		return v & mask, nil
	case WLCObjectSub:
		return (arg(0) - arg(1)) & mask, nil
	case WLCObjectMultiply:
		return (arg(0) * arg(1)) & mask, nil
	case WLCObjectDivide, WLCObjectRemainder, WLCObjectModulus:
		if arg(1) == 0 {
			if o.Type == WLCObjectDivide {
				return mask, nil
			}
			return arg(0) & mask, nil
		}
		if o.Signed {
			a, c := wlcSigned(arg(0), argWidth(0)), wlcSigned(arg(1), argWidth(1))
			if c == 0 {
				return 0, ErrInvariant
			}
			if o.Type == WLCObjectDivide {
				return uint64(a/c) & mask, nil
			}
			r := a % c
			if o.Type == WLCObjectModulus && r != 0 && ((r < 0) != (c < 0)) {
				r += c
			}
			return uint64(r) & mask, nil
		}
		if o.Type == WLCObjectDivide {
			return arg(0) / arg(1) & mask, nil
		}
		return arg(0) % arg(1) & mask, nil
	case WLCObjectPower:
		base, exp, res := arg(0)&mask, arg(1), uint64(1)
		for exp > 0 {
			if err := b.use(1); err != nil {
				return 0, err
			}
			if exp&1 != 0 {
				res = res * base & mask
			}
			exp >>= 1
			if exp > 0 {
				base = base * base & mask
			}
		}
		return res & mask, nil
	case WLCObjectMinus:
		return (-arg(0)) & mask, nil
	case WLCObjectSqrt:
		return wlcISqrt(arg(0)) & mask, nil
	case WLCObjectSquare:
		return (arg(0) * arg(0)) & mask, nil
	case WLCObjectAddSub:
		if arg(0)&1 != 0 {
			return (arg(2) + arg(3) + (arg(1) & 1)) & mask, nil
		}
		return (arg(2) - arg(3) - (arg(1) & 1)) & mask, nil
	case WLCObjectSelect:
		var v uint64
		for i := 1; i < len(o.Fanins); i++ {
			if arg(0)&(uint64(1)<<(i-1)) != 0 {
				v |= arg(i)
			}
		}
		return v & mask, nil
	case WLCObjectDecoder:
		idx := arg(0)
		if idx >= 64 {
			return 0, nil
		}
		return uint64(1) << idx & mask, nil
	case WLCObjectLUT:
		idx := arg(0)
		if idx >= 64 {
			return 0, ErrLimit
		}
		return o.Table >> idx & mask, nil
	default:
		return 0, ErrUnsupported
	}
}

func (n *WLCNetwork) Evaluate(inputs, state []uint64) (WLCStep, error) {
	if n == nil || n.closed || len(inputs) != len(n.inputs) || len(state) != len(n.flops) {
		return WLCStep{}, ErrInvariant
	}
	if err := n.Validate(); err != nil {
		return WLCStep{}, err
	}
	tempWords := len(n.objects) + len(n.outputs) + len(n.flops)
	if tempWords < len(n.objects) || tempWords > (n.limits.MaxBytes-n.bytes)/8 {
		return WLCStep{}, ErrLimit
	}
	b := wlcBudget{limits: n.limits}
	values := make([]uint64, len(n.objects))
	for i, id := range n.inputs {
		w, _ := n.objects[id].Range.Width()
		values[id] = inputs[i] & wlcMask(w)
	}
	for i, f := range n.flops {
		w, _ := n.objects[f.Object].Range.Width()
		values[f.Object] = state[i] & wlcMask(w)
	}
	for i := 1; i < len(n.objects); i++ {
		o := n.objects[i]
		if o.Type == WLCObjectPI || o.Type == WLCObjectFO {
			continue
		}
		v, e := n.evalObject(o, values, &b)
		if e != nil {
			return WLCStep{}, e
		}
		values[i] = v
	}
	r := WLCStep{Values: slices.Clone(values), Outputs: make([]uint64, len(n.outputs)), NextState: make([]uint64, len(n.flops))}
	for i, id := range n.outputs {
		r.Outputs[i] = values[id]
	}
	for i, f := range n.flops {
		w, _ := n.objects[f.Object].Range.Width()
		r.NextState[i] = values[f.Next] & wlcMask(w)
	}
	return r, nil
}

func (n *WLCNetwork) InitialState() ([]uint64, error) {
	if n == nil || n.closed {
		return nil, ErrInvariant
	}
	if len(n.flops) > (n.limits.MaxBytes-n.bytes)/8 {
		return nil, ErrLimit
	}
	if n.limits.Stop != nil && n.limits.Stop() {
		return nil, ErrCancelled
	}
	r := make([]uint64, len(n.flops))
	for i, f := range n.flops {
		r[i] = f.Initial
	}
	return r, nil
}

func (n *WLCNetwork) Validate() error {
	if n == nil || n.closed || len(n.objects) == 0 || len(n.names) != len(n.objects)-1 {
		return ErrInvariant
	}
	objects, edges, bitsTotal, bytesTotal := len(n.objects)-1, 0, 0, len(n.Name)+8*len(n.inputs)+8*len(n.outputs)
	if objects > n.limits.MaxObjects {
		return ErrLimit
	}
	b := wlcBudget{limits: n.limits}
	if n.bytes > n.limits.MaxBytes || objects > (n.limits.MaxBytes-n.bytes)/48 {
		return ErrLimit
	}
	seen := make(map[string]bool, objects)
	levels := make([]int, len(n.objects))
	inputSeen := make([]bool, len(n.objects))
	outputSeen := make([]bool, len(n.objects))
	flopSeen := make([]bool, len(n.objects))
	fiSeen := make([]bool, len(n.objects))
	for i := 1; i < len(n.objects); i++ {
		if err := b.use(1 + len(n.objects[i].Fanins)); err != nil {
			return err
		}
		o := n.objects[i]
		if o.ID != i || !wlcLegalName(o.Name) || seen[o.Name] || o.Type <= WLCObjectNone || o.Type >= WLCObjectTypeCount || !wlcShape(o.Type, len(o.Fanins)) || len(o.Fanins) > n.limits.MaxFanins {
			return ErrInvariant
		}
		seen[o.Name] = true
		if n.names[o.Name] != i {
			return ErrInvariant
		}
		w, e := o.Range.Width()
		if e != nil {
			return e
		}
		bitsTotal += w
		edges += len(o.Fanins)
		bytesTotal += len(o.Name) + 48 + 8*len(o.Fanins)
		for _, f := range o.Fanins {
			if f <= 0 || f >= i {
				return ErrInvariant
			}
			if levels[i] < levels[f]+1 {
				levels[i] = levels[f] + 1
			}
		}
		if o.Type == WLCObjectPI || o.Type == WLCObjectFO {
			levels[i] = 0
		}
		if levels[i] > n.limits.MaxDepth {
			return ErrLimit
		}
		if err := wlcObjectSemantics(n.objects, o); err != nil {
			return err
		}
	}
	if edges != n.edges || bitsTotal != n.bits || bytesTotal+24*len(n.flops) != n.bytes || edges > n.limits.MaxEdges || bitsTotal > n.limits.MaxBits || n.bytes > n.limits.MaxBytes {
		return ErrInvariant
	}
	for _, id := range n.inputs {
		if id <= 0 || id >= len(n.objects) || n.objects[id].Type != WLCObjectPI || inputSeen[id] {
			return ErrInvariant
		}
		inputSeen[id] = true
	}
	for i := 1; i < len(n.objects); i++ {
		if (n.objects[i].Type == WLCObjectPI) != inputSeen[i] {
			return ErrInvariant
		}
	}
	for _, id := range n.outputs {
		if id <= 0 || id >= len(n.objects) || !n.objects[id].IsOutput || outputSeen[id] {
			return ErrInvariant
		}
		outputSeen[id] = true
	}
	for i := 1; i < len(n.objects); i++ {
		if n.objects[i].IsOutput != outputSeen[i] {
			return ErrInvariant
		}
	}
	for _, f := range n.flops {
		if f.Object <= 0 || f.Object >= len(n.objects) || n.objects[f.Object].Type != WLCObjectFO || f.Next <= 0 || f.Next >= len(n.objects) || flopSeen[f.Object] || fiSeen[f.Next] {
			return ErrInvariant
		}
		ow, _ := n.objects[f.Object].Range.Width()
		nw, _ := n.objects[f.Next].Range.Width()
		if ow != nw || !n.objects[f.Next].IsFI {
			return ErrInvariant
		}
		flopSeen[f.Object], fiSeen[f.Next] = true, true
	}
	for i := 1; i < len(n.objects); i++ {
		if (n.objects[i].Type == WLCObjectFO) != flopSeen[i] || n.objects[i].IsFI != fiSeen[i] {
			return ErrInvariant
		}
	}
	return nil
}

func (n *WLCNetwork) Levels() ([]int, error) {
	if n == nil || n.closed {
		return nil, ErrInvariant
	}
	if err := n.checkWork(2, 0); err != nil {
		return nil, err
	}
	if err := n.checkTempBytes(8 * len(n.objects)); err != nil {
		return nil, err
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	levels := make([]int, len(n.objects))
	b := wlcBudget{limits: n.limits}
	for i := 1; i < len(n.objects); i++ {
		if err := b.use(1 + len(n.objects[i].Fanins)); err != nil {
			return nil, err
		}
		if n.objects[i].Type == WLCObjectPI || n.objects[i].Type == WLCObjectFO {
			continue
		}
		m := 0
		for _, f := range n.objects[i].Fanins {
			if levels[f] > m {
				m = levels[f]
			}
		}
		levels[i] = m + 1
		if levels[i] > n.limits.MaxDepth {
			return nil, ErrLimit
		}
	}
	return levels, nil
}

func (n *WLCNetwork) ReverseLevels() ([]int, error) {
	if n == nil || n.closed {
		return nil, ErrInvariant
	}
	if err := n.checkWork(2, len(n.outputs)+len(n.flops)); err != nil {
		return nil, err
	}
	if err := n.checkTempBytes(8 * (len(n.objects) + len(n.outputs) + len(n.flops))); err != nil {
		return nil, err
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	r := make([]int, len(n.objects))
	b := wlcBudget{limits: n.limits}
	roots := slices.Clone(n.outputs)
	for _, f := range n.flops {
		roots = append(roots, f.Next)
	}
	for _, id := range roots {
		if r[id] < 1 {
			r[id] = 1
		}
	}
	for i := len(n.objects) - 1; i > 0; i-- {
		if err := b.use(1 + len(n.objects[i].Fanins)); err != nil {
			return nil, err
		}
		for _, f := range n.objects[i].Fanins {
			if r[f] < r[i]+1 {
				r[f] = r[i] + 1
			}
			if r[f] > n.limits.MaxDepth {
				return nil, ErrLimit
			}
		}
	}
	return r, nil
}

func (n *WLCNetwork) References() ([]int, error) {
	if n == nil || n.closed {
		return nil, ErrInvariant
	}
	if err := n.checkWork(2, len(n.outputs)+len(n.flops)); err != nil {
		return nil, err
	}
	if err := n.checkTempBytes(8 * len(n.objects)); err != nil {
		return nil, err
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	r := make([]int, len(n.objects))
	b := wlcBudget{limits: n.limits}
	for i := 1; i < len(n.objects); i++ {
		if err := b.use(1 + len(n.objects[i].Fanins)); err != nil {
			return nil, err
		}
		for _, f := range n.objects[i].Fanins {
			r[f]++
		}
	}
	for _, id := range n.outputs {
		r[id]++
	}
	for _, f := range n.flops {
		r[f.Next]++
	}
	return r, nil
}

func (n *WLCNetwork) Support(root int) ([]int, error) {
	if n == nil || n.closed || root <= 0 || root >= len(n.objects) {
		return nil, ErrInvariant
	}
	if err := n.checkWork(2, 0); err != nil {
		return nil, err
	}
	if err := n.checkTempBytes(9 * len(n.objects)); err != nil {
		return nil, err
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	seen := make([]bool, len(n.objects))
	supp := make([]int, 0)
	b := wlcBudget{limits: n.limits}
	var visit func(int, int) error
	visit = func(id, d int) error {
		if d > n.limits.MaxDepth {
			return ErrLimit
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		if err := b.use(1 + len(n.objects[id].Fanins)); err != nil {
			return err
		}
		o := n.objects[id]
		if o.Type == WLCObjectPI || o.Type == WLCObjectFO {
			supp = append(supp, id)
			return nil
		}
		for _, f := range o.Fanins {
			if err := visit(f, d+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(root, 0); err != nil {
		return nil, err
	}
	sort.Ints(supp)
	return supp, nil
}

func (n *WLCNetwork) ArithmeticWindow(root int) (leaves, nodes []int, err error) {
	if n == nil || n.closed || root <= 0 || root >= len(n.objects) {
		return nil, nil, ErrInvariant
	}
	if err = n.checkWork(2, 0); err != nil {
		return nil, nil, err
	}
	if err = n.checkTempBytes(17 * len(n.objects)); err != nil {
		return nil, nil, err
	}
	if err = n.Validate(); err != nil {
		return nil, nil, err
	}
	seen := make([]bool, len(n.objects))
	b := wlcBudget{limits: n.limits}
	var visit func(int, int) error
	visit = func(id, d int) error {
		if d > n.limits.MaxDepth {
			return ErrLimit
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		if e := b.use(1 + len(n.objects[id].Fanins)); e != nil {
			return e
		}
		o := n.objects[id]
		arith := o.Type >= WLCObjectAdd && o.Type <= WLCObjectSquare
		if !arith {
			leaves = append(leaves, id)
			return nil
		}
		for _, f := range o.Fanins {
			if e := visit(f, d+1); e != nil {
				return e
			}
		}
		nodes = append(nodes, id)
		return nil
	}
	err = visit(root, 0)
	sort.Ints(leaves)
	return
}

func (n *WLCNetwork) Stats() (WLCStats, error) {
	if n == nil || n.closed {
		return WLCStats{}, ErrInvariant
	}
	if err := n.checkWork(2, len(n.objects)-1); err != nil {
		return WLCStats{}, err
	}
	levels, err := n.Levels()
	if err != nil {
		return WLCStats{}, err
	}
	s := WLCStats{Objects: len(n.objects) - 1, Edges: n.edges, Bits: n.bits, Inputs: len(n.inputs), Outputs: len(n.outputs), Flops: len(n.flops)}
	for i := 1; i < len(n.objects); i++ {
		if n.limits.Stop != nil && n.limits.Stop() {
			return WLCStats{}, ErrCancelled
		}
		s.TypeCounts[n.objects[i].Type]++
		if levels[i] > s.MaxLevel {
			s.MaxLevel = levels[i]
		}
	}
	return s, nil
}

func (n *WLCNetwork) Clone() (*WLCNetwork, error) {
	if n == nil || n.closed {
		return nil, ErrInvariant
	}
	if err := n.checkWork(1, 2*(len(n.objects)-1)); err != nil {
		return nil, err
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	if n.bytes > n.limits.MaxBytes-n.bytes {
		return nil, ErrLimit
	}
	r := &WLCNetwork{Name: n.Name, objects: make([]WLCObject, len(n.objects)), inputs: slices.Clone(n.inputs), outputs: slices.Clone(n.outputs), flops: slices.Clone(n.flops), names: make(map[string]int, len(n.names)), edges: n.edges, bits: n.bits, bytes: n.bytes, limits: n.limits}
	for i := 1; i < len(n.objects); i++ {
		if n.limits.Stop != nil && n.limits.Stop() {
			return nil, ErrCancelled
		}
		r.objects[i] = n.objects[i].clone()
	}
	for k, v := range n.names {
		r.names[k] = v
	}
	return r, nil
}

func (n *WLCNetwork) Close() {
	if n == nil || n.closed {
		return
	}
	n.closed = true
	n.Name = ""
	n.objects = nil
	n.inputs = nil
	n.outputs = nil
	n.flops = nil
	n.names = nil
	n.edges = 0
	n.bits = 0
	n.bytes = 0
}

// WLCReport returns bounded deterministic diagnostics instead of process output.
func (n *WLCNetwork) Report() (string, error) {
	s, err := n.Stats()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	appendLine := func(line string) error {
		if len(line) > n.limits.MaxTextBytes-b.Len() {
			return ErrLimit
		}
		b.WriteString(line)
		if n.limits.Stop != nil && n.limits.Stop() {
			return ErrCancelled
		}
		return nil
	}
	if err := appendLine(fmt.Sprintf("wlc %s objects=%d edges=%d bits=%d pi=%d po=%d ff=%d level=%d\n", n.Name, s.Objects, s.Edges, s.Bits, s.Inputs, s.Outputs, s.Flops, s.MaxLevel)); err != nil {
		return "", err
	}
	for t := WLCObjectType(1); t < WLCObjectTypeCount; t++ {
		if s.TypeCounts[t] > 0 {
			if err := appendLine(fmt.Sprintf("%s=%d\n", t.String(), s.TypeCounts[t])); err != nil {
				return "", err
			}
		}
	}
	return b.String(), nil
}

func wlcToCBAType(t WLCObjectType) (CBAObjectType, bool) {
	switch t {
	case WLCObjectBuf:
		return CBABoxBuf, true
	case WLCObjectBitNot:
		return CBABoxInv, true
	case WLCObjectBitAnd:
		return CBABoxAnd, true
	case WLCObjectBitNand:
		return CBABoxNand, true
	case WLCObjectBitOr:
		return CBABoxOr, true
	case WLCObjectBitNor:
		return CBABoxNor, true
	case WLCObjectBitXor:
		return CBABoxXor, true
	case WLCObjectBitXnor:
		return CBABoxXnor, true
	case WLCObjectMux:
		return CBABoxMux, true
	case WLCObjectReduceAnd:
		return CBABoxReduceAnd, true
	case WLCObjectReduceNand:
		return CBABoxReduceNand, true
	case WLCObjectReduceOr:
		return CBABoxReduceOr, true
	case WLCObjectReduceNor:
		return CBABoxReduceNor, true
	case WLCObjectReduceXor:
		return CBABoxReduceXor, true
	case WLCObjectReduceXnor:
		return CBABoxReduceXnor, true
	case WLCObjectLogicNot:
		return CBABoxLogicalNot, true
	case WLCObjectLogicAnd:
		return CBABoxLogicalAnd, true
	case WLCObjectLogicOr:
		return CBABoxLogicalOr, true
	case WLCObjectAdd:
		return CBABoxAdd, true
	case WLCObjectSub:
		return CBABoxSub, true
	case WLCObjectMultiply:
		return CBABoxMul, true
	case WLCObjectDivide:
		return CBABoxDiv, true
	case WLCObjectRemainder:
		return CBABoxRem, true
	case WLCObjectModulus:
		return CBABoxMod, true
	case WLCObjectPower:
		return CBABoxPow, true
	case WLCObjectSqrt:
		return CBABoxSqrt, true
	case WLCObjectLess:
		return CBABoxLess, true
	case WLCObjectLessEqual:
		return CBABoxLessEqual, true
	case WLCObjectGreater:
		return CBABoxMore, true
	case WLCObjectGreaterEqual:
		return CBABoxMoreEqual, true
	case WLCObjectEqual:
		return CBABoxEqual, true
	case WLCObjectNotEqual:
		return CBABoxNotEqual, true
	case WLCObjectShiftLeft:
		return CBABoxShiftLeft, true
	case WLCObjectShiftRight:
		return CBABoxShiftRight, true
	case WLCObjectArithmeticShiftLeft:
		return CBABoxArithmeticShiftLeft, true
	case WLCObjectArithmeticShiftRight:
		return CBABoxArithmeticShiftRight, true
	case WLCObjectRotateLeft:
		return CBABoxRotateLeft, true
	case WLCObjectRotateRight:
		return CBABoxRotateRight, true
	case WLCObjectMinus:
		return CBABoxMin, true
	case WLCObjectBitConcat:
		return CBABoxConcat, true
	default:
		return 0, false
	}
}

func cbaToWLCType(t CBAObjectType) (WLCObjectType, bool) {
	for w := WLCObjectType(1); w < WLCObjectTypeCount; w++ {
		if c, ok := wlcToCBAType(w); ok && c == t {
			return w, true
		}
	}
	return 0, false
}

// WLCWriteVerilog writes the concrete combinational subset through the reviewed CBA structural writer.
func WLCWriteVerilog(n *WLCNetwork) (string, error) {
	if err := n.Validate(); err != nil {
		return "", err
	}
	if len(n.flops) > 0 {
		return "", ErrUnsupported
	}
	cl := CBALimits{MaxObjects: n.limits.MaxObjects, MaxEdges: n.limits.MaxEdges, MaxBits: n.limits.MaxBits, MaxOperations: n.limits.MaxOperations, MaxBytes: n.limits.MaxBytes, MaxTextBytes: n.limits.MaxTextBytes, MaxDepth: n.limits.MaxDepth, Stop: n.limits.Stop}
	cn, err := NewCBANetwork(n.Name, cl)
	if err != nil {
		return "", err
	}
	ids := make([]int, len(n.objects))
	for i := 1; i < len(n.objects); i++ {
		o := n.objects[i]
		w, _ := o.Range.Width()
		switch o.Type {
		case WLCObjectPI:
			ids[i], err = cn.AddInput(o.Name, w, o.Signed)
		case WLCObjectConst:
			ids[i], err = cn.AddConstant(o.Name, w, o.Value)
		default:
			ct, ok := wlcToCBAType(o.Type)
			if !ok {
				return "", ErrUnsupported
			}
			fs := make([]int, len(o.Fanins))
			for j, f := range o.Fanins {
				fs[j] = ids[f]
			}
			if o.Type == WLCObjectBitConcat {
				ids[i], err = cn.AddNode(o.Name, ct, w, o.Signed, fs)
			} else {
				ids[i], err = cn.AddNode(o.Name, ct, w, o.Signed, fs)
			}
		}
		if err != nil {
			return "", err
		}
	}
	for _, id := range n.outputs {
		o := n.objects[id]
		w, _ := o.Range.Width()
		if _, err = cn.AddOutput(o.Name+"_out", w, ids[id]); err != nil {
			return "", err
		}
	}
	d, err := NewCBADesign(n.Name, cl)
	if err != nil {
		return "", err
	}
	if _, err = d.AddNetwork(cn); err != nil {
		return "", err
	}
	return CBAWriteVerilog(d, cl)
}

// WLCParseVerilog parses the reviewed combinational structural subset.
func WLCParseVerilog(text string, limits WLCLimits) (*WLCNetwork, error) {
	limits, err := normalizeWLCLimits(limits)
	if err != nil {
		return nil, err
	}
	cl := CBALimits{MaxModules: 1, MaxObjects: limits.MaxObjects, MaxEdges: limits.MaxEdges, MaxBits: limits.MaxBits, MaxOperations: limits.MaxOperations, MaxBytes: limits.MaxBytes, MaxTextBytes: limits.MaxTextBytes, MaxDepth: limits.MaxDepth, Stop: limits.Stop}
	d, err := CBAParseVerilog(text, cl)
	if err != nil {
		return nil, err
	}
	cn, err := d.Root()
	if err != nil {
		return nil, err
	}
	n, err := NewWLCNetwork(cn.Name, limits)
	if err != nil {
		return nil, err
	}
	objs, err := cn.Objects()
	if err != nil {
		return nil, err
	}
	ids := make(map[int]int, len(objs))
	for _, o := range objs {
		switch o.Type {
		case CBAObjectPI:
			ids[o.ID], err = n.AddInput(o.Name, WLCRange{End: o.Width - 1}, o.Signed)
		case CBABoxConstFalse, CBABoxConstTrue:
			ids[o.ID], err = n.AddConstant(o.Name, WLCRange{End: o.Width - 1}, o.Value, false)
		case CBAObjectPO:
			src := ids[o.Inputs[0]]
			err = n.MarkOutput(src)
			ids[o.ID] = src
		default:
			wt, ok := cbaToWLCType(o.Type)
			if !ok {
				return nil, ErrUnsupported
			}
			fs := make([]int, len(o.Inputs))
			for i, f := range o.Inputs {
				fs[i] = ids[f]
			}
			ids[o.ID], err = n.AddNode(o.Name, wt, WLCRange{End: o.Width - 1}, o.Signed, fs)
		}
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}
