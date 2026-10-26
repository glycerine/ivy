package abc

// Truth-table bi-decomposition from src/bool/bdc/bdc.h, bdcInt.h,
// bdcCore.c, bdcDec.c, bdcTable.c, and bdcSpfd.c.

import (
	"encoding/binary"
	"math/bits"
	"slices"
	"strconv"
)

// BdcNodeType is the concrete node kind retained by the decomposition.
// OR and MUX choices are lowered to complemented two-input AND nodes, as in
// the source implementation.
type BdcNodeType uint8

const (
	BdcNodeNone BdcNodeType = iota
	BdcNodeConst1
	BdcNodePI
	BdcNodeAND
)

func bdcMakeLit(id int, complement bool) BdcLit {
	lit := BdcLit(uint32(id) << 1)
	if complement {
		return lit.Not()
	}
	return lit
}

func (l BdcLit) NotIf(condition bool) BdcLit {
	if condition {
		return l.Not()
	}
	return l
}

// BdcNode is an owned decomposition node. Truth returns a clone through the
// manager; callers cannot mutate the cached function.
type BdcNode struct {
	Type           BdcNodeType
	Fanin0, Fanin1 BdcLit
	Support        uint32
	CopyInt        int
	CopyToken      uint64
	truth          *TruthTable
}

type BdcParameters struct {
	MaxVariables int
	MaxWords     int
	MaxNodes     int
}

type BdcDecomposeOptions struct {
	MaxNewNodes int
	MaxDivisors int
	Stop        func() bool
}

type BdcStatistics struct {
	Calls, Nodes, Reused, Muxes, AndChoices, OrChoices, WeakChoices int
}

// BdcManager owns one decomposition at a time. Failed and cancelled calls are
// atomic and leave the previous successful result unchanged.
type BdcManager struct {
	parameters BdcParameters
	variables  int
	nodes      []BdcNode
	root       BdcLit
	newNodes   int
	statistics BdcStatistics
}

func NewBdcManager(parameters BdcParameters) (*BdcManager, error) {
	if parameters.MaxVariables < 2 || parameters.MaxVariables > 15 || parameters.MaxWords < 0 || parameters.MaxNodes < 0 {
		return nil, ErrDimension
	}
	if parameters.MaxWords == 0 {
		parameters.MaxWords = 1 << 16
	}
	if parameters.MaxNodes == 0 {
		parameters.MaxNodes = 512
	}
	return &BdcManager{parameters: parameters}, nil
}

func (m *BdcManager) Variables() int {
	if m == nil {
		return 0
	}
	return m.variables
}
func (m *BdcManager) NodeCount() int {
	if m == nil {
		return 0
	}
	return len(m.nodes)
}
func (m *BdcManager) AndCount() int {
	if m == nil {
		return 0
	}
	if len(m.nodes) <= m.variables+1 {
		return 0
	}
	return len(m.nodes) - m.variables - 1
}
func (m *BdcManager) GateCount() int {
	if m == nil {
		return 0
	}
	n := 0
	for i := range m.nodes {
		if m.nodes[i].Type == BdcNodeAND {
			n++
		}
	}
	return n
}
func (m *BdcManager) Root() (BdcLit, error) {
	if m == nil || len(m.nodes) == 0 {
		return InvalidBdcLit, ErrDimension
	}
	return m.root, nil
}
func (m *BdcManager) Statistics() BdcStatistics {
	if m == nil {
		return BdcStatistics{}
	}
	return m.statistics
}
func (m *BdcManager) Node(id int) (BdcNode, error) {
	if m == nil || id < 0 || id >= len(m.nodes) {
		return BdcNode{}, ErrDimension
	}
	n := m.nodes[id]
	n.truth = nil
	return n, nil
}
func (m *BdcManager) NodeTruth(id int) (*TruthTable, error) {
	if m == nil || id < 0 || id >= len(m.nodes) {
		return nil, ErrDimension
	}
	return m.nodes[id].truth.Clone(), nil
}
func (m *BdcManager) SetCopyInt(id, value int) error {
	if m == nil || id < 0 || id >= len(m.nodes) {
		return ErrDimension
	}
	m.nodes[id].CopyInt = value
	return nil
}
func (m *BdcManager) SetCopyToken(id int, value uint64) error {
	if m == nil || id < 0 || id >= len(m.nodes) {
		return ErrDimension
	}
	m.nodes[id].CopyToken = value
	return nil
}
func (m *BdcManager) ObjectCopyToken(lit BdcLit) (uint64, error) {
	id := int(lit.ID())
	if m == nil || !lit.Valid() || id >= len(m.nodes) {
		return 0, ErrDimension
	}
	return m.nodes[id].CopyToken, nil
}
func (m *BdcManager) ObjectCopy(lit BdcLit) (int, error) {
	id := int(lit.ID())
	if m == nil || !lit.Valid() || id >= len(m.nodes) {
		return 0, ErrDimension
	}
	value := m.nodes[id].CopyInt
	if lit.Complemented() {
		value ^= 1
	}
	return value, nil
}
func (m *BdcManager) FaninCopy(id, fanin int) (int, error) {
	if m == nil || id < 0 || id >= len(m.nodes) || fanin < 0 || fanin > 1 {
		return 0, ErrDimension
	}
	n := m.nodes[id]
	if n.Type != BdcNodeAND {
		return 0, ErrDimension
	}
	lit := n.Fanin0
	if fanin == 1 {
		lit = n.Fanin1
	}
	return m.ObjectCopy(lit)
}

