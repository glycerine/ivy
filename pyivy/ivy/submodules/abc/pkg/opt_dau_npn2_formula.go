package abc

// Checked Boolean-formula parsing and evaluation from opt/dau/dauNpn2.c.

import (
	"fmt"
	"strconv"
	"strings"
)

type DauLogicFormulaOptions struct {
	MaxBytes, MaxNodes, MaxDepth int
	Stop                         func() bool
}

type dauLogicKind uint8

const (
	dauLogicVariable dauLogicKind = iota
	dauLogicAnd
	dauLogicXor
)

type dauLogicNode struct {
	kind        dauLogicKind
	variable    int
	left, right int
	complement  bool
}

type DauLogicFormula struct {
	nodes     []dauLogicNode
	root      int
	variables int
}

type dauLogicParser struct {
	text    string
	pos     int
	options DauLogicFormulaOptions
	nodes   []dauLogicNode
	maximum int
}

func normalizeDauLogicFormulaOptions(options DauLogicFormulaOptions) (DauLogicFormulaOptions, error) {
	if options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxDepth < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 4096
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 4096
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 256
	}
	return options, nil
}

func ParseDauLogicFormula(text string, options DauLogicFormulaOptions) (*DauLogicFormula, error) {
	o, err := normalizeDauLogicFormulaOptions(options)
	if err != nil {
		return nil, err
	}
	if len(text) == 0 {
		return nil, ErrFormat
	}
	if len(text) > o.MaxBytes {
		return nil, ErrLimit
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	p := dauLogicParser{text: text, options: o, maximum: -1}
	root, err := p.expression(1)
	if err != nil {
		return nil, err
	}
	if p.pos != len(text) {
		return nil, fmt.Errorf("%w at byte %d", ErrFormat, p.pos)
	}
	return &DauLogicFormula{nodes: p.nodes, root: root, variables: p.maximum + 1}, nil
}

func (p *dauLogicParser) stopped() bool { return p.options.Stop != nil && p.options.Stop() }

func (p *dauLogicParser) add(node dauLogicNode) (int, error) {
	if p.stopped() {
		return 0, ErrCancelled
	}
	if len(p.nodes) >= p.options.MaxNodes {
		return 0, ErrLimit
	}
	p.nodes = append(p.nodes, node)
	return len(p.nodes) - 1, nil
}

func (p *dauLogicParser) expression(depth int) (int, error) {
	if depth > p.options.MaxDepth {
		return 0, ErrLimit
	}
	left, err := p.atom(depth)
	if err != nil {
		return 0, err
	}
	if p.pos >= len(p.text) || p.text[p.pos] == ')' {
		return left, nil
	}
	operator := p.text[p.pos]
	if operator != '&' && operator != '^' {
		return 0, fmt.Errorf("%w at byte %d", ErrFormat, p.pos)
	}
	p.pos++
	right, err := p.atom(depth)
	if err != nil {
		return 0, err
	}
	if p.pos < len(p.text) && p.text[p.pos] != ')' {
		return 0, fmt.Errorf("%w at byte %d", ErrFormat, p.pos)
	}
	kind := dauLogicAnd
	if operator == '^' {
		kind = dauLogicXor
	}
	return p.add(dauLogicNode{kind: kind, left: left, right: right})
}

func (p *dauLogicParser) atom(depth int) (int, error) {
	complement := false
	for p.pos < len(p.text) && p.text[p.pos] == '~' {
		complement = !complement
		p.pos++
	}
	if p.pos >= len(p.text) {
		return 0, ErrFormat
	}
	var index int
	var err error
	if p.text[p.pos] == '(' {
		p.pos++
		index, err = p.expression(depth + 1)
		if err != nil {
			return 0, err
		}
		if p.pos >= len(p.text) || p.text[p.pos] != ')' {
			return 0, fmt.Errorf("%w at byte %d", ErrFormat, p.pos)
		}
		p.pos++
	} else {
		variable := int(p.text[p.pos] - 'a')
		if variable < 0 || variable >= 6 {
			return 0, fmt.Errorf("%w at byte %d", ErrFormat, p.pos)
		}
		p.pos++
		index, err = p.add(dauLogicNode{kind: dauLogicVariable, variable: variable})
		if err != nil {
			return 0, err
		}
		p.maximum = max(p.maximum, variable)
	}
	if complement {
		p.nodes[index].complement = !p.nodes[index].complement
	}
	return index, nil
}

func (formula *DauLogicFormula) VariableCount() int {
	if formula == nil {
		return 0
	}
	return formula.variables
}

func (formula *DauLogicFormula) String() string {
	if formula == nil || formula.root < 0 || formula.root >= len(formula.nodes) {
		return ""
	}
	var render func(int) string
	render = func(index int) string {
		node := formula.nodes[index]
		prefix := ""
		if node.complement {
			prefix = "~"
		}
		if node.kind == dauLogicVariable {
			return prefix + string(rune('a'+node.variable))
		}
		operator := "&"
		if node.kind == dauLogicXor {
			operator = "^"
		}
		return prefix + "(" + render(node.left) + operator + render(node.right) + ")"
	}
	return render(formula.root)
}

func (formula *DauLogicFormula) Evaluate(inputs []bool) (bool, error) {
	if formula == nil || formula.root < 0 || formula.root >= len(formula.nodes) || len(inputs) < formula.variables {
		return false, ErrDimension
	}
	values := make([]bool, len(formula.nodes))
	for index, node := range formula.nodes {
		var value bool
		switch node.kind {
		case dauLogicVariable:
			value = inputs[node.variable]
		case dauLogicAnd:
			if node.left < 0 || node.left >= index || node.right < 0 || node.right >= index {
				return false, ErrInvariant
			}
			value = values[node.left] && values[node.right]
		case dauLogicXor:
			if node.left < 0 || node.left >= index || node.right < 0 || node.right >= index {
				return false, ErrInvariant
			}
			value = values[node.left] != values[node.right]
		default:
			return false, ErrInvariant
		}
		values[index] = value != node.complement
	}
	return values[formula.root], nil
}

func (formula *DauLogicFormula) TruthTable(variables, maxWords int, stop func() bool) (*TruthTable, error) {
	if formula == nil || variables < formula.variables || variables > 6 {
		return nil, ErrDimension
	}
	table, err := NewConstantTruthTable(variables, false, maxWords)
	if err != nil {
		return nil, err
	}
	inputs := make([]bool, variables)
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		if stop != nil && stop() {
			return nil, ErrCancelled
		}
		for variable := range inputs {
			inputs[variable] = assignment&(1<<uint(variable)) != 0
		}
		value, err := formula.Evaluate(inputs)
		if err != nil {
			return nil, err
		}
		if value {
			if err = table.Set(uint64(assignment), true); err != nil {
				return nil, err
			}
		}
	}
	return table, nil
}

