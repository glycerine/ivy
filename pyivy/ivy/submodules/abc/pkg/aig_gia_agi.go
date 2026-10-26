package abc

// Compact alternate AIG experiment from src/aig/gia/giaAgi.c.

import "slices"

type AgiID uint32
type AgiLit uint32

const (
	InvalidAgiID  = AgiID(^uint32(0))
	InvalidAgiLit = AgiLit(^uint32(0))
	AgiConst0     = AgiLit(0)
	AgiConst1     = AgiLit(1)
)

func (literal AgiLit) Valid() bool        { return literal <= AgiLit(GiaNone<<1|1) }
func (literal AgiLit) ID() AgiID          { return AgiID(literal >> 1) }
func (literal AgiLit) Complemented() bool { return literal&1 != 0 }
func (literal AgiLit) Not() AgiLit        { return literal ^ 1 }
func (literal AgiLit) NotIf(value bool) AgiLit {
	if value {
		return literal.Not()
	}
	return literal
}

type AgiObjectKind uint8

const (
	AgiObjectConstant AgiObjectKind = iota
	AgiObjectCI
	AgiObjectAND
	AgiObjectCO
)

type AgiObject struct {
	Kind          AgiObjectKind
	First, Second AgiLit
}

type AgiManager struct {
	Name, Spec  string
	objects     []AgiObject
	cis, cos    []AgiID
	registers   int
	objectLimit int
}

type AgiSupportOptions struct {
	MaxObjects, MaxVisits int
	Stop                  func() bool
}

type AgiSupportReport struct {
	Nodes []AgiID
	Sizes []int
	Small int
}

type GiaAgiSupportComparison struct {
	Agi        *AgiManager
	AgiReport  AgiSupportReport
	GiaSizes   []int
	GiaSmall   int
	Equivalent bool
}

func NewAgiManager(maxObjects int) (*AgiManager, error) {
	if maxObjects < 1 {
		return nil, ErrDimension
	}
	if uint64(maxObjects) > uint64(MaxGiaID)+1 {
		return nil, ErrLimit
	}
	return &AgiManager{objects: []AgiObject{{Kind: AgiObjectConstant}}, objectLimit: maxObjects}, nil
}

func (manager *AgiManager) ObjectCount() int {
	if manager == nil {
		return 0
	}
	return len(manager.objects)
}
func (manager *AgiManager) InputCount() int {
	if manager == nil {
		return 0
	}
	return len(manager.cis)
}
func (manager *AgiManager) OutputCount() int {
	if manager == nil {
		return 0
	}
	return len(manager.cos)
}
func (manager *AgiManager) NodeCount() int {
	if manager == nil {
		return 0
	}
	return len(manager.objects) - len(manager.cis) - len(manager.cos) - 1
}
func (manager *AgiManager) RegisterCount() int {
	if manager == nil {
		return 0
	}
	return manager.registers
}
func (manager *AgiManager) Inputs() []AgiID {
	if manager == nil {
		return nil
	}
	return slices.Clone(manager.cis)
}
func (manager *AgiManager) Outputs() []AgiID {
	if manager == nil {
		return nil
	}
	return slices.Clone(manager.cos)
}
func (manager *AgiManager) Object(id AgiID) (AgiObject, error) {
	if manager == nil || uint64(id) >= uint64(len(manager.objects)) {
		return AgiObject{}, ErrInvalidLiteral
	}
	return manager.objects[id], nil
}

func agiOrdinal(ids []AgiID, id AgiID) int {
	for index, candidate := range ids {
		if candidate == id {
			return index
		}
	}
	return -1
}

