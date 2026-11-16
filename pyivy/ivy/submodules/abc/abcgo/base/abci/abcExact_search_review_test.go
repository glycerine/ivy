package abci

import (
	"slices"
	"testing"
)

func TestExactSearchDirections(t *testing.T) {
	for _, tc := range []struct {
		name  string
		truth uint64
		gates byte
	}{
		{"and", 0x80, 2},
		{"parity", 0x96, 2},
		{"mux", 0xca, 3},
		{"majority", 0xe8, 4},
	} {
		for _, topDown := range []bool{false, true} {
			name := tc.name + "/bottomUp"
			if topDown {
				name = tc.name + "/topDown"
			}
			t.Run(name, func(t *testing.T) {
				truth := [4]word{tc.truth * 0x0101010101010101}
				arrivals := [3]int32{}
				p := Ses_ManAlloc(&truth[0], 3, 1, 4, &arrivals[0], 0, 10000, 0)
				defer Ses_ManClean(p)
				p.nStartGates = 1
				// Start top-down at an attainable gate count. It must retain
				// this model when proving that a smaller network is impossible.
				p.nMaxGates = int32(tc.gates)
				var model string
				if topDown {
					model = Ses_ManFindMinimumSizeTopDown(p, 1)
				} else {
					model = Ses_ManFindMinimumSizeBottomUp(p)
				}
				if model == "" || model[2] != tc.gates || p.fHitResLimit != 0 || p.nSatCalls == 0 {
					t.Fatalf("search gates: model%v limit%d sat%d", []byte(model), p.fHitResLimit, p.nSatCalls)
				}
				exactSynthesisReviewModel(t, model, 3, truth[:1])
			})
		}
	}
}

func TestExactSearchRetainsSolutionAfterResourceLimit(t *testing.T) {
	exactMappingReviewStoreMode(t, 0)
	s_pSesStore.nBTLimit = 1
	// (a AND b) XOR c has a two-gate, depth-two implementation. With
	// this small budget, later depth-search attempts hit the resource limit.
	truth := []word{0x7878787878787878}
	arrivals := []int32{0, 0, 0}
	pins := make([]int8, 3)
	var cost int32
	delay := Abc_ExactDelayCost(&truth[0], 3, &arrivals[0], pins, &cost, 4)
	if delay != 2 || cost != 2 || s_pSesStore.nUndefCalls == 0 || s_pSesStore.nSynthesizedRL != 1 || s_pSesStore.nUnsynthesizedRL != 0 {
		t.Fatalf("lost feasible solution: delay%d cost%d undef%d successRL%d failedRL%d", delay, cost, s_pSesStore.nUndefCalls, s_pSesStore.nSynthesizedRL, s_pSesStore.nUnsynthesizedRL)
	}
	var model string
	if Ses_StoreGetEntry(s_pSesStore, &truth[0], 3, &arrivals[0], &model) != 1 || model == "" {
		t.Fatal("resource-limited search discarded its prior model")
	}
	if physical := exactSynthesisReviewModel(t, model, 3, truth); !slices.Equal(pins, physical) {
		t.Fatal("retained solution has incorrect pins")
	}
	calls, hits := s_pSesStore.nSatCalls+s_pSesStore.nUnsatCalls+s_pSesStore.nUndefCalls, s_pSesStore.nCacheHits
	if next := Abc_ExactDelayCost(&truth[0], 3, &arrivals[0], pins, &cost, 2); next != 2 || cost != 2 || s_pSesStore.nCacheHits != hits+1 || s_pSesStore.nSatCalls+s_pSesStore.nUnsatCalls+s_pSesStore.nUndefCalls != calls {
		t.Fatal("resource-limited valid model was not reused")
	}
}