type bdcBuild struct {
	manager          *BdcManager
	nodes            []BdcNode
	variables        int
	maxNew, newNodes int
	stop             func() bool
	table            *BdcTable
	andHash          map[[4]int]BdcLit
	statistics       BdcStatistics
}

type bdcISFState struct {
	on, off         *TruthTable
	support, unique uint32
}

func bdcSupportMask(t *TruthTable) (uint32, error) {
	s, e := t.Support()
	if e != nil {
		return 0, e
	}
	var mask uint32
	for _, v := range s {
		mask |= 1 << uint(v)
	}
	return mask, nil
}

func (b *bdcBuild) addBase(t BdcNodeType, truth *TruthTable) (BdcLit, error) {
	if len(b.nodes) >= b.manager.parameters.MaxNodes {
		return InvalidBdcLit, ErrLimit
	}
	s, e := bdcSupportMask(truth)
	if e != nil {
		return InvalidBdcLit, e
	}
	id := len(b.nodes)
	b.nodes = append(b.nodes, BdcNode{Type: t, Support: s, truth: truth.Clone()})
	lit, e := b.table.Add(s, truth)
	if e != nil {
		b.nodes = b.nodes[:id]
		return InvalidBdcLit, e
	}
	if int(lit.ID()) != id {
		b.nodes = b.nodes[:id]
		return InvalidBdcLit, ErrInvariant
	}
	return bdcMakeLit(id, false), nil
}

func (b *bdcBuild) litTruth(l BdcLit) *TruthTable {
	t := b.nodes[int(l.ID())].truth.Clone()
	if l.Complemented() {
		return t.Not()
	}
	return t
}

func bdcIntervalContains(on, off, candidate *TruthTable) bool {
	if on == nil || off == nil || candidate == nil {
		return false
	}
	if on.variables != candidate.variables || off.variables != candidate.variables {
		return false
	}
	for i, w := range candidate.words {
		if on.words[i]&^w != 0 || off.words[i]&w != 0 {
			return false
		}
	}
	return true
}

func (b *bdcBuild) lookupState(isf bdcISFState) (BdcLit, bool, error) {
	return b.table.Lookup(BdcISF{Support: isf.support, On: isf.on, Off: isf.off})
}

func bdcLitLess(a, b BdcLit) bool {
	if a.ID() != b.ID() {
		return a.ID() < b.ID()
	}
	return !a.Complemented() && b.Complemented()
}

func (b *bdcBuild) and(a, c BdcLit) (BdcLit, error) {
	one, zero := bdcMakeLit(0, false), bdcMakeLit(0, true)
	if a == zero || c == zero {
		return zero, nil
	}
	if a == one {
		return c, nil
	}
	if c == one {
		return a, nil
	}
	if a == c {
		return a, nil
	}
	if a == c.Not() {
		return zero, nil
	}
	if b.stop != nil && b.stop() {
		return InvalidBdcLit, ErrCancelled
	}
	if !bdcLitLess(a, c) {
		a, c = c, a
	}
	key := [4]int{int(a.ID()), int(c.ID()), 0, 0}
	if a.Complemented() {
		key[2] = 1
	}
	if c.Complemented() {
		key[3] = 1
	}
	if l, ok := b.andHash[key]; ok {
		return l, nil
	}
	if b.newNodes >= b.maxNew || len(b.nodes) >= b.manager.parameters.MaxNodes {
		return InvalidBdcLit, ErrLimit
	}
	t, e := b.litTruth(a).And(b.litTruth(c), false, false)
	if e != nil {
		return InvalidBdcLit, e
	}
	if t.IsConstant(true) {
		return bdcMakeLit(0, false), nil
	}
	if t.IsConstant(false) {
		return bdcMakeLit(0, true), nil
	}
	s, e := bdcSupportMask(t)
	if e != nil {
		return InvalidBdcLit, e
	}
	id := len(b.nodes)
	b.nodes = append(b.nodes, BdcNode{Type: BdcNodeAND, Fanin0: a, Fanin1: c, Support: s, truth: t})
	lit, e := b.table.Add(s, t)
	if e != nil {
		b.nodes = b.nodes[:id]
		return InvalidBdcLit, e
	}
	if int(lit.ID()) != id {
		b.nodes = b.nodes[:id]
		return InvalidBdcLit, ErrInvariant
	}
	b.andHash[key] = bdcMakeLit(id, false)
	b.newNodes++
	return bdcMakeLit(id, false), nil
}

func (b *bdcBuild) mux(control, when1, when0 BdcLit) (BdcLit, error) {
	if when1 == when0 {
		return when1, nil
	}
	a, e := b.and(control.Not(), when0)
	if e != nil {
		return InvalidBdcLit, e
	}
	c, e := b.and(control, when1)
	if e != nil {
		return InvalidBdcLit, e
	}
	d, e := b.and(a.Not(), c.Not())
	if e != nil {
		return InvalidBdcLit, e
	}
	return d.Not(), nil
}

func (b *bdcBuild) or(a, c BdcLit) (BdcLit, error) {
	lit, err := b.and(a.Not(), c.Not())
	if err != nil {
		return InvalidBdcLit, err
	}
	return lit.Not(), nil
}

func bdcCofactorPair(on, off *TruthTable, v int, value bool) (*TruthTable, *TruthTable, error) {
	a, e := on.Cofactor(v, value)
	if e != nil {
		return nil, nil, e
	}
	c, e := off.Cofactor(v, value)
	return a, c, e
}

