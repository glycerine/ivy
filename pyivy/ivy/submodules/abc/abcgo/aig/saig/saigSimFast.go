package saig

// Fast packed-word sequential activity simulation, based on ABC's
// saigSimFast.c (Alan Mishchenko, UC Berkeley).
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"math/bits"
)

type fastNode struct {
	kind   aig.NodeKind
	fanins [2]int32
}
type Faig_Man_t struct {
	nPis, nFfs int32
	ids        []aig.Edge // canonical compact order -> original graph object
	nodes      []fastNode
	outputs    []int32
}

func (p *Faig_Man_t) pos() int32                 { return int32(len(p.outputs)) - p.nFfs }
func (p *Faig_Man_t) cis() int32                 { return p.nPis + p.nFfs }
func (p *Faig_Man_t) nodeStart() int32           { return 1 + p.cis() }
func (p *Faig_Man_t) outputStart() int32         { return p.nodeStart() + int32(len(p.nodes)) }
func (p *Faig_Man_t) size() int32                { return int32(len(p.ids)) }
func Faig_CheckIdPi(p *Faig_Man_t, i int32) bool { return i >= 1 && i <= p.nPis }
func Faig_CheckIdLo(p *Faig_Man_t, i int32) bool { return i > p.nPis && i <= p.cis() }
func Faig_CheckIdNo(p *Faig_Man_t, i int32) bool { return i >= p.nodeStart() && i < p.outputStart() }
func Faig_CheckIdPo(p *Faig_Man_t, i int32) bool {
	return i >= p.outputStart() && i < p.outputStart()+p.pos()
}
func Faig_CheckIdLi(p *Faig_Man_t, i int32) bool  { return i >= p.outputStart()+p.pos() && i < p.size() }
func Faig_CheckIdCo(p *Faig_Man_t, i int32) bool  { return i >= p.outputStart() && i < p.size() }
func Faig_CheckIdObj(p *Faig_Man_t, i int32) bool { return i >= 0 && i < p.size() }
func fastRequire(ok bool) {
	if !ok {
		panic("saig: invalid fast simulation object")
	}
}
func Faig_ObjIdToNumPi(p *Faig_Man_t, i int32) int32 { fastRequire(Faig_CheckIdPi(p, i)); return i - 1 }
func Faig_ObjIdToNumLo(p *Faig_Man_t, i int32) int32 {
	fastRequire(Faig_CheckIdLo(p, i))
	return i - p.nPis - 1
}
func Faig_ObjIdToNumNo(p *Faig_Man_t, i int32) int32 {
	fastRequire(Faig_CheckIdNo(p, i))
	return i - p.nodeStart()
}
func Faig_ObjIdToNumPo(p *Faig_Man_t, i int32) int32 {
	fastRequire(Faig_CheckIdPo(p, i))
	return i - p.outputStart()
}
func Faig_ObjIdToNumLi(p *Faig_Man_t, i int32) int32 {
	fastRequire(Faig_CheckIdLi(p, i))
	return i - p.outputStart() - p.pos()
}
func Faig_ObjIdToNumCo(p *Faig_Man_t, i int32) int32 {
	fastRequire(Faig_CheckIdCo(p, i))
	return i - p.outputStart()
}
func Faig_ObjLoToLi(p *Faig_Man_t, i int32) int32 {
	return p.outputStart() + p.pos() + Faig_ObjIdToNumLo(p, i)
}
func Faig_ObjLiToLo(p *Faig_Man_t, i int32) int32  { return p.nPis + 1 + Faig_ObjIdToNumLi(p, i) }
func Faig_NodeChild0(p *Faig_Man_t, n int32) int32 { return p.nodes[n].fanins[0] }
func Faig_NodeChild1(p *Faig_Man_t, n int32) int32 { return p.nodes[n].fanins[1] }
func Faig_CoChild0(p *Faig_Man_t, n int32) int32   { return p.outputs[n] }
func Faig_ObjFaninC(lit int32) int32               { return lit & 1 }
func Faig_ObjFanin(lit int32) int32                { return lit >> 1 }
func Faig_ManIsCorrect(graph *aig.Graph) int32 {
	if graph == nil || graph.Check() != nil {
		return 0
	}
	if aig.Aig_ManObjNumMax(graph) == 1+aig.Aig_ManCiNum(graph)+aig.Aig_ManNodeNum(graph)+aig.Aig_ManCoNum(graph) {
		return 1
	}
	return 0
}