type DauLogicLibraryEntry struct {
	Expected, Actual uint64
	Formula          string
}

func (entry DauLogicLibraryEntry) Matches() bool { return entry.Expected == entry.Actual }

type DauLogicLibraryReport struct {
	entries []DauLogicLibraryEntry
	fails   int
}

func (report DauLogicLibraryReport) Entries() []DauLogicLibraryEntry {
	return append([]DauLogicLibraryEntry(nil), report.entries...)
}

func (report DauLogicLibraryReport) Failures() int { return report.fails }

func parseDauLogicLibrary(data []byte, variables int, options DauLogicFormulaOptions) ([]DauLogicLibraryEntry, []*DauLogicFormula, error) {
	if variables < 1 || variables > 6 {
		return nil, nil, ErrDimension
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	entries := make([]DauLogicLibraryEntry, 0, len(lines))
	formulas := make([]*DauLogicFormula, 0, len(lines))
	for lineNumber, raw := range lines {
		line := strings.TrimSpace(strings.TrimSuffix(raw, "\r"))
		if line == "" {
			continue
		}
		separator := strings.IndexByte(line, '=')
		if separator < 1 {
			return nil, nil, fmt.Errorf("%w at line %d", ErrFormat, lineNumber+1)
		}
		expected, err := strconv.ParseUint(strings.TrimSpace(line[:separator]), 16, 64)
		if err != nil || expected&^dauTruthMask(variables) != 0 {
			return nil, nil, fmt.Errorf("%w at line %d", ErrFormat, lineNumber+1)
		}
		text := strings.TrimSpace(line[separator+1:])
		formula, err := ParseDauLogicFormula(text, options)
		if err != nil {
			return nil, nil, fmt.Errorf("line %d: %w", lineNumber+1, err)
		}
		if formula.variables > variables {
			return nil, nil, fmt.Errorf("%w at line %d", ErrDimension, lineNumber+1)
		}
		table, err := formula.TruthTable(variables, 0, options.Stop)
		if err != nil {
			return nil, nil, err
		}
		actual := table.words[0] & dauTruthMask(variables)
		entries = append(entries, DauLogicLibraryEntry{Expected: expected, Actual: actual, Formula: formula.String()})
		formulas = append(formulas, formula)
	}
	return entries, formulas, nil
}

func VerifyDauLogicLibrary(data []byte, variables int, options DauLogicFormulaOptions) (DauLogicLibraryReport, error) {
	entries, _, err := parseDauLogicLibrary(data, variables, options)
	if err != nil {
		return DauLogicLibraryReport{}, err
	}
	report := DauLogicLibraryReport{entries: entries}
	for _, entry := range entries {
		if !entry.Matches() {
			report.fails++
		}
	}
	return report, nil
}
