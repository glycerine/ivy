package sbd

import (
	"encoding/binary"
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func (p *Network) ensureSimulation() {
	if p.SimWords == 0 {
		p.SimWords = 1
	}
	if p.SimWords < 0 {
		panic("negative simulation width")
	}
	count := len(p.Graph.Nodes()) * int(p.SimWords)
	if len(p.SimValues) < count {
		p.SimValues = append(p.SimValues, make([]uint64, count-len(p.SimValues))...)
	}
	if len(p.SimControls) < count {
		p.SimControls = append(p.SimControls, make([]uint64, count-len(p.SimControls))...)
	}
}
func Sbd_ObjSims(p *Network, id int32) []uint64 {
	p.ensureSimulation()
	start := int(id) * int(p.SimWords)
	return p.SimValues[start : start+int(p.SimWords)]
}
func Sbd_ObjCtrl(p *Network, id int32) []uint64 {
	p.ensureSimulation()
	start := int(id) * int(p.SimWords)
	return p.SimControls[start : start+int(p.SimWords)]
}
func (p *Network) randomWord() uint64 {
	if p.random == nil {
		p.random = aig.NewAigRandomSource()
	}
	high, low := p.random.Uint32(), p.random.Uint32()
	return uint64(high)<<32 | uint64(low)
}
func sbdSimLiteral(p *Network, edge aig.Edge, word int) uint64 {
	lit := sourceLiteral(edge)
	value := Sbd_ObjSims(p, lit>>1)[word]
	if lit&1 != 0 {
		value = ^value
	}
	return value
}

// This is the source's experimental backward heuristic, explicitly marked
// "This does not work" in C. It is retained as a diagnostic, not used to prove
// satisfiability or justify any rewrite.
func Sbd_GiaSimRoundBack2(p *Network) {
	p.ensureSimulation()
	for _, co := range p.Graph.Outputs() {
		sims := Sbd_ObjSims(p, int32(co.NodeID()))
		for w := range sims {
			sims[w] = p.randomWord()
		}
		driver := p.Graph.Node(co).Fanins[0]
		for w, value := range sims {
			if sourceLiteral(driver)&1 != 0 {
				value = ^value
			}
			Sbd_ObjSims(p, int32(driver.NodeID()))[w] = value
		}
	}
	edges := p.Graph.LiveEdges()
	for i := len(edges) - 1; i >= 0; i-- {
		edge := edges[i]
		node := p.Graph.Node(edge)
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			continue
		}
		sims := Sbd_ObjSims(p, int32(edge.NodeID()))
		a, b := Sbd_ObjSims(p, int32(node.Fanins[0].NodeID())), Sbd_ObjSims(p, int32(node.Fanins[1].NodeID()))
		random := p.randomWord()
		for w, value := range sims {
			a[w], b[w] = value|random, value|^random
			if node.Fanins[0].Complemented() {
				a[w] = ^a[w]
			}
			if node.Fanins[1].Complemented() {
				b[w] = ^b[w]
			}
		}
	}
}
func Sbd_GiaSatOne_rec(p *Network, id, value, first, pattern int32) int32 {
	if p.trial == nil {
		p.trial = map[int32]int32{}
	}
	if previous, ok := p.trial[id]; ok {
		if previous == value {
			return 1
		}
		return 0
	}
	p.trial[id] = value
	node := p.object(id)
	if node.Kind == aig.Input {
		data := Sbd_ObjSims(p, id)
		mask := uint64(1) << uint(pattern&63)
		if value != 0 {
			data[pattern>>6] |= mask
		} else {
			data[pattern>>6] &^= mask
		}
		return 1
	}
	if node.Kind != aig.And {
		panic("backward justification requires AND gates")
	}
	child := func(edge aig.Edge, desired int32) int32 {
		lit := sourceLiteral(edge)
		if lit>>1 == 0 {
			if int32(lit&1) == desired {
				return 1
			}
			return 0
		}
		return Sbd_GiaSatOne_rec(p, lit>>1, desired^(lit&1), first, pattern)
	}
	if value != 0 {
		if child(node.Fanins[0], 1) == 0 {
			return 0
		}
		return child(node.Fanins[1], 1)
	}
	index := 0
	if first == 0 {
		index = 1
	}
	return child(node.Fanins[index], 0)
}
func sbdPhases(p *Network) []int32 {
	values := make([]int32, len(p.Graph.Nodes()))
	read := func(edge aig.Edge) int32 { lit := sourceLiteral(edge); return values[lit>>1] ^ (lit & 1) }
	for _, edge := range p.Graph.LiveEdges() {
		node := p.Graph.Node(edge)
		switch node.Kind {
		case aig.And:
			values[edge.NodeID()] = read(node.Fanins[0]) & read(node.Fanins[1])
		case aig.ExclusiveOr:
			values[edge.NodeID()] = read(node.Fanins[0]) ^ read(node.Fanins[1])
		case aig.Buffer, aig.Output:
			values[edge.NodeID()] = read(node.Fanins[0])
		}
	}
	return values
}
func Sbd_GiaSatOne(p *Network, pairs *vec.Vec_Int_t) int32 {
	p.ensureSimulation()
	if vec.Vec_IntSize(pairs)%2 != 0 {
		panic("unpaired simulation requests")
	}
	phases := sbdPhases(p)
	pattern := int32(0)
	for i := 0; i < vec.Vec_IntSize(pairs); i += 2 {
		a, b := vec.Vec_IntEntry(pairs, i), vec.Vec_IntEntry(pairs, i+1)
		if b <= 0 {
			panic("invalid justification target")
		}
		success := false
		for first := int32(0); first < 2; first++ {
			p.trial = map[int32]int32{}
			if a == 0 {
				success = Sbd_GiaSatOne_rec(p, b, 1-phases[b], first, pattern) != 0
			} else {
				success = Sbd_GiaSatOne_rec(p, a, 1-phases[a], first, pattern) != 0 && Sbd_GiaSatOne_rec(p, b, phases[b], first, pattern) != 0
				if !success {
					p.trial = map[int32]int32{}
					success = Sbd_GiaSatOne_rec(p, a, phases[a], first, pattern) != 0 && Sbd_GiaSatOne_rec(p, b, 1-phases[b], first, pattern) != 0
				}
			}
			if success {
				break
			}
		}
		if success {
			pattern++
			if a == 0 {
				fmt.Print("c")
			} else {
				fmt.Print("e")
			}
		} else {
			fmt.Print(".")
		}
		if pattern == 64*p.SimWords-1 {
			break
		}
	}
	fmt.Println()
	return pattern
}
func Sbd_GiaSimRoundBack(p *Network) {
	mapping := Sbd_GiaSimRound(p, 0)
	representatives := make([]int32, len(p.Graph.Nodes()))
	pairs := vec.Vec_IntAlloc(1000)
	constants, classes := 0, 0
	for _, edge := range p.Graph.LiveEdges() {
		kind := p.Graph.Node(edge).Kind
		if kind != aig.And && kind != aig.ExclusiveOr {
			continue
		}
		id := int32(edge.NodeID())
		class := vec.Vec_IntEntry(mapping, int(id))
		if class == 0 {
			vec.Vec_IntPush(pairs, 0)
			vec.Vec_IntPush(pairs, id)
			constants++
		} else if representatives[class] == 0 {
			representatives[class] = id
		} else if representatives[class] != -1 {
			vec.Vec_IntPush(pairs, representatives[class])
			vec.Vec_IntPush(pairs, id)
			representatives[class] = -1
			classes++
		}
	}
	fmt.Printf("Constants = %d. Classes = %d.\n", constants, classes)
	fmt.Printf("Generated %d patterns.\n", Sbd_GiaSatOne(p, pairs))
}
func Sbd_GiaSimRound(p *Network, try int32) *vec.Vec_Int_t {
	p.ensureSimulation()
	if try != 0 {
		Sbd_GiaSimRoundBack(p)
		for _, edge := range p.Graph.Inputs() {
			Sbd_ObjSims(p, int32(edge.NodeID()))[0] <<= 1
		}
	} else {
		for _, edge := range p.Graph.Inputs() {
			for w := int32(0); w < p.SimWords; w++ {
				value := p.randomWord()
				if w == 0 {
					value <<= 1
				}
				Sbd_ObjSims(p, int32(edge.NodeID()))[w] = value
			}
		}
	}
	clear(Sbd_ObjSims(p, 0))
	mapping := vec.Vec_IntStart(len(p.Graph.Nodes()))
	index := map[string]int32{}
	key := func(values []uint64, compl bool) string {
		data := make([]byte, 8*len(values))
		for i, value := range values {
			if compl {
				value = ^value
			}
			binary.LittleEndian.PutUint64(data[8*i:], value)
		}
		return string(data)
	}
	index[key(Sbd_ObjSims(p, 0), false)] = 0
	for _, edge := range p.Graph.LiveEdges() {
		node := p.Graph.Node(edge)
		id := int32(edge.NodeID())
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			sims := Sbd_ObjSims(p, id)
			for w := range sims {
				a, b := sbdSimLiteral(p, node.Fanins[0], w), sbdSimLiteral(p, node.Fanins[1], w)
				if node.Kind == aig.ExclusiveOr {
					sims[w] = a ^ b
				} else {
					sims[w] = a & b
				}
			}
			signature := key(sims, sims[0]&1 != 0)
			class, ok := index[signature]
			if !ok {
				class = int32(len(index))
				index[signature] = class
			}
			vec.Vec_IntWriteEntry(mapping, int(id), class)
		case aig.Buffer, aig.Output:
			sims := Sbd_ObjSims(p, id)
			for w := range sims {
				sims[w] = sbdSimLiteral(p, node.Fanins[0], w)
			}
		}
	}
	fmt.Printf("Objects = %6d. Unique = %6d.\n", len(p.LUTs), len(index))
	return mapping
}
func Sbd_GiaSimTest(p *Network) {
	p.SimWords = 32
	p.SimValues = nil
	p.SimControls = nil
	p.random = aig.NewAigRandomSource()
	p.random.Uint32()
	for i := 0; i < 3; i++ {
		Sbd_GiaSimRound(p, 0)
	}
	for i := 0; i < 3; i++ {
		Sbd_GiaSimRound(p, 1)
	}
	p.SimValues, p.SimControls = nil, nil
}
