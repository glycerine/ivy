package abc

// Word-level network support ported from src/base/wln.

import (
	"fmt"
	"math/bits"
	"slices"
	"strconv"
	"strings"
)

type WlnOperation uint8

const (
	WlnNone WlnOperation = iota
	WlnPI
	WlnPO
	WlnCI
	WlnCO
	WlnFON
	WlnBox
	WlnConstFalse
	WlnConstTrue
	WlnConstX
	WlnConstZ
	WlnBitBuf
	WlnBitInv
	WlnBitAnd
	WlnBitNand
	WlnBitOr
	WlnBitNor
	WlnBitXor
	WlnBitNxor
	WlnBitSharp
	WlnBitSharpL
	WlnBitMux
	WlnBitMaj
	WlnABC
	WlnBA
	WlnBO
	WlnBX
	WlnBN
	WlnBAO
	WlnBOA
	WlnReduceAnd
	WlnReduceNand
	WlnReduceOr
	WlnReduceNor
	WlnReduceXor
	WlnReduceNxor
	WlnLogicNot
	WlnLogicAnd
	WlnLogicNand
	WlnLogicOr
	WlnLogicNor
	WlnLogicXor
	WlnLogicXnor
	WlnSelectNMux
	WlnSelect
	WlnPrioritySelect
	WlnEncode
	WlnPriorityEncode
	WlnDecode
	WlnEnableDecode
	WlnAdd
	WlnSub
	WlnMul
	WlnSignedMul
	WlnDiv
	WlnMod
	WlnRem
	WlnPow
	WlnNeg
	WlnSqrt
	WlnAbs
	WlnSignedLess
	WlnLess
	WlnLessEqual
	WlnMoreEqual
	WlnMore
	WlnEqual
	WlnNotEqual
	WlnShiftLeft
	WlnShiftRight
	WlnArithShiftLeft
	WlnArithShiftRight
	WlnRotateLeft
	WlnRotateRight
	WlnNode
	WlnLUT
	WlnGate
	WlnTable
	WlnTri
	WlnRAM
	WlnRAMRead
	WlnRAMWrite
	WlnRAMWriteControl
	WlnRAMLatch
	WlnRAMSelect
	WlnRAMBox
	WlnLatch
	WlnLatchRS
	WlnDFF
	WlnDFFRSE
	WlnDFFLast
	WlnSlice
	WlnConcat
	WlnZeroPad
	WlnSignExtend
	WlnLogicImplies
	WlnSquare
	WlnConst
	WlnAddSub
	WlnOperationLast
)

var wlnOperationNames = [...]string{
	"none", "pi", "po", "ci", "co", "fon", "box", "const0", "const1", "constX", "constZ",
	"buf", "~", "&", "~&", "|", "~|", "^", "~^", "sharp", "sharpL", "mux", "maj",
	"abc", "ba", "bo", "bx", "bn", "bao", "boa", "&", "~&", "|", "~|", "^", "~^",
	"!", "&&", "~&&", "||", "~||", "^^", "~^^", "nmux", "sel", "psel", "enc", "penc", "dec", "edec",
	"+", "-", "*", "*", "/", "mod", "%", "**", "-", "sqrt", "abs", "s<", "<", "<=", ">=", ">", "==", "!=",
	"<<", ">>", "<<<", ">>>", "rotL", "rotR", "node", "lut", "gate", "table", "tri", "ram", "ramr", "ramw", "ramwc", "raml", "rams", "rambox",
	"latch", "latchrs", "dff", "DFFRSE", "dfflast", "[:]", "{}", "zPad", "sExt", "=>", "squar", "const", "addsub",
}

func (operation WlnOperation) String() string {
	if operation >= WlnOperationLast {
		return "unknown"
	}
	return wlnOperationNames[operation]
}

type WlnRange struct{ End, Begin int }

func (wordRange WlnRange) Width() int {
	difference := wordRange.End - wordRange.Begin
	if difference < 0 {
		difference = -difference
	}
	return difference + 1
}

type WlnObject struct {
	Type       WlnOperation
	Signed     bool
	Range      WlnRange
	Fanins     []int
	Name       string
	NameID     int
	InstanceID int
	Parameter  int
	Constant   string
	Table      []uint64
}

type WlnLimits struct {
	MaxObjects, MaxFanins, MaxWidth, MaxSteps int
	Stop                                      func() bool
}

type WlnNetwork struct {
	Name, Spec string
	SMTLib     bool
	// UseInstanceDelays corresponds to the source network's allocated vInstIds.
	// When set, zero InstanceID values are intentional zero delays.
	UseInstanceDelays bool
	objects           []WlnObject
	cis, cos          []int
	ffs               []int
	limits            WlnLimits
}

func normalizeWlnLimits(limits WlnLimits) (WlnLimits, error) {
	if limits.MaxObjects < 0 || limits.MaxFanins < 0 || limits.MaxWidth < 0 || limits.MaxSteps < 0 {
		return limits, ErrDimension
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 20
	}
	if limits.MaxFanins == 0 {
		limits.MaxFanins = 1 << 22
	}
	if limits.MaxWidth == 0 {
		limits.MaxWidth = 1 << 20
	}
	if limits.MaxSteps == 0 {
		limits.MaxSteps = 1 << 20
	}
	return limits, nil
}

func NewWlnNetwork(name string, limits WlnLimits) (*WlnNetwork, error) {
	var err error
	limits, err = normalizeWlnLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &WlnNetwork{Name: name, objects: []WlnObject{{Type: WlnNone}}, limits: limits}, nil
}

func cloneWlnObject(object WlnObject) WlnObject {
	object.Fanins = slices.Clone(object.Fanins)
	object.Table = slices.Clone(object.Table)
	return object
}

