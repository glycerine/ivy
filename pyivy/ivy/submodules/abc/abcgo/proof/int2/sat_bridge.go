package int2

import "github.com/glycerine/abcgo/sat/bsat"

// Keep the source solver name local while using the public Go solver API.
type sat_solver = bsat.Solver

const (
	l_Undef = int32(bsat.Unknown)
	l_True  = int32(bsat.Satisfiable)
	l_False = int32(bsat.Unsatisfiable)
)

func sat_solver_new() *sat_solver { return bsat.NewSolver() }

func sat_solver_from_clauses(variableCount int32, clauses [][]int32) *sat_solver {
	if variableCount < 0 {
		panic("int2: negative CNF variable count")
	}
	solver := sat_solver_new()
	sat_solver_setnvars(solver, variableCount)
	for _, clause := range clauses {
		if sat_solver_addclause(solver, clause) == 0 {
			return nil
		}
	}
	return solver
}

func sat_solver_setnvars(s *sat_solver, count int32) {
	for s.NumVars() < count {
		s.NewVariable()
	}
}

func sat_solver_nvars(s *sat_solver) int32 { return s.NumVars() }

func sat_solver_solve(s *sat_solver, assumptions []int32, conflicts, propagations, globalConflicts, globalPropagations int64) int32 {
	literals := make([]bsat.Literal, len(assumptions))
	for i, literal := range assumptions {
		literals[i] = bsat.Literal(literal)
	}
	return int32(s.SolveWithResourceLimits(literals, conflicts, propagations, globalConflicts, globalPropagations))
}

func sat_solver_addclause(s *sat_solver, literals []int32) int32 {
	maxVariable := int32(-1)
	for _, literal := range literals {
		if literal < 0 {
			panic("int2: negative SAT literal")
		}
		if variable := literal >> 1; variable > maxVariable {
			maxVariable = variable
		}
	}
	if maxVariable >= 0 {
		sat_solver_setnvars(s, maxVariable+1)
	}
	clause := make([]bsat.Literal, len(literals))
	for i, literal := range literals {
		clause[i] = bsat.Literal(literal)
	}
	if s.AddClause(clause...) {
		return 1
	}
	return 0
}

func sat_solver_add_buffer(s *sat_solver, output, input, complement int32) int32 {
	if output < 0 || input < 0 || (complement != 0 && complement != 1) {
		panic("int2: invalid buffer gate")
	}
	if sat_solver_addclause(s, []int32{sat_literal(output, 0), sat_literal(input, 1^complement)}) == 0 {
		return 0
	}
	if sat_solver_addclause(s, []int32{sat_literal(output, 1), sat_literal(input, complement)}) == 0 {
		return 0
	}
	return 2
}

func sat_solver_add_and(s *sat_solver, output, input0, input1, complement0, complement1, complement int32) int32 {
	if output < 0 || input0 < 0 || input1 < 0 {
		panic("int2: negative SAT gate variable")
	}
	if sat_solver_addclause(s, []int32{sat_literal(output, 1^complement), sat_literal(input0, complement0)}) == 0 {
		return 0
	}
	if sat_solver_addclause(s, []int32{sat_literal(output, 1^complement), sat_literal(input1, complement1)}) == 0 {
		return 0
	}
	if sat_solver_addclause(s, []int32{
		sat_literal(output, complement),
		sat_literal(input0, 1^complement0),
		sat_literal(input1, 1^complement1),
	}) == 0 {
		return 0
	}
	return 3
}

func sat_solver_var_value(s *sat_solver, variable int32) int32 {
	value, ok := s.Value(variable)
	if ok && value {
		return 1
	}
	return 0
}

func sat_solver_final(s *sat_solver, output *[]int32) int32 {
	core := s.UnsatCore()
	*output = make([]int32, len(core))
	for i, literal := range core {
		(*output)[i] = int32(literal)
	}
	return int32(len(core))
}

func sat_literal(variable, complement int32) int32 {
	if variable < 0 || (complement != 0 && complement != 1) {
		panic("int2: invalid SAT literal")
	}
	return variable*2 + complement
}
