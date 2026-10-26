package abc

// Tests for src/misc/parse/parseEqn.c and parseStack.c.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

type parseEquationOracle struct {
	Cases []struct {
		Formula  string `json:"formula"`
		Accepted int    `json:"accepted"`
		Truth    string `json:"truth"`
	} `json:"cases"`
	Invalid []struct {
		Formula  string `json:"formula"`
		Accepted int    `json:"accepted"`
	} `json:"invalid"`
}

func parseEquationManager(t testing.TB) *AigManager {
	t.Helper()
	manager := NewAigManager()
	for index := 0; index < 4; index++ {
		if _, err := manager.CreateCI(); err != nil {
			t.Fatal(err)
		}
	}
	return manager
}

func parseEquationTruth(t testing.TB, manager *AigManager, literal AigLit) uint64 {
	t.Helper()
	var truth uint64
	inputs := make([]bool, 4)
	for assignment := 0; assignment < 64; assignment++ {
		for variable := range inputs {
			inputs[variable] = assignment&(1<<variable) != 0
		}
		if aigTestLiteralValue(t, manager, literal, inputs) {
			truth |= uint64(1) << assignment
		}
	}
	return truth
}

func TestParseBooleanFormulaNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/parse_equation.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle parseEquationOracle
	if err = json.Unmarshal(data, &oracle); err != nil || len(oracle.Cases) != 17 || len(oracle.Invalid) != 9 {
		t.Fatal(len(oracle.Cases), len(oracle.Invalid), err)
	}
	names := []string{"a", "b", "c", "aa"}
	for _, row := range oracle.Cases {
		manager := parseEquationManager(t)
		literal, parseErr := ParseBooleanFormula(manager, row.Formula, names, BooleanFormulaParseOptions{})
		if parseErr != nil {
			t.Fatal(row.Formula, literal, parseErr)
		}
		truth := parseEquationTruth(t, manager, literal)
		if row.Accepted != 1 || fmt.Sprintf("%016x", truth) != row.Truth {
			t.Fatal(row.Formula, literal, parseErr, fmt.Sprintf("%016x", truth), row)
		}
		if err = manager.Check(); err != nil {
			t.Fatal(row.Formula, err)
		}
	}
	for _, row := range oracle.Invalid {
		manager := parseEquationManager(t)
		before := manager.Clone()
		literal, parseErr := ParseBooleanFormula(manager, row.Formula, names, BooleanFormulaParseOptions{})
		if literal != InvalidAigLit || parseErr == nil || row.Accepted != 0 || !reflect.DeepEqual(manager, before) {
			t.Fatal(row.Formula, literal, parseErr, row.Accepted)
		}
	}
}

