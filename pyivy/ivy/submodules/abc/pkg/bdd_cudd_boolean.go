package abc

// Boolean BDD operations from src/bdd/cudd/cuddBddIte.c,
// src/bdd/cudd/cuddAndAbs.c, src/bdd/cudd/cuddBddAbs.c,
// src/bdd/cudd/cuddCof.c, and the BDD portions of src/bdd/cudd/cuddCompose.c.

type extraBDDIteKey struct {
	condition, then, otherwise ExtraBDDLit
}

type extraBDDPairKey struct {
	first, second ExtraBDDLit
}

type ExtraBDDAssignment struct {
	Variable int
	Value    bool
}

func (manager *ExtraBDDManager) runBooleanOperation(roots []ExtraBDDLit, options ExtraBDDOptions, operation func(*extraBDDWork) (ExtraBDDLit, error)) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	for _, root := range roots {
		if !manager.valid(root) {
			return 0, ErrInvalidLiteral
		}
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	result, err := operation(&work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	return result, nil
}

func (manager *ExtraBDDManager) iteRecursive(condition, then, otherwise ExtraBDDLit, work *extraBDDWork, memo map[extraBDDIteKey]ExtraBDDLit) (ExtraBDDLit, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	switch {
	case condition == ExtraBDDTrue:
		return then, nil
	case condition == ExtraBDDFalse:
		return otherwise, nil
	case then == otherwise:
		return then, nil
	case then == ExtraBDDTrue && otherwise == ExtraBDDFalse:
		return condition, nil
	case then == ExtraBDDFalse && otherwise == ExtraBDDTrue:
		return condition.Not(), nil
	}
	key := extraBDDIteKey{condition: condition, then: then, otherwise: otherwise}
	if result, ok := memo[key]; ok {
		return result, nil
	}
	top := manager.variableOf(condition)
	if variable := manager.variableOf(then); variable < top {
		top = variable
	}
	if variable := manager.variableOf(otherwise); variable < top {
		top = variable
	}
	cofactors := func(root ExtraBDDLit) (ExtraBDDLit, ExtraBDDLit) {
		if manager.variableOf(root) == top {
			return manager.thenOf(root), manager.elseOf(root)
		}
		return root, root
	}
	conditionHigh, conditionLow := cofactors(condition)
	thenHigh, thenLow := cofactors(then)
	elseHigh, elseLow := cofactors(otherwise)
	high, err := manager.iteRecursive(conditionHigh, thenHigh, elseHigh, work, memo)
	if err != nil {
		return 0, err
	}
	low, err := manager.iteRecursive(conditionLow, thenLow, elseLow, work, memo)
	if err != nil {
		return 0, err
	}
	result, err := manager.uniqueCreate(top, high, low, work.transaction)
	if err != nil {
		return 0, err
	}
	memo[key] = result
	return result, nil
}

func (manager *ExtraBDDManager) Ite(condition, then, otherwise ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.runBooleanOperation([]ExtraBDDLit{condition, then, otherwise}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		return manager.iteRecursive(condition, then, otherwise, work, make(map[extraBDDIteKey]ExtraBDDLit))
	})
}

func (manager *ExtraBDDManager) IteConstant(condition, then, otherwise ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, bool, error) {
	result, err := manager.Ite(condition, then, otherwise, options)
	if err != nil {
		return 0, false, err
	}
	return result, result < 2, nil
}

func (manager *ExtraBDDManager) Intersect(first, second ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.And(first, second, options)
}

func (manager *ExtraBDDManager) AndLimit(first, second ExtraBDDLit, maximumNewNodes int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if maximumNewNodes < 0 {
		return 0, ErrDimension
	}
	return manager.runBooleanOperation([]ExtraBDDLit{first, second}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		result, err := manager.andRecursive(first, second, work)
		if err != nil {
			return 0, err
		}
		if len(manager.nodes)-work.transaction.nodeCount > maximumNewNodes {
			return 0, ErrLimit
		}
		return result, nil
	})
}

