package abc

// Port of src/sat/bsat/satInter.c.

import (
	"fmt"
	"slices"
	"strings"
)

const satTruthInterpolationMaxGlobals = 8

type SatTruthInterpolationOptions struct {
	GlobalVariables []SatVar
	Replay          SatProofCoreLimits
	MaxTruthWords   int // zero: 16M uint64 words
}

type SatTruthInterpolationStats struct {
	Variables, Roots, Clauses, Globals, Words int
	ResolutionSteps, RootAssignments          int
	Stopped                                   bool
}

type SatTruthInterpolation struct {
	Variables []SatVar
	Words     []uint64
	Bits      int
}

type SatTruthInterpolator struct {
	core          *SatProofCoreManager
	globals       []SatVar
	maxTruthWords int
	stats         SatTruthInterpolationStats
	busy, stopped bool
}

type satTruthInterpolationRun struct {
	variableTypes []int
	tables        [][]uint64
	words, bits   int
}

func checkSatTruthInterpolationGlobals(variables []SatVar) error {
	if len(variables) > satTruthInterpolationMaxGlobals {
		return ErrLimit
	}
	for i, variable := range variables {
		if _, err := MakeSatLit(variable, false); err != nil {
			return err
		}
		if slices.Contains(variables[:i], variable) {
			return ErrInvariant
		}
	}
	return nil
}

func NewSatTruthInterpolator(options SatTruthInterpolationOptions) (*SatTruthInterpolator, error) {
	if options.MaxTruthWords < 0 {
		return nil, ErrDimension
	}
	if err := checkSatTruthInterpolationGlobals(options.GlobalVariables); err != nil {
		return nil, err
	}
	core, err := NewSatProofCoreManager(options.Replay)
	if err != nil {
		return nil, err
	}
	maximum := options.MaxTruthWords
	if maximum == 0 {
		maximum = 16 << 20
	}
	return &SatTruthInterpolator{core: core, globals: slices.Clone(options.GlobalVariables), maxTruthWords: maximum}, nil
}

func (m *SatTruthInterpolator) SetGlobalVariables(variables []SatVar) error {
	if m == nil || m.stopped || m.busy {
		return ErrInvariant
	}
	if err := checkSatTruthInterpolationGlobals(variables); err != nil {
		return err
	}
	m.globals = slices.Clone(variables)
	return nil
}

func satTruthInterpolationVariables(run *satProofCoreRun, clausesA int, explicit []SatVar) ([]SatVar, []int, error) {
	types := make([]int, len(run.assigns))
	for clauseID := 0; clauseID < clausesA; clauseID++ {
		for _, literal := range run.clauses[clauseID].literals {
			types[literal.Var()] = 1
		}
	}
	if len(explicit) > 0 {
		for index, variable := range explicit {
			if int(variable) >= len(types) {
				return nil, nil, ErrInvalidLiteral
			}
			types[variable] = -index - 1
		}
		return slices.Clone(explicit), types, nil
	}
	for clauseID := clausesA; clauseID < run.manager.stats.Roots; clauseID++ {
		for _, literal := range run.clauses[clauseID].literals {
			variable := literal.Var()
			if types[variable] == 1 {
				types[variable] = -1
			}
		}
	}
	globals := make([]SatVar, 0)
	for variable := range types {
		if types[variable] == -1 {
			types[variable] = -len(globals) - 1
			globals = append(globals, SatVar(variable))
		}
	}
	if len(globals) > satTruthInterpolationMaxGlobals {
		return nil, nil, ErrLimit
	}
	return globals, types, nil
}

func newSatTruthInterpolationRun(run *satProofCoreRun, clausesA, maximumWords int, explicit []SatVar) (*satTruthInterpolationRun, []SatVar, error) {
	globals, variableTypes, err := satTruthInterpolationVariables(run, clausesA, explicit)
	if err != nil {
		return nil, nil, err
	}
	bits := 1 << len(globals)
	words := (bits + 63) / 64
	if words <= 0 || len(run.clauses) > maximumWords/words {
		return nil, nil, ErrLimit
	}
	interpolation := &satTruthInterpolationRun{
		variableTypes: variableTypes,
		tables:        make([][]uint64, len(run.clauses)),
		words:         words,
		bits:          bits,
	}
	for clauseID := range interpolation.tables {
		interpolation.tables[clauseID] = make([]uint64, words)
	}
	for clauseID := 0; clauseID < run.manager.stats.Roots; clauseID++ {
		if clauseID >= clausesA {
			interpolation.fill(clauseID)
			continue
		}
		for _, literal := range run.clauses[clauseID].literals {
			kind := variableTypes[literal.Var()]
			if kind >= 0 {
				continue
			}
			global := -kind - 1
			for assignment := 0; assignment < bits; assignment++ {
				value := assignment&(1<<global) != 0
				if literal.Negative() {
					value = !value
				}
				if value {
					interpolation.tables[clauseID][assignment/64] |= uint64(1) << (assignment % 64)
				}
			}
		}
	}
	return interpolation, globals, nil
}

