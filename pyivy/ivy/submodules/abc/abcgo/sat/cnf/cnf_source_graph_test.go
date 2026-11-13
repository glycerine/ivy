package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceGraphOwnership(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const1())
	data, err := EncodeAIG(g, 0)
	if err != nil || data.SourceGraph() != g {
		t.Fatalf("source graph lost: %v", err)
	}
	if NewMappedCNF(0, nil, nil, nil, nil).SourceGraph() != nil {
		t.Fatal("unexpected source graph")
	}
}
