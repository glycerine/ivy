package abc

// Owned Go port of the ten src/aig/hop module.make units, the hop.h inline
// operations and iterators, and the 16-function cudd2.c adapter.

import (
	"math"
	"math/bits"
	"slices"
	"strconv"
	"strings"
)

type HopEdge uint32

func NewHopEdge(id int, complemented bool) (HopEdge, error) {
	if id < 0 || id > math.MaxInt32 {
		return 0, ErrDimension
	}
	edge := HopEdge(uint32(id) << 1)
	if complemented {
		edge |= 1
	}
	return edge, nil
}

func (edge HopEdge) ID() int            { return int(uint32(edge) >> 1) }
func (edge HopEdge) Complemented() bool { return edge&1 != 0 }
func (edge HopEdge) Regular() HopEdge   { return edge &^ 1 }
func (edge HopEdge) Not() HopEdge       { return edge ^ 1 }
func (edge HopEdge) NotIf(value bool) HopEdge {
	if value {
		return edge.Not()
	}
	return edge
}

type HopObjectType uint8

const (
	HopObjectNone HopObjectType = iota
	HopObjectConst1
	HopObjectPI
	HopObjectPO
	HopObjectAnd
	HopObjectXor
)

type HopLimits struct {
	MaxObjects int
	MaxVars    int
	Stop       func() bool
}

type HopObject struct {
	ID        int
	Type      HopObjectType
	Fanin0    HopEdge
	Fanin1    HopEdge
	PIONum    int
	Phase     bool
	Level     int
	Alive     bool
	Choice    HopEdge
	HasChoice bool
}

type hopHashKey struct {
	typeID HopObjectType
	first  HopEdge
	second HopEdge
}

type HopStats struct {
	PIs, POs, Ands, Xors int
	Created, Deleted     int
	Levels, HashEntries  int
	MemoryBytes          int
}

type HopManager struct {
	objects []HopObject
	pis     []HopEdge
	pos     []HopEdge
	table   map[hopHashKey]HopEdge
	limits  HopLimits
	deleted int
	travID  int
	closed  bool
}

func normalizeHopLimits(limits HopLimits) (HopLimits, error) {
	if limits.MaxObjects < 0 || limits.MaxVars < 0 {
		return limits, ErrDimension
	}
	if limits.MaxObjects == 0 {
		limits.MaxObjects = 1 << 24
	}
	if limits.MaxVars == 0 {
		limits.MaxVars = 20
	}
	if limits.MaxObjects < 1 {
		return limits, ErrLimit
	}
	return limits, nil
}

func NewHopManager(limits HopLimits) (*HopManager, error) {
	limits, err := normalizeHopLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	manager := &HopManager{table: make(map[hopHashKey]HopEdge), limits: limits, travID: 1}
	manager.objects = append(manager.objects, HopObject{ID: 0, Type: HopObjectConst1, Phase: true, Alive: true})
	return manager, nil
}

func (manager *HopManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed || len(manager.objects) == 0 || manager.objects[0].Type != HopObjectConst1 || !manager.objects[0].Alive {
		return ErrInvariant
	}
	return nil
}

func (manager *HopManager) stopped() bool   { return manager.limits.Stop != nil && manager.limits.Stop() }
func (manager *HopManager) Const1() HopEdge { return 0 }
func (manager *HopManager) Const0() HopEdge { return 1 }

func (manager *HopManager) valid(edge HopEdge) bool {
	id := edge.ID()
	return id >= 0 && id < len(manager.objects) && manager.objects[id].Alive
}

func (manager *HopManager) Object(edge HopEdge) (HopObject, error) {
	if err := manager.check(); err != nil {
		return HopObject{}, err
	}
	if !manager.valid(edge) {
		return HopObject{}, ErrDimension
	}
	return manager.objects[edge.ID()], nil
}

func (manager *HopManager) preflight(add int) error {
	if err := manager.check(); err != nil {
		return err
	}
	if add < 0 || len(manager.objects) > manager.limits.MaxObjects-add {
		return ErrLimit
	}
	if manager.stopped() {
		return ErrCancelled
	}
	return nil
}

func (manager *HopManager) appendObject(object HopObject) HopEdge {
	object.ID = len(manager.objects)
	object.Alive = true
	manager.objects = append(manager.objects, object)
	return HopEdge(uint32(object.ID) << 1)
}

func (manager *HopManager) IthVar(index int) (HopEdge, error) {
	if index < 0 || index >= manager.limits.MaxVars {
		return 0, ErrDimension
	}
	if index < len(manager.pis) {
		return manager.pis[index], nil
	}
	count := index + 1 - len(manager.pis)
	if err := manager.preflight(count); err != nil {
		return 0, err
	}
	for len(manager.pis) <= index {
		edge := manager.appendObject(HopObject{Type: HopObjectPI, PIONum: len(manager.pis)})
		manager.pis = append(manager.pis, edge)
	}
	return manager.pis[index], nil
}

func (manager *HopManager) rollback(size int) {
	for i := size; i < len(manager.objects); i++ {
		object := manager.objects[i]
		if object.Type == HopObjectAnd || object.Type == HopObjectXor {
			delete(manager.table, hopHashKey{object.Type, object.Fanin0, object.Fanin1})
		}
	}
	manager.objects = manager.objects[:size]
}

func (manager *HopManager) And(first, second HopEdge) (HopEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if !manager.valid(first) || !manager.valid(second) {
		return 0, ErrDimension
	}
	if first == second {
		return first, nil
	}
	if first == second.Not() {
		return manager.Const0(), nil
	}
	if first.Regular() == manager.Const1() {
		if first == manager.Const1() {
			return second, nil
		}
		return manager.Const0(), nil
	}
	if second.Regular() == manager.Const1() {
		if second == manager.Const1() {
			return first, nil
		}
		return manager.Const0(), nil
	}
	if first.ID() > second.ID() {
		first, second = second, first
	}
	key := hopHashKey{HopObjectAnd, first, second}
	if edge, exists := manager.table[key]; exists {
		return edge, nil
	}
	if err := manager.preflight(1); err != nil {
		return 0, err
	}
	level := max(manager.objects[first.ID()].Level, manager.objects[second.ID()].Level) + 1
	phase := manager.edgePhase(first) && manager.edgePhase(second)
	edge := manager.appendObject(HopObject{Type: HopObjectAnd, Fanin0: first, Fanin1: second, Level: level, Phase: phase})
	manager.table[key] = edge
	return edge, nil
}

