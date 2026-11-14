package aig

import (
	"fmt"
	"io"
	"math/bits"
	"os"
)

// Aig_ManPack_t packs compatible partial assignments into 64 simulation lanes.
// AddPattern reserves lane zero for the all-zero input assignment. Literal
// variable numbers refer to CI positions, not graph object IDs.
type Aig_ManPack_t struct {
	pAig                *Graph
	vSigns              []uint64
	vPiPats             []uint64
	vPiCare             []uint64
	nPatTotal, nPatSkip int32
}

func Aig_Word6CountOnes(word uint64) int32 { return int32(bits.OnesCount64(word)) }

// ABC's historical name includes zero: this tests whether at most one bit is set.
func Aig_Word6HasOneBit(word uint64) int32 {
	if word&(word-1) == 0 {
		return 1
	}
	return 0
}
func Aig_ManPackAlloc(graph *Graph) *Aig_ManPack_t {
	if graph == nil {
		panic("aig: nil packing graph")
	}
	return &Aig_ManPack_t{pAig: graph, vSigns: make([]uint64, len(graph.nodes)), vPiPats: make([]uint64, len(graph.inputs)), vPiCare: make([]uint64, len(graph.inputs))}
}
func (p *Aig_ManPack_t) checkInputs() {
	if p == nil || p.pAig == nil {
		panic("aig: packing manager stopped")
	}
	if len(p.vPiPats) != len(p.pAig.inputs) {
		panic("aig: input interface changed while packing")
	}
}
func Aig_ManPackCountCares(p *Aig_ManPack_t) int32 {
	p.checkInputs()
	var count int32
	for _, care := range p.vPiCare {
		count += Aig_Word6CountOnes(care)
	}
	return count
}

// WriteCare makes the old optional bitset diagnostic usable with any Go writer.
func (p *Aig_ManPack_t) WriteCare(writer io.Writer) {
	p.checkInputs()
	for _, care := range p.vPiCare {
		fmt.Fprintf(writer, "%064b\n", care)
	}
}

// C leaves this diagnostic disabled; WriteCare is the explicit Go alternative.
func Aig_ManPackPrintCare(p *Aig_ManPack_t) { p.checkInputs() }
func Aig_ManPackFree(p *Aig_ManPack_t) {
	if p == nil {
		return
	}
	p.checkInputs()
	percent := 0.0
	if len(p.vPiPats) != 0 {
		percent = 100 * float64(Aig_ManPackCountCares(p)) / float64(len(p.vPiPats)*64)
	}
	fmt.Printf("Patterns: Total = %6d. Skipped = %6d. Cares = %6.2f %%  \n", p.nPatTotal, p.nPatSkip, percent)
	*p = Aig_ManPack_t{}
}

