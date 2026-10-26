package abc

// Canonical zero-suppressed decision diagrams from src/bdd/cudd/cuddZddFuncs.c,
// cuddZddSetop.c, cuddZddCount.c, cuddZddMisc.c, cuddZddPort.c, and cuddZddIsop.c.
// This concrete manager has a fixed ascending variable order and a 63-variable
// bitset boundary. Dynamic reordering, C reference-count/ABI compatibility,
// subtable printing, and paired-positive/negative-literal cover algorithms are
// intentionally outside it. ISOP therefore returns a correct full-minterm
// cover, not CUDD's order-dependent irredundant paired-literal cover.

import (
	"math"
	"sort"
	"sync"
)

type CuddZDDEdge uint32

const (
	CuddZDDEmpty CuddZDDEdge = 0
	CuddZDDBase  CuddZDDEdge = 1
)

type CuddZDDOptions struct {
	MaxOperations int
	MaxBytes      int
	Stop          func() bool
}

type CuddZDDStatistics struct {
	Variables  int
	Nodes      int
	Capacity   int
	Operations uint64
}

type cuddZDDNode struct {
	variable  uint8
	high, low CuddZDDEdge
}

type cuddZDDKey struct {
	variable  uint8
	high, low CuddZDDEdge
}

type CuddZDDManager struct {
	mu         sync.Mutex
	variables  int
	capacity   int
	nodes      []cuddZDDNode
	unique     map[cuddZDDKey]CuddZDDEdge
	operations uint64
	closed     bool
}

type cuddZDDWork struct {
	options CuddZDDOptions
	steps   int
	bytes   int
}

func (work *cuddZDDWork) step() error {
	work.steps++
	if work.steps > work.options.MaxOperations {
		return ErrLimit
	}
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *cuddZDDWork) addSet(family map[uint64]struct{}, set uint64) error {
	if err := work.step(); err != nil {
		return err
	}
	if _, found := family[set]; found {
		return nil
	}
	if work.bytes > work.options.MaxBytes-16 {
		return ErrLimit
	}
	work.bytes += 16
	family[set] = struct{}{}
	return nil
}

func normalizeCuddZDDOptions(options CuddZDDOptions) (CuddZDDOptions, error) {
	if options.MaxOperations < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 26
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	return options, nil
}

func NewCuddZDDManager(variables, capacity int) (*CuddZDDManager, error) {
	if variables < 0 || variables > 63 || capacity < 2 || capacity > math.MaxInt32 {
		return nil, ErrDimension
	}
	return &CuddZDDManager{
		variables: variables,
		capacity:  capacity,
		nodes: []cuddZDDNode{
			{variable: 0xff}, {variable: 0xff},
		},
		unique: make(map[cuddZDDKey]CuddZDDEdge),
	}, nil
}

func (manager *CuddZDDManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed || manager.variables < 0 || manager.variables > 63 || manager.capacity < 2 || len(manager.nodes) < 2 || len(manager.nodes) > manager.capacity || manager.unique == nil {
		return ErrInvariant
	}
	return nil
}

func (manager *CuddZDDManager) valid(edge CuddZDDEdge) bool {
	return uint64(edge) < uint64(len(manager.nodes))
}

func (manager *CuddZDDManager) makeNode(variable int, high, low CuddZDDEdge, work *cuddZDDWork) (CuddZDDEdge, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if high == CuddZDDEmpty {
		return low, nil
	}
	key := cuddZDDKey{variable: uint8(variable), high: high, low: low}
	if edge, found := manager.unique[key]; found {
		return edge, nil
	}
	if len(manager.nodes) >= manager.capacity || (len(manager.nodes)+1)*16 > work.options.MaxBytes {
		return 0, ErrLimit
	}
	edge := CuddZDDEdge(len(manager.nodes))
	manager.nodes = append(manager.nodes, cuddZDDNode{variable: uint8(variable), high: high, low: low})
	manager.unique[key] = edge
	return edge, nil
}

func (manager *CuddZDDManager) rollback(count int) {
	for edge := count; edge < len(manager.nodes); edge++ {
		node := manager.nodes[edge]
		delete(manager.unique, cuddZDDKey{variable: node.variable, high: node.high, low: node.low})
	}
	manager.nodes = manager.nodes[:count]
}

