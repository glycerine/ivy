package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
)

// cec4SimulateGraph propagates current input patterns through combinational
// nodes and outputs. The caller supplies input bits and initializes classes.
func cec4SimulateGraph(s *cec3Simulation, graph *gia.Gia_Man_t) {
	for id := int32(1); id < gia.Gia_ManObjNum(graph); id++ {
		_, _, _, kind := graph.Gate(id)
		switch kind {
		case 0, 1:
		case 2, 3, 4, 5, 6:
			cec4SimulateNode(s, graph, id)
		default:
			panic("cec4: unsupported graph node")
		}
	}
}

// cec4FirstFailingOutput returns the first output with a simulated one,
// together with a combinational counterexample for that output.
func cec4FirstFailingOutput(s *cec3Simulation, graph *gia.Gia_Man_t) (int32, *util.Abc_Cex_t) {
	for i := int32(0); i < gia.Gia_ManCoNum(graph); i++ {
		id := gia.Gia_ObjId(graph, gia.Gia_ManCo(graph, i))
		for word, values := range s.object(id) {
			if values == 0 {
				continue
			}
			bit := word*64 + bits.TrailingZeros64(values)
			cex := util.Abc_CexAlloc(0, gia.Gia_ManCiNum(graph), 1)
			cex.IPo = i
			for j := int32(0); j < gia.Gia_ManCiNum(graph); j++ {
				input := gia.Gia_ObjId(graph, gia.Gia_ManCi(graph, j))
				if s.object(input)[bit>>6]&(uint64(1)<<uint(bit&63)) != 0 {
					cex.PData[j>>5] |= uint32(1) << uint(j&31)
				}
			}
			return i, cex
		}
	}
	return -1, nil
}
