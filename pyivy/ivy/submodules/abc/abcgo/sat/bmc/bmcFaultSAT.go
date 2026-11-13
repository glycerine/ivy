package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"os"
	"time"
)

// Gia_ManFaultAddOne adds a cofactored graph and connects its fault inputs to
// an existing encoding. addOr selects some output true instead of all false.
func Gia_ManFaultAddOne(m *gia.Gia_Man_t, base *cnf.Cnf_Dat_t, solver *bsat.Solver, values *vec.Vec_Int_t, nFuncVars, addOr int32, baseGraph *gia.Gia_Man_t) int32 {
	return faultAddOne(m, base, solver, values, nFuncVars, addOr, baseGraph, nil)
}

func faultAddOne(m *gia.Gia_Man_t, base *cnf.Cnf_Dat_t, solver *bsat.Solver, values *vec.Vec_Int_t, nFuncVars, addOr int32, baseGraph *gia.Gia_Man_t, capture *cnf.Cnf_Dat_t) int32 {
	if m == nil || solver == nil || nFuncVars < 0 {
		panic("bmc: invalid fault SAT arguments")
	}
	copyGraph := Gia_ManFaultCofactor(m, values)
	enc, err := gia.Mf_ManGenerateCnfResult(copyGraph, 6, true, false, false)
	if err != nil {
		panic(err)
	}
	offset := solver.NumVars()
	if capture != nil {
		capture.Variables = enc.Variables + int(offset)
		capture.VarNums = append([]int32(nil), enc.VarNums...)
		for i, v := range capture.VarNums {
			if v >= 0 {
				capture.VarNums[i] = v + offset
			}
		}
	}
	for solver.NumVars() < offset+int32(enc.Variables) {
		solver.NewVariable()
	}
	for _, clause := range enc.Clauses {
		lifted := make([]bsat.Literal, len(clause))
		for i, lit := range clause {
			lifted[i] = lit + 2*offset
		}
		if !solver.AddClause(lifted...) {
			return 0
		}
	}
	var outputs []bsat.Literal
	for i := int32(0); i < gia.Gia_ManPoNum(copyGraph); i++ {
		id := gia.Gia_ObjId(copyGraph, gia.Gia_ManCo(copyGraph, i))
		v := enc.VarNums[id] + offset
		if addOr != 0 {
			outputs = append(outputs, bsat.Positive(v))
		} else if !solver.AddClause(bsat.Negative(v)) {
			return 0
		}
	}
	if addOr != 0 && !solver.AddClause(outputs...) {
		return 0
	}
	if baseGraph != nil {
		if base == nil || nFuncVars > gia.Gia_ManPiNum(baseGraph) || gia.Gia_ManPiNum(baseGraph) != gia.Gia_ManPiNum(copyGraph) {
			panic("bmc: invalid base CNF")
		}
		for i := nFuncVars; i < gia.Gia_ManPiNum(baseGraph); i++ {
			oldID := gia.Gia_ObjId(baseGraph, gia.Gia_ManCi(baseGraph, i))
			newID := gia.Gia_ObjId(copyGraph, gia.Gia_ManCi(copyGraph, i))
			if int(oldID) >= len(base.VarNums) || base.VarNums[oldID] < 0 || enc.VarNums[newID] < 0 {
				panic("bmc: missing input CNF variable")
			}
			a, b := base.VarNums[oldID], enc.VarNums[newID]+offset
			if !solver.AddClause(bsat.Negative(a), bsat.Positive(b)) || !solver.AddClause(bsat.Positive(a), bsat.Negative(b)) {
				return 0
			}
		}
	}
	return 1
}

// Gia_ManDumpUntests enumerates remaining fault assignments and writes the
// indices of enabled faults in each nonempty assignment.
func Gia_ManDumpUntests(m *gia.Gia_Man_t, data *cnf.Cnf_Dat_t, solver *bsat.Solver, nFuncVars int32, filename string, verbose int32) int32 {
	if m == nil || data == nil || solver == nil || nFuncVars < 0 || nFuncVars > gia.Gia_ManPiNum(m) {
		panic("bmc: invalid untestable fault arguments")
	}
	f, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Cannot open output file %q: %v\n", filename, err)
		return 0
	}
	defer f.Close()
	count := int32(0)
	for iteration := int32(0); iteration < 10000; iteration++ {
		status := solver.Solve()
		if status == bsat.Unknown {
			fmt.Printf("Timeout reached after dumping %d untestable faults.\n", iteration)
			return count
		}
		if status == bsat.Unsatisfiable {
			return count
		}
		var block []bsat.Literal
		var active []int
		for i := nFuncVars; i < gia.Gia_ManPiNum(m); i++ {
			id := gia.Gia_ObjId(m, gia.Gia_ManCi(m, i))
			if int(id) >= len(data.VarNums) || data.VarNums[id] < 0 {
				panic("bmc: missing fault CNF variable")
			}
			variable := data.VarNums[id]
			value, ok := solver.Value(variable)
			if !ok {
				panic("bmc: missing solver model")
			}
			if value {
				block = append(block, bsat.Negative(variable))
				active = append(active, int(i-nFuncVars))
			} else {
				block = append(block, bsat.Positive(variable))
			}
		}
		if len(active) > 0 {
			count++
			if verbose != 0 {
				fmt.Printf("Untestable fault %4d : ", count)
			}
			for _, i := range active {
				fmt.Fprintf(f, "%d ", i)
				if verbose != 0 {
					fmt.Printf("%d ", i)
				}
			}
			fmt.Fprintln(f)
			if verbose != 0 {
				fmt.Println()
			}
		}
		if !solver.AddClause(block...) {
			return count
		}
	}
	return count
}

// Gia_ManFaultDumpNewFaults enumerates faults found by newTest and missed by
// every earlier test. It writes ABC's traditional newfaults.txt file.
func Gia_ManFaultDumpNewFaults(m *gia.Gia_Man_t, nFuncVars int32, tests, newTest *vec.Vec_Int_t, params *Bmc_ParFf_t) int32 {
	if m == nil || params == nil || nFuncVars <= 0 || vec.Vec_IntSize(newTest) != int(nFuncVars) || vec.Vec_IntSize(tests)%int(nFuncVars) != 0 {
		panic("bmc: invalid fault test patterns")
	}
	baseGraph := Gia_ManFaultCofactor(m, newTest)
	s := bsat.NewSolver()
	s.NewVariable()
	if params.nTimeOut > 0 {
		s.SetDeadline(time.Now().Add(time.Duration(params.nTimeOut) * time.Second))
	}
	base := new(cnf.Cnf_Dat_t)
	if faultAddOne(m, nil, s, newTest, nFuncVars, 1, nil, base) == 0 {
		return 0
	}
	all := vec.Vec_IntArray(tests)
	for start := 0; start < len(all); start += int(nFuncVars) {
		pattern := vec.Vec_IntAllocArrayCopy(all[start:start+int(nFuncVars)], int(nFuncVars))
		if Gia_ManFaultAddOne(m, base, s, pattern, nFuncVars, 0, baseGraph) == 0 {
			return 0
		}
	}
	start := time.Now()
	count := Gia_ManDumpUntests(baseGraph, base, s, nFuncVars, "newfaults.txt", params.fVerbose)
	fmt.Printf("Dumped %d new multiple faults into file %q. Time = %s\n", count, "newfaults.txt", time.Since(start))
	return 1
}
