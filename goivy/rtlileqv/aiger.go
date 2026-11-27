// Package rtlileqv compares RTLIL output and next-state logic at matched cuts.
package rtlileqv

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
)

type bitName struct {
	name string
	bit  int
}

func (b bitName) String() string { return fmt.Sprintf("%s[%d]", b.name, b.bit) }

type aiger struct {
	inputs, outputs []int
	latches         [][2]int
	gates           [][3]int
	inputNames      [][]bitName
	outputNames     [][]bitName
	latchNames      [][]bitName
	clockEdge       string
}

func readAiger(graph, names io.Reader) (*aiger, error) {
	lines := bufio.NewScanner(graph)
	if !lines.Scan() {
		return nil, fmt.Errorf("missing ASCII AIGER header")
	}
	header := strings.Fields(lines.Text())
	if len(header) < 6 || header[0] != "aag" {
		return nil, fmt.Errorf("invalid ASCII AIGER header %q", lines.Text())
	}
	counts, err := parseNonnegative(header[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid AIGER header: %w", err)
	}
	for _, extra := range counts[5:] {
		if extra != 0 {
			return nil, fmt.Errorf("AIGER property sections are unsupported")
		}
	}
	maxVar := counts[0]
	if maxVar > (math.MaxInt-1)/2 {
		return nil, fmt.Errorf("AIGER variable count is too large")
	}
	maxLiteral := 2*maxVar + 1
	read := func(minFields, maxFields int) ([]int, error) {
		if !lines.Scan() {
			if err := lines.Err(); err != nil {
				return nil, err
			}
			return nil, io.ErrUnexpectedEOF
		}
		fields := strings.Fields(lines.Text())
		if len(fields) < minFields || len(fields) > maxFields {
			return nil, fmt.Errorf("invalid AIGER record %q", lines.Text())
		}
		literals, err := parseNonnegative(fields)
		if err != nil {
			return nil, err
		}
		for _, lit := range literals {
			if lit > maxLiteral {
				return nil, fmt.Errorf("AIGER literal %d exceeds maximum %d", lit, maxLiteral)
			}
		}
		return literals, nil
	}

	var circuit aiger
	defined := map[int]bool{0: true}
	declare := func(lit int) error {
		if lit == 0 || lit%2 != 0 || defined[lit] {
			return fmt.Errorf("invalid or repeated AIGER variable literal %d", lit)
		}
		defined[lit] = true
		return nil
	}
	for range counts[1] {
		fields, err := read(1, 1)
		if err != nil {
			return nil, fmt.Errorf("reading AIGER input: %w", err)
		}
		if err := declare(fields[0]); err != nil {
			return nil, err
		}
		circuit.inputs = append(circuit.inputs, fields[0])
	}
	for range counts[2] {
		fields, err := read(2, 3)
		if err != nil {
			return nil, fmt.Errorf("reading AIGER latch: %w", err)
		}
		if err := declare(fields[0]); err != nil {
			return nil, err
		}
		circuit.latches = append(circuit.latches, [2]int{fields[0], fields[1]})
	}
	for range counts[3] {
		fields, err := read(1, 1)
		if err != nil {
			return nil, fmt.Errorf("reading AIGER output: %w", err)
		}
		circuit.outputs = append(circuit.outputs, fields[0])
	}
	for range counts[4] {
		fields, err := read(3, 3)
		if err != nil {
			return nil, fmt.Errorf("reading AIGER gate: %w", err)
		}
		lhs, left, right := fields[0], fields[1], fields[2]
		if !defined[left&^1] || !defined[right&^1] {
			return nil, fmt.Errorf("AIGER gate %d references an undefined input", lhs)
		}
		if err := declare(lhs); err != nil {
			return nil, err
		}
		circuit.gates = append(circuit.gates, [3]int{lhs, left, right})
	}
	for _, lit := range circuit.outputs {
		if !defined[lit&^1] {
			return nil, fmt.Errorf("AIGER output references undefined literal %d", lit)
		}
	}
	for _, latch := range circuit.latches {
		if !defined[latch[1]&^1] {
			return nil, fmt.Errorf("AIGER latch %d references undefined literal %d", latch[0], latch[1])
		}
	}
	if err := circuit.readNameMap(names); err != nil {
		return nil, err
	}
	circuit.restoreOutputLatchAliases()
	return &circuit, nil
}

func (a *aiger) readNameMap(source io.Reader) error {
	a.inputNames = make([][]bitName, len(a.inputs))
	a.outputNames = make([][]bitName, len(a.outputs))
	a.latchNames = make([][]bitName, len(a.latches))
	byKind := map[string][][]bitName{"input": a.inputNames, "output": a.outputNames, "latch": a.latchNames}
	names := bufio.NewScanner(source)
	for names.Scan() {
		fields := strings.Fields(names.Text())
		if len(fields) == 0 {
			continue
		}
		table, known := byKind[fields[0]]
		if !known {
			continue
		}
		if len(fields) < 4 {
			return fmt.Errorf("invalid AIGER name map entry %q", names.Text())
		}
		values, err := parseNonnegative(fields[1:3])
		if err != nil || values[0] >= len(table) {
			return fmt.Errorf("invalid AIGER name map position %q", names.Text())
		}
		name := strings.TrimPrefix(strings.Join(fields[3:], " "), `\`)
		table[values[0]] = append(table[values[0]], bitName{name, values[1]})
	}
	return names.Err()
}

func (a *aiger) restoreOutputLatchAliases() {
	buffered := make(map[int]int)
	for _, gate := range a.gates {
		if gate[1] == gate[2] {
			input := gate[1]
			if source, ok := buffered[input]; ok {
				input = source
			}
			buffered[gate[0]] = input
		}
	}
	latchPositions := make(map[int]int, len(a.latches))
	for pos, latch := range a.latches {
		latchPositions[latch[0]] = pos
	}
	for i, literal := range a.outputs {
		if input, ok := buffered[literal]; ok {
			literal = input
		}
		if pos, ok := latchPositions[literal]; ok {
			for _, alias := range a.outputNames[i] {
				if !slices.Contains(a.latchNames[pos], alias) {
					a.latchNames[pos] = append(a.latchNames[pos], alias)
				}
			}
		}
	}
}

func parseNonnegative(fields []string) ([]int, error) {
	values := make([]int, len(fields))
	for i, field := range fields {
		value, err := strconv.Atoi(field)
		if err != nil || value < 0 {
			return nil, fmt.Errorf("invalid nonnegative integer %q", field)
		}
		values[i] = value
	}
	return values, nil
}
