// Go port of the SAT-sweeping wrapper in cecCore.c.
package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec_ManSatSolving runs the selected SAT backend and transfers its witnesses.
func Cec_ManSatSolving(graph *Gia_Man_t, pars *Cec_ParSat_t, proved int32) *Gia_Man_t {
	patterns := Cec_ManPatStart()
	defer Cec_ManPatStop(patterns)
	if pars.SolverType == -1 {
		Cec_ManSatSolve(patterns, graph, pars, nil, nil, nil, proved)
	} else {
		CecG_ManSatSolve(patterns, graph, pars, proved)
	}
	result := gia.Gia_ManCleanup(graph)
	models := graph.CombinationalCounterexamples()
	if models != nil {
		if err := result.SetCombinationalCounterexamples(models); err != nil {
			panic(err)
		}
		graph.ClearSequentialCounterexamples()
	}
	return result
}