func (manager *ExtraBDDManager) Nand(first, second ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	result, err := manager.And(first, second, options)
	if err != nil {
		return 0, err
	}
	return result.Not(), nil
}

func (manager *ExtraBDDManager) Nor(first, second ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	result, err := manager.Or(first, second, options)
	if err != nil {
		return 0, err
	}
	return result.Not(), nil
}

func (manager *ExtraBDDManager) Xor(first, second ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.Ite(first, second.Not(), second, options)
}

func (manager *ExtraBDDManager) Xnor(first, second ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	result, err := manager.Xor(first, second, options)
	if err != nil {
		return 0, err
	}
	return result.Not(), nil
}

func (manager *ExtraBDDManager) leqRecursive(first, second ExtraBDDLit, work *extraBDDWork, memo map[extraBDDPairKey]bool) (bool, error) {
	if err := work.step(); err != nil {
		return false, err
	}
	if first == ExtraBDDFalse || second == ExtraBDDTrue || first == second {
		return true, nil
	}
	if first == ExtraBDDTrue || second == ExtraBDDFalse || first == second.Not() {
		return false, nil
	}
	key := extraBDDPairKey{first: first, second: second}
	if result, ok := memo[key]; ok {
		return result, nil
	}
	top := manager.variableOf(first)
	if variable := manager.variableOf(second); variable < top {
		top = variable
	}
	branches := func(root ExtraBDDLit) (ExtraBDDLit, ExtraBDDLit) {
		if manager.variableOf(root) == top {
			return manager.thenOf(root), manager.elseOf(root)
		}
		return root, root
	}
	firstHigh, firstLow := branches(first)
	secondHigh, secondLow := branches(second)
	high, err := manager.leqRecursive(firstHigh, secondHigh, work, memo)
	if err != nil || !high {
		return high, err
	}
	low, err := manager.leqRecursive(firstLow, secondLow, work, memo)
	if err == nil {
		memo[key] = low
	}
	return low, err
}

func (manager *ExtraBDDManager) Leq(first, second ExtraBDDLit, options ExtraBDDOptions) (bool, error) {
	if err := manager.checkLiteral(first); err != nil {
		return false, err
	}
	if err := manager.checkLiteral(second); err != nil {
		return false, err
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return false, err
	}
	work := extraBDDWork{options: normalized}
	return manager.leqRecursive(first, second, &work, make(map[extraBDDPairKey]bool))
}

func (manager *ExtraBDDManager) abstractionSet(variables []int) ([]bool, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	set := make([]bool, manager.variables)
	for _, variable := range variables {
		if variable < 0 || variable >= manager.variables {
			return nil, ErrDimension
		}
		set[variable] = true
	}
	return set, nil
}

func (manager *ExtraBDDManager) abstractRecursive(root ExtraBDDLit, variables []bool, parity bool, universal bool, work *extraBDDWork, iteMemo map[extraBDDIteKey]ExtraBDDLit, memo map[ExtraBDDLit]ExtraBDDLit) (ExtraBDDLit, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if root < 2 {
		return root, nil
	}
	if result, ok := memo[root]; ok {
		return result, nil
	}
	variable := manager.variableOf(root)
	high, err := manager.abstractRecursive(manager.thenOf(root), variables, parity, universal, work, iteMemo, memo)
	if err != nil {
		return 0, err
	}
	low, err := manager.abstractRecursive(manager.elseOf(root), variables, parity, universal, work, iteMemo, memo)
	if err != nil {
		return 0, err
	}
	var result ExtraBDDLit
	if variables[variable] {
		switch {
		case parity:
			result, err = manager.iteRecursive(high, low.Not(), low, work, iteMemo)
		case universal:
			result, err = manager.iteRecursive(high, low, ExtraBDDFalse, work, iteMemo)
		default:
			result, err = manager.iteRecursive(high, ExtraBDDTrue, low, work, iteMemo)
		}
	} else {
		result, err = manager.uniqueCreate(variable, high, low, work.transaction)
	}
	if err != nil {
		return 0, err
	}
	memo[root] = result
	return result, nil
}

