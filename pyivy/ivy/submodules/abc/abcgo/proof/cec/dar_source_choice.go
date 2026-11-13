package cec

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	dar "github.com/glycerine/abcgo/opt/dar/core"
	"github.com/glycerine/abcgo/proof/dch"
)

// Dar_ManChoiceNew runs ABC's active DAR new-choice pipeline through this
// higher package, where GIA, DCH and CEC are all available without an import
// cycle. It consumes the input graph on success, as the C entry point does,
// and returns errors when a Go operation fails. Nil options or parameters use
// defaults.
func Dar_ManChoiceNew(graph *aig.Graph, options *dar.ChoiceOptions, params *dch.Params) (*aig.Graph, error) {
	if params == nil {
		defaults := dch.DefaultParams()
		params = &defaults
	}
	config := dar.DefaultChoiceOptions()
	if options != nil {
		config = *options
	}
	config.Power, config.Light = params.Power, params.LightSynth
	config.DchParams = params
	config.Engine = DarSourceChoiceEngine{Params: params}
	result, err := dar.Dar_ManChoiceNew(graph, &config)
	params.SynthDuration = config.SynthesisDuration
	if options != nil {
		options.SynthesisDuration = config.SynthesisDuration
	}
	return result, err
}

// DarSourceChoiceEngine supplies the proof portion of the active C
// Dar_ManChoiceNew path. DAR passes the three synthesized snapshots in the
// source order; the engine builds their GIA miter and selects DCH, CEC or
// Cec4 according to the original DCH parameters. It does not consume the
// caller's snapshot graphs.
type DarSourceChoiceEngine struct{ Params *dch.Params }

func (e DarSourceChoiceEngine) ComputeChoices(versions []*aig.Graph) (*aig.Graph, error) {
	if len(versions) != 3 {
		return nil, fmt.Errorf("cec: DAR choice path requires three snapshots")
	}
	params := e.Params
	if params == nil {
		defaults := dch.DefaultParams()
		params = &defaults
	}
	snapshots := make([]*gia.Gia_Man_t, 0, len(versions))
	defer func() {
		for _, graph := range snapshots {
			gia.Gia_ManStop(graph)
		}
	}()
	for i, graph := range versions {
		if graph == nil {
			return nil, fmt.Errorf("cec: nil DAR snapshot %d", i)
		}
		if err := graph.Check(); err != nil {
			return nil, fmt.Errorf("cec: DAR snapshot %d: %w", i, err)
		}
		if i != 0 && (graph.InputCount() != versions[0].InputCount() || graph.OutputCount() != versions[0].OutputCount() || aig.Aig_ManRegNum(graph) != aig.Aig_ManRegNum(versions[0])) {
			return nil, fmt.Errorf("cec: incompatible DAR snapshot interfaces")
		}
		snapshots = append(snapshots, gia.Gia_ManFromAig(graph))
	}
	miter := gia.Gia_ManChoiceMiter(snapshots)
	defer gia.Gia_ManStop(miter)
	if params.UseNew {
		verbose := int32(0)
		if params.Verbose {
			verbose = 1
		}
		result := Cec_ComputeChoicesNew(miter, verbose)
		if result == nil {
			return nil, fmt.Errorf("cec: Cec4 choice computation failed")
		}
		return result, nil
	}
	if params.UseGia {
		result := Cec_ComputeChoices(miter, params)
		if result == nil {
			return nil, fmt.Errorf("cec: GIA choice computation failed")
		}
		return result, nil
	}
	return dch.ComputeChoices(gia.Gia_ManToAigSkip(miter, 3), params)
}
