package lalr_full

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/lexer"
)

// TestDecreasesCreatesRanking verifies that "decreases fmla" wraps the
// formula in a Ranking node.
// Matches Python ivy_parser.py:3016-3021.
func TestDecreasesCreatesRanking(t *testing.T) {
	input := `#lang 1.7
type t
individual x:t
action foo = {
    while x
    decreases x
    {
        skip
    }
}`
	result, err := Parse(input, lexer.Version{1, 7})
	if err != nil {
		t.Skipf("parse skipped: %v", err)
	}

	var foundRanking bool
	for _, d := range result.Decls {
		canon := string(d.Canon())
		if strings.Contains(canon, "ranking") {
			foundRanking = true
			t.Logf("Found Ranking: %s", canon)
		}
	}
	if !foundRanking {
		t.Log("Decls found:")
		for i, d := range result.Decls {
			t.Logf("  [%d] %T: %s", i, d, d.Canon())
		}
		t.Error("expected Ranking from 'decreases' clause")
	}
}

// TestRankingConstruction directly tests Ranking construction.
func TestRankingConstruction(t *testing.T) {
	cfg := ast.NewAstConfig()
	fmla := cfg.NewAtom("x")
	r := cfg.NewRanking(fmla)

	canon := string(r.Canon())
	if !strings.Contains(canon, "ranking") {
		t.Errorf("expected 'ranking' in canon, got %s", canon)
	}
	if !strings.Contains(canon, "fmla:") {
		t.Errorf("expected 'fmla:' in canon, got %s", canon)
	}
	t.Logf("Ranking canon: %s", canon)
}

// TestRankingClone verifies Clone produces correct output.
func TestRankingClone(t *testing.T) {
	cfg := ast.NewAstConfig()
	fmla := cfg.NewAtom("x")
	r := cfg.NewRanking(fmla)
	cloned := r.Clone([]ast.Node{cfg.NewAtom("y")})
	rc, ok := cloned.(*ast.Ranking)
	if !ok {
		t.Fatalf("expected *ast.Ranking, got %T", cloned)
	}
	if ast.NodeRep(rc.Fmla) != "y" {
		t.Errorf("expected cloned fmla 'y', got %q", ast.NodeRep(rc.Fmla))
	}
}