func (manager *AgiManager) IsConstant(id AgiID) bool {
	return manager != nil && id == 0 && len(manager.objects) != 0
}
func (manager *AgiManager) IsCI(id AgiID) bool {
	return manager != nil && uint64(id) < uint64(len(manager.objects)) && manager.objects[id].Kind == AgiObjectCI
}
func (manager *AgiManager) IsCO(id AgiID) bool {
	return manager != nil && uint64(id) < uint64(len(manager.objects)) && manager.objects[id].Kind == AgiObjectCO
}
func (manager *AgiManager) IsAND(id AgiID) bool {
	return manager != nil && uint64(id) < uint64(len(manager.objects)) && manager.objects[id].Kind == AgiObjectAND
}
func (manager *AgiManager) IsPI(id AgiID) bool {
	if manager == nil {
		return false
	}
	index := agiOrdinal(manager.cis, id)
	return index >= 0 && index < len(manager.cis)-manager.registers
}
func (manager *AgiManager) IsRO(id AgiID) bool {
	if manager == nil {
		return false
	}
	index := agiOrdinal(manager.cis, id)
	return index >= len(manager.cis)-manager.registers && index >= 0
}
func (manager *AgiManager) IsPO(id AgiID) bool {
	if manager == nil {
		return false
	}
	index := agiOrdinal(manager.cos, id)
	return index >= 0 && index < len(manager.cos)-manager.registers
}
func (manager *AgiManager) IsRI(id AgiID) bool {
	if manager == nil {
		return false
	}
	index := agiOrdinal(manager.cos, id)
	return index >= len(manager.cos)-manager.registers && index >= 0
}

func (manager *AgiManager) validSignal(literal AgiLit) bool {
	return manager != nil && literal.Valid() && uint64(literal.ID()) < uint64(len(manager.objects)) && manager.objects[literal.ID()].Kind != AgiObjectCO
}

func (manager *AgiManager) appendObject(object AgiObject) (AgiLit, error) {
	if manager == nil || len(manager.objects) == 0 {
		return InvalidAgiLit, ErrInvariant
	}
	if len(manager.objects) >= manager.objectLimit {
		return InvalidAgiLit, ErrLimit
	}
	id := AgiID(len(manager.objects))
	manager.objects = append(manager.objects, object)
	return AgiLit(id) << 1, nil
}

func (manager *AgiManager) AppendCI() (AgiLit, error) {
	literal, err := manager.appendObject(AgiObject{Kind: AgiObjectCI})
	if err == nil {
		manager.cis = append(manager.cis, literal.ID())
	}
	return literal, err
}

func (manager *AgiManager) AppendAND(first, second AgiLit) (AgiLit, error) {
	if !manager.validSignal(first) || !manager.validSignal(second) || first.ID() == second.ID() {
		return InvalidAgiLit, ErrInvalidLiteral
	}
	if first > second {
		first, second = second, first
	}
	return manager.appendObject(AgiObject{Kind: AgiObjectAND, First: first, Second: second})
}

func (manager *AgiManager) AppendCO(driver AgiLit) (AgiLit, error) {
	if !manager.validSignal(driver) {
		return InvalidAgiLit, ErrInvalidLiteral
	}
	literal, err := manager.appendObject(AgiObject{Kind: AgiObjectCO, First: driver})
	if err == nil {
		manager.cos = append(manager.cos, literal.ID())
	}
	return literal, err
}

func (manager *AgiManager) SetRegisterCount(registers int) error {
	if manager == nil || registers < 0 || registers > len(manager.cis) || registers > len(manager.cos) {
		return ErrDimension
	}
	manager.registers = registers
	return nil
}

func (manager *AgiManager) Check() error {
	if manager == nil || len(manager.objects) == 0 || manager.objects[0].Kind != AgiObjectConstant || len(manager.objects) > manager.objectLimit || manager.registers < 0 || manager.registers > len(manager.cis) || manager.registers > len(manager.cos) {
		return ErrInvariant
	}
	ci, co := 0, 0
	for index, object := range manager.objects {
		id := AgiID(index)
		switch object.Kind {
		case AgiObjectConstant:
			if index != 0 {
				return ErrInvariant
			}
		case AgiObjectCI:
			if ci >= len(manager.cis) || manager.cis[ci] != id {
				return ErrInvariant
			}
			ci++
		case AgiObjectAND:
			if !object.First.Valid() || !object.Second.Valid() || object.First >= object.Second || object.Second.ID() >= id {
				return ErrInvariant
			}
		case AgiObjectCO:
			if co >= len(manager.cos) || manager.cos[co] != id || !object.First.Valid() || object.First.ID() >= id {
				return ErrInvariant
			}
			co++
		default:
			return ErrInvariant
		}
	}
	if ci != len(manager.cis) || co != len(manager.cos) {
		return ErrInvariant
	}
	return nil
}

