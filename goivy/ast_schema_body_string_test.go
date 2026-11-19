package goivy

import "testing"

// Python SchemaBody.__str__ (ivy_ast.py) prints the premises and the
// conclusion, one per line, indented; showgoals prints 'theorem ' +
// str(goal). goivy printed the placeholder "{...}" (test/cut1.ivy).
func TestSchemaBodyStringMatchesPython(t *testing.T) {
	cfg := NewAstConfig()
	baz := cfg.NewLabeledFormula(cfg.NewAtom("baz"), NewConst("b", Boolean))
	body := cfg.NewSchemaBody(baz, NewConst("c", Boolean))
	goal := cfg.NewLabeledFormula(cfg.NewAtom("prop2"), body)
	want := "[prop2] {\n    property [baz] b\n    property c\n}\n"
	if got := goal.String(); got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}
