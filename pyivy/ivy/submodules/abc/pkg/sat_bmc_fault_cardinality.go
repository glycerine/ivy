package abc

// SAT cardinality encodings from src/sat/bmc/bmcFault.c.

import "slices"

type SatCardinalityEncoding uint8

const (
	SatCardinalityPairwise SatCardinalityEncoding = iota
	SatCardinalityGeneral
	SatCardinalityAtMostOneTree
)

type SatCardinalityOptions struct {
	MaxVariables int
	MaxClauses   int
	MaxLiterals  int
	Stop         func() bool
}

type SatCardinalityResult struct {
	Encoding           SatCardinalityEncoding
	Inputs             []SatVar
	Outputs            []SatVar
	AuxiliaryVariables int
	Clauses            int
	Literals           int
	Inconsistent       bool
}

type satCardinalityBuilder struct {
	next, maxVariables, maxClauses, maxLiterals int
	stop                                        func() bool
	clauses                                     [][]SatLit
	literals                                    int
}

func (b *satCardinalityBuilder) stopped() bool {
	return b.stop != nil && b.stop()
}

func (b *satCardinalityBuilder) variable() (SatVar, error) {
	if b.stopped() {
		return 0, ErrCancelled
	}
	if b.next >= b.maxVariables || uint64(b.next) > uint64(MaxLiteralID) {
		return 0, ErrLimit
	}
	v := SatVar(b.next)
	b.next++
	return v, nil
}

func (b *satCardinalityBuilder) clause(literals ...SatLit) error {
	if b.stopped() {
		return ErrCancelled
	}
	if len(b.clauses) >= b.maxClauses || len(literals) > b.maxLiterals-b.literals {
		return ErrLimit
	}
	b.literals += len(literals)
	b.clauses = append(b.clauses, slices.Clone(literals))
	return nil
}

func satCardinalityLit(variable SatVar, negative bool) SatLit {
	literal, _ := MakeSatLit(variable, negative)
	return literal
}

func (b *satCardinalityBuilder) and(a, c SatVar) (SatVar, error) {
	z, err := b.variable()
	if err != nil {
		return 0, err
	}
	for _, clause := range [][]SatLit{
		{satCardinalityLit(z, true), satCardinalityLit(a, false)},
		{satCardinalityLit(z, true), satCardinalityLit(c, false)},
		{satCardinalityLit(z, false), satCardinalityLit(a, true), satCardinalityLit(c, true)},
	} {
		if err = b.clause(clause...); err != nil {
			return 0, err
		}
	}
	return z, nil
}

func (b *satCardinalityBuilder) or(a, c SatVar) (SatVar, error) {
	z, err := b.variable()
	if err != nil {
		return 0, err
	}
	for _, clause := range [][]SatLit{
		{satCardinalityLit(z, false), satCardinalityLit(a, true)},
		{satCardinalityLit(z, false), satCardinalityLit(c, true)},
		{satCardinalityLit(z, true), satCardinalityLit(a, false), satCardinalityLit(c, false)},
	} {
		if err = b.clause(clause...); err != nil {
			return 0, err
		}
	}
	return z, nil
}

func (b *satCardinalityBuilder) buffer(a SatVar) (SatVar, error) {
	z, err := b.variable()
	if err != nil {
		return 0, err
	}
	if err = b.clause(satCardinalityLit(z, true), satCardinalityLit(a, false)); err != nil {
		return 0, err
	}
	if err = b.clause(satCardinalityLit(z, false), satCardinalityLit(a, true)); err != nil {
		return 0, err
	}
	return z, nil
}

func (b *satCardinalityBuilder) sorter(outputs []SatVar, i, k int) error {
	high, err := b.or(outputs[i], outputs[k])
	if err != nil {
		return err
	}
	low, err := b.and(outputs[i], outputs[k])
	if err != nil {
		return err
	}
	outputs[i], outputs[k] = high, low
	return nil
}

func (b *satCardinalityBuilder) pairwise(inputs []SatVar) ([]SatVar, error) {
	size := 1
	for size < len(inputs) {
		if size > int(MaxLiteralID)/2 {
			return nil, ErrLimit
		}
		size *= 2
	}
	outputs := make([]SatVar, size)
	copy(outputs, inputs)
	if size != len(inputs) {
		zero, err := b.variable()
		if err != nil {
			return nil, err
		}
		if err = b.clause(satCardinalityLit(zero, true)); err != nil {
			return nil, err
		}
		for i := len(inputs); i < size; i++ {
			outputs[i] = zero
		}
	}
	var merge func(int, int, int) error
	merge = func(lo, hi, stride int) error {
		step := stride * 2
		if step < hi-lo {
			if err := merge(lo, hi-stride, step); err != nil {
				return err
			}
			if err := merge(lo+stride, hi, step); err != nil {
				return err
			}
			for i := lo + stride; i < hi-stride; i += step {
				if err := b.sorter(outputs, i, i+stride); err != nil {
					return err
				}
			}
		}
		return nil
	}
	var build func(int, int) error
	build = func(lo, hi int) error {
		if hi-lo < 1 {
			return nil
		}
		mid := lo + (hi-lo)/2
		for i := lo; i <= mid; i++ {
			if err := b.sorter(outputs, i, i+(hi-lo+1)/2); err != nil {
				return err
			}
		}
		if err := build(lo, mid); err != nil {
			return err
		}
		if err := build(mid+1, hi); err != nil {
			return err
		}
		return merge(lo, hi, 1)
	}
	if err := build(0, size-1); err != nil {
		return nil, err
	}
	return outputs[:len(inputs)], nil
}

