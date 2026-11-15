package gia

import (
	"fmt"
	"testing"
)

func TestRehashPreservesOutputRoles(t *testing.T) {
	for additional := int32(0); additional < 2; additional++ {
		src := partitionReviewGraph(0, false, false)
		if err := src.SetAigerProperties(2, 2); err != nil {
			t.Fatal(err)
		}
		out := Gia_ManRehash(src, additional)
		if out.aigerBadOutputs != 2 || out.nConstrs != 2 {
			t.Errorf("additional%d: roles bad%d constraints%d, want2/2", additional, out.aigerBadOutputs, out.nConstrs)
		}
	}
}

func TestRehashPreservesInterleavedInterface(t *testing.T) {
	for additional := int32(0); additional < 2; additional++ {
		t.Run(fmt.Sprint(additional), func(t *testing.T) {
			src := partitionReviewGraph(5, false, true)
			out := Gia_ManRehash(src, additional)
			for mask := 0; mask < 32; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0, mask&16 != 0}
				want, err := src.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				checkQuantificationOutputs(t, out, inputs, want)
			}
		})
	}
}

func TestRehashPreservesRealGateLogic(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		for additional := int32(0); additional < 2; additional++ {
			t.Run(fmt.Sprintf("%s_%d", kind, additional), func(t *testing.T) {
				src := Gia_ManStart(1)
				a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
				var root int32
				switch kind {
				case "xor":
					root = Gia_ManAppendXorReal(src, a, b)
				case "mux":
					src.EnableMuxes()
					root = Gia_ManAppendMuxReal(src, c, a, b)
				case "buffer":
					root = Gia_ManAppendBuf(src, a^1)
				}
				Gia_ManAppendCo(src, root)
				out := Gia_ManRehash(src, additional)
				for mask := 0; mask < 8; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					want, err := src.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					checkQuantificationOutputs(t, out, inputs, want)
				}
			})
		}
	}
}
