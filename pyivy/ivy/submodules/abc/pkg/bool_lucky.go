package abc

// Lucky semi-canonicalization from src/bool/lucky/lucky.c,
// luckyFast6.c, luckyFast16.c, luckyRead.c, luckySimple.c, luckySwap.c,
// luckySwapIJ.c, lucky.h, and luckyInt.h.

import (
	"fmt"
	"math/bits"
	"os"
	"slices"
	"strconv"
	"strings"
)

const LuckyMaximumVariables = 16

// LuckyOptions bounds a fast semi-canonicalization. Thorough enables the
// second all-input-phase pass used by luckyCanonicizer_final_fast1.
type LuckyOptions struct {
	MaxOperations int
	Stop          func() bool
	Thorough      bool
}

// LuckyCanonicalization owns a Lucky representative and its transformation.
// Permutation is target-position to source-variable, matching pCanonPerm.
// InputPhases are indexed by target position. The representative at x equals
// the source at y, where y[Permutation[i]] = x[i] XOR InputPhases[i], followed
// by OutputPhase.
type LuckyCanonicalization struct {
	Table       *TruthTable
	Permutation []int
	InputPhases []bool
	OutputPhase bool
	Operations  int
}

func (r LuckyCanonicalization) Clone() LuckyCanonicalization {
	r.Table = r.Table.Clone()
	r.Permutation = slices.Clone(r.Permutation)
	r.InputPhases = slices.Clone(r.InputPhases)
	return r
}

type luckyBudget struct {
	left, operations int
	limited          bool
	stop             func() bool
}

func newLuckyBudget(options LuckyOptions) (*luckyBudget, error) {
	if options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	b := &luckyBudget{left: options.MaxOperations, limited: options.MaxOperations != 0, stop: options.Stop}
	if b.stop != nil && b.stop() {
		return nil, ErrCancelled
	}
	return b, nil
}

func (b *luckyBudget) step() error {
	if b.stop != nil && b.stop() {
		return ErrCancelled
	}
	if b.limited {
		if b.left == 0 {
			return ErrLimit
		}
		b.left--
	}
	b.operations++
	return nil
}

type luckyState struct {
	table       *TruthTable
	permutation []int
	phases      []bool
	output      bool
}

func newLuckyState(table *TruthTable) luckyState {
	permutation := make([]int, table.variables)
	for i := range permutation {
		permutation[i] = i
	}
	return luckyState{table: table.Clone(), permutation: permutation, phases: make([]bool, table.variables)}
}

func (s luckyState) clone() luckyState {
	s.table = s.table.Clone()
	s.permutation = slices.Clone(s.permutation)
	s.phases = slices.Clone(s.phases)
	return s
}

func (s *luckyState) flip(variable int) {
	s.table.flipVariable(variable)
	s.phases[variable] = !s.phases[variable]
}

func (s *luckyState) swap(first, second int) {
	s.table.swapVariables(first, second)
	s.permutation[first], s.permutation[second] = s.permutation[second], s.permutation[first]
	s.phases[first], s.phases[second] = s.phases[second], s.phases[first]
}

func (s *luckyState) complement() {
	for i := range s.table.words {
		s.table.words[i] = ^s.table.words[i]
	}
	s.output = !s.output
}

func luckyInitial(state *luckyState) (counts []int, inputTied, outputTied bool, err error) {
	populations, err := state.table.CofactorPopulations()
	if err != nil {
		return nil, false, false, err
	}
	ones := populations.Ones
	counts = populations.Zero
	half := len(state.table.words) * 32
	outputTied = ones == half
	if ones > half {
		state.complement()
		ones = 2*half - ones
		for i := range counts {
			counts[i] = half - counts[i]
		}
	}
	for variable, count := range counts {
		if 2*count == ones {
			inputTied = true
			continue
		}
		if count < ones-count {
			state.flip(variable)
			counts[variable] = ones - count
		}
	}
	// Lucky uses a stable adjacent bubble sort, unlike the later utilTruth
	// selection-sort implementation.
	for {
		changed := false
		for i := 0; i+1 < len(counts); i++ {
			if counts[i] <= counts[i+1] {
				continue
			}
			counts[i], counts[i+1] = counts[i+1], counts[i]
			state.swap(i, i+1)
			changed = true
		}
		if !changed {
			break
		}
	}
	return counts, inputTied, outputTied, nil
}