func (manager *CuddZDDManager) decode(edge CuddZDDEdge, work *cuddZDDWork, memo map[CuddZDDEdge]map[uint64]struct{}) (map[uint64]struct{}, error) {
	if err := work.step(); err != nil {
		return nil, err
	}
	if family, found := memo[edge]; found {
		return family, nil
	}
	family := make(map[uint64]struct{})
	if edge == CuddZDDBase {
		if err := work.addSet(family, 0); err != nil {
			return nil, err
		}
	} else if edge != CuddZDDEmpty {
		node := manager.nodes[edge]
		low, err := manager.decode(node.low, work, memo)
		if err != nil {
			return nil, err
		}
		high, err := manager.decode(node.high, work, memo)
		if err != nil {
			return nil, err
		}
		for set := range low {
			if err := work.addSet(family, set); err != nil {
				return nil, err
			}
		}
		bit := uint64(1) << node.variable
		for set := range high {
			if err := work.addSet(family, set|bit); err != nil {
				return nil, err
			}
		}
	}
	memo[edge] = family
	return family, nil
}

func (manager *CuddZDDManager) build(family map[uint64]struct{}, variable int, work *cuddZDDWork) (CuddZDDEdge, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if len(family) == 0 {
		return CuddZDDEmpty, nil
	}
	if variable == manager.variables {
		if _, found := family[0]; found {
			return CuddZDDBase, nil
		}
		return CuddZDDEmpty, nil
	}
	low := make(map[uint64]struct{})
	high := make(map[uint64]struct{})
	bit := uint64(1) << variable
	for set := range family {
		if set&bit == 0 {
			if err := work.addSet(low, set); err != nil {
				return 0, err
			}
		} else {
			if err := work.addSet(high, set&^bit); err != nil {
				return 0, err
			}
		}
	}
	highEdge, err := manager.build(high, variable+1, work)
	if err != nil {
		return 0, err
	}
	lowEdge, err := manager.build(low, variable+1, work)
	if err != nil {
		return 0, err
	}
	return manager.makeNode(variable, highEdge, lowEdge, work)
}

func (manager *CuddZDDManager) run(roots []CuddZDDEdge, options CuddZDDOptions, operation func([]map[uint64]struct{}, *cuddZDDWork) (map[uint64]struct{}, error)) (CuddZDDEdge, error) {
	if manager == nil {
		return 0, ErrDimension
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return 0, err
	}
	for _, root := range roots {
		if !manager.valid(root) {
			return 0, ErrInvalidLiteral
		}
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return 0, err
	}
	work := &cuddZDDWork{options: normalized}
	families := make([]map[uint64]struct{}, len(roots))
	memo := make(map[CuddZDDEdge]map[uint64]struct{})
	for index, root := range roots {
		families[index], err = manager.decode(root, work, memo)
		if err != nil {
			return 0, err
		}
	}
	resultFamily, err := operation(families, work)
	if err != nil {
		return 0, err
	}
	oldCount := len(manager.nodes)
	result, err := manager.build(resultFamily, 0, work)
	if err != nil {
		manager.rollback(oldCount)
		return 0, err
	}
	manager.operations += uint64(work.steps)
	return result, nil
}

func (manager *CuddZDDManager) FromBitsets(sets []uint64, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.run(nil, options, func(_ []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		family := make(map[uint64]struct{}, len(sets))
		limit := uint64(1)<<manager.variables - 1
		for _, set := range sets {
			if set&^limit != 0 {
				return nil, ErrDimension
			}
			if err := work.addSet(family, set); err != nil {
				return nil, err
			}
		}
		return family, nil
	})
}

func (manager *CuddZDDManager) ToBitsets(root CuddZDDEdge, options CuddZDDOptions) ([]uint64, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return nil, err
	}
	if !manager.valid(root) {
		return nil, ErrInvalidLiteral
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return nil, err
	}
	work := &cuddZDDWork{options: normalized}
	family, err := manager.decode(root, work, make(map[CuddZDDEdge]map[uint64]struct{}))
	if err != nil {
		return nil, err
	}
	sets := make([]uint64, 0, len(family))
	for set := range family {
		sets = append(sets, set)
	}
	sort.Slice(sets, func(i, j int) bool { return sets[i] < sets[j] })
	manager.operations += uint64(work.steps)
	return sets, nil
}

func (manager *CuddZDDManager) binary(first, second CuddZDDEdge, options CuddZDDOptions, operation func(map[uint64]struct{}, map[uint64]struct{}, *cuddZDDWork) (map[uint64]struct{}, error)) (CuddZDDEdge, error) {
	return manager.run([]CuddZDDEdge{first, second}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		return operation(f[0], f[1], work)
	})
}

func (manager *CuddZDDManager) Union(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.binary(first, second, options, func(a, b map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		r := make(map[uint64]struct{}, len(a)+len(b))
		for set := range a {
			if err := work.addSet(r, set); err != nil {
				return nil, err
			}
		}
		for set := range b {
			if err := work.addSet(r, set); err != nil {
				return nil, err
			}
		}
		return r, nil
	})
}

