package abc

// Shared-variable preparation and purity analysis from opt/dau/dauMerge.c.

import "slices"

func cloneDauDSDFormula(source *DauDSDFormula) *DauDSDFormula {
	if source == nil {
		return nil
	}
	result := &DauDSDFormula{nodes: make([]dauDSDNode, len(source.nodes)), root: source.root, variables: source.variables}
	for i, node := range source.nodes {
		result.nodes[i] = node
		result.nodes[i].children = slices.Clone(node.children)
		result.nodes[i].bindings = slices.Clone(node.bindings)
		result.nodes[i].primeBits = slices.Clone(node.primeBits)
	}
	return result
}

// dauDSDMergeRemap applies a source-variable-to-global-variable mapping while
// leaving the local template before a mux binding list untouched.
func dauDSDMergeRemap(source *DauDSDFormula, mapping []int, stop func() bool) (*DauDSDFormula, error) {
	if source == nil {
		return nil, ErrDimension
	}
	result := cloneDauDSDFormula(source)
	maximum := -1
	var visit func(int, bool) error
	visit = func(index int, local bool) error {
		if stop != nil && stop() {
			return ErrCancelled
		}
		node := &result.nodes[index]
		if node.kind == DauDSDVariable {
			if local {
				return nil
			}
			if node.value < 0 || node.value >= len(mapping) || mapping[node.value] < 0 || mapping[node.value] >= 26 {
				return ErrDimension
			}
			node.value = mapping[node.value]
			maximum = max(maximum, node.value)
			return nil
		}
		childLocal := local
		if node.kind == DauDSDMux && len(node.bindings) != 0 {
			childLocal = true
		}
		for _, child := range node.children {
			if err := visit(child, childLocal); err != nil {
				return err
			}
		}
		for _, child := range node.bindings {
			if err := visit(child, local); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(result.root, false); err != nil {
		return nil, err
	}
	result.variables = maximum + 1
	return result, nil
}

// dauDSDMergePresence reports global variables used by a formula. Variables in
// a mux-local template are intentionally excluded; its bindings are included.
func dauDSDMergePresence(formula *DauDSDFormula, variables int, stop func() bool) ([]bool, error) {
	if formula == nil || variables < 0 || variables > 26 {
		return nil, ErrDimension
	}
	present := make([]bool, variables)
	var visit func(int, bool) error
	visit = func(index int, local bool) error {
		if stop != nil && stop() {
			return ErrCancelled
		}
		node := &formula.nodes[index]
		if node.kind == DauDSDVariable {
			if local {
				return nil
			}
			if node.value < 0 || node.value >= variables {
				return ErrDimension
			}
			present[node.value] = true
			return nil
		}
		childLocal := local
		if node.kind == DauDSDMux && len(node.bindings) != 0 {
			childLocal = true
		}
		for _, child := range node.children {
			if err := visit(child, childLocal); err != nil {
				return err
			}
		}
		for _, child := range node.bindings {
			if err := visit(child, local); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(formula.root, false); err != nil {
		return nil, err
	}
	return present, nil
}

// dauDSDMergeMaps places shared variables first and then variables unique to
// either formula, retaining ascending original-variable order in each class.
func dauDSDMergeMaps(first, second []bool) (shared int, oldToNew, newToOld []int, err error) {
	if len(first) != len(second) || len(first) > 26 {
		return 0, nil, nil, ErrDimension
	}
	for i := range first {
		if first[i] && second[i] {
			shared++
		}
	}
	oldToNew = make([]int, len(first))
	for i := range oldToNew {
		oldToNew[i] = -1
	}
	newToOld = make([]int, 0, len(first))
	for _, wantShared := range []bool{true, false} {
		for variable := range first {
			used := first[variable] || second[variable]
			isShared := first[variable] && second[variable]
			if !used || isShared != wantShared {
				continue
			}
			oldToNew[variable] = len(newToOld)
			newToOld = append(newToOld, variable)
		}
	}
	return shared, oldToNew, newToOld, nil
}

// dauDSDMergeStatuses classifies each reachable expression like the source:
// 0 has no pure direct fanin, 1 has one, 2 has multiple but not all, and 3 has
// only pure fanins. A variable is pure when its remapped index is not shared.
// Skipped mux-local template nodes remain -1.
func dauDSDMergeStatuses(formula *DauDSDFormula, shared int, stop func() bool) ([]int, error) {
	if formula == nil || shared < 0 || shared > formula.variables {
		return nil, ErrDimension
	}
	status := make([]int, len(formula.nodes))
	for i := range status {
		status[i] = -1
	}
	var visit func(int) (int, error)
	visit = func(index int) (int, error) {
		if stop != nil && stop() {
			return 0, ErrCancelled
		}
		node := &formula.nodes[index]
		if node.kind == DauDSDVariable {
			if node.value < 0 || node.value >= formula.variables {
				return 0, ErrDimension
			}
			if node.value >= shared {
				status[index] = 3
			} else {
				status[index] = 0
			}
			return status[index], nil
		}
		if node.kind == DauDSDConstant {
			status[index] = 3
			return 3, nil
		}
		children := node.children
		if node.kind == DauDSDMux && len(node.bindings) != 0 {
			children = node.bindings
		}
		pure := 0
		for _, child := range children {
			childStatus, err := visit(child)
			if err != nil {
				return 0, err
			}
			if childStatus == 3 {
				pure++
			}
		}
		switch {
		case pure == 0:
			status[index] = 0
		case pure == 1:
			status[index] = 1
		case pure == len(children):
			status[index] = 3
		default:
			status[index] = 2
		}
		return status[index], nil
	}
	if _, err := visit(formula.root); err != nil {
		return nil, err
	}
	return status, nil
}
