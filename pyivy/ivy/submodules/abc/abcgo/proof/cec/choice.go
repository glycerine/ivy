package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/opt/dar"
	"github.com/glycerine/abcgo/proof/cec/core"
	"github.com/glycerine/abcgo/proof/dch"
)

type Dch_Pars_t = dch.Params

func cecChoiceParams(params *Cec_ParChc_t) dch.Params {
	result := dch.DefaultParams()
	if params != nil {
		result.ConflictLimit = int(params.nBTLimit)
		result.Words = int(params.nWords)
		result.Verbose = params.fVerbose != 0
		result.UseCSat = params.fUseCSat != 0
	}
	return result
}

// ChoiceSnapshots interprets the interleaved output layout used by the
// original Gia_ManEquivToChoices: output i belongs to snapshot i%snapshots.
// Inputs are shared and the first snapshot supplies the final output function.
func ChoiceSnapshots(graph *Gia_Man_t, snapshots int, params *dch.Params) (*aig.Graph, error) {
	if graph == nil || snapshots <= 0 || len(graph.OutputLiterals())%snapshots != 0 {
		return nil, fmt.Errorf("cec: invalid choice snapshot dimensions")
	}
	if _, err := core.NewModel(graph); err != nil {
		return nil, err
	}
	converted := gia.Gia_ManToAigSimple(graph)
	outputs := converted.Outputs()
	versions := make([]*aig.Graph, snapshots)
	for index := range versions {
		selected := make([]aig.Edge, 0, len(outputs)/snapshots)
		for i := index; i < len(outputs); i += snapshots {
			selected = append(selected, outputs[i])
		}
		if graph.RegisterCount() > len(selected) {
			return nil, fmt.Errorf("cec: snapshot has fewer outputs than registers")
		}
		versions[index] = aig.Aig_ManDupSimpleDfsPart(converted, converted.Inputs(), selected)
		versions[index].Name = converted.Name
		versions[index].Spec = converted.Spec
		aig.Aig_ManSetRegNum(versions[index], int32(graph.RegisterCount()))
	}
	return (dch.Engine{Params: params}).ComputeChoices(versions)
}

func Cec_ManChoiceComputation_int(graph *Gia_Man_t, params *Cec_ParChc_t) int32 {
	if params == nil {
		params = new(Cec_ParChc_t)
		Cec_ManChcSetDefaultParams(params)
	}
	sweep := new(Cec_ParFra_t)
	Cec_ManFraSetDefaultParams(sweep)
	sweep.nWords = params.nWords
	sweep.nRounds = params.nRounds
	sweep.nBTLimit = params.nBTLimit
	reduced, err := SatSweep(graph, sweep)
	params.Error = err
	if reduced != nil {
		gia.Gia_ManStop(reduced)
	}
	if err != nil {
		return -1
	}
	return 0
}
func Cec_ManChoiceComputationVec(graph *Gia_Man_t, snapshots int32, params *Cec_ParChc_t) *Gia_Man_t {
	if params == nil {
		params = new(Cec_ParChc_t)
		Cec_ManChcSetDefaultParams(params)
	}
	options := cecChoiceParams(params)
	result, err := ChoiceSnapshots(graph, int(snapshots), &options)
	params.Error = err
	if err != nil {
		return nil
	}
	return gia.Gia_ManFromAig(result)
}

// Cec_ManChoiceComputation runs the original three-version DAR synthesis
// sequence with the native DCH proof engine, then exports real choice links.
func Cec_ManChoiceComputation(graph *Gia_Man_t, params *Cec_ParChc_t) *Gia_Man_t {
	if params == nil {
		params = new(Cec_ParChc_t)
		Cec_ManChcSetDefaultParams(params)
	}
	if _, err := core.NewModel(graph); err != nil {
		params.Error = err
		return nil
	}
	options := cecChoiceParams(params)
	source := gia.Gia_ManToAigSimple(graph)
	result, err := dar.Dar_ManChoiceNew(source, &dar.ChoiceOptions{UpdateLevel: true, Engine: dch.Engine{Params: &options}})
	params.Error = err
	if err != nil {
		return nil
	}
	return gia.Gia_ManFromAig(result)
}
func Cec_ComputeChoices(graph *Gia_Man_t, params *Dch_Pars_t) *Aig_Man_t {
	result, err := ChoiceSnapshots(graph, 3, params)
	if err != nil {
		if params != nil && params.Verbose {
			fmt.Printf("CEC choices: %v\n", err)
		}
		return nil
	}
	return result
}
func Cec_ComputeChoicesNew(graph *Gia_Man_t, verbose int32) *Aig_Man_t {
	params := dch.DefaultParams()
	params.Verbose = verbose != 0
	return Cec_ComputeChoices(graph, &params)
}
