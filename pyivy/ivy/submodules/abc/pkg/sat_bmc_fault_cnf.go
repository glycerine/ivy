package abc

// GIA CNF derivation and lifting from src/sat/bmc/bmcFault.c.

import "slices"

type GiaFaultCNFOptions struct {
	MaxObjects   int
	MaxVariables int
	MaxClauses   int
	MaxLiterals  int
	Stop         func() bool
}

type GiaFaultCNF struct {
	variables      int
	clauses        [][]SatLit
	objectLiterals []SatLit
	inputs         []SatLit
	outputs        []SatLit
}

func (c *GiaFaultCNF) VariableCount() int { return c.variables }
func (c *GiaFaultCNF) Clauses() [][]SatLit {
	result := make([][]SatLit, len(c.clauses))
	for i := range c.clauses {
		result[i] = slices.Clone(c.clauses[i])
	}
	return result
}
func (c *GiaFaultCNF) InputLiterals() []SatLit  { return slices.Clone(c.inputs) }
func (c *GiaFaultCNF) OutputLiterals() []SatLit { return slices.Clone(c.outputs) }
func (c *GiaFaultCNF) ObjectLiterals() []SatLit { return slices.Clone(c.objectLiterals) }
func (c *GiaFaultCNF) ObjectLiteral(id GiaID) (SatLit, error) {
	if c == nil || uint64(id) >= uint64(len(c.objectLiterals)) || c.objectLiterals[id] == InvalidSatLit {
		return InvalidSatLit, ErrInvalidLiteral
	}
	return c.objectLiterals[id], nil
}

func giaFaultCNFLimits(options GiaFaultCNFOptions) (GiaFaultCNFOptions, error) {
	if options.MaxObjects < 0 || options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
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
	return options, nil
}

func checkGiaFaultCNF(c *GiaFaultCNF, options GiaFaultCNFOptions) error {
	if c == nil || c.variables < 1 || c.variables > options.MaxVariables || uint64(c.variables) > uint64(MaxLiteralID)+1 || len(c.clauses) > options.MaxClauses || len(c.objectLiterals) > options.MaxObjects {
		if c == nil || c.variables < 1 {
			return ErrDimension
		}
		return ErrLimit
	}
	literals := 0
	for _, clause := range c.clauses {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if len(clause) > options.MaxLiterals-literals {
			return ErrLimit
		}
		literals += len(clause)
		for _, literal := range clause {
			if !literal.Valid() || uint64(literal.Var()) >= uint64(c.variables) {
				return ErrInvalidLiteral
			}
		}
	}
	return nil
}

func DeriveGiaFaultCNF(manager *GiaManager, options GiaFaultCNFOptions) (*GiaFaultCNF, error) {
	options, err := giaFaultCNFLimits(options)
	if err != nil || manager == nil {
		return nil, ErrDimension
	}
	if manager.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	aig, mapping, err := manager.ToAig()
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	relation, err := EncodeAigCNF(aig)
	if err != nil {
		return nil, err
	}
	result := &GiaFaultCNF{
		variables:      relation.variables,
		clauses:        relation.Clauses(),
		objectLiterals: make([]SatLit, manager.ObjectCount()),
	}
	for i := range result.objectLiterals {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		aigLiteral := mapping[i]
		literal, err := relation.ObjectLiteral(aigLiteral.ID())
		if err != nil {
			result.objectLiterals[i] = InvalidSatLit
			continue
		}
		result.objectLiterals[i] = literal
		if aigLiteral.Complemented() {
			result.objectLiterals[i] = literal.Not()
		}
	}
	for _, id := range manager.cis {
		result.inputs = append(result.inputs, result.objectLiterals[id])
	}
	for _, id := range manager.cos {
		result.outputs = append(result.outputs, result.objectLiterals[id])
	}
	if err = checkGiaFaultCNF(result, options); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *GiaFaultCNF) Lift(offset int, options GiaFaultCNFOptions) (*GiaFaultCNF, error) {
	options, err := giaFaultCNFLimits(options)
	if err != nil || offset < 0 {
		return nil, ErrDimension
	}
	if err = checkGiaFaultCNF(c, options); err != nil {
		return nil, err
	}
	variables := uint64(c.variables) + uint64(offset)
	if variables > uint64(options.MaxVariables) || variables > uint64(MaxLiteralID)+1 {
		return nil, ErrLimit
	}
	shift := func(literal SatLit) (SatLit, error) {
		if literal == InvalidSatLit {
			return InvalidSatLit, nil
		}
		variable := uint64(literal.Var()) + uint64(offset)
		if variable > uint64(MaxLiteralID) {
			return InvalidSatLit, ErrLimit
		}
		return SatLit(variable)<<1 | literal&1, nil
	}
	result := &GiaFaultCNF{variables: int(variables), objectLiterals: make([]SatLit, len(c.objectLiterals))}
	for _, clause := range c.clauses {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		copyClause := make([]SatLit, len(clause))
		for i, literal := range clause {
			copyClause[i], err = shift(literal)
			if err != nil {
				return nil, err
			}
		}
		result.clauses = append(result.clauses, copyClause)
	}
	for i, literal := range c.objectLiterals {
		result.objectLiterals[i], err = shift(literal)
		if err != nil {
			return nil, err
		}
	}
	for _, literal := range c.inputs {
		mapped, err := shift(literal)
		if err != nil {
			return nil, err
		}
		result.inputs = append(result.inputs, mapped)
	}
	for _, literal := range c.outputs {
		mapped, err := shift(literal)
		if err != nil {
			return nil, err
		}
		result.outputs = append(result.outputs, mapped)
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}

func (c *GiaFaultCNF) LoadInto(solver *SatSolver) (bool, error) {
	if c == nil {
		return false, ErrDimension
	}
	return (&AigCNF{variables: c.variables, clauses: c.clauses}).LoadInto(solver, 0)
}