// Aig_ManPackSetRandom reinitializes all input words, as in C. Packing care
// masks remain separate; Start calls this before inserting any patterns.
func Aig_ManPackSetRandom(p *Aig_ManPack_t) {
	p.checkInputs()
	for i := range p.vPiPats {
		word := uint64(Aig_ManRandom(0))<<32 | uint64(Aig_ManRandom(0))
		p.vPiPats[i] = word << 1
	}
}
func Aig_ManPackSimulate(p *Aig_ManPack_t) {
	p.checkInputs()
	g := p.pAig
	if len(p.vSigns) != len(g.nodes) {
		p.vSigns = make([]uint64, len(g.nodes))
	}
	state := make([]uint8, len(g.nodes))
	p.vSigns[0] = ^uint64(0)
	state[0] = 2
	for i, input := range g.inputs {
		p.vSigns[input.NodeID()] = p.vPiPats[i]
		state[input.NodeID()] = 2
	}
	var simulate func(Edge) uint64
	simulate = func(edge Edge) uint64 {
		if !g.valid(edge) {
			panic("aig: invalid packed simulation edge")
		}
		id := edge.NodeID()
		if state[id] == 1 {
			panic("aig: cyclic packed simulation graph")
		}
		if state[id] != 2 {
			state[id] = 1
			n := g.nodes[id]
			var value uint64
			switch n.Kind {
			case And:
				value = simulate(n.Fanins[0]) & simulate(n.Fanins[1])
			case ExclusiveOr:
				value = simulate(n.Fanins[0]) ^ simulate(n.Fanins[1])
			case Buffer, Output:
				value = simulate(n.Fanins[0])
			default:
				panic("aig: unbound simulation terminal")
			}
			p.vSigns[id] = value
			state[id] = 2
		}
		value := p.vSigns[id]
		if edge.Complemented() {
			value = ^value
		}
		return value
	}
	for id := range g.nodes {
		if !g.deleted[id] {
			simulate(Edge(id) << 1)
		}
	}
}
func (p *Aig_ManPack_t) histogram() (counts [33]int, total int) {
	p.checkInputs()
	for id, n := range p.pAig.nodes {
		if p.pAig.deleted[id] || (n.Kind != And && n.Kind != ExclusiveOr) {
			continue
		}
		ones := bits.OnesCount64(p.vSigns[id])
		if ones > 32 {
			ones = 64 - ones
		}
		counts[ones]++
		total++
	}
	return
}
func (p *Aig_ManPack_t) WriteStats(writer io.Writer) {
	counts, total := p.histogram()
	sum := 0
	for i, count := range counts {
		sum += count
		percent, cumulative := 0.0, 0.0
		if total != 0 {
			percent = 100 * float64(count) / float64(total)
			cumulative = 100 * float64(sum) / float64(total)
		}
		fmt.Fprintf(writer, "%2d : %6d  %6.1f %%%6.1f %%\n", i, count, percent, cumulative)
	}
}
func Aig_ManPackPrintStats(p *Aig_ManPack_t) { p.WriteStats(os.Stdout) }
func Aig_ManPackConstNodes(p *Aig_ManPack_t) []int32 {
	p.checkInputs()
	var nodes []int32
	for id, n := range p.pAig.nodes {
		if p.pAig.deleted[id] || (n.Kind != And && n.Kind != ExclusiveOr) {
			continue
		}
		signature := p.vSigns[id]
		if Aig_Word6HasOneBit(signature) != 0 || Aig_Word6HasOneBit(^signature) != 0 {
			nodes = append(nodes, int32(id))
		}
	}
	return nodes
}
func Aig_ManPackAddPatternTry(p *Aig_ManPack_t, bit int32, literals []int32) int32 {
	p.checkInputs()
	if bit < 0 || bit >= 64 {
		panic("aig: simulation lane out of range")
	}
	mask := uint64(1) << uint(bit)
	// Validate the whole pattern before changing storage, including contradictory
	// duplicate literals that the original two-loop implementation overlooked.
	required := make(map[int]bool, len(literals))
	for _, literal := range literals {
		if literal < 0 || int(literal>>1) >= len(p.vPiPats) {
			panic("aig: pattern literal outside CI interface")
		}
		variable := int(literal >> 1)
		value := literal&1 == 0
		if previous, exists := required[variable]; exists && previous != value {
			return 0
		}
		required[variable] = value
		if p.vPiCare[variable]&mask != 0 && (p.vPiPats[variable]&mask != 0) != value {
			return 0
		}
	}
	for variable, value := range required {
		p.vPiCare[variable] |= mask
		if value {
			p.vPiPats[variable] |= mask
		} else {
			p.vPiPats[variable] &= ^mask
		}
	}
	return 1
}
func Aig_ManPackAddPattern(p *Aig_ManPack_t, literals []int32) {
	p.checkInputs()
	packed := false
	for bit := int32(1); bit < 64; bit++ {
		if Aig_ManPackAddPatternTry(p, bit, literals) != 0 {
			packed = true
			break
		}
	}
	if !packed {
		p.nPatSkip++
	}
	p.nPatTotal++
}
func Aig_ManPackStart(graph *Graph) *Aig_ManPack_t {
	p := Aig_ManPackAlloc(graph)
	Aig_ManPackSetRandom(p)
	Aig_ManPackSimulate(p)
	Aig_ManPackPrintStats(p)
	return p
}
func Aig_ManPackStop(p *Aig_ManPack_t) {
	Aig_ManPackSimulate(p)
	Aig_ManPackPrintStats(p)
	Aig_ManPackFree(p)
}

// Snapshot accessors keep callers from accidentally invalidating packing state.
func (p *Aig_ManPack_t) InputWords() []uint64                  { return append([]uint64(nil), p.vPiPats...) }
func (p *Aig_ManPack_t) CareWords() []uint64                   { return append([]uint64(nil), p.vPiCare...) }
func (p *Aig_ManPack_t) Signatures() []uint64                  { return append([]uint64(nil), p.vSigns...) }
func (p *Aig_ManPack_t) PatternCounts() (total, skipped int32) { return p.nPatTotal, p.nPatSkip }

// SetInputWord seeds uncared-for lanes without changing already packed cubes.
func (p *Aig_ManPack_t) SetInputWord(input int, word uint64) {
	p.checkInputs()
	p.vPiPats[input] = word&^p.vPiCare[input] | p.vPiPats[input]&p.vPiCare[input]
}
