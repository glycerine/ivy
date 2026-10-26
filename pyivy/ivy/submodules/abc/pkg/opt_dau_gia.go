package abc

// DSD-to-GIA lowering from opt/dau/dauGia.c.

import "sort"

type DauGiaOptions struct {
	NativeGates, Balance bool
	MaxObjects           int
	MaxBytes             int
	MaxNodes             int
	MaxDepth             int
	MaxWords             int
	Stop                 func() bool
}

type DauGiaResult struct {
	Graph      *GiaManager
	Root       GiaLit
	PrimeCount int
}

type dauGiaBuilder struct {
	graph      *GiaManager
	options    DauGiaOptions
	primeCount int
	retainHash bool
}

// WithDauDSD appends a parsed DSD formula to an owned graph clone. Structural
// hashing is used during construction; the receiver, formula and inputs remain
// unchanged, and every error returns an empty result rather than a partial graph.
func (m *GiaManager) WithDauDSD(text string, inputs []GiaLit, options DauGiaOptions) (DauGiaResult, error) {
	var result DauGiaResult
	if m == nil || options.MaxObjects < 0 || options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxDepth < 0 || options.MaxWords < 0 {
		return result, ErrDimension
	}
	if err := m.Check(); err != nil {
		return result, err
	}
	formula, err := ParseDauDSD(text, DauDSDOptions{
		MaxBytes: options.MaxBytes, MaxNodes: options.MaxNodes, MaxDepth: options.MaxDepth,
		MaxVariables: DauMaxVariables, MaxWords: options.MaxWords, Stop: options.Stop,
	})
	if err != nil {
		return result, err
	}
	if len(inputs) < formula.VariableCount() || len(inputs) > DauMaxVariables {
		return result, ErrDimension
	}
	for _, input := range inputs {
		if !m.validSignal(input) {
			return result, ErrInvalidLiteral
		}
	}
	if options.Stop != nil && options.Stop() {
		return result, ErrCancelled
	}
	limit := options.MaxObjects
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return result, ErrLimit
	}
	target := m.Clone()
	target.objectLimit = min(limit, int(MaxGiaID)+1)
	retainHash := target.Hashing()
	if !retainHash {
		if err = target.HashStart(); err != nil {
			return result, err
		}
	}
	builder := dauGiaBuilder{graph: target, options: options, retainHash: retainHash}
	root, err := builder.buildFormula(formula, formula.root, inputs)
	if err != nil {
		return result, err
	}
	if options.Stop != nil && options.Stop() {
		return result, ErrCancelled
	}
	if !retainHash {
		target.HashStop()
	}
	if err = target.Check(); err != nil {
		return result, err
	}
	return DauGiaResult{Graph: target, Root: root, PrimeCount: builder.primeCount}, nil
}

func (b *dauGiaBuilder) stopped() bool {
	return b.options.Stop != nil && b.options.Stop()
}

func (b *dauGiaBuilder) and(left, right GiaLit) (GiaLit, error) {
	return b.graph.HashAnd(left, right)
}

func (b *dauGiaBuilder) xor(left, right GiaLit) (GiaLit, error) {
	if b.options.NativeGates {
		return b.graph.HashXorReal(left, right)
	}
	return b.graph.HashXor(left, right)
}

func (b *dauGiaBuilder) mux(control, whenTrue, whenFalse GiaLit) (GiaLit, error) {
	if b.options.NativeGates {
		return b.graph.HashMuxReal(control, whenTrue, whenFalse)
	}
	return b.graph.HashMux(control, whenTrue, whenFalse)
}

func (b *dauGiaBuilder) combine(fanins []GiaLit, xor bool) (GiaLit, error) {
	if len(fanins) == 0 {
		if xor {
			return GiaConst0, nil
		}
		return GiaConst1, nil
	}
	work := append([]GiaLit(nil), fanins...)
	combine := b.and
	if xor {
		combine = b.xor
	}
	for len(work) > 1 {
		if b.stopped() {
			return InvalidGiaLit, ErrCancelled
		}
		if b.options.Balance {
			levels, _, err := b.graph.Levels(0)
			if err != nil {
				return InvalidGiaLit, err
			}
			sort.SliceStable(work, func(i, j int) bool {
				left, right := levels[work[i].ID()], levels[work[j].ID()]
				return left > right
			})
		}
		last := len(work) - 1
		joined, err := combine(work[last-1], work[last])
		if err != nil {
			return InvalidGiaLit, err
		}
		work = append(work[:last-1], joined)
	}
	return work[0], nil
}

func (b *dauGiaBuilder) buildFormula(formula *DauDSDFormula, index int, environment []GiaLit) (GiaLit, error) {
	if b.stopped() {
		return InvalidGiaLit, ErrCancelled
	}
	node := formula.nodes[index]
	var result GiaLit
	var err error
	switch node.kind {
	case DauDSDConstant:
		result = GiaConst0.NotIf(node.value != 0)
	case DauDSDVariable:
		if node.value < 0 || node.value >= len(environment) {
			return InvalidGiaLit, ErrDimension
		}
		result = environment[node.value]
	case DauDSDAnd, DauDSDXor:
		fanins := make([]GiaLit, len(node.children))
		for i, child := range node.children {
			fanins[i], err = b.buildFormula(formula, child, environment)
			if err != nil {
				return InvalidGiaLit, err
			}
		}
		result, err = b.combine(fanins, node.kind == DauDSDXor)
	case DauDSDMux:
		local := environment
		if len(node.bindings) != 0 {
			local = make([]GiaLit, len(node.bindings))
			for i, child := range node.bindings {
				local[i], err = b.buildFormula(formula, child, environment)
				if err != nil {
					return InvalidGiaLit, err
				}
			}
		}
		parts := [3]GiaLit{}
		for i, child := range node.children {
			parts[i], err = b.buildFormula(formula, child, local)
			if err != nil {
				return InvalidGiaLit, err
			}
		}
		result, err = b.mux(parts[0], parts[1], parts[2])
	case DauDSDPrime:
		fanins := make([]GiaLit, len(node.children))
		for i, child := range node.children {
			fanins[i], err = b.buildFormula(formula, child, environment)
			if err != nil {
				return InvalidGiaLit, err
			}
		}
		b.primeCount++
		result, err = b.buildPrime(node.primeBits, fanins)
	default:
		return InvalidGiaLit, ErrInvariant
	}
	if err != nil {
		return InvalidGiaLit, err
	}
	return result.NotIf(node.complement), nil
}

func (b *dauGiaBuilder) buildPrime(values []bool, fanins []GiaLit) (GiaLit, error) {
	if b.stopped() {
		return InvalidGiaLit, ErrCancelled
	}
	if len(values) != 1<<uint(len(fanins)) {
		return InvalidGiaLit, ErrInvariant
	}
	allFalse, allTrue := true, true
	for _, value := range values {
		allFalse = allFalse && !value
		allTrue = allTrue && value
	}
	if allFalse {
		return GiaConst0, nil
	}
	if allTrue {
		return GiaConst1, nil
	}
	if len(fanins) == 0 {
		return InvalidGiaLit, ErrInvariant
	}
	variable := len(fanins) - 1
	half := len(values) / 2
	whenFalse, err := b.buildPrime(values[:half], fanins[:variable])
	if err != nil {
		return InvalidGiaLit, err
	}
	whenTrue, err := b.buildPrime(values[half:], fanins[:variable])
	if err != nil {
		return InvalidGiaLit, err
	}
	return b.mux(fanins[variable], whenTrue, whenFalse)
}
