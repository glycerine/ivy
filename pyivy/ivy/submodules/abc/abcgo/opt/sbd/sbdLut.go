package sbd

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func NewLUTStructure(inputs []int32) Sbd_Str_t {
	if len(inputs) > 6 {
		panic("LUT structure exceeds six inputs")
	}
	s := Sbd_Str_t{fLut: 1, nVarIns: int32(len(inputs))}
	copy(s.VarIns[:], inputs)
	return s
}
func NewSelectorStructure(inputs []int32) Sbd_Str_t {
	if len(inputs) < 1 || len(inputs) > SBD_DIV_MAX {
		panic("selector structure requires 1..10 inputs")
	}
	s := Sbd_Str_t{nVarIns: int32(len(inputs))}
	copy(s.VarIns[:], inputs)
	return s
}
func sbdStructureParams(s *Sbd_Str_t) int32 {
	if s.fLut != 0 {
		if s.nVarIns < 0 || s.nVarIns > 6 {
			panic("invalid LUT structure")
		}
		return 1 << s.nVarIns
	}
	if s.nVarIns < 1 || s.nVarIns > SBD_DIV_MAX {
		panic("invalid selector structure")
	}
	return s.nVarIns
}
func Sbd_ProblemCountParams(count int32, structures []Sbd_Str_t) int32 {
	var result int32
	for i := range structures[:count] {
		result += sbdStructureParams(&structures[i])
	}
	return result
}
func Sbd_ProblemAddClauses(solver *bsat.Solver, vars, count int32, mapping []int32, structures []Sbd_Str_t) int32 {
	output, parameter := vars, vars+count
	lit := func(index, sign int32) bsat.Literal {
		if index < 0 || int(index) >= len(mapping) || mapping[index] < 0 {
			panic("invalid structure variable mapping")
		}
		return bsat.Literal(2*mapping[index] + sign)
	}
	for _, s := range structures[:count] {
		size := sbdStructureParams(&s)
		if s.fLut != 0 {
			for m := int32(0); m < size; m++ {
				clause := make([]bsat.Literal, s.nVarIns+2)
				for k := int32(0); k < s.nVarIns; k++ {
					clause[k] = lit(s.VarIns[k], m>>k&1)
				}
				for polarity := int32(0); polarity < 2; polarity++ {
					clause[s.nVarIns] = lit(parameter, polarity)
					clause[s.nVarIns+1] = lit(output, 1-polarity)
					if !solver.AddClause(clause...) {
						return 0
					}
				}
				parameter++
			}
		} else {
			for k := int32(0); k < s.nVarIns; k++ {
				for polarity := int32(0); polarity < 2; polarity++ {
					if !solver.AddClause(lit(parameter, 1), lit(output, polarity), lit(s.VarIns[k], 1-polarity)) {
						return 0
					}
				}
				parameter++
			}
		}
		output++
	}
	return 1
}
func Sbd_ProblemAddClausesInit(solver *bsat.Solver, vars, count int32, mapping []int32, structures []Sbd_Str_t) {
	parameter := vars + count
	for _, s := range structures[:count] {
		size := sbdStructureParams(&s)
		if s.fLut == 0 {
			clause := make([]bsat.Literal, s.nVarIns)
			for k := range clause {
				clause[k] = bsat.Positive(mapping[int(parameter)+k])
			}
			if !solver.AddClause(clause...) {
				panic("inconsistent selector initialization")
			}
			for a := int32(0); a < s.nVarIns; a++ {
				for b := a + 1; b < s.nVarIns; b++ {
					if !solver.AddClause(bsat.Negative(mapping[parameter+a]), bsat.Negative(mapping[parameter+b])) {
						panic("inconsistent selector exclusivity")
					}
				}
			}
		}
		parameter += size
	}
}
func Sbd_ProblemPrintSolution(count int32, structures []Sbd_Str_t, literals *vec.Vec_Int_t) {
	fmt.Println("Solution found:")
	at := 0
	for i, s := range structures[:count] {
		kind := "SEL"
		if s.fLut != 0 {
			kind = "LUT"
		}
		fmt.Printf("%s%d : ", kind, i)
		for m := int32(0); m < sbdStructureParams(&s); m++ {
			fmt.Printf("%d", 1-(vec.Vec_IntEntry(literals, at)&1))
			at++
		}
		fmt.Printf("    {%v}\n", s.VarIns[:s.nVarIns])
	}
	if at != vec.Vec_IntSize(literals) {
		panic("solution parameter count mismatch")
	}
}
func Sbd_ProblemCollectSolution(count int32, structures []Sbd_Str_t, literals *vec.Vec_Int_t) {
	at := 0
	for i := range structures[:count] {
		s := &structures[i]
		s.Res = 0
		selected := 0
		for m := int32(0); m < sbdStructureParams(s); m++ {
			value := vec.Vec_IntEntry(literals, at)&1 == 0
			at++
			if value {
				if s.fLut != 0 {
					s.Res |= uint64(1) << m
				} else {
					s.Res = uint64(s.VarIns[m])
					selected++
				}
			}
		}
		if s.fLut != 0 {
			s.Res = util.Abc_Tt6Stretch(s.Res, s.nVarIns)
		} else if selected != 1 {
			panic("selector solution is not one-hot")
		}
	}
	if at != vec.Vec_IntSize(literals) {
		panic("solution parameter count mismatch")
	}
}
func Sbd_ProblemSolve(network *Network, mirrors *vec.Vec_Int_t, pivot int32, window, objToVar, tfo, roots, divSet *vec.Vec_Int_t, count int32, structures []Sbd_Str_t) int32 {
	cec := Sbd_ManSatSolver(nil, network, mirrors, pivot, window, objToVar, tfo, roots, 1)
	if cec == nil {
		return 0
	}
	qbf := bsat.NewSolver()
	vars := int32(vec.Vec_IntSize(divSet))
	params := Sbd_ProblemCountParams(count, structures)
	if vars > SBD_DIV_MAX || vars+count+params > 256 {
		panic("SBD synthesis structure exceeds bounds")
	}
	cecOutput := int32(vec.Vec_IntSize(window) + vec.Vec_IntSize(tfo) + vec.Vec_IntSize(roots))
	cecParameter := cecOutput + count
	qbfFree := params
	cecVars, qbfVars := make([]int32, vars+count+params), make([]int32, vars+count+params)
	copy(cecVars, vec.Vec_IntArray(divSet))
	for i := int32(0); i < count; i++ {
		cecVars[vars+i] = cecOutput + i
	}
	for i := int32(0); i < params; i++ {
		cecVars[vars+count+i] = cecParameter + i
		qbfVars[vars+count+i] = i
	}
	for i := int32(0); i < vars+count; i++ {
		qbfVars[i] = -1
	}
	if Sbd_ProblemAddClauses(cec, vars, count, cecVars, structures) == 0 {
		return 0
	}
	sbdEnsureVariables(qbf, params)
	Sbd_ProblemAddClausesInit(qbf, vars, count, qbfVars, structures)
	parameters := make([]bsat.Literal, params)
	for i := range parameters {
		parameters[i] = bsat.Negative(cecParameter + int32(i))
	}
	for iteration := 0; iteration < 1<<vars; iteration++ {
		status := cec.Solve(parameters...)
		if status == bsat.Unknown {
			return 0
		}
		if status == bsat.Unsatisfiable {
			break
		}
		parameters = nil
		for i := int32(0); i < vars+count; i++ {
			qbfVars[i] = qbfFree
			qbfFree++
			sbdEnsureVariables(qbf, qbfFree)
		}
		for i, v := range vec.Vec_IntArray(divSet) {
			literal := bsat.Literal(2*qbfVars[i] + 1 - sbdModelValue(cec, v))
			if !qbf.AddClause(literal) {
				return 0
			}
		}
		desired := bsat.Literal(2*qbfVars[vars] + sbdModelValue(cec, cecOutput))
		if !qbf.AddClause(desired) {
			return 0
		}
		if Sbd_ProblemAddClauses(qbf, vars, count, qbfVars, structures) == 0 {
			return 0
		}
		if qbf.Solve() != bsat.Satisfiable {
			return 0
		}
		parameters = make([]bsat.Literal, params)
		for i := range parameters {
			parameters[i] = bsat.Literal(2*(cecParameter+int32(i)) + 1 - sbdModelValue(qbf, int32(i)))
		}
	}
	if len(parameters) == 0 {
		return 0
	}
	literals := vec.Vec_IntAlloc(len(parameters))
	for _, literal := range parameters {
		vec.Vec_IntPush(literals, int32(literal))
	}
	Sbd_ProblemCollectSolution(count, structures, literals)
	return 1
}
