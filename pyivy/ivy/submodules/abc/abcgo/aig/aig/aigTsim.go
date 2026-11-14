// Ternary sequential simulation, ported from ABC aigTsim.c (Alan Mishchenko).
package aig

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	TSI_MAX_ROUNDS       = 1000
	TSI_ONE_SERIES       = 300
	AIG_XVS0       int32 = 1
	AIG_XVS1       int32 = 2
	AIG_XVSX       int32 = 3
)

func validTernary(value int32) {
	if value < AIG_XVS0 || value > AIG_XVSX {
		panic("aig: invalid ternary value")
	}
}
func Aig_ObjSetXsim(p *Graph, e Edge, value int32) { validTernary(value); p.node(e).Xsim = value }
func Aig_ObjGetXsim(p *Graph, e Edge) int32        { return p.node(e).Xsim }
func Aig_XsimInv(value int32) int32 {
	validTernary(value)
	if value == AIG_XVSX {
		return value
	}
	return 3 - value
}
func Aig_XsimAnd(a, b int32) int32 {
	validTernary(a)
	validTernary(b)
	if a == AIG_XVS0 || b == AIG_XVS0 {
		return AIG_XVS0
	}
	if a == AIG_XVSX || b == AIG_XVSX {
		return AIG_XVSX
	}
	return AIG_XVS1
}
func Aig_XsimRand2() int32 { return 1 + int32(Aig_ManRandom(0)&1) }
func Aig_XsimRand3() int32 {
	for {
		if value := int32(Aig_ManRandom(0) & 3); value != 0 {
			return value
		}
	}
}
func Aig_ObjGetXsimFanin0(p *Graph, e Edge) int32 { return ternaryEdgeValue(p, p.node(e).Fanins[0]) }
func Aig_ObjGetXsimFanin1(p *Graph, e Edge) int32 { return ternaryEdgeValue(p, p.node(e).Fanins[1]) }
func ternaryEdgeValue(p *Graph, e Edge) int32 {
	value := Aig_ObjGetXsim(p, e)
	if e.Complemented() {
		return Aig_XsimInv(value)
	}
	validTernary(value)
	return value
}
func Aig_XsimPrint(w io.Writer, value int32) {
	validTernary(value)
	fmt.Fprint(w, []string{"", "0", "1", "x"}[value])
}

// States contain only packed two-bit values. Go's map replaces the C allocator
// and pointer chain embedded after each state buffer.
type Aig_Tsi_t struct {
	pAig   *Graph
	states [][]uint32
	seen   map[string]struct{}
}

func Aig_TsiStart(p *Graph) *Aig_Tsi_t {
	if p.nRegs < 0 || int(p.nRegs) > len(p.inputs) || int(p.nRegs) > len(p.outputs) {
		panic("aig: invalid sequential interface")
	}
	return &Aig_Tsi_t{pAig: p, seen: make(map[string]struct{})}
}
func Aig_TsiStop(p *Aig_Tsi_t)  { p.states = nil; p.seen = nil }
func (p *Aig_Tsi_t) words() int { return (2*int(p.pAig.nRegs) + 31) / 32 }
func ternaryStateKey(state []uint32, n int32) string {
	if n < 0 || int(n) > len(state) {
		panic("aig: invalid state word count")
	}
	key := make([]byte, int(n)*4)
	for i, x := range state[:n] {
		binary.LittleEndian.PutUint32(key[i*4:], x)
	}
	return string(key)
}
func Aig_TsiStateLookup(p *Aig_Tsi_t, state []uint32, n int32) int32 {
	_, ok := p.seen[ternaryStateKey(state, n)]
	if ok {
		return 1
	}
	return 0
}
func Aig_TsiStateInsert(p *Aig_Tsi_t, state []uint32, n int32) {
	key := ternaryStateKey(state, n)
	if _, ok := p.seen[key]; ok {
		panic("aig: duplicate ternary state")
	}
	p.seen[key] = struct{}{}
}
func Aig_TsiStateNew(p *Aig_Tsi_t) []uint32 {
	state := make([]uint32, p.words())
	p.states = append(p.states, state)
	return state
}
func ternaryStateValue(state []uint32, index int) int32 {
	value := int32((state[index/16] >> uint(2*(index%16))) & 3)
	validTernary(value)
	return value
}
func (p *Aig_Tsi_t) WriteState(w io.Writer, state []uint32) {
	counts := [4]int{}
	for i := 0; i < int(p.pAig.nRegs); i++ {
		value := ternaryStateValue(state, i)
		counts[value]++
		Aig_XsimPrint(w, value)
	}
	fmt.Fprintf(w, " (0=%5d, 1=%5d, x=%5d)\n", counts[AIG_XVS0], counts[AIG_XVS1], counts[AIG_XVSX])
}
func Aig_TsiStatePrint(p *Aig_Tsi_t, state []uint32) { p.WriteState(os.Stdout, state) }
func Aig_TsiStateCount(p *Aig_Tsi_t, state []uint32) int32 {
	var count int32
	for i := 0; i < int(p.pAig.nRegs); i++ {
		if ternaryStateValue(state, i) != AIG_XVSX {
			count++
		}
	}
	return count
}
func Aig_TsiStateOrAll(p *Aig_Tsi_t, state []uint32) {
	if len(state) != p.words() {
		panic("aig: wrong state width")
	}
	for _, old := range p.states {
		for i, x := range old {
			state[i] |= x
		}
	}
}

