// Go port of xsatCnfReader.c by Bruno Schmitt (2016).
package xsat

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ParseDIMACS(r io.Reader) (*Solver, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	var s *Solver
	var clause []int
	declared, count := 0, 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "c") {
			continue
		}
		if strings.HasPrefix(line, "p") {
			if s != nil {
				return nil, fmt.Errorf("duplicate DIMACS header")
			}
			var kind string
			var variables int
			if _, err := fmt.Sscanf(line, "p %s %d %d", &kind, &variables, &declared); err != nil || kind != "cnf" || variables < 0 || declared < 0 {
				return nil, fmt.Errorf("invalid DIMACS header")
			}
			s = New()
			for v := 0; v < variables; v++ {
				s.AddVariable(1)
			}
			continue
		}
		if s == nil {
			return nil, fmt.Errorf("missing DIMACS header")
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
			if n <= 0 || n > len(s.assigns) {
				return nil, fmt.Errorf("DIMACS variable out of range")
			}
			lit := 2 * (n - 1)
			if negative {
				lit++
			}
			clause = append(clause, lit)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if s == nil || len(clause) != 0 || count != declared {
		return nil, fmt.Errorf("incomplete DIMACS input")
	}
	return s, nil
}
func xSAT_SolverParseDimacs(r io.Reader, out **Solver) int {
	s, err := ParseDIMACS(r)
	if err != nil {
		return 0
	}
	*out = s
	return boolInt(s.okay)
}
