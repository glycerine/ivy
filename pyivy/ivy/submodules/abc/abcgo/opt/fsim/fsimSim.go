package fsim

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
	"time"
)

func (p *Manager) Round(cis [][]uint32) ([][]uint32, error) {
	if len(cis) != p.CIs {
		return nil, fmt.Errorf("fsim: CI dimensions differ")
	}
	words := p.words
	if len(cis) > 0 {
		words = len(cis[0])
	}
	if words <= 0 {
		return nil, fmt.Errorf("fsim: simulation word count must be positive")
	}
	for _, row := range cis {
		if len(row) != words {
			return nil, fmt.Errorf("fsim: inconsistent simulation word widths")
		}
	}
	if p.program == nil && !p.compressed {
		if err := Fsim_ManFront(p, false); err != nil {
			return nil, err
		}
	}
	if len(p.data) != p.Frontier*words {
		p.data = make([]uint32, p.Frontier*words)
	}
	p.words = words
	outputs := make([][]uint32, p.COs)
	for i := range outputs {
		outputs[i] = make([]uint32, words)
	}
	if p.constSlot > 0 {
		row := p.data[p.constSlot%p.Frontier*words:][:words]
		for i := range row {
			row[i] = ^uint32(0)
		}
	}
	if p.observer != nil {
		one := make([]uint32, words)
		for i := range one {
			one[i] = ^uint32(0)
		}
		p.observer(0, one)
	}
	get := func(lit, word int) uint32 {
		value := p.data[(lit>>1)%p.Frontier*words+word]
		if lit&1 != 0 {
			return ^value
		}
		return value
	}
	err := p.forEachOperation(func(o Fsim_Obj_t) {
		var dst []uint32
		if o.Kind == aig.Output {
			dst = outputs[o.Index]
		} else {
			dst = p.data[o.Node%p.Frontier*words:][:words]
		}
		switch o.Kind {
		case aig.Input:
			copy(dst, cis[o.Index])
		case aig.And:
			for w := range dst {
				dst[w] = get(o.Fan0, w) & get(o.Fan1, w)
			}
		case aig.ExclusiveOr:
			for w := range dst {
				dst[w] = get(o.Fan0, w) ^ get(o.Fan1, w)
			}
		case aig.Buffer, aig.Output:
			for w := range dst {
				dst[w] = get(o.Fan0, w)
			}
		}
		if p.observer != nil {
			p.observer(o.Source.NodeID(), dst)
		}
	})
	if err != nil {
		return nil, err
	}
	return outputs, nil
}
func Fsim_ManSimInfoIsZero(info []uint32) int {
	for w := len(info) - 1; w >= 0; w-- {
		if info[w] != 0 {
			return w*32 + bits.TrailingZeros32(info[w])
		}
	}
	return -1
}
func Fsim_ManCheckPos(outputs [][]uint32, nPOs int) (output, pattern int, found bool) {
	for i := 0; i < nPOs; i++ {
		if bit := Fsim_ManSimInfoIsZero(outputs[i]); bit >= 0 {
			return i, bit, true
		}
	}
	return 0, 0, false
}
func Fsim_ManGenerateCounter(g *aig.Graph, frame, output, words, pattern int, ciOrder []int) (*util.Abc_Cex_t, error) {
	regs := int(aig.Aig_ManRegNum(g))
	pis := g.InputCount() - regs
	if frame < 0 || output < 0 || output >= g.OutputCount()-regs || words <= 0 || pattern < 0 || pattern >= 32*words || len(ciOrder) != g.InputCount() {
		return nil, fmt.Errorf("fsim: invalid witness dimensions")
	}
	seen := make([]bool, len(ciOrder))
	for _, i := range ciOrder {
		if i < 0 || i >= len(seen) || seen[i] {
			return nil, fmt.Errorf("fsim: invalid witness CI permutation")
		}
		seen[i] = true
	}
	model := util.Abc_CexAlloc(int32(regs), int32(pis), int32(frame+1))
	model.IFrame, model.IPo = int32(frame), int32(output)
	rng := aig.NewAigRandomSource()
	rng.Uint32()
	for f := 0; f <= frame; f++ {
		for _, input := range ciOrder {
			if input >= pis {
				continue
			}
			for word := words - 1; word >= 0; word-- {
				value := rng.Uint32()
				if word == pattern/32 && value&(1<<uint(pattern&31)) != 0 {
					bit := regs + f*pis + input
					model.PData[bit>>5] |= 1 << uint(bit&31)
				}
			}
		}
	}
	if err := verifyCounter(g, model); err != nil {
		return nil, err
	}
	return model, nil
}
func verifyCounter(g *aig.Graph, model *util.Abc_Cex_t) error {
	regs := int(aig.Aig_ManRegNum(g))
	pis, pos := g.InputCount()-regs, g.OutputCount()-regs
	state := make([]bool, regs)
	for r := range state {
		state[r] = model.PData[r>>5]&(1<<uint(r&31)) != 0
	}
	for frame := 0; frame <= int(model.IFrame); frame++ {
		inputs := make([]bool, pis+regs)
		for i := 0; i < pis; i++ {
			bit := regs + frame*pis + i
			inputs[i] = model.PData[bit>>5]&(1<<uint(bit&31)) != 0
		}
		copy(inputs[pis:], state)
		out, err := g.Evaluate(inputs)
		if err != nil {
			return err
		}
		if frame == int(model.IFrame) && !out[model.IPo] {
			return fmt.Errorf("fsim: generated witness does not fail the reported output")
		}
		copy(state, out[pos:])
	}
	return nil
}
func RunSimulation(g *aig.Graph, parameters *Fsim_ParSim_t) (*SimulationResult, error) {
	pars := Fsim_ParSim_t{}
	Fsim_ManSetDefaultParamsSim(&pars)
	if parameters != nil {
		pars = *parameters
	}
	if pars.Words <= 0 || pars.Iterations <= 0 || pars.TimeLimit < 0 {
		return nil, fmt.Errorf("fsim: invalid simulation parameters")
	}
	p, err := Fsim_ManCreate(g)
	if err != nil {
		return nil, err
	}
	defer Fsim_ManDelete(p)
	p.words = pars.Words
	if aig.Aig_ManRegNum(g) <= 0 {
		return nil, fmt.Errorf("fsim: sequential simulation needs registers")
	}
	if err = Fsim_ManFront(p, pars.CompressAIG); err != nil {
		return nil, err
	}
	if pars.CheckMiter {
		g.ClearSequentialCounterexample()
	}
	cis := make([][]uint32, p.CIs)
	for i := range cis {
		cis[i] = make([]uint32, pars.Words)
	}
	rng := aig.NewAigRandomSource()
	rng.Uint32()
	randomInputs := func() {
		for _, i := range p.CIOrder {
			if i >= p.PIs {
				continue
			}
			for w := pars.Words - 1; w >= 0; w-- {
				cis[i][w] = rng.Uint32()
			}
		}
	}
	randomInputs()
	result := &SimulationResult{}
	start := time.Now()
	for frame := 0; frame < pars.Iterations; frame++ {
		outputs, err := p.Round(cis)
		if err != nil {
			return nil, err
		}
		result.Frames = frame + 1
		if pars.CheckMiter {
			if output, pattern, found := Fsim_ManCheckPos(outputs, p.POs); found {
				model, err := Fsim_ManGenerateCounter(g, frame, output, pars.Words, pattern, p.CIOrder)
				if err != nil {
					return nil, err
				}
				if err = g.SetSequentialCounterexample(model); err != nil {
					return nil, err
				}
				result.Found, result.Counterexample = true, model
				break
			}
		}
		if pars.TimeLimit > 0 && time.Since(start) >= pars.TimeLimit {
			result.TimedOut = true
			break
		}
		for r := p.PIs; r < p.CIs; r++ {
			copy(cis[r], outputs[p.POs+r-p.PIs])
		}
		randomInputs()
	}
	if pars.Verbose {
		fmt.Printf("FSIM: frames=%d frontier=%d objects=%d failed=%t elapsed=%s\n", result.Frames, p.Frontier, len(p.ops), result.Found, time.Since(start))
	}
	return result, nil
}
func Fsim_ManSimulate(g *aig.Graph, pars *Fsim_ParSim_t) (bool, error) {
	result, err := RunSimulation(g, pars)
	if err != nil {
		return false, err
	}
	return result.Found, nil
}
func Fsim_ManSimulateRound(p *Manager, cis [][]uint32) ([][]uint32, error)  { return p.Round(cis) }
func Fsim_ManSimulateRound2(p *Manager, cis [][]uint32) ([][]uint32, error) { return p.Round(cis) }
func Fsim_ManSimulateRoundTest(p *Manager, cis [][]uint32) error {
	compressed := p.compressed
	if err := Fsim_ManFront(p, false); err != nil {
		return err
	}
	left, err := p.Round(cis)
	if err != nil {
		return err
	}
	if err = Fsim_ManFront(p, true); err != nil {
		return err
	}
	right, err := p.Round(cis)
	if err != nil {
		return err
	}
	for i := range left {
		for w := range left[i] {
			if left[i][w] != right[i][w] {
				return fmt.Errorf("fsim: compressed round differs")
			}
		}
	}
	return Fsim_ManFront(p, compressed)
}