func ternarySimulateFrame(p *Graph) {
	done, active := make([]bool, len(p.nodes)), make([]bool, len(p.nodes))
	for _, e := range p.inputs {
		done[e.NodeID()] = true
	}
	p.nodes[0].Xsim = AIG_XVS1
	done[0] = true
	var visit func(Edge) int32
	visit = func(e Edge) int32 {
		if !p.valid(e) {
			panic("aig: invalid simulation edge")
		}
		id := e.NodeID()
		if !done[id] {
			if active[id] {
				panic("aig: combinational cycle during simulation")
			}
			active[id] = true
			n := p.node(e)
			switch n.Kind {
			case And:
				n.Xsim = Aig_XsimAnd(visit(n.Fanins[0]), visit(n.Fanins[1]))
			case ExclusiveOr:
				a, b := visit(n.Fanins[0]), visit(n.Fanins[1])
				if a == AIG_XVSX || b == AIG_XVSX {
					n.Xsim = AIG_XVSX
				} else if a == b {
					n.Xsim = AIG_XVS0
				} else {
					n.Xsim = AIG_XVS1
				}
			case Buffer, Output:
				n.Xsim = visit(n.Fanins[0])
			case Constant:
				n.Xsim = AIG_XVS1
			case Input:
				validTernary(n.Xsim)
			default:
				panic("aig: unknown simulation node")
			}
			active[id] = false
			done[id] = true
		}
		return ternaryEdgeValue(p, e)
	}
	for id := range p.nodes {
		if !p.deleted[id] {
			visit(Edge(id) << 1)
		}
	}
}