func (network *WlnNetwork) Objects() []WlnObject {
	if network == nil {
		return nil
	}
	objects := make([]WlnObject, len(network.objects))
	for index := range objects {
		objects[index] = cloneWlnObject(network.objects[index])
	}
	return objects
}
func (network *WlnNetwork) CIs() []int {
	if network == nil {
		return nil
	}
	return slices.Clone(network.cis)
}
func (network *WlnNetwork) COs() []int {
	if network == nil {
		return nil
	}
	return slices.Clone(network.cos)
}
func (network *WlnNetwork) Flops() []int {
	if network == nil {
		return nil
	}
	return slices.Clone(network.ffs)
}
func (network *WlnNetwork) ObjectCount() int {
	if network == nil {
		return 0
	}
	return len(network.objects)
}
func (network *WlnNetwork) PICount() int {
	if network == nil {
		return 0
	}
	return len(network.cis)
}
func (network *WlnNetwork) POCount() int {
	if network == nil {
		return 0
	}
	return len(network.cos)
}

func (network *WlnNetwork) AddObject(object WlnObject) (int, error) {
	if network == nil || object.Type >= WlnOperationLast || object.Range.Width() > network.limits.MaxWidth || len(object.Fanins) > network.limits.MaxFanins {
		return 0, ErrDimension
	}
	if len(network.objects) == network.limits.MaxObjects {
		return 0, ErrLimit
	}
	if network.limits.Stop != nil && network.limits.Stop() {
		return 0, ErrCancelled
	}
	for _, fanin := range object.Fanins {
		if fanin < 0 || fanin >= len(network.objects) {
			return 0, ErrDimension
		}
	}
	id := len(network.objects)
	object = cloneWlnObject(object)
	if object.Name == "" {
		object.Name = "n" + strconv.Itoa(id)
	}
	network.objects = append(network.objects, object)
	switch object.Type {
	case WlnCI:
		network.cis = append(network.cis, id)
	case WlnCO:
		network.cos = append(network.cos, id)
	case WlnDFFRSE:
		network.ffs = append(network.ffs, id)
	}
	return id, nil
}

func (network *WlnNetwork) SetFanin(object, index, fanin int) error {
	if err := network.checkID(object); err != nil {
		return err
	}
	if fanin < 0 || fanin >= len(network.objects) || index < 0 || index >= len(network.objects[object].Fanins) {
		return ErrDimension
	}
	network.objects[object].Fanins[index] = fanin
	return nil
}
func (network *WlnNetwork) checkID(id int) error {
	if network == nil || id <= 0 || id >= len(network.objects) {
		return ErrDimension
	}
	return nil
}

func (network *WlnNetwork) Check() error {
	if network == nil || len(network.objects) == 0 || network.objects[0].Type != WlnNone || len(network.objects) > network.limits.MaxObjects {
		return ErrInvariant
	}
	if network.limits.Stop != nil && network.limits.Stop() {
		return ErrCancelled
	}
	ci, co, ff, fanins := 0, 0, 0, 0
	for id := 1; id < len(network.objects); id++ {
		if network.limits.Stop != nil && id&255 == 0 && network.limits.Stop() {
			return ErrCancelled
		}
		object := network.objects[id]
		if object.Type == WlnNone || object.Type >= WlnOperationLast || object.Range.Width() > network.limits.MaxWidth {
			return ErrInvariant
		}
		for _, fanin := range object.Fanins {
			if fanin < 0 || fanin >= len(network.objects) {
				return ErrInvariant
			}
		}
		fanins += len(object.Fanins)
		if object.Type == WlnCI {
			if ci >= len(network.cis) || network.cis[ci] != id {
				return ErrInvariant
			}
			ci++
		}
		if object.Type == WlnCO {
			if co >= len(network.cos) || network.cos[co] != id {
				return ErrInvariant
			}
			co++
		}
		if object.Type == WlnDFFRSE {
			if ff >= len(network.ffs) || network.ffs[ff] != id {
				return ErrInvariant
			}
			ff++
		}
	}
	if ci != len(network.cis) || co != len(network.cos) || ff != len(network.ffs) || fanins > network.limits.MaxFanins {
		return ErrInvariant
	}
	return nil
}

func (network *WlnNetwork) References() ([]int, error) {
	if err := network.Check(); err != nil {
		return nil, err
	}
	refs := make([]int, len(network.objects))
	for _, object := range network.objects[1:] {
		for _, fanin := range object.Fanins {
			if fanin != 0 {
				refs[fanin]++
			}
		}
	}
	return refs, nil
}

func (network *WlnNetwork) Fanouts() ([][]int, error) {
	if err := network.Check(); err != nil {
		return nil, err
	}
	result := make([][]int, len(network.objects))
	for id, object := range network.objects[1:] {
		for _, fanin := range object.Fanins {
			if fanin != 0 {
				result[fanin] = append(result[fanin], id+1)
			}
		}
	}
	return result, nil
}

func (network *WlnNetwork) IsAcyclic() (bool, error) {
	if err := network.Check(); err != nil {
		return false, err
	}
	state := make([]uint8, len(network.objects))
	for root := 1; root < len(network.objects); root++ {
		if state[root] != 0 {
			continue
		}
		type frame struct{ id, next int }
		stack := []frame{{id: root}}
		for len(stack) != 0 {
			if network.limits.Stop != nil && network.limits.Stop() {
				return false, ErrCancelled
			}
			top := &stack[len(stack)-1]
			if state[top.id] == 0 {
				state[top.id] = 1
			}
			object := network.objects[top.id]
			if object.Type == WlnCI || object.Type == WlnDFFRSE || top.next == len(object.Fanins) {
				state[top.id] = 2
				stack = stack[:len(stack)-1]
				continue
			}
			fanin := object.Fanins[top.next]
			top.next++
			if fanin == 0 {
				continue
			}
			if state[fanin] == 1 {
				return false, nil
			}
			if state[fanin] == 0 {
				stack = append(stack, frame{id: fanin})
			}
		}
	}
	return true, nil
}

