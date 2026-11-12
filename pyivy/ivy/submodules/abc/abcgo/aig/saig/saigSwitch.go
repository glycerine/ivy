package saig

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"math/rand"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_SimObj_t is a compact per-node simulation record. PData stores one
// 32-pattern simulation word, while Number accumulates ones over sampled
// frames after the requested prefix.
type Saig_SimObj_t struct {
	iFan0, iFan1 int32
	Type         aig.NodeKind
	Number       uint32
	pData        uint32
	isLatch      bool
	valid        bool
}

func Saig_SimObjFaninC0(p *Saig_SimObj_t) int32 { return p.iFan0 & 1 }
func Saig_SimObjFaninC1(p *Saig_SimObj_t) int32 { return p.iFan1 & 1 }
func Saig_SimObjFanin0(p *Saig_SimObj_t) int32  { return p.iFan0 >> 1 }
func Saig_SimObjFanin1(p *Saig_SimObj_t) int32  { return p.iFan1 >> 1 }

// Saig_ManCreateMan builds the compact simulation view for a sequential AIG.
func Saig_ManCreateMan(g *aig.Graph) []Saig_SimObj_t {
	if g == nil {
		return nil
	}
	nodes, inputs, outputs := g.Nodes(), g.Inputs(), g.Outputs()
	regs := int(aig.Aig_ManRegNum(g))
	piCount := len(inputs) - regs
	poCount := len(outputs) - regs
	if regs < 0 || piCount < 0 || poCount < 0 {
		return nil
	}
	result := make([]Saig_SimObj_t, len(nodes)+1)
	for id, n := range nodes {
		entry := &result[id]
		if !g.ValidEdge(aig.Edge(id) << 1) {
			continue
		}
		entry.Type, entry.valid = n.Kind, true
		switch n.Kind {
		case aig.Constant:
		case aig.Input:
			for i, input := range inputs {
				if input.NodeID() == id {
					if i < piCount {
						entry.iFan0 = 0 // primary input marker
					} else {
						li := outputs[poCount+i-piCount]
						entry.iFan0 = int32(aig.Aig_ObjChild0(g, li))
						entry.isLatch = true
					}
					break
				}
			}
		case aig.And, aig.ExclusiveOr:
			entry.iFan0, entry.iFan1 = int32(n.Fanins[0]), int32(n.Fanins[1])
		case aig.Buffer, aig.Output:
			entry.iFan0 = int32(n.Fanins[0])
		}
	}
	result[len(nodes)].valid = false // sentinel
	return result
}

func Saig_ManSimulateNode2(rows []Saig_SimObj_t, obj int) {
	if obj < 0 || obj >= len(rows) {
		return
	}
	Saig_ManSimulateNode(rows, obj)
}

func Saig_ManSimulateNode(rows []Saig_SimObj_t, obj int) {
	entry := &rows[obj]
	a := rows[Saig_SimObjFanin0(entry)].pData
	b := rows[Saig_SimObjFanin1(entry)].pData
	if Saig_SimObjFaninC0(entry) != 0 {
		a = ^a
	}
	if Saig_SimObjFaninC1(entry) != 0 {
		b = ^b
	}
	if entry.Type == aig.ExclusiveOr {
		entry.pData = a ^ b
	} else {
		entry.pData = a & b
	}
}

func Saig_ManSimulateOneInput(rows []Saig_SimObj_t, obj int) {
	entry := &rows[obj]
	value := rows[Saig_SimObjFanin0(entry)].pData
	if Saig_SimObjFaninC0(entry) != 0 {
		value = ^value
	}
	entry.pData = value
}

// Saig_ManSimulateFrames runs bit-parallel random simulation, counting node
// activity only after the warm-up prefix.
func Saig_ManSimulateFrames(rows []Saig_SimObj_t, frames, prefix int32) {
	if frames <= 0 {
		return
	}
	rng := rand.New(rand.NewSource(1))
	for frame := int32(0); frame < frames; frame++ {
		for id := 0; id < len(rows)-1; id++ {
			entry := &rows[id]
			if !entry.valid {
				continue
			}
			switch entry.Type {
			case aig.Constant:
				entry.pData = ^uint32(0)
			case aig.Input:
				if !entry.isLatch {
					entry.pData = rng.Uint32()
				} else if frame > 0 {
					Saig_ManSimulateOneInput(rows, id)
				}
			case aig.And, aig.ExclusiveOr:
				Saig_ManSimulateNode(rows, id)
			case aig.Buffer, aig.Output:
				Saig_ManSimulateOneInput(rows, id)
			}
			if frame >= prefix {
				entry.Number += uint32(bits.OnesCount32(entry.pData))
			}
		}
	}
}

func Saig_ManComputeSwitching(ones, simulationWords int32) float32 {
	total := float32(32 * simulationWords)
	if total == 0 {
		return 0
	}
	p := float32(ones) / total
	return 2 * p * (1 - p)
}

func Saig_ManComputeProbOne(ones, simulationWords int32) float32 {
	total := float32(32 * simulationWords)
	if total == 0 {
		return 0
	}
	return float32(ones) / total
}

func Saig_ManComputeProbOnePlus(ones, simulationWords, complemented int32) float32 {
	p := Saig_ManComputeProbOne(ones, simulationWords)
	if complemented != 0 {
		return 1 - p
	}
	return p
}