func (manager *CuddZDDManager) Intersect(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.binary(first, second, options, func(a, b map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		r := make(map[uint64]struct{})
		for set := range a {
			if _, ok := b[set]; ok {
				if err := work.addSet(r, set); err != nil {
					return nil, err
				}
			}
		}
		return r, nil
	})
}

func (manager *CuddZDDManager) Difference(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.binary(first, second, options, func(a, b map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		r := make(map[uint64]struct{})
		for set := range a {
			if _, ok := b[set]; !ok {
				if err := work.addSet(r, set); err != nil {
					return nil, err
				}
			}
		}
		return r, nil
	})
}

func (manager *CuddZDDManager) DifferenceEmpty(first, second CuddZDDEdge, options CuddZDDOptions) (bool, error) {
	result, err := manager.Difference(first, second, options)
	return result == CuddZDDEmpty, err
}

func (manager *CuddZDDManager) Ite(condition, then, otherwise CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.run([]CuddZDDEdge{condition, then, otherwise}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		r := make(map[uint64]struct{})
		for set := range f[1] {
			if _, ok := f[0][set]; ok {
				if err := work.addSet(r, set); err != nil {
					return nil, err
				}
			}
		}
		for set := range f[2] {
			if _, ok := f[0][set]; !ok {
				if err := work.addSet(r, set); err != nil {
					return nil, err
				}
			}
		}
		return r, nil
	})
}

func (manager *CuddZDDManager) Change(root CuddZDDEdge, variable int, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.run([]CuddZDDEdge{root}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		if variable < 0 || variable >= manager.variables {
			return nil, ErrDimension
		}
		r := make(map[uint64]struct{}, len(f[0]))
		bit := uint64(1) << variable
		for set := range f[0] {
			if err := work.addSet(r, set^bit); err != nil {
				return nil, err
			}
		}
		return r, nil
	})
}

func (manager *CuddZDDManager) Subset1(root CuddZDDEdge, variable int, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.run([]CuddZDDEdge{root}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		if variable < 0 || variable >= manager.variables {
			return nil, ErrDimension
		}
		r := make(map[uint64]struct{})
		bit := uint64(1) << variable
		for set := range f[0] {
			if set&bit != 0 {
				if err := work.addSet(r, set&^bit); err != nil {
					return nil, err
				}
			}
		}
		return r, nil
	})
}

func (manager *CuddZDDManager) Subset0(root CuddZDDEdge, variable int, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.run([]CuddZDDEdge{root}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		if variable < 0 || variable >= manager.variables {
			return nil, ErrDimension
		}
		r := make(map[uint64]struct{})
		bit := uint64(1) << variable
		for set := range f[0] {
			if set&bit == 0 {
				if err := work.addSet(r, set); err != nil {
					return nil, err
				}
			}
		}
		return r, nil
	})
}

func (manager *CuddZDDManager) Product(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.run([]CuddZDDEdge{first, second}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		r := make(map[uint64]struct{})
		for a := range f[0] {
			for b := range f[1] {
				if err := work.step(); err != nil {
					return nil, err
				}
				if err := work.addSet(r, a|b); err != nil {
					return nil, err
				}
			}
		}
		return r, nil
	})
}

func (manager *CuddZDDManager) UnateProduct(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.Product(first, second, options)
}

// Divide returns all cubes q for which every cube q union g, g in divisor,
// occurs in dividend. This is the standard unate-cover quotient definition.
func (manager *CuddZDDManager) Divide(dividend, divisor CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.run([]CuddZDDEdge{dividend, divisor}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		if len(f[1]) == 0 {
			return map[uint64]struct{}{}, nil
		}
		var candidates map[uint64]struct{}
		for g := range f[1] {
			current := make(map[uint64]struct{})
			for cube := range f[0] {
				if err := work.step(); err != nil {
					return nil, err
				}
				if cube&g == g {
					if err := work.addSet(current, cube&^g); err != nil {
						return nil, err
					}
				}
			}
			if candidates == nil {
				candidates = current
			} else {
				for q := range candidates {
					if _, ok := current[q]; !ok {
						delete(candidates, q)
					}
				}
			}
		}
		return candidates, nil
	})
}

func (manager *CuddZDDManager) WeakDivide(dividend, divisor CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.Divide(dividend, divisor, options)
}

// DivideF retains the historical C entry point. In this fixed-order unate
// manager it has the same mathematical quotient as Divide.
func (manager *CuddZDDManager) DivideF(dividend, divisor CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.Divide(dividend, divisor, options)
}