func (network *WlnNetwork) DuplicateDFS() (*WlnNetwork, []int, error) {
	acyclic, err := network.IsAcyclic()
	if err != nil {
		return nil, nil, err
	}
	if !acyclic {
		return nil, nil, ErrInvariant
	}
	result, _ := NewWlnNetwork(network.Name, network.limits)
	result.Spec, result.SMTLib, result.UseInstanceDelays = network.Spec, network.SMTLib, network.UseInstanceDelays
	mapping := make([]int, len(network.objects))
	var duplicate func(int) (int, error)
	duplicate = func(id int) (int, error) {
		if id == 0 || mapping[id] != 0 {
			return mapping[id], nil
		}
		object := cloneWlnObject(network.objects[id])
		if object.Type != WlnDFFRSE {
			for index, fanin := range object.Fanins {
				mapped, duplicateErr := duplicate(fanin)
				if duplicateErr != nil {
					return 0, duplicateErr
				}
				object.Fanins[index] = mapped
			}
		}
		newID, addErr := result.AddObject(object)
		if addErr != nil {
			return 0, addErr
		}
		mapping[id] = newID
		return newID, nil
	}
	for _, id := range network.cis {
		if _, err = duplicate(id); err != nil {
			return nil, nil, err
		}
	}
	for _, id := range network.ffs {
		object := cloneWlnObject(network.objects[id])
		object.Fanins = nil
		mapping[id], err = result.AddObject(object)
		if err != nil {
			return nil, nil, err
		}
	}
	for _, id := range network.cos {
		if _, err = duplicate(id); err != nil {
			return nil, nil, err
		}
	}
	for _, id := range network.ffs {
		for _, fanin := range network.objects[id].Fanins {
			mapped, duplicateErr := duplicate(fanin)
			if duplicateErr != nil {
				return nil, nil, duplicateErr
			}
			result.objects[mapping[id]].Fanins = append(result.objects[mapping[id]].Fanins, mapped)
		}
	}
	return result, mapping, nil
}

type WlnNDRObject struct {
	ID                       int
	Type                     WlnOperation
	Signed                   bool
	Range                    WlnRange
	Fanins                   []int
	Table                    []uint64
	Name, Instance, Constant string
	NameID                   int
	Parameter                int
}
type WlnNDRDesign struct {
	Name, Spec        string
	SMTLib            bool
	UseInstanceDelays bool
	Objects           []WlnNDRObject
}

func (network *WlnNetwork) ToNDR() (*WlnNDRDesign, error) {
	if err := network.Check(); err != nil {
		return nil, err
	}
	design := &WlnNDRDesign{Name: network.Name, Spec: network.Spec, SMTLib: network.SMTLib, UseInstanceDelays: network.UseInstanceDelays, Objects: make([]WlnNDRObject, 0, len(network.objects)-1)}
	order := make([]int, 0, len(network.objects)-1)
	order = append(order, network.cis...)
	for id := 1; id < len(network.objects); id++ {
		if network.objects[id].Type != WlnCI && network.objects[id].Type != WlnCO {
			order = append(order, id)
		}
	}
	order = append(order, network.cos...)
	for _, id := range order {
		object := network.objects[id]
		design.Objects = append(design.Objects, WlnNDRObject{ID: id, Type: object.Type, Signed: object.Signed, Range: object.Range, Fanins: slices.Clone(object.Fanins), Table: slices.Clone(object.Table), Name: object.Name, NameID: object.NameID, Instance: strconv.Itoa(object.InstanceID), Constant: object.Constant, Parameter: object.Parameter})
	}
	return design, nil
}

func WlnNetworkFromNDR(design *WlnNDRDesign, limits WlnLimits) (*WlnNetwork, error) {
	if design == nil {
		return nil, ErrDimension
	}
	network, err := NewWlnNetwork(design.Name, limits)
	if err != nil {
		return nil, err
	}
	network.UseInstanceDelays = design.UseInstanceDelays
	network.Spec, network.SMTLib = design.Spec, design.SMTLib
	mapping := make(map[int]int, len(design.Objects))
	rawFanins := make([][]int, len(design.Objects))
	for index, record := range design.Objects {
		if record.ID <= 0 {
			record.ID = index + 1
		}
		if _, exists := mapping[record.ID]; exists {
			return nil, ErrInvariant
		}
		instance, parseErr := strconv.Atoi(record.Instance)
		if record.Instance == "" {
			instance, parseErr = 0, nil
		}
		if parseErr != nil {
			return nil, ErrDimension
		}
		var id int
		id, err = network.AddObject(WlnObject{Type: record.Type, Signed: record.Signed, Range: record.Range, Name: record.Name, NameID: record.NameID, InstanceID: instance, Constant: record.Constant, Table: slices.Clone(record.Table), Parameter: record.Parameter})
		if err != nil {
			return nil, err
		}
		mapping[record.ID] = id
		rawFanins[index] = slices.Clone(record.Fanins)
	}
	for index, record := range design.Objects {
		id := mapping[record.ID]
		if record.ID <= 0 {
			id = mapping[index+1]
		}
		for _, old := range rawFanins[index] {
			if old == 0 {
				network.objects[id].Fanins = append(network.objects[id].Fanins, 0)
				continue
			}
			mapped, exists := mapping[old]
			if !exists {
				return nil, ErrDimension
			}
			network.objects[id].Fanins = append(network.objects[id].Fanins, mapped)
		}
	}
	acyclic, err := network.IsAcyclic()
	if err != nil {
		return nil, err
	}
	if !acyclic {
		return nil, ErrInvariant
	}
	ordered, _, err := network.DuplicateDFS()
	if err != nil {
		return nil, err
	}
	return ordered, nil
}