type bdcDecompositionType uint8

const (
	bdcDecompositionAND bdcDecompositionType = iota + 1
	bdcDecompositionOR
	bdcDecompositionMUX
)

func bdcExist(t *TruthTable, variable int) (*TruthTable, error) {
	a, err := t.Cofactor(variable, false)
	if err != nil {
		return nil, err
	}
	b, err := t.Cofactor(variable, true)
	if err != nil {
		return nil, err
	}
	return a.Or(b)
}

func bdcForall(t *TruthTable, variable int) (*TruthTable, error) {
	a, err := t.Cofactor(variable, false)
	if err != nil {
		return nil, err
	}
	b, err := t.Cofactor(variable, true)
	if err != nil {
		return nil, err
	}
	return a.And(b, false, false)
}

func bdcExistMask(t *TruthTable, mask uint32) (*TruthTable, error) {
	r := t.Clone()
	var err error
	for variable := 0; variable < t.Variables(); variable++ {
		if mask&(1<<uint(variable)) != 0 {
			r, err = bdcExist(r, variable)
			if err != nil {
				return nil, err
			}
		}
	}
	return r, nil
}

func bdcDisjoint(a, b *TruthTable) (bool, error) {
	intersects, err := a.Intersects(b, false)
	return !intersects, err
}

func bdcDisjoint3(a, b, c *TruthTable) (bool, error) {
	pair, err := b.And(c, false, false)
	if err != nil {
		return false, err
	}
	return bdcDisjoint(a, pair)
}

func bdcMinimizeISF(isf *bdcISFState) error {
	if isf == nil || isf.on == nil || isf.off == nil {
		return ErrDimension
	}
	isf.support = 0
	for variable := 0; variable < isf.on.Variables(); variable++ {
		onHas, err := isf.on.HasVariable(variable)
		if err != nil {
			return err
		}
		offHas, err := isf.off.HasVariable(variable)
		if err != nil {
			return err
		}
		if !onHas && !offHas {
			continue
		}
		onExist, err := bdcExist(isf.on, variable)
		if err != nil {
			return err
		}
		offExist, err := bdcExist(isf.off, variable)
		if err != nil {
			return err
		}
		vacuous, err := bdcDisjoint(onExist, offExist)
		if err != nil {
			return err
		}
		if vacuous {
			isf.on, isf.off = onExist, offExist
		} else {
			isf.support |= 1 << uint(variable)
		}
	}
	return nil
}

func (b *bdcBuild) weakOR(parent bdcISFState) (int, bdcISFState, bdcISFState, error) {
	bestVariable, bestCount := -1, 0
	for variable := 0; variable < b.variables; variable++ {
		if parent.support&(1<<uint(variable)) == 0 {
			continue
		}
		existOff, err := bdcExist(parent.off, variable)
		if err != nil {
			return 0, bdcISFState{}, bdcISFState{}, err
		}
		implies, err := parent.on.Implies(existOff)
		if err != nil {
			return 0, bdcISFState{}, bdcISFState{}, err
		}
		if implies {
			continue
		}
		forallOn, err := bdcForall(parent.on, variable)
		if err != nil {
			return 0, bdcISFState{}, bdcISFState{}, err
		}
		count := forallOn.CountOnes()
		if count == 0 {
			count = 1
		}
		if count > bestCount {
			bestVariable, bestCount = variable, count
		}
	}
	if bestVariable < 0 {
		return 0, bdcISFState{}, bdcISFState{}, nil
	}
	existOff, err := bdcExist(parent.off, bestVariable)
	if err != nil {
		return 0, bdcISFState{}, bdcISFState{}, err
	}
	leftOn, err := parent.on.And(existOff, false, false)
	if err != nil {
		return 0, bdcISFState{}, bdcISFState{}, err
	}
	cost := bestCount * 1000 / (1 << uint(b.variables))
	if cost == 0 {
		cost = 1
	}
	left := bdcISFState{on: leftOn, off: parent.off.Clone(), unique: 1 << uint(bestVariable)}
	return cost, left, bdcISFState{}, nil
}

