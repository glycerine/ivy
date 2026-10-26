package abc

// Atomic GIA lowering and in-memory library construction from dauNpn2.c.

type DauLogicGiaOptions struct {
	NativeGates bool
	MaxObjects  int
	Parse       DauLogicFormulaOptions
	Stop        func() bool
}

type DauLogicGiaResult struct {
	Graph *GiaManager
	Root  GiaLit
}

func buildDauLogicGia(graph *GiaManager, formula *DauLogicFormula, inputs []GiaLit, native bool, stop func() bool) (GiaLit, error) {
	values := make([]GiaLit, len(formula.nodes))
	for index, node := range formula.nodes {
		if stop != nil && stop() {
			return InvalidGiaLit, ErrCancelled
		}
		var value GiaLit
		var err error
		switch node.kind {
		case dauLogicVariable:
			if node.variable < 0 || node.variable >= len(inputs) {
				return InvalidGiaLit, ErrDimension
			}
			value = inputs[node.variable]
		case dauLogicAnd:
			value, err = graph.HashAnd(values[node.left], values[node.right])
		case dauLogicXor:
			if native {
				value, err = graph.HashXorReal(values[node.left], values[node.right])
			} else {
				value, err = graph.HashXor(values[node.left], values[node.right])
			}
		default:
			return InvalidGiaLit, ErrInvariant
		}
		if err != nil {
			return InvalidGiaLit, err
		}
		values[index] = value.NotIf(node.complement)
	}
	return values[formula.root], nil
}

func (manager *GiaManager) WithDauLogicFormula(text string, inputs []GiaLit, options DauLogicGiaOptions) (DauLogicGiaResult, error) {
	var result DauLogicGiaResult
	if manager == nil || options.MaxObjects < 0 {
		return result, ErrDimension
	}
	if err := manager.Check(); err != nil {
		return result, err
	}
	parse := options.Parse
	if parse.Stop == nil {
		parse.Stop = options.Stop
	}
	formula, err := ParseDauLogicFormula(text, parse)
	if err != nil {
		return result, err
	}
	if len(inputs) < formula.variables {
		return result, ErrDimension
	}
	for _, input := range inputs {
		if !manager.validSignal(input) {
			return result, ErrInvalidLiteral
		}
	}
	limit := options.MaxObjects
	if limit == 0 {
		limit = 1 << 20
	}
	if len(manager.objects) > limit {
		return result, ErrLimit
	}
	target := manager.Clone()
	target.objectLimit = min(limit, int(MaxGiaID)+1)
	retainHash := target.Hashing()
	if !retainHash {
		if err = target.HashStart(); err != nil {
			return result, err
		}
	}
	root, err := buildDauLogicGia(target, formula, inputs, options.NativeGates, options.Stop)
	if err != nil {
		return result, err
	}
	if !retainHash {
		target.HashStop()
	}
	if err = target.Check(); err != nil {
		return result, err
	}
	return DauLogicGiaResult{Graph: target, Root: root}, nil
}

func BuildDauLogicLibrary(data []byte, variables int, options DauLogicGiaOptions) (*GiaManager, error) {
	parse := options.Parse
	if parse.Stop == nil {
		parse.Stop = options.Stop
	}
	_, formulas, err := parseDauLogicLibrary(data, variables, parse)
	if err != nil {
		return nil, err
	}
	limit := options.MaxObjects
	if limit == 0 {
		limit = 1 << 20
	}
	if limit < variables+1 {
		return nil, ErrLimit
	}
	graph, err := NewGiaManagerWithLimit(min(limit, int(MaxGiaID)+1))
	if err != nil {
		return nil, err
	}
	inputs := make([]GiaLit, variables)
	for index := range inputs {
		inputs[index], err = graph.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	if err = graph.HashStart(); err != nil {
		return nil, err
	}
	for _, formula := range formulas {
		root, err := buildDauLogicGia(graph, formula, inputs, options.NativeGates, options.Stop)
		if err != nil {
			return nil, err
		}
		if _, err = graph.AppendCO(root); err != nil {
			return nil, err
		}
	}
	graph.HashStop()
	if err = graph.Check(); err != nil {
		return nil, err
	}
	return graph, nil
}
