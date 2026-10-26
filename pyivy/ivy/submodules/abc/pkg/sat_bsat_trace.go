package abc

// Port of src/sat/bsat/satTrace.c as a checked streaming trace writer.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const satTraceHeaderWidth = 40

type SatTraceLimits struct {
	MaxClauses, MaxLiterals int // zero: 16M clauses, 64M literals
}

type SatTraceStats struct {
	Clauses, Roots, Literals int
	Open                     bool
}

type SatSolverTrace struct {
	file                     *os.File
	limits                   SatTraceLimits
	clauses, roots, literals int
}

func checkSatTraceLimits(limits SatTraceLimits) (SatTraceLimits, error) {
	if limits.MaxClauses < 0 || limits.MaxLiterals < 0 {
		return limits, ErrDimension
	}
	if limits.MaxClauses == 0 {
		limits.MaxClauses = 16 << 20
	}
	if limits.MaxLiterals == 0 {
		limits.MaxLiterals = 64 << 20
	}
	return limits, nil
}

func StartSatSolverTrace(name string, limits SatTraceLimits) (*SatSolverTrace, error) {
	limits, err := checkSatTraceLimits(limits)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, ErrDimension
	}
	file, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		return nil, err
	}
	placeholder := strings.Repeat(" ", satTraceHeaderWidth) + "\n"
	if _, err = file.WriteString(placeholder); err != nil {
		_ = file.Close()
		return nil, err
	}
	return &SatSolverTrace{file: file, limits: limits}, nil
}

func FormatSatTraceClause(literals []SatLit) (string, error) {
	var result strings.Builder
	for _, literal := range literals {
		if !literal.Valid() {
			return "", ErrInvalidLiteral
		}
		value := int64(literal.Var()) + 1
		if literal.Negative() {
			value = -value
		}
		result.WriteByte(' ')
		result.WriteString(strconv.FormatInt(value, 10))
	}
	result.WriteString(" 0\n")
	return result.String(), nil
}

func (t *SatSolverTrace) Write(literals []SatLit, root bool) error {
	if t == nil || t.file == nil {
		return ErrInvariant
	}
	if t.clauses == t.limits.MaxClauses || len(literals) > t.limits.MaxLiterals-t.literals {
		return ErrLimit
	}
	line, err := FormatSatTraceClause(literals)
	if err != nil {
		return err
	}
	if _, err = t.file.WriteString(line); err != nil {
		return err
	}
	t.clauses++
	t.literals += len(literals)
	if root {
		t.roots++
	}
	return nil
}

func (t *SatSolverTrace) Stats() SatTraceStats {
	if t == nil {
		return SatTraceStats{}
	}
	return SatTraceStats{t.clauses, t.roots, t.literals, t.file != nil}
}

func (t *SatSolverTrace) Stop(variableCount int) error {
	if t == nil || t.file == nil {
		return nil
	}
	if variableCount < 0 {
		return ErrDimension
	}
	header := fmt.Sprintf("p %d %d %d", variableCount, t.clauses, t.roots)
	if len(header) > satTraceHeaderWidth {
		return ErrLimit
	}
	file := t.file
	t.file = nil
	if _, err := file.Seek(0, 0); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.WriteString(header); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func (t *SatSolverTrace) StopForSolver(solver *SatSolver) error {
	if solver == nil {
		return ErrDimension
	}
	return t.Stop(solver.VariableCount())
}