func (b *bdcBuild) decomposeOR(parent bdcISFState) (int, bdcISFState, bdcISFState, error) {
	variables := make([]int, 0, b.variables)
	for variable := 0; variable < b.variables; variable++ {
		if parent.support&(1<<uint(variable)) != 0 {
			variables = append(variables, variable)
		}
	}
	left, right := bdcISFState{}, bdcISFState{}
	found := false
	for beg := 0; beg < len(variables) && !found; beg++ {
		leftOff, err := bdcExist(parent.off, variables[beg])
		if err != nil {
			return 0, left, right, err
		}
		for end := len(variables) - 1; end > beg; end-- {
			rightOff, err := bdcExist(parent.off, variables[end])
			if err != nil {
				return 0, left, right, err
			}
			ok, err := bdcDisjoint3(parent.on, leftOff, rightOff)
			if err != nil {
				return 0, left, right, err
			}
			if ok {
				left.unique = 1 << uint(variables[beg])
				right.unique = 1 << uint(variables[end])
				found = true
				break
			}
		}
	}
	if !found {
		return b.weakOR(parent)
	}
	existLeft, err := bdcExistMask(parent.off, left.unique)
	if err != nil {
		return 0, left, right, err
	}
	existRight, err := bdcExistMask(parent.off, right.unique)
	if err != nil {
		return 0, left, right, err
	}
	nLeft, nRight := 1, 1
	remaining := parent.support &^ left.unique &^ right.unique
	for variable := 0; variable < b.variables; variable++ {
		if remaining&(1<<uint(variable)) == 0 {
			continue
		}
		tempLeft, err := bdcExist(existLeft, variable)
		if err != nil {
			return 0, left, right, err
		}
		tempRight, err := bdcExist(existRight, variable)
		if err != nil {
			return 0, left, right, err
		}
		tryLeft := func() (bool, error) { return bdcDisjoint3(parent.on, tempLeft, existRight) }
		tryRight := func() (bool, error) { return bdcDisjoint3(parent.on, tempRight, existLeft) }
		if nLeft < nRight {
			ok, e := tryLeft()
			if e != nil {
				return 0, left, right, e
			}
			if ok {
				left.unique |= 1 << uint(variable)
				nLeft++
				existLeft = tempLeft
				continue
			}
			ok, e = tryRight()
			if e != nil {
				return 0, left, right, e
			}
			if ok {
				right.unique |= 1 << uint(variable)
				nRight++
				existRight = tempRight
			}
		} else {
			ok, e := tryRight()
			if e != nil {
				return 0, left, right, e
			}
			if ok {
				right.unique |= 1 << uint(variable)
				nRight++
				existRight = tempRight
				continue
			}
			ok, e = tryLeft()
			if e != nil {
				return 0, left, right, e
			}
			if ok {
				left.unique |= 1 << uint(variable)
				nLeft++
				existLeft = tempLeft
			}
		}
	}
	left.on, err = parent.on.And(existLeft, false, false)
	if err != nil {
		return 0, left, right, err
	}
	left.on, err = bdcExistMask(left.on, right.unique)
	if err != nil {
		return 0, left, right, err
	}
	left.off = existRight.Clone()
	right.on, err = parent.on.And(existRight, false, false)
	if err != nil {
		return 0, left, right, err
	}
	right.on, err = bdcExistMask(right.on, left.unique)
	if err != nil {
		return 0, left, right, err
	}
	right.off = existLeft.Clone()
	cost := 1000 * (b.variables*min(nLeft, nRight) + max(nLeft, nRight))
	return cost, left, right, nil
}

func (b *bdcBuild) cachedComponent(isf *bdcISFState) (int, error) {
	if isf.unique == 0 {
		return 0, nil
	}
	if err := bdcMinimizeISF(isf); err != nil {
		return 0, err
	}
	_, found, err := b.lookupState(*isf)
	if err != nil {
		return 0, err
	}
	if found {
		return 1, nil
	}
	return 0, nil
}

func (b *bdcBuild) decompositionStep(parent bdcISFState) (bdcDecompositionType, bdcISFState, bdcISFState, error) {
	orCost, orLeft, orRight, err := b.decomposeOR(parent)
	if err != nil {
		return 0, orLeft, orRight, err
	}
	swapped := bdcISFState{on: parent.off.Clone(), off: parent.on.Clone(), support: parent.support}
	andCost, andLeft, andRight, err := b.decomposeOR(swapped)
	if err != nil {
		return 0, andLeft, andRight, err
	}
	if andLeft.on != nil {
		andLeft.on, andLeft.off = andLeft.off, andLeft.on
	}
	if andRight.on != nil {
		andRight.on, andRight.off = andRight.off, andRight.on
	}
	if orCost == 0 && andCost == 0 {
		return bdcDecompositionMUX, bdcISFState{}, bdcISFState{}, nil
	}
	ol, err := b.cachedComponent(&orLeft)
	if err != nil {
		return 0, orLeft, orRight, err
	}
	orr, err := b.cachedComponent(&orRight)
	if err != nil {
		return 0, orLeft, orRight, err
	}
	al, err := b.cachedComponent(&andLeft)
	if err != nil {
		return 0, andLeft, andRight, err
	}
	ar, err := b.cachedComponent(&andRight)
	if err != nil {
		return 0, andLeft, andRight, err
	}
	if ol+orr > al+ar {
		b.statistics.Reused++
		b.statistics.OrChoices++
		return bdcDecompositionOR, orLeft, orRight, nil
	}
	if ol+orr < al+ar {
		b.statistics.Reused++
		b.statistics.AndChoices++
		return bdcDecompositionAND, andLeft, andRight, nil
	}
	if orCost > andCost {
		if orCost < 1000 {
			b.statistics.WeakChoices++
		}
		b.statistics.OrChoices++
		return bdcDecompositionOR, orLeft, orRight, nil
	}
	if andCost < 1000 {
		b.statistics.WeakChoices++
	}
	b.statistics.AndChoices++
	return bdcDecompositionAND, andLeft, andRight, nil
}

func (b *bdcBuild) updateRight(parent, left bdcISFState, right *bdcISFState, leftLit BdcLit, typ bdcDecompositionType) (bool, error) {
	truth := b.litTruth(leftLit)
	var err error
	if typ == bdcDecompositionOR {
		right.on, err = parent.on.And(truth, false, true)
		if err != nil {
			return false, err
		}
		right.on, err = bdcExistMask(right.on, left.unique)
		if err != nil {
			return false, err
		}
		right.off, err = bdcExistMask(parent.off, left.unique)
		if err != nil {
			return false, err
		}
		return right.on.IsConstant(false), nil
	}
	right.off, err = parent.off.And(truth, false, false)
	if err != nil {
		return false, err
	}
	right.off, err = bdcExistMask(right.off, left.unique)
	if err != nil {
		return false, err
	}
	right.on, err = bdcExistMask(parent.on, left.unique)
	if err != nil {
		return false, err
	}
	return right.off.IsConstant(false), nil
}

