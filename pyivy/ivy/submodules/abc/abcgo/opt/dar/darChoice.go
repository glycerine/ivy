package dar

import (
	"context"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/dch"
	"io"
	"time"
)

// ChoiceEngine computes proved equivalences between independent synthesis
// versions with identical CI/CO order. DCH/CEC adapters live above DAR so GIA
// compression can use DAR without creating a dependency cycle.
type ChoiceEngine interface {
	ComputeChoices([]*aig.Graph) (*aig.Graph, error)
}

// PartitionedChoiceEngine uses ABC's partitioned choice driver with the
// supplied FRAIG backend. This is distinct from Dar_ManChoiceNew's C path.
type PartitionedChoiceEngine struct {
	PartitionSize, ConflictLimit, LevelLimit int32
	Verbose                                  bool
	Backend                                  aig.PartitionFraigFunc
}

func (e PartitionedChoiceEngine) ComputeChoices(versions []*aig.Graph) (*aig.Graph, error) {
	if len(versions) < 2 {
		return nil, fmt.Errorf("dar: partitioned choices require at least two versions")
	}
	if e.Backend == nil {
		return nil, fmt.Errorf("dar: partitioned choices require a FRAIG backend")
	}
	size := e.PartitionSize
	if size == 0 {
		size = 300
	}
	verbose := int32(0)
	if e.Verbose {
		verbose = 1
	}
	result := aig.Aig_ManChoicePartitionedWithFraig(versions, size, e.ConflictLimit, e.LevelLimit, verbose, e.Backend)
	return result, result.Check()
}

type ChoiceOptions struct {
	Context                            context.Context
	Balance, UpdateLevel, Power, Light bool
	Progress                           io.Writer
	Engine                             ChoiceEngine
	SynthesisDuration                  time.Duration
}

func Dar_NewChoiceSynthesis(g *aig.Graph, p *ChoiceOptions) ([]*aig.Graph, error) {
	if g == nil || p == nil {
		return nil, fmt.Errorf("dar: missing graph or choice options")
	}
	update := p.UpdateLevel && !Dar_NewChoiceSynthesisGuard(g)
	options := CompressOptions{Context: p.Context, Balance: p.Balance, UpdateLevel: update, Power: p.Power, Fanout: true, Progress: p.Progress}
	original := aig.Aig_ManDupDfs(g)
	first, err := Compress(original, options)
	if err != nil {
		return nil, err
	}
	options.Light = p.Light
	second, err := compress2(first, options, true)
	if err != nil {
		return nil, err
	}
	return []*aig.Graph{second, first, original}, nil
}

// Dar_ManChoiceNew runs the newer light-synthesis-aware version sequence and
// consumes g on success, transferring its timing manager to the result.
// The source's default DCH/GIA path is not yet wired across package boundaries;
// callers must supply an engine explicitly until that integration is complete.
func Dar_ManChoiceNew(g *aig.Graph, p *ChoiceOptions) (*aig.Graph, error) {
	return runNewChoice(g, p, false)
}
func Dar_ManChoiceNewAig(g *aig.Graph, p *ChoiceOptions) (*aig.Graph, error) {
	return runNewChoice(g, p, true)
}
func runNewChoice(g *aig.Graph, p *ChoiceOptions, classicSynthesis bool) (*aig.Graph, error) {
	if g == nil || p == nil {
		return nil, fmt.Errorf("dar: missing graph or choice options")
	}
	engine := p.Engine
	if engine == nil {
		if !classicSynthesis {
			return nil, fmt.Errorf("dar: source DCH/CEC choice path is not wired; supply an explicit engine")
		}
		params := dch.DefaultParams()
		params.Power = p.Power
		engine = dch.Engine{Params: &params}
	}
	timing := g.Timing
	if !classicSynthesis {
		// The active C new-choice path removes timing before synthesis and
		// transfers the original manager to the final AIG.
		g.Timing = nil
	}
	consumed := false
	defer func() {
		if !consumed {
			g.Timing = timing
		}
	}()
	started := time.Now()
	var versions []*aig.Graph
	var err error
	if classicSynthesis {
		options := CompressOptions{Context: p.Context, Balance: true, UpdateLevel: true, Power: p.Power, Fanout: true, Progress: p.Progress}
		original := aig.Aig_ManDupDfs(g)
		first, e := Compress(original, options)
		if e != nil {
			return nil, e
		}
		second, e := Compress2(first, options)
		if e != nil {
			return nil, e
		}
		versions = []*aig.Graph{second, first, original}
	} else {
		options := *p
		options.Balance, options.UpdateLevel = true, true
		versions, err = Dar_NewChoiceSynthesis(g, &options)
		if err != nil {
			return nil, err
		}
	}
	defer func() {
		for _, version := range versions {
			aig.Aig_ManStop(version)
		}
	}()
	p.SynthesisDuration = time.Since(started)
	if p.Context != nil {
		if err = p.Context.Err(); err != nil {
			return nil, err
		}
	}
	result, err := engine.ComputeChoices(versions)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("dar: choice engine returned nil graph")
	}
	if err = result.Check(); err != nil {
		return nil, err
	}
	if len(result.Inputs()) != len(g.Inputs()) || len(result.Outputs()) != len(g.Outputs()) {
		return nil, fmt.Errorf("dar: choice engine changed the ordered interface size")
	}
	order := aig.Aig_ManOrderPios(result, g)
	name, spec := g.Name, g.Spec
	g.Timing = nil
	aig.Aig_ManStop(g)
	consumed = true
	intermediate := result
	result = aig.Aig_ManDupDfsGuided(intermediate, order)
	aig.Aig_ManStop(intermediate)
	result.Timing = timing
	aig.Aig_ManChoiceLevel(result)
	result.Name, result.Spec = name, spec
	return result, result.Check()
}

func DefaultChoiceOptions() ChoiceOptions { return ChoiceOptions{Balance: true, UpdateLevel: true} }