func (r *satTruthInterpolationRun) lastMask() uint64 {
	if r.bits%64 == 0 {
		return ^uint64(0)
	}
	return uint64(1)<<(r.bits%64) - 1
}

func (r *satTruthInterpolationRun) fill(clauseID int) {
	for word := range r.tables[clauseID] {
		r.tables[clauseID][word] = ^uint64(0)
	}
	r.tables[clauseID][len(r.tables[clauseID])-1] &= r.lastMask()
}

func (r *satTruthInterpolationRun) copy(destination, source int) {
	copy(r.tables[destination], r.tables[source])
}

func (r *satTruthInterpolationRun) resolve(destination, reason, variable int) {
	if r.variableTypes[variable] == 1 {
		for word := range r.tables[destination] {
			r.tables[destination][word] |= r.tables[reason][word]
		}
		return
	}
	for word := range r.tables[destination] {
		r.tables[destination][word] &= r.tables[reason][word]
	}
}

func (m *SatTruthInterpolator) Interpolate(store *SatProofStore) (SatTruthInterpolation, error) {
	if m == nil || m.stopped || m.busy || m.core == nil || m.core.stopped {
		return SatTruthInterpolation{}, ErrInvariant
	}
	if store == nil || store.stopped || store.clausesA < 0 || store.clausesA > store.roots {
		return SatTruthInterpolation{}, ErrInvariant
	}
	m.busy = true
	defer func() { m.busy, m.core.busy = false, false }()
	run, err := m.core.newRun(store)
	m.core.busy = true
	if err != nil {
		return SatTruthInterpolation{}, err
	}
	interpolation, globals, err := newSatTruthInterpolationRun(run, store.clausesA, m.maxTruthWords, m.globals)
	if err != nil {
		return SatTruthInterpolation{}, err
	}
	run.interpolation = interpolation
	if err = run.derive(); err != nil {
		return SatTruthInterpolation{}, err
	}
	words := slices.Clone(interpolation.tables[run.empty])
	coreStats := m.core.Stats()
	m.stats = SatTruthInterpolationStats{
		Variables:       coreStats.Variables,
		Roots:           coreStats.Roots,
		Clauses:         coreStats.Clauses,
		Globals:         len(globals),
		Words:           len(words),
		ResolutionSteps: coreStats.ResolutionSteps,
		RootAssignments: coreStats.RootAssignments,
	}
	return SatTruthInterpolation{Variables: globals, Words: words, Bits: interpolation.bits}, nil
}

func (r SatTruthInterpolation) Evaluate(assignment uint64) (bool, error) {
	if r.Bits <= 0 || r.Bits > 1<<satTruthInterpolationMaxGlobals || len(r.Variables) > satTruthInterpolationMaxGlobals ||
		r.Bits != 1<<len(r.Variables) || len(r.Words) != (r.Bits+63)/64 || assignment >= uint64(r.Bits) {
		return false, ErrDimension
	}
	return r.Words[assignment/64]&(uint64(1)<<(assignment%64)) != 0, nil
}

func FormatSatTruthBits(words []uint64, bits int) (string, error) {
	if bits <= 0 || bits > 1<<satTruthInterpolationMaxGlobals || len(words) != (bits+63)/64 {
		return "", ErrDimension
	}
	var result strings.Builder
	result.Grow(bits)
	for bit := bits - 1; bit >= 0; bit-- {
		if words[bit/64]&(uint64(1)<<(bit%64)) != 0 {
			result.WriteByte('1')
		} else {
			result.WriteByte('0')
		}
	}
	return result.String(), nil
}

func FormatSatTruthInterpolationClause(clauseID int, interpolation SatTruthInterpolation) (string, error) {
	if clauseID < 0 {
		return "", ErrDimension
	}
	bits, err := FormatSatTruthBits(interpolation.Words, interpolation.Bits)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Clause %2d :  %s\n", clauseID, bits), nil
}

func (m *SatTruthInterpolator) Stats() SatTruthInterpolationStats {
	if m == nil {
		return SatTruthInterpolationStats{Stopped: true}
	}
	stats := m.stats
	stats.Stopped = m.stopped
	return stats
}

func (m *SatTruthInterpolator) Stop() {
	if m == nil || m.stopped || m.busy {
		return
	}
	m.core.Stop()
	m.globals = nil
	m.stats = SatTruthInterpolationStats{Stopped: true}
	m.stopped = true
}