func WlnConstantFromBits(words []uint64, width int) (string, error) {
	if width < 0 || len(words) < (width+63)/64 {
		return "", ErrDimension
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d'b", width)
	for bit := width - 1; bit >= 0; bit-- {
		builder.WriteByte('0' + byte(words[bit>>6]>>uint(bit&63)&1))
	}
	return builder.String(), nil
}
func WlnConstantFromString(value string, width int) (string, error) {
	if width < 0 || len(value) != width {
		return "", ErrDimension
	}
	for _, digit := range value {
		if !strings.ContainsRune("01xXzZ", digit) {
			return "", ErrDimension
		}
	}
	return strconv.Itoa(width) + "'b" + value, nil
}

func wlnMask(width int) uint64 {
	if width >= 64 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(width) - 1
}

func wlnSignedValue(value uint64, width int) int64 {
	value &= wlnMask(width)
	if width < 64 && value&(uint64(1)<<uint(width-1)) != 0 {
		value |= ^wlnMask(width)
	}
	return int64(value)
}

func (network *WlnNetwork) Evaluate(primaryInputs, state []uint64) ([]uint64, []uint64, error) {
	if err := network.Check(); err != nil {
		return nil, nil, err
	}
	if len(primaryInputs) != network.PICount() || len(state) != len(network.ffs) {
		return nil, nil, ErrDimension
	}
	for _, object := range network.objects[1:] {
		if object.Range.Width() > 64 {
			return nil, nil, ErrUnsupported
		}
	}
	values := make([]uint64, len(network.objects))
	pi := 0
	for index, id := range network.cis {
		if index < network.PICount() {
			values[id] = primaryInputs[pi]
			pi++
		} else {
			values[id] = state[index-network.PICount()]
		}
		values[id] &= wlnMask(network.objects[id].Range.Width())
	}
	for index, id := range network.ffs {
		values[id] = state[index] & wlnMask(network.objects[id].Range.Width())
	}
	for id := 1; id < len(network.objects); id++ {
		object := network.objects[id]
		if object.Type == WlnCI || object.Type == WlnDFFRSE {
			continue
		}
		fanin := func(index int) uint64 { return values[object.Fanins[index]] }
		width, value := object.Range.Width(), uint64(0)
		switch object.Type {
		case WlnCO, WlnBitBuf, WlnZeroPad, WlnSignExtend:
			value = fanin(0)
		case WlnConst:
			parsed, ok := parseWlnConstant(object.Constant)
			if !ok {
				return nil, nil, ErrUnsupported
			}
			value = parsed
		case WlnBitInv:
			value = ^fanin(0)
		case WlnBitAnd:
			value = fanin(0) & fanin(1)
		case WlnBitNand:
			value = ^(fanin(0) & fanin(1))
		case WlnBitOr:
			value = fanin(0) | fanin(1)
		case WlnBitNor:
			value = ^(fanin(0) | fanin(1))
		case WlnBitXor:
			value = fanin(0) ^ fanin(1)
		case WlnBitNxor:
			value = ^(fanin(0) ^ fanin(1))
		case WlnBitMux:
			if fanin(0)&1 != 0 {
				value = fanin(1)
			} else {
				value = fanin(2)
			}
		case WlnBitMaj:
			value = fanin(0)&fanin(1) | fanin(0)&fanin(2) | fanin(1)&fanin(2)
		case WlnReduceAnd:
			if fanin(0)&wlnMask(network.objects[object.Fanins[0]].Range.Width()) == wlnMask(network.objects[object.Fanins[0]].Range.Width()) {
				value = 1
			}
		case WlnReduceNand:
			if fanin(0)&wlnMask(network.objects[object.Fanins[0]].Range.Width()) != wlnMask(network.objects[object.Fanins[0]].Range.Width()) {
				value = 1
			}
		case WlnReduceOr:
			if fanin(0) != 0 {
				value = 1
			}
		case WlnReduceNor:
			if fanin(0) == 0 {
				value = 1
			}
		case WlnReduceXor:
			value = uint64(bits.OnesCount64(fanin(0)) & 1)
		case WlnReduceNxor:
			value = uint64(1 - bits.OnesCount64(fanin(0))&1)
		case WlnLogicNot:
			if fanin(0) == 0 {
				value = 1
			}
		case WlnLogicAnd:
			if fanin(0) != 0 && fanin(1) != 0 {
				value = 1
			}
		case WlnLogicOr:
			if fanin(0) != 0 || fanin(1) != 0 {
				value = 1
			}
		case WlnLogicXor:
			if (fanin(0) != 0) != (fanin(1) != 0) {
				value = 1
			}
		case WlnLogicImplies:
			if fanin(0) == 0 || fanin(1) != 0 {
				value = 1
			}
		case WlnAdd:
			value = fanin(0) + fanin(1)
			if len(object.Fanins) == 3 {
				value += fanin(2)
			}
		case WlnSub:
			value = fanin(0) - fanin(1)
		case WlnMul, WlnSignedMul:
			value = fanin(0) * fanin(1)
		case WlnDiv:
			if fanin(1) == 0 {
				return nil, nil, ErrDimension
			}
			value = fanin(0) / fanin(1)
		case WlnMod, WlnRem:
			if fanin(1) == 0 {
				return nil, nil, ErrDimension
			}
			value = fanin(0) % fanin(1)
		case WlnNeg:
			value = uint64(-int64(fanin(0)))
		case WlnEqual:
			if fanin(0) == fanin(1) {
				value = 1
			}
		case WlnNotEqual:
			if fanin(0) != fanin(1) {
				value = 1
			}
		case WlnSignedLess:
			leftWidth := network.objects[object.Fanins[0]].Range.Width()
			rightWidth := network.objects[object.Fanins[1]].Range.Width()
			if wlnSignedValue(fanin(0), leftWidth) < wlnSignedValue(fanin(1), rightWidth) {
				value = 1
			}
		case WlnLess:
			if fanin(0) < fanin(1) {
				value = 1
			}
		case WlnLessEqual:
			if fanin(0) <= fanin(1) {
				value = 1
			}
		case WlnMore:
			if fanin(0) > fanin(1) {
				value = 1
			}
		case WlnMoreEqual:
			if fanin(0) >= fanin(1) {
				value = 1
			}
		case WlnShiftLeft, WlnArithShiftLeft:
			value = fanin(0) << uint(fanin(1))
		case WlnShiftRight:
			value = fanin(0) >> uint(fanin(1))
		case WlnArithShiftRight:
			faninWidth := network.objects[object.Fanins[0]].Range.Width()
			value = uint64(wlnSignedValue(fanin(0), faninWidth) >> uint(fanin(1)))
		case WlnRotateLeft:
			if object.Parameter <= 0 || object.Parameter >= width {
				return nil, nil, ErrDimension
			}
			shift := uint(object.Parameter)
			value = (fanin(0) << shift) | (fanin(0) >> uint(width-int(shift)))
		case WlnRotateRight:
			if object.Parameter <= 0 || object.Parameter >= width {
				return nil, nil, ErrDimension
			}
			shift := uint(object.Parameter)
			value = (fanin(0) >> shift) | (fanin(0) << uint(width-int(shift)))
		case WlnSlice:
			low := object.Range.Begin
			if low > object.Range.End {
				low = object.Range.End
			}
			value = fanin(0) >> uint(low)
		case WlnConcat:
			for _, source := range object.Fanins {
				value = value<<uint(network.objects[source].Range.Width()) | values[source]
			}
		default:
			return nil, nil, ErrUnsupported
		}
		values[id] = value & wlnMask(width)
	}
	outputs := make([]uint64, network.POCount())
	for index := range outputs {
		outputs[index] = values[network.cos[index]]
	}
	next := make([]uint64, len(network.ffs))
	for index, id := range network.ffs {
		if len(network.objects[id].Fanins) == 0 {
			next[index] = values[id]
		} else {
			next[index] = values[network.objects[id].Fanins[0]] & wlnMask(network.objects[id].Range.Width())
		}
	}
	return outputs, next, nil
}

func parseWlnConstant(text string) (uint64, bool) {
	parts := strings.Split(text, "'b")
	if len(parts) != 2 || len(parts[1]) > 64 {
		return 0, false
	}
	value, err := strconv.ParseUint(parts[1], 2, 64)
	return value, err == nil
}

type WlnRetimeMove struct {
	Delay   int
	Forward bool
	Objects []int
}
type WlnRetimeResult struct {
	Moves        []WlnRetimeMove
	Instructions []int
	InitialDelay int
	BestDelay    int
	MoveCount    int
}

type WlnRetimeOptions struct {
	SkipSimple             bool
	MaxMoves, MaxEdgeLinks int
	Stop                   func() bool
}

func (network *WlnNetwork) RetimeWithOptions(delays []int, options WlnRetimeOptions) (WlnRetimeResult, error) {
	if network == nil || options.MaxMoves < 0 || options.MaxEdgeLinks < 0 {
		return WlnRetimeResult{}, ErrDimension
	}
	clone := *network
	originalStop := network.limits.Stop
	clone.limits = network.limits
	clone.limits.Stop = func() bool { return (originalStop != nil && originalStop()) || (options.Stop != nil && options.Stop()) }
	result, err := clone.Retime(delays, options.MaxMoves)
	if err != nil {
		return WlnRetimeResult{}, err
	}
	links := 2
	for id := 1; id < len(network.objects); id++ {
		for _, fanin := range network.objects[id].Fanins {
			if fanin != 0 && network.objects[fanin].Type == WlnDFFRSE {
				links += 2
			}
		}
	}
	if options.MaxEdgeLinks > 0 && links > options.MaxEdgeLinks {
		return WlnRetimeResult{}, ErrLimit
	}
	fanouts, _ := network.Fanouts()
	result.Instructions = []int{result.InitialDelay, 0}
	for _, move := range result.Moves {
		result.Instructions = append(result.Instructions, move.Delay)
		for _, id := range move.Objects {
			object := network.objects[id]
			if move.Forward {
				links += 2 * len(fanouts[id])
			} else {
				effective := 0
				for k, fanin := range object.Fanins {
					if fanin != 0 && !(object.Type == WlnDFFRSE && k > 0) {
						effective++
					}
				}
				links += 2 * effective
			}
			if options.MaxEdgeLinks > 0 && links > options.MaxEdgeLinks {
				return WlnRetimeResult{}, ErrLimit
			}
			if options.SkipSimple && (object.Type == WlnDFFRSE || object.Type == WlnSlice || object.Type == WlnConcat) {
				continue
			}
			if object.NameID <= 0 {
				return WlnRetimeResult{}, ErrInvariant
			}
			nameID := object.NameID
			if move.Forward {
				nameID = -nameID
			}
			result.Instructions = append(result.Instructions, nameID)
		}
		result.Instructions = append(result.Instructions, 0)
	}
	result.MoveCount = len(result.Moves)
	return result, nil
}

func (network *WlnNetwork) DefaultDelays() ([]int, error) {
	if err := network.Check(); err != nil {
		return nil, err
	}
	result := make([]int, len(network.objects))
	if network.UseInstanceDelays {
		for id := 1; id < len(result); id++ {
			result[id] = network.objects[id].InstanceID
		}
		return result, nil
	}
	for id := 1; id < len(result); id++ {
		object := network.objects[id]
		if object.Type == WlnDFFRSE || object.Type == WlnSlice || object.Type == WlnConcat {
			result[id] = 1
		} else if object.Type != WlnCI && object.Type != WlnCO && len(object.Fanins) > 0 {
			result[id] = 10
		}
	}
	for _, co := range network.cos {
		if len(network.objects[co].Fanins) > 0 {
			driver := network.objects[co].Fanins[0]
			if network.objects[driver].Type != WlnLUT {
				result[driver] = 1
			}
		}
	}
	return result, nil
}

// Retime returns deterministic delay-reducing moves on the source network's
// edge-register abstraction. It does not rewrite object storage; moves are
// recorded using object IDs in source order.
func (network *WlnNetwork) Retime(delays []int, maxMoves int) (WlnRetimeResult, error) {
	if err := network.Check(); err != nil {
		return WlnRetimeResult{}, err
	}
	if maxMoves < 0 || len(delays) != len(network.objects) {
		return WlnRetimeResult{}, ErrDimension
	}
	if maxMoves == 0 {
		maxMoves = network.limits.MaxSteps
	}
	for _, delay := range delays {
		if delay < 0 {
			return WlnRetimeResult{}, ErrDimension
		}
	}
	classes := make([]int, len(network.objects))
	classKeys := make(map[[6]int]int)
	for _, id := range network.ffs {
		if len(network.objects[id].Fanins) < 7 {
			return WlnRetimeResult{}, ErrInvariant
		}
		var key [6]int
		copy(key[:], network.objects[id].Fanins[1:7])
		class, found := classKeys[key]
		if !found {
			class = len(classKeys) + 1
			classKeys[key] = class
		}
		classes[id] = class
	}
	registers := make([][][]int, len(network.objects))
	fanouts, _ := network.Fanouts()
	for id := 1; id < len(network.objects); id++ {
		registers[id] = make([][]int, len(network.objects[id].Fanins))
		for index, fanin := range network.objects[id].Fanins {
			if fanin != 0 && network.objects[fanin].Type == WlnDFFRSE {
				registers[id][index] = []int{classes[fanin]}
			}
		}
	}
	compute := func() (int, []int, []int, error) {
		distance := make([]int, len(network.objects))
		indegree := make([]int, len(network.objects))
		for id := 1; id < len(network.objects); id++ {
			for k, f := range network.objects[id].Fanins {
				if f != 0 && len(registers[id][k]) == 0 && (network.objects[id].Type != WlnDFFRSE || k == 0) {
					indegree[id]++
				}
			}
		}
		queue := []int{}
		for id := 1; id < len(indegree); id++ {
			if indegree[id] == 0 {
				queue = append(queue, id)
			}
		}
		seen, maxDelay := 0, 0
		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]
			seen++
			if network.objects[id].Type != WlnCI {
				distance[id] += delays[id]
			}
			if distance[id] > maxDelay && !isWlnTerminal(network.objects[id].Type) {
				maxDelay = distance[id]
			}
			for _, out := range fanouts[id] {
				for k, f := range network.objects[out].Fanins {
					if f == id && len(registers[out][k]) == 0 && (network.objects[out].Type != WlnDFFRSE || k == 0) {
						if distance[out] < distance[id] {
							distance[out] = distance[id]
						}
						indegree[out]--
						if indegree[out] == 0 {
							queue = append(queue, out)
						}
					}
				}
			}
		}
		if seen < len(network.objects)-1 {
			return 0, nil, nil, ErrInvariant
		}
		critical := []int{}
		for id := 1; id < len(distance); id++ {
			if distance[id] == maxDelay && !isWlnTerminal(network.objects[id].Type) {
				critical = append(critical, id)
			}
		}
		return maxDelay, critical, distance, nil
	}
	findSources := func(sinks, distance []int) []int {
		visited := make([]bool, len(network.objects))
		result := make([]int, 0)
		type frame struct{ id, next int }
		for _, sink := range sinks {
			if visited[sink] || network.objects[sink].Type == WlnCI {
				continue
			}
			stack := []frame{{id: sink}}
			for len(stack) > 0 {
				top := &stack[len(stack)-1]
				id := top.id
				if !visited[id] {
					visited[id] = true
				}
				object := network.objects[id]
				faninDelay := distance[id] - delays[id]
				advanced := false
				for top.next < len(object.Fanins) {
					k := top.next
					top.next++
					if object.Type == WlnDFFRSE && k > 0 {
						continue
					}
					fanin := object.Fanins[k]
					if fanin == 0 || len(registers[id][k]) != 0 || distance[fanin] != faninDelay || visited[fanin] || network.objects[fanin].Type == WlnCI {
						continue
					}
					stack = append(stack, frame{id: fanin})
					advanced = true
					break
				}
				if advanced {
					continue
				}
				if faninDelay == 0 {
					result = append(result, id)
				}
				stack = stack[:len(stack)-1]
			}
		}
		return result
	}
	delay, critical, distance, err := compute()
	if err != nil {
		return WlnRetimeResult{}, err
	}
	result := WlnRetimeResult{InitialDelay: delay, BestDelay: delay}
	previousForward, previousBackward := false, false
	nonImproving, improvementChanges := 0, 0
	for len(result.Moves) < maxMoves && len(critical) > 0 {
		if network.limits.Stop != nil && network.limits.Stop() {
			return WlnRetimeResult{}, ErrCancelled
		}
		sources := findSources(critical, distance)
		slices.Sort(sources)
		sinks := slices.Clone(critical)
		slices.Sort(sinks)
		forwardClass, forward := -1, len(sources) != 0
		for _, id := range sources {
			local := -1
			for k, chain := range registers[id] {
				if network.objects[id].Type == WlnDFFRSE && k > 0 {
					continue
				}
				if network.objects[id].Fanins[k] == 0 {
					continue
				}
				if len(chain) == 0 {
					forward = false
					break
				}
				if local < 0 {
					local = chain[0]
				} else if local != chain[0] {
					forward = false
					break
				}
			}
			if !forward || local < 0 {
				forward = false
				break
			}
			if forwardClass < 0 {
				forwardClass = local
			} else if forwardClass != local {
				forward = false
				break
			}
		}
		backwardClass, backward := -1, len(sinks) != 0
		for _, id := range sinks {
			local := -1
			if len(fanouts[id]) == 0 {
				backward = false
				break
			}
			for _, out := range fanouts[id] {
				for k, fanin := range network.objects[out].Fanins {
					if fanin != id {
						continue
					}
					chain := registers[out][k]
					if len(chain) == 0 {
						backward = false
						break
					}
					tail := chain[len(chain)-1]
					if local < 0 {
						local = tail
					} else if local != tail {
						backward = false
						break
					}
				}
				if !backward {
					break
				}
			}
			if !backward || local < 0 {
				backward = false
				break
			}
			if backwardClass < 0 {
				backwardClass = local
			} else if backwardClass != local {
				backward = false
				break
			}
		}
		if !forward && !backward {
			break
		}
		common := false
		i, j := 0, 0
		for i < len(sources) && j < len(sinks) {
			if sources[i] == sinks[j] {
				common = true
				break
			}
			if sources[i] < sinks[j] {
				i++
			} else {
				j++
			}
		}
		if common {
			break
		}
		chooseForward := (previousForward && forward) || (!(previousBackward && backward) && ((forward && !backward) || (forward && backward && len(sources) < len(sinks))))
		frontier, class := sinks, backwardClass
		if chooseForward {
			frontier, class = sources, forwardClass
		}
		for _, candidate := range frontier {
			if chooseForward {
				for k := range registers[candidate] {
					if network.objects[candidate].Fanins[k] != 0 && !(network.objects[candidate].Type == WlnDFFRSE && k > 0) {
						registers[candidate][k] = registers[candidate][k][1:]
					}
				}
				for _, out := range fanouts[candidate] {
					for k, fanin := range network.objects[out].Fanins {
						if fanin == candidate {
							registers[out][k] = append(registers[out][k], class)
						}
					}
				}
			} else {
				for _, out := range fanouts[candidate] {
					for k, fanin := range network.objects[out].Fanins {
						if fanin == candidate {
							registers[out][k] = registers[out][k][:len(registers[out][k])-1]
						}
					}
				}
				for k := range registers[candidate] {
					if network.objects[candidate].Fanins[k] != 0 && !(network.objects[candidate].Type == WlnDFFRSE && k > 0) {
						registers[candidate][k] = append([]int{class}, registers[candidate][k]...)
					}
				}
			}
		}
		newDelay, next, nextDistance, computeErr := compute()
		if computeErr != nil {
			return WlnRetimeResult{}, computeErr
		}
		result.Moves = append(result.Moves, WlnRetimeMove{Delay: newDelay, Forward: chooseForward, Objects: slices.Clone(frontier)})
		if newDelay < result.BestDelay {
			result.BestDelay = newDelay
		}
		if chooseForward {
			previousForward = true
		} else {
			previousBackward = true
		}
		if newDelay >= delay {
			nonImproving++
		} else {
			if nonImproving > 0 {
				improvementChanges++
			}
			nonImproving = 0
		}
		if nonImproving > 3 || improvementChanges > 5 {
			break
		}
		delay, critical, distance = newDelay, next, nextDistance
	}
	return result, nil
}
func isWlnTerminal(operation WlnOperation) bool {
	return operation == WlnCI || operation == WlnCO
}

