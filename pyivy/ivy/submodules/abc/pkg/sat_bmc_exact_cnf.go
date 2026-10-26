package abc

// Exact-synthesis structural and minterm CNF from src/sat/bmc/bmcMaj.c.

import "slices"

type ExactCNFOptions struct {
	MaxVariables, MaxClauses, MaxLiterals, MaxAssignments int
	OnlyAND, OrderNodes                                   bool
	Stop                                                  func() bool
}

func exactCNFLimits(options ExactCNFOptions) (ExactCNFOptions, error) {
	if options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxAssignments < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 1 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 4 << 20
	}
	if options.MaxAssignments == 0 {
		options.MaxAssignments = 1 << 20
	}
	return options, nil
}

func checkExactTopology(topology ExactSynthesisTopology, options ExactCNFOptions) error {
	if topology.Inputs < 1 || topology.InternalNodes < 1 || len(topology.Nodes) != topology.InternalNodes || topology.LUTSize < 2 || topology.LUTSize > 6 || topology.VariableCount < 1 || topology.VariableCount > options.MaxVariables {
		return ErrDimension
	}
	majority := topology.ConstantObjects == 2
	if majority && topology.LUTSize != 3 || !majority && topology.ConstantObjects != 0 {
		return ErrDimension
	}
	seen := make([]bool, topology.VariableCount)
	seen[0] = true
	offset := topology.Inputs + topology.ConstantObjects
	for row, node := range topology.Nodes {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if node.Object != offset+row || len(node.Pins) != topology.LUTSize {
			return ErrDimension
		}
		wantFunctions := 0
		if !majority {
			wantFunctions = (1 << topology.LUTSize) - 1
		}
		if len(node.FunctionVariables) != wantFunctions {
			return ErrDimension
		}
		for _, variable := range node.FunctionVariables {
			if int(variable) <= 0 || int(variable) >= len(seen) || seen[variable] {
				return ErrDimension
			}
			seen[variable] = true
		}
		for _, pin := range node.Pins {
			if len(pin.Candidates) == 0 || len(pin.Candidates) != len(pin.Variables) {
				return ErrDimension
			}
			candidateSeen := make([]bool, node.Object)
			for i, candidate := range pin.Candidates {
				variable := pin.Variables[i]
				if candidate < 0 || candidate >= node.Object || candidateSeen[candidate] || int(variable) <= 0 || int(variable) >= len(seen) || seen[variable] {
					return ErrDimension
				}
				candidateSeen[candidate], seen[variable] = true, true
			}
		}
	}
	for _, used := range seen {
		if !used {
			return ErrDimension
		}
	}
	return nil
}

type exactClauseBuilder struct {
	clauses              [][]SatLit
	literals, maxClauses int
	maxLiterals          int
	stop                 func() bool
}

func (b *exactClauseBuilder) add(literals ...SatLit) error {
	if b.stop != nil && b.stop() {
		return ErrCancelled
	}
	if len(b.clauses) >= b.maxClauses || len(literals) > b.maxLiterals-b.literals {
		return ErrLimit
	}
	b.clauses = append(b.clauses, slices.Clone(literals))
	b.literals += len(literals)
	return nil
}

type ExactSynthesisSession struct {
	Topology ExactSynthesisTopology
	Solver   *SatSolver
	Options  ExactCNFOptions
	Minterms []int
}

type ExactCNFStatistics struct {
	FirstVariable, Variables, Clauses, Literals int
	Inconsistent                                bool
}

