package abc

// Ordered gate clauses from src/sat/bsat/satSolver3.h.

import "slices"

func satSolver3GateClause(variables []int, negative []bool) ([]SatLit, error) {
	if len(variables) == 0 || len(variables) != len(negative) {
		return nil, ErrDimension
	}
	result := make([]SatLit, len(variables))
	for index, variable := range variables {
		if variable < 0 || uint64(variable) > uint64(MaxLiteralID) {
			return nil, ErrDimension
		}
		literal, err := MakeSatLit(SatVar(variable), negative[index])
		if err != nil {
			return nil, err
		}
		result[index] = literal
	}
	return result, nil
}

func satSolver3GateClauses(variables [][]int, negative [][]bool) ([][]SatLit, error) {
	if len(variables) != len(negative) {
		return nil, ErrDimension
	}
	result := make([][]SatLit, len(variables))
	for index := range variables {
		clause, err := satSolver3GateClause(variables[index], negative[index])
		if err != nil {
			return nil, err
		}
		result[index] = clause
	}
	return result, nil
}

func CloneSatSolver3Clauses(clauses [][]SatLit) [][]SatLit {
	result := make([][]SatLit, len(clauses))
	for index := range clauses {
		result[index] = slices.Clone(clauses[index])
	}
	return result
}

// AddClauses atomically installs an ordered gate encoding.
func (s *SatSolver3) AddClauses(clauses [][]SatLit) (bool, error) {
	if !s.usable() || len(s.pushed) != 0 {
		return false, ErrInvariant
	}
	trial := s.core.cloneState()
	trialStore := s.store.clone()
	for _, clause := range clauses {
		for _, literal := range clause {
			if !literal.Valid() {
				return false, ErrInvalidLiteral
			}
		}
		if trialStore != nil {
			if _, err := trialStore.AddClause(clause); err != nil {
				return false, err
			}
		}
		ok, err := trial.AddClause(clause)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	s.core = &trial
	if trialStore != nil {
		s.store = trialStore
	}
	if s.printClauses {
		s.printedClauses = append(s.printedClauses, CloneSatSolver3Clauses(clauses)...)
	}
	return true, nil
}

func SatSolver3ConstantClauses(variable int, complement bool) ([][]SatLit, error) {
	return satSolver3GateClauses([][]int{{variable}}, [][]bool{{complement}})
}

func SatSolver3BufferClauses(variableA, variableB int, complement bool) ([][]SatLit, error) {
	return satSolver3GateClauses(
		[][]int{{variableA, variableB}, {variableA, variableB}},
		[][]bool{{false, !complement}, {true, complement}},
	)
}

func SatSolver3EnabledBufferClauses(variableA, variableB, enable int, complement bool) ([][]SatLit, error) {
	return satSolver3GateClauses(
		[][]int{{variableA, variableB, enable}, {variableA, variableB, enable}},
		[][]bool{{false, !complement, true}, {true, complement, true}},
	)
}

func SatSolver3AndClauses(output, input0, input1 int, complement0, complement1, complementOutput bool) ([][]SatLit, error) {
	return satSolver3GateClauses(
		[][]int{{output, input0}, {output, input1}, {output, input0, input1}},
		[][]bool{{!complementOutput, complement0}, {!complementOutput, complement1}, {complementOutput, !complement0, !complement1}},
	)
}

func SatSolver3XorClauses(output, input0, input1 int, complement bool) ([][]SatLit, error) {
	return satSolver3GateClauses(
		[][]int{{output, input0, input1}, {output, input0, input1}, {output, input0, input1}, {output, input0, input1}},
		[][]bool{{!complement, true, true}, {!complement, false, false}, {complement, true, false}, {complement, false, true}},
	)
}

func SatSolver3MuxClauses(output, control, thenInput, elseInput int, complementControl, complementThen, complementElse, complementOutput bool) ([][]SatLit, error) {
	variables := [][]int{
		{control, thenInput, output},
		{control, thenInput, output},
		{control, elseInput, output},
		{control, elseInput, output},
	}
	negative := [][]bool{
		{!complementControl, !complementThen, false},
		{!complementControl, complementThen, !complementOutput},
		{complementControl, !complementElse, complementOutput},
		{complementControl, complementElse, !complementOutput},
	}
	if thenInput != elseInput {
		variables = append(variables, []int{thenInput, elseInput, output}, []int{thenInput, elseInput, output})
		negative = append(negative,
			[]bool{complementThen, complementElse, !complementOutput},
			[]bool{!complementThen, !complementElse, complementOutput},
		)
	}
	return satSolver3GateClauses(variables, negative)
}

func SatSolver3Mux4Clauses(output, control0, control1, data0, data1, data2, data3 int) ([][]SatLit, error) {
	data := []int{data0, data1, data2, data3}
	variables := make([][]int, 0, 8)
	negative := make([][]bool, 0, 8)
	selectors := [][2]bool{{false, false}, {true, false}, {false, true}, {true, true}}
	for index := 0; index < 4; index++ {
		variables = append(variables, []int{data[index], control0, control1, output})
		negative = append(negative, []bool{true, selectors[index][0], selectors[index][1], false})
	}
	for index := 0; index < 4; index++ {
		variables = append(variables, []int{data[index], control0, control1, output})
		negative = append(negative, []bool{false, selectors[index][0], selectors[index][1], true})
	}
	return satSolver3GateClauses(variables, negative)
}

func SatSolver3XorAndClauses(output, inputA, inputB, inputC int) ([][]SatLit, error) {
	return satSolver3GateClauses(
		[][]int{{output, inputA, inputB}, {output, inputA, inputB}, {output, inputC}, {output, inputA, inputB, inputC}, {output, inputA, inputB, inputC}},
		[][]bool{{true, true, true}, {true, false, false}, {true, false}, {false, true, false, true}, {false, false, true, true}},
	)
}

func SatSolver3ConstraintClauses(variable, auxiliary int, complement bool) ([][]SatLit, error) {
	return satSolver3GateClauses(
		[][]int{{variable, auxiliary}, {variable, auxiliary}},
		[][]bool{{complement, false}, {complement, true}},
	)
}

func SatSolver3HalfSorterClauses(outputA, outputB, input0, input1 int) ([][]SatLit, error) {
	return satSolver3GateClauses(
		[][]int{{outputA, input0}, {outputA, input1}, {outputB, input0, input1}},
		[][]bool{{false, true}, {false, true}, {false, true, true}},
	)
}
