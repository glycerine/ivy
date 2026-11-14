// Symbolic/multivalued sequential simulation, ported from saigSimMv.c.
package aig

import (
	"encoding/binary"
	"fmt"
)

const (
	SAIG_DIFF_VALUES       = 8
	SAIG_UNDEF_VALUE int32 = 0x1ffffffe
)

type mvPair struct{ a, b int32 }

// MvManager stores symbolic functions as complemented integer literals. A
// distinct literal is allocated for each PI at each symbolic frame. Retired
// values use one distinguished unknown value whose complement stays unknown.
// The graph and sequential interface must remain fixed while the manager lives.
type MvManager struct {
	graph       *Graph
	values      []int32
	frameOrder  []Edge
	nextState   []int32
	nextLiteral int32
	hashed      map[mvPair]int32
	states      [][]int32
	seen        map[string]int32
	skip        []int32
}

func mvNot(value int32) int32 {
	if value == SAIG_UNDEF_VALUE {
		return value
	}
	return value ^ 1
}
func Aig_MvManStart(p *Graph, satur int32) *MvManager {
	if p.nRegs <= 0 || int(p.nRegs) > len(p.inputs) || int(p.nRegs) > len(p.outputs) || satur < 1 || satur > (1<<30)-1 {
		panic("aig: invalid multivalued simulation interface")
	}
	// Preserve object order on source-topological graphs, resolving forward
	// dependencies once for Go graphs whose objects have been rewired.
	order := p.TopologicalObjects()
	latches := make(map[Edge]bool, p.nRegs)
	for _, lo := range p.inputs[len(p.inputs)-int(p.nRegs):] {
		latches[lo] = true
	}
	frameOrder := order[:0]
	for _, e := range order {
		if !latches[e] {
			frameOrder = append(frameOrder, e)
		}
	}
	return &MvManager{
		graph: p, values: make([]int32, len(p.nodes)), frameOrder: frameOrder,
		nextState: make([]int32, p.nRegs), nextLiteral: 2,
		hashed: map[mvPair]int32{}, seen: map[string]int32{},
	}
}
func Aig_MvManStop(p *MvManager) { *p = MvManager{} }

