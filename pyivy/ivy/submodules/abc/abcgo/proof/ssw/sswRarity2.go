package ssw

// Rarity-guided simulation and class filtering from sswRarity2.c.
import (
	"fmt"
	"math"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type sswRar2Man struct {
	graph                          *aig.Graph
	words, frames, binSize, groups int
	verbose                        bool
	sim                            *Ssw_Sml_t
	classes                        *Ssw_Cla_t
	inits                          *vec.Vec_Int_t
	rarity, groupValues            []int
	costs                          []float64
}

func sswRar2Start(g *aig.Graph, words, frames, bin int32, verbose bool) *sswRar2Man {
	if g == nil || bin <= 0 || g.RegisterCount() < bin {
		return nil
	}
	if words <= 0 || frames <= 0 || bin >= 31 {
		panic("ssw: invalid rarity simulation dimensions")
	}
	groups := int(g.RegisterCount() / bin)
	return &sswRar2Man{graph: g, words: int(words), frames: int(frames), binSize: int(bin), groups: groups, verbose: verbose, sim: Ssw_SmlStart(g, 0, frames, words), rarity: make([]int, groups*(1<<uint(bin))), groupValues: make([]int, groups), costs: make([]float64, int(words)*32)}
}
func (p *sswRar2Man) patternValues(pattern int) {
	clear(p.groupValues)
	outputs := p.graph.Outputs()
	regs := int(p.graph.RegisterCount())
	for i, out := range outputs[len(outputs)-regs:] {
		group := i / p.binSize
		if group >= p.groups {
			continue
		}
		words := Ssw_SmlSimInfo(p.sim, out)
		index := (p.frames-1)*p.words + pattern/32
		if words[index]&(uint32(1)<<uint(pattern%32)) != 0 {
			p.groupValues[group] |= 1 << uint(i%p.binSize)
		}
	}
}
func (p *sswRar2Man) updateCounters() {
	for pattern := range p.costs {
		p.patternValues(pattern)
		for group, value := range p.groupValues {
			p.rarity[group*(1<<uint(p.binSize))+value]++
		}
	}
}
func (p *sswRar2Man) transferPatterns() {
	for pattern := range p.costs {
		p.patternValues(pattern)
		cost := 0.0
		for group, value := range p.groupValues {
			count := p.rarity[group*(1<<uint(p.binSize))+value]
			if count <= 0 {
				panic("ssw: missing rarity count")
			}
			cost += 1 / (float64(count) * float64(count))
		}
		p.costs[pattern] = cost
	}
	vec.Vec_IntClear(p.inits)
	outputs := p.graph.Outputs()
	regs := int(p.graph.RegisterCount())
	for i := 0; i < p.words; i++ {
		best := -1
		score := math.Inf(-1)
		for j, cost := range p.costs {
			if cost > score {
				score = cost
				best = j
			}
		}
		if best < 0 {
			panic("ssw: no rarity pattern")
		}
		p.costs[best] = math.Inf(-1)
		for _, out := range outputs[len(outputs)-regs:] {
			data := Ssw_SmlSimInfo(p.sim, out)
			index := (p.frames-1)*p.words + best/32
			vec.Vec_IntPush(p.inits, int32((data[index]>>uint(best%32))&1))
		}
	}
}

// sswRar2StartingState returns the state entering the final trace frame.
func sswRar2StartingState(g *aig.Graph, cex *util.Abc_Cex_t) *vec.Vec_Int_t {
	regs, pis := int(g.RegisterCount()), g.InputCount()-int(g.RegisterCount())
	if cex == nil || int(cex.NRegs) != regs || int(cex.NPis) != pis || cex.IFrame < 0 || cex.NBits != int32(regs+pis*(int(cex.IFrame)+1)) || len(cex.PData)*32 < int(cex.NBits) {
		panic("ssw: invalid rarity counterexample")
	}
	bit := func(i int) bool { return cex.PData[i/32]&(uint32(1)<<uint(i%32)) != 0 }
	state := make([]bool, regs)
	for i := range state {
		state[i] = bit(i)
	}
	for f := 0; f < int(cex.IFrame); f++ {
		in := make([]bool, pis+regs)
		for i := 0; i < pis; i++ {
			in[i] = bit(regs + f*pis + i)
		}
		copy(in[pis:], state)
		out, err := g.Evaluate(in)
		if err != nil {
			panic(err)
		}
		copy(state, out[len(out)-regs:])
	}
	result := vec.Vec_IntAlloc(regs)
	for _, v := range state {
		if v {
			vec.Vec_IntPush(result, 1)
		} else {
			vec.Vec_IntPush(result, 0)
		}
	}
	return result
}
func sswRar2Run(g *aig.Graph, frames, words, bin, rounds, timeout int32, cex *util.Abc_Cex_t, latchOnly, verbose int32, filter bool) int32 {
	if g == nil || g.RegisterCount() <= 0 || aig.Aig_ManConstrNum(g) != 0 {
		panic("ssw: rarity requires an unconstrained sequential AIG")
	}
	if aig.Aig_ManNodeNum(g) == 0 {
		return -1
	}
	aig.Aig_ManRandom(1)
	p := sswRar2Start(g, words, frames, bin, verbose != 0)
	if p == nil {
		return -1
	}
	defer Ssw_SmlStop(p.sim)
	p.inits = vec.Vec_IntStart(int(g.RegisterCount()))
	if cex != nil {
		p.inits = sswRar2StartingState(g, cex)
	}
	base := append([]int32(nil), vec.Vec_IntArray(p.inits)...)
	for i := int32(1); i < words; i++ {
		for _, v := range base {
			vec.Vec_IntPush(p.inits, v)
		}
	}
	Ssw_SmlInitializeSpecial(p.sim, p.inits)
	if filter {
		if aig.Aig_ManHasReprs(g) {
			p.classes = Ssw_ClassesPrepareFromReprs(g)
		} else {
			p.classes = Ssw_ClassesPrepareSimple(g, latchOnly, 0)
		}
		defer Ssw_ClassesStop(p.classes)
		Ssw_ClassesSetData(p.classes, p.sim, nil, func(data any, e aig.Edge) int32 { return Ssw_SmlObjIsConstWord(data.(*Ssw_Sml_t), e) }, func(data any, a, b aig.Edge) int32 { return Ssw_SmlObjsAreEqualWord(data.(*Ssw_Sml_t), a, b) })
	}
	started := time.Now()
	for r := int32(0); r < rounds; r++ {
		if filter && Ssw_ClassesCand1Num(p.classes) == 0 && Ssw_ClassesClassNum(p.classes) == 0 {
			fmt.Println("All equivalences are refined away.")
			break
		}
		Ssw_SmlSimulateOne(p.sim)
		if !filter && Ssw_SmlCheckNonConstOutputs(p.sim) != 0 {
			fmt.Printf("Simulation asserted a PO in frame f: %d <= f < %d.\n", r*frames, (r+1)*frames)
			return 0
		}
		if filter {
			Ssw_ClassesRefineConst1(p.classes, 1)
			Ssw_ClassesRefine(p.classes, 1)
			if verbose != 0 {
				fmt.Printf("Round %3d: ", r)
				Ssw_ClassesPrint(p.classes, 0)
			}
		}
		p.updateCounters()
		p.transferPatterns()
		Ssw_SmlInitializeSpecial(p.sim, p.inits)
		if timeout > 0 && time.Since(started) > time.Duration(timeout)*time.Second {
			fmt.Printf("Reached timeout (%d seconds).\n", timeout)
			break
		}
	}
	return -1
}
func Ssw_RarSimulate2(g *aig.Graph, frames, words, bin, rounds, timeout, verbose int32) int32 {
	return sswRar2Run(g, frames, words, bin, rounds, timeout, nil, 0, verbose, false)
}
func Ssw_RarSignalFilter2(g *aig.Graph, frames, words, bin, rounds, timeout int32, cex *util.Abc_Cex_t, latchOnly, verbose int32) int32 {
	return sswRar2Run(g, frames, words, bin, rounds, timeout, cex, latchOnly, verbose, true)
}
func Ssw_RarSignalFilterGia2(g *gia.Gia_Man_t, frames, words, bin, rounds, timeout int32, cex *util.Abc_Cex_t, latchOnly, verbose int32) int32 {
	graph := gia.Gia_ManToAigSimple(g)
	if g.HasRepresentatives() {
		gia.Gia_ManReprToAigRepr2(graph, g)
		g.ClearRepresentatives()
	}
	result := Ssw_RarSignalFilter2(graph, frames, words, bin, rounds, timeout, cex, latchOnly, verbose)
	gia.Gia_ManReprFromAigRepr(graph, g)
	return result
}
