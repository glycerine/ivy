package abc

import (
	"bytes"
	"strings"
	"testing"
)

func TestFlopDependencySupport(t *testing.T) {
	n := NewNetwork("dependency")
	input := n.NewObject(ObjectPI)
	first := Abc_NtkAddLatch(n, input, ABC_INIT_ZERO)
	second := Abc_NtkAddLatch(n, first, ABC_INIT_ZERO)
	third := Abc_NtkAddLatch(n, n.AndEdges(second.Signal(), input.Signal()).Object, ABC_INIT_ZERO)
	_ = third
	var out bytes.Buffer
	if err := WriteFlopDependency(&out, n); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"0 [label=\"0(1)\"]", "1 [label=\"1(0)\"]", "2 [label=\"2(1)\"]", "   0 ->    1", "   1 ->    2"} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("missing %q in %s", expected, out.String())
		}
	}
}