func (manager *ExtraBDDManager) abstract(root ExtraBDDLit, variables []int, parity, universal bool, options ExtraBDDOptions) (ExtraBDDLit, error) {
	set, err := manager.abstractionSet(variables)
	if err != nil {
		return 0, err
	}
	return manager.runBooleanOperation([]ExtraBDDLit{root}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		return manager.abstractRecursive(root, set, parity, universal, work, make(map[extraBDDIteKey]ExtraBDDLit), make(map[ExtraBDDLit]ExtraBDDLit))
	})
}

func (manager *ExtraBDDManager) ExistAbstract(root ExtraBDDLit, variables []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.abstract(root, variables, false, false, options)
}

func (manager *ExtraBDDManager) UnivAbstract(root ExtraBDDLit, variables []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.abstract(root, variables, false, true, options)
}

func (manager *ExtraBDDManager) XorExistAbstract(first, second ExtraBDDLit, variables []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	set, err := manager.abstractionSet(variables)
	if err != nil {
		return 0, err
	}
	return manager.runBooleanOperation([]ExtraBDDLit{first, second}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		iteMemo := make(map[extraBDDIteKey]ExtraBDDLit)
		xor, xorErr := manager.iteRecursive(first, second.Not(), second, work, iteMemo)
		if xorErr != nil {
			return 0, xorErr
		}
		return manager.abstractRecursive(xor, set, false, false, work, iteMemo, make(map[ExtraBDDLit]ExtraBDDLit))
	})
}

func (manager *ExtraBDDManager) AndAbstract(first, second ExtraBDDLit, variables []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.AndAbstractLimit(first, second, variables, -1, options)
}

func (manager *ExtraBDDManager) AndAbstractLimit(first, second ExtraBDDLit, variables []int, maximumNewNodes int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if maximumNewNodes < -1 {
		return 0, ErrDimension
	}
	set, err := manager.abstractionSet(variables)
	if err != nil {
		return 0, err
	}
	return manager.runBooleanOperation([]ExtraBDDLit{first, second}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		iteMemo := make(map[extraBDDIteKey]ExtraBDDLit)
		and, andErr := manager.iteRecursive(first, second, ExtraBDDFalse, work, iteMemo)
		if andErr != nil {
			return 0, andErr
		}
		result, abstractErr := manager.abstractRecursive(and, set, false, false, work, iteMemo, make(map[ExtraBDDLit]ExtraBDDLit))
		if abstractErr != nil {
			return 0, abstractErr
		}
		if maximumNewNodes >= 0 && len(manager.nodes)-work.transaction.nodeCount > maximumNewNodes {
			return 0, ErrLimit
		}
		return result, nil
	})
}

func (manager *ExtraBDDManager) restrictRecursive(root ExtraBDDLit, variable int, value bool, work *extraBDDWork, memo map[ExtraBDDLit]ExtraBDDLit) (ExtraBDDLit, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if root < 2 || manager.variableOf(root) > variable {
		return root, nil
	}
	if manager.variableOf(root) == variable {
		if value {
			return manager.thenOf(root), nil
		}
		return manager.elseOf(root), nil
	}
	if result, ok := memo[root]; ok {
		return result, nil
	}
	high, err := manager.restrictRecursive(manager.thenOf(root), variable, value, work, memo)
	if err != nil {
		return 0, err
	}
	low, err := manager.restrictRecursive(manager.elseOf(root), variable, value, work, memo)
	if err != nil {
		return 0, err
	}
	result, err := manager.uniqueCreate(manager.variableOf(root), high, low, work.transaction)
	if err != nil {
		return 0, err
	}
	memo[root] = result
	return result, nil
}

