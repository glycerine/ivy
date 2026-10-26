package abc

// Boolean analysis from src/bdd/cudd/cuddEssent.c, cuddLiteral.c,
// cuddBddCorr.c, cuddDecomp.c, cuddGenCof.c, cuddSat.c, cuddSolve.c,
// and cuddSplit.c.

import (
	"math"
	"slices"
)

type CuddLiteral struct {
	Variable int
	Value    bool
}

type CuddTwoLiteralClause struct {
	First  CuddLiteral
	Second CuddLiteral
	Unit   bool
}

type CuddCube struct {
	Assignments []ExtraBDDAssignment
	Root        ExtraBDDLit
	Minterms    uint64
	Exists      bool
}

type CuddDecomposition struct {
	First, Second ExtraBDDLit
}

type CuddEquationSolution struct {
	Consistency ExtraBDDLit
	Unknowns    []int
	Solutions   []ExtraBDDLit
}

func (manager *ExtraBDDManager) analysisCount(options ExtraBDDOptions, tables int) (int, ExtraBDDOptions, error) {
	if err := manager.check(); err != nil {
		return 0, options, err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, options, err
	}
	if tables < 1 || manager.variables >= 63 {
		return 0, options, ErrLimit
	}
	count64 := uint64(1) << uint(manager.variables)
	if count64 > uint64(math.MaxInt) || count64 > uint64(options.MaxBytes/tables) {
		return 0, options, ErrLimit
	}
	return int(count64), options, nil
}

func (manager *ExtraBDDManager) analysisDense(root ExtraBDDLit, count int, work *extraBDDWork) ([]bool, error) {
	if !manager.valid(root) {
		return nil, ErrInvalidLiteral
	}
	values := make([]bool, count)
	for assignment := 0; assignment < count; assignment++ {
		current := root
		for current >= 2 {
			if err := work.step(); err != nil {
				return nil, err
			}
			variable := manager.variableOf(current)
			if assignment&(1<<uint(variable)) != 0 {
				current = manager.thenOf(current)
			} else {
				current = manager.elseOf(current)
			}
		}
		values[assignment] = current == ExtraBDDTrue
	}
	return values, nil
}

func (manager *ExtraBDDManager) analysisBuild(values []bool, work *extraBDDWork) (ExtraBDDLit, error) {
	count := 1 << uint(manager.variables)
	if len(values) != count {
		return 0, ErrDimension
	}
	var build func(int, int) (ExtraBDDLit, error)
	build = func(variable, assignment int) (ExtraBDDLit, error) {
		if err := work.step(); err != nil {
			return 0, err
		}
		if variable == manager.variables {
			if values[assignment] {
				return ExtraBDDTrue, nil
			}
			return ExtraBDDFalse, nil
		}
		low, err := build(variable+1, assignment)
		if err != nil {
			return 0, err
		}
		high, err := build(variable+1, assignment|(1<<uint(variable)))
		if err != nil {
			return 0, err
		}
		return manager.uniqueCreate(variable, high, low, work.transaction)
	}
	return build(0, 0)
}