func (b *bdcBuild) decompose(isf bdcISFState, depth int) (BdcLit, error) {
	if b.stop != nil && b.stop() {
		return InvalidBdcLit, ErrCancelled
	}
	if depth > b.manager.parameters.MaxNodes {
		return InvalidBdcLit, ErrInvariant
	}
	if lit, found, err := b.lookupState(isf); err != nil {
		return InvalidBdcLit, err
	} else if found {
		return lit, nil
	}
	typ, left, right, err := b.decompositionStep(isf)
	if err != nil {
		return InvalidBdcLit, err
	}
	if typ == bdcDecompositionMUX {
		variable := -1
		for v := 0; v < b.variables; v++ {
			if isf.support&(1<<uint(v)) != 0 {
				variable = v
				break
			}
		}
		if variable < 0 {
			return InvalidBdcLit, ErrInvariant
		}
		left.on, left.off, err = bdcCofactorPair(isf.on, isf.off, variable, false)
		if err != nil {
			return InvalidBdcLit, err
		}
		right.on, right.off, err = bdcCofactorPair(isf.on, isf.off, variable, true)
		if err != nil {
			return InvalidBdcLit, err
		}
		if err = bdcMinimizeISF(&left); err != nil {
			return InvalidBdcLit, err
		}
		if err = bdcMinimizeISF(&right); err != nil {
			return InvalidBdcLit, err
		}
		b.statistics.Muxes++
		low, err := b.decompose(left, depth+1)
		if err != nil {
			return InvalidBdcLit, err
		}
		high, err := b.decompose(right, depth+1)
		if err != nil {
			return InvalidBdcLit, err
		}
		return b.mux(bdcMakeLit(variable+1, false), high, low)
	}
	leftLit, err := b.decompose(left, depth+1)
	if err != nil {
		return InvalidBdcLit, err
	}
	absorbed, err := b.updateRight(isf, left, &right, leftLit, typ)
	if err != nil {
		return InvalidBdcLit, err
	}
	if absorbed {
		if b.newNodes > 0 {
			b.newNodes--
		}
		return leftLit, nil
	}
	if err = bdcMinimizeISF(&right); err != nil {
		return InvalidBdcLit, err
	}
	rightLit, err := b.decompose(right, depth+1)
	if err != nil {
		return InvalidBdcLit, err
	}
	if typ == bdcDecompositionOR {
		return b.or(leftLit, rightLit)
	}
	return b.and(leftLit, rightLit)
}

// Decompose synthesizes an incompletely specified function. Function and care
// are owned by the caller; divisor tables are copied. A nil care means full
// care. The result is an AND/inverter graph whose truth contains the on-set
// and excludes the cared off-set.
func (m *BdcManager) Decompose(function, care *TruthTable, divisors []*TruthTable, options BdcDecomposeOptions) (int, error) {
	if m == nil || function == nil {
		return 0, ErrDimension
	}
	if e := function.Check(); e != nil {
		return 0, e
	}
	vars := function.Variables()
	if vars < 0 || vars > m.parameters.MaxVariables {
		return 0, ErrDimension
	}
	if care != nil {
		if e := function.checkPair(care); e != nil {
			return 0, e
		}
	}
	if options.MaxNewNodes < 0 || options.MaxDivisors < 0 {
		return 0, ErrDimension
	}
	maxNew := options.MaxNewNodes
	if maxNew == 0 {
		maxNew = m.parameters.MaxNodes
	}
	maxDivs := options.MaxDivisors
	if maxDivs == 0 {
		maxDivs = 201
	}
	if options.Stop != nil && options.Stop() {
		return 0, ErrCancelled
	}
	table, e := NewBdcTable(BdcTableOptions{Variables: vars, MaxFunctions: m.parameters.MaxNodes, Stop: options.Stop})
	if e != nil {
		return 0, e
	}
	b := &bdcBuild{manager: m, variables: vars, maxNew: maxNew, stop: options.Stop, table: table, andHash: make(map[[4]int]BdcLit)}
	one, e := NewConstantTruthTable(vars, true, m.parameters.MaxWords)
	if e != nil {
		return 0, e
	}
	if _, e = b.addBase(BdcNodeConst1, one); e != nil {
		return 0, e
	}
	for v := 0; v < vars; v++ {
		t, x := NewVariableTruthTable(vars, v, m.parameters.MaxWords)
		if x != nil {
			return 0, x
		}
		if _, x = b.addBase(BdcNodePI, t); x != nil {
			return 0, x
		}
	}
	if len(divisors) > maxDivs {
		divisors = divisors[:maxDivs]
	}
	for _, t := range divisors {
		if t == nil {
			return 0, ErrDimension
		}
		if e = function.checkPair(t); e != nil {
			return 0, e
		}
		if t.IsConstant(true) || t.IsConstant(false) {
			continue
		}
		if _, e = b.addBase(BdcNodePI, t); e != nil {
			return 0, e
		}
	}
	if care != nil && care.IsConstant(false) {
		m.variables = vars
		m.nodes = b.nodes
		m.root = bdcMakeLit(0, true)
		m.newNodes = 0
		m.statistics.Calls++
		return 0, nil
	}
	on := function.Clone()
	off := function.Not()
	if care != nil {
		var x error
		on, x = function.And(care, false, false)
		if x != nil {
			return 0, x
		}
		off, x = function.Not().And(care, false, false)
		if x != nil {
			return 0, x
		}
	}
	isf := bdcISFState{on: on, off: off}
	if e = bdcMinimizeISF(&isf); e != nil {
		return 0, e
	}
	root, e := b.decompose(isf, 0)
	if e != nil {
		return 0, e
	}
	if !bdcIntervalContains(on, off, b.litTruth(root)) {
		return 0, ErrInvariant
	}
	m.variables = vars
	m.nodes = b.nodes
	m.root = root
	m.newNodes = b.newNodes
	m.statistics.Calls++
	m.statistics.Nodes += b.newNodes
	m.statistics.Reused += b.statistics.Reused
	m.statistics.Muxes += b.statistics.Muxes
	m.statistics.AndChoices += b.statistics.AndChoices
	m.statistics.OrChoices += b.statistics.OrChoices
	m.statistics.WeakChoices += b.statistics.WeakChoices
	return b.newNodes, nil
}