// WeakDivideF retains the historical C entry point. Paired-literal weak
// division is not represented; for unate families it is the same as Divide.
func (manager *CuddZDDManager) WeakDivideF(dividend, divisor CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.Divide(dividend, divisor, options)
}

func (manager *CuddZDDManager) Count(root CuddZDDEdge, options CuddZDDOptions) (uint64, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return 0, err
	}
	if !manager.valid(root) {
		return 0, ErrInvalidLiteral
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return 0, err
	}
	work := &cuddZDDWork{options: normalized}
	memo := make(map[CuddZDDEdge]uint64)
	var visit func(CuddZDDEdge) (uint64, error)
	visit = func(edge CuddZDDEdge) (uint64, error) {
		if err := work.step(); err != nil {
			return 0, err
		}
		if edge == CuddZDDEmpty {
			return 0, nil
		}
		if edge == CuddZDDBase {
			return 1, nil
		}
		if count, found := memo[edge]; found {
			return count, nil
		}
		node := manager.nodes[edge]
		high, err := visit(node.high)
		if err != nil {
			return 0, err
		}
		low, err := visit(node.low)
		if err != nil {
			return 0, err
		}
		if high > math.MaxUint64-low || work.bytes > normalized.MaxBytes-16 {
			return 0, ErrLimit
		}
		work.bytes += 16
		memo[edge] = high + low
		return high + low, nil
	}
	count, err := visit(root)
	if err != nil {
		return 0, err
	}
	manager.operations += uint64(work.steps)
	return count, nil
}

func (manager *CuddZDDManager) CountDouble(root CuddZDDEdge, options CuddZDDOptions) (float64, error) {
	count, err := manager.Count(root, options)
	return float64(count), err
}

func (manager *CuddZDDManager) CountMinterms(root CuddZDDEdge, dontCareVariables int, options CuddZDDOptions) (float64, error) {
	if dontCareVariables < 0 || dontCareVariables > 1023 {
		return 0, ErrDimension
	}
	count, err := manager.CountDouble(root, options)
	if err != nil {
		return 0, err
	}
	return math.Ldexp(count, -dontCareVariables), nil
}

func (manager *CuddZDDManager) Support(root CuddZDDEdge, options CuddZDDOptions) ([]int, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return nil, err
	}
	if !manager.valid(root) {
		return nil, ErrInvalidLiteral
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return nil, err
	}
	work := &cuddZDDWork{options: normalized}
	seen := make(map[CuddZDDEdge]struct{})
	mask := uint64(0)
	stack := []CuddZDDEdge{root}
	for len(stack) != 0 {
		if err := work.step(); err != nil {
			return nil, err
		}
		edge := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if edge <= CuddZDDBase {
			continue
		}
		if _, found := seen[edge]; found {
			continue
		}
		if work.bytes > normalized.MaxBytes-16 {
			return nil, ErrLimit
		}
		work.bytes += 16
		seen[edge] = struct{}{}
		node := manager.nodes[edge]
		mask |= uint64(1) << node.variable
		stack = append(stack, node.high, node.low)
	}
	result := make([]int, 0)
	for variable := 0; variable < manager.variables; variable++ {
		if mask&(uint64(1)<<variable) != 0 {
			result = append(result, variable)
		}
	}
	manager.operations += uint64(work.steps)
	return result, nil
}

func (manager *CuddZDDManager) DagSize(root CuddZDDEdge) (int, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return 0, err
	}
	if !manager.valid(root) {
		return 0, ErrInvalidLiteral
	}
	seen := make(map[CuddZDDEdge]struct{})
	var visit func(CuddZDDEdge)
	visit = func(edge CuddZDDEdge) {
		if edge <= CuddZDDBase {
			return
		}
		if _, ok := seen[edge]; ok {
			return
		}
		seen[edge] = struct{}{}
		n := manager.nodes[edge]
		visit(n.high)
		visit(n.low)
	}
	visit(root)
	return len(seen), nil
}

func (manager *CuddZDDManager) Statistics() (CuddZDDStatistics, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return CuddZDDStatistics{}, err
	}
	return CuddZDDStatistics{Variables: manager.variables, Nodes: len(manager.nodes), Capacity: manager.capacity, Operations: manager.operations}, nil
}

func (manager *CuddZDDManager) Close() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return err
	}
	manager.closed = true
	manager.nodes = nil
	manager.unique = nil
	return nil
}

