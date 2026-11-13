package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"strings"
	"testing"
)

// A reconvergent fanout graph requires the placement to supply the first
// primary input to two gates before their outputs meet at the root.
func TestMeshReconvergentFanoutWitness(t *testing.T) {
	for _, backend := range []struct {
		name string
		run  func(*gia.Gia_Man_t, int32, int32, int32, int32)
	}{{"satoko", Bmc_MeshTest}, {"bsat", Bmc_MeshTest2}} {
		t.Run(backend.name, func(t *testing.T) {
			graph := gia.Gia_ManStart(16)
			a, b, c := gia.Gia_ManAppendCi(graph), gia.Gia_ManAppendCi(graph), gia.Gia_ManAppendCi(graph)
			x, y := gia.Gia_ManAppendAnd(graph, a, b), gia.Gia_ManAppendAnd(graph, a, c)
			gia.Gia_ManAppendCo(graph, gia.Gia_ManAppendAnd(graph, x, y))
			text := captureMeshOutput(t, func() { backend.run(graph, 5, 5, 4, 0) })
			if !strings.Contains(text, "Satisfying solution found") {
				t.Fatalf("expected placement:\n%s", text)
			}
			checkMeshWitness(t, text, 5, 5, 3, [][2]int{{0, 1}, {0, 2}, {3, 4}})
		})
	}
}
