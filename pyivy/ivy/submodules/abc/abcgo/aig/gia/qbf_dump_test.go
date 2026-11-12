package gia

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestQbfDIMACSQuantifiersAndClauses(t *testing.T) {
	g := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, b))
	var output bytes.Buffer
	if err := qbfWriteDIMACS(&output, g, 1); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	var variables, clauses int
	if _, err := fmt.Sscanf(lines[0], "p cnf %d %d", &variables, &clauses); err != nil || variables < 3 || clauses != len(lines)-3 {
		t.Fatalf("header: %q", output.String())
	}
	if lines[1] != "a 1 0" || !strings.HasPrefix(lines[2], "e 2 ") {
		t.Fatalf("quantifiers: %q", output.String())
	}
	if lines[3] != "-3 0" {
		t.Fatalf("constant clause missing: %q", output.String())
	}
}