func NewExactSynthesisSession(topology ExactSynthesisTopology, options ExactCNFOptions) (*ExactSynthesisSession, ExactCNFStatistics, error) {
	statistics := ExactCNFStatistics{}
	var err error
	options, err = exactCNFLimits(options)
	if err != nil {
		return nil, statistics, err
	}
	if err = checkExactTopology(topology, options); err != nil {
		return nil, statistics, err
	}
	builder := exactClauseBuilder{maxClauses: options.MaxClauses, maxLiterals: options.MaxLiterals, stop: options.Stop}
	fanouts := make([][]SatLit, topology.Inputs+topology.ConstantObjects+topology.InternalNodes)
	for row, node := range topology.Nodes {
		for pinIndex, pin := range node.Pins {
			atLeast := make([]SatLit, len(pin.Variables))
			for i, variable := range pin.Variables {
				atLeast[i] = satCardinalityLit(variable, false)
				fanouts[pin.Candidates[i]] = append(fanouts[pin.Candidates[i]], atLeast[i])
			}
			if err = builder.add(atLeast...); err != nil {
				return nil, statistics, err
			}
			for i := range pin.Variables {
				for j := i + 1; j < len(pin.Variables); j++ {
					if err = builder.add(satCardinalityLit(pin.Variables[i], true), satCardinalityLit(pin.Variables[j], true)); err != nil {
						return nil, statistics, err
					}
				}
			}
			if pinIndex+1 < len(node.Pins) {
				next := node.Pins[pinIndex+1]
				for i, source := range pin.Candidates {
					for j, nextSource := range next.Candidates {
						if nextSource >= source {
							if err = builder.add(satCardinalityLit(pin.Variables[i], true), satCardinalityLit(next.Variables[j], true)); err != nil {
								return nil, statistics, err
							}
						}
					}
				}
			}
		}
		if options.OrderNodes && row > 0 {
			first := node.Pins[0]
			for priorRow := 0; priorRow < row; priorRow++ {
				prior := topology.Nodes[priorRow].Pins[0]
				for i, source := range first.Candidates {
					for j, priorSource := range prior.Candidates {
						if priorSource > source {
							if err = builder.add(satCardinalityLit(first.Variables[i], true), satCardinalityLit(prior.Variables[j], true)); err != nil {
								return nil, statistics, err
							}
						}
					}
				}
			}
		}
		if topology.ConstantObjects == 0 && topology.LUTSize == 2 {
			f := node.FunctionVariables
			if err = builder.add(satCardinalityLit(f[0], false), satCardinalityLit(f[1], false), satCardinalityLit(f[2], true)); err != nil {
				return nil, statistics, err
			}
			if err = builder.add(satCardinalityLit(f[0], true), satCardinalityLit(f[1], false), satCardinalityLit(f[2], true)); err != nil {
				return nil, statistics, err
			}
			if err = builder.add(satCardinalityLit(f[0], false), satCardinalityLit(f[1], true), satCardinalityLit(f[2], true)); err != nil {
				return nil, statistics, err
			}
			if options.OnlyAND {
				if err = builder.add(satCardinalityLit(f[0], true), satCardinalityLit(f[1], true), satCardinalityLit(f[2], false)); err != nil {
					return nil, statistics, err
				}
			}
		}
	}
	firstUsed := topology.ConstantObjects
	if firstUsed == 2 {
		firstUsed = 2
	}
	for object := firstUsed; object < len(fanouts)-1; object++ {
		if len(fanouts[object]) == 0 {
			return nil, statistics, ErrDimension
		}
		if err = builder.add(fanouts[object]...); err != nil {
			return nil, statistics, err
		}
	}
	solver := NewSatSolver()
	if err = solver.SetVariableCount(topology.VariableCount); err != nil {
		return nil, statistics, err
	}
	inconsistent := false
	for _, clause := range builder.clauses {
		if inconsistent {
			break
		}
		var ok bool
		ok, err = solver.AddClause(clause)
		if err != nil {
			return nil, statistics, err
		}
		inconsistent = !ok
	}
	statistics = ExactCNFStatistics{Variables: topology.VariableCount, Clauses: len(builder.clauses), Literals: builder.literals, Inconsistent: inconsistent}
	return &ExactSynthesisSession{Topology: topology.Clone(), Solver: solver, Options: options}, statistics, nil
}

func exactMintermSource(topology ExactSynthesisTopology, nodeOutputs []SatVar, source, assignment int) (fixed bool, value bool, variable SatVar, err error) {
	if topology.ConstantObjects == 2 {
		if source == 0 {
			return true, false, 0, nil
		}
		if source == 1 {
			return true, true, 0, nil
		}
		if source < topology.Inputs+2 {
			return true, assignment&(1<<(source-2)) != 0, 0, nil
		}
		source -= topology.Inputs + 2
	} else if source < topology.Inputs {
		return true, assignment&(1<<source) != 0, 0, nil
	} else {
		source -= topology.Inputs
	}
	if source < 0 || source >= len(nodeOutputs) {
		return false, false, 0, ErrDimension
	}
	return false, false, nodeOutputs[source], nil
}

