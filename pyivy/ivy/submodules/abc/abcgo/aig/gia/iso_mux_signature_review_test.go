package gia

import (
	"fmt"
	"testing"
)

func TestIsoDormantMuxRecognitionPreservesSourceSalts(t *testing.T) {
	for _, xor := range []bool{false, true} {
		patterns := 8
		if xor {
			patterns = 4
		}
		for phases := 0; phases < patterns; phases++ {
			t.Run(fmt.Sprintf("xor%t/phases%d", xor, phases), func(t *testing.T) {
				graph := Gia_ManStart(10)
				a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
				control := int32(0)
				var root int32
				if xor {
					root = Gia_ManAppendXor(graph, a^int32(phases&1), b^int32((phases>>1)&1))
				} else {
					control = Gia_ManAppendCi(graph)
					root = Gia_ManAppendMux(graph, control^int32(phases&1), a^int32((phases>>1)&1), b^int32((phases>>2)&1))
				}
				Gia_ManAppendCo(graph, root)
				expected := make([]uint64, Gia_ManObjNum(graph))
				fields := make([][8]uint32, len(expected))
				for id := range expected {
					expected[id] = (0xfffffff0 + uint64(id)*101) & 0xffffffff
					obj := Gia_ManObj(graph, int32(id))
					obj.Value = uint32(expected[id])
					obj.fMark0, obj.fMark1 = uint32(id%2), uint32((id+1)%2)
					fields[id] = isoSignatureReviewFields(obj)
				}
				if xor {
					isoSignatureReviewAdd(expected, int(root>>1), uint64(s_256Primes[233]))
					for _, input := range []int32{a, b} {
						isoSignatureReviewAdd(expected, int(input>>1), uint64(s_256Primes[234]))
					}
				} else {
					isoSignatureReviewAdd(expected, int(root>>1), uint64(s_256Primes[235]))
					isoSignatureReviewAdd(expected, int(control>>1), uint64(s_256Primes[236]))
					for _, input := range []int32{a, b} {
						isoSignatureReviewAdd(expected, int(input>>1), uint64(s_256Primes[237]))
					}
				}
				Gia_IsoRecognizeMuxes(graph)
				checkIsoSignatureReview(t, graph, expected, fields, "dormant XOR/MUX recognition")
			})
		}
	}
}
