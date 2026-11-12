// Package dch computes SAT-proven AIG equivalences and acyclic mapping choices.
// The simulation, sweeping and choice construction follow ABC's DCH engine
// (Alan Mishchenko, UC Berkeley, 2008), using explicit edges and Go-owned state.
package dch

import (
	"time"

	aig "github.com/glycerine/abcgo/aig/aig"
)

// Params retains the old DCH controls. Synthesis/GIA/new-engine fields are
// consumed by the surrounding choice script, as in the source, not by old DCH.
type Params struct {
	Words, ConflictLimit, SatVarMax, NodesAhead, CallsRecycle          int
	Synthesis, PolarFlip, SimulateTFO, Power                           bool
	UseGia, UseCSat, UseNew, LightSynth, SkipRedundantSupport, Verbose bool
	SynthDuration                                                      time.Duration
}
type Dch_Pars_t = Params

func DefaultParams() Params {
	return Params{Words: 8, ConflictLimit: 1000, SatVarMax: 5000, Synthesis: true, PolarFlip: true, SimulateTFO: true, NodesAhead: 1000, CallsRecycle: 100}
}
func Dch_ManSetDefaultParams(p *Params) { *p = DefaultParams() }
func Dch_ManReadVerbose(p *Params) int {
	if p.Verbose {
		return 1
	}
	return 0
}

// Engine adapts DCH to the typed choice engine interface used by DAR and CEC.
type Engine struct{ Params *Params }

func (e Engine) ComputeChoices(variants []*aig.Graph) (*aig.Graph, error) {
	total, err := DeriveTotalAig(variants)
	if err != nil {
		return nil, err
	}
	return ComputeChoices(total, e.Params)
}
func Dch_ComputeChoices(g *aig.Graph, p *Params) (*aig.Graph, error) { return ComputeChoices(g, p) }
func Dch_ComputeEquivalences(g *aig.Graph, p *Params) error          { return ComputeEquivalences(g, p) }
