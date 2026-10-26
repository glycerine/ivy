package abc

// Port of src/sat/bsat/satUtil.c.

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type SatSolverDIMACSOptions struct {
	IncrementVariables bool
	MaxBytes           int // zero: 1 GiB
	Stop               func() bool
}

func FormatSatSolverClauseDIMACS(clause []SatLit, incrementVariables bool) (string, error) {
	var result strings.Builder
	for _, literal := range clause {
		if !literal.Valid() {
			return "", ErrInvalidLiteral
		}
		if literal.Negative() {
			result.WriteByte('-')
		}
		variable := uint64(literal.Var())
		if incrementVariables {
			variable++
		}
		fmt.Fprintf(&result, "%d ", variable)
	}
	if incrementVariables {
		result.WriteByte('0')
	}
	result.WriteByte('\n')
	return result.String(), nil
}

func normalizeSatSolverDIMACSOptions(options SatSolverDIMACSOptions) (SatSolverDIMACSOptions, error) {
	if options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 30
	}
	return options, nil
}

func SatSolverDIMACS(solver *SatSolver, assumptions []SatLit, options SatSolverDIMACSOptions) ([]byte, error) {
	if solver == nil || solver.solving || len(solver.clauses) == 0 {
		return nil, ErrInvariant
	}
	options, err := normalizeSatSolverDIMACSOptions(options)
	if err != nil {
		return nil, err
	}
	for _, literal := range assumptions {
		if !literal.ValidForVariables(len(solver.assigns)) {
			return nil, ErrInvalidLiteral
		}
	}
	work := solver.cloneState()
	work.bookmark = nil
	work.prepareRoot(nil)
	clauses := make([][]SatLit, 0)
	for id, clause := range work.clauses {
		if id == 0 || clause.deleted || clause.learnt || len(clause.lits) <= 1 {
			continue
		}
		clauses = append(clauses, slices.Clone(clause.lits))
	}
	units := make([]SatLit, 0)
	for variable, value := range work.assigns {
		if value < 0 || work.levels[variable] != 0 {
			continue
		}
		literal, makeErr := MakeSatLit(SatVar(variable), value == 0)
		if makeErr != nil {
			return nil, makeErr
		}
		units = append(units, literal)
	}
	if work.inconsistent {
		clauses = append(clauses, nil)
	}
	var result strings.Builder
	appendText := func(text string) error {
		if len(text) > options.MaxBytes-result.Len() {
			return ErrLimit
		}
		result.WriteString(text)
		return nil
	}
	if err = appendText(fmt.Sprintf("p cnf %d %d\n", len(work.assigns), len(clauses)+len(units)+len(assumptions))); err != nil {
		return nil, err
	}
	writeClause := func(clause []SatLit) error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		text, formatErr := FormatSatSolverClauseDIMACS(clause, options.IncrementVariables)
		if formatErr != nil {
			return formatErr
		}
		return appendText(text)
	}
	for _, clause := range clauses {
		if err = writeClause(clause); err != nil {
			return nil, err
		}
	}
	for _, literal := range units {
		if err = writeClause([]SatLit{literal}); err != nil {
			return nil, err
		}
	}
	for _, literal := range assumptions {
		if err = writeClause([]SatLit{literal}); err != nil {
			return nil, err
		}
	}
	if err = appendText("\n"); err != nil {
		return nil, err
	}
	return []byte(result.String()), nil
}

func WriteSatSolverDIMACSFile(solver *SatSolver, name string, assumptions []SatLit, options SatSolverDIMACSOptions) error {
	if name == "" {
		return ErrDimension
	}
	data, err := SatSolverDIMACS(solver, assumptions, options)
	if err != nil {
		return err
	}
	return os.WriteFile(name, data, 0o666)
}

func SatLegacyCounterFloat(value uint64) float64 {
	return float64(uint32(value&0x3fffffff)) + float64(uint64(1)<<30)*float64(value>>30)
}

func FormatSatSolverStatistics(stats SatStats, secondGeneration bool) string {
	if secondGeneration {
		return fmt.Sprintf("starts        : %10d\nconflicts     : %10d\ndecisions     : %10d\npropagations  : %10d\n",
			stats.Restarts, stats.Conflicts, stats.Decisions, stats.Propagations)
	}
	return fmt.Sprintf("starts        : %16.0f\nconflicts     : %16.0f\ndecisions     : %16.0f\npropagations  : %16.0f\n",
		SatLegacyCounterFloat(stats.Restarts), SatLegacyCounterFloat(stats.Conflicts),
		SatLegacyCounterFloat(stats.Decisions), SatLegacyCounterFloat(stats.Propagations))
}

func SatSolver2VariableBytes(pointerBytes int, floatActivity bool) (int, error) {
	if pointerBytes != 4 && pointerBytes != 8 {
		return 0, ErrDimension
	}
	vectorBytes := 8 + pointerBytes
	if remainder := vectorBytes % pointerBytes; remainder != 0 {
		vectorBytes += pointerBytes - remainder
	}
	activityBytes := 4
	if floatActivity {
		activityBytes = 8
	}
	return activityBytes + 2*vectorBytes + 7*4, nil
}

func NativeSatSolver2VariableBytes() int {
	bytes, _ := SatSolver2VariableBytes(strconv.IntSize/8, false)
	return bytes
}

func SatSolverModel(result SatResult, variables []SatVar) ([]int, error) {
	if result.Status != SatSatisfiable || len(result.Model) == 0 {
		return nil, ErrInvariant
	}
	model := make([]int, len(variables))
	for i, variable := range variables {
		if int(variable) >= len(result.Model) {
			return nil, ErrInvalidLiteral
		}
		if result.Model[variable] {
			model[i] = 1
		}
	}
	return model, nil
}

func (s *SatSolver) DoubleClauses(complementVariable SatVar) error {
	if s == nil || s.solving || int(complementVariable) >= len(s.assigns) {
		return ErrInvalidLiteral
	}
	return ErrUnsupported
}
