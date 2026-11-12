package gia

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"time"
)

// qbfSolve decides exists parameters, forall remaining inputs, output=1.
// Status is 0 for SAT, 1 for UNSAT, and -1 for an exhausted limit.
func qbfSolve(g *Gia_Man_t, nPars, nIterLimit, nConfLimit, nTimeOut int32) (int32, []bool) {
	if g == nil {
		panic("gia: invalid QBF problem")
	}
	nInputs := Gia_ManPiNum(g)
	if g.nRegs != 0 || Gia_ManPoNum(g) != 1 || nPars < 0 || nPars > nInputs {
		panic("gia: invalid QBF problem")
	}
	verify := satoko.New()
	verify.SetVariables(int(nInputs))
	inputs := make([]int, nInputs)
	for i := range inputs {
		inputs[i] = 2 * i
	}
	output := qbfEncodeGraph(verify, g, inputs)
	verify.AddClause([]int{output ^ 1})
	synth := satoko.New()
	synth.SetVariables(int(nPars))
	params := make([]bool, nPars)
	start := time.Now()
	for iteration := int32(0); ; iteration++ {
		assumptions := make([]int, nPars)
		for i, value := range params {
			assumptions[i] = 2 * i
			if !value {
				assumptions[i] ^= 1
			}
		}
		status := verify.SolveAssumptions(assumptions)
		if status == satoko.SATOKO_UNSAT {
			return 0, params
		}
		if status != satoko.SATOKO_SAT {
			return -1, nil
		}
		constant := 2 * synth.AddVariable(false)
		synth.AddClause([]int{constant ^ 1})
		cofactorInputs := make([]int, nInputs)
		for i := range cofactorInputs {
			if i < int(nPars) {
				cofactorInputs[i] = 2 * i
			} else {
				cofactorInputs[i] = constant
				if verify.ModelValue(i) == 1 {
					cofactorInputs[i] ^= 1
				}
			}
		}
		condition := qbfEncodeGraph(synth, g, cofactorInputs)
		if !synth.AddClause([]int{condition}) {
			return 1, nil
		}
		status = synth.SolveAssumptionsLimit(nil, int64(nConfLimit))
		if status == satoko.SATOKO_UNSAT {
			return 1, nil
		}
		if status != satoko.SATOKO_SAT {
			return -1, nil
		}
		for i := range params {
			params[i] = synth.ModelValue(i) == 1
		}
		if nIterLimit > 0 && iteration+1 >= nIterLimit || nTimeOut > 0 && time.Since(start) >= time.Duration(nTimeOut)*time.Second {
			return -1, nil
		}
	}
}
