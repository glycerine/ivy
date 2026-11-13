package cec

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
)

// cecComputeChoicesCec4 follows Cec_ComputeChoicesNew in cecChoice.c.
// Every retained equivalence is rechecked on the original logic before use.
func cecComputeChoicesCec4(graph *Gia_Man_t, verbose int32) (*aig.Graph, error) {
	if graph == nil || gia.Gia_ManCoNum(graph)%3 != 0 || gia.Gia_ManRegNum(graph) > gia.Gia_ManCoNum(graph)/3 {
		return nil, fmt.Errorf("cec: invalid Cec4 choice snapshot dimensions")
	}
	if _, err := core.NewModel(graph); err != nil {
		return nil, err
	}
	work := graph
	var sourceIDs []int32
	for id := int32(0); int(id) < graph.ObjectCount(); id++ {
		_, _, _, kind := graph.Gate(id)
		if kind == core.Xor || kind == core.Mux || kind == core.Buffer {
			var err error
			work, sourceIDs, _, err = cecLowerSweepGraph(graph)
			if err != nil {
				return nil, err
			}
			defer gia.Gia_ManStop(work)
			break
		}
	}
	Cec4_ManSimulateTest2(work, verbose)
	var p Cec_ParFra_t
	Cec4_ManSetParams(&p)
	var simPars Cec_ParSim_t
	Cec_ManSimSetDefaultParams(&simPars)
	sim := Cec_ManSimStart(work, &simPars)
	defer Cec_ManSimStop(sim)
	if err := cecValidateChoiceClasses(work, sim, p.nBTLimit); err != nil {
		return nil, err
	}
	if sourceIDs != nil {
		if err := graph.TransferRepresentativesFrom(work, sourceIDs); err != nil {
			return nil, err
		}
	}
	choices := gia.Gia_ManEquivToChoices(work, 3)
	defer gia.Gia_ManStop(choices)
	return gia.Gia_ManToAig(choices, 1), nil
}