func (manager *HopManager) edgePhase(edge HopEdge) bool {
	phase := manager.objects[edge.ID()].Phase
	if edge.Complemented() {
		return !phase
	}
	return phase
}

func (manager *HopManager) atomicCompound(build func() (HopEdge, error)) (HopEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if manager.stopped() {
		return 0, ErrCancelled
	}
	size := len(manager.objects)
	piSize, poSize := len(manager.pis), len(manager.pos)
	result, err := build()
	if err != nil {
		manager.rollback(size)
		manager.pis = manager.pis[:piSize]
		manager.pos = manager.pos[:poSize]
	}
	return result, err
}

func (manager *HopManager) Or(first, second HopEdge) (HopEdge, error) {
	return manager.atomicCompound(func() (HopEdge, error) {
		result, err := manager.And(first.Not(), second.Not())
		return result.Not(), err
	})
}

func (manager *HopManager) Xor(first, second HopEdge) (HopEdge, error) {
	return manager.atomicCompound(func() (HopEdge, error) {
		left, err := manager.And(first, second.Not())
		if err != nil {
			return 0, err
		}
		right, err := manager.And(first.Not(), second)
		if err != nil {
			return 0, err
		}
		return manager.Or(left, right)
	})
}

func (manager *HopManager) Mux(control, whenTrue, whenFalse HopEdge) (HopEdge, error) {
	return manager.atomicCompound(func() (HopEdge, error) {
		left, err := manager.And(control, whenTrue)
		if err != nil {
			return 0, err
		}
		right, err := manager.And(control.Not(), whenFalse)
		if err != nil {
			return 0, err
		}
		return manager.Or(left, right)
	})
}

func (manager *HopManager) Majority(a, b, c HopEdge) (HopEdge, error) {
	return manager.atomicCompound(func() (HopEdge, error) {
		ab, err := manager.And(a, b)
		if err != nil {
			return 0, err
		}
		ac, err := manager.And(a, c)
		if err != nil {
			return 0, err
		}
		bc, err := manager.And(b, c)
		if err != nil {
			return 0, err
		}
		left, err := manager.Or(ab, ac)
		if err != nil {
			return 0, err
		}
		return manager.Or(left, bc)
	})
}

func (manager *HopManager) Multi(edges []HopEdge, xor bool) (HopEdge, error) {
	if len(edges) == 0 {
		return 0, ErrDimension
	}
	for _, edge := range edges {
		if !manager.valid(edge) {
			return 0, ErrDimension
		}
	}
	copyEdges := slices.Clone(edges)
	return manager.atomicCompound(func() (HopEdge, error) {
		var rec func(int, int) (HopEdge, error)
		rec = func(start, stop int) (HopEdge, error) {
			if stop-start == 1 {
				return copyEdges[start], nil
			}
			middle := start + (stop-start)/2
			left, err := rec(start, middle)
			if err != nil {
				return 0, err
			}
			right, err := rec(middle, stop)
			if err != nil {
				return 0, err
			}
			if xor {
				return manager.Xor(left, right)
			}
			return manager.And(left, right)
		}
		return rec(0, len(copyEdges))
	})
}

func (manager *HopManager) CreateConjunction(variableCount int) (HopEdge, error) {
	if variableCount < 0 {
		return 0, ErrDimension
	}
	result := manager.Const1()
	return manager.atomicCompound(func() (HopEdge, error) {
		for i := 0; i < variableCount; i++ {
			variable, err := manager.IthVar(i)
			if err != nil {
				return 0, err
			}
			result, err = manager.And(result, variable)
			if err != nil {
				return 0, err
			}
		}
		return result, nil
	})
}

func (manager *HopManager) CreateDisjunction(variableCount int) (HopEdge, error) {
	if variableCount < 0 {
		return 0, ErrDimension
	}
	result := manager.Const0()
	return manager.atomicCompound(func() (HopEdge, error) {
		for i := 0; i < variableCount; i++ {
			variable, err := manager.IthVar(i)
			if err != nil {
				return 0, err
			}
			result, err = manager.Or(result, variable)
			if err != nil {
				return 0, err
			}
		}
		return result, nil
	})
}

func (manager *HopManager) CreateParity(variableCount int) (HopEdge, error) {
	if variableCount < 0 {
		return 0, ErrDimension
	}
	result := manager.Const0()
	return manager.atomicCompound(func() (HopEdge, error) {
		for i := 0; i < variableCount; i++ {
			variable, err := manager.IthVar(i)
			if err != nil {
				return 0, err
			}
			result, err = manager.Xor(result, variable)
			if err != nil {
				return 0, err
			}
		}
		return result, nil
	})
}

func (manager *HopManager) Miter(pairs []HopEdge) (HopEdge, error) {
	if len(pairs) == 0 || len(pairs)&1 != 0 {
		return 0, ErrDimension
	}
	equalities := make([]HopEdge, len(pairs)/2)
	return manager.atomicCompound(func() (HopEdge, error) {
		for i := range equalities {
			different, err := manager.Xor(pairs[2*i], pairs[2*i+1])
			if err != nil {
				return 0, err
			}
			equalities[i] = different.Not()
		}
		all, err := manager.Multi(equalities, false)
		return all.Not(), err
	})
}

func (manager *HopManager) CreatePO(driver HopEdge) (HopEdge, error) {
	if !manager.valid(driver) {
		return 0, ErrDimension
	}
	if err := manager.preflight(1); err != nil {
		return 0, err
	}
	edge := manager.appendObject(HopObject{Type: HopObjectPO, Fanin0: driver, PIONum: len(manager.pos), Phase: manager.edgePhase(driver)})
	manager.pos = append(manager.pos, edge)
	return edge, nil
}

func (manager *HopManager) DFS(root HopEdge) ([]HopEdge, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	if !manager.valid(root) {
		return nil, ErrDimension
	}
	if manager.stopped() {
		return nil, ErrCancelled
	}
	seen := make([]bool, len(manager.objects))
	result := make([]HopEdge, 0)
	var visit func(HopEdge)
	visit = func(edge HopEdge) {
		id := edge.ID()
		if seen[id] {
			return
		}
		seen[id] = true
		object := manager.objects[id]
		if object.Type == HopObjectPO {
			visit(object.Fanin0)
		}
		if object.Type == HopObjectAnd || object.Type == HopObjectXor {
			visit(object.Fanin0)
			visit(object.Fanin1)
			result = append(result, HopEdge(uint32(id)<<1))
		}
	}
	visit(root)
	return result, nil
}

