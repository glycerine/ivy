package cec

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// cecLowerSweepGraph constructs an AND-inverter graph and retains an inverse
// source-node mapping. Synthetic AND nodes have no original ID. Outputs added
// temporarily for every source gate keep dangling candidate nodes reachable.
func cecLowerSweepGraph(src *Gia_Man_t) (*Gia_Man_t, []int32, *Vec_Int_t, error) {
	aigGraph := aig.NewGraph()
	mapped := make([]aig.Edge, src.ObjectCount())
	mapped[0] = aigGraph.Const0()
	for _, lit := range src.InputLiterals() {
		mapped[lit>>1] = aigGraph.NewInput()
	}
	read := func(lit int32) aig.Edge { return mapped[lit>>1] ^ aig.Edge(lit&1) }
	var nodes []int32
	for id := int32(1); int(id) < src.ObjectCount(); id++ {
		a, b, c, kind := src.Gate(id)
		switch kind {
		case 1, 6:
			continue
		case 2:
			mapped[id] = aigGraph.AND(read(a), read(b))
		case 3:
			mapped[id] = aigGraph.NewXor(read(a), read(b))
		case 4:
			mapped[id] = aigGraph.Mux(read(a), read(b), read(c))
		case 5:
			mapped[id] = read(a)
		default:
			return nil, nil, nil, fmt.Errorf("cec: cannot lower gate %d of kind %d", id, kind)
		}
		nodes = append(nodes, id)
	}
	outputs := src.OutputLiterals()
	for _, lit := range outputs {
		aigGraph.AddOutput(read(lit))
	}
	for _, id := range nodes {
		aigGraph.AddOutput(mapped[id])
	}
	intermediate := gia.Gia_ManFromAig(aigGraph)
	defer gia.Gia_ManStop(intermediate)
	result := gia.Gia_ManStart(int32(intermediate.ObjectCount()))
	result.CopyIdentityFrom(src)
	copies := make([]int32, intermediate.ObjectCount())
	copies[0] = 0
	for _, lit := range intermediate.InputLiterals() {
		copies[lit>>1] = gia.Gia_ManAppendCi(result)
	}
	remap := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
	for id := int32(1); int(id) < intermediate.ObjectCount(); id++ {
		a, b, _, kind := intermediate.Gate(id)
		switch kind {
		case 1, 6:
			continue
		case 2:
			copies[id] = gia.Gia_ManAppendAnd(result, remap(a), remap(b))
		default:
			return nil, nil, nil, fmt.Errorf("cec: lowering retained gate kind %d", kind)
		}
	}
	loweredOutputs := intermediate.OutputLiterals()
	for _, lit := range loweredOutputs[:len(outputs)] {
		gia.Gia_ManAppendCo(result, remap(lit))
	}
	gia.Gia_ManSetRegNum(result, gia.Gia_ManRegNum(src))
	gia.Gia_ManSetConstrNum(result, gia.Gia_ManConstrNum(src))
	inverse := make([]int32, result.ObjectCount())
	forward := make([]int32, src.ObjectCount())
	for i := range inverse {
		inverse[i] = -1
	}
	for i := range forward {
		forward[i] = -1
	}
	inverse[0] = 0
	forward[0] = 0
	sourceInputs, lowerInputs := src.InputLiterals(), result.InputLiterals()
	for i, lit := range sourceInputs {
		inverse[lowerInputs[i]>>1] = lit >> 1
		forward[lit>>1] = lowerInputs[i] >> 1
	}
	aliases := vec.Vec_IntAlloc(0)
	for i, id := range nodes {
		lit := remap(loweredOutputs[len(outputs)+i])
		varID := lit >> 1
		forward[id] = varID
		if inverse[varID] < 0 {
			inverse[varID] = id
		} else {
			vec.Vec_IntPushTwo(aliases, inverse[varID], id)
		}
	}
	if src.HasRepresentatives() {
		if err := result.TransferRepresentativesFrom(src, forward); err != nil {
			gia.Gia_ManStop(result)
			return nil, nil, nil, fmt.Errorf("cec: remap native equivalence classes: %w", err)
		}
	}
	return result, inverse, aliases, nil
}