// Aig_ManTernarySimulate returns an input mapping only when a register is
// constant across the observed state sequence. All latches begin at zero.
func Aig_ManTernarySimulate(p *Graph, verbose, veryVerbose int32) []Edge {
	return ternarySimulate(p, TSI_MAX_ROUNDS, TSI_ONE_SERIES, verbose, veryVerbose)
}
func ternarySimulate(p *Graph, maxRounds, oneSeries int, verbose, veryVerbose int32) []Edge {
	manager := Aig_TsiStart(p)
	defer Aig_TsiStop(manager)
	if p.nRegs == 0 {
		return nil
	}
	if maxRounds <= 0 || oneSeries < 0 {
		panic("aig: invalid simulation round limits")
	}
	for _, e := range p.inputs {
		p.node(e).Xsim = AIG_XVSX
	}
	p.nodes[0].Xsim = AIG_XVS1
	piCount := len(p.inputs) - int(p.nRegs)
	poCount := len(p.outputs) - int(p.nRegs)
	for _, e := range p.inputs[piCount:] {
		p.node(e).Xsim = AIG_XVS0
	}
	round := 0
	for ; round < maxRounds; round++ {
		state := Aig_TsiStateNew(manager)
		for i, e := range p.inputs[piCount:] {
			value := p.node(e).Xsim
			validTernary(value)
			state[i/16] |= uint32(value) << uint(2*(i%16))
		}
		if veryVerbose != 0 {
			fmt.Printf("%3d : ", round)
			Aig_TsiStatePrint(manager, state)
		}
		if Aig_TsiStateLookup(manager, state, int32(len(state))) != 0 {
			break
		}
		Aig_TsiStateInsert(manager, state, int32(len(state)))
		ternarySimulateFrame(p)
		next := make([]int32, p.nRegs)
		for i, e := range p.outputs[poCount:] {
			next[i] = ternaryEdgeValue(p, e)
		}
		for i, e := range p.inputs[piCount:] {
			if round < oneSeries {
				p.node(e).Xsim = next[i]
			} else if p.node(e).Xsim != next[i] {
				p.node(e).Xsim = AIG_XVSX
			}
		}
	}
	if round == maxRounds {
		fmt.Printf("Aig_ManTernarySimulate(): Did not reach a fixed point after %d iterations (not a bug).\n", maxRounds)
		return nil
	}
	state := manager.states[0]
	Aig_TsiStateOrAll(manager, state)
	count := Aig_TsiStateCount(manager, state)
	if count == 0 {
		if verbose != 0 {
			fmt.Printf("Detected 0 constants after %d iterations of ternary simulation.\n", round)
		}
		return nil
	}
	mapping := append([]Edge(nil), p.inputs[:piCount]...)
	for i, e := range p.inputs[piCount:] {
		switch ternaryStateValue(state, i) {
		case AIG_XVS0:
			mapping = append(mapping, p.Const0())
		case AIG_XVS1:
			mapping = append(mapping, p.Const1())
		case AIG_XVSX:
			mapping = append(mapping, e)
		}
	}
	if verbose != 0 {
		fmt.Printf("Detected %d constants after %d iterations of ternary simulation.\n", count, round)
	}
	return mapping
}

var ternaryHashPrimes = [...]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
	1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
	2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
	2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
	3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
	3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
	4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
	4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
	6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
	6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
	7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
	8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

// Aig_TsiStateHash retains ABC's 32-bit hash for callers that persist hashes;
// the simulation itself uses immutable map keys instead of manual hash bins.
func Aig_TsiStateHash(state []uint32, nWords, nBins int32) int32 {
	if nWords < 0 || int(nWords) > len(state) || nBins <= 0 {
		panic("aig: invalid state hash dimensions")
	}
	var hash uint32
	for i, value := range state[:nWords] {
		hash ^= value * ternaryHashPrimes[i&127]
	}
	return int32(hash % uint32(nBins))
}

// MappingSimulator is the boundary to optional symbolic/multivalued engines.
// Keeping it injected avoids an aig -> saig -> aig package cycle.
type MappingSimulator func(*Graph, int32, int32, int32, int32) []Edge

// Aig_ManConstReduce repeatedly applies ternary mappings and sequential
// reachability cleanup. The built-in multivalued engine can be overridden.
func Aig_ManConstReduce(p *Graph, useMV, framesSymbolic, framesSatur, verbose, veryVerbose int32, engines ...MappingSimulator) *Graph {
	var engine MappingSimulator
	if useMV != 0 {
		if len(engines) == 0 {
			engine = Aig_MvManSimulate
		} else if len(engines) == 1 && engines[0] != nil {
			engine = engines[0]
		} else {
			panic("aig: invalid multivalued MappingSimulator override")
		}
	}
	for p.nRegs > 0 {
		var mapping []Edge
		if useMV != 0 {
			mapping = engine(p, framesSymbolic, framesSatur, verbose, veryVerbose)
		} else {
			mapping = Aig_ManTernarySimulate(p, verbose, veryVerbose)
		}
		if mapping == nil {
			break
		}
		next := ternaryRemapAndCleanup(p, mapping)
		if verbose != 0 {
			Aig_ManReportImprovement(p, next)
		}
		if next.nRegs >= p.nRegs {
			panic("aig: constant mapping made no sequential progress")
		}
		Aig_ManStop(p)
		p = next
	}
	return p
}

func ternaryRemapAndCleanup(p *Graph, mapping []Edge) *Graph {
	result := Aig_ManRemap(p, mapping)
	Aig_ManSeqCleanup(result)
	return result
}
