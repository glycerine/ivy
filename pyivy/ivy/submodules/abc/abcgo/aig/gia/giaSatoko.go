package gia

// Original Gia Satoko drivers, with source MF eight-cut CNF construction.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia/mf"
	"github.com/glycerine/abcgo/aig/gia/satcost"
	"github.com/glycerine/abcgo/sat/satoko"
	"os"
	"time"
)

type satoko_opts_t = satoko.Options
type satoko_t = satoko.Solver

func Gia_ManCollectVars_rec(variable int32, mapping, result *Vec_Int_t, visited *Vec_Bit_t) {
	if Vec_BitEntry(visited, int(variable)) != 0 {
		return
	}
	Vec_BitWriteEntry(visited, int(variable), 1)
	offset := Vec_IntEntry(mapping, int(variable))
	if offset != 0 {
		count := Vec_IntEntry(mapping, int(offset))
		for i := int32(1); i <= count; i++ {
			Gia_ManCollectVars_rec(Vec_IntEntry(mapping, int(offset+i)), mapping, result, visited)
		}
	}
	Vec_IntPush(result, variable)
}
func Gia_ManCollectVars(root int32, mapping *Vec_Int_t, variables int32) *Vec_Int_t {
	if Vec_IntEntry(mapping, int(root)) == 0 {
		panic("gia: cone root has no mapping")
	}
	result := Vec_IntAlloc(100)
	visited := Vec_BitStart(int(variables))
	Gia_ManCollectVars_rec(root, mapping, result, visited)
	return result
}
func Gia_ManSatokoReport(output, status int32, elapsed time.Duration) {
	if output >= 0 {
		fmt.Printf("Output %6d : ", output)
	} else {
		fmt.Print("Total: ")
	}
	switch int(status) {
	case satoko.SATOKO_UNDEC:
		fmt.Print("UNDECIDED      ")
	case satoko.SATOKO_SAT:
		fmt.Print("SATISFIABLE    ")
	default:
		fmt.Print("UNSATISFIABLE  ")
	}
	fmt.Printf("Time = %.2f sec\n", elapsed.Seconds())
}
func giaSatokoOptions(options *satoko.Options) satoko.Options {
	if options == nil {
		return satoko.DefaultOptions()
	}
	return *options
}
func Gia_ManSatokoFromDimacs(path string, options *satoko_opts_t) *satoko_t {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	defer file.Close()
	solver, err := satoko.ParseDIMACS(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	*solver.Options() = giaSatokoOptions(options)
	return solver
}
func Gia_ManSatokoDimacs(path string, options *satoko_opts_t) {
	start := time.Now()
	status := satoko.SATOKO_UNSAT
	if solver := Gia_ManSatokoFromDimacs(path, options); solver != nil {
		status = solver.Solve()
	}
	Gia_ManSatokoReport(-1, int32(status), time.Since(start))
}
func Gia_ManSatokoInit(cnf *Cnf_Dat_t, options *satoko_opts_t) *satoko_t {
	if cnf == nil {
		return nil
	}
	solver := satoko.New()
	for _, clause := range cnf.Clauses {
		literals := make([]int, len(clause))
		for i, lit := range clause {
			literals[i] = int(lit)
		}
		if !solver.AddClause(literals) {
			return nil
		}
	}
	*solver.Options() = giaSatokoOptions(options)
	return solver
}
func Gia_ManSatokoCreate(p *Gia_Man_t, options *satoko_opts_t) *satoko_t {
	cnf, err := mf.GenerateCNF(p, 8, false, true, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	solver := Gia_ManSatokoInit(cnf, options)
	if solver == nil || !solver.Simplify() {
		return nil
	}
	return solver
}
func Gia_ManSatokoCallOneResult(p *Gia_Man_t, options *satoko_opts_t) (satcost.Result, error) {
	cnf, err := mf.GenerateCNF(p, 8, false, true, false)
	if err != nil {
		return satcost.Result{}, err
	}
	result, err := satcost.RunCNF(cnf, giaSatokoOptions(options))
	result.Backend = (satcost.MFBackend{}).Name()
	return result, err
}
func Gia_ManSatokoCallOne(p *Gia_Man_t, options *satoko_opts_t, output int32) int32 {
	start := time.Now()
	result, err := Gia_ManSatokoCallOneResult(p, options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return -1
	}
	Gia_ManSatokoReport(output, int32(result.Status), time.Since(start))
	return int32(result.Conflicts)
}
func Gia_ManSatokoCall(p *Gia_Man_t, options *satoko_opts_t, split, incremental int32) {
	start := time.Now()
	if incremental != 0 {
		cnf, err := mf.GenerateCNF(p, 8, false, false, true)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		solver := Gia_ManSatokoInit(cnf, options)
		if solver == nil {
			for i := int32(0); i < Gia_ManCoNum(p); i++ {
				Gia_ManSatokoReport(i, satoko.SATOKO_UNSAT, 0)
			}
			return
		}
		mapping := Vec_IntAlloc(len(cnf.Mapping))
		for _, value := range cnf.Mapping {
			Vec_IntPush(mapping, value)
		}
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			queryStart := time.Now()
			root := i + 1
			cone := Gia_ManCollectVars(root, mapping, int32(cnf.Variables))
			variables := make([]int, Vec_IntSize(cone))
			for k, id := range Vec_IntArray(cone) {
				variables[k] = int(id)
			}
			solver.PushAssumption(int(root << 1))
			solver.MarkCone(variables, true)
			status := solver.Solve()
			solver.MarkCone(variables, false)
			solver.PopAssumption()
			Gia_ManSatokoReport(i, int32(status), time.Since(queryStart))
		}
	} else if split != 0 {
		for i := int32(0); i < Gia_ManCoNum(p); i++ {
			cone := Gia_ManDupDfsCone(p, Gia_ManCo(p, i))
			Gia_ManSatokoCallOne(cone, options, i)
			Gia_ManStop(cone)
		}
	} else {
		Gia_ManSatokoCallOne(p, options, -1)
		return
	}
	fmt.Printf("Total time = %.2f sec\n", time.Since(start).Seconds())
}