func (manager *HopManager) DagSize(root HopEdge) (int, error) {
	nodes, err := manager.DFS(root)
	return len(nodes), err
}

func (manager *HopManager) References() ([]int, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	refs := make([]int, len(manager.objects))
	for _, object := range manager.objects {
		if !object.Alive {
			continue
		}
		if object.Type == HopObjectAnd || object.Type == HopObjectXor {
			refs[object.Fanin0.ID()]++
			refs[object.Fanin1.ID()]++
		}
		if object.Type == HopObjectPO {
			refs[object.Fanin0.ID()]++
		}
	}
	return refs, nil
}

func (manager *HopManager) CollectMulti(root HopEdge) ([]HopEdge, error) {
	if !manager.valid(root) || root.Complemented() {
		return nil, ErrDimension
	}
	rootType := manager.objects[root.ID()].Type
	if rootType != HopObjectAnd && rootType != HopObjectXor {
		return []HopEdge{root}, nil
	}
	seen := make(map[HopEdge]struct{})
	result := make([]HopEdge, 0)
	var rec func(HopEdge)
	rec = func(edge HopEdge) {
		object := manager.objects[edge.ID()]
		if edge != root && (edge.Complemented() || object.Type != rootType) {
			if _, ok := seen[edge]; !ok {
				seen[edge] = struct{}{}
				result = append(result, edge)
			}
			return
		}
		rec(object.Fanin0)
		rec(object.Fanin1)
	}
	rec(root)
	return result, nil
}

func (manager *HopManager) IsMuxType(root HopEdge) (bool, error) {
	if !manager.valid(root) || root.Complemented() {
		return false, ErrDimension
	}
	object := manager.objects[root.ID()]
	if object.Type != HopObjectAnd || !object.Fanin0.Complemented() || !object.Fanin1.Complemented() {
		return false, nil
	}
	left, right := manager.objects[object.Fanin0.ID()], manager.objects[object.Fanin1.ID()]
	if left.Type != HopObjectAnd || right.Type != HopObjectAnd {
		return false, nil
	}
	leftFans := [2]HopEdge{left.Fanin0, left.Fanin1}
	rightFans := [2]HopEdge{right.Fanin0, right.Fanin1}
	for _, a := range leftFans {
		for _, b := range rightFans {
			if a.Regular() == b.Regular() && a.Complemented() != b.Complemented() {
				return true, nil
			}
		}
	}
	return false, nil
}

func (manager *HopManager) RecognizeXor(root HopEdge) (HopEdge, HopEdge, bool, error) {
	if !manager.valid(root) || root.Complemented() {
		return 0, 0, false, ErrDimension
	}
	object := manager.objects[root.ID()]
	if object.Type == HopObjectXor {
		return object.Fanin0, object.Fanin1, true, nil
	}
	if object.Type != HopObjectAnd || !object.Fanin0.Complemented() || !object.Fanin1.Complemented() {
		return 0, 0, false, nil
	}
	left, right := manager.objects[object.Fanin0.ID()], manager.objects[object.Fanin1.ID()]
	if left.Type != HopObjectAnd || right.Type != HopObjectAnd || left.Fanin0.Regular() != right.Fanin0.Regular() || left.Fanin1.Regular() != right.Fanin1.Regular() || left.Fanin0.Complemented() == right.Fanin0.Complemented() || left.Fanin1.Complemented() == right.Fanin1.Complemented() {
		return 0, 0, false, nil
	}
	return left.Fanin0, left.Fanin1, true, nil
}

func (manager *HopManager) RecognizeMux(root HopEdge) (control, whenTrue, whenFalse HopEdge, ok bool, err error) {
	mux, checkErr := manager.IsMuxType(root)
	if checkErr != nil || !mux {
		return 0, 0, 0, false, checkErr
	}
	object := manager.objects[root.ID()]
	left, right := manager.objects[object.Fanin0.ID()], manager.objects[object.Fanin1.ID()]
	leftFans := [2]HopEdge{left.Fanin0, left.Fanin1}
	rightFans := [2]HopEdge{right.Fanin0, right.Fanin1}
	for i, a := range leftFans {
		for j, b := range rightFans {
			if a.Regular() != b.Regular() || a.Complemented() == b.Complemented() {
				continue
			}
			if !a.Complemented() {
				return a, leftFans[1-i].Not(), rightFans[1-j].Not(), true, nil
			}
			return b, rightFans[1-j].Not(), leftFans[1-i].Not(), true, nil
		}
	}
	return 0, 0, 0, false, nil
}

func (manager *HopManager) FanoutCount(root, pivot HopEdge) (int, error) {
	nodes, err := manager.DFS(root)
	if err != nil {
		return 0, err
	}
	if !manager.valid(pivot) {
		return 0, ErrDimension
	}
	count := 0
	for _, edge := range nodes {
		object := manager.objects[edge.ID()]
		if (object.Type == HopObjectAnd || object.Type == HopObjectXor || object.Type == HopObjectPO) && object.Fanin0.Regular() == pivot.Regular() {
			count++
		}
		if (object.Type == HopObjectAnd || object.Type == HopObjectXor) && object.Fanin1.Regular() == pivot.Regular() {
			count++
		}
	}
	return count, nil
}

func (manager *HopManager) Evaluate(root HopEdge, inputs []bool) (bool, error) {
	if err := manager.check(); err != nil {
		return false, err
	}
	if !manager.valid(root) || len(inputs) > len(manager.pis) {
		return false, ErrDimension
	}
	if manager.stopped() {
		return false, ErrCancelled
	}
	values := make([]bool, len(manager.objects))
	values[0] = true
	for i, edge := range manager.pis[:len(inputs)] {
		values[edge.ID()] = inputs[i]
	}
	for id := 1; id < len(manager.objects); id++ {
		object := manager.objects[id]
		if !object.Alive {
			continue
		}
		value := func(edge HopEdge) bool {
			v := values[edge.ID()]
			if edge.Complemented() {
				return !v
			}
			return v
		}
		switch object.Type {
		case HopObjectAnd:
			values[id] = value(object.Fanin0) && value(object.Fanin1)
		case HopObjectXor:
			values[id] = value(object.Fanin0) != value(object.Fanin1)
		case HopObjectPO:
			values[id] = value(object.Fanin0)
		}
	}
	result := values[root.ID()]
	if root.Complemented() {
		result = !result
	}
	return result, nil
}