func (b *satCardinalityBuilder) general(inputs []SatVar) ([]SatVar, error) {
	current := slices.Clone(inputs)
	for level := 0; level < len(inputs); level++ {
		next := make([]SatVar, len(inputs))
		start := level & 1
		if start != 0 {
			v, err := b.buffer(current[0])
			if err != nil {
				return nil, err
			}
			next[0] = v
		}
		k := start
		for ; k+1 < len(inputs); k += 2 {
			high, err := b.or(current[k], current[k+1])
			if err != nil {
				return nil, err
			}
			low, err := b.and(current[k], current[k+1])
			if err != nil {
				return nil, err
			}
			next[k], next[k+1] = high, low
		}
		if k == len(inputs)-1 {
			v, err := b.buffer(current[k])
			if err != nil {
				return nil, err
			}
			next[k] = v
		}
		current = next
	}
	return current, nil
}

func (b *satCardinalityBuilder) atMostOneTree(inputs []SatVar, exact bool) ([]SatVar, error) {
	work := slices.Clone(inputs)
	for len(work) > 1 {
		next := make([]SatVar, 0, (len(work)+1)/2)
		for i := 0; i+1 < len(work); i += 2 {
			if err := b.clause(satCardinalityLit(work[i], true), satCardinalityLit(work[i+1], true)); err != nil {
				return nil, err
			}
			parent, err := b.or(work[i], work[i+1])
			if err != nil {
				return nil, err
			}
			next = append(next, parent)
		}
		if len(work)&1 != 0 {
			next = append(next, work[len(work)-1])
		}
		work = next
	}
	if exact {
		if err := b.clause(satCardinalityLit(work[0], false)); err != nil {
			return nil, err
		}
	}
	return work, nil
}

// AddSatCardinalityConstraint appends an at-most-k or exactly-k constraint.
// Inputs and returned slices are owned. Construction is atomic on validation,
// cancellation, and resource errors. Pairwise and General return descending
// threshold outputs; AtMostOneTree accepts k=1 and returns its OR root.
func AddSatCardinalityConstraint(solver *SatSolver, variables []SatVar, k int, exact bool, encoding SatCardinalityEncoding, options SatCardinalityOptions) (SatCardinalityResult, error) {
	result := SatCardinalityResult{Encoding: encoding}
	if solver == nil || solver.solving || len(variables) == 0 || k < 0 || k > len(variables) || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return result, ErrDimension
	}
	if encoding > SatCardinalityAtMostOneTree || encoding == SatCardinalityAtMostOneTree && k != 1 {
		return result, ErrDimension
	}
	seen := make([]bool, solver.VariableCount())
	for _, variable := range variables {
		if uint64(variable) >= uint64(len(seen)) || seen[variable] {
			return result, ErrDimension
		}
		seen[variable] = true
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = solver.maxVariables
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 1 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 4 << 20
	}
	if options.MaxVariables < solver.VariableCount() {
		return result, ErrLimit
	}
	options.MaxVariables = min(options.MaxVariables, solver.maxVariables)
	builder := satCardinalityBuilder{
		next: solver.VariableCount(), maxVariables: options.MaxVariables,
		maxClauses: options.MaxClauses, maxLiterals: options.MaxLiterals, stop: options.Stop,
	}
	if builder.stopped() {
		return result, ErrCancelled
	}
	var outputs []SatVar
	var err error
	switch encoding {
	case SatCardinalityPairwise:
		outputs, err = builder.pairwise(variables)
	case SatCardinalityGeneral:
		outputs, err = builder.general(variables)
	case SatCardinalityAtMostOneTree:
		outputs, err = builder.atMostOneTree(variables, exact)
	}
	if err != nil {
		return result, err
	}
	if encoding != SatCardinalityAtMostOneTree {
		if k < len(variables) {
			err = builder.clause(satCardinalityLit(outputs[k], true))
		}
		if err == nil && exact && k > 0 {
			err = builder.clause(satCardinalityLit(outputs[k-1], false))
		}
		if err != nil {
			return result, err
		}
	}
	if builder.stopped() {
		return result, ErrCancelled
	}
	trial := solver.cloneState()
	trial.bookmark = solver.bookmark
	if err = trial.SetVariableCount(builder.next); err != nil {
		return result, err
	}
	inconsistent := trial.inconsistent
	for _, clause := range builder.clauses {
		if inconsistent {
			break
		}
		var ok bool
		ok, err = trial.AddClause(clause)
		if err != nil {
			return result, err
		}
		inconsistent = !ok
	}
	*solver = trial
	result.Inputs = slices.Clone(variables)
	result.Outputs = slices.Clone(outputs)
	result.AuxiliaryVariables = builder.next - len(seen)
	result.Clauses = len(builder.clauses)
	result.Literals = builder.literals
	result.Inconsistent = inconsistent
	return result, nil
}