// Literal IDs alone encode symbolic identities; the hash map owns AND keys.
// No parallel node-capacity/level arrays are needed by this simulation.
func (p *MvManager) newLiteral() int32 {
	if p.nextLiteral >= SAIG_UNDEF_VALUE {
		panic("aig: symbolic literal space exhausted")
	}
	literal := p.nextLiteral
	p.nextLiteral += 2
	return literal
}
func (p *MvManager) and(a, b int32, first bool) int32 {
	if a == b {
		return a
	}
	if a == mvNot(b) {
		return 1
	}
	if a < 2 {
		if a == 0 {
			return b
		}
		return 1
	}
	if b < 2 {
		if b == 0 {
			return a
		}
		return 1
	}
	if a == SAIG_UNDEF_VALUE || b == SAIG_UNDEF_VALUE {
		return SAIG_UNDEF_VALUE
	}
	if !first {
		return SAIG_UNDEF_VALUE
	}
	if a > b {
		a, b = b, a
	}
	key := mvPair{a, b}
	if value, ok := p.hashed[key]; ok {
		return value
	}
	value := p.newLiteral()
	p.hashed[key] = value
	return value
}
func (p *MvManager) xor(a, b int32, first bool) int32 {
	x := p.and(a, mvNot(b), first)
	y := p.and(mvNot(a), b, first)
	return mvNot(p.and(mvNot(x), mvNot(y), first))
}
func Aig_MvSimulateFrame(p *MvManager, first, verbose int32) {
	g := p.graph
	value := func(e Edge) int32 {
		result := p.values[e.NodeID()]
		if e.Complemented() {
			return mvNot(result)
		}
		return result
	}
	for _, e := range p.frameOrder {
		id := e.NodeID()
		n := g.node(e)
		switch n.Kind {
		case Constant:
			p.values[id] = 0
		case Input:
			if first != 0 {
				p.values[id] = p.newLiteral()
			} else {
				p.values[id] = SAIG_UNDEF_VALUE
			}
		case And:
			p.values[id] = p.and(value(n.Fanins[0]), value(n.Fanins[1]), first != 0)
		case ExclusiveOr:
			p.values[id] = p.xor(value(n.Fanins[0]), value(n.Fanins[1]), first != 0)
		case Buffer, Output:
			p.values[id] = value(n.Fanins[0])
		default:
			panic("aig: invalid symbolic node")
		}
	}
	for i, e := range g.outputs[len(g.outputs)-int(g.nRegs):] {
		p.nextState[i] = value(e)
	}
	for i, e := range g.inputs[len(g.inputs)-int(g.nRegs):] {
		p.values[e.NodeID()] = p.nextState[i]
	}
}
func (p *MvManager) currentState() []int32 {
	state := make([]int32, p.graph.nRegs)
	for i, e := range p.graph.inputs[len(p.graph.inputs)-len(state):] {
		state[i] = p.values[e.NodeID()]
	}
	return state
}
func mvStateKey(values []int32) string {
	bytes := make([]byte, 4*len(values))
	for i, value := range values {
		binary.LittleEndian.PutUint32(bytes[i*4:], uint32(value))
	}
	return string(bytes)
}
func Aig_MvSaveState(p *MvManager) int32 {
	state := p.currentState()
	key := mvStateKey(state)
	if index, ok := p.seen[key]; ok {
		return index
	}
	p.seen[key] = int32(len(p.states)) + 1
	p.states = append(p.states, state)
	return -1
}
func Aig_MvPrintState(iter int32, p *MvManager) {
	fmt.Printf("%3d : ", iter)
	for _, value := range p.currentState() {
		if value == SAIG_UNDEF_VALUE {
			fmt.Print("    *")
		} else {
			fmt.Printf("%5d", value)
		}
	}
	fmt.Println()
}
func Aig_MvManFindXFlops(p *MvManager) []int32 {
	result := make([]int32, p.graph.nRegs)
	for _, state := range p.states {
		for i, value := range state {
			if value == SAIG_UNDEF_VALUE {
				result[i] = 1
			}
		}
	}
	return result
}
func Aig_MvManFindConstBinaryFlops(p *MvManager, binary *[]int32) []int32 {
	var constants, known []int32
	for i := int32(0); i < p.graph.nRegs; i++ {
		constant, undefined := true, false
		for _, state := range p.states {
			if state[i] != 1 {
				constant = false
			}
			if state[i] == SAIG_UNDEF_VALUE {
				undefined = true
				break
			}
		}
		if undefined {
			continue
		}
		if constant {
			constants = append(constants, i)
		} else {
			known = append(known, i)
		}
	}
	if binary != nil {
		*binary = known
	}
	return constants
}
func Aig_MvManCheckOscilator(p *MvManager, flop int32) int32 {
	if flop < 0 || flop >= p.graph.nRegs {
		panic("aig: invalid oscillator register")
	}
	values := make([]int32, len(p.states))
	for i, state := range p.states {
		values[i] = state[flop]
	}
	if len(values) == 0 || values[0] != 1 {
		panic("aig: oscillator history must start at zero")
	}
	// Follow ABC's candidate-period search: skip an initial zero run, then the
	// following nonzero run, and compare the resulting repeated prefix.
	for period := 0; period < len(values)/2; period++ {
		for period < len(values) && values[period] == 1 {
			period++
		}
		if period == len(values) {
			break
		}
		for period < len(values) && values[period] != 1 {
			period++
		}
		if period == len(values) {
			break
		}
		i := period
		for ; i < len(values); i++ {
			if values[i-period] != values[i] {
				break
			}
		}
		if i == len(values) {
			return int32(period)
		}
	}
	return 0
}
func Aig_MvManFindOscilators(p *MvManager, constant *[]int32) []int32 {
	var binary []int32
	zero := Aig_MvManFindConstBinaryFlops(p, &binary)
	if constant != nil {
		*constant = zero
	}
	var result []int32
	for _, index := range binary {
		if Aig_MvManCheckOscilator(p, index) != 0 {
			result = append(result, index)
		}
	}
	return result
}
func Aig_MvManCreateNextSkip(p *MvManager) []int32 {
	var constants []int32
	oscillators := Aig_MvManFindOscilators(p, &constants)
	skip := make([]int32, p.graph.nRegs)
	for i := range skip {
		skip[i] = 1
	}
	for _, index := range constants {
		skip[index] = 0
	}
	for _, index := range oscillators {
		skip[index] = 0
	}
	return skip
}
func Aig_MvManDeriveMap(p *MvManager, verbose int32) []Edge {
	var binary []int32
	constants := Aig_MvManFindConstBinaryFlops(p, &binary)
	mapping := append([]Edge(nil), p.graph.inputs...)
	base := len(mapping) - int(p.graph.nRegs)
	for _, index := range constants {
		mapping[base+int(index)] = p.graph.Const0()
	}
	equivalent := 0
	for k, first := range binary {
		if first < 0 {
			continue
		}
		for j := k + 1; j < len(binary); j++ {
			second := binary[j]
			if second < 0 {
				continue
			}
			same := true
			for _, state := range p.states {
				if state[first] != state[second] {
					same = false
					break
				}
			}
			if same {
				mapping[base+int(second)] = p.graph.inputs[base+int(first)]
				binary[j] = -1
				equivalent++
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Detected %d const0 flops and %d pairs of equiv binary flops.\n", len(constants), equivalent)
	}
	if len(constants) == 0 && equivalent == 0 {
		return nil
	}
	return mapping
}
func Aig_MvManPostProcess(p *MvManager, state int32) {
	// C's pRegsUndef counters are never populated. Use actual histories for
	// this diagnostic, consistent with the mapping and retirement routines.
	known := Aig_MvManFindXFlops(p)
	count := 0
	for _, flag := range known {
		if flag == 0 {
			count++
		}
	}
	fmt.Printf("The number of registers that never became undef = %d. (Total = %d.)\n", count, p.graph.nRegs)
	groups := map[string][]int{}
	var keys []string
	for i, flag := range known {
		if flag != 0 {
			continue
		}
		history := make([]int32, len(p.states))
		for k, s := range p.states {
			history[k] = s[i]
		}
		key := mvStateKey(history)
		if _, exists := groups[key]; !exists {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], i)
	}
	for _, key := range keys {
		members := groups[key]
		fmt.Printf("FLOP %5d : (%3d) ", members[0], len(members))
		for i, s := range p.states {
			if int32(i) == state {
				fmt.Print(" # ")
			}
			if s[members[0]] == SAIG_UNDEF_VALUE {
				fmt.Print("*")
			} else {
				fmt.Print(s[members[0]])
			}
		}
		fmt.Println()
	}
}

func Aig_MvManSimulate(g *Graph, symbolic, satur, verbose, veryVerbose int32) []Edge {
	if symbolic < 1 || symbolic > satur || satur > (1<<30)-1 {
		panic("aig: symbolic frames must be between one and saturation frames")
	}
	if g.nRegs == 0 {
		return nil
	}
	p := Aig_MvManStart(g, satur)
	defer Aig_MvManStop(p)
	for _, lo := range g.inputs[len(g.inputs)-int(g.nRegs):] {
		p.values[lo.NodeID()] = 1
	}
	Aig_MvSaveState(p)
	if veryVerbose != 0 {
		Aig_MvPrintState(0, p)
	}
	for frame := int32(0); ; frame++ {
		if frame == satur {
			if verbose != 0 {
				fmt.Printf("Beginning to saturate simulation after %d frames\n", frame)
			}
			p.skip = Aig_MvManFindXFlops(p)
		}
		if frame == 2*satur {
			if verbose != 0 {
				fmt.Printf("Aggressively saturating simulation after %d frames\n", frame)
			}
			p.skip = Aig_MvManCreateNextSkip(p)
		}
		if p.skip != nil {
			for i, lo := range g.inputs[len(g.inputs)-int(g.nRegs):] {
				if p.skip[i] != 0 {
					p.values[lo.NodeID()] = SAIG_UNDEF_VALUE
				}
			}
		}
		first := int32(0)
		if frame < symbolic {
			first = 1
		}
		Aig_MvSimulateFrame(p, first, verbose)
		state := Aig_MvSaveState(p)
		if veryVerbose != 0 {
			Aig_MvPrintState(frame+1, p)
		}
		if state >= 0 {
			if verbose != 0 {
				fmt.Printf("Converged after %d frames with lasso in state %d. Cycle = %d.\n", frame+1, state-1, frame+2-state)
			}
			break
		}
	}
	return Aig_MvManDeriveMap(p, verbose)
}

func Aig_MvSimHash(state []uint32, nFlops, nBins int32) int32 {
	if nFlops < 0 || int(nFlops) > len(state) || nBins <= 0 {
		panic("aig: invalid multivalued hash dimensions")
	}
	primes := [16]uint32{1610612741, 805306457, 402653189, 201326611, 100663319, 50331653, 25165843, 12582917, 6291469, 3145739, 1572869, 786433, 393241, 196613, 98317, 49157}
	var hash uint32
	for i, value := range state[:nFlops] {
		hash ^= value * primes[i&15]
	}
	return int32(hash % uint32(nBins))
}