// luckyRefinePair directly evaluates the eight pair configurations encoded by
// Extra_Truth6MinimumRoundOne and the corresponding quotient routines in
// luckyFast16.c. This retains their deterministic first-minimum tie rule while
// avoiding architecture-specific masks and fixed scratch arrays.
func luckyRefinePair(state luckyState, variable int, allowPhases bool, budget *luckyBudget) (luckyState, error) {
	type transform struct {
		swap, first, second bool
	}
	transforms := []transform{{}}
	if allowPhases {
		transforms = append(transforms,
			transform{first: true}, transform{second: true}, transform{first: true, second: true})
	}
	transforms = append(transforms, transform{swap: true})
	if allowPhases {
		// C visits swapped !b,a; b,!a; !b,!a in this order.
		transforms = append(transforms,
			transform{swap: true, first: true}, transform{swap: true, second: true}, transform{swap: true, first: true, second: true})
	}
	best := state
	for _, operation := range transforms[1:] {
		if err := budget.step(); err != nil {
			return luckyState{}, err
		}
		candidate := state.clone()
		if operation.swap {
			candidate.swap(variable, variable+1)
		}
		if operation.first {
			candidate.flip(variable)
		}
		if operation.second {
			candidate.flip(variable + 1)
		}
		if truthWordsLessReverse(candidate.table.words, best.table.words) {
			best = candidate
		}
	}
	return best, nil
}

func luckyRefine(state luckyState, counts []int, allowPhases bool, budget *luckyBudget) (luckyState, error) {
	for {
		before := slices.Clone(state.table.words)
		for i := 0; i+1 < len(counts); i++ {
			if counts[i] != counts[i+1] {
				continue
			}
			var err error
			state, err = luckyRefinePair(state, i, allowPhases, budget)
			if err != nil {
				return luckyState{}, err
			}
		}
		if slices.Equal(before, state.table.words) {
			return state, nil
		}
	}
}

func luckyInvertAllInputs(state *luckyState) {
	for variable := range state.phases {
		state.flip(variable)
	}
}

func luckyLess(first, second luckyState) bool {
	return truthWordsLessReverse(first.table.words, second.table.words)
}

// LuckyCanonicalize computes the deterministic fast Lucky representative for
// packed truth tables from zero through sixteen variables. It is atomic: any
// validation, limit, or cancellation error returns a zero result.
func (t *TruthTable) LuckyCanonicalize(options LuckyOptions) (LuckyCanonicalization, error) {
	fail := func(err error) (LuckyCanonicalization, error) { return LuckyCanonicalization{}, err }
	if t == nil || t.variables > LuckyMaximumVariables {
		return fail(ErrDimension)
	}
	if err := t.Check(); err != nil {
		return fail(err)
	}
	budget, err := newLuckyBudget(options)
	if err != nil {
		return fail(err)
	}
	state := newLuckyState(t)
	counts, inputTied, outputTied, err := luckyInitial(&state)
	if err != nil {
		return fail(err)
	}
	initial := state.clone()
	state, err = luckyRefine(state, counts, inputTied, budget)
	if err != nil {
		return fail(err)
	}
	// The six-variable routine considers the balanced output complement in
	// both entry points. The wide routine does so in its "fast1" entry point.
	if outputTied && (t.variables <= 6 || options.Thorough) {
		other := initial.clone()
		other.complement()
		other, err = luckyRefine(other, counts, inputTied, budget)
		if err != nil {
			return fail(err)
		}
		if luckyLess(other, state) {
			state = other
		}
	}
	if options.Thorough {
		other := state.clone()
		luckyInvertAllInputs(&other)
		inverted := other.clone()
		other, err = luckyRefine(other, counts, inputTied, budget)
		if err != nil {
			return fail(err)
		}
		if outputTied {
			complement := inverted.clone()
			complement.complement()
			complement, err = luckyRefine(complement, counts, inputTied, budget)
			if err != nil {
				return fail(err)
			}
			if luckyLess(complement, other) {
				other = complement
			}
		}
		// The native fast1 entry points deliberately return this second basin;
		// they do not compare it with the result before all-input inversion.
		state = other
	}
	if budget.stop != nil && budget.stop() {
		return fail(ErrCancelled)
	}
	return LuckyCanonicalization{
		Table:       state.table.Clone(),
		Permutation: slices.Clone(state.permutation),
		InputPhases: slices.Clone(state.phases),
		OutputPhase: state.output,
		Operations:  budget.operations,
	}, nil
}

