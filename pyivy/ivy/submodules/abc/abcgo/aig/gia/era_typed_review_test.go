package gia

import (
	"fmt"
	"testing"
)

func TestExplicitReachabilityTypedInitialViolation(t *testing.T) {
	for _, inputs := range []int{1, 6} {
		for _, registers := range []int{1, 33, 65} {
			for _, kind := range []string{"xor", "mux", "buffer", "simple opposite"} {
				t.Run(fmt.Sprintf("%s/inputs%d/registers%d", kind, inputs, registers), func(t *testing.T) {
					p := Gia_ManStart(16)
					defer Gia_ManStop(p)
					var a, r int32
					var stateInputs []int32
					for i := 0; i < inputs; i++ {
						a = Gia_ManAppendCi(p)
					}
					for i := 0; i < registers; i++ {
						r = Gia_ManAppendCi(p)
						stateInputs = append(stateInputs, r)
					}
					var output int32
					want := int32(0) // PI=true violates at the initial all-zero state.
					switch kind {
					case "xor":
						output = Gia_ManAppendXorReal(p, a, r)
					case "mux":
						p.EnableMuxes()
						output = Gia_ManAppendMuxReal(p, a, 1, r)
					case "buffer":
						output = Gia_ManAppendBuf(p, a)
					case "simple opposite":
						p.fGiaSimple = 1
						output, want = Gia_ManAppendAnd(p, a, a^1), 1
					}
					Gia_ManAppendCo(p, output)
					for _, ro := range stateInputs {
						Gia_ManAppendCo(p, ro)
					}
					Gia_ManSetRegNum(p, int32(registers))
					var got int32
					hashReviewStdout(t, func() { got = Gia_ManCollectReachable(p, 100, 1, 0, 0) })
					m := Gia_ManEraCreate(p)
					initial := Gia_ManEraCreateState(m)
					m.pBins[Gia_ManEraStateHash(initial.pData, m.nWordsDat, m.nBins)] = initial.Num
					Gia_ManInsertState(m, initial)
					Gia_ManPerformOneIter(m)
					failure := Gia_ManAnalyzeResult(m, initial, 1, 0)
					if want == 0 {
						if failure != 1 {
							t.Error("word simulation missed the initial violation")
						}
						assertIDs(t, m.vBugTrace, 0, 1<<uint(inputs-1))
					} else if failure != 0 {
						t.Error("opposite-fanin AND falsely violated")
					}
					Gia_ManEraFree(m)
					if got != want {
						t.Fatalf("reachability returned %d, want %d (0 unsafe, 1 safe)", got, want)
					}
				})
			}

		}
	}
}