func (network *WlnNetwork) Verilog() (string, error) {
	if err := network.Check(); err != nil {
		return "", err
	}
	var b strings.Builder
	for id := 1; id < len(network.objects); id++ {
		o := network.objects[id]
		if o.Type != WlnTable {
			continue
		}
		if len(o.Fanins) != 1 || o.Range.Width() <= 0 || o.Range.Width() > 64 || 64%o.Range.Width() != 0 {
			return "", ErrDimension
		}
		inputWidth := network.objects[o.Fanins[0]].Range.Width()
		if inputWidth < 0 || inputWidth > 20 {
			return "", ErrLimit
		}
		entries := 1 << uint(inputWidth)
		needed := (entries*o.Range.Width() + 63) / 64
		if len(o.Table) < needed {
			return "", ErrDimension
		}
		fmt.Fprintf(&b, "module table%d(ind, val);\n  input [%d:0] ind;\n  output [%d:0] val;\n  reg [%d:0] val;\n  always @(ind)\n  begin\n    case (ind)\n", o.Parameter, inputWidth-1, o.Range.Width()-1, o.Range.Width()-1)
		mask := wlnMask(o.Range.Width())
		for entry := 0; entry < entries; entry++ {
			offset := entry * o.Range.Width()
			value := (o.Table[offset>>6] >> uint(offset&63)) & mask
			fmt.Fprintf(&b, "      %d'h%x: val = %d'h%x;\n", inputWidth, entry, o.Range.Width(), value)
		}
		b.WriteString("    endcase\n  end\nendmodule\n\n")
	}
	name := network.Name
	if name == "" {
		name = "top"
	}
	fmt.Fprintf(&b, "module %s (", name)
	ports := append(slices.Clone(network.cis), network.cos...)
	for i, id := range ports {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(network.objects[id].Name)
	}
	b.WriteString(");\n")
	for id := 1; id < len(network.objects); id++ {
		o := network.objects[id]
		kind := "wire"
		if o.Type == WlnCI {
			kind = "input wire"
		} else if o.Type == WlnCO {
			kind = "output wire"
		}
		signed := ""
		if o.Signed && !network.SMTLib {
			signed = " signed"
		}
		if o.Type == WlnSelectNMux || o.Type == WlnSelect {
			kind = "reg"
		}
		fmt.Fprintf(&b, "  %s%s [%d:%d] %s", kind, signed, o.Range.End, o.Range.Begin, o.Name)
		if o.Type == WlnCI || o.Type == WlnDFFRSE {
			b.WriteString(";\n")
			continue
		}
		if o.Type == WlnTable {
			fmt.Fprintf(&b, "; table%d table%d_Index(%s, %s);\n", o.Parameter, id, network.objects[o.Fanins[0]].Name, o.Name)
			continue
		}
		if o.Type == WlnLUT {
			b.WriteString("; LUT lut" + strconv.Itoa(id) + " (")
			for _, fanin := range o.Fanins {
				b.WriteString(network.objects[fanin].Name + ", ")
			}
			b.WriteString(o.Name + ");\n")
			continue
		}
		if o.Type == WlnRAMRead || o.Type == WlnRAMWrite {
			b.WriteString("; ")
			module := "ABC_READ"
			if o.Type == WlnRAMWrite {
				module = "ABC_WRITE"
			}
			fmt.Fprintf(&b, "%s (", module)
			ports := []string{"mem_in", "addr", "data"}
			for k, fanin := range o.Fanins {
				fmt.Fprintf(&b, ".%s(%s), ", ports[k], network.objects[fanin].Name)
			}
			outPort := "data"
			if o.Type == WlnRAMWrite {
				outPort = "mem_out"
			}
			fmt.Fprintf(&b, ".%s(%s));\n", outPort, o.Name)
			continue
		}
		if o.Type == WlnSelectNMux || o.Type == WlnSelect {
			b.WriteString(";\n  always @(")
			for k, fanin := range o.Fanins {
				if k > 0 {
					b.WriteString(" or ")
				}
				b.WriteString(network.objects[fanin].Name)
			}
			fmt.Fprintf(&b, ") begin\n    case (%s)\n", network.objects[o.Fanins[0]].Name)
			for k := 1; k < len(o.Fanins); k++ {
				if o.Type == WlnSelectNMux {
					fmt.Fprintf(&b, "      %d: %s = %s;\n", k-1, o.Name, network.objects[o.Fanins[k]].Name)
				} else {
					fmt.Fprintf(&b, "      %d'b", len(o.Fanins)-1)
					for j := len(o.Fanins) - 1; j > 0; j-- {
						if j == k {
							b.WriteByte('1')
						} else {
							b.WriteByte('0')
						}
					}
					fmt.Fprintf(&b, ": %s = %s;\n", o.Name, network.objects[o.Fanins[k]].Name)
				}
			}
			if o.Type == WlnSelect {
				fmt.Fprintf(&b, "      default: %s = %d'b0;\n", o.Name, o.Range.Width())
			}
			b.WriteString("    endcase\n  end\n")
			continue
		}
		if o.Type == WlnDecode {
			if len(o.Fanins) != 1 {
				return "", ErrInvariant
			}
			fmt.Fprintf(&b, " = ({{%d{1'b0}}} | 1'b1) << %s;\n", o.Range.Width()-1, network.objects[o.Fanins[0]].Name)
			continue
		}
		if o.Type == WlnAddSub {
			if len(o.Fanins) != 4 {
				return "", ErrInvariant
			}
			fmt.Fprintf(&b, " = %s ? %s + %s + %s : %s - %s - %s;\n", network.objects[o.Fanins[0]].Name, network.objects[o.Fanins[2]].Name, network.objects[o.Fanins[3]].Name, network.objects[o.Fanins[1]].Name, network.objects[o.Fanins[2]].Name, network.objects[o.Fanins[3]].Name, network.objects[o.Fanins[1]].Name)
			continue
		}
		expression, exprErr := network.verilogExpression(id)
		if exprErr != nil {
			return "", exprErr
		}
		comment := ""
		if network.SMTLib && o.Signed {
			comment = " // signed SMT-LIB operator"
		}
		fmt.Fprintf(&b, " = %s;%s\n", expression, comment)
	}
	if len(network.ffs) > 0 {
		b.WriteByte('\n')
	}
	for _, id := range network.ffs {
		o := network.objects[id]
		names := [8]string{"d", "clk", "reset", "set", "enable", "async", "sre", "init"}
		fmt.Fprintf(&b, "  ABC_DFFRSE (")
		first := true
		for k, f := range o.Fanins {
			if f == 0 {
				continue
			}
			if !first {
				b.WriteString(", ")
			}
			first = false
			fmt.Fprintf(&b, ".%s(%s)", names[k], network.objects[f].Name)
		}
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, ".q(%s));\n", o.Name)
	}
	b.WriteString("endmodule\n")
	return b.String(), nil
}

