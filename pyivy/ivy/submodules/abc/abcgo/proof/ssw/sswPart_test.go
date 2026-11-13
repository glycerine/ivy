package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestPartitionRejectsConstraintsWithoutChangingParameters(t *testing.T) {
	g := aig.NewGraph()
	p := &Ssw_Pars_t{fConstrs: 1, nPartSize: 3, fVerbose: 1}
	if Ssw_SignalCorrespondencePart(g, p) != nil || p.nPartSize != 3 || p.fVerbose != 1 {
		t.Fatal("constraint rejection changed parameters")
	}
}
