package io

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

// TestSMVFollowsIoWriteSmv: a strashed network is written through the
// bench netlist, with ABC's sections, init values and specs.
func TestSMVFollowsIoWriteSmv(t *testing.T) {
	source := ".model smv\n.inputs a|b c\n.outputs assert_safety_y\n.latch n q 0\n.names a|b c n\n11 1\n.names q c assert_safety_y\n11 0\n.end\n"
	n, err := ParseBLIF("state.blif", source, false, true)
	if err != nil {
		t.Fatal(err)
	}
	logic := abc.Abc_NtkToLogic(n)
	strashed := abc.Abc_NtkStrash(logic, 0, 1, 0)
	var text bytes.Buffer
	if err = WriteSMV(&text, strashed); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"MODULE main\n\nVAR  -- inputs\n    a_b : boolean;\n", "VAR  -- state variables\n    q : boolean;\n", "init(q) := 0;", "next(q) := ", "INVARSPEC assert_safety_y;", " & "} {
		if !strings.Contains(text.String(), expected) {
			t.Fatalf("missing %q in\n%s", expected, text.String())
		}
	}
}
