package cec

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/cec/core"
)

// cecComputeChoicesCec4 follows Cec_ComputeChoicesNew in cecChoice.c.
func cecComputeChoicesCec4(graph *Gia_Man_t, verbose int32) (*aig.Graph, error) {
	if graph == nil || gia.Gia_ManCoNum(graph)%3 != 0 || gia.Gia_ManRegNum(graph) > gia.Gia_ManCoNum(graph)/3 {
		return nil, fmt.Errorf("cec: invalid Cec4 choice snapshot dimensions")
	}
	if _, err := core.NewModel(graph); err != nil {
		return nil, err
	}
	if err := cecRequireAIG(graph); err != nil {
		return nil, err
	}
	Cec4_ManSimulateTest2(graph, verbose)
	choices := gia.Gia_ManEquivToChoices(graph, 3)
	defer gia.Gia_ManStop(choices)
	return gia.Gia_ManToAig(choices, 1), nil
}