func (m *BdcManager) Evaluate(assignment uint64) (bool, error) {
	if m == nil || len(m.nodes) == 0 {
		return false, ErrDimension
	}
	if assignment >= uint64(1)<<uint(m.variables) {
		return false, ErrDimension
	}
	values := make([]bool, len(m.nodes))
	values[0] = true
	value := func(l BdcLit) bool {
		x := values[int(l.ID())]
		if l.Complemented() {
			return !x
		}
		return x
	}
	for id := 1; id < len(m.nodes); id++ {
		n := m.nodes[id]
		switch n.Type {
		case BdcNodePI:
			if id <= m.variables {
				values[id] = assignment&(1<<uint(id-1)) != 0
			} else {
				x, e := n.truth.Get(assignment)
				if e != nil {
					return false, e
				}
				values[id] = x
			}
		case BdcNodeAND:
			values[id] = value(n.Fanin0) && value(n.Fanin1)
		default:
			return false, ErrInvariant
		}
	}
	return value(m.root), nil
}

func (m *BdcManager) Format() (string, error) {
	if m == nil || len(m.nodes) == 0 {
		return "", ErrDimension
	}
	var rec func(BdcLit) (string, error)
	rec = func(l BdcLit) (string, error) {
		id := int(l.ID())
		if !l.Valid() || id >= len(m.nodes) {
			return "", ErrInvariant
		}
		n := m.nodes[id]
		var s string
		switch n.Type {
		case BdcNodeConst1:
			s = "1"
		case BdcNodePI:
			if id <= m.variables {
				s = "x" + strconv.Itoa(id-1)
			} else {
				s = "d" + strconv.Itoa(id-m.variables-1)
			}
		case BdcNodeAND:
			a, e := rec(n.Fanin0)
			if e != nil {
				return "", e
			}
			b, e := rec(n.Fanin1)
			if e != nil {
				return "", e
			}
			s = "(" + a + " & " + b + ")"
		default:
			return "", ErrInvariant
		}
		if l.Complemented() {
			s = "!" + s
		}
		return s, nil
	}
	return rec(m.root)
}

// BdcSPFDAdjacencyCost sums Boolean-edge transitions of a six-variable word.
func BdcSPFDAdjacencyCost(truth uint64) int {
	c := 0
	for v := 0; v < 6; v++ {
		a, _ := BdcSPFDCofactor(truth, v, false)
		b, _ := BdcSPFDCofactor(truth, v, true)
		c += bits.OnesCount64(a ^ b)
	}
	return c
}
func BdcSPFDCofactor(truth uint64, variable int, value bool) (uint64, error) {
	if variable < 0 || variable >= 6 {
		return 0, ErrDimension
	}
	mask := truthVariablePattern(variable)
	shift := uint(1) << uint(variable)
	if value {
		x := truth & mask
		return x | (x >> shift), nil
	}
	x := truth &^ mask
	return x | (x << shift), nil
}
func BdcSPFDSeparationCount(candidate, target uint64) int {
	return bits.OnesCount64(^candidate&^target)*bits.OnesCount64(candidate&target) + bits.OnesCount64(candidate&^target)*bits.OnesCount64(^candidate&target)
}
func BdcSPFDHash(truth uint64, size int) (int, error) {
	if size <= 0 || uint64(size) > uint64(^uint32(0)>>1) {
		return 0, ErrDimension
	}
	p := [8]uint32{12582917, 25165843, 50331653, 100663319, 201326611, 402653189, 805306457, 1610612741}
	v := uint32(0)
	for i := 0; i < 8; i++ {
		v ^= p[i] * uint32(byte(truth>>uint(8*i)))
	}
	return int(v % uint32(size)), nil
}

type BdcSPFDSelection struct {
	Truth                uint64
	Index, Score, Weight int
}