func (manager *HopManager) Truth(root HopEdge, variables int, msbFirst bool) ([]uint64, error) {
	// Hop_TruthWordNum uses a 32-bit word count.  Reject shifts outside that
	// domain before computing 1<<variables, even when a caller deliberately
	// configured a manager with more PIs.
	if variables < 0 || variables > 30 || variables > len(manager.pis) || variables > manager.limits.MaxVars {
		return nil, ErrDimension
	}
	if !manager.valid(root) {
		return nil, ErrDimension
	}
	maximumPI, err := manager.maximumPIVariable(root)
	if err != nil || maximumPI >= variables {
		return nil, ErrDimension
	}
	bitsCount := 1 << variables
	wordCount := max(1, (bitsCount+63)/64)
	words := make([]uint64, wordCount)
	inputs := make([]bool, variables)
	for assignment := 0; assignment < bitsCount; assignment++ {
		if assignment&1023 == 0 && manager.stopped() {
			return nil, ErrCancelled
		}
		for variable := 0; variable < variables; variable++ {
			bit := variable
			if msbFirst {
				bit = variables - 1 - variable
			}
			inputs[variable] = assignment&(1<<bit) != 0
		}
		value, err := manager.Evaluate(root, inputs)
		if err != nil {
			return nil, err
		}
		if value {
			words[assignment>>6] |= uint64(1) << (assignment & 63)
		}
	}
	if bitsCount < 64 {
		pattern := words[0] & ((uint64(1) << bitsCount) - 1)
		for shift := bitsCount; shift < 64; shift += bitsCount {
			words[0] |= pattern << shift
		}
	}
	return words, nil
}

func (manager *HopManager) Truth6(root HopEdge, variables int) (uint64, error) {
	if variables < 0 || variables > 6 {
		return 0, ErrDimension
	}
	truth, err := manager.Truth(root, variables, false)
	if err != nil {
		return 0, err
	}
	if len(truth) == 0 {
		return 0, nil
	}
	return truth[0], nil
}

func (manager *HopManager) TruthWords32(root HopEdge, variables int, msbFirst bool) ([]uint32, error) {
	truth, err := manager.Truth(root, variables, msbFirst)
	if err != nil {
		return nil, err
	}
	count, err := HopTruthWordCount(variables)
	if err != nil {
		return nil, err
	}
	words := make([]uint32, count)
	for i := range words {
		word := truth[i/2]
		if i&1 != 0 {
			word >>= 32
		}
		words[i] = uint32(word)
	}
	return words, nil
}

func (manager *HopManager) maximumPIVariable(root HopEdge) (int, error) {
	if !manager.valid(root) {
		return -1, ErrDimension
	}
	seen := make([]bool, len(manager.objects))
	maximum := -1
	var visit func(HopEdge)
	visit = func(edge HopEdge) {
		id := edge.ID()
		if seen[id] {
			return
		}
		seen[id] = true
		object := manager.objects[id]
		if object.Type == HopObjectPI {
			maximum = max(maximum, object.PIONum)
		}
		if object.Type == HopObjectPO {
			visit(object.Fanin0)
		}
		if object.Type == HopObjectAnd || object.Type == HopObjectXor {
			visit(object.Fanin0)
			visit(object.Fanin1)
		}
	}
	visit(root)
	return maximum, nil
}

func (manager *HopManager) Compose(root, function HopEdge, variable int) (HopEdge, error) {
	if !manager.valid(root) || !manager.valid(function) || variable < 0 || variable >= len(manager.pis) {
		return 0, ErrDimension
	}
	memo := make(map[HopEdge]HopEdge)
	return manager.atomicCompound(func() (HopEdge, error) {
		var rec func(HopEdge) (HopEdge, error)
		rec = func(edge HopEdge) (HopEdge, error) {
			regular := edge.Regular()
			if found, ok := memo[regular]; ok {
				return found.NotIf(edge.Complemented()), nil
			}
			object := manager.objects[regular.ID()]
			var result HopEdge
			var err error
			switch object.Type {
			case HopObjectConst1:
				result = manager.Const1()
			case HopObjectPI:
				if object.PIONum == variable {
					result = function
				} else {
					result = regular
				}
			case HopObjectAnd, HopObjectXor:
				first, e := rec(object.Fanin0)
				if e != nil {
					return 0, e
				}
				second, e := rec(object.Fanin1)
				if e != nil {
					return 0, e
				}
				if object.Type == HopObjectXor {
					result, err = manager.Xor(first, second)
				} else {
					result, err = manager.And(first, second)
				}
			default:
				return 0, ErrDimension
			}
			if err != nil {
				return 0, err
			}
			memo[regular] = result
			return result.NotIf(edge.Complemented()), nil
		}
		return rec(root)
	})
}

func (manager *HopManager) ComplementVariable(root HopEdge, variable int) (HopEdge, error) {
	if variable < 0 || variable >= len(manager.pis) {
		return 0, ErrDimension
	}
	return manager.Compose(root, manager.pis[variable].Not(), variable)
}

func (manager *HopManager) Permute(root HopEdge, permutation []int) (HopEdge, error) {
	if !manager.valid(root) || len(permutation) > len(manager.pis) {
		return 0, ErrDimension
	}
	maximumPI, err := manager.maximumPIVariable(root)
	if err != nil || maximumPI >= len(permutation) {
		return 0, ErrDimension
	}
	for _, target := range permutation {
		if target < 0 || target >= len(manager.pis) {
			return 0, ErrDimension
		}
	}
	originalObjects, originalPIs := len(manager.objects), len(manager.pis)
	memo := make(map[HopEdge]HopEdge)
	result, err := manager.atomicCompound(func() (HopEdge, error) {
		var rec func(HopEdge) (HopEdge, error)
		rec = func(edge HopEdge) (HopEdge, error) {
			regular := edge.Regular()
			if value, ok := memo[regular]; ok {
				return value.NotIf(edge.Complemented()), nil
			}
			object := manager.objects[regular.ID()]
			result := regular
			var err error
			if object.Type == HopObjectPI {
				result = manager.pis[permutation[object.PIONum]]
			}
			if object.Type == HopObjectAnd || object.Type == HopObjectXor {
				first, e := rec(object.Fanin0)
				if e != nil {
					return 0, e
				}
				second, e := rec(object.Fanin1)
				if e != nil {
					return 0, e
				}
				if object.Type == HopObjectXor {
					result, err = manager.Xor(first, second)
				} else {
					result, err = manager.And(first, second)
				}
			}
			if err != nil {
				return 0, err
			}
			memo[regular] = result
			return result.NotIf(edge.Complemented()), nil
		}
		return rec(root)
	})
	if err != nil {
		manager.rollback(originalObjects)
		manager.pis = manager.pis[:originalPIs]
	}
	return result, err
}