func (manager *ExtraBDDManager) restrictMany(root ExtraBDDLit, assignments []ExtraBDDAssignment, work *extraBDDWork) (ExtraBDDLit, error) {
	result := root
	seen := make([]int8, manager.variables)
	for _, assignment := range assignments {
		if assignment.Variable < 0 || assignment.Variable >= manager.variables {
			return 0, ErrDimension
		}
		value := int8(1)
		if assignment.Value {
			value = 2
		}
		if seen[assignment.Variable] != 0 && seen[assignment.Variable] != value {
			return 0, ErrInvariant
		}
		seen[assignment.Variable] = value
		var err error
		result, err = manager.restrictRecursive(result, assignment.Variable, assignment.Value, work, make(map[ExtraBDDLit]ExtraBDDLit))
		if err != nil {
			return 0, err
		}
	}
	return result, nil
}

func (manager *ExtraBDDManager) Restrict(root ExtraBDDLit, assignments []ExtraBDDAssignment, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.runBooleanOperation([]ExtraBDDLit{root}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		return manager.restrictMany(root, assignments, work)
	})
}

func (manager *ExtraBDDManager) cubeAssignments(cube ExtraBDDLit) ([]ExtraBDDAssignment, error) {
	if err := manager.checkLiteral(cube); err != nil {
		return nil, err
	}
	if cube == ExtraBDDFalse {
		return nil, ErrInvariant
	}
	assignments := make([]ExtraBDDAssignment, 0)
	for cube >= 2 {
		variable := manager.variableOf(cube)
		high, low := manager.thenOf(cube), manager.elseOf(cube)
		switch {
		case low == ExtraBDDFalse && high != ExtraBDDFalse:
			assignments = append(assignments, ExtraBDDAssignment{Variable: variable, Value: true})
			cube = high
		case high == ExtraBDDFalse && low != ExtraBDDFalse:
			assignments = append(assignments, ExtraBDDAssignment{Variable: variable})
			cube = low
		default:
			return nil, ErrInvariant
		}
	}
	if cube != ExtraBDDTrue {
		return nil, ErrInvariant
	}
	return assignments, nil
}

func (manager *ExtraBDDManager) CheckCube(cube ExtraBDDLit) (bool, error) {
	_, err := manager.cubeAssignments(cube)
	if err == ErrInvariant {
		return false, nil
	}
	return err == nil, err
}

func (manager *ExtraBDDManager) Cofactor(root, cube ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	assignments, err := manager.cubeAssignments(cube)
	if err != nil {
		return 0, err
	}
	return manager.Restrict(root, assignments, options)
}

func (manager *ExtraBDDManager) BooleanDifference(root ExtraBDDLit, variable int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if variable < 0 || variable >= manager.variables {
		return 0, ErrDimension
	}
	return manager.runBooleanOperation([]ExtraBDDLit{root}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		high, err := manager.restrictRecursive(root, variable, true, work, make(map[ExtraBDDLit]ExtraBDDLit))
		if err != nil {
			return 0, err
		}
		low, err := manager.restrictRecursive(root, variable, false, work, make(map[ExtraBDDLit]ExtraBDDLit))
		if err != nil {
			return 0, err
		}
		return manager.iteRecursive(high, low.Not(), low, work, make(map[extraBDDIteKey]ExtraBDDLit))
	})
}

