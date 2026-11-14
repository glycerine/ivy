package gia

import (
	"bufio"
	"fmt"
	"github.com/glycerine/abcgo/misc/extra"
	"io"
	"os"
)

type qbfFormula struct {
	variables int
	clauses   [][]int
}

func (p *qbfFormula) AddVariable(_ bool) int { id := p.variables; p.variables++; return id }
func (p *qbfFormula) AddClause(clause []int) bool {
	p.clauses = append(p.clauses, append([]int(nil), clause...))
	return true
}
func qbfWriteDIMACS(w io.Writer, g *Gia_Man_t, nPars int32) error {
	if g == nil || nPars < 0 || nPars > Gia_ManPiNum(g) {
		return fmt.Errorf("gia: invalid QBF parameter count")
	}
	n := int(Gia_ManPiNum(g))
	formula := &qbfFormula{variables: n}
	inputs := make([]int, n)
	for i := range inputs {
		inputs[i] = 2 * i
	}
	output := qbfEncodeGraph(formula, g, inputs)
	formula.AddClause([]int{output})
	bw := bufio.NewWriter(w)
	if _, err := fmt.Fprintf(bw, "p cnf %d %d\n", formula.variables, len(formula.clauses)); err != nil {
		return err
	}
	if nPars > 0 {
		if _, err := fmt.Fprint(bw, "a "); err != nil {
			return err
		}
		for i := int32(0); i < nPars; i++ {
			if _, err := fmt.Fprintf(bw, "%d ", i+1); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(bw, "0"); err != nil {
			return err
		}
	}
	if int(nPars) < formula.variables {
		if _, err := fmt.Fprint(bw, "e "); err != nil {
			return err
		}
		for i := int(nPars); i < formula.variables; i++ {
			if _, err := fmt.Fprintf(bw, "%d ", i+1); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(bw, "0"); err != nil {
			return err
		}
	}
	for _, clause := range formula.clauses {
		for _, lit := range clause {
			value := lit/2 + 1
			if lit&1 != 0 {
				value = -value
			}
			if _, err := fmt.Fprintf(bw, "%d ", value); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(bw, "0"); err != nil {
			return err
		}
	}
	return bw.Flush()
}
func Gia_QbfDumpFile(g *Gia_Man_t, nPars int32) {
	name := extra.Extra_FileNameGenericAppend(g.PSpec, ".qdimacs")
	file, err := os.Create(name)
	if err != nil {
		fmt.Printf("Cannot write QBF file %q: %v\n", name, err)
		return
	}
	err = qbfWriteDIMACS(file, g, nPars)
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Printf("Cannot write QBF file %q: %v\n", name, err)
		return
	}
	fmt.Printf("The 2QBF formula was written into file %q.\n", name)
}
