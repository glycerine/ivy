package lalr_full

import (
	"testing"

	"github.com/glycerine/ivy/goivy/lexer"
)

func TestParseEmpty(t *testing.T) {
	result, err := Parse("", lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Decls) != 0 {
		t.Fatalf("expected 0 decls, got %d", len(result.Decls))
	}
}

func TestParseTypeDecl(t *testing.T) {
	result, err := Parse("type node", lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(result.Decls))
	}
	t.Logf("parsed decl: %T", result.Decls[0])
}

func TestParseRelation(t *testing.T) {
	result, err := Parse("relation r(X:node, Y:node)", lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(result.Decls))
	}
	t.Logf("parsed decl: %T", result.Decls[0])
}

func TestParseAxiom(t *testing.T) {
	result, err := Parse("axiom [myax] forall X:t . r(X)", lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(result.Decls))
	}
	t.Logf("parsed decl: %T", result.Decls[0])
}

func TestParseAction(t *testing.T) {
	result, err := Parse("action foo(x:t) = { x := x }", lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(result.Decls))
	}
	t.Logf("parsed decl: %T", result.Decls[0])
}

func TestParseMultiDecl(t *testing.T) {
	src := `type node
type value
relation table(N:node, V:value)
axiom forall N:node, V1:value, V2:value . table(N,V1) & table(N,V2) -> V1 = V2
action write(n:node, v:value) = {
    table(n, V) := V = v
}
export write`
	result, err := Parse(src, lexer.Version{1, 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Decls) < 4 {
		t.Fatalf("expected at least 4 decls, got %d", len(result.Decls))
	}
	for i, d := range result.Decls {
		t.Logf("decl[%d]: %T", i, d)
	}
}
