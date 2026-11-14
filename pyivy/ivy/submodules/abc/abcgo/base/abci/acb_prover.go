package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/base/acb"
)

func init() { acb.Acb_ProveMiter = acbProveMiter }

// acbProveMiter ports Acb_NtkSolve from src/base/acb/acbTest.c.
// It lives beside the ABCI functions it calls to avoid an ACB/ABCI import cycle.
func acbProveMiter(p *gia.Gia_Man_t) *int32 {
	manager := gia.Gia_ManToAig(p, 0)
	network := Abc_NtkFromAigPhase(manager)
	var params Prove_Params_t
	Prove_ParamsSetDefault(&params)
	params.fUseRewriting, params.fVerbose = 1, 0
	Aig_ManStop(manager)
	if network == nil {
		return nil
	}
	start := Abc_Clock()
	status := Abc_NtkIvyProve(&network, &params)
	model := baseabc.Abc_NtkReadModel(network)
	baseabc.Abc_NtkSetModel(network, nil)
	Abc_NtkDelete(network)
	result := "UNDECIDED"
	if status == 1 {
		result = "equivalent"
	} else if status == 0 {
		result = "NOT equivalent"
	}
	fmt.Printf("The networks are %s.  ", result)
	Abc_PrintTime(1, "Time", Abc_Clock()-start)
	if status == 0 {
		if len(model) == 0 {
			// A satisfiable zero-input miter still needs a non-nil witness:
			// Acb_OutputFile distinguishes NEQ from EQ by this pointer.
			return new(int32)
		}
		return &model[0]
	}
	return nil
}
