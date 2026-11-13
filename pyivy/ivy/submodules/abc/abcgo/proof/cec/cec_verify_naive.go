package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"time"
)

// Cec_ManVerifyNaive follows cecCec.c: generate eight-input MF CNF, load
// BSAT, and check both implications for each pair of primary outputs.
// Like ABC's routine, it reports the aggregate result without storing a model.
func Cec_ManVerifyNaive(graph *Gia_Man_t, params *Cec_ParCec_t) int32 {
	if params == nil {
		params = new(Cec_ParCec_t)
		Cec_ManCecSetDefaultParams(params)
	}
	params.Error = nil
	if graph == nil || gia.Gia_ManPoNum(graph)%2 != 0 || params.nBTLimit < 0 {
		params.Error = fmt.Errorf("cec: invalid naive verification input")
		return -1
	}
	coIDs := make([]int32, gia.Gia_ManCoNum(graph))
	for i := range coIDs {
		coIDs[i] = gia.Gia_ObjId(graph, gia.Gia_ManCo(graph, int32(i)))
	}
	data, varNums, err := cecNaiveMappedCNF(graph, coIDs)
	if err != nil {
		params.Error = err
		return -1
	}
	solver := cnf.Cnf_DataWriteIntoSolver(&cnf.Cnf_Dat_t{Variables: data.Variables, Clauses: data.Clauses}, 1, 0)
	if solver == nil {
		params.Error = fmt.Errorf("cec: failed to load mapped CNF")
		return -1
	}
	solver.SetLimits(int64(params.nBTLimit), 0)
	started := time.Now()
	nPairs := gia.Gia_ManPoNum(graph) / 2
	var unsat, sat, undec, trivial int32
	for i := int32(0); i < nPairs; i++ {
		a := gia.Gia_ManPo(graph, 2*i)
		b := gia.Gia_ManPo(graph, 2*i+1)
		if gia.Gia_ObjFaninLit0(a, gia.Gia_ObjId(graph, a)) == gia.Gia_ObjFaninLit0(b, gia.Gia_ObjId(graph, b)) {
			unsat++
			trivial++
			continue
		}
		if params.TimeLimit != 0 && time.Since(started) >= time.Duration(params.TimeLimit)*time.Second {
			fmt.Printf("Timeout (%d sec) is reached.\n", params.TimeLimit)
			undec = nPairs - unsat - sat
			break
		}
		// The C MF wrapper maps coarsened CO variables back to source object IDs.
		v0, v1 := varNums[gia.Gia_ObjId(graph, a)], varNums[gia.Gia_ObjId(graph, b)]
		if v0 < 0 || v1 < 0 {
			panic("cec: missing MF output variable")
		}
		l0, l1 := bsat.Negative(v0), bsat.Positive(v1)
		status := solver.Solve(l0, l1)
		if status == bsat.Satisfiable {
			fmt.Printf("Output %d is SAT.\n", i)
			sat++
			continue
		}
		if status != bsat.Unsatisfiable {
			undec++
			continue
		}
		l0, l1 = l0^1, l1^1
		if !solver.AddClause(l0, l1) {
			panic("cec: source implication clause is inconsistent")
		}
		status = solver.Solve(l0, l1)
		if status == bsat.Satisfiable {
			fmt.Printf("Output %d is SAT.\n", i)
			sat++
			continue
		}
		if status != bsat.Unsatisfiable {
			undec++
			continue
		}
		if !solver.AddClause(l0^1, l1^1) {
			panic("cec: source implication clause is inconsistent")
		}
		unsat++
	}
	fmt.Printf("UNSAT = %6d.  SAT = %6d.   UNDEC = %6d.  Trivial = %6d.  Time = %8.2f sec\n", unsat, sat, undec, trivial, time.Since(started).Seconds())
	if sat > 0 {
		return 0
	}
	if undec > 0 {
		return -1
	}
	return 1
}