func (manager *HopManager) Remap(root HopEdge, support uint32, variableCount int) (HopEdge, error) {
	if !manager.valid(root) || variableCount < 0 || variableCount > len(manager.pis) || variableCount > 32 {
		return 0, ErrDimension
	}
	if support == 0 {
		return manager.Const0().NotIf(manager.edgePhase(root)), nil
	}
	mapping := make([]HopEdge, variableCount)
	next := 0
	for i := 0; i < variableCount; i++ {
		if support&(uint32(1)<<i) != 0 {
			mapping[i] = manager.pis[next]
			next++
		} else {
			mapping[i] = manager.Const0()
		}
	}
	memo := make(map[HopEdge]HopEdge)
	return manager.atomicCompound(func() (HopEdge, error) {
		var rec func(HopEdge) (HopEdge, error)
		rec = func(edge HopEdge) (HopEdge, error) {
			regular := edge.Regular()
			if value, ok := memo[regular]; ok {
				return value.NotIf(edge.Complemented()), nil
			}
			object := manager.objects[regular.ID()]
			result := regular
			var err error
			if object.Type == HopObjectPI && object.PIONum < variableCount {
				result = mapping[object.PIONum]
			}
			if object.Type == HopObjectAnd || object.Type == HopObjectXor {
				first, e := rec(object.Fanin0)
				if e != nil {
					return 0, e
				}
				second, e := rec(object.Fanin1)
				if e != nil {
					return 0, e
				}
				if object.Type == HopObjectXor {
					result, err = manager.Xor(first, second)
				} else {
					result, err = manager.And(first, second)
				}
			}
			if err != nil {
				return 0, err
			}
			memo[regular] = result
			return result.NotIf(edge.Complemented()), nil
		}
		return rec(root)
	})
}

func (manager *HopManager) TransferTo(destination *HopManager, root HopEdge, variables int) (HopEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if err := destination.check(); err != nil {
		return 0, err
	}
	if !manager.valid(root) || variables < 0 || variables > len(manager.pis) {
		return 0, ErrDimension
	}
	maximumPI, err := manager.maximumPIVariable(root)
	if err != nil || maximumPI >= variables {
		return 0, ErrDimension
	}
	originalObjects, originalPIs := len(destination.objects), len(destination.pis)
	for i := 0; i < variables; i++ {
		if _, err := destination.IthVar(i); err != nil {
			destination.rollback(originalObjects)
			destination.pis = destination.pis[:originalPIs]
			return 0, err
		}
	}
	memo := map[HopEdge]HopEdge{manager.Const1(): destination.Const1()}
	// Hop_Transfer maps only the caller-selected PI prefix.  Mapping all source
	// PIs would index beyond destination.pis when the source manager has extra,
	// unused variables.
	for i, edge := range manager.pis[:variables] {
		memo[edge] = destination.pis[i]
	}
	result, err := destination.atomicCompound(func() (HopEdge, error) {
		var rec func(HopEdge) (HopEdge, error)
		rec = func(edge HopEdge) (HopEdge, error) {
			regular := edge.Regular()
			if value, ok := memo[regular]; ok {
				return value.NotIf(edge.Complemented()), nil
			}
			object := manager.objects[regular.ID()]
			first, err := rec(object.Fanin0)
			if err != nil {
				return 0, err
			}
			second, err := rec(object.Fanin1)
			if err != nil {
				return 0, err
			}
			var result HopEdge
			if object.Type == HopObjectXor {
				result, err = destination.Xor(first, second)
			} else {
				result, err = destination.And(first, second)
			}
			if err != nil {
				return 0, err
			}
			memo[regular] = result
			return result.NotIf(edge.Complemented()), nil
		}
		return rec(root)
	})
	if err != nil {
		destination.rollback(originalObjects)
		destination.pis = destination.pis[:originalPIs]
	}
	return result, err
}

func (manager *HopManager) Balance(root HopEdge) (HopEdge, error) {
	if !manager.valid(root) {
		return 0, ErrDimension
	}
	leaves := make([]HopEdge, 0)
	polarities := make(map[HopEdge]bool)
	contradiction := false
	appendLeaf := func(edge HopEdge) {
		regular := edge.Regular()
		if complemented, exists := polarities[regular]; exists {
			if complemented != edge.Complemented() {
				contradiction = true
			}
			return
		}
		polarities[regular] = edge.Complemented()
		leaves = append(leaves, edge)
	}
	var collect func(HopEdge)
	collect = func(edge HopEdge) {
		object := manager.objects[edge.ID()]
		if !edge.Complemented() && object.Type == HopObjectAnd {
			collect(object.Fanin0)
			collect(object.Fanin1)
			return
		}
		appendLeaf(edge)
	}
	collect(root)
	if contradiction {
		return manager.Const0(), nil
	}
	slices.SortFunc(leaves, func(a, b HopEdge) int { return manager.objects[a.ID()].Level - manager.objects[b.ID()].Level })
	return manager.Multi(leaves, false)
}

