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

// getVarName and the parser have intentionally different whitespace rules in
// ltl_parser.c: CR, VT and FF are skipped at token starts but kept in names.
func TestLTLSourceNameDelimiters(t *testing.T) {
	for _, separator := range []string{"\r", "\v", "\f"} {
		n, err := ParseFormula("\talpha" + separator + "beta")
		if err != nil || n.Type != BOOL || n.Name != "alpha"+separator+"beta" {
			t.Fatalf("separator %q: node %v error %v", separator, n, err)
		}
		n, err = ParseFormula(separator + "alpha")
		if err != nil || n.Name != "alpha" {
			t.Fatalf("leading %q: node %v error %v", separator, n, err)
		}
	}
}
