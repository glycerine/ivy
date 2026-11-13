package fracomb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestCecFallbackConsumesReplacedManagers(t *testing.T) {
	g := cecPigeonholeMiter()
	original := g
	var swept *aig.Graph
	status := Fra_FraigCec(&g, 100000, 0, func(input *aig.Graph, _, _ int32) *aig.Graph {
		swept = input
		result := aig.Aig_ManDupDfs(input)
		result.Data = make([]int32, result.InputCount())
		return result
	})
	if status != 0 || swept == nil {
		t.Fatalf("status=%d swept=%v", status, swept != nil)
	}
	if original.InputCount() != 0 || swept.InputCount() != 0 {
		t.Fatal("CEC retained a manager replaced during fallback")
	}
	if g.InputCount() == 0 {
		t.Fatal("CEC consumed its result manager")
	}
}