func SelectBdcSPFDDivisor(divisors []uint64, weights []int, off, on uint64) (BdcSPFDSelection, error) {
	if len(divisors) == 0 || len(divisors) != len(weights) || off&on != 0 {
		return BdcSPFDSelection{}, ErrDimension
	}
	best := BdcSPFDSelection{Index: -1, Score: -1}
	for i, f := range divisors {
		cands := [4]struct {
			valid  bool
			f, rem uint64
		}{{f&off == 0, f, f & on}, {f&on == 0, f, f & off}, {^f&off == 0, ^f, ^f & on}, {^f&on == 0, ^f, ^f & off}}
		for _, c := range cands {
			if !c.valid {
				continue
			}
			ones := bits.OnesCount64(c.rem)
			if ones == 0 {
				continue
			}
			score := 7*ones + 10*(8-weights[i])
			if score > best.Score {
				best = BdcSPFDSelection{c.f, i, score, weights[i]}
			}
		}
	}
	if best.Index < 0 {
		return BdcSPFDSelection{}, ErrInvariant
	}
	return best, nil
}

// BdcSPFDEntry is a concrete gate catalogue entry used by overlap checks and
// bounded stand-alone dataset generation.
type BdcSPFDEntry struct {
	Fanin0, Fanin1                int
	Complement0, Complement1, XOR bool
	Truth                         uint64
}

type BdcSPFDSearchOptions struct {
	Variables, Candidates, MaxGates, MaxGenerated int
	Stop                                          func() bool
}

type BdcSPFDSearchResult struct {
	Entries      []BdcSPFDEntry
	Root         int
	Complemented bool
	GateCount    int
	Generated    int
}

// SearchBdcSPFD is the bounded SPFD-guided decomposition experiment from
// bdcSpfd.c without printing or process-global allocation. Each level retains
// the candidates with the largest separation score. Returned fanins are
// topological indices into Entries.
func SearchBdcSPFD(target uint64, options BdcSPFDSearchOptions) (BdcSPFDSearchResult, error) {
	if options.Variables < 1 || options.Variables > 6 || options.Candidates <= 0 || options.MaxGates < 0 || options.MaxGenerated < 0 {
		return BdcSPFDSearchResult{}, ErrDimension
	}
	if options.MaxGenerated == 0 {
		options.MaxGenerated = 1 << 20
	}
	if options.Stop != nil && options.Stop() {
		return BdcSPFDSearchResult{}, ErrCancelled
	}
	result := BdcSPFDSearchResult{Root: -1}
	levels := make([][]int, options.MaxGates+1)
	for variable := 0; variable < options.Variables; variable++ {
		entry := BdcSPFDEntry{Fanin0: -1, Fanin1: variable, Truth: truthVariablePattern(variable)}
		result.Entries = append(result.Entries, entry)
		levels[0] = append(levels[0], len(result.Entries)-1)
		if entry.Truth == target || entry.Truth == ^target {
			result.Root = len(result.Entries) - 1
			result.Complemented = entry.Truth != target
			return result, nil
		}
	}
	if target == 0 || target == ^uint64(0) {
		result.Entries = append(result.Entries, BdcSPFDEntry{Fanin0: -1, Fanin1: -1, Truth: ^uint64(0)})
		result.Root = len(result.Entries) - 1
		result.Complemented = target == 0
		return result, nil
	}
	type candidate struct {
		entry  BdcSPFDEntry
		weight int
	}
	for gates := 1; gates <= options.MaxGates; gates++ {
		if options.Stop != nil && options.Stop() {
			return BdcSPFDSearchResult{}, ErrCancelled
		}
		candidates := make([]candidate, 0)
		addPair := func(left, right int) error {
			a, c := result.Entries[left].Truth, result.Entries[right].Truth
			variants := [5]BdcSPFDEntry{{Fanin0: left, Fanin1: right, Truth: a & c}, {Fanin0: left, Fanin1: right, Complement0: true, Truth: ^a & c}, {Fanin0: left, Fanin1: right, Complement1: true, Truth: a &^ c}, {Fanin0: left, Fanin1: right, Complement0: true, Complement1: true, Truth: ^a &^ c}, {Fanin0: left, Fanin1: right, XOR: true, Truth: a ^ c}}
			for _, entry := range variants {
				result.Generated++
				if result.Generated&1023 == 0 && options.Stop != nil && options.Stop() {
					return ErrCancelled
				}
				if result.Generated > options.MaxGenerated {
					return ErrLimit
				}
				if entry.Truth == target || entry.Truth == ^target {
					result.Entries = append(result.Entries, entry)
					result.Root = len(result.Entries) - 1
					result.Complemented = entry.Truth != target
					result.GateCount = gates
					return nil
				}
				candidates = append(candidates, candidate{entry, BdcSPFDSeparationCount(entry.Truth, target)})
			}
			return nil
		}
		previous := levels[gates-1]
		if gates > 1 {
			for level := 0; level < gates-1; level++ {
				for _, left := range levels[level] {
					for _, right := range previous {
						if err := addPair(left, right); err != nil {
							return BdcSPFDSearchResult{}, err
						}
						if result.Root >= 0 {
							return result, nil
						}
					}
				}
			}
		}
		for i := 0; i < len(previous); i++ {
			for j := i + 1; j < len(previous); j++ {
				if err := addPair(previous[i], previous[j]); err != nil {
					return BdcSPFDSearchResult{}, err
				}
				if result.Root >= 0 {
					return result, nil
				}
			}
		}
		slices.SortFunc(candidates, func(a, c candidate) int {
			if a.weight != c.weight {
				if a.weight > c.weight {
					return -1
				}
				return 1
			}
			if a.entry.Truth != c.entry.Truth {
				if a.entry.Truth < c.entry.Truth {
					return -1
				}
				return 1
			}
			if a.entry.Fanin0 != c.entry.Fanin0 {
				if a.entry.Fanin0 < c.entry.Fanin0 {
					return -1
				}
				return 1
			}
			if a.entry.Fanin1 < c.entry.Fanin1 {
				return -1
			}
			if a.entry.Fanin1 > c.entry.Fanin1 {
				return 1
			}
			return 0
		})
		if gates > 1 && len(candidates) > options.Candidates {
			candidates = candidates[:options.Candidates]
		}
		for _, c := range candidates {
			result.Entries = append(result.Entries, c.entry)
			levels[gates] = append(levels[gates], len(result.Entries)-1)
		}
	}
	return result, nil
}

