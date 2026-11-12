package io

import (
	"os"
	"strings"
	"testing"
)

func TestEDIFSyntaxQuotedRenameAndMultiline(t *testing.T) {
	text, err := os.ReadFile("test_vectors/gates.edif")
	if err != nil {
		t.Fatal(err)
	}
	root, err := parseEdifSyntax("gates.edif", string(text))
	if err != nil {
		t.Fatal(err)
	}
	if len(root.children("external")) != 1 || root.child("design") == nil {
		t.Fatal("lost forms")
	}
	renamed, err := parseEdifSyntax("names.edif", `(edif x (rename internal "quoted (name)"))`)
	if err != nil {
		t.Fatal(err)
	}
	name, err := edifIdentifier(renamed.child("rename"))
	if err != nil || name != "internal" {
		t.Fatal(name, err)
	}
}
func TestEDIFSyntaxMalformedLocations(t *testing.T) {
	for _, text := range []string{"(edif x", "(edif x))", `(edif "unclosed)`, "(wrong x)", "(edif x) (edif y)"} {
		if _, err := parseEdifSyntax("bad.edif", text); err == nil {
			t.Fatal("accepted", text)
		} else if !strings.Contains(err.Error(), "bad.edif") {
			t.Fatal(err)
		}
	}
}
