package abc

import "testing"

func TestAbcExtractSfmMappedNetwork(t *testing.T) {
	n, _, _, pi, driver, consumer := abcSCLFixture(t)
	mapped, objects, err := Abc_NtkExtractSfm(n)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 4 || objects[0] != pi || objects[1] != driver || objects[2] != consumer || objects[3] != n.Outputs()[0] {
		t.Fatalf("object order: %v", objects)
	}
	for _, input := range []bool{false, true} {
		out, err := mapped.Evaluate([]bool{input})
		if err != nil || len(out) != 1 || out[0] != input {
			t.Fatalf("input %v: %v %v", input, out, err)
		}
	}
	consumer.fanins[0].Complement = true
	if _, _, err := Abc_NtkExtractSfm(n); err == nil {
		t.Fatal("complemented mapped edge accepted")
	}
}
