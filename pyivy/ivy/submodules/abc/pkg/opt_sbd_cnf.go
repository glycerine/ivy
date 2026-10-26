package abc

// Owned CNF generation, formatting and translation from src/opt/sbd/sbdCnf.c.

import (
	"strconv"
	"strings"
)

type SBDCNFOptions struct {
	MaxClauses, MaxLiterals, MaxBytes int
	ISOP                              TruthISOPOptions
	Stop                              func() bool
}

type SBDCNF struct {
	variables int
	clauses   [][]SatLit
}

func (cnf *SBDCNF) VariableCount() int {
	if cnf == nil {
		return 0
	}
	return cnf.variables
}

func (cnf *SBDCNF) CubeCount() int {
	if cnf == nil {
		return 0
	}
	return len(cnf.clauses)
}

func (cnf *SBDCNF) Clauses() [][]SatLit {
	if cnf == nil {
		return nil
	}
	result := make([][]SatLit, len(cnf.clauses))
	for index, clause := range cnf.clauses {
		result[index] = append([]SatLit(nil), clause...)
	}
	return result
}

func normalizeSBDCNFOptions(options SBDCNFOptions) (SBDCNFOptions, error) {
	if options.MaxClauses < 0 || options.MaxLiterals < 0 || options.MaxBytes < 0 || options.ISOP.MaxElements < 0 || options.ISOP.MaxVisits < 0 {
		return options, ErrDimension
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 1 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 4 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 20
	}
	return options, nil
}

func checkSBDCNF(cnf *SBDCNF) error {
	if cnf == nil || cnf.variables < 0 || uint64(cnf.variables) > uint64(MaxLiteralID) {
		return ErrDimension
	}
	for _, clause := range cnf.clauses {
		if len(clause) == 0 {
			return ErrInvariant
		}
		for _, literal := range clause {
			if !literal.ValidForVariables(cnf.variables) {
				return ErrInvalidLiteral
			}
		}
	}
	return nil
}

// SBDTruthToCNF emits the direct and complement ISOP clauses defining output
// variable inputVariables. A zero-input constant uses variable zero, matching
// the native encoding.
func SBDTruthToCNF(truth uint64, inputVariables int, options SBDCNFOptions) (*SBDCNF, error) {
	options, err := normalizeSBDCNFOptions(options)
	if err != nil || inputVariables < 0 || inputVariables > 6 {
		return nil, ErrDimension
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	table, err := TruthTableFromWords(inputVariables, []uint64{truth}, 1)
	if err != nil {
		return nil, err
	}
	if inputVariables == 0 {
		literal, literalErr := MakeSatLit(0, truth == 0)
		if literalErr != nil {
			return nil, literalErr
		}
		if options.MaxClauses < 1 || options.MaxLiterals < 1 {
			return nil, ErrLimit
		}
		return &SBDCNF{variables: 1, clauses: [][]SatLit{{literal}}}, nil
	}
	variables := make([]SatVar, inputVariables)
	for index := range variables {
		variables[index] = SatVar(index)
	}
	priorStop := options.ISOP.Stop
	options.ISOP.TryBoth = false
	options.ISOP.Stop = func() bool {
		return options.Stop != nil && options.Stop() || priorStop != nil && priorStop()
	}
	covers := make([]TruthISOP, 2)
	covers[0], err = table.ISOP(options.ISOP)
	if err == nil {
		covers[1], err = table.Not().ISOP(options.ISOP)
	}
	if err != nil {
		return nil, err
	}
	clauses := make([][]SatLit, 0, len(covers[0].Cubes)+len(covers[1].Cubes))
	literalCount := 0
	for polarity, cover := range covers {
		for _, cube := range cover.Cubes {
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			if len(clauses) == options.MaxClauses {
				return nil, ErrLimit
			}
			body, bodyErr := CNFISOPCubeLiterals(cube, variables)
			if bodyErr != nil {
				return nil, bodyErr
			}
			if len(body)+1 > options.MaxLiterals-literalCount {
				return nil, ErrLimit
			}
			output, outputErr := MakeSatLit(SatVar(inputVariables), polarity != 0)
			if outputErr != nil {
				return nil, outputErr
			}
			clause := make([]SatLit, len(body), len(body)+1)
			copy(clause, body)
			clause = append(clause, output)
			clauses = append(clauses, clause)
			literalCount += len(clause)
		}
	}
	return &SBDCNF{variables: inputVariables + 1, clauses: clauses}, nil
}

// SBDFormatCNF returns Sbd_PrintCnf's zero-based literal text without writing
// process output.
func SBDFormatCNF(cnf *SBDCNF, options SBDCNFOptions) (string, error) {
	options, err := normalizeSBDCNFOptions(options)
	if err != nil {
		return "", err
	}
	if err = checkSBDCNF(cnf); err != nil {
		return "", err
	}
	var builder strings.Builder
	write := func(part string) error {
		if len(part) > options.MaxBytes-builder.Len() {
			return ErrLimit
		}
		builder.WriteString(part)
		return nil
	}
	for _, clause := range cnf.clauses {
		for _, literal := range clause {
			if options.Stop != nil && options.Stop() {
				return "", ErrCancelled
			}
			prefix := ""
			if literal.Negative() {
				prefix = "-"
			}
			if err = write(prefix + strconv.FormatUint(uint64(literal.Var()), 10) + " "); err != nil {
				return "", err
			}
		}
		if err = write("\n"); err != nil {
			return "", err
		}
	}
	return builder.String(), nil
}

// SBDTranslateCNF substitutes local variables through faninMap and flips every
// literal mapped to pivot. Structured clauses intentionally omit the native
// vector routine's final unused empty level.
func SBDTranslateCNF(cnf *SBDCNF, faninMap []SatVar, pivot SatVar, options SBDCNFOptions) (*SBDCNF, error) {
	options, err := normalizeSBDCNFOptions(options)
	if err != nil {
		return nil, err
	}
	if err = checkSBDCNF(cnf); err != nil {
		return nil, err
	}
	if len(faninMap) < cnf.variables || uint32(pivot) > MaxLiteralID {
		return nil, ErrDimension
	}
	if len(cnf.clauses) > options.MaxClauses {
		return nil, ErrLimit
	}
	maximum, literalCount := uint32(0), 0
	for _, variable := range faninMap[:cnf.variables] {
		if uint32(variable) > MaxLiteralID {
			return nil, ErrInvalidLiteral
		}
		if uint32(variable) > maximum {
			maximum = uint32(variable)
		}
	}
	result := &SBDCNF{variables: int(maximum) + 1, clauses: make([][]SatLit, len(cnf.clauses))}
	for clauseIndex, clause := range cnf.clauses {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if len(clause) > options.MaxLiterals-literalCount {
			return nil, ErrLimit
		}
		translated := make([]SatLit, len(clause))
		for literalIndex, literal := range clause {
			mapped := faninMap[literal.Var()]
			translatedLiteral, literalErr := MakeSatLit(mapped, literal.Negative())
			if literalErr != nil {
				return nil, literalErr
			}
			if mapped == pivot {
				translatedLiteral = translatedLiteral.Not()
			}
			translated[literalIndex] = translatedLiteral
		}
		result.clauses[clauseIndex] = translated
		literalCount += len(clause)
	}
	return result, nil
}
