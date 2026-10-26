package abc

// Simple native-order encoders from sat/cnf/cnfWrite.c.

type AigSimpleCNFOptions struct {
	MaxObjects, MaxVariables, MaxClauses, MaxLiterals int // zero: 1M, 1M, 1M, 4M
	Stop                                              func() bool
}

type AigSimpleCNF struct {
	*AigMappedCNF
	exposedOutputs int
	retiming       bool
}

func (c *AigSimpleCNF) ExposedOutputCount() int {
	if c == nil {
		return 0
	}
	return c.exposedOutputs
}

func (c *AigSimpleCNF) RetimingInitialization() bool {
	return c != nil && c.retiming
}

// DeriveAigSimpleCNF preserves Cnf_DeriveSimple's raw variable and clause
// order. All outputs receive variables when exposedOutputs is nonzero, although
// only the final exposedOutputs are equated to their drivers; earlier outputs
// are asserted through driver literals. OutputLiterals returns meaningful
// driver literals for asserted outputs and CO variables for exposed outputs.
func DeriveAigSimpleCNF(g *AigManager, exposedOutputs int, o AigSimpleCNFOptions) (*AigSimpleCNF, error) {
	return deriveAigSimpleCNF(g, exposedOutputs, false, o)
}

// DeriveAigSimpleRetimingCNF preserves Cnf_DeriveSimpleForRetiming: every CO
// receives a variable, is equated to its driver, and is constrained to zero.
func DeriveAigSimpleRetimingCNF(g *AigManager, o AigSimpleCNFOptions) (*AigSimpleCNF, error) {
	if g == nil {
		return nil, ErrDimension
	}
	return deriveAigSimpleCNF(g, len(g.cos), true, o)
}

func deriveAigSimpleCNF(g *AigManager, exposedOutputs int, retiming bool, o AigSimpleCNFOptions) (*AigSimpleCNF, error) {
	if g == nil || exposedOutputs < 0 || exposedOutputs > len(g.cos) || o.MaxObjects < 0 || o.MaxVariables < 0 || o.MaxClauses < 0 || o.MaxLiterals < 0 {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxVariables == 0 {
		o.MaxVariables = 1 << 20
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = 1 << 20
	}
	if o.MaxLiterals == 0 {
		o.MaxLiterals = 4 << 20
	}
	stop := func() bool { return o.Stop != nil && o.Stop() }
	if stop() {
		return nil, ErrCancelled
	}
	if len(g.objects) > o.MaxObjects {
		return nil, ErrLimit
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	ands := 0
	for _, object := range g.objects {
		switch object.Type {
		case AigAND:
			ands++
		case AigBuffer, AigXOR, AigVoid:
			return nil, ErrUnsupported
		}
	}
	coVariables := 0
	if exposedOutputs > 0 || retiming {
		coVariables = len(g.cos)
	}
	variables, e := checkedSize(2, 1, coVariables+ands+len(g.cis), min(o.MaxVariables, int(MaxLiteralID)+1))
	if e != nil {
		return nil, e
	}
	clauses, literals := 0, 0
	if retiming {
		clauses, e = checkedSize(1, 3, ands+len(g.cos), o.MaxClauses)
		if e == nil {
			literals, e = checkedSize(1, 7, ands, o.MaxLiterals)
		}
		if e == nil {
			literals, e = checkedSize(literals, 5, len(g.cos), o.MaxLiterals)
		}
	} else {
		clauses, e = checkedSize(1, 3, ands, o.MaxClauses)
		if e == nil {
			clauses, e = checkedSize(clauses, 1, len(g.cos)+exposedOutputs, o.MaxClauses)
		}
		if e == nil {
			literals, e = checkedSize(1, 7, ands, o.MaxLiterals)
		}
		if e == nil {
			literals, e = checkedSize(literals, 1, len(g.cos)+3*exposedOutputs, o.MaxLiterals)
		}
	}
	if e != nil {
		return nil, e
	}
	_ = literals
	mapped := &AigMappedCNF{variables: variables, objectVariables: make([]int, len(g.objects)), clauses: make([][]SatLit, 0, clauses)}
	for i := range mapped.objectVariables {
		mapped.objectVariables[i] = -1
	}
	number := 1
	if coVariables != 0 {
		for _, id := range g.cos {
			mapped.objectVariables[id] = number
			number++
		}
	}
	for id, object := range g.objects {
		if object.Type == AigAND {
			mapped.objectVariables[id] = number
			number++
		}
	}
	for _, id := range g.cis {
		mapped.objectVariables[id] = number
		mapped.inputs = append(mapped.inputs, SatLit(number)<<1)
		number++
	}
	mapped.objectVariables[0] = number
	number++
	if number != variables {
		return nil, ErrInvariant
	}
	literal := func(signal AigLit) (SatLit, error) {
		variable := mapped.objectVariables[signal.ID()]
		if variable < 0 {
			return 0, ErrInvariant
		}
		return SatLit(variable)<<1 | SatLit(signal&1), nil
	}
	for id, object := range g.objects {
		if object.Type != AigAND {
			continue
		}
		if stop() {
			return nil, ErrCancelled
		}
		output := SatLit(mapped.objectVariables[id]) << 1
		first, e := literal(object.Fanin0)
		if e != nil {
			return nil, e
		}
		second, e := literal(object.Fanin1)
		if e != nil {
			return nil, e
		}
		mapped.clauses = append(mapped.clauses, []SatLit{output, first.Not(), second.Not()}, []SatLit{output.Not(), first}, []SatLit{output.Not(), second})
	}
	mapped.clauses = append(mapped.clauses, []SatLit{SatLit(mapped.objectVariables[0]) << 1})
	firstExposed := len(g.cos) - exposedOutputs
	for index, id := range g.cos {
		if stop() {
			return nil, ErrCancelled
		}
		driver, e := literal(g.objects[id].Fanin0)
		if e != nil {
			return nil, e
		}
		if retiming || index >= firstExposed {
			output := SatLit(mapped.objectVariables[id]) << 1
			mapped.clauses = append(mapped.clauses, []SatLit{output, driver.Not()}, []SatLit{output.Not(), driver})
			mapped.outputs = append(mapped.outputs, output)
			if retiming {
				mapped.clauses = append(mapped.clauses, []SatLit{output.Not()})
			}
		} else {
			mapped.clauses = append(mapped.clauses, []SatLit{driver})
			mapped.outputs = append(mapped.outputs, driver)
		}
	}
	if stop() {
		return nil, ErrCancelled
	}
	if len(mapped.clauses) != clauses {
		return nil, ErrInvariant
	}
	return &AigSimpleCNF{AigMappedCNF: mapped, exposedOutputs: exposedOutputs, retiming: retiming}, nil
}
