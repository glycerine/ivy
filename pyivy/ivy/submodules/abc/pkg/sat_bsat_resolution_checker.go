package abc

// Port of src/sat/bsat/satChecker.c as an owned resolution-proof checker.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SatResolutionProofLimits struct {
	MaxBytes, MaxClauses, MaxLiterals, MaxLineBytes int // zero: 64M, 16M, 64M, 4M
}

type SatResolutionProofReport struct {
	Roots, ResolutionSteps int
	Clauses                [][]int
	DerivedEmpty           bool
}

func checkSatResolutionProofLimits(limits SatResolutionProofLimits) (SatResolutionProofLimits, error) {
	if limits.MaxBytes < 0 || limits.MaxClauses < 0 || limits.MaxLiterals < 0 || limits.MaxLineBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 64 << 20
	}
	if limits.MaxClauses == 0 {
		limits.MaxClauses = 16 << 20
	}
	if limits.MaxLiterals == 0 {
		limits.MaxLiterals = 64 << 20
	}
	if limits.MaxLineBytes == 0 {
		limits.MaxLineBytes = 4 << 20
	}
	return limits, nil
}

func FormatSatProofClause(index int, clause []int) string {
	var result strings.Builder
	result.WriteString("Clause ")
	result.WriteString(strconv.Itoa(index))
	result.WriteString(":  {")
	for _, literal := range clause {
		result.WriteByte(' ')
		result.WriteString(strconv.Itoa(literal))
	}
	result.WriteString(" }\n")
	return result.String()
}

// ResolveSatProofClauses resolves two signed-DIMACS clauses. It preserves
// first occurrence order and requires exactly one complementary literal pair,
// including multiplicity, matching Sat_ProofResolve.
func ResolveSatProofClauses(first, second []int) ([]int, error) {
	pivot, pairs := 0, 0
	for _, left := range first {
		if left == 0 || left == mathMinInt() {
			return nil, ErrInvalidLiteral
		}
		for _, right := range second {
			if right == 0 || right == mathMinInt() {
				return nil, ErrInvalidLiteral
			}
			if left == -right {
				pivot, pairs = left, pairs+1
			}
		}
	}
	if pairs != 1 {
		return nil, fmt.Errorf("%w: resolution has %d pivot pairs", ErrInvariant, pairs)
	}
	result := make([]int, 0, len(first)+len(second)-2)
	pushUnique := func(literal int) {
		for _, old := range result {
			if old == literal {
				return
			}
		}
		result = append(result, literal)
	}
	for _, literal := range first {
		if literal != pivot && literal != -pivot {
			pushUnique(literal)
		}
	}
	if len(result)+1 != len(first) {
		return nil, fmt.Errorf("%w: first antecedent is not a unique clause", ErrInvariant)
	}
	for _, literal := range second {
		if literal != pivot && literal != -pivot {
			pushUnique(literal)
		}
	}
	return result, nil
}

func mathMinInt() int { return -int(^uint(0)>>1) - 1 }

func parseSatProofInteger(token string) (int, error) {
	value, err := strconv.ParseInt(token, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid proof integer %q", ErrDimension, token)
	}
	return int(value), nil
}

func CheckSatResolutionProof(data []byte, limits SatResolutionProofLimits) (SatResolutionProofReport, error) {
	limits, err := checkSatResolutionProofLimits(limits)
	if err != nil {
		return SatResolutionProofReport{}, err
	}
	if len(data) > limits.MaxBytes {
		return SatResolutionProofReport{}, ErrLimit
	}
	text := string(data)
	if strings.HasSuffix(text, "\n") {
		text = strings.TrimSuffix(text, "\n")
	}
	if text == "" {
		return SatResolutionProofReport{}, ErrDimension
	}
	lines := strings.Split(text, "\n")
	if len(lines) > limits.MaxClauses {
		return SatResolutionProofReport{}, ErrLimit
	}
	report := SatResolutionProofReport{Clauses: make([][]int, 1, len(lines)+1)}
	literalCount := 0
	for lineIndex, line := range lines {
		if len(line) > limits.MaxLineBytes || strings.TrimSpace(line) == "" {
			if len(line) > limits.MaxLineBytes {
				return SatResolutionProofReport{}, ErrLimit
			}
			return SatResolutionProofReport{}, fmt.Errorf("%w: empty proof line %d", ErrDimension, lineIndex+1)
		}
		fields := strings.Fields(line)
		index, parseErr := parseSatProofInteger(fields[0])
		if parseErr != nil || index != lineIndex+1 {
			return SatResolutionProofReport{}, fmt.Errorf("%w: expected clause %d", ErrInvariant, lineIndex+1)
		}
		var clause []int
		if len(fields) > 1 && fields[1] == "*" {
			if len(fields) != 5 || fields[4] != "0" {
				return SatResolutionProofReport{}, fmt.Errorf("%w: malformed resolution line %d", ErrDimension, index)
			}
			left, leftErr := parseSatProofInteger(fields[2])
			right, rightErr := parseSatProofInteger(fields[3])
			if leftErr != nil || rightErr != nil || left < 1 || right < 1 || left >= index || right >= index {
				return SatResolutionProofReport{}, fmt.Errorf("%w: antecedents on line %d", ErrInvariant, index)
			}
			clause, err = ResolveSatProofClauses(report.Clauses[left], report.Clauses[right])
			if err != nil {
				return SatResolutionProofReport{}, fmt.Errorf("clause %d resolve(%d,%d): %w", index, left, right, err)
			}
			report.ResolutionSteps++
		} else {
			if len(fields) < 3 || fields[len(fields)-1] != "0" || fields[len(fields)-2] != "0" {
				return SatResolutionProofReport{}, fmt.Errorf("%w: malformed root line %d", ErrDimension, index)
			}
			clause = make([]int, len(fields)-3)
			for i, token := range fields[1 : len(fields)-2] {
				literal, literalErr := parseSatProofInteger(token)
				if literalErr != nil || literal == 0 || literal == mathMinInt() {
					return SatResolutionProofReport{}, fmt.Errorf("%w: root literal on line %d", ErrInvalidLiteral, index)
				}
				clause[i] = literal
			}
			report.Roots++
		}
		if len(clause) > limits.MaxLiterals-literalCount {
			return SatResolutionProofReport{}, ErrLimit
		}
		literalCount += len(clause)
		report.Clauses = append(report.Clauses, clause)
	}
	report.DerivedEmpty = len(report.Clauses[len(report.Clauses)-1]) == 0
	return report, nil
}

func CheckSatResolutionProofFile(name string, limits SatResolutionProofLimits) (SatResolutionProofReport, error) {
	limits, err := checkSatResolutionProofLimits(limits)
	if err != nil {
		return SatResolutionProofReport{}, err
	}
	info, err := os.Stat(name)
	if err != nil {
		return SatResolutionProofReport{}, err
	}
	if info.Size() > int64(limits.MaxBytes) {
		return SatResolutionProofReport{}, ErrLimit
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return SatResolutionProofReport{}, err
	}
	return CheckSatResolutionProof(data, limits)
}