func TestParseBooleanFormulaErrorsLimitsAndAtomicity(t *testing.T) {
	names := []string{"a", "b", "c", "aa"}
	manager := parseEquationManager(t)
	baseline := manager.Clone()
	tests := []struct {
		formula string
		names   []string
		options BooleanFormulaParseOptions
		want    error
	}{
		{"a\x00+b", names, BooleanFormulaParseOptions{}, ErrFormat},
		{"a+b", names, BooleanFormulaParseOptions{MaxTokens: 2}, ErrLimit},
		{"((a))", names, BooleanFormulaParseOptions{MaxDepth: 1}, ErrLimit},
		{"a", []string{""}, BooleanFormulaParseOptions{}, ErrDimension},
		{"a", []string{"0bad"}, BooleanFormulaParseOptions{}, ErrDimension},
		{"a", []string{"bad+name"}, BooleanFormulaParseOptions{}, ErrDimension},
		{"a", names, BooleanFormulaParseOptions{MaxTokens: -1}, ErrDimension},
		{"a", names, BooleanFormulaParseOptions{MaxDepth: -1}, ErrDimension},
		{"a", names, BooleanFormulaParseOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for _, test := range tests {
		literal, err := ParseBooleanFormula(manager, test.formula, test.names, test.options)
		if literal != InvalidAigLit || !errors.Is(err, test.want) || !reflect.DeepEqual(manager, baseline) {
			t.Fatal(test.formula, literal, err, test.want)
		}
	}
	calls := 0
	if literal, err := ParseBooleanFormula(manager, "a", names, BooleanFormulaParseOptions{Stop: func() bool {
		calls++
		return calls > 1
	}}); literal != InvalidAigLit || !errors.Is(err, ErrCancelled) || !reflect.DeepEqual(manager, baseline) {
		t.Fatal(literal, err, calls)
	}
	if literal, err := ParseBooleanFormula(nil, "a", names, BooleanFormulaParseOptions{}); literal != InvalidAigLit || !errors.Is(err, ErrDimension) {
		t.Fatal(literal, err)
	}
	if literal, err := ParseBooleanFormula(&AigManager{}, "a", names, BooleanFormulaParseOptions{}); literal != InvalidAigLit || !errors.Is(err, ErrInvariant) {
		t.Fatal(literal, err)
	}
}

func TestParseBooleanFormulaDuplicateNamesAndScale(t *testing.T) {
	manager := parseEquationManager(t)
	literal, err := ParseBooleanFormula(manager, "a", []string{"a", "a", "c", "aa"}, BooleanFormulaParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		if got := aigTestLiteralValue(t, manager, literal, inputs); got != inputs[0] {
			t.Fatal(assignment, got)
		}
	}
	deep := strings.Repeat("(", 10_000) + "a" + strings.Repeat(")", 10_000)
	manager = parseEquationManager(t)
	if literal, err = ParseBooleanFormula(manager, deep, []string{"a"}, BooleanFormulaParseOptions{MaxTokens: 25_000, MaxDepth: 10_001}); err != nil || literal.ID() != manager.cis[0] {
		t.Fatal(literal, err)
	}
	flat := strings.Repeat("a+", 4_999) + "a"
	manager = parseEquationManager(t)
	if literal, err = ParseBooleanFormula(manager, flat, []string{"a"}, BooleanFormulaParseOptions{MaxTokens: 10_001}); err != nil || literal.ID() != manager.cis[0] {
		t.Fatal(literal, err)
	}
}

func parseEquationLeaf(value byte) (string, uint16) {
	switch value % 6 {
	case 0:
		return "0", 0x0000
	case 1:
		return "1", 0xffff
	case 2:
		return "a", 0xaaaa
	case 3:
		return "b", 0xcccc
	case 4:
		return "c", 0xf0f0
	default:
		return "aa", 0xff00
	}
}

func FuzzParseBooleanFormula(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6})
	f.Add([]byte("a!b+(c^aa)"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		names := []string{"a", "b", "c", "aa"}
		rawManager := parseEquationManager(t)
		before := rawManager.Clone()
		rawLiteral, rawErr := ParseBooleanFormula(rawManager, string(data), names, BooleanFormulaParseOptions{MaxTokens: 256, MaxDepth: 128})
		if rawErr != nil {
			if rawLiteral != InvalidAigLit || !reflect.DeepEqual(rawManager, before) {
				t.Fatal("non-atomic raw rejection", rawLiteral, rawErr)
			}
		} else if !rawManager.validSignal(rawLiteral) || rawManager.Check() != nil {
			t.Fatal("invalid accepted graph", rawLiteral)
		}

		formula, truth := parseEquationLeaf(0)
		for index, value := range data {
			leaf, leafTruth := parseEquationLeaf(value >> 3)
			switch value % 4 {
			case 0:
				formula = "!(" + formula + ")"
				truth = ^truth
			case 1:
				formula = "(" + formula + ")*(" + leaf + ")"
				truth &= leafTruth
			case 2:
				formula = "(" + formula + ")^(" + leaf + ")"
				truth ^= leafTruth
			case 3:
				if index&1 == 0 {
					formula = "(" + formula + ")!(" + leaf + ")"
					truth &= ^leafTruth
				} else {
					formula = "(" + formula + ")+(" + leaf + ")"
					truth |= leafTruth
				}
			}
		}
		manager := parseEquationManager(t)
		literal, err := ParseBooleanFormula(manager, formula, names, BooleanFormulaParseOptions{MaxTokens: 1 << 14, MaxDepth: 1 << 12})
		if err != nil {
			t.Fatal(formula, err)
		}
		if got := uint16(parseEquationTruth(t, manager, literal)); got != truth {
			t.Fatal(formula, fmt.Sprintf("%04x != %04x", got, truth))
		}
	})
}