// Balanced rebuilds the complete manager, preserving PI/PO order and stopping
// supergate collection at shared fanout boundaries as Hop_ManBalance does.
func (manager *HopManager) Balanced(updateLevels bool) (*HopManager, error) {
	_ = updateLevels
	if err := manager.check(); err != nil {
		return nil, err
	}
	if manager.stopped() {
		return nil, ErrCancelled
	}
	result, err := NewHopManager(manager.limits)
	if err != nil {
		return nil, err
	}
	for i := range manager.pis {
		if _, err = result.IthVar(i); err != nil {
			return nil, err
		}
	}
	refs, err := manager.References()
	if err != nil {
		return nil, err
	}
	memo := map[HopEdge]HopEdge{manager.Const1(): result.Const1()}
	for i, edge := range manager.pis {
		memo[edge] = result.pis[i]
	}
	var rec func(HopEdge) (HopEdge, error)
	rec = func(edge HopEdge) (HopEdge, error) {
		regular := edge.Regular()
		if mapped, ok := memo[regular]; ok {
			return mapped.NotIf(edge.Complemented()), nil
		}
		object := manager.objects[regular.ID()]
		if object.Type != HopObjectAnd && object.Type != HopObjectXor {
			return 0, ErrDimension
		}
		leaves := make([]HopEdge, 0)
		polarities := make(map[HopEdge]bool)
		contradiction := false
		appendLeaf := func(leaf HopEdge) {
			regularLeaf := leaf.Regular()
			if complemented, exists := polarities[regularLeaf]; exists {
				if complemented != leaf.Complemented() {
					contradiction = true
				}
				return
			}
			polarities[regularLeaf] = leaf.Complemented()
			leaves = append(leaves, leaf)
		}
		var collect func(HopEdge, bool)
		collect = func(current HopEdge, isRoot bool) {
			node := manager.objects[current.ID()]
			if current.Complemented() || node.Type != object.Type || (!isRoot && refs[current.ID()] > 1) {
				appendLeaf(current)
				return
			}
			collect(node.Fanin0, false)
			collect(node.Fanin1, false)
		}
		collect(regular, true)
		if contradiction {
			memo[regular] = result.Const0()
			return result.Const0().NotIf(edge.Complemented()), nil
		}
		mapped := make([]HopEdge, len(leaves))
		for i, leaf := range leaves {
			mapped[i], err = rec(leaf)
			if err != nil {
				return 0, err
			}
		}
		for len(mapped) > 1 {
			slices.SortFunc(mapped, func(a, b HopEdge) int {
				levelA, levelB := result.objects[a.ID()].Level, result.objects[b.ID()].Level
				if levelA != levelB {
					return levelA - levelB
				}
				return a.ID() - b.ID()
			})
			var combined HopEdge
			if object.Type == HopObjectXor {
				combined, err = result.Xor(mapped[0], mapped[1])
			} else {
				combined, err = result.And(mapped[0], mapped[1])
			}
			if err != nil {
				return 0, err
			}
			mapped = append(mapped[2:], combined)
		}
		built := mapped[0]
		memo[regular] = built
		return built.NotIf(edge.Complemented()), nil
	}
	for _, output := range manager.pos {
		driver, buildErr := rec(manager.objects[output.ID()].Fanin0)
		if buildErr != nil {
			return nil, buildErr
		}
		if _, buildErr = result.CreatePO(driver); buildErr != nil {
			return nil, buildErr
		}
	}
	return result, nil
}

func (manager *HopManager) Check() error {
	if err := manager.check(); err != nil {
		return err
	}
	for id, object := range manager.objects {
		if object.ID != id || !object.Alive {
			continue
		}
		if object.Type == HopObjectPI && (object.PIONum < 0 || object.PIONum >= len(manager.pis) || manager.pis[object.PIONum].ID() != id) {
			return ErrInvariant
		}
		if object.Type == HopObjectPI && (object.Fanin0 != 0 || object.Fanin1 != 0) {
			return ErrInvariant
		}
		if object.Type == HopObjectPO && (!manager.valid(object.Fanin0) || object.Fanin1 != 0 || object.PIONum < 0 || object.PIONum >= len(manager.pos) || manager.pos[object.PIONum].ID() != id) {
			return ErrInvariant
		}
		if object.Type == HopObjectAnd || object.Type == HopObjectXor {
			if !manager.valid(object.Fanin0) || !manager.valid(object.Fanin1) || object.Fanin0.ID() >= id || object.Fanin1.ID() >= id {
				return ErrInvariant
			}
			if manager.table[hopHashKey{object.Type, object.Fanin0, object.Fanin1}].ID() != id {
				return ErrInvariant
			}
			if object.Fanin0.ID() >= object.Fanin1.ID() || object.Level != max(manager.objects[object.Fanin0.ID()].Level, manager.objects[object.Fanin1.ID()].Level)+1 || object.Phase != (manager.edgePhase(object.Fanin0) && manager.edgePhase(object.Fanin1)) {
				return ErrInvariant
			}
		}
		if object.HasChoice && !manager.valid(object.Choice) {
			return ErrInvariant
		}
	}
	stats, _ := manager.Stats()
	if stats.HashEntries != stats.Ands+stats.Xors {
		return ErrInvariant
	}
	return nil
}

func (manager *HopManager) Cleanup() (int, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if manager.stopped() {
		return 0, ErrCancelled
	}
	reachable := make([]bool, len(manager.objects))
	reachable[0] = true
	for _, edge := range manager.pis {
		reachable[edge.ID()] = true
	}
	for _, output := range manager.pos {
		nodes, _ := manager.DFS(output)
		for _, edge := range nodes {
			reachable[edge.ID()] = true
		}
	}
	removed := 0
	for id := len(manager.objects) - 1; id > 0; id-- {
		object := &manager.objects[id]
		if object.Alive && !reachable[id] && (object.Type == HopObjectAnd || object.Type == HopObjectXor) {
			delete(manager.table, hopHashKey{object.Type, object.Fanin0, object.Fanin1})
			object.Alive = false
			object.Type = HopObjectNone
			removed++
		}
	}
	manager.deleted += removed
	return removed, nil
}

func (manager *HopManager) Stats() (HopStats, error) {
	if err := manager.check(); err != nil {
		return HopStats{}, err
	}
	stats := HopStats{PIs: len(manager.pis), POs: len(manager.pos), Created: len(manager.objects), Deleted: manager.deleted, HashEntries: len(manager.table)}
	for _, object := range manager.objects {
		if !object.Alive {
			continue
		}
		if object.Type == HopObjectAnd {
			stats.Ands++
		}
		if object.Type == HopObjectXor {
			stats.Xors++
		}
		stats.Levels = max(stats.Levels, object.Level)
	}
	stats.MemoryBytes = len(manager.objects)*64 + len(manager.table)*32 + (len(manager.pis)+len(manager.pos))*4
	return stats, nil
}

func (manager *HopManager) MemorySnapshot() ([]HopObject, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	return slices.Clone(manager.objects), nil
}

