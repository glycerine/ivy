package abc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// MVTerm is a finite value set, a wildcard, or equality to another input.
// EqualInput is -1 when unused; parser/formatter preserve that distinction.
type MVTerm struct {
	Any        bool
	EqualInput int
	Values     []int
}
type MVCube struct {
	Inputs []MVTerm
	Output MVTerm
}

// MVTable stores normalized numeric BLIF-MV covers. DomainSizes lists inputs
// followed by output; Default is nil when uncovered assignments are undefined.
type MVTable struct {
	DomainSizes []int
	Default     *MVTerm
	Cubes       []MVCube
}

func parseMVTerm(text string, domain int, domains []int) (MVTerm, error) {
	term := MVTerm{EqualInput: -1}
	if text == "-" {
		term.Any = true
		return term, nil
	}
	if strings.HasPrefix(text, "=") {
		index, err := strconv.Atoi(text[1:])
		if err != nil || index < 0 || index >= len(domains)-1 || domains[index] != domain {
			return term, fmt.Errorf("invalid equality literal %q", text)
		}
		term.EqualInput = index
		return term, nil
	}
	complement := strings.HasPrefix(text, "!")
	if complement {
		text = text[1:]
	}
	if len(text) >= 2 && ((text[0] == '(' && text[len(text)-1] == ')') || (text[0] == '{' && text[len(text)-1] == '}')) {
		text = text[1 : len(text)-1]
	}
	seen := map[int]bool{}
	for _, part := range strings.Split(text, ",") {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value < 0 || value >= domain {
			return term, fmt.Errorf("value %q outside domain %d", part, domain)
		}
		seen[value] = true
	}
	for value := 0; value < domain; value++ {
		if seen[value] != complement {
			term.Values = append(term.Values, value)
		}
	}
	if len(term.Values) == 0 {
		return term, fmt.Errorf("empty value set")
	}
	return term, nil
}
func mvFields(line string) ([]string, error) {
	var fields []string
	start := -1
	depth := 0
	for i, c := range line {
		switch c {
		case '(', '{':
			depth++
		case ')', '}':
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("unbalanced value set")
			}
		}
		if (c == ' ' || c == '\t' || c == '\r') && depth == 0 {
			if start >= 0 {
				fields = append(fields, line[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("unbalanced value set")
	}
	if start >= 0 {
		fields = append(fields, line[start:])
	}
	return fields, nil
}
func ParseMVTable(text string, domains []int) (*MVTable, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("missing output domain")
	}
	for _, size := range domains {
		if size < 2 {
			return nil, fmt.Errorf("domain must have at least two values")
		}
	}
	table := &MVTable{DomainSizes: append([]int(nil), domains...)}
	for lineIndex, line := range strings.Split(strings.TrimRight(text, "\x00"), "\n") {
		if at := strings.IndexByte(line, '#'); at >= 0 {
			line = line[:at]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "d") {
			if table.Default != nil || len(table.Cubes) > 0 {
				return nil, fmt.Errorf("line %d: default must occur once before cubes", lineIndex+1)
			}
			value := strings.TrimSpace(line[1:])
			if value == "-" {
				value = "0"
			}
			term, err := parseMVTerm(value, domains[len(domains)-1], domains)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineIndex+1, err)
			}
			table.Default = &term
			continue
		}
		fields, err := mvFields(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineIndex+1, err)
		}
		if len(fields) != len(domains) {
			return nil, fmt.Errorf("line %d: expected %d literals, got %d", lineIndex+1, len(domains), len(fields))
		}
		cube := MVCube{}
		for i, field := range fields {
			if len(domains) == 1 && strings.HasSuffix(field, "-") {
				field = "-"
			}
			term, err := parseMVTerm(field, domains[i], domains)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineIndex+1, err)
			}
			if i == len(domains)-1 {
				cube.Output = term
			} else {
				if term.EqualInput == i {
					return nil, fmt.Errorf("line %d: equality refers to itself", lineIndex+1)
				}
				cube.Inputs = append(cube.Inputs, term)
			}
		}
		table.Cubes = append(table.Cubes, cube)
	}
	if len(table.Cubes) == 0 && table.Default == nil {
		return nil, fmt.Errorf("empty cover")
	}
	return table, nil
}
func termString(term MVTerm) string {
	if term.Any {
		return "-"
	}
	if term.EqualInput >= 0 {
		return fmt.Sprintf("=%d", term.EqualInput)
	}
	if len(term.Values) == 1 {
		return strconv.Itoa(term.Values[0])
	}
	parts := make([]string, len(term.Values))
	for i, v := range term.Values {
		parts[i] = strconv.Itoa(v)
	}
	return "(" + strings.Join(parts, ",") + ")"
}
func (t *MVTable) String() string {
	var out strings.Builder
	if t.Default != nil {
		out.WriteString("d" + termString(*t.Default) + "\n")
	}
	for _, cube := range t.Cubes {
		for _, input := range cube.Inputs {
			out.WriteString(termString(input) + " ")
		}
		out.WriteString(termString(cube.Output) + "\n")
	}
	return out.String()
}
func termMatches(term MVTerm, value int, inputs []int) bool {
	if term.Any {
		return true
	}
	if term.EqualInput >= 0 {
		return value == inputs[term.EqualInput]
	}
	for _, v := range term.Values {
		if v == value {
			return true
		}
	}
	return false
}
func (t *MVTable) Evaluate(inputs []int) ([]int, error) {
	if len(inputs)+1 != len(t.DomainSizes) {
		return nil, fmt.Errorf("input count mismatch")
	}
	for i, v := range inputs {
		if v < 0 || v >= t.DomainSizes[i] {
			return nil, fmt.Errorf("input %d outside domain", i)
		}
	}
	values := map[int]bool{}
	outputSize := t.DomainSizes[len(inputs)]
	for _, cube := range t.Cubes {
		matches := true
		for i, input := range cube.Inputs {
			if !termMatches(input, inputs[i], inputs) {
				matches = false
				break
			}
		}
		if matches {
			for value := 0; value < outputSize; value++ {
				if termMatches(cube.Output, value, inputs) {
					values[value] = true
				}
			}
		}
	}
	if t.Default != nil {
		term := *t.Default
		if term.EqualInput < 0 && !term.Any && len(term.Values) == 1 {
			def := term.Values[0]
			other := false
			for value := range values {
				if value != def {
					other = true
				}
			}
			if !other {
				values[def] = true
			} else {
				delete(values, def)
			}
		} else if len(values) == 0 {
			for value := 0; value < outputSize; value++ {
				if termMatches(term, value, inputs) {
					values[value] = true
				}
			}
		}
	}
	result := make([]int, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Ints(result)
	return result, nil
}
func (t *MVTable) CloneForNetwork(_ *Network) any {
	copy, err := ParseMVTable(t.String(), t.DomainSizes)
	if err != nil {
		panic(err)
	}
	return copy
}