func (s *ExactSynthesisSession) AddMinterm(assignment int, target bool) (ExactCNFStatistics, error) {
	statistics := ExactCNFStatistics{}
	if s == nil || s.Solver == nil || s.Topology.Inputs >= 63 || assignment < 0 || uint64(assignment) >= uint64(1)<<s.Topology.Inputs {
		return statistics, ErrDimension
	}
	for _, old := range s.Minterms {
		if old == assignment {
			return statistics, ErrInvariant
		}
	}
	perNode := s.Topology.LUTSize + 1
	first := s.Solver.VariableCount()
	count := perNode * len(s.Topology.Nodes)
	if count > s.Options.MaxVariables-first {
		return statistics, ErrLimit
	}
	pinVariables := make([][]SatVar, len(s.Topology.Nodes))
	outputs := make([]SatVar, len(s.Topology.Nodes))
	for row := range s.Topology.Nodes {
		pinVariables[row] = make([]SatVar, s.Topology.LUTSize)
		for pin := range pinVariables[row] {
			pinVariables[row][pin] = SatVar(first + row*perNode + pin)
		}
		outputs[row] = SatVar(first + row*perNode + s.Topology.LUTSize)
	}
	builder := exactClauseBuilder{maxClauses: s.Options.MaxClauses, maxLiterals: s.Options.MaxLiterals, stop: s.Options.Stop}
	for row, node := range s.Topology.Nodes {
		for pinIndex, pin := range node.Pins {
			for index, source := range pin.Candidates {
				fixed, value, sourceVariable, err := exactMintermSource(s.Topology, outputs[:row], source, assignment)
				if err != nil {
					return statistics, err
				}
				selection := satCardinalityLit(pin.Variables[index], true)
				if fixed {
					if err = builder.add(selection, satCardinalityLit(pinVariables[row][pinIndex], !value)); err != nil {
						return statistics, err
					}
				} else {
					if err = builder.add(selection, satCardinalityLit(pinVariables[row][pinIndex], false), satCardinalityLit(sourceVariable, true)); err != nil {
						return statistics, err
					}
					if err = builder.add(selection, satCardinalityLit(pinVariables[row][pinIndex], true), satCardinalityLit(sourceVariable, false)); err != nil {
						return statistics, err
					}
				}
			}
		}
		final := row == len(s.Topology.Nodes)-1
		if s.Topology.ConstantObjects == 2 {
			for value := 0; value < 2; value++ {
				if final && (value != 0) == target {
					continue
				}
				for omitted := 0; omitted < 3; omitted++ {
					clause := make([]SatLit, 0, 3)
					for pin := 0; pin < 3; pin++ {
						if pin != omitted {
							clause = append(clause, satCardinalityLit(pinVariables[row][pin], value != 0))
						}
					}
					if !final {
						clause = append(clause, satCardinalityLit(outputs[row], value == 0))
					}
					if err := builder.add(clause...); err != nil {
						return statistics, err
					}
				}
			}
		} else {
			mask := (1 << s.Topology.LUTSize) - 1
			for value := 0; value < 2; value++ {
				if final && (value != 0) == target {
					continue
				}
				for pattern := 0; pattern <= mask; pattern++ {
					if pattern == 0 && value == 1 {
						continue
					}
					clause := make([]SatLit, 0, s.Topology.LUTSize+2)
					for pin := 0; pin < s.Topology.LUTSize; pin++ {
						clause = append(clause, satCardinalityLit(pinVariables[row][pin], pattern&(1<<pin) != 0))
					}
					if !final {
						clause = append(clause, satCardinalityLit(outputs[row], value == 0))
					}
					if pattern > 0 {
						clause = append(clause, satCardinalityLit(node.FunctionVariables[pattern-1], value != 0))
					}
					if err := builder.add(clause...); err != nil {
						return statistics, err
					}
				}
			}
		}
	}
	trial := s.Solver.cloneState()
	trial.bookmark = s.Solver.bookmark
	if err := trial.SetVariableCount(first + count); err != nil {
		return statistics, err
	}
	inconsistent := trial.inconsistent
	for _, clause := range builder.clauses {
		if inconsistent {
			break
		}
		ok, err := trial.AddClause(clause)
		if err != nil {
			return statistics, err
		}
		inconsistent = !ok
	}
	*s.Solver = trial
	s.Minterms = append(s.Minterms, assignment)
	statistics = ExactCNFStatistics{FirstVariable: first, Variables: count, Clauses: len(builder.clauses), Literals: builder.literals, Inconsistent: inconsistent}
	return statistics, nil
}

func (s *ExactSynthesisSession) Decode(model []bool, complementOutput bool) (ExactNetwork, error) {
	if s == nil || len(model) < s.Topology.VariableCount {
		return ExactNetwork{}, ErrDimension
	}
	result := ExactNetwork{Inputs: s.Topology.Inputs, ComplementOutput: complementOutput, Gates: make([]ExactGate, len(s.Topology.Nodes))}
	for row, node := range s.Topology.Nodes {
		gate := ExactGate{Fanins: make([]int, len(node.Pins)), Majority: s.Topology.ConstantObjects == 2}
		for pinIndex, pin := range node.Pins {
			found := -1
			for i, variable := range pin.Variables {
				if int(variable) >= len(model) {
					return ExactNetwork{}, ErrDimension
				}
				if model[variable] {
					if found != -1 {
						return ExactNetwork{}, ErrInvariant
					}
					found = pin.Candidates[i]
				}
			}
			if found == -1 {
				return ExactNetwork{}, ErrInvariant
			}
			if s.Topology.ConstantObjects == 2 {
				switch found {
				case 0:
					found = ExactConst0
				case 1:
					found = ExactConst1
				default:
					found -= 2
				}
			}
			gate.Fanins[pinIndex] = found
		}
		for i, variable := range node.FunctionVariables {
			if int(variable) >= len(model) {
				return ExactNetwork{}, ErrDimension
			}
			if model[variable] {
				gate.Truth |= uint64(1) << (i + 1)
			}
		}
		result.Gates[row] = gate
	}
	if err := result.Check(ExactSynthesisLimits{}); err != nil {
		return ExactNetwork{}, err
	}
	return result, nil
}
