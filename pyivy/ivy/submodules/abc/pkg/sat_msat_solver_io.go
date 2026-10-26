package abc

// Deterministic diagnostics and DIMACS output from src/sat/msat/msatSolverIo.c.

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func (s *MsatSolver) AssignmentDiagnostic() (string, error) {
	if err := s.checkLive(); err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("Current assignments are: \n")
	for variable := 0; variable < s.VariableCount(); variable++ {
		builder.WriteByte(byte('0' + variable%10))
	}
	builder.WriteByte('\n')
	for _, assignment := range s.Assignments() {
		switch {
		case assignment < 0:
			builder.WriteByte('.')
		case assignment&1 != 0:
			builder.WriteByte('0')
		default:
			builder.WriteByte('1')
		}
	}
	builder.WriteByte('\n')
	return builder.String(), nil
}

func (s *MsatSolver) ClausesDiagnostic() (string, error) {
	if err := s.checkLive(); err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("Original clauses: \n")
	for index, clause := range s.original {
		fmt.Fprintf(&builder, "%3d: %s", index, FormatMsatClause(clause))
	}
	builder.WriteString("Learned clauses: \n")
	learnedIndex := 0
	for identifier := 1; identifier < len(s.core.clauses); identifier++ {
		clause := &s.core.clauses[identifier]
		if !clause.learnt || clause.deleted {
			continue
		}
		fmt.Fprintf(&builder, "%3d: Act = %.4f  %s", learnedIndex, float64(clause.activity), strings.TrimPrefix(msatClauseText(clause.lits), " "))
		learnedIndex++
	}
	builder.WriteString("Variable activity: \n")
	for variable := 0; variable < s.VariableCount(); variable++ {
		fmt.Fprintf(&builder, "%3d : %.4f\n", variable, float64(s.core.activity[variable]))
	}
	return builder.String(), nil
}

func formatMsatTimestamp(value time.Time) string {
	return value.Format("Mon Jan _2 15:04:05 2006")
}

func (s *MsatSolver) DIMACS(timestamp time.Time) ([]byte, error) {
	if err := s.checkLive(); err != nil {
		return nil, err
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	clauses := make([][]SatLit, 0)
	for identifier := 1; identifier < len(s.core.clauses); identifier++ {
		clause := &s.core.clauses[identifier]
		if clause.deleted || len(clause.lits) < 2 {
			continue
		}
		clauses = append(clauses, clause.lits)
	}
	assignments := s.Assignments()
	rootCount := 0
	for variable, assignment := range assignments {
		if assignment >= 0 && s.core.levels[variable] == 0 {
			rootCount++
		}
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "c Produced by Msat_SolverWriteDimacs() on %s\n", formatMsatTimestamp(timestamp))
	conflictCount := 0
	if s.core.inconsistent {
		conflictCount = 1
	}
	fmt.Fprintf(&builder, "p cnf %d %d\n", s.VariableCount(), len(clauses)+rootCount+conflictCount)
	for _, clause := range clauses {
		for _, literal := range clause {
			if literal.Negative() {
				fmt.Fprintf(&builder, "-%d ", literal.Var()+1)
			} else {
				fmt.Fprintf(&builder, "%d ", literal.Var()+1)
			}
		}
		builder.WriteString("0\n")
	}
	for variable, assignment := range assignments {
		if assignment < 0 || s.core.levels[variable] != 0 {
			continue
		}
		if assignment&1 != 0 {
			fmt.Fprintf(&builder, "-%d 0\n", variable+1)
		} else {
			fmt.Fprintf(&builder, "%d 0\n", variable+1)
		}
	}
	if conflictCount != 0 {
		builder.WriteString("0\n")
	}
	builder.WriteByte('\n')
	return []byte(builder.String()), nil
}

func (s *MsatSolver) WriteDIMACS(path string, timestamp time.Time) error {
	if path == "" {
		return ErrDimension
	}
	data, err := s.DIMACS(timestamp)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
