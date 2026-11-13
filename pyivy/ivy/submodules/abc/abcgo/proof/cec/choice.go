package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
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
	params.Error = nil
	if _, err := core.NewModel(graph); err != nil {
		params.Error = err
		return -1
	}
	if params.nWords <= 0 || params.nRounds < 0 || params.nBTLimit < 0 {
		params.Error = fmt.Errorf("cec: invalid choice computation parameters")
		return -1
	}
	for id := int32(0); int(id) < graph.ObjectCount(); id++ {
		_, _, _, kind := graph.Gate(id)
		if kind != core.Constant && kind != core.Input && kind != core.And && kind != core.Output {
			params.Error = fmt.Errorf("cec: choice refinement requires an AND-inverter graph")
			return -1
		}
	}
	graph.ClearRepresentatives()
	gia.Gia_ManRandom(1)
	var simPars Cec_ParSim_t
	Cec_ManSimSetDefaultParams(&simPars)
	simPars.nWords, simPars.nFrames, simPars.fVerbose = params.nWords, params.nRounds, params.fVerbose
	sim := Cec_ManSimStart(graph, &simPars)
	defer Cec_ManSimStop(sim)
	Cec_ManSimClassesPrepare(sim, -1)
	Cec_ManSimClassesRefine(sim)
	var satPars Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&satPars)
	satPars.nBTLimit, satPars.fVerbose = params.nBTLimit, params.fVerbose
	for round := 0; round < 1000; round++ {
		var outputs *Vec_Int_t
		query := Cec_ManCombSpecReduce(graph, &outputs, params.fUseRings)
		if gia.Gia_ManCoNum(query) == 0 {
			gia.Gia_ManStop(query)
			break
		}
		var status *Vec_Str_t
		var store *Vec_Int_t
		if params.fUseCSat != 0 {
			store = gia.Cbs_ManSolveMiterNc(query, params.nBTLimit, &status, 0, 0)
		} else {
			store = Cec_ManSatSolveMiter(query, &satPars, &status)
		}
		gia.Gia_ManStop(query)
		if vec.Vec_IntSize(store) == 0 {
			break
		}
		Cec_ManResimulateCounterExamplesComb(sim, store)
		Gia_ManCheckRefinements(graph, status, outputs, sim, params.fUseRings)
	}
	return 0
}

func Cec_ManChoiceComputationVec(graph *Gia_Man_t, snapshots int32, params *Cec_ParChc_t) *Gia_Man_t {
	if params == nil {
		params = new(Cec_ParChc_t)
		Cec_ManChcSetDefaultParams(params)
	}
	if graph == nil || snapshots <= 0 || gia.Gia_ManCoNum(graph)%snapshots != 0 || gia.Gia_ManRegNum(graph) > gia.Gia_ManCoNum(graph)/snapshots {
		params.Error = fmt.Errorf("cec: invalid choice snapshot dimensions")
		return nil
	}
	if err := cecRequireAIG(graph); err != nil {
		params.Error = err
		return nil
	}
	if Cec_ManChoiceComputation_int(graph, params) != 0 {
		return nil
	}
	return gia.Gia_ManEquivToChoices(graph, snapshots)
}

// Cec_ManChoiceComputation runs the original DAR new-choice sequence with
// GIA-based CEC refinement, then exports real choice links.
func Cec_ManChoiceComputation(graph *Gia_Man_t, params *Cec_ParChc_t) *Gia_Man_t {
	if params == nil {
		params = new(Cec_ParChc_t)
		Cec_ManChcSetDefaultParams(params)
	}
	if err := cecRequireAIG(graph); err != nil {
		params.Error = err
		return nil
	}
	if _, err := core.NewModel(graph); err != nil {
		params.Error = err
		return nil
	}
	options := dch.DefaultParams()
	options.UseGia = true
	options.ConflictLimit = int(params.nBTLimit)
	options.UseCSat = params.fUseCSat != 0
	options.Verbose = params.fVerbose != 0
	source := gia.Gia_ManToAig(graph, 0)
	result, err := Dar_ManChoiceNew(source, nil, &options)
	params.Error = err
	if err != nil {
		return nil
	}
	defer aig.Aig_ManStop(result)
	return gia.Gia_ManFromAig(result)
}
func Cec_ComputeChoices(graph *Gia_Man_t, params *Dch_Pars_t) *Aig_Man_t {
	if params == nil {
		defaults := dch.DefaultParams()
		params = &defaults
	}
	var choice Cec_ParChc_t
	Cec_ManChcSetDefaultParams(&choice)
	choice.nBTLimit = int32(params.ConflictLimit)
	if params.UseCSat {
		choice.fUseCSat = 1
		if choice.nBTLimit > 100 {
			choice.nBTLimit = 100
		}
	}
	if params.Verbose {
		choice.fVerbose = 1
	}
	result := Cec_ManChoiceComputationVec(graph, 3, &choice)
	if result == nil {
		if params.Verbose {
			fmt.Printf("CEC choices: %v\n", choice.Error)
		}
		return nil
	}
	defer gia.Gia_ManStop(result)
	return gia.Gia_ManToAig(result, 1)
}
func Cec_ComputeChoicesNew(graph *Gia_Man_t, verbose int32) *Aig_Man_t {
	result, err := cecComputeChoicesCec4(graph, verbose)
	if err != nil {
		if verbose != 0 {
			fmt.Printf("CEC choices: %v\n", err)
		}
		return nil
	}
	return result
}