func Faig_ManAlloc(graph *aig.Graph) *Faig_Man_t { return Faig_ManCreate(graph) }
func Faig_ManCreate(graph *aig.Graph) *Faig_Man_t {
	if graph == nil {
		panic("saig: nil simulation graph")
	}
	if err := graph.Check(); err != nil {
		panic(err)
	}
	registers := aig.Aig_ManRegNum(graph)
	p := &Faig_Man_t{nPis: int32(graph.InputCount()) - registers, nFfs: registers, ids: []aig.Edge{graph.Const1()}}
	mapping := map[int]int32{0: 0}
	for _, input := range graph.Inputs() {
		mapping[input.NodeID()] = int32(len(p.ids)) << 1
		p.ids = append(p.ids, input)
	}
	active := make(map[int]bool)
	var copyEdge func(aig.Edge) int32
	copyEdge = func(edge aig.Edge) int32 {
		id := edge.NodeID()
		lit, ok := mapping[id]
		if !ok {
			if active[id] {
				panic("saig: cycle in simulation graph")
			}
			active[id] = true
			node := graph.Node(aig.Edge(id) << 1)
			entry := fastNode{kind: node.Kind}
			switch node.Kind {
			case aig.And, aig.ExclusiveOr:
				entry.fanins[0] = copyEdge(node.Fanins[0])
				entry.fanins[1] = copyEdge(node.Fanins[1])
			case aig.Buffer:
				entry.fanins[0] = copyEdge(node.Fanins[0])
			default:
				panic("saig: invalid internal simulation node")
			}
			lit = int32(len(p.ids)) << 1
			mapping[id] = lit
			p.ids = append(p.ids, aig.Edge(id)<<1)
			p.nodes = append(p.nodes, entry)
			delete(active, id)
		}
		if edge.Complemented() {
			lit ^= 1
		}
		return lit
	}
	for id, node := range graph.Nodes() {
		edge := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(graph, edge) || node.Kind == aig.Constant || node.Kind == aig.Input || node.Kind == aig.Output {
			continue
		}
		copyEdge(edge)
	}
	for _, output := range graph.Outputs() {
		p.outputs = append(p.outputs, copyEdge(graph.Node(output).Fanins[0]))
	}
	p.ids = append(p.ids, graph.Outputs()...)
	return p
}
func fastLitValue(lit int32, values []uint32) uint32 {
	v := values[lit>>1]
	if lit&1 != 0 {
		v = ^v
	}
	return v
}
func Faig_SimulateNode(p *Faig_Man_t, id int32, values []uint32) uint32 {
	node := p.nodes[Faig_ObjIdToNumNo(p, id)]
	a := fastLitValue(node.fanins[0], values)
	if node.kind == aig.Buffer {
		return a
	}
	b := fastLitValue(node.fanins[1], values)
	if node.kind == aig.ExclusiveOr {
		return a ^ b
	}
	return a & b
}
func Faig_SimulateCo(p *Faig_Man_t, id int32, values []uint32) uint32 {
	return fastLitValue(p.outputs[Faig_ObjIdToNumCo(p, id)], values)
}
func Faig_SimulateRandomShift(old uint32) uint32         { return old<<16 | (old^aig.Aig_ManRandom(0))&0xffff }
func Faig_SimulateTransferShift(old, next uint32) uint32 { return old<<16 | next&0xffff }
func Faig_ManSimulateFrames(p *Faig_Man_t, frames, prefix, transitions int32) []int32 {
	if frames < 0 || prefix < 0 || prefix > frames {
		panic("saig: invalid simulation frames")
	}
	samplesPerFrame := int64(32)
	if transitions != 0 {
		samplesPerFrame = 16
	}
	if int64(frames-prefix)*samplesPerFrame > 1<<31-1 {
		panic("saig: simulation count exceeds int32 range")
	}
	counts := make([]int32, p.size())
	values := make([]uint32, p.size())
	values[0] = ^uint32(0)
	for frame := int32(0); frame < frames; frame++ {
		for id := int32(1); id <= p.nPis; id++ {
			if transitions != 0 && frame > 0 {
				values[id] = Faig_SimulateRandomShift(values[id])
			} else {
				values[id] = aig.Aig_ManRandom(0)
			}
		}
		for id := p.nPis + 1; id <= p.cis(); id++ {
			next := uint32(0)
			if frame > 0 {
				next = values[Faig_ObjLoToLi(p, id)]
			}
			if transitions != 0 && frame > 0 {
				next = Faig_SimulateTransferShift(values[id], next)
			}
			values[id] = next
		}
		for id := p.nodeStart(); id < p.outputStart(); id++ {
			values[id] = Faig_SimulateNode(p, id, values)
		}
		for id := p.outputStart(); id < p.size(); id++ {
			values[id] = Faig_SimulateCo(p, id, values)
		}
		if frame < prefix {
			continue
		}
		for id, value := range values {
			if transitions != 0 {
				value = (value ^ (value >> 16)) & 0xffff
			}
			counts[id] += int32(bits.OnesCount32(value))
		}
	}
	return counts
}
func Faig_ManComputeSwitching(ones, words int32) float32 {
	total := float32(32 * words)
	return 2 * float32(ones) / total * (total - float32(ones)) / total
}
func Faig_ManComputeProbOne(ones, words int32) float32 {
	return float32(ones) / float32(32*words)
}

// SwitchingProbabilities returns one probability per original object ID. It
// retains all nodes, including dangling nodes, and remaps sparse/reordered IDs.
func SwitchingProbabilities(graph *aig.Graph, frames, prefix int32, probabilityOne bool) ([]float32, error) {
	if frames <= prefix || prefix < 0 {
		return nil, fmt.Errorf("saig: total frames must exceed nonnegative prefix")
	}
	if graph == nil {
		return nil, fmt.Errorf("saig: nil simulation graph")
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	p := Faig_ManCreate(graph)
	aig.Aig_ManRandom(1)
	transitions := int32(1)
	words := (frames - prefix) / 2
	if probabilityOne {
		transitions = 0
		words = frames - prefix
	}
	counts := Faig_ManSimulateFrames(p, frames, prefix, transitions)
	result := make([]float32, aig.Aig_ManObjNumMax(graph))
	for id, edge := range p.ids {
		result[edge.NodeID()] = Faig_ManComputeProbOne(counts[id], words)
	}
	return result, nil
}
func Faig_ManComputeSwitchProbs4(graph *aig.Graph, frames, prefix, probabilityOne int32) *vec.Vec_Int_t {
	probabilities, err := SwitchingProbabilities(graph, frames, prefix, probabilityOne != 0)
	if err != nil {
		panic(err)
	}
	result := vec.Vec_IntAlloc(len(probabilities))
	for _, p := range probabilities {
		vec.Vec_IntPush(result, int32(math.Float32bits(p)))
	}
	return result
}
func Saig_ManComputeSwitchProb3s(graph *aig.Graph, frames, prefix, probabilityOne int32) *vec.Vec_Int_t {
	return Faig_ManComputeSwitchProbs4(graph, frames, prefix, probabilityOne)
}
