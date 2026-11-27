package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Gia_ManMoObj returns the packed simulation words for one object.
func Gia_ManMoObj(p *gia.Gia_Man_t, id int32) []uint64 { return gia.Gia_ManObjSim(p, id) }
func Gia_ManMoSetCi(p *gia.Gia_Man_t, id int32)        { gia.Gia_ManObjSimPi(p, id) }
func Gia_ManMoSimAnd(p *gia.Gia_Man_t, id int32)       { gia.Gia_ManObjSimAnd(p, id) }
func Gia_ManMoSetCo(p *gia.Gia_Man_t, id int32)        { gia.Gia_ManObjSimPo(p, id) }

// Gia_ManMoFindSimulate initializes deterministic random packed simulations.
func Gia_ManMoFindSimulate(p *gia.Gia_Man_t, nWords int32) {
	gia.Gia_ManRandomW(1)
	gia.Gia_ManSimulateWords(p, nWords)
}

// Gia_ManTestSatEnum runs the source's repeated SAT-enumeration experiment.
// The computed simulation assignments are not applied to the solver, matching
// the original source where the activity and polarity calls are commented out.
func Gia_ManTestSatEnum(p *gia.Gia_Man_t) int32 {
	if p == nil || gia.Gia_ManCoNum(p) == 0 {
		return 0
	}
	clk := util.Abc_Clock()
	data, err := gia.Mf_ManGenerateCnfResult(p, 8, true, false, false)
	if err != nil {
		fmt.Printf("Cannot generate CNF: %v\n", err)
		return 0
	}
	s := bsat.NewSolver()
	for i := 0; i < data.Variables; i++ {
		s.NewVariable()
	}
	for _, clause := range data.Clauses {
		if !s.AddClause(clause...) {
			return 0
		}
	}
	outID := gia.Gia_ObjId(p, gia.Gia_ManCo(p, 0))
	if !s.AddClause(bsat.Positive(data.VarNums[outID])) {
		return 0
	}
	Gia_ManMoFindSimulate(p, 1)
	sim := Gia_ManMoObj(p, outID)
	for i := int32(0); i < 64; i++ {
		fmt.Printf("%d", util.Abc_TtGetBit(sim, i))
	}
	fmt.Println()
	count := 0
	var clkTotal int64
	for i := int32(0); i < 64; i++ {
		clk2 := util.Abc_Clock()
		status := s.Solve()
		clkTotal += util.Abc_Clock() - clk2
		if status == bsat.Satisfiable {
			fmt.Print("+")
			count++
		} else {
			fmt.Print("-")
		}
	}
	fmt.Printf("\n")
	fmt.Printf("Finished generating %d assignments.  ", count)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	util.Abc_PrintTime(1, "SAT solver time", clkTotal)
	return 1
}