func (network *WlnNetwork) verilogExpression(id int) (string, error) {
	o := network.objects[id]
	f := func(k int) string { return network.objects[o.Fanins[k]].Name }
	binary := map[WlnOperation]string{WlnBitAnd: "&", WlnBitNand: "~&", WlnBitOr: "|", WlnBitNor: "~|", WlnBitXor: "^", WlnBitNxor: "~^", WlnLogicAnd: "&&", WlnLogicOr: "||", WlnLogicXor: "^^", WlnLogicImplies: "=>", WlnAdd: "+", WlnSub: "-", WlnMul: "*", WlnSignedMul: "*", WlnDiv: "/", WlnMod: "%", WlnRem: "%", WlnPow: "**", WlnSqrt: "@", WlnSquare: "#", WlnEqual: "==", WlnNotEqual: "!=", WlnLess: "<", WlnLessEqual: "<=", WlnMore: ">", WlnMoreEqual: ">=", WlnShiftLeft: "<<", WlnShiftRight: ">>", WlnArithShiftLeft: "<<<", WlnArithShiftRight: ">>>"}
	if op, ok := binary[o.Type]; ok {
		if len(o.Fanins) < 2 {
			return "", ErrInvariant
		}
		result := f(0) + " " + op + " " + f(1)
		if o.Type == WlnAdd && len(o.Fanins) == 3 {
			result += " + " + f(2)
		}
		return result, nil
	}
	switch o.Type {
	case WlnCO, WlnBitBuf:
		return f(0), nil
	case WlnConst:
		return o.Constant, nil
	case WlnBitInv:
		return "~" + f(0), nil
	case WlnLogicNot:
		return "!" + f(0), nil
	case WlnNeg:
		return "-" + f(0), nil
	case WlnReduceAnd:
		return "&" + f(0), nil
	case WlnReduceNand:
		return "~&" + f(0), nil
	case WlnReduceOr:
		return "|" + f(0), nil
	case WlnReduceNor:
		return "~|" + f(0), nil
	case WlnReduceXor:
		return "^" + f(0), nil
	case WlnReduceNxor:
		return "~^" + f(0), nil
	case WlnBitMux:
		return f(0) + " ? " + f(1) + " : " + f(2), nil
	case WlnSlice:
		return fmt.Sprintf("%s [%d:%d]", f(0), o.Range.End, o.Range.Begin), nil
	case WlnConcat:
		parts := make([]string, len(o.Fanins))
		for i := range parts {
			parts[i] = f(i)
		}
		return "{ " + strings.Join(parts, ", ") + " }", nil
	case WlnZeroPad:
		return fmt.Sprintf("{ {%d{1'b0}}, %s }", o.Range.Width()-network.objects[o.Fanins[0]].Range.Width(), f(0)), nil
	case WlnSignExtend:
		w := network.objects[o.Fanins[0]].Range.Width()
		return fmt.Sprintf("{ {%d{%s[%d]}}, %s }", o.Range.Width()-w, f(0), w-1, f(0)), nil
	case WlnRotateLeft, WlnRotateRight:
		shift := o.Parameter
		other := o.Range.Width() - shift
		if shift <= 0 || other <= 0 {
			return "", ErrDimension
		}
		if o.Type == WlnRotateLeft {
			return fmt.Sprintf("(%s << %d) | (%s >> %d)", f(0), shift, f(0), other), nil
		}
		return fmt.Sprintf("(%s >> %d) | (%s << %d)", f(0), shift, f(0), other), nil
	}
	return "", ErrUnsupported
}