func (manager *CuddZDDManager) PortToBDD(root CuddZDDEdge, bdd *ExtraBDDManager, options CuddZDDOptions) (ExtraBDDLit, error) {
	if bdd == nil {
		return 0, ErrDimension
	}
	if err := bdd.check(); err != nil {
		return 0, err
	}
	manager.mu.Lock()
	if err := manager.check(); err != nil {
		manager.mu.Unlock()
		return 0, err
	}
	if bdd.variables != manager.variables {
		manager.mu.Unlock()
		return 0, ErrDimension
	}
	if !manager.valid(root) {
		manager.mu.Unlock()
		return 0, ErrInvalidLiteral
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		manager.mu.Unlock()
		return 0, err
	}
	zddWork := &cuddZDDWork{options: normalized}
	family, err := manager.decode(root, zddWork, make(map[CuddZDDEdge]map[uint64]struct{}))
	if err != nil {
		manager.mu.Unlock()
		return 0, err
	}
	sets := make([]uint64, 0, len(family))
	for set := range family {
		sets = append(sets, set)
	}
	sort.Slice(sets, func(i, j int) bool { return sets[i] < sets[j] })
	manager.operations += uint64(zddWork.steps)
	variables := manager.variables
	manager.mu.Unlock()
	remaining := normalized.MaxOperations - zddWork.steps
	if len(sets) != 0 && remaining <= 0 {
		return 0, ErrLimit
	}
	transaction := bdd.beginTransaction()
	bddWork := extraBDDWork{
		options:     ExtraBDDOptions{MaxOperations: remaining, MaxBytes: normalized.MaxBytes, Stop: normalized.Stop},
		transaction: &transaction,
	}
	result := ExtraBDDFalse
	for _, set := range sets {
		cube := ExtraBDDTrue
		for variable := 0; variable < variables; variable++ {
			literal := ExtraBDDLit(variable+1) << 1
			if set&(uint64(1)<<variable) == 0 {
				literal = literal.Not()
			}
			cube, err = bdd.andRecursive(cube, literal, &bddWork)
			if err != nil {
				bdd.rollback(&transaction)
				return 0, err
			}
		}
		result, err = bdd.andRecursive(result.Not(), cube.Not(), &bddWork)
		if err != nil {
			bdd.rollback(&transaction)
			return 0, err
		}
		result = result.Not()
	}
	return result, nil
}

func (manager *CuddZDDManager) PortFromBDD(bdd *ExtraBDDManager, root ExtraBDDLit, options CuddZDDOptions) (CuddZDDEdge, error) {
	if bdd == nil {
		return 0, ErrDimension
	}
	if err := bdd.checkLiteral(root); err != nil {
		return 0, err
	}
	return manager.run(nil, options, func(_ []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		if bdd.variables != manager.variables {
			return nil, ErrDimension
		}
		family := make(map[uint64]struct{})
		limit := uint64(1) << manager.variables
		for assignment := uint64(0); assignment < limit; assignment++ {
			if err := work.step(); err != nil {
				return nil, err
			}
			cursor := root
			for cursor.ID() != 0 {
				if err := work.step(); err != nil {
					return nil, err
				}
				node := bdd.nodes[cursor.ID()]
				if assignment&(uint64(1)<<node.variable) != 0 {
					cursor = bdd.thenOf(cursor)
				} else {
					cursor = bdd.elseOf(cursor)
				}
			}
			if cursor == ExtraBDDTrue {
				if err := work.addSet(family, assignment); err != nil {
					return nil, err
				}
			}
		}
		return family, nil
	})
}

// Isop returns the lower bound and a deterministic full-minterm cover of it.
// It is valid whenever lower implies upper; it deliberately makes no minimality claim.
func (manager *CuddZDDManager) Isop(bdd *ExtraBDDManager, lower, upper ExtraBDDLit, options CuddZDDOptions) (ExtraBDDLit, CuddZDDEdge, error) {
	if bdd == nil {
		return 0, 0, ErrDimension
	}
	leq, err := bdd.Leq(lower, upper, ExtraBDDOptions{MaxOperations: options.MaxOperations, MaxBytes: options.MaxBytes, Stop: options.Stop})
	if err != nil {
		return 0, 0, err
	}
	if !leq {
		return 0, 0, ErrInvariant
	}
	cover, err := manager.PortFromBDD(bdd, lower, options)
	if err != nil {
		return 0, 0, err
	}
	return lower, cover, nil
}

func (manager *CuddZDDManager) BDDIsop(bdd *ExtraBDDManager, lower, upper ExtraBDDLit, options CuddZDDOptions) (ExtraBDDLit, error) {
	result, _, err := manager.Isop(bdd, lower, upper, options)
	return result, err
}
