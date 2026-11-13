// Go port of DIMACS reading in msatRead.c.
package msat

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ParseDIMACS(r io.Reader) (*Msat_Solver_t, error) {
	return parseDIMACS(r, true)
}

func parseDIMACS(r io.Reader, initializeOrder bool) (*Msat_Solver_t, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	var p *Msat_Solver_t
	var clause []int
	declared, count := 0, 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "c") {
			continue
		}
		if strings.HasPrefix(line, "p") {
			if p != nil {
				return nil, fmt.Errorf("duplicate DIMACS header")
			}
			var kind string
			var n int
			if _, err := fmt.Sscanf(line, "p %s %d %d", &kind, &n, &declared); err != nil || kind != "cnf" || n < 0 || declared < 0 {
				return nil, fmt.Errorf("invalid DIMACS header")
			}
			p = Msat_SolverAlloc(n, 1, 1, 1, 1, 0)
			Msat_SolverClean(p, n)
			continue
		}
		if p == nil {
			return nil, fmt.Errorf("missing DIMACS header")
		}
		for _, s := range strings.Fields(line) {
			n, err := strconv.Atoi(s)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				if p.okay {
					if Msat_SolverAddClause(p, &Msat_IntVec_t{data: clause}) == 0 {
						p.okay = false
					}
				}
				clause = nil
				count++
				continue
			}
			negative := n < 0
			if negative {
				n = -n
			}
			if n <= 0 || n > p.nVars {
				return nil, fmt.Errorf("DIMACS variable out of range")
			}
			lit := (n - 1) * 2
			if negative {
				lit++
			}
			clause = append(clause, lit)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if p == nil || len(clause) > 0 || count != declared {
		return nil, fmt.Errorf("incomplete DIMACS input")
	}
	if initializeOrder {
		msatInitializeOrder(p)
	}
	return p, nil
}
func Msat_SolverParseDimacs(r io.Reader, out **Msat_Solver_t, verbose int) int {
	p, err := parseDIMACS(r, false)
	if err != nil {
		return 0
	}
	p.fVerbose = verbose
	*out = p
	return boolInt(p.okay)
}
