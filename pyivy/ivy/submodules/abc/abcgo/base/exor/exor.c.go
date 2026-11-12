package exor

import (
	"fmt"
	"time"
)

func GetQCost(vars, negative int) int {
	if vars < 0 || negative < 0 || negative > vars {
		panic("exor: invalid quantum cost dimensions")
	}
	switch vars {
	case 0:
		return 1
	case 1:
		if negative == 0 {
			return 1
		}
		return 2
	case 2:
		if negative <= 1 {
			return 5
		}
		return 6
	case 3:
		if negative <= 1 {
			return 14
		}
		return 12 + 2*negative
	}
	return 20 + 12*(vars-4) + 2*max(0, negative-vars/2)
}
func GetQCostTest() {
	for v := 0; v < 10; v++ {
		for n := 0; n <= v; n++ {
			fmt.Printf("%4d ", GetQCost(v, n))
		}
		fmt.Println()
	}
}
func ComputeQCost(literals []int32) int {
	neg := 0
	for _, lit := range literals {
		if lit < 0 {
			panic("exor: output marker in literal vector")
		}
		neg += int(lit & 1)
	}
	return GetQCost(len(literals), neg)
}
func ComputeQCostBits(c *Cube) int {
	vars, neg := 0, 0
	for _, w := range c.Inputs {
		for w != 0 {
			switch w & 3 {
			case 1:
				vars++
				neg++
			case 2:
				vars++
			}
			w >>= 2
		}
	}
	return GetQCost(vars, neg)
}
func ToffoliGateCount(controls, lines int) int {
	if controls < 0 || lines < 0 {
		panic("exor: negative gate dimensions")
	}
	switch controls {
	case 0, 1:
		return 0
	case 2:
		return 1
	case 3:
		return 4
	case 4:
		if (lines+1)/2 >= controls {
			return 8
		}
		return 10
	}
	if (lines+1)/2 >= controls {
		return 4 * (controls - 2)
	}
	return 8 * (controls - 3)
}
func ComputeQCostTcount(m *Manager, literals []int32) int {
	return 7 * ToffoliGateCount(len(literals), m.Inputs+1)
}
func ComputeQCostTcountBits(m *Manager, c *Cube) int {
	return 7 * ToffoliGateCount(literalCount(c), m.Inputs+1)
}
func CountLiterals(m *Manager) int {
	n := 0
	for _, c := range m.Cubes {
		n += literalCount(c)
	}
	return n
}
func CountLiteralsCheck(m *Manager) int { return CountLiterals(m) }
func CountQCost(m *Manager) int {
	n := 0
	for _, c := range m.Cubes {
		n += ComputeQCostBits(c)
	}
	return n
}
func ReduceEsopCover(m *Manager) int {
	before := len(m.Cubes)
	m.Normalize()
	for {
		gain := 0
		for distance := 2; distance <= 4; distance++ {
			gain += m.applyDistance(distance)
		}
		if gain == 0 {
			if m.Options.Quality > 0 && m.escapePlateau(m.Options.Quality) {
				continue
			}
			break
		}
	}
	m.reindex()
	return before - len(m.Cubes)
}

// AddCubesToStartingCover accepts ABC literal rows terminated by -(output+1).
// It copies rows instead of destructively popping the caller's output marker.
func AddCubesToStartingCover(m *Manager, rows [][]int32) error {
	for _, row := range rows {
		if len(row) == 0 || row[len(row)-1] >= 0 {
			return fmt.Errorf("exor: cube missing negative output marker")
		}
		output := -int64(row[len(row)-1]) - 1
		if output < 0 || output >= int64(m.Outputs) {
			return fmt.Errorf("exor: output index outside cover")
		}
		values := make([]VarValue, m.Inputs)
		for i := range values {
			values[i] = VAR_ABS
		}
		for _, lit := range row[:len(row)-1] {
			if lit < 0 || int(lit>>1) >= m.Inputs {
				return fmt.Errorf("exor: input literal outside cover")
			}
			v := VAR_POS
			if lit&1 != 0 {
				v = VAR_NEG
			}
			index := int(lit >> 1)
			if values[index] != VAR_ABS && values[index] != v {
				return fmt.Errorf("exor: contradictory cube literals")
			}
			values[index] = v
		}
		out := make([]bool, m.Outputs)
		out[output] = true
		if err := m.Add(NewCube(values, out)); err != nil {
			return err
		}
	}
	return nil
}
func (m *Manager) Minimize() error {
	for _, c := range m.Cubes {
		if err := m.validate(c); err != nil {
			return err
		}
	}
	if m.Options.MaxCubes > 0 && len(m.Cubes) > m.Options.MaxCubes {
		return fmt.Errorf("exor: starting cover exceeds cube limit")
	}
	m.Stats = Statistics{CubesBefore: len(m.Cubes), LiteralsBefore: CountLiterals(m), QCostBefore: CountQCost(m)}
	ReduceEsopCover(m)
	return nil
}
func Exorcism(rows [][]int32, nIns, nOuts int, path string) (*Manager, error) {
	return Abc_ExorcismMain(rows, nIns, nOuts, path, 1, 0, 0, false)
}
func Abc_ExorcismMain(rows [][]int32, nIns, nOuts int, path string, quality, verbosity, maxCubes int, useQCost bool) (*Manager, error) {
	m, err := NewManager(nIns, nOuts, Options{quality, verbosity, maxCubes, useQCost})
	if err != nil {
		return nil, err
	}
	if err = AddCubesToStartingCover(m, rows); err != nil {
		return nil, err
	}
	start := time.Now()
	if err = m.Minimize(); err != nil {
		return nil, err
	}
	if path != "" {
		if err = WriteResultIntoFile(m, path); err != nil {
			return nil, err
		}
	}
	if verbosity != 0 {
		fmt.Printf("EXORCISM: %d → %d cubes, %d → %d literals (%s)\n", m.Stats.CubesBefore, m.Stats.CubesAfter, m.Stats.LiteralsBefore, m.Stats.LiteralsAfter, time.Since(start))
	}
	return m, nil
}
