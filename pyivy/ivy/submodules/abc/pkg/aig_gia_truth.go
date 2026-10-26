package abc

// Small truth/GIA bridge from aig/gia/giaTruth.c: Gia_Truth5ToGia,
// Gia_Truth6ToGia and Gia_LutComputeTruth6Simple/Map (including rec helpers).

type GiaTruthBuildOptions struct {
	Hash       bool
	MaxObjects int // zero: one million complete graph objects
	Stop       func() bool
}

// WithTruthTable6 appends C's highest-dependent-input-first Shannon mux tree
// to an owned copy of the graph. inputs bind truth variables to existing signal
// literals and may include phases, aliases or constants. Hash selects native
// strashing; raw mode retains C's three-AND mux construction. Degenerate raw
// bindings fold equal-ID ANDs that C cannot represent. Five-input truth uses
// the same logic as C's uint32 path with canonical 64-bit stretching.
// Errors return no partial graph/root; the source and truth remain unchanged.
func (m *GiaManager) WithTruthTable6(truth *TruthTable, inputs []GiaLit, options GiaTruthBuildOptions) (*GiaManager, GiaLit, error) {
	fail := func(e error) (*GiaManager, GiaLit, error) { return nil, InvalidGiaLit, e }
	if options.MaxObjects < 0 {
		return fail(ErrDimension)
	}
	if m == nil {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if e := truth.Check(); e != nil {
		return fail(e)
	}
	if truth.variables > 6 || len(inputs) != truth.variables {
		return fail(ErrDimension)
	}
	for _, lit := range inputs {
		if !m.validSignal(lit) {
			return fail(ErrInvalidLiteral)
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	limit := options.MaxObjects
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return fail(ErrLimit)
	}
	target := m.Clone()
	target.objectLimit = min(limit, int(MaxGiaID)+1)
	if options.Hash && !target.Hashing() {
		if e := target.HashStart(); e != nil {
			return fail(e)
		}
	}
	and := func(a, b GiaLit) (GiaLit, error) {
		if a.ID() == b.ID() {
			return target.AppendAndSimplified(a, b)
		}
		return target.AppendAnd(a, b)
	}
	mux := func(control, yes, no GiaLit) (GiaLit, error) {
		if options.Hash {
			return target.HashMux(control, yes, no)
		}
		low, e := and(control.Not(), no)
		if e != nil {
			return InvalidGiaLit, e
		}
		high, e := and(control, yes)
		if e != nil {
			return InvalidGiaLit, e
		}
		joined, e := and(low.Not(), high.Not())
		if e != nil {
			return InvalidGiaLit, e
		}
		return joined.Not(), nil
	}
	var build func(uint64, int) (GiaLit, error)
	build = func(word uint64, variables int) (GiaLit, error) {
		if stopped() {
			return InvalidGiaLit, ErrCancelled
		}
		if word == 0 {
			return GiaConst0, nil
		}
		if word == ^uint64(0) {
			return GiaConst1, nil
		}
		variable := variables - 1
		for ; variable >= 0; variable-- {
			shift := uint(1) << uint(variable)
			if ((word>>shift)^word)&^truthVariablePattern(variable) != 0 {
				break
			}
		}
		if variable < 0 {
			return InvalidGiaLit, ErrInvariant
		}
		pattern := truthVariablePattern(variable)
		shift := uint(1) << uint(variable)
		low, high := word&^pattern, word&pattern
		a, e := build(low|(low<<shift), variable)
		if e != nil {
			return InvalidGiaLit, e
		}
		b, e := build(high|(high>>shift), variable)
		if e != nil {
			return InvalidGiaLit, e
		}
		return mux(inputs[variable], b, a)
	}
	root, e := build(truth.words[0], truth.variables)
	if e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return target, root, nil
}

// OutputTruthWord6 evaluates one CO cone into the six-slot packed truth format.
// inputMap, when nonnil, maps CI ordinals to slots 0..5; -1 is allowed for CIs
// outside this cone. Nil uses ordinal identity. Aliased slots model tied CIs.
// The iterative postorder preserves source scratch and handles constants,
// buffers, native XOR/mux nodes and register terminals explicitly. Options cap
// visited nonconstant cone nodes and arena-indexed scratch entries; zero caps
// select one million. Only reached CIs require valid slots.
func (m *GiaManager) OutputTruthWord6(output int, inputMap []int, options GiaCofactorOptions) (uint64, error) {
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return 0, ErrDimension
	}
	if m == nil {
		return 0, ErrDimension
	}
	if e := m.Check(); e != nil {
		return 0, e
	}
	if output < 0 || output >= len(m.cos) || inputMap != nil && len(inputMap) != len(m.cis) {
		return 0, ErrDimension
	}
	maxObjects, maxEntries := options.MaxObjects, options.MaxMappingEntries
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if len(m.objects) > maxEntries {
		return 0, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return 0, ErrCancelled
	}
	values := make([]uint64, len(m.objects))
	seen := make([]uint8, len(m.objects))
	seen[0] = 2
	type frame struct {
		id    GiaID
		ready bool
	}
	stack := []frame{{id: m.cos[output]}}
	visited := 0
	litValue := func(lit GiaLit) uint64 {
		word := values[lit.ID()]
		if lit.Complemented() {
			word = ^word
		}
		return word
	}
	for len(stack) > 0 {
		if stopped() {
			return 0, ErrCancelled
		}
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[item.id] == 2 {
			continue
		}
		object := m.objects[item.id]
		if !item.ready {
			if seen[item.id] != 0 {
				continue
			}
			visited++
			if visited > maxObjects {
				return 0, ErrLimit
			}
			seen[item.id] = 1
			if object.IsCI() {
				slot := int(object.Diff1())
				if inputMap != nil {
					slot = inputMap[slot]
				}
				if slot < 0 || slot >= 6 {
					return 0, ErrDimension
				}
				values[item.id] = truthVariablePattern(slot)
				seen[item.id] = 2
				continue
			}
			children, n, e := m.crossCutFanins(item.id)
			if e != nil {
				return 0, e
			}
			stack = append(stack, frame{item.id, true})
			for i := 0; i < n; i++ {
				if seen[children[i]] != 2 {
					stack = append(stack, frame{id: children[i]})
				}
			}
			continue
		}
		a, e := object.Fanin0(item.id)
		if e != nil {
			return 0, e
		}
		first := litValue(a)
		if object.IsCO() || object.IsBuffer() {
			values[item.id] = first
		} else {
			b, e := object.Fanin1(item.id)
			if e != nil {
				return 0, e
			}
			second := litValue(b)
			if control, mux := m.MuxControl(item.id); mux {
				c := litValue(control)
				values[item.id] = (c & second) | (^c & first)
			} else if object.IsXor() {
				values[item.id] = first ^ second
			} else {
				values[item.id] = first & second
			}
		}
		seen[item.id] = 2
	}
	return values[m.cos[output]], nil
}
