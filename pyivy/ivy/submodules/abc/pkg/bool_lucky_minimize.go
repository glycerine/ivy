package abc

// Exact truth minimizers from bool/lucky/luckySimple.c:
// simpleMinimal, simpleMinimalGroups and minWord/minWord1/minWord3.

type TruthExactOptions struct {
	MaxScheduleSteps int // zero: one million stored swaps/flips, shared across sizes
	MaxCandidates    int // zero: one million packed candidates, including both outputs
	MaxTransitions   int // zero: one million mixed-radix iterations, including no-op carries
	Stop             func() bool
}

// TruthExactMinimization owns the minimum packed truth. Candidates counts
// compared tables, including output complements and mixed-radix intermediate
// transitions/repeats. Changes counts strict minimum replacements. These APIs
// return a representative; C's minimizers do not retain transformation metadata.
type TruthExactMinimization struct {
	Table               *TruthTable
	Candidates, Changes int
	Transitions         int // mixed-radix iterations; zero for unrestricted simple search
}
type truthExactSearch struct {
	options TruthExactOptions
	result  TruthExactMinimization
	limit   int
}

func newTruthExactSearch(t *TruthTable, options TruthExactOptions) (*truthExactSearch, error) {
	if options.MaxScheduleSteps < 0 || options.MaxCandidates < 0 || options.MaxTransitions < 0 {
		return nil, ErrDimension
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	if e := t.Check(); e != nil {
		return nil, e
	}
	limit := options.MaxCandidates
	if limit == 0 {
		limit = 1 << 20
	}
	return &truthExactSearch{options: options, limit: limit}, nil
}
func (s *truthExactSearch) visit(table *TruthTable, output bool) error {
	count := 1
	if output {
		count = 2
	}
	for polarity := 0; polarity < count; polarity++ {
		if s.options.Stop != nil && s.options.Stop() {
			return ErrCancelled
		}
		if s.result.Candidates >= s.limit {
			return ErrLimit
		}
		s.result.Candidates++
		candidate := table
		if polarity != 0 {
			candidate = table.Not()
		}
		if s.result.Table == nil || truthWordsLessReverse(candidate.words, s.result.Table.words) {
			s.result.Table = candidate.Clone()
			s.result.Changes++
		}
	}
	return nil
}
func (s *truthExactSearch) finish() (TruthExactMinimization, error) {
	if s.options.Stop != nil && s.options.Stop() {
		return TruthExactMinimization{}, ErrCancelled
	}
	return s.result, nil
}

// MinimizeNPN exhaustively minimizes over all input permutations/phases and
// both output phases using C's simpleMinimal traversal. The single-input case
// skips C's invalid swap with dummy input 1, preserving canonical small-table
// padding and the actual NPN orbit. Limits/cancellation return no partial result.
func (t *TruthTable) MinimizeNPN(options TruthExactOptions) (TruthExactMinimization, error) {
	search, e := newTruthExactSearch(t, options)
	if e != nil {
		return TruthExactMinimization{}, e
	}
	schedule, e := NewTruthPermutationSchedule(t.variables, options.MaxScheduleSteps)
	if e != nil {
		return TruthExactMinimization{}, e
	}
	work := t.Clone()
	if e = search.visit(work, true); e != nil {
		return TruthExactMinimization{}, e
	}
	walk := func() error {
		for _, index := range schedule.swaps {
			work.swapVariables(index, index+1)
			if e := search.visit(work, true); e != nil {
				return e
			}
		}
		return nil
	}
	if e = walk(); e != nil {
		return TruthExactMinimization{}, e
	}
	for _, variable := range schedule.flips {
		if t.variables >= 2 {
			work.swapVariables(0, 1)
		}
		work.flipVariable(variable)
		if e = search.visit(work, true); e != nil {
			return TruthExactMinimization{}, e
		}
		if e = walk(); e != nil {
			return TruthExactMinimization{}, e
		}
	}
	return search.finish()
}

// MinimizeGroups ports simpleMinimalGroups over adjacent groups whose positive
// sizes sum to Variables(). All input permutations and phases inside groups
// are explored; flipSingletonInputs controls only size-one input phases,
// matching C. flipOutput includes both output polarities. Group order stays
// fixed. Mixed-radix carry transitions, including repeats, preserve native
// minimum behavior. Schedules are built only for sizes actually used.
func (t *TruthTable) MinimizeGroups(groups []int, flipOutput, flipSingletonInputs bool, options TruthExactOptions) (TruthExactMinimization, error) {
	search, e := newTruthExactSearch(t, options)
	if e != nil {
		return TruthExactMinimization{}, e
	}
	sum := 0
	for _, size := range groups {
		if size <= 0 || size > t.variables-sum {
			return TruthExactMinimization{}, ErrDimension
		}
		sum += size
	}
	if sum != t.variables {
		return TruthExactMinimization{}, ErrDimension
	}
	scheduleLimit := options.MaxScheduleSteps
	if scheduleLimit == 0 {
		scheduleLimit = 1 << 20
	}
	schedules := make([]*TruthPermutationSchedule, t.variables+1)
	offsets := make([]int, len(groups))
	sum = 0
	for i, size := range groups {
		offsets[i] = sum
		sum += size
		if schedules[size] == nil {
			if scheduleLimit <= 0 {
				return TruthExactMinimization{}, ErrLimit
			}
			schedule, e := NewTruthPermutationSchedule(size, scheduleLimit)
			if e != nil {
				return TruthExactMinimization{}, e
			}
			schedules[size] = schedule
			scheduleLimit -= len(schedule.swaps) + len(schedule.flips)
		}
	}
	digits, changed, radices := make([]int, 2*len(groups)+1), make([]bool, 2*len(groups)+1), make([]int, 2*len(groups)+1)
	radices[0] = 2
	for i, size := range groups {
		radices[i+1] = len(schedules[size].flips) + 1
		radices[len(groups)+i+1] = len(schedules[size].swaps) + 1
	}
	work := t.Clone()
	if e = search.visit(work, flipOutput); e != nil {
		return TruthExactMinimization{}, e
	}
	transitionLimit := options.MaxTransitions
	if transitionLimit == 0 {
		transitionLimit = 1 << 20
	}
	for {
		if options.Stop != nil && options.Stop() {
			return TruthExactMinimization{}, ErrCancelled
		}
		if search.result.Transitions >= transitionLimit {
			return TruthExactMinimization{}, ErrLimit
		}
		search.result.Transitions++

		for i, size := range groups {
			digit := i + 1
			if !changed[digit] || size == 1 && !flipSingletonInputs {
				continue
			}
			flips := schedules[size].flips
			index := digits[digit] - 1
			if digits[digit] == 0 {
				index = len(flips) - 1
			}
			work.flipVariable(offsets[i] + flips[index])
			if e = search.visit(work, flipOutput); e != nil {
				return TruthExactMinimization{}, e
			}
		}
		for i, size := range groups {
			digit := len(groups) + i + 1
			if !changed[digit] || size == 1 {
				continue
			}
			swaps := schedules[size].swaps
			index := 0
			if digits[digit] != len(swaps) {
				index = swaps[digits[digit]]
			}
			work.swapVariables(offsets[i]+index, offsets[i]+index+1)
			if e = search.visit(work, flipOutput); e != nil {
				return TruthExactMinimization{}, e
			}
		}
		clear(changed)
		digit := len(digits) - 1
		for digits[digit] == radices[digit]-1 {
			changed[digit] = true
			digits[digit] = 0
			digit--
		}
		if digit == 0 {
			break
		}
		changed[digit] = true
		digits[digit]++
	}
	return search.finish()
}
