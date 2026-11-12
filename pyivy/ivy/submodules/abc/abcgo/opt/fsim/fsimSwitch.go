package fsim

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/bits"
)

// The C switching file contained no implementation. This implements its
// advertised activity API on the same frontier kernels, with explicit rates.
func Fsim_ManSwitchSimulate(g *aig.Graph, parameters *Fsim_ParSwitch_t) (*Switching, error) {
	pars := Fsim_ParSwitch_t{}
	Fsim_ManSetDefaultParamsSwitch(&pars)
	if parameters != nil {
		pars = *parameters
	}
	if pars.Words <= 0 || pars.Iterations <= 0 || pars.Prefix < 0 || pars.Prefix >= pars.Iterations || pars.RandomPI < 0 || pars.RandomPI > 31 || !pars.ProbabilityOne && !pars.ProbabilityTransition {
		return nil, fmt.Errorf("fsim: invalid switching parameters")
	}
	p, err := Fsim_ManCreate(g)
	if err != nil {
		return nil, err
	}
	defer Fsim_ManDelete(p)
	if err = Fsim_ManFront(p, false); err != nil {
		return nil, err
	}
	count := len(g.Nodes())
	ones, transitions := make([]uint64, count), make([]uint64, count)
	previous := make([][]uint32, count)
	for i := range previous {
		previous[i] = make([]uint32, pars.Words)
	}
	cis := make([][]uint32, p.CIs)
	for i := range cis {
		cis[i] = make([]uint32, pars.Words)
	}
	rng := aig.NewAigRandomSource()
	rng.Uint32()
	for _, i := range p.CIOrder {
		if i < p.PIs {
			for w := pars.Words - 1; w >= 0; w-- {
				cis[i][w] = rng.Uint32()
			}
		}
	}
	result := &Switching{}
	for frame := 0; frame < pars.Iterations; frame++ {
		p.observer = func(id int, words []uint32) {
			for w, value := range words {
				if frame >= pars.Prefix && pars.ProbabilityOne {
					ones[id] += uint64(bits.OnesCount32(value))
				}
				if frame >= max(pars.Prefix, 1) && pars.ProbabilityTransition {
					transitions[id] += uint64(bits.OnesCount32(value ^ previous[id][w]))
				}
				previous[id][w] = value
			}
		}
		outputs, err := p.Round(cis)
		if err != nil {
			return nil, err
		}
		if frame >= pars.Prefix {
			result.OneFrames++
		}
		if frame >= max(pars.Prefix, 1) {
			result.TransitionFrames++
		}
		for r := p.PIs; r < p.CIs; r++ {
			copy(cis[r], outputs[p.POs+r-p.PIs])
		}
		for _, i := range p.CIOrder {
			if i >= p.PIs {
				continue
			}
			for w := pars.Words - 1; w >= 0; w-- {
				mask := rng.Uint32()
				for k := 0; k < pars.RandomPI; k++ {
					mask &= rng.Uint32()
				}
				cis[i][w] ^= mask
			}
		}
	}
	if pars.ProbabilityOne {
		result.One = make([]float32, count)
		for i, n := range ones {
			result.One[i] = float32(n) / float32(32*pars.Words*result.OneFrames)
		}
	}
	if pars.ProbabilityTransition {
		result.Transition = make([]float32, count)
		if result.TransitionFrames > 0 {
			for i, n := range transitions {
				result.Transition[i] = float32(n) / float32(32*pars.Words*result.TransitionFrames)
			}
		}
	}
	if pars.Verbose {
		fmt.Printf("FSIM activity: samples=%d transition samples=%d\n", 32*pars.Words*result.OneFrames, 32*pars.Words*result.TransitionFrames)
	}
	return result, nil
}
