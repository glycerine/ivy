package live

import (
	_ "embed"
	"encoding/json"
	"testing"
)

//go:embed test_vectors/ltl_oracle.json
var ltlOracle []byte

func TestLTLOriginalCParserAndFragment(t *testing.T) {
	var fixtures []struct {
		Formula, Tree string
		Well          int32
	}
	if err := json.Unmarshal(ltlOracle, &fixtures); err != nil {
		t.Fatal(err)
	}
	for trial, fixture := range fixtures {
		node, err := ParseFormula(fixture.Formula)
		if err != nil {
			t.Fatalf("trial%d %v", trial, err)
		}
		if node.String() != fixture.Tree || isWellFormed(node) != fixture.Well {
			t.Fatalf("trial%d C/native parser differs: %s", trial, node)
		}
	}
}
