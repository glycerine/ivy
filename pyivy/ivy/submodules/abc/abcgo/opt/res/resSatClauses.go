package res

import "github.com/glycerine/abcgo/sat/bsat"

func resLiteral(variable, complemented int32) bsat.Literal {
	if complemented != 0 {
		return bsat.Negative(variable)
	}
	return bsat.Positive(variable)
}

func resInvertPhase(phase int32) int32 {
	if phase != 0 {
		return 0
	}
	return 1
}

// Res_SatAddConst1 constrains a variable to the requested phase of true.
func Res_SatAddConst1(s *bsat.Solver, variable, complemented int32) int32 {
	if s.AddClause(resLiteral(variable, complemented)) {
		return 1
	}
	return 0
}

// Res_SatAddEqual constrains variable0 to variable1 XOR complemented.
func Res_SatAddEqual(s *bsat.Solver, variable0, variable1, complemented int32) int32 {
	clauses := [][]bsat.Literal{
		{bsat.Positive(variable0), resLiteral(variable1, resInvertPhase(complemented))},
		{bsat.Negative(variable0), resLiteral(variable1, complemented)},
	}
	for _, clause := range clauses {
		if !s.AddClause(clause...) {
			return 0
		}
	}
	return 1
}

// Res_SatAddAnd constrains output to the conjunction of the phased inputs.
func Res_SatAddAnd(s *bsat.Solver, output, input0, input1, phase0, phase1 int32) int32 {
	clauses := [][]bsat.Literal{
		{bsat.Negative(output), resLiteral(input0, phase0)},
		{bsat.Negative(output), resLiteral(input1, phase1)},
		{bsat.Positive(output), resLiteral(input0, resInvertPhase(phase0)), resLiteral(input1, resInvertPhase(phase1))},
	}
	for _, clause := range clauses {
		if !s.AddClause(clause...) {
			return 0
		}
	}
	return 1
}