func BdcSPFDOverlap(entries []BdcSPFDEntry, first, second int) (int, error) {
	if first < 0 || second < 0 || first >= len(entries) || second >= len(entries) {
		return 0, ErrDimension
	}
	seen := make([]bool, len(entries))
	var mark func(int) error
	mark = func(i int) error {
		if i < 0 || i >= len(entries) {
			return ErrInvariant
		}
		e := entries[i]
		if e.Fanin0 < 0 {
			return nil
		}
		if seen[i] {
			return nil
		}
		seen[i] = true
		if e.Fanin0 >= i || e.Fanin1 >= i {
			return ErrInvariant
		}
		if x := mark(e.Fanin0); x != nil {
			return x
		}
		return mark(e.Fanin1)
	}
	if e := mark(first); e != nil {
		return 0, e
	}
	var visit func(int) (int, error)
	done := make([]bool, len(entries))
	visit = func(i int) (int, error) {
		if i < 0 || i >= len(entries) {
			return 0, ErrInvariant
		}
		e := entries[i]
		if e.Fanin0 < 0 {
			return 0, nil
		}
		if done[i] {
			return 0, nil
		}
		done[i] = true
		if e.Fanin0 >= i || e.Fanin1 >= i {
			return 0, ErrInvariant
		}
		count := 0
		if seen[i] {
			count = 1
		}
		a, x := visit(e.Fanin0)
		if x != nil {
			return 0, x
		}
		c, x := visit(e.Fanin1)
		if x != nil {
			return 0, x
		}
		return count + a + c, nil
	}
	return visit(second)
}

// Marshal/ParseBdcSPFDDataset replace the source's fixed global binary files.
// Records are little-endian (truth, signed weight) and entirely caller-owned.
func MarshalBdcSPFDDataset(truths []uint64, weights []int32, maxBytes int) ([]byte, error) {
	if len(truths) != len(weights) || maxBytes < 0 {
		return nil, ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 1 << 28
	}
	if len(truths) > maxBytes/12 {
		return nil, ErrLimit
	}
	out := make([]byte, 12*len(truths))
	for i, t := range truths {
		binary.LittleEndian.PutUint64(out[12*i:], t)
		binary.LittleEndian.PutUint32(out[12*i+8:], uint32(weights[i]))
	}
	return out, nil
}
func ParseBdcSPFDDataset(data []byte, maxRecords int) ([]uint64, []int32, error) {
	if len(data)%12 != 0 || maxRecords < 0 {
		return nil, nil, ErrDimension
	}
	n := len(data) / 12
	if maxRecords == 0 {
		maxRecords = 1 << 20
	}
	if n > maxRecords {
		return nil, nil, ErrLimit
	}
	t := make([]uint64, n)
	w := make([]int32, n)
	for i := range t {
		t[i] = binary.LittleEndian.Uint64(data[12*i:])
		w[i] = int32(binary.LittleEndian.Uint32(data[12*i+8:]))
	}
	return t, w, nil
}

// ValidateBdcSPFDEntries checks the catalogue truth and topology independently.
func ValidateBdcSPFDEntries(entries []BdcSPFDEntry) error {
	for i, e := range entries {
		if e.Fanin0 < 0 {
			if e.Fanin1 == -1 && e.Truth == ^uint64(0) {
				continue
			}
			if e.Fanin1 < 0 || e.Fanin1 >= 6 || e.Truth != truthVariablePattern(e.Fanin1) {
				return ErrInvariant
			}
			continue
		}
		if e.Fanin0 >= i || e.Fanin1 < 0 || e.Fanin1 >= i {
			return ErrInvariant
		}
		a, b := entries[e.Fanin0].Truth, entries[e.Fanin1].Truth
		if e.Complement0 {
			a = ^a
		}
		if e.Complement1 {
			b = ^b
		}
		want := a & b
		if e.XOR {
			want = a ^ b
		}
		if want != e.Truth {
			return ErrInvariant
		}
	}
	return nil
}

func FormatBdcSPFD(entries []BdcSPFDEntry, index int) (string, error) {
	if e := ValidateBdcSPFDEntries(entries); e != nil {
		return "", e
	}
	if index < 0 || index >= len(entries) {
		return "", ErrDimension
	}
	var rec func(int) string
	rec = func(i int) string {
		e := entries[i]
		if e.Fanin0 < 0 {
			if e.Fanin1 < 0 {
				return "1"
			}
			return string(rune('a' + e.Fanin1))
		}
		a, b := rec(e.Fanin0), rec(e.Fanin1)
		if e.Complement0 {
			a = "!" + a
		}
		if e.Complement1 {
			b = "!" + b
		}
		op := "*"
		if e.XOR {
			op = "+"
		}
		return "(" + a + op + b + ")"
	}
	return rec(index), nil
}