// Saig_ManComputeSwitchProb4s returns IEEE-754 float32 bit patterns in the
// legacy integer vector, preserving the C ABI representation.
func Saig_ManComputeSwitchProb4s(g *aig.Graph, frames, prefix, probability int32) *vec.Vec_Int_t {
	if g == nil || frames <= 0 || prefix < 0 {
		return vec.Vec_IntAlloc(0)
	}
	rows := Saig_ManCreateMan(g)
	Saig_ManSimulateFrames(rows, frames, prefix)
	result := vec.Vec_IntStart(len(rows) - 1)
	simulatedFrames := frames - prefix
	if simulatedFrames <= 0 {
		return result
	}
	for id := 0; id < len(rows)-1; id++ {
		var value float32
		if probability != 0 {
			value = Saig_ManComputeProbOne(int32(rows[id].Number), simulatedFrames)
		} else {
			value = Saig_ManComputeSwitching(int32(rows[id].Number), simulatedFrames)
		}
		vec.Vec_IntWriteEntry(result, id, int32(math.Float32bits(value)))
	}
	return result
}

// Aig_CMan_t stores a compact, delta-encoded AIG node/output stream.
type Aig_CMan_t struct {
	nIns, nNodes, nOuts int32
	iNode, iPrev        int32
	Data                []byte
	readAt              int
}

func Aig_CManStart(inputs, nodes, outputs int32) *Aig_CMan_t {
	if inputs < 0 || nodes < 0 || outputs < 0 {
		panic("saig: negative compact-AIG size")
	}
	return &Aig_CMan_t{nIns: inputs, nNodes: nodes, nOuts: outputs, iNode: inputs + 1, iPrev: -1, Data: make([]byte, 0, 2*(2*int(nodes)+int(outputs)))}
}

func Aig_CManStop(p *Aig_CMan_t) {
	if p != nil {
		p.Data = nil
	}
}

func Aig_CManRestart(p *Aig_CMan_t) {
	if p.iNode != 1+p.nIns+p.nNodes+p.nOuts {
		panic("saig: compact-AIG cursor has not reached the end")
	}
	p.iNode, p.iPrev, p.readAt = 1+p.nIns, -1, 0
}

func Aig_CManStoreNum(p *Aig_CMan_t, value uint32) {
	var buf [binary.MaxVarintLen32]byte
	n := binary.PutUvarint(buf[:], uint64(value))
	p.Data = append(p.Data, buf[:n]...)
}

func Aig_CManRestoreNum(p *Aig_CMan_t) int32 {
	if p.readAt >= len(p.Data) {
		panic("saig: compact-AIG stream exhausted")
	}
	value, n := binary.Uvarint(p.Data[p.readAt:])
	if n <= 0 || value > (1<<31)-1 {
		panic("saig: invalid compact-AIG varint")
	}
	p.readAt += n
	return int32(value)
}

func Aig_CManAddNode(p *Aig_CMan_t, fanin0, fanin1 int32) {
	if fanin0 >= fanin1 || fanin1 >= p.iNode<<1 {
		panic("saig: compact-AIG fanins must be ordered and earlier than their node")
	}
	Aig_CManStoreNum(p, uint32((p.iNode<<1)-fanin1))
	Aig_CManStoreNum(p, uint32(fanin1-fanin0))
	p.iNode++
}

func Aig_CManAddPo(p *Aig_CMan_t, fanin int32) {
	switch {
	case p.iPrev == -1:
		Aig_CManStoreNum(p, uint32(p.iNode-fanin))
	case p.iPrev <= fanin:
		Aig_CManStoreNum(p, uint32((fanin-p.iPrev)<<1))
	default:
		Aig_CManStoreNum(p, uint32(((p.iPrev-fanin)<<1)|1))
	}
	p.iPrev = fanin
	p.iNode++
}

func Aig_CManGetNode(p *Aig_CMan_t, fanin0, fanin1 *int32) {
	*fanin1 = (p.iNode << 1) - Aig_CManRestoreNum(p)
	*fanin0 = *fanin1 - Aig_CManRestoreNum(p)
	p.iNode++
}

func Aig_CManGetPo(p *Aig_CMan_t) int32 {
	num := Aig_CManRestoreNum(p)
	if p.iPrev == -1 {
		p.iPrev = p.iNode
	}
	p.iNode++
	if num&1 != 0 {
		p.iPrev += num >> 1
	} else {
		p.iPrev -= num >> 1
	}
	return p.iPrev
}

func Aig_CManCreate(g *aig.Graph) *Aig_CMan_t {
	if g == nil {
		return nil
	}
	nodes, outputs := g.Nodes(), g.Outputs()
	andCount := 0
	for id, node := range nodes {
		if !g.ValidEdge(aig.Edge(id) << 1) {
			continue
		}
		if node.Kind == aig.ExclusiveOr {
			panic("saig: compact-AIG format does not encode XOR node kinds")
		}
		if node.Kind == aig.And {
			andCount++
		}
	}
	manager := Aig_CManStart(int32(len(g.Inputs())), int32(andCount), int32(len(outputs)))
	for _, node := range nodes {
		if node.Kind == aig.And {
			Aig_CManAddNode(manager, int32(node.Fanins[0]), int32(node.Fanins[1]))
		}
	}
	for _, output := range outputs {
		Aig_CManAddPo(manager, int32(aig.Aig_ObjChild0(g, output)))
	}
	return manager
}

func (p *Aig_CMan_t) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Aig_CMan(inputs=%d nodes=%d outputs=%d bytes=%d)", p.nIns, p.nNodes, p.nOuts, len(p.Data))
}