func (manager *ExtraBDDManager) analysisRun(roots []ExtraBDDLit, tables int, options ExtraBDDOptions, operation func(int, *extraBDDWork) (ExtraBDDLit, error)) (ExtraBDDLit, error) {
	count, normalized, err := manager.analysisCount(options, tables)
	if err != nil {
		return 0, err
	}
	for _, root := range roots {
		if !manager.valid(root) {
			return 0, ErrInvalidLiteral
		}
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	result, err := operation(count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	return result, nil
}

func (manager *ExtraBDDManager) EssentialLiterals(root ExtraBDDLit, options ExtraBDDOptions) ([]CuddLiteral, error) {
	count, normalized, err := manager.analysisCount(options, 1)
	if err != nil {
		return nil, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	values, err := manager.analysisDense(root, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return nil, err
	}
	seen := false
	canFalse := make([]bool, manager.variables)
	canTrue := make([]bool, manager.variables)
	for assignment, value := range values {
		if !value {
			continue
		}
		seen = true
		for variable := 0; variable < manager.variables; variable++ {
			if assignment&(1<<uint(variable)) != 0 {
				canTrue[variable] = true
			} else {
				canFalse[variable] = true
			}
		}
	}
	if !seen {
		return []CuddLiteral{}, nil
	}
	result := make([]CuddLiteral, 0, manager.variables)
	for variable := 0; variable < manager.variables; variable++ {
		if canFalse[variable] != canTrue[variable] {
			result = append(result, CuddLiteral{Variable: variable, Value: canTrue[variable]})
		}
	}
	return result, nil
}

func (manager *ExtraBDDManager) VariableEssential(root ExtraBDDLit, variable int, value bool, options ExtraBDDOptions) (bool, error) {
	if variable < 0 || variable >= manager.variables {
		return false, ErrDimension
	}
	literal, err := manager.Variable(variable)
	if err != nil {
		return false, err
	}
	if !value {
		literal = literal.Not()
	}
	return manager.Leq(root, literal, options)
}

func (manager *ExtraBDDManager) TwoLiteralClauses(root ExtraBDDLit, options ExtraBDDOptions) ([]CuddTwoLiteralClause, error) {
	count, normalized, err := manager.analysisCount(options, 1)
	if err != nil {
		return nil, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	values, err := manager.analysisDense(root, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return nil, err
	}
	if root < 2 {
		return []CuddTwoLiteralClause{}, nil
	}
	result := make([]CuddTwoLiteralClause, 0)
	for first := 0; first < manager.variables; first++ {
		for firstValue := 0; firstValue < 2; firstValue++ {
			unit := true
			for assignment, enabled := range values {
				if enabled && ((assignment>>uint(first))&1 != firstValue) {
					unit = false
					break
				}
			}
			if unit {
				result = append(result, CuddTwoLiteralClause{First: CuddLiteral{first, firstValue != 0}, Unit: true})
			}
			for second := first + 1; second < manager.variables; second++ {
				for secondValue := 0; secondValue < 2; secondValue++ {
					implied := true
					for assignment, enabled := range values {
						if !enabled {
							continue
						}
						one := (assignment>>uint(first))&1 == firstValue
						two := (assignment>>uint(second))&1 == secondValue
						if !one && !two {
							implied = false
							break
						}
					}
					if implied {
						result = append(result, CuddTwoLiteralClause{First: CuddLiteral{first, firstValue != 0}, Second: CuddLiteral{second, secondValue != 0}})
					}
				}
			}
		}
	}
	return result, nil
}

func CuddLiteralSetIntersection(first, second []CuddLiteral) ([]CuddLiteral, error) {
	left := make(map[int]bool, len(first))
	for _, literal := range first {
		if literal.Variable < 0 {
			return nil, ErrDimension
		}
		if old, ok := left[literal.Variable]; ok && old != literal.Value {
			return nil, ErrInvariant
		}
		left[literal.Variable] = literal.Value
	}
	result := make([]CuddLiteral, 0)
	seen := make(map[int]bool, len(second))
	for _, literal := range second {
		if literal.Variable < 0 {
			return nil, ErrDimension
		}
		if old, ok := seen[literal.Variable]; ok && old != literal.Value {
			return nil, ErrInvariant
		}
		if _, duplicate := seen[literal.Variable]; duplicate {
			continue
		}
		seen[literal.Variable] = literal.Value
		if value, ok := left[literal.Variable]; ok && value == literal.Value {
			result = append(result, literal)
		}
	}
	slices.SortFunc(result, func(a, b CuddLiteral) int { return a.Variable - b.Variable })
	return result, nil
}

func (manager *ExtraBDDManager) Correlation(first, second ExtraBDDLit, probabilities []float64, options ExtraBDDOptions) (float64, error) {
	if len(probabilities) != 0 && len(probabilities) != manager.variables {
		return 0, ErrDimension
	}
	for _, probability := range probabilities {
		if math.IsNaN(probability) || probability < 0 || probability > 1 {
			return 0, ErrDimension
		}
	}
	count, normalized, err := manager.analysisCount(options, 2)
	if err != nil {
		return 0, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	left, err := manager.analysisDense(first, count, &work)
	if err == nil {
		var right []bool
		right, err = manager.analysisDense(second, count, &work)
		if err == nil {
			total := 0.0
			for assignment := 0; assignment < count; assignment++ {
				if left[assignment] != right[assignment] {
					continue
				}
				weight := 1.0
				if len(probabilities) == 0 {
					weight = 1 / float64(count)
				} else {
					for variable, probability := range probabilities {
						if assignment&(1<<uint(variable)) != 0 {
							weight *= probability
						} else {
							weight *= 1 - probability
						}
					}
				}
				total += weight
			}
			return total, nil
		}
	}
	manager.rollback(&transaction)
	return 0, err
}

func denseBinary(left, right []bool, operation byte) []bool {
	result := make([]bool, len(left))
	for index := range left {
		switch operation {
		case '&':
			result[index] = left[index] && right[index]
		case '|':
			result[index] = left[index] || right[index]
		case '^':
			result[index] = left[index] != right[index]
		}
	}
	return result
}

func (manager *ExtraBDDManager) decomposition(root ExtraBDDLit, disjunctive bool, options ExtraBDDOptions) (CuddDecomposition, error) {
	var decomposition CuddDecomposition
	result, err := manager.analysisRun([]ExtraBDDLit{root}, 4, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		values, err := manager.analysisDense(root, count, work)
		if err != nil {
			return 0, err
		}
		for split := 1; split < (1<<uint(manager.variables))-1; split++ {
			first := make([]bool, count)
			second := make([]bool, count)
			if disjunctive {
				for assignment := 0; assignment < count; assignment++ {
					first[assignment] = true
					second[assignment] = true
				}
			}
			for assignment := 0; assignment < count; assignment++ {
				if err := work.step(); err != nil {
					return 0, err
				}
				for completion := 0; completion < count; completion++ {
					if completion&split == assignment&split {
						if disjunctive {
							first[assignment] = first[assignment] && values[completion]
						} else {
							first[assignment] = first[assignment] || values[completion]
						}
					}
					other := (^split) & (count - 1)
					if completion&other == assignment&other {
						if disjunctive {
							second[assignment] = second[assignment] && values[completion]
						} else {
							second[assignment] = second[assignment] || values[completion]
						}
					}
				}
			}
			combined := denseBinary(first, second, map[bool]byte{false: '&', true: '|'}[disjunctive])
			if slices.Equal(combined, values) {
				one, err := manager.analysisBuild(first, work)
				if err != nil {
					return 0, err
				}
				two, err := manager.analysisBuild(second, work)
				if err != nil {
					return 0, err
				}
				decomposition = CuddDecomposition{one, two}
				return root, nil
			}
		}
		if disjunctive {
			decomposition = CuddDecomposition{root, ExtraBDDFalse}
		} else {
			decomposition = CuddDecomposition{root, ExtraBDDTrue}
		}
		return root, nil
	})
	_ = result
	return decomposition, err
}

func (manager *ExtraBDDManager) ConjunctiveDecomposition(root ExtraBDDLit, options ExtraBDDOptions) (CuddDecomposition, error) {
	return manager.decomposition(root, false, options)
}

func (manager *ExtraBDDManager) DisjunctiveDecomposition(root ExtraBDDLit, options ExtraBDDOptions) (CuddDecomposition, error) {
	return manager.decomposition(root, true, options)
}

func (manager *ExtraBDDManager) Constrain(root, care ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.runBooleanOperation([]ExtraBDDLit{root, care}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		type pair struct{ root, care ExtraBDDLit }
		memo := make(map[pair]ExtraBDDLit)
		var constrain func(ExtraBDDLit, ExtraBDDLit) (ExtraBDDLit, error)
		constrain = func(function, condition ExtraBDDLit) (ExtraBDDLit, error) {
			if err := work.step(); err != nil {
				return 0, err
			}
			switch {
			case condition == ExtraBDDFalse:
				return ExtraBDDFalse, nil
			case condition == ExtraBDDTrue || function < 2:
				return function, nil
			case function == condition:
				return ExtraBDDTrue, nil
			case function == condition.Not():
				return ExtraBDDFalse, nil
			}
			key := pair{function, condition}
			if result, ok := memo[key]; ok {
				return result, nil
			}
			top := manager.variableOf(function)
			if variable := manager.variableOf(condition); variable < top {
				top = variable
			}
			branches := func(literal ExtraBDDLit) (ExtraBDDLit, ExtraBDDLit) {
				if manager.variableOf(literal) == top {
					return manager.thenOf(literal), manager.elseOf(literal)
				}
				return literal, literal
			}
			functionHigh, functionLow := branches(function)
			careHigh, careLow := branches(condition)
			if careLow == ExtraBDDFalse {
				return constrain(functionHigh, careHigh)
			}
			if careHigh == ExtraBDDFalse {
				return constrain(functionLow, careLow)
			}
			high, err := constrain(functionHigh, careHigh)
			if err != nil {
				return 0, err
			}
			low, err := constrain(functionLow, careLow)
			if err != nil {
				return 0, err
			}
			result, err := manager.uniqueCreate(top, high, low, work.transaction)
			if err == nil {
				memo[key] = result
			}
			return result, err
		}
		return constrain(root, care)
	})
}

func (manager *ExtraBDDManager) RestrictCare(root, care ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.runBooleanOperation([]ExtraBDDLit{root, care}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		type pair struct{ root, care ExtraBDDLit }
		memo := make(map[pair]ExtraBDDLit)
		var restrict func(ExtraBDDLit, ExtraBDDLit) (ExtraBDDLit, error)
		restrict = func(function, condition ExtraBDDLit) (ExtraBDDLit, error) {
			if err := work.step(); err != nil {
				return 0, err
			}
			switch {
			case condition == ExtraBDDTrue:
				return function, nil
			case condition == ExtraBDDFalse:
				return ExtraBDDFalse, nil
			case function < 2:
				return function, nil
			case function == condition:
				return ExtraBDDTrue, nil
			case function == condition.Not():
				return ExtraBDDFalse, nil
			}
			key := pair{function, condition}
			if result, ok := memo[key]; ok {
				return result, nil
			}
			functionTop := manager.variableOf(function)
			careTop := manager.variableOf(condition)
			if careTop < functionTop {
				careHigh, careLow := manager.thenOf(condition), manager.elseOf(condition)
				projected, err := manager.andRecursive(careHigh.Not(), careLow.Not(), work)
				if err != nil {
					return 0, err
				}
				result, err := restrict(function, projected.Not())
				if err == nil {
					memo[key] = result
				}
				return result, err
			}
			functionHigh, functionLow := manager.thenOf(function), manager.elseOf(function)
			careHigh, careLow := condition, condition
			if careTop == functionTop {
				careHigh, careLow = manager.thenOf(condition), manager.elseOf(condition)
			}
			if careHigh == ExtraBDDFalse {
				return restrict(functionLow, careLow)
			}
			if careLow == ExtraBDDFalse {
				return restrict(functionHigh, careHigh)
			}
			high, err := restrict(functionHigh, careHigh)
			if err != nil {
				return 0, err
			}
			low, err := restrict(functionLow, careLow)
			if err != nil {
				return 0, err
			}
			result, err := manager.uniqueCreate(functionTop, high, low, work.transaction)
			if err == nil {
				memo[key] = result
			}
			return result, err
		}

		result, err := restrict(root, care)
		if err != nil {
			return 0, err
		}
		if manager.analysisNodeCount(root) <= manager.analysisNodeCount(result) {
			manager.rollback(work.transaction)
			return root, nil
		}
		return result, nil
	})
}

func (manager *ExtraBDDManager) analysisNodeCount(root ExtraBDDLit) int {
	seen := make(map[uint32]struct{})
	stack := []ExtraBDDLit{root}
	for len(stack) != 0 {
		literal := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		id := literal.ID()
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if literal >= 2 {
			stack = append(stack, manager.thenOf(literal), manager.elseOf(literal))
		}
	}
	return len(seen)
}

func (manager *ExtraBDDManager) NPAnd(first, second ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.analysisRun([]ExtraBDDLit{first, second}, 4, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		left, err := manager.analysisDense(first, count, work)
		if err != nil {
			return 0, err
		}
		right, err := manager.analysisDense(second, count, work)
		if err != nil {
			return 0, err
		}
		supportMask := 0
		for variable := 0; variable < manager.variables; variable++ {
			bit := 1 << uint(variable)
			for assignment := 0; assignment < count; assignment++ {
				if assignment&bit == 0 && left[assignment] != left[assignment|bit] {
					supportMask |= bit
					break
				}
			}
		}
		result := make([]bool, count)
		for assignment := 0; assignment < count; assignment++ {
			if !left[assignment] {
				continue
			}
			for completion := 0; completion < count; completion++ {
				if err := work.step(); err != nil {
					return 0, err
				}
				if completion&supportMask == assignment&supportMask && right[completion] {
					result[assignment] = true
					break
				}
			}
		}
		return manager.analysisBuild(result, work)
	})
}

func (manager *ExtraBDDManager) Squeeze(lower, upper ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	leq, err := manager.Leq(lower, upper, options)
	if err != nil {
		return 0, err
	}
	if !leq {
		return 0, ErrInvariant
	}
	return lower, nil
}

func (manager *ExtraBDDManager) SubsetCompress(root, care ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.And(root, care, options)
}

func (manager *ExtraBDDManager) SupersetCompress(root, care ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.Or(root, care.Not(), options)
}

func (manager *ExtraBDDManager) SatCount(root ExtraBDDLit, options ExtraBDDOptions) (uint64, error) {
	count, normalized, err := manager.analysisCount(options, 1)
	if err != nil {
		return 0, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	values, err := manager.analysisDense(root, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	var result uint64
	for _, value := range values {
		if value {
			result++
		}
	}
	return result, nil
}

func cubeMatches(assignment, mask, bits int) bool { return assignment&mask == bits&mask }

func (manager *ExtraBDDManager) LargestCube(root ExtraBDDLit, options ExtraBDDOptions) (CuddCube, error) {
	var result CuddCube
	_, err := manager.analysisRun([]ExtraBDDLit{root}, 3, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		values, err := manager.analysisDense(root, count, work)
		if err != nil {
			return 0, err
		}
		bestMask, bestBits, bestLiterals := 0, 0, manager.variables+1
		for bits, enabled := range values {
			if !enabled {
				continue
			}
			for mask := 0; mask < count; mask++ {
				if err := work.step(); err != nil {
					return 0, err
				}
				literals := bitsCount(uint(mask))
				if literals >= bestLiterals {
					continue
				}
				valid := true
				for assignment, value := range values {
					if cubeMatches(assignment, mask, bits) && !value {
						valid = false
						break
					}
				}
				if valid {
					bestMask, bestBits, bestLiterals = mask, bits, literals
				}
			}
		}
		if bestLiterals == manager.variables+1 {
			result = CuddCube{Root: ExtraBDDFalse}
			return ExtraBDDFalse, nil
		}
		cubeValues := make([]bool, count)
		for assignment := 0; assignment < count; assignment++ {
			cubeValues[assignment] = cubeMatches(assignment, bestMask, bestBits)
		}
		cubeRoot, err := manager.analysisBuild(cubeValues, work)
		if err != nil {
			return 0, err
		}
		assignments := make([]ExtraBDDAssignment, 0, bestLiterals)
		for variable := 0; variable < manager.variables; variable++ {
			if bestMask&(1<<uint(variable)) != 0 {
				assignments = append(assignments, ExtraBDDAssignment{Variable: variable, Value: bestBits&(1<<uint(variable)) != 0})
			}
		}
		result = CuddCube{Assignments: assignments, Root: cubeRoot, Minterms: uint64(1) << uint(manager.variables-bestLiterals), Exists: true}
		return cubeRoot, nil
	})
	return result, err
}

func bitsCount(value uint) int {
	count := 0
	for value != 0 {
		value &= value - 1
		count++
	}
	return count
}

// ShortestPath returns a minimum-literal satisfying cube. With unit arc
// weights this is the Boolean specialization of Cudd_ShortestPath.
func (manager *ExtraBDDManager) ShortestPath(root ExtraBDDLit, options ExtraBDDOptions) (CuddCube, error) {
	return manager.LargestCube(root, options)
}

func (manager *ExtraBDDManager) ShortestLength(root ExtraBDDLit, options ExtraBDDOptions) (int, error) {
	cube, err := manager.ShortestPath(root, options)
	if err != nil {
		return 0, err
	}
	if !cube.Exists {
		return -1, nil
	}
	return len(cube.Assignments), nil
}

func (manager *ExtraBDDManager) MakePrime(cube, root ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	assignments, err := manager.cubeAssignments(cube)
	if err != nil {
		return 0, err
	}
	return manager.analysisRun([]ExtraBDDLit{cube, root}, 3, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		values, err := manager.analysisDense(root, count, work)
		if err != nil {
			return 0, err
		}
		implies := func(candidate []ExtraBDDAssignment) (bool, error) {
			for assignment := 0; assignment < count; assignment++ {
				if err := work.step(); err != nil {
					return false, err
				}
				matches := true
				for _, literal := range candidate {
					if (assignment&(1<<uint(literal.Variable)) != 0) != literal.Value {
						matches = false
						break
					}
				}
				if matches && !values[assignment] {
					return false, nil
				}
			}
			return true, nil
		}
		valid, err := implies(assignments)
		if err != nil {
			return 0, err
		}
		if !valid {
			return 0, ErrInvariant
		}
		for index := 0; index < len(assignments); {
			candidate := slices.Concat(assignments[:index], assignments[index+1:])
			expand, err := implies(candidate)
			if err != nil {
				return 0, err
			}
			if expand {
				assignments = candidate
			} else {
				index++
			}
		}
		prime := make([]bool, count)
		for assignment := 0; assignment < count; assignment++ {
			prime[assignment] = true
			for _, literal := range assignments {
				if (assignment&(1<<uint(literal.Variable)) != 0) != literal.Value {
					prime[assignment] = false
					break
				}
			}
		}
		return manager.analysisBuild(prime, work)
	})
}

func (manager *ExtraBDDManager) Increasing(root ExtraBDDLit, variable int, options ExtraBDDOptions) (bool, error) {
	if variable < 0 || variable >= manager.variables {
		return false, ErrDimension
	}
	low, err := manager.Restrict(root, []ExtraBDDAssignment{{Variable: variable, Value: false}}, options)
	if err != nil {
		return false, err
	}
	high, err := manager.Restrict(root, []ExtraBDDAssignment{{Variable: variable, Value: true}}, options)
	if err != nil {
		return false, err
	}
	return manager.Leq(low, high, options)
}

func (manager *ExtraBDDManager) Decreasing(root ExtraBDDLit, variable int, options ExtraBDDOptions) (bool, error) {
	if variable < 0 || variable >= manager.variables {
		return false, ErrDimension
	}
	low, err := manager.Restrict(root, []ExtraBDDAssignment{{Variable: variable, Value: false}}, options)
	if err != nil {
		return false, err
	}
	high, err := manager.Restrict(root, []ExtraBDDAssignment{{Variable: variable, Value: true}}, options)
	if err != nil {
		return false, err
	}
	return manager.Leq(high, low, options)
}

func (manager *ExtraBDDManager) EquivalentDontCare(first, second, dontCare ExtraBDDLit, options ExtraBDDOptions) (bool, error) {
	difference, err := manager.Xor(first, second, options)
	if err != nil {
		return false, err
	}
	return manager.Leq(difference, dontCare, options)
}

func (manager *ExtraBDDManager) LeqUnless(first, second, exception ExtraBDDLit, options ExtraBDDOptions) (bool, error) {
	right, err := manager.Or(second, exception, options)
	if err != nil {
		return false, err
	}
	return manager.Leq(first, right, options)
}

func (manager *ExtraBDDManager) SplitSet(root ExtraBDDLit, minterms uint64, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.analysisRun([]ExtraBDDLit{root}, 2, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		values, err := manager.analysisDense(root, count, work)
		if err != nil {
			return 0, err
		}
		selected := make([]bool, count)
		for assignment, value := range values {
			if value && minterms > 0 {
				selected[assignment] = true
				minterms--
			}
		}
		if minterms != 0 {
			return 0, ErrDimension
		}
		return manager.analysisBuild(selected, work)
	})
}