func (manager *ExtraBDDManager) VariableDependent(root ExtraBDDLit, variable int, options ExtraBDDOptions) (bool, error) {
	if err := manager.check(); err != nil {
		return false, err
	}
	if !manager.valid(root) {
		return false, ErrInvalidLiteral
	}
	if variable < 0 || variable >= manager.variables {
		return false, ErrDimension
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return false, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	high, err := manager.restrictRecursive(root, variable, true, &work, make(map[ExtraBDDLit]ExtraBDDLit))
	if err != nil {
		manager.rollback(&transaction)
		return false, err
	}
	low, err := manager.restrictRecursive(root, variable, false, &work, make(map[ExtraBDDLit]ExtraBDDLit))
	if err != nil {
		manager.rollback(&transaction)
		return false, err
	}
	result, err := manager.leqRecursive(high, low.Not(), &work, make(map[extraBDDPairKey]bool))
	if err != nil {
		manager.rollback(&transaction)
		return false, err
	}
	return result, nil
}

func (manager *ExtraBDDManager) composeRecursive(root ExtraBDDLit, replacements []ExtraBDDLit, work *extraBDDWork, iteMemo map[extraBDDIteKey]ExtraBDDLit, memo map[ExtraBDDLit]ExtraBDDLit) (ExtraBDDLit, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if root < 2 {
		return root, nil
	}
	if result, ok := memo[root]; ok {
		return result, nil
	}
	variable := manager.variableOf(root)
	high, err := manager.composeRecursive(manager.thenOf(root), replacements, work, iteMemo, memo)
	if err != nil {
		return 0, err
	}
	low, err := manager.composeRecursive(manager.elseOf(root), replacements, work, iteMemo, memo)
	if err != nil {
		return 0, err
	}
	result, err := manager.iteRecursive(replacements[variable], high, low, work, iteMemo)
	if err != nil {
		return 0, err
	}
	memo[root] = result
	return result, nil
}

func (manager *ExtraBDDManager) VectorCompose(root ExtraBDDLit, replacements []ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if len(replacements) != manager.variables {
		return 0, ErrDimension
	}
	roots := make([]ExtraBDDLit, 1, len(replacements)+1)
	roots[0] = root
	roots = append(roots, replacements...)
	return manager.runBooleanOperation(roots, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		return manager.composeRecursive(root, replacements, work, make(map[extraBDDIteKey]ExtraBDDLit), make(map[ExtraBDDLit]ExtraBDDLit))
	})
}

func (manager *ExtraBDDManager) Compose(root, replacement ExtraBDDLit, variable int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if variable < 0 || variable >= manager.variables {
		return 0, ErrDimension
	}
	replacements := make([]ExtraBDDLit, manager.variables)
	for index := range replacements {
		replacements[index] = ExtraBDDLit(index+1) << 1
	}
	replacements[variable] = replacement
	return manager.VectorCompose(root, replacements, options)
}

func (manager *ExtraBDDManager) Permute(root ExtraBDDLit, permutation []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if len(permutation) != manager.variables {
		return 0, ErrDimension
	}
	seen := make([]bool, manager.variables)
	replacements := make([]ExtraBDDLit, manager.variables)
	for variable, image := range permutation {
		if image < 0 || image >= manager.variables || seen[image] {
			return 0, ErrDimension
		}
		seen[image] = true
		replacements[variable] = ExtraBDDLit(image+1) << 1
	}
	return manager.VectorCompose(root, replacements, options)
}

func (manager *ExtraBDDManager) SwapVariables(root ExtraBDDLit, first, second []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if len(first) != len(second) {
		return 0, ErrDimension
	}
	permutation := make([]int, manager.variables)
	for index := range permutation {
		permutation[index] = index
	}
	used := make([]bool, manager.variables)
	for index := range first {
		a, b := first[index], second[index]
		if a < 0 || a >= manager.variables || b < 0 || b >= manager.variables || a == b || used[a] || used[b] {
			return 0, ErrDimension
		}
		used[a], used[b] = true, true
		permutation[a], permutation[b] = b, a
	}
	return manager.Permute(root, permutation, options)
}

// AdjacentPermuteX preserves Cudd_bddAdjPermuteX: variables are interpreted
// in groups of three, and the first two variables of each complete group are
// swapped. A trailing group shorter than three is ignored.
func (manager *ExtraBDDManager) AdjacentPermuteX(root ExtraBDDLit, variables []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	permutation := make([]int, manager.variables)
	for variable := range permutation {
		permutation[variable] = variable
	}
	seen := make([]bool, manager.variables)
	for _, variable := range variables {
		if variable < 0 || variable >= manager.variables || seen[variable] {
			return 0, ErrDimension
		}
		seen[variable] = true
	}
	for index := 0; index < len(variables)-2; index += 3 {
		first, second := variables[index], variables[index+1]
		permutation[first], permutation[second] = second, first
	}
	return manager.Permute(root, permutation, options)
}
