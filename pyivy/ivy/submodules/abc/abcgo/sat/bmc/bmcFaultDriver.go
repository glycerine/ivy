package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"strings"
	"time"
)

// Gia_ManFaultTest iteratively constructs tests and optionally checks faults
// that survive the resulting suite.
func Gia_ManFaultTest(p, golden *gia.Gia_Man_t, opts *Bmc_ParFf_t) {
	if p == nil || golden == nil || opts == nil {
		return
	}
	var nvars, nparams int32
	if opts.Algo == 0 && Gia_FormStrCount(opts.PFormStr, &nvars, &nparams) != 0 {
		return
	}
	if opts.Algo < 0 || opts.Algo > 4 {
		fmt.Printf("Unrecognized algorithm (%d).\n", opts.Algo)
		return
	}
	start := time.Now()
	nfunc := gia.Gia_ManCiNum(p)
	if opts.Algo == 1 {
		nfunc = gia.Gia_ManRegNum(p) + 2*gia.Gia_ManPiNum(p)
	}
	if nfunc <= 0 {
		fmt.Println("Fault test requires input variables.")
		return
	}
	tests := vec.Vec_IntAlloc(10000)
	if opts.PFileName != "" {
		tests = Gia_ManGetTestPatterns(opts.PFileName)
		if tests == nil {
			return
		}
	}
	faultMap := vec.Vec_IntAlloc(0)
	switch opts.Algo {
	case 2:
		vec.Vec_IntFill(faultMap, int(2*gia.Gia_ManAndNum(p)), 1)
	case 3:
		vec.Vec_IntFill(faultMap, int(gia.Gia_ManAndNum(p)), 1)
	case 4:
		vec.Vec_IntFill(faultMap, int(4*gia.Gia_ManAndNum(p)), 1)
	}
	scratch := vec.Vec_IntAlloc(int(gia.Gia_ManCoNum(p)))
	var m *gia.Gia_Man_t
	var data *cnf.Cnf_Dat_t
	var solver *bsat.Solver
	prepared := Gia_ManFaultPrepare(p, golden, opts, nfunc, faultMap, tests, scratch, &m, &data, &solver, 1)
	if prepared == 0 && (m == nil || data == nil || solver == nil) {
		return
	}
	for iteration := int32(vec.Vec_IntSize(tests)) / nfunc; prepared != 0 && iteration < 1000000; iteration++ {
		if iteration > 0 && opts.NIterCheck > 0 && iteration%opts.NIterCheck == 0 && vec.Vec_IntSize(faultMap) > 0 {
			params := vec.Vec_IntAlloc(int(gia.Gia_ManPiNum(m) - nfunc))
			for i := nfunc; i < gia.Gia_ManPiNum(m); i++ {
				id := gia.Gia_ObjId(m, gia.Gia_ManCi(m, i))
				vec.Vec_IntPush(params, data.VarNums[id])
			}
			Gia_ManFaultAnalyze(solver, params, faultMap, scratch, iteration)
			if Gia_ManFaultPrepare(p, golden, opts, nfunc, faultMap, tests, scratch, &m, &data, &solver, 0) == 0 {
				break
			}
		}
		status := solver.Solve()
		if status != bsat.Satisfiable {
			fmt.Printf("Fault test stopped after %d iterations (status %v).\n", iteration, status)
			break
		}
		vec.Vec_IntClear(scratch)
		for i := int32(0); i < nfunc; i++ {
			id := gia.Gia_ObjId(m, gia.Gia_ManCi(m, i))
			value, ok := solver.Value(data.VarNums[id])
			if !ok {
				panic("bmc: missing fault-test model")
			}
			if value {
				vec.Vec_IntPush(scratch, 1)
			} else {
				vec.Vec_IntPush(scratch, 0)
			}
		}
		if opts.FDumpNewFaults != 0 {
			if vec.Vec_IntSize(tests) > 0 {
				Gia_ManFaultDumpNewFaults(m, nfunc, tests, scratch, opts)
			}
			return
		}
		for _, v := range vec.Vec_IntArray(scratch) {
			vec.Vec_IntPush(tests, v)
		}
		if Gia_ManFaultAddOne(m, data, solver, scratch, nfunc, 0, m) == 0 {
			break
		}
	}
	count := int32(vec.Vec_IntSize(tests)) / nfunc
	fmt.Printf("Testing runtime: %s; %d test patterns.\n", time.Since(start), count)
	if opts.FDump != 0 && count > 0 {
		var err error
		if opts.FDumpDelay != 0 && opts.Algo == 1 {
			err = Gia_ManDumpTestsDelay(tests, count, faultOutputName(p, "_tests.txt", "tests.txt"), p)
		} else {
			err = Gia_ManDumpTests(tests, count, faultOutputName(p, "_tests.txt", "tests.txt"))
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	if count == 0 || (p == golden && opts.FDumpUntest == 0 && opts.FCheckUntest == 0) {
		return
	}
	remaining := bsat.NewSolver()
	for i := 0; i < data.Variables; i++ {
		remaining.NewVariable()
	}
	if opts.NTimeOut > 0 {
		remaining.SetDeadline(time.Now().Add(time.Duration(opts.NTimeOut) * time.Second))
	}
	for _, clause := range data.Clauses {
		if !remaining.AddClause(clause...) {
			return
		}
	}
	var faults []bsat.Literal
	vars := vec.Vec_IntAlloc(int(gia.Gia_ManPiNum(m) - nfunc))
	for i := nfunc; i < gia.Gia_ManPiNum(m); i++ {
		id := gia.Gia_ObjId(m, gia.Gia_ManCi(m, i))
		v := data.VarNums[id]
		faults = append(faults, bsat.Positive(v))
		vec.Vec_IntPush(vars, v)
	}
	if !remaining.AddClause(faults...) {
		return
	}
	if opts.FBasic != 0 {
		Cnf_AddCardinConstr(remaining, vars)
	}
	for i := int32(0); i < gia.Gia_ManCoNum(m); i++ {
		id := gia.Gia_ObjId(m, gia.Gia_ManCo(m, i))
		if !remaining.AddClause(bsat.Negative(data.VarNums[id])) {
			return
		}
	}
	patterns := vec.Vec_IntArray(tests)
	for i := 0; i < len(patterns); i += int(nfunc) {
		pattern := vec.Vec_IntAllocArrayCopy(patterns[i:i+int(nfunc)], int(nfunc))
		if Gia_ManFaultAddOne(m, data, remaining, pattern, nfunc, 0, m) == 0 {
			fmt.Println("There are no remaining faults.")
			return
		}
	}
	status := remaining.Solve()
	fmt.Printf("Remaining fault status: %v.\n", status)
	if status == bsat.Satisfiable && opts.FDumpUntest != 0 {
		Gia_ManDumpUntests(m, data, remaining, nfunc, faultOutputName(p, "_untest.txt", "untest.txt"), opts.FVerbose)
	}
}

// faultOutputName preserves ABC's file naming convention for input designs.
func faultOutputName(p *gia.Gia_Man_t, suffix, fallback string) string {
	_, spec := p.Identity()
	if spec == "" {
		return fallback
	}
	if dot := strings.LastIndexByte(spec, '.'); dot >= 0 {
		spec = spec[:dot]
	}
	return spec + suffix
}