// Evaluate computes all COs in source order; sequential CIs are ordinary
// caller-supplied inputs, matching one combinational transition relation.
func (manager *AgiManager) Evaluate(inputs []bool) ([]bool, error) {
	if manager == nil || len(inputs) != len(manager.cis) {
		return nil, ErrDimension
	}
	if err := manager.Check(); err != nil {
		return nil, err
	}
	values := make([]bool, len(manager.objects))
	for index, id := range manager.cis {
		values[id] = inputs[index]
	}
	read := func(literal AgiLit) bool { return values[literal.ID()] != literal.Complemented() }
	for index, object := range manager.objects {
		switch object.Kind {
		case AgiObjectAND:
			values[index] = read(object.First) && read(object.Second)
		case AgiObjectCO:
			values[index] = read(object.First)
		}
	}
	outputs := make([]bool, len(manager.cos))
	for index, id := range manager.cos {
		outputs[index] = values[id]
	}
	return outputs, nil
}

func checkAgiSupportOptions(options AgiSupportOptions) (AgiSupportOptions, error) {
	if options.MaxObjects < 0 || options.MaxVisits < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 26
	}
	return options, nil
}

// AgiFromGia converts an ordinary-AND GIA without borrowing Value or traversal
// scratch. Object IDs and terminal order are retained exactly.
func AgiFromGia(source *GiaManager, maxObjects int) (*AgiManager, []AgiLit, error) {
	if source == nil || maxObjects < 0 {
		return nil, nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := source.Check(); err != nil {
		return nil, nil, err
	}
	if source.ObjectCount() > maxObjects {
		return nil, nil, ErrLimit
	}
	if source.XorCount() != 0 || source.MuxCount() != 0 || source.BufferCount() != 0 {
		return nil, nil, ErrUnsupported
	}
	target, err := NewAgiManager(maxObjects)
	if err != nil {
		return nil, nil, err
	}
	target.Name, target.Spec = source.Name, source.Spec
	mapping := make([]AgiLit, source.ObjectCount())
	for i := range mapping {
		mapping[i] = InvalidAgiLit
	}
	mapping[0] = AgiConst0
	for index, object := range source.objects[1:] {
		id := GiaID(index + 1)
		switch {
		case object.IsCI():
			mapping[id], err = target.AppendCI()
		case object.IsAnd():
			first, firstErr := object.Fanin0(id)
			second, secondErr := object.Fanin1(id)
			if firstErr != nil {
				err = firstErr
			} else if secondErr != nil {
				err = secondErr
			} else {
				mapping[id], err = target.AppendAND(mapping[first.ID()].NotIf(first.Complemented()), mapping[second.ID()].NotIf(second.Complemented()))
			}
		case object.IsCO():
			first, firstErr := object.Fanin0(id)
			if firstErr != nil {
				err = firstErr
			} else {
				mapping[id], err = target.AppendCO(mapping[first.ID()].NotIf(first.Complemented()))
			}
		default:
			err = ErrInvariant
		}
		if err != nil {
			return nil, nil, err
		}
	}
	if err = target.SetRegisterCount(source.RegisterCount()); err != nil {
		return nil, nil, err
	}
	if err = target.Check(); err != nil {
		return nil, nil, err
	}
	return target, mapping, nil
}

func (manager *AgiManager) supportSize(root AgiID, options AgiSupportOptions, visits *int) (int, error) {
	if uint64(root) >= uint64(len(manager.objects)) {
		return 0, ErrInvalidLiteral
	}
	seen := make([]bool, len(manager.objects))
	stack := []AgiID{root}
	size := 0
	for len(stack) != 0 {
		if options.Stop != nil && options.Stop() {
			return 0, ErrCancelled
		}
		*visits++
		if *visits > options.MaxVisits {
			return 0, ErrLimit
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[id] {
			continue
		}
		seen[id] = true
		object := manager.objects[id]
		switch object.Kind {
		case AgiObjectConstant:
		case AgiObjectCI:
			size++
		case AgiObjectAND:
			stack = append(stack, object.Second.ID(), object.First.ID())
		case AgiObjectCO:
			stack = append(stack, object.First.ID())
		default:
			return 0, ErrInvariant
		}
	}
	return size, nil
}

func (manager *AgiManager) SupportSize(root AgiID, options AgiSupportOptions) (int, error) {
	if manager == nil {
		return 0, ErrDimension
	}
	options, err := checkAgiSupportOptions(options)
	if err != nil {
		return 0, err
	}
	if err = manager.Check(); err != nil {
		return 0, err
	}
	if manager.ObjectCount() > options.MaxObjects {
		return 0, ErrLimit
	}
	visits := 0
	return manager.supportSize(root, options, &visits)
}

func (manager *AgiManager) AnalyzeSupports(maximum int, options AgiSupportOptions) (AgiSupportReport, error) {
	if manager == nil || maximum < 0 {
		return AgiSupportReport{}, ErrDimension
	}
	options, err := checkAgiSupportOptions(options)
	if err != nil {
		return AgiSupportReport{}, err
	}
	if err = manager.Check(); err != nil {
		return AgiSupportReport{}, err
	}
	if manager.ObjectCount() > options.MaxObjects {
		return AgiSupportReport{}, ErrLimit
	}
	report := AgiSupportReport{}
	visits := 0
	for id, object := range manager.objects {
		if object.Kind != AgiObjectAND {
			continue
		}
		size, sizeErr := manager.supportSize(AgiID(id), options, &visits)
		if sizeErr != nil {
			return AgiSupportReport{}, sizeErr
		}
		report.Nodes = append(report.Nodes, AgiID(id))
		report.Sizes = append(report.Sizes, size)
		if size <= maximum {
			report.Small++
		}
	}
	return report, nil
}

// CompareGiaAgiSupports replaces the source's stdout-only experiment with an
// owned comparison against the existing GIA support-set implementation.
func CompareGiaAgiSupports(source *GiaManager, maximum int, options AgiSupportOptions) (GiaAgiSupportComparison, error) {
	if source == nil || maximum < 0 {
		return GiaAgiSupportComparison{}, ErrDimension
	}
	options, err := checkAgiSupportOptions(options)
	if err != nil {
		return GiaAgiSupportComparison{}, err
	}
	agi, _, err := AgiFromGia(source, options.MaxObjects)
	if err != nil {
		return GiaAgiSupportComparison{}, err
	}
	report, err := agi.AnalyzeSupports(maximum, options)
	if err != nil {
		return GiaAgiSupportComparison{}, err
	}
	nodes := make([]GiaID, len(report.Nodes))
	for i, id := range report.Nodes {
		nodes[i] = GiaID(id)
	}
	supports, err := source.NodeSupports(nodes, options.MaxObjects, options.MaxVisits)
	if err != nil {
		return GiaAgiSupportComparison{}, err
	}
	result := GiaAgiSupportComparison{Agi: agi, AgiReport: report, GiaSizes: make([]int, len(supports)), Equivalent: true}
	for i, support := range supports {
		result.GiaSizes[i] = len(support)
		if len(support) <= maximum {
			result.GiaSmall++
		}
		if result.GiaSizes[i] != report.Sizes[i] {
			result.Equivalent = false
		}
	}
	return result, nil
}

func (manager *AgiManager) Close() {
	if manager == nil {
		return
	}
	manager.objects = nil
	manager.cis = nil
	manager.cos = nil
	manager.registers = 0
}