func (manager *HopManager) HashProfile() ([]int, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	profile := make([]int, 16)
	for _, object := range manager.objects {
		if object.Alive && (object.Type == HopObjectAnd || object.Type == HopObjectXor) {
			hash := uint32(object.Fanin0.ID())*7937 + uint32(object.Fanin1.ID())*2971
			if object.Fanin0.Complemented() {
				hash += 911
			}
			if object.Fanin1.Complemented() {
				hash += 353
			}
			if object.Type == HopObjectXor {
				hash += 1699
			}
			profile[hash&15]++
		}
	}
	return profile, nil
}

func (manager *HopManager) Equation(root HopEdge) (string, error) { return manager.format(root, false) }
func (manager *HopManager) Verilog(root HopEdge) (string, error)  { return manager.format(root, true) }
func (manager *HopManager) Verbose(root HopEdge) (string, error) {
	object, err := manager.Object(root)
	if err != nil {
		return "", err
	}
	return "id=" + strconv.Itoa(object.ID) + " type=" + strconv.Itoa(int(object.Type)) + " level=" + strconv.Itoa(object.Level) + " phase=" + strconv.FormatBool(manager.edgePhase(root)) + "\n", nil
}
func (manager *HopManager) ManagerVerbose() (string, error) {
	if err := manager.check(); err != nil {
		return "", err
	}
	stats, _ := manager.Stats()
	return "PI/PO = " + strconv.Itoa(stats.PIs) + "/" + strconv.Itoa(stats.POs) + ". A = " + strconv.Itoa(stats.Ands) + ". X = " + strconv.Itoa(stats.Xors) + ". Lev = " + strconv.Itoa(stats.Levels) + ".\n", nil
}
func (manager *HopManager) BLIF(root HopEdge, model string) (string, error) {
	if model == "" {
		return "", ErrDimension
	}
	nodes, err := manager.DFS(root)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString(".model ")
	builder.WriteString(model)
	builder.WriteString("\n.inputs")
	for i := range manager.pis {
		builder.WriteString(" x")
		builder.WriteString(strconv.Itoa(i))
	}
	builder.WriteString("\n.outputs z\n")
	name := func(edge HopEdge) string {
		object := manager.objects[edge.ID()]
		if object.Type == HopObjectConst1 {
			return "const1"
		}
		if object.Type == HopObjectPI {
			return "x" + strconv.Itoa(object.PIONum)
		}
		return "n" + strconv.Itoa(edge.ID())
	}
	for _, edge := range nodes {
		object := manager.objects[edge.ID()]
		builder.WriteString(".names ")
		builder.WriteString(name(object.Fanin0))
		builder.WriteByte(' ')
		builder.WriteString(name(object.Fanin1))
		builder.WriteByte(' ')
		builder.WriteString(name(edge))
		builder.WriteByte('\n')
		cube0, cube1 := '1', '1'
		if object.Fanin0.Complemented() {
			cube0 = '0'
		}
		if object.Fanin1.Complemented() {
			cube1 = '0'
		}
		builder.WriteRune(cube0)
		builder.WriteRune(cube1)
		builder.WriteString(" 1\n")
	}
	// The source dumper emits a constant node when a PO is constant.  Keeping
	// the regular constant name here lets the output cube carry the complement
	// exactly once (and makes both constant roots valid BLIF).
	if root.Regular() == manager.Const1() {
		builder.WriteString(".names const1\n1\n")
	}
	builder.WriteString(".names ")
	builder.WriteString(name(root))
	builder.WriteString(" z\n")
	if root.Complemented() {
		builder.WriteString("0 1\n")
	} else {
		builder.WriteString("1 1\n")
	}
	builder.WriteString(".end\n")
	return builder.String(), nil
}
func (manager *HopManager) format(root HopEdge, verilog bool) (string, error) {
	if !manager.valid(root) {
		return "", ErrDimension
	}
	memo := make(map[HopEdge]string)
	var rec func(HopEdge) string
	rec = func(edge HopEdge) string {
		regular := edge.Regular()
		text, ok := memo[regular]
		if !ok {
			object := manager.objects[regular.ID()]
			switch object.Type {
			case HopObjectConst1:
				text = "1"
				if verilog {
					text = "1'b1"
				}
			case HopObjectPI:
				text = "x" + strconv.Itoa(object.PIONum)
			case HopObjectAnd:
				text = "(" + rec(object.Fanin0) + " & " + rec(object.Fanin1) + ")"
			case HopObjectXor:
				text = "(" + rec(object.Fanin0) + " ^ " + rec(object.Fanin1) + ")"
			}
			memo[regular] = text
		}
		if edge.Complemented() {
			if verilog {
				return "~" + text
			}
			return "!" + text
		}
		return text
	}
	return rec(root), nil
}

func (manager *HopManager) CreateChoice(old, replacement HopEdge) error {
	if !manager.valid(old) || !manager.valid(replacement) {
		return ErrDimension
	}
	if manager.stopped() {
		return ErrCancelled
	}
	representative, err := manager.Representative(old)
	if err != nil {
		return err
	}
	newRepresentative, err := manager.Representative(replacement)
	if err != nil {
		return err
	}
	if representative.Regular() == newRepresentative.Regular() {
		return nil
	}
	object := &manager.objects[newRepresentative.ID()]
	object.Choice = representative.NotIf(newRepresentative.Complemented())
	object.HasChoice = true
	return nil
}
func (manager *HopManager) Representative(edge HopEdge) (HopEdge, error) {
	if !manager.valid(edge) {
		return 0, ErrDimension
	}
	complemented := edge.Complemented()
	regular := edge.Regular()
	seen := make([]bool, len(manager.objects))
	for {
		if seen[regular.ID()] {
			return 0, ErrInvariant
		}
		seen[regular.ID()] = true
		object := manager.objects[regular.ID()]
		if !object.HasChoice {
			return regular.NotIf(complemented), nil
		}
		complemented = complemented != object.Choice.Complemented()
		regular = object.Choice.Regular()
	}
}
func (manager *HopManager) IncrementTraversalID() (int, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if manager.travID == math.MaxInt {
		manager.travID = 1
	} else {
		manager.travID++
	}
	return manager.travID, nil
}
func (manager *HopManager) Close() error {
	if manager == nil || manager.closed {
		return nil
	}
	manager.objects = nil
	manager.pis = nil
	manager.pos = nil
	manager.table = nil
	manager.closed = true
	return nil
}

