package io

import (
	"fmt"
	stdio "io"
	"strings"

	"github.com/glycerine/abcgo/aig/ioa"
)

// AigerCNF returns a definitional encoding with the sole output asserted.
// The three clauses of each AND are its complete set of prime implicates.
func AigerCNF(model *ioa.Model) (int, [][]int, error) {
	if err := model.Check(); err != nil {
		return 0, nil, err
	}
	if len(model.Outputs) != 1 || len(model.Latches) != 0 {
		return 0, nil, fmt.Errorf("CNF requires one combinational miter output")
	}
	if len(model.Bad)+len(model.Constraints)+len(model.Justice)+len(model.Fairness) > 0 {
		return 0, nil, fmt.Errorf("CNF miter must not contain additional verification properties")
	}
	var clauses [][]int
	add := func(literals ...uint32) {
		clause := make([]int, 0, len(literals))
		seen := map[int]bool{}
		for _, literal := range literals {
			if literal == 1 {
				return
			}
			if literal == 0 {
				continue
			}
			value := int(literal >> 1)
			if literal&1 != 0 {
				value = -value
			}
			if seen[-value] {
				return
			}
			if !seen[value] {
				seen[value] = true
				clause = append(clause, value)
			}
		}
		clauses = append(clauses, clause)
	}
	for _, gate := range model.ANDs {
		add(gate.Left^1, gate.Right^1, gate.Literal)
		add(gate.Left, gate.Literal^1)
		add(gate.Right, gate.Literal^1)
	}
	add(model.Outputs[0])
	return int(model.MaxVar), clauses, nil
}
func WriteAigerCNF(w stdio.Writer, model *ioa.Model) error {
	variables, clauses, err := AigerCNF(model)
	if err != nil {
		return err
	}
	var out strings.Builder
	out.WriteString("c PI variable numbers: <PI_name> <SAT_var_number>\n")
	for i, literal := range model.Inputs {
		name := model.Symbols[ioa.Symbol{Kind: 'i', Index: uint32(i)}]
		if name == "" {
			name = fmt.Sprintf("pi%d", i)
		}
		name = strings.NewReplacer("\n", " ", "\r", " ").Replace(name)
		fmt.Fprintf(&out, "c %s %d\n", name, literal>>1)
	}
	fmt.Fprintf(&out, "p cnf %d %d\n", variables, len(clauses))
	for _, clause := range clauses {
		for _, literal := range clause {
			fmt.Fprintf(&out, "%d ", literal)
		}
		out.WriteString("0\n")
	}
	return writeText(w, out.String())
}
