package saig

// Reference-counted packed sequential simulation from ABC's saigSimSeq.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"time"
)

type randomObject struct {
	kind   aig.NodeKind
	fanins [2]int32
	input  int
	refs   int32
}
type Raig_Man_t struct {
	pAig                      *aig.Graph
	nWords                    int32
	objects                   []randomObject // 0 unused, 1 constant one
	mapping                   map[int]int32
	inputs, outputs, los, lis []int32
	vCis2Ids                  *vec.Vec_Int_t
	slots                     []int
	blocks                    [][]uint32
	free                      []int
	live, maxLive             int
}

func Raig_Var2Lit(variable, complement int32) int32 { return variable<<1 | complement }
func Raig_Lit2Var(lit int32) int32                  { return lit >> 1 }
func Raig_LitIsCompl(lit int32) int32               { return lit & 1 }
func Raig_LitNot(lit int32) int32                   { return lit ^ 1 }
func Raig_LitNotCond(lit, condition int32) int32 {
	if condition > 0 {
		return lit ^ 1
	}
	return lit
}
func Raig_LitRegular(lit int32) int32 { return lit &^ 1 }
func Raig_ManFindPo(p *Raig_Man_t, id int32) int32 {
	for i, obj := range p.outputs {
		if obj == id {
			return int32(i)
		}
	}
	return -1
}
func Raig_ManCreate_rec(p *Raig_Man_t, edge aig.Edge) int32 {
	id := edge.NodeID()
	if mapped, ok := p.mapping[id]; ok {
		return mapped
	}
	node := p.pAig.Node(aig.Edge(id) << 1)
	obj := randomObject{kind: node.Kind, input: -1}
	child := func(edge aig.Edge) int32 {
		lit := Raig_ManCreate_rec(p, edge) << 1
		if edge.Complemented() {
			lit ^= 1
		}
		return lit
	}
	switch node.Kind {
	case aig.Input:
		for i, input := range p.pAig.Inputs() {
			if input.NodeID() == id {
				obj.input = i
				break
			}
		}
		vec.Vec_IntPush(p.vCis2Ids, int32(obj.input))
	case aig.And, aig.ExclusiveOr:
		obj.fanins[0] = child(node.Fanins[0])
		obj.fanins[1] = child(node.Fanins[1])
	case aig.Buffer, aig.Output:
		obj.fanins[0] = child(node.Fanins[0])
	default:
		panic("saig: invalid random simulation object")
	}
	mapped := int32(len(p.objects))
	p.mapping[id] = mapped
	p.objects = append(p.objects, obj)
	return mapped
}
func Raig_ManCreate(graph *aig.Graph) *Raig_Man_t {
	if graph == nil {
		panic("saig: nil random simulation graph")
	}
	if err := graph.Check(); err != nil {
		panic(err)
	}
	p := &Raig_Man_t{pAig: graph, nWords: 1, objects: make([]randomObject, 2), mapping: map[int]int32{0: 1}, vCis2Ids: vec.Vec_IntAlloc(graph.InputCount())}
	// Unused CIs come first, matching the source's random-number replay order.
	for _, input := range graph.Inputs() {
		if aig.Aig_ObjRefs(graph, input) == 0 {
			Raig_ManCreate_rec(p, input)
		}
	}
	for _, output := range graph.Outputs() {
		p.outputs = append(p.outputs, Raig_ManCreate_rec(p, output))
	}
	for _, input := range graph.Inputs() {
		if _, ok := p.mapping[input.NodeID()]; !ok {
			Raig_ManCreate_rec(p, input)
		}
	}
	for _, input := range graph.Inputs() {
		p.inputs = append(p.inputs, p.mapping[input.NodeID()])
	}
	for id := 2; id < len(p.objects); id++ {
		obj := p.objects[id]
		switch obj.kind {
		case aig.And, aig.ExclusiveOr:
			p.objects[obj.fanins[1]>>1].refs++
			fallthrough
		case aig.Buffer, aig.Output:
			p.objects[obj.fanins[0]>>1].refs++
		}
	}
	registers := int(aig.Aig_ManRegNum(graph))
	p.los = append(p.los, p.inputs[len(p.inputs)-registers:]...)
	p.lis = append(p.lis, p.outputs[len(p.outputs)-registers:]...)
	for _, id := range p.lis {
		p.objects[id].refs++
	}
	p.slots = make([]int, len(p.objects))
	return p
}
func Raig_ManDelete(p *Raig_Man_t) {
	if p != nil {
		p.blocks = nil
		p.slots = nil
		p.free = nil
		p.objects = nil
		p.mapping = nil
	}
}
func (p *Raig_Man_t) initialize() {
	if p.nWords <= 0 {
		panic("saig: nonpositive simulation word count")
	}
	if len(p.blocks) > 0 {
		if len(p.blocks[0]) != int(p.nWords)+1 {
			panic("saig: word count changed during simulation")
		}
		return
	}
	constant := make([]uint32, int(p.nWords)+1)
	for i := range constant {
		constant[i] = ^uint32(0)
	}
	p.blocks = [][]uint32{constant}
	p.live = 1
	p.maxLive = 1
}
func Raig_ManSimRef(p *Raig_Man_t, id int32) []uint32 {
	p.initialize()
	if id <= 1 || int(id) >= len(p.objects) || p.slots[id] != 0 {
		panic("saig: invalid simulation reference")
	}
	slot := 0
	if n := len(p.free); n > 0 {
		slot = p.free[n-1]
		p.free = p.free[:n-1]
	} else {
		slot = len(p.blocks)
		p.blocks = append(p.blocks, make([]uint32, int(p.nWords)+1))
	}
	p.slots[id] = slot
	block := p.blocks[slot]
	block[0] = uint32(p.objects[id].refs)
	p.live++
	p.maxLive = max(p.maxLive, p.live)
	return block
}
func Raig_ManSimDeref(p *Raig_Man_t, id int32) []uint32 {
	p.initialize()
	if id == 1 {
		return p.blocks[0]
	}
	if id <= 1 || int(id) >= len(p.slots) || p.slots[id] == 0 {
		panic("saig: missing simulation reference")
	}
	slot := p.slots[id]
	block := p.blocks[slot]
	if block[0] == 0 {
		panic("saig: simulation reference underflow")
	}
	block[0]--
	if block[0] == 0 {
		p.slots[id] = 0
		p.free = append(p.free, slot)
		p.live--
	}
	return block
}
func (p *Raig_Man_t) releaseUnused(id int32, block []uint32) {
	if block[0] == 0 {
		block[0] = 1
		Raig_ManSimDeref(p, id)
	}
}
func Raig_ManSimulateRound(p *Raig_Man_t, miter, first int32, pattern *int32) int32 {
	p.initialize()
	// Transfer all previous-frame register inputs before evaluating any new gate.
	for i, id := range p.los {
		block := Raig_ManSimRef(p, id)
		if first != 0 {
			clear(block[1:])
		} else {
			copy(block[1:], Raig_ManSimDeref(p, p.lis[i])[1:])
		}
		p.releaseUnused(id, block)
	}
	primaryInputs := len(p.inputs) - len(p.los)
	primaryOutputs := len(p.outputs) - len(p.lis)
	outputNumber := make(map[int32]int, len(p.outputs))
	for i, id := range p.outputs {
		outputNumber[id] = i
	}
	for id := int32(2); int(id) < len(p.objects); id++ {
		obj := p.objects[id]
		if obj.kind == aig.Input {
			if obj.input >= primaryInputs {
				continue
			}
			block := Raig_ManSimRef(p, id)
			for w := 1; w < len(block); w++ {
				block[w] = aig.Aig_ManRandom(0)
			}
			p.releaseUnused(id, block)
			continue
		}
		// Allocate before dereferencing fanins so live operands cannot be reused.
		result := Raig_ManSimRef(p, id)
		left := Raig_ManSimDeref(p, obj.fanins[0]>>1)
		var right []uint32
		if obj.kind == aig.And || obj.kind == aig.ExclusiveOr {
			right = Raig_ManSimDeref(p, obj.fanins[1]>>1)
		}
		for w := 1; w < len(result); w++ {
			a := left[w]
			if obj.fanins[0]&1 != 0 {
				a = ^a
			}
			if right == nil {
				result[w] = a
				continue
			}
			b := right[w]
			if obj.fanins[1]&1 != 0 {
				b = ^b
			}
			if obj.kind == aig.ExclusiveOr {
				result[w] = a ^ b
			} else {
				result[w] = a & b
			}
		}
		if obj.kind == aig.Output && miter != 0 && outputNumber[id] < primaryOutputs {
			for w, value := range result[1:] {
				if value != 0 {
					if pattern != nil {
						*pattern = int32(w*32 + bits.TrailingZeros32(value))
					}
					return id
				}
			}
		}
		p.releaseUnused(id, result)
	}
	if p.live != 1+len(p.lis) {
		panic("saig: leaked simulation blocks")
	}
	return 0
}
func Raig_ManGenerateCounter(graph *aig.Graph, frame, output, words, pattern int32, order *vec.Vec_Int_t) *util.Abc_Cex_t {
	registers := aig.Aig_ManRegNum(graph)
	pis := int32(graph.InputCount()) - registers
	if frame < 0 || words <= 0 || pattern < 0 || pattern >= 32*words {
		panic("saig: invalid simulation witness coordinates")
	}
	cex := util.Abc_CexAlloc(registers, pis, frame+1)
	cex.IFrame = frame
	cex.IPo = output
	aig.Aig_ManRandom(1)
	for f := int32(0); f <= frame; f++ {
		for _, input := range vec.Vec_IntArray(order) {
			if input >= pis {
				continue
			}
			for w := int32(0); w < words; w++ {
				value := aig.Aig_ManRandom(0)
				if w == pattern/32 && value&(1<<uint(pattern%32)) != 0 {
					bit := registers + f*pis + input
					cex.PData[bit>>5] |= 1 << uint(bit&31)
				}
			}
		}
	}
	return cex
}
func Raig_ManSimulate(graph *aig.Graph, words, iterations, timeLimit, miter, verbose int32) int32 {
	if words <= 0 || iterations < 0 || timeLimit < 0 || aig.Aig_ManRegNum(graph) <= 0 {
		panic("saig: invalid sequential simulation parameters")
	}
	p := Raig_ManCreate(graph)
	defer Raig_ManDelete(p)
	p.nWords = words
	aig.Aig_ManRandom(1)
	started := time.Now()
	for frame := int32(0); frame < iterations; frame++ {
		var pattern int32
		first := int32(0)
		if frame == 0 {
			first = 1
		}
		failed := Raig_ManSimulateRound(p, miter, first, &pattern)
		if failed != 0 {
			cex := Raig_ManGenerateCounter(graph, frame, Raig_ManFindPo(p, failed), words, pattern, p.vCis2Ids)
			if err := validateRandomCounterexample(graph, cex); err != nil {
				panic(err)
			}
			if err := graph.SetSequentialCounterexample(cex); err != nil {
				panic(err)
			}
			if verbose != 0 {
				fmt.Printf("Simulation found output %d at frame %d\n", cex.IPo, frame)
			}
			return 1
		}
		if timeLimit > 0 && time.Since(started) >= time.Duration(timeLimit)*time.Second {
			break
		}
	}
	if verbose != 0 {
		fmt.Printf("Simulation finished: max live blocks %d, elapsed %s\n", p.maxLive, time.Since(started))
	}
	return 0
}
func validateRandomCounterexample(graph *aig.Graph, cex *util.Abc_Cex_t) error {
	registers := int(cex.NRegs)
	pis := int(cex.NPis)
	state := make([]bool, registers)
	for i := range state {
		state[i] = cex.PData[i>>5]&(1<<uint(i&31)) != 0
	}
	for frame := 0; frame <= int(cex.IFrame); frame++ {
		inputs := make([]bool, pis+registers)
		copy(inputs[pis:], state)
		for i := 0; i < pis; i++ {
			bit := registers + frame*pis + i
			inputs[i] = cex.PData[bit>>5]&(1<<uint(bit&31)) != 0
		}
		outputs, err := graph.Evaluate(inputs)
		if err != nil {
			return err
		}
		if frame == int(cex.IFrame) {
			if cex.IPo < 0 || int(cex.IPo) >= len(outputs)-registers || !outputs[cex.IPo] {
				return fmt.Errorf("saig: random simulation witness does not reproduce")
			}
			return nil
		}
		copy(state, outputs[len(outputs)-registers:])
	}
	return fmt.Errorf("saig: empty simulation witness")
}
