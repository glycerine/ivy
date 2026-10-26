package abc

// Derived from ABC src/aig/aig/aigOper.c (Alan Mishchenko).

type AigOperationOptions struct {
	MaxOperations int // zero: 1M high-level gates
	Stop          func() bool
}

type aigOperationBudget struct {
	options AigOperationOptions
	used    int
}

func newAigOperationBudget(options AigOperationOptions) (*aigOperationBudget, error) {
	if options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 20
	}
	return &aigOperationBudget{options: options}, nil
}

func (b *aigOperationBudget) step() error {
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	if b.used == b.options.MaxOperations {
		return ErrLimit
	}
	b.used++
	return nil
}

// And implements Aig_And with the default additional-strashing and special
// XOR-node modes disabled. Fanins are ordered by regular object ID, as in C.
func (m *AigManager) And(a, b AigLit) (AigLit, error) {
	if m == nil || !m.validSignal(a) || !m.validSignal(b) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	switch {
	case a == b:
		return a, nil
	case a == b.Not():
		return AigConst0, nil
	case a == AigConst0 || b == AigConst0:
		return AigConst0, nil
	case a == AigConst1:
		return b, nil
	case b == AigConst1:
		return a, nil
	}
	if a.ID() > b.ID() {
		a, b = b, a
	}
	key := aigHashKey{a, b}
	if id, found := m.hash[key]; found {
		return AigLit(id) << 1, nil
	}
	if err := m.room(); err != nil {
		return InvalidAigLit, err
	}
	if m.objects[a.ID()].RefCount == ^uint32(0) || m.objects[b.ID()].RefCount == ^uint32(0) {
		return InvalidAigLit, ErrLimit
	}
	level := 1 + max(m.objects[a.ID()].Level, m.objects[b.ID()].Level)
	id := m.appendObject(AigObject{Type: AigAND, Fanin0: a, Fanin1: b, Level: level, Phase: m.phase(a) && m.phase(b)})
	m.objects[a.ID()].RefCount++
	m.objects[b.ID()].RefCount++
	m.hash[key] = id
	return AigLit(id) << 1, nil
}

// Operate implements Aig_Oper for the two native gate operations.
func (m *AigManager) Operate(a, b AigLit, operation AigType) (AigLit, error) {
	switch operation {
	case AigAND:
		return m.And(a, b)
	case AigXOR:
		return m.Xor(a, b)
	default:
		return InvalidAigLit, ErrDimension
	}
}

// LookupAnd implements Aig_TableLookupInt without creating a node. Found is
// true for a trivial result or an existing structural-hash entry.
func (m *AigManager) LookupAnd(a, b AigLit) (result AigLit, found bool, err error) {
	if m == nil || !m.validSignal(a) || !m.validSignal(b) {
		return InvalidAigLit, false, ErrInvalidLiteral
	}
	switch {
	case a == b:
		return a, true, nil
	case a == AigConst0 || b == AigConst0 || a == b.Not():
		return AigConst0, true, nil
	case a == AigConst1:
		return b, true, nil
	case b == AigConst1:
		return a, true, nil
	}
	if a.ID() > b.ID() {
		a, b = b, a
	}
	id, found := m.hash[aigHashKey{a, b}]
	if !found {
		return InvalidAigLit, false, nil
	}
	return AigLit(id) << 1, true, nil
}

func (m *AigManager) Or(a, b AigLit) (AigLit, error) {
	l, err := m.And(a.Not(), b.Not())
	if err != nil {
		return InvalidAigLit, err
	}
	return l.Not(), nil
}
func (m *AigManager) Xor(a, b AigLit) (AigLit, error) {
	if m == nil || !m.validSignal(a) || !m.validSignal(b) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	switch {
	case a == b:
		return AigConst0, nil
	case a == b.Not():
		return AigConst1, nil
	case a.ID() == 0:
		return b.NotIf(a == AigConst1), nil
	case b.ID() == 0:
		return a.NotIf(b == AigConst1), nil
	}
	// The pinned native GCC build evaluates Aig_Or's right argument first.
	// Preserve that observable object-allocation order.
	right, err := m.And(a.Not(), b)
	if err != nil {
		return InvalidAigLit, err
	}
	left, err := m.And(a, b.Not())
	if err != nil {
		return InvalidAigLit, err
	}
	return m.Or(left, right)
}

// Mux implements Aig_Mux: control ? whenTrue : whenFalse.
func (m *AigManager) Mux(control, whenTrue, whenFalse AigLit) (AigLit, error) {
	if m == nil || !m.validSignal(control) || !m.validSignal(whenTrue) || !m.validSignal(whenFalse) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	// The pinned native GCC build evaluates the second Aig_And argument to
	// Aig_Or first, which determines stable intermediate object IDs.
	right, err := m.And(control.Not(), whenFalse)
	if err != nil {
		return InvalidAigLit, err
	}
	left, err := m.And(control, whenTrue)
	if err != nil {
		return InvalidAigLit, err
	}
	return m.Or(left, right)
}

// Mux2 is the direct Aig_Mux2 spelling. Aig_Mux uses the same construction in
// the native source because its optional canonicalizing branch is disabled.
func (m *AigManager) Mux2(control, whenTrue, whenFalse AigLit) (AigLit, error) {
	return m.Mux(control, whenTrue, whenFalse)
}

