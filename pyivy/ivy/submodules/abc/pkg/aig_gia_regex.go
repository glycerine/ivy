package abc

// Regular-expression automata from src/aig/gia/giaRex.c.

import (
	"math/rand"
	"slices"
	"sort"
)

type GiaRegexOptions struct {
	MaxExpression int
	MaxObjects    int
	MaxEntries    int
	Stop          func() bool
}

type GiaRegexAutomaton struct {
	Graph              *GiaManager
	Alphabet           []byte
	Preprocessed       string
	StatePositions     []int
	EpsilonTransitions []int
	InitialState       []bool
}

type GiaAutomatonWalkOptions struct {
	Steps     int
	Walks     int
	Seed      int64
	MaxStates int
	MaxVisits int
	Stop      func() bool
}

type GiaAutomatonWalk struct {
	Inputs   []int
	Accepted []bool
	States   []uint64
}

type GiaAutomatonWalkResult struct {
	Walks       []GiaAutomatonWalk
	States      []uint64
	Counts      []int
	InitialMask uint64
}

func giaRegexSpecial(symbol byte) bool {
	return symbol == '(' || symbol == ')' || symbol == '*' || symbol == '|'
}

func checkGiaRegexOptions(options GiaRegexOptions) (GiaRegexOptions, error) {
	if options.MaxExpression < 0 || options.MaxObjects < 0 || options.MaxEntries < 0 {
		return options, ErrDimension
	}
	if options.MaxExpression == 0 {
		options.MaxExpression = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	return options, nil
}

func giaRegexPreprocess(expression string, maximum int) ([]byte, error) {
	if len(expression) > maximum {
		return nil, ErrLimit
	}
	result := make([]byte, 0, 2*len(expression)+4)
	result = append(result, '(', '(')
	for i := 0; i < len(expression); i++ {
		symbol := expression[i]
		if symbol == 0 {
			return nil, ErrFormat
		}
		if symbol == '(' {
			result = append(result, '(')
		} else if symbol == ')' {
			result = append(result, ')')
		}
		switch symbol {
		case ' ', '\t', '\n', '\r':
		default:
			result = append(result, symbol)
		}
		if len(result) > maximum {
			return nil, ErrLimit
		}
	}
	result = append(result, ')', ')')
	if len(result) > maximum {
		return nil, ErrLimit
	}
	return result, nil
}

func giaRegexAnalyze(expression []byte, order bool, maximum int) (alphabet []byte, stateAt, statePositions, transitions []int, err error) {
	stateAt = make([]int, len(expression)+1)
	for i := range stateAt {
		stateAt[i] = -1
	}
	alphabetSeen := make(map[byte]bool)
	for position, symbol := range expression {
		if giaRegexSpecial(symbol) {
			continue
		}
		if !alphabetSeen[symbol] {
			alphabetSeen[symbol] = true
			alphabet = append(alphabet, symbol)
		}
		stateAt[position] = len(statePositions)
		statePositions = append(statePositions, position)
	}
	if len(alphabet) == 0 || len(statePositions) > maximum || len(expression)+1 > maximum {
		if len(alphabet) == 0 {
			return nil, nil, nil, nil, ErrFormat
		}
		return nil, nil, nil, nil, ErrLimit
	}
	stateAt[len(expression)] = len(statePositions)
	if order {
		sort.Slice(alphabet, func(i, j int) bool { return alphabet[i] < alphabet[j] })
	}
	transitions = make([]int, len(expression))
	for i := range transitions {
		transitions[i] = -1
	}
	stack := make([]int, 0, len(expression))
	for position, symbol := range expression {
		left := position
		switch symbol {
		case '(', '|':
			stack = append(stack, position)
		case ')':
			if len(stack) == 0 {
				return nil, nil, nil, nil, ErrFormat
			}
			or := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if expression[or] == '|' {
				if len(stack) == 0 {
					return nil, nil, nil, nil, ErrFormat
				}
				left = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if expression[left] != '(' {
					return nil, nil, nil, nil, ErrFormat
				}
				transitions[left] = or + 1
				transitions[or] = position
			} else {
				left = or
				if expression[left] != '(' {
					return nil, nil, nil, nil, ErrFormat
				}
			}
		}
		if position+1 < len(expression) && expression[position+1] == '*' {
			transitions[left] = position + 1
			transitions[position+1] = left
		}
	}
	if len(stack) != 0 {
		return nil, nil, nil, nil, ErrFormat
	}
	return alphabet, stateAt, statePositions, transitions, nil
}

func giaRegexReached(expression []byte, transitions []int, start int, maximum int, stop func() bool) ([]int, error) {
	if start < 0 || start > len(expression) {
		return nil, ErrFormat
	}
	reached := make([]int, 0)
	seen := make([]bool, len(expression)+1)
	stack := []int{start}
	visits := 0
	for len(stack) != 0 {
		if stop != nil && stop() {
			return nil, ErrCancelled
		}
		position := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[position] {
			continue
		}
		seen[position] = true
		visits++
		if visits > maximum {
			return nil, ErrLimit
		}
		if position == len(expression) {
			reached = append(reached, position)
			continue
		}
		symbol := expression[position]
		if !giaRegexSpecial(symbol) {
			reached = append(reached, position)
		}
		if transitions[position] >= 0 {
			stack = append(stack, transitions[position])
		}
		if giaRegexSpecial(symbol) && symbol != '|' {
			stack = append(stack, position+1)
		}
	}
	return reached, nil
}

type giaRegexOneHotPair struct{ zero, one GiaLit }

func giaRegexExactlyOne(graph *GiaManager, literals []GiaLit, stop func() bool) (GiaLit, error) {
	if len(literals) == 0 {
		return InvalidGiaLit, ErrDimension
	}
	pairs := make([]giaRegexOneHotPair, len(literals))
	for i, literal := range literals {
		pairs[i] = giaRegexOneHotPair{zero: literal.Not(), one: literal}
	}
	for len(pairs) > 1 {
		next := make([]giaRegexOneHotPair, 0, (len(pairs)+1)/2)
		for i := 0; i < len(pairs); i += 2 {
			if stop != nil && stop() {
				return InvalidGiaLit, ErrCancelled
			}
			if i+1 == len(pairs) {
				next = append(next, pairs[i])
				continue
			}
			zero, err := graph.HashAnd(pairs[i].zero, pairs[i+1].zero)
			if err != nil {
				return InvalidGiaLit, err
			}
			left, err := graph.HashAnd(pairs[i].zero, pairs[i+1].one)
			if err != nil {
				return InvalidGiaLit, err
			}
			right, err := graph.HashAnd(pairs[i+1].zero, pairs[i].one)
			if err != nil {
				return InvalidGiaLit, err
			}
			one, err := graph.HashOr(left, right)
			if err != nil {
				return InvalidGiaLit, err
			}
			next = append(next, giaRegexOneHotPair{zero: zero, one: one})
		}
		pairs = next
	}
	return pairs[0].one, nil
}

// CompileGiaRegex converts the source's byte-oriented regular-expression
// language to a zero-initialized sequential GIA. Only (), |, and * are
// operators; ASCII whitespace is discarded and every other byte is a symbol.
func CompileGiaRegex(expression string, orderAlphabet bool, options GiaRegexOptions) (GiaRegexAutomaton, error) {
	fail := func(err error) (GiaRegexAutomaton, error) { return GiaRegexAutomaton{}, err }
	options, err := checkGiaRegexOptions(options)
	if err != nil {
		return fail(err)
	}
	if options.Stop != nil && options.Stop() {
		return fail(ErrCancelled)
	}
	prepared, err := giaRegexPreprocess(expression, options.MaxExpression)
	if err != nil {
		return fail(err)
	}
	alphabet, stateAt, statePositions, transitions, err := giaRegexAnalyze(prepared, orderAlphabet, options.MaxEntries)
	if err != nil {
		return fail(err)
	}
	graph, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	graph.Name = expression
	inputs := make([]GiaLit, len(alphabet))
	for i := range inputs {
		if options.Stop != nil && options.Stop() {
			return fail(ErrCancelled)
		}
		inputs[i], err = graph.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	states := make([]GiaLit, len(statePositions))
	for i := range states {
		states[i], err = graph.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	alphabetIndex := make(map[byte]int, len(alphabet))
	for i, symbol := range alphabet {
		alphabetIndex[symbol] = i
	}
	next := make([]GiaLit, len(states)+1)
	if err = graph.HashStart(); err != nil {
		return fail(err)
	}
	for position := 1; position < len(prepared); position++ {
		if giaRegexSpecial(prepared[position]) {
			continue
		}
		reached, reachErr := giaRegexReached(prepared, transitions, position+1, options.MaxEntries, options.Stop)
		if reachErr != nil {
			return fail(reachErr)
		}
		term, termErr := graph.HashAnd(states[stateAt[position]], inputs[alphabetIndex[prepared[position]]])
		if termErr != nil {
			return fail(termErr)
		}
		for _, destination := range reached {
			index := stateAt[destination]
			if index < 0 {
				return fail(ErrInvariant)
			}
			next[index], err = graph.HashOr(next[index], term)
			if err != nil {
				return fail(err)
			}
		}
	}
	one, err := giaRegexExactlyOne(graph, inputs, options.Stop)
	if err != nil {
		return fail(err)
	}
	for i := range next {
		next[i], err = graph.HashAnd(one, next[i])
		if err != nil {
			return fail(err)
		}
	}
	graph.HashStop()
	if _, err = graph.AppendCO(next[len(states)]); err != nil {
		return fail(err)
	}
	for _, driver := range next[:len(states)] {
		if _, err = graph.AppendCO(driver); err != nil {
			return fail(err)
		}
	}
	if err = graph.SetRegisterCount(len(states)); err != nil {
		return fail(err)
	}
	initialPositions, err := giaRegexReached(prepared, transitions, 1, options.MaxEntries, options.Stop)
	if err != nil {
		return fail(err)
	}
	initial := make([]bool, len(states))
	for _, position := range initialPositions {
		if position != len(prepared) {
			initial[stateAt[position]] = true
		}
	}
	clean, _, err := graph.Cleanup()
	if err != nil {
		return fail(err)
	}
	resets := make([]LatchInit, len(initial))
	for i, value := range initial {
		if value {
			resets[i] = LatchInitOne
		}
	}
	normalized, err := clean.NormalizeInitialState(resets, 0, GiaCofactorOptions{MaxObjects: options.MaxObjects, MaxMappingEntries: options.MaxEntries, Stop: options.Stop})
	if err != nil {
		return fail(err)
	}
	return GiaRegexAutomaton{
		Graph:              normalized.Graph,
		Alphabet:           slices.Clone(alphabet),
		Preprocessed:       string(prepared),
		StatePositions:     slices.Clone(statePositions),
		EpsilonTransitions: slices.Clone(transitions),
		InitialState:       initial,
	}, nil
}

// Simulate returns acceptance after each consumed symbol. Unknown symbols
// drive every alphabet input low and therefore cannot advance the automaton.
func (a GiaRegexAutomaton) Simulate(sequence string) ([]bool, error) {
	if a.Graph == nil || len(a.Alphabet) != a.Graph.PrimaryInputCount() {
		return nil, ErrDimension
	}
	index := make(map[byte]int, len(a.Alphabet))
	for i, symbol := range a.Alphabet {
		index[symbol] = i
	}
	state := make([]bool, a.Graph.RegisterCount())
	accepted := make([]bool, len(sequence))
	for frame := 0; frame < len(sequence); frame++ {
		input := make([]bool, len(a.Alphabet))
		if i, found := index[sequence[frame]]; found {
			input[i] = true
		}
		outputs, next, err := a.Graph.Step(input, state)
		if err != nil || len(outputs) != 1 {
			if err == nil {
				err = ErrInvariant
			}
			return nil, err
		}
		accepted[frame] = outputs[0]
		state = next
	}
	return accepted, nil
}

// GiaTranspose64 transposes a 64-by-64 bit matrix in place.
func GiaTranspose64(matrix *[64]uint64) {
	mask := uint64(0x00000000ffffffff)
	for width := uint(32); width != 0; {
		for k := uint(0); k < 64; k = (k + width + 1) &^ width {
			temporary := (matrix[k] ^ (matrix[k+width] >> width)) & mask
			matrix[k] ^= temporary
			matrix[k+width] ^= temporary << width
		}
		width >>= 1
		mask ^= mask << width
	}
}

// AutomatonStep64 evaluates every one-hot primary input in parallel. State and
// returned successor words use bit 63-index for register indices.
func (m *GiaManager) AutomatonStep64(state uint64) (uint64, []uint64, error) {
	if m == nil || m.PrimaryOutputCount() != 1 || m.PrimaryInputCount() > 64 || m.RegisterCount() > 64 {
		return 0, nil, ErrDimension
	}
	if err := m.Check(); err != nil {
		return 0, nil, err
	}
	if m.XorCount() != 0 || m.MuxCount() != 0 {
		return 0, nil, ErrUnsupported
	}
	values := make([]uint64, len(m.objects))
	for i, id := range m.cis[:m.PrimaryInputCount()] {
		values[id] = uint64(1) << (63 - i)
	}
	for i, id := range m.cis[m.PrimaryInputCount():] {
		if state&(uint64(1)<<(63-i)) != 0 {
			values[id] = ^uint64(0)
		}
	}
	read := func(literal GiaLit) uint64 {
		value := values[literal.ID()]
		if literal.Complemented() {
			return ^value
		}
		return value
	}
	for i, object := range m.objects {
		if !object.IsAnd() {
			continue
		}
		first, _ := object.Fanin0(GiaID(i))
		if object.IsBuffer() {
			values[i] = read(first)
			continue
		}
		second, _ := object.Fanin1(GiaID(i))
		values[i] = read(first) & read(second)
	}
	var rows [64]uint64
	for i, co := range m.cos[m.PrimaryOutputCount():] {
		driver, _ := m.objects[co].Fanin0(co)
		rows[i] = read(driver)
	}
	GiaTranspose64(&rows)
	next := make([]uint64, m.PrimaryInputCount())
	copy(next, rows[:len(next)])
	po := m.cos[0]
	driver, _ := m.objects[po].Fanin0(po)
	return read(driver), next, nil
}

// AutomatonInitialMask preserves the source walker's structural inference of
// register phases from complemented RO fanins.
func (m *GiaManager) AutomatonInitialMask() (uint64, error) {
	if m == nil || m.RegisterCount() > 64 {
		return 0, ErrDimension
	}
	if err := m.Check(); err != nil {
		return 0, err
	}
	mask := uint64(0)
	for i, object := range m.objects {
		if !object.IsAnd() {
			continue
		}
		id := GiaID(i)
		fanins := make([]GiaLit, 0, 2)
		first, _ := object.Fanin0(id)
		fanins = append(fanins, first)
		if !object.IsBuffer() {
			second, _ := object.Fanin1(id)
			fanins = append(fanins, second)
		}
		for _, fanin := range fanins {
			object := m.objects[fanin.ID()]
			if !fanin.Complemented() || !object.IsCI() {
				continue
			}
			terminal, _ := object.TerminalIndex()
			register := terminal - m.PrimaryInputCount()
			if register >= 0 {
				mask |= uint64(1) << (63 - register)
			}
		}
	}
	return mask, nil
}

// WalkAutomaton performs deterministic caller-seeded coverage-guided walks.
func (m *GiaManager) WalkAutomaton(options GiaAutomatonWalkOptions) (GiaAutomatonWalkResult, error) {
	fail := func(err error) (GiaAutomatonWalkResult, error) { return GiaAutomatonWalkResult{}, err }
	if m == nil || options.Steps < 0 || options.Walks < 0 || options.MaxStates < 0 || options.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if m.PrimaryInputCount() == 0 || m.PrimaryInputCount() > 64 || m.RegisterCount() > 64 || m.PrimaryOutputCount() != 1 {
		return fail(ErrDimension)
	}
	if options.MaxStates == 0 {
		options.MaxStates = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 26
	}
	if options.Stop != nil && options.Stop() {
		return fail(ErrCancelled)
	}
	initial, err := m.AutomatonInitialMask()
	if err != nil {
		return fail(err)
	}
	result := GiaAutomatonWalkResult{States: []uint64{0}, Counts: []int{1}, InitialMask: initial}
	positions := map[uint64]int{0: 0}
	random := rand.New(rand.NewSource(options.Seed))
	visits := 0
	for walkIndex := 0; walkIndex < options.Walks; walkIndex++ {
		walk := GiaAutomatonWalk{States: []uint64{0}}
		state := uint64(0)
		for step := 0; step < options.Steps; step++ {
			if options.Stop != nil && options.Stop() {
				return fail(ErrCancelled)
			}
			visits++
			if visits > options.MaxVisits {
				return fail(ErrLimit)
			}
			output, next, stepErr := m.AutomatonStep64(state)
			if stepErr != nil {
				return fail(stepErr)
			}
			chosen, minimum, ties := -1, int(^uint(0)>>1), 0
			for input, successor := range next {
				if successor == initial {
					continue
				}
				count := 0
				if position, found := positions[successor]; found {
					count = result.Counts[position]
				}
				if count < minimum {
					chosen, minimum, ties = input, count, 1
				} else if count == minimum {
					ties++
					if random.Intn(ties) == 0 {
						chosen = input
					}
				}
			}
			if chosen < 0 {
				break
			}
			successor := next[chosen]
			position, found := positions[successor]
			if !found {
				if len(result.States) == options.MaxStates {
					return fail(ErrLimit)
				}
				position = len(result.States)
				positions[successor] = position
				result.States = append(result.States, successor)
				result.Counts = append(result.Counts, 0)
			}
			result.Counts[position]++
			walk.Inputs = append(walk.Inputs, chosen)
			walk.Accepted = append(walk.Accepted, output&(uint64(1)<<(63-chosen)) != 0)
			walk.States = append(walk.States, successor)
			state = successor
		}
		result.Walks = append(result.Walks, walk)
	}
	return result, nil
}
