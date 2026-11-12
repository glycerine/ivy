// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func ParseDIMACS(r io.Reader) (*Solver, error) {
	s := New()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 16*1024*1024)
	var clause []int
	declaredVars, declaredClauses := -1, -1
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "c") {
			continue
		}
		if strings.HasPrefix(line, "p") {
			if declaredVars >= 0 {
				return nil, fmt.Errorf("duplicate DIMACS header")
			}
			var kind string
			if _, err := fmt.Sscanf(line, "p %s %d %d", &kind, &declaredVars, &declaredClauses); err != nil || kind != "cnf" || declaredVars < 0 || declaredClauses < 0 {
				return nil, fmt.Errorf("invalid DIMACS header: %s", line)
			}
			s.SetVariables(declaredVars)
			continue
		}
		for _, field := range strings.Fields(line) {
			n, err := strconv.Atoi(field)
			if err != nil {
				return nil, fmt.Errorf("invalid DIMACS literal %q", field)
			}
			if n == 0 {
				s.AddClause(clause)
				clause = nil
				count++
				continue
			}
			negative := n < 0
			if negative {
				n = -n
			}
			if n <= 0 || declaredVars >= 0 && n > declaredVars {
				return nil, fmt.Errorf("DIMACS variable out of range")
			}
			clause = append(clause, Literal(n-1, negative))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(clause) != 0 {
		return nil, fmt.Errorf("unterminated DIMACS clause")
	}
	if declaredClauses >= 0 && count != declaredClauses {
		return nil, fmt.Errorf("DIMACS clause count: got %d want %d", count, declaredClauses)
	}
	return s, nil
}
func satoko_parse_dimacs(path string, out **Solver) int {
	f, err := os.Open(path)
	if err != nil {
		return SATOKO_ERR
	}
	defer f.Close()
	s, err := ParseDIMACS(f)
	if err != nil {
		return SATOKO_ERR
	}
	*out = s
	return SATOKO_OK
}
