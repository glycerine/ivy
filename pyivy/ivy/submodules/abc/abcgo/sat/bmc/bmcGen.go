package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
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
	start := time.Now()
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
	var satTime time.Duration
	for i := int32(0); i < 64; i++ {
		before := time.Now()
		status := s.Solve()
		satTime += time.Since(before)
		if status == bsat.Satisfiable {
			fmt.Print("+")
			count++
		} else {
			fmt.Print("-")
		}
	}
	fmt.Printf("\nFinished generating %d assignments.  Time = %s. SAT solver time = %s.\n", count, time.Since(start), satTime)
	return 1
}
