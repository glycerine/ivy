package fracomb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceCecFailedSweepConsumesInputAndPanics(t *testing.T) {
	g := cecPigeonholeMiter()
	calls := 0
	var swept *aig.Graph
	defer func() {
		if got := recover(); got != "fra: source dereferences failed FRAIG sweep" {
			t.Fatalf("panic=%v", got)
		}
		if calls != 1 || swept == nil || swept.InputCount() != 0 {
			t.Fatalf("calls=%d swept=%v", calls, swept)
		}
	}()
	Fra_FraigCec(&g, 100000, 0, func(input *aig.Graph, _, _ int32) *aig.Graph { calls++; swept = input; return nil })
}