func (m *AigManager) Majority(a, b, c AigLit) (AigLit, error) {
	if m == nil || !m.validSignal(a) || !m.validSignal(b) || !m.validSignal(c) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	// Match the pinned native build's right-to-left argument evaluation:
	// (b&c), (a&c), then (a&b).
	bc, err := m.And(b, c)
	if err != nil {
		return InvalidAigLit, err
	}
	ac, err := m.And(a, c)
	if err != nil {
		return InvalidAigLit, err
	}
	ab, err := m.And(a, b)
	if err != nil {
		return InvalidAigLit, err
	}
	partial, err := m.Or(ab, ac)
	if err != nil {
		return InvalidAigLit, err
	}
	return m.Or(partial, bc)
}

// Multi constructs the balanced operation tree of Aig_Multi. Empty inputs
// are rejected, matching C's nArgs > 0 precondition.
func (m *AigManager) Multi(arguments []AigLit, operation AigType) (AigLit, error) {
	if m == nil || len(arguments) == 0 || (operation != AigAND && operation != AigXOR) {
		return InvalidAigLit, ErrDimension
	}
	for _, a := range arguments {
		if !m.validSignal(a) {
			return InvalidAigLit, ErrInvalidLiteral
		}
	}
	return m.multi(arguments, operation)
}

func (m *AigManager) multiBounded(arguments []AigLit, operation AigType, budget *aigOperationBudget) (AigLit, error) {
	if len(arguments) == 1 {
		return arguments[0], nil
	}
	split := len(arguments) / 2
	a, err := m.multiBounded(arguments[:split], operation, budget)
	if err != nil {
		return InvalidAigLit, err
	}
	b, err := m.multiBounded(arguments[split:], operation, budget)
	if err != nil {
		return InvalidAigLit, err
	}
	if err = budget.step(); err != nil {
		return InvalidAigLit, err
	}
	return m.Operate(a, b, operation)
}

// MiterPairs implements Aig_Miter over adjacent pairs without mutating the
// caller's slice. The result is the OR of all pairwise XORs.
func (m *AigManager) MiterPairs(pairs [][2]AigLit, options AigOperationOptions) (AigLit, error) {
	if m == nil || len(pairs) == 0 {
		return InvalidAigLit, ErrDimension
	}
	left, right := make([]AigLit, len(pairs)), make([]AigLit, len(pairs))
	for i, pair := range pairs {
		left[i], right[i] = pair[0], pair[1]
	}
	return m.Miter(left, right, options)
}

// Miter implements Aig_MiterTwo. Multi-step construction is committed only
// after all inputs, limits and cancellation checks succeed.
func (m *AigManager) Miter(left, right []AigLit, options AigOperationOptions) (AigLit, error) {
	if m == nil || len(left) == 0 || len(left) != len(right) {
		return InvalidAigLit, ErrDimension
	}
	if err := m.Check(); err != nil {
		return InvalidAigLit, err
	}
	for i := range left {
		if !m.validSignal(left[i]) || !m.validSignal(right[i]) {
			return InvalidAigLit, ErrInvalidLiteral
		}
	}
	budget, err := newAigOperationBudget(options)
	if err != nil {
		return InvalidAigLit, err
	}
	work := m.Clone()
	equivalent := make([]AigLit, len(left))
	for i := range left {
		if err = budget.step(); err != nil {
			return InvalidAigLit, err
		}
		xor, err := work.Xor(left[i], right[i])
		if err != nil {
			return InvalidAigLit, err
		}
		equivalent[i] = xor.Not()
	}
	allEquivalent, err := work.multiBounded(equivalent, AigAND, budget)
	if err != nil {
		return InvalidAigLit, err
	}
	result := allEquivalent.Not()
	*m = *work
	return result, nil
}

func (m *AigManager) createNaryFunction(variables int, operation AigType, options AigOperationOptions) (AigLit, error) {
	if m == nil || variables < 0 || (operation != AigAND && operation != AigXOR && operation != AigVoid) {
		return InvalidAigLit, ErrDimension
	}
	if err := m.Check(); err != nil {
		return InvalidAigLit, err
	}
	budget, err := newAigOperationBudget(options)
	if err != nil {
		return InvalidAigLit, err
	}
	work := m.Clone()
	result := AigConst1
	if operation == AigXOR || operation == AigVoid {
		result = AigConst0
	}
	for i := 0; i < variables; i++ {
		if err = budget.step(); err != nil {
			return InvalidAigLit, err
		}
		variable, err := work.IthVar(i)
		if err != nil {
			return InvalidAigLit, err
		}
		if operation == AigVoid {
			result, err = work.Or(result, variable)
		} else {
			result, err = work.Operate(result, variable, operation)
		}
		if err != nil {
			return InvalidAigLit, err
		}
	}
	*m = *work
	return result, nil
}

func (m *AigManager) CreateAndFunction(variables int, options AigOperationOptions) (AigLit, error) {
	return m.createNaryFunction(variables, AigAND, options)
}

func (m *AigManager) CreateOrFunction(variables int, options AigOperationOptions) (AigLit, error) {
	return m.createNaryFunction(variables, AigVoid, options)
}

func (m *AigManager) CreateXorFunction(variables int, options AigOperationOptions) (AigLit, error) {
	return m.createNaryFunction(variables, AigXOR, options)
}
func (m *AigManager) multi(arguments []AigLit, operation AigType) (AigLit, error) {
	if len(arguments) == 1 {
		return arguments[0], nil
	}
	split := len(arguments) / 2
	a, err := m.multi(arguments[:split], operation)
	if err != nil {
		return InvalidAigLit, err
	}
	b, err := m.multi(arguments[split:], operation)
	if err != nil {
		return InvalidAigLit, err
	}
	if operation == AigXOR {
		return m.Xor(a, b)
	}
	return m.And(a, b)
}
