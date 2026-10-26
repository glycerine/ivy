package abc

// Input-phase canonicalization from src/misc/extra/extraUtilCanon.c.

import "slices"

type ExtraTruthNCanonicalOptions struct {
	MaxCandidates int
	Stop          func() bool
}

type ExtraTruthNCanonicalization struct {
	Table               *TruthTable
	Phases              []uint8
	RealVariables       int
	EnumeratedVariables int
	Candidates          int
}

// ExtraTruthCanonFastN finds the numerically smallest truth table under input
// complementation. For zero through three real inputs it deliberately
// enumerates three phase bits, preserving the source lookup-table multiplicity
// for unused inputs. All returned storage is owned.
func ExtraTruthCanonFastN(table *TruthTable, realVariables int, options ExtraTruthNCanonicalOptions) (ExtraTruthNCanonicalization, error) {
	fail := func(e error) (ExtraTruthNCanonicalization, error) {
		return ExtraTruthNCanonicalization{}, e
	}
	if e := table.Check(); e != nil {
		return fail(e)
	}
	if table.variables > 6 || realVariables < 0 || realVariables > table.variables || options.MaxCandidates < 0 {
		return fail(ErrDimension)
	}
	for variable := realVariables; variable < table.variables; variable++ {
		used, e := table.HasVariable(variable)
		if e != nil {
			return fail(e)
		}
		if used {
			return fail(ErrUnsupported)
		}
	}
	enumerated := realVariables
	if enumerated < 3 {
		enumerated = 3
	}
	candidates := 1 << uint(enumerated)
	limit := options.MaxCandidates
	if limit == 0 {
		limit = 64
	}
	if candidates > limit {
		return fail(ErrLimit)
	}
	var best *TruthTable
	phases := make([]uint8, 0, candidates)
	for phase := 0; phase < candidates; phase++ {
		if options.Stop != nil && options.Stop() {
			return fail(ErrCancelled)
		}
		candidate := table.Clone()
		for variable := 0; variable < realVariables; variable++ {
			if phase&(1<<uint(variable)) != 0 {
				candidate.flipVariable(variable)
			}
		}
		if best == nil || candidate.words[0] < best.words[0] {
			best = candidate
			phases = phases[:0]
			phases = append(phases, uint8(phase))
		} else if candidate.words[0] == best.words[0] {
			phases = append(phases, uint8(phase))
		}
	}
	return ExtraTruthNCanonicalization{
		Table:               best,
		Phases:              slices.Clone(phases),
		RealVariables:       realVariables,
		EnumeratedVariables: enumerated,
		Candidates:          candidates,
	}, nil
}

type ExtraTruth3CanonicalRow struct {
	Truth, Canonical uint8
	Phases           []uint8
}

// ExtraTruth3CanonicalRows replaces the source's static 256-row tables and
// print/test drivers with deterministic caller-owned data.
func ExtraTruth3CanonicalRows(options ExtraTruthNCanonicalOptions) ([]ExtraTruth3CanonicalRow, error) {
	rows := make([]ExtraTruth3CanonicalRow, 256)
	for truth := range rows {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		table, e := NewConstantTruthTable(3, false, 0)
		if e != nil {
			return nil, e
		}
		for assignment := 0; assignment < 8; assignment++ {
			if truth&(1<<uint(assignment)) != 0 {
				if e = table.Set(uint64(assignment), true); e != nil {
					return nil, e
				}
			}
		}
		canonical, e := ExtraTruthCanonFastN(table, 3, options)
		if e != nil {
			return nil, e
		}
		rows[truth] = ExtraTruth3CanonicalRow{
			Truth:     uint8(truth),
			Canonical: uint8(canonical.Table.words[0]),
			Phases:    slices.Clone(canonical.Phases),
		}
	}
	return rows, nil
}
