package abc

// Intermediate-function cache from src/bool/bdc/bdcTable.c.

import (
	"math"
	"slices"
)

// BdcFunctionID identifies an owned regular function in one BdcTable.
type BdcFunctionID uint32

// BdcLit is a function ID with a low complement bit.
type BdcLit uint32

const InvalidBdcLit = BdcLit(math.MaxUint32)

func (l BdcLit) Valid() bool        { return l <= math.MaxInt32 }
func (l BdcLit) ID() BdcFunctionID  { return BdcFunctionID(l >> 1) }
func (l BdcLit) Complemented() bool { return l&1 != 0 }
func (l BdcLit) Not() BdcLit        { return l ^ 1 }
func (l BdcLit) Regular() BdcLit    { return l &^ 1 }

// BdcISF is an incompletely specified function. On and Off must be disjoint;
// assignments in neither table are don't-cares. Unique is carried for the
// decomposition stages and must be a subset of Support.
type BdcISF struct {
	Support uint32
	Unique  uint32
	On      *TruthTable
	Off     *TruthTable
}

// BdcFunction is a detached snapshot. Truth never aliases table storage.
type BdcFunction struct {
	Support uint32
	Truth   *TruthTable
}

type BdcTableOptions struct {
	Variables    int
	MaxFunctions int
	Stop         func() bool
}

type bdcTableFunction struct {
	support uint32
	truth   *TruthTable
	next    int
}

// BdcTable owns the support-indexed cache used by bi-decomposition. Bucket
// chains are newest-first, matching Bdc_TableAdd and Bdc_TableLookup.
type BdcTable struct {
	variables    int
	maxFunctions int
	stop         func() bool
	heads        []int
	occupied     []uint32
	functions    []bdcTableFunction
}