// VerifyLuckyCanonicalization verifies transformation metadata independently
// by evaluating every packed assignment.
func VerifyLuckyCanonicalization(source *TruthTable, result LuckyCanonicalization) (bool, error) {
	if source == nil || result.Table == nil || source.variables != result.Table.variables {
		return false, ErrDimension
	}
	if err := source.Check(); err != nil {
		return false, err
	}
	if err := result.Table.Check(); err != nil {
		return false, err
	}
	n := source.variables
	if len(result.Permutation) != n || len(result.InputPhases) != n {
		return false, ErrDimension
	}
	seen := make([]bool, n)
	for _, sourceVariable := range result.Permutation {
		if sourceVariable < 0 || sourceVariable >= n || seen[sourceVariable] {
			return false, ErrDimension
		}
		seen[sourceVariable] = true
	}
	assignments := 1 << uint(n)
	for target := 0; target < assignments; target++ {
		sourceAssignment := 0
		for position, sourceVariable := range result.Permutation {
			bit := target>>uint(position)&1 != 0
			if result.InputPhases[position] {
				bit = !bit
			}
			if bit {
				sourceAssignment |= 1 << uint(sourceVariable)
			}
		}
		want := source.words[sourceAssignment>>6]>>uint(sourceAssignment&63)&1 != 0
		if result.OutputPhase {
			want = !want
		}
		got := result.Table.words[target>>6]>>uint(target&63)&1 != 0
		if got != want {
			return false, nil
		}
	}
	return true, nil
}

// LuckyTruthStore is an owned collection in luckyRead.c's hexadecimal format.
type LuckyTruthStore struct {
	variables int
	tables    []*TruthTable
}

func (s *LuckyTruthStore) Variables() int {
	if s == nil {
		return 0
	}
	return s.variables
}

func (s *LuckyTruthStore) Tables() []*TruthTable {
	if s == nil {
		return nil
	}
	tables := make([]*TruthTable, len(s.tables))
	for i, table := range s.tables {
		tables[i] = table.Clone()
	}
	return tables
}

func luckyVariablesFromDigits(digits int) (int, bool) {
	if digits <= 0 || digits&(digits-1) != 0 {
		return 0, false
	}
	bitCount := digits * 4
	variables := bits.TrailingZeros(uint(bitCount))
	return variables, variables >= 2 && variables <= LuckyMaximumVariables && 1<<uint(variables) == bitCount
}

// ParseLuckyTruthStore parses the initial block of fixed-width hex tables. A
// lowercase 0x prefix and trailing space-delimited fields are accepted as in C.
func ParseLuckyTruthStore(text string, maximumTables int) (*LuckyTruthStore, error) {
	if maximumTables < 0 {
		return nil, ErrDimension
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	store := &LuckyTruthStore{}
	for _, line := range lines {
		if line == "" {
			break
		}
		hex := line
		if end := strings.IndexAny(hex, " \r"); end >= 0 {
			hex = hex[:end]
		}
		if strings.HasPrefix(hex, "0x") {
			hex = hex[2:]
		}
		variables, ok := luckyVariablesFromDigits(len(hex))
		if !ok || store.variables != 0 && store.variables != variables {
			return nil, ErrDimension
		}
		if maximumTables != 0 && len(store.tables) >= maximumTables {
			return nil, ErrLimit
		}
		store.variables = variables
		words := make([]uint64, 1<<uint(max(0, variables-6)))
		for nibble := 0; nibble < len(hex); nibble++ {
			value, err := strconv.ParseUint(hex[len(hex)-1-nibble:len(hex)-nibble], 16, 4)
			if err != nil {
				return nil, ErrDimension
			}
			words[nibble>>4] |= value << uint(4*(nibble&15))
		}
		if variables < 6 {
			words[0], _ = NormalizeSmallTruthWord(words[0], variables)
		}
		table, err := TruthTableFromWords(variables, words, len(words))
		if err != nil {
			return nil, err
		}
		store.tables = append(store.tables, table)
	}
	if len(store.tables) == 0 {
		return nil, ErrDimension
	}
	return store, nil
}

func (s *LuckyTruthStore) Format() (string, error) {
	if s == nil || s.variables < 2 || s.variables > LuckyMaximumVariables || len(s.tables) == 0 {
		return "", ErrDimension
	}
	digits := 1 << uint(s.variables-2)
	var output strings.Builder
	for _, table := range s.tables {
		if table == nil || table.variables != s.variables {
			return "", ErrDimension
		}
		if err := table.Check(); err != nil {
			return "", err
		}
		for nibble := digits - 1; nibble >= 0; nibble-- {
			value := table.words[nibble>>4] >> uint(4*(nibble&15)) & 15
			output.WriteByte("0123456789ABCDEF"[value])
		}
		output.WriteByte('\n')
	}
	return output.String(), nil
}

func ReadLuckyTruthStore(path string, maximumTables int) (*LuckyTruthStore, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseLuckyTruthStore(string(contents), maximumTables)
}

func (s *LuckyTruthStore) Write(path string) error {
	text, err := s.Format()
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, []byte(text), 0o666); err != nil {
		return fmt.Errorf("write Lucky truth store: %w", err)
	}
	return nil
}
