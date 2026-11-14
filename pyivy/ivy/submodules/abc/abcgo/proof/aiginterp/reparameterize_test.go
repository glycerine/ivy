package aiginterp

import (
	"strings"
	"testing"
)

func TestReparameterizationPreservesSourceClauseIDFailure(t *testing.T) {
	g := predicate(10)
	defer func() {
		failure := recover()
		message, ok := failure.(string)
		if !ok || !strings.Contains(message, "invalid solver2 problem clause") {
			t.Fatalf("expected source zero-ID rejection, got %v", failure)
		}
		if err := g.Check(); err != nil {
			t.Fatal(err)
		}
	}()
	Aig_ManInterRepar(g, 0)
}