func HopBitWordCount(bitCount int) (int, error) {
	if bitCount < 0 {
		return 0, ErrDimension
	}
	return (bitCount + 31) / 32, nil
}
func HopTruthWordCount(variableCount int) (int, error) {
	if variableCount < 0 || variableCount > 30 {
		return 0, ErrDimension
	}
	if variableCount <= 5 {
		return 1, nil
	}
	return 1 << (variableCount - 5), nil
}
func HopBase2Log(value uint) int {
	if value < 2 {
		return int(value)
	}
	return bits.Len(value - 1)
}
func HopBase10Log(value uint) int {
	if value < 2 {
		return int(value)
	}
	result := 0
	for value--; value != 0; value /= 10 {
		result++
	}
	return result
}

type Cudd2Adapter struct {
	manager      *HopManager
	variables    int
	slots, cache int
	maxMemory    uint64
	bindings     map[uint64]HopEdge
	closed       bool
}

func NewCudd2Adapter(variables, zVariables, slots, cache int, maxMemory uint64, limits HopLimits) (*Cudd2Adapter, error) {
	if variables < 0 || zVariables < 0 || slots < 0 || cache < 0 {
		return nil, ErrDimension
	}
	manager, err := NewHopManager(limits)
	if err != nil {
		return nil, err
	}
	for i := 0; i < variables; i++ {
		if _, err = manager.IthVar(i); err != nil {
			return nil, err
		}
	}
	return &Cudd2Adapter{manager: manager, variables: variables, slots: slots, cache: cache, maxMemory: maxMemory, bindings: make(map[uint64]HopEdge)}, nil
}

type Cudd2Reference struct {
	Identity     uint64
	Complemented bool
}

func (adapter *Cudd2Adapter) Register(reference Cudd2Reference, edge HopEdge) error {
	if err := adapter.check(); err != nil {
		return err
	}
	if reference.Identity == 0 || !adapter.manager.valid(edge) {
		return ErrDimension
	}
	if _, exists := adapter.bindings[reference.Identity]; exists {
		return nil
	}
	adapter.bindings[reference.Identity] = edge.NotIf(reference.Complemented)
	return nil
}
func (adapter *Cudd2Adapter) Resolve(reference Cudd2Reference) (HopEdge, error) {
	if err := adapter.check(); err != nil {
		return 0, err
	}
	edge, exists := adapter.bindings[reference.Identity]
	if !exists {
		return 0, ErrDimension
	}
	return edge.NotIf(reference.Complemented), nil
}
func (adapter *Cudd2Adapter) check() error {
	if adapter == nil || adapter.closed {
		return ErrInvariant
	}
	return adapter.manager.check()
}
func (adapter *Cudd2Adapter) One() (HopEdge, error) {
	if err := adapter.check(); err != nil {
		return 0, err
	}
	return adapter.manager.Const1(), nil
}
func (adapter *Cudd2Adapter) IthVar(index int) (HopEdge, error) {
	if err := adapter.check(); err != nil {
		return 0, err
	}
	return adapter.manager.IthVar(index)
}
func (adapter *Cudd2Adapter) And(a, b HopEdge) (HopEdge, error) {
	if err := adapter.check(); err != nil {
		return 0, err
	}
	return adapter.manager.And(a, b)
}
func (adapter *Cudd2Adapter) Or(a, b HopEdge) (HopEdge, error) {
	if err := adapter.check(); err != nil {
		return 0, err
	}
	return adapter.manager.Or(a, b)
}
func (adapter *Cudd2Adapter) Nand(a, b HopEdge) (HopEdge, error) {
	v, e := adapter.And(a, b)
	return v.Not(), e
}
func (adapter *Cudd2Adapter) Nor(a, b HopEdge) (HopEdge, error) {
	v, e := adapter.Or(a, b)
	return v.Not(), e
}
func (adapter *Cudd2Adapter) Xor(a, b HopEdge) (HopEdge, error) {
	if err := adapter.check(); err != nil {
		return 0, err
	}
	return adapter.manager.Xor(a, b)
}
func (adapter *Cudd2Adapter) Xnor(a, b HopEdge) (HopEdge, error) {
	v, e := adapter.Xor(a, b)
	return v.Not(), e
}
func (adapter *Cudd2Adapter) ITE(c, t, e HopEdge) (HopEdge, error) {
	if err := adapter.check(); err != nil {
		return 0, err
	}
	return adapter.manager.Mux(c, t, e)
}
func (adapter *Cudd2Adapter) Compose(root, function HopEdge, variable int) (HopEdge, error) {
	if err := adapter.check(); err != nil {
		return 0, err
	}
	return adapter.manager.Compose(root, function, variable)
}
func (adapter *Cudd2Adapter) Leq(a, b HopEdge) (bool, error) {
	if err := adapter.check(); err != nil {
		return false, err
	}
	size := len(adapter.manager.objects)
	violation, err := adapter.manager.And(a, b.Not())
	if err != nil {
		return false, err
	}
	ta, e := adapter.manager.Truth(a, len(adapter.manager.pis), false)
	if e != nil {
		adapter.manager.rollback(size)
		return false, e
	}
	tb, e := adapter.manager.Truth(b, len(adapter.manager.pis), false)
	if e != nil {
		adapter.manager.rollback(size)
		return false, e
	}
	if _, err = adapter.manager.CreatePO(violation); err != nil {
		adapter.manager.rollback(size)
		return false, err
	}
	for i := range ta {
		if ta[i]&^tb[i] != 0 {
			return false, nil
		}
	}
	return true, nil
}
func (adapter *Cudd2Adapter) Equal(a, b HopEdge) (bool, error) {
	if err := adapter.check(); err != nil {
		return false, err
	}
	size := len(adapter.manager.objects)
	difference, err := adapter.manager.Xor(a, b)
	if err != nil {
		return false, err
	}
	truth, err := adapter.manager.Truth(difference, len(adapter.manager.pis), false)
	if err != nil {
		adapter.manager.rollback(size)
		return false, err
	}
	if _, err = adapter.manager.CreatePO(difference); err != nil {
		adapter.manager.rollback(size)
		return false, err
	}
	for _, word := range truth {
		if word != 0 {
			return false, nil
		}
	}
	return true, nil
}
func (adapter *Cudd2Adapter) Manager() (*HopManager, error) {
	if err := adapter.check(); err != nil {
		return nil, err
	}
	return adapter.manager, nil
}
func (adapter *Cudd2Adapter) Close() error {
	if adapter == nil || adapter.closed {
		return nil
	}
	adapter.closed = true
	adapter.bindings = nil
	return adapter.manager.Close()
}
