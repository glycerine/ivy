package exor

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func (m *Manager) Graph() (*aig.Graph, error) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, m.Inputs)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	outputs := make([]aig.Edge, m.Outputs)
	for i := range outputs {
		outputs[i] = g.Const0()
	}
	for _, c := range m.Cubes {
		if err := m.validate(c); err != nil {
			return nil, err
		}
		term := g.Const1()
		for i, x := range inputs {
			switch GetVar(c, i) {
			case VAR_NEG:
				term = g.AND(term, x.Not())
			case VAR_POS:
				term = g.AND(term, x)
			}
		}
		for i := range outputs {
			if c.Output(i) {
				outputs[i] = g.XOR(outputs[i], term)
			}
		}
	}
	for _, o := range outputs {
		g.AddOutput(o)
	}
	return g, nil
}

// FromGraph computes the exact positive-polarity Reed-Muller expansion by a
// Boolean Möbius transform. The caller bounds exponential truth-table work.
func FromGraph(g *aig.Graph, maxVariables int) (*Manager, error) {
	if err := g.Check(); err != nil {
		return nil, err
	}
	if aig.Aig_ManRegNum(g) != 0 {
		return nil, fmt.Errorf("exor: graph must be combinational")
	}
	n := g.InputCount()
	if maxVariables < 0 || n > maxVariables || n >= 31 {
		return nil, fmt.Errorf("exor: %d inputs exceed truth-table budget %d", n, maxVariables)
	}
	m, err := NewManager(n, g.OutputCount(), Options{})
	if err != nil {
		return nil, err
	}
	count := 1 << uint(n)
	truth := make([][]bool, count)
	input := make([]bool, n)
	for assignment := 0; assignment < count; assignment++ {
		for i := range input {
			input[i] = assignment&(1<<uint(i)) != 0
		}
		truth[assignment], err = g.Evaluate(input)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < n; i++ {
		for mask := 0; mask < count; mask++ {
			if mask&(1<<uint(i)) != 0 {
				for output := 0; output < m.Outputs; output++ {
					truth[mask][output] = truth[mask][output] != truth[mask^(1<<uint(i))][output]
				}
			}
		}
	}
	for mask, outputs := range truth {
		nonzero := false
		for _, v := range outputs {
			nonzero = nonzero || v
		}
		if !nonzero {
			continue
		}
		values := make([]VarValue, n)
		for i := range values {
			values[i] = VAR_ABS
			if mask&(1<<uint(i)) != 0 {
				values[i] = VAR_POS
			}
		}
		m.Add(NewCube(values, outputs))
	}
	return m, nil
}
func Abc_ExorcismNtk2Esop(g *aig.Graph, maxVariables int) ([][]int32, error) {
	m, err := FromGraph(g, maxVariables)
	if err != nil {
		return nil, err
	}
	var rows [][]int32
	for _, c := range m.Cubes {
		var literals []int32
		for i := 0; i < m.Inputs; i++ {
			if GetVar(c, i) == VAR_POS {
				literals = append(literals, int32(i*2))
			}
		}
		for output := 0; output < m.Outputs; output++ {
			if c.Output(output) {
				row := append([]int32(nil), literals...)
				row = append(row, -int32(output+1))
				rows = append(rows, row)
			}
		}
	}
	return rows, nil
}