func NewBdcTable(options BdcTableOptions) (*BdcTable, error) {
	if options.Variables < 0 || options.Variables >= 16 || options.MaxFunctions < 0 {
		return nil, ErrDimension
	}
	maximum := options.MaxFunctions
	if maximum == 0 {
		maximum = 512
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	tableSize := 1 << uint(options.Variables)
	heads := make([]int, tableSize)
	for index := range heads {
		heads[index] = -1
	}
	return &BdcTable{
		variables:    options.Variables,
		maxFunctions: maximum,
		stop:         options.Stop,
		heads:        heads,
		occupied:     make([]uint32, 0, min(tableSize, 256)),
		functions:    make([]bdcTableFunction, 0, min(maximum, 512)),
	}, nil
}

func (t *BdcTable) checkLive() error {
	if t == nil || t.variables < 0 || t.variables >= 16 || t.maxFunctions <= 0 || len(t.heads) != 1<<uint(t.variables) {
		return ErrDimension
	}
	if t.stop != nil && t.stop() {
		return ErrCancelled
	}
	return nil
}

func (t *BdcTable) validateTruth(truth *TruthTable) error {
	if err := t.checkLive(); err != nil {
		return err
	}
	if err := truth.Check(); err != nil {
		return err
	}
	if truth.Variables() != t.variables {
		return ErrDimension
	}
	return nil
}

func bdcTruthSupportMask(truth *TruthTable) (uint32, error) {
	support, err := truth.Support()
	if err != nil {
		return 0, err
	}
	var mask uint32
	for _, variable := range support {
		mask |= uint32(1) << uint(variable)
	}
	return mask, nil
}

// Add copies truth into the newest position of its support bucket. The first
// function after construction or Clear must be constant one at support zero,
// which preserves the native cache's distinguished node-zero contract.
func (t *BdcTable) Add(support uint32, truth *TruthTable) (BdcLit, error) {
	if err := t.validateTruth(truth); err != nil {
		return InvalidBdcLit, err
	}
	if support >= uint32(len(t.heads)) {
		return InvalidBdcLit, ErrDimension
	}
	actual, err := bdcTruthSupportMask(truth)
	if err != nil {
		return InvalidBdcLit, err
	}
	if actual != support {
		return InvalidBdcLit, ErrInvariant
	}
	if len(t.functions) == 0 {
		if support != 0 || !truth.IsConstant(true) {
			return InvalidBdcLit, ErrInvariant
		}
	} else if support == 0 {
		return InvalidBdcLit, ErrInvariant
	}
	if len(t.functions) >= t.maxFunctions || len(t.functions) > int(MaxLiteralID) {
		return InvalidBdcLit, ErrLimit
	}
	if t.stop != nil && t.stop() {
		return InvalidBdcLit, ErrCancelled
	}
	head := t.heads[support]
	if head < 0 {
		t.occupied = append(t.occupied, support)
	}
	id := BdcFunctionID(len(t.functions))
	t.functions = append(t.functions, bdcTableFunction{support: support, truth: truth.Clone(), next: head})
	t.heads[support] = int(id)
	return BdcLit(id << 1), nil
}

func validateBdcISF(variables int, isf BdcISF) error {
	if variables < 0 || variables >= 16 || isf.On == nil || isf.Off == nil ||
		isf.Support >= uint32(1<<uint(variables)) || isf.Unique&^isf.Support != 0 {
		return ErrDimension
	}
	if err := isf.On.Check(); err != nil {
		return err
	}
	if err := isf.Off.Check(); err != nil {
		return err
	}
	if isf.On.Variables() != variables || isf.Off.Variables() != variables {
		return ErrDimension
	}
	intersects, err := isf.On.Intersects(isf.Off, false)
	if err != nil {
		return err
	}
	if intersects {
		return ErrInvariant
	}
	return nil
}

// BdcTableCheckContainment reports whether truth includes the ISF on-set and
// excludes its off-set.
func BdcTableCheckContainment(isf BdcISF, truth *TruthTable) (bool, error) {
	if truth == nil {
		return false, ErrDimension
	}
	if err := truth.Check(); err != nil {
		return false, err
	}
	if err := validateBdcISF(truth.Variables(), isf); err != nil {
		return false, err
	}
	implied, err := isf.On.Implies(truth)
	if err != nil || !implied {
		return false, err
	}
	intersects, err := truth.Intersects(isf.Off, false)
	return !intersects, err
}

func (t *BdcTable) lookupPolarity(isf BdcISF) (BdcLit, bool, error) {
	for id := t.heads[isf.Support]; id >= 0; id = t.functions[id].next {
		contained, err := BdcTableCheckContainment(isf, t.functions[id].truth)
		if err != nil {
			return InvalidBdcLit, false, err
		}
		if contained {
			return BdcLit(uint32(id) << 1), true, nil
		}
	}
	return InvalidBdcLit, false, nil
}

// Lookup first searches direct functions newest-first and then searches the
// same bucket with on/off exchanged, returning a complemented literal.
func (t *BdcTable) Lookup(isf BdcISF) (BdcLit, bool, error) {
	if err := t.checkLive(); err != nil {
		return InvalidBdcLit, false, err
	}
	if err := validateBdcISF(t.variables, isf); err != nil {
		return InvalidBdcLit, false, err
	}
	if len(t.functions) == 0 || t.heads[0] != 0 || !t.functions[0].truth.IsConstant(true) {
		return InvalidBdcLit, false, ErrInvariant
	}
	if isf.Support == 0 {
		if isf.On.IsConstant(true) {
			return BdcLit(0), true, nil
		}
		if !isf.Off.IsConstant(true) {
			return InvalidBdcLit, false, ErrInvariant
		}
		return BdcLit(1), true, nil
	}
	if literal, found, err := t.lookupPolarity(isf); found || err != nil {
		return literal, found, err
	}
	isf.On, isf.Off = isf.Off, isf.On
	literal, found, err := t.lookupPolarity(isf)
	if found {
		literal = literal.Not()
	}
	return literal, found, err
}

// Function returns the regular function named by literal. Complemented is
// represented by the literal and is not applied to the returned truth table.
func (t *BdcTable) Function(literal BdcLit) (BdcFunction, error) {
	if err := t.checkLive(); err != nil {
		return BdcFunction{}, err
	}
	if !literal.Valid() || int(literal.ID()) >= len(t.functions) {
		return BdcFunction{}, ErrInvalidLiteral
	}
	function := t.functions[literal.ID()]
	return BdcFunction{Support: function.support, Truth: function.truth.Clone()}, nil
}

func (t *BdcTable) Len() int {
	if t == nil {
		return 0
	}
	return len(t.functions)
}

func (t *BdcTable) OccupiedSupports() []uint32 {
	if t == nil {
		return nil
	}
	return slices.Clone(t.occupied)
}

// Clear resets only occupied buckets and releases all owned functions. The next
// successful Add must establish constant one as function zero again.
func (t *BdcTable) Clear() error {
	if err := t.checkLive(); err != nil {
		return err
	}
	for _, support := range t.occupied {
		t.heads[support] = -1
	}
	clear(t.functions)
	t.functions = t.functions[:0]
	t.occupied = t.occupied[:0]
	return nil
}