func (manager *ExtraBDDManager) SolveZeroEquation(root ExtraBDDLit, unknowns []int, options ExtraBDDOptions) (CuddEquationSolution, error) {
	result := CuddEquationSolution{Unknowns: slices.Clone(unknowns)}
	seen := make([]bool, manager.variables)
	for _, variable := range unknowns {
		if variable < 0 || variable >= manager.variables || seen[variable] {
			return result, ErrDimension
		}
		seen[variable] = true
	}
	_, err := manager.analysisRun([]ExtraBDDLit{root}, len(unknowns)+3, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		values, err := manager.analysisDense(root, count, work)
		if err != nil {
			return 0, err
		}
		consistency := make([]bool, count)
		solutions := make([][]bool, len(unknowns))
		for index := range solutions {
			solutions[index] = make([]bool, count)
		}
		unknownMask := 0
		for _, variable := range unknowns {
			unknownMask |= 1 << uint(variable)
		}
		for base := 0; base < count; base++ {
			if base&unknownMask != 0 {
				continue
			}
			found := -1
			for choice := 0; choice < (1 << uint(len(unknowns))); choice++ {
				assignment := base
				for index, variable := range unknowns {
					if choice&(1<<uint(index)) != 0 {
						assignment |= 1 << uint(variable)
					}
				}
				if !values[assignment] {
					found = choice
					break
				}
			}
			for assignment := base; assignment < count; assignment++ {
				if assignment&^unknownMask != base {
					continue
				}
				consistency[assignment] = found >= 0
				for index := range unknowns {
					solutions[index][assignment] = found >= 0 && found&(1<<uint(index)) != 0
				}
			}
		}
		result.Consistency, err = manager.analysisBuild(consistency, work)
		if err != nil {
			return 0, err
		}
		result.Solutions = make([]ExtraBDDLit, len(solutions))
		for index, table := range solutions {
			result.Solutions[index], err = manager.analysisBuild(table, work)
			if err != nil {
				return 0, err
			}
		}
		return result.Consistency, nil
	})
	return result, err
}

func (manager *ExtraBDDManager) VerifyZeroEquation(root ExtraBDDLit, solution CuddEquationSolution, options ExtraBDDOptions) (bool, error) {
	if len(solution.Unknowns) != len(solution.Solutions) {
		return false, ErrDimension
	}
	replacements := make([]ExtraBDDLit, manager.variables)
	for variable := 0; variable < manager.variables; variable++ {
		replacements[variable] = ExtraBDDLit(variable+1) << 1
	}
	for index, variable := range solution.Unknowns {
		if variable < 0 || variable >= manager.variables || !manager.valid(solution.Solutions[index]) {
			return false, ErrDimension
		}
		replacements[variable] = solution.Solutions[index]
	}
	composed, err := manager.VectorCompose(root, replacements, options)
	if err != nil {
		return false, err
	}
	return manager.Leq(solution.Consistency, composed.Not(), options)
}
